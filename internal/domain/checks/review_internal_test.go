package checks

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// An agreement with no run in it never comes out of Examine: its runs come
// from a verdict that returned no error, and such a verdict holds at least the
// first run. It is still a value its type allows, and a review handed one
// refuses the solver rather than trusting it or falling over — while the
// self-check is held against the task as always.
func TestAnAgreementWithNoRunsIsRefused(t *testing.T) {
	t.Parallel()

	var r reviewer
	problems, unchecked := r.solverRuns(&solver.Agreement{})
	if len(problems) != 1 || problems[0].Code != CodeSolverError || problems[0].Message == "" || unchecked != "" {
		t.Errorf("solverRuns() = %v, %q, want one refusal of the solver, in words", problems, unchecked)
	}

	draft := Draft{Task: &Task{CorrectAnswer: "C"}, SelfCheck: &SelfCheck{FinalAnswer: "B"}}
	problems, _ = r.answers(&solver.Agreement{}, draft)
	if len(problems) != 1 || !strings.Contains(problems[0].Message, "self_check.final_answer is not task.correct_answer") {
		t.Errorf("answers() = %v, want the self-check's disagreement and nothing about a solver that never ran",
			problems)
	}
}
