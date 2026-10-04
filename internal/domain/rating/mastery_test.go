package rating_test

import (
	"math"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The uncertainty of the worked examples, as the specification prints it: of a
// child of 80 answers in a topic of 10, past the floor of the step, and of one
// of 9 answers in a topic of 7, too new to call anything mastered. Before any
// answer it is the trial series' spread squared and a topic's twenty answers'
// worth.
func TestTheUncertaintyOfTheSpecification(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		answers, inTopic int
		want             float64
	}{
		{0, 0, 6.5834},
		{9, 7, 0.9093},
		{80, 10, 0.3045},
	} {
		nearly(t, rating.Uncertainty(tc.answers, tc.inTopic), tc.want, tolerance, "the uncertainty")
	}
}

// A count below zero, which only a file edited by hand can hold, is no count
// at all: the uncertainty of no answers, rather than one below zero.
func TestTheUncertaintyOfACountBelowZeroIsOfNone(t *testing.T) {
	t.Parallel()

	if got, want := rating.Uncertainty(-20, -40), rating.Uncertainty(0, 0); got != want {
		t.Errorf("Uncertainty(-20, -40) = %v, want %v, the uncertainty of no answers", got, want)
	}
}

// The level a topic is mastered at is the highest of its own levels, at or
// below the task's, whose middle task the cautious level clears: a right
// answer to a task of 3-4 can master a topic at 1-2 while 3-4 is still beyond
// the child; a topic not taught at 1-2 is not mastered there; and nine answers
// in, the margin keeps a child the bare estimate would call ready from it.
func TestTheLevelATopicIsMasteredAt(t *testing.T) {
	t.Parallel()

	every := rating.GradeLevels()
	for _, tc := range []struct {
		name             string
		level            float64
		answers, inTopic int
		task             rating.GradeLevel
		taught           []rating.GradeLevel
		want             rating.GradeLevel
		clears           bool
	}{
		{"a task of 3-4 masters the topic at 1-2", 1.7603, 80, 10, rating.Grades34, every, rating.Grades12, true},
		{"a topic not taught at 1-2 is not mastered there", 1.7603, 80, 10, rating.Grades34, []rating.GradeLevel{rating.Grades34, rating.Grades56}, "", false},
		{"nine answers in, the margin holds the topic back", 0.9878, 9, 7, rating.Grades12, every, "", false},
		{"a strong child is mastered no higher than the task", 5, 200, 50, rating.Grades12, every, rating.Grades12, true},
		{"a strong child is mastered at the task's level when it clears", 5, 200, 50, rating.Grades34, every, rating.Grades34, true},
		{"a level beyond the child falls back to the one below", 5, 200, 50, rating.Grades56, every, rating.Grades34, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, clears := rating.MasteredAt(tc.level, tc.answers, tc.inTopic, tc.task, tc.taught)
			if got != tc.want || clears != tc.clears {
				t.Errorf("mastered at %q, %v; want %q, %v", got, clears, tc.want, tc.clears)
			}
		})
	}
}

// Whatever the level and the counts, a topic is mastered at the highest of its
// levels the cautious level clears, and every answer narrows the margin that
// level is taken by.
func TestMasteryHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	counts, inTopic := gen.IntRange(0, 300), gen.IntRange(0, 300)
	task := gen.IntRange(0, 2).Map(func(place int) rating.GradeLevel { return rating.GradeLevels()[place] })

	// What is declared is a level the topic is taught at, no higher than the
	// task answered, and one whose middle task the cautious level answers at
	// the corridor's middle — and no higher such level does.
	properties.Property("a topic is mastered at the highest of its levels the cautious level clears", prop.ForAll(
		masteredAtTheHighestLevelCleared,
		genLevel(), counts, inTopic, task, genTopicLevels(),
	))

	// Every answer tells something, so more of them never leave a level more
	// uncertain, and no count leaves it certain.
	properties.Property("every answer narrows the uncertainty", prop.ForAll(
		func(answers, topicAnswers int) bool {
			now := rating.Uncertainty(answers, topicAnswers)
			return now > 0 && rating.Uncertainty(answers+1, topicAnswers) < now && rating.Uncertainty(answers, topicAnswers+1) < now
		},
		counts, inTopic,
	))

	properties.TestingRun(t)
}

// masteredAtTheHighestLevelCleared says whether MasteredAt declared the
// highest of the taught levels, at or below the task's, whose middle task the
// cautious level clears — and nothing when none of them does.
func masteredAtTheHighestLevelCleared(level float64, answers, topicAnswers int, task rating.GradeLevel, taught []rating.GradeLevel) bool {
	cautious := level - math.Sqrt(rating.Uncertainty(answers, topicAnswers))
	held, clears := rating.MasteredAt(level, answers, topicAnswers, task, taught)
	if clears && (held.Shift() > task.Shift() || !slices.Contains(taught, held) || !clearsTheMiddle(cautious, held)) {
		return false
	}
	for _, candidate := range taught {
		higher := !clears || candidate.Shift() > held.Shift()
		if candidate.Shift() <= task.Shift() && higher && clearsTheMiddle(cautious, candidate) {
			return false
		}
	}
	return true
}

// clearsTheMiddle says whether a child at this level answers the middle task
// of a grade level at the corridor's middle or better.
func clearsTheMiddle(level float64, at rating.GradeLevel) bool {
	middle := rating.Point{GradeLevel: at, Difficulty: (1 + rating.Difficulties) / 2}
	return rating.Probability(level, middle.Beta()) >= rating.CorridorMiddle
}
