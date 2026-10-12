package rating_test

import (
	"math"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// easiestOfTheLadder is difficulty 1 of the youngest level, the easiest point
// of a topic taught from grades 1–2.
var easiestOfTheLadder = rating.Point{GradeLevel: rating.Grades12, Difficulty: 1}

func TestReachedAtIsTheLowestLevelInReach(t *testing.T) {
	t.Parallel()
	for _, point := range rating.Points(rating.GradeLevels()...) {
		level := rating.ReachedAt(point)
		if !rating.InReach(level, point) {
			t.Errorf("ReachedAt(%v) = %v, which is out of reach", point, level)
		}
		if below := math.Nextafter(level, math.Inf(-1)); rating.InReach(below, point) {
			t.Errorf("ReachedAt(%v) = %v, but %v below it is in reach already", point, level, below)
		}
	}
}

func TestReachedAtOfAPointThatStandsNowhereIsNoNumber(t *testing.T) {
	t.Parallel()
	if got := rating.ReachedAt(rating.Point{GradeLevel: "7-8", Difficulty: 1}); !math.IsNaN(got) {
		t.Errorf("ReachedAt of a level that is none of the three = %v, want NaN", got)
	}
}

// The worked example: a child at 127 answers whose topic was put on pause, at
// −1.746, answers its probe, difficulty 1 of 1-2, rightly. The step moves the
// topic to −1.649, still out of reach, and the lift takes it the rest of the
// way: to −1.4892, where the probe's point is in reach, the rating 1241.
func TestARightAnswerToAProbeLiftsItsTopicIntoReach(t *testing.T) {
	t.Parallel()
	before := rating.State{Theta: -0.535, Delta: -1.211, Answers: 127, TopicAnswers: 15}
	stepped := rating.Update(before, easiestOfTheLadder.Beta(), true).State
	if rating.InReach(stepped.Level(), easiestOfTheLadder) {
		t.Fatalf("the step alone took the topic to %v, in reach: the example lifts nothing", stepped.Level())
	}

	lifted := rating.Lifted(stepped, before.Level(), easiestOfTheLadder)

	if !rating.InReach(lifted.Level(), easiestOfTheLadder) {
		t.Errorf("lifted to %v, want in reach of %v", lifted.Level(), easiestOfTheLadder)
	}
	if got, want := lifted.Level(), rating.ReachedAt(easiestOfTheLadder); got-want > 1e-12 {
		t.Errorf("lifted to %v, want the lowest level in reach, %v", got, want)
	}
	if got := rating.Shown(rating.Elo(lifted.Level())); got != 1241 {
		t.Errorf("the rating after the lift is %d, want 1241", got)
	}
	if lifted.Theta != stepped.Theta || lifted.Answers != stepped.Answers || lifted.TopicAnswers != stepped.TopicAnswers {
		t.Errorf("the lift moved %+v to %+v, want only the topic's correction moved", stepped, lifted)
	}
}

// A topic just short of reach is lifted by ProbeLift, which takes it past the
// floor: the least a right answer to a probe does is the lift, whatever the
// step was.
func TestALiftIsAtLeastProbeLift(t *testing.T) {
	t.Parallel()
	before := rating.ReachedAt(easiestOfTheLadder) - 0.01
	stepped := rating.State{Theta: 0, Delta: before + 0.02, Answers: 300, TopicAnswers: 40}

	lifted := rating.Lifted(stepped, before, easiestOfTheLadder)

	if got, want := lifted.Level(), before+rating.ProbeLift; got < want || got-want > 1e-12 {
		t.Errorf("lifted to %v, want %v: the step plus nothing less than the lift", got, want)
	}
}

func TestAStepThatWentFurtherIsLeftAsItIs(t *testing.T) {
	t.Parallel()
	before := rating.ReachedAt(easiestOfTheLadder) - 0.3
	stepped := rating.State{Theta: 0.1, Delta: before + 0.5, Answers: 9, TopicAnswers: 2}

	if lifted := rating.Lifted(stepped, before, easiestOfTheLadder); lifted != stepped {
		t.Errorf("Lifted(%+v) = %+v, want it unchanged: the step went past the lift and the floor", stepped, lifted)
	}
}

func TestAStateThatIsNoNumberIsNotLifted(t *testing.T) {
	t.Parallel()
	stepped := rating.State{Theta: math.NaN(), Delta: -1, Answers: 50, TopicAnswers: 5}
	if lifted := rating.Lifted(stepped, -2, easiestOfTheLadder); !math.IsNaN(lifted.Theta) || lifted.Delta != -1 {
		t.Errorf("Lifted of a level that is no number = %+v, want it as it was", lifted)
	}
	stepped = rating.State{Theta: 0, Delta: -2}
	if lifted := rating.Lifted(stepped, math.NaN(), easiestOfTheLadder); lifted != stepped {
		t.Errorf("Lifted from a level before that is no number = %+v, want the step as it was", lifted)
	}
}

func TestARightAnswerToAProbeHoldsItsProperties(t *testing.T) {
	t.Parallel()
	properties := gopter.NewProperties(nil)

	// Two generators rather than one used twice, so that each argument reads
	// as the count it stands for, and the easiest point apart from the probe's.
	answerCounts, topicCounts := genAnswers(), genAnswers()
	easiestPoints, taskPoints := genPoint(), genPoint()

	properties.Property("a right answer to a probe leaves its topic in reach, at least ProbeLift higher, "+
		"and moves nothing but the topic's correction, never down", prop.ForAll(
		func(theta, delta float64, answers, topicAnswers int, easiest, point rating.Point) bool {
			before := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			stepped := rating.Update(before, point.Beta(), true).State
			lifted := rating.Lifted(stepped, before.Level(), easiest)
			return rating.InReach(lifted.Level(), easiest) &&
				lifted.Level() >= before.Level()+rating.ProbeLift &&
				lifted.Theta == stepped.Theta &&
				lifted.Answers == stepped.Answers &&
				lifted.TopicAnswers == stepped.TopicAnswers &&
				lifted.Delta >= stepped.Delta
		},
		gen.Float64Range(-10, 10), gen.Float64Range(-5, 5), answerCounts, topicCounts, easiestPoints, taskPoints,
	))

	properties.Property("a lift stops where its target is: no higher than the lift or the floor asks", prop.ForAll(
		func(theta, delta float64, easiest rating.Point) bool {
			stepped := rating.State{Theta: theta, Delta: delta}
			before := stepped.Level() - 0.05
			lifted := rating.Lifted(stepped, before, easiest)
			target := max(before+rating.ProbeLift, rating.ReachedAt(easiest))
			if stepped.Level() >= target {
				return lifted == stepped
			}
			return lifted.Level() >= target && lifted.Level()-target < 1e-9
		},
		gen.Float64Range(-10, 10), gen.Float64Range(-5, 5), genPoint(),
	))

	properties.TestingRun(t)
}
