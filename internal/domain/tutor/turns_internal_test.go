package tutor

import (
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// oldestBefore is the order of the turns as it was before a topic behind
// counted its wait twice: the day each topic was last given, the older first,
// and between two given on one day, the one whose last task stands earlier
// among the tasks behind the child, catalog order settling the rest.
func oldestBefore(p *profile.Profile, candidates []string) string {
	places := lastPlaces(p)
	best := candidates[0]
	for _, topic := range candidates[1:] {
		mine, theirs := p.Topics[topic].LastIssued, p.Topics[best].LastIssued
		var longer bool
		switch {
		case mine.IsZero() && theirs.IsZero():
		case mine.IsZero():
			longer = true
		case theirs.IsZero():
		case !mine.Equal(theirs.Time):
			longer = mine.Before(theirs.Time)
		default:
			longer = places[topic] < places[best]
		}
		if longer {
			best = topic
		}
	}
	return best
}

// history is a child's turns in plain values: for each of six topics the
// days since it was last given, or −1 for never; the topics of the tasks in
// the window, oldest first; and the topic of a task on the card waiting for
// its answer, or −1 for none.
type history struct {
	Days     []int
	Window   []int
	InFlight int
}

var historyTopics = []string{"a.one", "b.two", "c.three", "d.four", "e.five", "f.six"}

func genHistory() gopter.Gen {
	return gen.Struct(reflect.TypeOf(history{}), map[string]gopter.Gen{
		"Days":     gen.SliceOfN(len(historyTopics), gen.IntRange(-1, 3)),
		"Window":   gen.SliceOf(gen.IntRange(0, len(historyTopics)-1)),
		"InFlight": gen.IntRange(-1, len(historyTopics)-1),
	}).Map(func(h history) *history { return &h })
}

// build is a child past the trial series whose topics all stand at the
// overall level, so that none is behind, with the days and the window of h.
func (h *history) build() *profile.Profile {
	today := time.Date(2026, 10, 12, 18, 0, 0, 0, time.UTC)
	p := profile.New(profile.Student{Grade: 3, Pseudonym: "Otter"}, "0.0.0-test", today)
	p.Ratings.Answers = 60
	for i, topic := range historyTopics {
		summary := profile.Topic{Answers: 4, Correct: 2}
		if days := h.Days[i]; days >= 0 {
			summary.LastIssued = profile.DateOf(today.AddDate(0, 0, -days))
		}
		p.Topics[topic] = summary
	}
	for i, at := range h.Window[:min(len(h.Window), profile.MaxRecent)] {
		p.Recent = append(p.Recent, profile.Answer{
			AnsweredAt: profile.At(today.Add(time.Duration(i) * time.Minute)),
			Topic:      historyTopics[at],
			Skipped:    true,
		})
	}
	if h.InFlight >= 0 {
		p.CurrentTask = &profile.CurrentTask{ID: "tsk_on_the_card", Topic: historyTopics[h.InFlight]}
	}
	return p
}

// A topic given today that the window no longer holds a task of was given
// before all the window holds, and comes first, however far back in the
// window the last task of a topic behind stands: the distance a topic behind
// counts twice is one within the window.
func TestATopicOutOfTheWindowComesBeforeOneBehind(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 10, 12, 18, 0, 0, 0, time.UTC)
	p := profile.New(profile.Student{Grade: 3, Pseudonym: "Otter"}, "0.0.0-test", today)
	p.Ratings.Answers = 60
	for topic, delta := range map[string]float64{"a.one": 0, "b.two": -1, "c.three": 0} {
		p.Topics[topic] = profile.Topic{Answers: 4, Correct: 2, Delta: delta, LastIssued: profile.DateOf(today)}
	}
	if !p.Behind("b.two") || p.Behind("a.one") {
		t.Fatalf("the child is not behind in b.two alone")
	}
	// b.two's last task stands fifth of twenty; a.one has none in the window.
	for i := range profile.MaxRecent {
		topic := "c.three"
		if i == 4 {
			topic = "b.two"
		}
		p.Recent = append(p.Recent, profile.Answer{AnsweredAt: profile.At(today.Add(time.Duration(i) * time.Minute)),
			Topic: topic, Skipped: true})
	}

	if got := oldest(p, []string{"a.one", "b.two"}); got != "a.one" {
		t.Errorf("oldest() = %s, want a.one, given before all the window holds", got)
	}
}

// Two topics given today that the window holds no task of go in catalog
// order, a topic behind among them or not: the distance a topic behind counts
// twice is one within the window.
func TestTwoTopicsOutOfTheWindowGoInCatalogOrder(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 10, 12, 18, 0, 0, 0, time.UTC)
	p := profile.New(profile.Student{Grade: 3, Pseudonym: "Otter"}, "0.0.0-test", today)
	p.Ratings.Answers = 60
	for topic, delta := range map[string]float64{"a.one": 0, "b.two": -1, "c.three": 0} {
		p.Topics[topic] = profile.Topic{Answers: 4, Correct: 2, Delta: delta, LastIssued: profile.DateOf(today)}
	}
	for i := range profile.MaxRecent {
		p.Recent = append(p.Recent, profile.Answer{AnsweredAt: profile.At(today.Add(time.Duration(i) * time.Minute)),
			Topic: "c.three", Skipped: true})
	}

	if got := oldest(p, []string{"a.one", "b.two"}); got != "a.one" {
		t.Errorf("oldest() = %s, want a.one, first in catalog order of two the window holds no task of", got)
	}
}

func TestWithNoTopicBehindTheTurnsAreAsTheyWere(t *testing.T) {
	t.Parallel()
	properties := gopter.NewProperties(nil)

	// The candidates are the topics from one place to another: two generators,
	// one for each end.
	firstPlaces, lastPlaces := gen.IntRange(0, len(historyTopics)-1), gen.IntRange(0, len(historyTopics)-1)

	properties.Property("with no topic behind, the topic that has waited longest is the one it was before "+
		"waits were weighed", prop.ForAll(
		func(h *history, from, to int) bool {
			p := h.build()
			for _, topic := range historyTopics {
				if p.Behind(topic) {
					return false
				}
			}
			candidates := historyTopics[min(from, to) : max(from, to)+1]
			return oldest(p, candidates) == oldestBefore(p, candidates)
		},
		genHistory(), firstPlaces, lastPlaces,
	))

	properties.TestingRun(t)
}
