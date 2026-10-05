package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// baselineRule is the step chosen, under the service's own rule of mastery.
func baselineRule(t *testing.T) *rule {
	t.Helper()
	step, err := chosenRule()
	if err != nil {
		t.Fatal(err)
	}
	return step
}

// The cautious estimate at a margin of one uncertainty, put in the service's
// place, is the service's rule of mastery: run beside the step chosen under
// the service's own mastery, on the same children of every generator, it
// leaves every topic of the profile and the overall level as the service's
// rule leaves them after every answer, declares and loses mastery on the same
// answers, and comes to the same measures.
func TestTheCautiousEstimateInTheServicesPlaceIsTheService(t *testing.T) {
	t.Parallel()
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	step := baselineRule(t)
	ours := modelOf(step, masteryCandidate{name: "cautious_z1", test: func() masteryTest { return cautious{z: 1} }})
	for _, gen := range allGenerators {
		for i := range 3 {
			wantTheSameMastery(t, w, step, ours, gen, i)
		}
	}
}

// wantTheSameMastery runs a child under two rules side by side and holds the
// second to the first after every answer — the profile's topics and ratings,
// the masteries declared and those declared falsely — and in its measures at
// the end.
func wantTheSameMastery(t *testing.T, w *world, theirs, ours *rule, gen generator, i int) {
	t.Helper()
	a, b := newSession(w, theirs, newChild(gen, i, w.topics)), newSession(w, ours, newChild(gen, i, w.topics))
	for k := range w.answers {
		if err := a.step(k); err != nil {
			t.Fatal(err)
		}
		if err := b.step(k); err != nil {
			t.Fatal(err)
		}
		when := fmt.Sprintf("%s child %d, answer %d", gen, i, k+1)
		if !reflect.DeepEqual(b.p.Topics, a.p.Topics) || b.p.Ratings != a.p.Ratings {
			t.Fatalf("%s: the profiles part", when)
		}
		if b.result.declared != a.result.declared || b.result.wrongly != a.result.wrongly {
			t.Fatalf("%s: %d declared, %d falsely; the service's rule %d and %d", when, b.result.declared, b.result.wrongly, a.result.declared, a.result.wrongly)
		}
	}
	ms := metrics()
	if !reflect.DeepEqual(vectorOf(b.result, ms), vectorOf(a.result, ms)) {
		t.Errorf("%s child %d: the measures part", gen, i)
	}
}

// An infinite margin never declares a topic mastered: on children the
// service's rule declares masteries to, the cautious estimate at an infinite
// margin, in the service's place, declares none, and the profile never holds
// one — which also shows that what a rule of mastery holds does reach the
// profile, over what the service's own rule writes there.
func TestAnInfiniteMarginNeverMasters(t *testing.T) {
	t.Parallel()
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	step := baselineRule(t)
	never := modelOf(step, masteryCandidate{name: "cautious_inf", test: func() masteryTest { return cautious{z: math.Inf(1)} }})
	declaredByTheService := 0
	for _, gen := range []generator{staticChildren, learning, farTopics} {
		for i := range 3 {
			baseline, err := run(w, step, newChild(gen, i, w.topics))
			if err != nil {
				t.Fatal(err)
			}
			declaredByTheService += baseline.declared
			wantNeverMastered(t, w, never, gen, i)
		}
	}
	if declaredByTheService == 0 {
		t.Error("the service's rule declared no mastery to these children, so the test shows nothing")
	}
}

// wantNeverMastered runs a child under a rule and holds it to no mastery: none
// held in the profile after any answer, none declared.
func wantNeverMastered(t *testing.T, w *world, r *rule, gen generator, i int) {
	t.Helper()
	s := newSession(w, r, newChild(gen, i, w.topics))
	for k := range w.answers {
		if err := s.step(k); err != nil {
			t.Fatal(err)
		}
		for _, topic := range w.topics {
			if summary := s.p.Topics[topic]; summary.MasteredSince != nil || summary.MasteredLevel != nil {
				t.Fatalf("%s child %d, answer %d: %s held mastered", gen, i, k+1, topic)
			}
		}
	}
	if s.result.declared != 0 {
		t.Errorf("%s child %d: %d masteries declared, want none", gen, i, s.result.declared)
	}
}

// Wald's test errs as it was built to: a child whose chance on the level's
// middle task is the corridor's bottom is let through as mastered at most
// α / (1 − β) of the time, and one at its top turned away at most
// β / (1 − α) of it, within three standard errors over twenty thousand single
// tests on tasks of every difficulty of the level; and almost every test
// decides within a few hundred answers.
func TestWaldsTestErrsAsItWasBuiltTo(t *testing.T) {
	t.Parallel()
	const tests, longest = 20000, 600
	rng := seeded("wald", "test")
	level := rating.Grades34
	for _, tc := range []struct {
		name        string
		above       float64
		wrongToPass bool
		most        float64
	}{
		{"at the bottom", levelAt(rating.CorridorLow), true, waldAlpha / (1 - waldBeta)},
		{"at the top", levelAt(rating.CorridorHigh), false, waldBeta / (1 - waldAlpha)},
	} {
		child := middleTaskOf(level) + tc.above
		wrongly, undecided := 0, 0
		for range tests {
			switch passed, decided := waldOnce(rng, level, child, longest); {
			case !decided:
				undecided++
			case passed == tc.wrongToPass:
				wrongly++
			}
		}
		share := float64(wrongly) / tests
		if se := math.Sqrt(tc.most * (1 - tc.most) / tests); share > tc.most+3*se {
			t.Errorf("%s: wrong %.4f of the time, want at most %.4f", tc.name, share, tc.most)
		}
		if float64(undecided)/tests > 0.001 {
			t.Errorf("%s: %d of %d tests undecided after %d answers", tc.name, undecided, tests, longest)
		}
	}
}

// waldOnce is one test of Wald's at a level, for a child at a level of their
// own, on tasks of every difficulty of the level drawn at random: whether it
// let the child through as mastered, and whether it decided within so many
// answers.
func waldOnce(rng *rand.Rand, level rating.GradeLevel, child float64, longest int) (passed, decided bool) {
	sum := 0.0
	for range longest {
		beta := middleTaskOf(level) + float64(rng.IntN(5)-2)
		sum = waldStep(sum, level, beta, rng.Float64() < rating.Probability(child, beta))
		if sum >= waldUpper || sum <= waldLower {
			return sum >= waldUpper, true
		}
	}
	return false, false
}

// A mastery is read at the level the topic is held mastered at, which a rule
// may set below the task's: a child mastered at the lower level and not at the
// task's is no false mastery there, and the mastery counts as one held below
// the task's level.
func TestAMasteryIsReadAtTheLevelItIsHeldAt(t *testing.T) {
	t.Parallel()
	topics := []string{"t"}
	c := newChild(exactlyAsWritten, 0, topics)
	c.theta, c.delta["t"] = middleTaskOf(rating.Grades12)+2, 0
	if c.truthAt("t", rating.Grades12) < masteredAt || c.truthAt("t", rating.Grades34) >= masteredAt {
		t.Fatalf("the child is mastered at 1-2 by %v and at 3-4 by %v; the case wants only the first", c.truthAt("t", rating.Grades12), c.truthAt("t", rating.Grades34))
	}
	p := profile.New(profile.Student{Grade: 3, Pseudonym: "sim"}, "simulation", firstDay)
	summary := p.Topics["t"]
	held, since := rating.Grades12, profile.DateOf(firstDay)
	summary.MasteredLevel, summary.MasteredSince = &held, &since
	p.Topics["t"] = summary
	s := &session{w: &world{catalog: oneTopic{}}, c: c, p: p, est: &stub{at: c.level("t")}}
	r := newChildResult(c)
	brief := profile.Brief{TargetConcept: "t", GradeLevel: rating.Grades34}
	r.after(s, &brief, &profile.Recorded{Mastered: true}, &observedAnswer{}, 30)
	if r.declared != 1 || r.wrongly != 0 || r.heldBelow != 1 {
		t.Errorf("%d declared, %d falsely, %d below the task's level; want 1, 0 and 1", r.declared, r.wrongly, r.heldBelow)
	}
}

// The cautious estimate masters a topic at the highest of its levels, at or
// below the task's, whose middle task the estimate less its margin clears;
// never at a level above the task's, never before enough answers in the
// topic, and never after a wrong or an aided answer.
func TestTheCautiousEstimateMastersTheHighestLevelItClears(t *testing.T) {
	t.Parallel()
	levels := []rating.GradeLevel{rating.Grades12, rating.Grades34, rating.Grades56}
	clears := func(level rating.GradeLevel) float64 {
		return middleTaskOf(level) + levelAt(rating.CorridorMiddle) + 0.01
	}
	answer := func(estimate float64, task rating.GradeLevel) *masteryAnswer {
		return &masteryAnswer{topic: "t", level: task, correct: true, inTopic: profile.MasteryAnswers, answers: 50, estimate: estimate, levels: levels}
	}
	for _, tc := range []struct {
		name  string
		a     *masteryAnswer
		want  rating.GradeLevel
		masts bool
	}{
		{"the task's level", answer(clears(rating.Grades34), rating.Grades34), rating.Grades34, true},
		{"a level below", answer(clears(rating.Grades12), rating.Grades34), rating.Grades12, true},
		{"not above the task's", answer(clears(rating.Grades56), rating.Grades34), rating.Grades34, true},
		{"nothing cleared", answer(middleTaskOf(rating.Grades12), rating.Grades34), "", false},
	} {
		if got, masts := (cautious{z: 0}).passed(tc.a); got != tc.want || masts != tc.masts {
			t.Errorf("%s: %q, %v; want %q, %v", tc.name, got, masts, tc.want, tc.masts)
		}
	}
	for name, a := range map[string]*masteryAnswer{
		"too few answers": {topic: "t", level: rating.Grades34, correct: true, inTopic: profile.MasteryAnswers - 1, estimate: 100, levels: levels},
		"a wrong answer":  {topic: "t", level: rating.Grades34, inTopic: 10, estimate: 100, levels: levels},
		"an aided answer": {topic: "t", level: rating.Grades34, correct: true, hint: true, inTopic: 10, estimate: 100, levels: levels},
	} {
		if _, masts := (cautious{z: 0}).passed(a); masts {
			t.Errorf("%s mastered the topic", name)
		}
	}
}

// The uncertainty of a level from the counts is the spread the trial series
// starts from, and the topic's the service's step implies, before any answer,
// and narrows with every answer.
func TestTheUncertaintyFromTheCountsNarrowsWithEveryAnswer(t *testing.T) {
	t.Parallel()
	if got, want := uncertaintyOf(0, 0), trialSpread*trialSpread+serviceDecay/answerInformation; math.Abs(got-want) > 1e-12 {
		t.Errorf("before any answer %v, want %v", got, want)
	}
	for n := range 100 {
		if uncertaintyOf(n+1, n/3) >= uncertaintyOf(n, n/3) || uncertaintyOf(n, n/3+1) >= uncertaintyOf(n, n/3) {
			t.Fatalf("after %d answers the uncertainty did not narrow with another", n)
		}
	}
}

// A model of mastery is no candidate of the step, and a step's candidate none
// of mastery: each choice weighs its own.
func TestEachChoiceWeighsItsOwnCandidates(t *testing.T) {
	t.Parallel()
	step := baselineRule(t)
	models, err := masteryModels()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range models {
		if isCandidateIn(m, forStep) {
			t.Errorf("%s is a candidate of the step", m.name)
		}
	}
	if isCandidateIn(step, forMastery) || isCandidateIn(uncertainStep{afterSeries: 0.4, topicSpread: 0.5, limit: noLimit}.rule(), forMastery) {
		t.Error("a step is a candidate of mastery")
	}
}

// The rules of mastery in the bench's sets come to the same bits every time
// they are run, Wald's sums kept in the levels' order among them.
func TestEveryModelOfMasteryRunsTheSameTwice(t *testing.T) {
	t.Parallel()
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	models, err := masteryModels()
	if err != nil {
		t.Fatal(err)
	}
	ms := metrics()
	for _, m := range models {
		first, err := run(w, m, newChild(learning, 2, w.topics))
		if err != nil {
			t.Fatal(err)
		}
		again, err := run(w, m, newChild(learning, 2, w.topics))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(vectorOf(first, ms), vectorOf(again, ms)) {
			t.Errorf("%s: two runs of one child differ", m.name)
		}
	}
}
