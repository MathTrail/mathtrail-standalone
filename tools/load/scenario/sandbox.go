package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// examined is the unit of the attacks on the sandbox: a hand-in whose solver
// the service ran before it turned the task away.
const examined = "examined hand-in"

// runSaturation hands in more solvers than the sandbox has slots, and then
// looks at whether the service came back. A hand-in that finds every slot
// taken for as long as it may wait is turned away as busy: that is the
// service keeping its answers in time, not failing.
func runSaturation(ctx context.Context, o *Options, target session.Target) []report.Run {
	return recovered(ctx, o, target, Saturation, handIns(ctx, o, target, Saturation, lesson.Loop))
}

// runAdversarial hands in each costly solver of the options in turn, each
// followed by a look at whether the service came back from it.
func runAdversarial(ctx context.Context, o *Options, target session.Target) []report.Run {
	runs := make([]report.Run, 0, 2*len(o.Variants))
	for _, variant := range o.Variants {
		if ctx.Err() != nil {
			break
		}
		name := Adversarial + ": " + variant
		runs = append(runs, recovered(ctx, o, target, name, handIns(ctx, o, target, name, variant))...)
	}
	return runs
}

// handIns hands in the costly solver named at the rate of the options, for as
// long as they say, each time by the next of the children in turn and for a
// request nobody opened: the sandbox runs every solver it takes, both runs,
// and the service writes nothing and spends nothing on it.
func handIns(ctx context.Context, o *Options, target session.Target, name, solver string) report.Run {
	run := report.Run{Scenario: name, Unit: examined, Began: time.Now()}
	task, err := lesson.CostlyTask(solver, o.Steps)
	if err != nil {
		run.Ended, run.Broken = run.Began, []string{err.Error()}
		return run
	}
	service := session.Open(target, o.Timeout)
	defer service.Close()
	children := make([]*session.Child, o.Children)
	began := stamp()
	for i := range children {
		children[i] = service.Child(fmt.Sprintf("%s-%s-%s-%d", o.Scenario, solver, began, i))
	}

	hits, unsent := attack(ctx, every(o.Rate), o.Duration, inFlight, func(ctx context.Context, n uint64) session.Answer {
		return lesson.HandInStale(ctx, children[n%uint64(len(children))], &task)
	})
	finish(ctx, &run)
	run.Calls, run.Unsent = callsOf(hits), unsent
	run.Requests = service.Requests()
	for _, child := range children {
		run.Reconnects += child.Reconnects()
		_ = child.Close()
	}
	run.Units = countOf(&run, stale)
	if run.Units == 0 {
		// Turned away every time, the attack never reached the sandbox it is
		// meant to load, and proves nothing about it.
		run.Broken = append(run.Broken, "no hand-in was examined: every one was turned away before its solver ran")
	}
	if other := kindsBut(&run, stale, session.Busy, session.Paced); other != "" {
		run.Broken = append(run.Broken, "hand-ins answered otherwise than a task for no request is: "+other)
	}
	return run
}

// stale is what a task handed in for no request comes to once its solver has
// run.
var stale = session.Kind{Class: session.Stale, Detail: "stale_request"}

// countOf is how many calls of the run came to the kind given.
func countOf(run *report.Run, kind session.Kind) int {
	count := 0
	for i := range run.Calls {
		if run.Calls[i].Kind == kind {
			count++
		}
	}
	return count
}
