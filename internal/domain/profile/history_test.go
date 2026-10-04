package profile_test

import (
	"fmt"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The history of the ratings is what lets a change be told afterwards: every
// answer keeps where the child stood before it, and the first answer of each
// day after the trial series keeps where the child stood as the day began.
// Only recording an answer writes either.

// atDay is the moment hours into the date days after the date of the child's
// first task, in UTC.
func atDay(days, hours int) time.Time {
	return issued.Truncate(24 * time.Hour).Add(time.Duration(days*24+hours) * time.Hour)
}

// dayOf is the date of atDay(days, 0), in UTC.
func dayOf(days int) profile.Date { return profile.DateOf(atDay(days, 0)) }

// answeredOn puts the n-th task, of topic and at the child's grade, on the card
// and answers it, right or wrong, at the moment given.
func answeredOn(t *testing.T, p *profile.Profile, n int, topic string, at time.Time, correct bool) {
	t.Helper()

	level, _ := rating.GradeLevelOf(p.Student.Grade)
	id := answeringAs(t, p, fmt.Sprintf("tsk_%03d", n), topic, rating.Point{GradeLevel: level, Difficulty: 3})
	choice := wrongLetter
	if correct {
		choice = rightLetter
	}
	give(t, p, id, choice, at)
}

// pastTheSeries is a child of grade 3 who answered the five tasks of the trial
// series on the day of their first task, two topics among them.
func pastTheSeries(t *testing.T) *profile.Profile {
	t.Helper()

	p := newChild(t, 3)
	for n := range rating.TrialAnswers {
		answeredOn(t, p, n, []string{"counting.gaps", "logic.ordering"}[n%2], atDay(0, n), n%3 != 0)
	}
	if p.Ratings.InTrial() {
		t.Fatalf("the child is still in the trial series after %d answers", p.Ratings.Answers)
	}
	return p
}

// deltasOf are the corrections of the topics the child has answered.
func deltasOf(p *profile.Profile) map[string]float64 {
	deltas := map[string]float64{}
	for id, topic := range p.Topics {
		if topic.Answers > 0 {
			deltas[id] = topic.Delta
		}
	}
	return deltas
}

func TestAnAnswerKeepsWhereTheChildStoodBeforeIt(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	theta, delta := p.Ratings.Theta, p.Topics["counting.gaps"].Delta

	answeredOn(t, p, 10, "counting.gaps", atDay(1, 0), true)

	last := p.Recent[len(p.Recent)-1]
	if last.Before == nil || *last.Before != (profile.Levels{Delta: delta, Theta: theta}) {
		t.Errorf("before = %+v, want the levels before the answer, delta %v and theta %v", last.Before, delta, theta)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile that can be written", err)
	}
}

// The history as recording answers writes it is a file the schema holds and
// the service reads back whole: every level it keeps comes back as it went.
func TestTheHistoryAsWrittenHoldsToTheSchemaAndReadsBack(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(1, 2), true)
	answeredOn(t, p, 11, "time.clocks", atDay(1, 3), false)
	answeredOn(t, p, 12, "logic.ordering", atDay(3, 1), true)

	file, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	if broken := againstTheSchema(t, file); broken != nil {
		t.Errorf("the file breaks the schema: %v", broken)
	}
	read, err := profile.Parse(file)
	if err != nil {
		t.Fatalf("Parse() error = %v, want the file read back", err)
	}
	if len(read.RatingDays) != 2 || !slices.EqualFunc(read.RatingDays, p.RatingDays, sameDay) {
		t.Errorf("rating_days read back as %+v, want the two days written, %+v", read.RatingDays, p.RatingDays)
	}
	for i, entry := range read.Recent {
		if written := p.Recent[i].Before; (entry.Before == nil) != (written == nil) || written != nil && *entry.Before != *written {
			t.Errorf("recent[%d].before read back as %+v, want %+v", i, entry.Before, written)
		}
	}
}

func TestATaskLeftWithoutAnAnswerKeepsNoLevelsBeforeIt(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	level, _ := rating.GradeLevelOf(p.Student.Grade)
	answeringAs(t, p, "tsk_left", "counting.gaps", rating.Point{GradeLevel: level, Difficulty: 3})
	days := slices.Clone(p.RatingDays)

	if _, skipped := p.Skip(atDay(1, 0)); !skipped {
		t.Fatal("Skip() left no entry, want the task kept as skipped")
	}

	if last := p.Recent[len(p.Recent)-1]; !last.Skipped || last.Before != nil {
		t.Errorf("the skipped entry is %+v, want one with no levels before it", last)
	}
	if !slices.EqualFunc(p.RatingDays, days, sameDay) {
		t.Errorf("rating_days = %+v after a skip, want them as they were, %+v", p.RatingDays, days)
	}
}

// A task left without an answer keeps no levels, and is no answer the history
// missed: a first day kept with nothing before it but answers that keep their
// levels, and a skip, marks none missed.
func TestATaskLeftWithoutAnAnswerIsNoAnswerTheHistoryMissed(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(1, 0), true)
	level, _ := rating.GradeLevelOf(p.Student.Grade)
	answeringAs(t, p, "tsk_left", "counting.gaps", rating.Point{GradeLevel: level, Difficulty: 3})
	if _, skipped := p.Skip(atDay(2, 0)); !skipped {
		t.Fatal("Skip() left no entry, want the task kept as skipped")
	}
	p.RatingDays = nil // no day kept, as a file edited by hand may be

	answeredOn(t, p, 11, "counting.gaps", atDay(2, 1), true)

	if gap := p.RatingDays[0].Unkept; gap != nil {
		t.Errorf("the first day kept after a skip says an answer was missed on %v, want none", *gap)
	}
}

func TestTheTrialSeriesKeepsNoDay(t *testing.T) {
	t.Parallel()

	p := newChild(t, 3)
	for n := range rating.TrialAnswers {
		answeredOn(t, p, n, "counting.gaps", atDay(n, 0), true)
		if len(p.RatingDays) != 0 {
			t.Fatalf("rating_days = %+v after answer %d of the series, want none", p.RatingDays, n+1)
		}
	}
}

func TestTheFirstAnswerAfterTheSeriesStartsTheDayFromWhereTheSeriesEnded(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	theta, deltas := p.Ratings.Theta, deltasOf(p)

	answeredOn(t, p, 10, "counting.gaps", atDay(0, 8), true)

	want := []profile.RatingDay{{Date: dayOf(0), Deltas: deltas, Theta: theta}}
	if !slices.EqualFunc(p.RatingDays, want, sameDay) {
		t.Errorf("rating_days = %+v, want the day started from where the series ended, %+v", p.RatingDays, want)
	}
}

func TestADayIsKeptOnceAtItsFirstAnswer(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(1, 0), true)
	firstDay := p.RatingDays[len(p.RatingDays)-1]
	answeredOn(t, p, 11, "counting.gaps", atDay(1, 5), false)

	if len(p.RatingDays) != 1 || !sameDay(p.RatingDays[0], firstDay) {
		t.Fatalf("rating_days = %+v after a second answer the same day, want only %+v", p.RatingDays, firstDay)
	}

	theta, deltas := p.Ratings.Theta, deltasOf(p)
	answeredOn(t, p, 12, "counting.gaps", atDay(2, 0), true)

	if got := p.RatingDays[len(p.RatingDays)-1]; !sameDay(got, profile.RatingDay{Date: dayOf(2), Deltas: deltas, Theta: theta}) {
		t.Errorf("the next day started at %+v, want the levels before its first answer, theta %v", got, theta)
	}
}

// A topic only skipped so far has a summary of its own and no answer, and its
// first answer still makes it new to the day.
func TestATopicFirstAnsweredOnADayIsNoneOfItsCorrections(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	level, _ := rating.GradeLevelOf(p.Student.Grade)
	answeringAs(t, p, "tsk_left", "time.clocks", rating.Point{GradeLevel: level, Difficulty: 3})
	if _, skipped := p.Skip(atDay(1, 0)); !skipped {
		t.Fatal("Skip() left no entry, want the task kept as skipped")
	}

	answeredOn(t, p, 10, "time.clocks", atDay(1, 1), true)

	if _, kept := p.RatingDays[0].Deltas["time.clocks"]; kept {
		t.Errorf("the day keeps a correction of a topic first answered that day: %+v", p.RatingDays[0].Deltas)
	}
}

func TestTheDaysKeptAreTheOnesTheWeekReaches(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	for day := 1; day <= 9; day++ {
		answeredOn(t, p, 10+day, "counting.gaps", atDay(day, 0), day%2 == 0)
	}

	dates := make([]profile.Date, 0, len(p.RatingDays))
	for _, day := range p.RatingDays {
		dates = append(dates, day.Date)
	}
	want := []profile.Date{dayOf(3), dayOf(4), dayOf(5), dayOf(6), dayOf(7), dayOf(8), dayOf(9)}
	if !slices.Equal(dates, want) {
		t.Errorf("the days kept are %v, want the %d dates up to the last answer's, %v", dates, profile.RatingDaysKept, want)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile that can be written", err)
	}
}

func TestAnAnswerDatedBeforeTheLastDayStartsNoDay(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(3, 0), true)
	days := slices.Clone(p.RatingDays)

	answeredOn(t, p, 11, "counting.gaps", atDay(2, 0), true)

	if !slices.EqualFunc(p.RatingDays, days, sameDay) {
		t.Errorf("rating_days = %+v after an answer from a day already past, want %+v", p.RatingDays, days)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want the days still in order", err)
	}
}

func TestTheWeekStartsFromTheFirstDayItReaches(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(1, 0), true)
	answeredOn(t, p, 11, "counting.gaps", atDay(4, 0), true)
	answeredOn(t, p, 12, "counting.gaps", atDay(9, 0), true)

	// The answer of day 9 let the day of day 1 go: the days kept are 4 and 9.
	for _, tc := range []struct {
		name  string
		today int
		from  int // the day the week starts from, or -1 for none
		known bool
	}{
		{"a week reaching both days kept", 9, 4, true},
		{"a week that begins on its first day", 10, 4, true},
		{"a week that no longer reaches its first day", 11, 9, true},
		{"a week whose only day is its first", 15, 9, true},
		{"a week with no answer in it", 16, -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			start, known := p.WeekStart(dayOf(tc.today))
			switch {
			case known != tc.known:
				t.Errorf("WeekStart() known = %v, want %v", known, tc.known)
			case tc.from < 0 && start != nil:
				t.Errorf("WeekStart() = %+v, want no start", *start)
			case tc.from >= 0 && (start == nil || start.Date != dayOf(tc.from)):
				t.Errorf("WeekStart() = %+v, want the day of %v", start, dayOf(tc.from))
			}
		})
	}
}

func TestAWeekWhoseDaysWereDroppedCannotBeTold(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(1, 0), true)
	p.RatingDays = nil // as an earlier build writes the file

	if start, known := p.WeekStart(dayOf(2)); known || start != nil {
		t.Errorf("WeekStart() = %+v, %v, want a week that cannot be told", start, known)
	}
}

// writtenByAnEarlierBuild is the profile as a build that kept no history
// writes it back: with no day kept, and no answer keeping the levels before
// it.
func writtenByAnEarlierBuild(p *profile.Profile) {
	p.RatingDays = nil
	for i := range p.Recent {
		p.Recent[i].Before = nil
	}
}

// An answer an earlier build wrote keeps no levels before it, and that build
// kept no day either: the first day kept after it is the middle of its week,
// which cannot be told until the answer is out of it. That day alone says so.
func TestAWeekWithAnAnswerTheHistoryMissedCannotBeTold(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(1, 0), true)
	writtenByAnEarlierBuild(p)
	answeredOn(t, p, 11, "counting.gaps", atDay(3, 0), true)
	answeredOn(t, p, 12, "counting.gaps", atDay(4, 0), true)

	if start, known := p.WeekStart(dayOf(4)); known || start != nil {
		t.Errorf("WeekStart() = %+v, %v with an answer of the week the history missed, want a week that cannot be told", start, known)
	}
	if start, known := p.WeekStart(dayOf(8)); !known || start == nil || start.Date != dayOf(3) {
		t.Errorf("WeekStart() = %+v, %v once that answer is out of the week, want the day of %v", start, known, dayOf(3))
	}
	if gap := p.RatingDays[0].Unkept; gap == nil || *gap != dayOf(1) || p.RatingDays[1].Unkept != nil {
		t.Errorf("rating_days = %+v, want the date of the missed answer on the first day after it alone", p.RatingDays)
	}
}

// A child in the middle of the trial series when an earlier build last wrote
// the file has answers of the series that keep no levels, and a week never
// measures the series: the day it ended at marks no answer missed.
func TestTheSeriesAnswersAnEarlierBuildWroteMissNothingAWeekMeasures(t *testing.T) {
	t.Parallel()

	p := newChild(t, 3)
	for n := range 3 {
		answeredOn(t, p, n, "counting.gaps", atDay(0, n), true)
	}
	writtenByAnEarlierBuild(p)
	for n := 3; n <= rating.TrialAnswers; n++ {
		answeredOn(t, p, n, "logic.ordering", atDay(1, n), n%2 == 0)
	}

	if gap := p.RatingDays[0].Unkept; gap != nil {
		t.Errorf("the day the series ended at says an answer was missed on %v, want none", *gap)
	}
	if start, known := p.WeekStart(dayOf(1)); !known || start == nil {
		t.Errorf("WeekStart() = %+v, %v, want the week told from where the series ended", start, known)
	}
}

// What the history missed stays missed once the answers that show it have left
// the window: the first day kept after it remembers when it was.
func TestAWeekStaysUntoldOnceTheAnswersTheHistoryMissedLeaveTheWindow(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(1, 0), true)
	writtenByAnEarlierBuild(p)
	for n := range profile.MaxRecent + 1 {
		answeredOn(t, p, 20+n, "counting.gaps", atDay(3, 1).Add(time.Duration(n)*time.Minute), n%2 == 0)
	}
	if slices.ContainsFunc(p.Recent, func(entry profile.Answer) bool { return !entry.Skipped && entry.Before == nil }) {
		t.Fatal("the window still holds an answer the history missed, want it gone from the window")
	}

	if start, known := p.WeekStart(dayOf(4)); known || start != nil {
		t.Errorf("WeekStart() = %+v, %v with an answer of the week the history missed, want a week that cannot be told", start, known)
	}
}

func TestADayLaterThanTodayIsNoPartOfTheWeek(t *testing.T) {
	t.Parallel()

	p := pastTheSeries(t)
	answeredOn(t, p, 10, "counting.gaps", atDay(5, 0), true)

	if start, known := p.WeekStart(dayOf(1)); known || start != nil {
		t.Errorf("WeekStart() = %+v, %v a day before the only day kept, want a week that cannot be told", start, known)
	}
}

// sameDay reports whether two days say the same: the same date, the same
// overall level and the same corrections, and the same answer missed before
// them, or none.
func sameDay(a, b profile.RatingDay) bool {
	sameGap := (a.Unkept == nil) == (b.Unkept == nil) && (a.Unkept == nil || *a.Unkept == *b.Unkept)
	return a.Date == b.Date && a.Theta == b.Theta && maps.Equal(a.Deltas, b.Deltas) && sameGap
}

// stood is where a child stood after one answer: its date and the levels.
type stood struct {
	date   profile.Date
	theta  float64
	deltas map[string]float64
}

// expectedStart is the start of the week ending on today that the answers
// logged imply: where the child stood before the first answer after the trial
// series that falls in the week, or none with the week known when no answer
// came in it, and none with the week unknown when only answers of the series
// did. The log is in the order of the answers, and their dates never go back.
func expectedStart(log []stood, today profile.Date) (start *stood, known bool) {
	first := profile.Date{Time: today.AddDate(0, 0, -(profile.RatingDaysKept - 1))}
	answeredInWeek := false
	for i, after := range log {
		if after.date.Before(first.Time) || after.date.After(today.Time) {
			continue
		}
		answeredInWeek = true
		if i >= rating.TrialAnswers {
			return &log[i-1], true
		}
	}
	return nil, !answeredInWeek
}

func TestTheWeekStartsWhereTheChildStoodAsItBegan(t *testing.T) {
	t.Parallel()

	gaps := gen.OneConstOf(0, 0, 0, 1, 2, 7, 8)
	topics := gen.OneConstOf("counting.gaps", "logic.ordering", "time.clocks")
	properties := gopter.NewProperties(nil)
	properties.Property("the week starts from the levels it began at", prop.ForAll(
		func(steps []int, names []string, rights []bool, later int) bool {
			p := newChild(t, 3)
			var log []stood
			day := 0
			for n := range min(len(steps), len(names), len(rights)) {
				day += steps[n]
				answeredOn(t, p, n, names[n], atDay(day, n%20), rights[n])
				log = append(log, stood{date: dayOf(day), theta: p.Ratings.Theta, deltas: deltasOf(p)})
			}
			today := dayOf(day + later)

			start, known := p.WeekStart(today)
			want, wantKnown := expectedStart(log, today)
			switch {
			case known != wantKnown:
				return false
			case want == nil:
				return start == nil
			default:
				return start != nil && start.Theta == want.theta && maps.Equal(start.Deltas, want.deltas)
			}
		},
		gen.SliceOfN(40, gaps), gen.SliceOfN(40, topics), gen.SliceOfN(40, gen.Bool()), gen.IntRange(0, 9),
	))
	properties.TestingRun(t)
}
