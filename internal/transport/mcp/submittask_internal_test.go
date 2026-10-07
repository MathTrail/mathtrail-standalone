package mcpserver

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// A part is measured with its characters as they are: the escapes that keep <,
// > and & out of a web page would count six bytes for each, in a task whose
// words are full of comparisons.
func TestAPartIsMeasuredWithItsCharactersAsTheyAre(t *testing.T) {
	t.Parallel()

	for _, part := range []any{
		map[string]any{"question": "Is 3 < 5 & 5 > 4?"},
		`{"question":"Is 3 < 5 & 5 > 4?"}`,
	} {
		encoded, err := partJSON(part)
		if err != nil {
			t.Fatalf("partJSON(%v) error = %v", part, err)
		}
		if got := string(encoded); got != `{"question":"Is 3 < 5 & 5 > 4?"}` {
			t.Errorf("partJSON(%v) = %s, want the characters as they are", part, got)
		}
		if strings.HasSuffix(string(encoded), "\n") {
			t.Errorf("partJSON(%v) ends in a line break", part)
		}
	}
}

// The idea is measured as the task's JSON holds it, as the parts are, so that
// it is a share of the task's size: a quotation mark inside takes the two
// bytes of its escape, and the marks around it take none.
func TestTheIdeaIsMeasuredAsTheTasksJSONHoldsIt(t *testing.T) {
	t.Parallel()

	field := coreIdeaSize(&checks.Task{CoreIdea: `Tom is "second", not first.`})
	if want := len(`Tom is \"second\", not first.`); field.Key != "core_idea_bytes" || field.Integer != int64(want) {
		t.Errorf("coreIdeaSize = %s %d, want core_idea_bytes %d", field.Key, field.Integer, want)
	}
}
