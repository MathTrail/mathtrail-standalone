package rating_test

import (
	"fmt"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The scale a family already knows: a child who has answered nothing starts at
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

// The ranks, boundary by boundary. Each step is one corridor wide and they sit
// either side of the starting rating, so a child who has answered nothing is
// at the bottom of the middle rank with room below and above.
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
		{shown: 2400, rank: 5},
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

	corridor := rating.NewCorridor(0)
	width := corridor.BetaMax - corridor.BetaMin
	inPoints := rating.Elo(width) - rating.Elo(0)

	nearly(t, inPoints, 166, 0.5, "the width of a rank in rating points")
}
