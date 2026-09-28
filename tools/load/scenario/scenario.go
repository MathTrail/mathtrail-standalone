// Package scenario is what a run does to the service. A lesson walks one
// child through the tools at a live pace; the runner sends calls at a constant
// pace however slowly they are answered, for the scenarios that load the
// service rather than walk through it.
package scenario

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// Options are what a run of a scenario is asked to do.
type Options struct {
	// Scenario is the name of the scenario.
	Scenario string
	// Tasks is how many tasks a lesson asks for.
	Tasks int
	// Pace is the pause before every step of a lesson: a model writing a task,
	// a child reading one.
	Pace time.Duration
	// Timeout is how long one call may take before it counts as unanswered.
	Timeout time.Duration
}

// Lesson is the name of the scenario of one child's lesson.
const Lesson = "lesson"

// scenario is what a scenario starts from, what it does to the service, and
// what it expects to have come to when it has run to its end.
type scenario struct {
	defaults Options
	run      func(context.Context, Options, *session.Service) report.Run
	expected func(*report.Run, *Options) []string
}

// scenarios are the scenarios there are, by name.
var scenarios = map[string]scenario{
	Lesson: {
		// Five tasks at a pause of five seconds is a lesson of a minute and a
		// half: short enough to run by hand, and about the pace at which a
		// child's lesson calls the tools at its busiest.
		defaults: Options{Scenario: Lesson, Tasks: 5, Pace: 5 * time.Second, Timeout: time.Minute},
		run:      runLesson,
		expected: lessonExpected,
	},
}

// Names are the names of the scenarios there are, in order.
func Names() []string { return slices.Sorted(maps.Keys(scenarios)) }

// Defaults are the options of the scenario named, before any is changed.
func Defaults(name string) (Options, error) {
	known, found := scenarios[name]
	if !found {
		return Options{}, fmt.Errorf("scenario: no scenario %q; there are %s", name, strings.Join(Names(), ", "))
	}
	return known.defaults, nil
}

// Check refuses options no run could keep to.
func (o *Options) Check() error {
	switch {
	case o.Tasks < 1 || o.Tasks > len(lesson.Written()):
		return fmt.Errorf("scenario: a lesson asks for 1 to %d tasks, not %d", len(lesson.Written()), o.Tasks)
	case o.Pace < 0:
		return fmt.Errorf("scenario: a pause of %v is no pause", o.Pace)
	case o.Timeout <= 0:
		return fmt.Errorf("scenario: a call may take no time at all: %v", o.Timeout)
	}
	return nil
}

// Steps is how many steps a run of the options takes at their pace: a
// reckoning of how long it will last, for whoever is waiting for it.
func (o *Options) Steps() int {
	// Opening the profile, three steps for each task, and the progress.
	return 1 + 3*o.Tasks + 1
}

// Run runs the scenario of the options against the service at the target. A
// run asked to stop before its end is the part that ran, marked as stopped,
// and is held to no expectation: it did not get to meet them. The error is
// for a run that could not start at all.
func Run(ctx context.Context, o Options, target session.Target) (report.Run, error) {
	known, found := scenarios[o.Scenario]
	if !found {
		return report.Run{}, fmt.Errorf("scenario: no scenario %q", o.Scenario)
	}
	if err := o.Check(); err != nil {
		return report.Run{}, err
	}
	service := session.Open(target, o.Timeout)
	defer service.Close()

	run := known.run(ctx, o, service)
	run.Stopped = ctx.Err() != nil
	if !run.Stopped {
		run.Broken = known.expected(&run, &o)
	}
	return run, nil
}

// student is the child of a lesson, as the card shows them.
var student = lesson.Student{Pseudonym: "Otter", Grade: 2}

// runLesson walks one child through a lesson: the profile, then each task
// asked for, handed in and answered — right and wrong in turn — and the
// progress at the end, with a pause before every step. The child is new to the
// service, so that a lesson finds no profile and no request left by another.
func runLesson(ctx context.Context, o Options, service *session.Service) report.Run {
	run := report.Run{Scenario: Lesson, Unit: "accepted task", Began: time.Now()}
	child := service.Child("lesson-" + strconv.FormatInt(run.Began.UnixNano(), 36))
	defer func() { _ = child.Close() }()
	w := &walk{child: child, pace: o.Pace, run: &run}

	if w.step(ctx) {
		w.record(lesson.Start(ctx, child, student)...)
	}
	tasks := lesson.Written()[:o.Tasks]
	for i := range tasks {
		if !w.step(ctx) {
			break
		}
		letter := tasks[i].Body.Correct
		if i%2 == 1 {
			letter = tasks[i].Wrong
		}
		w.task(ctx, &tasks[i], letter)
	}
	if w.step(ctx) {
		w.record(lesson.Progress(ctx, child))
	}

	run.Ended = time.Now()
	run.Requests = service.Requests()
	run.Reconnects = child.Reconnects()
	return run
}

// walk is one child's way through a lesson: the pause before every step, and
// the run that keeps every call of it.
type walk struct {
	child *session.Child
	pace  time.Duration
	run   *report.Run
}

// step waits the pause before a step, and says whether the run may take it.
func (w *walk) step(ctx context.Context) bool { return pause(ctx, w.pace) }

// record keeps the answers of a step among the calls of the run.
func (w *walk) record(answers ...session.Answer) {
	for i := range answers {
		w.run.Calls = append(w.run.Calls, report.CallOf(&answers[i], answers[i].Started))
	}
}

// task walks the child through one task: asked for, handed in and, once the
// service accepted it, answered with the letter given. The task counts as
// soon as the service accepted it, whether or not the run goes on to the
// answer.
func (w *walk) task(ctx context.Context, written *lesson.Task, letter string) {
	request, asked, err := lesson.Ask(ctx, w.child, written.Choice)
	w.record(asked)
	if err != nil || !w.step(ctx) {
		return
	}
	card, handedIn := lesson.HandIn(ctx, w.child, request, written, student)
	w.record(handedIn)
	if handedIn.Kind != session.Answered {
		return
	}
	w.run.Units++
	if w.step(ctx) {
		w.record(lesson.AnswerTask(ctx, w.child, card, letter))
	}
}

// lessonExpected is what a lesson that ran to its end did not come to that it
// should have: every call answered as asked, and every task it asked for
// accepted.
func lessonExpected(run *report.Run, o *Options) []string {
	var broken []string
	if run.Units < o.Tasks {
		broken = append(broken, fmt.Sprintf("%d of %d tasks accepted", run.Units, o.Tasks))
	}
	var other []string
	for _, kind := range report.Kinds(run.Calls) {
		if kind != session.Answered {
			other = append(other, kind.String())
		}
	}
	if len(other) > 0 {
		broken = append(broken, "calls of a lesson answered otherwise than as asked: "+strings.Join(other, ", "))
	}
	return broken
}

// pause waits the pause given, and says whether the run may go on: a run
// asked to stop stops in the middle of a pause rather than at its end.
func pause(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
