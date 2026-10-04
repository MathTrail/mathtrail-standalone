package main

// cell is one rule run on one generator's children.
type cell struct {
	rule      *rule
	generator generator
}

// name is how a cell is written in the results.
func (c *cell) name() string {
	return c.rule.name + "/" + string(c.rule.shape) + "/" + string(c.generator)
}

// allGenerators are the generators every rule is run on: the nine the bench
// was carried over with, then those that move how far the model misses, how
// fast and how steadily a child learns, which way a child changes at once,
// and how widely children and their topics spread.
var allGenerators = []generator{
	staticChildren, misplaced, learning, jumping, harderHost, linkedTopics, otherSlope, otherFloor, hintsUsed,
	exactlyAsWritten, smallMisses, largeMisses, learningHalf, learningFades, dropping,
	narrowStart, wideStart, closeTopics, farTopics,
}

// steps makes a step rule, tuned by a variant.
func steps(tune func(s *stepRule)) func(c *child, start float64) estimator {
	return func(_ *child, start float64) estimator {
		s := newStepRule(start)
		if tune != nil {
			tune(s)
		}
		return s
	}
}

// rules are the service's own path — the shrinking step in the structure of
// both — and the rules it is compared with: the step kept constant, at the
// service's first step and at half of it; the shrinking step with a floor
// under the overall level's; the service's step with no trial series; and
// Glicko-2 that knows a child can guess, with one level and with a level per
// topic — and, last, the oracle, which every rule is measured against as its
// ceiling. A rule's name and structure name its cells, and so every seed its
// numbers draw.
func rules() []*rule {
	glicko := func(shape structure) *rule {
		return &rule{name: "glicko2_floor", shape: shape, make: func(_ *child, start float64) estimator { return newGlicko(shape, start) }}
	}
	return []*rule{
		{name: "shrinking", shape: both, service: true},
		{name: "constant", shape: both, trial: true, make: steps(func(s *stepRule) { s.decay = 0 })},
		{name: "constant_slow", shape: both, trial: true, make: steps(func(s *stepRule) { s.decay, s.k0Theta, s.k0Delta = 0, 0.1, 0.2 })},
		{name: "floor_0.05", shape: both, trial: true, make: steps(func(s *stepRule) { s.floor = 0.05 })},
		{name: "no_trial", shape: both, make: steps(nil)},
		glicko(general),
		glicko(topics),
		{name: "oracle", shape: both, ceiling: true, make: func(c *child, _ float64) estimator { return oracle{c: c} }},
	}
}

// cells are every cell of a run: every rule on every generator.
func cells() []cell {
	var all []cell
	for _, r := range rules() {
		for _, g := range allGenerators {
			all = append(all, cell{rule: r, generator: g})
		}
	}
	return all
}
