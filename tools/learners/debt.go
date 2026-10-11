package main

import "slices"

// owing is what wrong answers still owe a step rule's levels, and how they pay
// it. A wrong answer takes at once the share of its charge that its place in
// its topic's run of wrong answers names — the last share for every place past
// them — and owes the rest: the overall level its part, the topic's offset its
// own. A right answer gives repayment of its gain to what is owed, each unit
// of it clearing burn units, and the rest raises the level. With burn at one
// every unit owed is charged in the end, so the level less what it owes moves
// as the step alone would; a burn of two clears what is owed with half of it
// charged, and forgives the rest.
type owing struct {
	shares    []float64
	repayment float64
	burn      float64
	theta     float64
	delta     map[string]float64
	run       map[string]int
}

func newOwing(shares []float64, repayment, burn float64) *owing {
	return &owing{shares: slices.Clone(shares), repayment: repayment, burn: burn, delta: map[string]float64{}, run: map[string]int{}}
}

// charged is how far an answer moves the overall level and the topic's offset,
// given how far the step alone would move them. A right answer ends the
// topic's run of wrong answers; a wrong one lengthens it.
func (o *owing) charged(topic string, theta, delta float64, correct bool) (movedTheta, movedDelta float64) {
	if correct {
		o.run[topic] = 0
		payTheta, payDelta := o.paid(o.theta, theta), o.paid(o.delta[topic], delta)
		o.theta = max(o.theta-o.burn*payTheta, 0)
		o.delta[topic] = max(o.delta[topic]-o.burn*payDelta, 0)
		return theta - payTheta, delta - payDelta
	}
	o.run[topic]++
	share := o.shareAt(o.run[topic])
	o.theta -= (1 - share) * theta
	o.delta[topic] -= (1 - share) * delta
	return share * theta, share * delta
}

// paid is what a right answer's gain gives to what a level owes: its share of
// the gain, no more than what the payment still has to clear.
func (o *owing) paid(owed, gain float64) float64 {
	return min(owed/o.burn, o.repayment*gain)
}

// shareAt is the share of a wrong answer's charge taken at once at its place in
// the run, counted from one.
func (o *owing) shareAt(place int) float64 {
	return o.shares[min(max(place, 1), len(o.shares))-1]
}

// counted takes an answer of the trial series into the topic's run without
// charging anything: the series moves the level its own way, and the service
// counts the run through it all the same.
func (o *owing) counted(topic string, correct bool) {
	if correct {
		o.run[topic] = 0
		return
	}
	o.run[topic]++
}

// owed is what a topic's level owes: the overall level's debt and the topic's
// own.
func (o *owing) owed(topic string) float64 { return o.theta + o.delta[topic] }
