package tutor_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The briefs in testdata/golden are the prototype's, built by its own rule on
// the five seed profiles. The profiles those briefs came from are the fixtures
// of the profile package, carried into this shape in T26 — so this is the one
// test that runs the whole way through: a real profile, the real catalogs of
// the binary, and a brief that has to match what another implementation
// produced from the same answers.
//
// The file names what v1 changed, and the vectors do not cover it: the goal
// "motivate" is gone, the setting rotates by a different count, and grades 5-6
// exist. Everything checked below is what did not change.

// goldenBrief is as much of a reference brief as v1 still promises to match.
type goldenBrief struct {
	Goal       string   `json:"pedagogical_goal"`
	Topic      string   `json:"target_concept"`
	Difficulty int      `json:"difficulty"`
	Setting    string   `json:"setting"`
	Traps      []string `json:"traps_to_use"`
	Excluded   []string `json:"excluded_skills"`
}

type goldenBriefs struct {
	Briefs []struct {
		Student string      `json:"student"`
		Brief   goldenBrief `json:"brief"`
	} `json:"briefs"`
}

func repoFile(t *testing.T, parts ...string) []byte {
	t.Helper()

	path := filepath.Join(append([]string{"..", "..", ".."}, parts...)...)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}

func fixture(t *testing.T, student string) *profile.Profile {
	t.Helper()

	p, err := profile.Parse(repoFile(t, "testdata", "profiles", student+".json"))
	if err != nil {
		t.Fatalf("parse the fixture of %s: %v", student, err)
	}
	return p
}

// embedded is the real catalog of the binary, which is what the reference
// briefs were built against.
func embedded(t *testing.T) *content.Content {
	t.Helper()

	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	return loaded
}

func TestTheRuleReproducesTheReferenceBriefs(t *testing.T) {
	t.Parallel()

	var golden goldenBriefs
	if err := json.Unmarshal(repoFile(t, "testdata", "golden", "rule_briefs.json"), &golden); err != nil {
		t.Fatalf("parse the reference briefs: %v", err)
	}
	if len(golden.Briefs) == 0 {
		t.Fatal("the reference carries no briefs")
	}

	catalog := embedded(t)
	for _, want := range golden.Briefs {
		t.Run(want.Student, func(t *testing.T) {
			t.Parallel()

			got, mode, err := tutor.Next(fixture(t, want.Student), catalog, tutor.Choice{})
			if err != nil {
				t.Fatalf("Next() error = %v, want nil", err)
			}
			if mode != profile.TutorRule {
				t.Errorf("mode = %q, want %q", mode, profile.TutorRule)
			}
			checkBrief(t, &got, &want.Brief)
		})
	}
}

// checkBrief holds one brief to one row of the reference.
//
// The setting is checked although v1 rotates by another count: these profiles
// have answered every task in their window and nothing has been pruned yet, so
// the two counts are the same number here. That is a property of the fixtures
// rather than of the rule, and it is why the agreement is worth stating.
func checkBrief(t *testing.T, got *profile.Brief, want *goldenBrief) {
	t.Helper()

	if string(got.PedagogicalGoal) != want.Goal {
		t.Errorf("goal = %q, want %q", got.PedagogicalGoal, want.Goal)
	}
	if got.TargetConcept != want.Topic {
		t.Errorf("topic = %q, want %q", got.TargetConcept, want.Topic)
	}
	if got.Difficulty != want.Difficulty {
		t.Errorf("difficulty = %d, want %d", got.Difficulty, want.Difficulty)
	}
	if got.Setting != want.Setting {
		t.Errorf("setting = %q, want %q", got.Setting, want.Setting)
	}
	if !slices.Equal(got.TrapsToUse, want.Traps) {
		t.Errorf("traps = %v, want %v", got.TrapsToUse, want.Traps)
	}
	if !slices.Equal(got.ExcludedSkills, want.Excluded) {
		t.Errorf("excluded skills = %v, want %v", got.ExcludedSkills, want.Excluded)
	}
}
