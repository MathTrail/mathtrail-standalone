package mcpserver_test

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark/starlarktest"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The endpoint has no sessions and one sandbox for everybody, so the one thing
// that stands between a crowd of hand-ins and a mixed-up or lost answer is the
// endpoint itself. Two children hold both slots of the sandbox with solvers
// only a clock stops, and three more hand in their tasks meanwhile: each of the
// three is turned away in a sentence that says the task was not checked, while
// the two are still running, and spends nothing. Then all five hand their
// tasks in at once — more calls than slots — and each is answered with its own
// card. Nothing on the way is a server's failure or a panic, and afterwards the
// slots are free again.
func TestMoreTasksThanSlotsEachGetTheirOwnAnswerAndNoneWaitsForEver(t *testing.T) {
	t.Parallel()

	// Two slots, and steps enough that only the clock stops a run.
	runner, err := starlark.New(starlark.Limits{
		Steps: 1 << 62, Timeout: 3 * time.Second, Concurrency: 2, Wait: 1500 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("starlark.New() error = %v", err)
	}
	kept := memory.New()
	h := newHarness(t)
	service := lessonService(t, h, kept, &clock{at: lessonDay}, runner)
	h.start(t, mcpserver.DevSignIn, slices.Concat(service.ProfileTools(), service.TaskTools())...)

	var failed atomic.Int32
	children := make([]*racingChild, len(entrants))
	for i, entrant := range entrants {
		children[i] = racingChildOf(t, h, kept, entrant, &failed)
	}
	holders, extras := children[:2], children[2:]

	// The run: the holders take both slots, and the others arrive while the
	// holders' clocks are still running.
	holding := make(chan []handedIn)
	holdersRaces := racesOf(t, holders, clockBound)
	go func() { holding <- handInTogether(t.Context(), holders, holdersRaces) }()
	starlarktest.EverySlotTaken(t, runner)
	turnedAway := handInTogether(t.Context(), extras, racesOf(t, extras, ""))
	held := <-holding

	wantHeldToTheirClocks(t, held)
	wantTurnedAwayMeanwhile(t, kept, turnedAway, held)

	// After the run: everybody at once, more calls than slots.
	wantEachTheirOwnCard(t, handInTogether(t.Context(), children, racesOf(t, children, "")), holders)

	starlarktest.WantEverySlotFree(t, runner, 2)
	h.settle()
	wantNothingOfOursFailed(t, h, &failed, len(extras))
}

// wantHeldToTheirClocks holds the hand-ins that took the slots to what their
// solvers earned: stopped by the clock, the attempt refused as the solver's,
// each on the card of its own child.
func wantHeldToTheirClocks(t *testing.T, held []handedIn) {
	t.Helper()

	for _, answer := range held {
		got := payloadOf[handedInPayload](t, answer.of(t))
		if got.Status != "rejected" || got.Code != "solver_error" {
			t.Errorf("%s held a slot to its clock: status %q, code %q, want rejected as solver_error",
				answer.child.name, got.Status, got.Code)
		}
		if got.Child == nil || got.Child.Pseudonym != answer.child.pseudonym {
			t.Errorf("%s's card is for %+v, want %q", answer.child.name, got.Child, answer.child.pseudonym)
		}
	}
}

// wantTurnedAwayMeanwhile holds the hand-ins that found every slot taken to a
// refusal told while the slots were still taken — before any holder's run
// ended — that spent nothing.
func wantTurnedAwayMeanwhile(t *testing.T, kept store.Storage, turnedAway, held []handedIn) {
	t.Helper()

	for _, answer := range turnedAway {
		wantOurSentence(t, answer.of(t), "MathTrail is checking too many tasks right now")
		for _, holder := range held {
			if !answer.ended.Before(holder.ended) {
				t.Errorf("%s was answered at %v, after %s's run ended at %v; want it turned away while the slots were taken",
					answer.child.name, answer.ended, holder.child.name, holder.ended)
			}
		}
		if spent := answer.child.openRequest(t, kept).Attempts; spent != 0 {
			t.Errorf("%s spent %d attempts on a task nothing checked, want none", answer.child.name, spent)
		}
	}
}

// wantEachTheirOwnCard holds every hand-in to its own answer: accepted, with
// the question that child handed in, on that child's card, at the attempt that
// child had come to — the second for a holder, whose first the clock spent,
// and the first for everybody else.
func wantEachTheirOwnCard(t *testing.T, answers []handedIn, holders []*racingChild) {
	t.Helper()

	for _, answer := range answers {
		got := payloadOf[handedInPayload](t, answer.of(t))
		if got.Task == nil {
			t.Errorf("%s's task: status %q, code %q, want it accepted", answer.child.name, got.Status, got.Code)
			continue
		}
		if got.Task.Question != answer.child.question {
			t.Errorf("%s's card asks %q, want their own question %q", answer.child.name, got.Task.Question, answer.child.question)
		}
		if got.Child == nil || got.Child.Pseudonym != answer.child.pseudonym {
			t.Errorf("%s's card is for %+v, want %q", answer.child.name, got.Child, answer.child.pseudonym)
		}
		wantAttempt := 1
		if slices.Contains(holders, answer.child) {
			wantAttempt = 2
		}
		if got.Attempt != wantAttempt {
			t.Errorf("%s's task was accepted at attempt %d, want %d", answer.child.name, got.Attempt, wantAttempt)
		}
	}
}

// wantNothingOfOursFailed holds the whole case to the service never giving up:
// no answer of 500 or more, no panic, no call told in the general sentence, and
// one busy line for each task turned away.
func wantNothingOfOursFailed(t *testing.T, h *harness, failed *atomic.Int32, turnedAway int) {
	t.Helper()

	if n := failed.Load(); n != 0 {
		t.Errorf("%d answers with a status of 500 or more, want none", n)
	}
	if panics := linesOf(h, "mcp_panic"); len(panics) != 0 {
		t.Errorf("%d mcp_panic lines, want none", len(panics))
	}
	busy := 0
	lines := h.toolLines()
	for i := range lines {
		switch kind, _ := lines[i].ContextMap()["error"].(string); kind {
		case "internal", "panic":
			t.Errorf("a tool_call line names the failure %q, want none of ours", kind)
		case "busy":
			busy++
		}
	}
	if busy != turnedAway {
		t.Errorf("tool_call lines naming busy = %d, want %d, one for each task turned away", busy, turnedAway)
	}
}

// entrants are the children of the case: whom each is signed in as, what the
// card calls them, and the names their race is run by, so that no two of them
// hand in the same question. The names are of one syllable, as the race's own
// are, so that the wording stays as easy to read as the race it came from.
var entrants = []entrant{
	{name: "child-0", pseudonym: "Otter", runners: [3]string{"Ann", "Ben", "Kim"}},
	{name: "child-1", pseudonym: "Badger", runners: [3]string{"Liz", "Max", "Tom"}},
	{name: "child-2", pseudonym: "Heron", runners: [3]string{"Pam", "Dan", "Joe"}},
	{name: "child-3", pseudonym: "Lynx", runners: [3]string{"Sam", "Ned", "Kit"}},
	{name: "child-4", pseudonym: "Marten", runners: [3]string{"Rob", "Jen", "Hal"}},
}

// entrant is who a child of the case is.
type entrant struct {
	name, pseudonym string
	runners         [3]string
}

// clockBound is a solver nothing but its clock stops, which holds its slot for
// all of that time.
const clockBound = "def solve(options):\n    while True:\n        pass\n    return []\n"

// racingChild is a child of the case, signed in and with a request open.
type racingChild struct {
	entrant
	account  store.Account
	session  *mcp.ClientSession
	request  *profile.OpenRequest
	renamed  *strings.Replacer
	question string
}

// racingChildOf keeps a profile for the child, signs a session in as them, and
// asks for the race, so that the child has a request open to hand it in for.
func racingChildOf(t *testing.T, h *harness, kept store.Storage, who entrant, failed *atomic.Int32) *racingChild {
	t.Helper()

	account := store.NewAccount(mcpserver.DevAccount+"-"+who.name, "", time.Time{})
	// A child with an interest: the brief's setting is made of it, and a brief
	// with none would be handed back refused.
	student := profile.Student{Grade: 2, Pseudonym: who.pseudonym, Interests: []string{"sport"}}
	if _, err := kept.Create(context.Background(), account, profile.New(student, "test", lessonDay)); err != nil {
		t.Fatalf("keep the profile of %s: %v", who.name, err)
	}

	session, err := h.connectThrough(t, "claude-code", bearing{token: who.name, failed: failed})
	if err != nil {
		t.Fatalf("connect as %s: %v", who.name, err)
	}

	renamed := strings.NewReplacer("Ann", who.runners[0], "Ben", who.runners[1], "Kim", who.runners[2])
	child := &racingChild{
		entrant: who, account: account, session: session, renamed: renamed, question: renamed.Replace(raceQuestion),
	}
	call(t, session, "next_task", raceChoice)
	child.request = child.openRequest(t, kept)
	return child
}

// openRequest is the child's open request as the store keeps it now.
func (c *racingChild) openRequest(t *testing.T, kept store.Storage) *profile.OpenRequest {
	t.Helper()

	p, _, err := kept.Load(context.Background(), c.account)
	if err != nil || p.OpenRequest == nil {
		t.Fatalf("%s's profile: error %v, want a request open", c.name, err)
	}
	return p.OpenRequest
}

// racesOf are the races of the children, each handed in for the child's own
// request — the brief as it was received, and every name in the task, its
// solver and its self-check the child's own — with the solver given, or the
// race's own when none is.
func racesOf(t *testing.T, children []*racingChild, solverSource string) []map[string]any {
	t.Helper()

	races := make([]map[string]any, len(children))
	for i, child := range children {
		race := raceOn(child.request)
		if solverSource != "" {
			race["solver"] = solverSource
		}
		for _, part := range []string{"task", "solver", "self_check"} {
			written, err := json.Marshal(race[part])
			if err != nil {
				t.Fatalf("encode the %s of %s's race: %v", part, child.name, err)
			}
			var renamed any
			if err := json.Unmarshal([]byte(child.renamed.Replace(string(written))), &renamed); err != nil {
				t.Fatalf("decode the %s of %s's race: %v", part, child.name, err)
			}
			race[part] = renamed
		}
		races[i] = race
	}
	return races
}

// handedIn is one hand-in as the child saw it: what came back, and when.
type handedIn struct {
	child  *racingChild
	result *mcp.CallToolResult
	err    error
	ended  time.Time
}

// of is the answer, holding the call to having been answered at all.
func (h handedIn) of(t *testing.T) *mcp.CallToolResult {
	t.Helper()

	if h.err != nil {
		t.Fatalf("%s's hand-in: CallTool error = %v, want an answer", h.child.name, h.err)
	}
	return h.result
}

// handInTogether hands every child's race in at the same moment, and is what
// came back to each, in the order of the children.
func handInTogether(ctx context.Context, children []*racingChild, races []map[string]any) []handedIn {
	answers := make([]handedIn, len(children))
	var together sync.WaitGroup
	for i, child := range children {
		together.Go(func() {
			result, err := child.session.CallTool(ctx, &mcp.CallToolParams{Name: "submit_task", Arguments: races[i]})
			answers[i] = handedIn{child: child, result: result, err: err, ended: time.Now()}
		})
	}
	together.Wait()
	return answers
}
