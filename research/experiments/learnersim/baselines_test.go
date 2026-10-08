package main

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The baselines are answered in two topics at random and never in a third.
var (
	baselineTopics = []string{"a", "b", "never answered"}
	answeredTopics = baselineTopics[:2]
)

// baselineStart is where the baselines start, away from zero so that a level
// left at zero by mistake shows.
const baselineStart = 1.5

// levelsIn are an estimator's levels in the baselines' topics.
func levelsIn(est estimator) []float64 {
	levels := make([]float64, 0, len(baselineTopics))
	for _, topic := range baselineTopics {
		levels = append(levels, est.level(topic))
	}
	return levels
}

// answerWithin gives an estimator one answer and fails the test when it moves
// a level its structure holds still: under one level every topic stands where
// the overall level does, and under a level per topic the overall level stays
// at the start and no topic moves but the one answered. It says how far the
// level of the topic answered moved.
func answerWithin(t *testing.T, est estimator, shape structure, topic string, beta float64, correct bool) float64 {
	t.Helper()
	before := levelsIn(est)
	est.answered(topic, beta, correct)
	after := levelsIn(est)
	for i, other := range baselineTopics {
		switch {
		case shape == general && after[i] != est.overall():
			t.Fatalf("after an answer in %s, %s stands at %v and the overall level at %v, want one level", topic, other, after[i], est.overall())
		case shape == topics && other != topic && after[i] != before[i]:
			t.Fatalf("an answer in %s moved %s from %v to %v", topic, other, before[i], after[i])
		}
	}
	if shape == topics && est.overall() != baselineStart {
		t.Fatalf("after an answer in %s the overall level is %v, want it held at the start %v", topic, est.overall(), baselineStart)
	}
	answered := slices.Index(baselineTopics, topic)
	return after[answered] - before[answered]
}

// Glicko-2 moves the level of the topic answered toward the outcome, keeps its
// chance between its floor and certainty, and moves only what its structure
// lets move: one level for every topic, or each topic's own with the overall
// level held at the start.
func TestGlickoFollowsTheOutcomeWithinItsStructure(t *testing.T) {
	t.Parallel()
	for _, shape := range []structure{general, topics} {
		for _, floor := range []float64{0, rating.Guess} {
			t.Run(fmt.Sprintf("%s, floor %v", shape, floor), func(t *testing.T) {
				t.Parallel()
				answerGlicko(t, shape, floor)
			})
		}
	}
}

// answerGlicko answers Glicko-2 2,000 times at random, at tasks within three
// logits of its level, and fails the test at a chance outside its floor and
// certainty, and at an answer that does not move the level toward its outcome
// or moves one the structure holds still.
func answerGlicko(t *testing.T, shape structure, floor float64) {
	t.Helper()
	g := newGlicko(shape, baselineStart, floor)
	rng := seeded(fmt.Sprintf("glicko/%s/%v", shape, floor), "test")
	for range 2000 {
		topic := answeredTopics[rng.IntN(len(answeredTopics))]
		level := g.level(topic)
		for _, away := range []float64{-40, -3, 0, 3, 40} {
			if chance := g.chance(topic, level+away); chance < floor || chance > 1 {
				t.Fatalf("the chance %v from the level is %v, want it within [%v, 1]", away, chance, floor)
			}
		}
		beta, correct := level+6*rng.Float64()-3, rng.IntN(2) == 0
		moved := answerWithin(t, g, shape, topic, beta, correct)
		if correct && moved <= 0 || !correct && moved >= 0 {
			t.Fatalf("an answer, correct %v, at %.3f from the level moved it by %v", correct, beta-level, moved)
		}
	}
}

// With no floor the floored term is the published Glicko-2 term, so that
// Glicko-2 with a floor departs from the published rating by the floor alone.
// Past some 37 logits above a task the floored term divides nothing by
// nothing, so it is compared within 20.
func TestTheFlooredTermWithoutAFloorIsThePublishedOne(t *testing.T) {
	t.Parallel()
	rng := seeded("floored term", "test")
	for range 10000 {
		mu := 10*rng.Float64() - 5
		beta, score := mu+40*rng.Float64()-20, float64(rng.IntN(2))
		got, want := flooredTerm(0, mu, beta, score), logisticTerm(mu, beta, 0, score)
		if math.Abs(got.information-want.information) > 1e-12 || math.Abs(got.gradient-want.gradient) > 1e-12 {
			t.Fatalf("μ %v, β %v, score %v: floored term %+v, published term %+v", mu, beta, score, got, want)
		}
	}
}

// The volatility falls after a period that went exactly as expected and rises
// after one more surprising than the rating's deviation and the period's
// variance together explain: a player whose results surprise nobody is a
// steady one. A surprise that large sets the far end of the interval the new
// volatility's logarithm is looked for in, and the logarithm lies within it.
func TestTheVolatilityFallsWithoutASurpriseAndRisesWithALargeOne(t *testing.T) {
	t.Parallel()
	phi, sigma := glickoDeviation, glickoVolatility
	for _, v := range []float64{0.25, 4, 40} {
		t.Run(fmt.Sprintf("variance %v", v), func(t *testing.T) {
			t.Parallel()
			if got := newVolatility(phi, sigma, v, 0); got >= sigma {
				t.Errorf("no surprise: volatility %v, want it below %v", got, sigma)
			}
			surprise := 2 * math.Sqrt(phi*phi+v+sigma*sigma)
			got := newVolatility(phi, sigma, v, surprise)
			if got <= sigma {
				t.Errorf("surprise %v: volatility %v, want it above %v", surprise, got, sigma)
			}
			// The far end of so large a surprise is set from the surprise
			// alone, with no search, so the function searched is left out.
			start := math.Log(sigma * sigma)
			end := bracket(start, phi, v, surprise, func(float64) float64 { return math.NaN() })
			if found := math.Log(got * got); found <= start || found > end {
				t.Errorf("surprise %v: the new logarithm %v lies outside the interval from %v to %v", surprise, found, start, end)
			}
		})
	}
}

// Urnings keeps every urn within its size, moves the urn of the topic answered
// only toward the outcome, moves nothing its structure holds still, and draws
// the same moves from the same seeds.
func TestUrningsFollowTheOutcomeWithinTheirUrns(t *testing.T) {
	t.Parallel()
	for _, shape := range []structure{general, topics} {
		t.Run(string(shape), func(t *testing.T) {
			t.Parallel()
			first := urnTrail(t, shape)
			if again := urnTrail(t, shape); !slices.Equal(first, again) {
				t.Error("two runs from the same seeds moved the urns differently")
			}
		})
	}
}

// urnTrail answers Urnings 2,000 times in two topics — all correctly first,
// then all wrongly, then at random, so that its urns fill and empty — and
// gives the count the urn of each topic answered held after the answer. It
// fails the test at an answer that takes an urn past its size or moves it
// against the outcome, and when no urn was ever full or ever empty.
func urnTrail(t *testing.T, shape structure) []int {
	t.Helper()
	u := newUrnings(shape, baselineStart, seeded("urnings/"+string(shape), "urn"))
	if got := u.level(baselineTopics[0]); got != baselineStart {
		t.Fatalf("a fresh urn says %v, want the start %v", got, baselineStart)
	}
	rng := seeded("urnings/"+string(shape), "answers")
	trail := make([]int, 0, 2000)
	full, empty := false, false
	for k := range 2000 {
		topic := answeredTopics[rng.IntN(len(answeredTopics))]
		correct := k < 500 || k >= 1000 && rng.IntN(2) == 0
		before := u.urn(topic)
		answerWithin(t, u, shape, topic, baselineStart+rng.NormFloat64(), correct)
		after := u.urn(topic)
		switch {
		case after < 0 || after > urnSize:
			t.Fatalf("answer %d left the urn of %s at %d, past its size %d", k+1, topic, after, urnSize)
		case correct && after < before || !correct && after > before:
			t.Fatalf("answer %d, correct %v, moved the urn of %s from %d to %d", k+1, correct, topic, before, after)
		}
		full, empty = full || after == urnSize, empty || after == 0
		trail = append(trail, after)
	}
	if !full || !empty {
		t.Errorf("the urns were full %v and empty %v, want both, so that their bounds were met", full, empty)
	}
	return trail
}
