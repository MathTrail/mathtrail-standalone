package drivestore

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/logger"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

// scope is what the spans of the calls to Drive are recorded under, so that a
// reader can tell them from the ones a library produced.
const scope = "github.com/MathTrail/mathtrail-standalone/internal/store/drive"

// eventDriveCall is the line every call to Drive leaves.
const eventDriveCall = "drive_call"

// The calls made to Drive, as a span and a line name them.
const (
	opList     = "list"
	opGet      = "get"
	opDownload = "download"
	opCreate   = "create"
	opUpdate   = "update"
	// opRevisions lists a file's history, opRevision reads one state of it,
	// opKeep keeps one forever and opDelete deletes one.
	opRevisions = "revisions"
	opRevision  = "revision"
	opKeep      = "keep"
	opDelete    = "delete"
)

// How a call to Drive ended, as a span and a line say it: a closed list, never
// the text of an error.
const (
	outcomeOK          = "ok"
	outcomeNotFound    = "not_found"
	outcomeRevoked     = "revoked"
	outcomeExpired     = "expired"
	outcomeRateLimited = "rate_limited"
	outcomeStorageFull = "storage_full"
	outcomeTooLarge    = "too_large"
	outcomeNotKept     = "not_kept"
	outcomeUnavailable = "unavailable"
	outcomeTimeout     = "timeout"
	outcomeCanceled    = "canceled"
	outcomeFailed      = "failed"
)

// pauses are how long a call Drive asked to pause, or failed, waits before it
// is tried again: twice, about half a second and then a second and a half.
// Drive's own advice is to back off up to half a minute, which no chat host
// waits for; this rides out a blip, and a real block reaches the parent as
// Drive out of reach.
var pauses = []time.Duration{500 * time.Millisecond, 1500 * time.Millisecond}

// parentsDrive is the Drive of the parent an account belongs to, as the store
// reaches it: every call made with the account's token, and every one watched.
type parentsDrive struct {
	files     drive.Files
	revisions drive.Revisions
	account   store.Account
	watch     *watch
}

// driveOf is the Drive the account's token reaches.
func (s *driveStore) driveOf(account store.Account) parentsDrive {
	return parentsDrive{files: s.files, revisions: s.revisions, account: account, watch: s.watch}
}

// find is the service's own files in the Drive that carry the marker, the most
// recently changed first: outside the bin, or in it.
func (p parentsDrive) find(ctx context.Context, marker string, inBin bool) ([]drive.File, error) {
	var found []drive.File
	err := p.watch.call(ctx, p.account, opList, func(ctx context.Context) (err error) {
		found, err = p.files.List(ctx, p.account.Token(), drive.Query{Key: markerKey, Value: marker, InBin: inBin})
		return err
	})
	return found, err
}

// get is what a file is called.
func (p parentsDrive) get(ctx context.Context, id string) (drive.File, error) {
	var file drive.File
	err := p.watch.call(ctx, p.account, opGet, func(ctx context.Context) (err error) {
		file, err = p.files.Get(ctx, p.account.Token(), id)
		return err
	})
	return file, err
}

// download is what a file holds, as far as a profile ever runs.
func (p parentsDrive) download(ctx context.Context, id string) ([]byte, error) {
	var raw []byte
	err := p.watch.call(ctx, p.account, opDownload, func(ctx context.Context) (err error) {
		raw, err = p.files.Download(ctx, p.account.Token(), id, maxFile)
		return err
	})
	return raw, err
}

// create makes a file, or a folder when there is no content.
func (p parentsDrive) create(ctx context.Context, meta *drive.File, content []byte) (drive.File, error) {
	var made drive.File
	err := p.watch.call(ctx, p.account, opCreate, func(ctx context.Context) (err error) {
		made, err = p.files.Create(ctx, p.account.Token(), meta, content)
		return err
	})
	return made, err
}

// update makes a change to a file, and answers the file as it is afterwards.
func (p parentsDrive) update(ctx context.Context, id string, change *drive.Change) (drive.File, error) {
	var changed drive.File
	err := p.watch.call(ctx, p.account, opUpdate, func(ctx context.Context) (err error) {
		changed, err = p.files.Update(ctx, p.account.Token(), id, change)
		return err
	})
	return changed, err
}

// history is every revision of a file Drive keeps, in no order.
func (p parentsDrive) history(ctx context.Context, id string) ([]drive.Revision, error) {
	var revisions []drive.Revision
	err := p.watch.call(ctx, p.account, opRevisions, func(ctx context.Context) (err error) {
		revisions, err = p.revisions.List(ctx, p.account.Token(), id)
		return err
	})
	return revisions, err
}

// downloadRevision is what one revision of a file held.
func (p parentsDrive) downloadRevision(ctx context.Context, id, revision string) ([]byte, error) {
	var raw []byte
	err := p.watch.call(ctx, p.account, opRevision, func(ctx context.Context) (err error) {
		raw, err = p.revisions.Download(ctx, p.account.Token(), id, revision, maxFile)
		return err
	})
	return raw, err
}

// keep keeps a revision of a file forever.
func (p parentsDrive) keep(ctx context.Context, id, revision string) error {
	return p.watch.call(ctx, p.account, opKeep, func(ctx context.Context) error {
		return p.revisions.Keep(ctx, p.account.Token(), id, revision)
	})
}

// delete deletes a revision of a file for good.
func (p parentsDrive) delete(ctx context.Context, id, revision string) error {
	return p.watch.call(ctx, p.account, opDelete, func(ctx context.Context) error {
		return p.revisions.Delete(ctx, p.account.Token(), id, revision)
	})
}

// watch makes every call to Drive the same way, and records it: a span inside
// the span of whatever asked for it, and a line.
type watch struct {
	tracer    trace.Tracer
	logger    *zap.Logger
	projectID string
	// timeout is how long one call to Drive may take.
	timeout time.Duration
	now     func() time.Time
	sleep   func(ctx context.Context, d time.Duration) error
}

// call makes one call to Drive on behalf of the account, as many times as it
// is worth trying, and records how it ended.
//
// No attempt is started that could outlive the account's token: it would be
// refused halfway, and a refusal of Drive's for a token that ended reads the
// same as access taken back. A call Drive refused for a pause, or failed on
// its side, is tried again after one of the pauses — unless it may have landed
// all the same, which only a write can have done.
//
// Neither the span nor the line names the file or carries anything it holds:
// the outcome is a word of a closed list, and a failure is marked with a
// sentence of ours rather than the error's own.
func (w *watch) call(ctx context.Context, account store.Account, op string, do func(context.Context) error) error {
	started := time.Now()
	ctx, span := w.tracer.Start(ctx, "drive "+op, trace.WithSpanKind(trace.SpanKindClient))
	retries, err := 0, w.attempt(ctx, account, do)
	for ; err != nil && retries < len(pauses) && worthRepeating(op, err); retries++ {
		if waited := w.sleep(ctx, jittered(pauses[retries])); waited != nil {
			err = fmt.Errorf("drivestore: %s: %w", op, waited)
			break
		}
		err = w.attempt(ctx, account, do)
	}

	outcome := outcomeOf(err)
	span.SetAttributes(attribute.String("mathtrail.drive.outcome", outcome), attribute.Int("mathtrail.drive.retries", retries))
	failed := isFailure(outcome)
	if failed {
		span.SetStatus(codes.Error, "drive did not do what was asked")
	}
	span.End()

	fields := append([]zap.Field{
		zap.String("op", op),
		zap.Int64("duration_ms", time.Since(started).Milliseconds()),
		zap.Int("retries", retries),
		zap.String("outcome", outcome),
	}, w.callerFields(ctx, account)...)
	if failed {
		w.logger.Warn(eventDriveCall, fields...)
	} else {
		w.logger.Info(eventDriveCall, fields...)
	}
	return err
}

// attempt makes the call once, if the account's token outlasts it.
func (w *watch) attempt(ctx context.Context, account store.Account, do func(context.Context) error) error {
	if ends := account.Ends(); !ends.IsZero() && w.now().Add(w.timeout).After(ends) {
		return fmt.Errorf("drivestore: %w: the token ends before a call could", store.ErrAccessExpired)
	}
	return do(ctx)
}

// worthRepeating reports whether a call that ended in err is worth making
// again: Drive asked for a pause, which it asks for before it does anything,
// or it failed on its side on a call that changes nothing, or nothing more
// when made twice. A creation or an upload Drive failed on may have landed all
// the same, and making it again could lay a stale state over a newer one or
// make a second file; whoever asked reads again instead. A call that ran out
// of its time, or that its caller gave up on, is never made again.
func worthRepeating(op string, err error) bool {
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return false
	case errors.Is(err, drive.ErrRateLimited):
		return true
	case errors.Is(err, drive.ErrUnavailable):
		return op != opCreate && op != opUpdate
	}
	return false
}

// jittered is a pause spread by a quarter either way, so that the calls of two
// instances Drive asked to pause at once do not come back at once.
func jittered(pause time.Duration) time.Duration {
	//nolint:gosec // the spread only keeps retries apart; nothing depends on guessing it
	return time.Duration(float64(pause) * (0.75 + rand.Float64()/2))
}

// outcomeOf is how a call ended, in the words a span and a line use. A call
// that ran out of time, or that its caller gave up on, is told as that before
// anything it wraps.
func outcomeOf(err error) string {
	switch {
	case err == nil:
		return outcomeOK
	case errors.Is(err, context.DeadlineExceeded):
		return outcomeTimeout
	case errors.Is(err, context.Canceled):
		return outcomeCanceled
	case errors.Is(err, store.ErrAccessExpired):
		return outcomeExpired
	case errors.Is(err, drive.ErrNotFound):
		return outcomeNotFound
	case errors.Is(err, drive.ErrUnauthorized):
		return outcomeRevoked
	case errors.Is(err, drive.ErrRateLimited):
		return outcomeRateLimited
	case errors.Is(err, drive.ErrStorageFull):
		return outcomeStorageFull
	case errors.Is(err, drive.ErrTooLarge):
		return outcomeTooLarge
	case errors.Is(err, drive.ErrNotKept):
		return outcomeNotKept
	case errors.Is(err, drive.ErrUnavailable):
		return outcomeUnavailable
	}
	return outcomeFailed
}

// isFailure reports whether a call ended in a failure, of Drive's or of ours,
// rather than in an answer about the parent's Drive: a file that is not
// there, access taken back, a full Drive, a file too large and a revision not
// kept are all things Drive says about the parent's side, and nothing went
// wrong in saying them. Nor did anything go wrong in a call its caller gave up
// on. A call not made because the token would end first is a failure of ours:
// the token was issued to outlast the calls it lets in.
func isFailure(outcome string) bool {
	switch outcome {
	case outcomeRateLimited, outcomeUnavailable, outcomeTimeout, outcomeExpired, outcomeFailed:
		return true
	}
	return false
}

// callerFields are what every line about the account's profile carries beside
// its own: the request, the user and the trace it belongs to.
func (w *watch) callerFields(ctx context.Context, account store.Account) []zap.Field {
	return append([]zap.Field{
		zap.String("request_id", logger.RequestID(ctx)),
		zap.String("user", account.ID),
	}, telemetry.LogFields(ctx, w.projectID)...)
}
