package lesson_test

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// otter is the child of these cases.
var otter = lesson.Student{Pseudonym: "Otter", Grade: 2}

// childOf is a child of a run against the target, closed when the case ends.
func childOf(t *testing.T, target session.Target, name string) *session.Child {
	t.Helper()

	service := session.Open(target, 30*time.Second)
	child := service.Child(name)
	t.Cleanup(func() {
		if err := child.Close(); err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
		service.Close()
	})
	return child
}

// The request is read out of the words of get_package the way a model reads
// it: the id its words name before the package, which carries a brief.
func TestTheRequestIsReadOutOfTheWords(t *testing.T) {
	t.Parallel()

	const id = "req_0f8fad5b-d9cb-469f-a165-70867728950e"
	const pack = "\n\nPackage:\n" + `{"brief":{"topic":"logic.ordering"},"guide":"write a task"}`
	for _, test := range []struct{ name, words string }{
		{
			name:  "the package of a request",
			words: "The package of request " + id + ". Write one task in en to it, and hand it in with submit_task and request_id " + id + "." + pack,
		},
		{
			name: "the package after a word about the language",
			words: "The package of request " + id + ".\n" +
				"The parent chose ru for the lessons: talk in it, and every task is written in it." + pack,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request, err := lesson.ReadRequest(test.words)
			if err != nil {
				t.Fatalf("ReadRequest() error = %v, want nil", err)
			}
			if request.ID != id {
				t.Errorf("ID = %q, want %q", request.ID, id)
			}
		})
	}

	for _, test := range []struct{ name, words string }{
		{"no package", "Request " + id + " is open."},
		{"no request", "The child has had enough for today." + pack},
		{"a package with no brief", "Request " + id + " is open.\n\nPackage:\n" + `{"guide":"write a task"}`},
		{"a package that does not read", "Request " + id + " is open.\n\nPackage:\nnot json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if request, err := lesson.ReadRequest(test.words); err == nil {
				t.Errorf("ReadRequest() = %+v, want an error", request)
			}
		})
	}
}

// A card that shows another question, or another child, belongs to another
// call, whatever the service said of it: the hand-in is told as a mismatch.
func TestACardThatShowsAnotherTaskIsAMismatch(t *testing.T) {
	t.Parallel()

	task := lesson.Written()[0]
	for _, test := range []struct {
		name, question, pseudonym string
		want                      session.Kind
	}{
		{"the card of the task handed in", task.Body.Question, otter.Pseudonym, session.Answered},
		{"a card with another question", "Who finished last?", otter.Pseudonym, session.Mismatched},
		{"a card of another child", task.Body.Question, "Badger", session.Mismatched},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			card := map[string]any{
				"screen":  "task",
				"attempt": 1,
				"child":   map[string]any{"pseudonym": test.pseudonym, "grade": 2},
				"task":    map[string]any{"id": "tsk_1", "question": test.question},
			}
			target := servicetest.Fake(t, map[string]mcp.ToolHandler{
				"submit_task": func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
					return &mcp.CallToolResult{
						Content:           []mcp.Content{&mcp.TextContent{Text: "Accepted at attempt 1."}},
						StructuredContent: card,
					}, nil
				},
			}, nil)

			_, answer := lesson.HandIn(t.Context(), childOf(t, target, "otter"), lesson.Request{ID: "req_1"}, &task, otter)
			if answer.Kind != test.want {
				t.Errorf("HandIn().Kind = %q, want %q", answer.Kind, test.want)
			}
		})
	}
}

// kindsOf are the kinds of the answers given, in the order they came.
func kindsOf(answers []session.Answer) []session.Kind {
	kinds := make([]session.Kind, 0, len(answers))
	for i := range answers {
		kinds = append(kinds, answers[i].Kind)
	}
	return kinds
}

// A card waiting for a task, or a package, that belongs to another call is told
// as a mismatch: next_task drawing no card for a request, a package of another
// request than the card waits for, and a card told of anything but the task
// being written, or then of anything but the task the hand-in put on it.
func TestWhatBelongsToAnotherCallIsAMismatch(t *testing.T) {
	t.Parallel()

	const id = "req_0f8fad5b-d9cb-469f-a165-70867728950e"
	const other = "req_1b4e28ba-2fa1-11d2-883f-0016d3cca427"
	answering := func(payload any, words string) mcp.ToolHandler {
		return func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: words}}, StructuredContent: payload}, nil
		}
	}
	packageOf := func(request string) string {
		return "The package of request " + request + ".\n\nPackage:\n" + `{"brief":{"topic":"logic.ordering"}}`
	}
	coming := map[string]any{"screen": "coming", "request_id": id}
	card := lesson.Card{TaskID: "tsk_1", Question: "Who finished first?"}
	shown := func(taskID string) map[string]any {
		return map[string]any{"screen": "task", "task": map[string]any{"id": taskID, "question": card.Question}}
	}

	for _, test := range []struct {
		name  string
		tools map[string]mcp.ToolHandler
		take  func(child *session.Child) session.Kind
		want  session.Kind
	}{
		{"a card drawn for the request, and its package", map[string]mcp.ToolHandler{
			"next_task": answering(coming, "Request "+id+" is open."), "get_package": answering(nil, packageOf(id)),
		}, asking, session.Answered},
		{"no card drawn for the request", map[string]mcp.ToolHandler{
			"next_task":   answering(map[string]any{"screen": "waiting", "request_id": id}, "No task."),
			"get_package": answering(nil, packageOf(id)),
		}, asking, session.Mismatched},
		{"a card drawn for no request", map[string]mcp.ToolHandler{
			"next_task": answering(map[string]any{"screen": "coming"}, "A task is coming."), "get_package": answering(nil, packageOf(id)),
		}, asking, session.Mismatched},
		{"the package of another request", map[string]mcp.ToolHandler{
			"next_task": answering(coming, "Request "+id+" is open."), "get_package": answering(nil, packageOf(other)),
		}, asking, session.Mismatched},
		{"told the task is being written", map[string]mcp.ToolHandler{
			"read_task": answering(map[string]any{"screen": "coming"}, "Being written."),
		}, func(child *session.Child) session.Kind {
			return lesson.AwaitWriting(context.Background(), child, lesson.Request{ID: id}).Kind
		}, session.Answered},
		{"told of a task while it is being written", map[string]mcp.ToolHandler{
			"read_task": answering(shown("tsk_1"), "On the card."),
		}, func(child *session.Child) session.Kind {
			return lesson.AwaitWriting(context.Background(), child, lesson.Request{ID: id}).Kind
		}, session.Mismatched},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			child := childOf(t, servicetest.Fake(t, test.tools, nil), "otter")
			if got := test.take(child); got != test.want {
				t.Errorf("the answer is %q, want %q", got, test.want)
			}
		})
	}
}

// The card's question held beside the hand-in says it has heard of no try
// turned down, so that the service holds it; is asked again while the task is
// still being written; and takes for a mismatch a task other than the one the
// hand-in put on the card.
func TestAHeldQuestionIsAskedAsTheCardAsksIt(t *testing.T) {
	t.Parallel()

	const id = "req_0f8fad5b-d9cb-469f-a165-70867728950e"
	card := lesson.Card{TaskID: "tsk_1", Question: "Who finished first?"}
	told := func(payload map[string]any) *mcp.CallToolResult {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Told."}}, StructuredContent: payload}
	}
	coming := map[string]any{"screen": "coming"}
	shown := func(taskID string) map[string]any {
		return map[string]any{"screen": "task", "task": map[string]any{"id": taskID, "question": card.Question}}
	}
	for _, test := range []struct {
		name  string
		after map[string]any
		want  session.Kind
	}{
		{"told of the task handed in", shown("tsk_1"), session.Answered},
		{"told of another task", shown("tsk_2"), session.Mismatched},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fake := &readTasks{
				answers: []*mcp.CallToolResult{told(coming), told(coming), told(test.after)},
				seen:    make(chan struct{}, 8),
			}
			child := childOf(t, servicetest.Fake(t, map[string]mcp.ToolHandler{"read_task": fake.handle}, nil), "otter")

			// The hand-in lands once the card has twice heard the task is
			// being written.
			held := lesson.Hold(t.Context(), child, lesson.Request{ID: id}, 0, quickPace)
			<-fake.seen
			<-fake.seen
			answers := held.Card(card)

			want := []session.Kind{session.Answered, session.Answered, test.want}
			if !slices.Equal(kindsOf(answers), want) {
				t.Errorf("the held question was told %v, want %v", kindsOf(answers), want)
			}
			asked := fake.heard()
			if len(asked) != 3 {
				t.Errorf("read_task was asked %d times, want 3", len(asked))
			}
			for _, arguments := range asked {
				if arguments["request_id"] != id || arguments["refused"] != float64(0) {
					t.Errorf("read_task was asked %v, want the request and no try turned down heard of", arguments)
				}
			}
		})
	}
}

// The held question says the tries turned down it has heard of, so that the
// service holds it rather than answering it at once with what it said before.
func TestAHeldQuestionSaysTheTriesItHeardOf(t *testing.T) {
	t.Parallel()

	card := lesson.Card{TaskID: "tsk_1", Question: "Who finished first?"}
	told := func(payload map[string]any) *mcp.CallToolResult {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Told."}}, StructuredContent: payload}
	}
	fake := &readTasks{answers: []*mcp.CallToolResult{
		told(map[string]any{"screen": "coming", "refused": 1}),
		told(map[string]any{"screen": "task", "task": map[string]any{"id": "tsk_1", "question": card.Question}}),
	}}
	child := childOf(t, servicetest.Fake(t, map[string]mcp.ToolHandler{"read_task": fake.handle}, nil), "otter")

	lesson.Hold(t.Context(), child, lesson.Request{ID: "req_0f8fad5b-d9cb-469f-a165-70867728950e"}, 0, quickPace).Card(card)

	asked := fake.heard()
	if len(asked) != 2 || asked[0]["refused"] != float64(0) || asked[1]["refused"] != float64(1) {
		t.Errorf("read_task was asked %v, want no try turned down heard of and then the one it was told of", asked)
	}
}

// A held question still told the task is being written once the hand-in has
// landed asks once more, as a read that lags behind a write may need, and then
// settles, the answer it settled on a mismatch, rather than waiting for good.
func TestAHeldQuestionStillWritingAfterTheHandInSettles(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	questions := 0
	child := childOf(t, servicetest.Fake(t, map[string]mcp.ToolHandler{
		"read_task": func(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			mu.Lock()
			questions++
			mu.Unlock()
			// As the service holds a question that has no news.
			select {
			case <-ctx.Done():
			case <-time.After(50 * time.Millisecond):
			}
			return &mcp.CallToolResult{
				Content:           []mcp.Content{&mcp.TextContent{Text: "Being written."}},
				StructuredContent: map[string]any{"screen": "coming"},
			}, nil
		},
	}, nil), "otter")

	answers := lesson.Hold(t.Context(), child, lesson.Request{ID: "req_0f8fad5b-d9cb-469f-a165-70867728950e"}, 0, quickPace).
		Card(lesson.Card{TaskID: "tsk_1", Question: "Who finished first?"})

	if last := len(answers) - 1; last < 0 || answers[last].Kind != session.Mismatched {
		t.Errorf("the held question was told %v, want it to settle on a mismatch", kindsOf(answers))
	}
	mu.Lock()
	defer mu.Unlock()
	if questions > 3 {
		t.Errorf("read_task was asked %d times, want it settled by the second question after the hand-in", questions)
	}
}

// quickPace is the pace the cases' held questions keep, so that a case waits
// no card's four seconds between them.
const quickPace = 20 * time.Millisecond

// A held question answered at once waits the rest of the card's pace before it
// asks again, as a card does, rather than asking as fast as it is answered.
func TestAHeldQuestionKeepsTheCardsPace(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var started []time.Time
	child := childOf(t, servicetest.Fake(t, map[string]mcp.ToolHandler{
		"read_task": func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			mu.Lock()
			started = append(started, time.Now())
			refused := len(started)
			mu.Unlock()
			screen := map[string]any{"screen": "coming", "refused": refused}
			if refused == 3 {
				screen = map[string]any{"screen": "task", "task": map[string]any{"id": "tsk_1", "question": "Q"}}
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Told."}}, StructuredContent: screen}, nil
		},
	}, nil), "otter")

	const pace = 150 * time.Millisecond
	held := lesson.Hold(t.Context(), child, lesson.Request{ID: "req_0f8fad5b-d9cb-469f-a165-70867728950e"}, 0, pace)
	time.Sleep(2*pace + pace/2)
	held.Card(lesson.Card{TaskID: "tsk_1", Question: "Q"})

	mu.Lock()
	defer mu.Unlock()
	if len(started) != 3 {
		t.Fatalf("read_task was asked %d times, want 3", len(started))
	}
	// The pace runs from a question's start as the card sends it, which the
	// service hears a moment later; a question asked at once would come in
	// well under half of it.
	for i := 1; i < len(started); i++ {
		if gap := started[i].Sub(started[i-1]); gap < pace/2 {
			t.Errorf("question %d came %v after the one before, want about the card's pace of %v", i+1, gap, pace)
		}
	}
}

// readTasks answers read_task with its answers in turn, the last again after
// them, and keeps what each question was asked with; given seen, it says on it
// that it has answered, once for each question it has room for.
type readTasks struct {
	mu      sync.Mutex
	answers []*mcp.CallToolResult
	asked   []map[string]any
	seen    chan struct{}
}

func (r *readTasks) handle(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var arguments map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &arguments); err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.asked = append(r.asked, arguments)
	answer := r.answers[min(len(r.asked), len(r.answers))-1]
	r.mu.Unlock()
	// Said without waiting, so that a question too many fails the case
	// rather than holding the answer.
	select {
	case r.seen <- struct{}{}:
	default:
	}
	return answer, nil
}

// heard is what each question was asked with.
func (r *readTasks) heard() []map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.asked
}

// A held question let go, the hand-in having put no task on the card, leaves
// nothing it was told after that among what the card was told: the question
// still held ends as the service's hold does, and is not counted.
func TestAHeldQuestionLetGoLeavesNoCallCutOff(t *testing.T) {
	t.Parallel()

	var first sync.Once
	asked := make(chan struct{})
	child := childOf(t, servicetest.Fake(t, map[string]mcp.ToolHandler{
		"read_task": func(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			first.Do(func() { close(asked) })
			// As the service holds a question that has no news: until the
			// call ends, or its hold does.
			select {
			case <-ctx.Done():
			case <-time.After(200 * time.Millisecond):
			}
			return &mcp.CallToolResult{
				Content:           []mcp.Content{&mcp.TextContent{Text: "Being written."}},
				StructuredContent: map[string]any{"screen": "coming"},
			}, nil
		},
	}, nil), "otter")

	held := lesson.Hold(t.Context(), child, lesson.Request{ID: "req_0f8fad5b-d9cb-469f-a165-70867728950e"}, 0, quickPace)
	<-asked

	if answers := held.Drop(); len(answers) != 0 && answers[len(answers)-1].Kind != session.Answered {
		t.Errorf("the held question let go was told %v, want no call cut off among them", kindsOf(answers))
	}
}

// asking asks for the first written task, and is the kind of the last answer
// asking for it took.
func asking(child *session.Child) session.Kind {
	_, answers, _ := lesson.Ask(context.Background(), child, lesson.Written()[0].Choice)
	return answers[len(answers)-1].Kind
}

// Every written task is accepted by the service's own checks, in the order a
// lesson hands them in to one child — so none is a near-copy of one the child
// was handed before it — and each is answered, right and wrong in turn. The
// card waiting for each is told it is being written, and then that it is on
// the card.
func TestEveryWrittenTaskIsAcceptedInTheOrderALessonHandsItIn(t *testing.T) {
	t.Parallel()

	// Called without a pause, the lesson is faster than an account's pace allows.
	child := childOf(t, servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600"), "otter")
	for _, answer := range lesson.Start(t.Context(), child, otter) {
		if answer.Kind != session.Answered {
			t.Fatalf("%s: %q, %s, want it answered", answer.Tool, answer.Kind, answer.Text())
		}
	}

	for i, task := range lesson.Written() {
		letter := task.Body.Correct
		if i%2 == 1 {
			letter = task.Wrong
		}
		// A task not walked to its end leaves its request open, and the tasks
		// after it would only say so.
		if !t.Run(fmt.Sprintf("task %d", i+1), func(t *testing.T) { walkTask(t, child, &task, letter) }) {
			break
		}
	}
	if progress := lesson.Progress(t.Context(), child); progress.Kind != session.Answered {
		t.Errorf("get_progress %q, want it answered", progress.Kind)
	}
}

// walkTask takes the child through one written task, as a model and a card
// take it: asked for with its package, the card told it is being written,
// handed in and accepted, the card told it is on it, and answered with the
// letter given.
func walkTask(t *testing.T, child *session.Child, task *lesson.Task, letter string) {
	t.Helper()

	request, asked, err := lesson.Ask(t.Context(), child, task.Choice)
	if err != nil {
		t.Fatalf("asked for as %v: %v, want a request", kindsOf(asked), err)
	}
	writing := lesson.AwaitWriting(t.Context(), child, request)
	if writing.Kind != session.Answered {
		t.Errorf("read_task %q, %s, want the task being written", writing.Kind, writing.Text())
	}
	held := lesson.Hold(t.Context(), child, request, lesson.TriesTurnedDown(&writing), lesson.CardPace)
	card, handedIn := lesson.HandIn(t.Context(), child, request, task, otter)
	if handedIn.Kind != session.Answered {
		held.Drop()
		t.Fatalf("submit_task %q, %s, want it accepted", handedIn.Kind, handedIn.Text())
	}
	shown := held.Card(card)
	if len(shown) == 0 {
		t.Fatalf("the held question told the card nothing, want the task on the card")
	}
	for i := range shown {
		if shown[i].Kind != session.Answered {
			t.Errorf("read_task %q, %s, want the task being written and then on the card", shown[i].Kind, shown[i].Text())
		}
	}
	if answered := lesson.AnswerTask(t.Context(), child, card, letter); answered.Kind != session.Answered {
		t.Errorf("submit_answer %q, %s, want it recorded", answered.Kind, answered.Text())
	}
}

// No written task is a near-copy of another, or of a reference task of its
// level, by the measure the service's own check uses: one child may be handed
// every one of them, and no model could have copied one out of its package.
func TestNoWrittenTaskIsANearCopyOfAnotherOrOfAReferenceTask(t *testing.T) {
	t.Parallel()

	shipped, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	written := lesson.Written()
	for i, task := range written {
		var others []string
		for j, other := range written {
			if j != i {
				others = append(others, checks.Fingerprint(other.Body.Question, "en"))
			}
		}
		references := shipped.ReferenceQuestions(rating.GradeLevel(task.Choice.GradeLevel))
		if problems := checks.NearDuplicate(task.Body.Question, "en", references, others); len(problems) != 0 {
			t.Errorf("task %d, %q: %v, want no near-copy", i+1, task.Body.Question, problems)
		}
	}
}

// Every written task draws its solution, as every task the service takes
// does.
func TestEveryWrittenTaskDrawsItsSolution(t *testing.T) {
	t.Parallel()

	for _, task := range lesson.Written() {
		if task.Body.SolutionPicture == nil {
			t.Errorf("%q draws no picture of its solution", task.Body.Question)
		}
	}
}

// Every written task is handed in as the model writes one: the parts of the
// task under the names the service reads them by, the picture and the picture
// of the solution of the first among them.
func TestAWrittenTaskIsHandedInUnderTheNamesTheServiceReads(t *testing.T) {
	t.Parallel()

	written, err := json.Marshal(lesson.Written()[0].Body)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var parts map[string]json.RawMessage
	if err := json.Unmarshal(written, &parts); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	for _, name := range []string{
		"core_idea", "question", "picture", "options", "correct_answer", "hint", "solution", "solution_picture",
		"distractors",
	} {
		if _, found := parts[name]; !found {
			t.Errorf("the task carries no %q: %s", name, written)
		}
	}
}
