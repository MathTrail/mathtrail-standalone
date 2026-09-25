package rating

import "math"

// Difficulties is how many difficulty levels a task can have.
const Difficulties = 5

const (
	// corridorLow and corridorHigh bound the chance of a correct answer a
	// child should be meeting: hard enough to be worth doing, easy enough to
	// be done. Below the floor a child mostly fails and stops; above the
	// ceiling nothing new is learned.
	corridorLow  = 0.70
	corridorHigh = 0.85

	// CorridorMiddle is what the recommendation aims at, and the line an
	// answer has to be at or below to count toward mastering a topic: at the
	// middle of the band or harder, and not the easy half of it.
	CorridorMiddle = (corridorLow + corridorHigh) / 2
)

// Fit says where the recommended difficulty landed against the corridor. It is
// needed because the corridor is narrower than the gap between two
// difficulties, so there is not always a level inside it.
type Fit string

const (
	// FitInside means the recommended difficulty is within the corridor.
	FitInside Fit = "inside"
	// FitTooHard means the nearest difficulty is still harder than the
	// corridor: the child would fail it more often than is useful.
	FitTooHard Fit = "too_hard"
	// FitTooEasy means the nearest difficulty is easier than the corridor.
	FitTooEasy Fit = "too_easy"
	// FitUnknown means the level is no number, so nothing can be said of where
	// any difficulty stands; the middle one stands in for it.
	FitUnknown Fit = "unknown"
)

// Corridor is the band of difficulty a child should be working in, and the
// one level to ask for next.
type Corridor struct {
	// BetaMin and BetaMax bound the corridor on the difficulty scale. They
	// are what a search for a task looks through, and they slide toward easier
	// tasks after every wrong answer.
	BetaMin float64
	BetaMax float64
	// Probabilities is the chance of a correct answer at each difficulty, from
	// 1 at the front to 5 at the back.
	Probabilities [Difficulties]float64
	// Inside lists the difficulties whose probability falls within the
	// corridor. The corridor is narrower than the distance between two of
	// them, so this holds one level, and sometimes none.
	Inside []int
	// Recommended is the difficulty to ask for next: the one whose chance of
	// success is closest to the middle of the corridor, which is defined even
	// when nothing is inside it.
	Recommended int
	// Fit is where Recommended landed.
	Fit Fit
}

// NewCorridor works out the corridor for a child at this level in a topic.
func NewCorridor(level float64) Corridor {
	corridor := Corridor{
		// The bounds are the corridor read backwards: the difficulty at which
		// this child's chance would be exactly the ceiling, and exactly the
		// floor. A higher chance means an easier task, so the ceiling gives
		// the lower bound.
		BetaMin: level - logit(corridorHigh),
		BetaMax: level - logit(corridorLow),
	}

	// The distance from the middle of the band that the best difficulty so far
	// stands at. Nothing has been chosen yet, so everything beats it. A level
	// that is no number leaves every distance unordered, and the middle
	// difficulty stands for it rather than none.
	best := math.Inf(1)
	corridor.Recommended = middleDifficulty

	for difficulty := 1; difficulty <= Difficulties; difficulty++ {
		probability := Probability(level, Beta(difficulty))
		corridor.Probabilities[difficulty-1] = probability

		if corridorLow <= probability && probability <= corridorHigh {
			corridor.Inside = append(corridor.Inside, difficulty)
		}
		// Ties go toward the corridor. Two difficulties equally far from the
		// middle on either side of it mean the child is between them, and the
		// easier is the one that can be finished. Every difficulty at the same
		// chance means the curve has run flat: at the floor the easiest is the
		// nearest to the corridor, and at the top the hardest is.
		from := math.Abs(probability - CorridorMiddle)
		if from < best || from == best && probability > CorridorMiddle {
			corridor.Recommended, best = difficulty, from
		}
	}

	switch recommended := corridor.Probability(corridor.Recommended); {
	case math.IsNaN(recommended):
		corridor.Fit = FitUnknown
	case recommended < corridorLow:
		corridor.Fit = FitTooHard
	case recommended > corridorHigh:
		corridor.Fit = FitTooEasy
	default:
		corridor.Fit = FitInside
	}
	return corridor
}

// Probability is the chance of a correct answer at this difficulty, 1 to 5.
//
//nolint:gocritic // hugeParam: a pointer receiver would make NewCorridor(level).Probability(d) illegal, and the copy is five numbers read at most five times
func (c Corridor) Probability(difficulty int) float64 {
	return c.Probabilities[difficulty-1]
}
