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
	Recommendation *recommendationOut `json:"recommendation"`
}

// standingOut is the overall rating and the rank above it, out of how many
// ranks there are, so that a card need not know the number.
type standingOut struct {
	Rating int `json:"rating"`
	Rank   int `json:"rank"`
	Ranks  int `json:"ranks"`
}

// topicOut is one topic the child has answered. Its rating is a plain number,
// null while the trial series runs, and it has no rank of its own.
type topicOut struct {
	Topic    string `json:"topic"`
	Rating   *int   `json:"rating"`
	Answers  int    `json:"answers"`
	Correct  int    `json:"correct"`
	Mastered bool   `json:"mastered"`
}

// recentOut is one of the latest answers.
type recentOut struct {
	Topic      string `json:"topic"`
	Correct    bool   `json:"correct"`
	AnsweredAt string `json:"answered_at"`
}

func (s *Service) getProgressTool() Tool {
	return Define(Spec{
		Name:  "get_progress",
		Title: "Show the child's progress",
		Description: "Shows the child's progress: the overall rating on a chess-like scale with its rank, the rating " +
			"of each topic met, the topics mastered, the latest answers and what comes next. The scale is one for " +
			"grades 1 to 6, so an older child's number is higher. During the trial series — the first five tasks — " +
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
	p, _, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[progressOut]{Text: firstRunText, Payload: progressOut{
			Screen: screenFirstRun,
			Topics: []topicOut{},
			Recent: []recentOut{},
		}}, nil
	case err != nil:
		return Reply[progressOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	summary, err := progress.Of(p, s.content)
	if err != nil {
		return Reply[progressOut]{}, err
	}
	return Reply[progressOut]{
		Text: s.progressText(p, &summary),
		Payload: progressOut{
			Screen:         screenProgress,
			LastAnswer:     lastAnswerOf(p),
			Profile:        childOf(&p.Student),
			Trial:          trialOf(summary.Trial),
			Overall:        standingOf(summary.Overall),
			Topics:         topicsOf(summary.Topics),
			Recent:         recentOf(summary.Recent),
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
		})
	}
	return out
}

func recentOf(answers []profile.Answer) []recentOut {
	out := make([]recentOut, 0, len(answers))
	for i := range answers {
		out = append(out, recentOut{
			Topic:      answers[i].Topic,
			Correct:    answers[i].Correct,
			AnsweredAt: moment(answers[i].AnsweredAt),
		})
	}
	return out
}

// progressText is the progress in words.
func (s *Service) progressText(p *profile.Profile, summary *progress.Summary) string {
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
		s.nextText(&summary.Next),
	)
}

// topicsText names every topic met and what it is, with its rating once there
// is one.
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
		parts = append(parts, part)
	}
	return "Topics: " + strings.Join(parts, "; ") + "."
}

// recentText names the latest few answers, the latest first.
func (s *Service) recentText(answers []profile.Answer) string {
	if len(answers) == 0 {
		return ""
	}
	parts := make([]string, 0, textAnswers)
	for i := range answers[:min(len(answers), textAnswers)] {
		parts = append(parts, s.topicName(answers[i].Topic)+" "+howItWent(answers[i].Correct))
	}
	return "Latest answers, the latest first: " + strings.Join(parts, ", ") + "."
}
