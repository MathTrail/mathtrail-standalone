package progress

import (
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Move is the word for how a standing moved, read off what is drawn of it:
// its rank, and how far through the rank it has come.
type Move string

const (
	// RankUp is a rank above the one it stood in.
	RankUp Move = "rank_up"
	// Forward is further through the same rank.
	Forward Move = "forward"
	// Same is no move that is drawn.
	Same Move = "same"
	// Back is less far through the same rank.
	Back Move = "back"
	// RankDown is a rank below the one it stood in.
	RankDown Move = "rank_down"
	// New is a topic first answered in the while, which stood nowhere before.
	New Move = "new"
)

// Change is how a standing moved over a while: where it stood before, as it
// was drawn, and the word for the move. A topic first answered in the while
// stood nowhere, and has no standing before.
type Change struct {
	Before *Standing
	Moved  Move
}

// Changes are how a standing moved since the last answer and over the week of
// seven dates in UTC ending today. A while that cannot be told is nil: during
// the trial series and at its last answer, which gave the first rank; for a
// last answer whose levels before it the file does not keep; and for a week
// the history does not cover. The overall standing has changes only when it
// is shown, and each one it has says where it stood before: only a topic can
// be new.
type Changes struct {
	LastTask *Change
	Week     *Change
}

// whiles are what a standing is measured from: the last answer, with the
// levels before it, and where the child stood as the week began — nil with
// the week known when nothing moved in it.
type whiles struct {
	last      *profile.Answer
	weekStart *profile.RatingDay
	weekKnown bool
}

// whilesOf are the whiles of the profile as of today. While the trial series
// runs, and at its last answer, there is nothing to measure from.
func whilesOf(p *profile.Profile, today profile.Date) whiles {
	if p.Ratings.Answers <= rating.TrialAnswers {
		return whiles{}
	}
	var measured whiles
	if last, answered := p.LastAnswer(); answered && last.Before != nil {
		measured.last = &last
	}
	measured.weekStart, measured.weekKnown = p.WeekStart(today)
	return measured
}

// overall is how the overall standing moved: from the overall level before
// the last answer, and from the one the week began at.
func (w whiles) overall(now *Standing) Changes {
	if now == nil {
		return Changes{}
	}
	var changes Changes
	if w.last != nil {
		changes.LastTask = changeFrom(standing(w.last.Before.Theta), now)
	}
	if w.weekKnown {
		before := now
		if w.weekStart != nil {
			before = standing(w.weekStart.Theta)
		}
		changes.Week = changeFrom(before, now)
	}
	return changes
}

// topic is how a topic's own standing moved: its correction alone, measured
// from where the overall level stands now. Every answer moves the overall
// level and with it every topic's rating, so a topic measured whole would
// seem to move with each task in another; measured by its own correction, a
// topic no answer of its own moved stands where it stood.
func (w whiles) topic(theta float64, topic *Topic) Changes {
	if topic.Standing == nil {
		return Changes{}
	}
	var changes Changes
	if last := w.last; last != nil {
		switch {
		case last.Topic != topic.ID:
			changes.LastTask = changeFrom(topic.Standing, topic.Standing)
		case topic.Answers == 1:
			changes.LastTask = &Change{Moved: New}
		default:
			before := rating.State{Theta: theta, Delta: last.Before.Delta}
			changes.LastTask = changeFrom(standing(before.Level()), topic.Standing)
		}
	}
	if w.weekKnown {
		changes.Week = w.weekChange(theta, topic)
	}
	return changes
}

// weekChange is how a topic's own standing moved over the week: from the
// correction it had as the week began, or new when it was first answered in
// the week, or no move at all when nothing moved in the week.
func (w whiles) weekChange(theta float64, topic *Topic) *Change {
	if w.weekStart == nil {
		return changeFrom(topic.Standing, topic.Standing)
	}
	delta, met := w.weekStart.Deltas[topic.ID]
	if !met {
		return &Change{Moved: New}
	}
	before := rating.State{Theta: theta, Delta: delta}
	return changeFrom(standing(before.Level()), topic.Standing)
}

// changeFrom is the move from before to now.
func changeFrom(before, now *Standing) *Change {
	return &Change{Before: before, Moved: moveOf(before, now)}
}

// moveOf is the word for the move from before to now, read off what is drawn:
// a rank above or below, or, in the same rank, further through it or less
// far. A move too small to change either is no move, as it is on the screen,
// and so is any move inside the highest rank, which is drawn full, or below
// the first rank's way, which is drawn empty.
func moveOf(before, now *Standing) Move {
	switch {
	case now.Rank > before.Rank:
		return RankUp
	case now.Rank < before.Rank:
		return RankDown
	case now.Share > before.Share:
		return Forward
	case now.Share < before.Share:
		return Back
	default:
		return Same
	}
}
