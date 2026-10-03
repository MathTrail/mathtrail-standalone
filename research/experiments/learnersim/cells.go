package main

import (
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// cell is one rule run on one generator's children.
type cell struct {
	rule      *rule
	generator generator
}

// name is how a cell is written in the results.
func (c *cell) name() string {
	return c.rule.name + "/" + string(c.rule.shape) + "/" + string(c.generator)
}

// allGenerators are the nine generators of the sweep.
var allGenerators = []generator{
	staticChildren, misplaced, learning, jumping, harderHost, linkedTopics, otherSlope, otherFloor, hintsUsed,
}

// steps makes a step rule in a structure, tuned by a variant.
func steps(shape structure, tune func(s *stepRule)) func(c *child, start float64) estimator {
	return func(_ *child, start float64) estimator {
		s := newStepRule(shape, start)
		if tune != nil {
			tune(s)
		}
		return s
	}
}

// sweepRules are the fifteen rules of the sweep: the three step rules in all
// three structures — the shrinking step in the structure of both being the
// service's own path — and Glicko-2, with and without the floor, and Urnings
// in the general and topic structures, where they have a published form.
func sweepRules() []*rule {
	var rules []*rule
	for _, shape := range []structure{general, topics, both} {
		shrinking := &rule{name: "shrinking", shape: shape, trial: true, make: steps(shape, nil)}
		if shape == both {
			shrinking = &rule{name: "shrinking", shape: both, service: true}
		}
		rules = append(rules,
			shrinking,
			&rule{name: "constant", shape: shape, trial: true, make: steps(shape, func(s *stepRule) { s.decay = 0 })},
			&rule{name: "gradient", shape: shape, trial: true, make: steps(shape, func(s *stepRule) { s.weigh = true })},
		)
	}
	for _, shape := range []structure{general, topics} {
		rules = append(rules,
			&rule{name: "glicko2", shape: shape, make: func(_ *child, start float64) estimator { return newGlicko(shape, start, 0) }},
			&rule{name: "glicko2_floor", shape: shape, make: func(_ *child, start float64) estimator { return newGlicko(shape, start, rating.Guess) }},
			&rule{name: "urnings", shape: shape, make: urningsMaker(shape)},
		)
	}
	return rules
}

// urningsMaker makes Urnings with a stream of its own per child, so that its
// simulated outcomes take nothing from the child's answers.
func urningsMaker(shape structure) func(c *child, start float64) estimator {
	return func(c *child, start float64) estimator {
		return newUrnings(shape, start, seeded(c.id+"/urnings/"+string(shape), "urnings"))
	}
}

// variant is a rule of the service's kind run on some generators only.
type variant struct {
	rule       *rule
	generators []generator
}

// variants are the variants of the service's rule and the bias of the
// corridor, each on the generators it is run on.
func variants() []variant {
	floored := func(name string, least float64) *rule {
		return &rule{name: name, shape: both, trial: true, make: steps(both, func(s *stepRule) { s.floor = least })}
	}
	modelled := func(name string, tune func(s *stepRule)) *rule {
		return &rule{name: name, shape: both, trial: true, target: rating.CorridorMiddle, make: steps(both, tune)}
	}
	return []variant{
		{floored("floor_0.01", 0.01), []generator{staticChildren, jumping}},
		{floored("floor_0.02", 0.02), []generator{staticChildren, jumping}},
		{floored("floor_0.05", 0.05), []generator{staticChildren, jumping}},
		{&rule{name: "k0_half", shape: both, trial: true, make: steps(both, func(s *stepRule) { s.k0Theta, s.k0Delta = 0.1, 0.2 })}, []generator{staticChildren, learning}},
		{&rule{name: "k0_double", shape: both, trial: true, make: steps(both, func(s *stepRule) { s.k0Theta, s.k0Delta = 0.4, 0.8 })}, []generator{staticChildren, learning}},
		{&rule{name: "constant_slow", shape: both, trial: true, make: steps(both, func(s *stepRule) { s.decay, s.k0Theta, s.k0Delta = 0, 0.1, 0.2 })}, []generator{staticChildren, learning}},
		{&rule{name: "no_trial", shape: both, make: steps(both, nil)}, []generator{staticChildren, misplaced}},
		{&rule{name: "target_0.70", shape: both, service: true, target: 0.70}, []generator{staticChildren, learning}},
		{&rule{name: "target_0.85", shape: both, service: true, target: 0.85}, []generator{staticChildren, learning}},
		{modelled("model_slope_0.5", func(s *stepRule) { s.slope = 0.5 }), []generator{staticChildren, otherSlope, otherFloor}},
		{modelled("model_slope_2", func(s *stepRule) { s.slope = 2 }), []generator{staticChildren, otherSlope, otherFloor}},
		{modelled("model_floor_0", func(s *stepRule) { s.guess = 0 }), []generator{staticChildren, otherSlope, otherFloor}},
		{modelled("model_floor_0.3", func(s *stepRule) { s.guess = 0.3 }), []generator{staticChildren, otherSlope, otherFloor}},
		{&rule{name: "bias_shrinking_0.5", shape: both, service: true, target: 0.5}, []generator{staticChildren}},
		{&rule{name: "bias_constant_0.5", shape: both, trial: true, target: 0.5, make: steps(both, func(s *stepRule) { s.decay = 0 })}, []generator{staticChildren}},
	}
}

// cells are every cell of the experiment: the sweep on every generator, then
// the variants on theirs.
func cells() []cell {
	var all []cell
	for _, r := range sweepRules() {
		for _, g := range allGenerators {
			all = append(all, cell{rule: r, generator: g})
		}
	}
	for _, v := range variants() {
		for _, g := range v.generators {
			all = append(all, cell{rule: v.rule, generator: g})
		}
	}
	return all
}
