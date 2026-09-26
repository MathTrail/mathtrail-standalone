package rating

import (
	"math"
	"slices"
)

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

// Fit says where the recommended point landed against the corridor. It is
// needed because the corridor is narrower than the gap between two
// difficulties of a level, so there is not always a point inside it.
type Fit string

const (
	// FitInside means the recommended point is within the corridor.
	FitInside Fit = "inside"
	// FitTooHard means the nearest point is still harder than the corridor:
	// the child would fail it more often than is useful.
	FitTooHard Fit = "too_hard"
	// FitTooEasy means the nearest point is easier than the corridor.
	FitTooEasy Fit = "too_easy"
	// FitUnknown means the level is no number, or there was no point to
	// weigh, so nothing can be said of where the recommendation stands.
	FitUnknown Fit = "unknown"
)

// Chance is the chance of a correct answer at one point of the ladder.
type Chance struct {
	Point
	Probability float64
}

// Corridor is the band of difficulty a child should be working in, and the
// one point of the ladder to ask for next.
type Corridor struct {
	// BetaMin and BetaMax bound the corridor on the difficulty scale. They
	// are what a search for a task looks through, and they slide toward easier
	// tasks after every wrong answer.
	BetaMin float64
	BetaMax float64
	// Chances are the points the corridor was worked out over, the easiest
	// first, each with the chance of a correct answer there.
	Chances []Chance
	// Inside lists the points whose chance falls within the corridor, the
	// easiest first. The corridor is narrower than the distance between two
	// difficulties of a level, so it holds at most one of each level; where
	// two levels overlap it can hold two points, and sometimes it holds none.
	Inside []Point
	// Recommended is the point to ask for next: the one whose chance of
	// success is closest to the middle of the corridor, which is defined even
	// when nothing is inside it.
	Recommended Point
	// Fit is where Recommended landed.
	Fit Fit
}

// NewCorridor works out the corridor for a child at this level in a topic,
// over the points the topic can be set at — a topic is not taught at every
// level, so it is the caller who knows which points there are. They may come
// in any order. With none at all there is nothing to recommend, and the
// recommendation is the zero point, its fit unknown.
func NewCorridor(level float64, points []Point) Corridor {
	corridor := Corridor{
		// The bounds are the corridor read backwards: the difficulty at which
		// this child's chance would be exactly the ceiling, and exactly the
		// floor. A higher chance means an easier task, so the ceiling gives
		// the lower bound.
		BetaMin: level - logit(corridorHigh),
		BetaMax: level - logit(corridorLow),
		Fit:     FitUnknown,
	}
	ordered := slices.Clone(points)
	slices.SortStableFunc(ordered, easierFirst)
	if len(ordered) == 0 {
		return corridor
	}

	// The distance from the middle of the band that the best point so far
	// stands at. Nothing has been chosen yet, so everything beats it. A level
	// that is no number leaves every distance unordered, and the middle point
	// stands for it rather than none.
	best, chance := math.Inf(1), math.NaN()
	corridor.Recommended = ordered[len(ordered)/2]

	for _, point := range ordered {
		probability := Probability(level, point.Beta())
		corridor.Chances = append(corridor.Chances, Chance{Point: point, Probability: probability})

		if corridorLow <= probability && probability <= corridorHigh {
			corridor.Inside = append(corridor.Inside, point)
		}
		// Ties go toward the corridor. Two points equally far from the middle
		// on either side of it mean the child is between them, and the easier
		// is the one that can be finished. Every point at the same chance means
		// the curve has run flat: at the floor the easiest is the nearest to
		// the corridor, and at the top the hardest is.
		from := math.Abs(probability - CorridorMiddle)
		if from < best || from == best && probability > CorridorMiddle {
			corridor.Recommended, chance, best = point, probability, from
		}
	}
	corridor.Fit = fitOf(chance)
	return corridor
}

// fitOf says where a chance stands against the corridor.
func fitOf(chance float64) Fit {
	switch {
	case math.IsNaN(chance):
		return FitUnknown
	case chance < corridorLow:
		return FitTooHard
	case chance > corridorHigh:
		return FitTooEasy
	default:
		return FitInside
	}
}

// Probability is the chance of a correct answer at this point, and no number
// for a point the corridor was not worked out over.
func (c *Corridor) Probability(point Point) float64 {
	for _, chance := range c.Chances {
		if chance.Point == point {
			return chance.Probability
		}
	}
	return math.NaN()
}
