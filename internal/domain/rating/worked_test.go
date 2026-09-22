package rating_test

import (
	"fmt"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The two worked examples of the specification, written out column by column.
// They exist so that a failure can be read by eye: the table below is the one
// a person checks the arithmetic against, and every number in it is printed to
// the four decimals the specification prints.
//
// The prototype's exported vectors are the larger reference and live in
// golden_test.go; these two are the ones anybody can follow by hand.

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
			result := rating.Update(state, rating.Beta(row.difficulty), row.correct)

			nearly(t, result.Probability, row.probability, tolerance, "the chance of a correct answer")
			nearly(t, result.Theta, row.thetaAfter, tolerance, "the overall level after")
			nearly(t, result.Delta, row.deltaAfter, tolerance, "the topic after")

			corridor := rating.NewCorridor(result.Level())
			nearly(t, corridor.BetaMin, row.betaMin, printed, "the hard end of the corridor")
			nearly(t, corridor.BetaMax, row.betaMax, printed, "the easy end of the corridor")
			if corridor.Recommended != row.next {
				t.Errorf("the next difficulty = %d, want %d", corridor.Recommended, row.next)
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
		t.Run(fmt.Sprintf("failure %d", number+1), func(t *testing.T) {
			nearly(t, theta, row.theta, tolerance, "the overall level before")
			nearly(t, delta, row.delta, tolerance, "the topic before")

			state := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			result := rating.Update(state, rating.Beta(row.difficulty), row.correct)

			nearly(t, result.Probability, row.probability, tolerance, "the chance of a correct answer")
			nearly(t, result.Theta, row.thetaAfter, tolerance, "the overall level after")
			nearly(t, result.Delta, row.deltaAfter, tolerance, "the topic after")

			corridor := rating.NewCorridor(result.Level())
			nearly(t, corridor.BetaMin, row.betaMin, printed, "the hard end of the corridor")
			nearly(t, corridor.BetaMax, row.betaMax, printed, "the easy end of the corridor")
			if corridor.Recommended != row.next {
				t.Errorf("the next difficulty = %d, want %d", corridor.Recommended, row.next)
			}

			theta, delta = result.Theta, result.Delta
			answers, topicAnswers = result.Answers, result.TopicAnswers
		})
	}
}
