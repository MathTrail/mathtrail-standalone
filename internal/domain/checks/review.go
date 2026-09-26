package checks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// Content is what a review needs of what the service ships: the catalogs a
// task is checked against, and the reference tasks it must not copy.
type Content interface {
	Catalog
	TrapDescriber
	// ReferenceQuestions are the questions of the reference tasks of this
	// level.
	ReferenceQuestions(level rating.GradeLevel) []string
}

// Submission is a task as it is handed in: its three parts as the JSON they
// arrived as, and the source of the program that is to prove its answer.
type Submission struct {
	Brief     json.RawMessage
	Task      json.RawMessage
	SelfCheck json.RawMessage
	Solver    string
}

// Against is what a task is judged against beside itself, all of it kept in
// the profile: the request the task answers and the tasks the child has had.
// The child's grade is not among it: a task is held to the level its brief
// asked for, whoever it is for.
type Against struct {
	// Asked is the brief the open request recorded, and a task is not judged
	// without it: nothing else shows that it is the task that was asked for,
	// and its level is what the wording is read against.
	Asked *profile.Brief
	// Language is what the task is written in, as the request recorded it.
	Language string
	// Fingerprints are the sketches of the tasks the child has had.
	Fingerprints []string
}

// Examined is a submission read and its solver run: the half of a review that
// takes time and needs nothing from the profile.
type Examined struct {
	// finished says Examine got to the end: any other value of this type has
	// nothing in it to judge.
	finished bool
	draft    Draft
	// read is what reading the submission refused.
	read []Problem
	// answer is what the runs of the solver came to, and nil when there were
	// not five options to run it on.
	answer *solver.Agreement
}

// Reviewer decides whether a task can be given to a child: accepted, or every
// reason it cannot be, in the order the model is told them.
//
// A review comes in two halves because the profile is read before it and
// written after it, and nothing slow may happen in between: what is written
// may have been overtaken by another tab in the meantime, and the shorter the
// window, the rarer that is. Examine is the slow half — the solver's runs — and
// needs no profile. Judge is every check, arithmetic only, against what the
// profile holds.
type Reviewer interface {
	// Examine reads a submission and runs its solver. The error is the
	// sandbox's own failure rather than the task's, and it costs the model no
	// attempt.
	Examine(ctx context.Context, submission *Submission) (Examined, error)
	// Judge runs every check, in the order their refusals are reported. The
	// error says there was nothing to judge — a submission Examine did not
	// finish, or no brief of an open request to hold it against — which, like
	// the error of Examine, is the service's fault rather than the task's.
	Judge(examined Examined, against Against) (Outcome, error)
}

// reviewer holds what every review needs beside the submission.
type reviewer struct {
	content Content
	runner  solver.Runner
	limits  DrawingLimits
}

// NewReviewer builds a reviewer over what the service ships, the sandbox a
// solver runs in and the limits a drawing is held to.
func NewReviewer(content Content, runner solver.Runner, limits DrawingLimits) Reviewer {
	return &reviewer{content: content, runner: runner, limits: limits}
}

func (r *reviewer) Examine(ctx context.Context, submission *Submission) (Examined, error) {
	draft, read := Decode(submission.Brief, submission.Task, submission.SelfCheck)
	examined := Examined{finished: true, draft: draft, read: read}

	options, complete := optionsOf(draft.Task)
	if !complete {
		return examined, nil
	}
	agreement, err := solver.Verdict(ctx, r.runner, submission.Solver, options)
	if err != nil {
		return Examined{}, fmt.Errorf("checks: run the solver: %w", err)
	}
	examined.answer = &agreement
	return examined, nil
}

// Judge runs every check a task has the input for, in the order the model is
// told what failed. A check whose input is missing does not run, and is noted
// instead of passed: the model learns it has still to pass it, and a refusal
// about a fault that is not there would send it looking in the wrong place.
func (r *reviewer) Judge(examined Examined, against Against) (Outcome, error) {
	switch {
	case !examined.finished:
		return Outcome{}, errors.New("checks: judge: the submission was not examined")
	case against.Asked == nil:
		return Outcome{}, errors.New("checks: judge: no brief of an open request to hold the task against")
	}
	draft, task := examined.draft, examined.draft.Task

	var found findings
	found.add(slices.Concat(examined.read, Structure(draft, against.Asked, r.content)), "")
	found.add(r.explanations(draft, against))
	found.add(r.drawingFormat(task), "")
	found.add(r.drawingMatch(task))
	found.add(r.readability(task, against))
	found.add(r.solverRuns(examined.answer))
	found.add(r.answers(examined.answer, draft))
	found.add(r.blockingIssues(draft.SelfCheck))
	found.add(r.nearDuplicates(task, against))

	outcome := Outcome{
		Draft:       draft,
		Problems:    found.problems,
		Unchecked:   found.unchecked,
		MinorIssues: minorIssues(draft.SelfCheck),
		judged:      true,
	}
	if examined.answer != nil {
		outcome.runs = examined.answer.Runs
	}
	return outcome, nil
}

// findings are what the checks of one review found, in the order they ran.
type findings struct {
	problems  []Problem
	unchecked []string
}

// add records what one check found, or why it could not look.
func (f *findings) add(problems []Problem, unchecked string) {
	f.problems = append(f.problems, problems...)
	if unchecked != "" {
		f.unchecked = append(f.unchecked, unchecked)
	}
}

// explanations checks the explanations behind the wrong options.
func (r *reviewer) explanations(draft Draft, against Against) (problems []Problem, unchecked string) {
	if draft.Task == nil {
		return nil, "the explanations behind the wrong options were not checked: that needs a task that can be read"
	}
	return Explanations(draft, against.Language, r.content), ""
}

// drawingFormat checks the format of the drawing, when the task has one: a
// drawing is not required, and a task without one has nothing to check here.
func (r *reviewer) drawingFormat(task *Task) []Problem {
	if task == nil {
		return nil
	}
	return DrawingFormat(task.Drawing, r.limits)
}

// drawingMatch checks the drawing against its structure and the wording.
func (r *reviewer) drawingMatch(task *Task) (problems []Problem, unchecked string) {
	if task == nil || strings.TrimSpace(task.Drawing) == "" {
		return nil, ""
	}
	if task.DrawingStructure == nil {
		return nil, "the drawing was not checked against the wording: that needs task.drawing_structure"
	}
	return DrawingMatch(task.Question, task.Drawing, task.DrawingStructure), ""
}

// readability checks that the question reads as a task of the level asked for.
func (r *reviewer) readability(task *Task, against Against) (problems []Problem, unchecked string) {
	if task == nil || solver.Blank(task.Question) {
		return nil, "readability was not checked: that needs task.question"
	}
	return Readability(task.Question, against.Language, against.Asked.GradeLevel), ""
}

// nearDuplicates checks that the question copies no reference task of the
// level asked for and repeats none of the child's own.
func (r *reviewer) nearDuplicates(task *Task, against Against) (problems []Problem, unchecked string) {
	if task == nil || solver.Blank(task.Question) {
		return nil, "the question was not compared with earlier tasks: that needs task.question"
	}
	references := r.content.ReferenceQuestions(against.Asked.GradeLevel)
	return NearDuplicate(task.Question, against.Language, references, against.Fingerprints), ""
}

// optionsOf are a task's five options as a solver takes them, and whether all
// five are there to take: a solver run on fewer would be judged on options no
// child is ever shown.
func optionsOf(task *Task) (solver.Options, bool) {
	var options solver.Options
	if task == nil {
		return options, false
	}
	for place := range solver.Count {
		text := task.Options[solver.Letter(place)]
		if solver.Blank(text) {
			return options, false
		}
		options[place] = text
	}
	return options, true
}
