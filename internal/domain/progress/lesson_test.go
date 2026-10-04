package progress_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
)

// What comes next is what the next task will be: the topic a person chose for
// the lessons once the trial series is over, said to be theirs, and the rule's
// own while the series runs.
func TestTheRecommendationFollowsAChosenTopic(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	for _, tc := range []struct {
		student string
		chosen  bool
	}{
		{"petya", true},
		{"masha", false},
	} {
		t.Run(tc.student, func(t *testing.T) {
			t.Parallel()

			p := fixture(t, tc.student)
			ruled, err := progress.Recommend(p, catalog)
			if err != nil {
				t.Fatalf("Recommend() error = %v", err)
			}
			if ruled.Topic == "arithmetic.tricks" || ruled.Chosen {
				t.Fatalf("the rule already sets %+v, so this case proves nothing", ruled)
			}
			p.Student.LessonTopic = "arithmetic.tricks"

			got, err := progress.Recommend(p, catalog)
			switch {
			case err != nil:
				t.Fatalf("Recommend() error = %v", err)
			case tc.chosen && (got.Topic != "arithmetic.tricks" || !got.Chosen):
				t.Errorf("next = %+v, want arithmetic.tricks, said to be chosen", got)
			case !tc.chosen && got != ruled:
				t.Errorf("next in the trial series = %+v, want the rule's own %+v", got, ruled)
			}
		})
	}
}

// The review suggests the topics its steps are for, in their order — the ones
// to develop, then the one to begin —, each once, three at most; a step for
// every topic names none, and no review suggests nothing.
func TestTheReviewSuggestsTheTopicsOfItsSteps(t *testing.T) {
	t.Parallel()

	review := &progress.Review{Steps: []progress.Step{
		{Kind: progress.StepTrap, Topic: "counting.gaps", Trap: "off_by_one"},
		{Kind: progress.StepPractice, Topic: "logic.ordering"},
		{Kind: progress.StepUnaided, Topic: "counting.gaps"},
		{Kind: progress.StepTrap, Trap: "missed_case"},
		{Kind: progress.StepBegin, Topic: "time.calendar", Base: "arithmetic.tricks"},
		{Kind: progress.StepRhythm, Topic: "parity.alternation"},
	}}
	if got, want := review.Suggested(), []string{"counting.gaps", "logic.ordering", "time.calendar"}; !slices.Equal(got, want) {
		t.Errorf("Suggested() = %v, want %v", got, want)
	}

	var none *progress.Review
	if got := none.Suggested(); got == nil || len(got) != 0 {
		t.Errorf("Suggested() of no review = %#v, want an empty list", got)
	}
}
