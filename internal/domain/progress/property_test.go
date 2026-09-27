package progress_test

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
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
