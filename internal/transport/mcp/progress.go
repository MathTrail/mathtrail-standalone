package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// textAnswers is how many of the latest answers the words for the model name.
// The payload holds the whole window; a sentence that listed twenty answers
// would be one nobody relays.
const textAnswers = 5

// progressOut is the progress as a card and a model read it. The child's
// details are the ones the progress screen shows, so that a card opening the
// progress inside itself draws the whole screen from one call. The parent's
// notes are not among them: the progress is what the child opens from a task
// card, and the notes are the parent's words about the child, which that screen
// has no reason to carry.
type progressOut struct {
	Screen         string             `json:"screen"`
	LastAnswer     *answerLine        `json:"last_answer"`
	Profile        *childOut          `json:"profile"`
	Trial          *trialOut          `json:"trial"`
	Overall        *standingOut       `json:"overall"`
	Topics         []topicOut         `json:"topics"`
	Recent         []recentOut        `json:"recent"`
	Mistakes       []mistakeOut       `json:"mistakes"`
	Recommendation *recommendationOut `json:"recommendation"`
}

// standingOut is the overall rating and the rank above it, out of how many
// ranks there are, so that a card need not know the number.
type standingOut struct {
	Rating int `json:"rating"`
	Rank   int `json:"rank"`
	Ranks  int `json:"ranks"`
}

// topicOut is one topic the child has answered or skipped a task of. Its
// rating is a plain number, null while the trial series runs and before its
// first answer, and it has no rank of its own. The skipped tasks are for the
// parent to see.
type topicOut struct {
	Topic    string `json:"topic"`
	Rating   *int   `json:"rating"`
	Answers  int    `json:"answers"`
	Correct  int    `json:"correct"`
	Mastered bool   `json:"mastered"`
	Skipped  int    `json:"skipped"`
}

// recentOut is one of the latest entries: an answer, or a task left without
// one, which is marked skipped and says nothing of right or wrong.
type recentOut struct {
	Topic      string `json:"topic"`
	Correct    *bool  `json:"correct"`
	Skipped    bool   `json:"skipped"`
	AnsweredAt string `json:"answered_at"`
}

// mistakeOut is a trap the child keeps falling for among the latest answers:
// its catalog id — the card names it in its own language, and the model has
// the catalog's words for it — and how many answers it was behind. Neither a
// task nor an answer is in it.
type mistakeOut struct {
	Trap  string `json:"trap"`
	Times int    `json:"times"`
}

func (s *Service) getProgressTool() Tool {
	return Define(Spec{
		Name:  "get_progress",
		Title: "Show the child's progress",
		Description: "Shows the child's progress: the overall rating on a chess-like scale with its rank, the rating " +
			"of each topic met, the topics mastered, the latest answers, the mistakes that keep coming back and " +
			"what comes next. The scale is one for grades 1 to 6, so an older child's number is higher. During " +
			"the trial series — the first five tasks — " +
			"there is no rating yet, only how many of the five are done. Present it encouragingly and name one " +
			"thing to practise next.",
		ReadOnly:   true,
		Idempotent: true,
		DrawsCard:  true,
	}, s.progress)
}

func (s *Service) readProgressTool() Tool {
	return Define(Spec{
		Name:  "read_progress",
		Title: "Read the child's progress for the card",
		Description: "The progress get_progress shows, for a card that opens it inside itself. " +
			"Only the card calls it.",
		ReadOnly:   true,
		Idempotent: true,
		WidgetOnly: true,
	}, s.progress)
}

// progress is the one answer of both progress tools: they differ in who calls
// them and in whether a host draws a card, never in what they say.
func (s *Service) progress(ctx context.Context, account store.Account, _ noArguments) (Reply[progressOut], error) {
	return afresh(ctx, func() (Reply[progressOut], error) { return s.readProgress(ctx, account) })
}

// readProgress is one read of the profile, and the progress it holds.
func (s *Service) readProgress(ctx context.Context, account store.Account) (Reply[progressOut], error) {
	p, _, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[progressOut]{Text: firstRunText, Payload: progressOut{
			Screen:   screenFirstRun,
			Topics:   []topicOut{},
			Recent:   []recentOut{},
			Mistakes: []mistakeOut{},
		}}, nil
	case err != nil:
		return Reply[progressOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	summary, err := progress.Of(p, s.content)
	if err != nil {
		return Reply[progressOut]{}, err
	}
	mistakes := progress.Mistakes(p.Recent, s.content, s.repeats)
	return Reply[progressOut]{
		Text: s.progressText(p, &summary, mistakes),
		Payload: progressOut{
			Screen:         screenProgress,
			LastAnswer:     lastAnswerOf(p),
			Profile:        childOf(&p.Student),
			Trial:          trialOf(summary.Trial),
			Overall:        standingOf(summary.Overall),
			Topics:         topicsOf(summary.Topics),
			Recent:         recentOf(summary.Recent),
			Mistakes:       mistakesOf(mistakes),
			Recommendation: recommendationOf(&summary.Next),
		},
	}, nil
}

func standingOf(overall *progress.Standing) *standingOut {
	if overall == nil {
		return nil
	}
	return &standingOut{Rating: overall.Rating, Rank: overall.Rank, Ranks: rating.Ranks}
}

func topicsOf(topics []progress.Topic) []topicOut {
	out := make([]topicOut, 0, len(topics))
	for _, topic := range topics {
		out = append(out, topicOut{
			Topic:    topic.ID,
			Rating:   topic.Rating,
			Answers:  topic.Answers,
			Correct:  topic.Correct,
			Mastered: topic.Mastered,
			Skipped:  topic.Skipped,
		})
	}
	return out
}

func mistakesOf(mistakes []progress.Mistake) []mistakeOut {
	out := make([]mistakeOut, 0, len(mistakes))
	for _, mistake := range mistakes {
		out = append(out, mistakeOut{Trap: mistake.Trap, Times: mistake.Times})
	}
	return out
}

func recentOf(entries []profile.Answer) []recentOut {
	out := make([]recentOut, 0, len(entries))
	for i := range entries {
		entry := recentOut{Topic: entries[i].Topic, Skipped: entries[i].Skipped, AnsweredAt: moment(entries[i].AnsweredAt)}
		if !entries[i].Skipped {
			correct := entries[i].Correct
			entry.Correct = &correct
		}
		out = append(out, entry)
	}
	return out
}

// progressText is the progress in words.
func (s *Service) progressText(p *profile.Profile, summary *progress.Summary, mistakes []progress.Mistake) string {
	standing := ""
	if summary.Overall != nil {
		standing = fmt.Sprintf("Overall rating %d, rank %d of %d.",
			summary.Overall.Rating, summary.Overall.Rank, rating.Ranks)
	}
	return joined(
		"The progress of "+quoted(p.Student.Pseudonym)+".",
		trialLine(summary.Trial),
		standing,
		s.topicsText(summary.Topics),
		s.recentText(summary.Recent),
		s.mistakesText(mistakes),
		s.nextText(&summary.Next),
	)
}

// mistakesText names the mistakes that keep coming back among the latest
// answers, the most frequent first, each by what the catalog says of it, or
// nothing when none does.
func (s *Service) mistakesText(mistakes []progress.Mistake) string {
	if len(mistakes) == 0 {
		return ""
	}
	parts := make([]string, 0, len(mistakes))
	for _, mistake := range mistakes {
		parts = append(parts, fmt.Sprintf("%s (%d times)", s.trapDescribed(mistake.Trap), mistake.Times))
	}
	return "Mistakes that keep coming back among the latest answers, the most frequent first: " +
		strings.Join(parts, "; ") + "."
}

// topicsText names every topic met and what it is, with its rating once there
// is one, and the tasks of it the child left without an answer.
func (s *Service) topicsText(topics []progress.Topic) string {
	if len(topics) == 0 {
		return "No topic has been answered yet."
	}
	parts := make([]string, 0, len(topics))
	for _, topic := range topics {
		part := fmt.Sprintf("%s, %d of %d right", s.topicDescribed(topic.ID), topic.Correct, topic.Answers)
		if topic.Rating != nil {
			part += fmt.Sprintf(", rating %d", *topic.Rating)
		}
		if topic.Mastered {
			part += ", mastered"
		}
		if topic.Skipped > 0 {
			part += fmt.Sprintf(", %d skipped", topic.Skipped)
		}
		parts = append(parts, part)
	}
	return "Topics: " + strings.Join(parts, "; ") + "."
}

// recentText names the latest few answers, the latest first, and how many of
// the tasks the window holds were left without one. The skips are counted
// apart, so that a run of them never hides how the answers went, and against
// the tasks they are counted among, so that a few old ones are not read as a
// habit of today.
func (s *Service) recentText(entries []profile.Answer) string {
	answers := make([]string, 0, textAnswers)
	skipped := 0
	for i := range entries {
		switch {
		case entries[i].Skipped:
			skipped++
		case len(answers) < textAnswers:
			answers = append(answers, s.topicName(entries[i].Topic)+" "+howItWent(entries[i].Correct))
		}
	}
	latest := ""
	if len(answers) > 0 {
		latest = "Latest answers, the latest first: " + strings.Join(answers, ", ") + "."
	}
	if skipped == 0 {
		return latest
	}
	return joined(latest, fmt.Sprintf("Of the last %d tasks, %d were left without an answer.", len(entries), skipped))
}
