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

// follower is an estimator that starts after the trial series. It takes each
// of the series' answers in without a step of its own and stands where the
// series' estimate stands, so that until its own first step it chooses and
// predicts from the estimate the service has; what it keeps of the answers is
// its own.
type follower interface {
	followed(topic string, beta float64, correct bool, estimate float64)
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

// oracle stands where the child truly stands, in every topic, at every moment,
// and knows the child's true chance: the ceiling of every rule that learns of
// a child from their answers. It learns nothing from an answer, having nothing
// to learn.
type oracle struct {
	c *child
}

func (o oracle) overall() float64                          { return o.c.theta }
func (o oracle) level(topic string) float64                { return o.c.level(topic) }
func (o oracle) chance(topic string, beta float64) float64 { return o.c.chance(topic, beta) }
func (o oracle) answered(string, float64, bool)            {}

// stepRule is an update in the style of Elo in the service's structure, an
// overall level and a topic offset: each moves by its own step times the
// surprise of the answer. The service's step is one such rule; so are a
// constant step, a floor under the overall level's step, and a step worked
// out from an uncertainty counted by answers, which never dies away. A limit
// on how far one answer may move the level in a topic holds a step that is
// large at first back from throwing a child about.
type stepRule struct {
	overallStep, topicStep stepCurve
	floor                  float64 // the least the overall level's step may shrink to
	limit                  float64 // the most one answer may move the level in a topic
	theta                  float64
	delta                  map[string]float64
	answers                int
	inTopic                map[string]int
}

// The service's own constants, which a variant departs from. The service's
// overall step has kept a floor since its step was chosen; the rules here,
// put forward against the step before it, start without one, and a variant
// that keeps one says so.
const (
	serviceK0Theta = 0.2
	serviceK0Delta = 0.4
	serviceDecay   = 0.05
	serviceFloor   = 0.05
)

// The service's steps as curves, as they stood when the choice of the step
// was made — before the floor the service then took, which is the step these
// curves under floorRule(serviceFloor) come to.
var (
	serviceOverallStep = stepCurve{first: serviceK0Theta, decay: serviceDecay}
	serviceTopicStep   = stepCurve{first: serviceK0Delta, decay: serviceDecay}
)

func newStepRule(start float64) *stepRule {
	return &stepRule{
		overallStep: serviceOverallStep, topicStep: serviceTopicStep, limit: noLimit,
		theta: start, delta: map[string]float64{}, inTopic: map[string]int{},
	}
}

func (s *stepRule) overall() float64 { return s.theta }

func (s *stepRule) level(topic string) float64 { return s.theta + s.delta[topic] }

// chance is the service's chance of a right answer at the rule's level.
func (s *stepRule) chance(topic string, beta float64) float64 {
	return rating.Guess + (1-rating.Guess)*logistic(s.level(topic)-beta)
}

// answered moves the overall level and the topic's offset, each by its step,
// both held back together where they would move the level in the topic past
// the limit.
func (s *stepRule) answered(topic string, beta float64, correct bool) {
	surprise := surpriseOf(correct, s.chance(topic, beta))
	kTheta, kDelta := limited(max(s.overallStep.at(s.answers), s.floor), s.topicStep.at(s.inTopic[topic]), surprise, s.limit)
	s.theta += kTheta * surprise
	s.delta[topic] += kDelta * surprise
	s.answers++
	s.inTopic[topic]++
}

// followed takes an answer of the trial series in without moving a level:
// the series' own estimate stands, and the counts go on as the service's do.
func (s *stepRule) followed(topic string, _ float64, _ bool, estimate float64) {
	s.answers++
	s.inTopic[topic]++
	s.theta = estimate
}

// surpriseOf is how far an answer fell from the chance it was given: the score
// less the chance.
func surpriseOf(correct bool, chance float64) float64 {
	score := 0.0
	if correct {
		score = 1
	}
	return score - chance
}
