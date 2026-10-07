package mcpserver_test

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// The lines of the calls a task took name its request, so that the calls can
// be joined up — the ask, the package, the card's questions, the hand-in and
// the answer — and each names what its call answered. A request the profile
// does not hold is not named: what a caller sent is not written down.
func TestTheCallsOfATaskNameItsRequest(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h, session := lesson(t, kept)
	request := askForTheRace(t, session, kept)
	call(t, session, "get_package", map[string]any{"request_id": request.ID})
	call(t, session, "read_task", map[string]any{"request_id": request.ID})
	call(t, session, "submit_task", raceOn(request))
	call(t, session, "read_task", map[string]any{"request_id": request.ID})
	answerIt(t, session, profile.TaskIDFor(request.ID), "C", false)
	call(t, session, "read_task", map[string]any{"request_id": "req_nobody-holds-this"})
	h.settle()

	type said struct{ tool, request, screen, code string }
	want := []said{
		{"next_task", request.ID, "coming", ""},
		{"get_package", request.ID, "", ""},
		{"read_task", request.ID, "coming", ""},
		{"submit_task", request.ID, "task", ""},
		{"read_task", request.ID, "task", ""},
		{"submit_answer", request.ID, "result", ""},
		{"read_task", "", "waiting", "stale_request"},
	}
	lines := h.toolLines()
	if len(lines) != len(want) {
		t.Fatalf("tool_call lines = %d, want %d", len(lines), len(want))
	}
	for i := range lines {
		fields := lines[i].ContextMap()
		got := said{
			tool:    stringOf(fields["tool"]),
			request: stringOf(fields["task_request"]),
			screen:  stringOf(fields["screen"]),
			code:    stringOf(fields["code"]),
		}
		if got != want[i] {
			t.Errorf("line %d = %+v, want %+v", i+1, got, want[i])
		}
	}
}

// stringOf is a field's value as text, or nothing when the line has no such
// field.
func stringOf(value any) string {
	text, _ := value.(string)
	return text
}
