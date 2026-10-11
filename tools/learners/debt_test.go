package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// within says whether two numbers agree to the last bits a sum of a few
// products may round differently.
func within(a, b float64) bool { return math.Abs(a-b) <= 1e-12 }

// The author's debt, worked by hand. A first wrong answer in a run takes a
// tenth of its charge and owes the rest; a second in a row takes two fifths; a
// right answer pays what is owed with four fifths of its gain and raises the
// level by the fifth left; a wrong answer after it starts a run again; and the
// third wrong answer in a row takes its whole charge and owes nothing more.
func TestTheAuthorsDebtWorkedByHand(t *testing.T) {
	t.Parallel()
	o := newOwing(authorsDebt.shares(), authorsDebt.repayment, authorsDebt.burn)
	for _, step := range []struct {
		name                       string
		theta, delta               float64
		correct                    bool
		wantTheta, wantDelta       float64
		wantOwedTheta, wantOwedTop float64
	}{
		{"a first wrong answer", -0.12, -0.24, false, -0.012, -0.024, 0.108, 0.216},
		{"a second in a row", -0.1, -0.2, false, -0.04, -0.08, 0.168, 0.336},
		{"a right answer", 0.05, 0.1, true, 0.01, 0.02, 0.128, 0.256},
		{"a wrong answer after it", -0.1, -0.2, false, -0.01, -0.02, 0.218, 0.436},
		{"the second of the new run", -0.1, -0.2, false, -0.04, -0.08, 0.278, 0.556},
		{"the third in a row", -0.1, -0.2, false, -0.1, -0.2, 0.278, 0.556},
		{"a right answer that pays more than its fifth", 0.3, 0.6, true, 0.06, 0.12, 0.038, 0.076},
		{"a right answer that clears what is owed", 0.3, 0.6, true, 0.262, 0.524, 0, 0},
	} {
		theta, delta := o.charged("a", step.theta, step.delta, step.correct)
		if !within(theta, step.wantTheta) || !within(delta, step.wantDelta) {
			t.Errorf("%s: moved %v and %v, want %v and %v", step.name, theta, delta, step.wantTheta, step.wantDelta)
		}
		if !within(o.theta, step.wantOwedTheta) || !within(o.delta["a"], step.wantOwedTop) {
			t.Errorf("%s: owes %v and %v, want %v and %v", step.name, o.theta, o.delta["a"], step.wantOwedTheta, step.wantOwedTop)
		}
	}
}

// What a debt does to a level, on any run of answers. With a payment that
// clears one unit for one, the level less what it owes moves by exactly the
// step's own move, so every unit owed is charged in the end; what is owed is
// never below nothing; a right answer raises the level by at least the fifth
// of its gain the payment leaves, and never lowers it; a wrong answer never
// lowers it further than the step would; and from the third wrong answer in a
// row on, a wrong answer is the step's and owes nothing more.
func TestADebtHoldsItsProperties(t *testing.T) {
	t.Parallel()
	properties := gopter.NewProperties(nil)
	properties.Property("a debt paid one unit for one charges every unit owed and keeps its bounds", prop.ForAll(
		func(answers []debtAnswer) string {
			o := newOwing(authorsDebt.shares(), authorsDebt.repayment, authorsDebt.burn)
			run := map[string]int{}
			for i, a := range answers {
				if a.correct {
					run[a.topic] = 0
				} else {
					run[a.topic]++
				}
				if err := chargedAsItShould(o, a, run[a.topic]); err != nil {
					return fmt.Sprintf("answer %d: %v", i+1, err)
				}
			}
			return ""
		},
		gen.SliceOfN(150, genDebtAnswer()),
	))
	properties.TestingRun(t)
}

// debtAnswer is an answer as a debt is given it: its topic, how far the step
// alone would move the overall level and the topic's offset, and whether it
// was right.
type debtAnswer struct {
	topic        string
	theta, delta float64
	correct      bool
}

// genDebtAnswer is an answer in one of three topics, right seven times in ten,
// with the step's moves up for a right answer and down for a wrong one, as far
// as the service's first steps go.
func genDebtAnswer() gopter.Gen {
	topics := []string{"a", "b", "c"}
	return gopter.CombineGens(gen.IntRange(0, len(topics)-1), gen.Float64Range(0, 1), gen.Float64Range(0, 0.2), gen.Float64Range(0, 0.4)).
		Map(func(values []any) debtAnswer {
			topic, _ := values[0].(int)
			draw, _ := values[1].(float64)
			theta, _ := values[2].(float64)
			delta, _ := values[3].(float64)
			a := debtAnswer{topic: topics[topic], correct: draw < 0.7, theta: -theta, delta: -delta}
			if a.correct {
				a.theta, a.delta = theta, delta
			}
			return a
		})
}

// chargedAsItShould charges an answer, the run in its topic already counting
// it, and says which property of a debt it broke, if any.
func chargedAsItShould(o *owing, a debtAnswer, run int) error {
	owedTheta, owedTopic := o.theta, o.delta[a.topic]
	theta, delta := o.charged(a.topic, a.theta, a.delta, a.correct)
	least := 1 - authorsDebt.repayment
	switch {
	case !within(theta-(o.theta-owedTheta), a.theta) || !within(delta-(o.delta[a.topic]-owedTopic), a.delta):
		return fmt.Errorf("the level less what it owes moved %v and %v, want the step's %v and %v",
			theta-(o.theta-owedTheta), delta-(o.delta[a.topic]-owedTopic), a.theta, a.delta)
	case o.theta < 0 || o.delta[a.topic] < 0:
		return fmt.Errorf("owes %v and %v, want nothing below zero", o.theta, o.delta[a.topic])
	case a.correct && (theta < least*a.theta-1e-15 || delta < least*a.delta-1e-15):
		return fmt.Errorf("a right answer raised %v and %v, want at least a fifth of %v and %v", theta, delta, a.theta, a.delta)
	case !a.correct && (theta < a.theta-1e-15 || delta < a.delta-1e-15 || theta > 0 || delta > 0):
		return fmt.Errorf("a wrong answer moved %v and %v, want between the step's %v and %v and nothing", theta, delta, a.theta, a.delta)
	case !a.correct && run >= 3 && (!within(theta, a.theta) || !within(delta, a.delta) || o.theta != owedTheta || o.delta[a.topic] != owedTopic):
		return fmt.Errorf("the wrong answer %d in a row moved %v and %v or owes more, want the step's %v and %v", run, theta, delta, a.theta, a.delta)
	}
	return nil
}

// A payment that clears twice what it charges forgives what it does not: a
// wrong answer owes nine tenths of its charge, a right answer clears all of it
// by charging half, and the level less what it owes has gained the half that
// was never charged.
func TestAPaymentAtTwiceThePaceForgivesHalfOfWhatIsOwed(t *testing.T) {
	t.Parallel()
	o := newOwing(twiceDebt.shares(), 1, twiceDebt.burn)
	o.charged("a", -0.2, -0.4, false)
	theta, delta := o.charged("a", 1, 1, true)
	if o.theta != 0 || o.delta["a"] != 0 {
		t.Fatalf("owes %v and %v after a large right answer, want nothing", o.theta, o.delta["a"])
	}
	if !within(theta, 1-0.09) || !within(delta, 1-0.18) {
		t.Errorf("the right answer raised %v and %v, want 0.91 and 0.82: half of 0.18 and 0.36 charged", theta, delta)
	}
}

// The trial series counts a topic's run of wrong answers as the service's
// does, charging nothing: after a wrong answer of the series in a topic, the
// first wrong answer after it there is the second in its run, and takes two
// fifths of its charge.
func TestTheTrialSeriesCountsTheRunAndChargesNothing(t *testing.T) {
	t.Parallel()
	s := newStepRule(0)
	s.debt = newOwing(authorsDebt.shares(), authorsDebt.repayment, authorsDebt.burn)
	s.followed("a", 0, false, -0.5)
	if s.theta != -0.5 || s.debt.theta != 0 || s.debt.delta["a"] != 0 {
		t.Fatalf("after a wrong answer of the series: level %v, owes %v and %v; want the series' -0.5 and nothing owed", s.theta, s.debt.theta, s.debt.delta["a"])
	}
	theta, delta := s.debt.charged("a", -0.1, -0.2, false)
	if !within(theta, -0.04) || !within(delta, -0.08) {
		t.Errorf("the first wrong answer after the series moved %v and %v, want two fifths of the charge: -0.04 and -0.08", theta, delta)
	}
	s.followed("b", 0, true, -0.4)
	if s.debt.run["b"] != 0 {
		t.Errorf("a right answer of the series left a run of %d, want none", s.debt.run["b"])
	}
}

// A rule that reads mastery off its settled level reads its level less what it
// owes; every other rule reads its level as it is, and a rule that keeps no
// debt owes nothing.
func TestMasteryReadsTheSettledLevelOnlyWhereARuleSaysSo(t *testing.T) {
	t.Parallel()
	s := newStepRule(1)
	s.debt = newOwing(authorsDebt.shares(), authorsDebt.repayment, authorsDebt.burn)
	s.debt.theta, s.debt.delta["a"] = 0.1, 0.2
	settled := &session{est: s, r: &rule{masteryOnSettled: true}}
	plain := &session{est: s, r: &rule{}}
	if got := settled.masteryEstimate("a"); !within(got, 0.7) {
		t.Errorf("the settled level %v, want 1 less 0.3 owed", got)
	}
	if got := plain.masteryEstimate("a"); got != 1 {
		t.Errorf("the level of a rule that reads it as it is: %v, want 1", got)
	}
	if got := newStepRule(1).owed("a"); got != 0 {
		t.Errorf("a rule with no debt owes %v, want nothing", got)
	}
}

// The choice of a debt weighs the debts put forward and nothing else: the
// author's and the backup are its candidates, the comparisons none, and no
// debt is a candidate of the step or of mastery. The baseline, the model
// chosen, keeps the candidacy it was chosen by in the choice of mastery, which
// a run of a debt does not read.
func TestTheChoiceOfADebtWeighsItsOwnCandidates(t *testing.T) {
	t.Parallel()
	all, err := debtRules()
	if err != nil {
		t.Fatal(err)
	}
	authors, backup := mustDebtName(t, authorsDebt), mustDebtName(t, backupDebt)
	candidates := map[string]candidateKind{}
	for _, r := range all {
		if r.name != chosenMastery && (isCandidateIn(r, forStep) || isCandidateIn(r, forMastery)) {
			t.Errorf("%s is a candidate of the step or of mastery", r.name)
		}
		if isCandidateIn(r, forDebt) {
			candidates[r.name] = r.candidacy.kind
		}
	}
	if len(candidates) != 2 || candidates[authors] != mainCandidate || candidates[backup] != backupCandidate {
		t.Errorf("the candidates of a debt %v, want %s as the main one and %s as the backup", candidates, authors, backup)
	}
	if criterionOf("debt").baseline != chosenMastery {
		t.Errorf("a debt is weighed against %s, want the model chosen, %s", criterionOf("debt").baseline, chosenMastery)
	}
}

// Of the rules a run of a debt weighs, one comparison alone reads mastery off
// the level less what it owes: the author's debt under that name.
func TestOneComparisonAloneReadsMasteryOffTheSettledLevel(t *testing.T) {
	t.Parallel()
	all, err := debtRules()
	if err != nil {
		t.Fatal(err)
	}
	var settled []string
	for _, r := range all {
		if r.masteryOnSettled {
			settled = append(settled, r.name)
		}
	}
	if want := mustDebtName(t, authorsDebt) + "_settled"; len(settled) != 1 || settled[0] != want {
		t.Errorf("the rules whose mastery reads the level less what it owes: %v, want %s alone", settled, want)
	}
}

// mustDebtName is the name a debt over the model chosen takes.
func mustDebtName(t *testing.T, d debtStep) string {
	t.Helper()
	step, model, err := chosenStepAndModel()
	if err != nil {
		t.Fatal(err)
	}
	r, err := debtOver(step, model, d)
	if err != nil {
		t.Fatal(err)
	}
	return r.name
}

// A debt goes over a step rule alone: over a rule that keeps its level another
// way, it is refused rather than run as a copy of that rule under a debt's
// name.
func TestADebtOverARuleThatIsNoStepRuleIsRefused(t *testing.T) {
	t.Parallel()
	_, model, err := chosenStepAndModel()
	if err != nil {
		t.Fatal(err)
	}
	glicko := &rule{name: "glicko2_floor", shape: both, make: func(_ *child, start float64) estimator { return newGlicko(both, start) }}
	if _, err := debtOver(glicko, model, authorsDebt); !errors.Is(err, errNotAStep) {
		t.Errorf("a debt over Glicko-2: error %v, want %v", err, errNotAStep)
	}
}

// The lag of a learner is read by its size, and its check says so: where the
// baseline lags 0.1 behind, a debt that keeps the level so high it overshoots
// the child by 0.2 is 0.1 further off and fails the check, though its signed
// lag is higher; and one that lags 0.05 behind is nearer and passes it, though
// its signed lag is higher too.
func TestTheDebtReadsALearnersLagByItsSize(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		lag  float64
		want string
	}{{0.2, notMet}, {-0.05, met}} {
		baseline, candidate := &rule{name: debtBaseline, shape: both}, &rule{name: "candidate", shape: both}
		rs := []*rule{
			{name: "shrinking", shape: both, service: true}, {name: measuringRule, shape: both}, baseline, candidate,
			{name: "oracle", shape: both, ceiling: true},
		}
		cr := fakeRunReadBy(debtCriterion(), rs, func(r *rule, _ generator, metric string) float64 {
			switch {
			case metric != "r6_lag":
				return 0
			case r == candidate:
				return tc.lag
			}
			return -0.1
		})
		if got := lagVerdict(t, cr, candidate); got != tc.want {
			t.Errorf("a lag of %v against the baseline's -0.1: %s, want %s", tc.lag, got, tc.want)
		}
	}
}

// lagVerdict is the verdict of a debt's check of the lag on G2-half for a
// rule.
func lagVerdict(t *testing.T, cr *criterionRun, r *rule) string {
	t.Helper()
	for _, nw := range debtCriterion().checks() {
		if nw.metric == "r6_lag" && nw.generator == learningHalf {
			res := cr.resolution(nw)
			rd := cr.readNotWorse(r, &res)
			if rd.metric != "abs(r6_lag)" {
				t.Errorf("the lag's check is shown as %s, want abs(r6_lag), read by its size", rd.metric)
			}
			return rd.verdict
		}
	}
	t.Fatal("the criterion of a debt reads no lag on G2-half")
	return ""
}

// A debt stands from the baseline, which charges every wrong answer whole, by
// how much of a first wrong answer's charge it defers: the author's, deferring
// nine tenths, farther than the backup, deferring four fifths.
func TestADebtStandsFromTheBaselineByWhatItDefers(t *testing.T) {
	t.Parallel()
	step, model, err := chosenStepAndModel()
	if err != nil {
		t.Fatal(err)
	}
	authors, err := debtCandidate(step, model, authorsDebt, mainCandidate)
	if err != nil {
		t.Fatal(err)
	}
	backup, err := debtCandidate(step, model, backupDebt, backupCandidate)
	if err != nil {
		t.Fatal(err)
	}
	if !within(authors.candidacy.distance, 0.9) || !within(backup.candidacy.distance, 0.8) {
		t.Errorf("distances %v and %v from the baseline, want 0.9 and 0.8", authors.candidacy.distance, backup.candidacy.distance)
	}
}

// The sets of a debt draw their children: the rough look the sweep's, the
// decision run the working ones, each reading the debt's criterion; and until
// a debt is chosen, the children held out for it are run on nothing. The test
// sets the seed and the name every other test draws with, so it does not run
// beside them.
func TestTheSetsOfADebtDrawTheirChildren(t *testing.T) {
	t.Cleanup(func() { masterSeed, experiment = paperSeed, paperExperiment })
	for _, tc := range []struct {
		set, directory, experiment string
	}{
		{"debt-look", filepath.Join("results", "debt-look"), sweepExperiment},
		{"debt", filepath.Join("results", "debt"), paperExperiment},
		{"debt-confirmation", filepath.Join("results", debtHeldOutDirectory), debtHeldOutExperiment},
	} {
		out, d, err := parse([]string{"-rules", tc.set}, io.Discard)
		if err != nil || out != tc.directory || d.set != tc.set || experiment != tc.experiment {
			t.Errorf("parse of %s = %q, %+v, %v with the experiment %s; want %s and %s", tc.set, out, d, err, experiment, tc.directory, tc.experiment)
		}
		if criterionOf(tc.set).choice != forDebt {
			t.Errorf("a run of %s reads another criterion than the debt's", tc.set)
		}
	}
	if chosenDebt == "" {
		if _, err := debtConfirmationRules(); !errors.Is(err, errNoDebtChosen) {
			t.Errorf("the confirmation of a debt with none chosen: error %v, want %v", err, errNoDebtChosen)
		}
	}
}

// The criterion of a debt reads its checks on the generators it names: the
// step's two and false masteries on every one, the lag on the three learners,
// the catching up on the jump and the drop, and the placement's two on the
// child placed a level off and the one who stays put.
func TestTheCriterionOfADebtReadsItsChecks(t *testing.T) {
	t.Parallel()
	if got, want := len(debtCriterion().checks()), 3*len(allGenerators)+3+2+2+2; got != want {
		t.Errorf("%d checks of not worse, want %d", got, want)
	}
	goals := debtGoals()
	if got, want := len(goals), 2*len(screenGenerators)*len(screenWindows)+1+len(screenGoals("the baseline's")); got != want {
		t.Errorf("%d goals, want %d", got, want)
	}
}
