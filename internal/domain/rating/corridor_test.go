package rating_test

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// sameDifficulties fails when two lists of difficulties differ, and prints
// both rather than the first place they part.
func sameDifficulties(t *testing.T, got, want []int) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("inside the corridor = %v, want %v", got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("inside the corridor = %v, want %v", got, want)
			return
		}
	}
}

// The corridor read back: at its own bounds the chance of a correct answer is
// exactly the band it was defined by.
func TestTheCorridorBoundsAreTheBandItself(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(0)
	nearly(t, rating.Probability(0, corridor.BetaMin), 0.85, tolerance, "the chance at the hard end")
	nearly(t, rating.Probability(0, corridor.BetaMax), 0.70, tolerance, "the chance at the easy end")
	nearly(t, corridor.BetaMin, -1.4663, tolerance, "the hard end for a child who has answered nothing")
	nearly(t, corridor.BetaMax, -0.5108, tolerance, "the easy end for a child who has answered nothing")
}

// The corridor is narrower than the gap between two difficulties. That is why
// it holds one level and sometimes none, and why the recommendation is defined
// without it: a child can stand between two difficulties, and one of them
// still has to be handed out.
func TestTheCorridorIsNarrowerThanTheGapBetweenDifficulties(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(0)
	width := corridor.BetaMax - corridor.BetaMin
	nearly(t, width, 0.9555, tolerance, "the width of the corridor")
	if width >= 1 {
		t.Errorf("the corridor is %v wide, want it under the 1.0 between difficulties", width)
	}
}

// Where a child lands, and what is recommended there.
func TestWhatIsRecommendedAtEachLevel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		level       float64
		inside      []int
		recommended int
		fit         rating.Fit
	}{
		{
			name:        "a child who has answered nothing",
			level:       0,
			inside:      []int{2},
			recommended: 2,
			fit:         rating.FitInside,
		},
		{
			// Neither difficulty falls in the band: 2 is a shade too easy at
			// 0.854 and 3 a shade too hard at 0.698, and 0.698 is the nearer
			// of the two to the middle.
			name:        "a child standing between two difficulties",
			level:       0.5,
			inside:      nil,
			recommended: 3,
			fit:         rating.FitTooHard,
		},
		{
			name:        "a child far below the easiest task",
			level:       -2,
			inside:      nil,
			recommended: 1,
			fit:         rating.FitTooHard,
		},
		{
			name:        "a child far above the hardest task",
			level:       4,
			inside:      nil,
			recommended: 5,
			fit:         rating.FitTooEasy,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			corridor := rating.NewCorridor(tc.level)
			sameDifficulties(t, corridor.Inside, tc.inside)
			if corridor.Recommended != tc.recommended {
				t.Errorf("recommended = %d, want %d", corridor.Recommended, tc.recommended)
			}
			if corridor.Fit != tc.fit {
				t.Errorf("fit = %q, want %q", corridor.Fit, tc.fit)
			}
		})
	}
}

// The corridor slides toward easier tasks after every wrong answer, whether or
// not the whole number the child is offered moves with it.
func TestTheCorridorSlidesAfterAWrongAnswer(t *testing.T) {
	t.Parallel()

	state := rating.State{Theta: 0.9, Delta: 0.3, Answers: 20, TopicAnswers: 8}
	before := rating.NewCorridor(state.Level())

	for failure := 1; failure <= 3; failure++ {
		result := rating.Update(state, rating.Beta(3), false)
		after := rating.NewCorridor(result.Level())

		if after.BetaMin >= before.BetaMin || after.BetaMax >= before.BetaMax {
			t.Errorf("failure %d left the corridor at [%v, %v], want it below [%v, %v]",
				failure, after.BetaMin, after.BetaMax, before.BetaMin, before.BetaMax)
		}
		before = after
		state = result.State
	}
}

// The chances are read out by difficulty, not by where they sit in the array.
func TestTheChancesAreReadByDifficulty(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(0.3)
	for difficulty := 1; difficulty <= rating.Difficulties; difficulty++ {
		nearly(t, corridor.Probability(difficulty), rating.Probability(0.3, rating.Beta(difficulty)),
			tolerance, "the chance read out of the corridor")
	}
}

// Two difficulties can be exactly as far from the middle of the band as each
// other, and then the easier one is handed out: a child standing between two
// levels can finish the lower one, and a task that is finished teaches more
// than one that is abandoned.
//
// The tie is shown at a level no child reaches, because that is where it can
// be shown exactly: far below every task the chance is the guessing floor at
// all five difficulties, so all five are equally far from the middle. Between
// two real difficulties the same rule decides, and there the two distances are
// never equal to the last bit of a float.
func TestATieGoesToTheEasierDifficulty(t *testing.T) {
	t.Parallel()

	for _, level := range []float64{-1000, 1000} {
		corridor := rating.NewCorridor(level)

		first := corridor.Probability(1)
		for difficulty := 2; difficulty <= rating.Difficulties; difficulty++ {
			if corridor.Probability(difficulty) != first {
				t.Fatalf("at level %v difficulty %d is at %v and difficulty 1 at %v, want a tie to read",
					level, difficulty, corridor.Probability(difficulty), first)
			}
		}
		if corridor.Recommended != 1 {
			t.Errorf("at level %v every difficulty is equally far from the middle and %d was recommended, want 1",
				level, corridor.Recommended)
		}
	}
}
