package rating_test

import (
	"math"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// sameDifficulties fails when two lists of difficulties differ, and prints
// both rather than the first place they part.
func sameDifficulties(t *testing.T, got, want []int) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("inside the corridor = %v, want %v", got, want)
	}
}

// difficultiesOf reads the difficulties off points of the youngest level, the
// level every vector of the prototype was measured at.
func difficultiesOf(t *testing.T, points []rating.Point) []int {
	t.Helper()

	var difficulties []int
	for _, point := range points {
		if point.GradeLevel != rating.Grades12 {
			t.Errorf("%+v is not a point of %s", point, rating.Grades12)
		}
		difficulties = append(difficulties, point.Difficulty)
	}
	return difficulties
}

// youngestLevel is the five points of the youngest level: a topic taught there
// alone, which is every topic the prototype had.
func youngestLevel() []rating.Point { return rating.Points(rating.Grades12) }

// The corridor read back: at its own bounds the chance of a correct answer is
// exactly the band it was defined by.
func TestTheCorridorBoundsAreTheBandItself(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(0, youngestLevel())
	nearly(t, rating.Probability(0, corridor.BetaMin), 0.85, tolerance, "the chance at the hard end")
	nearly(t, rating.Probability(0, corridor.BetaMax), 0.70, tolerance, "the chance at the easy end")
	nearly(t, corridor.BetaMin, -1.4663, tolerance, "the hard end for a child who has answered nothing")
	nearly(t, corridor.BetaMax, -0.5108, tolerance, "the easy end for a child who has answered nothing")
}

// The corridor is narrower than the gap between two difficulties of a level.
// That is why it holds one of them and sometimes none, and why the
// recommendation is defined without it: a child can stand between two
// difficulties, and one of them still has to be handed out.
func TestTheCorridorIsNarrowerThanTheGapBetweenDifficulties(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(0, youngestLevel())
	width := corridor.BetaMax - corridor.BetaMin
	nearly(t, width, 0.9555, tolerance, "the width of the corridor")
	if width >= 1 {
		t.Errorf("the corridor is %v wide, want it under the 1.0 between difficulties", width)
	}
}

// Where a child lands among the difficulties of one level, and what is
// recommended there.
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

			corridor := rating.NewCorridor(tc.level, youngestLevel())
			sameDifficulties(t, difficultiesOf(t, corridor.Inside), tc.inside)
			if corridor.Recommended != youngest(tc.recommended) {
				t.Errorf("recommended = %+v, want difficulty %d", corridor.Recommended, tc.recommended)
			}
			if corridor.Fit != tc.fit {
				t.Errorf("fit = %q, want %q", corridor.Fit, tc.fit)
			}
		})
	}
}

// Where two levels overlap their points stand half a difficulty apart, and the
// corridor can hold one of each. The recommendation is still one point: the
// nearer to the middle of the band, whichever level it belongs to. A child of
// grade 3 a little above the start is the case: difficulty 5 of the youngest
// level is nearer the middle than difficulty 2 of the child's own, and the
// grade is not asked.
func TestWhereTheLevelsOverlapTheNearerPointIsRecommended(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(2.79, rating.Points(rating.GradeLevels()...))

	own, younger := rating.Point{GradeLevel: rating.Grades34, Difficulty: 2}, youngest(5)
	if want := []rating.Point{own, younger}; !slices.Equal(corridor.Inside, want) {
		t.Errorf("inside the corridor = %+v, want %+v, the easier first", corridor.Inside, want)
	}
	if corridor.Recommended != younger {
		t.Errorf("recommended = %+v, want %+v, at %.4f against %.4f", corridor.Recommended, younger,
			corridor.Probability(younger), corridor.Probability(own))
	}
	if corridor.Fit != rating.FitInside {
		t.Errorf("fit = %q, want %q", corridor.Fit, rating.FitInside)
	}
}

// The corridor is worked out over the points it is given, whatever order they
// come in, and lists them easiest first.
func TestTheCorridorListsItsPointsEasiestFirst(t *testing.T) {
	t.Parallel()

	shuffled := []rating.Point{
		{GradeLevel: rating.Grades56, Difficulty: 1},
		{GradeLevel: rating.Grades34, Difficulty: 5},
		{GradeLevel: rating.Grades34, Difficulty: 1},
	}
	corridor := rating.NewCorridor(3, shuffled)

	var got []rating.Point
	for _, chance := range corridor.Chances {
		got = append(got, chance.Point)
	}
	want := []rating.Point{shuffled[2], shuffled[0], shuffled[1]}
	if !slices.Equal(got, want) {
		t.Errorf("the corridor's points = %+v, want %+v", got, want)
	}
}

// The corridor slides toward easier tasks after every wrong answer, whether or
// not the whole number the child is offered moves with it.
func TestTheCorridorSlidesAfterAWrongAnswer(t *testing.T) {
	t.Parallel()

	state := rating.State{Theta: 0.9, Delta: 0.3, Answers: 20, TopicAnswers: 8}
	before := rating.NewCorridor(state.Level(), youngestLevel())

	for failure := 1; failure <= 3; failure++ {
		result := rating.Update(state, youngest(3).Beta(), false)
		after := rating.NewCorridor(result.Level(), youngestLevel())

		if after.BetaMin >= before.BetaMin || after.BetaMax >= before.BetaMax {
			t.Errorf("failure %d left the corridor at [%v, %v], want it below [%v, %v]",
				failure, after.BetaMin, after.BetaMax, before.BetaMin, before.BetaMax)
		}
		before = after
		state = result.State
	}
}

// The chances are read out by point, not by where they sit in the list, and a
// point the corridor was not worked out over has no chance to read.
func TestTheChancesAreReadByPoint(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(0.3, youngestLevel())
	for difficulty := 1; difficulty <= rating.Difficulties; difficulty++ {
		nearly(t, corridor.Probability(youngest(difficulty)), rating.Probability(0.3, youngest(difficulty).Beta()),
			tolerance, "the chance read out of the corridor")
	}
	if chance := corridor.Probability(rating.Point{GradeLevel: rating.Grades34, Difficulty: 1}); !math.IsNaN(chance) {
		t.Errorf("the chance at a point the corridor was not given = %v, want no number", chance)
	}
}

// Two points can be exactly as far from the middle of the band as each other,
// and then the one toward the corridor is handed out. Between two real points
// that is the easier one: a child standing between two can finish the lower
// one, and a task that is finished teaches more than one that is abandoned.
// When the curve has run flat, it is the one nearest the band.
//
// The tie is shown at levels no child reaches, because that is where it can
// be shown exactly: far below every task the chance is the guessing floor at
// every point, and far above it is certainty at every point, so all of them
// are equally far from the middle. Between two real points the two distances
// are never equal to the last bit of a float.
func TestATieGoesTowardTheCorridor(t *testing.T) {
	t.Parallel()

	ladder := rating.Points(rating.GradeLevels()...)
	for level, want := range map[float64]rating.Point{-1000: ladder[0], 1000: ladder[len(ladder)-1]} {
		corridor := rating.NewCorridor(level, ladder)

		first := corridor.Chances[0].Probability
		for _, chance := range corridor.Chances[1:] {
			if chance.Probability != first {
				t.Fatalf("at level %v %+v is at %v and the easiest point at %v, want a tie to read",
					level, chance.Point, chance.Probability, first)
			}
		}
		if corridor.Recommended != want {
			t.Errorf("at level %v every point is equally far from the middle and %+v was recommended, want %+v",
				level, corridor.Recommended, want)
		}
	}
}

// A level that is no number is a state no profile passes validation with, and
// a corridor built from one still recommends a point it was given, rather than
// none a caller could use — and says it cannot tell where that stands, rather
// than that it is inside.
func TestALevelThatIsNoNumberStillRecommendsAPoint(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(math.NaN(), youngestLevel())
	if !slices.Contains(youngestLevel(), corridor.Recommended) {
		t.Fatalf("Recommended = %+v, want one of the points given", corridor.Recommended)
	}
	if corridor.Fit != rating.FitUnknown {
		t.Errorf("Fit = %q, want %q", corridor.Fit, rating.FitUnknown)
	}
}

// A corridor over no points has nothing to recommend, and says so rather than
// recommending a point nobody gave it.
func TestACorridorOverNoPointsRecommendsNothing(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(0, nil)
	if corridor.Recommended != (rating.Point{}) || corridor.Fit != rating.FitUnknown {
		t.Errorf("recommended %+v with fit %q, want the zero point and %q",
			corridor.Recommended, corridor.Fit, rating.FitUnknown)
	}
	if len(corridor.Chances) != 0 || len(corridor.Inside) != 0 {
		t.Errorf("chances %v and inside %v, want neither", corridor.Chances, corridor.Inside)
	}
}
