package rating

import "math"

const (
	// eloBase is where a child starts, and eloScale turns the levels of this
	// package into the chess numbers a family already understands: a
	// difference of 400 points is the same ten-to-one odds it is at a board.
	eloBase  = 1500
	eloScale = 400 / math.Ln10

	// rankStep is how far apart the ranks stand: one corridor — the distance
	// from "this is a stretch" to "this is a warm-up" — in rating points,
	// rounded to a whole number.
	rankStep = 166
)

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

// Rank is the step drawn above the rating, from 1 to 5. It is worked out
// wherever it is drawn and stored nowhere: a stored rank is one more thing
// that can disagree with the number beside it.
func Rank(elo float64) int {
	// The lowest rating of every rank above the first, either side of the
	// rating a child starts at, so that going up a rank means tasks a whole
	// corridor harder are now within reach. They are worked out from the step
	// rather than written down, and they live here because nothing else needs
	// them: a table at the top of the package is a table any other file can
	// edit by accident.
	floors := [...]int{eloBase - rankStep, eloBase, eloBase + rankStep, eloBase + 2*rankStep}

	rank, shown := 1, Shown(elo)
	for _, floor := range floors {
		if shown < floor {
			break
		}
		rank++
	}
	return rank
}
