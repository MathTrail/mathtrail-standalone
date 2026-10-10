package scenario

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// student is the child of a lesson, as the card shows them.
var student = lesson.Student{Pseudonym: "Otter", Grade: 2}

// runLesson walks one child through a lesson, and holds it to what a lesson
// should come to.
func runLesson(ctx context.Context, o *Options, launch Launch) ([]report.Run, error) {
	return served(ctx, launch, Lesson, func(up *Launched) []report.Run {
		service := session.Open(up.Target, o.Timeout)
		defer service.Close()

		run := walkLesson(ctx, o, service, "lesson-"+stamp())
		run.Requests = service.Requests()
		run.Broken = lessonExpected(&run, o.Tasks)
		return []report.Run{run}
	})
}

// lessonLasts is about how long a lesson of the options takes.
func lessonLasts(o *Options) time.Duration { return time.Duration(lessonSteps(o)) * o.Pace }

// lessonSteps is how many steps a lesson of the options takes: opening the
// profile, three steps for each task, and the progress.
func lessonSteps(o *Options) int { return 1 + 3*o.Tasks + 1 }

// walkLesson walks one child, signed in by the name given, through a lesson.
// The child is new to the service, so that a lesson finds no profile and no
// request left by another.
func walkLesson(ctx context.Context, o *Options, service *session.Service, name string) report.Run {
	return walkLessonAs(ctx, o, service.Child(name))
}

// walkLessonAs walks the child given through a lesson, and closes the child
// after: the profile, then each task asked for, handed in, answered — right
// and wrong in turn — and its result shown, and the progress at the end, with
// a pause before every step. What the service was sent in all is the caller's
// to count, since the service may be shared.
func walkLessonAs(ctx context.Context, o *Options, child *session.Child) report.Run {
	run := report.Run{Scenario: Lesson, Unit: "accepted task", Began: time.Now()}
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

	finish(ctx, &run)
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

// task walks the child through one task: asked for, its package fetched, the
// card it will come to told it is being written, handed in, the card told it
// is on it and, once the service accepted it, answered with the letter given,
// and the card of how the answer went asked for.
// The card asks once on either side of the hand-in, as few times as says
// whether it turned into the task. The task counts as soon as the service
// accepted it, whether or not the run goes on to the answer.
func (w *walk) task(ctx context.Context, written *lesson.Task, letter string) {
	request, asked, err := lesson.Ask(ctx, w.child, written.Choice)
	w.record(asked...)
	if err != nil {
		return
	}
	w.record(lesson.AwaitWriting(ctx, w.child, request))
	if !w.step(ctx) {
		return
	}
	card, handedIn := lesson.HandIn(ctx, w.child, request, written, student)
	w.record(handedIn)
	if handedIn.Kind != session.Answered {
		return
	}
	w.run.Units++
	w.record(lesson.AwaitCard(ctx, w.child, request, card))
	if !w.step(ctx) {
		return
	}
	answered := lesson.AnswerTask(ctx, w.child, card, letter)
	w.record(answered)
	if answered.Kind == session.Answered && w.step(ctx) {
		w.record(lesson.ShowResult(ctx, w.child, card))
	}
}

// lessonExpected is what lessons that ran to their end did not come to that
// they should have: every call answered as asked, and the tasks they asked
// for accepted, as many as given.
func lessonExpected(run *report.Run, tasks int) []string {
	var broken []string
	if run.Units < tasks {
		broken = append(broken, fmt.Sprintf("%d of %d tasks accepted", run.Units, tasks))
	}
	if other := kindsBut(run, session.Answered); other != "" {
		broken = append(broken, "calls of a lesson answered otherwise than as asked: "+other)
	}
	return broken
}

// kindsBut are the kinds of answer among the calls of the run other than the
// ones given, as a report writes them, or nothing when there is none.
func kindsBut(run *report.Run, allowed ...session.Kind) string {
	var other []string
	for _, kind := range report.Kinds(run.Calls) {
		if !slices.Contains(allowed, kind) {
			other = append(other, kind.String())
		}
	}
	return strings.Join(other, ", ")
}

// stamp is a name no run before this one has used: a child's name is the
// account it signs in, and a child of an earlier run would find the profile
// that run left.
func stamp() string { return strconv.FormatInt(time.Now().UnixNano(), 36) }

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
