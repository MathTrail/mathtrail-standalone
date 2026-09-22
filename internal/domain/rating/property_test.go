package rating_test

import (
	"math"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The properties here name what has to hold for every child rather than for
// the ten in the worked examples. Three things in this package have that
// shape: the curve, which must never leave the band between guessing and
// certainty; the corridor, which promises to hold at most one difficulty and
// to recommend the nearest one whether or not it does; and an answer, which
// may move a rating only in the direction it argues for and only as far as
// the step allows.

// genLevel covers far more than a child ever reaches: the levels of the
// reference run from −2 to +3.5, and a property that only held there would be
// a property of the reference rather than of the package.
func genLevel() gopter.Gen { return gen.Float64Range(-12, 12) }

func genDifficulty() gopter.Gen { return gen.IntRange(1, rating.Difficulties) }

func genAnswers() gopter.Gen { return gen.IntRange(0, 5000) }

func TestTheCurveHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// Two generators rather than one used twice, so that each argument reads as
	// the side of the comparison it stands for.
	level, beta := genLevel(), genLevel()

	properties.Property("a chance is never below the guessing floor and never certain", prop.ForAll(
		func(level float64, difficulty int) bool {
			p := rating.Probability(level, rating.Beta(difficulty))
			return p > rating.Guess && p < 1
		},
		genLevel(), genDifficulty(),
	))

	properties.Property("knowing more never lowers the chance", prop.ForAll(
		func(level, gain float64, difficulty int) bool {
			if gain <= 0 {
				return true
			}
			beta := rating.Beta(difficulty)
			return rating.Probability(level+gain, beta) > rating.Probability(level, beta)
		},
		genLevel(), gen.Float64Range(0, 6), genDifficulty(),
	))

	properties.Property("a harder task is never the likelier one", prop.ForAll(
		func(level float64, easier int) bool {
			if easier >= rating.Difficulties {
				return true
			}
			return rating.Probability(level, rating.Beta(easier)) >
				rating.Probability(level, rating.Beta(easier+1))
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
			corridor := rating.NewCorridor(level)
			return math.Abs(rating.Probability(level, corridor.BetaMin)-0.85) < 1e-9 &&
				math.Abs(rating.Probability(level, corridor.BetaMax)-0.70) < 1e-9
		},
		genLevel(),
	))

	properties.Property("it holds at most one difficulty", prop.ForAll(
		func(level float64) bool { return len(rating.NewCorridor(level).Inside) <= 1 },
		genLevel(),
	))

	properties.Property("a difficulty inside it is the one recommended", prop.ForAll(
		func(level float64) bool {
			corridor := rating.NewCorridor(level)
			if len(corridor.Inside) == 0 {
				return true
			}
			return corridor.Recommended == corridor.Inside[0] && corridor.Fit == rating.FitInside
		},
		genLevel(),
	))

	// The recommendation is defined even where the corridor is empty, and this
	// is what it means: the difficulty closest to the middle of the band, and
	// the easier one when two are equally close.
	properties.Property("it recommends the difficulty nearest the middle, ties to the easier", prop.ForAll(
		func(level float64) bool {
			corridor := rating.NewCorridor(level)
			const middle = (0.70 + 0.85) / 2

			best, distance := 0, math.Inf(1)
			for difficulty := 1; difficulty <= rating.Difficulties; difficulty++ {
				if from := math.Abs(corridor.Probability(difficulty) - middle); from < distance {
					best, distance = difficulty, from
				}
			}
			return corridor.Recommended == best
		},
		genLevel(),
	))

	properties.Property("the mark on the recommendation says where it landed", prop.ForAll(
		func(level float64) bool {
			corridor := rating.NewCorridor(level)
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
		genLevel(),
	))

	properties.TestingRun(t)
}

func TestAnAnswerHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// The overall level and a topic's correction are drawn apart: they are two
	// numbers a child carries, not one number used twice.
	theta, delta := genLevel(), genLevel()
	answers, topicAnswers := genAnswers(), genAnswers()

	properties.Property("an answer moves both levels the way it argues", prop.ForAll(
		func(theta, delta float64, difficulty, answers, topicAnswers int, correct bool) bool {
			state := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			result := rating.Update(state, rating.Beta(difficulty), correct)
			if correct {
				return result.Theta > state.Theta && result.Delta > state.Delta
			}
			return result.Theta < state.Theta && result.Delta < state.Delta
		},
		theta, delta, genDifficulty(), answers, topicAnswers, gen.Bool(),
	))

	// One answer can only ever be worth its own step, however surprising it
	// was: without that a child could be moved a whole difficulty by a single
	// unlucky task.
	properties.Property("no answer moves a level further than its step", prop.ForAll(
		func(theta, delta float64, difficulty, answers, topicAnswers int, correct bool) bool {
			state := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			result := rating.Update(state, rating.Beta(difficulty), correct)
			return math.Abs(result.Theta-state.Theta) <= result.KTheta &&
				math.Abs(result.Delta-state.Delta) <= result.KDelta
		},
		theta, delta, genDifficulty(), answers, topicAnswers, gen.Bool(),
	))

	properties.Property("every answer leaves the next one a smaller step", prop.ForAll(
		func(answers int, correct bool) bool {
			now := rating.Update(rating.State{Answers: answers, TopicAnswers: answers}, rating.Beta(3), correct)
			later := rating.Update(now.State, rating.Beta(3), correct)
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
			easy := rating.Update(state, rating.Beta(easier), true)
			hard := rating.Update(state, rating.Beta(easier+1), true)
			return hard.Theta > easy.Theta
		},
		genLevel(), genDifficulty(),
	))

	properties.TestingRun(t)
}
