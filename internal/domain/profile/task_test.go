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

// Answering a task is not what the daily limit counts: the limit is on tasks
// handed out, and one task is answered once whatever else happens.
func TestAnAnswerDoesNotTouchTheDailyCounters(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	p.CountAccepted(issued)
	before := p.Daily

	id := answering(t, p, "counting.gaps", 3)
	if _, err := p.Record(profile.Answered{TaskID: id, Correct: true, At: issued.Add(time.Minute)}); err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}
	if p.Daily != before {
		t.Errorf("the counters moved to %+v on an answer, want %+v", p.Daily, before)
	}
}
