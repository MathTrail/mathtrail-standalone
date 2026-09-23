package checks

import (
	"slices"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// Outcome is what a review came to.
type Outcome struct {
	// Draft is the submission as it was read.
	Draft Draft
	// Problems are every check that failed, in the order the model is told
	// them.
	Problems []Problem
	// Unchecked are the checks that could not run for want of their input,
	// each a sentence saying what it needs.
	Unchecked []string
	// MinorIssues are the types of the issues the self-check found and let
	// through, one for each issue: two issues of one type count as two.
	MinorIssues []string
	// runs are the runs of the solver, kept for what they cost.
	runs []solver.Result
}

// Accepted says whether the task passed every check there is. A check goes
// unrun only for want of an input the format requires, so a task with a check
// left unrun always has a problem of its format too.
func (o *Outcome) Accepted() bool { return len(o.Problems) == 0 }

// Primary is the code of the first check that failed: the one an attempt is
// counted by and the log is read by. It is empty when nothing failed.
func (o *Outcome) Primary() Code {
	if len(o.Problems) == 0 {
		return ""
	}
	return o.Problems[0].Code
}

// Reason is one check that failed, and everything it found.
type Reason struct {
	Code     Code
	Messages []string
}

// Reasons are the problems grouped by the check that found them: every code
// once, in the order it first appears, with everything found under it.
func (o *Outcome) Reasons() []Reason {
	var reasons []Reason
	for _, problem := range o.Problems {
		at := slices.IndexFunc(reasons, func(reason Reason) bool { return reason.Code == problem.Code })
		if at < 0 {
			at = len(reasons)
			reasons = append(reasons, Reason{Code: problem.Code})
		}
		reasons[at].Messages = append(reasons[at].Messages, problem.Message)
	}
	return reasons
}

// The two ways a review ends, as the log names them.
const (
	outcomeAccepted = "accepted"
	outcomeRejected = "rejected"
)

// Submitted is what the log keeps of one review: codes, types and what the
// solver cost, and not a word of the task. The service keeps no text of a
// task anywhere, and a log line is the easiest place to forget that, so the
// line is built from these fields and nothing else.
type Submitted struct {
	// Outcome is "accepted" or "rejected".
	Outcome string
	// Primary is the code the attempt is counted by, empty when accepted.
	Primary Code
	// Failed is the code of every check that failed, each once, in order.
	Failed []Code
	// MinorIssues are the types of the issues the self-check let through.
	MinorIssues []string
	// SolverSteps is what the runs of the solver cost the step budget,
	// together.
	SolverSteps uint64
	// SolverTime is how long they took, together.
	SolverTime time.Duration
}

// Event is what the log keeps of this review.
func (o *Outcome) Event() Submitted {
	event := Submitted{Outcome: outcomeAccepted, Primary: o.Primary(), MinorIssues: slices.Clone(o.MinorIssues)}
	if !o.Accepted() {
		event.Outcome = outcomeRejected
	}
	for _, reason := range o.Reasons() {
		event.Failed = append(event.Failed, reason.Code)
	}
	for _, run := range o.runs {
		event.SolverSteps += run.Steps
		event.SolverTime += run.Duration
	}
	return event
}
