package main

import (
	"math"
	"math/rand/v2"
)

// The learner's urn: the size Math Garden gave learners in Bolsinova et al.'s
// analysis, started half full.
const (
	urnSize  = 20
	urnStart = 10
)

// urnings is Urnings for a learner against tasks of known difficulty, one urn
// for the child or one per topic. A task is an urn of fixed proportion
// σ(β − start) that is never updated. After an answer X, a simulated outcome
// X* is drawn from the two urns' game, and the proposal r + X − X* is accepted
// with probability min(1, q(r)/q(r̃)), q(r) = r(1 − π) + (n − r)π: for a
// learner who answers with the urn game's own chance, that makes
// Binomial(n, σ(θ − start)) the urn's stationary law, the same for every task.
// The correction for adaptive selection is left out: the service selects
// deterministically, and the correction would reject almost every move.
type urnings struct {
	shape structure
	start float64
	one   int
	urns  map[string]int
	rng   *rand.Rand
}

func newUrnings(shape structure, start float64, rng *rand.Rand) *urnings {
	return &urnings{shape: shape, start: start, one: urnStart, urns: map[string]int{}, rng: rng}
}

func (u *urnings) urn(topic string) int {
	if u.shape == general {
		return u.one
	}
	if r, found := u.urns[topic]; found {
		return r
	}
	return urnStart
}

// urnLevel is the level an urn says, on the scale of the start.
func urnLevel(start float64, r, size int) float64 {
	return start + math.Log(float64(r+1)/float64(size-r+1))
}

func (u *urnings) overall() float64 {
	if u.shape == general {
		return urnLevel(u.start, u.one, urnSize)
	}
	return u.start
}

func (u *urnings) level(topic string) float64 { return urnLevel(u.start, u.urn(topic), urnSize) }

func (u *urnings) chance(topic string, beta float64) float64 { return logistic(u.level(topic) - beta) }

func (u *urnings) answered(topic string, beta float64, correct bool) {
	r := fixedItemStep(u.urn(topic), urnSize, logistic(beta-u.start), correct, u.rng)
	if u.shape == general {
		u.one = r
		return
	}
	u.urns[topic] = r
}

// fixedItemStep is one update of a learner's urn of size n holding r green
// balls after an answer to an item of fixed proportion pi.
func fixedItemStep(r, n int, pi float64, correct bool, rng *rand.Rand) int {
	q := func(r int) float64 { return float64(r)*(1-pi) + float64(n-r)*pi }
	simulated := rng.Float64() < float64(r)*(1-pi)/q(r)
	proposal := r
	if correct {
		proposal++
	}
	if simulated {
		proposal--
	}
	if proposal == r || rng.Float64() >= min(1, q(r)/q(proposal)) {
		return r
	}
	return proposal
}

// tournament is the simulated example of Bolsinova et al. at a smaller scale:
// players whose logits are equally spaced quantiles of the standard normal,
// urns of a size starting half full, games matched with probability
// proportional to exp(−2(ℓ_i − ℓ_j)²) on the smoothed logits, and the full
// update with its Metropolis–Hastings correction for the matching.
type tournament struct {
	size    int
	truth   []float64
	urns    []int
	weights [][]float64
	total   float64
	rng     *rand.Rand
}

func newTournament(players, size int, rng *rand.Rand) *tournament {
	t := &tournament{size: size, truth: make([]float64, players), urns: make([]int, players), rng: rng}
	for i := range players {
		t.truth[i] = normalQuantile((float64(i) + 0.5) / float64(players))
		t.urns[i] = size / 2
	}
	t.weights = make([][]float64, players)
	for i := range players {
		t.weights[i] = make([]float64, players)
	}
	for i := range players {
		for j := i + 1; j < players; j++ {
			t.setWeight(i, j)
			t.total += t.weights[i][j]
		}
	}
	return t
}

// smoothed is the logit an urn says, smoothed so that an empty or a full urn
// says something finite.
func (t *tournament) smoothed(r int) float64 {
	return math.Log(float64(r+1) / float64(t.size-r+1))
}

func (t *tournament) weightOf(ri, rj int) float64 {
	d := t.smoothed(ri) - t.smoothed(rj)
	return math.Exp(-2 * d * d)
}

func (t *tournament) setWeight(i, j int) {
	w := t.weightOf(t.urns[i], t.urns[j])
	t.weights[i][j], t.weights[j][i] = w, w
}

// pick draws a pair with probability proportional to its weight, by
// rejection: a uniform pair is kept with probability equal to its weight,
// which is at most one.
func (t *tournament) pick() (i, j int) {
	players := len(t.urns)
	for {
		i, j = t.rng.IntN(players), t.rng.IntN(players)
		if i != j && t.rng.Float64() < t.weights[i][j] {
			return i, j
		}
	}
}

// totalAfter is the sum of the weights of every pair if players i and j held
// these urns instead.
func (t *tournament) totalAfter(i, j, ri, rj int) float64 {
	total := t.total
	for k := range t.urns {
		if k == i || k == j {
			continue
		}
		total += t.weightOf(ri, t.urns[k]) - t.weights[i][k]
		total += t.weightOf(rj, t.urns[k]) - t.weights[j][k]
	}
	return total + t.weightOf(ri, rj) - t.weights[i][j]
}

// game plays one game and updates the two urns by the full algorithm.
func (t *tournament) game() {
	i, j := t.pick()
	ri, rj, n := t.urns[i], t.urns[j], t.size
	observed := t.rng.Float64() < logistic(t.truth[i]-t.truth[j])
	// Two empty urns, or two full ones, never draw balls of different colours,
	// and their game never ends; such a pair plays no game.
	ends := ri*(n-rj) + (n-ri)*rj
	if ends == 0 {
		return
	}
	simulated := t.rng.Float64() < float64(ri*(n-rj))/float64(ends)
	step := 0
	if observed {
		step++
	}
	if simulated {
		step--
	}
	if step == 0 {
		return
	}
	pi, pj := ri+step, rj-step
	if pi < 0 || pi > n || pj < 0 || pj > n {
		return
	}
	totalAfter := t.totalAfter(i, j, pi, pj)
	ratio := float64(ri*(n-rj)+(n-ri)*rj) / float64(pi*(n-pj)+pj*(n-pi))
	ratio *= (t.weightOf(pi, pj) / totalAfter) / (t.weights[i][j] / t.total)
	if t.rng.Float64() >= min(1, ratio) {
		return
	}
	t.urns[i], t.urns[j], t.total = pi, pj, totalAfter
	for k := range t.urns {
		if k != i {
			t.setWeight(i, k)
		}
		if k != j {
			t.setWeight(j, k)
		}
	}
}

// normalQuantile is the standard normal's quantile, by bisection on its
// distribution function.
func normalQuantile(p float64) float64 {
	low, high := -10.0, 10.0
	for range 200 {
		middle := (low + high) / 2
		if 0.5*math.Erfc(-middle/math.Sqrt2) < p {
			low = middle
		} else {
			high = middle
		}
	}
	return (low + high) / 2
}
