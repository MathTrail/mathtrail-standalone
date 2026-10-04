// Package progress is where a child stands, as the child and the parent are
// shown it: the overall rating with its rank and how far through the rank it
// has come, each topic the child has met or could be set now with a rank of
// its own beside the overall one, what is mastered, the latest answers and the
// tasks left without one, the mistakes that keep coming back, and what the
// rule would set next.
//
// It is worked out from the profile every time it is asked for and kept
// nowhere, so it cannot disagree with the file it came from. While the trial
// series runs there is no rating to show at all — the series is still finding
// where the child stands on the ladder — and the summary says how far the
// series has got instead.
package progress

import (
	"fmt"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// Summary is where the child stands.
type Summary struct {
	// Trial is how far the trial series has got, or nil once it is over.
	Trial *Trial
	// Overall is the child's rating on the one ladder of grades 1 to 6, with
	// the rank drawn with it, or nil while the trial series runs.
	Overall *Standing
	// Changes are how the overall standing moved since the last answer and
	// over the week.
	Changes Changes
	// Topics are the topics the child has met — answered or skipped a task of
	// — and those the rule could set now, in catalog order.
	Topics []Topic
	// Recent are the entries of the history window, the latest first: the
	// answers, and the tasks skipped between them where they were skipped.
	Recent []profile.Answer
	// Skipped is how many tasks were left without an answer in all, in every
	// topic the profile counted, one no longer in the catalog too: one number,
	// so that every reader of the progress says the same one.
	Skipped int
	// Next is what the rule would set next.
	Next Recommendation
}

// Trial is how far the trial series has got.
type Trial struct {
	// Answered is how many of its answers are in, and Of how many it takes.
	Answered int
	Of       int
}

// Standing is a rating as it is shown: the whole number, the rank drawn with
// it, and how far through that rank it has come, as a whole percent.
type Standing struct {
	Rating int
	Rank   int
	Share  int
}

// Comparison is how a topic's rank stands to the overall one.
type Comparison string

const (
	// Ahead is a rank above the overall one.
	Ahead Comparison = "ahead"
	// Even is the overall rank itself.
	Even Comparison = "even"
	// Behind is a rank below the overall one.
	Behind Comparison = "behind"
)

// Topic is one topic as the progress shows it.
type Topic struct {
	// ID is the topic's catalog id. What it is called is the reader's
	// business, in the reader's language.
	ID string
	// Standing is the child's rating in the topic with its rank, or nil while
	// the trial series runs, and for a topic with no answer yet, whose rating
	// would only be the overall one again.
	Standing *Standing
	// Compared is how the topic's rank stands to the overall one, and empty
	// when the topic has no standing.
	Compared Comparison
	// Answers and Correct are how many answers the topic rests on, and how
	// many of them were right.
	Answers int
	Correct int
	// Mastered says the topic is mastered at the level its tasks now come
	// from.
	Mastered bool
	// Skipped is how many of its tasks were left without an answer. It is for
	// the parent, who can see from it that hard tasks are being leafed past.
	Skipped int
	// Changes are how the topic's own standing moved since the last answer
	// and over the week, nil together with its standing.
	Changes Changes
}

// Recommendation is what the rule would set next: the topic, the point of the
// ladder, and why now.
type Recommendation struct {
	Topic      string
	GradeLevel rating.GradeLevel
	Difficulty int
	Goal       profile.Goal
}

// Of is where the child whose profile this is stands today, a date in UTC,
// and how that moved since the last answer and over the week ending today. It
// fails only when the rule can suggest nothing at all, which is a catalog with
// no topic taught at any level.
func Of(p *profile.Profile, catalog tutor.Catalog, today profile.Date) (Summary, error) {
	next, err := Recommend(p, catalog)
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{
		Trial:  TrialOf(p),
		Recent: latestFirst(p.Recent),
		Next:   next,
	}
	if summary.Trial == nil {
		summary.Overall = standing(p.Ratings.Theta)
	}
	summary.Topics = topics(p, catalog, summary.Overall)
	for _, topic := range p.Topics {
		summary.Skipped += topic.Skipped
	}
	measured := whilesOf(p, today)
	summary.Changes = measured.overall(summary.Overall)
	for i := range summary.Topics {
		summary.Topics[i].Changes = measured.topic(p.Ratings.Theta, &summary.Topics[i])
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

// topics are the topics the child has met — answered or skipped a task of —
// and those the rule could set now, in catalog order, so that the list grows
// as the child does. A topic no longer in the catalog is not shown at all.
// A topic stands somewhere once the trial series is over and the topic has an
// answer, overall being nil until then, and its rank is compared with the
// overall one rank for rank, so that the word beside it agrees with the names
// drawn.
func topics(p *profile.Profile, catalog tutor.Catalog, overall *Standing) []Topic {
	reachable := tutor.WithinReach(p, catalog)
	listed := []Topic{}
	for _, id := range catalog.TopicIDs() {
		summary := p.Topics[id]
		met := summary.Answers > 0 || summary.Skipped > 0
		if !met && !slices.Contains(reachable, id) {
			continue
		}
		topic := Topic{
			ID:       id,
			Answers:  summary.Answers,
			Correct:  summary.Correct,
			Mastered: tutor.Mastered(p, catalog, id),
			Skipped:  summary.Skipped,
		}
		if overall != nil && summary.Answers > 0 {
			topic.Standing = standing(p.LevelIn(id))
			topic.Compared = compared(topic.Standing.Rank, overall.Rank)
		}
		listed = append(listed, topic)
	}
	return listed
}

// compared is how rank stands to the overall rank.
func compared(rank, overall int) Comparison {
	switch {
	case rank > overall:
		return Ahead
	case rank < overall:
		return Behind
	default:
		return Even
	}
}

// standing is the rating of a level, as it is shown.
func standing(level float64) *Standing {
	elo := rating.Elo(level)
	return &Standing{Rating: rating.Shown(elo), Rank: rating.Rank(elo), Share: rating.Share(elo)}
}

// latestFirst is the history window turned round, sharing nothing with it.
func latestFirst(window []profile.Answer) []profile.Answer {
	latest := make([]profile.Answer, 0, len(window))
	for i := len(window) - 1; i >= 0; i-- {
		latest = append(latest, window[i])
	}
	return latest
}
