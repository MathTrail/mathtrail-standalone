package tutor_test

import (
	"slices"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The worked example of the specification, run against the real catalogs of
// the binary. Every number in it was arrived at by hand from the profile
// printed in the architecture document, so a failure here is readable without
// a debugger: the topic just failed, the point of the ladder the corridor
// recommends for it, the setting the rotation lands on, and the two mistakes
// this child keeps making in that topic.
func TestTheWorkedExampleOfTheSpecification(t *testing.T) {
	t.Parallel()

	answered := time.Date(2026, 9, 20, 19, 2, 55, 0, time.UTC)
	p := profile.New(profile.Student{
		ExcludedSkills: []string{"division_with_remainder"},
		Grade:          3,
		Interests:      []string{"space", "dinosaurs", "football"},
		Pseudonym:      "Otter",
	}, "1.0.0", answered)

	p.Ratings = profile.Ratings{Answers: 57, ConsecutiveFailures: 1, Start: 2.5, Theta: 2.92}
	p.Topics = map[string]profile.Topic{
		"combinatorics.enumeration": {
			Answers: 9, Correct: 7, Delta: 0.31,
			LastIssued: profile.DateOf(answered),
			Traps:      map[string]int{"missed_case": 3, "double_count": 1},
		},
		"logic.truth_tellers": {
			Answers: 6, Correct: 4, Delta: -0.18,
			LastIssued: profile.DateOf(answered),
			Traps:      map[string]int{"negation_slip": 2},
		},
	}
	p.Recent = []profile.Answer{{
		AnsweredAt: profile.At(answered),
		Difficulty: 3,
		GradeLevel: rating.Grades34,
		Pace:       profile.PaceSlow,
		TaskID:     "tsk_01J9Z2A1B7",
		Topic:      "combinatorics.enumeration",
		HintUsed:   true,
		Chosen:     "B",
		Trap:       "missed_case",
	}}

	got, mode, err := tutor.Next(p, embedded(t), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}

	// A failure just happened, so the same ground is worked over again.
	if got.PedagogicalGoal != profile.GoalReinforce {
		t.Errorf("goal = %q, want %q", got.PedagogicalGoal, profile.GoalReinforce)
	}
	if got.TargetConcept != "combinatorics.enumeration" {
		t.Errorf("topic = %q, want the one just failed", got.TargetConcept)
	}

	// The child stands at 2.92 + 0.31 = 3.23 in that topic, which puts
	// difficulty 3 of grades 3–4 nearest the middle of the corridor and inside
	// it — difficulty 5 of grades 1–2 is inside too, further from the middle.
	// It is the same task they have just failed, because the corridor has slid
	// but not yet far enough to cross a point.
	if got.Difficulty != 3 || got.GradeLevel != rating.Grades34 {
		t.Errorf("difficulty %d of %s, want difficulty 3 of %s", got.Difficulty, got.GradeLevel, rating.Grades34)
	}

	// Fifty-seven answers over three interests lands on the first.
	if got.Setting != "space" {
		t.Errorf("setting = %q, want space", got.Setting)
	}

	if want := []string{"missed_case", "double_count"}; !slices.Equal(got.TrapsToUse, want) {
		t.Errorf("traps = %v, want %v", got.TrapsToUse, want)
	}
	if want := []string{"division_with_remainder"}; !slices.Equal(got.ExcludedSkills, want) {
		t.Errorf("excluded skills = %v, want %v", got.ExcludedSkills, want)
	}
	if mode != profile.TutorRule {
		t.Errorf("mode = %q, want %q", mode, profile.TutorRule)
	}
}
