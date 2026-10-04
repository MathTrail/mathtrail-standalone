package progress_test

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// What moved is told twice: since the last answer, and over the seven dates in
// UTC ending today. The overall rating is measured whole, and a topic by its
// own correction alone, from where the overall level stands now, so that a
// topic no answer of its own moved stands where it stood.

// lastTopic is the topic of Petya's last answer.
const lastTopic = "pigeonhole.basic"

// levelAt is the level whose rating is elo.
func levelAt(elo float64) float64 { return (elo - rating.Elo(0)) / (rating.Elo(1) - rating.Elo(0)) }

// drawn is the rating elo as it is drawn.
func drawn(elo float64) progress.Standing {
	return progress.Standing{Rating: rating.Shown(elo), Rank: rating.Rank(elo), Share: rating.Share(elo)}
}

// dated is a date of the days before today, in UTC.
func dated(month time.Month, day int) profile.Date {
	return profile.DateOf(time.Date(2026, month, day, 0, 0, 0, 0, time.UTC))
}

// answered is Petya after his last answer moved his overall rating from
// overallBefore to overallNow and his rating in the topic of the answer,
// measured from the overall level now, from topicBefore to topicNow. It is his
// second answer in the topic, so the topic stood somewhere before it. He last
// answered a month ago, so that the week has nothing in it.
func answered(t *testing.T, overallBefore, overallNow, topicBefore, topicNow float64) *profile.Profile {
	t.Helper()

	p := fixture(t, "petya")
	p.Ratings.Theta = levelAt(overallNow)
	topic := p.Topics[lastTopic]
	topic.Answers, topic.Correct = 2, 2
	topic.Delta = levelAt(topicNow) - p.Ratings.Theta
	p.Topics[lastTopic] = topic
	p.Recent[len(p.Recent)-1].Before = &profile.Levels{
		Delta: levelAt(topicBefore) - p.Ratings.Theta,
		Theta: levelAt(overallBefore),
	}
	return p
}

// topicOf is the topic of the summary with the id, failing the test when it is
// not listed.
func topicOf(t *testing.T, summary *progress.Summary, id string) progress.Topic {
	t.Helper()

	for _, topic := range summary.Topics {
		if topic.ID == id {
			return topic
		}
	}
	t.Fatalf("%s is not listed among %v", id, idsOf(summary.Topics))
	return progress.Topic{}
}

// sameChange says whether a change told is the one wanted: the same move, from
// the same standing as drawn, or from none.
func sameChange(got *progress.Change, want progress.Change) bool {
	if got == nil || got.Moved != want.Moved || (got.Before == nil) != (want.Before == nil) {
		return false
	}
	return got.Before == nil || *got.Before == *want.Before
}

// wantChange fails the test unless the change told is the one wanted.
func wantChange(t *testing.T, what string, got *progress.Change, want progress.Change) {
	t.Helper()

	if !sameChange(got, want) {
		t.Errorf("%s changed %s, want %s", what, describe(got), describe(&want))
	}
}

// describe is a change as a failure message says it.
func describe(change *progress.Change) string {
	switch {
	case change == nil:
		return "nothing that can be told"
	case change.Before == nil:
		return string(change.Moved) + " from nowhere"
	default:
		return fmt.Sprintf("%s from rank %d, %d%% through it, at %d",
			change.Moved, change.Before.Rank, change.Before.Share, change.Before.Rating)
	}
}

// A move is read off what is drawn — the rank, and how far through it the
// rating has come — so a point lost that the bar does not draw is no move,
// and neither is any move inside the highest rank, drawn full, or below the
// first rank's way, drawn empty.
func TestAMoveIsReadOffTheRankAndTheShareDrawn(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	for _, tc := range []struct {
		name        string
		before, now float64
		want        progress.Move
	}{
		{"a rank up", 1800, 1985, progress.RankUp},
		{"further through the rank", 1900, 1985, progress.Forward},
		{"a point the bar does not draw", 1986, 1985, progress.Same},
		{"less far through the rank", 1990, 1985, progress.Back},
		{"a rank down", 2010, 1985, progress.RankDown},
		{"into the highest rank", 2800, 2900, progress.RankUp},
		{"inside the highest rank, drawn full", 2900, 3200, progress.Same},
		{"onto the first rank's way", 1100, 1200, progress.Forward},
		{"below the first rank's way, drawn empty", 1000, 1160, progress.Same},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			summary := summaryOf(t, answered(t, tc.before, tc.now, 2010, 2010), catalog)
			wantChange(t, "the overall rating", summary.Changes.LastTask, progress.Change{Before: new(drawn(tc.before)), Moved: tc.want})
		})
	}
}

// The last answer moved the overall rating and the topic it was given in, and
// no other: every other topic stands where it stood, since its correction did
// not move, though the overall level under it did.
func TestTheLastAnswerMovesItsOwnTopicAlone(t *testing.T) {
	t.Parallel()

	p := answered(t, 1980, 1985, 1990, 2010)
	summary := summaryOf(t, p, embedded(t))

	wantChange(t, "the overall rating", summary.Changes.LastTask, progress.Change{Before: new(drawn(1980)), Moved: progress.Forward})
	others := 0
	for _, topic := range summary.Topics {
		switch {
		case topic.ID == lastTopic:
			wantChange(t, topic.ID, topic.Changes.LastTask, progress.Change{Before: new(drawn(1990)), Moved: progress.RankUp})
		case topic.Standing != nil:
			others++
			wantChange(t, topic.ID, topic.Changes.LastTask, progress.Change{Before: topic.Standing, Moved: progress.Same})
		default:
			if topic.Changes != (progress.Changes{}) {
				t.Errorf("%s, with no answer, changed %+v, want nothing told", topic.ID, topic.Changes)
			}
		}
	}
	if others == 0 {
		t.Fatal("no other topic stands anywhere, want some to check")
	}
}

// A topic the last answer was the first in stood nowhere before it: it is new,
// with no standing to have moved from.
func TestATopicTheLastAnswerWasTheFirstInIsNew(t *testing.T) {
	t.Parallel()

	p := answered(t, 1980, 1985, 1990, 2010)
	topic := p.Topics[lastTopic]
	topic.Answers, topic.Correct = 1, 1
	p.Topics[lastTopic] = topic
	summary := summaryOf(t, p, embedded(t))

	wantChange(t, lastTopic, topicOf(t, &summary, lastTopic).Changes.LastTask, progress.Change{Moved: progress.New})
}

// The last task is the last answer: a task left without an answer after it
// moved nothing, and changes nothing that is told.
func TestATaskLeftAfterTheLastAnswerChangesNothingTold(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	p := answered(t, 1980, 1985, 1990, 2010)
	want := summaryOf(t, p, catalog)
	last := p.Recent[len(p.Recent)-1]
	p.Recent = append(p.Recent, profile.Answer{
		AnsweredAt: last.AnsweredAt, Difficulty: last.Difficulty, GradeLevel: last.GradeLevel,
		Skipped: true, TaskID: "tsk_left", Topic: "counting.gaps",
	})

	got := summaryOf(t, p, catalog)
	if !reflect.DeepEqual(got.Changes, want.Changes) {
		t.Errorf("changes = %+v after a task left, want %+v", got.Changes, want.Changes)
	}
	if !reflect.DeepEqual(topicOf(t, &got, lastTopic).Changes, topicOf(t, &want, lastTopic).Changes) {
		t.Errorf("%s changed differently after a task left in another topic", lastTopic)
	}
}

// An answer an earlier build wrote keeps no levels before it, and what it moved
// cannot be told.
func TestWhatAnAnswerWithNoLevelsBeforeItMovedCannotBeTold(t *testing.T) {
	t.Parallel()

	summary := summaryOf(t, fixture(t, "petya"), embedded(t))

	if summary.Changes.LastTask != nil {
		t.Errorf("the overall rating changed %s, want nothing told", describe(summary.Changes.LastTask))
	}
	for _, topic := range summary.Topics {
		if topic.Changes.LastTask != nil {
			t.Errorf("%s changed %s, want nothing told", topic.ID, describe(topic.Changes.LastTask))
		}
	}
}

// During the trial series, and at its last answer, which gave the first rank,
// there is nothing to measure a move from — whatever the file keeps.
func TestNothingMovesDuringTheTrialSeriesNorAtItsLastAnswer(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	for _, student := range []string{"masha", "olya"} {
		t.Run(student, func(t *testing.T) {
			t.Parallel()

			p := fixture(t, student)
			p.Recent[len(p.Recent)-1].Before = &profile.Levels{Theta: p.Ratings.Theta - 0.5}
			p.RatingDays = []profile.RatingDay{{Date: today, Deltas: map[string]float64{}, Theta: p.Ratings.Theta - 0.5}}
			summary := summaryOf(t, p, catalog)

			if summary.Changes != (progress.Changes{}) {
				t.Errorf("the overall rating changed %+v, want nothing told", summary.Changes)
			}
			for _, topic := range summary.Topics {
				if topic.Changes != (progress.Changes{}) {
					t.Errorf("%s changed %+v, want nothing told", topic.ID, topic.Changes)
				}
			}
		})
	}
}

// The week is measured from where the child stood as it began: from the first
// day it reaches, and not from a day before it nor a later one. A topic the
// week began without was first answered in it, and is new.
func TestTheWeekIsMeasuredFromWhereItBegan(t *testing.T) {
	t.Parallel()

	p := answered(t, 1980, 1985, 1990, 2010)
	theta := p.Ratings.Theta
	gaps := p.Topics["counting.gaps"].Delta
	p.RatingDays = []profile.RatingDay{
		{Date: dated(time.September, 27), Deltas: map[string]float64{"counting.gaps": 0}, Theta: levelAt(1500)},
		{Date: dated(time.September, 29), Theta: levelAt(1900), Deltas: map[string]float64{
			"counting.gaps": gaps, "logic.ordering": levelAt(2100) - theta,
		}},
		{Date: dated(time.October, 2), Deltas: map[string]float64{"counting.gaps": 1}, Theta: levelAt(2500)},
	}
	summary := summaryOf(t, p, embedded(t))

	wantChange(t, "the overall rating", summary.Changes.Week, progress.Change{Before: new(drawn(1900)), Moved: progress.Forward})
	ordering := topicOf(t, &summary, "logic.ordering")
	wantChange(t, "logic.ordering", ordering.Changes.Week, progress.Change{Before: new(drawn(2100)), Moved: progress.Back})
	untouched := topicOf(t, &summary, "counting.gaps")
	wantChange(t, "counting.gaps", untouched.Changes.Week, progress.Change{Before: untouched.Standing, Moved: progress.Same})
	wantChange(t, lastTopic, topicOf(t, &summary, lastTopic).Changes.Week, progress.Change{Moved: progress.New})
}

// A week with no answer in it moved nothing, and says so: the overall rating
// and every topic stand where they stood.
func TestAWeekWithNoAnswerInItMovedNothing(t *testing.T) {
	t.Parallel()

	summary := summaryOf(t, fixture(t, "petya"), embedded(t))

	wantChange(t, "the overall rating", summary.Changes.Week, progress.Change{Before: summary.Overall, Moved: progress.Same})
	for _, topic := range summary.Topics {
		if topic.Standing != nil {
			wantChange(t, topic.ID, topic.Changes.Week, progress.Change{Before: topic.Standing, Moved: progress.Same})
		}
	}
}

// A week with an answer in it and no start the file keeps — the file of an
// earlier build — cannot be told, while the last answer still can.
func TestAWeekWhoseStartIsNotKeptCannotBeTold(t *testing.T) {
	t.Parallel()

	p := answered(t, 1980, 1985, 1990, 2010)
	p.Recent[len(p.Recent)-1].AnsweredAt = profile.Time{Time: time.Date(2026, time.October, 2, 18, 0, 0, 0, time.UTC)}
	summary := summaryOf(t, p, embedded(t))

	if summary.Changes.Week != nil {
		t.Errorf("the overall rating changed %s over the week, want nothing told", describe(summary.Changes.Week))
	}
	for _, topic := range summary.Topics {
		if topic.Changes.Week != nil {
			t.Errorf("%s changed %s over the week, want nothing told", topic.ID, describe(topic.Changes.Week))
		}
	}
	if summary.Changes.LastTask == nil {
		t.Error("the last answer's move is not told, want it told from the levels it keeps")
	}
}

// What kind of change a topic has in a generated week: its correction as the
// week began was the one it has now, another one, or none at all.
const (
	untouched = iota
	movedInWeek
	metInWeek
)

// Whatever the levels, a topic whose correction the last answer did not move
// stands where it stood since it, one whose correction the week began with
// stands where it stood over the week, and one the week began without is new;
// and one profile on one day is always told the same.
func TestATopicNoAnswerOfItsOwnMovedStandsWhereItStood(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	ids := slices.Sorted(maps.Keys(fixture(t, "petya").Topics))
	properties := gopter.NewProperties(nil)
	properties.Property("still when untouched, new when first met, told the same every time", prop.ForAll(
		func(theta float64, deltas, moves []float64, kinds []int, last int) bool {
			p := fixture(t, "petya")
			p.Ratings.Theta = theta
			day := profile.RatingDay{Date: today, Deltas: map[string]float64{}, Theta: theta - 0.1}
			for i, id := range ids {
				topic := p.Topics[id]
				topic.Answers, topic.Delta = 2, deltas[i]
				p.Topics[id] = topic
				switch kinds[i] {
				case untouched:
					day.Deltas[id] = deltas[i]
				case movedInWeek:
					day.Deltas[id] = deltas[i] - moves[i]
				}
			}
			p.RatingDays = []profile.RatingDay{day}
			lastAnswer := &p.Recent[len(p.Recent)-1]
			lastAnswer.Topic = ids[last]
			lastAnswer.Before = &profile.Levels{Delta: deltas[last] - moves[last], Theta: theta - 0.2}

			summary, err := progress.Of(p, catalog, today)
			again, _ := progress.Of(p, catalog, today)
			if err != nil || !reflect.DeepEqual(summary, again) {
				return false
			}
			return eachTopicStill(&summary, ids, kinds, ids[last])
		},
		gen.Float64Range(-4, 9),
		gen.SliceOfN(len(ids), gen.Float64Range(-3, 3)),
		gen.SliceOfN(len(ids), gen.Float64Range(-1, 1)),
		gen.SliceOfN(len(ids), gen.IntRange(untouched, metInWeek)),
		gen.IntRange(0, len(ids)-1),
	))
	properties.TestingRun(t)
}

// eachTopicStill says whether every topic that stands somewhere is told what
// its own answers moved: no move since the last answer unless it was in the
// topic, no move over the week unless its correction moved in it, and new when
// the week began without it.
func eachTopicStill(summary *progress.Summary, ids []string, kinds []int, last string) bool {
	for _, topic := range summary.Topics {
		if topic.Standing == nil {
			continue
		}
		still := progress.Change{Before: topic.Standing, Moved: progress.Same}
		if topic.ID != last && !sameChange(topic.Changes.LastTask, still) {
			return false
		}
		switch kinds[slices.Index(ids, topic.ID)] {
		case untouched:
			if !sameChange(topic.Changes.Week, still) {
				return false
			}
		case metInWeek:
			if !sameChange(topic.Changes.Week, progress.Change{Moved: progress.New}) {
				return false
			}
		}
	}
	return true
}

// Whatever the answer, it moves the overall level and its topic's correction
// the same way, so the two are never told as moving apart; and each move is
// read off the ranks and shares drawn: up exactly when the standing drawn rose,
// down exactly when it fell.
func TestAnAnswerNeverMovesItsTopicAndTheOverallRatingApart(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	properties := gopter.NewProperties(nil)
	properties.Property("never apart, and read off what is drawn", prop.ForAll(
		func(theta, delta, beta float64, answers, topicAnswers int, correct bool) bool {
			before := rating.State{Theta: theta, Delta: delta, Answers: answers, TopicAnswers: topicAnswers}
			moved := rating.Update(before, beta, correct)
			p := fixture(t, "petya")
			p.Ratings.Theta, p.Ratings.Answers = moved.Theta, moved.Answers
			topic := p.Topics[lastTopic]
			topic.Answers, topic.Correct, topic.Delta = moved.TopicAnswers, 0, moved.Delta
			p.Topics[lastTopic] = topic
			p.Recent[len(p.Recent)-1].Before = &profile.Levels{Delta: delta, Theta: theta}

			summary, err := progress.Of(p, catalog, today)
			if err != nil {
				return false
			}
			overall := summary.Changes.LastTask
			own := topicOf(t, &summary, lastTopic)
			return overall != nil && own.Changes.LastTask != nil &&
				!apart(overall.Moved, own.Changes.LastTask.Moved) &&
				readOff(overall, drawn(rating.Elo(theta)), *summary.Overall) &&
				readOff(own.Changes.LastTask, drawn(rating.Elo(moved.Theta+delta)), *own.Standing)
		},
		gen.Float64Range(-4, 9), gen.Float64Range(-3, 3), gen.Float64Range(-4, 12),
		gen.IntRange(rating.TrialAnswers, 300), gen.IntRange(1, 100), gen.Bool(),
	))
	properties.TestingRun(t)
}

// apart says whether one move went up and the other down.
func apart(a, b progress.Move) bool { return up(a) && down(b) || down(a) && up(b) }

func up(move progress.Move) bool   { return move == progress.RankUp || move == progress.Forward }
func down(move progress.Move) bool { return move == progress.Back || move == progress.RankDown }

// readOff says whether a change is told from before, as drawn, and goes up
// exactly when the standing drawn rose to now, down exactly when it fell, and
// nowhere when neither the rank nor the share moved.
func readOff(change *progress.Change, before, now progress.Standing) bool {
	rose := now.Rank > before.Rank || now.Rank == before.Rank && now.Share > before.Share
	fell := now.Rank < before.Rank || now.Rank == before.Rank && now.Share < before.Share
	return change.Before != nil && *change.Before == before &&
		up(change.Moved) == rose && down(change.Moved) == fell &&
		(change.Moved == progress.Same) == (!rose && !fell)
}
