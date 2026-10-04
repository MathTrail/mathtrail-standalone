package main

import (
	"math"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// How the estimate over the whole history is looked for: at most so many
// rounds of Fisher's scoring, each taking as much of its step as raises the
// estimate's probability, halving it as often as it has to, and done once no
// level moves by more than the last distance worth telling apart.
const (
	historyRounds  = 50
	historyHalving = 30
	historySettled = 1e-12
)

// wholeHistory is the most likely overall level and topic offsets given every
// answer at once: the start's belief about the overall level, as the trial
// series holds it, the spread of topics' belief about every offset, and the
// chance of every answer as it came, weighed together and found anew after
// every answer. Nothing is added for a child who changes, so for a child who
// stays put it is the best any rule of answers could do, and how far a rule
// stands from it is how much that rule leaves on the table. It follows the
// trial series as the others do, and weighs the series' answers with the rest.
//
// The topics take their places in the order they are first answered, so that
// every sum over them comes to the same bits every time.
type wholeHistory struct {
	start         float64
	topicVariance float64
	levels        []float64 // the overall level, then each topic's offset
	places        map[string]int
	answers       []placedAnswer
}

// placedAnswer is an answer as the estimate weighs it: the place of its
// topic among the levels, its task's difficulty, and whether it was solved.
type placedAnswer struct {
	place   int
	beta    float64
	correct bool
}

func newWholeHistory(start, topicSpread float64) *wholeHistory {
	return &wholeHistory{start: start, topicVariance: topicSpread * topicSpread, levels: []float64{start}, places: map[string]int{}}
}

func (h *wholeHistory) overall() float64 { return h.levels[0] }

func (h *wholeHistory) level(topic string) float64 {
	if at, seen := h.places[topic]; seen {
		return h.levels[0] + h.levels[at]
	}
	return h.levels[0]
}

func (h *wholeHistory) chance(topic string, beta float64) float64 {
	return rating.Guess + (1-rating.Guess)*logistic(h.level(topic)-beta)
}

// followed keeps an answer of the trial series and stands where the series'
// estimate stands.
func (h *wholeHistory) followed(topic string, beta float64, correct bool, estimate float64) {
	h.answers = append(h.answers, placedAnswer{place: h.placeOf(topic), beta: beta, correct: correct})
	h.levels[0] = estimate
}

// answered keeps the answer and finds the most likely levels anew, from where
// they stood.
func (h *wholeHistory) answered(topic string, beta float64, correct bool) {
	h.answers = append(h.answers, placedAnswer{place: h.placeOf(topic), beta: beta, correct: correct})
	for range historyRounds {
		step := h.scoringStep()
		if !h.climb(step) {
			return
		}
	}
}

// placeOf is a topic's place among the levels, given it at its first answer.
func (h *wholeHistory) placeOf(topic string) int {
	if at, seen := h.places[topic]; seen {
		return at
	}
	at := len(h.levels)
	h.places[topic] = at
	h.levels = append(h.levels, 0)
	return at
}

// scoringStep is a round of Fisher's scoring: the step that solves the
// expected information against the slope of the estimate's log-probability.
// The information couples the overall level with every topic and no topic
// with another, so the step is found topic by topic once the overall level's
// share is known.
func (h *wholeHistory) scoringStep() []float64 {
	n := len(h.levels)
	slope, information, coupling := make([]float64, n), make([]float64, n), make([]float64, n)
	slope[0] = -(h.levels[0] - h.start) / (trialSpread * trialSpread)
	information[0] = 1 / (trialSpread * trialSpread)
	for t := 1; t < n; t++ {
		slope[t] = -h.levels[t] / h.topicVariance
		information[t] = 1 / h.topicVariance
	}
	for _, a := range h.answers {
		above := h.levels[0] + h.levels[a.place] - a.beta
		pull, told := answerSlope(above, a.correct), informationAt(rating.Probability(above, 0))
		slope[0] += pull
		slope[a.place] += pull
		information[0] += told
		information[a.place] += told
		coupling[a.place] += told
	}
	overallShare, overallSlope := information[0], slope[0]
	for t := 1; t < n; t++ {
		overallShare -= coupling[t] * coupling[t] / information[t]
		overallSlope -= coupling[t] * slope[t] / information[t]
	}
	step := make([]float64, n)
	step[0] = overallSlope / overallShare
	for t := 1; t < n; t++ {
		step[t] = (slope[t] - coupling[t]*step[0]) / information[t]
	}
	return step
}

// climb takes as much of a step as raises the estimate's log-probability,
// halving it until it does, and says whether the levels moved by enough to
// look again.
func (h *wholeHistory) climb(step []float64) bool {
	here := h.logProbability(h.levels)
	there := make([]float64, len(h.levels))
	size := 1.0
	for range historyHalving {
		farthest := 0.0
		for i, level := range h.levels {
			there[i] = level + size*step[i]
			farthest = max(farthest, math.Abs(size*step[i]))
		}
		if h.logProbability(there) >= here {
			copy(h.levels, there)
			return farthest > historySettled
		}
		size /= 2
	}
	return false
}

// logProbability is how probable levels are given the beliefs and the
// answers, as a logarithm and up to a constant.
func (h *wholeHistory) logProbability(levels []float64) float64 {
	fromStart := (levels[0] - h.start) / trialSpread
	sum := -fromStart * fromStart / 2
	for _, offset := range levels[1:] {
		sum -= offset * offset / (2 * h.topicVariance)
	}
	for _, a := range h.answers {
		above := levels[0] + levels[a.place] - a.beta
		if a.correct {
			sum += math.Log(rating.Guess + (1-rating.Guess)*logistic(above))
		} else {
			sum += math.Log(1-rating.Guess) - softplus(above)
		}
	}
	return sum
}

// answerSlope is how fast an answer's log-chance rises with the level, where
// the level stands this far above the task.
func answerSlope(above float64, correct bool) float64 {
	s := logistic(above)
	if correct {
		return (1 - rating.Guess) * s * (1 - s) / (rating.Guess + (1-rating.Guess)*s)
	}
	return -s
}

// softplus is ln(1 + eˣ), worked out so that it neither overflows for a large
// x nor loses a small one: the logarithm of the chance of a wrong answer, less
// that of the wrong options, without forming a chance that underflows.
func softplus(x float64) float64 {
	if x > 0 {
		return x + math.Log1p(math.Exp(-x))
	}
	return math.Log1p(math.Exp(x))
}
