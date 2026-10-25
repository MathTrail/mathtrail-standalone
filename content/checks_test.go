package content_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The checks read the catalogs through interfaces of their own, and the content
// of the binary is what answers them.
var (
	_ checks.TrapDescriber = (*content.Content)(nil)
	_ checks.Content       = (*content.Content)(nil)
)

// A new task is compared with the reference tasks of its own level, and with
// no others: the questions handed out for a level are exactly those of its
// reference tasks, and a level there is not gets none.
func TestTheReferenceQuestionsAreThoseOfTheLevel(t *testing.T) {
	t.Parallel()

	shipped := loaded(t)
	perLevel := map[rating.GradeLevel][]string{}
	for _, task := range shipped.Examples() {
		perLevel[task.GradeLevel] = append(perLevel[task.GradeLevel], task.Question)
	}
	for _, level := range rating.GradeLevels() {
		if got, want := shipped.ReferenceQuestions(level), perLevel[level]; !slices.Equal(got, want) || len(want) == 0 {
			t.Errorf("ReferenceQuestions(%s) = %d questions, want the %d of the level", level, len(got), len(want))
		}
	}
	for _, level := range []rating.GradeLevel{"", "7-8"} {
		if got := shipped.ReferenceQuestions(level); len(got) != 0 {
			t.Errorf("ReferenceQuestions(%q) = %d questions, want none", level, len(got))
		}
	}
}

// The answers are the catalogs' own: an id they list is known, and one they
// do not — including one of another catalog — is not.
func TestTheCatalogsKnowTheirOwnIDs(t *testing.T) {
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
		for _, problem := range checks.Explanations(draft, "en", embedded) {
			t.Errorf("%s: %s", example.ID, problem.Message)
		}
	}
}

// The reference tasks are in English, and every one of them is written in the
// letters a task in English is held to: their labels, numbers and names cost
// none of them a refusal, as they would cost the task a model writes.
func TestEveryReferenceTaskIsWrittenInTheLettersOfEnglish(t *testing.T) {
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
		task := &checks.Task{
			Question: example.Question, Options: example.Options, Hint: example.Hint, Solution: example.Solution,
			Distractors: distractors,
		}
		for _, problem := range checks.Language(task, "en") {
			t.Errorf("%s: %s", example.ID, problem.Message)
		}
	}
}

// The reference tasks written for this service — every task of grades 5–6, and
// each task of grades 1–4 that took the place of one of the prototype's — are
// held to what a task the model writes is held to wherever it applies to them:
// a hint, and sentences a child of the youngest grade of their level can read,
// the strictest limits the level has.
func TestEveryTaskWrittenForTheServiceIsWrittenAsATaskMustBe(t *testing.T) {
	t.Parallel()

	ported := portedTasks(t)
	checked := 0
	for _, example := range loaded(t).Examples() {
		if ported[example.ID] {
			continue
		}
		checked++
		if strings.TrimSpace(example.Hint) == "" {
			t.Errorf("%s: got no hint, want one, as every task the model writes has", example.ID)
		}
		for _, problem := range checks.Readability(example.Question, "en", example.GradeLevel) {
			t.Errorf("%s: %s", example.ID, problem.Message)
		}
	}
	if checked == 0 {
		t.Fatal("no reference task written for this service, so nothing here was tested")
	}
}

// No reference task written for this service reads as a copy of another task
// of its level. A new task is held to the same measure against the reference
// tasks of its child's level, so a pair of reference tasks that close would
// mean a model following either one is refused as copying the other. Pairs of
// the prototype's own tasks are left out: it wrote many that close.
func TestNoTaskWrittenForTheServiceNearlyRepeatsATaskOfItsLevel(t *testing.T) {
	t.Parallel()

	ported := portedTasks(t)
	tasks := loaded(t).Examples()
	measured, closest, between := 0, 0.0, ""
	for i := range tasks {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].GradeLevel != tasks[j].GradeLevel || (ported[tasks[i].ID] && ported[tasks[j].ID]) {
				continue
			}
			measured++
			similarity := checks.Similarity(tasks[i].Question, tasks[j].Question)
			if similarity >= checks.Threshold {
				t.Errorf("%s and %s: similarity %.2f, want below %.2f", tasks[i].ID, tasks[j].ID, similarity, checks.Threshold)
			}
			if similarity > closest {
				closest, between = similarity, tasks[i].ID+" and "+tasks[j].ID
			}
		}
	}
	if measured == 0 {
		t.Fatal("no pair holds a reference task written for this service, so nothing here was tested")
	}
	t.Logf("the closest of %d pairs: %s, at %.2f", measured, between, closest)
}

// portedTasks are the ids of the reference tasks ported from the prototype, as
// its golden vectors name them, and only those: a task written for this service
// in the place of one of them has an id of its own.
func portedTasks(t *testing.T) map[string]bool {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "testdata", "golden", "readability.json"))
	if err != nil {
		t.Fatalf("read the prototype's golden vectors: %v", err)
	}
	var golden struct {
		Examples []struct {
			ID string `json:"id"`
		} `json:"examples"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse the prototype's golden vectors: %v", err)
	}
	ported := make(map[string]bool, len(golden.Examples))
	for _, example := range golden.Examples {
		ported[example.ID] = true
	}
	return ported
}
