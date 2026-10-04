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

// cellsOf are every cell of a run of these rules: every rule on every
// generator.
func cellsOf(rules []*rule) []cell {
	var all []cell
	for _, r := range rules {
		for _, g := range allGenerators {
			all = append(all, cell{rule: r, generator: g})
		}
	}
	return all
}
