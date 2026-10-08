package mcpserver_test

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// noticed is a store that says, once armed, when the file has been read.
type noticed struct {
	store.Storage
	armed atomic.Bool
	once  sync.Once
	read  chan struct{}
}

func (n *noticed) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	p, revision, err := n.Storage.Load(ctx, account)
	if n.armed.Load() {
		n.once.Do(func() { close(n.read) })
	}
	return p, revision, err
}

// readWithin waits for the store to say the file has been read, which has to
// come within a few seconds.
func readWithin(t *testing.T, read <-chan struct{}) {
	t.Helper()

	select {
	case <-read:
	case <-time.After(5 * time.Second):
		t.Fatal("read_task never read the file, want the question to read it before it waits")
	}
}

// newsLesson is a lesson over kept whose card's questions wait for news no
// longer than hold.
func newsLesson(t *testing.T, kept store.Storage, hold time.Duration) (*harness, *mcp.ClientSession) {
	t.Helper()

	h := newHarness(t)
	service := lessonService(t, h, kept, &clock{at: lessonDay}, nil)
	mcpserver.HoldForNews(service, hold)
	h.start(t, mcpserver.DevSignIn, slices.Concat(service.ProfileTools(), service.TaskTools())...)
	session := h.connect(t, "")
	return h, session
}

// question is a card's question put in the background: its answer comes on
// the channel, with how long it took.
type question struct {
	result *mcp.CallToolResult
	took   time.Duration
	err    error
}

// ask puts a card's question in the background.
func ask(t *testing.T, session *mcp.ClientSession, args map[string]any) <-chan question {
	t.Helper()

	asked := make(chan question, 1)
	began := time.Now()
	go func() {
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "read_task", Arguments: args})
		asked <- question{result: result, took: time.Since(began), err: err}
	}()
	return asked
}

// answeredWithin is a question's answer, which has to come within the bound.
func answeredWithin(t *testing.T, asked <-chan question, bound time.Duration) (awaitedPayload, time.Duration) {
	t.Helper()

	select {
	case got := <-asked:
		if got.err != nil {
			t.Fatalf("read_task failed at the protocol: %v", got.err)
		}
		return payloadOf[awaitedPayload](t, got.result), got.took
	case <-time.After(bound):
		t.Fatalf("read_task did not answer within %v", bound)
		return awaitedPayload{}, 0
	}
}

// A card's question that has heard the task is being written waits for news,
// and hears it as the hand-in lands: the task accepted, or a try turned down.
func TestAHeldQuestionHearsOfTheHandInAsItLands(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		handIn   func(*profile.OpenRequest) map[string]any
		screen   string
		refused  int
		wantTask bool
	}{
		{name: "the task accepted", handIn: raceOn, screen: "task", wantTask: true},
		{name: "a try turned down", handIn: broken, screen: "coming", refused: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := &noticed{Storage: racer(t), read: make(chan struct{})}
			_, session := newsLesson(t, kept, time.Minute)
			request := askForTheRace(t, session, kept)
			if first := awaited(t, session, request.ID); first.Screen != "coming" || first.Refused != 0 {
				t.Fatalf("read_task without what the card heard = %+v, want the task being written, at once", first)
			}

			kept.armed.Store(true)
			asked := ask(t, session, map[string]any{"request_id": request.ID, "refused": 0})
			readWithin(t, kept.read)
			call(t, session, "submit_task", tc.handIn(request))
			got, _ := answeredWithin(t, asked, 5*time.Second)
			if got.Screen != tc.screen || got.Refused != tc.refused || (got.Task != nil) != tc.wantTask {
				t.Errorf("read_task held for news = %+v, want %s with %d tries turned down", got, tc.screen, tc.refused)
			}
		})
	}
}

// A held question with no news is answered when its hold ends, with what the
// card knew.
func TestAQuestionWithNoNewsIsAnsweredWhenItsHoldEnds(t *testing.T) {
	t.Parallel()

	const hold = 300 * time.Millisecond
	kept := racer(t)
	_, session := newsLesson(t, kept, hold)
	request := askForTheRace(t, session, kept)

	got, took := answeredWithin(t, ask(t, session, map[string]any{"request_id": request.ID, "refused": 0}), 5*time.Second)
	if got.Screen != "coming" || got.Refused != 0 {
		t.Errorf("read_task held with no news = %+v, want the task still being written", got)
	}
	if took < hold {
		t.Errorf("read_task held with no news took %v, want its hold, %v", took, hold)
	}
}

// A question is answered at once whenever the answer is news to the card, or
// the card says nothing of what it heard: only a task still being written,
// with as many tries turned down as the card has heard of, waits.
func TestAQuestionWithNewsForTheCardIsAnsweredAtOnce(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		args    func(*profile.OpenRequest) map[string]any
		prepare func(*testing.T, *mcp.ClientSession, *profile.OpenRequest)
		screen  string
	}{
		{
			name:   "a card that says nothing of what it heard",
			args:   func(r *profile.OpenRequest) map[string]any { return map[string]any{"request_id": r.ID} },
			screen: "coming",
		},
		{
			name:   "a try turned down the card has not heard of",
			args:   func(r *profile.OpenRequest) map[string]any { return map[string]any{"request_id": r.ID, "refused": 1} },
			screen: "coming",
		},
		{
			name: "the task on the card",
			args: func(r *profile.OpenRequest) map[string]any { return map[string]any{"request_id": r.ID, "refused": 0} },
			prepare: func(t *testing.T, session *mcp.ClientSession, r *profile.OpenRequest) {
				t.Helper()
				call(t, session, "submit_task", raceOn(r))
			},
			screen: "task",
		},
		{
			name: "a request no file holds",
			args: func(*profile.OpenRequest) map[string]any {
				return map[string]any{"request_id": "req_nobody-holds-this", "refused": 0}
			},
			screen: "waiting",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			_, session := newsLesson(t, kept, time.Minute)
			request := askForTheRace(t, session, kept)
			if tc.prepare != nil {
				tc.prepare(t, session, request)
			}
			if got, _ := answeredWithin(t, ask(t, session, tc.args(request)), 5*time.Second); got.Screen != tc.screen {
				t.Errorf("read_task = %+v, want %s at once", got, tc.screen)
			}
		})
	}
}

// lagging is a store whose first read after a write, once armed, lags behind
// that write, as Drive's may straight after it.
type lagging struct {
	noticed
	wrote atomic.Bool
}

func (l *lagging) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	revision, err := l.Storage.Save(ctx, account, p, expected)
	if err == nil && l.armed.Load() {
		l.wrote.Store(true)
	}
	return revision, err
}

func (l *lagging) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if l.wrote.CompareAndSwap(true, false) {
		return nil, "", fmt.Errorf("drive: the read lags behind the write: %w", store.ErrBehind)
	}
	return l.noticed.Load(ctx, account)
}

// A question the news wakes is answered from the profile the write left, and
// reads no file: a read straight after the write may lag behind it, and one
// that lags lets the next read see the profile from before the write.
func TestAQuestionWokenByNewsAnswersFromTheProfileWritten(t *testing.T) {
	t.Parallel()

	kept := &lagging{noticed: noticed{Storage: racer(t), read: make(chan struct{})}}
	_, session := newsLesson(t, kept, time.Minute)
	request := askForTheRace(t, session, kept)

	kept.armed.Store(true)
	asked := ask(t, session, map[string]any{"request_id": request.ID, "refused": 0})
	readWithin(t, kept.read)
	call(t, session, "submit_task", raceOn(request))
	got, _ := answeredWithin(t, asked, 5*time.Second)
	if got.Screen != "task" || got.Task == nil {
		t.Errorf("read_task woken by the hand-in = %+v, want the task, from the profile the hand-in wrote", got)
	}
	if !kept.wrote.Load() {
		t.Error("the question woken by the hand-in read the file after it, want the profile written answered from")
	}
}

// A write that leaves the task as the card knows it — the adult's note saved
// while the task is written — keeps the question waiting, and the hand-in
// that follows is what the question answers with.
func TestAQuestionKeepsWaitingThroughNewsThatChangesNothingForTheCard(t *testing.T) {
	t.Parallel()

	kept := &noticed{Storage: racer(t), read: make(chan struct{})}
	_, session := newsLesson(t, kept, time.Minute)
	request := askForTheRace(t, session, kept)

	kept.armed.Store(true)
	asked := ask(t, session, map[string]any{"request_id": request.ID, "refused": 0})
	readWithin(t, kept.read)
	call(t, session, "save_profile", map[string]any{"notes": "Likes puzzles."})
	select {
	case <-asked:
		t.Fatal("read_task answered once a note was saved, want it waiting on for news of the task")
	case <-time.After(200 * time.Millisecond):
	}
	call(t, session, "submit_task", raceOn(request))
	if got, _ := answeredWithin(t, asked, 5*time.Second); got.Screen != "task" || got.Task == nil {
		t.Errorf("read_task held through a note saved = %+v, want the task the hand-in accepted", got)
	}
}
