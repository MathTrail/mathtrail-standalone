package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// probe is the path the platform asks whether the service is alive.
const probe = "/health"

// probeAgain is how long a service that has not come back is left before it
// is asked again, and the first pause before a lesson a pace held back is
// walked again.
const probeAgain = 250 * time.Millisecond

// recovery is the look at the service after an attack, as a run of its own:
// the probe answers, and then a child new to the service walks a lesson of one
// task through it without a pause, all within as long as a call may take. A
// service that never came back, or came back unable to give a lesson, fails
// the run.
//
// A lesson a pace held back is walked again, by another new child and after
// twice the pause each time: a pace shedding what is left of the attack is the
// service doing as it should, and it has the same time to come back as the
// probe has. Only the probe the service settled on and the last lesson are
// counted as its calls, the ones the recovery is judged by, while its
// requests are every one it sent, as the platform bills them; and the run
// begins when the service was first asked, so that its length is how long
// coming back took.
func recovery(ctx context.Context, o *Options, target session.Target, after string) report.Run {
	service := session.Open(target, o.Timeout)
	defer service.Close()
	name := after + ": recovery"
	began := time.Now()
	until := began.Add(o.Timeout)

	probed, alive := comeBack(ctx, service, until)
	asked := report.CallOf(&probed, probed.Started)
	if !alive {
		run := report.Run{
			Scenario: name, Unit: "accepted task", Began: began,
			Calls: []report.Call{asked}, Requests: service.Requests(),
			Broken: []string{fmt.Sprintf("the service did not come back within %v: its probe answered %s", o.Timeout, probed.Kind)},
		}
		finish(ctx, &run)
		return run
	}

	one := Options{Tasks: 1, Timeout: o.Timeout}
	run := walkLesson(ctx, &one, service, "recovery-"+stamp())
	for wait := probeAgain; heldBackOnly(&run); wait *= 2 {
		if time.Now().Add(wait).After(until) || !pause(ctx, wait) {
			break
		}
		run = walkLesson(ctx, &one, service, "recovery-"+stamp())
	}
	run.Scenario, run.Began = name, began
	run.Calls = append([]report.Call{asked}, run.Calls...)
	run.Requests = service.Requests()
	run.Broken = lessonExpected(&run, one.Tasks)
	return run
}

// recovered is the runs of an attack followed by the look at whether the
// service came back from it — unless the run was asked to stop, when the
// attack is not over for anything to come back from.
func recovered(ctx context.Context, o *Options, target session.Target, after string, attack ...report.Run) []report.Run {
	if ctx.Err() != nil {
		return attack
	}
	return append(attack, recovery(ctx, o, target, after))
}

// comeBack asks the probe until it answers or the moment given has come, and
// is the last answer it gave and whether that was an answer: a service may
// take a moment to come back from an attack, and only the answer it settled
// on says whether it did.
func comeBack(ctx context.Context, service *session.Service, until time.Time) (session.Answer, bool) {
	for {
		answer := service.Fetch(ctx, probe, "")
		if answer.Kind == session.Answered {
			return answer, true
		}
		if time.Now().Add(probeAgain).After(until) || !pause(ctx, probeAgain) {
			return answer, false
		}
	}
}

// heldBackOnly says whether a pace held back some call of the run, and
// nothing else stood in its way.
func heldBackOnly(run *report.Run) bool {
	return countOf(run, session.Paced) > 0 && kindsBut(run, session.Answered, session.Paced) == ""
}
