package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// starts is how many times a cold run starts the service: enough for a
// middle that one slow start does not move.
const starts = 5

// startLasts is about how long one start of a cold run takes: its container
// asked for, come up, answering, and stopped again.
const startLasts = 5 * time.Second

// runCold starts the service again and again, every start an instance of its
// own, and measures how long each took to come up: from the moment it was
// asked for and from the moment its process began to the first answer of its
// probe, and to the first answer of the tool a child calls first, the profile;
// and how much memory it held once it had answered. Every call is counted from
// the moment its service was asked for.
func runCold(ctx context.Context, o *Options, launch Launch) ([]report.Run, error) {
	run := report.Run{Scenario: Cold, Began: time.Now()}
	began := stamp()
	for i := range starts {
		if ctx.Err() != nil {
			break
		}
		if err := measureStart(ctx, o, launch, &run, fmt.Sprintf("%s-%s-%d", Cold, began, i)); err != nil {
			finish(ctx, &run)
			return []report.Run{run}, err
		}
	}
	finish(ctx, &run)
	return []report.Run{run}, nil
}

// measureStart starts the service once more and adds the start to the run:
// when it came up and answered its first call — made by the child named — the
// memory it held then, and the life of its instance.
func measureStart(ctx context.Context, o *Options, launch Launch, run *report.Run, child string) error {
	up, err := launch(ctx)
	if err != nil {
		return err
	}
	start := report.Start{Asked: up.Asked, Began: up.Began, Healthy: up.Healthy}
	if up.Healthy.IsZero() {
		run.Broken = append(run.Broken, fmt.Sprintf("start %d never answered its probe", len(run.Starts)+1))
	} else {
		answer, requests := firstCall(ctx, o, up.Target, child)
		run.Calls = append(run.Calls, report.CallOf(&answer, up.Asked))
		run.Requests += requests
		if answer.Kind == session.Answered {
			start.Answered = answer.Ended
		}
		if up.Peak != nil {
			start.Memory, _ = up.Peak()
		}
	}
	run.Starts = append(run.Starts, start)
	if up.Stop == nil {
		return nil
	}
	life, err := up.Stop(ctx)
	if err != nil {
		return err
	}
	run.Lives = append(run.Lives, life)
	return nil
}

// firstCall is the first call a child makes of a service just come up — its
// profile, with the handshake before it — and how many requests it took.
func firstCall(ctx context.Context, o *Options, target session.Target, name string) (answer session.Answer, requests int) {
	service := session.Open(target, o.Timeout)
	defer service.Close()
	child := service.Child(name)
	defer func() { _ = child.Close() }()
	answer = child.Call(ctx, "get_profile", map[string]any{})
	return answer, service.Requests()
}
