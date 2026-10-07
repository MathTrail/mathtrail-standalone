package app_test

import (
	"encoding/json"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The reviewer the container builds is the whole review over the real content
// and the real sandbox: the race, written the way a model is asked to write a
// task, is accepted from one end to the other, its solver having run in the
// sandbox.
func TestTheContainersReviewerAcceptsAGoodTask(t *testing.T) {
	t.Parallel()

	brief := &profile.Brief{
		PedagogicalGoal: profile.GoalNewTopic, TargetConcept: "logic.ordering", GradeLevel: rating.Grades12,
		Difficulty: 2, Setting: "sport",
		TrapsToUse:     []string{"reversed_relation", "stopped_early"},
		ExcludedSkills: []string{}, Constraints: []string{}, Rationale: "A topic the child has not met yet.",
	}

	reviewer := newTestContainer(t).Reviewer
	examined, err := reviewer.Examine(t.Context(), &checks.Submission{
		Task: marshal(t, raceTask()), SelfCheck: marshal(t, raceSelfCheck()), Solver: raceSolver,
	})
	if err != nil {
		t.Fatalf("Examine() error = %v, want nil", err)
	}
	outcome, err := reviewer.Judge(examined, checks.Against{Asked: brief, Language: "en"})
	if err != nil {
		t.Fatalf("Judge() error = %v, want nil", err)
	}

	if !outcome.Accepted() {
		t.Fatalf("problems %v and unchecked %v, want the task accepted", outcome.Problems, outcome.Unchecked)
	}
	if event := outcome.Event(); event.SolverSteps == 0 || event.SolverTime == 0 {
		t.Errorf("Event() = %+v, want the cost of the solver's runs in the sandbox", event)
	}
}

func marshal(t *testing.T, part any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(part)
	if err != nil {
		t.Fatalf("marshal %T: %v", part, err)
	}
	return raw
}
