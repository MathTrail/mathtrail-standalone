package rating_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The scale a family already knows: a child of the youngest level starts at
// 1500, and one level of this package is about 174 points.
func TestTheRatingShown(t *testing.T) {
	t.Parallel()

	for level, want := range map[float64]int{-1: 1326, 0: 1500, 1: 1674} {
		if got := rating.Shown(rating.Elo(level)); got != want {
			t.Errorf("the rating at level %v = %d, want %d", level, got, want)
		}
	}
	nearly(t, rating.Elo(1)-rating.Elo(0), 173.7178, tolerance, "one level in rating points")
}

// One scale for every child: the number is a place on the ladder, so an older
// child starts higher, in the rank the specification names.
func TestWhereEachGradeStartsOnTheScale(t *testing.T) {
	t.Parallel()

	for grade, want := range map[int]struct{ shown, rank int }{1: {1500, 3}, 3: {1934, 5}, 5: {2369, 8}} {
		elo := rating.Elo(rating.Start(grade))
		if shown, rank := rating.Shown(elo), rating.Rank(elo); shown != want.shown || rank != want.rank {
			t.Errorf("grade %d starts at %d in rank %d, want %d in rank %d", grade, shown, rank, want.shown, want.rank)
		}
	}
}

// The ranks, boundary by boundary. Each step is one corridor wide, counted
// from the rating the youngest children start at, so such a child is at the
// bottom of the third rank with two below and eight above.
func TestTheRanks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		shown int
		rank  int
	}{
		{shown: 800, rank: 1},
		{shown: 1333, rank: 1},
		{shown: 1334, rank: 2},
		{shown: 1499, rank: 2},
		{shown: 1500, rank: 3},
		{shown: 1665, rank: 3},
		{shown: 1666, rank: 4},
		{shown: 1831, rank: 4},
		{shown: 1832, rank: 5},
		{shown: 1997, rank: 5},
		{shown: 1998, rank: 6},
		{shown: 2163, rank: 6},
		{shown: 2164, rank: 7},
		{shown: 2329, rank: 7},
		{shown: 2330, rank: 8},
		{shown: 2495, rank: 8},
		{shown: 2496, rank: 9},
		{shown: 2661, rank: 9},
		{shown: 2662, rank: 10},
		{shown: 2827, rank: 10},
		{shown: 2828, rank: 11},
		{shown: 4000, rank: 11},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%d", tc.shown), func(t *testing.T) {
			t.Parallel()

			if got := rating.Rank(float64(tc.shown)); got != tc.rank {
				t.Errorf("the rank at %d = %d, want %d", tc.shown, got, tc.rank)
			}
		})
	}
}

// The ranks cover the ladder: a child for whom the easiest task of the
// youngest level is just right stands in the first, and one for whom the
// hardest task of the oldest level is, in the last.
func TestTheRanksCoverTheLadder(t *testing.T) {
	t.Parallel()

	ladder := rating.Points(rating.GradeLevels()...)
	// Just right is a chance of 0.775, which a child has 0.9383 above a task.
	const justRight = 0.9383
	for _, end := range []struct {
		point rating.Point
		rank  int
	}{
		{ladder[0], 1},
		{ladder[len(ladder)-1], rating.Ranks},
	} {
		level := end.point.Beta() + justRight
		nearly(t, rating.Probability(level, end.point.Beta()), rating.CorridorMiddle, tolerance,
			"the chance of a child for whom the task is just right")
		if got := rating.Rank(rating.Elo(level)); got != end.rank {
			t.Errorf("a child at home at %+v stands at %d in rank %d, want rank %d",
				end.point, rating.Shown(rating.Elo(level)), got, end.rank)
		}
	}
}

// A rank is drawn beside a number, so it reads the number as drawn. A rating
// of 1333.6 is shown as 1334 and must not carry the rank of 1333.
func TestARankFollowsTheNumberBesideIt(t *testing.T) {
	t.Parallel()

	if shown, rank := rating.Shown(1333.6), rating.Rank(1333.6); shown != 1334 || rank != 2 {
		t.Errorf("1333.6 is shown as %d with rank %d, want 1334 and rank 2", shown, rank)
	}
	if shown, rank := rating.Shown(1333.4), rating.Rank(1333.4); shown != 1333 || rank != 1 {
		t.Errorf("1333.4 is shown as %d with rank %d, want 1333 and rank 1", shown, rank)
	}
}

// Every rank is one corridor wide, which is what makes a step up mean "tasks a
// whole corridor harder are now within reach".
func TestARankIsOneCorridorWide(t *testing.T) {
	t.Parallel()

	corridor := rating.NewCorridor(0, youngestLevel())
	width := corridor.BetaMax - corridor.BetaMin
	inPoints := rating.Elo(width) - rating.Elo(0)

	nearly(t, inPoints, 166, 0.5, "the width of a rank in rating points")
}

// A rating past what four bytes of a whole number hold is shown at that limit,
// and ranked as the highest or the lowest it is, and one that is no number as
// the rating a child of the youngest level starts at, rather than either
// turned into whatever the machine makes of it.
func TestARatingPastAnyLimitIsShownAtIt(t *testing.T) {
	t.Parallel()

	if got := rating.Shown(1e30); got != math.MaxInt32 || rating.Rank(1e30) != rating.Ranks {
		t.Errorf("Shown(1e30) = %d at rank %d, want %d at rank %d", got, rating.Rank(1e30), math.MaxInt32, rating.Ranks)
	}
	if got := rating.Shown(-1e30); got != math.MinInt32 || rating.Rank(-1e30) != 1 {
		t.Errorf("Shown(-1e30) = %d at rank %d, want %d at rank 1", got, rating.Rank(-1e30), math.MinInt32)
	}
	if got := rating.Shown(math.NaN()); got != 1500 {
		t.Errorf("Shown(NaN) = %d, want 1500, the rating a child starts at", got)
	}
}
