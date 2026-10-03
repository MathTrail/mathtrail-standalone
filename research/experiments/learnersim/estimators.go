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

// stepRule is an update in the style of Elo with the service's step and its
// variants: a constant step, a step weighted toward maximum likelihood, a
// floor under the overall level's step, other first steps, and another slope
// or floor in the model the step is computed against.
type stepRule struct {
	shape   structure
	k0Theta float64
	k0Delta float64
	decay   float64
	floor   float64 // the least the overall level's step may shrink to
	weigh   bool    // whether the step is weighed by w(P), toward maximum likelihood
	guess   float64 // the model's floor
	slope   float64 // the model's slope
	start   float64
	theta   float64
	// base is the offset of a topic the rule has not moved yet: zero, or the
	// trial series' estimate where only the topics have levels.
	base    float64
	delta   map[string]float64
	answers int
	inTopic map[string]int
}

// The service's own constants, which a variant departs from one at a time.
const (
	serviceK0Theta = 0.2
	serviceK0Delta = 0.4
	serviceDecay   = 0.05
	serviceGuess   = rating.Guess
)

func newStepRule(shape structure, start float64) *stepRule {
	return &stepRule{
		shape: shape, k0Theta: serviceK0Theta, k0Delta: serviceK0Delta, decay: serviceDecay,
		guess: serviceGuess, slope: 1, start: start, theta: start,
		delta: map[string]float64{}, inTopic: map[string]int{},
	}
}

func (s *stepRule) overall() float64 { return s.theta }

func (s *stepRule) level(topic string) float64 { return s.theta + s.offset(topic) }

// offset is a topic's offset from the overall level.
func (s *stepRule) offset(topic string) float64 {
	if moved, found := s.delta[topic]; found {
		return moved
	}
	return s.base
}

func (s *stepRule) chance(topic string, beta float64) float64 {
	return s.guess + (1-s.guess)*logistic(s.slope*(s.level(topic)-beta))
}

// answered moves the levels the structure lets move by the step of each.
func (s *stepRule) answered(topic string, beta float64, correct bool) {
	p := s.chance(topic, beta)
	score := 0.0
	if correct {
		score = 1
	}
	surprise := score - p
	if s.weigh {
		surprise *= (p - s.guess) / (p * (1 - s.guess))
	}
	kTheta := max(s.k0Theta/(1+s.decay*float64(s.answers)), s.floor)
	kDelta := s.k0Delta / (1 + s.decay*float64(s.inTopic[topic]))
	switch s.shape {
	case general:
		s.theta += kTheta * surprise
	case topics:
		s.delta[topic] = s.offset(topic) + kDelta*surprise
	case both:
		s.theta += kTheta * surprise
		s.delta[topic] = s.offset(topic) + kDelta*surprise
	}
	s.answers++
	s.inTopic[topic]++
}

// counted takes an answer of the trial series in without moving a level: the
// trial's own estimate stands, and the counts go on as the service's do.
func (s *stepRule) counted(topic string) {
	s.answers++
	s.inTopic[topic]++
}

// takeOver starts the rule from the trial series' estimate: the overall level
// it set, or, where only the topics have levels, every topic at that level and
// the overall level back at the start.
func (s *stepRule) takeOver(estimate float64) {
	if s.shape == topics {
		s.theta, s.base = s.start, estimate-s.start
		return
	}
	s.theta = estimate
}
