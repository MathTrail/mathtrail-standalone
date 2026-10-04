package scenario_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/scenario"
	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// The paces of a deployed service hold back the greedy alone: the account
// calling past its pace and the address fetching past its own, never the
// account walking its lesson beside them. No recovery follows, since nothing
// was attacked that a child new to the service would have to find working.
func TestPacesHoldBackTheGreedyAlone(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t,
		"MATHTRAIL_RATE_USER_PER_MIN=60", "MATHTRAIL_RATE_IP_PER_MIN=120", "MATHTRAIL_RATE_INSTANCE_PER_MIN=6000")
	o := optionsOf(t, scenario.Paces)
	o.Rate, o.Duration, o.Pace, o.Timeout, o.Tasks = 50, 2*time.Second, 0, 10*time.Second, 1

	came := runs(t.Context(), t, &o, target)
	if len(came) != 3 {
		t.Fatalf("paces came to %d runs, want the three parts and nothing after them", len(came))
	}
	for _, part := range []struct {
		run      int
		wantHeld bool
	}{
		{0, true},
		{1, false},
		{2, true},
	} {
		counted := kinds(&came[part.run])
		if held := counted["limited:paced"] > 0; held != part.wantHeld || counted["ok"] == 0 {
			t.Errorf("%s came to %v; want held back: %v, and answered as asked at least once",
				came[part.run].Scenario, counted, part.wantHeld)
		}
	}
	failedNone(t, came)
}

// The children of a run of accounts are the accounts, in order: the first is
// the greedy one, and the next walks its lesson with its own token, so that the
// profile the lesson leaves is the second account's.
func TestARunOfAccountsSignsItsChildrenInWithThem(t *testing.T) {
	t.Parallel()

	// The development sign-in takes a token for the name of an account, which
	// stands in here for the tokens a parent's sign-in gives.
	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=60", "MATHTRAIL_RATE_IP_PER_MIN=120")
	o := optionsOf(t, scenario.Paces)
	o.Rate, o.Duration, o.Pace, o.Timeout, o.Tasks = 50, time.Second, 0, 10*time.Second, 1
	o.Accounts = []scenario.Account{bearing("paces-first"), bearing("paces-second")}

	came := runs(t.Context(), t, &o, target)
	failedNone(t, came)
	if got := kinds(&came[1])["ok"]; got == 0 {
		t.Fatalf("the second account's lesson came to %v, want it answered", kinds(&came[1]))
	}

	service := session.Open(target, 10*time.Second)
	defer service.Close()
	for _, check := range []struct {
		account     string
		wantStudent bool
	}{
		{"paces-first", false},
		{"paces-second", true},
	} {
		profile := service.Child(check.account).Call(t.Context(), "get_profile", map[string]any{})
		if hasStudent := strings.Contains(profile.Text(), "Otter"); hasStudent != check.wantStudent {
			t.Errorf("the profile of %s holds the lesson's student: %v, want %v; it says %q",
				check.account, hasStudent, check.wantStudent, profile.Text())
		}
	}
}

// An account that walked the lessons of paces before is handed the same tasks
// again, which its profile refuses as near-copies: the run fails, and says that
// the account's profile is why rather than leaving it to look like the service.
func TestPacesOnAnAccountThatWalkedThemSaysWhy(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=60", "MATHTRAIL_RATE_IP_PER_MIN=120")
	o := optionsOf(t, scenario.Paces)
	o.Rate, o.Duration, o.Pace, o.Timeout, o.Tasks = 50, time.Second, 0, 10*time.Second, 1
	o.Accounts = []scenario.Account{bearing("again-first"), bearing("again-second")}

	failedNone(t, runs(t.Context(), t, &o, target))
	again := runs(t.Context(), t, &o, target)
	const says = "had been handed the load's tasks before"
	if broken := strings.Join(again[1].Broken, "; "); !strings.Contains(broken, says) {
		t.Errorf("%s the second time: Broken = %q, want the run failed as %q", again[1].Scenario, broken, says)
	}
}

// An account whose tokens cannot be renewed never reaches the service: each of
// its calls ends unanswered, as unsigned, and fails its part of the run.
func TestAnAccountThatCannotSignFailsItsPart(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600", "MATHTRAIL_RATE_IP_PER_MIN=600")
	o := optionsOf(t, scenario.Paces)
	o.Rate, o.Duration, o.Pace, o.Timeout, o.Tasks = 5, time.Second, 0, 10*time.Second, 1
	o.Accounts = []scenario.Account{
		{Name: "lapsed", Token: func() (string, error) { return "", errors.New("the refresh token has ended") }},
		bearing("paces-beside"),
	}

	came := runs(t.Context(), t, &o, target)
	const says = "answered otherwise than as asked or held back: no_answer:unsigned"
	if broken := strings.Join(came[0].Broken, "; "); !strings.Contains(broken, says) {
		t.Errorf("%s: Broken = %q, want the run failed as %q", came[0].Scenario, broken, says)
	}
	if came[0].Requests != 0 {
		t.Errorf("%s sent %d requests, want none: a request that cannot be signed is not sent",
			came[0].Scenario, came[0].Requests)
	}
}

// The day refuses the failing child once its requests have failed as often as
// the day allows — every hand-in refused by the checks, every request out of
// attempts — and the child reading its profile beside it is never refused.
func TestTheDayRefusesTheFailingChildAlone(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_DAILY_FAILED=2", "MATHTRAIL_RATE_USER_PER_MIN=600")
	o := optionsOf(t, scenario.Ceiling)
	o.Pace, o.QuietEvery, o.Duration, o.Timeout = 0, 50*time.Millisecond, 30*time.Second, 10*time.Second

	came := runs(t.Context(), t, &o, target)
	if len(came) != 2 {
		t.Fatalf("the ceiling came to %d runs, want the failing child and the others", len(came))
	}
	failedNone(t, came)
	if came[0].Units != 2 {
		t.Errorf("%s failed %d requests, want the day's two", came[0].Scenario, came[0].Units)
	}
	counted := kinds(&came[0])
	if counted["limited:limit_reached"] != 1 {
		t.Errorf("%s came to %v, want the day's refusal once", came[0].Scenario, counted)
	}
	refused := 0
	for kind, count := range counted {
		if strings.HasPrefix(kind, "rejected:") {
			refused += count
		}
	}
	if refused != 2*3 {
		t.Errorf("%s had %d hand-ins refused by the checks, want three for each failed request; it came to %v",
			came[0].Scenario, refused, counted)
	}
	if kinds(&came[1])["ok"] == 0 {
		t.Errorf("%s came to %v, want reads answered", came[1].Scenario, kinds(&came[1]))
	}
}

// A day that never refuses the failing child within the run's time fails the
// run: the ceiling it was to show is not there.
func TestADayThatNeverRefusesFailsTheRun(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_DAILY_FAILED=100", "MATHTRAIL_RATE_USER_PER_MIN=6000")
	o := optionsOf(t, scenario.Ceiling)
	o.Pace, o.QuietEvery, o.Duration, o.Timeout = 0, 50*time.Millisecond, 500*time.Millisecond, 10*time.Second

	came := runs(t.Context(), t, &o, target)
	const says = "the day never refused the failing child a request"
	if broken := strings.Join(came[0].Broken, "; "); !strings.Contains(broken, says) {
		t.Errorf("%s: Broken = %q, want the run failed as %q", came[0].Scenario, broken, says)
	}
}

// The task the ceiling hands in is one the checks refuse for its answer alone:
// everything else of it is the race as a lesson hands it in.
func TestTheRefusedTaskDiffersFromTheRaceInItsAnswerAlone(t *testing.T) {
	t.Parallel()

	race, refused := lesson.Written()[0], lesson.Refused()
	if refused.Body.Question != race.Body.Question || refused.Solver != race.Solver {
		t.Fatal("the refused task is not the race")
	}
	if refused.Body.Correct != race.Wrong || refused.SelfCheck.FinalAnswer != race.Wrong {
		t.Errorf("the refused task answers %q with a self-check of %q, want the race's wrong letter %q for both",
			refused.Body.Correct, refused.SelfCheck.FinalAnswer, race.Wrong)
	}
	if _, explains := refused.Body.Distractors[race.Body.Correct]; !explains {
		t.Errorf("the refused task explains no wrong answer at %q, the option its solver proves", race.Body.Correct)
	}
	if _, explains := refused.Body.Distractors[refused.Body.Correct]; explains {
		t.Errorf("the refused task explains its own answer %q as a wrong one", refused.Body.Correct)
	}
}

// bearing is an account whose token is the name given, which the development
// sign-in signs in as an account of that name.
func bearing(name string) scenario.Account {
	return scenario.Account{Name: name, Token: func() (string, error) { return name, nil }}
}
