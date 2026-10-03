package rating

import "math"

const (
	// eloBase is where a child of the youngest level starts, and eloScale
	// turns the levels of this package into the chess numbers a family
	// already understands: a difference of 400 points is the same ten-to-one
	// odds it is at a board. The scale is one for every child, so the number
	// is a place on the ladder and an older child starts higher.
	eloBase  = 1500
	eloScale = 400 / math.Ln10

	// rankStep is how far apart the ranks stand: one corridor — the distance
	// from "this is a stretch" to "this is a warm-up" — in rating points,
	// rounded to a whole number.
	rankStep = 166

	// ranksBelowStart is how many ranks lie under the one a child of the
	// youngest level starts in.
	ranksBelowStart = 2
)

// Ranks is how many ranks there are: enough steps of one corridor, counted
// from where the youngest children start, to cover the whole ladder — a child
// for whom the easiest task of the youngest level is just right stands in the
// first, and one for whom the hardest task of the oldest level is, in the last.
const Ranks = 11

// Elo turns a level into the number shown to the child: the overall level
// gives the overall rating, and a level in one topic gives that topic's.
func Elo(level float64) float64 { return eloBase + eloScale*level }

// Shown is the rating as a reader sees it, which is the whole number the
// ranks are drawn against. A rating past what a whole number of four bytes
// holds is shown at that limit, and one that is no number at all as the rating
// a child starts at, rather than either turned into whatever the machine makes
// of it.
func Shown(elo float64) int {
	if math.IsNaN(elo) {
		return eloBase
	}
	return int(min(max(math.Round(elo), math.MinInt32), math.MaxInt32))
}

// Rank is the step drawn above the rating, from 1 to Ranks. It is worked out
// wherever it is drawn and stored nowhere: a stored rank is one more thing
// that can disagree with the number beside it.
//
// Every rank above the first begins one step above the one before, counted
// from the rating the youngest children start at, so that going up a rank
// means tasks a whole corridor harder are now within reach. The floors are
// worked out from the step rather than written down: a table of eleven numbers
// is eleven chances for one of them to be typed wrong.
func Rank(elo float64) int {
	rank, shown := 1, Shown(elo)
	for rank < Ranks && shown >= floorOf(rank+1) {
		rank++
	}
	return rank
}

// Share is how far through its rank the rating has come: a whole percent of
// the way from the rank's floor to the next one's, rounded down, so that the
// next rank is never drawn as reached before it is. Like the rank, it reads
// the number as drawn. The first rank is counted from a step below the
// second's floor, and a rating below even that has come no way at all; the
// last rank has no next one to come towards, and is drawn with the whole way
// behind it.
func Share(elo float64) int {
	rank := Rank(elo)
	if rank == Ranks {
		return 100
	}
	// The rating is held at the rank's floor before the floor is taken from
	// it: the lowest rating a whole number holds would wrap round otherwise,
	// where a whole number has four bytes.
	floor := floorOf(rank)
	return (max(Shown(elo), floor) - floor) * 100 / rankStep
}

// floorOf is the lowest rating drawn in rank, one step above the floor of the
// rank before, counted from the rating the youngest children start at. The
// first rank has no floor of its own: the one this gives it is where the way
// through it is counted from.
func floorOf(rank int) int {
	return eloBase + (rank-1-ranksBelowStart)*rankStep
}
