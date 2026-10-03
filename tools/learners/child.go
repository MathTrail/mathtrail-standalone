package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand/v2"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The seed and the name of the run whose numbers the paper about the service
// reports: a run that is given neither repeats them.
const (
	paperSeed       = 20261001
	paperExperiment = "E-A3"
)

// masterSeed is where every random draw of a run comes from, and experiment
// names the run in every seed it draws. The command line may set either before
// the run starts; nothing changes them during it.
var (
	masterSeed uint64 = paperSeed
	experiment        = paperExperiment
)

// seeded is the random stream of one unit of the experiment for one purpose.
// The draws stand for children's answers and a model's writing, and need to be
// repeatable rather than unpredictable.
func seeded(unit, purpose string) *rand.Rand {
	hash := fnv.New64a()
	fmt.Fprintf(hash, "%d|%s|%s|%s", masterSeed, experiment, unit, purpose)
	return rand.New(rand.NewPCG(hash.Sum64(), masterSeed)) //nolint:gosec // a seeded, repeatable simulation, not a secret
}

// The base population: how far a child stands from the start their grade
// gives, how far their topics stand from their overall level, how far a written
// task's difficulty misses the one asked for, and how children answer.
const (
	spreadAroundStart = 1.0
	spreadOfTopics    = 0.5
	writingError      = 0.5
	baseSlope         = 1.0
	baseFloor         = 0.2
)

// The generators' departures from the base population.
const (
	misplacedBy       = 2.5
	learningPerAnswer = 0.01
	topicLearning     = 0.02
	forgetting        = 0.01
	jumpBy            = 1.0
	jumpEarliest      = 50
	jumpLatest        = 150
	harderHostBias    = 0.75
	groupOffset       = 0.75
	spreadInGroup     = 0.25
	hintShare         = 0.2
)

// generator is a population of simulated children: the base population, with
// one way children depart from the model.
type generator string

// The generators, each changing the base population in one respect.
const (
	staticChildren   generator = "G0"
	misplaced        generator = "G1"
	learning         generator = "G2"
	jumping          generator = "G3"
	harderHost       generator = "G4"
	linkedTopics     generator = "G5"
	otherSlope       generator = "G6"
	otherFloor       generator = "G7"
	hintsUsed        generator = "G8"
	exactlyAsWritten generator = "G0-exact"
)

// child is one simulated child: where they truly stand, how they answer, and
// how they change.
type child struct {
	id       string
	grade    int
	start    float64
	theta    float64
	delta    map[string]float64
	first    map[string]float64
	slope    float64
	floor    float64
	hint     float64
	writing  float64
	jumpAt   int
	learns   bool
	twoHosts bool
	draws    *rand.Rand
}

// newChild draws child number i of a generator over the catalog's topics. The
// same child is drawn for every rule it is run under: its parameters and its
// answers come from streams of its own.
func newChild(gen generator, i int, topics []string) *child {
	id := fmt.Sprintf("%s/%d", gen, i)
	params := seeded(id, "params")
	c := &child{
		id:      id,
		grade:   1 + params.IntN(6),
		delta:   make(map[string]float64, len(topics)),
		first:   make(map[string]float64, len(topics)),
		slope:   baseSlope,
		floor:   baseFloor,
		writing: writingError,
		jumpAt:  -1,
		draws:   seeded(id, "answers"),
	}
	c.start = rating.Start(c.grade)
	c.theta = c.start + spreadAroundStart*params.NormFloat64()
	sign := 1.0
	if params.IntN(2) == 0 {
		sign = -1
	}
	for place, topic := range topics {
		if gen == linkedTopics {
			c.delta[topic] = groupSide(place, len(topics))*sign*groupOffset + spreadInGroup*params.NormFloat64()
		} else {
			c.delta[topic] = spreadOfTopics * params.NormFloat64()
		}
		c.first[topic] = c.delta[topic]
	}
	c.depart(gen, i, params)
	return c
}

// groupSide is +1 for the first ⌈T/2⌉ topics of the catalog's order and −1
// for the rest.
func groupSide(place, topics int) float64 {
	if place < (topics+1)/2 {
		return 1
	}
	return -1
}

// depart applies the one respect in which a generator's children differ from
// the base population. Half the children of a two-sided generator go each way.
func (c *child) depart(gen generator, i int, params *rand.Rand) {
	lowerHalf := i%2 == 0
	switch gen {
	case misplaced:
		c.theta = c.start + misplacedBy
		if lowerHalf {
			c.theta = c.start - misplacedBy
		}
	case learning:
		c.learns = true
	case jumping:
		c.jumpAt = jumpEarliest + params.IntN(jumpLatest-jumpEarliest+1)
	case harderHost:
		c.twoHosts = true
	case otherSlope:
		c.slope = 2
		if lowerHalf {
			c.slope = 0.5
		}
	case otherFloor:
		c.floor = 0.3
		if lowerHalf {
			c.floor = 0
		}
	case hintsUsed:
		c.hint = hintShare
	case exactlyAsWritten:
		c.writing = 0
	}
}

// answerDraws are what one answer takes from the child's stream, the same
// draws in the same order whatever happens, so that the k-th answer meets the
// same draws under every rule.
type answerDraws struct {
	writing float64 // a standard normal, for the written task's error
	correct float64 // uniform, compared with the chance of a correct answer
	wrong   int     // which of the four wrong letters a wrong answer picks
	hint    float64 // uniform, compared with the share of answers with the hint
	key     int     // which letter the task's key is
}

func (c *child) next() answerDraws {
	return answerDraws{
		writing: c.draws.NormFloat64(),
		correct: c.draws.Float64(),
		wrong:   c.draws.IntN(4),
		hint:    c.draws.Float64(),
		key:     c.draws.IntN(5),
	}
}

// level is where the child truly stands in a topic.
func (c *child) level(topic string) float64 { return c.theta + c.delta[topic] }

// writtenDifficulty is the difficulty a task asked at beta truly has: the
// model misses it, and in the generator of two hosts the second host writes
// every second task harder.
func (c *child) writtenDifficulty(beta float64, k int, z float64) float64 {
	mean := 0.0
	if c.twoHosts && k%2 == 1 {
		mean = harderHostBias
	}
	return beta + mean + c.writing*z
}

// chance is the child's true chance of answering a task of this true
// difficulty in this topic.
func (c *child) chance(topic string, beta float64) float64 {
	return c.floor + (1-c.floor)*logistic(c.slope*(c.level(topic)-beta))
}

// truthAt is the child's true chance on a task of this level and topic at
// difficulty 3, written exactly as asked: what "truly mastered" is read from.
func (c *child) truthAt(topic string, level rating.GradeLevel) float64 {
	return c.chance(topic, rating.Point{GradeLevel: level, Difficulty: 3}.Beta())
}

// after moves the child on after answer k in a topic: learning and forgetting,
// and a jump at its answer.
func (c *child) after(topic string, k int) {
	if c.learns {
		c.theta += learningPerAnswer
		for other, offset := range c.delta {
			if other == topic {
				c.delta[other] = offset + topicLearning
				continue
			}
			c.delta[other] = offset + forgetting*(c.first[other]-offset)
		}
	}
	if k+1 == c.jumpAt {
		c.theta += jumpBy
	}
}

func logistic(x float64) float64 { return 1 / (1 + math.Exp(-x)) }
