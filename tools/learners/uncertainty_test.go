package main

import (
	"math"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// An answer in the middle of the corridor tells 0.15 of a unit of uncertainty
// and moves the level by 0.927 for every unit: the numbers the step from
// counted answers is built on, read off the service's own chance.
func TestAnAnswerInTheMiddleOfTheCorridorTellsWhatTheModelSays(t *testing.T) {
	t.Parallel()
	if math.Abs(answerInformation-0.15) > 0.0005 || math.Abs(gainPerUncertainty-0.927) > 0.0005 {
		t.Errorf("information %v and gain %v at the middle of the corridor, want 0.15 and 0.927", answerInformation, gainPerUncertainty)
	}
	if got, want := informationAt(rating.CorridorMiddle), gainAt(rating.CorridorMiddle)*slopeAt(rating.CorridorMiddle); got != want {
		t.Errorf("information %v, want the gain times the slope, %v", got, want)
	}
}

// The service's step, read as one from an uncertainty narrowed by answers in
// the middle of the corridor, is that step again: its numbers in the
// step's natural terms — an uncertainty of a third of the answers' worth before
// any answer, gains of 0.6 and 1.2 — give back its first steps and its decay.
func TestTheServicesStepIsAStepFromAnUncertainty(t *testing.T) {
	t.Parallel()
	before := serviceDecay / answerInformation
	afterSeries := before / (1 + float64(rating.TrialAnswers)*answerInformation*before)
	for _, tc := range []struct {
		name string
		got  stepCurve
		want stepCurve
	}{
		{"overall", uncertain(afterSeries, rating.TrialAnswers, 0, serviceK0Theta/before), serviceOverallStep},
		{"topic", uncertain(before, 0, 0, serviceK0Delta/before), serviceTopicStep},
	} {
		if math.Abs(tc.got.first-tc.want.first) > 1e-12 || math.Abs(tc.got.decay-tc.want.decay) > 1e-12 || tc.got.added != 0 {
			t.Errorf("%s: %+v, want %+v", tc.name, tc.got, tc.want)
		}
	}
	if math.Abs(afterSeries-0.2667) > 0.0001 || math.Abs(math.Sqrt(before)-0.577) > 0.001 {
		t.Errorf("the service's uncertainty after the series %v and at the start %v, want 0.267 and 0.577 squared", afterSeries, before)
	}
}

// With nothing added the step is the formula, worked out as the formula and
// not carried answer by answer, which rounds differently; the two are one
// curve to twelve places.
func TestWithNothingAddedTheStepIsTheFormula(t *testing.T) {
	t.Parallel()
	c := uncertain(0.4, rating.TrialAnswers, 0, gainPerUncertainty)
	carried := c.first
	for n := range 300 {
		if got := c.at(n); got != c.first/(1+c.decay*float64(n)) {
			t.Fatalf("after %d answers the step is %v, want the formula's %v to the bit", n, got, c.first/(1+c.decay*float64(n)))
		}
		if math.Abs(c.at(n)-carried) > 1e-12*carried {
			t.Fatalf("after %d answers the formula gives %v and the step carried answer by answer %v", n, c.at(n), carried)
		}
		carried = carried/(1+c.decay/c.first*carried) + c.added
	}
}

// An uncertainty that answers narrow and q widens stays between where it
// started and where it settles, moves toward the latter with every answer and
// comes to it, and never falls below the same uncertainty with nothing added:
// on ten thousand draws of a start, q and gain, over their first sixty answers,
// and after two thousand.
func TestAnUncertaintyStaysBetweenItsStartAndItsSteadyValueAndSettles(t *testing.T) {
	t.Parallel()
	rng := seeded("uncertainty", "test")
	for range 10000 {
		start, q, gain := 0.01+2*rng.Float64(), 0.001+0.05*rng.Float64(), 0.3+1.2*rng.Float64()
		c, bare := uncertain(start, 0, q, gain), uncertain(start, 0, 0, gain)
		steady := gain * steadyUncertainty(q)
		low, high := min(c.at(0), steady), max(c.at(0), steady)
		previous := c.at(0)
		for n := 1; n <= 60; n++ {
			step := c.at(n)
			if step < low*(1-1e-12) || step > high*(1+1e-12) {
				t.Fatalf("start %v, q %v, gain %v: after %d answers the step %v lies outside [%v, %v]", start, q, gain, n, step, low, high)
			}
			if math.Abs(step-steady) > math.Abs(previous-steady)*(1+1e-12) {
				t.Fatalf("start %v, q %v, gain %v: after %d answers the step %v moved away from %v", start, q, gain, n, step, steady)
			}
			if step < bare.at(n)*(1-1e-12) {
				t.Fatalf("start %v, q %v, gain %v: after %d answers the step %v is below the one with nothing added, %v", start, q, gain, n, step, bare.at(n))
			}
			previous = step
		}
		if settled := c.at(2000); math.Abs(settled-steady) > 1e-9*steady {
			t.Fatalf("start %v, q %v, gain %v: after 2000 answers the step %v, want it settled at %v", start, q, gain, settled, steady)
		}
	}
}

// The steady uncertainty is where an answer takes away exactly what q adds.
func TestTheSteadyUncertaintyIsWhereAnAnswerTakesAwayWhatIsAdded(t *testing.T) {
	t.Parallel()
	for _, q := range []float64{0.0001, 0.003, 0.01, 0.03, 0.5} {
		v := steadyUncertainty(q)
		if after := v/(1+answerInformation*v) + q; math.Abs(after-v) > 1e-12 {
			t.Errorf("q %v: the steady uncertainty %v goes to %v after an answer", q, v, after)
		}
	}
}

// No answer moves the level in a topic, which the overall level and the
// topic's offset make up, by more than the limit; a move within the limit, or
// one with no limit, is left as it is, to the bit.
func TestTheLimitHoldsTheMoveOfALevel(t *testing.T) {
	t.Parallel()
	rng := seeded("limit", "test")
	for range 10000 {
		kTheta, kDelta, surprise, limit := rng.Float64(), rng.Float64(), 2*rng.Float64()-1, 0.05+rng.Float64()
		theta, delta := limited(kTheta, kDelta, surprise, limit)
		if move := (theta + delta) * math.Abs(surprise); move > limit*(1+1e-12) {
			t.Fatalf("steps %v and %v, surprise %v: a move of %v past the limit %v", kTheta, kDelta, surprise, move, limit)
		}
		if (kTheta+kDelta)*math.Abs(surprise) <= limit && (theta != kTheta || delta != kDelta) {
			t.Fatalf("steps %v and %v within the limit became %v and %v", kTheta, kDelta, theta, delta)
		}
		if theta, delta := limited(kTheta, kDelta, surprise, noLimit); theta != kTheta || delta != kDelta {
			t.Fatalf("steps %v and %v with no limit became %v and %v", kTheta, kDelta, theta, delta)
		}
	}
}

// An uncertainty after the series that no uncertainty before it narrows to
// in five answers is a mistake in the numbers, refused at once; one just
// within reach is taken.
func TestAnUncertaintyFiveAnswersCannotReachIsRefused(t *testing.T) {
	t.Parallel()
	reach := 1 / (float64(rating.TrialAnswers) * answerInformation)
	_ = uncertain(0.99*reach, rating.TrialAnswers, 0, gainPerUncertainty)
	for _, v := range []float64{reach, 1.01 * reach, 0, -1, math.NaN()} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("an uncertainty of %v after the series was taken, want it refused", v)
				}
			}()
			_ = uncertain(v, rating.TrialAnswers, 0, gainPerUncertainty)
		}()
	}
}
