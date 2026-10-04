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

// recentNamed is how many of the latest entries the words for the model name:
// as many as the card lists, a skip among them where it fell, so that the two
// say the same. The payload holds the whole window; a sentence that listed
// twenty entries would be one nobody relays.
const recentNamed = 5

// progressOut is the progress as a card and a model read it. The child's
// details are the ones the progress screen shows, and where the profile's file
// is, so that a card opening the progress inside itself draws the whole
// screen, the parent's data with it, from one call. The parent's notes are not
// among them: the progress is what the child opens from a task card, and the
// notes are the parent's words about the child, which that screen has no
// reason to carry.
type progressOut struct {
	Screen         string             `json:"screen"`
	LastAnswer     *answerLine        `json:"last_answer"`
	Profile        *childOut          `json:"profile"`
	Trial          *trialOut          `json:"trial"`
	Overall        *standingOut       `json:"overall"`
	Topics         []topicOut         `json:"topics"`
	Recent         []recentOut        `json:"recent"`
	Skipped        int                `json:"skipped"`
	Mistakes       []mistakeOut       `json:"mistakes"`
	Recommendation *recommendationOut `json:"recommendation"`
	Location       *locationOut       `json:"location,omitempty"`
}

// standingOut is the overall rating, its rank out of how many ranks there are,
// so that a card need not know the number, and how far through the rank it has
// come, as a whole percent; and how it moved since the last answer and over
// the week, absent where neither can be told.
type standingOut struct {
	Rating int         `json:"rating"`
	Rank   int         `json:"rank"`
	Ranks  int         `json:"ranks"`
	Share  int         `json:"share"`
	Change *changesOut `json:"change,omitempty"`
}

// topicOut is one topic the child has met, or the rule could set now. Its
// rating, rank and share of the way through the rank, and how the rank stands
// to the overall one, are null together: while the trial series runs, and
// before its first answer; how it moved by its own answers is absent with
// them. The skipped tasks are for the parent to see.
type topicOut struct {
	Topic    string      `json:"topic"`
	Rating   *int        `json:"rating"`
	Rank     *int        `json:"rank"`
	Share    *int        `json:"share"`
	Compared *string     `json:"compared"`
	Answers  int         `json:"answers"`
	Correct  int         `json:"correct"`
	Mastered bool        `json:"mastered"`
	Skipped  int         `json:"skipped"`
	Change   *changesOut `json:"change,omitempty"`
}

// changesOut is how a standing moved since the last answer and over the week
// of seven dates in UTC ending today, each null where it cannot be told.
type changesOut struct {
	LastTask *changeOut `json:"last_task"`
	Week     *changeOut `json:"week"`
}

// changeOut is one move: where the standing stood before, in what a card draws
// of a standing — its rank and how far through it — and the word for the move.
// The overall standing's earlier rating comes with it, as its rating does, for
// the model. A topic's is left out: no card draws a topic's rating, and this
// one, measured by the topic's own answers from where the overall level stands
// now, is no rating the topic ever had. A topic first answered in the while
// stood nowhere, and has only its word.
type changeOut struct {
	Rating *int   `json:"rating,omitempty"`
	Rank   *int   `json:"rank,omitempty"`
	Share  *int   `json:"share,omitempty"`
	Moved  string `json:"moved"`
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
		Description: "Shows the child's progress: the overall rating on a chess-like scale with its rank and how far " +
			"through the rank it has come, each topic met or within reach now with a rank of its own, ahead of or " +
			"behind the overall one, the topics mastered, the latest answers and how many tasks were left without " +
			"one, the mistakes that keep coming back, what comes next, and where the profile's file is kept. It " +
			"also tells what moved since the last answer and over the last seven days, each topic by its own " +
			"answers: say a step back gently, as part of learning, never as a score against the child. Its " +
			"card draws the rank, not the rating's number: say the number yourself. The scale is one for grades 1 " +
			"to 6, so an older child's number is higher. During the trial series — the first five tasks — there " +
			"is no rating yet, only how many of the five are done. Call it only when someone asks to see the " +
			"progress. Present it encouragingly and name one thing to practise next.",
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

// readProgress is one read of the profile, the progress it holds, and where it
// is kept.
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
	location, where, err := s.whereKept(ctx, account)
	if err != nil {
		return Reply[progressOut]{}, err
	}

	summary, err := progress.Of(p, s.content, profile.DateOf(s.now()))
	if err != nil {
		return Reply[progressOut]{}, err
	}
	mistakes := progress.Mistakes(p.Recent, s.content, s.repeats)
	return Reply[progressOut]{
		Text: joined(s.progressText(p, &summary, mistakes), where),
		Payload: progressOut{
			Screen:         screenProgress,
			LastAnswer:     lastAnswerOf(p),
			Profile:        childOf(&p.Student),
			Trial:          trialOf(summary.Trial),
			Overall:        standingOf(summary.Overall, summary.Changes),
			Topics:         topicsOf(summary.Topics),
			Recent:         recentOf(summary.Recent),
			Skipped:        summary.Skipped,
			Mistakes:       mistakesOf(mistakes),
			Recommendation: recommendationOf(&summary.Next),
			Location:       location,
		},
	}, nil
}

func standingOf(overall *progress.Standing, changes progress.Changes) *standingOut {
	if overall == nil {
		return nil
	}
	return &standingOut{
		Rating: overall.Rating,
		Rank:   overall.Rank,
		Ranks:  rating.Ranks,
		Share:  overall.Share,
		Change: changesOf(changes, true),
	}
}

// changesOf is how a standing moved, with its earlier rating where it is one
// a card drew, or nil where neither move can be told.
func changesOf(changes progress.Changes, withRating bool) *changesOut {
	if changes.LastTask == nil && changes.Week == nil {
		return nil
	}
	return &changesOut{
		LastTask: changeOf(changes.LastTask, withRating),
		Week:     changeOf(changes.Week, withRating),
	}
}

func changeOf(change *progress.Change, withRating bool) *changeOut {
	if change == nil {
		return nil
	}
	out := &changeOut{Moved: string(change.Moved)}
	if before := change.Before; before != nil {
		out.Rank, out.Share = &before.Rank, &before.Share
		if withRating {
			out.Rating = &before.Rating
		}
	}
	return out
}

func topicsOf(topics []progress.Topic) []topicOut {
	out := make([]topicOut, 0, len(topics))
	for _, topic := range topics {
		entry := topicOut{
			Topic:    topic.ID,
			Answers:  topic.Answers,
			Correct:  topic.Correct,
			Mastered: topic.Mastered,
			Skipped:  topic.Skipped,
		}
		if standing := topic.Standing; standing != nil {
			compared := string(topic.Compared)
			entry.Rating, entry.Rank, entry.Share = &standing.Rating, &standing.Rank, &standing.Share
			entry.Compared = &compared
			entry.Change = changesOf(topic.Changes, false)
		}
		out = append(out, entry)
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
	return joined(
		"The progress of "+quoted(p.Student.Pseudonym)+".",
		trialLine(summary.Trial),
		standingText(summary.Overall),
		s.changeText(summary),
		s.topicsText(summary.Topics),
		s.recentText(summary.Recent, summary.Skipped),
		s.mistakesText(mistakes),
		s.nextText(&summary.Next),
	)
}

// standingText is the overall rating in words: its rank, and how far through
// the rank it has come, or nothing while the trial series runs.
func standingText(overall *progress.Standing) string {
	if overall == nil {
		return ""
	}
	if overall.Rank == rating.Ranks {
		return fmt.Sprintf("Overall rating %d, rank %d of %d, the highest.", overall.Rating, overall.Rank, rating.Ranks)
	}
	return fmt.Sprintf("Overall rating %d, rank %d of %d, %d%% of the way to rank %d.",
		overall.Rating, overall.Rank, rating.Ranks, overall.Share, overall.Rank+1)
}

// topicMoves are the topics' moves the words name, in the order they name
// them, each with the words for it. A topic's move is its own: what its own
// answers moved, beside what the overall level did.
var topicMoves = []struct {
	move  progress.Move
	words string
}{
	{progress.RankUp, "a rank up by their own answers"},
	{progress.Forward, "forward by their own answers"},
	{progress.New, "answered for the first time"},
	{progress.Back, "back by their own answers"},
	{progress.RankDown, "a rank down by their own answers"},
}

// changeText is how the overall rating and the topics moved since the last
// answer and over the last seven days, in words, saying nothing of a while
// that cannot be told.
func (s *Service) changeText(summary *progress.Summary) string {
	return joined(
		s.whileText("Since the last answer", summary, func(changes progress.Changes) *progress.Change {
			return changes.LastTask
		}),
		s.whileText("Over the last seven days", summary, func(changes progress.Changes) *progress.Change {
			return changes.Week
		}),
	)
}

// whileText is the moves of one while: the overall rating from where it
// stood to where it stands, and the topics whose own answers moved them, by
// how. A while with no move of the overall rating to tell says nothing.
func (s *Service) whileText(lead string, summary *progress.Summary, of func(progress.Changes) *progress.Change) string {
	overall := of(summary.Changes)
	if overall == nil {
		return ""
	}
	parts := []string{overallMoveText(overall.Before, summary.Overall)}
	for _, kind := range topicMoves {
		var named []string
		for i := range summary.Topics {
			if change := of(summary.Topics[i].Changes); change != nil && change.Moved == kind.move {
				named = append(named, s.topicName(summary.Topics[i].ID))
			}
		}
		if len(named) > 0 {
			parts = append(parts, kind.words+": "+strings.Join(named, ", "))
		}
	}
	return lead + ", " + strings.Join(parts, "; ") + "."
}

// overallMoveText is the overall rating from where it stood to where it
// stands, and its rank with it when the rank changed.
func overallMoveText(before, now *progress.Standing) string {
	words := fmt.Sprintf("the overall rating stayed at %d", now.Rating)
	if before.Rating != now.Rating {
		words = fmt.Sprintf("the overall rating went from %d to %d", before.Rating, now.Rating)
	}
	if before.Rank != now.Rank {
		words += fmt.Sprintf(", rank %d to rank %d", before.Rank, now.Rank)
	}
	return words
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

// topicsText names every topic met and what it is, with its rating and rank
// once there are some, how the rank stands to the overall one, and the tasks
// of it the child left without an answer; then, by name, the topics within
// reach that the child has not met yet.
func (s *Service) topicsText(topics []progress.Topic) string {
	met := make([]string, 0, len(topics))
	var unmet []string
	for _, topic := range topics {
		if topic.Answers == 0 && topic.Skipped == 0 {
			unmet = append(unmet, s.topicName(topic.ID))
			continue
		}
		met = append(met, s.topicMet(&topic))
	}
	words := "No topic has been answered yet."
	if len(met) > 0 {
		words = "Topics: " + strings.Join(met, "; ") + "."
	}
	if len(unmet) == 0 {
		return words
	}
	return joined(words, "Not met yet, and within reach now: "+strings.Join(unmet, ", ")+".")
}

// topicMet is one topic the child has met, in words.
func (s *Service) topicMet(topic *progress.Topic) string {
	part := s.topicDescribed(topic.ID) + ", no answers yet"
	if topic.Answers > 0 {
		part = fmt.Sprintf("%s, %d of %d right", s.topicDescribed(topic.ID), topic.Correct, topic.Answers)
	}
	if standing := topic.Standing; standing != nil {
		part += fmt.Sprintf(", rating %d, rank %d", standing.Rating, standing.Rank)
		switch topic.Compared {
		case progress.Ahead:
			part += ", ahead of the overall rank"
		case progress.Behind:
			part += ", behind the overall rank"
		}
	}
	if topic.Mastered {
		part += ", mastered"
	}
	if topic.Skipped > 0 {
		part += fmt.Sprintf(", %d skipped", topic.Skipped)
	}
	return part
}

// recentText names the latest entries, the latest first, as many as the card
// lists — a task left without an answer among them as skipped — and how many
// tasks were left without an answer in all.
func (s *Service) recentText(entries []profile.Answer, skipped int) string {
	named := make([]string, 0, recentNamed)
	for i := range entries[:min(len(entries), recentNamed)] {
		went := "skipped"
		if !entries[i].Skipped {
			went = howItWent(entries[i].Correct)
		}
		named = append(named, s.topicName(entries[i].Topic)+" "+went)
	}
	latest := ""
	if len(named) > 0 {
		latest = "The latest tasks, the latest first: " + strings.Join(named, ", ") + "."
	}
	if skipped == 0 {
		return latest
	}
	return joined(latest, fmt.Sprintf("Tasks left without an answer in all: %d.", skipped))
}
