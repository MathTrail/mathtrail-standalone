package main

import (
	"maps"
	"math"
)

// twoSpeeds is a slow step and a fast one taken side by side over the same
// answers, the slow one's levels the estimate's, and a running sum of how much
// better the fast one has foretold the answers than the slow one, never let
// fall below nothing. Once the sum passes the threshold the child has most
// likely changed faster than the slow step follows, and the slow one takes the
// fast one's levels and the sum starts again. A sum that forgets would not
// do: a jump of a level makes each answer only some hundredths of a nat likelier
// under the fast step, and a sum that forgets a tenth of itself with every
// answer settles near ten times that, half a nat, short of any threshold worth
// the name.
type twoSpeeds struct {
	slow, fast *stepRule
	lead       float64
	threshold  float64
}

func newTwoSpeeds(start, threshold float64) *twoSpeeds {
	fast := newStepRule(start)
	fast.overallStep.decay, fast.topicStep.decay = 0, 0
	return &twoSpeeds{slow: newStepRule(start), fast: fast, threshold: threshold}
}

func (s *twoSpeeds) overall() float64                          { return s.slow.overall() }
func (s *twoSpeeds) level(topic string) float64                { return s.slow.level(topic) }
func (s *twoSpeeds) chance(topic string, beta float64) float64 { return s.slow.chance(topic, beta) }

// followed has both steps follow the trial series.
func (s *twoSpeeds) followed(topic string, beta float64, correct bool, estimate float64) {
	s.slow.followed(topic, beta, correct, estimate)
	s.fast.followed(topic, beta, correct, estimate)
}

// answered weighs how much likelier the answer was under the fast step than
// under the slow one, takes it into both, and moves the slow one onto the fast
// one's levels once the fast one's lead passes the threshold.
func (s *twoSpeeds) answered(topic string, beta float64, correct bool) {
	s.lead = max(0, s.lead+logChanceOf(correct, s.fast.chance(topic, beta))-logChanceOf(correct, s.slow.chance(topic, beta)))
	s.slow.answered(topic, beta, correct)
	s.fast.answered(topic, beta, correct)
	if s.lead > s.threshold {
		s.slow.theta = s.fast.theta
		maps.Copy(s.slow.delta, s.fast.delta)
		s.lead = 0
	}
}

// logChanceOf is the logarithm of the chance an answer as it came had.
func logChanceOf(correct bool, chance float64) float64 {
	if correct {
		return math.Log(chance)
	}
	return math.Log(1 - chance)
}
