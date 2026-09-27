// Package progress is where a child stands, as the child and the parent are
// shown it: the overall rating and the rank above it, the rating of each topic
// the child has met, what is mastered, the latest answers, and what the rule
// would set next.
//
// It is worked out from the profile every time it is asked for and kept
// nowhere, so it cannot disagree with the file it came from. While the trial
// series runs there is no rating to show at all — the series is still finding
// where the child stands on the ladder — and the summary says how far the
// series has got instead.
package progress

import (
	"fmt"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// Summary is where the child stands.
type Summary struct {
	// Trial is how far the trial series has got, or nil once it is over.
	Trial *Trial
	// Overall is the child's rating on the one ladder of grades 1 to 6, with
	// the rank drawn above it, or nil while the trial series runs.
	Overall *Standing
	// Topics are the topics the child has answered, in catalog order.
	Topics []Topic
	// Recent are the answers of the history window, the latest first.
	Recent []profile.Answer
	// Next is what the rule would set next.
	Next Recommendation
}

// Trial is how far the trial series has got.
type Trial struct {
	// Answered is how many of its answers are in, and Of how many it takes.
	Answered int
	Of       int
}

// Standing is a rating as it is shown: the whole number, and the rank drawn
// above it.
type Standing struct {
	Rating int
	Rank   int
}

// Topic is one topic as the progress shows it.
type Topic struct {
	// ID is the topic's catalog id. What it is called is the reader's
	// business, in the reader's language.
	ID string
	// Rating is the child's rating in the topic, or nil while the trial series
	// runs. A topic has no rank of its own: the one rank there is stands above
	// the overall rating, which is a number a child can remember and climb.
	Rating *int
	// Answers and Correct are how many answers the topic rests on, and how
	// many of them were right.
	Answers int
	Correct int
	// Mastered says the topic is mastered at the level its tasks now come
	// from.
	Mastered bool
}

// Recommendation is what the rule would set next: the topic, the point of the
// ladder, and why now.
type Recommendation struct {
	Topic      string
	GradeLevel rating.GradeLevel
	Difficulty int
	Goal       profile.Goal
}

// Of is where the child whose profile this is stands. It fails only when the
// rule can suggest nothing at all, which is a catalog with no topic taught at
// any level.
func Of(p *profile.Profile, catalog tutor.Catalog) (Summary, error) {
	next, err := Recommend(p, catalog)
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{
		Trial:  TrialOf(p),
		Topics: topics(p, catalog),
		Recent: latestFirst(p.Recent),
		Next:   next,
	}
	if summary.Trial == nil {
		summary.Overall = standing(p.Ratings.Theta)
	}
	return summary, nil
}

// TrialOf is how far the trial series has got, or nil once it is over.
func TrialOf(p *profile.Profile) *Trial {
	if !p.Ratings.InTrial() {
		return nil
	}
	return &Trial{Answered: max(p.Ratings.Answers, 0), Of: rating.TrialAnswers}
}

// Recommend is what the rule would set next, as a reader is told it. The rest
// of the brief — the plot, the traps, the rule's own reasoning — is for the
// model that writes the task, and reaches it when a task is asked for.
func Recommend(p *profile.Profile, catalog tutor.Catalog) (Recommendation, error) {
	brief, _, err := tutor.Next(p, catalog, tutor.Choice{})
	if err != nil {
		return Recommendation{}, fmt.Errorf("progress: %w", err)
	}
	return Recommendation{
		Topic:      brief.TargetConcept,
		GradeLevel: brief.GradeLevel,
		Difficulty: brief.Difficulty,
		Goal:       brief.PedagogicalGoal,
	}, nil
}

// topics are the topics the child has answered, in catalog order. A topic
// given but never answered has nothing to show yet: its rating would only be
// the overall one again. A topic no longer in the catalog is not shown at all.
func topics(p *profile.Profile, catalog tutor.Catalog) []Topic {
	listed := []Topic{}
	for _, id := range catalog.TopicIDs() {
		summary := p.Topics[id]
		if summary.Answers == 0 {
			continue
		}
		topic := Topic{
			ID:       id,
			Answers:  summary.Answers,
			Correct:  summary.Correct,
			Mastered: tutor.Mastered(p, catalog, id),
		}
		if !p.Ratings.InTrial() {
			shown := rating.Shown(rating.Elo(p.LevelIn(id)))
			topic.Rating = &shown
		}
		listed = append(listed, topic)
	}
	return listed
}

// standing is the overall rating of a level, as it is shown.
func standing(level float64) *Standing {
	elo := rating.Elo(level)
	return &Standing{Rating: rating.Shown(elo), Rank: rating.Rank(elo)}
}

// latestFirst is the history window turned round, sharing nothing with it.
func latestFirst(window []profile.Answer) []profile.Answer {
	latest := make([]profile.Answer, 0, len(window))
	for i := len(window) - 1; i >= 0; i-- {
		latest = append(latest, window[i])
	}
	return latest
}
