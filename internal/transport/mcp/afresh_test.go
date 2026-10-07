package mcpserver_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// racing is a store another tab gets to first: the first write a call makes
// is held just before it reaches the store, until the case has had another
// call write in between. The held write then finds the profile moved on.
type racing struct {
	store.Storage
	once sync.Once
	// reached is closed when the first write is held, and letGo lets it on.
	reached chan struct{}
	letGo   chan struct{}
}

func racingOver(kept store.Storage) *racing {
	return &racing{Storage: kept, reached: make(chan struct{}), letGo: make(chan struct{})}
}

func (r *racing) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	first := false
	r.once.Do(func() { first = true })
	if first {
		close(r.reached)
		<-r.letGo
	}
	return r.Storage.Save(ctx, account, p, expected)
}

// raceTwoCalls makes a call that is held at its first write while another tab
// makes a second one, which lands; the held call then goes on. It answers the
// held call's result and the second one's.
func raceTwoCalls(t *testing.T, h *harness, session *mcp.ClientSession, raced *racing, tool string, first, second any) (held, landed *mcp.CallToolResult) {
	t.Helper()

	otherTab := h.connect(t, "")
	done := make(chan *mcp.CallToolResult, 1)
	go func() {
		result, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: first})
		if err != nil {
			result = nil
		}
		done <- result
	}()
	<-raced.reached
	landed = call(t, otherTab, tool, second)
	close(raced.letGo)
	if held = <-done; held == nil {
		t.Fatalf("the held %s call failed at the protocol", tool)
	}
	return held, landed
}

// A write that lost to another tab is made again from what won: nothing the
// first read computed is sent, so both changes stay.
func TestAWriteThatLostIsMadeAgainFromWhatWon(t *testing.T) {
	t.Parallel()

	raced := racingOver(keptWith(t, "masha"))
	h, session := lesson(t, raced)

	held, landed := raceTwoCalls(t, h, session, raced, "save_profile",
		map[string]any{"interests": []string{"chess"}}, map[string]any{"notes": "Likes puzzles."})
	for _, result := range []*mcp.CallToolResult{held, landed} {
		if result.IsError {
			t.Errorf("a save that raced another: %s", textOf(t, result))
		}
	}
	p, _ := loadKept(t, raced)
	if !slices.Equal(p.Student.Interests, []string{"chess"}) || p.Student.Notes != "Likes puzzles." {
		t.Errorf("the profile holds interests %v and notes %q, want both changes", p.Student.Interests, p.Student.Notes)
	}
}

// An answer another instance recorded while this one's write was on its way
// counts once. Each tab read the task without an answer, and was told its own
// recorded, since a call answers before it writes; the write that comes second
// finds the answer recorded, and leaves no line of it.
func TestAnAnswerRecordedOnAnotherInstanceMeanwhileCountsOnce(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	raced := racingOver(keptAsIs(t, p))
	h, session := lesson(t, raced)
	other, otherTab := lesson(t, raced)
	answer := map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"}

	first := callEarly(t, session, "submit_answer", answer)
	<-raced.reached
	second := call(t, otherTab, "submit_answer", answer)
	close(raced.letGo)
	h.settle()
	other.settle()

	for _, told := range []*mcp.CallToolResult{first, second} {
		if got := answered(t, told); got.AlreadyAnswered {
			t.Errorf("a tab was told %+v, want the answer recorded, as it read the task without one", got)
		}
	}
	if lines := linesOf(other, "answer_recorded"); len(lines) != 1 {
		t.Errorf("answer_recorded lines of the instance that wrote first = %d, want 1", len(lines))
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 0 {
		t.Errorf("answer_recorded lines of the instance whose write came second = %d, want none", len(lines))
	}
	if outcome := lateOutcome(t, h, "submit_answer"); outcome != "already" {
		t.Errorf("the write that came second went %q, want already", outcome)
	}
}

// unreached is a store another tab gets to first while the first write cannot
// reach the file: that write is held until the case has had other calls write
// in between, and then fails as Drive out of reach, having written nothing.
type unreached struct{ *racing }

func (u unreached) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	first := false
	u.once.Do(func() { first = true })
	if first {
		close(u.reached)
		<-u.letGo
		return "", fmt.Errorf("%w: no answer in time", store.ErrUnavailable)
	}
	return u.Storage.Save(ctx, account, p, expected)
}

// A write that could not reach the file, and finds on the fresh read the same
// answer recorded by another instance on top of a change of its own, does not
// take that answer for its own: the file holds it at another number than the
// write gave the profile, so the write leaves no line, and the answer counts
// once.
func TestAnUnreachedWriteDoesNotCountAnotherInstancesAnswer(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	raced := unreached{racingOver(keptAsIs(t, p))}
	h, session := lesson(t, raced)
	other, otherTab := lesson(t, raced)
	answer := map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"}

	callEarly(t, session, "submit_answer", answer)
	<-raced.reached
	call(t, otherTab, "save_profile", map[string]any{"notes": "Likes puzzles."})
	call(t, otherTab, "submit_answer", answer)
	close(raced.letGo)
	h.settle()
	other.settle()

	if lines := linesOf(other, "answer_recorded"); len(lines) != 1 {
		t.Errorf("answer_recorded lines of the instance that wrote the answer = %d, want 1", len(lines))
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 0 {
		t.Errorf("answer_recorded lines of the instance whose write did not reach the file = %d, want none", len(lines))
	}
	if outcome := lateOutcome(t, h, "submit_answer"); outcome != "already" {
		t.Errorf("the write that did not reach the file went %q, want already", outcome)
	}
}

// Another answer to the same task, recorded on another instance while this
// one's write was on its way, keeps the task: each tab was told its own answer
// recorded, the file holds the one written first, and the write that came
// second is lost, which its line says, as a warning.
func TestAnotherAnswerRecordedOnAnotherInstanceMeanwhileLosesTheWrite(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	raced := racingOver(keptAsIs(t, p))
	h, session := lesson(t, raced)
	other, otherTab := lesson(t, raced)

	callEarly(t, session, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"})
	<-raced.reached
	call(t, otherTab, "submit_answer", map[string]any{"task_id": p.CurrentTask.ID, "answer": "A"})
	close(raced.letGo)
	h.settle()
	other.settle()

	if kept, _ := loadKept(t, raced); kept.CurrentTask.Answered == nil || kept.CurrentTask.Answered.Choice != "A" {
		t.Errorf("the file holds the answer %+v, want A, written first", kept.CurrentTask.Answered)
	}
	if line := lateLine(t, h, "submit_answer"); line["outcome"] != "lost" || line["level"] != zapcore.WarnLevel {
		t.Errorf("the write that came second = %v, want lost, as a warning", line)
	}
	if lines := linesOf(h, "answer_recorded"); len(lines) != 0 {
		t.Errorf("answer_recorded lines of the instance whose write came second = %d, want none", len(lines))
	}
}

// A request another tab opened meanwhile is the one the call hands back: a
// second request beside it would race the first for the same attempts.
func TestARequestOpenedByAnotherTabMeanwhileIsHandedBack(t *testing.T) {
	t.Parallel()

	raced := racingOver(racer(t))
	h, session := lesson(t, raced)

	held, landed := raceTwoCalls(t, h, session, raced, "next_task", raceChoice, raceChoice)
	p, _ := loadKept(t, raced)
	if p.OpenRequest == nil {
		t.Fatal("no request is open after two asks, want one")
	}
	if card, text := wantComing(t, landed), textOf(t, landed); card.RequestID != p.OpenRequest.ID || card.AlreadyOpen ||
		!strings.Contains(text, "Request "+p.OpenRequest.ID+" is open") {
		t.Errorf("the tab that landed drew %+v and said %q, want request %s opened", card, text, p.OpenRequest.ID)
	}
	if card, text := wantComing(t, held), textOf(t, held); card.RequestID != p.OpenRequest.ID || !card.AlreadyOpen ||
		!strings.Contains(text, "Request "+p.OpenRequest.ID+" is already open") {
		t.Errorf("the tab that was overtaken drew %+v and said %q, want request %s handed back", card, text, p.OpenRequest.ID)
	}
}

// A task another tab handed in meanwhile, and had accepted, closed the
// request: the call made again finds it closed, spends nothing, and shows the
// task on the card.
func TestATaskAcceptedByAnotherTabMeanwhileIsNotHandedInTwice(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, asking := lesson(t, kept)
	request := askForTheRace(t, asking, kept)
	raced := racingOver(kept)
	h, session := lesson(t, raced)

	held, landed := raceTwoCalls(t, h, session, raced, "submit_task", raceOn(request), raceOn(request))
	wantOnTheCard(t, landed)
	told := payloadOf[handedInPayload](t, held)
	if told.Status != "stale" || told.Code != "stale_request" || told.Task == nil {
		t.Errorf("the tab that was overtaken was told %+v, want the request closed and the task on the card", told)
	}
	h.settle()
	wantLessonLines(t, h, map[string]int{"task_submitted": 1, "task_accepted": 1})
}

// counted is a store every write of which somebody else gets to first, which
// counts how many times it was tried.
type counted struct {
	store.Storage
	writes atomic.Int32
}

func (c *counted) Save(context.Context, store.Account, *profile.Profile, store.Revision) (store.Revision, error) {
	c.writes.Add(1)
	return "", store.ErrConflict
}

// A call whose write loses every time is made three times, and then told as a
// conflict.
func TestAWriteThatKeepsLosingIsToldAfterThreeTries(t *testing.T) {
	t.Parallel()

	kept := &counted{Storage: keptWith(t, "masha")}
	h, session := lesson(t, kept)
	wantOurSentence(t, call(t, session, "save_profile", map[string]any{"grade": 4}),
		"The child's profile was changed somewhere else at the same moment")
	if got := kept.writes.Load(); got != 3 {
		t.Errorf("the write was tried %d times, want 3", got)
	}
	h.settle()
	wantFailed(t, h, "save_profile", "conflict")
}

// vanishing is a store whose profile is deleted for good between a call's read
// and its write: the write finds no profile, and neither does any read after.
type vanishing struct {
	store.Storage
	gone atomic.Bool
}

func (v *vanishing) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if v.gone.Load() {
		return nil, "", store.ErrNotFound
	}
	return v.Storage.Load(ctx, account)
}

func (v *vanishing) Save(context.Context, store.Account, *profile.Profile, store.Revision) (store.Revision, error) {
	v.gone.Store(true)
	return "", store.ErrNotFound
}

// A profile deleted for good between the read and the write is found gone
// when the call reads again, and told as no profile — not as a fault of ours.
func TestAProfileDeletedBeforeTheWriteIsToldAsGone(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	_, session := lesson(t, &vanishing{Storage: keptAsIs(t, p)})
	result := call(t, session, "next_task", raceChoice)
	if result.IsError {
		t.Fatalf("next_task for a profile deleted meanwhile failed: %s", textOf(t, result))
	}
	if told := payloadOf[requestPayload](t, result); told.Screen != "first_run" {
		t.Errorf("next_task for a profile deleted meanwhile = %+v, want the first run", told)
	}
}

// lostOnce is a store whose first read is refused as a conflict, as a read
// that set out to put a damaged file back finds it mended meanwhile.
type lostOnce struct {
	store.Storage
	once sync.Once
}

func (l *lostOnce) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	conflict := false
	l.once.Do(func() { conflict = true })
	if conflict {
		return nil, "", store.ErrConflict
	}
	return l.Storage.Load(ctx, account)
}

// A read the store could not finish because the file moved on is made again,
// by the tools that only read as much as by those that write.
func TestAReadThatFoundTheFileMovingIsMadeAgain(t *testing.T) {
	t.Parallel()

	for _, tool := range []string{"get_profile", "get_progress", "read_progress"} {
		t.Run(tool, func(t *testing.T) {
			t.Parallel()

			_, session := lesson(t, &lostOnce{Storage: keptWith(t, "masha")})
			if result := call(t, session, tool, map[string]any{}); result.IsError {
				t.Errorf("%s after a read that found the file moving: %s", tool, textOf(t, result))
			}
		})
	}
}
