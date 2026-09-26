package solver

import (
	"context"
	"fmt"
	"strings"
)

// Runner executes one program once, against one set of options.
//
// It is declared here, by the side that needs it: the rule below is about what
// two runs mean together, and it has no business knowing what an interpreter
// is. The error is for the runner's own failures — it could not start, or
// whoever asked stopped waiting — while everything the program itself did
// comes back as a status.
type Runner interface {
	Run(ctx context.Context, source string, options Options) (Result, error)
}

// Agreement is what the runs of one program came to.
type Agreement struct {
	// Letter is the option both runs pointed at, empty when they pointed at
	// different things or when one of them never got there.
	Letter string
	// Runs are the runs that happened: both of them, or only the first when it
	// left the second nothing to agree with.
	Runs []Result
}

// Agreed reports whether the two runs pointed at one and the same option.
func (a Agreement) Agreed() bool { return a.Letter != "" }

// Failed is the run that did not reach an answer, or nil when both did.
func (a Agreement) Failed() *Result {
	for i, run := range a.Runs {
		if run.Status != StatusOK {
			return &a.Runs[i]
		}
	}
	return nil
}

// Verdict runs one program twice and reports what the two runs agreed on.
//
// The second run sees the same five texts under shifted labels. A program that
// works the answer out and looks it up among the options follows the text and
// returns the new label; a program that writes the letter out by hand returns
// the old one and is caught. This does not make cheating impossible — a
// program that hard-codes the value rather than the letter still passes — and
// it is not meant to. It removes the laziest failure, which is also the most
// likely one.
//
// The second run happens only when the first reached exactly one option: there
// is otherwise nothing for it to agree with, and a program that spends the
// whole clock would spend it twice.
func Verdict(ctx context.Context, runner Runner, source string, options Options) (Agreement, error) {
	first, err := runner.Run(ctx, source, options)
	if err != nil {
		return Agreement{}, fmt.Errorf("solver: run as handed in: %w", err)
	}
	if !first.One() {
		return Agreement{Runs: []Result{first}}, nil
	}

	second, err := runner.Run(ctx, source, options.Rotated())
	if err != nil {
		return Agreement{}, fmt.Errorf("solver: run with the labels shifted: %w", err)
	}

	agreement := Agreement{Runs: []Result{first, second}}
	if second.One() && second.Letters[0] == Relabelled(first.Letters[0]) {
		agreement.Letter = first.Letters[0]
	}
	return agreement, nil
}

// Explain is one sentence about what stands between these runs and a verdict,
// empty when they agreed. It names the letters the runs arrived at, which is
// what a person reading a run needs and what the model is never shown: a
// sentence that named them would name the answer.
func (a Agreement) Explain() string {
	if a.Agreed() {
		return ""
	}
	if len(a.Runs) == 0 {
		return "the solver did not run"
	}
	if failed := a.Failed(); failed != nil {
		return failed.Message
	}
	if len(a.Runs) == 1 {
		return found(a.Runs[0])
	}
	return fmt.Sprintf(
		"the solver pointed at %s, and %s once the options were relabelled: it is reading the letters rather than working the answer out",
		a.Runs[0].Letters[0], pointed(a.Runs[1]))
}

// found says what one run came to when it did not come to a single option.
func found(run Result) string {
	if len(run.Letters) == 0 {
		return "the solver found no option that fits what it computed"
	}
	return "the solver found more than one option that fits: " + strings.Join(run.Letters, ", ")
}

// pointed is the same for the second run, in a form that continues a sentence.
func pointed(run Result) string {
	switch len(run.Letters) {
	case 0:
		return "at nothing"
	case 1:
		return "at " + run.Letters[0]
	default:
		return "at " + strings.Join(run.Letters, " and ")
	}
}
