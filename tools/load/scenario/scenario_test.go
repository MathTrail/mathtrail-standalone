package scenario_test

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/scenario"
	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// optionsOf are the options of the scenario named, as a case starts from them.
func optionsOf(t *testing.T, name string) scenario.Options {
	t.Helper()

	o, err := scenario.Defaults(name)
	if err != nil {
		t.Fatalf("Defaults(%q) error = %v", name, err)
	}
	return o
}

// lessonAtOnce is a lesson with no pause between its steps.
func lessonAtOnce(t *testing.T) scenario.Options {
	t.Helper()

	o := optionsOf(t, scenario.Lesson)
	o.Pace = 0
	return o
}

// runs runs the options against the target, and is what the run came to.
func runs(ctx context.Context, t *testing.T, o *scenario.Options, target session.Target) []report.Run {
	t.Helper()

	came, err := scenario.Run(ctx, o, scenario.At(target))
	if err != nil {
		t.Fatalf("Run(%s) error = %v", o.Scenario, err)
	}
	return came
}

// one is the one run a scenario came to.
func one(t *testing.T, came []report.Run) *report.Run {
	t.Helper()

	if len(came) != 1 {
		t.Fatalf("the scenario came to %d runs, want one", len(came))
	}
	return &came[0]
}

// inFront is the service at the target behind what the case puts in front of
// it.
func inFront(t *testing.T, target session.Target, front func(next http.Handler) http.Handler) session.Target {
	t.Helper()

	backend, err := url.Parse(target.URL)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", target.URL, err)
	}
	server := httptest.NewServer(front(httputil.NewSingleHostReverseProxy(backend)))
	t.Cleanup(server.Close)
	return session.Target{URL: server.URL, Host: target.Host}
}

// A lesson against the service walks one child through every task: each asked
// for with its package, accepted with the card told on either side of the
// hand-in, and answered, and the progress read at the end — with every call
// answered as asked, nothing the service should never do, and what one
// accepted task cost.
func TestALessonIsAcceptedTaskByTask(t *testing.T) {
	t.Parallel()

	// Walked without a pause, a lesson is faster than an account's pace allows.
	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600")
	o := lessonAtOnce(t)
	run := one(t, runs(t.Context(), t, &o, target))

	if run.Units != o.Tasks {
		t.Errorf("%d tasks accepted, want all %d", run.Units, o.Tasks)
	}
	if hards := run.Verdict(); len(hards) != 0 {
		t.Errorf("Verdict() = %+v, want nothing", hards)
	}
	if got, want := len(run.Calls), 2+6*o.Tasks+1; got != want {
		t.Errorf("%d calls, want %d: the profile read and saved, six for each task — asked for, its package, the "+
			"card's two looks, handed in, answered — and the progress", got, want)
	}
	if run.Requests <= len(run.Calls) {
		t.Errorf("%d requests for %d calls, want the handshake counted too", run.Requests, len(run.Calls))
	}
	if _, done := run.Cost(report.Instance{VCPU: 1, GiB: 0.5}); !done {
		t.Error("Cost() says no work was done, want the cost of an accepted task")
	}
}

// A lesson that does not come to every task accepted, every call answered as
// asked, is a failed run, whatever held it back: here a pace nothing gets far
// under.
func TestALessonTheServiceHoldsBackFailsTheRun(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=3")
	o := lessonAtOnce(t)
	run := one(t, runs(t.Context(), t, &o, target))

	hards := run.Verdict()
	if len(hards) == 0 || !strings.Contains(hards[len(hards)-1].What, session.Paced.String()) {
		t.Errorf("Verdict() = %+v, want the lesson failed and the pace named", hards)
	}
}

// A scenario nobody wrote has no options to start from, and says which do.
func TestAScenarioNobodyWroteHasNoDefaults(t *testing.T) {
	t.Parallel()

	if _, err := scenario.Defaults("stampede"); err == nil || !strings.Contains(err.Error(), scenario.Lesson) {
		t.Errorf("Defaults(stampede) error = %v, want one naming the scenarios there are", err)
	}
}

// Every scenario there is starts from options a run can keep to, and reads
// some of the options the command line sets and nothing else.
func TestEveryScenarioStartsFromOptionsItCanKeepTo(t *testing.T) {
	t.Parallel()

	set := []string{"tasks", "pace", "timeout", "rate", "duration", "children", "variants", "steps", "accounts"}
	for _, name := range scenario.Names() {
		o := optionsOf(t, name)
		if err := o.Check(); err != nil || o.Lasts() <= 0 {
			t.Errorf("Defaults(%q) = %+v, lasting %v, and Check() = %v; want options a run keeps to", name, o, o.Lasts(), err)
		}
		reads := scenario.Reads(name)
		if len(reads) == 0 || slices.ContainsFunc(reads, func(option string) bool { return !slices.Contains(set, option) }) {
			t.Errorf("Reads(%q) = %q, want some of %q and nothing else", name, reads, set)
		}
	}
}

// Options no run could keep to are refused before anything is sent.
func TestOptionsNoRunCouldKeepToAreRefused(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, scenario string
		spoil          func(*scenario.Options)
	}{
		{"no task", scenario.Lesson, func(o *scenario.Options) { o.Tasks = 0 }},
		{"more tasks than written", scenario.Lesson, func(o *scenario.Options) { o.Tasks = 6 }},
		{"a pause that goes back", scenario.Lesson, func(o *scenario.Options) { o.Pace = -time.Second }},
		{"quiet addresses that never fetch", scenario.Limits, func(o *scenario.Options) { o.QuietEvery = 0 }},
		{"an attack over before a quiet fetch", scenario.Limits, func(o *scenario.Options) { o.Duration = o.QuietEvery }},
		{"a call that may not last", scenario.Lesson, func(o *scenario.Options) { o.Timeout = 0 }},
		{"no rate", scenario.Saturation, func(o *scenario.Options) { o.Rate = 0 }},
		{"a rate that is no number", scenario.Saturation, func(o *scenario.Options) { o.Rate = math.NaN() }},
		{"a flood", scenario.Saturation, func(o *scenario.Options) { o.Rate = 1001 }},
		{"a rate too slow to call", scenario.Saturation, func(o *scenario.Options) { o.Rate = 1e-12 }},
		{"an attack of no time", scenario.Saturation, func(o *scenario.Options) { o.Duration = 0 }},
		{"an attack over before its first call", scenario.Saturation, func(o *scenario.Options) { o.Duration = 300 * time.Millisecond }},
		{"no child", scenario.Saturation, func(o *scenario.Options) { o.Children = 0 }},
		{"a crowd", scenario.Saturation, func(o *scenario.Options) { o.Children = 1001 }},
		{"a ceiling no solver is sized to", scenario.Adversarial, func(o *scenario.Options) { o.Steps = lesson.MinSteps - 1 }},
		{"no variant", scenario.Adversarial, func(o *scenario.Options) { o.Variants = nil }},
		{"a variant nobody wrote", scenario.Adversarial, func(o *scenario.Options) { o.Variants = []string{lesson.Pairs, "sleep"} }},
		{"one account alone", scenario.Paces, func(o *scenario.Options) { o.Accounts = []scenario.Account{{Name: "parent"}} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			o := optionsOf(t, test.scenario)
			test.spoil(&o)
			if err := o.Check(); err == nil {
				t.Errorf("Check(%+v) = nil, want the options refused", o)
			}
		})
	}
}

// A run asked to stop before its end is the part that ran, marked as stopped,
// not an error: what it did is still worth reporting. It is held to no
// expectation, since it did not get to meet them.
func TestARunAskedToStopIsThePartThatRan(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600")
	o := lessonAtOnce(t)
	o.Pace = time.Second
	ctx, cancel := context.WithTimeout(t.Context(), 1500*time.Millisecond)
	defer cancel()

	run := one(t, runs(ctx, t, &o, target))
	if !run.Stopped || len(run.Calls) == 0 || run.Units == o.Tasks || len(run.Broken) != 0 {
		t.Errorf("Run() = stopped %v with %d calls, %d tasks accepted and expectations %q, "+
			"want a stopped run of the part that ran, held to no expectation",
			run.Stopped, len(run.Calls), run.Units, run.Broken)
	}
}

// A task the service accepted counts, even when the run is stopped right after
// it, before the child gets to answer it.
func TestATaskAcceptedJustBeforeTheRunStopsCounts(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(t.Context())
	defer stop()
	var handedIn atomic.Bool
	target := inFront(t, servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600"), func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The run stops at the child's first call after the hand-in: the
			// task has counted by then, and the child has not answered it. No
			// clock decides it, so a slow machine cannot let a second task in.
			tool := r.Header.Get("Mcp-Name")
			if tool != "" && handedIn.Load() {
				stop()
			}
			if tool == "submit_task" {
				handedIn.Store(true)
			}
			next.ServeHTTP(w, r)
		})
	})
	o := lessonAtOnce(t)
	o.Pace = time.Second

	run := one(t, runs(ctx, t, &o, target))
	if !run.Stopped || run.Units != 1 {
		t.Errorf("Run() = stopped %v with %d tasks accepted, want a stopped run that counts the one accepted",
			run.Stopped, run.Units)
	}
}
