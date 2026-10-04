package main

import (
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
)

// candidate is a rule of a fake run that is a candidate, of a kind, adding so
// many numbers, fields or not, at a distance from the service's step.
func candidate(name string, kind candidateKind, added int, fields bool, distance float64) *rule {
	return &rule{name: name, shape: both, candidacy: &candidacy{kind: kind, added: added, fields: fields, distance: distance}}
}

// choiceOf is the choice a fake run of the service, the given rules and the
// ceiling comes to, every rule as bad as badness gives — 0.5 the service's,
// 0.1 the ceiling's — and every candidate meeting every constraint at a
// badness of 0.22 or less.
func choiceOf(t *testing.T, badness map[string]float64, rs ...*rule) (criterionChoice, *criterionRun) {
	t.Helper()
	all := append([]*rule{{name: "shrinking", shape: both, service: true}}, rs...)
	all = append(all, &rule{name: "oracle", shape: both, ceiling: true})
	cr := fakeRunOf(all, func(r *rule, g generator, metric string) float64 {
		if bad, given := badness[r.name]; given {
			return badAt(bad)(r, g, metric)
		}
		return badnessAt(0.5)(r, g, metric)
	})
	read, _ := readCriterion(cr)
	c := cr.choose(read, false)
	return c, cr
}

// badAt is badnessAt for a rule that is as bad as given whatever it is.
func badAt(bad float64) func(r *rule, g generator, metric string) float64 {
	return func(_ *rule, g generator, metric string) float64 {
		return badnessAt(bad)(&rule{}, g, metric)
	}
}

func nameOf(rc *ruleCriterion) string {
	if rc == nil {
		return "nothing"
	}
	return rc.rule.name
}

// A* is the main step of the highest score among those that meet every
// constraint and are better than the service; its equals are the main steps
// whose score less A*'s holds nothing; and the step chosen is the simplest of
// them — fewer numbers added first, then no field in the profile, then the
// nearer the service's step.
func TestTheSimplestOfAStarAndItsEqualsIsChosen(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		rules []*rule
		want  string
	}{
		{"fewer numbers", []*rule{candidate("best", mainCandidate, 3, false, 0.1), candidate("equal", mainCandidate, 1, false, 0.5)}, "equal"},
		{"no fields", []*rule{candidate("best", mainCandidate, 1, true, 0.1), candidate("equal", mainCandidate, 1, false, 0.5)}, "equal"},
		{"nearer", []*rule{candidate("best", mainCandidate, 1, false, 0.5), candidate("equal", mainCandidate, 1, false, 0.1)}, "equal"},
		{"A* the simplest", []*rule{candidate("best", mainCandidate, 1, false, 0.1), candidate("equal", mainCandidate, 2, false, 0.1)}, "best"},
		{"worse is no equal", []*rule{candidate("best", mainCandidate, 3, false, 0.1), candidate("worse", mainCandidate, 0, false, 0)}, "best"},
	} {
		badness := map[string]float64{"best": 0.15, "equal": 0.15, "worse": 0.2}
		c, _ := choiceOf(t, badness, tc.rules...)
		if nameOf(c.chosen) != tc.want || c.exit {
			t.Errorf("%s: chose %s (the exit %v), want %s", tc.name, nameOf(c.chosen), c.exit, tc.want)
		}
	}
}

// A candidate that misses a constraint, or is no better than the service, is
// not chosen, however simple.
func TestACandidateThatMissesAConstraintIsNotChosen(t *testing.T) {
	t.Parallel()
	c, _ := choiceOf(t, map[string]float64{"good": 0.15, "lagging": 0.3},
		candidate("good", mainCandidate, 3, false, 1), candidate("lagging", mainCandidate, 0, false, 0))
	if nameOf(c.chosen) != "good" || len(c.eligible) != 1 {
		t.Errorf("chose %s of %d eligible, want good, the one eligible", nameOf(c.chosen), len(c.eligible))
	}
}

// The backup is chosen over the main step only where its score less the main
// step's lies above the margin with its whole interval, and where no main
// step meets the constraints it is chosen on its own.
func TestTheBackupIsChosenOnlyByTheMargin(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		badness map[string]float64
		rules   []*rule
		want    string
	}{
		{"within the margin", map[string]float64{"main": 0.15, "backup": 0.14},
			[]*rule{candidate("main", mainCandidate, 1, false, 0), candidate("backup", backupCandidate, 1, true, 0)}, "main"},
		{"past the margin", map[string]float64{"main": 0.2, "backup": 0.1},
			[]*rule{candidate("main", mainCandidate, 1, false, 0), candidate("backup", backupCandidate, 1, true, 0)}, "backup"},
		{"no main step", map[string]float64{"main": 0.4, "backup": 0.15},
			[]*rule{candidate("main", mainCandidate, 1, false, 0), candidate("backup", backupCandidate, 1, true, 0)}, "backup"},
	} {
		c, _ := choiceOf(t, tc.badness, tc.rules...)
		if nameOf(c.chosen) != tc.want {
			t.Errorf("%s: chose %s, want %s", tc.name, nameOf(c.chosen), tc.want)
		}
	}
}

// With no candidate to choose, the exit is the floor of the highest score
// among those that meet the constraints of not worse and of the screen, the
// goals left aside; with none of those, nothing is chosen and the service's
// step stays.
func TestTheExitIsTheBestFloorThatIsNotWorse(t *testing.T) {
	t.Parallel()
	floors := []*rule{candidate("floor_low", exitCandidate, 1, false, 0.01), candidate("floor_high", exitCandidate, 1, false, 0.05)}
	c, _ := choiceOf(t, map[string]float64{"floor_low": 0.45, "floor_high": 0.4}, floors...)
	if nameOf(c.chosen) != "floor_high" || !c.exit {
		t.Errorf("chose %s (the exit %v), want floor_high as the exit", nameOf(c.chosen), c.exit)
	}
	c, _ = choiceOf(t, map[string]float64{"floor_low": 0.6, "floor_high": 0.7}, floors...)
	if c.chosen != nil {
		t.Errorf("with every floor worse than the service chose %s, want nothing", nameOf(c.chosen))
	}
	var b strings.Builder
	writeChoice(&b, &c)
	if !strings.Contains(b.String(), "the service's step stays") {
		t.Errorf("the choice of nothing reads\n%s\nwant it to say the service's step stays", b.String())
	}
}

// One rule's score less another's, read on the same children, is the
// difference of their scores.
func TestTheDifferenceOfTwoScoresIsReadOnTheSameChildren(t *testing.T) {
	t.Parallel()
	_, cr := choiceOf(t, map[string]float64{"a": 0.1, "b": 0.3},
		candidate("a", mainCandidate, 0, false, 0), candidate("b", mainCandidate, 0, false, 0))
	a, b := cr.all[len(allGenerators)].rule, cr.all[2*len(allGenerators)].rule
	d := cr.scoreDifference(a, b)
	if want := cr.scoreOf(a).value - cr.scoreOf(b).value; !d.has || math.Abs(d.value-want) > 1e-12 || math.Abs(want-0.5) > 1e-12 {
		t.Errorf("a's score less b's is %+v, want %v, which is 0.5", d, want)
	}
}

// A run of the held-out children confirms its candidate when it meets every
// constraint and is better than the service there, and says the exit is taken
// when it does not.
func TestTheConfirmationSaysWhetherTheCandidateHolds(t *testing.T) {
	t.Parallel()
	for bad, want := range map[float64]string{0.15: "**Confirmed:**", 0.4: "**Not confirmed:**"} {
		all := []*rule{{name: "shrinking", shape: both, service: true}, candidate("chosen", mainCandidate, 1, false, 0), {name: "oracle", shape: both, ceiling: true}}
		cr := fakeRunOf(all, func(r *rule, g generator, metric string) float64 {
			if r.candidacy != nil {
				return badAt(bad)(r, g, metric)
			}
			return badnessAt(0.5)(r, g, metric)
		})
		read, _ := readCriterion(cr)
		c := cr.choose(read, true)
		var b strings.Builder
		writeChoice(&b, &c)
		if !strings.Contains(b.String(), want) {
			t.Errorf("a candidate as bad as %v: the confirmation reads\n%s\nwant %s", bad, b.String(), want)
		}
	}
}

// A constraint met with room to spare holds on new children almost surely,
// one met on its edge half the time, and one missed by as much almost never;
// a check of not worse has to keep the worse end of its interval within the
// tolerance, and so holds less often than its value alone would.
func TestTheChanceAConstraintHoldsGrowsWithItsRoom(t *testing.T) {
	t.Parallel()
	se := 0.01
	for _, tc := range []struct {
		name string
		rd   reading
		want float64
	}{
		{"room", reading{kind: goalKind, value: 0.1, low: 0.1 - normalQuantile*se, high: 0.1 + normalQuantile*se, bound: 0.2}, 1},
		{"edge", reading{kind: goalKind, value: 0.2, low: 0.2 - normalQuantile*se, high: 0.2 + normalQuantile*se, bound: 0.2}, 0.5},
		{"missed", reading{kind: goalKind, value: 0.3, low: 0.3 - normalQuantile*se, high: 0.3 + normalQuantile*se, bound: 0.2}, 0},
		{"at least", reading{kind: goalKind, atLeast: true, value: 0.3, low: 0.3 - normalQuantile*se, high: 0.3 + normalQuantile*se, bound: 0.2}, 1},
		{"not worse on its edge", reading{kind: notWorseKind, value: 0.02 - normalQuantile*se, low: 0.02 - 2*normalQuantile*se, high: 0.02, bound: 0.02}, 0.5},
		{"more is better", reading{kind: notWorseKind, higherIsBetter: true, value: -0.02 + normalQuantile*se, low: -0.02, high: -0.02 + 2*normalQuantile*se, bound: 0.02}, 0.5},
	} {
		if got := holdChance(&tc.rd); math.Abs(got-tc.want) > 1e-3 {
			t.Errorf("%s: the chance it holds is %v, want %v", tc.name, got, tc.want)
		}
	}
}

// The exit is taken without a confirmation and has no parts to take apart:
// the runs that follow the choice refuse it, and take a candidate as it is,
// so that the held-out children stay held out for a candidate.
func TestTheExitIsNeitherConfirmedNorTakenApart(t *testing.T) {
	t.Parallel()
	if _, err := confirmable(floorRule(0.05)); !errors.Is(err, errExitTaken) {
		t.Errorf("confirming a floor: error %v, want %v", err, errExitTaken)
	}
	main := uncertainStep{afterSeries: 0.4, topicSpread: 0.5, limit: noLimit}.rule()
	if got, err := confirmable(main); err != nil || got != main {
		t.Errorf("confirming a main step: %v, %v; want it as it is", got, err)
	}
}

// The chance a chosen candidate's constraints hold on new children is read
// for every one of them, before its confirmation draws the children; the exit
// is taken without a confirmation, and none is read for it.
func TestOnlyACandidateHasItsChancesOnNewChildrenRead(t *testing.T) {
	t.Parallel()
	c, _ := choiceOf(t, map[string]float64{"floor": 0.4}, candidate("floor", exitCandidate, 1, false, 0))
	if !c.exit || len(c.margins) != 0 {
		t.Errorf("chose %s (the exit %v) with %d chances read, want the floor as the exit, with none", nameOf(c.chosen), c.exit, len(c.margins))
	}
	c, _ = choiceOf(t, map[string]float64{"main": 0.15}, candidate("main", mainCandidate, 1, false, 0))
	if nameOf(c.chosen) != "main" || len(c.margins) != len(c.chosen.readings) {
		t.Errorf("chose %s with %d chances read, want main with one for each of its constraints", nameOf(c.chosen), len(c.margins))
	}
}

// In a decision run, the main step of the highest score, whatever either
// meets, is set against every rule but the service and the ceiling, by its
// score less theirs on the same children; the table says so, a row a rule.
func TestEveryRuleIsSetAgainstTheHighestMainStep(t *testing.T) {
	t.Parallel()
	c, cr := choiceOf(t, map[string]float64{"top": 0.3, "part": 0.4, "comparison": 0.35},
		candidate("top", mainCandidate, 1, false, 0), candidate("part", mainCandidate, 0, false, 0), &rule{name: "comparison", shape: both})
	if nameOf(c.top) != "top" || len(c.againstTop) != 2 {
		t.Fatalf("the highest main step %s, set against %d rules; want top, against 2", nameOf(c.top), len(c.againstTop))
	}
	for _, r := range c.againstTop {
		if want := cr.scoreOf(c.top.rule).value - cr.scoreOf(r.rc.rule).value; !r.difference.has || math.Abs(r.difference.value-want) > 1e-12 {
			t.Errorf("the top less %s: %+v, want %v", r.rc.rule.name, r.difference, want)
		}
	}
	var b strings.Builder
	writeChoice(&b, &c)
	if want := "| comparison/both | 0.375 [0.375, 0.375] | 0.125 [0.125, 0.125] |"; !strings.Contains(b.String(), want) {
		t.Errorf("the choice reads\n%s\nwant the row %s", b.String(), want)
	}
}

// The step chosen is a rule of the decision run, and of the bench's own set:
// a name that matched no rule would drop it from both without a word.
func TestTheChosenStepIsARuleOfTheRuns(t *testing.T) {
	t.Parallel()
	chosen, err := chosenRule()
	if err != nil || chosen.name != chosenStep {
		t.Fatalf("the chosen step %q: %v, %v; want a rule of the decision run", chosenStep, chosen, err)
	}
	if !containsRule(benchRules(), chosen) {
		t.Errorf("the bench's own set lacks the chosen step %s", chosenStep)
	}
}

// The model of mastery chosen is a model of the decision run of mastery, of
// the bench's own set and of the confirmation, which reads it by the
// criterion of mastery: a name that matched no model would drop it from all
// three without a word.
func TestTheChosenModelIsARuleOfTheRuns(t *testing.T) {
	t.Parallel()
	model, err := chosenModel()
	if chosenMastery == "" {
		if !errors.Is(err, errNoMasteryChosen) {
			t.Errorf("no model of mastery chosen: error %v, want %v", err, errNoMasteryChosen)
		}
		return
	}
	if err != nil || model.name != chosenMastery {
		t.Fatalf("the chosen model %q: %v, %v; want a model of the decision run of mastery", chosenMastery, model, err)
	}
	if !containsRule(benchRules(), model) {
		t.Errorf("the bench's own set lacks the chosen model %s", chosenMastery)
	}
	if confirmed, err := confirmationRules(); err != nil || !containsRule(confirmed, model) {
		t.Errorf("the confirmation's rules lack the chosen model %s: error %v", chosenMastery, err)
	}
	if confirmationCriterion().choice != forMastery {
		t.Errorf("the confirmation reads the step's criterion, want mastery's")
	}
}

// masteryChoiceOf is the choice a fake run read by the criterion of mastery
// comes to: the service as bad as 0.9, the baseline — the step chosen under
// the service's mastery — as bad as 0.5, the ceiling 0.1, and every candidate
// as bad as badness gives; a candidate meets every constraint at a badness of
// 0.2 or less.
func masteryChoiceOf(t *testing.T, badness map[string]float64, rs ...*rule) (criterionChoice, *criterionRun) {
	t.Helper()
	all := append([]*rule{{name: "shrinking", shape: both, service: true}, {name: masteryBaseline, shape: both}}, rs...)
	all = append(all, &rule{name: "oracle", shape: both, ceiling: true})
	cr := fakeRunReadBy(masteryCriterion(), all, func(r *rule, g generator, metric string) float64 {
		switch bad, given := badness[r.name]; {
		case given:
			return badAt(bad)(r, g, metric)
		case r.name == masteryBaseline:
			return badAt(0.5)(r, g, metric)
		case r.service:
			return badAt(0.9)(r, g, metric)
		}
		return badnessAt(0.5)(r, g, metric)
	})
	read, _ := readCriterion(cr)
	c := cr.choose(read, false)
	return c, cr
}

// masteryCandidateRule is a rule of a fake run put forward in the choice of
// mastery.
func masteryCandidateRule(name string, kind candidateKind, added int, fields bool) *rule {
	return &rule{name: name, shape: both, candidacy: &candidacy{choice: forMastery, kind: kind, added: added, fields: fields}}
}

// The choice of mastery compares with the baseline, not with the service: a
// candidate much better than the service but worse than the baseline is no
// candidate to choose; its score is the share of the way from the baseline to
// perfect, which is nothing, rather than to the ceiling; and with no candidate
// the exit is the baseline itself, kept as it is though it meets none of the
// goals, the false masteries of half its declarations among them.
func TestTheChoiceOfMasteryComparesWithTheBaseline(t *testing.T) {
	t.Parallel()
	c, cr := masteryChoiceOf(t, map[string]float64{"worse": 0.6, "halfway": 0.25},
		masteryCandidateRule("worse", mainCandidate, 0, false), masteryCandidateRule("halfway", mainCandidate, 1, false))
	if cr.baseline.name != masteryBaseline {
		t.Fatalf("the choice of mastery compares with %s, want %s", cr.baseline.name, masteryBaseline)
	}
	if score := cr.scoreOf(cr.all[2*len(allGenerators)].rule); !score.has || math.Abs(score.value-(-0.2)) > 1e-12 {
		t.Errorf("a candidate as bad as 0.6 against the baseline's 0.5 scores %+v, want −0.2 of the way to perfect", score)
	}
	if score := cr.scoreOf(cr.all[3*len(allGenerators)].rule); !score.has || math.Abs(score.value-0.5) > 1e-12 {
		t.Errorf("a candidate as bad as 0.25 against the baseline's 0.5 scores %+v, want half the way to perfect", score)
	}
	if nameOf(c.chosen) != masteryBaseline || !c.exit || len(c.margins) != 0 {
		t.Errorf("chose %s (the exit %v, %d chances read), want the baseline as the exit, with none", nameOf(c.chosen), c.exit, len(c.margins))
	}
	var b strings.Builder
	writeChoice(&b, &c)
	if want := "**The exit: " + masteryBaseline + "/both**, the baseline itself, kept as it is."; !strings.Contains(b.String(), want) {
		t.Errorf("the choice reads\n%s\nwant %s", b.String(), want)
	}
}

// In the choice of mastery, as in the step's, the simplest of the best and its
// equals is chosen, and the backup — Wald's test, which keeps its sums in the
// profile — only by the margin.
func TestTheChoiceOfMasteryIsMadeAsTheStepsIs(t *testing.T) {
	t.Parallel()
	c, _ := masteryChoiceOf(t, map[string]float64{"run5": 0.15, "cautious": 0.15, "wald": 0.14},
		masteryCandidateRule("cautious", mainCandidate, 1, false), masteryCandidateRule("run5", mainCandidate, 0, false),
		masteryCandidateRule("wald", backupCandidate, 2, true))
	if nameOf(c.chosen) != "run5" || c.exit {
		t.Errorf("chose %s (the exit %v), want run5, the simplest of the equals, the backup within the margin", nameOf(c.chosen), c.exit)
	}
}

// The report of the choice of mastery names the baseline wherever it compares
// with it — its checks of not worse, what the bench resolves on them, and the
// score, whose way ends at perfect — and measures its candidates' distances
// from the service's rule of mastery, Wald's test, which stands apart from it,
// at none: nothing in it reads as compared with the service or its step.
func TestTheReportOfMasteryNamesTheBaseline(t *testing.T) {
	t.Parallel()
	models := map[string]*rule{}
	for _, mc := range masteryCandidates() {
		models[mc.name] = modelOf(&rule{name: masteryBaseline, shape: both}, mc)
	}
	run5, wald := models["run5"], models["wald"]
	c, cr := masteryChoiceOf(t, map[string]float64{run5.name: 0.15, wald.name: 0.14}, run5, wald)
	read, resolutions := readCriterion(cr)
	text := criterionText(read, resolutions, &c, design{children: decisionChildren, answers: 200, set: "mastery"}, cr.crit)
	for _, want := range []string{
		"## Not worse than the baseline", "its difference from the baseline,", "between " + measuringRule + " and the baseline,",
		"as good as the baseline passes", "| Distance from the service's rule of mastery |",
		"a distance of 0.000 from the service's rule of mastery.", "| 2 | yes | — |",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the report of mastery lacks %q:\n%s", want, text)
		}
	}
	for _, wrong := range []string{"than the service", "from the service,", "and the service,", "as good as the service", "the service's step"} {
		if strings.Contains(text, wrong) {
			t.Errorf("the report of mastery says %q, as if compared with the service:\n%s", wrong, text)
		}
	}
	named := map[string][]string{}
	for _, row := range criterionTable(read, cr.crit)[1:] {
		if kind, name := row[2], row[3]; !slices.Contains(named[kind], name) {
			named[kind] = append(named[kind], name)
		}
	}
	if got, want := named[notWorseKind], []string{"not worse than the baseline"}; !slices.Equal(got, want) {
		t.Errorf("the checks of not worse are named %q, want %q", got, want)
	}
	if got, want := named["score"], []string{"the share of the way to perfect"}; !slices.Equal(got, want) {
		t.Errorf("the scores are named %q, want %q", got, want)
	}
}

// A choice of mastery whose run lacks the baseline chooses nothing, and says
// so in its own terms: its exit is the baseline, which is not there to keep,
// rather than a floor under the step.
func TestAChoiceOfMasteryWithoutItsBaselineChoosesNothing(t *testing.T) {
	t.Parallel()
	rs := []*rule{
		{name: "shrinking", shape: both, service: true}, masteryCandidateRule("cautious", mainCandidate, 1, false),
		{name: "oracle", shape: both, ceiling: true},
	}
	cr := fakeRunReadBy(masteryCriterion(), rs, badAt(0.15))
	read, _ := readCriterion(cr)
	c := cr.choose(read, false)
	if c.chosen != nil {
		t.Fatalf("chose %s with no baseline in the run, want nothing", nameOf(c.chosen))
	}
	var b strings.Builder
	writeChoice(&b, &c)
	if text := b.String(); !strings.Contains(text, "the run lacks the baseline, which is the exit: **nothing is chosen.**") ||
		strings.Contains(text, "the service's step stays") {
		t.Errorf("the choice reads\n%s\nwant it to say the baseline is missing and nothing is chosen", text)
	}
}

// The choice of mastery takes the price of a rule that declares only what it
// is sure of: a candidate worse than the baseline on the error, and on
// masteries never declared on the widest spread of topics, within that price
// is chosen; one past either, or past the narrower tolerance of masteries
// never declared on the child who stays put, is not.
func TestTheChoiceOfMasteryTakesItsPrice(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		worseError float64
		neverOn    generator
		moreNever  float64
		chosen     bool
	}{
		{"within the price", masteryErrorTolerance - 0.005, farTopics, wideNeverTolerance - 0.01, true},
		{"the error past it", masteryErrorTolerance + 0.005, farTopics, 0, false},
		{"masteries never declared past it on the widest spread", 0, farTopics, wideNeverTolerance + 0.01, false},
		{"masteries never declared past the tolerance on the child who stays put", 0, staticChildren, falseTolerance + 0.01, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			candidate := masteryCandidateRule("cautious", mainCandidate, 1, false)
			rs := []*rule{{name: masteryBaseline, shape: both}, candidate}
			cr := fakeRunReadBy(masteryCriterion(), rs, func(r *rule, g generator, metric string) float64 {
				baseline := badAt(0.5)(r, g, metric)
				switch {
				case r != candidate:
					return baseline
				case metric == "r1_rms_200":
					return baseline + tc.worseError
				case metric == "r5_never" && g == tc.neverOn:
					return baseline + tc.moreNever
				}
				return badAt(0.15)(r, g, metric)
			})
			read, _ := readCriterion(cr)
			c := cr.choose(read, false)
			if chosen := nameOf(c.chosen) == candidate.name && !c.exit; chosen != tc.chosen {
				t.Errorf("chose %s (the exit %v), want the candidate chosen %v", nameOf(c.chosen), c.exit, tc.chosen)
			}
		})
	}
}

// The masteries never declared are read on the children who stay put alone,
// where the goals of mastery are read; every other measure of not worse, on
// every generator, and the step's checks as they were.
func TestMasteriesNeverDeclaredAreReadWhereTheGoalsOfMasteryAre(t *testing.T) {
	t.Parallel()
	read := map[string][]generator{}
	for _, nw := range masteryCriterion().checks() {
		read[nw.metric] = append(read[nw.metric], nw.generator)
	}
	if !slices.Equal(read["r5_never"], masteryGenerators) {
		t.Errorf("masteries never declared are read on %v, want %v", read["r5_never"], masteryGenerators)
	}
	for _, metric := range []string{"r1_rms_200", "r3_inside", "r4_false"} {
		if !slices.Equal(read[metric], allGenerators) {
			t.Errorf("%s is read on %v, want every generator", metric, read[metric])
		}
	}
	if got, want := len(stepCriterion().checks()), 2*len(allGenerators); got != want {
		t.Errorf("the step has %d checks of not worse, want %d", got, want)
	}
}
