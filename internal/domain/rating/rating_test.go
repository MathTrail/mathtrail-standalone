package rating_test

import (
	"fmt"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// What a child meets at each difficulty of the level they start in. This table
// is the one a reader checks the scale against: difficulty 3 is an even chance,
// and the two difficulties either side of it are the stretch and the warm-up.
func TestTheChanceAtEachDifficulty(t *testing.T) {
	t.Parallel()

	for difficulty, want := range map[int]float64{
		1: 0.9046,
		2: 0.7848,
		3: 0.6000,
		4: 0.4152,
		5: 0.2954,
	} {
		t.Run(fmt.Sprintf("difficulty %d", difficulty), func(t *testing.T) {
			t.Parallel()

			got := rating.Probability(0, youngest(difficulty).Beta())
			nearly(t, got, want, tolerance, "the chance for a child who has answered nothing")
		})
	}
}

// Difficulty 3 of the youngest level sits at zero, and its five difficulties
// are one apart.
func TestDifficultyOnTheLevelScale(t *testing.T) {
	t.Parallel()

	for difficulty, want := range map[int]float64{1: -2, 2: -1, 3: 0, 4: 1, 5: 2} {
		if got := youngest(difficulty).Beta(); got != want {
			t.Errorf("β of difficulty %d of the youngest level = %v, want %v", difficulty, got, want)
		}
	}
}

// The floor is what five options and no penalty are worth: a child who knows
// nothing at all still answers one task in five.
func TestTheGuessingFloor(t *testing.T) {
	t.Parallel()

	nearly(t, rating.Probability(-50, 0), rating.Guess, tolerance, "the chance far below the task")
	nearly(t, rating.Probability(50, 0), 1, tolerance, "the chance far above the task")
	nearly(t, rating.Probability(0, 0), 0.6, tolerance, "the chance at the task's own level")
}

// A correct answer raises both levels and a wrong one lowers them — and the
// same answer is worth more the less it was expected.
func TestAnAnswerMovesTheLevelsByHowSurprisingItWas(t *testing.T) {
	t.Parallel()

	state := rating.State{}
	easy, hard := youngest(1).Beta(), youngest(5).Beta()

	correct := rating.Update(state, youngest(3).Beta(), true)
	wrong := rating.Update(state, youngest(3).Beta(), false)
	if correct.Theta <= state.Theta {
		t.Errorf("a correct answer left the level at %v, want it higher than %v", correct.Theta, state.Theta)
	}
	if wrong.Theta >= state.Theta {
		t.Errorf("a wrong answer left the level at %v, want it lower than %v", wrong.Theta, state.Theta)
	}

	// An easy task answered correctly says almost nothing; a hard one says a
	// great deal. Wrong answers read the other way round.
	if rating.Update(state, easy, true).Theta >= rating.Update(state, hard, true).Theta {
		t.Error("a correct answer to an easy task moved the level as much as one to a hard task")
	}
	if rating.Update(state, hard, false).Theta <= rating.Update(state, easy, false).Theta {
		t.Error("a wrong answer to a hard task cost as much as one to an easy task")
	}
}

// The two levels move at their own speeds: a topic has to find its place from
// a standing start, while the level across every topic is already supported by
// everything answered so far.
func TestTheTopicMovesFasterThanTheLevelBehindIt(t *testing.T) {
	t.Parallel()

	first := rating.Update(rating.State{}, youngest(3).Beta(), true)
	if first.KDelta <= first.KTheta {
		t.Errorf("the topic moved by %v and the overall level by %v, want the topic to move further",
			first.KDelta, first.KTheta)
	}

	// A topic never met starts from the overall level, so the first answer in
	// it is measured against what the child can do generally.
	settled := rating.State{Theta: 0.8, Answers: 30}
	nearly(t, rating.Update(settled, youngest(3).Beta(), true).Probability,
		rating.Probability(0.8, 0), tolerance, "the chance in a topic never met")
}

// Every answer narrows the next one: a rating resting on fifty answers is not
// thrown by one bad day, while a new child's finds its place in a handful.
func TestTheStepNarrowsWithEveryAnswer(t *testing.T) {
	t.Parallel()

	previous := 0.0
	for answers := 0; answers < 40; answers++ {
		step := rating.Update(rating.State{Answers: answers, TopicAnswers: answers}, youngest(3).Beta(), true)
		if step.KTheta <= 0 || step.KDelta <= 0 {
			t.Fatalf("after %d answers the steps were %v and %v, want both above zero",
				answers, step.KTheta, step.KDelta)
		}
		if answers > 0 && step.KTheta >= previous {
			t.Errorf("after %d answers the step was %v, want it below the previous %v",
				answers, step.KTheta, previous)
		}
		previous = step.KTheta
	}

	// The counts are read apart: a child on their fiftieth answer meeting a
	// new topic moves that topic as far as a beginner would.
	mixed := rating.Update(rating.State{Answers: 50, TopicAnswers: 0}, youngest(3).Beta(), true)
	fresh := rating.Update(rating.State{}, youngest(3).Beta(), true)
	nearly(t, mixed.KDelta, fresh.KDelta, tolerance, "the step in a topic never met")
	if mixed.KTheta >= fresh.KTheta {
		t.Error("fifty answers left the overall level moving as fast as on the first")
	}
}

// Where the child stands in a topic is the level and its correction together,
// before the answer and after it.
func TestTheLevelInATopic(t *testing.T) {
	t.Parallel()

	state := rating.State{Theta: 0.4, Delta: -0.1}
	nearly(t, state.Level(), 0.3, tolerance, "the level in the topic")

	result := rating.Update(state, youngest(3).Beta(), true)
	nearly(t, result.Level(), result.Theta+result.Delta, tolerance, "the level after the answer")
}

// The counts that narrow the next step are carried by the answer itself. They
// are part of the formula, and a caller who forgot to raise them would leave a
// rating moving at a beginner's speed for ever — so a caller is not asked to.
func TestAnAnswerCountsItself(t *testing.T) {
	t.Parallel()

	state := rating.State{Theta: 0.2, Delta: -0.1, Answers: 7, TopicAnswers: 3}
	result := rating.Update(state, youngest(3).Beta(), true)

	if result.Answers != 8 {
		t.Errorf("answers = %d, want 8", result.Answers)
	}
	if result.TopicAnswers != 4 {
		t.Errorf("answers in the topic = %d, want 4", result.TopicAnswers)
	}

	// The first answer in a second topic advances the child's total and that
	// topic's own count, and says nothing about the first topic.
	elsewhere := rating.Update(rating.State{Theta: result.Theta, Answers: result.Answers}, youngest(3).Beta(), true)
	if elsewhere.Answers != 9 || elsewhere.TopicAnswers != 1 {
		t.Errorf("a first answer in a new topic left %d answers and %d in the topic, want 9 and 1",
			elsewhere.Answers, elsewhere.TopicAnswers)
	}

	// And what comes back is a state, so the next answer is given it as it is.
	next := rating.Update(result.State, youngest(3).Beta(), true)
	if next.Answers != 9 || next.TopicAnswers != 5 {
		t.Errorf("the next answer left %d answers and %d in the topic, want 9 and 5",
			next.Answers, next.TopicAnswers)
	}
	if next.KTheta >= result.KTheta {
		t.Error("the step did not narrow when the state was carried forward")
	}
}

// A count below zero is no count at all: the step is the widest there is, and
// a correct answer still raises the level. A profile never passes validation
// with one, and nothing here should turn it into a level without end.
func TestACountBelowZeroIsNone(t *testing.T) {
	t.Parallel()

	fresh := rating.Update(rating.State{}, 0, true)
	for _, answers := range []int{-20, -40} {
		t.Run(fmt.Sprintf("%d answers", answers), func(t *testing.T) {
			t.Parallel()
			result := rating.Update(rating.State{Answers: answers, TopicAnswers: answers}, 0, true)
			if result.KTheta != fresh.KTheta || result.KDelta != fresh.KDelta || result.Theta <= 0 {
				t.Errorf("Update() = steps %v and %v and level %v, want the steps of no answers and a level raised",
					result.KTheta, result.KDelta, result.Theta)
			}
		})
	}
}

// The guessing floor is one option in however many a task offers. The rating
// does not depend on the package that says how many, and this is what keeps
// the two from drifting apart.
func TestTheGuessingFloorIsOneOptionOfAll(t *testing.T) {
	t.Parallel()

	if want := 1 / float64(solver.Count); rating.Guess != want {
		t.Errorf("Guess = %v, want %v, one option of the %d a task offers", rating.Guess, want, solver.Count)
	}
}
