package rating_test

import (
	"fmt"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The worked examples of the specification, written out column by column.
// They exist so that a failure can be read by eye: the table below is the one
// a person checks the arithmetic against, and every number in it is printed to
// the four decimals the specification prints.
//
// The prototype's exported vectors are the larger reference and live in
// golden_test.go; these are the ones anybody can follow by hand.

// The corridor and the chance at the recommended difficulty are printed to two
// decimals in the tables, and are held to that.
const printed = 5e-3

// step is one row: the state before the answer, the answer, and everything the
// table prints after it.
type step struct {
	topic       string
	difficulty  int
	theta       float64 // before
	delta       float64 // before
	probability float64
	correct     bool
	thetaAfter  float64
	deltaAfter  float64
	betaMin     float64 // the corridor after the answer
	betaMax     float64
	next        int     // the difficulty the rule would ask for next
	nextChance  float64 // and the chance of a correct answer at it
}

// topicState is what one topic carries between the answers of a table.
type topicState struct {
	delta   float64
	answers int
}

// replay walks a table, holding the package to every column of every row. The
// state is carried from row to row rather than taken from the table, so that
// one wrong number fails its own row and every row after it, the way a person
// checking the arithmetic by hand would notice.
func replay(t *testing.T, table []step) (theta float64, answers int, topics map[string]topicState) {
	t.Helper()

	topics = map[string]topicState{}
	for number, row := range table {
		t.Run(fmt.Sprintf("answer %d", number+1), func(t *testing.T) {
			topic := topics[row.topic]
			nearly(t, theta, row.theta, tolerance, "the overall level before")
			nearly(t, topic.delta, row.delta, tolerance, "the topic before")

			state := rating.State{Theta: theta, Delta: topic.delta, Answers: answers, TopicAnswers: topic.answers}
			result := rating.Update(state, youngest(row.difficulty).Beta(), row.correct)

			nearly(t, result.Probability, row.probability, tolerance, "the chance of a correct answer")
			nearly(t, result.Theta, row.thetaAfter, tolerance, "the overall level after")
			nearly(t, result.Delta, row.deltaAfter, tolerance, "the topic after")

			corridor := rating.NewCorridor(result.Level(), rating.Points(rating.Grades12))
			nearly(t, corridor.BetaMin, row.betaMin, printed, "the hard end of the corridor")
			nearly(t, corridor.BetaMax, row.betaMax, printed, "the easy end of the corridor")
			if corridor.Recommended != youngest(row.next) {
				t.Errorf("the next point = %+v, want difficulty %d of %s", corridor.Recommended, row.next, rating.Grades12)
			}
			nearly(t, corridor.Probability(corridor.Recommended), row.nextChance, printed, "the chance at it")
			if corridor.Fit != rating.FitInside {
				t.Errorf("fit = %q, want %q", corridor.Fit, rating.FitInside)
			}

			theta, answers = result.Theta, result.Answers
			topics[row.topic] = topicState{delta: result.Delta, answers: result.TopicAnswers}
		})
	}
	return theta, answers, topics
}

// A new child, ten answers, two topics. Topic A is answered seven times and
// topic B three, and the two never touch each other's correction while both
// move the level they share.
func TestANewChildOverTenAnswers(t *testing.T) {
	t.Parallel()

	const (
		a = "combinatorics.enumeration"
		b = "parity.alternation"
	)

	theta, answers, topics := replay(t, []step{
		{a, 3, 0.0000, 0.0000, 0.6000, true, 0.0800, 0.1600, -1.23, -0.27, 2, 0.82},
		{a, 3, 0.0800, 0.1600, 0.6478, true, 0.1471, 0.2942, -1.03, -0.07, 2, 0.85},
		{a, 4, 0.1471, 0.2942, 0.4911, false, 0.0578, 0.1156, -1.29, -0.34, 2, 0.81},
		{a, 3, 0.0578, 0.1156, 0.6346, true, 0.1214, 0.2427, -1.10, -0.15, 2, 0.84},
		{b, 3, 0.1214, 0.0000, 0.6242, false, 0.0173, -0.2497, -1.70, -0.74, 2, 0.75},
		{b, 2, 0.0173, -0.2497, 0.7464, true, 0.0579, -0.1531, -1.56, -0.61, 2, 0.77},
		{a, 4, 0.0579, 0.2427, 0.4656, true, 0.1401, 0.4209, -0.91, 0.05, 3, 0.71},
		{a, 4, 0.1401, 0.4209, 0.5136, true, 0.2122, 0.5765, -0.68, 0.28, 3, 0.75},
		{a, 4, 0.2122, 0.5765, 0.5579, true, 0.2753, 0.7125, -0.48, 0.48, 3, 0.78},
		{b, 3, 0.2753, -0.1531, 0.6244, true, 0.3271, -0.0165, -1.16, -0.20, 2, 0.83},
	})

	if answers != 10 {
		t.Errorf("answers = %d, want 10", answers)
	}
	nearly(t, theta, 0.3271, tolerance, "the overall level at the end")
	if shown := rating.Shown(rating.Elo(theta)); shown != 1557 {
		t.Errorf("the overall rating = %d, want 1557", shown)
	}

	for _, want := range []struct {
		topic   string
		level   float64
		answers int
		rating  int
	}{
		{a, 1.0397, 7, 1681},
		{b, 0.3106, 3, 1554},
	} {
		topic := topics[want.topic]
		nearly(t, theta+topic.delta, want.level, tolerance, want.topic+": the level at the end")
		if topic.answers != want.answers {
			t.Errorf("%s: answers = %d, want %d", want.topic, topic.answers, want.answers)
		}
		if shown := rating.Shown(rating.Elo(theta + topic.delta)); shown != want.rating {
			t.Errorf("%s: the rating = %d, want %d", want.topic, shown, want.rating)
		}
	}
}

// A settled child fails three times. The corridor slides toward easier tasks
// after every one of them, while the difficulty the child is offered — which
// can only be a whole number — holds for two failures and drops on the third.
// That is the behaviour worth knowing before somebody reports it as a bug.
func TestThreeFailuresInARow(t *testing.T) {
	t.Parallel()

	// Twenty answers behind the overall level and eight in this topic, so the
	// steps are already narrow: this is not a child who moves on one answer.
	const settled, inTopic = 20, 8

	table := []step{
		{"a", 3, 0.9000, 0.3000, 0.8148, false, 0.8185, 0.0672, -0.58, 0.37, 3, 0.77},
		{"a", 3, 0.8185, 0.0672, 0.7664, false, 0.7437, -0.1442, -0.87, 0.09, 3, 0.72},
		{"a", 3, 0.7437, -0.1442, 0.7164, false, 0.6755, -0.3353, -1.13, -0.17, 2, 0.83},
		{"a", 2, 0.6755, -0.3353, 0.8340, true, 0.6910, -0.2924, -1.07, -0.11, 2, 0.84},
	}

	theta, delta, answers, topicAnswers := 0.9, 0.3, settled, inTopic
	for number, row := range table {
		outcome := "wrong"
		if row.correct {
			outcome = "right"
		}
		t.Run(fmt.Sprintf("answer %d, %s", number+1, outcome), func(t *testing.T) {
			nearly(t, theta, row.theta, tolerance, "the overall level before")
			nearly(t, delta, row.delta, tolerance, "the topic before")

			state := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			result := rating.Update(state, youngest(row.difficulty).Beta(), row.correct)

			nearly(t, result.Probability, row.probability, tolerance, "the chance of a correct answer")
			nearly(t, result.Theta, row.thetaAfter, tolerance, "the overall level after")
			nearly(t, result.Delta, row.deltaAfter, tolerance, "the topic after")

			corridor := rating.NewCorridor(result.Level(), rating.Points(rating.Grades12))
			nearly(t, corridor.BetaMin, row.betaMin, printed, "the hard end of the corridor")
			nearly(t, corridor.BetaMax, row.betaMax, printed, "the easy end of the corridor")
			if corridor.Recommended != youngest(row.next) {
				t.Errorf("the next point = %+v, want difficulty %d of %s", corridor.Recommended, row.next, rating.Grades12)
			}
			nearly(t, corridor.Probability(corridor.Recommended), row.nextChance, printed, "the chance at it")

			theta, delta = result.Theta, result.Delta
			answers, topicAnswers = result.Answers, result.TopicAnswers
		})
	}
}

// pastTheFloor is one answer of a child whose overall level's step has
// reached its floor: the state before it, the answer, the two steps it was
// allowed and the levels it left.
type pastTheFloor struct {
	point          rating.Point
	theta, delta   float64 // before
	probability    float64
	correct        bool
	kTheta, kDelta float64
	thetaAfter     float64
	deltaAfter     float64
}

// A child of 79 answers answers three tasks in a topic taught at every level.
// The formula alone would give the overall level a step of 0.0404, 0.0400 and
// 0.0396; the floor holds it at 0.05 at every one, while the topic's step,
// counted by the topic's own nine answers and more, narrows on. What the
// example says of mastery is the profile's to reproduce.
func TestAChildPastTheFloorOfTheStep(t *testing.T) {
	t.Parallel()

	const settled, inTopic = 79, 9
	topic := rating.Points(rating.Grades12, rating.Grades34, rating.Grades56)

	table := []pastTheFloor{
		{rating.Point{GradeLevel: rating.Grades34, Difficulty: 1}, 1.2000, 0.5000, 0.8148, true, 0.0500, 0.2759, 1.2093, 0.5511},
		{youngest(4), 1.2093, 0.5511, 0.7451, false, 0.0500, 0.2667, 1.1720, 0.3524},
		{youngest(3), 1.1720, 0.3524, 0.8569, false, 0.0500, 0.2581, 1.1292, 0.1312},
	}

	theta, delta, answers, topicAnswers := 1.2, 0.5, settled, inTopic
	after := make([]rating.Result, len(table))
	for number, row := range table {
		t.Run(fmt.Sprintf("answer %d", number+1), func(t *testing.T) {
			nearly(t, theta, row.theta, tolerance, "the overall level before")
			nearly(t, delta, row.delta, tolerance, "the topic before")

			state := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			result := rating.Update(state, row.point.Beta(), row.correct)

			nearly(t, result.Probability, row.probability, tolerance, "the chance of a correct answer")
			nearly(t, result.KTheta, row.kTheta, tolerance, "the overall level's step")
			nearly(t, result.KDelta, row.kDelta, tolerance, "the topic's step")
			nearly(t, result.Theta, row.thetaAfter, tolerance, "the overall level after")
			nearly(t, result.Delta, row.deltaAfter, tolerance, "the topic after")
			if result.Answers != answers+1 || result.TopicAnswers != topicAnswers+1 {
				t.Errorf("answers = %d and %d in the topic, want %d and %d", result.Answers, result.TopicAnswers, answers+1, topicAnswers+1)
			}

			theta, delta = result.Theta, result.Delta
			answers, topicAnswers = result.Answers, result.TopicAnswers
			after[number] = result
		})
	}

	for _, want := range []struct {
		answer           int
		overall, inTopic int
	}{
		{1, 1710, 1806},
		{3, 1696, 1719},
	} {
		result := after[want.answer-1]
		if shown := rating.Shown(rating.Elo(result.Theta)); shown != want.overall {
			t.Errorf("after answer %d the overall rating = %d, want %d", want.answer, shown, want.overall)
		}
		if shown := rating.Shown(rating.Elo(result.Level())); shown != want.inTopic {
			t.Errorf("after answer %d the topic's rating = %d, want %d", want.answer, shown, want.inTopic)
		}
	}

	// After the first answer the topic is asked for at difficulty 4 of the
	// youngest level, inside the corridor, which is the level it is mastered at.
	corridor := rating.NewCorridor(after[0].Level(), topic)
	if corridor.Recommended != youngest(4) || corridor.Fit != rating.FitInside {
		t.Errorf("after the first answer the next point = %+v, %q; want difficulty 4 of %s, %q",
			corridor.Recommended, corridor.Fit, rating.Grades12, rating.FitInside)
	}
	nearly(t, corridor.Probability(corridor.Recommended), 0.7451, tolerance, "the chance at it")
}
