package scenario_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/scenario"
	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// lessonAtOnce is a lesson with no pause between its steps.
func lessonAtOnce(t *testing.T) scenario.Options {
	t.Helper()

	o, err := scenario.Defaults(scenario.Lesson)
	if err != nil {
		t.Fatalf("Defaults(%q) error = %v", scenario.Lesson, err)
	}
	o.Pace = 0
	return o
}

// A lesson against the service walks one child through every task: each asked
// for, accepted, answered, and the progress read at the end — with every call
// answered as asked, nothing the service should never do, and what one
// accepted task cost.
func TestALessonIsAcceptedTaskByTask(t *testing.T) {
	t.Parallel()

	// Walked without a pause, a lesson is faster than an account's pace allows.
	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600")
	o := lessonAtOnce(t)
	run, err := scenario.Run(t.Context(), o, target)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if run.Units != o.Tasks {
		t.Errorf("%d tasks accepted, want all %d", run.Units, o.Tasks)
	}
	if hards := run.Verdict(); len(hards) != 0 {
		t.Errorf("Verdict() = %+v, want nothing", hards)
	}
	if got, want := len(run.Calls), 2+3*o.Tasks+1; got != want {
		t.Errorf("%d calls, want %d: the profile read and saved, three for each task, and the progress", got, want)
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
	run, err := scenario.Run(t.Context(), lessonAtOnce(t), target)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

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

// Options no run could keep to are refused before anything is sent.
func TestOptionsNoRunCouldKeepToAreRefused(t *testing.T) {
	t.Parallel()

	for name, spoil := range map[string]func(*scenario.Options){
		"no task":                  func(o *scenario.Options) { o.Tasks = 0 },
		"more tasks than written":  func(o *scenario.Options) { o.Tasks = 6 },
		"a pause that goes back":   func(o *scenario.Options) { o.Pace = -time.Second },
		"a call that may not last": func(o *scenario.Options) { o.Timeout = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			o := lessonAtOnce(t)
			spoil(&o)
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

	run, err := scenario.Run(ctx, o, target)
	if err != nil {
		t.Fatalf("Run() error = %v, want the part that ran", err)
	}
	if !run.Stopped || len(run.Calls) == 0 || run.Units == o.Tasks || len(run.Broken) != 0 {
		t.Errorf("Run() = stopped %v with %d calls, %d tasks accepted and expectations %q, "+
			"want a stopped run of the part that ran, held to no expectation",
			run.Stopped, len(run.Calls), run.Units, run.Broken)
	}
}

// A task the service accepted counts, even when the run is stopped in the
// pause after it, before the child gets to answer it.
func TestATaskAcceptedJustBeforeTheRunStopsCounts(t *testing.T) {
	t.Parallel()

	service := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600")
	backend, err := url.Parse(service.URL)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", service.URL, err)
	}
	ctx, stop := context.WithCancel(t.Context())
	defer stop()
	proxy := httputil.NewSingleHostReverseProxy(backend)
	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
		if r.Header.Get("Mcp-Name") == "submit_task" {
			// Well after the answer is back, and well before the pause after it
			// is over.
			time.AfterFunc(300*time.Millisecond, stop)
		}
	}))
	defer front.Close()
	o := lessonAtOnce(t)
	o.Pace = time.Second

	run, err := scenario.Run(ctx, o, session.Target{URL: front.URL, Host: service.Host})
	if err != nil {
		t.Fatalf("Run() error = %v, want the part that ran", err)
	}
	if !run.Stopped || run.Units != 1 {
		t.Errorf("Run() = stopped %v with %d tasks accepted, want a stopped run that counts the one accepted",
			run.Stopped, run.Units)
	}
}
