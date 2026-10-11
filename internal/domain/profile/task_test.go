package profile_test

import (
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// The daily counters are in the file because every instance of the service has
// to see the same numbers, and they belong to a day: at the turn of it they
// start again, and nobody has to remember to ask.
func TestTheDailyCountersBelongToADay(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	day := time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC)

	p.CountAccepted(day)
	p.CountAccepted(day.Add(6 * time.Hour))
	p.CountFailed(day.Add(7 * time.Hour))

	if p.Daily.Accepted != 2 || p.Daily.Failed != 1 {
		t.Errorf("the day counted %d accepted and %d failed, want 2 and 1", p.Daily.Accepted, p.Daily.Failed)
	}
	if got := p.Daily.Date.Format("2006-01-02"); got != "2026-09-20" {
		t.Errorf("the counters belong to %s, want 2026-09-20", got)
	}

	// The next day starts from nothing, and it starts by itself.
	p.CountAccepted(day.AddDate(0, 0, 1))
	if p.Daily.Accepted != 1 || p.Daily.Failed != 0 {
		t.Errorf("the new day counted %d accepted and %d failed, want 1 and none",
			p.Daily.Accepted, p.Daily.Failed)
	}
	if got := p.Daily.Date.Format("2006-01-02"); got != "2026-09-21" {
		t.Errorf("the counters belong to %s, want the new day", got)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile that can be written", err)
	}
}

// A limit reads the counters of the day it is asked on: the day's own as they
// stand, and nothing counted for a day the counters do not belong to — the
// counters themselves left as the file holds them.
func TestTheCountersOfTodayAreTodaysAlone(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	day := time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC)
	p.CountAccepted(day)
	p.CountFailed(day)
	kept := p.Daily

	if got := p.Daily.Today(day.Add(14 * time.Hour)); got != kept {
		t.Errorf("Today() later that day = %+v, want the day's own %+v", got, kept)
	}
	next := day.AddDate(0, 0, 1)
	want := profile.Daily{Date: profile.DateOf(next), Ends: profile.At(time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))}
	if got := p.Daily.Today(next); got != want {
		t.Errorf("Today() the next day = %+v, want nothing counted yet, %+v", got, want)
	}
	if p.Daily != kept {
		t.Errorf("the counters became %+v by being read, want them as they were, %+v", p.Daily, kept)
	}
}

// The family's day ends at the family's midnight, by the clock the card told:
// for a family five hours behind UTC, a task given at half past six in the
// evening and one given at half past seven are counted on the same day,
// although UTC has turned over between them, and the day starts again only at
// the family's own midnight.
func TestTheDayEndsAtTheFamilysMidnight(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	evening := time.Date(2026, 10, 10, 23, 30, 0, 0, time.UTC) // 18:30 for the family
	p.SetClock(-5 * 60)

	p.CountAccepted(evening)
	p.CountAccepted(evening.Add(time.Hour)) // 19:30 for the family, the next day in UTC
	if p.Daily.Accepted != 2 {
		t.Errorf("the evening counted %d tasks, want 2: UTC's midnight is not the family's", p.Daily.Accepted)
	}
	if got := p.Daily.Date.Format("2006-01-02"); got != "2026-10-10" {
		t.Errorf("the counters belong to %s, want the family's 2026-10-10", got)
	}

	beforeMidnight := time.Date(2026, 10, 11, 4, 59, 59, 0, time.UTC)
	if got := p.Daily.Today(beforeMidnight); got.Accepted != 2 {
		t.Errorf("Today() a second before the family's midnight = %+v, want the day's two tasks", got)
	}
	midnight := time.Date(2026, 10, 11, 5, 0, 0, 0, time.UTC)
	want := profile.Daily{
		Date: profile.DateOf(midnight), Ends: profile.At(time.Date(2026, 10, 12, 5, 0, 0, 0, time.UTC)), UTCOffset: -5 * 60,
	}
	if got := p.Daily.Today(midnight); got != want {
		t.Errorf("Today() at the family's midnight = %+v, want nothing counted and the clock kept, %+v", got, want)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile that can be written", err)
	}
}

// A clock is taken only at an offset some time zone keeps: twelve hours behind
// UTC to fourteen ahead, in quarters of an hour, such as Nepal's five hours
// and three quarters. Any other offset leaves the clock as it was.
func TestOnlyAClockSomeTimeZoneKeepsIsTaken(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		offset int
		want   int
	}{
		{"five hours behind", -300, -300},
		{"Nepal's", 5*60 + 45, 5*60 + 45},
		{"the furthest ahead", 14 * 60, 14 * 60},
		{"the furthest behind", -12 * 60, -12 * 60},
		{"UTC itself", 0, 0},
		{"past the furthest ahead", 14*60 + 15, 60},
		{"past the furthest behind", -12*60 - 15, 60},
		{"off the quarters", 61, 60},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "dima")
			p.SetClock(60)
			p.SetClock(tc.offset)
			if p.Daily.UTCOffset != tc.want {
				t.Errorf("SetClock(%d) after an hour ahead keeps %d, want %d", tc.offset, p.Daily.UTCOffset, tc.want)
			}
		})
	}
}

// A clock the card tells counts from the next day on: the day under way ends
// when it was to end. A family whose first answer on a card says it is five
// hours behind UTC keeps the tasks of its afternoon until UTC's midnight, the
// one its day began under, and two devices whose clocks take turns neither end
// that day sooner nor stretch it. The day after begins by the clock told last
// and ends at the family's midnight.
func TestAClockCountsFromTheNextDayOn(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	afternoon := time.Date(2026, 10, 10, 22, 0, 0, 0, time.UTC) // five in the afternoon for the family
	for range 3 {
		p.CountAccepted(afternoon)
	}

	for _, offset := range []int{-5 * 60, 0, -5 * 60, 0, -5 * 60} {
		p.SetClock(offset)
		if got := p.Daily.Today(afternoon.Add(time.Hour)); got.Accepted != 3 {
			t.Fatalf("Today() once the clock is %d = %+v, want the afternoon's three tasks", offset, got)
		}
	}
	utcMidnight := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	if got := p.Daily.Today(utcMidnight.Add(-time.Second)); got.Accepted != 3 {
		t.Errorf("Today() a second before the midnight the day began under = %+v, want its three tasks", got)
	}
	want := profile.Daily{
		Date: profile.DateOf(afternoon), Ends: profile.At(time.Date(2026, 10, 11, 5, 0, 0, 0, time.UTC)), UTCOffset: -5 * 60,
	}
	if got := p.Daily.Today(utcMidnight); got != want {
		t.Errorf("Today() at that midnight = %+v, want a day begun by the family's clock, %+v", got, want)
	}
}

// A clock told late in the family's day carries none of its tasks into the
// next: a family five hours behind UTC that reached twenty tasks, and whose
// clock a second device tells as UTC's at eight in the evening, starts the
// next day at its own midnight with nothing counted.
func TestAStrayClockCarriesNoTaskIntoTheNextDay(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	p.SetClock(-5 * 60)
	evening := time.Date(2026, 10, 10, 23, 0, 0, 0, time.UTC) // six in the evening for the family
	for range 20 {
		p.CountAccepted(evening)
	}

	p.SetClock(0) // eight in the evening for the family
	if got := p.Daily.Today(time.Date(2026, 10, 11, 4, 59, 59, 0, time.UTC)); got.Accepted != 20 {
		t.Errorf("Today() a second before the family's midnight = %+v, want the day's twenty tasks", got)
	}
	if got := p.Daily.Today(time.Date(2026, 10, 11, 5, 0, 0, 0, time.UTC)); got.Accepted != 0 {
		t.Errorf("Today() at the family's midnight = %+v, want nothing counted: the tasks were the day before's", got)
	}
}

// A day that lies ahead of the clock, which only a file edited by hand or an
// instance whose clock runs fast leaves, counts as none: the counters of a day
// not begun hold nobody back.
func TestADayAheadOfTheClockCountsAsNone(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	ahead := profile.DateOf(now.AddDate(0, 0, 3))
	p.Daily = profile.Daily{Date: ahead, Ends: profile.At(ahead.AddDate(0, 0, 1)), Accepted: 20}

	if got := p.Daily.Today(now); got.Accepted != 0 || !got.Date.Equal(profile.DateOf(now).Time) {
		t.Errorf("Today() = %+v, want nothing counted on the day now falls in", got)
	}
}

// Answering a task is not what the daily limit counts: the limit is on tasks
// handed out, and one task is answered once whatever else happens.
func TestAnAnswerDoesNotTouchTheDailyCounters(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	p.CountAccepted(issued)
	before := p.Daily

	give(t, p, answering(t, p, "counting.gaps", 3), rightLetter, issued.Add(time.Minute))
	if p.Daily != before {
		t.Errorf("the counters moved to %+v on an answer, want %+v", p.Daily, before)
	}
}
