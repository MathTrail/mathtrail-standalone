package mcpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The tools that take a task from the rule to the child. Every task here is a
// race of three runners, written the way a model is asked to write one — a
// question, five options, a solver in Starlark, a self-check — and handed in
// against a request opened for it with the model's own choice of topic, level
// and difficulty.

// raceChoice asks for the race: its topic, level and difficulty, with a reason.
var raceChoice = map[string]any{
	"language": "en", "topic": "logic.ordering", "grade_level": "1-2", "difficulty": 2,
	"reason": "The child asked for a race.",
}

// raceSolver proves the race by trying every order of the three.
const raceSolver = `def solve(options):
    firsts = []
    for order in permutations(["Ann", "Ben", "Kim"]):
        place = {name: i for i, name in enumerate(order)}
        if place["Ben"] < place["Kim"] and place["Kim"] < place["Ann"]:
            firsts.append(order[0])
    return match(options, firsts[0])
`

// What would give the race away: its solution, the explanations behind its
// wrong options and the traps they are for. None of it may be seen anywhere
// before the child has answered.
var (
	raceSolution  = "Ben is ahead of Kim and Kim is ahead of Ann, so Ben crosses the line first."
	raceExplained = []string{
		"Ann finished after Kim, so she came last.",
		"Kim sits in the middle: Ben beat her.",
		"In any race somebody always finishes first.",
		"The three finished one after another.",
	}
	raceTraps = []string{"reversed_relation", "stopped_early", "ignored_condition", "answered_other_question"}
)

// raceQuestion is the race's wording, which the child reads.
const raceQuestion = "Ann, Ben and Kim ran a race. Ben finished before Kim. Ann finished after Kim. Who finished first?"

// raceOn is the race handed in for a request, the brief handed back as it was
// received.
func raceOn(request *profile.OpenRequest) map[string]any {
	return map[string]any{
		"request_id": request.ID,
		"brief":      request.Brief,
		"task":       raceTask(raceDistractors()),
		"solver":     raceSolver,
		"self_check": map[string]any{
			"issues": []any{},
			"option_check": map[string]string{
				"A": "Ann is last.", "B": "Kim is second.", "C": "Ben is first.", "D": "Someone was first.", "E": "Nobody tied.",
			},
			"final_answer": "C",
		},
	}
}

// raceTask is the race as the model writes it, with these explanations behind
// its wrong options.
func raceTask(distractors map[string]map[string]string) map[string]any {
	return map[string]any{
		"core_idea":              "Order three runners from two comparisons.",
		"design_thought_process": "Plot: a race. Traps: a reversed comparison, stopping early.",
		"question":               raceQuestion,
		"options":                map[string]string{"A": "Ann", "B": "Kim", "C": "Ben", "D": "Nobody", "E": "All at once"},
		"correct_answer":         "C",
		"hint":                   "Who finished before Kim?",
		"solution":               raceSolution,
		"distractors":            distractors,
	}
}

// raceDistractors are the wrong options of the race, each with its trap and
// what the child who chose it is told.
func raceDistractors() map[string]map[string]string {
	return map[string]map[string]string{
		"A": {"trap": raceTraps[0], "text": raceExplained[0]},
		"B": {"trap": raceTraps[1], "text": raceExplained[1]},
		"D": {"trap": raceTraps[2], "text": raceExplained[2]},
		"E": {"trap": raceTraps[3], "text": raceExplained[3]},
	}
}

// broken is the race with two faults, reported in this order: no hint, which
// is its structure, and an explanation too short to say anything.
func broken(request *profile.OpenRequest) map[string]any {
	distractors := raceDistractors()
	distractors["E"]["text"] = "No."
	task := raceTask(distractors)
	task["hint"] = ""

	race := raceOn(request)
	race["task"] = task
	return race
}

// racer is a store holding a child of grade 2 who has answered nothing yet.
func racer(t *testing.T, excluded ...string) store.Storage {
	t.Helper()

	return keptAsIs(t, profile.New(profile.Student{
		Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}, ExcludedSkills: excluded,
	}, "test", lessonDay))
}

// askForTheRace asks for the race and is the request opened for it, as the
// profile keeps it.
func askForTheRace(t *testing.T, session *mcp.ClientSession, kept store.Storage) *profile.OpenRequest {
	t.Helper()

	asked := leadOf(wantWordsAlone(t, call(t, session, "next_task", raceChoice)))
	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil || !strings.Contains(asked, "Request "+p.OpenRequest.ID+" is open") {
		t.Fatalf("next_task said %q and the profile holds %+v, want the one request", asked, p.OpenRequest)
	}
	return p.OpenRequest
}

// refusedRequestPayload is what next_task hands back when it opens no request:
// the one result of the tool with a payload.
type refusedRequestPayload struct {
	Screen   string `json:"screen"`
	Status   string `json:"status"`
	Code     string `json:"code"`
	Problems []struct {
		Field string `json:"field"`
	} `json:"problems"`
}

// wantWordsAlone holds a result of next_task to carrying no payload, and is its
// words. What the tool says is for the model alone, and a host that shows the
// model a payload in place of the words would show it nothing to write from.
func wantWordsAlone(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()

	if result.IsError || result.StructuredContent != nil {
		t.Fatalf("next_task = %+v, %s, want words alone", result.StructuredContent, textOf(t, result))
	}
	return textOf(t, result)
}

// leadOf is the words of next_task before the package they carry.
func leadOf(text string) string {
	lead, _, _ := strings.Cut(text, "\n\nPackage:\n")
	return lead
}

// handedInPayload is what submit_task hands a card.
type handedInPayload struct {
	Screen  string `json:"screen"`
	Status  string `json:"status"`
	Code    string `json:"code"`
	Reasons []struct {
		Code     string   `json:"code"`
		Messages []string `json:"messages"`
	} `json:"reasons"`
	Attempt      int  `json:"attempt"`
	AttemptsLeft *int `json:"attempts_left"`
	Child        *struct {
		Pseudonym  string  `json:"pseudonym"`
		Grade      int     `json:"grade"`
		UILanguage *string `json:"ui_language"`
	} `json:"child"`
	Task *struct {
		ID       string            `json:"id"`
		Topic    string            `json:"topic"`
		Language string            `json:"language"`
		Question string            `json:"question"`
		Options  map[string]string `json:"options"`
		Hint     string            `json:"hint"`
	} `json:"task"`
}

// packageIn is the package the words of next_task carry after their lead.
func packageIn(t *testing.T, text string) map[string]json.RawMessage {
	t.Helper()

	_, pack, found := strings.Cut(text, "\n\nPackage:\n")
	var parts map[string]json.RawMessage
	if !found || json.Unmarshal([]byte(pack), &parts) != nil {
		t.Fatalf("the words carry no package after their lead:\n%s", text)
	}
	return parts
}

// linesOf are the lines of one event.
func linesOf(h *harness, event string) []observer.LoggedEntry {
	return h.logs.FilterMessage(event).All()
}

// The whole way, as a model goes it: a task asked for with a choice of the
// model's own, the package it is written from, the task handed in and
// accepted at the first attempt — on the child's card without its answer,
// sealed in the profile with it, and every step of it a line.
func TestATaskGoesFromTheRuleToTheChildsCard(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)

	open := wantRequestOpened(t, call(t, session, "next_task", raceChoice), kept)
	handed := call(t, session, "submit_task", raceOn(open))
	card := wantOnTheCard(t, handed)
	if text := textOf(t, handed); !strings.HasPrefix(text, "Accepted at attempt 1") || !strings.Contains(text, raceQuestion) {
		t.Errorf("the words are %q, want the acceptance and the question to read out", text)
	}
	wantHandedOut(t, kept, card.Task.ID)

	h.settle()
	wantLessonLines(t, h, map[string]int{
		"task_requested": 1, "solver_run": 2, "task_submitted": 1, "task_accepted": 1,
	})
	if accepted := linesOf(h, "task_accepted")[0].ContextMap(); accepted["topic"] != "logic.ordering" ||
		accepted["attempts"] != int64(1) {
		t.Errorf("task_accepted = %v, want the race accepted at the first attempt", accepted)
	}
}

// wantRequestOpened holds what next_task answered the race with to a new
// request, whose package carries its brief, and is the request as the profile
// keeps it: the model's choice, in English, with no attempt spent.
func wantRequestOpened(t *testing.T, asked *mcp.CallToolResult, kept store.Storage) *profile.OpenRequest {
	t.Helper()

	lead := leadOf(wantWordsAlone(t, asked))
	var brief profile.Brief
	if err := json.Unmarshal(packageIn(t, textOf(t, asked))["brief"], &brief); err != nil || brief.TargetConcept != "logic.ordering" {
		t.Errorf("the package carries the brief %+v (%v), want the race's topic", brief, err)
	}
	p, _ := loadKept(t, kept)
	open := p.OpenRequest
	if open == nil || !strings.Contains(lead, "Request "+open.ID+" is open") || open.Language != "en" || open.Attempts != 0 ||
		open.TutorMode != profile.TutorLLM || open.Brief.Difficulty != 2 {
		t.Fatalf("the request is %+v, want the model's choice in English with no attempt spent", open)
	}
	return open
}

// wantOnTheCard holds what submit_task answered to the race on the child's card
// at the first attempt, beside the name and the grade of the child it is for,
// and no language of the cards' own: the parent chose none.
func wantOnTheCard(t *testing.T, handed *mcp.CallToolResult) handedInPayload {
	t.Helper()

	card := payloadOf[handedInPayload](t, handed)
	if card.Screen != "task" || card.Status != "" || card.Attempt != 1 || card.Task == nil {
		t.Fatalf("submit_task = %+v, want the task on the card at the first attempt", card)
	}
	if card.Child == nil || card.Child.Pseudonym != "Otter" || card.Child.Grade != 2 || card.Child.UILanguage != nil {
		t.Errorf("child = %+v, want Otter of grade 2 beside the task, the cards in the chat's language", card.Child)
	}
	if card.Task.Question != raceQuestion || len(card.Task.Options) != 5 || card.Task.Hint == "" ||
		card.Task.Language != "en" || card.Task.Topic != "logic.ordering" {
		t.Errorf("task = %+v, want the race as the child reads it", card.Task)
	}
	return card
}

// wantHandedOut holds the profile to the race handed out as this task: in
// flight with its answer sealed, the request closed, the task counted and
// remembered.
func wantHandedOut(t *testing.T, kept store.Storage, id string) {
	t.Helper()

	p, _ := loadKept(t, kept)
	if p.CurrentTask == nil || p.CurrentTask.ID != id || p.OpenRequest != nil || p.Daily.Accepted != 1 ||
		len(p.TaskFingerprints) != 1 || p.Topics["logic.ordering"].LastIssued.IsZero() {
		t.Errorf("the profile holds task %+v, request %+v, %d accepted, %d fingerprints, want the task handed out",
			p.CurrentTask, p.OpenRequest, p.Daily.Accepted, len(p.TaskFingerprints))
	}
	if secret, err := p.OpenTask(sealer(t)); err != nil || secret.Answer != "C" || secret.Solution != raceSolution {
		t.Errorf("the sealed part opens to %+v, %v, want the race's answer and solution", secret, err)
	}
}

// wantLessonLines holds the lines of the lesson to how many of each there are,
// and every one of them to carrying the request, the account and the version
// of the instructions.
func wantLessonLines(t *testing.T, h *harness, want map[string]int) {
	t.Helper()

	loaded, _ := shipped()
	for _, event := range slices.Sorted(maps.Keys(want)) {
		lines := linesOf(h, event)
		if len(lines) != want[event] {
			t.Errorf("%s lines = %d, want %d", event, len(lines), want[event])
		}
		for i := range lines {
			fields := lines[i].ContextMap()
			if fields["request_id"] == "" || fields["user"] != devAccount.ID ||
				fields["instructions_version"] != loaded.InstructionsVersion() {
				t.Errorf("a %s line carries %v, want the request, the account and the instructions", event, fields)
			}
		}
	}
}

// A task that fails its checks spends an attempt and is told every reason at
// once, the first of them the code the attempt is counted by; the card waits.
// The same request takes the task again, mended.
func TestARefusedTaskSpendsAnAttemptAndHearsEveryReason(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)

	refused := payloadOf[handedInPayload](t, call(t, session, "submit_task", broken(request)))
	codes := make([]string, 0, len(refused.Reasons))
	for _, reason := range refused.Reasons {
		codes = append(codes, reason.Code)
	}
	if refused.Status != "rejected" || refused.Code != "bad_structure" || refused.Screen != "waiting" ||
		refused.Attempt != 1 || refused.AttemptsLeft == nil || *refused.AttemptsLeft != 2 || refused.Task != nil {
		t.Errorf("submit_task = %+v, want the first attempt refused with two left and the card waiting", refused)
	}
	if want := []string{"bad_structure", "distractor_explanations"}; !slices.Equal(codes, want) {
		t.Errorf("reasons = %v, want %v", codes, want)
	}
	if refused.Child == nil || refused.Child.Pseudonym != "Otter" {
		t.Errorf("child = %+v, want the waiting card to know whom it waits for", refused.Child)
	}
	if p, _ := loadKept(t, kept); p.OpenRequest == nil || p.OpenRequest.Attempts != 1 || p.CurrentTask != nil {
		t.Errorf("the profile holds %+v and task %+v, want one attempt spent and nothing handed out",
			p.OpenRequest, p.CurrentTask)
	}

	accepted := payloadOf[handedInPayload](t, call(t, session, "submit_task", raceOn(request)))
	if accepted.Screen != "task" || accepted.Attempt != 2 {
		t.Errorf("the mended task = %+v, want it on the card at the second attempt", accepted)
	}

	h.settle()
	submitted := linesOf(h, "task_submitted")
	if len(submitted) != 2 {
		t.Fatalf("task_submitted lines = %d, want 2", len(submitted))
	}
	first := submitted[0].ContextMap()
	if first["outcome"] != "rejected" || first["primary"] != "bad_structure" ||
		fmt.Sprint(first["failed"]) != "[bad_structure distractor_explanations]" {
		t.Errorf("the refusal's line = %v, want it counted by its first reason and naming both", first)
	}
}

// A client may send a part of the task as a string holding its JSON rather
// than as the JSON itself: the part is read for the JSON it holds, and the task
// reaches the card as it would have. A string that holds no JSON is still no
// part, and is refused in the structure check's own words.
func TestAPartSentAsAStringOfItsJSONIsReadForIt(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)

	worded := raceOn(request)
	worded["brief"] = "the brief as it came"
	refused := payloadOf[handedInPayload](t, call(t, session, "submit_task", worded))
	if refused.Code != "bad_structure" || len(refused.Reasons) == 0 ||
		!slices.Contains(refused.Reasons[0].Messages, "brief must be a JSON object") {
		t.Errorf("a brief of words = %+v, want it refused as no object", refused)
	}

	// JSON with a fault in it is told as broken JSON: the model can mend that,
	// and cannot send the part as anything but a string.
	faulty := raceOn(request)
	faulty["brief"] = `{"setting":"park",}`
	refused = payloadOf[handedInPayload](t, call(t, session, "submit_task", faulty))
	if refused.Code != "bad_structure" || len(refused.Reasons) == 0 ||
		!slices.Contains(refused.Reasons[0].Messages, "brief is not valid JSON") {
		t.Errorf("a brief of broken JSON = %+v, want it refused as broken JSON", refused)
	}

	// A whole number written as 2.0 reads as 2 in a part sent as JSON; so it
	// does in a part sent as a string of it.
	race := partsAsStrings(t, raceOn(request))
	brief, isText := race["brief"].(string)
	if !isText || !strings.Contains(brief, `"difficulty":2`) {
		t.Fatalf("the brief as a string = %v, want its difficulty in it", race["brief"])
	}
	race["brief"] = strings.Replace(brief, `"difficulty":2`, `"difficulty":2.0`, 1)
	accepted := payloadOf[handedInPayload](t, call(t, session, "submit_task", race))
	if accepted.Screen != "task" || accepted.Task == nil || accepted.Task.Question != raceQuestion {
		t.Errorf("the race with its parts as strings = %+v, want it on the card", accepted)
	}
}

// partsAsStrings is a task handed in with its brief, task and self-check each
// sent as a string of its JSON, as a client may send a part whose schema names
// no type.
func partsAsStrings(t testing.TB, race map[string]any) map[string]any {
	t.Helper()

	for _, part := range []string{"brief", "task", "self_check"} {
		encoded, err := json.Marshal(race[part])
		if err != nil {
			t.Fatalf("encode the %s: %v", part, err)
		}
		race[part] = string(encoded)
	}
	return race
}

// The third refusal closes the request: nothing reaches the child, the day's
// failed generations go up and its accepted tasks do not. A task handed in
// after that is for a request that is over, and spends nothing.
func TestTheThirdRefusalClosesTheRequest(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)

	var last handedInPayload
	for range profile.MaxAttempts {
		last = payloadOf[handedInPayload](t, call(t, session, "submit_task", broken(request)))
	}
	if last.Code != "attempts_exhausted" || last.AttemptsLeft == nil || *last.AttemptsLeft != 0 || last.Attempt != 3 {
		t.Errorf("the third refusal = %+v, want attempts_exhausted with none left", last)
	}
	p, revision := loadKept(t, kept)
	if p.OpenRequest != nil || p.Daily.Failed != 1 || p.Daily.Accepted != 0 {
		t.Errorf("the profile holds %+v and the counters %+v, want the request closed and one failure counted",
			p.OpenRequest, p.Daily)
	}

	fourth := payloadOf[handedInPayload](t, call(t, session, "submit_task", raceOn(request)))
	if fourth.Status != "stale" || fourth.Code != "stale_request" {
		t.Errorf("a fourth attempt = %+v, want it refused as stale", fourth)
	}
	if _, now := loadKept(t, kept); now != revision {
		t.Error("a fourth attempt wrote the profile, want nothing written")
	}
}

// The task handed in has to be the one asked for: the topic, the level and the
// difficulty the request recorded, and every skill the child has not met kept
// out. A program that crashes is a solver that did not run.
func TestATaskIsHeldToWhatWasAskedFor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		change func(race map[string]any, asked profile.Brief)
		code   string
	}{
		{"another difficulty", func(race map[string]any, asked profile.Brief) {
			asked.Difficulty = 3
			race["brief"] = asked
		}, "bad_structure"},
		{"a skill the child has not met let back in", func(race map[string]any, asked profile.Brief) {
			asked.ExcludedSkills = []string{}
			race["brief"] = asked
		}, "bad_structure"},
		{"a program that crashes", func(race map[string]any, _ profile.Brief) {
			race["solver"] = "def solve(options):\n    return 1 // 0\n"
		}, "solver_error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t, "fractions")
			_, session := lesson(t, kept)
			request := askForTheRace(t, session, kept)
			race := raceOn(request)
			tc.change(race, request.Brief)

			if got := payloadOf[handedInPayload](t, call(t, session, "submit_task", race)); got.Code != tc.code {
				t.Errorf("submit_task = %+v, want it refused with %s", got, tc.code)
			}
		})
	}
}

// A task handed in for no request that is open spends nothing and writes
// nothing: not for a request never asked, nor another request's, nor one
// already handed out, nor one that waited past its window. A task handed in
// twice — the answer to the first gone astray — leaves the child's card on
// the task it holds rather than on a wait.
func TestATaskForNoOpenRequestSpendsNothing(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		before func(t *testing.T, session *mcp.ClientSession, kept store.Storage, moving *clock) map[string]any
		screen string
	}{
		{"no request asked for", func(t *testing.T, _ *mcp.ClientSession, _ store.Storage, _ *clock) map[string]any {
			return raceOn(&profile.OpenRequest{ID: "req_nobody_asked", Brief: profile.Brief{}})
		}, "waiting"},
		{"another request's id", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) map[string]any {
			race := raceOn(askForTheRace(t, session, kept))
			race["request_id"] = "req_another"
			return race
		}, "waiting"},
		{"a request handed out already", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) map[string]any {
			race := raceOn(askForTheRace(t, session, kept))
			call(t, session, "submit_task", race)
			return race
		}, "task"},
		{"a request whose task was answered", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) map[string]any {
			race := raceOn(askForTheRace(t, session, kept))
			card := wantOnTheCard(t, call(t, session, "submit_task", race))
			call(t, session, "submit_answer", map[string]any{"task_id": card.Task.ID, "answer": "C"})
			return race
		}, "waiting"},
		{"a request past its window", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, moving *clock) map[string]any {
			race := raceOn(askForTheRace(t, session, kept))
			moving.advance(config.DefaultRequestWindow)
			return race
		}, "waiting"},
		{"a request with every attempt spent", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) map[string]any {
			race := raceOn(askForTheRace(t, session, kept))
			spendEveryAttempt(t, kept)
			return race
		}, "waiting"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept, moving := racer(t), &clock{at: lessonDay}
			h, session := lessonWith(t, kept, moving, nil)
			race := tc.before(t, session, kept, moving)
			wantStaleSpendingNothing(t, h, session, kept, race, tc.screen)
		})
	}
}

// wantStaleSpendingNothing hands in a task for no open request and holds what
// follows to a stale refusal on this screen that judged nothing and wrote
// nothing. The card shows the task the child holds when the screen is the
// task's.
func wantStaleSpendingNothing(t *testing.T, h *harness, session *mcp.ClientSession, kept store.Storage,
	race map[string]any, screen string,
) {
	t.Helper()

	before, revision := loadKept(t, kept)
	judged := len(linesOf(h, "task_submitted"))

	got := payloadOf[handedInPayload](t, call(t, session, "submit_task", race))
	if got.Status != "stale" || got.Code != "stale_request" || got.Screen != screen || got.Attempt != 0 {
		t.Errorf("submit_task = %+v, want it refused as stale on the %s screen", got, screen)
	}
	if screen == "task" && (got.Task == nil || got.Task.ID != before.CurrentTask.ID) {
		t.Errorf("the card shows %+v, want the task the child holds", got.Task)
	}
	if _, now := loadKept(t, kept); now != revision {
		t.Error("the profile was written, want nothing written")
	}
	h.settle()
	if after := len(linesOf(h, "task_submitted")); after != judged {
		t.Errorf("task_submitted lines went from %d to %d, want no line for a task nothing judged", judged, after)
	}
}

// failingSandbox is a sandbox whose every run fails the way the case gives,
// which the real one cannot be made to do on demand.
type failingSandbox struct{ err error }

func (s failingSandbox) Run(context.Context, string, solver.Options) (solver.Result, error) {
	return solver.Result{}, s.err
}

// A sandbox that cannot run a task is the service's failure, not the task's:
// the model is told in the service's own words — in the words of its own when
// every slot stayed taken, which say to hand the same task in again — and no
// attempt is spent.
func TestASandboxThatCannotRunSpendsNoAttempt(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, sentence, kind string
		err                  error
	}{
		{
			name: "a sandbox that is gone", err: errors.New("the sandbox is gone"),
			sentence: "Something went wrong inside MathTrail.", kind: "internal",
		},
		{
			name: "a sandbox whose every slot stayed taken", err: fmt.Errorf("starlark: wait for a free slot: %w", solver.ErrBusy),
			sentence: "MathTrail is checking too many tasks right now", kind: "busy",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			h, session := lessonWith(t, kept, &clock{at: lessonDay}, failingSandbox{err: test.err})
			request := askForTheRace(t, session, kept)
			_, revision := loadKept(t, kept)

			result := call(t, session, "submit_task", raceOn(request))
			wantOurSentence(t, result, test.sentence)
			if p, now := loadKept(t, kept); now != revision || p.OpenRequest.Attempts != 0 {
				t.Errorf("the profile is at revision %s with %d attempts spent, want it untouched", now, p.OpenRequest.Attempts)
			}

			h.settle()
			wantFailed(t, h, "submit_task", test.kind)
		})
	}
}

// Asked again while a task is being written, next_task hands back the same
// request, marked open and aged, with the package again, and writes nothing —
// and the choice of the second call is not applied. Once the request has
// waited past its window, the next ask opens a new one.
func TestAskingAgainGivesTheSameRequest(t *testing.T) {
	t.Parallel()

	kept, moving := racer(t), &clock{at: lessonDay}
	h, session := lessonWith(t, kept, moving, nil)
	first := askForTheRace(t, session, kept)
	_, revision := loadKept(t, kept)

	moving.advance(30 * time.Second)
	again := call(t, session, "next_task", map[string]any{"language": "ru", "topic": "time.clocks", "reason": "clocks"})
	text := wantWordsAlone(t, again)
	if lead := leadOf(text); !strings.Contains(lead, "Request "+first.ID+" is already open, since 30 seconds ago") ||
		!strings.Contains(lead, "If you have written its task, hand it in") ||
		!strings.Contains(lead, "If you have not, a turn cut short say, the task is yours to write") ||
		!strings.Contains(lead, "not applied") {
		t.Errorf("the words are %q, want the request open for 30 seconds, its task asked for if written and the model's "+
			"to write if not, and the choice not applied", lead)
	}
	packageIn(t, text)
	if p, now := loadKept(t, kept); now != revision || p.OpenRequest.Brief.TargetConcept != "logic.ordering" {
		t.Error("asking again wrote the profile or changed the request, want neither")
	}
	for language, ignored := range map[string]bool{"de": true, "EN": false} {
		if said := textOf(t, call(t, session, "next_task", map[string]any{"language": language})); strings.Contains(said, "not applied") != ignored {
			t.Errorf("asked again in %s, the words are %q, want the arguments said to be ignored: %v", language, said, ignored)
		}
	}

	moving.advance(config.DefaultRequestWindow)
	if later := leadOf(wantWordsAlone(t, call(t, session, "next_task", raceChoice))); strings.Contains(later, first.ID) ||
		strings.Contains(later, "already open") {
		t.Errorf("next_task past the window says %q, want a new request", later)
	}

	h.settle()
	var open []bool
	for _, line := range linesOf(h, "task_requested") {
		open = append(open, line.ContextMap()["already_open"] == true)
	}
	if want := []bool{false, true, true, true, false}; !slices.Equal(open, want) {
		t.Errorf("task_requested lines say already open %v, want %v", open, want)
	}
}

// Arguments no request can be opened from are refused one by one in the
// service's words, and nothing is written: the language a task is to be
// written in, always, and a choice of the model's own that the catalog can
// carry, with its reason.
func TestArgumentsNoRequestCanBeOpenedFromAreRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		arguments map[string]any
		fields    []string
	}{
		{"no language", map[string]any{"language": ""}, []string{"language"}},
		{"not a language", map[string]any{"language": "not a language"}, []string{"language"}},
		{"a topic with no reason", map[string]any{"language": "en", "topic": "time.clocks"}, []string{"reason"}},
		{"a topic of spaces", map[string]any{"language": "en", "topic": "  "}, []string{"topic", "reason"}},
		{"a topic nobody has", map[string]any{"language": "en", "topic": "astronomy.stars", "reason": "stars"}, []string{"topic"}},
		{
			"a level the topic is not taught at",
			map[string]any{"language": "en", "topic": "percent.basic", "grade_level": "1-2", "reason": "easier"},
			[]string{"grade_level"},
		},
		{
			"everything at once",
			map[string]any{"language": "", "topic": "astronomy.stars", "difficulty": 9},
			[]string{"language", "topic", "difficulty", "reason"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			_, session := lesson(t, kept)
			_, revision := loadKept(t, kept)

			got := payloadOf[refusedRequestPayload](t, call(t, session, "next_task", tc.arguments))
			fields := make([]string, 0, len(got.Problems))
			for _, problem := range got.Problems {
				fields = append(fields, problem.Field)
			}
			if got.Status != "rejected" || got.Code != "invalid_arguments" || !slices.Equal(fields, tc.fields) {
				t.Errorf("next_task = %+v, want %v refused as invalid_arguments", got, tc.fields)
			}
			if p, now := loadKept(t, kept); now != revision || p.OpenRequest != nil {
				t.Error("a refusal wrote the profile, want nothing written")
			}
		})
	}
}

// A task on the child's card with no answer, when another is asked for, is
// recorded as skipped: in the window, in its topic and in a line of its own.
// Nothing that learns from answers moves — not the level, not the runs, not
// the trial series — and the progress shows the skip to the parent.
func TestAnUnansweredTaskIsSkippedByTheNextAsk(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	before, _ := loadKept(t, kept)
	left := before.CurrentTask
	h, session := lesson(t, kept)

	asked := call(t, session, "next_task", map[string]any{"language": "en"})
	if text := textOf(t, asked); !strings.Contains(text, left.ID+", left on the child's card without an answer, is recorded as skipped") {
		t.Errorf("the words are %q, want them to say the task left was recorded as skipped", text)
	}

	p, _ := loadKept(t, kept)
	last := p.Recent[len(p.Recent)-1]
	if !last.Skipped || last.TaskID != left.ID || p.CurrentTask != nil || p.Topics[left.Topic].Skipped != 1 {
		t.Errorf("the window ends with %+v, the task in flight is %+v, want the skip recorded", last, p.CurrentTask)
	}
	if p.Ratings != before.Ratings {
		t.Errorf("the ratings moved to %+v on a skip, want %+v", p.Ratings, before.Ratings)
	}

	progress := payloadOf[skipsPayload](t, call(t, session, "get_progress", nil))
	if len(progress.Recent) == 0 || !progress.Recent[0].Skipped || progress.Recent[0].Correct != nil {
		t.Errorf("the latest entry of the progress is %+v, want the skip, with no outcome", progress.Recent)
	}
	at := slices.IndexFunc(progress.Topics, func(topic skippedTopic) bool { return topic.Topic == left.Topic })
	if at < 0 || progress.Topics[at].Skipped != 1 {
		t.Errorf("the topics of the progress are %+v, want %s with one skip", progress.Topics, left.Topic)
	}

	h.settle()
	if skipped := linesOf(h, "task_skipped"); len(skipped) != 1 || skipped[0].ContextMap()["topic"] != left.Topic {
		t.Errorf("task_skipped lines = %d, want one about %s", len(skipped), left.Topic)
	}
}

// Nothing that gives the answer away leaves the seal before the child has
// answered: not in the payload a card draws, not in the words for the model,
// not in the open part of the profile, not on a span and not in a line. The
// card is drawn from two results — the task accepted, and the same task handed
// in again, which shows the child the task they hold — and neither carries it.
func TestTheAnswerStaysSealed(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	race := raceOn(askForTheRace(t, session, kept))
	handed := call(t, session, "submit_task", race)
	again := call(t, session, "submit_task", race)

	secrets := slices.Concat([]string{raceSolution, "correct_answer", "solution", "distractors"}, raceExplained, raceTraps)
	for _, card := range []struct {
		name   string
		result *mcp.CallToolResult
	}{{"accepted", handed}, {"handed in again", again}} {
		wantNoneOf(t, "the payload of the task "+card.name, []string{string(rawPayload(t, card.result))}, secrets)
		wantNoneOf(t, "the words of the task "+card.name, []string{textOf(t, card.result)}, secrets)

		var payload struct {
			Screen string                     `json:"screen"`
			Task   map[string]json.RawMessage `json:"task"`
		}
		if err := json.Unmarshal(rawPayload(t, card.result), &payload); err != nil {
			t.Fatalf("the payload of the task %s does not read: %v", card.name, err)
		}
		if payload.Screen != "task" {
			t.Errorf("the task %s draws the %q screen, want the task's card", card.name, payload.Screen)
		}
		if keys := slices.Sorted(maps.Keys(payload.Task)); !slices.Equal(keys,
			[]string{"drawing", "hint", "id", "language", "options", "question", "topic"}) {
			t.Errorf("the card of the task %s has %v, want what a child may see and nothing else", card.name, keys)
		}
	}

	p, _ := loadKept(t, kept)
	p.CurrentTask.Sealed = "sealed"
	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	wantNoneOf(t, "the open part of the profile", []string{string(written)}, secrets)

	h.settle()
	secrets = append(secrets, raceQuestion, "Ben")
	for _, span := range h.spans.Ended() {
		wantNoneOf(t, "span "+span.Name(), spanTexts(span), secrets)
	}
	lines := h.logs.All()
	for i := range lines {
		wantNoneOf(t, "line "+lines[i].Message, lineTexts(t, &lines[i]), secrets)
	}
}

// A card is told the language the parent chose for the cards, the waiting card
// a refusal draws as well as the task's, so that it speaks that language in
// place of the chat's.
func TestTheCardIsToldTheLanguageChosenForTheCards(t *testing.T) {
	t.Parallel()

	russian := "ru"
	kept := keptAsIs(t, profile.New(profile.Student{
		Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}, UILanguage: &russian,
	}, "test", lessonDay))
	_, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)

	refused := payloadOf[handedInPayload](t, call(t, session, "submit_task", broken(request)))
	handed := payloadOf[handedInPayload](t, call(t, session, "submit_task", raceOn(request)))
	for _, card := range []handedInPayload{refused, handed} {
		if card.Child == nil || card.Child.UILanguage == nil || *card.Child.UILanguage != russian {
			t.Errorf("the %s card is for %+v, want the cards' language, %s", card.Screen, card.Child, russian)
		}
	}
	if refused.Screen != "waiting" || handed.Screen != "task" {
		t.Errorf("the cards are %q and %q, want a waiting card and then the task's", refused.Screen, handed.Screen)
	}
}

// The name the child goes by is on the card, beside the task, and nowhere
// else: not in the package a task is written from, not in the words of the
// task, not in the task itself, not on a span and not in a line.
func TestThePseudonymStaysBesideTheTask(t *testing.T) {
	t.Parallel()

	const name = "Zebra-Quill"
	kept := keptAsIs(t, profile.New(profile.Student{Grade: 2, Pseudonym: name}, "test", lessonDay))
	h, session := lesson(t, kept)

	asked := call(t, session, "next_task", raceChoice)
	wantNoneOf(t, "the package", []string{textOf(t, asked)}, []string{name})
	p, _ := loadKept(t, kept)
	handed := call(t, session, "submit_task", raceOn(p.OpenRequest))
	card := payloadOf[handedInPayload](t, handed)
	if card.Child == nil || card.Child.Pseudonym != name {
		t.Errorf("child = %+v, want the name the card shows at its top", card.Child)
	}
	taskJSON, err := json.Marshal(card.Task)
	if err != nil {
		t.Fatalf("the task does not encode: %v", err)
	}
	wantNoneOf(t, "the task and its words", []string{string(taskJSON), textOf(t, handed)}, []string{name})

	h.settle()
	for _, span := range h.spans.Ended() {
		wantNoneOf(t, "span "+span.Name(), spanTexts(span), []string{name})
	}
	lines := h.logs.All()
	for i := range lines {
		wantNoneOf(t, "line "+lines[i].Message, lineTexts(t, &lines[i]), []string{name})
	}
}

// A task's review is seen in the trace under the call that asked for it: the
// examination, with both runs of the solver under it, and the judgement, which
// says how the review ended in words from a closed list.
func TestTheReviewIsTracedUnderTheCall(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	call(t, session, "submit_task", raceOn(askForTheRace(t, session, kept)))
	h.settle()

	submitted := h.spanNamed(t, "tools/call submit_task")
	examined := h.spanNamed(t, "examine_task")
	judged := h.spanNamed(t, "judge_task")
	if examined.Parent().SpanID() != submitted.SpanContext().SpanID() ||
		judged.Parent().SpanID() != submitted.SpanContext().SpanID() {
		t.Error("the examination and the judgement are not under the call")
	}
	runs := 0
	for _, span := range h.spans.Ended() {
		if span.Name() == "solver_run" && span.Parent().SpanID() == examined.SpanContext().SpanID() {
			runs++
		}
	}
	if runs != 2 {
		t.Errorf("solver runs under the examination = %d, want both", runs)
	}
	for _, attr := range judged.Attributes() {
		if attr.Key == "mathtrail.review.outcome" && attr.Value.AsString() != "accepted" {
			t.Errorf("the judgement ended %q, want accepted", attr.Value.AsString())
		}
	}
}

// With no profile there is no task to ask for or to hand in: both say so, show
// the first sign-in and write nothing.
func TestWithNoProfileThereIsNoTask(t *testing.T) {
	t.Parallel()

	kept := memory.New()
	_, session := lesson(t, kept)

	asked := wantWordsAlone(t, call(t, session, "next_task", raceChoice))
	handed := payloadOf[handedInPayload](t, call(t, session, "submit_task",
		raceOn(&profile.OpenRequest{ID: "req_nobody"})))
	if !strings.HasPrefix(asked, "No task can be asked for yet. There is no profile yet.") || handed.Screen != "first_run" ||
		handed.Code != "stale_request" {
		t.Errorf("next_task says %q and submit_task = %+v, want both to point at the first sign-in", asked, handed)
	}
	if _, _, err := kept.Load(t.Context(), devAccount); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Load() error = %v, want still no profile", err)
	}
}

// The model can only choose a topic of its own with ids it has been told, and
// only within the limit it has been told: next_task lists every topic with the
// levels it is taught at, and says how long a reason may be.
func TestNextTaskStatesWhatItTakes(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, memory.New())
	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	at := slices.IndexFunc(listed.Tools, func(tool *mcp.Tool) bool { return tool.Name == "next_task" })
	if at < 0 {
		t.Fatal("next_task is not listed")
	}
	tool := listed.Tools[at]
	loaded, _ := shipped()
	for _, topic := range loaded.Topics() {
		levels := make([]string, 0, len(topic.GradeLevels))
		for _, level := range topic.GradeLevels {
			levels = append(levels, string(level))
		}
		if line := "- " + topic.ID + " (" + strings.Join(levels, ", ") + ")"; !strings.Contains(tool.Description, line) {
			t.Errorf("the description of next_task does not name the topic %s with its levels, %q", topic.ID, line)
		}
	}
	schema, err := json.Marshal(tool.InputSchema)
	if err != nil {
		t.Fatalf("the input schema does not encode: %v", err)
	}
	if want := fmt.Sprintf("at most %d characters", tutor.MaxReason); !strings.Contains(string(schema), want) {
		t.Errorf("the input schema does not say %q about the reason", want)
	}
}

// skipsPayload is what the progress shows of the skipped tasks.
type skipsPayload struct {
	Recent []struct {
		Correct *bool `json:"correct"`
		Skipped bool  `json:"skipped"`
	} `json:"recent"`
	Topics []skippedTopic `json:"topics"`
}

// skippedTopic is a topic of the progress, as far as its skips go.
type skippedTopic struct {
	Topic   string `json:"topic"`
	Skipped int    `json:"skipped"`
}

// keyGone is a seal whose key is gone: nothing can be sealed with it, which a
// real key ring cannot be made to do on demand.
type keyGone struct{}

func (keyGone) Seal([]byte, ...string) (string, error) { return "", errors.New("the key is gone") }
func (keyGone) Open(string, ...string) ([]byte, error) { return nil, errors.New("the key is gone") }

// A task that cannot be sealed is not handed out: a task on the card with its
// answer in the open is worse than none. The failure is ours, told in our
// words, and nothing is written — not even the attempt.
func TestATaskThatCannotBeSealedIsNotHandedOut(t *testing.T) {
	t.Parallel()

	asked := fuzzProfile(t)
	kept := keptAsIs(t, asked)
	h := newHarness(t)
	parts := allParts(t)
	parts.Store, parts.Sealer, parts.Logger, parts.Traces = kept, keyGone{}, h.log, h.traces
	service, err := mcpserver.NewService(parts)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	h.start(t, mcpserver.DevSignIn, service.TaskTools()...)
	session := h.connect(t, "")
	_, revision := loadKept(t, kept)

	wantOurSentence(t, call(t, session, "submit_task", raceOn(asked.OpenRequest)), "Something went wrong inside MathTrail.")
	if p, now := loadKept(t, kept); now != revision || p.CurrentTask != nil || p.OpenRequest.Attempts != 0 {
		t.Errorf("the profile is at revision %s with task %+v, want it untouched", now, p.CurrentTask)
	}
}

// Where no card is drawn the drawing is read out as it is to be shown: in a
// block of its own, fenced by more backticks than any run of them inside it.
func TestTheDrawingIsReadOutInABlockOfItsOwn(t *testing.T) {
	t.Parallel()

	const drawing = "+---+\n|```|\n+---+\n"
	on := fuzzProfile(t)
	on.CurrentTask = &profile.CurrentTask{
		Difficulty: 2, Drawing: drawing, Fingerprint: "sketch", GradeLevel: "1-2", Hint: "Look at the box.",
		ID: "tsk_boxed", InstructionsVersion: "test", IssuedAt: profile.At(lessonDay), Language: "en",
		Options: map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"},
		Sealed:  "mt1.t.kid.sealed", Topic: "logic.ordering", Wording: "How many boxes are drawn?",
	}
	_, session := lesson(t, keptAsIs(t, on))

	race := raceOn(on.OpenRequest)
	race["request_id"] = "req_another"
	text := textOf(t, call(t, session, "submit_task", race))
	if want := "Drawing:\n````\n+---+\n|```|\n+---+\n````\n"; !strings.Contains(text, want) {
		t.Errorf("the words are %q, want the drawing fenced as %q", text, want)
	}
	if want := `Options: A) "1" B) "2" C) "3" D) "4" E) "5"`; !strings.Contains(text, want) {
		t.Errorf("the words are %q, want the options read out in order as %q", text, want)
	}
}

// spendEveryAttempt writes the open request with every attempt spent, as only a
// hand editing the file can leave it: the last refusal closes a request.
func spendEveryAttempt(t *testing.T, kept store.Storage) {
	t.Helper()

	p, revision := loadKept(t, kept)
	p.OpenRequest.Attempts = profile.MaxAttempts
	p.Touch("test", lessonDay)
	if _, err := kept.Save(t.Context(), devAccount, p, revision); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
}

// A request with every attempt spent is not waited for: asking again opens a
// new one rather than handing back a request no task can be handed in for.
func TestARequestWithNoAttemptLeftIsNotAskedAgain(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	spent := askForTheRace(t, session, kept)
	spendEveryAttempt(t, kept)

	if again := leadOf(wantWordsAlone(t, call(t, session, "next_task", raceChoice))); strings.Contains(again, "already open") ||
		strings.Contains(again, spent.ID) {
		t.Errorf("next_task says %q, want a new request in place of %s", again, spent.ID)
	}
}

// A task handed out over one still in flight — a file edited by hand holding
// both — records the one it replaces as skipped, with its line, rather than
// letting it vanish from what the parent sees.
func TestATaskHandedOutOverAnotherSkipsIt(t *testing.T) {
	t.Parallel()

	on := fuzzProfile(t)
	on.CurrentTask = &profile.CurrentTask{
		Difficulty: 3, Fingerprint: "sketch", GradeLevel: "3-4", Hint: "Count them.",
		ID: "tsk_left", InstructionsVersion: "test", IssuedAt: profile.At(lessonDay), Language: "en",
		Options: map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"},
		Sealed:  "mt1.t.kid.sealed", Topic: "time.clocks", Wording: "What time is it?",
	}
	kept := keptAsIs(t, on)
	h, session := lesson(t, kept)

	if card := payloadOf[handedInPayload](t, call(t, session, "submit_task", raceOn(on.OpenRequest))); card.Screen != "task" {
		t.Fatalf("submit_task = %+v, want the race handed out", card)
	}
	p, _ := loadKept(t, kept)
	if last := p.Recent[len(p.Recent)-1]; !last.Skipped || last.TaskID != "tsk_left" {
		t.Errorf("the window ends with %+v, want the task it replaced skipped", last)
	}
	h.settle()
	if skipped := linesOf(h, "task_skipped"); len(skipped) != 1 || skipped[0].ContextMap()["topic"] != "time.clocks" {
		t.Errorf("task_skipped lines = %d, want one about time.clocks", len(skipped))
	}
}

// A task handed out over one that has had its answer — a file edited by hand
// holding an open request beside it — replaces it with nothing recorded: its
// answer is in the window already, and it was never left.
func TestATaskHandedOutOverAnAnsweredOneSkipsNothing(t *testing.T) {
	t.Parallel()

	on := raceOnTheCard(t, 0)
	if _, err := on.Record(profile.Answered{TaskID: on.CurrentTask.ID, Choice: "C", At: lessonDay}, sealer(t)); err != nil {
		t.Fatalf("Record() error = %v, want the race answered", err)
	}
	asked := fuzzProfile(t).OpenRequest.Brief
	request := on.Ask(&asked, profile.TutorLLM, "en", lessonDay)
	kept := keptAsIs(t, on)
	h, session := lesson(t, kept)
	window := len(on.Recent)

	task := raceTask(raceDistractors())
	task["question"] = "Ann, Ben and Kim ran a race. Kim finished after Ben. Ann finished last. Who won?"
	race := raceOn(request)
	race["task"] = task
	if card := payloadOf[handedInPayload](t, call(t, session, "submit_task", race)); card.Screen != "task" {
		t.Fatalf("submit_task = %+v, want the new race handed out", card)
	}
	if p, _ := loadKept(t, kept); len(p.Recent) != window {
		t.Errorf("the window holds %d entries, want %d: nothing recorded of the answered task", len(p.Recent), window)
	}
	h.settle()
	if skipped := linesOf(h, "task_skipped"); len(skipped) != 0 {
		t.Errorf("task_skipped lines = %d, want none", len(skipped))
	}
}

// The words of the progress name the latest answers and count the skips apart,
// so that a run of skipped tasks never hides how the answers went.
func TestTheProgressCountsTheSkipsApart(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	p, revision := loadKept(t, kept)
	for i := range 6 {
		p.Recent = append(p.Recent, profile.Answer{
			AnsweredAt: profile.At(lessonDay), Difficulty: 2, GradeLevel: "3-4", Skipped: true,
			TaskID: fmt.Sprintf("tsk_skipped_%d", i), Topic: "time.clocks",
		})
	}
	p.Touch("test", lessonDay)
	if _, err := kept.Save(t.Context(), devAccount, p, revision); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	_, session := lesson(t, kept)

	text := textOf(t, call(t, session, "get_progress", nil))
	if !strings.Contains(text, "Latest answers, the latest first: Clocks wrong,") ||
		!strings.Contains(text, "Of the last 9 tasks, 6 were left without an answer.") {
		t.Errorf("the words are %q, want the answers named and the six skips counted apart", text)
	}
}

// unjudging is a review whose second half cannot run, which the real reviewer
// does only when handed nothing to judge.
type unjudging struct{ checks.Reviewer }

func (unjudging) Judge(checks.Examined, checks.Against) (checks.Outcome, error) {
	return checks.Outcome{}, errors.New("nothing to judge")
}

// A review that cannot be judged is the service's failure: the model is told in
// our words, and nothing is spent or written.
func TestAReviewThatCannotBeJudgedSpendsNothing(t *testing.T) {
	t.Parallel()

	asked := fuzzProfile(t)
	kept := keptAsIs(t, asked)
	h := newHarness(t)
	parts := allParts(t)
	parts.Store, parts.Reviewer, parts.Logger, parts.Traces = kept, unjudging{parts.Reviewer}, h.log, h.traces
	service, err := mcpserver.NewService(parts)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	h.start(t, mcpserver.DevSignIn, service.TaskTools()...)
	session := h.connect(t, "")
	_, revision := loadKept(t, kept)

	wantOurSentence(t, call(t, session, "submit_task", raceOn(asked.OpenRequest)), "Something went wrong inside MathTrail.")
	if p, now := loadKept(t, kept); now != revision || p.OpenRequest.Attempts != 0 {
		t.Errorf("the profile is at revision %s with %d attempts, want it untouched", now, p.OpenRequest.Attempts)
	}
	h.settle()
	if judged := h.spanNamed(t, "judge_task"); judged.Status().Description != "the task could not be judged" {
		t.Errorf("judge_task status = %q, want it marked failed", judged.Status().Description)
	}
}

// A request no package can be built for — its topic, or a skill the child has
// not met, gone from the catalog in a file edited by hand — is the service's
// failure, and nothing is written, whether the request is new or open already.
func TestARequestNoPackageCanBeBuiltForWritesNothing(t *testing.T) {
	t.Parallel()

	gone := fuzzProfile(t)
	gone.OpenRequest.Brief.TargetConcept = "gone.topic"
	unknown := profile.New(profile.Student{
		Grade: 2, Pseudonym: "Otter", ExcludedSkills: []string{"gone_skill"},
	}, "test", lessonDay)

	for name, p := range map[string]*profile.Profile{"an open request": gone, "a new request": unknown} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			kept := keptAsIs(t, p)
			_, session := lesson(t, kept)
			_, revision := loadKept(t, kept)

			wantOurSentence(t, call(t, session, "next_task", map[string]any{"language": "en"}),
				"Something went wrong inside MathTrail.")
			if _, now := loadKept(t, kept); now != revision {
				t.Error("the profile was written, want nothing written")
			}
		})
	}
}
