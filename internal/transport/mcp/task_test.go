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
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
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
// profile keeps it: the one the card next_task drew waits for.
func askForTheRace(t *testing.T, session *mcp.ClientSession, kept store.Storage) *profile.OpenRequest {
	t.Helper()

	asked := call(t, session, "next_task", raceChoice)
	coming := wantComing(t, asked)
	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil || coming.RequestID != p.OpenRequest.ID ||
		!strings.Contains(textOf(t, asked), "Request "+p.OpenRequest.ID+" is open") {
		t.Fatalf("next_task = %+v, %q and the profile holds %+v, want the one request", coming, textOf(t, asked),
			p.OpenRequest)
	}
	return p.OpenRequest
}

// requestPayload is what next_task hands the card it draws: a task on its way,
// or why no request was opened.
type requestPayload struct {
	Screen   string `json:"screen"`
	Status   string `json:"status"`
	Code     string `json:"code"`
	Problems []struct {
		Field string `json:"field"`
		Code  string `json:"code"`
	} `json:"problems"`
	RequestID   string `json:"request_id"`
	AlreadyOpen bool   `json:"already_open"`
	AgeSeconds  int    `json:"age_seconds"`
	Child       *struct {
		Pseudonym string `json:"pseudonym"`
		Grade     int    `json:"grade"`
	} `json:"child"`
	Language string `json:"language"`
}

// wantComing holds a result of next_task to a card waiting for a task on its
// way — the request, whom it is for and the language it is written in — and
// is what the card was handed.
func wantComing(t *testing.T, result *mcp.CallToolResult) requestPayload {
	t.Helper()

	coming := payloadOf[requestPayload](t, result)
	if coming.Screen != "coming" || coming.Status != "" || coming.RequestID == "" || coming.Child == nil ||
		coming.Language == "" {
		t.Fatalf("next_task = %+v, want a card waiting for the task of a request", coming)
	}
	return coming
}

// fetchPackage asks for the package of a request, as the model does once
// next_task has answered, and is the words it comes in.
func fetchPackage(t *testing.T, session *mcp.ClientSession, requestID string) string {
	t.Helper()

	return wantWordsAlone(t, call(t, session, "get_package", map[string]any{"request_id": requestID}))
}

// wantWordsAlone holds a result of get_package to carrying no payload, and is
// its words. The package is for the model alone, and a host that shows the
// model a payload in place of the words would show it nothing to write from.
func wantWordsAlone(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()

	if result.IsError || result.StructuredContent != nil {
		t.Fatalf("get_package = %+v, %s, want words alone", result.StructuredContent, textOf(t, result))
	}
	return textOf(t, result)
}

// leadOf is the words of get_package before the package they carry, or the
// whole of words that carry none.
func leadOf(text string) string {
	lead, _, _ := strings.Cut(text, "\n\nPackage:\n")
	return lead
}

// awaitedPayload is what read_task tells the card that waits for a task.
type awaitedPayload struct {
	Screen  string `json:"screen"`
	Status  string `json:"status"`
	Code    string `json:"code"`
	Refused int    `json:"refused"`
	Child   *struct {
		Pseudonym string `json:"pseudonym"`
	} `json:"child"`
	Task *struct {
		ID       string            `json:"id"`
		Question string            `json:"question"`
		Options  map[string]string `json:"options"`
	} `json:"task"`
	Language string `json:"language"`
}

// awaited is what read_task tells the card waiting for a request's task.
func awaited(t *testing.T, session *mcp.ClientSession, requestID string) awaitedPayload {
	t.Helper()

	return payloadOf[awaitedPayload](t, call(t, session, "read_task", map[string]any{"request_id": requestID}))
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
	Language string `json:"language"`
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

// The whole way, as a model and a card go it: a task asked for with a choice
// of the model's own, which draws the card it will come to; the package it is
// written from, fetched by the model; the card told the task is being written;
// the task handed in and accepted at the first attempt — sealed in the profile
// with its answer — and the card told it is on the card, without its answer;
// the child's answer from the card; and every step of it a line.
func TestATaskGoesFromTheRuleToTheChildsCard(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)

	open := wantRequestOpened(t, session, call(t, session, "next_task", raceChoice), kept)
	if waiting := awaited(t, session, open.ID); waiting.Screen != "coming" || waiting.Task != nil || waiting.Status != "" {
		t.Errorf("read_task before the task is handed in = %+v, want it being written", waiting)
	}
	handed := call(t, session, "submit_task", raceOn(open))
	card := wantOnTheCard(t, handed)
	if text := textOf(t, handed); !strings.HasPrefix(text, "Accepted at attempt 1") || !strings.Contains(text, raceQuestion) {
		t.Errorf("the words are %q, want the acceptance and the question to read out", text)
	}
	wantHandedOut(t, kept, card.Task.ID)

	shown := awaited(t, session, open.ID)
	if shown.Screen != "task" || shown.Task == nil || shown.Task.ID != card.Task.ID || shown.Task.Question != raceQuestion ||
		shown.Child == nil || shown.Child.Pseudonym != "Otter" || shown.Language != "en" {
		t.Fatalf("read_task once the task is accepted = %+v, want the race on the card, as submit_task put it there", shown)
	}
	answered := payloadOf[answerPayload](t, call(t, session, "submit_answer",
		map[string]any{"task_id": shown.Task.ID, "answer": "C"}))
	if answered.Screen != "result" || answered.Result == nil || !answered.Result.Correct {
		t.Errorf("the answer from the card = %+v, want it recorded as right", answered)
	}

	h.settle()
	wantLessonLines(t, h, map[string]int{
		"task_requested": 1, "solver_run": 2, "task_submitted": 1, "task_accepted": 1, "answer_recorded": 1,
	})
	if accepted := linesOf(h, "task_accepted")[0].ContextMap(); accepted["topic"] != "logic.ordering" ||
		accepted["attempts"] != int64(1) {
		t.Errorf("task_accepted = %v, want the race accepted at the first attempt", accepted)
	}
}

// wantRequestOpened holds what next_task answered the race with to a card
// waiting for a new request's task, with nothing of its package — which the
// model fetches with get_package, and whose brief is the race's — and is the
// request as the profile keeps it: the model's choice, in English, with no
// attempt spent.
func wantRequestOpened(t *testing.T, session *mcp.ClientSession, asked *mcp.CallToolResult, kept store.Storage) *profile.OpenRequest {
	t.Helper()

	coming := wantComing(t, asked)
	text := fetchPackage(t, session, coming.RequestID)
	pack := packageIn(t, text)
	var brief profile.Brief
	if err := json.Unmarshal(pack["brief"], &brief); err != nil || brief.TargetConcept != "logic.ordering" {
		t.Errorf("the package carries the brief %+v (%v), want the race's topic", brief, err)
	}
	wantNoPackage(t, asked, pack)
	p, _ := loadKept(t, kept)
	open := p.OpenRequest
	if open == nil || coming.RequestID != open.ID || !strings.Contains(textOf(t, asked), "Request "+open.ID+" is open") ||
		!strings.Contains(leadOf(text), "The package of request "+open.ID) || open.Language != "en" ||
		open.Attempts != 0 || open.TutorMode != profile.TutorLLM || open.Brief.Difficulty != 2 {
		t.Fatalf("the request is %+v, want the model's choice in English with no attempt spent", open)
	}
	return open
}

// wantNoPackage holds a result of next_task to carrying nothing of the package
// a task is written from, in its words or in its payload: a card is drawn from
// the whole of the result, and the package — reference tasks with their
// answers among it — is for the model alone.
func wantNoPackage(t *testing.T, result *mcp.CallToolResult, pack map[string]json.RawMessage) {
	t.Helper()

	var examples []struct {
		Question string `json:"question"`
	}
	if err := json.Unmarshal(pack["examples"], &examples); err != nil || len(examples) == 0 {
		t.Fatalf("the package carries the examples %s (%v), want some to look for", pack["examples"], err)
	}
	var guide string
	if err := json.Unmarshal(pack["guide"], &guide); err != nil || len(guide) < 80 {
		t.Fatalf("the package carries the guide %q (%v), want one to look for", guide, err)
	}
	marks := []string{"Package:", `"examples"`, `"guide"`, `"solver_templates"`, examples[0].Question, guide[:80]}
	wantNoneOf(t, "the result of next_task", []string{textOf(t, result), string(rawPayload(t, result))}, marks)
}

// The words around a task keep the model from giving it away: while the task
// is written the child hears only that one is coming — so say the request and
// the package alike — and once the card shows it — accepted, or handed in
// again after it was — the model adds nothing of its own until the child
// answers or asks. A host may keep only the start of the instructions, so the
// words that come with the task are where the rule is sure to be read.
func TestTheWordsAroundATaskKeepTheModelQuiet(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)

	asked := call(t, session, "next_task", raceChoice)
	request := wantRequestOpened(t, session, asked, kept)
	race := raceOn(request)
	again := call(t, session, "next_task", raceChoice)
	pack := call(t, session, "get_package", map[string]any{"request_id": request.ID})
	for name, result := range map[string]*mcp.CallToolResult{
		"next_task opening the request": asked, "next_task asked again": again, "get_package": pack,
	} {
		if lead := leadOf(textOf(t, result)); !strings.Contains(lead, "tell the child only that one is on its way") {
			t.Errorf("%s says %q, want the child told only that a task is coming", name, lead)
		}
	}
	for _, handedIn := range []string{"when accepted", "when handed in again"} {
		if text := textOf(t, call(t, session, "submit_task", race)); !strings.Contains(text, "add nothing of your own about the task") {
			t.Errorf("submit_task %s says %q, want the model to add nothing about the task on the card", handedIn, text)
		}
	}
}

// The card records the answer the child gives on it, so the model asks for no
// answer in the chat: in a live lesson a model told only to record the answer
// asked the parent to type the letter in, and every such message spends one of
// the parent's chat messages on an answer the card already has. The model is
// told so where it reads it as the task comes — the words of the task accepted,
// or handed in again while the card shows it — and in the descriptions of the
// tools that hand a task in and record an answer; each says it of a host that
// shows cards, and the words of the task say what to do without one: read the
// task out and record the answer given in the chat.
func TestTheModelAsksForNoAnswerTheCardRecords(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	race := raceOn(wantRequestOpened(t, session, call(t, session, "next_task", raceChoice), kept))
	for _, handedIn := range []string{"when accepted", "when handed in again"} {
		text := textOf(t, call(t, session, "submit_task", race))
		for _, says := range []string{
			"Where the card next_task drew shows it, the child answers there, and the card records the answer itself: " +
				"do not ask for the answer in the chat",
			"Without a card, or if the child says the card shows no task, read out the question",
			"record the answer the child gives in the chat with submit_answer",
		} {
			if !strings.Contains(text, says) {
				t.Errorf("submit_task %s says %q, want it saying %q", handedIn, text, says)
			}
		}
	}

	listed, err := session.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	described := map[string]string{}
	for _, tool := range listed.Tools {
		described[tool.Name] = tool.Description
	}
	for tool, says := range map[string]string{
		"submit_task": "where cards are shown, the child answers on the card, which records the answer itself, so do " +
			"not ask for the answer in the chat",
		"submit_answer": "Where cards are shown, an answer given on the card is recorded by the card itself: do not ask " +
			"for one in the chat",
	} {
		if !strings.Contains(described[tool], says) {
			t.Errorf("%s is described as %q, want it saying %q", tool, described[tool], says)
		}
	}
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

// The package is handed out for the request that is open and still awaited,
// and for no other: not for a request never asked, nor another request's, nor
// one handed out already, nor one past its window. A model writing for such a
// request is told so, and sent to the request open when there is one, or else
// to the task on the card; nothing is written.
func TestAPackageIsHandedOutOnlyForTheOpenRequest(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		before func(t *testing.T, session *mcp.ClientSession, kept store.Storage, moving *clock) string
		says   string
	}{
		{"no request asked for", func(*testing.T, *mcp.ClientSession, store.Storage, *clock) string {
			return "req_nobody_asked"
		}, "Ask for a new task with next_task."},
		{"another request's id", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) string {
			askForTheRace(t, session, kept)
			return "req_another"
		}, "is the open one, and the card waits for its task: get its package with get_package"},
		{"a request handed out already", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) string {
			request := askForTheRace(t, session, kept)
			call(t, session, "submit_task", raceOn(request))
			return request.ID
		}, "is on the child's card: wait for the child's answer to it."},
		{"a request past its window", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, moving *clock) string {
			request := askForTheRace(t, session, kept)
			moving.advance(config.DefaultRequestWindow)
			return request.ID
		}, "Ask for a new task with next_task."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept, moving := racer(t), &clock{at: lessonDay}
			_, session := lessonWith(t, kept, moving, nil)
			id := tc.before(t, session, kept, moving)
			_, revision := loadKept(t, kept)

			result := call(t, session, "get_package", map[string]any{"request_id": id})
			refused, text := payloadOf[requestPayload](t, result), textOf(t, result)
			if refused.Screen != "waiting" || refused.Status != "stale" || refused.Code != "stale_request" ||
				!strings.Contains(text, "there is no package for it") || !strings.Contains(text, tc.says) ||
				strings.Contains(text, "Package:") {
				t.Errorf("get_package = %+v, %q, want it refused as stale, saying %q, with no package", refused, text, tc.says)
			}
			if _, now := loadKept(t, kept); now != revision {
				t.Error("get_package wrote the profile, want nothing written")
			}
		})
	}
}

// A card that waits for a request's task is told how it stands each time it
// asks: being written, and how many tries the checks have turned down; on the
// card, once it is handed out; and not coming once the request is over —
// out of attempts, past its window or replaced by the next — or when it asks
// about a request that never was. Asking writes nothing, and the answer is
// never a refusal.
func TestACardIsToldHowItsTaskStands(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		before  func(t *testing.T, session *mcp.ClientSession, kept store.Storage, moving *clock) string
		screen  string
		refused int
	}{
		{"just asked for", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) string {
			return askForTheRace(t, session, kept).ID
		}, "coming", 0},
		{"a try turned down", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) string {
			request := askForTheRace(t, session, kept)
			call(t, session, "submit_task", broken(request))
			return request.ID
		}, "coming", 1},
		{"every try turned down", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) string {
			request := askForTheRace(t, session, kept)
			for range profile.MaxAttempts {
				call(t, session, "submit_task", broken(request))
			}
			return request.ID
		}, "waiting", 0},
		{"handed out and answered", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) string {
			request := askForTheRace(t, session, kept)
			card := wantOnTheCard(t, call(t, session, "submit_task", raceOn(request)))
			call(t, session, "submit_answer", map[string]any{"task_id": card.Task.ID, "answer": "C"})
			return request.ID
		}, "task", 0},
		{"past its window", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, moving *clock) string {
			request := askForTheRace(t, session, kept)
			moving.advance(config.DefaultRequestWindow)
			return request.ID
		}, "waiting", 0},
		{"its task left for the next one", func(t *testing.T, session *mcp.ClientSession, kept store.Storage, _ *clock) string {
			request := askForTheRace(t, session, kept)
			call(t, session, "submit_task", raceOn(request))
			call(t, session, "next_task", raceChoice)
			return request.ID
		}, "waiting", 0},
		{"a request that never was", func(*testing.T, *mcp.ClientSession, store.Storage, *clock) string {
			return "req_never"
		}, "waiting", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept, moving := racer(t), &clock{at: lessonDay}
			h, session := lessonWith(t, kept, moving, nil)
			id := tc.before(t, session, kept, moving)
			_, revision := loadKept(t, kept)

			told := awaited(t, session, id)
			wantTold(t, &told, tc.screen, tc.refused)
			if _, now := loadKept(t, kept); now != revision {
				t.Error("read_task wrote the profile, want nothing written")
			}
			h.settle()
			if outcome := field(t, h.lineOf(t, "read_task"), "outcome"); outcome != "ok" {
				t.Errorf("the line of read_task says %q, want an answer, never a refusal", outcome)
			}
		})
	}
}

// wantTold holds what read_task told a card to this screen, with this many
// tries turned down, for the child the card is for: the task only on the
// task's screen, a request that is over named so, and never a status.
func wantTold(t *testing.T, got *awaitedPayload, screen string, refused int) {
	t.Helper()

	if got.Screen != screen || got.Refused != refused || got.Status != "" || got.Child == nil ||
		(got.Task != nil) != (screen == "task") {
		t.Errorf("read_task = %+v, want the %s screen with %d tries turned down, and no status", got, screen, refused)
	}
	if screen == "waiting" && got.Code != "stale_request" {
		t.Errorf("read_task = %+v, want a request that is over named so", got)
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
// request, marked open and aged, to a card that waits for its task like the
// first, sends a model that has not written the task for its package, and
// writes nothing — and the choice of the second call is not applied. Once the
// request has waited past its window, the next ask opens a new one.
func TestAskingAgainGivesTheSameRequest(t *testing.T) {
	t.Parallel()

	kept, moving := racer(t), &clock{at: lessonDay}
	h, session := lessonWith(t, kept, moving, nil)
	first := askForTheRace(t, session, kept)
	_, revision := loadKept(t, kept)

	moving.advance(30 * time.Second)
	again := call(t, session, "next_task", map[string]any{"language": "ru", "topic": "time.clocks", "reason": "clocks"})
	if card := wantComing(t, again); card.RequestID != first.ID || !card.AlreadyOpen || card.AgeSeconds != 30 ||
		card.Language != "en" {
		t.Errorf("next_task asked again = %+v, want request %s open for 30 seconds, in English", card, first.ID)
	}
	if text := textOf(t, again); !strings.Contains(text, "Request "+first.ID+" is already open, since 30 seconds ago") ||
		!strings.Contains(text, "If you have written its task, hand it in") ||
		!strings.Contains(text, "If you have not, a turn cut short say, the task is yours to write: get its package with get_package") ||
		!strings.Contains(text, "not applied") || strings.Contains(text, "Package:") {
		t.Errorf("the words are %q, want the request open for 30 seconds, its task asked for if written and its "+
			"package fetched if not, the choice not applied, and no package", text)
	}
	if p, now := loadKept(t, kept); now != revision || p.OpenRequest.Brief.TargetConcept != "logic.ordering" {
		t.Error("asking again wrote the profile or changed the request, want neither")
	}
	for language, ignored := range map[string]bool{"de": true, "EN": false} {
		if said := textOf(t, call(t, session, "next_task", map[string]any{"language": language})); strings.Contains(said, "not applied") != ignored {
			t.Errorf("asked again in %s, the words are %q, want the arguments said to be ignored: %v", language, said, ignored)
		}
	}

	moving.advance(config.DefaultRequestWindow)
	if later := wantComing(t, call(t, session, "next_task", raceChoice)); later.RequestID == first.ID || later.AlreadyOpen {
		t.Errorf("next_task past the window = %+v, want a new request", later)
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
// service's words, each by the code of the rule it broke, and nothing is
// written: the language a task is to be written in, always, and a choice of
// the model's own that the catalog can carry, with its reason. The card the
// refusal draws knows whose it is, and that no task comes to it.
func TestArgumentsNoRequestCanBeOpenedFromAreRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		arguments map[string]any
		problems  []string
	}{
		{"no language", map[string]any{"language": ""}, []string{"language required"}},
		{"not a language", map[string]any{"language": "not a language"}, []string{"language not_a_language"}},
		{"a topic with no reason", map[string]any{"language": "en", "topic": "time.clocks"}, []string{"reason required"}},
		{
			"a topic of spaces",
			map[string]any{"language": "en", "topic": "  "},
			[]string{"topic not_in_catalog", "reason required"},
		},
		{
			"a topic nobody has",
			map[string]any{"language": "en", "topic": "astronomy.stars", "reason": "stars"},
			[]string{"topic not_in_catalog"},
		},
		{
			"a level nobody has",
			map[string]any{"language": "en", "grade_level": "7-8", "reason": "older"},
			[]string{"grade_level not_one_of"},
		},
		{
			"a level the topic is not taught at",
			map[string]any{"language": "en", "topic": "percent.basic", "grade_level": "1-2", "reason": "easier"},
			[]string{"grade_level not_taught"},
		},
		{
			"a reason past its length",
			map[string]any{"language": "en", "topic": "time.clocks", "reason": strings.Repeat("why ", 100)},
			[]string{"reason too_long"},
		},
		{
			"everything at once",
			map[string]any{"language": "", "topic": "astronomy.stars", "difficulty": 9},
			[]string{"language required", "topic not_in_catalog", "difficulty out_of_range", "reason required"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			_, session := lesson(t, kept)
			_, revision := loadKept(t, kept)

			got := payloadOf[requestPayload](t, call(t, session, "next_task", tc.arguments))
			wantArgumentsRefused(t, &got, tc.problems)
			if p, now := loadKept(t, kept); now != revision || p.OpenRequest != nil {
				t.Error("a refusal wrote the profile, want nothing written")
			}
		})
	}
}

// wantArgumentsRefused holds what next_task answered to a refusal of these
// problems, each a field and the code of the rule it broke, on a card that
// knows whose it is and that no task comes to it.
func wantArgumentsRefused(t *testing.T, got *requestPayload, want []string) {
	t.Helper()

	problems := make([]string, 0, len(got.Problems))
	for _, problem := range got.Problems {
		problems = append(problems, problem.Field+" "+problem.Code)
	}
	if got.Screen != "waiting" || got.Status != "rejected" || got.Code != "invalid_arguments" ||
		!slices.Equal(problems, want) || got.RequestID != "" {
		t.Errorf("next_task = %+v, want %v refused as invalid_arguments, on a card no task comes to", got, want)
	}
	if got.Child == nil || got.Child.Pseudonym != "Otter" {
		t.Errorf("child = %+v, want the card a refusal draws to know whose it is", got.Child)
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

// The package names the idea of the topic its task is built on by the tasks
// of the topic the child has left behind, answered or skipped: a task skipped
// by the next ask moves the next one on to another idea, as an answer does.
func TestThePackageNamesTheIdeaByTheTopicsTasksLeftBehind(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	if p.Topics == nil {
		p.Topics = map[string]profile.Topic{}
	}
	ordering := p.Topics["logic.ordering"]
	ordering.Answers, ordering.Correct, ordering.Skipped = 12, 9, 3
	p.Topics["logic.ordering"] = ordering
	_, session := lesson(t, keptAsIs(t, p))

	// The race on the card is left for another race: the sixteenth task of the
	// topic, and the second time round the list.
	coming := wantComing(t, call(t, session, "next_task", raceChoice))
	type idea struct{ Number, Of, Round int }
	var got idea
	if err := json.Unmarshal(packageIn(t, fetchPackage(t, session, coming.RequestID))["idea"], &got); err != nil {
		t.Fatalf("read the package's idea: %v", err)
	}
	if want := (idea{Number: 7, Of: 10, Round: 2}); got != want {
		t.Errorf("idea = %+v after 12 answers, 3 skips and the race left, want %+v", got, want)
	}
}

// Nothing that gives the answer away leaves the seal before the child has
// answered: not in the payload a card draws, not in the words for the model,
// not in the open part of the profile, not on a span and not in a line. A card
// shows the task from three results — the task as the card next_task drew is
// told of it, the task accepted, and the same task handed in again, which shows
// the child the task they hold — and none of them carries it.
func TestTheAnswerStaysSealed(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)
	race := raceOn(request)
	handed := call(t, session, "submit_task", race)
	again := call(t, session, "submit_task", race)
	shown := call(t, session, "read_task", map[string]any{"request_id": request.ID})

	secrets := slices.Concat([]string{raceSolution, "correct_answer", "solution", "distractors"}, raceExplained, raceTraps)
	for _, card := range []struct {
		name   string
		result *mcp.CallToolResult
	}{{"accepted", handed}, {"handed in again", again}, {"read by the card that waits for it", shown}} {
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

// A lesson is held in the language the parent chose, whatever the chat is in:
// the request is opened in it and the model is told to write the task in it,
// the waiting card a refusal draws and the task's card speak it, and the same
// ask in the chat's language again finds nothing to differ from.
func TestALessonIsHeldInTheLanguageTheParentChose(t *testing.T) {
	t.Parallel()

	russian := "ru"
	kept := keptAsIs(t, profile.New(profile.Student{
		Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}, UILanguage: &russian,
	}, "test", lessonDay))
	_, session := lesson(t, kept)
	first := leadOf(textOf(t, call(t, session, "next_task", raceChoice)))
	p, _ := loadKept(t, kept)
	request := p.OpenRequest
	if request == nil || request.Language != russian {
		t.Fatalf("the request is %+v, want it in the language the parent chose, %s, over the chat's", request, russian)
	}
	again := leadOf(textOf(t, call(t, session, "next_task", map[string]any{"language": "en"})))
	if strings.Contains(again, "not applied") || !strings.Contains(again, "in "+russian+".") {
		t.Errorf("the ask again in the chat's language says %q, want the request kept in %s with nothing set aside", again, russian)
	}
	for name, lead := range map[string]string{"opening the request": first, "asked again": again} {
		if !strings.Contains(lead, "The parent chose "+russian+" for the lessons: talk to the child in it") {
			t.Errorf("next_task %s says %q, want the model told the language the parent chose, and to talk in it", name, lead)
		}
	}

	refused := payloadOf[handedInPayload](t, call(t, session, "submit_task", broken(request)))
	handed := payloadOf[handedInPayload](t, call(t, session, "submit_task", raceOn(request)))
	for _, card := range []handedInPayload{refused, handed} {
		if card.Child == nil || card.Child.UILanguage == nil || *card.Child.UILanguage != russian {
			t.Errorf("the %s card is for %+v, want the language the parent chose, %s", card.Screen, card.Child, russian)
		}
	}
	if refused.Screen != "waiting" || refused.Language != russian {
		t.Errorf("the refusal draws %q in %q, want a waiting card in %s", refused.Screen, refused.Language, russian)
	}
	if handed.Screen != "task" || handed.Task == nil || handed.Task.Language != russian || handed.Language != russian {
		t.Errorf("the task's card is %q in %q with %+v, want the task and its card in %s", handed.Screen, handed.Language, handed.Task, russian)
	}
}

// A task handed in for a request that is not the open one draws a card that
// waits for the one that is, and speaks that request's language. A request
// past its window is open no longer, and the card names no language of its
// own: it falls back to the parent's choice, or the host's.
func TestAStaleWaitingCardSpeaksTheOpenRequestsLanguage(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name   string
		waited time.Duration
		speaks bool
	}{
		{name: "while the other request is open", waited: 0, speaks: true},
		{name: "once it is past its window", waited: config.DefaultRequestWindow, speaks: false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			kept, moving := racer(t), &clock{at: lessonDay}
			_, session := lessonWith(t, kept, moving, nil)
			request := askForTheRace(t, session, kept)
			moving.advance(c.waited)
			race := raceOn(request)
			race["request_id"] = "req_another"

			want := ""
			if c.speaks {
				want = request.Language
			}
			stale := payloadOf[handedInPayload](t, call(t, session, "submit_task", race))
			if stale.Screen != "waiting" || stale.Status != "stale" || stale.Language != want {
				t.Errorf("submit_task = %+v, want a stale waiting card in %q", stale, want)
			}
		})
	}
}

// A language the parent chooses while a request is open changes nothing the
// model asked for: the same ask again is not told its arguments were set
// aside, and the request keeps the language it was opened in.
func TestALanguageChosenWhileARequestIsOpenSetsNoArgumentAside(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)
	saved := textOf(t, call(t, session, "save_profile", map[string]any{"ui_language": "ru"}))

	again := leadOf(textOf(t, call(t, session, "next_task", map[string]any{"language": "en"})))
	if strings.Contains(again, "not applied") || !strings.Contains(again, "Request "+request.ID+" is already open") {
		t.Errorf("the same ask again says %q, want the request handed back with nothing set aside", again)
	}
	stays := "The task already asked for stays in en, and its card speaks it; the next one comes in ru."
	for name, words := range map[string]string{"the saved profile": saved, "the ask again": again} {
		if !strings.Contains(words, stays) {
			t.Errorf("%s says %q, want the model told the task asked for keeps its language", name, words)
		}
	}
}

// A language typed into the file by hand that names no language is no choice:
// the model is told the lessons follow the chat's, as the requests do, and the
// words it reads never carry the text typed.
func TestALanguageTypedIntoTheFileThatNamesNoneIsNoChoice(t *testing.T) {
	t.Parallel()

	typed := "Russian please"
	kept := keptAsIs(t, profile.New(profile.Student{
		Grade: 2, Pseudonym: "Otter", Interests: []string{"sport"}, UILanguage: &typed,
	}, "test", lessonDay))
	_, session := lesson(t, kept)

	read := textOf(t, call(t, session, "get_profile", map[string]any{}))
	if !strings.Contains(read, "The lessons follow the chat's language.") || strings.Contains(read, typed) {
		t.Errorf("get_profile says %q, want the lessons following the chat's language and nothing of the text typed", read)
	}
	if request := askForTheRace(t, session, kept); request.Language != "en" {
		t.Errorf("the request is in %q, want the chat's, en", request.Language)
	}
}

// With no language chosen a lesson is held in the chat's, and the waiting card
// a refusal draws speaks it too, where it would otherwise fall back to the
// host's.
func TestWithNoLanguageChosenALessonIsHeldInTheChats(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)

	refused := payloadOf[handedInPayload](t, call(t, session, "submit_task", broken(request)))
	if request.Language != "en" || refused.Language != "en" {
		t.Errorf("the request is in %q and its refusal draws a card in %q, want both in the chat's, en", request.Language, refused.Language)
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
	p, _ := loadKept(t, kept)
	pack := fetchPackage(t, session, p.OpenRequest.ID)
	wantNoneOf(t, "the words of the request and the package", []string{textOf(t, asked), pack}, []string{name})
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

// With no profile there is no task to ask for, to write or to hand in, and
// none for a card to wait for: each says so, shows the first sign-in and
// writes nothing. A lesson starts with next_task, so its words are where the
// adult is first asked to say they are the child's parent or tutor, before any
// detail of the child.
func TestWithNoProfileThereIsNoTask(t *testing.T) {
	t.Parallel()

	kept := memory.New()
	_, session := lesson(t, kept)

	result := call(t, session, "next_task", raceChoice)
	asked, card := textOf(t, result), payloadOf[requestPayload](t, result)
	handed := payloadOf[handedInPayload](t, call(t, session, "submit_task",
		raceOn(&profile.OpenRequest{ID: "req_nobody"})))
	if !strings.HasPrefix(asked, "No task can be asked for yet. There is no profile yet.") || card.Screen != "first_run" ||
		handed.Screen != "first_run" || handed.Code != "stale_request" {
		t.Errorf("next_task says %q and draws %+v, and submit_task = %+v, want all to point at the first sign-in",
			asked, card, handed)
	}
	pack := wantWordsAlone(t, call(t, session, "get_package", map[string]any{"request_id": "req_nobody"}))
	if waiting := awaited(t, session, "req_nobody"); !strings.Contains(pack, "There is no profile yet.") ||
		waiting.Screen != "first_run" {
		t.Errorf("get_package says %q and read_task = %+v, want both to point at the first sign-in", pack, waiting)
	}
	if confirmed, pseudonym := strings.Index(asked, "parent or tutor"), strings.Index(asked, "ask for a pseudonym"); confirmed < 0 || pseudonym < confirmed {
		t.Errorf("next_task says %q, want the adult asked to say they are the parent or tutor before the pseudonym", asked)
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

	if again := wantComing(t, call(t, session, "next_task", raceChoice)); again.AlreadyOpen || again.RequestID == spent.ID {
		t.Errorf("next_task = %+v, want a new request in place of %s", again, spent.ID)
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

// A skipped task is read from the profile, which a person can edit, so its line
// names the task's topic only while the catalog has it: a name typed over the
// topic in the file stays in the file. So whichever way the task is left — for
// the next ask, or under a task handed out over it.
func TestTheLineOfASkippedTaskNamesOnlyACatalogTopic(t *testing.T) {
	t.Parallel()

	ways := map[string]func(t *testing.T, topic string) (*harness, *mcp.CallToolResult){
		"asked past": func(t *testing.T, topic string) (*harness, *mcp.CallToolResult) {
			p := raceOnTheCard(t, rating.TrialAnswers)
			p.CurrentTask.Topic = topic
			h, session := lesson(t, keptAsIs(t, p))
			return h, call(t, session, "next_task", map[string]any{"language": "en"})
		},
		"handed out over": func(t *testing.T, topic string) (*harness, *mcp.CallToolResult) {
			on := fuzzProfile(t)
			on.CurrentTask = &profile.CurrentTask{
				Difficulty: 3, Fingerprint: "sketch", GradeLevel: "3-4", Hint: "Count them.",
				ID: "tsk_left", InstructionsVersion: "test", IssuedAt: profile.At(lessonDay), Language: "en",
				Options: map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"},
				Sealed:  "mt1.t.kid.sealed", Topic: topic, Wording: "What time is it?",
			}
			h, session := lesson(t, keptAsIs(t, on))
			return h, call(t, session, "submit_task", raceOn(on.OpenRequest))
		},
	}
	for way, leave := range ways {
		for _, tc := range []struct{ name, topic, want string }{
			{"as the service wrote it", "time.clocks", "time.clocks"},
			{"a name typed over", pseudonym, other},
		} {
			t.Run(way+", "+tc.name, func(t *testing.T) {
				t.Parallel()

				h, left := leave(t, tc.topic)
				if left.IsError {
					t.Fatalf("the call that leaves the task failed: %s", textOf(t, left))
				}
				h.settle()
				skipped := linesOf(h, "task_skipped")
				if len(skipped) != 1 {
					t.Fatalf("task_skipped lines = %d, want 1", len(skipped))
				}
				if got := skipped[0].ContextMap()["topic"]; got != tc.want {
					t.Errorf("task_skipped names the topic %v, want %s", got, tc.want)
				}
			})
		}
	}
}

// A task handed out over one that has had its answer — a file edited by hand
// holding an open request beside it — replaces it with nothing recorded: its
// answer is in the window already, and it was never left.
func TestATaskHandedOutOverAnAnsweredOneSkipsNothing(t *testing.T) {
	t.Parallel()

	on := raceOnTheCard(t, 0)
	if _, err := on.Record(profile.Answered{TaskID: on.CurrentTask.ID, Choice: "C", At: lessonDay}, sealer(t), rating.GradeLevels()); err != nil {
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

// The progress counts the tasks left without an answer once, in all, and its
// words name the latest entries the card lists, a skip among them as skipped:
// the words and the card say the same.
func TestTheProgressCountsTheSkipsAsTheCardDoes(t *testing.T) {
	t.Parallel()

	kept := keptWith(t, "masha")
	p, revision := loadKept(t, kept)
	for i := range 6 {
		p.Recent = append(p.Recent, profile.Answer{
			AnsweredAt: profile.At(lessonDay), Difficulty: 2, GradeLevel: "3-4", Skipped: true,
			TaskID: fmt.Sprintf("tsk_skipped_%d", i), Topic: "time.clocks",
		})
	}
	clocks := p.Topics["time.clocks"]
	clocks.Skipped += 6
	p.Topics["time.clocks"] = clocks
	p.Touch("test", lessonDay)
	if _, err := kept.Save(t.Context(), devAccount, p, revision); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	_, session := lesson(t, kept)

	result := call(t, session, "get_progress", nil)
	if counted := payloadOf[progressPayload](t, result).Skipped; counted != 6 {
		t.Errorf("skipped = %d, want the 6 tasks left without an answer", counted)
	}
	text := textOf(t, result)
	if !strings.Contains(text, "The latest tasks, the latest first: Clocks skipped, Clocks skipped, "+
		"Clocks skipped, Clocks skipped, Clocks skipped.") ||
		!strings.Contains(text, "Tasks left without an answer in all: 6.") {
		t.Errorf("the words are %q, want the five latest entries named as the card lists them and the six skips "+
			"counted in all", text)
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
