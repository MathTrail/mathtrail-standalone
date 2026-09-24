// Package solver is the contract between the service and the program the
// chat's model writes to prove a task's answer.
//
// One check of a task is not an opinion. The model hands in a short program
// that finds the answer by brute force, the service runs it, and what it
// arrives at is held against what the model claimed. Two witnesses to the same
// number — the reasoning that wrote the task and a program that enumerates it
// — are the strongest guarantee the service has.
//
// This package says what such a program must return, how a run of it is
// reported, and how two runs of it become one verdict. It runs nothing: the
// interpreter is infrastructure, and keeping the contract away from it is what
// lets every rule here be read and tested on its own.
package solver

import "time"

// Status is what became of one run.
type Status string

const (
	// StatusOK means the program ran to the end and returned letters. It is
	// the only status a verdict can be built from.
	StatusOK Status = "ok"
	// StatusBadSource means the source does not parse, or asks for something
	// the dialect does not have.
	StatusBadSource Status = "bad_source"
	// StatusBadFormat means the program stopped at a string it formats with %
	// in a way Starlark cannot fill: a stray %, a width or a precision, or a
	// number of values the string does not ask for. A string like that which
	// the run never reaches stops nothing, and is no reason to refuse.
	StatusBadFormat Status = "bad_format"
	// StatusNoEntryPoint means there is no solve to call, or what is called
	// solve cannot be called with the options.
	StatusNoEntryPoint Status = "no_entry_point"
	// StatusError means the program failed while running.
	StatusError Status = "error"
	// StatusTimeout means a limit tripped: the steps or the clock.
	StatusTimeout Status = "timeout"
	// StatusBadOutput means the program returned something that is not a list
	// of distinct option letters.
	StatusBadOutput Status = "bad_output"
)

// Result is one run of one program.
type Result struct {
	// Status is what became of the run.
	Status Status
	// Letters are the options the program said the conditions allow, sorted
	// and without repeats. An empty list with a status of ok is an answer
	// rather than a failure: it says no option fits, which is worth knowing.
	Letters []string
	// Message is one sentence about what went wrong, empty when the run
	// succeeded: which limit tripped, or the first thing the program or the
	// interpreter said, without the trace beneath it. What the program said
	// can quote an option, so a sentence for the model is built from the
	// status rather than from this.
	Message string
	// Steps is what the run cost the step budget.
	Steps uint64
	// Duration is how long it took.
	Duration time.Duration
}

// One reports whether the run pointed at exactly one option, which is what a
// verdict needs of it.
func (r Result) One() bool { return r.Status == StatusOK && len(r.Letters) == 1 }
