package rating_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The third worked example of the specification: a child of grade 3 who
// stands in fact about a level lower, through the five answers of the trial
// series. Every estimate is made from the start and all the answers so far,
// so each row is checked against the whole series up to it rather than
// against the row before — the way the estimate itself is made.
func TestTheTrialSeriesOfTheSpecification(t *testing.T) {
	t.Parallel()

	middle := func(difficulty int) rating.Point {
		return rating.Point{GradeLevel: rating.Grades34, Difficulty: difficulty}
	}
	everyLevel := rating.GradeLevels()
	rows := []struct {
		topic       string
		levels      []rating.GradeLevel // the levels the topic is taught at
		point       rating.Point
		before      float64
		probability float64
		correct     bool
		after       float64
	}{
		{"logic.ordering", everyLevel, middle(2), 2.5000, 0.7848, false, 0.6407},
		{"combinatorics.enumeration", everyLevel, youngest(3), 0.6407, 0.7239, true, 1.1189},
		{"logic.knights_liars", everyLevel[1:], middle(1), 1.1189, 0.7200, false, 0.2536},
		{"counting.gaps", everyLevel, youngest(2), 0.2536, 0.8223, true, 0.4532},
		{"time.clocks", everyLevel, youngest(2), 0.4532, 0.8484, true, 0.6039},
	}

	start := rating.Start(3)
	theta := start
	var answers []rating.Answer
	for number, row := range rows {
		t.Run(fmt.Sprintf("answer %d, %s", number+1, row.topic), func(t *testing.T) {
			nearly(t, theta, row.before, tolerance, "the estimate before")

			// The rule sets the topic at the point its corridor recommends.
			corridor := rating.NewCorridor(theta, rating.Points(row.levels...))
			if corridor.Recommended != row.point {
				t.Errorf("the point = %+v, want %+v", corridor.Recommended, row.point)
			}
			nearly(t, rating.Probability(theta, row.point.Beta()), row.probability, tolerance, "the chance")

			answers = append(answers, rating.Answer{Point: row.point, Correct: row.correct})
			theta = rating.Estimate(start, answers)
			nearly(t, theta, row.after, tolerance, "the estimate after")
		})
	}

	if shown, rank := rating.Shown(rating.Elo(theta)), rating.Rank(rating.Elo(theta)); shown != 1605 || rank != 3 {
		t.Errorf("after the series the child is shown %d in rank %d, want 1605 in rank 3", shown, rank)
	}
}

// Knights and liars is taught from grades 3–4, and the example passes over it
// for the second task and takes it for the third: out of reach at the estimate
// after one wrong answer, in reach once a right one has lifted it.
func TestTheExampleTakesKnightsOnlyOnceTheyAreInReach(t *testing.T) {
	t.Parallel()

	easiest := rating.Point{GradeLevel: rating.Grades34, Difficulty: 1}
	if rating.InReach(0.6407, easiest) {
		t.Errorf("%+v is in reach at 0.6407, at %.4f", easiest, rating.Probability(0.6407, easiest.Beta()))
	}
	if !rating.InReach(1.1189, easiest) {
		t.Errorf("%+v is out of reach at 1.1189, at %.4f", easiest, rating.Probability(1.1189, easiest.Beta()))
	}
}

// Step 1 by hand, as the specification works it: one wrong answer, and at the
// estimate the pull of the start balances the pull of the answer.
func TestTheFirstStepBalancesByHand(t *testing.T) {
	t.Parallel()

	answer := rating.Answer{Point: rating.Point{GradeLevel: rating.Grades34, Difficulty: 2}, Correct: false}
	theta := rating.Estimate(2.5, []rating.Answer{answer})

	start := (theta - 2.5) / (2.5 * 2.5)
	wrong := -(rating.Probability(theta, answer.Point.Beta()) - rating.Guess) / (1 - rating.Guess)
	nearly(t, start, -0.2975, tolerance, "the pull of the start")
	nearly(t, start, wrong, 1e-12, "the pull of the start against the pull of the answer")
}

// Five right answers to the hardest tasks there are, from the youngest start,
// give the estimate two peaks: one near the start, which a search that climbs
// from the start would stop at, and a higher one near the tasks. The estimate
// is the higher.
func TestTheEstimateIsTheHigherOfTwoPeaks(t *testing.T) {
	t.Parallel()

	hardest := rating.Answer{Point: rating.Point{GradeLevel: rating.Grades56, Difficulty: 5}, Correct: true}
	answers := []rating.Answer{hardest, hardest, hardest, hardest, hardest}

	if got := rating.Estimate(0, answers); got < 7 || got > 8.5 {
		t.Errorf("Estimate() = %.4f, want the peak near the tasks, about 7.7", got)
	}
}

// An answer to a point off the ladder says nothing about where the child
// stands, and a start that is no number gives an estimate that is none either
// rather than a search that never ends.
func TestWhatTheEstimateCannotReadIsLeftOut(t *testing.T) {
	t.Parallel()

	off := rating.Answer{Point: rating.Point{GradeLevel: "7-8", Difficulty: 3}, Correct: false}
	if got := rating.Estimate(2.5, []rating.Answer{off}); got != 2.5 {
		t.Errorf("Estimate() over an answer off the ladder = %v, want the start", got)
	}

	answer := rating.Answer{Point: youngest(3), Correct: true}
	if got := rating.Estimate(math.NaN(), []rating.Answer{answer}); !math.IsNaN(got) {
		t.Errorf("Estimate() from a start that is no number = %v, want no number", got)
	}
	if got := rating.Estimate(math.Inf(1), []rating.Answer{answer}); !math.IsInf(got, 1) {
		t.Errorf("Estimate() from an endless start = %v, want the start", got)
	}
}
