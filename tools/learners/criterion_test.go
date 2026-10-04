package main

import (
	"math"
	"strings"
	"testing"
)

// A value held to at most a bound meets it when it is no more than the bound;
// it reaches the bound when its whole interval does, stands on its edge when
// its interval holds the bound, and does not reach it when the whole interval
// lies past it.
func TestAGoalIsJudgedByItsValueAndMarkedByItsInterval(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		value, low, high, bound float64
		verdict, mark           string
	}{
		{0.2, 0.1, 0.3, 0.3, met, reached},
		{0.2, 0.1, 0.4, 0.3, met, onTheEdge},
		{0.3, 0.2, 0.4, 0.3, met, onTheEdge},
		{0.4, 0.3, 0.5, 0.3, notMet, onTheEdge},
		{0.4, 0.35, 0.5, 0.3, notMet, notReached},
	} {
		if verdict, mark := judged(tc.value, tc.low, tc.high, tc.bound); verdict != tc.verdict || mark != tc.mark {
			t.Errorf("%v [%v, %v] against %v: %s, %s; want %s, %s", tc.value, tc.low, tc.high, tc.bound, verdict, mark, tc.verdict, tc.mark)
		}
	}
}

// The size of a number keeps its interval's ends on one side of zero, turns
// them over on the other, and starts the interval at nothing across zero.
func TestASizeHasTheIntervalItsSizeHas(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in        summary
		low, high float64
	}{
		{summary{value: 0.3, low: 0.2, high: 0.4}, 0.2, 0.4},
		{summary{value: -0.3, low: -0.4, high: -0.2}, 0.2, 0.4},
		{summary{value: -0.1, low: -0.3, high: 0.2}, 0, 0.3},
	} {
		value, low, high := sizeOf(tc.in)
		if value != math.Abs(tc.in.value) || low != tc.low || high != tc.high {
			t.Errorf("sizeOf(%+v) = %v [%v, %v], want %v [%v, %v]", tc.in, value, low, high, math.Abs(tc.in.value), tc.low, tc.high)
		}
	}
}

// A paired difference with the service meets a check of not worse when even
// its worse end is within the tolerance, the edge included; it fails it when
// even its better end is past it; and between the two a decision run fails it
// where a rough look lets it through. For a measure where more is better, the
// worse end is the lower one.
func TestNotWorseIsReadOnTheWorseEndOfTheInterval(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		low, high      float64
		higherIsBetter bool
		decides        bool
		verdict, mark  string
	}{
		{-0.01, 0.01, false, true, met, reached},
		{0.0, 0.02, false, true, notMet, onTheEdge},
		{0.0, 0.02, false, false, met, onTheEdge},
		{0.011, 0.03, false, false, notMet, notReached},
		{-0.01, 0.02, true, true, met, reached},
		{-0.03, 0.0, true, true, notMet, onTheEdge},
		{-0.03, -0.011, true, false, notMet, notReached},
	} {
		diff := summary{value: (tc.low + tc.high) / 2, low: tc.low, high: tc.high, has: true}
		verdict, mark := notWorseVerdict(diff, tc.higherIsBetter, 0.01, tc.decides)
		if verdict != tc.verdict || mark != tc.mark {
			t.Errorf("[%v, %v], more is better %v, deciding %v: %s, %s; want %s, %s",
				tc.low, tc.high, tc.higherIsBetter, tc.decides, verdict, mark, tc.verdict, tc.mark)
		}
	}
}

// fakeRun is a run of three rules — the service, a candidate and the ceiling —
// on every generator, two alike children a cell, whose every number is what
// value gives for its rule, generator and metric.
func fakeRun(value func(r *rule, g generator, metric string) float64) *criterionRun {
	return fakeRunOf([]*rule{
		{name: "shrinking", shape: both, service: true},
		{name: "candidate", shape: both},
		{name: "oracle", shape: both, ceiling: true},
	}, value)
}

// fakeRunOf is a run of these rules on every generator, two alike children a
// cell, whose every number is what value gives for its rule, generator and
// metric, read by the step's criterion.
func fakeRunOf(rs []*rule, value func(r *rule, g generator, metric string) float64) *criterionRun {
	return fakeRunReadBy(stepCriterion(), rs, value)
}

// fakeRunReadBy is fakeRunOf read by a criterion.
func fakeRunReadBy(crit *criterion, rs []*rule, value func(r *rule, g generator, metric string) float64) *criterionRun {
	ms := metrics()
	names := metricNames(ms)
	all := cellsOf(rs)
	summaries := make([][]summary, len(all))
	results := make([][]vector, len(all))
	for c := range all {
		summaries[c] = make([]summary, len(names))
		results[c] = []vector{{parts: make([]part, len(ms))}, {parts: make([]part, len(ms))}}
		for i, name := range names {
			v := value(all[c].rule, all[c].generator, name)
			summaries[c][i] = summary{value: v, low: v, high: v, has: true}
			if i < len(ms) {
				for child := range results[c] {
					results[c][child].parts[i] = part{a: v, b: 1}
				}
			}
		}
	}
	return newCriterionRun(all, summaries, results, ms, decisionChildren, crit)
}

// badnessAt is a value of a fake run in which every measure the score reads
// is 0.5 bad under the service, 0.1 under the ceiling, and as bad under the
// candidate as given, the corridor read as its shortfall and the lag as behind.
func badnessAt(candidate float64) func(r *rule, g generator, metric string) float64 {
	return func(r *rule, _ generator, metric string) float64 {
		bad := candidate
		switch {
		case r.service:
			bad = 0.5
		case r.ceiling:
			bad = 0.1
		}
		switch metric {
		case "r3_inside":
			return 1 - bad
		case "r6_lag":
			return -bad
		}
		return bad
	}
}

// The score is the share of the way from the service to the ceiling a rule
// closes: none for a rule as bad as the service, all of it for one as good as
// the ceiling, half for one halfway, and less than none for one worse than
// the service.
func TestTheScoreIsTheShareOfTheWayClosed(t *testing.T) {
	t.Parallel()
	for candidate, want := range map[float64]float64{0.5: 0, 0.1: 1, 0.3: 0.5, 0.7: -0.5} {
		cr := fakeRun(badnessAt(candidate))
		score := cr.scoreOf(cr.all[len(allGenerators)].rule)
		if !score.has || math.Abs(score.value-want) > 1e-12 {
			t.Errorf("a candidate as bad as %v scores %+v, want %v", candidate, score, want)
		}
	}
}

// A score whose way to the ceiling is no way at all — the ceiling no better
// than the service on one of its measures — is not read.
func TestAScoreWithNoWayIsNotRead(t *testing.T) {
	t.Parallel()
	ceilingNoBetter := func(r *rule, g generator, metric string) float64 {
		if r.ceiling && g == staticChildren && metric == "r1_rms_200" {
			return 0.5
		}
		return badnessAt(0.3)(r, g, metric)
	}
	cr := fakeRun(ceilingNoBetter)
	if score := cr.scoreOf(cr.all[len(allGenerators)].rule); score.has {
		t.Errorf("a score with no way on G0's error = %+v, want none", score)
	}
}

// A check of not worse the run cannot read is listed as unread beside the
// checks a rule does not meet, never with numbers it does not have.
func TestAnUnreadCheckIsListedAsUnread(t *testing.T) {
	t.Parallel()
	rc := ruleCriterion{
		rule: &rule{name: "candidate", shape: both},
		readings: []reading{
			{kind: "not worse", generator: staticChildren, metric: "r4_false", verdict: unread},
			{kind: "not worse", generator: learning, metric: "r1_rms_200", value: 0.1, low: 0.05, high: 0.15, bound: 0.02, verdict: notMet},
			{kind: "not worse", generator: jumping, metric: "r3_inside", verdict: met},
		},
	}
	var b strings.Builder
	writeNotWorse(&b, []ruleCriterion{rc})
	want := "| candidate/both | 1 of 3 | G0 r4_false unread; G2 r1_rms_200 0.100 [0.050, 0.150] against 0.020 |"
	if !strings.Contains(b.String(), want) {
		t.Errorf("the table of not worse is\n%s\nwant the row %s", b.String(), want)
	}
}

// A score reads only measures that mean something under the rule: a rule that
// gives no error of placement of its own — as one that keeps no overall level
// does not — has no score, rather than one that counts the child's distance
// from the start as the rule's.
func TestAScoreReadsOnlyTheRulesOwnMeasures(t *testing.T) {
	t.Parallel()
	cr := fakeRun(badnessAt(0.3))
	candidate := cr.all[len(allGenerators)].rule
	c, _ := cr.cellOf(candidate, misplaced)
	cr.summaries[c][cr.metricAt["r7_error_10"]].has = false
	if score := cr.scoreOf(candidate); score.has {
		t.Errorf("a candidate with no error of placement of its own scores %+v, want no score", score)
	}
}

// A goal holds a rule to a bound the service sets: on the main learner, a lag
// of at most 45 % of the service's, read as its size.
func TestAGoalHoldsARuleToTheServicesBound(t *testing.T) {
	t.Parallel()
	lagGoal := stepGoals()[0]
	for candidate, want := range map[float64]string{0.2: met, 0.3: notMet} {
		cr := fakeRun(badnessAt(candidate))
		rd := cr.readGoal(cr.all[len(allGenerators)].rule, lagGoal)
		if rd.verdict != want || math.Abs(rd.bound-0.225) > 1e-12 || rd.value != candidate {
			t.Errorf("a lag of %v against the service's 0.5: %s with %v against %v, want %s against 0.225",
				-candidate, rd.verdict, rd.value, rd.bound, want)
		}
	}
}

// A candidate exactly as good as the service, in a decision run, passes a
// check read with a tolerance of 4.3 standard errors 99 times in 100, one read
// with the interval's own half-width half the time, and one with no tolerance
// at all 2.5 times in 100; in a rough look, which fails it only when the
// better end of the interval is past the tolerance, it passes that last one
// 97.5 times in 100.
func TestAnEqualCandidatePassesAsTheToleranceAllows(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		tolerance float64
		decides   bool
		want      float64
	}{
		{resolutionWidths, true, 0.990},
		{normalQuantile, true, 0.5},
		{0, true, 0.025},
		{0, false, 0.975},
	} {
		if got := passChance(tc.tolerance, 1, tc.decides); math.Abs(got-tc.want) > 0.001 {
			t.Errorf("pass chance at %v standard errors, deciding %v = %v, want %v", tc.tolerance, tc.decides, got, tc.want)
		}
	}
}
