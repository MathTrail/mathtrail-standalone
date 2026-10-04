package main

import (
	"math"
	"math/rand/v2"
	"strconv"
)

// The criterion a new step for the rating is chosen by, written down before
// any candidate ran, so that what the candidates turn out to give cannot shape
// what they are judged by. A run reads it for every rule: the constraints a
// candidate must meet, and its score, the share of the way from the service
// to the ceiling it closes. Choosing among the rules that meet it — the best,
// its equals, the simplest of those, the confirmation on held-out seeds and
// the exit — is done from these numbers by whoever runs the candidates.

// decisionChildren is how many children a cell needs for its criterion to
// decide anything. A run of fewer is a rough look, and reads not worse as no
// clear harm.
const decisionChildren = 4000

// The goals of the step: the main learner's lag at most this share of the
// service's, its corridor closing this share of the way from the service's to
// the ceiling's, and at most this share of children not caught up after a
// jump. corridorShare is the share the goal set for children who learn fast —
// 35 % of their tasks in the corridor instead of the service's 26 % —
// closes of the way to the ceiling's 61 %, read once, on the run that wrote
// the criterion down, and fixed: (0.35 − 0.257) / (0.605 − 0.257).
const (
	lagShare      = 0.45
	corridorShare = 0.27
	unsettledMost = 0.25
)

// The tolerances of not worse, as the author set them: the error after 200
// answers in logits, the corridor and false masteries as shares. A tolerance
// is held to no less than resolutionWidths standard errors of the paired
// difference, which a candidate exactly as good as the service stays within
// 99 times in 100: a finer one would fail good candidates by chance across
// some fifty checks.
const (
	errorTolerance    = 0.01
	corridorTolerance = 0.01
	falseTolerance    = 0.02
	resolutionWidths  = 4.3
)

// normalQuantile is the 97.5th percentile of the standard normal, which turns
// the half-width of a 95 % interval into a standard error.
const normalQuantile = 1.959964

// verdicts and marks a reading of a constraint ends in.
const (
	met        = "met"
	notMet     = "not met"
	unread     = "unread"
	reached    = "reached"
	onTheEdge  = "on the edge"
	notReached = "not reached"
)

// reading is one constraint as a run reads it for one rule: its value, with
// the interval the run gives it, the bound it is held to, whether it holds,
// and whether the whole interval lies on the bound's right side, across it, or
// on the wrong side.
type reading struct {
	kind, name       string
	generator        generator
	metric           string
	bound            float64
	value, low, high float64
	verdict, mark    string
}

// goal is a constraint read on a rule's own point estimate, against a bound
// read off the service's cell, and the ceiling's for a share of the way.
type goal struct {
	name      string
	generator generator
	metric    string
	// size reads the value as its size: the lag is as far behind as ahead.
	size   bool
	atMost bool
	bound  func(cr *criterionRun) (float64, bool)
}

// notWorse is a constraint read on the paired difference between a rule and
// the service, on the same children: the worse end of its interval lies
// within the tolerance.
type notWorse struct {
	generator      generator
	metric         string
	higherIsBetter bool
	tolerance      float64
}

// badness is a measure the score reads, as a size less of which is better.
type badness struct {
	metric string
	size   func(v float64) float64
}

// scoredGenerator is a generator the score is read on, with its weight and
// its measures.
type scoredGenerator struct {
	generator generator
	weight    float64
	measures  []badness
}

// The measures the score reads: the error after 200 answers, the shortfall
// of the corridor, the size of the lag, the answers until the estimate catches
// up after a jump or a drop, the error of the overall level after 10 answers,
// and the share of hard first tasks.
var (
	errorAfter200 = badness{metric: "r1_rms_200", size: same}
	shortfall     = badness{metric: "r3_inside", size: func(v float64) float64 { return 1 - v }}
	lagSize       = badness{metric: "r6_lag", size: math.Abs}
	catchingUp    = badness{metric: "r6_jump_answers", size: same}
	placement     = badness{metric: "r7_error_10", size: same}
	hardFirst     = badness{metric: "r7_hard_first", size: same}
)

func same(v float64) float64 { return v }

// scoredGenerators are the generators the score is read on: a child who
// stays put weighs 1, children who learn 2 between them, children who change
// at once 1, and a child placed a level off 0.5.
var scoredGenerators = []scoredGenerator{
	{generator: staticChildren, weight: 1, measures: []badness{errorAfter200, shortfall}},
	{generator: learningHalf, weight: 1.25, measures: []badness{lagSize, shortfall}},
	{generator: learningFades, weight: 0.5, measures: []badness{lagSize, shortfall}},
	{generator: learning, weight: 0.25, measures: []badness{lagSize, shortfall}},
	{generator: jumping, weight: 0.5, measures: []badness{catchingUp, shortfall}},
	{generator: dropping, weight: 0.5, measures: []badness{catchingUp, shortfall}},
	{generator: misplaced, weight: 0.5, measures: []badness{placement, hardFirst}},
}

// screenGenerators are the generators what the child is shown is held on:
// the child who stays put, and the main learner.
var screenGenerators = []generator{staticChildren, learningHalf}

// goals are the step's goals: the main learner's lag and corridor, the jump,
// and the screen.
func goals() []goal {
	all := []goal{
		{
			name: "the lag at most " + percentOf(lagShare) + " of the service's", generator: learningHalf, metric: "r6_lag", size: true, atMost: true,
			bound: func(cr *criterionRun) (float64, bool) {
				lag, has := cr.serviceValue(learningHalf, "r6_lag")
				return lagShare * math.Abs(lag), has
			},
		},
		{
			name: "the corridor a share of the way to the ceiling", generator: learningHalf, metric: "r3_inside",
			bound: func(cr *criterionRun) (float64, bool) {
				service, hasService := cr.serviceValue(learningHalf, "r3_inside")
				ceiling, hasCeiling := cr.ceilingValue(learningHalf, "r3_inside")
				return service + corridorShare*(ceiling-service), hasService && hasCeiling
			},
		},
		{
			name: "children not caught up after the jump at most " + percentOf(unsettledMost), generator: jumping, metric: "r6_jump_unsettled", atMost: true,
			bound: func(*criterionRun) (float64, bool) { return unsettledMost, true },
		},
	}
	early, earlyLabel := screenWindows[0].name(), screenWindows[0].label()
	for _, g := range screenGenerators {
		for _, w := range screenWindows {
			all = append(all,
				goal{
					name: "the card's move in answers " + w.label() + " no more than the service's in answers " + earlyLabel, generator: g,
					metric: "r8_move_p95_" + w.name(), atMost: true,
					bound: func(cr *criterionRun) (float64, bool) { return cr.serviceValue(g, "r8_move_p95_"+early) },
				},
				goal{
					name: "the rank's changes in answers " + w.label() + " no more than the service's in answers " + earlyLabel, generator: g,
					metric: "r8_rank_" + w.name(), atMost: true,
					bound: func(cr *criterionRun) (float64, bool) { return cr.serviceValue(g, "r8_rank_"+early) },
				},
			)
		}
	}
	return all
}

// notWorseChecks are the checks of not worse: on every generator the error
// after 200 answers, the corridor and false masteries.
func notWorseChecks() []notWorse {
	var all []notWorse
	for _, g := range allGenerators {
		all = append(all,
			notWorse{generator: g, metric: "r1_rms_200", tolerance: errorTolerance},
			notWorse{generator: g, metric: "r3_inside", higherIsBetter: true, tolerance: corridorTolerance},
			notWorse{generator: g, metric: "r4_false", tolerance: falseTolerance},
		)
	}
	return all
}

// criterionRun is a run as the criterion reads it: its cells and their
// numbers, the service, the ceiling and the rule the bench's resolution is
// measured with, and whether the run has the children a decision needs.
type criterionRun struct {
	all       []cell
	summaries [][]summary
	results   [][]vector
	ms        []metric
	metricAt  map[string]int
	cellAt    map[string]int
	service   *rule
	ceiling   *rule
	measuring *rule
	decides   bool
}

// measuringRule is the rule the bench's resolution is measured with: one that
// steps differently from the service, but not wildly.
const measuringRule = "constant_slow"

func newCriterionRun(all []cell, summaries [][]summary, results [][]vector, ms []metric, children int) *criterionRun {
	cr := &criterionRun{
		all: all, summaries: summaries, results: results, ms: ms,
		metricAt: map[string]int{}, cellAt: map[string]int{}, decides: children >= decisionChildren,
	}
	for i, name := range metricNames(ms) {
		cr.metricAt[name] = i
	}
	for c := range all {
		cr.cellAt[all[c].name()] = c
		switch r := all[c].rule; {
		case r.service:
			cr.service = r
		case r.ceiling:
			cr.ceiling = r
		case r.name == measuringRule:
			cr.measuring = r
		}
	}
	return cr
}

// cellOf is the place of a rule's cell on a generator among the run's cells.
func (cr *criterionRun) cellOf(r *rule, g generator) (int, bool) {
	if r == nil {
		return 0, false
	}
	c, found := cr.cellAt[(&cell{rule: r, generator: g}).name()]
	return c, found
}

// summaryOf is a metric's number in a rule's cell on a generator.
func (cr *criterionRun) summaryOf(r *rule, g generator, metric string) (summary, bool) {
	c, hasCell := cr.cellOf(r, g)
	i, hasMetric := cr.metricAt[metric]
	if !hasCell || !hasMetric || !cr.summaries[c][i].has {
		return summary{}, false
	}
	return cr.summaries[c][i], true
}

func (cr *criterionRun) serviceValue(g generator, metric string) (float64, bool) {
	s, has := cr.summaryOf(cr.service, g, metric)
	return s.value, has
}

func (cr *criterionRun) ceilingValue(g generator, metric string) (float64, bool) {
	s, has := cr.summaryOf(cr.ceiling, g, metric)
	return s.value, has
}

// readGoal reads a goal for a rule on its own point estimate.
func (cr *criterionRun) readGoal(r *rule, g goal) reading {
	rd := reading{kind: "goal", name: g.name, generator: g.generator, metric: g.metric, verdict: unread}
	own, hasOwn := cr.summaryOf(r, g.generator, g.metric)
	bound, hasBound := g.bound(cr)
	rd.bound = bound
	if !hasOwn || !hasBound {
		return rd
	}
	rd.value, rd.low, rd.high = own.value, own.low, own.high
	if g.size {
		rd.value, rd.low, rd.high = sizeOf(own)
	}
	if !g.atMost {
		// Held to at least the bound: read as at most, on the other side of
		// zero, so that one comparison serves both.
		rd.verdict, rd.mark = judged(-rd.value, -rd.high, -rd.low, -bound)
		return rd
	}
	rd.verdict, rd.mark = judged(rd.value, rd.low, rd.high, bound)
	return rd
}

// judged is the verdict and the mark of a value held to at most a bound: met
// when the value is, reached when its whole interval is, not reached when the
// whole interval lies past it.
func judged(value, low, high, bound float64) (verdict, mark string) {
	verdict = notMet
	if value <= bound {
		verdict = met
	}
	switch {
	case high <= bound:
		return verdict, reached
	case low > bound:
		return verdict, notReached
	}
	return verdict, onTheEdge
}

// sizeOf is a number read as its size, with the interval its size has: across
// zero, the smallest size is none.
func sizeOf(s summary) (value, low, high float64) {
	value = math.Abs(s.value)
	switch {
	case s.low >= 0:
		return value, s.low, s.high
	case s.high <= 0:
		return value, -s.high, -s.low
	}
	return value, 0, max(-s.low, s.high)
}

// pairedDifference is a rule's cell on a generator less the service's, on the
// same children, with its interval, drawn on a stream named after what it
// reads.
func (cr *criterionRun) pairedDifference(r *rule, g generator, metric string) (summary, bool) {
	c, hasCell := cr.cellOf(r, g)
	s, hasService := cr.cellOf(cr.service, g)
	i, hasMetric := cr.metricAt[metric]
	if !hasCell || !hasService || !hasMetric {
		return summary{}, false
	}
	stream := seeded("criterion/"+cr.all[c].name()+"/"+metric, "bootstrap")
	diff := compare(readerOf(i, cr.ms), cr.results[c], cr.results[s], stream)
	return diff, diff.has
}

// resolution is what the bench resolves on a check: the paired difference
// between the measuring rule and the service on it, and the standard error
// that difference shows, if the run gives it.
func (cr *criterionRun) resolution(nw notWorse) checkResolution {
	res := checkResolution{check: nw}
	res.measured, res.has = cr.pairedDifference(cr.measuring, nw.generator, nw.metric)
	if res.has {
		res.se = (res.measured.high - res.measured.low) / 2 / normalQuantile
	}
	res.tolerance = toleranceOf(nw, res.se, res.has)
	return res
}

// toleranceOf is the tolerance a check is read with, given the bench's
// resolution on it: the author's, widened to the resolution where the run
// cannot tell less.
func toleranceOf(nw notWorse, se float64, has bool) float64 {
	if has {
		return max(nw.tolerance, resolutionWidths*se)
	}
	return nw.tolerance
}

// readNotWorse reads a check of not worse for a rule: the worse end of its
// paired difference with the service within the tolerance, in a decision
// run; in a rough look, the better end not past it. The service less itself
// is nothing on every child, and the measuring rule's difference was read for
// the check's resolution already, so neither is drawn again.
func (cr *criterionRun) readNotWorse(r *rule, res *checkResolution) reading {
	nw := res.check
	rd := reading{kind: "not worse", name: "not worse than the service", generator: nw.generator, metric: nw.metric, bound: res.tolerance, verdict: unread}
	var diff summary
	var has bool
	switch r {
	case cr.service:
		_, has = cr.summaryOf(r, nw.generator, nw.metric)
		diff.has = has
	case cr.measuring:
		diff, has = res.measured, res.has
	default:
		diff, has = cr.pairedDifference(r, nw.generator, nw.metric)
	}
	if !has {
		return rd
	}
	rd.value, rd.low, rd.high = diff.value, diff.low, diff.high
	rd.verdict, rd.mark = notWorseVerdict(diff, nw.higherIsBetter, res.tolerance, cr.decides)
	return rd
}

// notWorseVerdict is the verdict and the mark of a paired difference with
// the service held to a tolerance: met and reached when even its worse end is
// within it; not met and not reached when even its better end is past it; and
// on the edge between, which a decision run does not let through and a rough
// look does.
func notWorseVerdict(diff summary, higherIsBetter bool, tolerance float64, decides bool) (verdict, mark string) {
	worse, better := diff.high, diff.low
	if higherIsBetter {
		worse, better = -diff.low, -diff.high
	}
	switch {
	case worse <= tolerance:
		return met, reached
	case better > tolerance:
		return notMet, notReached
	case decides:
		return notMet, onTheEdge
	}
	return met, onTheEdge
}

// scoreSheet is what a rule's score is read from: for every generator of the
// score, the rule's and the service's children, and for every measure its
// reader and the way from the service to the ceiling on it.
type scoreSheet struct {
	parts []scorePart
	total float64
}

// scorePart is one generator of a score.
type scorePart struct {
	weight        float64
	rule, service []vector
	terms         []scoreTerm
}

// scoreTerm is one measure of a generator of a score.
type scoreTerm struct {
	read reader
	size func(float64) float64
	way  float64
}

// sheetOf is the score sheet of a rule, or false when the run cannot read the
// score: a cell is missing, a number the rule, the service or the ceiling
// gives — the rule's own included, which a measure that means nothing under
// the rule does not give — or a way of more than nothing.
func (cr *criterionRun) sheetOf(r *rule) (scoreSheet, bool) {
	var sheet scoreSheet
	for _, sg := range scoredGenerators {
		c, hasCell := cr.cellOf(r, sg.generator)
		s, hasService := cr.cellOf(cr.service, sg.generator)
		if !hasCell || !hasService {
			return scoreSheet{}, false
		}
		part := scorePart{weight: sg.weight, rule: cr.results[c], service: cr.results[s]}
		for _, b := range sg.measures {
			_, hasOwn := cr.summaryOf(r, sg.generator, b.metric)
			service, hasServiceValue := cr.serviceValue(sg.generator, b.metric)
			ceiling, hasCeiling := cr.ceilingValue(sg.generator, b.metric)
			way := b.size(service) - b.size(ceiling)
			if !hasOwn || !hasServiceValue || !hasCeiling || way <= 0 {
				return scoreSheet{}, false
			}
			part.terms = append(part.terms, scoreTerm{read: readerOf(cr.metricAt[b.metric], cr.ms), size: b.size, way: way})
		}
		sheet.parts = append(sheet.parts, part)
		sheet.total += sg.weight
	}
	return sheet, true
}

// read is the score on a sample of each generator's children: the weighted
// mean, over the generators, of the mean share of the way the rule closes on
// its measures.
func (sheet *scoreSheet) read(samples [][]int) (float64, bool) {
	score := 0.0
	for p, part := range sheet.parts {
		closed := 0.0
		for _, t := range part.terms {
			service, hasService := t.read(part.service, samples[p])
			own, hasOwn := t.read(part.rule, samples[p])
			if !hasService || !hasOwn {
				return 0, false
			}
			closed += (t.size(service) - t.size(own)) / t.way
		}
		score += part.weight * closed / float64(len(part.terms))
	}
	return score / sheet.total, true
}

// scoreOf is a rule's score, with an interval from samples of each
// generator's children drawn apart, the rule and the service on the same
// children.
func (cr *criterionRun) scoreOf(r *rule) summary {
	sheet, has := cr.sheetOf(r)
	if !has {
		return summary{}
	}
	whole := make([][]int, len(sheet.parts))
	for p, part := range sheet.parts {
		whole[p] = everyone(len(part.rule))
	}
	value, has := sheet.read(whole)
	if !has {
		return summary{}
	}
	rng := seeded("criterion/"+r.name+"/"+string(r.shape)+"/score", "bootstrap")
	draws := make([]float64, 0, resamples)
	samples := make([][]int, len(sheet.parts))
	for p := range samples {
		samples[p] = make([]int, len(whole[p]))
	}
	for range resamples {
		drawSamples(samples, rng)
		if v, ok := sheet.read(samples); ok {
			draws = append(draws, v)
		}
	}
	if len(draws) == 0 {
		return summary{}
	}
	low, high := interval(draws)
	return summary{value: value, low: low, high: high, has: true}
}

// drawSamples draws every sample anew, each from its own children.
func drawSamples(samples [][]int, rng *rand.Rand) {
	for _, sample := range samples {
		for i := range sample {
			sample[i] = rng.IntN(len(sample))
		}
	}
}

// ruleCriterion is the criterion as a run reads it for one rule.
type ruleCriterion struct {
	rule     *rule
	readings []reading
	score    summary
}

// checkResolution is a check of not worse with what the bench resolves on it:
// the measuring rule's paired difference with the service, its standard
// error, whether the run gives them, and the tolerance the check is read with.
type checkResolution struct {
	check     notWorse
	measured  summary
	se        float64
	has       bool
	tolerance float64
}

// readCriterion reads the criterion for every rule of a run, the rules at
// once, and what the bench resolves on each check of not worse.
func readCriterion(cr *criterionRun) ([]ruleCriterion, []checkResolution) {
	checks := notWorseChecks()
	resolutions := make([]checkResolution, len(checks))
	eachAtOnce(len(checks), func(i int) { resolutions[i] = cr.resolution(checks[i]) })
	var ruleList []*rule
	for c := range cr.all {
		if !containsRule(ruleList, cr.all[c].rule) {
			ruleList = append(ruleList, cr.all[c].rule)
		}
	}
	read := make([]ruleCriterion, len(ruleList))
	eachAtOnce(len(ruleList), func(i int) {
		r := ruleList[i]
		rc := ruleCriterion{rule: r, score: cr.scoreOf(r)}
		for _, g := range goals() {
			rc.readings = append(rc.readings, cr.readGoal(r, g))
		}
		for c := range resolutions {
			rc.readings = append(rc.readings, cr.readNotWorse(r, &resolutions[c]))
		}
		read[i] = rc
	})
	return read, resolutions
}

// containsRule says whether a list holds a rule of the same name and
// structure.
func containsRule(rules []*rule, r *rule) bool {
	for _, held := range rules {
		if sameRule(held, r) {
			return true
		}
	}
	return false
}

// passChance is the chance a candidate exactly as good as the service passes
// a check of not worse read with a tolerance, given the standard error of the
// paired difference: in a decision run, that the worse end of its interval
// stays within the tolerance; in a rough look, that its better end does.
func passChance(tolerance, se float64, decides bool) float64 {
	if se <= 0 {
		return 1
	}
	margin := tolerance/se - normalQuantile
	if !decides {
		margin = tolerance/se + normalQuantile
	}
	return 0.5 * (1 + math.Erf(margin/math.Sqrt2))
}

// percentOf is a share as a goal's name says it: 45 % for 0.45.
func percentOf(share float64) string { return strconv.Itoa(int(math.Round(100*share))) + " %" }
