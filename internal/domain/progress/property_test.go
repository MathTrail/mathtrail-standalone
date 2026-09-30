package progress_test

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// Whatever the ratings say, a child in the trial series is shown no rating and
// no rank anywhere, and a child past it is shown the overall rating with one
// rank from the ladder's own, and a rating for every topic met. The worked
// cases above pin the numbers; this is what catches the profile nobody wrote a
// case for.
func TestTheTrialSeriesDecidesWhetherARatingIsShown(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	properties := gopter.NewProperties(nil)
	properties.Property("no rating during the series, one rank in range after it", prop.ForAll(
		func(answers int, theta, delta float64) bool {
			p := fixture(t, "petya")
			p.Ratings.Answers, p.Ratings.Theta = answers, theta
			gaps := p.Topics["counting.gaps"]
			gaps.Delta = delta
			p.Topics["counting.gaps"] = gaps

			summary, err := progress.Of(p, catalog)
			if err != nil {
				return false
			}
			if p.Ratings.InTrial() {
				return summary.Overall == nil && summary.Trial != nil &&
					summary.Trial.Answered == answers && noTopicRated(&summary)
			}
			return summary.Trial == nil && summary.Overall != nil &&
				summary.Overall.Rank >= 1 && summary.Overall.Rank <= rating.Ranks &&
				everyTopicRated(&summary)
		},
		gen.IntRange(0, 2*rating.TrialAnswers), gen.Float64Range(-4, 9), gen.Float64Range(-3, 3),
	))
	properties.TestingRun(t)
}

func noTopicRated(summary *progress.Summary) bool {
	for _, topic := range summary.Topics {
		if topic.Rating != nil {
			return false
		}
	}
	return true
}

func everyTopicRated(summary *progress.Summary) bool {
	for _, topic := range summary.Topics {
		if topic.Rating == nil {
			return false
		}
	}
	return len(summary.Topics) > 0
}

// Whatever the window holds, the map lists exactly the traps behind at least
// the threshold of its answers, each with how many answers it was behind, the
// most frequent first — and says of each of them, and of no other, that it
// repeats. The worked cases pin the order of a tie; this is what catches the
// window nobody wrote a case for.
func TestTheMapCountsTheWindow(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	properties := gopter.NewProperties(nil)
	properties.Property("every trap at the threshold, counted, the most frequent first", prop.ForAll(
		func(picks []int, repeats int) bool {
			window, times := windowOf(picks)
			mistakes := progress.Mistakes(window, catalog, repeats)
			return countedInOrder(mistakes, times, repeats) && listedAsRepeating(window, catalog, mistakes, times, repeats)
		},
		gen.SliceOf(gen.IntRange(0, len(windowTraps)-1)), gen.IntRange(1, 5),
	))
	properties.TestingRun(t)
}

// windowTraps are what a generated answer has behind it: no trap, or one of
// three.
var windowTraps = []string{"", "off_by_one", "missed_case", "double_count"}

// windowOf is a window of answers with the traps picks name behind them, and
// how many of its answers each trap is behind.
func windowOf(picks []int) (window []profile.Answer, times map[string]int) {
	window = make([]profile.Answer, len(picks))
	times = map[string]int{}
	for at, pick := range picks {
		window[at] = profile.Answer{Topic: "counting.gaps", Trap: windowTraps[pick]}
		if windowTraps[pick] != "" {
			times[windowTraps[pick]]++
		}
	}
	return window, times
}

// countedInOrder says whether every mistake on the map is counted as the
// window has it, at the threshold at least, and comes after none less
// frequent.
func countedInOrder(mistakes []progress.Mistake, times map[string]int, repeats int) bool {
	for at, mistake := range mistakes {
		if mistake.Times != times[mistake.Trap] || mistake.Times < repeats {
			return false
		}
		if at > 0 && mistakes[at-1].Times < mistake.Times {
			return false
		}
	}
	return true
}

// listedAsRepeating says whether the traps on the map, and the traps said to
// repeat, are exactly those behind at least the threshold of the window's
// answers.
func listedAsRepeating(
	window []profile.Answer, catalog tutor.Catalog, mistakes []progress.Mistake, times map[string]int, repeats int,
) bool {
	listed := map[string]bool{}
	for _, mistake := range mistakes {
		listed[mistake.Trap] = true
	}
	for trap, counted := range times {
		repeated := counted >= repeats
		if repeated != listed[trap] || progress.Repeats(window, catalog, trap, repeats) != repeated {
			return false
		}
	}
	return true
}
