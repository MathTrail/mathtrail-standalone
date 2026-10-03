package main

import "math/rand/v2"

// The run of qualifying answers that earns mastery: three in a row on
// eligible tasks.
const streakNeeded = 3

// chainChance is the chance that the run of qualifying answers reaches three
// within m eligible attempts, for a child whose true chance on an eligible task
// is p and who uses the hint on a share h of them: a correct answer without
// the hint moves the run up, one with the hint leaves it, a wrong one returns
// it to nothing.
func chainChance(p, h float64, m int) float64 {
	var state [streakNeeded + 1]float64
	state[0] = 1
	up, stay, reset := p*(1-h), p*h, 1-p
	for range m {
		var next [streakNeeded + 1]float64
		next[streakNeeded] = state[streakNeeded]
		for s := range streakNeeded {
			next[s+1] += state[s] * up
			next[s] += state[s] * stay
			next[0] += state[s] * reset
		}
		state = next
	}
	return state[streakNeeded]
}

// simulatedChain is the same chance found by running the chain itself.
func simulatedChain(p, h float64, m, runs int, rng *rand.Rand) float64 {
	reached := 0
	for range runs {
		run := 0
		for range m {
			u := rng.Float64()
			switch {
			case u < 1-p:
				run = 0
			case u < 1-p+p*h:
			default:
				run++
			}
			if run == streakNeeded {
				reached++
				break
			}
		}
	}
	return float64(reached) / float64(runs)
}

// chainRow is one point of the computed chain.
type chainRow struct {
	p, h   float64
	m      int
	chance float64
}

// chainGrid is the chain over its grid: p from 0.05 to 0.95 in steps of 0.05,
// a share of hints of 0, 0.1 or 0.2, and the lengths of chainLengths.
func chainGrid() []chainRow {
	var rows []chainRow
	for step := 1; step <= 19; step++ {
		p := float64(step) / 20
		for _, h := range []float64{0, 0.1, 0.2} {
			for _, m := range chainLengths {
				rows = append(rows, chainRow{p: p, h: h, m: m, chance: chainChance(p, h, m)})
			}
		}
	}
	return rows
}
