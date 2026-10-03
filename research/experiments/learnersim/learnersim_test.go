package main

import (
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The step rules of the harness, at the service's constants, are the
// service's own update: the same θ and δ to within 1e-12.
func TestStepRulesAreTheServicesUpdate(t *testing.T) {
	t.Parallel()
	rng := seeded("step rules", "test")
	for range 10000 {
		state := rating.State{
			Theta: 6*rng.Float64() - 1, Delta: 2*rng.Float64() - 1,
			Answers: rng.IntN(300), TopicAnswers: rng.IntN(60),
		}
		beta, correct := 8*rng.Float64()-2, rng.IntN(2) == 0
		want := rating.Update(state, beta, correct)
		s := newStepRule(both, 0)
		s.theta, s.delta["t"], s.answers, s.inTopic["t"] = state.Theta, state.Delta, state.Answers, state.TopicAnswers
		s.answered("t", beta, correct)
		if math.Abs(s.theta-want.Theta) > 1e-12 || math.Abs(s.delta["t"]-want.Delta) > 1e-12 {
			t.Fatalf("from %+v at β %v, %v: got θ %v δ %v, want θ %v δ %v", state, beta, correct, s.theta, s.delta["t"], want.Theta, want.Delta)
		}
	}
}

// On children the model describes exactly, with tasks written as asked, the
// service's estimate converges: its error after 200 answers is smaller than
// after 20.
func TestTheEstimateConvergesOnTheTrivialLearner(t *testing.T) {
	t.Parallel()
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	service := &rule{name: "shrinking", shape: both, service: true}
	var early, late float64
	const children = 100
	for i := range children {
		r, err := run(w, service, newChild(exactlyAsWritten, i, w.topics))
		if err != nil {
			t.Fatal(err)
		}
		early += r.errorRMS[20]
		late += r.errorRMS[200]
	}
	early, late = early/children, late/children
	if late >= early {
		t.Errorf("root mean square error after 200 answers %.3f, after 20 %.3f; want it smaller after 200", late, early)
	}
}

// Glicko-2 reproduces Glickman's worked example to the figures he prints. He
// rounds as he goes: the new μ to −0.2069 and φ to 0.8722 before he converts
// them, which makes the rating 1464.06 where the unrounded μ makes it 1464.05,
// and he cuts the volatility of 0.059996 to 0.05999. The intermediate figures
// are matched at his precision, and the final ones within his rounding.
func TestGlickoReproducesGlickmansExample(t *testing.T) {
	t.Parallel()
	r, rd, volatility := glickmanExample()
	mu, phi := (r-1500)/glickoScale, rd/glickoScale
	if math.Round(mu*1e4)/1e4 != -0.2069 || math.Round(phi*1e4)/1e4 != 0.8722 || math.Trunc(volatility*1e5)/1e5 != 0.05999 {
		t.Errorf("got μ %.4f, φ %.4f, σ %.5f; want −0.2069, 0.8722 and 0.05999", mu, phi, volatility)
	}
	if math.Abs(r-1464.06) > 0.01 || math.Abs(rd-151.52) > 0.005 || math.Abs(volatility-0.05999) > 0.00001 {
		t.Errorf("got rating %.4f, RD %.4f, volatility %.6f; want 1464.06, 151.52 and 0.05999 within his rounding", r, rd, volatility)
	}
}

// Urnings, in its own simulated example at a smaller scale, holds each
// player's average scaled urning near the player's true proportion.
func TestUrningsReproducesItsSimulatedExample(t *testing.T) {
	t.Parallel()
	if gap := tournamentGap(tournamentPlayers, tournamentUrn, tournamentGames, seeded("tournament", "check")); gap > 0.02 {
		t.Errorf("mean gap between average scaled urnings and true proportions %.4f, want at most 0.02", gap)
	}
}

// Urnings against items of fixed proportion, drawn independently of the urn,
// keeps the urn at the binomial law of the learner's own proportion: at every
// count within the protocol's tolerance, and, with tasks around the learner's
// level, within a total variation the urn misses without its acceptance step.
func TestFixedItemUrningsHaveABinomialLaw(t *testing.T) {
	t.Parallel()
	if gap := fixedItemCheck(); gap > 0.01 {
		t.Errorf("largest gap between the urn's time at a count and the binomial chance %.4f, want at most 0.01", gap)
	}
	if variation := fixedItemNearCheck(); variation > 0.01 {
		t.Errorf("total variation from the binomial law with tasks near the learner %.4f, want at most 0.01", variation)
	}
}

// The chain of mastery agrees with a simulation of itself, at the points and
// with the draws whose gaps the paper reports.
func TestTheChainAgreesWithItsSimulation(t *testing.T) {
	t.Parallel()
	for i, gap := range chainGaps() {
		if point := chainPoints[i]; math.Abs(gap) > 0.005 {
			t.Errorf("p %v, h %v, m %d: the simulation is %.4f from the computed chance", point.p, point.h, point.m, gap)
		}
	}
}

// A child is drawn the same every time, and a generator departs from the base
// population in its one respect.
func TestChildrenAreRepeatableAndDepartAsTheirGenerator(t *testing.T) {
	t.Parallel()
	topics := []string{"a", "b", "c"}
	one, again := newChild(staticChildren, 7, topics), newChild(staticChildren, 7, topics)
	if one.theta != again.theta || one.next() != again.next() {
		t.Error("the same child was drawn twice differently")
	}
	low, high := newChild(misplaced, 0, topics), newChild(misplaced, 1, topics)
	if low.theta != low.start-misplacedBy || high.theta != high.start+misplacedBy {
		t.Errorf("misplaced children stand at %v and %v from their starts, want −2.5 and +2.5", low.theta-low.start, high.theta-high.start)
	}
	if c := newChild(jumping, 3, topics); c.jumpAt < jumpEarliest || c.jumpAt > jumpLatest {
		t.Errorf("a jump at answer %d, want from %d to %d", c.jumpAt, jumpEarliest, jumpLatest)
	}
	if c := newChild(exactlyAsWritten, 0, topics); c.writtenDifficulty(1, 0, 2) != 1 {
		t.Error("a child of tasks written exactly met a writing error")
	}
}

// The chance of a false "mastered" within m attempts counts a topic whose
// count stopped early for the attempts it made. Of four topics — declared at
// the first attempt, stopped undeclared after two, declared at the third, and
// undeclared after five — one in four is declared at the first attempt, and
// one in two of the two still counted at the third: the chance within three is
// 1 − 3/4 · 1/2 = 5/8, where leaving out the topic that stopped would make it
// two in three.
func TestDeclaredWithinCountsTopicsForTheAttemptsTheyMade(t *testing.T) {
	t.Parallel()
	r := &childResult{pairs: map[string]*pairTrack{
		"early":   {attempts: 1, declaredAt: 1},
		"stopped": {attempts: 2},
		"middle":  {attempts: 3, declaredAt: 3},
		"long":    {attempts: 5},
	}}
	children := []vector{vectorOf(r, nil)}
	for m, want := range map[int]float64{1: 1.0 / 4, 3: 5.0 / 8, 5: 5.0 / 8} {
		if got, read := declaredWithin(m)(children, []int{0}); !read || math.Abs(got-want) > 1e-12 {
			t.Errorf("within %d: %v (read %v), want %v", m, got, read, want)
		}
	}
	if _, read := declaredWithin(3)([]vector{vectorOf(&childResult{}, nil)}, []int{0}); read {
		t.Error("a child with no eligible attempts gave a chance, want none")
	}
}

// The paper cites a point of the chain by a key that names the point, so the
// key must say the chance, the share of hints and the attempts it was computed
// at.
func TestChainNumbersNameTheirPoint(t *testing.T) {
	t.Parallel()
	got := chainNumbers([]chainRow{{p: 0.7, h: 0, m: 5, chance: 0.549}, {p: 0.05, h: 0.2, m: 50, chance: 0.1}})
	want := []string{"chain_p70_h00_m5=0.5490", "chain_p05_h20_m50=0.1000"}
	if !slices.Equal(got, want) {
		t.Errorf("chainNumbers = %q, want %q", got, want)
	}
}

// Until its own first step, a step rule that starts after the trial series
// stands where the series stands, in every topic, so that it chooses and
// predicts from the estimate the service has.
func TestAStepRuleHoldsTheTrialsEstimateDuringTheSeries(t *testing.T) {
	t.Parallel()
	w, err := newWorld(rating.TrialAnswers)
	if err != nil {
		t.Fatal(err)
	}
	for _, shape := range []structure{general, topics, both} {
		r := &rule{name: "constant", shape: shape, trial: true, target: 0.5, make: steps(shape, func(s *stepRule) { s.decay = 0 })}
		s := newSession(w, r, newChild(misplaced, 1, w.topics))
		for k := range rating.TrialAnswers {
			if err := s.step(k); err != nil {
				t.Fatal(err)
			}
			for _, topic := range w.topics {
				if got, want := s.est.level(topic), s.p.LevelIn(topic); math.Abs(got-want) > 1e-12 {
					t.Fatalf("%s, after answer %d: the rule stands at %v in %s, the series at %v", shape, k+1, got, topic, want)
				}
			}
		}
	}
}

// A mastery declared on the very answer at which the child is first truly
// mastered counts as declared after no answers, not as one that never began.
func TestAMasteryDeclaredAtOnceIsLateByNoAnswers(t *testing.T) {
	t.Parallel()
	p := profile.New(profile.Student{Grade: 3, Pseudonym: "sim"}, "simulation", firstDay)
	level := rating.Grades34
	summary := p.Topics["t"]
	summary.MasteredLevel = &level
	p.Topics["t"] = summary
	r := newChildResult(nil)
	r.followLate(&session{p: p}, topicLevel{topic: "t", level: level}, masteredAt, false, true)
	if num, den := lateAnswers(r); num != 0 || den != 1 {
		t.Errorf("late answers %v over %v items, want 0 over 1", num, den)
	}
	if num, den := neverDeclared(r); num != 0 || den != 1 {
		t.Errorf("never declared %v of %v, want 0 of 1", num, den)
	}
}

// The true chances on the attempts R4b counts over are shared out into bins
// that leave none out and count none twice.
func TestTheChancesOfR4bAreSharedOutWhole(t *testing.T) {
	t.Parallel()
	r := &childResult{pairs: map[string]*pairTrack{
		"a": {chances: []float64{0.45, 0.5, 0.55}},
		"b": {chances: []float64{0.65, 0.7, 0.8, 0.99}},
	}}
	want := []float64{1, 2, 1, 1, 2}
	total := 0.0
	for i, b := range chanceBins {
		num, den := chancesIn(b.low, b.high)(r)
		if num != want[i] || den != 7 {
			t.Errorf("%s: %v of %v, want %v of 7", b.name, num, den, want[i])
		}
		total += num
	}
	if total != 7 {
		t.Errorf("the bins hold %v attempts, want all 7", total)
	}
}

// The error of the overall level is read only off rules that keep one: the
// structure of topics holds it at the start.
func TestPlacementIsReadOnlyOffRulesThatKeepAnOverallLevel(t *testing.T) {
	t.Parallel()
	ms := metrics()
	at := slices.IndexFunc(ms, func(m metric) bool { return m.name == "r7_error_5" })
	if at < 0 {
		t.Fatal("no metric r7_error_5")
	}
	placed := newChildResult(nil)
	placed.placed[5] = 0.3
	children := []vector{vectorOf(placed, ms)}
	for shape, want := range map[structure]bool{general: true, topics: false, both: true} {
		all := []cell{{rule: &rule{name: "constant", shape: shape}, generator: misplaced}}
		if got := summarizeCells(all, [][]vector{children}, ms)[0][at].has; got != want {
			t.Errorf("%s: r7_error_5 read = %v, want %v", shape, got, want)
		}
	}
}

// A primary comparison that names a cell or a metric the run lacks fails the
// run instead of leaving the comparison out.
func TestAComparisonOfWhatTheRunLacksFails(t *testing.T) {
	t.Parallel()
	if _, err := primaryComparisons(nil, nil, metrics()); err == nil {
		t.Error("primaryComparisons with no cells = nil error, want one")
	}
}

// A primary comparison finds a difference only when its whole interval lies on
// one side of zero, and the numbers file counts the comparisons and those found.
func TestComparisonNumbersCountTheDifferencesFound(t *testing.T) {
	t.Parallel()
	comparisons := []comparison{
		{group: "1", metric: "above", first: "a", second: "b", result: summary{value: 0.2, low: 0.1, high: 0.3, has: true}},
		{group: "1", metric: "below", first: "a", second: "b", result: summary{value: -0.2, low: -0.3, high: -0.1, has: true}},
		{group: "1", metric: "across", first: "a", second: "b", result: summary{value: 0.1, low: -0.1, high: 0.3, has: true}},
		{group: "1", metric: "touching", first: "a", second: "b", result: summary{value: 0.1, low: 0, high: 0.2, has: true}},
	}
	lines := comparisonNumbers(comparisons)
	for _, want := range []string{"primary_comparisons=4", "primary_comparisons_found=2"} {
		if !slices.Contains(lines, want) {
			t.Errorf("comparisonNumbers lack %q: %v", want, lines)
		}
	}
}

// A child's run comes to the same bits every time it is run.
func TestARunIsTheSameEveryTime(t *testing.T) {
	t.Parallel()
	w, err := newWorld(120)
	if err != nil {
		t.Fatal(err)
	}
	service := &rule{name: "shrinking", shape: both, service: true}
	for i := range 3 {
		first, err := run(w, service, newChild(jumping, i, w.topics))
		if err != nil {
			t.Fatal(err)
		}
		again, err := run(w, service, newChild(jumping, i, w.topics))
		if err != nil {
			t.Fatal(err)
		}
		if !maps.Equal(first.errorRMS, again.errorRMS) || !maps.Equal(first.errorMean, again.errorMean) || first.jumpFollowed != again.jumpFollowed {
			t.Errorf("child %d: two runs differ: %v and %v", i, first.errorRMS, again.errorRMS)
		}
	}
}

// A primary comparison the run gives no values for fails the run instead of
// being left out.
func TestAComparisonWithNoValuesFails(t *testing.T) {
	t.Parallel()
	ms := metrics()
	all := cells()
	empty := make([][]vector, len(all))
	for c := range empty {
		empty[c] = []vector{vectorOf(newChildResult(nil), ms)}
	}
	if _, err := primaryComparisons(all, empty, ms); err == nil {
		t.Error("primaryComparisons over children with no values = nil error, want one")
	}
}

// The error of the overall level after five or ten answers is read only off a
// child who gave that many.
func TestPlacementIsReadOnlyAfterTheAnswersItNames(t *testing.T) {
	t.Parallel()
	r := newChildResult(nil)
	for n, read := range map[int]bool{5: false, 10: false} {
		if _, has := placedAfter(n)(r); has != read {
			t.Errorf("after %d answers with none given: read %v, want %v", n, has, read)
		}
	}
	r.placed[5] = -0.75
	if off, has := placedAfter(5)(r); !has || off != 0.75 {
		t.Errorf("after 5 answers: %v (read %v), want 0.75", off, has)
	}
	if _, has := placedAfter(10)(r); has {
		t.Error("after 10 answers with five given: read, want nothing")
	}
}

// figureSummaries gives every cell a value for the figure's two metrics: the
// share in the corridor 1 on children who stay put and 0 elsewhere, and a
// share of false declarations of 0.5.
func figureSummaries(all []cell, names []string) [][]summary {
	summaries := make([][]summary, len(all))
	for c := range all {
		summaries[c] = make([]summary, len(names))
		inside := 0.0
		if all[c].generator == staticChildren {
			inside = 1
		}
		summaries[c][slices.Index(names, "r3_inside")] = summary{value: inside, has: true}
		summaries[c][slices.Index(names, "r4_false")] = summary{value: 0.5, has: true}
	}
	return summaries
}

// The figure of the sweep has one row for each of the fifteen rules on
// children who stay put, the service among them, in percent.
func TestTheCorridorTableHoldsTheSweepOnChildrenWhoStayPut(t *testing.T) {
	t.Parallel()
	names := metricNames(metrics())
	all := cells()
	table, err := corridorTable(all, figureSummaries(all, names), names)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(table)-1, len(sweepRules()); got != want {
		t.Fatalf("%d rows, want one for each of the %d rules", got, want)
	}
	services := 0
	for _, row := range table[1:] {
		if row[1] == "service" {
			services++
		}
		if row[2] != "100.0" || row[5] != "50.0" {
			t.Errorf("row %v: inside %s and declared wrongly %s, want 100.0 and 50.0, the values of children who stay put", row, row[2], row[5])
		}
	}
	if services != 1 {
		t.Errorf("%d rows of the service, want one", services)
	}
}

// The rows run from the most tasks in the corridor to the fewest, and a rule
// with no value or no interval for the figure fails it rather than being drawn
// at zero or without its bars.
func TestTheCorridorTableIsOrderedAndRefusesAGap(t *testing.T) {
	t.Parallel()
	names := metricNames(metrics())
	all := cells()
	summaries := figureSummaries(all, names)
	inside := slices.Index(names, "r3_inside")
	for c := range all {
		summaries[c][inside].value = float64(c) / 1000
	}
	table, err := corridorTable(all, summaries, names)
	if err != nil {
		t.Fatal(err)
	}
	previous := math.Inf(1)
	for _, row := range table[1:] {
		value, err := strconv.ParseFloat(row[2], 64)
		if err != nil || value > previous {
			t.Errorf("row %v comes after %v", row, previous)
		}
		previous = value
	}
	gaps := map[string]summary{
		"missing":          {},
		"with no interval": {value: 0.5, low: math.NaN(), high: math.NaN(), has: true},
	}
	for name, gap := range gaps {
		broken := figureSummaries(all, names)
		for c := range all {
			if all[c].generator == staticChildren && all[c].rule.service {
				broken[c][slices.Index(names, "r4_false")] = gap
			}
		}
		if _, err := corridorTable(all, broken, names); err == nil {
			t.Errorf("corridorTable with the service's false declarations %s = nil error, want one", name)
		}
	}
}

// Every rule of the sweep has a label of its own, named, with no comma or
// quotation mark, which the figure's reader of the table would split or keep.
func TestRuleLabelsAreNamedAndDistinct(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, r := range sweepRules() {
		label := ruleLabel(r)
		if strings.ContainsAny(label, ",\"") || strings.HasPrefix(label, " ") || seen[label] {
			t.Errorf("%s/%s is labelled %q: a comma, a quotation mark, no name or a label already used", r.name, r.shape, label)
		}
		seen[label] = true
	}
}

// The design the numbers file opens with is the run's own: the children and
// answers it was given, a sweep of the service and its alternatives, and the
// lag read from the answer after the hundredth to the run's last answer.
func TestTheDesignNumbersAreTheRunsOwn(t *testing.T) {
	t.Parallel()
	lines := designNumbers(40, 120)
	want := []string{
		"children_per_cell=40", "answers_per_child=120",
		"sweep_rules=" + strconv.Itoa(len(sweepRules())), "sweep_alternatives=" + strconv.Itoa(len(sweepRules())-1),
		"generators=" + strconv.Itoa(len(allGenerators)),
		"lag_window_first=101", "lag_window_last=120",
		"writing_error_sd=0.5", "learning_per_answer=0.01",
	}
	for _, line := range want {
		if !slices.Contains(lines, line) {
			t.Errorf("designNumbers lack %q: %v", line, lines)
		}
	}
}

// A number that no sample of the children drawn again can read has no
// interval, and counts as no number at all rather than one with undefined
// ends, which would reach the paper as NaN or count as no difference found.
func TestANumberWithNoIntervalCountsAsNone(t *testing.T) {
	t.Parallel()
	whole := func(children []vector, sample []int) (float64, bool) {
		return 1, slices.Equal(sample, everyone(len(children)))
	}
	children := make([]vector, 50)
	if s := summarize(whole, children, seeded("no interval", "test")); s.has {
		t.Errorf("summarize = %+v, want no number", s)
	}
	if s := compare(whole, children, children, seeded("no interval", "test")); s.has {
		t.Errorf("compare = %+v, want no number", s)
	}
}
