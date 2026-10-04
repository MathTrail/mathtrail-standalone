package main

import "github.com/MathTrail/mathtrail-standalone/internal/domain/rating"

// trialSpread is how far from the start the trial series' estimate lets a
// child stand before any answer: the spread of its belief about the start,
// which the rules that weigh a child's answers all at once take as theirs.
const trialSpread = 2.5

// storedUncertainty is a step from an uncertainty each child carries, as the
// profile would have to store it: one for the overall level and one for each
// topic. An answer narrows them by what it tells at the chance the child had,
// not at the middle of the corridor, and widens them by what is added for every
// answer; its surprise is shared between the overall level and the topic by
// their uncertainties, as a Kalman filter shares it when it takes the two to
// be independent. The overall level's uncertainty starts at the end of the
// trial series, from what the series' answers tell at the level it found on
// top of what the start allowed, the expected information rather than the
// curvature of the series' answers, which a guess can turn the wrong way; a
// topic's starts at the spread of topics. The limit holds a large first step
// back as it does the step from counted answers.
type storedUncertainty struct {
	theta, overallUncertainty float64
	delta, topicUncertainty   map[string]float64
	topicSpread               float64
	addedOverall, addedTopic  float64
	limit                     float64
	series                    []float64 // the difficulties of the trial series' answers, until the first step
	started                   bool
}

func newStoredUncertainty(start, topicSpread, addedOverall, addedTopic, limit float64) *storedUncertainty {
	return &storedUncertainty{
		theta: start, overallUncertainty: trialSpread * trialSpread,
		delta: map[string]float64{}, topicUncertainty: map[string]float64{},
		topicSpread: topicSpread, addedOverall: addedOverall, addedTopic: addedTopic, limit: limit,
	}
}

func (s *storedUncertainty) overall() float64 { return s.theta }

func (s *storedUncertainty) level(topic string) float64 { return s.theta + s.delta[topic] }

func (s *storedUncertainty) chance(topic string, beta float64) float64 {
	return rating.Guess + (1-rating.Guess)*logistic(s.level(topic)-beta)
}

// followed keeps the difficulty of an answer of the trial series and stands
// where the series' estimate stands.
func (s *storedUncertainty) followed(_ string, beta float64, _ bool, estimate float64) {
	s.series = append(s.series, beta)
	s.theta = estimate
}

// answered moves the overall level and the topic by their shares of the
// surprise, then narrows and widens their uncertainties.
func (s *storedUncertainty) answered(topic string, beta float64, correct bool) {
	s.start()
	p := s.chance(topic, beta)
	overallBefore, topicBefore := s.overallUncertainty, s.topicUncertaintyOf(topic)
	gain, information := gainAt(p), informationAt(p)
	together := 1 + information*(overallBefore+topicBefore)
	surprise := surpriseOf(correct, p)
	kTheta, kDelta := limited(overallBefore*gain/together, topicBefore*gain/together, surprise, s.limit)
	s.theta += kTheta * surprise
	s.delta[topic] += kDelta * surprise
	s.overallUncertainty = overallBefore - overallBefore*overallBefore*information/together + s.addedOverall
	s.topicUncertainty[topic] = topicBefore - topicBefore*topicBefore*information/together + s.addedTopic
}

// start sets the overall level's uncertainty, once, from the trial series'
// answers at the level the series found: what the start allowed, narrowed by
// what each answer tells there.
func (s *storedUncertainty) start() {
	if s.started {
		return
	}
	s.started = true
	precision := 1 / s.overallUncertainty
	for _, beta := range s.series {
		precision += informationAt(rating.Probability(s.theta, beta))
	}
	s.overallUncertainty = 1 / precision
	s.series = nil
}

// topicUncertaintyOf is a topic's uncertainty, the spread of topics until its
// first answer.
func (s *storedUncertainty) topicUncertaintyOf(topic string) float64 {
	if v, seen := s.topicUncertainty[topic]; seen {
		return v
	}
	return s.topicSpread * s.topicSpread
}
