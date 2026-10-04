package main

import (
	"fmt"
	"math"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// What one answer at the middle of the corridor, where the rule keeps a
// child's tasks, tells of a level under the chance with a guess: its
// information, by how much it narrows the uncertainty about the level, and the
// gain, how far a surprise of one moves the level for every unit of
// uncertainty, as a Kalman filter moves it there. A step worked out from an
// uncertainty counted by answers takes both from there, since it knows nothing
// of the chance a particular answer had.
var (
	answerInformation  = informationAt(rating.CorridorMiddle)
	gainPerUncertainty = gainAt(rating.CorridorMiddle)
)

// informationAt is the information an answer given at chance p carries about
// the level: the square of how fast the chance rises with the level, over the
// variance of the answer.
func informationAt(p float64) float64 {
	return gainAt(p) * slopeAt(p)
}

// gainAt is how far a surprise of one moves the level, for every unit of
// uncertainty about it, at chance p: how fast the chance rises with the level,
// over the variance of the answer. It is worked out without dividing by that
// variance, which is nothing where the chance is certain.
func gainAt(p float64) float64 {
	return (p - rating.Guess) / (p * (1 - rating.Guess))
}

// slopeAt is how fast the chance rises with the level where it stands at p.
func slopeAt(p float64) float64 {
	return (p - rating.Guess) * (1 - p) / (1 - rating.Guess)
}

// stepCurve is the step of one level, the overall one or a topic's, after n
// answers, in the step's own units: the first step, how fast it shrinks, and
// how much an answer adds back to it, from the answer `from` on.
//
// With nothing added back it is the service's formula, first/(1 + decay·n),
// worked out as the service works it out, so that the service's numbers give
// the service's step to the last bit. With something added back it is an
// uncertainty carried answer by answer from `from`, where it stands as the
// formula has it: each answer narrows it by what the answer tells and widens
// it by what is added, so that it never dies away. The two are one curve when
// nothing is added, but carrying it answer by answer rounds differently from
// the formula, which is why the formula is kept where it applies.
type stepCurve struct {
	first, decay, added float64
	from                int
}

// at is the step after n answers.
func (c stepCurve) at(n int) float64 {
	if c.added == 0 || n <= c.from {
		return c.first / (1 + c.decay*float64(n))
	}
	step := c.first / (1 + c.decay*float64(c.from))
	narrowing := c.decay / c.first
	for range n - c.from {
		step = step/(1+narrowing*step) + c.added
	}
	return step
}

// uncertain is the step of a level whose uncertainty stands at v after `from`
// answers and grows by q with every answer from there on, moved by gain for
// every unit of uncertainty. Before `from` the step is the formula's, from the
// uncertainty the level would have had before any answer for answers that add
// nothing to narrow it to v by `from`: no answer is taken in by the step
// before then, and the formula is what keeps the curve the service's when its
// own numbers are given. That uncertainty is v / (1 − from·I·v), which no
// count of answers reaches once from·I·v is one or more; such a v is a mistake
// in the numbers given, and is refused at once.
func uncertain(v float64, from int, q, gain float64) stepCurve {
	before := 1 / (1/v - float64(from)*answerInformation)
	if v <= 0 || math.IsNaN(before) || before <= 0 || math.IsInf(before, 0) {
		panic(fmt.Sprintf("learners: no uncertainty before %d answers narrows to %v", from, v))
	}
	return stepCurve{first: gain * before, decay: answerInformation * before, added: gain * q, from: from}
}

// steadyUncertainty is where the uncertainty an answer narrows and q widens
// settles: the v at which an answer takes away exactly q.
func steadyUncertainty(q float64) float64 {
	return q/2 + math.Sqrt(q*q/4+q/answerInformation)
}

// limited scales the steps of the overall level and of a topic down together,
// so that a surprise moves the level in the topic, which both make up, by no
// more than the limit. Steps within it are kept to the bit.
func limited(kTheta, kDelta, surprise, limit float64) (theta, delta float64) {
	scale := heldTo(limit, (kTheta+kDelta)*math.Abs(surprise))
	return kTheta * scale, kDelta * scale
}

// heldTo is the factor that brings a move down to the limit, or one, which
// changes nothing, for a move within it.
func heldTo(limit, move float64) float64 {
	if move > limit {
		return limit / move
	}
	return 1
}

// noLimit is the limit of a step that moves the level as far as it says.
var noLimit = math.Inf(1)
