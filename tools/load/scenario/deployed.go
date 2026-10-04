package scenario

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// The scenarios of a deployed service, paces and the ceiling, are the limits
// as a deployed service lets a run try them. Its children are accounts a
// parent signed in for real, whose profiles live from one run to the next, so
// they are never children new to the service; and the platform in front of it
// writes the tool's own address into every request, so a run speaks from one
// address alone. Neither scenario makes up quiet addresses, and neither ends
// with a lesson for a child new to the service. Against a service with the
// development sign-in, both make their children up as the other scenarios do.

// dayIsFull is the refusal of a day whose ceiling an account has reached: no
// new task until tomorrow, in the words the child is told.
var dayIsFull = session.Kind{Class: session.Limited, Detail: "limit_reached"}

// childOf is the i-th child of a run: the i-th account the options give,
// signed in with its own tokens, or else a child the development sign-in makes
// up under the name given.
func (o *Options) childOf(service *session.Service, i int, name string) *session.Child {
	if i < len(o.Accounts) {
		return service.SignedIn(o.Accounts[i].Name, o.Accounts[i].Token)
	}
	return service.Child(name)
}

// others is how many children of a run go beside its first: the accounts
// after the first, or the children the options ask for.
func (o *Options) others() int {
	if len(o.Accounts) > 0 {
		return len(o.Accounts) - 1
	}
	return o.Children
}

// runPaces runs three things at once for as long as the options say — the
// first account past its pace, the others walking lessons within theirs, and
// the tool's own address past the pace of an address — and holds each to what
// its pace should have done to it. The lessons hand in the load's own tasks,
// and an account's profile keeps every task it was handed: an account that
// has had them before is handed them again as near-copies, and refused.
func runPaces(ctx context.Context, o *Options, launch Launch) ([]report.Run, error) {
	return served(ctx, launch, Paces, func(up *Launched) []report.Run {
		var group sync.WaitGroup
		var runs [3]report.Run
		group.Go(func() { runs[0] = greedyChild(ctx, o, up.Target, Paces) })
		group.Go(func() { runs[1] = otherChildren(ctx, o, up.Target, Paces) })
		group.Go(func() { runs[2] = greedyAddress(ctx, o, up.Target, Paces) })
		group.Wait()
		return runs[:]
	})
}

// runCeiling has the first account fail request after request until the day
// refuses it one more, while the other accounts read their profiles: the day
// refuses that one alone, in the words the child is told, and the counter it
// is refused by lives in the account's profile, which every instance reads.
func runCeiling(ctx context.Context, o *Options, launch Launch) ([]report.Run, error) {
	return served(ctx, launch, Ceiling, func(up *Launched) []report.Run {
		failed := make(chan struct{})
		var group sync.WaitGroup
		var runs [2]report.Run
		group.Go(func() {
			defer close(failed)
			runs[0] = failingChild(ctx, o, up.Target)
		})
		group.Go(func() { runs[1] = readingChildren(ctx, o, up.Target, failed) })
		group.Wait()
		return runs[:]
	})
}

// failingChild asks for a task, hands in one the checks refuse until the
// request has no attempt left, and asks again, until the day refuses to open a
// request or the options' time is up, pausing before every call as the
// options say. Its profile is the student of the load for the rest of it.
func failingChild(ctx context.Context, o *Options, target session.Target) report.Run {
	service := session.Open(target, o.Timeout)
	defer service.Close()
	child := o.childOf(service, 0, Ceiling+"-failing-"+stamp())
	defer func() { _ = child.Close() }()

	run := report.Run{Scenario: Ceiling + ": the failing child", Unit: "failed request", Began: time.Now()}
	w := &walk{child: child, pace: o.Pace, run: &run}
	w.record(lesson.Start(ctx, child, student)...)
	refused := lesson.Refused()
	until := run.Began.Add(o.Duration)
	full := false
	for !full && time.Now().Before(until) && w.step(ctx) {
		full = w.fail(ctx, &refused)
	}
	finish(ctx, &run)
	run.Requests, run.Reconnects = service.Requests(), child.Reconnects()
	run.Broken = ceilingExpected(&run, full)
	return run
}

// fail asks for a task and hands the refused one in for it, as many times as a
// request has attempts, each after a pause: one failed request. It says
// whether the day refused to open the request instead.
func (w *walk) fail(ctx context.Context, refused *lesson.Task) (full bool) {
	request, asked, err := lesson.Ask(ctx, w.child, refused.Choice)
	w.record(asked...)
	if len(asked) > 0 && asked[0].Kind == dayIsFull {
		return true
	}
	if err != nil {
		return false
	}
	for range profile.MaxAttempts {
		if !w.step(ctx) {
			return false
		}
		_, handedIn := lesson.HandIn(ctx, w.child, request, refused, student)
		w.record(handedIn)
		if handedIn.Kind.Class != session.Rejected {
			return false
		}
	}
	w.run.Units++
	return false
}

// ceilingExpected is what the failing child's run did not come to that it
// should have: refused by the day, and otherwise every call answered as asked
// or, a hand-in, refused by the checks.
func ceilingExpected(run *report.Run, full bool) []string {
	var broken []string
	if !full {
		broken = append(broken, "the day never refused the failing child a request")
	}
	var other []string
	for _, kind := range report.Kinds(run.Calls) {
		if kind != session.Answered && kind != dayIsFull && kind.Class != session.Rejected {
			other = append(other, kind.String())
		}
	}
	if len(other) > 0 {
		broken = append(broken, "calls of the failing child answered otherwise than as asked, "+
			"refused by the checks or by the day: "+strings.Join(other, ", "))
	}
	return broken
}

// readingChildren read their profiles, the children after the first, each
// once at once and then once every quiet interval of the options, until the
// failing one is done: none of them should ever be refused.
func readingChildren(ctx context.Context, o *Options, target session.Target, failed <-chan struct{}) report.Run {
	service := session.Open(target, o.Timeout)
	defer service.Close()

	run := report.Run{Scenario: Ceiling + ": the other children", Unit: "read", Began: time.Now()}
	var (
		mu    sync.Mutex
		group sync.WaitGroup
	)
	began := stamp()
	for i := range o.others() {
		child := o.childOf(service, i+1, fmt.Sprintf("%s-%s-reading-%d", Ceiling, began, i))
		group.Go(func() {
			defer func() { _ = child.Close() }()
			for read := true; read; read = waitOrDone(ctx, o.QuietEvery, failed) {
				answer := child.Call(ctx, "get_profile", map[string]any{})
				mu.Lock()
				run.Calls = append(run.Calls, report.CallOf(&answer, answer.Started))
				mu.Unlock()
			}
		})
	}
	group.Wait()
	finish(ctx, &run)
	run.Requests = service.Requests()
	run.Units = len(run.Calls)
	if other := kindsBut(&run, session.Answered); other != "" {
		run.Broken = []string{"reads of the other children answered otherwise than as asked: " + other}
	}
	return run
}

// waitOrDone waits the pause given, and says whether to go on: not once the
// channel given is closed, or the run was asked to stop.
func waitOrDone(ctx context.Context, d time.Duration, done <-chan struct{}) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-done:
		return false
	case <-ctx.Done():
		return false
	}
}
