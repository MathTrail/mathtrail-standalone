package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"testing"
)

// ruleNamed is the rule of the bench of that name.
func ruleNamed(t *testing.T, name string) *rule {
	t.Helper()
	for _, r := range benchRules() {
		if r.name == name {
			return r
		}
	}
	t.Fatalf("no rule named %s", name)
	return nil
}

// Every generator added to the nine departs from the base population in its
// one respect: the model's miss, the speed and the steadiness of learning, or
// a drop instead of a jump.
func TestTheAddedGeneratorsDepartInTheirOneRespect(t *testing.T) {
	t.Parallel()
	topics := []string{"a", "b", "c"}
	for gen, want := range map[generator]float64{staticChildren: writingError, exactlyAsWritten: 0, smallMisses: 0.25, largeMisses: 1} {
		if got := newChild(gen, 0, topics).writing; got != want {
			t.Errorf("%s: a miss of %v, want %v", gen, got, want)
		}
	}
	fast, half, fading := newChild(learning, 0, topics), newChild(learningHalf, 0, topics), newChild(learningFades, 0, topics)
	if !half.learns || half.fades || half.learnRate != fast.learnRate/2 || half.topicRate != fast.topicRate/2 {
		t.Errorf("a child of G2-half learns by %v and %v, fading %v; want half of G2's %v and %v, steadily",
			half.learnRate, half.topicRate, half.fades, fast.learnRate, fast.topicRate)
	}
	if !fading.learns || !fading.fades || fast.fades || fading.learnRate != fast.learnRate || fading.topicRate != fast.topicRate {
		t.Errorf("a child of G2-fading learns by %v and %v, fading %v; want G2's %v and %v, fading", fading.learnRate, fading.topicRate,
			fading.fades, fast.learnRate, fast.topicRate)
	}
	for i := range 20 {
		jump, drop := newChild(jumping, i, topics), newChild(dropping, i, topics)
		if jump.jump != jumpBy || drop.jump != -jumpBy || drop.jumpAt < jumpEarliest || drop.jumpAt > jumpLatest {
			t.Errorf("child %d: a jump of %v and a drop of %v at answer %d, want %v and %v from answer %d to %d",
				i, jump.jump, drop.jump, drop.jumpAt, jumpBy, -jumpBy, jumpEarliest, jumpLatest)
		}
	}
}

// A child whose learning fades gains G2's step at first, half of it at the
// 50th answer, and a fifth of it at the 200th.
func TestFadingLearningSlowsAsTheAnswersGoBy(t *testing.T) {
	t.Parallel()
	c := newChild(learningFades, 0, []string{"a", "b"})
	for _, at := range []struct {
		k     int
		share float64
	}{{0, 1}, {50, 0.5}, {200, 0.2}} {
		before := c.theta
		c.after("a", at.k)
		if got := (c.theta - before) / learningPerAnswer; math.Abs(got-at.share) > 1e-12 {
			t.Errorf("after answer %d the level grew by %v of G2's step, want %v", at.k+1, got, at.share)
		}
	}
}

// The children of the generators that move a spread stand as far around
// their start, and their topics as far around their level, as the generator
// says: over two thousand children the spread comes within a tenth of it, and
// what a topic began at is what it was drawn at.
func TestSpreadsAreTheGeneratorsOwn(t *testing.T) {
	t.Parallel()
	topics := []string{"a", "b", "c", "d"}
	for _, tc := range []struct {
		gen                   generator
		aroundStart, ofTopics float64
	}{
		{staticChildren, 1, 0.5}, {narrowStart, 0.5, 0.5}, {wideStart, 2, 0.5}, {closeTopics, 1, 0.3}, {farTopics, 1, 1},
	} {
		var levels, offsets []float64
		for i := range 2000 {
			c := newChild(tc.gen, i, topics)
			levels = append(levels, c.theta-c.start)
			for _, topic := range topics {
				offsets = append(offsets, c.delta[topic])
				if c.first[topic] != c.delta[topic] {
					t.Fatalf("%s child %d began %s at %v, drawn at %v", tc.gen, i, topic, c.first[topic], c.delta[topic])
				}
			}
		}
		if got := spreadOf(levels); math.Abs(got/tc.aroundStart-1) > 0.1 {
			t.Errorf("%s: children spread around their start by %.3f, want %v", tc.gen, got, tc.aroundStart)
		}
		if got := spreadOf(offsets); math.Abs(got/tc.ofTopics-1) > 0.1 {
			t.Errorf("%s: topics spread around the level by %.3f, want %v", tc.gen, got, tc.ofTopics)
		}
	}
}

// spreadOf is the root mean square of values around zero.
func spreadOf(values []float64) float64 {
	squares := 0.0
	for _, v := range values {
		squares += v * v
	}
	return math.Sqrt(squares / float64(len(values)))
}

// The oracle stands where the child stands: the error of its level, of its
// overall level after five and ten answers, and its lag are exactly nothing on
// every generator, the children who learn, jump and drop among them.
func TestTheOracleStandsWhereTheChildDoes(t *testing.T) {
	t.Parallel()
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	oracleRule := ruleNamed(t, "oracle")
	for _, gen := range allGenerators {
		for i := range 3 {
			r, err := run(w, oracleRule, newChild(gen, i, w.topics))
			if err != nil {
				t.Fatal(err)
			}
			wantNoError(t, fmt.Sprintf("%s child %d", gen, i), r)
		}
	}
}

// wantNoError holds a run of a child to no error at all: of its level at every
// checkpoint, of its overall level after five and ten answers, and in its lag
// over the answers the lag is read on.
func wantNoError(t *testing.T, whose string, r *childResult) {
	t.Helper()
	for at, e := range r.errorRMS {
		if e != 0 || r.errorMean[at] != 0 {
			t.Errorf("%s: after %d answers an error of %v and %v, want none", whose, at, e, r.errorMean[at])
		}
	}
	for at, off := range r.placed {
		if off != 0 {
			t.Errorf("%s: after %d answers the overall level %v off, want on it", whose, at, off)
		}
	}
	if r.lag != 0 || r.lagged == 0 {
		t.Errorf("%s: a lag of %v over %d answers, want none over the answers it is read on", whose, r.lag, r.lagged)
	}
}

// With no miss of the model, the oracle's task is in the corridor exactly
// when its topic's ladder held a task there: answer by answer, the tasks in
// the corridor grow with the tasks the ladder allowed, not only in the sum.
func TestTheOracleIsInTheCorridorWheneverTheLadderAllows(t *testing.T) {
	t.Parallel()
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	oracleRule := ruleNamed(t, "oracle")
	denied := 0
	for i := range 20 {
		s := newSession(w, oracleRule, newChild(exactlyAsWritten, i, w.topics))
		for k := range w.answers {
			inside, reachable := s.result.inside, s.result.reachable
			if err := s.step(k); err != nil {
				t.Fatal(err)
			}
			gotInside, allowed := s.result.inside-inside, s.result.reachable-reachable
			if gotInside != allowed {
				t.Fatalf("child %d, answer %d: in the corridor %d, the ladder allowed it %d", i, k+1, gotInside, allowed)
			}
			denied += 1 - allowed
		}
	}
	if denied == 0 {
		t.Log("the ladder allowed a task in the corridor on every answer, so the check never met a task it could not give")
	}
}

// The held-out seeds and the sweep's draw other children than the working
// ones and than each other: their seeds or names differ, the first draws of
// no child's streams in one set are drawn in another, and their results go
// into directories of their own. The test sets the seed and the name every
// other test draws with, so it does not run beside them.
func TestTheHeldOutAndTheSweepsSeedsDrawOtherChildren(t *testing.T) {
	t.Cleanup(func() { masterSeed, experiment = paperSeed, paperExperiment })
	if heldOutSeed == paperSeed || heldOutExperiment == paperExperiment || sweepExperiment == paperExperiment {
		t.Fatalf("the held-out seed %d and name %s and the sweep's name %s, want none the working ones", heldOutSeed, heldOutExperiment, sweepExperiment)
	}
	sets := []struct {
		name, directory string
		args            []string
	}{
		{"working", "results", []string{"-out", "results"}},
		{"held-out", filepath.Join("results", heldOutDirectory), []string{"-out", "results", "-held-out"}},
		{"sweep", filepath.Join("results", "sweep"), []string{"-out", "results", "-rules", "sweep"}},
	}
	drawn := map[float64]string{}
	for _, set := range sets {
		out, draws := firstDraws(t, set.args...)
		if out != set.directory {
			t.Errorf("the %s results go to %s, want %s", set.name, out, set.directory)
		}
		for draw := range draws {
			if other, seen := drawn[draw]; seen {
				t.Errorf("a first draw of %v among the %s children and the %s ones", draw, other, set.name)
			}
			drawn[draw] = set.name
		}
	}
}

// A command line that asks for two things at once is refused: a set of rules
// that draws children of its own and a seed or a name besides, or the
// held-out children and a set of rules other than the confirmation, which is
// all they are run on. One that names no set is refused too, and a decision
// run goes to its own directory on the working children. The test sets the
// seed and the name every other test draws with, so it does not run beside
// them.
func TestACommandLineAskingForTwoThingsAtOnceIsRefused(t *testing.T) {
	t.Cleanup(func() { masterSeed, experiment = paperSeed, paperExperiment })
	if out, _, err := parse([]string{"-held-out", "-out", ""}, io.Discard); err != nil || out != "" {
		t.Errorf("parse with -held-out and no directory = %q, %v; want no directory, for the run to refuse", out, err)
	}
	for _, args := range [][]string{{"-held-out", "-seed", "7"}, {"-experiment", "another", "-held-out"}, {"-rules", "sweep", "-seed", "7"}} {
		if _, _, err := parse(args, io.Discard); !errors.Is(err, errOwnSeeds) {
			t.Errorf("parse(%q) error = %v, want %v", args, err, errOwnSeeds)
		}
	}
	if _, _, err := parse([]string{"-held-out", "-rules", "bench"}, io.Discard); !errors.Is(err, errHeldOutBesideRules) {
		t.Errorf("parse with -held-out and the bench's rules: error %v, want %v", err, errHeldOutBesideRules)
	}
	if _, _, err := parse([]string{"-rules", "nothing"}, io.Discard); err == nil {
		t.Error("parse with a set of rules nobody named = nil error, want one")
	}
	if out, d, err := parse([]string{"-rules", "decision", "-children", "4000"}, io.Discard); err != nil || out != filepath.Join("results", "decision") ||
		d.set != "decision" || masterSeed != paperSeed || experiment != paperExperiment {
		t.Errorf("parse of a decision run = %q, %+v, %v, seed %d and %s; want results/decision on the working children", out, d, err, masterSeed, experiment)
	}
}

// firstDraws reads a command line, as a run would, and draws for a thousand
// children of every generator the first draw of each of their two streams: the
// directory the run would write to, and the draws.
func firstDraws(t *testing.T, args ...string) (out string, draws map[float64]bool) {
	t.Helper()
	out, _, err := parse(args, io.Discard)
	if err != nil {
		t.Fatalf("parse(%q) error = %v, want nil", args, err)
	}
	draws = map[float64]bool{}
	for _, gen := range allGenerators {
		for i := range 1000 {
			id := fmt.Sprintf("%s/%d", gen, i)
			draws[seeded(id, "params").Float64()] = true
			draws[seeded(id, "answers").Float64()] = true
		}
	}
	return out, draws
}
