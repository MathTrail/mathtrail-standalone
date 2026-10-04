package profile

import "github.com/MathTrail/mathtrail-standalone/internal/domain/rating"

// RatingDaysKept is how many days the history of the ratings reaches back:
// the week the progress tells the change over, seven dates in UTC with today
// among them. A day the week no longer reaches is let go of, so the file never
// holds more.
const RatingDaysKept = 7

// Levels is where the child stood at one moment: the overall level and the
// correction of one topic.
type Levels struct {
	Delta float64 `json:"delta"`
	Theta float64 `json:"theta"`
}

// RatingDay is where the child stood at the start of one day with answers:
// the overall level, and the correction of every topic answered by then. A
// topic answered for the first time that day is not among them, which is how
// a topic new to the week is told.
type RatingDay struct {
	// Date is the day, in UTC.
	Date Date `json:"date"`
	// Deltas are the corrections of the topics answered before the day began,
	// by catalog id.
	Deltas map[string]float64 `json:"deltas"`
	// Theta is the overall level.
	Theta float64 `json:"theta"`
	// Unkept is the latest date of an answer that keeps no levels before it,
	// on the first day kept after a build that kept no history wrote the file,
	// and absent on every other day. That build let go of the days kept before
	// it and of the levels every answer kept, so a week that reaches back to
	// this date has a start no day holds.
	Unkept *Date `json:"unkept,omitempty"`
}

// startTheDay keeps where the child stands as the start of day when the
// answer about to move the levels is the first of that day, and lets go of
// the days the week from day no longer reaches. The answers of the trial
// series keep no day: a week never starts in the middle of the series, and the
// first answer after it keeps the levels the series ended at. A day kept for a
// later date than this one — the clock of another instance ran ahead — keeps
// this answer from starting a day of its own, so the days stay in order. The
// first day kept with none before it says when the latest answer that kept no
// levels was given — a build that kept no history wrote the file — unless it
// is the day the series ended at: the answers before it are the series',
// which a week never measures.
func (p *Profile) startTheDay(day Date) {
	if !p.Ratings.InTrial() && !p.holdsDayFrom(day) {
		started := RatingDay{Date: day, Deltas: p.answeredDeltas(), Theta: p.Ratings.Theta}
		if len(p.RatingDays) == 0 && p.Ratings.Answers > rating.TrialAnswers {
			started.Unkept = p.lastUnkeptAnswer()
		}
		p.RatingDays = append(p.RatingDays, started)
	}
	p.RatingDays = keptFrom(p.RatingDays, day)
}

// lastUnkeptAnswer is the latest date of an answer in the window that keeps no
// levels before it, or nil when every answer keeps them. A task left without
// an answer moved nothing and keeps none, and is passed over.
func (p *Profile) lastUnkeptAnswer() *Date {
	var latest *Date
	for i := range p.Recent {
		entry := &p.Recent[i]
		if entry.Skipped || entry.Before != nil {
			continue
		}
		if day := DateOf(entry.AnsweredAt.Time); latest == nil || day.After(latest.Time) {
			latest = &day
		}
	}
	return latest
}

// holdsDayFrom reports whether the last day kept is day or a later one.
func (p *Profile) holdsDayFrom(day Date) bool {
	last := len(p.RatingDays) - 1
	return last >= 0 && !p.RatingDays[last].Date.Before(day.Time)
}

// answeredDeltas are the corrections of the topics answered so far.
func (p *Profile) answeredDeltas() map[string]float64 {
	deltas := map[string]float64{}
	for id, topic := range p.Topics {
		if topic.Answers > 0 {
			deltas[id] = topic.Delta
		}
	}
	return deltas
}

// keptFrom are the days the week ending on day still reaches, in a slice of
// their own: a copy of the profile taken before shares nothing that is cut.
// None left is none at all, which the file leaves out.
func keptFrom(days []RatingDay, day Date) []RatingDay {
	first := weekFrom(day)
	var kept []RatingDay
	for _, held := range days {
		if !held.Date.Before(first.Time) {
			kept = append(kept, held)
		}
	}
	return kept
}

// weekFrom is the first date of the week that ends on day.
func weekFrom(day Date) Date { return Date{day.AddDate(0, 0, -(RatingDaysKept - 1))} }

// WeekStart is where the child stood as the week ending today began, and
// whether the week can be told at all. It began from the first day kept that
// falls within the week; a day kept for a date after today is no part of it.
// That day cannot start the week when an answer that kept no levels came
// within the week before it: the first day kept is then somewhere in the
// middle of the week, and a change told from it would pass part of a week off
// as the whole. With no day in the week, nothing moved in it when no answer
// came in it — the start is nil and the week is known — and the week cannot be
// told when an answer did: its days were dropped by an earlier build, or the
// history began after it.
func (p *Profile) WeekStart(today Date) (start *RatingDay, known bool) {
	first := weekFrom(today)
	for _, day := range p.RatingDays {
		if day.Date.Before(first.Time) || day.Date.After(today.Time) {
			continue
		}
		if day.Unkept != nil && !day.Unkept.Before(first.Time) {
			return nil, false
		}
		return &day, true
	}
	last, answered := p.LastAnswer()
	return nil, !answered || last.AnsweredAt.Before(first.Time)
}
