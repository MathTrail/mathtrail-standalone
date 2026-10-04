package main

import (
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
)

// The interval every number is reported with: a 95 % percentile bootstrap
// over children.
const (
	resamples = 2000
	tail      = 0.025
)

// part is what one child brings to one metric: a value and whether the child
// has one, for a mean over children; a numerator and a denominator, for a
// ratio of sums over them.
type part struct {
	a, b float64
}

// metric reads one number off a cell's children: a mean over the children
// that have a value, or a ratio of sums over them.
type metric struct {
	name string
	// extract is what one child brings to the metric.
	extract func(r *childResult) part
	// applies says whether the metric means anything under a rule; nil when it
	// does under every rule.
	applies func(r *rule) bool
}

// placesOverall says whether a rule keeps an overall level of its own: the
// structure of topics holds it at the start, so its error there would be the
// child's own distance from the start and say nothing of the rule.
func placesOverall(r *rule) bool { return r.shape != topics }

// vector is what one child brings to every metric, in the order of metrics(),
// with its calibration bins, its topics' eligible attempts and the moves of
// the card's rating in each window: all of a child a summary needs, kept
// small.
type vector struct {
	parts    []part
	bins     [calibrationBins]bin
	attempts [longestChain]atAttempt
	moves    [len(screenWindows)][]uint16
}

// atAttempt is, at one eligible attempt, how many of a child's topics made it
// while not yet declared mastered and not truly mastered, and how many of
// those were declared mastered on it.
type atAttempt struct {
	atRisk, declared int32
}

func vectorOf(r *childResult, all []metric) vector {
	v := vector{parts: make([]part, len(all)), bins: r.bins, attempts: attemptsOf(r), moves: r.screen.moves}
	for i, m := range all {
		v.parts[i] = m.extract(r)
	}
	return v
}

// attemptsOf counts a child's topics at every eligible attempt up to the
// longest chain: a topic counts at each attempt it made, up to the one it was
// declared mastered at or the last it made undeclared.
func attemptsOf(r *childResult) [longestChain]atAttempt {
	var at [longestChain]atAttempt
	for _, pair := range r.pairs {
		last, declared := pair.attempts, pair.declaredAt > 0
		if declared {
			last = pair.declaredAt
		}
		for j := range min(last, longestChain) {
			at[j].atRisk++
		}
		if declared && last <= longestChain {
			at[last-1].declared++
		}
	}
	return at
}

// read is the metric at index i over a sample of children, given by their
// indices: the sum of what they bring to it over the sum of their weights,
// which is a mean over the children that have a value or a ratio of sums. It
// says false when the sample holds nothing to read it from.
func read(children []vector, i int, sample []int) (float64, bool) {
	sumA, sumB := 0.0, 0.0
	for _, c := range sample {
		sumA += children[c].parts[i].a
		sumB += children[c].parts[i].b
	}
	if sumB == 0 {
		return 0, false
	}
	return sumA / sumB, true
}

// mean is a metric that is the mean over children of one number each has,
// where it has one.
func mean(name string, value func(r *childResult) (float64, bool)) metric {
	return metric{name: name, extract: func(r *childResult) part {
		if v, has := value(r); has {
			return part{a: v, b: 1}
		}
		return part{}
	}}
}

// ratio is a metric that is a ratio of sums over children.
func ratio(name string, parts func(r *childResult) (num, den float64)) metric {
	return metric{name: name, extract: func(r *childResult) part {
		num, den := parts(r)
		return part{a: num, b: den}
	}}
}

// metrics are every number the experiment reads off a cell, but the
// calibration error, which is read off the bins.
func metrics() []metric {
	var all []metric
	for _, at := range checkpoints {
		all = append(all,
			mean("r1_rms_"+strconv.Itoa(at), func(r *childResult) (float64, bool) { v, has := r.errorRMS[at]; return v, has }),
			mean("r1_mean_"+strconv.Itoa(at), func(r *childResult) (float64, bool) { v, has := r.errorMean[at]; return v, has }),
		)
	}
	all = append(all,
		mean("r2_brier", func(r *childResult) (float64, bool) { return r.brier / float64(r.answers), r.answers > 0 }),
		mean("r2_bias", func(r *childResult) (float64, bool) {
			return (r.predicted - r.observed) / float64(r.answers), r.answers > 0
		}),
		mean("r3_inside", func(r *childResult) (float64, bool) { return float64(r.inside) / float64(r.answers), r.answers > 0 }),
		mean("r3_below_0.5", func(r *childResult) (float64, bool) { return float64(r.below) / float64(r.answers), r.answers > 0 }),
		mean("r3_above_0.95", func(r *childResult) (float64, bool) { return float64(r.over) / float64(r.answers), r.answers > 0 }),
		ratio("r4_false", func(r *childResult) (num, den float64) { return float64(r.wrongly), float64(r.declared) }),
		ratio("r4_false_0.70", func(r *childResult) (num, den float64) { return float64(r.looselyWrong), float64(r.declared) }),
		ratio("r4_below", func(r *childResult) (num, den float64) { return float64(r.heldBelow), float64(r.declared) }),
		ratio("r4_false_at_task", func(r *childResult) (num, den float64) { return float64(r.wronglyAtTask), float64(r.atTask) }),
		mean("r4_declared", func(r *childResult) (float64, bool) { return float64(r.declared), true }),
		ratio("r5_late_answers", lateAnswers),
		ratio("r5_never", neverDeclared),
		mean("r6_lag", func(r *childResult) (float64, bool) { return r.lag / float64(r.lagged), r.lagged > 0 }),
		mean("r6_jump_answers", jumpAnswers),
		ratio("r6_jump_unsettled", jumpUnsettled),
		overallOnly(mean("r7_error_5", placedAfter(5))),
		overallOnly(mean("r7_error_10", placedAfter(10))),
		mean("r7_longest_wrong", func(r *childResult) (float64, bool) { return float64(r.longest), true }),
		mean("r7_hard_first", func(r *childResult) (float64, bool) { return float64(r.hardFirst) / firstTasks, true }),
	)
	for _, b := range chanceBins {
		all = append(all, ratio("r4b_chance_"+b.name, chancesIn(b.low, b.high)))
	}
	all = append(all, mean("r3_reachable", func(r *childResult) (float64, bool) {
		return float64(r.reachable) / float64(r.answers), r.answers > 0
	}))
	for at, w := range screenWindows {
		all = append(all,
			shown(overallOnly(mean("r8_rank_"+w.name(), rankChanges(at, false)))),
			shown(mean("r8_topic_rank_"+w.name(), rankChanges(at, true))),
		)
	}
	return all
}

func overallOnly(m metric) metric {
	m.applies = placesOverall
	return m
}

// placedAfter is how far the overall level stood from the child's after n
// answers, for a child who gave that many.
func placedAfter(n int) func(r *childResult) (float64, bool) {
	return func(r *childResult) (float64, bool) {
		off, reached := r.placed[n]
		return math.Abs(off), reached
	}
}

// chanceBins share out the true chances on the attempts R4b counts over, so
// that R4b can be read beside how likely the child truly was to answer the
// tasks it counts over.
var chanceBins = []struct {
	name      string
	low, high float64
}{
	{"below_0.5", 0, 0.5},
	{"0.5_to_0.6", 0.5, 0.6},
	{"0.6_to_0.7", 0.6, 0.7},
	{"0.7_to_0.8", 0.7, 0.8},
	{"0.8_up", 0.8, math.Inf(1)},
}

// chancesIn is the share of the attempts R4b counts over, up to the longest
// chain, whose true chance lies in [low, high).
func chancesIn(low, high float64) func(r *childResult) (num, den float64) {
	return func(r *childResult) (num, den float64) {
		for _, pair := range r.pairs {
			for _, chance := range pair.chances[:min(len(pair.chances), longestChain)] {
				den++
				if chance >= low && chance < high {
					num++
				}
			}
		}
		return num, den
	}
}

func lateAnswers(r *childResult) (num, den float64) {
	for _, item := range r.lateItems {
		if item.declared {
			num += float64(item.answers)
			den++
		}
	}
	return num, den
}

func neverDeclared(r *childResult) (num, den float64) {
	for _, item := range r.lateItems {
		if !item.declared {
			num++
		}
		den++
	}
	return num, den
}

// jumpAnswers is how many answers the rule took to follow a child's jump; a
// child who never settled counts at the horizon, and is counted apart too.
func jumpAnswers(r *childResult) (float64, bool) {
	return float64(r.jumpFollowed), r.jumpFollowed > 0 || r.jumpDone
}

func jumpUnsettled(r *childResult) (num, den float64) {
	switch {
	case r.jumpFollowed == 0 && !r.jumpDone:
		return 0, 0
	case r.jumpDone:
		return 0, 1
	default:
		return 1, 1
	}
}

// declaredWithin is the chance that a topic a child has not truly mastered is
// declared mastered within m eligible attempts: one less the Kaplan–Meier
// estimate of its going undeclared that long. A topic whose count stops early,
// because the run ended or the child came to master it truly, counts for the
// attempts it made rather than being left out, which would leave the topics
// declared early overrepresented.
func declaredWithin(m int) reader {
	return func(children []vector, sample []int) (float64, bool) {
		undeclared, read := 1.0, false
		for j := range min(m, longestChain) {
			var atRisk, declared int
			for _, c := range sample {
				atRisk += int(children[c].attempts[j].atRisk)
				declared += int(children[c].attempts[j].declared)
			}
			if atRisk == 0 {
				break
			}
			read = true
			undeclared *= 1 - float64(declared)/float64(atRisk)
		}
		return 1 - undeclared, read
	}
}

// calibrationError is the expected calibration error of a sample's pooled
// answers: over ten bins of the predicted chance, the mean gap between
// predicted and observed, weighted by the answers in each bin.
func calibrationError(children []vector, sample []int) (float64, bool) {
	var bins [calibrationBins]bin
	total := 0
	for _, c := range sample {
		for b := range calibrationBins {
			add := children[c].bins[b]
			bins[b].predicted += add.predicted
			bins[b].observed += add.observed
			bins[b].n += add.n
			total += add.n
		}
	}
	if total == 0 {
		return 0, false
	}
	gap := 0.0
	for _, b := range bins {
		gap += math.Abs(b.predicted-b.observed) / float64(total)
	}
	return gap, true
}

// summary is a number with its interval.
type summary struct {
	value, low, high float64
	has              bool
}

// reader reads one number off a sample of children.
type reader func(children []vector, sample []int) (float64, bool)

// pooled is a number read off a sample of children as a whole rather than as
// one sum over what each brings: the calibration error, the chance of a false
// "mastered" within m attempts, and how far the card's rating moves on most
// answers.
type pooled struct {
	name string
	read reader
	// applies says whether the number means anything under a rule; nil when
	// it does under every rule.
	applies func(r *rule) bool
}

// calibrationName is the name the expected calibration error is written
// under.
const calibrationName = "r2_ece"

// pooledMetrics are the numbers read off a sample as a whole, which come after
// the metrics in every list of them.
func pooledMetrics() []pooled {
	all := []pooled{{name: calibrationName, read: calibrationError}}
	for _, m := range chainLengths {
		all = append(all, pooled{name: "r4b_" + strconv.Itoa(m), read: declaredWithin(m)})
	}
	for at, w := range screenWindows {
		all = append(all, pooled{name: "r8_move_p95_" + w.name(), read: movesAt(at, 95), applies: notCeiling})
	}
	return all
}

// readerOf is the reader of the metric at index i, or of a pooled number when
// i is past the metrics.
func readerOf(i int, all []metric) reader {
	if i >= len(all) {
		return pooledMetrics()[i-len(all)].read
	}
	return func(children []vector, sample []int) (float64, bool) { return read(children, i, sample) }
}

func everyone(n int) []int {
	all := make([]int, n)
	for i := range all {
		all[i] = i
	}
	return all
}

// summarize reads a number off all the children, and its interval off
// samples of them drawn again. A number no sample can read has no interval,
// and counts as no number rather than one with undefined ends.
func summarize(readOff reader, children []vector, rng *rand.Rand) summary {
	value, has := readOff(children, everyone(len(children)))
	if !has {
		return summary{}
	}
	draws := make([]float64, 0, resamples)
	sample := make([]int, len(children))
	for range resamples {
		for i := range sample {
			sample[i] = rng.IntN(len(children))
		}
		if v, ok := readOff(children, sample); ok {
			draws = append(draws, v)
		}
	}
	if len(draws) == 0 {
		return summary{}
	}
	low, high := interval(draws)
	return summary{value: value, low: low, high: high, has: true}
}

// compare reads a number off two cells' runs of the same children and gives
// the difference, the first less the second, with an interval from samples of
// the children drawn again, the same children for both. A difference no sample
// can read counts as none, as in summarize.
func compare(readOff reader, first, second []vector, rng *rand.Rand) summary {
	a, hasA := readOff(first, everyone(len(first)))
	b, hasB := readOff(second, everyone(len(second)))
	if !hasA || !hasB {
		return summary{}
	}
	draws := make([]float64, 0, resamples)
	sample := make([]int, len(first))
	for range resamples {
		for i := range sample {
			sample[i] = rng.IntN(len(first))
		}
		x, okX := readOff(first, sample)
		y, okY := readOff(second, sample)
		if okX && okY {
			draws = append(draws, x-y)
		}
	}
	if len(draws) == 0 {
		return summary{}
	}
	low, high := interval(draws)
	return summary{value: a - b, low: low, high: high, has: true}
}

func interval(draws []float64) (low, high float64) {
	if len(draws) == 0 {
		return math.NaN(), math.NaN()
	}
	slices.Sort(draws)
	at := func(share float64) float64 { return draws[min(len(draws)-1, int(share*float64(len(draws))))] }
	return at(tail), at(1 - tail)
}
