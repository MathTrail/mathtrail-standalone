package mcpserver

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// An answer after the trial series falls in the range of its number, the last
// answer of a range included in it and the next one in the range after; past
// the last range, the answers are counted together.
func TestAnAnswerFallsInTheRangeOfItsNumber(t *testing.T) {
	t.Parallel()

	for answers, want := range map[int]string{
		6: "6-20", 20: "6-20", 21: "21-50", 50: "21-50", 51: "51-100", 100: "51-100",
		101: "101-200", 200: "101-200", 201: "201+", 1000: "201+",
	} {
		if got := answersBucket(answers); got != want {
			t.Errorf("answersBucket(%d) = %q, want %q", answers, got, want)
		}
	}
}

// Who chose a task is one of three words on a line whatever the task says: the
// rule, the model, or unknown — for a task that says nothing, and for one that
// names anybody else, which only a file edited by hand can.
func TestWhoChoseATaskIsOneOfThreeWords(t *testing.T) {
	t.Parallel()

	for mode, want := range map[profile.TutorMode]string{
		profile.TutorRule: "rule", profile.TutorLLM: "llm", "": "unknown", "teacher": "unknown",
	} {
		if got := chooserOf(mode); got != want {
			t.Errorf("chooserOf(%q) = %q, want %q", mode, got, want)
		}
	}
}
