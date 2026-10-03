package main

import (
	"math"
	"math/rand/v2"
)

// glickmanExample is Glickman's worked example (2022): a player at 1500 with
// RD 200 and volatility 0.06 meets players at 1400, 1550 and 1700, with RDs
// 30, 100 and 300, wins the first game and loses the other two, with τ = 0.5.
// It returns the new rating, RD and volatility, which he prints as 1464.06,
// 151.52 and 0.05999.
func glickmanExample() (r, rd, volatility float64) {
	player := glickoRating{mu: 0, phi: 200 / glickoScale, sigma: 0.06}
	opponents := []struct{ rating, rd, score float64 }{{1400, 30, 1}, {1550, 100, 0}, {1700, 300, 0}}
	terms := make([]term, 0, len(opponents))
	for _, o := range opponents {
		terms = append(terms, logisticTerm(player.mu, (o.rating-1500)/glickoScale, o.rd/glickoScale, o.score))
	}
	after := player.period(terms)
	return glickoScale*after.mu + 1500, glickoScale * after.phi, after.sigma
}

// tournamentGap reproduces Urnings' simulated example at a smaller scale and
// says how far, on average over the players, the mean of each player's scaled
// urning after the first tenth of the games lies from the player's true
// proportion.
func tournamentGap(players, size, games int, rng *rand.Rand) float64 {
	t := newTournament(players, size, rng)
	sums := make([]float64, players)
	samples := 0
	const every = 10
	for g := range games {
		t.game()
		if g < games/10 || g%every != 0 {
			continue
		}
		for i, r := range t.urns {
			sums[i] += float64(r) / float64(size)
		}
		samples++
	}
	gap := 0.0
	for i := range players {
		gap += math.Abs(sums[i]/float64(samples) - logistic(t.truth[i]))
	}
	return gap / float64(players)
}

// fixedItemGap runs the urn of a learner of fixed ability against tasks of
// random difficulty — normal around an offset from the start, drawn
// independently of the urn — the learner answering with the urn game's own
// chance. It says the largest gap, over the counts the urn can hold, between
// the share of time spent at a count and the Binomial(n, σ(θ − start)) chance
// of it, and the total variation between the two laws.
func fixedItemGap(ability, offset, spread float64, steps int, rng *rand.Rand) (worst, variation float64) {
	start := 0.0
	r := urnStart
	time := make([]float64, urnSize+1)
	for step := range steps {
		beta := start + offset + spread*rng.NormFloat64()
		correct := rng.Float64() < logistic(ability-beta)
		r = fixedItemStep(r, urnSize, logistic(beta-start), correct, rng)
		if step >= steps/10 {
			time[r]++
		}
	}
	counted := float64(steps - steps/10)
	pi := logistic(ability - start)
	for k := range urnSize + 1 {
		gap := math.Abs(time[k]/counted - binomial(urnSize, k, pi))
		worst = max(worst, gap)
		variation += gap / 2
	}
	return worst, variation
}

func binomial(n, k int, p float64) float64 {
	whole, _ := math.Lgamma(float64(n + 1))
	first, _ := math.Lgamma(float64(k + 1))
	rest, _ := math.Lgamma(float64(n - k + 1))
	return math.Exp(whole - first - rest + float64(k)*math.Log(p) + float64(n-k)*math.Log1p(-p))
}

// fixedItemCheck is the protocol's check of the fixed-item urn: tasks normal
// around the start, the largest gap at any count.
func fixedItemCheck() float64 {
	worst, _ := fixedItemGap(fixedItemAbility, 0, 1, fixedItemSteps, seeded("fixed item", "check"))
	return worst
}

// fixedItemNearCheck is a stricter companion: tasks normal around the
// learner's own level, where the urn game's chance of ending varies most with
// the count, and the total variation between the urn's law and the binomial.
// At the protocol's tolerance the first check passes even with the acceptance
// step left out; this one does not.
func fixedItemNearCheck() float64 {
	_, variation := fixedItemGap(fixedItemAbility, fixedItemAbility, 0.5, fixedItemSteps, seeded("fixed item near", "check"))
	return variation
}
