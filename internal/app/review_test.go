package app_test

import (
	"encoding/json"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// raceSolver proves the answer to the race below the way a model is asked to:
// it tries every order, keeps the ones the wording allows, and looks up who
// came first among the options.
const raceSolver = `def solve(options):
    firsts = []
    for order in permutations(["Ann", "Ben", "Kim"]):
        place = {name: i for i, name in enumerate(order)}
        if place["Ben"] < place["Kim"] and place["Kim"] < place["Ann"]:
            firsts.append(order[0])
    return match(options, firsts[0])
`

// The reviewer the container builds is the whole review over the real content
// and the real sandbox: a task written the way a model is asked to write one —
// a topic and traps from the catalogs, a solver in Starlark — is accepted from
// one end to the other, its solver having run in the sandbox.
func TestTheContainersReviewerAcceptsAGoodTask(t *testing.T) {
	t.Parallel()

	brief := &profile.Brief{
		PedagogicalGoal: profile.GoalNewTopic, TargetConcept: "logic.ordering", Difficulty: 2, Setting: "sport",
		TrapsToUse:     []string{"reversed_relation", "stopped_early"},
		ExcludedSkills: []string{}, Constraints: []string{}, Rationale: "A topic the child has not met yet.",
	}
	task := map[string]any{
		"core_idea":              "Order three runners from two comparisons.",
		"design_thought_process": "Plot: a race. Traps: a reversed comparison, stopping early.",
		"question":               "Ann, Ben and Kim ran a race. Ben finished before Kim. Ann finished after Kim. Who finished first?",
		"options":                map[string]string{"A": "Ann", "B": "Kim", "C": "Ben", "D": "Nobody", "E": "All at once"},
		"correct_answer":         "C",
		"hint":                   "Who finished before Kim?",
		"solution":               "Ben is before Kim, and Kim is before Ann. So Ben is first.",
		"distractors": map[string]map[string]string{
			"A": {"trap": "reversed_relation", "text": "Ann finished after Kim, so she is last."},
			"B": {"trap": "stopped_early", "text": "Kim is in the middle: Ben beat her."},
			"D": {"trap": "ignored_condition", "text": "Someone always finishes first."},
			"E": {"trap": "answered_other_question", "text": "They finished one after another."},
		},
	}
	selfCheck := map[string]any{
		"issues": []any{},
		"option_check": map[string]string{
			"A": "Ann is last.", "B": "Kim is second.", "C": "Ben is first.", "D": "Someone was first.", "E": "Nobody tied.",
		},
		"final_answer": "C",
	}

	reviewer := newTestContainer(t).Reviewer
	examined, err := reviewer.Examine(t.Context(), &checks.Submission{
		Brief: marshal(t, brief), Task: marshal(t, task), SelfCheck: marshal(t, selfCheck), Solver: raceSolver,
	})
	if err != nil {
		t.Fatalf("Examine() error = %v, want nil", err)
	}
	outcome, err := reviewer.Judge(examined, checks.Against{Asked: brief, Language: "en", Grade: 3})
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
