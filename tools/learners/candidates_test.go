package main

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// A step from counted answers given the service's numbers — its floor, nothing
// added, no limit — is the service's step to the last bit: run beside the service on
// the same children of every generator, it stands where the service stands
// after every answer, overall and in every topic, hands out the same task,
// holds the same masteries, and comes to the same measures. The service's
// numbers were rounded on amd64, where Go fuses no multiplication with an
// addition; elsewhere the last bits of the two may part, and the check is not
// made.
func TestAStepAtTheServicesNumbersIsTheService(t *testing.T) {
	t.Parallel()
	if runtime.GOARCH != "amd64" {
		t.Skipf("the bits are compared on amd64, and this is %s", runtime.GOARCH)
	}
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	same := &rule{name: "same", shape: both, trial: true, make: steps(func(s *stepRule) {
		s.overallStep = stepCurve{first: serviceK0Theta, decay: serviceDecay, from: rating.TrialAnswers}
		s.topicStep = stepCurve{first: serviceK0Delta, decay: serviceDecay}
		s.floor = serviceFloor
		s.limit = noLimit
	})}
	for _, gen := range allGenerators {
		for i := range 3 {
			wantTheService(t, w, same, gen, i)
		}
	}
}

// wantTheService runs a child under a rule and under the service side by
// side, holds the rule to the service after every answer, and holds its
// measures to the service's at the end.
func wantTheService(t *testing.T, w *world, r *rule, gen generator, i int) {
	t.Helper()
	whose := fmt.Sprintf("%s child %d", gen, i)
	theService := newSession(w, serviceRule(), newChild(gen, i, w.topics))
	theRule := newSession(w, r, newChild(gen, i, w.topics))
	for k := range w.answers {
		if err := theService.step(k); err != nil {
			t.Fatal(err)
		}
		if err := theRule.step(k); err != nil {
			t.Fatal(err)
		}
		wantInStep(t, fmt.Sprintf("%s, answer %d", whose, k+1), theRule, theService)
	}
	ms := metrics()
	if got, want := vectorOf(theRule.result, ms), vectorOf(theService.result, ms); !reflect.DeepEqual(got, want) {
		t.Errorf("%s: the measures of the rule and of the service differ", whose)
	}
}

// wantInStep holds a session to the service's after the same answer: the
// same level overall and in every topic, the same task, the same masteries.
func wantInStep(t *testing.T, when string, got, service *session) {
	t.Helper()
	if got.est.overall() != service.p.Ratings.Theta {
		t.Fatalf("%s: overall %v, the service %v", when, got.est.overall(), service.p.Ratings.Theta)
	}
	for _, topic := range got.w.topics {
		if got.est.level(topic) != service.p.LevelIn(topic) {
			t.Fatalf("%s: %v in %s, the service %v", when, got.est.level(topic), topic, service.p.LevelIn(topic))
		}
		if a, b := got.p.Topics[topic].MasteredLevel, service.p.Topics[topic].MasteredLevel; (a == nil) != (b == nil) || (a != nil && *a != *b) {
			t.Fatalf("%s: %s mastered at %v, by the service at %v", when, topic, a, b)
		}
	}
	mine, _ := got.p.LastAnswer()
	theirs, _ := service.p.LastAnswer()
	if mine.Topic != theirs.Topic || mine.GradeLevel != theirs.GradeLevel || mine.Difficulty != theirs.Difficulty || mine.Correct != theirs.Correct {
		t.Fatalf("%s: answered %s at %s/%d, the service %s at %s/%d", when, mine.Topic, mine.GradeLevel, mine.Difficulty, theirs.Topic, theirs.GradeLevel, theirs.Difficulty)
	}
}

// Every rule that starts after the trial series can follow it, in every set
// a run can be given and among the rules a choice is made of.
func TestEveryRuleThatStartsAfterTheSeriesFollowsIt(t *testing.T) {
	t.Parallel()
	u := uncertainStep{afterSeries: 0.4, topicSpread: 0.5, addedOverall: 0.003, addedTopic: 0.006, limit: 0.42}
	all := []*rule{u.rule(), storedLike(u).rule(), filterRule(u), historyRule(0.5), switchRule(4), floorRule(0.02)}
	for _, part := range u.withoutParts() {
		all = append(all, part.rule())
	}
	for _, set := range ruleSets() {
		if rules, err := set.rules(); err == nil {
			all = append(all, rules...)
		}
	}
	c := newChild(staticChildren, 0, []string{"a"})
	for _, r := range all {
		if r.service || !r.trial {
			continue
		}
		if _, follows := r.make(c, 0).(follower); !follows {
			t.Errorf("%s starts after the trial series and cannot follow it", r.name)
		}
	}
}

// Each new rule comes to the same bits every time it is run: whatever it sums
// over topics it sums in the same order.
func TestEveryNewRuleRunsTheSameTwice(t *testing.T) {
	t.Parallel()
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	u := uncertainStep{afterSeries: 1, topicSpread: 0.5, addedOverall: 0.01, addedTopic: 0.03, limit: 0.3}
	ms := metrics()
	for _, r := range []*rule{u.rule(), storedLike(u).rule(), filterRule(u), historyRule(0.5), switchRule(2)} {
		for _, gen := range []generator{jumping, learning} {
			first, err := run(w, r, newChild(gen, 1, w.topics))
			if err != nil {
				t.Fatal(err)
			}
			again, err := run(w, r, newChild(gen, 1, w.topics))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(vectorOf(first, ms), vectorOf(again, ms)) {
				t.Errorf("%s on %s: two runs of one child differ", r.name, gen)
			}
		}
	}
}

// The full filter and the backup take a first answer alike when they start
// alike: both are a Kalman filter over two levels whose errors do not yet go
// together.
func TestTheFullFilterAndTheBackupTakeAFirstAnswerAlike(t *testing.T) {
	t.Parallel()
	rng := seeded("first answer", "test")
	for range 1000 {
		start, overallUncertainty, spread := 4*rng.Float64(), 0.05+rng.Float64(), 0.2+rng.Float64()
		beta, correct := start+2*rng.Float64()-1, rng.IntN(2) == 0
		f := newFullFilter(start, overallUncertainty, spread, 0.003, 0.006, noLimit)
		s := newStoredUncertainty(start, spread, 0.003, 0.006, noLimit)
		s.overallUncertainty, s.started = overallUncertainty, true
		f.answered("t", beta, correct)
		s.answered("t", beta, correct)
		for _, pair := range [][2]float64{
			{f.overall(), s.overall()}, {f.level("t"), s.level("t")},
			{f.uncertainty[0][0], s.overallUncertainty}, {f.uncertainty[1][1], s.topicUncertainty["t"]},
		} {
			if math.Abs(pair[0]-pair[1]) > 1e-12 {
				t.Fatalf("from %v with uncertainties %v and %v, β %v, %v: the filter %v, the backup %v",
					start, overallUncertainty, spread*spread, beta, correct, pair[0], pair[1])
			}
		}
	}
}

// With the topics held where they start, the estimate over the whole history
// is the trial series' own estimate: the most likely level given the start
// and the answers, which the service finds by another road.
func TestTheWholeHistoryWithTopicsHeldIsTheSeriesEstimate(t *testing.T) {
	t.Parallel()
	start := rating.Start(3)
	answers := []rating.Answer{
		{Point: rating.Point{GradeLevel: rating.Grades34, Difficulty: 3}, Correct: true},
		{Point: rating.Point{GradeLevel: rating.Grades34, Difficulty: 4}, Correct: false},
		{Point: rating.Point{GradeLevel: rating.Grades34, Difficulty: 2}, Correct: true},
		{Point: rating.Point{GradeLevel: rating.Grades34, Difficulty: 4}, Correct: true},
		{Point: rating.Point{GradeLevel: rating.Grades34, Difficulty: 5}, Correct: false},
	}
	h := newWholeHistory(start, 1e-6)
	for i, a := range answers {
		h.answered(fmt.Sprintf("topic %d", i%2), a.Point.Beta(), a.Correct)
	}
	if want := rating.Estimate(start, answers); math.Abs(h.overall()-want) > 1e-6 {
		t.Errorf("the whole history stands at %v, the series' estimate at %v", h.overall(), want)
	}
}

// The steps one number away from a step on the ladders are the values next to
// each of its numbers, below and above, and only those the ladder has.
func TestTheNeighboursOfAStepAreOneNumberAway(t *testing.T) {
	t.Parallel()
	ladders := numbers{
		afterSeries: []float64{0.1, 0.2, 0.3}, topicSpread: []float64{0.5},
		addedOverall: []float64{0, 0.01}, addedTopic: []float64{0.02}, limit: []float64{0.3, noLimit},
	}
	u := uncertainStep{afterSeries: 0.2, topicSpread: 0.5, addedOverall: 0, addedTopic: 0.02, limit: noLimit}
	want := []uncertainStep{
		{afterSeries: 0.1, topicSpread: 0.5, addedTopic: 0.02, limit: noLimit},
		{afterSeries: 0.3, topicSpread: 0.5, addedTopic: 0.02, limit: noLimit},
		{afterSeries: 0.2, topicSpread: 0.5, addedOverall: 0.01, addedTopic: 0.02, limit: noLimit},
		{afterSeries: 0.2, topicSpread: 0.5, addedTopic: 0.02, limit: 0.3},
	}
	if got := u.neighbours(&ladders); !reflect.DeepEqual(got, want) {
		t.Errorf("neighbours %+v, want %+v", got, want)
	}
}

// A step's parts taken away are each part alone, all three, and the
// service's gains — but none that leaves the step as it is.
func TestAStepsPartsAreTakenAwayOneAtATime(t *testing.T) {
	t.Parallel()
	u := uncertainStep{afterSeries: 0.4, topicSpread: 0.5, addedOverall: 0.003, limit: 0.42}
	got := u.withoutParts()
	want := []uncertainStep{
		{afterSeries: 0.4, topicSpread: 0.5, addedOverall: 0.003, limit: noLimit},
		{afterSeries: 0.4, topicSpread: 0.5, limit: 0.42},
		{afterSeries: 0.4, topicSpread: 0.5, limit: noLimit},
		{afterSeries: 0.4, topicSpread: 0.5, addedOverall: 0.003, limit: 0.42, serviceGains: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parts taken away %+v, want %+v", got, want)
	}
	if added := u.added(); added != 2 {
		t.Errorf("the step adds %d numbers, want 2", added)
	}
}

// A step stands nowhere from the service's, as the choice of the step found
// it, when it is that step, and further the more its steps part from it.
func TestAStepsDistanceFromTheServicesGrowsWithTheirDifference(t *testing.T) {
	t.Parallel()
	if d := stepDistance(serviceOverallStep, serviceTopicStep, 0); d != 0 {
		t.Errorf("the service's step stands %v from itself, want 0", d)
	}
	near, far := stepDistance(serviceOverallStep, serviceTopicStep, 0.02), stepDistance(serviceOverallStep, serviceTopicStep, 0.1)
	if near <= 0 || far <= near {
		t.Errorf("floors of 0.02 and 0.1 stand %v and %v from the service's step before its floor, want 0 < the first < the second", near, far)
	}
}
