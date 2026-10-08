package mcpserver_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// changesPayload is how a standing moved, as a card reads it: since the last
// answer and over the week, each null where it cannot be told.
type changesPayload struct {
	LastTask *changePayload `json:"last_task"`
	Week     *changePayload `json:"week"`
}

// changePayload is one move: where the standing stood, as it was drawn, and
// the word for the move.
type changePayload struct {
	Rating *int   `json:"rating"`
	Rank   *int   `json:"rank"`
	Share  *int   `json:"share"`
	Moved  string `json:"moved"`
}

// The progress tells what the answers moved, from the history recording them
// keeps: a child past the trial series answers a task one day and another of
// the same topic the next, and the card is told what the progress says moved —
// the overall rating since the last answer and over the week, and every topic
// by its own answers — the same whichever tool asks, and read without writing
// anything.
func TestTheProgressTellsWhatTheAnswersMoved(t *testing.T) {
	t.Parallel()

	p := raceOnTheCard(t, rating.TrialAnswers)
	kept, moving := keptAsIs(t, p), &clock{at: lessonDay}
	_, session := lessonWith(t, kept, moving, nil)
	answered(t, answerIt(t, session, p.CurrentTask.ID, "C", false))
	moving.advance(24 * time.Hour)
	handOutAnother(t, kept)
	answered(t, answerIt(t, session, loadedOf(t, kept).CurrentTask.ID, "B", false))

	stored, revision := loadKept(t, kept)
	forModel, forCard := call(t, session, "get_progress", nil), call(t, session, "read_progress", nil)
	if !bytes.Equal(rawPayload(t, forModel), rawPayload(t, forCard)) {
		t.Errorf("get_progress and read_progress differ:\n%s\n%s", rawPayload(t, forModel), rawPayload(t, forCard))
	}
	if _, after := loadKept(t, kept); after != revision {
		t.Errorf("reading the progress left the profile at revision %s, want it unwritten at %s", after, revision)
	}

	loaded, err := shipped()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	summary, err := progress.Of(stored, loaded, profile.DateOf(moving.now()))
	if err != nil {
		t.Fatalf("Of() error = %v, want nil", err)
	}
	payload := payloadOf[progressPayload](t, forModel)
	if payload.Overall == nil || payload.Overall.Change == nil || len(payload.Topics) != len(summary.Topics) {
		t.Fatalf("progress = %+v, want the overall rating with its change and the topics the progress lists", payload)
	}
	wantMoveOnTheCard(t, "the overall rating since the last answer", payload.Overall.Change.LastTask, summary.Changes.LastTask, true)
	wantMoveOnTheCard(t, "the overall rating over the week", payload.Overall.Change.Week, summary.Changes.Week, true)
	for i, topic := range payload.Topics {
		wantTopicMovesOnTheCard(t, &topic, &summary.Topics[i])
	}

	ordering := summary.Topics[0]
	if ordering.ID != "logic.ordering" || ordering.Changes.LastTask == nil || ordering.Changes.LastTask.Moved == progress.New ||
		ordering.Changes.Week == nil || ordering.Changes.Week.Moved != progress.New {
		t.Errorf("%s changed %+v, want a move of its own since its second answer, and new over the week of its first",
			ordering.ID, ordering.Changes)
	}
	words := textOf(t, forModel)
	for _, want := range []string{"Since the last answer, the overall rating", "Over the last seven days, the overall rating",
		"answered for the first time: Ordering"} {
		if !strings.Contains(words, want) {
			t.Errorf("the words are %q, want them to say %q", words, want)
		}
	}
}

// wantTopicMovesOnTheCard fails the test unless a topic's moves reached the
// card as the progress tells them: absent for a topic that stands nowhere, and
// never with an earlier rating, which no card drew.
func wantTopicMovesOnTheCard(t *testing.T, got *topicPayload, want *progress.Topic) {
	t.Helper()

	if want.Changes == (progress.Changes{}) {
		if got.Change != nil {
			t.Errorf("%s changed %+v on the card, want nothing told", got.Topic, *got.Change)
		}
		return
	}
	if got.Change == nil {
		t.Errorf("%s has no change on the card, want %+v", got.Topic, want.Changes)
		return
	}
	wantMoveOnTheCard(t, got.Topic+" since the last answer", got.Change.LastTask, want.Changes.LastTask, false)
	wantMoveOnTheCard(t, got.Topic+" over the week", got.Change.Week, want.Changes.Week, false)
}

// wantMoveOnTheCard fails the test unless one move reached the card as the
// progress tells it: null where it cannot be told, its word, and where it
// stood — its earlier rating only where a card drew that rating.
func wantMoveOnTheCard(t *testing.T, what string, got *changePayload, want *progress.Change, withRating bool) {
	t.Helper()

	switch {
	case want == nil && got != nil:
		t.Errorf("%s = %+v, want null", what, *got)
	case want == nil:
	case got == nil:
		t.Errorf("%s = null, want %s", what, want.Moved)
	case got.Moved != string(want.Moved):
		t.Errorf("%s moved %q, want %q", what, got.Moved, want.Moved)
	case want.Before == nil:
		if got.Rating != nil || got.Rank != nil || got.Share != nil {
			t.Errorf("%s = %+v, want it to stand nowhere before", what, *got)
		}
	case got.Rank == nil || *got.Rank != want.Before.Rank || got.Share == nil || *got.Share != want.Before.Share:
		t.Errorf("%s stood at rank %v, share %v, want %d, %d", what, got.Rank, got.Share, want.Before.Rank, want.Before.Share)
	case (got.Rating != nil) != withRating || withRating && *got.Rating != want.Before.Rating:
		t.Errorf("%s stood at rating %v, want %d told: %v", what, got.Rating, want.Before.Rating, withRating)
	}
}

// levelAt is the level whose rating is elo.
func levelAt(elo float64) float64 { return (elo - rating.Elo(0)) / (rating.Elo(1) - rating.Elo(0)) }

// petyaMoved is Petya, his overall rating at 1985, whose last answer, his
// second in the pigeonhole principle, moved his overall rating from
// overallBefore and that topic, measured from the overall level now, from
// topicBefore to 2010.
func petyaMoved(overallBefore, topicBefore float64) func(p *profile.Profile) {
	return func(p *profile.Profile) {
		p.Ratings.Theta = levelAt(1985)
		topic := p.Topics["pigeonhole.basic"]
		topic.Answers, topic.Correct, topic.Delta = 2, 2, levelAt(2010)-p.Ratings.Theta
		p.Topics["pigeonhole.basic"] = topic
		p.Recent[len(p.Recent)-1].Before = &profile.Levels{
			Delta: levelAt(topicBefore) - p.Ratings.Theta,
			Theta: levelAt(overallBefore),
		}
	}
}

// petyaThisWeek is Petya as petyaMoved(1966, 1990) leaves him, whose week
// began five days ago at an overall rating of 1800 and with every topic's
// correction where it is now, but for two: Ordering stood at 2100, measured
// from the overall level now, and the pigeonhole principle was not yet
// answered.
func petyaThisWeek(p *profile.Profile) {
	petyaMoved(1966, 1990)(p)
	deltas := map[string]float64{}
	for id, topic := range p.Topics {
		deltas[id] = topic.Delta
	}
	deltas["logic.ordering"] = levelAt(2100) - p.Ratings.Theta
	delete(deltas, "pigeonhole.basic")
	p.RatingDays = []profile.RatingDay{{Date: profile.DateOf(lessonDay.AddDate(0, 0, -5)), Deltas: deltas, Theta: levelAt(1800)}}
}

// The words say what moved in each while: the overall rating from where it
// stood to where it stands, with its rank when the rank changed, and then the
// topics whose own answers moved them, by how; a topic nothing of its own
// moved is not named, and a while that cannot be told is not spoken of. A step
// back, of the overall rating or of a topic, is to be told gently, and words
// that name none say nothing of it.
func TestTheWordsSayWhatMovedEachTopicByItsOwnAnswers(t *testing.T) {
	t.Parallel()

	const gently = "Say a step back gently, as part of learning, never as a score against the child."
	for _, tc := range []struct {
		name   string
		change func(p *profile.Profile)
		want   []string
		unsaid []string
	}{
		{
			name:   "moves in both whiles",
			change: petyaThisWeek,
			want: []string{
				"Since the last answer, the overall rating went from 1966 to 1985; a rank up by their own answers: Pigeonhole principle.",
				"Over the last seven days, the overall rating went from 1800 to 1985, rank 4 to rank 5; " +
					"answered for the first time: Pigeonhole principle; back by their own answers: Ordering.",
				gently,
			},
		},
		{
			name:   "the overall rating down since the last answer",
			change: petyaMoved(2000, 2010),
			want:   []string{"Since the last answer, the overall rating went from 2000 to 1985", gently},
		},
		{
			name:   "nothing but steps forward",
			change: petyaMoved(1966, 1990),
			want:   []string{"Since the last answer, the overall rating went from 1966 to 1985"},
			unsaid: []string{"step back"},
		},
		{
			name:   "a quiet week after an answer that keeps nothing before it",
			change: func(*profile.Profile) {},
			want:   []string{"Over the last seven days, the overall rating stayed at 1985."},
			unsaid: []string{"Since the last answer", "step back"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, session := lesson(t, keptAs(t, "petya", tc.change))
			wantWords(t, textOf(t, call(t, session, "get_progress", nil)), tc.want, tc.unsaid)
		})
	}
}

// wantWords fails the test unless the words say every sentence wanted, and
// nothing of what is to stay unsaid.
func wantWords(t *testing.T, words string, want, unsaid []string) {
	t.Helper()

	for _, sentence := range want {
		if !strings.Contains(words, sentence) {
			t.Errorf("the words are %q, want them to say %q", words, sentence)
		}
	}
	for _, sentence := range unsaid {
		if strings.Contains(words, sentence) {
			t.Errorf("the words are %q, want nothing said %q", words, sentence)
		}
	}
}

// petyaFirstInTopic is Petya as petyaMoved(1900, 1990) leaves him, but with the
// last answer his first in its topic.
func petyaFirstInTopic(p *profile.Profile) {
	petyaMoved(1900, 1990)(p)
	topic := p.Topics["pigeonhole.basic"]
	topic.Answers, topic.Correct = 1, 1
	p.Topics["pigeonhole.basic"] = topic
}

// Every move reaches the card in the shape the tool declares, which the
// service holds every payload to: each word, a topic new with nothing before
// it, and a while that cannot be told as null.
func TestEveryMoveReachesTheCardInTheDeclaredShape(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		change  func(p *profile.Profile)
		overall string // the overall rating's move since the last answer, or "" for null
		topic   string // the pigeonhole principle's, or "" for null
	}{
		{"a rank up", petyaMoved(1800, 2010), "rank_up", "same"},
		{"forward", petyaMoved(1900, 1990), "forward", "rank_up"},
		{"no move drawn", petyaMoved(1985, 2005), "same", "forward"},
		{"back", petyaMoved(1990, 2030), "back", "back"},
		{"a rank down", petyaMoved(2010, 2200), "rank_down", "rank_down"},
		{"a topic answered for the first time", petyaFirstInTopic, "forward", "new"},
		{"an answer that keeps no levels before it", func(*profile.Profile) {}, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, session := lesson(t, keptAs(t, "petya", tc.change))
			payload := payloadOf[progressPayload](t, call(t, session, "get_progress", nil))
			if payload.Overall == nil || payload.Overall.Change == nil || payload.Overall.Change.Week == nil {
				t.Fatalf("overall = %+v, want a change with the week told", payload.Overall)
			}
			if got := movedOf(payload.Overall.Change.LastTask); got != tc.overall {
				t.Errorf("the overall rating moved %q since the last answer, want %q", got, tc.overall)
			}
			if got := topicMovedOf(payload.Topics, "pigeonhole.basic"); got != tc.topic {
				t.Errorf("the pigeonhole principle moved %q since the last answer, want %q", got, tc.topic)
			}
		})
	}
}

// A progress with nothing that can be told of what moved carries no change at
// all, rather than one of nulls: the trial series, and a child past it whose
// last answer keeps nothing before it in a week the history does not cover.
func TestAProgressWithNothingToTellCarriesNoChange(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		kept store.Storage
	}{
		{"the trial series", keptWith(t, "masha")},
		{"a week the history does not cover", keptAs(t, "petya", func(p *profile.Profile) {
			p.Recent[len(p.Recent)-1].AnsweredAt = profile.Time{Time: lessonDay}
		})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, session := lesson(t, tc.kept)
			if raw := rawPayload(t, call(t, session, "get_progress", nil)); bytes.Contains(raw, []byte(`"change"`)) {
				t.Errorf("the progress is %s, want no change in it", raw)
			}
		})
	}
}

// movedOf is the word of a move as the card reads it, or "" for null.
func movedOf(change *changePayload) string {
	if change == nil {
		return ""
	}
	return change.Moved
}

// topicMovedOf is the word of the move of the topic with the id since the last
// answer, as the card reads it, or "" when the card is told none.
func topicMovedOf(topics []topicPayload, id string) string {
	for _, topic := range topics {
		if topic.Topic == id && topic.Change != nil {
			return movedOf(topic.Change.LastTask)
		}
	}
	return ""
}
