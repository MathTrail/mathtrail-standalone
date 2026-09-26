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

// The properties here name what has to hold for every child rather than for
// the ten in the worked examples. Five things in this package have that shape:
// the curve, which must never leave the band between guessing and certainty;
// the corridor, which promises to hold at most one difficulty of a level and
// to recommend the nearest of the points it is given whether or not it holds
// any; an answer, which may move a rating only in the direction it argues for
// and only as far as the step allows; the estimate of the trial series, which
// only answers may move away from the start; and the ranks, which never fall
// as the rating rises.

// genLevel covers far more than a child ever reaches: the ladder runs from −2
// to +7, and a property that only held there would be a property of the ladder
// rather than of the package.
func genLevel() gopter.Gen { return gen.Float64Range(-12, 12) }

func genDifficulty() gopter.Gen { return gen.IntRange(1, rating.Difficulties) }

func genAnswers() gopter.Gen { return gen.IntRange(0, 5000) }

// genPoint is any point of the ladder.
func genPoint() gopter.Gen {
	return gopter.CombineGens(gen.IntRange(0, 2), genDifficulty()).Map(func(values []any) rating.Point {
		place, _ := values[0].(int)
		difficulty, _ := values[1].(int)
		return rating.Point{GradeLevel: rating.GradeLevels()[place], Difficulty: difficulty}
	})
}

// genTopicLevels is the levels a topic can be taught at: a run of neighbouring
// levels, as every topic of the catalog is.
func genTopicLevels() gopter.Gen {
	return gen.OneConstOf(
		[]rating.GradeLevel{rating.Grades12},
		[]rating.GradeLevel{rating.Grades56},
		[]rating.GradeLevel{rating.Grades34, rating.Grades56},
		rating.GradeLevels(),
	)
}

// genTrial is up to a whole trial series of answers anywhere on the ladder.
func genTrial() gopter.Gen {
	answer := gopter.CombineGens(genPoint(), gen.Bool()).Map(func(values []any) rating.Answer {
		point, _ := values[0].(rating.Point)
		correct, _ := values[1].(bool)
		return rating.Answer{Point: point, Correct: correct}
	})
	return gopter.CombineGens(gen.SliceOfN(rating.TrialAnswers, answer), gen.IntRange(0, rating.TrialAnswers)).
		Map(func(values []any) []rating.Answer {
			answers, _ := values[0].([]rating.Answer)
			count, _ := values[1].(int)
			return answers[:min(count, len(answers))]
		})
}

func TestTheCurveHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// Two generators rather than one used twice, so that each argument reads as
	// the side of the comparison it stands for.
	level, beta := genLevel(), genLevel()

	properties.Property("a chance is never below the guessing floor and never certain", prop.ForAll(
		func(level float64, point rating.Point) bool {
			p := rating.Probability(level, point.Beta())
			return p > rating.Guess && p < 1
		},
		genLevel(), genPoint(),
	))

	properties.Property("knowing more never lowers the chance", prop.ForAll(
		func(level, gain float64, point rating.Point) bool {
			if gain <= 0 {
				return true
			}
			beta := point.Beta()
			return rating.Probability(level+gain, beta) > rating.Probability(level, beta)
		},
		genLevel(), gen.Float64Range(0, 6), genPoint(),
	))

	properties.Property("a harder task is never the likelier one", prop.ForAll(
		func(level float64, easier int) bool {
			if easier >= rating.Difficulties {
				return true
			}
			return rating.Probability(level, youngest(easier).Beta()) >
				rating.Probability(level, youngest(easier+1).Beta())
		},
		genLevel(), genDifficulty(),
	))

	// The chess scale is the same curve in the numbers a family reads: the
	// odds between two ratings 400 points apart are ten to one, which is what
	// makes the number worth showing at all.
	properties.Property("the rating shown is the same curve in chess numbers", prop.ForAll(
		func(level, beta float64) bool {
			chess := 1 / (1 + math.Pow(10, (rating.Elo(beta)-rating.Elo(level))/400))
			ours := (rating.Probability(level, beta) - rating.Guess) / (1 - rating.Guess)
			return math.Abs(chess-ours) < 1e-9
		},
		level, beta,
	))

	properties.TestingRun(t)
}

func TestTheCorridorHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("its own bounds give back the band it was cut from", prop.ForAll(
		func(level float64) bool {
			corridor := rating.NewCorridor(level, youngestLevel())
			return math.Abs(rating.Probability(level, corridor.BetaMin)-0.85) < 1e-9 &&
				math.Abs(rating.Probability(level, corridor.BetaMax)-0.70) < 1e-9
		},
		genLevel(),
	))

	// The corridor is narrower than a difficulty, so one level can put at most
	// one of its points inside; where two levels overlap their points stand
	// half a difficulty apart, and two of them fit.
	properties.Property("it holds at most one difficulty of a level, and two points of the ladder", prop.ForAll(
		func(level float64, levels []rating.GradeLevel) bool {
			inside := rating.NewCorridor(level, rating.Points(levels...)).Inside
			return len(inside) <= 2 && atMostOneOfALevel(inside)
		},
		genLevel(), genTopicLevels(),
	))

	properties.Property("when a point is inside, the one recommended is inside", prop.ForAll(
		func(level float64, levels []rating.GradeLevel) bool {
			corridor := rating.NewCorridor(level, rating.Points(levels...))
			if len(corridor.Inside) == 0 {
				return true
			}
			return slices.Contains(corridor.Inside, corridor.Recommended) && corridor.Fit == rating.FitInside
		},
		genLevel(), genTopicLevels(),
	))

	// The recommendation is defined even where the corridor is empty, and this
	// is what it means: of the points given, the one closest to the middle of
	// the band, and the easier one when two are equally close.
	properties.Property("it recommends the given point nearest the middle, ties to the easier", prop.ForAll(
		func(level float64, levels []rating.GradeLevel) bool {
			points := rating.Points(levels...)
			corridor := rating.NewCorridor(level, points)

			best, distance := rating.Point{}, math.Inf(1)
			for _, point := range points {
				if from := math.Abs(corridor.Probability(point) - rating.CorridorMiddle); from < distance {
					best, distance = point, from
				}
			}
			return corridor.Recommended == best
		},
		genLevel(), genTopicLevels(),
	))

	properties.Property("the mark on the recommendation says where it landed", prop.ForAll(
		func(level float64, levels []rating.GradeLevel) bool {
			corridor := rating.NewCorridor(level, rating.Points(levels...))
			switch p := corridor.Probability(corridor.Recommended); corridor.Fit {
			case rating.FitInside:
				return p >= 0.70 && p <= 0.85
			case rating.FitTooHard:
				return p < 0.70
			case rating.FitTooEasy:
				return p > 0.85
			default:
				return false
			}
		},
		genLevel(), genTopicLevels(),
	))

	// A child's level only ever raises a chance, so a point in reach stays in
	// reach as the child grows: a topic that was open does not close.
	properties.Property("a higher level never takes a point out of reach", prop.ForAll(
		func(level, gain float64, point rating.Point) bool {
			return !rating.InReach(level, point) || rating.InReach(level+math.Abs(gain), point)
		},
		genLevel(), gen.Float64Range(0, 6), genPoint(),
	))

	properties.TestingRun(t)
}

// atMostOneOfALevel reports whether no two of these points share a level.
func atMostOneOfALevel(points []rating.Point) bool {
	seen := map[rating.GradeLevel]bool{}
	for _, point := range points {
		if seen[point.GradeLevel] {
			return false
		}
		seen[point.GradeLevel] = true
	}
	return true
}

func TestAnAnswerHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// The overall level and a topic's correction are drawn apart: they are two
	// numbers a child carries, not one number used twice.
	theta, delta := genLevel(), genLevel()
	answers, topicAnswers := genAnswers(), genAnswers()

	properties.Property("an answer moves both levels the way it argues", prop.ForAll(
		func(theta, delta float64, point rating.Point, answers, topicAnswers int, correct bool) bool {
			state := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			result := rating.Update(state, point.Beta(), correct)
			if correct {
				return result.Theta > state.Theta && result.Delta > state.Delta
			}
			return result.Theta < state.Theta && result.Delta < state.Delta
		},
		theta, delta, genPoint(), answers, topicAnswers, gen.Bool(),
	))

	// One answer can only ever be worth its own step, however surprising it
	// was: without that a child could be moved a whole difficulty by a single
	// unlucky task.
	properties.Property("no answer moves a level further than its step", prop.ForAll(
		func(theta, delta float64, point rating.Point, answers, topicAnswers int, correct bool) bool {
			state := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			result := rating.Update(state, point.Beta(), correct)
			return math.Abs(result.Theta-state.Theta) <= result.KTheta &&
				math.Abs(result.Delta-state.Delta) <= result.KDelta
		},
		theta, delta, genPoint(), answers, topicAnswers, gen.Bool(),
	))

	properties.Property("every answer leaves the next one a smaller step", prop.ForAll(
		func(answers int, correct bool) bool {
			now := rating.Update(rating.State{Answers: answers, TopicAnswers: answers}, youngest(3).Beta(), correct)
			later := rating.Update(now.State, youngest(3).Beta(), correct)
			return later.KTheta < now.KTheta && later.KDelta < now.KDelta && later.KTheta > 0
		},
		genAnswers(), gen.Bool(),
	))

	// The same answer to a task the child was expected to fail is worth more
	// than one to a task they were expected to pass.
	properties.Property("the less expected an answer, the further it moves a level", prop.ForAll(
		func(theta float64, easier int) bool {
			if easier >= rating.Difficulties {
				return true
			}
			state := rating.State{Theta: theta}
			easy := rating.Update(state, youngest(easier).Beta(), true)
			hard := rating.Update(state, youngest(easier+1).Beta(), true)
			return hard.Theta > easy.Theta
		},
		genLevel(), genDifficulty(),
	))

	properties.TestingRun(t)
}

func TestTheTrialEstimateHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	start := gen.OneConstOf(rating.Start(1), rating.Start(3), rating.Start(5))

	properties.Property("with no answers the estimate is the start", prop.ForAll(
		func(start float64) bool { return rating.Estimate(start, nil) == start },
		genLevel(),
	))

	// A right answer is evidence of a higher level than a wrong one, whatever
	// else was answered, so turning one wrong answer right can only raise the
	// estimate or leave it. The comparison allows a hair of rounding: the two
	// estimates are pinned down separately, each to the last bits of a float.
	properties.Property("turning a wrong trial answer right never lowers the estimate", prop.ForAll(
		func(start float64, answers []rating.Answer, which int) bool {
			if len(answers) == 0 {
				return true
			}
			wrong := slices.Clone(answers)
			wrong[which%len(wrong)].Correct = false
			right := slices.Clone(wrong)
			right[which%len(right)].Correct = true
			return rating.Estimate(start, right) >= rating.Estimate(start, wrong)-1e-9
		},
		start, genTrial(), gen.IntRange(0, rating.TrialAnswers-1),
	))

	// The start pulls back harder the further the estimate is from it, and no
	// answer pulls with a slope steeper than one, so the answers can take the
	// estimate only so far: that bound is where the search for it looks.
	properties.Property("the estimate stays within reach of the start", prop.ForAll(
		func(start float64, answers []rating.Answer) bool {
			estimate := rating.Estimate(start, answers)
			reach := float64(len(answers)) * 2.5 * 2.5
			return math.Abs(estimate-start) <= reach
		},
		start, genTrial(),
	))

	// All right, the estimate is above the start; all wrong, below it.
	properties.Property("answers all one way move the estimate that way", prop.ForAll(
		func(start float64, answers []rating.Answer, correct bool) bool {
			if len(answers) == 0 {
				return true
			}
			same := slices.Clone(answers)
			for i := range same {
				same[i].Correct = correct
			}
			estimate := rating.Estimate(start, same)
			if correct {
				return estimate > start
			}
			return estimate < start
		},
		start, genTrial(), gen.Bool(),
	))

	properties.TestingRun(t)
}

func TestTheRanksHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("a rank never falls as the rating rises", prop.ForAll(
		func(elo, gain float64) bool {
			return rating.Rank(elo+math.Abs(gain)) >= rating.Rank(elo)
		},
		gen.Float64Range(0, 4000), gen.Float64Range(0, 1000),
	))

	properties.Property("a rank is one of the ranks there are", prop.ForAll(
		func(elo float64) bool {
			rank := rating.Rank(elo)
			return rank >= 1 && rank <= rating.Ranks
		},
		gen.Float64Range(-1e6, 1e6),
	))

	properties.TestingRun(t)
}
