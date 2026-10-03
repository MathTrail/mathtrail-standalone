package mcpserver_test

import (
	"context"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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

// An answer another tab recorded while this one was on its way is found
// recorded when the call reads again, and told as it was: it counts once.
func TestAnAnswerRecordedByAnotherTabMeanwhileCountsOnce(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	raced := racingOver(keptAsIs(t, p))
	h, session := lesson(t, raced)
	answer := map[string]any{"task_id": p.CurrentTask.ID, "answer": "C"}

	held, landed := raceTwoCalls(t, h, session, raced, "submit_answer", answer, answer)
	if told := answered(t, landed); told.AlreadyAnswered {
		t.Errorf("the tab that landed was told %+v, want the answer recorded anew", told)
	}
	if told := answered(t, held); !told.AlreadyAnswered {
		t.Errorf("the tab that was overtaken was told %+v, want the answer told again", told)
	}
	h.settle()
	if lines := linesOf(h, "answer_recorded"); len(lines) != 1 {
		t.Errorf("answer_recorded lines = %d, want the one answer", len(lines))
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
	result := answerIt(t, session, p.CurrentTask.ID, "C", false)
	if result.IsError {
		t.Fatalf("submit_answer for a profile deleted meanwhile failed: %s", textOf(t, result))
	}
	if told := payloadOf[answerPayload](t, result); told.Screen != "first_run" || told.Status != "stale" {
		t.Errorf("submit_answer for a profile deleted meanwhile = %+v, want the first run", told)
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
