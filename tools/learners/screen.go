package main

import (
	"strconv"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// answerWindow is a run of answers, from the first to the last, counted from
// one.
type answerWindow struct {
	first, last int
}

// name is how a measure read over the window names it: 6_20 for answers 6 to
// 20.
func (w answerWindow) name() string { return strconv.Itoa(w.first) + "_" + strconv.Itoa(w.last) }

// label is how a reader is told the window: 6–20.
func (w answerWindow) label() string { return strconv.Itoa(w.first) + "–" + strconv.Itoa(w.last) }

// screenWindows are the answers what a child is shown is read over: the first
// ones after the trial series, during which no rating is shown, and the last
// of a run of two hundred.
var screenWindows = [2]answerWindow{{first: 6, last: 20}, {first: 150, last: 200}}

// windowOf is the place in screenWindows of the window answer k, counted from
// zero, falls in, or -1 for an answer in none.
func windowOf(k int) int {
	for at, w := range screenWindows {
		if k+1 >= w.first && k+1 <= w.last {
			return at
		}
	}
	return -1
}

// moveCap is the largest move of a rating counted as itself, a larger one
// counting as it: no rule moves a child's rating that far on one answer, and a
// fixed number of possible moves lets a sample's moves be counted without
// sorting them.
const moveCap = 1023

// screenRecord is what a child was shown in each window: how far the rating
// of the topic answered moved on the card after each answer, how often the
// overall rank and the answered topic's rank changed, and over how many
// answers.
type screenRecord struct {
	moves                     [len(screenWindows)][]uint16
	rankChanges, topicChanges [len(screenWindows)]int
	answers                   [len(screenWindows)]int
}

// after records what the child is shown once the rule has taken answer k in,
// read off the rule's own estimate as the card and the progress screen would
// show it: the topic's rating before and after the answer, and the ranks of
// the overall rating and the topic's. A rule's estimate rather than the
// profile's, because a rule that is not the service's has its level written
// into the profile only before the next task.
func (sr *screenRecord) after(s *session, topic string, o *observedAnswer, k int) {
	at := windowOf(k)
	if at < 0 {
		return
	}
	before, after := rating.Elo(o.estimate), rating.Elo(s.est.level(topic))
	move := rating.Shown(after) - rating.Shown(before)
	if move < 0 {
		move = -move
	}
	sr.moves[at] = append(sr.moves[at], uint16(min(move, moveCap))) //nolint:gosec // at most moveCap, which a uint16 holds
	if rating.Rank(rating.Elo(o.overall)) != rating.Rank(rating.Elo(s.est.overall())) {
		sr.rankChanges[at]++
	}
	if rating.Rank(before) != rating.Rank(after) {
		sr.topicChanges[at]++
	}
	sr.answers[at]++
}

// rankChanges is how often a rank changed in a window, in changes a hundred
// answers: the overall rank's, or, for the topic, the rank of the topic
// answered.
func rankChanges(at int, topic bool) func(r *childResult) (float64, bool) {
	return func(r *childResult) (float64, bool) {
		changes := r.screen.rankChanges[at]
		if topic {
			changes = r.screen.topicChanges[at]
		}
		answers := r.screen.answers[at]
		return 100 * float64(changes) / float64(answers), answers > 0
	}
}

// movesAt reads the given percentile of the moves on the card in a window,
// over every answer of the sample's children together, by the nearest rank:
// the smallest move that at least that share of the moves are no larger than.
func movesAt(at, percentile int) reader {
	return func(children []vector, sample []int) (float64, bool) {
		var counts [moveCap + 1]int
		total := 0
		for _, c := range sample {
			for _, move := range children[c].moves[at] {
				counts[move]++
			}
			total += len(children[c].moves[at])
		}
		if total == 0 {
			return 0, false
		}
		rank, seen := (percentile*total+99)/100, 0
		for move, n := range &counts {
			if seen += n; seen >= rank {
				return float64(move), true
			}
		}
		return moveCap, true
	}
}

// shown marks a measure of what the child is shown, which says nothing of the
// ceiling: its rating moves only when the child does.
func shown(m metric) metric {
	also := m.applies
	m.applies = func(r *rule) bool { return notCeiling(r) && (also == nil || also(r)) }
	return m
}

// notCeiling says whether a rule is one a measure of what the child is shown
// applies to.
func notCeiling(r *rule) bool { return !r.ceiling }
