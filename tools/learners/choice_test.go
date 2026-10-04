package main

import (
	"math"
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
func choiceOf(t *testing.T, badness map[string]float64, rs ...*rule) (stepChoice, *criterionRun) {
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
