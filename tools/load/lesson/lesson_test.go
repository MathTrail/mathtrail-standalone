package lesson_test

import (
	"context"
	"encoding/json"
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

// The request is read out of the words of next_task the way a model reads it:
// the id its words name before the package, and the brief of the package.
func TestTheRequestIsReadOutOfTheWords(t *testing.T) {
	t.Parallel()

	const id = "req_0f8fad5b-d9cb-469f-a165-70867728950e"
	const pack = "\n\nPackage:\n" + `{"brief":{"topic":"logic.ordering"},"guide":"write a task"}`
	for _, test := range []struct {
		name, words string
		wantBrief   string
	}{
		{
			name:      "a request opened",
			words:     "Request " + id + " is open. Write one task in en to the package below, and hand it in with submit_task and request_id " + id + "." + pack,
			wantBrief: `{"topic":"logic.ordering"}`,
		},
		{
			name: "a request opened after a task was skipped",
			words: "Task tsk_1b4e28ba-2fa1-11d2-883f-0016d3cca427, left on the child's card without an answer, is recorded as skipped.\n" +
				"Request " + id + " is open." + pack,
			wantBrief: `{"topic":"logic.ordering"}`,
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
			if string(request.Brief) != test.wantBrief {
				t.Errorf("Brief = %s, want %s", request.Brief, test.wantBrief)
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

// Every written task is accepted by the service's own checks, in the order a
// lesson hands them in to one child — so none is a near-copy of one the child
// was handed before it — and each is answered, right and wrong in turn.
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
		request, asked, err := lesson.Ask(t.Context(), child, task.Choice)
		if err != nil {
			t.Fatalf("task %d: next_task %q: %v, want a request", i+1, asked.Kind, err)
		}
		card, handedIn := lesson.HandIn(t.Context(), child, request, &task, otter)
		if handedIn.Kind != session.Answered {
			t.Fatalf("task %d: submit_task %q, %s, want it accepted", i+1, handedIn.Kind, handedIn.Text())
		}
		letter := task.Body.Correct
		if i%2 == 1 {
			letter = task.Wrong
		}
		if answered := lesson.AnswerTask(t.Context(), child, card, letter); answered.Kind != session.Answered {
			t.Errorf("task %d: submit_answer %q, %s, want it recorded", i+1, answered.Kind, answered.Text())
		}
	}
	if progress := lesson.Progress(t.Context(), child); progress.Kind != session.Answered {
		t.Errorf("get_progress %q, want it answered", progress.Kind)
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

// Every written task is handed in as the model writes one: the parts of the
// task under the names the service reads them by.
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
		"core_idea", "design_thought_process", "question", "options", "correct_answer", "hint", "solution", "distractors",
	} {
		if _, found := parts[name]; !found {
			t.Errorf("the task carries no %q: %s", name, written)
		}
	}
}
