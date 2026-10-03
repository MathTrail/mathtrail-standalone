package main

import (
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// estimator says where a child stands from their answers: the rule under test.
type estimator interface {
	// overall is its level for the child across topics.
	overall() float64
	// level is its level for the child in a topic.
	level(topic string) float64
	// chance is its own predicted chance of a correct answer in a topic, at the
	// difficulty a task was asked at.
	chance(topic string, beta float64) float64
	// answered takes one answer in.
	answered(topic string, beta float64, correct bool)
}

// structure is how an estimator splits a child's level.
type structure string

// The three structures.
const (
	// general is one level per child, the same in every topic.
	general structure = "general"
	// topics is one level per topic, the overall level held at the start.
	topics structure = "topics"
	// both is the overall level plus a topic offset, as the service keeps it.
	both structure = "both"
)

// service is the product's own estimate: the profile the service keeps, which
// its own code updates after every answer.
type service struct {
	p *profile.Profile
}

func (s service) overall() float64               { return s.p.Ratings.Theta }
func (s service) level(topic string) float64     { return s.p.LevelIn(topic) }
func (s service) answered(string, float64, bool) {}
func (s service) chance(topic string, beta float64) float64 {
	return rating.Probability(s.level(topic), beta)
}

// stepRule is an update in the style of Elo in the service's structure, an
// overall level and a topic offset, with the service's step and its variants:
// a constant step, a floor under the overall level's step, and other first
// steps.
type stepRule struct {
	k0Theta float64
	k0Delta float64
	decay   float64
	floor   float64 // the least the overall level's step may shrink to
	theta   float64
	delta   map[string]float64
	answers int
	inTopic map[string]int
}

// The service's own constants, which a variant departs from one at a time.
const (
	serviceK0Theta = 0.2
	serviceK0Delta = 0.4
	serviceDecay   = 0.05
)

func newStepRule(start float64) *stepRule {
	return &stepRule{
		k0Theta: serviceK0Theta, k0Delta: serviceK0Delta, decay: serviceDecay, theta: start,
		delta: map[string]float64{}, inTopic: map[string]int{},
	}
}

func (s *stepRule) overall() float64 { return s.theta }

func (s *stepRule) level(topic string) float64 { return s.theta + s.delta[topic] }

// chance is the service's chance of a right answer at the rule's level.
func (s *stepRule) chance(topic string, beta float64) float64 {
	return rating.Guess + (1-rating.Guess)*logistic(s.level(topic)-beta)
}

// answered moves the overall level and the topic's offset, each by its step.
func (s *stepRule) answered(topic string, beta float64, correct bool) {
	p := s.chance(topic, beta)
	score := 0.0
	if correct {
		score = 1
	}
	surprise := score - p
	kTheta := max(s.k0Theta/(1+s.decay*float64(s.answers)), s.floor)
	kDelta := s.k0Delta / (1 + s.decay*float64(s.inTopic[topic]))
	s.theta += kTheta * surprise
	s.delta[topic] += kDelta * surprise
	s.answers++
	s.inTopic[topic]++
}

// counted takes an answer of the trial series in without moving a level: the
// trial's own estimate stands, and the counts go on as the service's do.
func (s *stepRule) counted(topic string) {
	s.answers++
	s.inTopic[topic]++
}

// takeOver starts the rule from the overall level the trial series' estimate
// set.
func (s *stepRule) takeOver(estimate float64) { s.theta = estimate }
