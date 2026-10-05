package mcpserver_test

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The two accounts of the cases about paces, and the tokens that sign them in.
const (
	oneToken, oneAccount         = "the-token-of-one", "one-account"
	anotherToken, anotherAccount = "the-token-of-another", "another-account"
)

// twoAccounts is a reader that vouches for two tokens, each signing in an
// account of its own, and refuses every other.
func twoAccounts(_ context.Context, token string) (store.Account, time.Time, error) {
	switch token {
	case oneToken:
		return store.NewAccount(oneAccount, "a-google-token", time.Time{}), time.Now().Add(time.Hour), nil
	case anotherToken:
		return store.NewAccount(anotherAccount, "another-google-token", time.Time{}), time.Now().Add(time.Hour), nil
	}
	return store.Account{}, time.Time{}, errors.New("a token of nobody's")
}

// counting is a tool that counts the calls that reach it.
func counting(reached *atomic.Int32) mcpserver.Tool {
	return mcpserver.Define(mcpserver.Spec{Name: "count", Title: "Count", Description: "Counts the calls that reach it.", Effect: mcpserver.Reads},
		func(_ context.Context, account store.Account, _ sayIn) (mcpserver.Reply[sayOut], error) {
			reached.Add(1)
			return mcpserver.Reply[sayOut]{Text: "counted", Payload: sayOut{For: account.ID}}, nil
		})
}

// pacedHarness serves the counting tool behind the sign-in given, held to the
// paces given.
func pacedHarness(t *testing.T, signIn mcpserver.SignIn, paces mcpserver.Limits, reached *atomic.Int32) *harness {
	t.Helper()

	h := newHarness(t)
	h.limits = paces
	h.start(t, signIn, counting(reached))
	return h
}

// countAs calls the counting tool with the token given, or with none, and is
// the result the call was answered with.
func (h *harness) countAs(t *testing.T, token string) map[string]any {
	t.Helper()

	header := http.Header{}
	if token != "" {
		header.Set("Authorization", "Bearer "+token)
	}
	result, _ := message(t, h.post(t, legacyCall("count", `{"say":"one more"}`), header))["result"].(map[string]any)
	return result
}

// heldBack says whether a result is the answer to a call a pace held back: one
// sentence telling the model to wait and call again, marked as an error, and
// no payload of the tool's shape.
func heldBack(result map[string]any) bool {
	if result["isError"] != true || result["structuredContent"] != nil {
		return false
	}
	content, _ := result["content"].([]any)
	if len(content) != 1 {
		return false
	}
	block, _ := content[0].(map[string]any)
	text, _ := block["text"].(string)
	return strings.Contains(text, "too many calls in a short time") && strings.Contains(text, "Wait a moment")
}

// hitsOf are the lines of the ceilings reached, as the ceiling and the account
// each names.
func hitsOf(h *harness) [][2]string {
	var named [][2]string
	hits := linesOf(h, "limit_hit")
	for i := range hits {
		limit, _ := hits[i].ContextMap()["limit"].(string)
		user, _ := hits[i].ContextMap()["user"].(string)
		named = append(named, [2]string{limit, user})
	}
	return named
}

// A call past its account's pace is held back before any tool runs. The model
// is told in one sentence to wait and call again, with no payload, since none
// of the tool's shape can be made without the tool; the call's line is a
// refusal of status limited rather than a failure; and the ceiling reached is
// written down once for the flood.
func TestACallPastItsAccountsPaceIsHeldBack(t *testing.T) {
	t.Parallel()

	var reached atomic.Int32
	h := pacedHarness(t, mcpserver.DevSignIn, pacesOf(t, 3, 1<<20), &reached)

	if first := h.countAs(t, ""); first["isError"] == true {
		t.Fatalf("the first call = %v, want it answered", first)
	}
	for i := range 2 {
		if again := h.countAs(t, ""); !heldBack(again) {
			t.Errorf("call %d past the pace = %v, want it held back", i+2, again)
		}
	}
	if got := reached.Load(); got != 1 {
		t.Errorf("the tool ran %d times, want once: a call held back runs nothing", got)
	}

	h.settle()
	var ended [][2]string
	for _, line := range h.toolLines() {
		status, _ := line.ContextMap()["status"].(string)
		if _, failed := line.ContextMap()["error"]; failed {
			t.Errorf("a tool_call line names an error, %v, want a call held back told as a refusal", line.ContextMap())
		}
		ended = append(ended, [2]string{field(t, &line, "outcome"), status})
	}
	if want := [][2]string{{"ok", ""}, {"refused", "limited"}, {"refused", "limited"}}; !slices.Equal(ended, want) {
		t.Errorf("the calls ended %v, want %v", ended, want)
	}
	if hits := hitsOf(h); !slices.Equal(hits, [][2]string{{"user_rate", mcpserver.DevAccount}}) {
		t.Errorf("limit_hit lines = %v, want one, naming user_rate and the account", hits)
	}
	if warned := h.logs.FilterMessage("limit_hit").FilterLevelExact(zapcore.WarnLevel).Len(); warned != 1 {
		t.Errorf("limit_hit warnings = %d, want the line a warning", warned)
	}
}

// One account past its pace costs no other account anything.
func TestOneAccountPastItsPaceCostsNoOtherAccountAnything(t *testing.T) {
	t.Parallel()

	var reached atomic.Int32
	h := pacedHarness(t, mcpserver.BearerSignIn(twoAccounts, metadata), pacesOf(t, 3, 1<<20), &reached)
	h.countAs(t, oneToken)
	if runaway := h.countAs(t, oneToken); !heldBack(runaway) {
		t.Fatalf("one account past its pace = %v, want it held back", runaway)
	}

	if other := h.countAs(t, anotherToken); other["isError"] == true {
		t.Errorf("another account's call = %v, want it answered", other)
	}
	if got := reached.Load(); got != 2 {
		t.Errorf("the tool ran %d times, want twice: once for each account", got)
	}
}

// The instance's pace holds every account together: past it, an account that
// has spent nothing of its own is held back too, the ceiling named is the
// instance's, and the call it turned away is given back to the account.
func TestTheInstancesPaceHoldsEveryAccountTogether(t *testing.T) {
	t.Parallel()

	var reached atomic.Int32
	paces := pacesOf(t, 3, 3)
	h := pacedHarness(t, mcpserver.BearerSignIn(twoAccounts, metadata), paces, &reached)
	h.countAs(t, oneToken)

	if other := h.countAs(t, anotherToken); !heldBack(other) {
		t.Errorf("another account past the instance's pace = %v, want it held back", other)
	}
	h.settle()
	if hits := hitsOf(h); !slices.Equal(hits, [][2]string{{"instance_rate", anotherAccount}}) {
		t.Errorf("limit_hit lines = %v, want one, naming instance_rate", hits)
	}
	if got := paces.PerAccount.Take(anotherAccount); !got.Allowed {
		t.Errorf("the account's own pace after = %+v, want the call the instance turned away given back", got)
	}
}

// A message that is no tool call — a host listing the tools, or reading the
// widget's page — counts against the pace too, and past it is refused in the
// protocol's own terms, since no model reads it.
func TestAMessageThatIsNoToolCallIsHeldToThePaceToo(t *testing.T) {
	t.Parallel()

	var reached atomic.Int32
	h := pacedHarness(t, mcpserver.DevSignIn, pacesOf(t, 3, 1<<20), &reached)
	list := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	if first := message(t, h.post(t, list, nil)); first["error"] != nil {
		t.Fatalf("the first list = %v, want it answered", first)
	}

	refused, _ := message(t, h.post(t, list, nil))["error"].(map[string]any)
	words, _ := refused["message"].(string)
	if refused["code"] != float64(-32000) || !strings.Contains(words, "too many requests") {
		t.Errorf("the list past the pace = %v, want the protocol's error for too many requests", refused)
	}
	if call := h.countAs(t, ""); !heldBack(call) {
		t.Errorf("a call after it = %v, want it held back: the list spent the account's pace", call)
	}
}

// dayOf is a child of grade 2 whose day's counters stand as given.
func dayOf(t *testing.T, daily profile.Daily) store.Storage {
	t.Helper()

	p := profile.New(profile.Student{Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}}, "test", lessonDay)
	p.Daily = daily
	return keptAsIs(t, p)
}

// A day at a ceiling of its own opens no request. The refusal is limited, as
// limit_reached, in the words the child is to hear — that the new tasks are
// over for today, that there will be more tomorrow, and what they can do
// meanwhile — with no number in them, and which ceiling it was is in the line
// alone. Nothing is written.
func TestADayAtItsCeilingOpensNoRequest(t *testing.T) {
	t.Parallel()

	today := profile.DateOf(lessonDay)
	for _, tc := range []struct {
		name  string
		daily profile.Daily
		limit string
		count int64
	}{
		{"every task of the day given", profile.Daily{Date: today, Accepted: config.DefaultDailyTasks}, "daily_tasks", config.DefaultDailyTasks},
		{"every failure of the day spent", profile.Daily{Date: today, Failed: config.DefaultDailyFailed}, "daily_failed", config.DefaultDailyFailed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := dayOf(t, tc.daily)
			h, session := lesson(t, kept)
			_, revision := loadKept(t, kept)

			wantTheDayFull(t, call(t, session, "next_task", raceChoice))
			if p, now := loadKept(t, kept); now != revision || p.OpenRequest != nil {
				t.Error("a day at its ceiling wrote the profile, want nothing written")
			}

			h.settle()
			wantOneHit(t, h, tc.limit, tc.count)
			if got := field(t, h.lineOf(t, "next_task"), "status"); got != "limited" {
				t.Errorf("the call's status = %q, want limited", got)
			}
		})
	}
}

// wantTheDayFull holds a result of next_task to the refusal of a day with no
// room left: limited, as limit_reached, naming no problem, on a card that
// knows whose it is, in words that say what happened, when it clears and what
// the child can do meanwhile, and no number.
func wantTheDayFull(t *testing.T, result *mcp.CallToolResult) {
	t.Helper()

	if got := payloadOf[requestPayload](t, result); got.Screen != "waiting" || got.Status != "limited" ||
		got.Code != "limit_reached" || len(got.Problems) != 0 || got.RequestID != "" || got.Child == nil {
		t.Errorf("next_task = %+v, want limited as limit_reached, naming no problem, on the child's card", got)
	}
	words := textOf(t, result)
	for _, said := range []string{"no more new tasks today", "more tomorrow", "look at their progress"} {
		if !strings.Contains(words, said) {
			t.Errorf("the words are %q, want them to say %q", words, said)
		}
	}
	if strings.ContainsAny(words, "0123456789") {
		t.Errorf("the words are %q, want no number in them", words)
	}
	if strings.Contains(words, "had") {
		t.Errorf("the words are %q, want no claim that the child had tasks: a day of failed requests gave none", words)
	}
}

// wantOneHit holds the lines of a case to one ceiling reached, the one named,
// with the counter's value and the account.
func wantOneHit(t *testing.T, h *harness, limit string, count int64) {
	t.Helper()

	hits := linesOf(h, "limit_hit")
	if len(hits) != 1 || hits[0].ContextMap()["limit"] != limit || hits[0].ContextMap()["count"] != count {
		t.Fatalf("limit_hit lines = %v, want one naming %s at %d", hits, limit, count)
	}
	if hits[0].Level != zapcore.WarnLevel {
		t.Errorf("the line is at %v, want a warning, as every limit_hit is", hits[0].Level)
	}
	if got := field(t, &hits[0], "user"); got != mcpserver.DevAccount {
		t.Errorf("the line is for %q, want the account", got)
	}
}

// A day with room left opens a request: one short of each ceiling, and a day
// whose counters are yesterday's, however full, since yesterday's tasks never
// hold today back.
func TestADayWithRoomLeftOpensARequest(t *testing.T) {
	t.Parallel()

	today, yesterday := profile.DateOf(lessonDay), profile.DateOf(lessonDay.AddDate(0, 0, -1))
	for _, tc := range []struct {
		name  string
		daily profile.Daily
	}{
		{"one task short", profile.Daily{Date: today, Accepted: config.DefaultDailyTasks - 1}},
		{"one failure short", profile.Daily{Date: today, Failed: config.DefaultDailyFailed - 1}},
		{"yesterday full", profile.Daily{Date: yesterday, Accepted: 99, Failed: 99}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := dayOf(t, tc.daily)
			_, session := lesson(t, kept)
			askForTheRace(t, session, kept)
		})
	}
}

// A request opened before the day filled up is still handed back: its task is
// already being written, and handing it back costs the day nothing more than
// the one task it was opened for.
func TestARequestOpenedBeforeTheDayFilledUpIsHandedBack(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	first := askForTheRace(t, session, kept)

	p, revision := loadKept(t, kept)
	p.Daily = profile.Daily{Date: profile.DateOf(lessonDay), Accepted: config.DefaultDailyTasks}
	p.Touch("test", lessonDay)
	if _, err := kept.Save(t.Context(), devAccount, p, revision); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	again := call(t, session, "next_task", raceChoice)
	if card := wantComing(t, again); card.RequestID != first.ID || !card.AlreadyOpen ||
		!strings.Contains(textOf(t, again), "Request "+first.ID+" is already open") {
		t.Errorf("next_task on a full day with a request open = %+v, %q, want that request handed back",
			card, textOf(t, again))
	}
}
