package app_test

import (
	"encoding/json"
	"slices"
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

// The picture reaches a solver through every layer the container puts around
// the sandbox: the race told with its three runners drawn, asked for as a task
// solved from its picture, is accepted with a solver that counts the runners
// on the picture, and refused as one that left the picture unread with the
// solver that names them itself.
func TestTheContainersReviewerHandsTheSolverThePicture(t *testing.T) {
	t.Parallel()

	brief := &profile.Brief{
		PedagogicalGoal: profile.GoalNewTopic, TargetConcept: "logic.ordering", GradeLevel: rating.Grades12,
		Difficulty: 2, Setting: "sport",
		TrapsToUse:     []string{"reversed_relation", "stopped_early"},
		ExcludedSkills: []string{}, Constraints: []string{profile.FactsOnThePicture},
		Rationale: "A topic the child has not met yet.",
	}
	task := raceTask()
	task["picture"] = map[string]any{"kind": "row", "items": []map[string]any{{}, {}, {}}}
	const readsThePicture = `def solve(options, picture):
    names = ["Ann", "Ben", "Kim"][:len(picture["items"])]
    firsts = []
    for order in permutations(names):
        place = {name: i for i, name in enumerate(order)}
        if place["Ben"] < place["Kim"] and place["Kim"] < place["Ann"]:
            firsts.append(order[0])
    return match(options, firsts[0])
`

	reviewer := newTestContainer(t).Reviewer
	for _, test := range []struct {
		name, solver string
		want         []checks.Code
	}{
		{"a solver that reads the picture", readsThePicture, nil},
		{"a solver that names the runners itself", raceSolver, []checks.Code{checks.CodeSolverUnread}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			examined, err := reviewer.Examine(t.Context(), &checks.Submission{
				Task: marshal(t, task), SelfCheck: marshal(t, raceSelfCheck()), Solver: test.solver,
			})
			if err != nil {
				t.Fatalf("Examine() error = %v, want nil", err)
			}
			outcome, err := reviewer.Judge(examined, checks.Against{Asked: brief, Language: "en"})
			if err != nil {
				t.Fatalf("Judge() error = %v, want nil", err)
			}
			var codes []checks.Code
			for _, problem := range outcome.Problems {
				codes = append(codes, problem.Code)
			}
			if !slices.Equal(codes, test.want) {
				t.Errorf("problems %v and unchecked %v, want %v", outcome.Problems, outcome.Unchecked, test.want)
			}
		})
	}
}
