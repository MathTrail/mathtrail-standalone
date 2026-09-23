package content_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// The checks read the catalogs through interfaces of their own, and the content
// of the binary is what answers them.
var (
	_ checks.Catalog       = (*content.Content)(nil)
	_ checks.TrapDescriber = (*content.Content)(nil)
	_ checks.Content       = (*content.Content)(nil)
)

// A new task is compared with the reference tasks of its child's level, and
// with no others: the questions handed out for a grade are exactly those of
// the level it belongs to, and a grade outside the levels gets none.
func TestTheReferenceQuestionsAreThoseOfTheGradesLevel(t *testing.T) {
	t.Parallel()

	shipped := loaded(t)
	perLevel := map[string][]string{}
	for _, task := range shipped.Examples() {
		perLevel[task.GradeLevel] = append(perLevel[task.GradeLevel], task.Question)
	}
	for grade := 1; grade <= 6; grade++ {
		level, _ := content.LevelOf(grade)
		if got, want := shipped.ReferenceQuestions(grade), perLevel[level]; !slices.Equal(got, want) {
			t.Errorf("ReferenceQuestions(%d) = %d questions, want the %d of level %s", grade, len(got), len(want), level)
		}
	}
	for _, grade := range []int{0, 7} {
		if got := shipped.ReferenceQuestions(grade); len(got) != 0 {
			t.Errorf("ReferenceQuestions(%d) = %d questions, want none", grade, len(got))
		}
	}
}

// The answers are the catalogs' own: an id they list is known, and one they
// do not — including one of another catalog — is not.
func TestTheCatalogsAnswerTheStructureCheck(t *testing.T) {
	t.Parallel()

	embedded, err := content.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	for _, test := range []struct {
		name string
		has  func(string) bool
		id   string
		want bool
	}{
		{"a topic", embedded.HasTopic, "combinatorics.enumeration", true},
		{"a trap", embedded.HasTrap, "off_by_one", true},
		{"a skill", embedded.HasSkill, "multiplication", true},
		{"a topic nobody has", embedded.HasTopic, "geometry.spheres", false},
		{"a trap id asked of the topics", embedded.HasTopic, "off_by_one", false},
		{"a topic id asked of the traps", embedded.HasTrap, "combinatorics.enumeration", false},
		{"a trap id asked of the skills", embedded.HasSkill, "off_by_one", false},
	} {
		if got := test.has(test.id); got != test.want {
			t.Errorf("%s: %q known = %v, want %v", test.name, test.id, got, test.want)
		}
	}
}

// The reference tasks are what the model imitates, so a check they failed would
// teach the model to fail it. Every one of them passes the check of the
// explanations behind the wrong options, as a submitted task has to.
func TestEveryReferenceTaskExplainsItsWrongOptions(t *testing.T) {
	t.Parallel()

	embedded, err := content.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	for _, example := range embedded.Examples() {
		distractors := make(map[string]checks.Distractor, len(example.Distractors))
		for letter, distractor := range example.Distractors {
			distractors[letter] = checks.Distractor{Trap: distractor.Trap, Text: distractor.Text}
		}
		draft := checks.Draft{Task: &checks.Task{
			Solution:    example.Solution,
			Hint:        example.Hint,
			Distractors: distractors,
		}}
		for _, problem := range checks.Explanations(draft, embedded) {
			t.Errorf("%s: %s", example.ID, problem.Message)
		}
	}
}
