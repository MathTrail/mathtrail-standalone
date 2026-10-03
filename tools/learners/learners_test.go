package main

import (
	"flag"
	"io"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
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
		r := &rule{name: "constant", shape: shape, trial: true, make: steps(shape, func(s *stepRule) { s.decay = 0 })}
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

// A number that no sample of the children drawn again can read has no
// interval, and counts as no number at all rather than one with undefined
// ends, which would reach the table as NaN or count as no difference found.
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

// update rewrites the summary's snapshot from the run the test makes, for the
// change to be read before it is kept.
var update = flag.Bool("update", false, "rewrite testdata/summary.md from the run the test makes")

// A run given neither a seed nor a name draws the children the paper's run
// drew, the same as one given its seed and name outright, and a run given
// another seed or another name draws other children. It sets the seed and the
// name every other test draws with, so it does not run beside them.
func TestTheSeedAndTheNameDecideWhoIsDrawn(t *testing.T) {
	t.Cleanup(func() { masterSeed, experiment = paperSeed, paperExperiment })
	drawn := func(args ...string) float64 {
		t.Helper()
		if _, _, err := parse(args, io.Discard); err != nil {
			t.Fatalf("parse(%q) error = %v, want nil", args, err)
		}
		return newChild(staticChildren, 0, []string{"a"}).theta
	}
	paper := drawn("-seed", "20261001", "-experiment", "E-A3")
	if got := drawn(); got != paper {
		t.Errorf("a run given no seed and no name draws a child at θ %v, want the paper's %v", got, paper)
	}
	for _, args := range [][]string{{"-seed", "1"}, {"-experiment", "another"}} {
		if got := drawn(args...); got == paper {
			t.Errorf("a run given %q draws the paper's child at θ %v, want another", args, got)
		}
	}
}

// A small run is summed up as its snapshot in testdata says, the command
// prints the summary it writes, and run.txt states what the run was given. A
// change to the numbers is a change to the snapshot, which go test -update
// rewrites from the run, to be read before it is kept. The run draws with the
// paper's seed and name, which it sets, so it does not run beside the tests
// that draw.
func TestASmallRunIsSummedUpAsItsSnapshot(t *testing.T) {
	out := t.TempDir()
	var stdout, stderr strings.Builder
	if code := runCommand([]string{"-out", out, "-children", "10", "-answers", "200"}, &stdout, &stderr); code != 0 {
		t.Fatalf("learners exited %d: %s", code, stderr.String())
	}
	got := readFile(t, filepath.Join(out, "summary.md"))
	if !strings.Contains(stdout.String(), got) {
		t.Errorf("the command printed\n%s\nwant the summary it wrote:\n%s", stdout.String(), got)
	}
	run := readFile(t, filepath.Join(out, "run.txt"))
	for _, line := range []string{"seed=20261001", "experiment=E-A3", "children=10", "answers=200"} {
		if !slices.Contains(strings.Split(run, "\n"), line) {
			t.Errorf("run.txt lacks %q:\n%s", line, run)
		}
	}
	snapshot := filepath.Join("testdata", "summary.md")
	if *update {
		if err := os.WriteFile(snapshot, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if want := readFile(t, snapshot); got != want {
		t.Errorf("the summary of a small run is not %s; if the change is meant, rewrite it with go test -run %s -update and read the diff:\n%s", snapshot, t.Name(), got)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// A run too short for the error its comparisons read is refused before its
// children run, rather than after.
func TestARunTooShortForItsComparisonsIsRefused(t *testing.T) {
	t.Parallel()
	err := experimentInto(t.TempDir(), design{children: 1, answers: 199}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "at least 200") {
		t.Errorf("a run of 199 answers: error %v, want one that asks for at least 200", err)
	}
}

// The summary of a run with no numbers has a dash for every rule and headline,
// and a run that lacks a cell or a metric the summary reads is refused rather
// than summed up in dashes.
func TestTheSummaryShowsWhatItLacksAndRefusesWhatTheRunLacks(t *testing.T) {
	t.Parallel()
	all, names := cells(), metricNames(metrics())
	empty := make([][]summary, len(all))
	for c := range empty {
		empty[c] = make([]summary, len(names))
	}
	text, err := summaryOf(all, empty, names, design{})
	if err != nil {
		t.Fatalf("summaryOf a whole run = %v, want nil", err)
	}
	if got, want := strings.Count(text, noNumber), len(rules())*len(headlines); got != want {
		t.Errorf("%d dashes in the summary of a run with no numbers, want %d:\n%s", got, want, text)
	}
	if _, err := summaryOf(all[1:], empty[1:], names, design{}); err == nil {
		t.Error("summaryOf a run without the service's cell on G0 = nil error, want one")
	}
	renamed := slices.Clone(names)
	renamed[slices.Index(renamed, "r6_lag")] = "r6_other"
	if _, err := summaryOf(all, empty, renamed, design{}); err == nil {
		t.Error("summaryOf a run without r6_lag = nil error, want one")
	}
}
