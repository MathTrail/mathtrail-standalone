package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/logger"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// A write the child need not wait for is made after the call has answered.
// The answer leaves as soon as it is ready, and the request that carried it is
// held open until the write is done: the platform keeps an instance's CPU for
// a request that is still open, and goes on doing so once its client has
// left, so the write is finished at full speed.

// lateDeadline is how long a write after the answer may take, its tries again
// included; holdBound is how long a request is held open for it, a little
// longer, so that a write that ends at its deadline is waited for. Both stay
// well under the time the server gives a response.
const (
	lateDeadline = 30 * time.Second
	holdBound    = lateDeadline + time.Second
)

// lateAttempts is how many times a write after the answer is made: the first,
// and again from a fresh read for each time the file had moved on or could not
// be reached.
const lateAttempts = 3

// How a write after the answer went, as its line says it.
const (
	// lateWritten is a change written as it was made.
	lateWritten = "written"
	// lateRemade is a change made again on a fresh read, because the file
	// moved on or could not be reached, and written then.
	lateRemade = "remade"
	// lateAlready is a change the file held already: another writer made it.
	lateAlready = "already"
	// lateLost is a change the file no longer allows: another writer took it
	// out of reach, and the answer the call gave is not what the file holds.
	lateLost = "lost"
	// lateFailed is a change that could not be written.
	lateFailed = "failed"
)

// eventWriteAfterAnswer is the line a write after the answer leaves.
const eventWriteAfterAnswer = "write_after_answer"

// heldKey holds the writes a request is held open for. It is a type of its
// own, so that nothing else a context carries can be taken for it.
type heldKey struct{}

// heldWrites are the writes after the answer the calls of one request began.
type heldWrites struct {
	mu   sync.Mutex
	done []<-chan struct{}
}

// holdOpen hands a request on, and once it has answered holds it open until
// the writes its calls began after the answer are done, or holdBound has
// passed.
func holdOpen(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		held := &heldWrites{}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), heldKey{}, held)))
		held.wait(holdBound)
	})
}

// add is one more write the request is held open for.
func (h *heldWrites) add(done <-chan struct{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.done = append(h.done, done)
}

// wait waits for every write the request began, but no longer than bound.
func (h *heldWrites) wait(bound time.Duration) {
	h.mu.Lock()
	writes := slices.Clone(h.done)
	h.mu.Unlock()
	if len(writes) == 0 {
		return
	}
	limit := time.NewTimer(bound)
	defer limit.Stop()
	for _, done := range writes {
		select {
		case <-done:
		case <-limit.C:
			return
		}
	}
}

// pending are the writes after the answer this instance has begun and not yet
// finished, by account. A call waits for its account's before it touches the
// profile, so that it finds what an answer said rather than the file as it was
// before.
type pending struct {
	mu     sync.Mutex
	writes map[string][]chan struct{}
}

// newPending is an instance with no write after the answer begun.
func newPending() *pending {
	return &pending{writes: map[string][]chan struct{}{}}
}

// begin records a write of the account as begun, and is what ends it.
func (p *pending) begin(account string) (done <-chan struct{}, end func()) {
	write := make(chan struct{})
	p.mu.Lock()
	p.writes[account] = append(p.writes[account], write)
	p.mu.Unlock()
	return write, func() {
		p.mu.Lock()
		left := slices.DeleteFunc(p.writes[account], func(other chan struct{}) bool { return other == write })
		if len(left) == 0 {
			delete(p.writes, account)
		} else {
			p.writes[account] = left
		}
		p.mu.Unlock()
		close(write)
	}
}

// wait waits for the writes of the account begun before it, and for none begun
// after: a call is held back by what came before it, never by what came since.
func (p *pending) wait(ctx context.Context, account string) error {
	p.mu.Lock()
	begun := slices.Clone(p.writes[account])
	p.mu.Unlock()
	for _, write := range begun {
		select {
		case <-write:
		case <-ctx.Done():
			return fmt.Errorf("mcp: wait for the profile to be written: %w", ctx.Err())
		}
	}
	return nil
}

// settled is the store as the tools reach it: every operation on an account
// waits first for the writes after the answer this instance has begun for it.
type settled struct {
	store.Storage
	pending *pending
}

// Load reads the profile once the writes begun before it are done.
func (s settled) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if err := s.pending.wait(ctx, account.ID); err != nil {
		return nil, "", err
	}
	return s.Storage.Load(ctx, account)
}

// Create keeps the first profile once the writes begun before it are done.
func (s settled) Create(ctx context.Context, account store.Account, p *profile.Profile) (store.Revision, error) {
	if err := s.pending.wait(ctx, account.ID); err != nil {
		return "", err
	}
	return s.Storage.Create(ctx, account, p)
}

// Save writes the profile once the writes begun before it are done.
func (s settled) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	if err := s.pending.wait(ctx, account.ID); err != nil {
		return "", err
	}
	return s.Storage.Save(ctx, account, p, expected)
}

// Export says where the file is once the writes begun before it are done.
func (s settled) Export(ctx context.Context, account store.Account) (store.Location, error) {
	if err := s.pending.wait(ctx, account.ID); err != nil {
		return store.Location{}, err
	}
	return s.Storage.Export(ctx, account)
}

// Restore mends the file once the writes begun before it are done.
func (s settled) Restore(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if err := s.pending.wait(ctx, account.ID); err != nil {
		return nil, "", err
	}
	return s.Storage.Restore(ctx, account)
}

// StartOver sets the file aside once the writes begun before it are done.
func (s settled) StartOver(ctx context.Context, account store.Account, p *profile.Profile) (store.Revision, error) {
	if err := s.pending.wait(ctx, account.ID); err != nil {
		return "", err
	}
	return s.Storage.StartOver(ctx, account, p)
}

// change is a call's change to a profile, made on whatever profile it is
// given: on the one the call read, and on a fresh one when that one has moved
// on by the time it is written.
type change func(p *profile.Profile) (made, error)

// made is how a change went on one profile, and the lines it leaves once the
// file holds it.
type made struct {
	state  madeState
	landed func(context.Context)
}

// madeState is whether a profile took a change, held it already, or no longer
// allows it.
type madeState int

const (
	// changed is a profile the change was made on.
	changed madeState = iota
	// already is a profile that holds the change already.
	already
	// gone is a profile the change no longer applies to.
	gone
)

// lateWrite is a change a call made to the profile it read, and answered with,
// to be written after the answer: the tool that made it, whose profile, the
// profile changed and touched, the revision it was read at, the change itself,
// for a fresh read, and the lines the change leaves once the file holds it.
type lateWrite struct {
	tool     string
	account  store.Account
	profile  *profile.Profile
	revision store.Revision
	again    change
	landed   func(context.Context)
}

// writeAfterAnswer writes a call's change once the call has answered, when the
// request that carries the call is held open for it, and at once otherwise —
// a call that came by no endpoint, as a test's may. A change written at once
// fails the call as any write does; one written after the answer can no
// longer tell the call, and its line says how it went. The call hands the
// write over as its last act, and reads the profile no more.
func (s *Service) writeAfterAnswer(ctx context.Context, w *lateWrite) error {
	held, holding := ctx.Value(heldKey{}).(*heldWrites)
	if !holding {
		if _, err := s.store.Save(ctx, w.account, w.profile, w.revision); err != nil {
			return fmt.Errorf("mcp: save the profile: %w", err)
		}
		w.landed(ctx)
		return nil
	}
	done, end := s.pending.begin(w.account.ID)
	held.add(done)
	go func() {
		defer end()
		defer s.recoverLate(ctx, w)
		s.land(context.WithoutCancel(ctx), w)
	}()
	return nil
}

// land writes a change after its call has answered. When the file has moved
// on since the read, or could not be reached, the file is read again and the
// same change made on what is there, while the change still applies and tries
// are left. An upload Drive failed on may have landed all the same: a fresh
// read that holds the change at the number this write gave the profile is
// this write, landed. Any other change the file holds already is another
// writer's, and leaves no lines of this call. Whatever happens is said in one
// line, which carries nothing the child could be known by.
func (s *Service) land(ctx context.Context, w *lateWrite) {
	ctx, cancel := context.WithTimeout(ctx, lateDeadline)
	defer cancel()
	// The lines are the call's, as every line of a call is, so that they are
	// joined to it; the calls to the store go inside a span of the write's own.
	call := ctx
	ctx, span := s.tracer.Start(ctx, eventWriteAfterAnswer)
	defer span.End()
	began := time.Now()
	finish := func(outcome string, attempts int, err error) {
		if outcome == lateLost || outcome == lateFailed {
			span.SetStatus(codes.Error, outcome)
		}
		s.lateLine(call, w, outcome, attempts, time.Since(began), err)
	}

	p, revision, landed, outcome := w.profile, w.revision, w.landed, lateWritten
	for attempt := 1; ; attempt++ {
		_, err := s.direct.Save(ctx, w.account, p, revision)
		switch {
		case err == nil:
			landed(call)
			finish(outcome, attempt, nil)
			return
		case attempt == lateAttempts || !worthAgain(err) || ctx.Err() != nil:
			finish(lateFailed, attempt, err)
			return
		}
		mayHaveLanded := errors.Is(err, store.ErrUnavailable)
		fresh, freshRevision, err := s.readAgain(ctx, w.account)
		switch {
		case errors.Is(err, store.ErrNotFound):
			finish(lateLost, attempt, nil)
			return
		case err != nil:
			finish(lateFailed, attempt, err)
			return
		}
		again, err := w.again(fresh)
		switch {
		case err != nil:
			finish(lateFailed, attempt, err)
			return
		case again.state == already && mayHaveLanded && fresh.Revision == p.Revision:
			landed(call)
			finish(outcome, attempt, nil)
			return
		case again.state == already:
			finish(lateAlready, attempt, nil)
			return
		case again.state == gone:
			finish(lateLost, attempt, nil)
			return
		}
		p, revision, landed, outcome = fresh, freshRevision, again.landed, lateRemade
	}
}

// worthAgain says whether a write that failed may go through when made again
// on a fresh read: the file moved on, was not found where it was, or could not
// be reached.
func worthAgain(err error) bool {
	return errors.Is(err, store.ErrConflict) || errors.Is(err, store.ErrNotFound) ||
		errors.Is(err, store.ErrUnavailable)
}

// readAgain reads the profile afresh once a short pause has passed, and once
// more when the read lags behind a write or cannot reach the file.
func (s *Service) readAgain(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	for read := 1; ; read++ {
		if err := sleep(ctx, pause()); err != nil {
			return nil, "", err
		}
		p, revision, err := s.direct.Load(ctx, account)
		lagged := errors.Is(err, store.ErrBehind) || errors.Is(err, store.ErrUnavailable)
		if !lagged || read == 2 {
			return p, revision, err
		}
	}
}

// recoverLate turns a panic in a write after the answer into the line of a
// write that failed, as a panic in a call is turned into a failure, so that
// the instance and every other call on it go on. The goroutine the write runs
// in defers it.
func (s *Service) recoverLate(ctx context.Context, w *lateWrite) {
	recovered := recover()
	if recovered == nil {
		return
	}
	s.events.logger.Error(eventWriteAfterAnswer, append([]zap.Field{
		zap.String("tool", w.tool),
		zap.String("outcome", lateFailed),
		zap.String("error", kindPanic),
		zap.String("instructions_version", s.events.instructionsVersion),
		logger.PanicValue(recovered),
		logger.PanicStack(),
	}, callerFields(ctx, w.account.ID, s.events.projectID)...)...)
}

// lateLine leaves the line of a write after the answer: the tool, how it went,
// how many tries it took, how long it took from its hand-over, as the call
// answered, to its end, the
// version of the instructions, and, when it failed, what stopped it, in the
// words a call's line uses. A write that lost or failed is a warning: the
// answer the call gave is not what the file holds.
func (s *Service) lateLine(ctx context.Context, w *lateWrite, outcome string, attempts int, took time.Duration, err error) {
	fields := []zap.Field{
		zap.String("tool", w.tool),
		zap.String("outcome", outcome),
		zap.Int("attempts", attempts),
		zap.Int64("duration_ms", took.Milliseconds()),
		zap.String("instructions_version", s.events.instructionsVersion),
	}
	if err != nil {
		fields = append(fields, zap.String("error", explain(err).kind))
	}
	level := zapcore.InfoLevel
	if outcome == lateLost || outcome == lateFailed {
		level = zapcore.WarnLevel
	}
	s.events.logger.Log(level, eventWriteAfterAnswer,
		append(fields, callerFields(ctx, w.account.ID, s.events.projectID)...)...)
}
