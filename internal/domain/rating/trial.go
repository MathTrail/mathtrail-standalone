package rating

import "math"

// TrialAnswers is how many answers the trial series takes. The start is only a
// guess made from the grade, and the step of Update is built for a level that
// is roughly right already, so until this many answers are in, the level is
// estimated from all of them at once instead: a child the grade placed a level
// too high meets easier tasks from the second or third answer, rather than
// failing their way down with a step that narrows as they go.
const TrialAnswers = 5

const (
	// startSpread is how far from the start the estimate allows a child to
	// stand before any answer is in: one level shift, so that the answers of
	// the series can move a child a level either way. It was chosen by
	// simulation, as the balance between the child whose grade was entered a
	// level off, whom a narrower spread leaves too near the start, and the
	// child whose grade was right, whom a wider one throws about.
	startSpread = levelShift

	// gridStep is how finely the estimate first looks for the peak. It only
	// has to land in the right cell: the peak is then found exactly inside it.
	gridStep = 0.05
)

// Answer is one answer of the trial series: where its task stood on the
// ladder, and whether it was solved.
type Answer struct {
	Point   Point
	Correct bool
}

// Estimate is where a child most likely stands, given where they started and
// every answer of the trial series so far, weighed all at once: the level at
// which the start and those answers together are most probable. Before any
// answer that is the start itself. An answer to a point off the ladder says
// nothing about where the child stands, and is left out.
//
// The estimate can have two peaks — a child who answers several tasks far
// above the start correctly gives it one near the start and one near the
// tasks — so it is looked for everywhere it can be before it is pinned down,
// rather than climbed to from the start, which would stop at the nearer peak.
func Estimate(start float64, answers []Answer) float64 {
	onLadder := make([]weighed, 0, len(answers))
	for _, answer := range answers {
		if beta := answer.Point.Beta(); !math.IsNaN(beta) {
			onLadder = append(onLadder, weighed{beta: beta, correct: answer.Correct})
		}
	}
	if len(onLadder) == 0 || math.IsNaN(start) || math.IsInf(start, 0) {
		return start
	}

	// The start pulls the estimate back in proportion to the distance, and no
	// answer pulls with a slope steeper than one, so the peak lies within this
	// distance of the start.
	reach := float64(len(onLadder)) * startSpread * startSpread
	cells := int(math.Ceil(2 * reach / gridStep))

	best, highest := start, math.Inf(-1)
	for cell := 0; cell <= cells; cell++ {
		level := start - reach + float64(cell)*gridStep
		if likelihood := logPosterior(level, start, onLadder); likelihood > highest {
			best, highest = level, likelihood
		}
	}

	// The peak is inside the best cell's neighbours, where the slope turns
	// from rising to falling; halving on its sign pins it down to the last
	// bit a float can tell apart, which takes fewer halvings than allowed here.
	low, high := best-gridStep, best+gridStep
	for range halvings {
		middle := low + (high-low)/2
		if middle <= low || middle >= high {
			break
		}
		if slope(middle, start, onLadder) > 0 {
			low = middle
		} else {
			high = middle
		}
	}
	return low + (high-low)/2
}

// halvings is more than enough to narrow a cell to the spacing of a float: a
// float has 52 bits of fraction, and the cell's first bits are its own.
const halvings = 128

// weighed is an answer as the estimate weighs it: where its task stood on the
// scale, worked out once rather than at every level the search tries, and
// whether it was solved.
type weighed struct {
	beta    float64
	correct bool
}

// logPosterior is how probable a level is given the start and the answers, as
// a logarithm and up to a constant: the start's own belief, a normal curve
// around it, and the chance of every answer as it came.
func logPosterior(level, start float64, answers []weighed) float64 {
	fromStart := (level - start) / startSpread
	sum := -fromStart * fromStart / 2
	for _, answer := range answers {
		above := level - answer.beta
		if answer.correct {
			sum += math.Log(Guess + (1-Guess)*sigmoid(above))
		} else {
			// The chance of a wrong answer is (1 − Guess)·σ(−above), and its
			// logarithm is taken without forming it: far above a task the chance
			// underflows to zero, while its logarithm is still a plain number.
			sum += math.Log(1-Guess) - softplus(above)
		}
	}
	return sum
}

// slope is how fast logPosterior rises at a level: the start pulling back, and
// every answer pulling toward the levels that explain it.
func slope(level, start float64, answers []weighed) float64 {
	pull := -(level - start) / (startSpread * startSpread)
	for _, answer := range answers {
		s := sigmoid(level - answer.beta)
		if answer.correct {
			pull += (1 - Guess) * s * (1 - s) / (Guess + (1-Guess)*s)
		} else {
			pull -= s
		}
	}
	return pull
}

// softplus is ln(1 + eˣ), worked out so that it neither overflows for a large
// x nor loses the small one.
func softplus(x float64) float64 {
	if x > 0 {
		return x + math.Log1p(math.Exp(-x))
	}
	return math.Log1p(math.Exp(x))
}
