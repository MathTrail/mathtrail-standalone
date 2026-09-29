package scenario_test

import (
	"context"
	"errors"
	"net/http"
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

// run runs the options against the services the launch brings up, and is
// what the run came to.
func run(ctx context.Context, t *testing.T, o *scenario.Options, launch scenario.Launch) []report.Run {
	t.Helper()

	came, err := scenario.Run(ctx, o, launch)
	if err != nil {
		t.Fatalf("Run(%s) error = %v", o.Scenario, err)
	}
	return came
}

// inProcess launches the service in the same process, as servicetest starts
// it, and a service of its own every time: measured by nothing, and stopped
// with the test.
func inProcess(t *testing.T, env ...string) scenario.Launch {
	t.Helper()
	return func(context.Context) (scenario.Launched, error) {
		asked := time.Now()
		target := servicetest.Start(t, env...)
		return scenario.Launched{Target: target, Asked: asked, Began: asked, Healthy: time.Now()}, nil
	}
}

// measured is the launch given, with an instance the run measures and ends:
// it spent a tenth of every stretch of time asked about and held the memory
// given, and its life is the one given.
func measured(launch scenario.Launch, memory uint64, life *report.Life, stopped *atomic.Int32) scenario.Launch {
	return func(ctx context.Context) (scenario.Launched, error) {
		up, err := launch(ctx)
		up.Spent = func(from, to time.Time) (report.Spent, bool) {
			return report.Spent{CPU: to.Sub(from) / 10, Memory: memory}, true
		}
		up.Peak = func() (uint64, bool) { return memory, true }
		up.Stop = func(context.Context) (report.Life, error) {
			stopped.Add(1)
			return *life, nil
		}
		return up, err
	}
}

// A service that never answered its probe is a run of its own, and a failed
// one: nothing the scenario would have done to it is done.
func TestAServiceThatNeverAnsweredItsProbeFailsTheRun(t *testing.T) {
	t.Parallel()

	launch := func(context.Context) (scenario.Launched, error) {
		return scenario.Launched{Target: session.Target{URL: "http://127.0.0.1:1"}, Asked: time.Now(), Began: time.Now()}, nil
	}
	o := lessonAtOnce(t)
	started := one(t, run(t.Context(), t, &o, launch))
	if started.Scenario != "lesson: start" || !strings.Contains(strings.Join(started.Broken, "; "), "never answered its probe") {
		t.Errorf("Run() = %q failed as %q, want the start of the lesson failed for a service that never answered",
			started.Scenario, started.Broken)
	}
}

// Each costly solver is handed to a service launched for it, so that one that
// ends its instance leaves the next a fresh one.
func TestEachVariantIsHandedToAServiceOfItsOwn(t *testing.T) {
	t.Parallel()

	var launched atomic.Int32
	inner := inProcess(t, "MATHTRAIL_SOLVER_STEPS=200000", "MATHTRAIL_RATE_USER_PER_MIN=600")
	launch := func(ctx context.Context) (scenario.Launched, error) {
		launched.Add(1)
		return inner(ctx)
	}
	o := optionsOf(t, scenario.Adversarial)
	o.Variants = []string{lesson.Loop, lesson.Appends}
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 400*time.Millisecond, 1, lesson.MinSteps, 10*time.Second

	came := run(t.Context(), t, &o, launch)
	if len(came) != 4 || launched.Load() != 2 {
		t.Errorf("adversarial came to %d runs over %d services, want two variants on a service each", len(came), launched.Load())
	}
	failedNone(t, came)
}

// What the instance spent while each run went on goes with the run, and the
// life of the instance with the last run it served: an instance killed for
// its memory fails that run, however well its calls were answered.
func TestWhatTheInstanceSpentAndHowItLivedGoWithTheRuns(t *testing.T) {
	t.Parallel()

	var stopped atomic.Int32
	killed := report.Life{Measured: true, OOMKilled: true, Exited: true, ExitCode: 137}
	launch := measured(inProcess(t, "MATHTRAIL_RATE_USER_PER_MIN=600"), 64<<20, &killed, &stopped)
	o := optionsOf(t, scenario.Saturation)
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 400*time.Millisecond, 1, lesson.MinSteps, 10*time.Second

	came := run(t.Context(), t, &o, launch)
	if len(came) != 2 || stopped.Load() != 1 {
		t.Fatalf("saturation came to %d runs and stopped %d instances, want the attack and its recovery on one", len(came), stopped.Load())
	}
	for i := range came {
		if came[i].Spent == nil || came[i].Spent.Memory != 64<<20 {
			t.Errorf("%s: Spent = %+v, want what the instance spent while it went on", came[i].Scenario, came[i].Spent)
		}
	}
	attack, recovered := &came[0], &came[1]
	var said []string
	for _, hard := range recovered.Verdict() {
		said = append(said, hard.What)
	}
	if len(attack.Lives) != 0 || len(recovered.Lives) != 1 || !strings.Contains(strings.Join(said, "; "), "killed for want of memory") {
		t.Errorf("the attack has %d lives and the recovery %d, failed as %q; want the life with the last run, failing it",
			len(attack.Lives), len(recovered.Lives), said)
	}
}

// A service that could not be launched at all stops the scenario there: what
// ran before it is kept, and the error says why nothing more could.
func TestALaunchThatFailsLeavesWhatRanBeforeIt(t *testing.T) {
	t.Parallel()

	var launched atomic.Int32
	inner := inProcess(t, "MATHTRAIL_SOLVER_STEPS=200000", "MATHTRAIL_RATE_USER_PER_MIN=600")
	launch := func(ctx context.Context) (scenario.Launched, error) {
		if launched.Add(1) == 2 {
			return scenario.Launched{}, errors.New("no daemon to start a container")
		}
		return inner(ctx)
	}
	o := optionsOf(t, scenario.Adversarial)
	o.Variants = []string{lesson.Loop, lesson.Appends}
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 400*time.Millisecond, 1, lesson.MinSteps, 10*time.Second

	came, err := scenario.Run(t.Context(), &o, launch)
	if err == nil || len(came) != 2 {
		t.Errorf("Run() = %d runs, error %v; want the first variant's two runs and the error of the second", len(came), err)
	}
}

// A cold run starts the service again and again, and measures every start:
// when it was asked for, when its process began, when its probe first
// answered and when its first call did, and the memory it held then — with
// the life of every instance it started.
func TestColdStartsAreMeasuredOneByOne(t *testing.T) {
	t.Parallel()

	var stopped atomic.Int32
	launch := measured(inProcess(t), 48<<20, &report.Life{Measured: true}, &stopped)
	o := optionsOf(t, scenario.Cold)
	o.Timeout = 10 * time.Second

	cold := one(t, run(t.Context(), t, &o, launch))
	if len(cold.Starts) != 5 || len(cold.Lives) != 5 || stopped.Load() != 5 {
		t.Fatalf("cold came to %d starts and %d lives, %d stopped; want five of each", len(cold.Starts), len(cold.Lives), stopped.Load())
	}
	for i, start := range cold.Starts {
		if start.Answered.IsZero() || start.Healthy.Before(start.Began) || start.Began.Before(start.Asked) ||
			start.Answered.Before(start.Healthy) || start.Memory != 48<<20 {
			t.Errorf("start %d = %+v, want it asked for, begun, answering its probe and its first call in turn, "+
				"and the memory it held", i+1, start)
		}
	}
	if hards := cold.Verdict(); len(hards) != 0 {
		t.Errorf("Verdict() = %+v, want nothing", hards)
	}
}

// Runs that went on together on one instance are given no spending of their
// own, since the cgroup counts the instance as a whole: what it spent then
// was theirs together. A run that went on alone is given what it spent.
func TestRunsThatWentOnTogetherAreGivenNoSpendingOfTheirOwn(t *testing.T) {
	t.Parallel()

	var stopped atomic.Int32
	launch := measured(inProcess(t,
		"MATHTRAIL_RATE_USER_PER_MIN=60", "MATHTRAIL_RATE_IP_PER_MIN=120", "MATHTRAIL_RATE_INSTANCE_PER_MIN=6000",
	), 64<<20, &report.Life{Measured: true}, &stopped)
	o := optionsOf(t, scenario.Limits)
	o.Rate, o.Duration, o.Pace, o.QuietEvery, o.Timeout = 50, time.Second, 0, 100*time.Millisecond, 10*time.Second

	came := run(t.Context(), t, &o, launch)
	if len(came) != 5 {
		t.Fatalf("limits came to %d runs, want its four parts and the recovery", len(came))
	}
	for i := range came[:4] {
		if came[i].Spent != nil {
			t.Errorf("%s: Spent = %+v, want none for a part that went on beside the others", came[i].Scenario, came[i].Spent)
		}
	}
	if came[4].Spent == nil {
		t.Errorf("%s: Spent = nil, want what the instance spent while the recovery went on alone", came[4].Scenario)
	}
}

// A recovery does not wait for a service whose instance is known to have
// ended: it fails at once, and says so, rather than at the end of its time.
func TestARecoveryDoesNotWaitForAnInstanceThatEnded(t *testing.T) {
	t.Parallel()

	target := inFront(t, servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600"), func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	launch := func(context.Context) (scenario.Launched, error) {
		now := time.Now()
		return scenario.Launched{Target: target, Asked: now, Began: now, Healthy: now, Alive: func() bool { return false }}, nil
	}
	o := optionsOf(t, scenario.Saturation)
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 400*time.Millisecond, 1, lesson.MinSteps, 30*time.Second

	came := run(t.Context(), t, &o, launch)
	recovered := &came[len(came)-1]
	if took := recovered.Ended.Sub(recovered.Began); took > 5*time.Second ||
		!strings.Contains(strings.Join(recovered.Broken, "; "), "its instance ended") {
		t.Errorf("%s: failed as %q after %v; want it failed at once for an instance that ended", recovered.Scenario, recovered.Broken, took)
	}
}

// An instance the run can end and not measure is ended all the same, and its
// runs are given no spending rather than a spending nothing counted.
func TestAnInstanceTheRunCannotMeasureIsEndedAllTheSame(t *testing.T) {
	t.Parallel()

	var stopped atomic.Int32
	inner := inProcess(t, "MATHTRAIL_RATE_USER_PER_MIN=600")
	launch := func(ctx context.Context) (scenario.Launched, error) {
		up, err := inner(ctx)
		up.Stop = func(context.Context) (report.Life, error) {
			stopped.Add(1)
			return report.Life{}, nil
		}
		return up, err
	}
	o := lessonAtOnce(t)
	o.Tasks = 1

	lessoned := one(t, run(t.Context(), t, &o, launch))
	if lessoned.Spent != nil || len(lessoned.Lives) != 1 || stopped.Load() != 1 {
		t.Errorf("Run() = spent %+v, %d lives, %d stopped; want no spending, and the instance ended and its life kept",
			lessoned.Spent, len(lessoned.Lives), stopped.Load())
	}
}
