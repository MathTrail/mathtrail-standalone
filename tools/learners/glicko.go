package main

import "math"

// Glicko-2 as Glickman describes it (2022): a rating, its deviation and its
// volatility, on the scale where a difference of one is a factor of e in the
// odds.
const (
	glickoScale      = 173.7178
	glickoDeviation  = 350 / glickoScale
	glickoVolatility = 0.06
	glickoTau        = 0.5
	glickoTolerance  = 0.000001
)

// glickoRating is one rating of Glicko-2.
type glickoRating struct {
	mu, phi, sigma float64
}

// term is one game's share of a rating period: its information about the
// rating, and the gradient of its log-likelihood.
type term struct {
	information, gradient float64
}

// logisticTerm is a game against an opponent of a rating and a deviation, as
// Glickman has it: his g, his E, and the score.
func logisticTerm(mu, opponent, deviation, score float64) term {
	g := 1 / math.Sqrt(1+3*deviation*deviation/(math.Pi*math.Pi))
	e := logistic(g * (mu - opponent))
	return term{information: g * g * e * (1 - e), gradient: g * (score - e)}
}

// flooredTerm is a task of known difficulty when a guess the child can always
// make lifts the expected score to floor + (1 − floor)·λ: the information and
// gradient of the Bernoulli likelihood of that score, which are Glickman's own
// when the floor is zero.
func flooredTerm(floor, mu, beta, score float64) term {
	l := logistic(mu - beta)
	e := floor + (1-floor)*l
	slope := (1 - floor) * l * (1 - l)
	return term{information: slope * slope / (e * (1 - e)), gradient: slope * (score - e) / (e * (1 - e))}
}

// period is one rating period, Glickman's steps 3 to 7.
func (r glickoRating) period(terms []term) glickoRating {
	information, gradient := 0.0, 0.0
	for _, t := range terms {
		information += t.information
		gradient += t.gradient
	}
	v := 1 / information
	sigma := newVolatility(r.phi, r.sigma, v, v*gradient)
	phiStar := math.Sqrt(r.phi*r.phi + sigma*sigma)
	phi := 1 / math.Sqrt(1/(phiStar*phiStar)+1/v)
	return glickoRating{mu: r.mu + phi*phi*gradient, phi: phi, sigma: sigma}
}

// newVolatility is Glickman's step 5, by the Illinois algorithm he gives.
func newVolatility(phi, sigma, v, delta float64) float64 {
	a := math.Log(sigma * sigma)
	f := func(x float64) float64 {
		ex := math.Exp(x)
		return ex*(delta*delta-phi*phi-v-ex)/(2*(phi*phi+v+ex)*(phi*phi+v+ex)) - (x-a)/(glickoTau*glickoTau)
	}
	upper, lower := a, bracket(a, phi, v, delta, f)
	fUpper, fLower := f(upper), f(lower)
	for math.Abs(lower-upper) > glickoTolerance {
		c := upper + (upper-lower)*fUpper/(fLower-fUpper)
		fc := f(c)
		if fc*fLower <= 0 {
			upper, fUpper = lower, fLower
		} else {
			fUpper /= 2
		}
		lower, fLower = c, fc
	}
	return math.Exp(upper / 2)
}

// bracket is the second end of the interval the new volatility's logarithm
// lies in, as step 5.2 sets it.
func bracket(a, phi, v, delta float64, f func(float64) float64) float64 {
	if delta*delta > phi*phi+v {
		return math.Log(delta*delta - phi*phi - v)
	}
	k := 1.0
	for f(a-k*glickoTau) < 0 {
		k++
	}
	return a - k*glickoTau
}

// glicko is the estimator of Glicko-2: one rating for the child, or one per
// topic, each starting at the child's start. Every answer is a rating period
// of its own, against a task of known difficulty, whose deviation is zero.
type glicko struct {
	shape   structure
	start   float64
	floor   float64
	one     glickoRating
	ratings map[string]glickoRating
}

func newGlicko(shape structure, start, floor float64) *glicko {
	return &glicko{shape: shape, start: start, floor: floor, one: freshGlicko(start), ratings: map[string]glickoRating{}}
}

func freshGlicko(start float64) glickoRating {
	return glickoRating{mu: start, phi: glickoDeviation, sigma: glickoVolatility}
}

func (g *glicko) rating(topic string) glickoRating {
	if g.shape == general {
		return g.one
	}
	if r, found := g.ratings[topic]; found {
		return r
	}
	return freshGlicko(g.start)
}

func (g *glicko) overall() float64 {
	if g.shape == general {
		return g.one.mu
	}
	return g.start
}

func (g *glicko) level(topic string) float64 { return g.rating(topic).mu }

func (g *glicko) chance(topic string, beta float64) float64 {
	return g.floor + (1-g.floor)*logistic(g.level(topic)-beta)
}

func (g *glicko) answered(topic string, beta float64, correct bool) {
	score := 0.0
	if correct {
		score = 1
	}
	r := g.rating(topic)
	t := logisticTerm(r.mu, beta, 0, score)
	if g.floor > 0 {
		t = flooredTerm(g.floor, r.mu, beta, score)
	}
	updated := r.period([]term{t})
	if g.shape == general {
		g.one = updated
		return
	}
	g.ratings[topic] = updated
}
