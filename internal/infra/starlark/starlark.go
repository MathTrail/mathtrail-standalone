// Package starlark runs the program the chat's model writes to prove a task's
// answer, and stops it when it asks for too much.
//
// It runs inside this process. Starlark is a language built to be embedded and
// cut down — no imports, no file access, no clock, no threads, no randomness —
// so the sandbox is mostly the language itself, and what is left to decide is
// how much of Python to allow back, what a brute force needs to be able to
// say, and what one run may spend.
//
// Two limits do the stopping: a budget of interpreter steps and a wall clock.
// Neither is a limit on memory, which the interpreter does not account for at
// all, so everything predeclared here refuses a sequence past a cap of its own
// and charges the step budget for what it walks. What that still leaves open is
// the language itself: its operators and the methods of its values — a list
// times a number, extend, union, += with a range — build a sequence of any
// length in a single step, past both limits, and only a limit on the memory of
// the whole process could stop them.
package starlark

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

const (
	// MaxSource is the longest solver accepted, in bytes. Past this a program
	// is a transcription of the answer rather than a search for it.
	MaxSource = 8 * 1024

	// entryPoint is the one function a solver must define.
	entryPoint = "solve"

	// programName is what the interpreter calls the source in its own
	// complaints, and what is taken back out of them before the model reads
	// one: the model never chose the name.
	programName = "solver.star"
)

// Limits are what one run may spend. Both runs of one program get their own
// of each: a budget shared between them would make the second run's verdict
// depend on how expensive the first was.
type Limits struct {
	// Steps is how many instructions the interpreter may execute. It counts
	// its own instructions and nothing else, which is why the vocabulary
	// charges it for the elements it walks.
	Steps uint64
	// Timeout is the wall clock of one run, from the first instruction.
	Timeout time.Duration
	// Concurrency is how many runs may be in flight at once. A solver holds a
	// core for as long as its clock allows, and an instance has few.
	Concurrency int
}

// sandbox is a place to run solvers: the limits, the slots and the vocabulary
// that every run in it shares.
type sandbox struct {
	limits      Limits
	slots       chan struct{}
	predeclared starlark.StringDict
}

// New builds a sandbox, or refuses limits nothing could run under.
func New(limits Limits) (solver.Runner, error) {
	switch {
	case limits.Steps == 0:
		return nil, errors.New("starlark: steps must be at least one")
	case limits.Timeout <= 0:
		return nil, errors.New("starlark: timeout must be a positive duration")
	case limits.Concurrency < 1:
		return nil, errors.New("starlark: concurrency must be at least one")
	}
	declared, err := vocabulary(limits.Steps)
	if err != nil {
		return nil, err
	}
	return &sandbox{
		limits:      limits,
		slots:       make(chan struct{}, limits.Concurrency),
		predeclared: declared,
	}, nil
}

// Run executes one program once and reports what became of it.
//
// The error is for the run not happening: no slot came free in the time
// allowed, or whoever asked for it stopped waiting. Everything the program
// itself did comes back as a status, because that is a fact about the solver
// rather than about this service.
func (s *sandbox) Run(ctx context.Context, source string, options solver.Options) (solver.Result, error) {
	// Whoever asked has gone: nothing here is worth starting, and the slot
	// below is worth leaving to a caller who is still waiting for an answer.
	if err := ctx.Err(); err != nil {
		return solver.Result{}, fmt.Errorf("starlark: %w", err)
	}

	// Before waiting for a slot, which a program too long to read never needs.
	if refusal := oversized(source); refusal != "" {
		return refused(refusal), nil
	}

	select {
	case s.slots <- struct{}{}:
	case <-ctx.Done():
		return solver.Result{}, fmt.Errorf("starlark: wait for a free slot: %w", ctx.Err())
	}
	defer func() { <-s.slots }()

	return s.run(ctx, source, options)
}

// run is one program from source to letters, once a slot is held.
func (s *sandbox) run(ctx context.Context, source string, options solver.Options) (solver.Result, error) {
	file, program, refusal := compile(programName, source, s.predeclared.Has)
	if refusal != "" {
		return refused(refusal), nil
	}

	arguments, err := argumentsOf(options)
	if err != nil {
		return solver.Result{}, err
	}

	runCtx, cancel := context.WithTimeout(ctx, s.limits.Timeout)
	defer cancel()

	thread := &starlark.Thread{
		Name: entryPoint,
		Print: func(*starlark.Thread, string) {
			// Deliberately nothing. What a solver prints is read by nobody:
			// the result is the answer, and the output of a program written
			// elsewhere has no place in this service's own stream.
		},
	}
	thread.SetMaxExecutionSteps(s.limits.Steps)
	// Cancel is the one method safe to call from another goroutine, and the
	// interpreter looks for it between instructions; a helper that works for
	// long inside one of them reads the run's own context instead.
	stop := context.AfterFunc(runCtx, func() { thread.Cancel("the time limit") })
	defer stop()
	thread.SetLocal(runContext, runCtx)

	started := time.Now()

	globals, err := program.Init(thread, s.predeclared)
	if err != nil {
		return s.failed(ctx, runCtx, thread, file, err, started)
	}

	solve, problem := solveOf(globals)
	if problem != "" {
		return spent(solver.StatusNoEntryPoint, nil, problem, thread, started), nil
	}

	returned, err := starlark.Call(thread, solve, starlark.Tuple{arguments}, nil)
	if err != nil {
		return s.failed(ctx, runCtx, thread, file, err, started)
	}

	letters, problem := lettersOf(returned)
	if problem != "" {
		return spent(solver.StatusBadOutput, nil, problem, thread, started), nil
	}
	return spent(solver.StatusOK, letters, "", thread, started), nil
}

// compile reads a program the way every run reads it: no longer than a solver
// may be, parsed in the dialect, its names bound against the vocabulary, and
// no module asked for. What stops it is said in the sentence a refusal
// carries, or nothing when it reads.
func compile(name, source string, isPredeclared func(string) bool) (*syntax.File, *starlark.Program, string) {
	if refusal := oversized(source); refusal != "" {
		return nil, nil, refusal
	}
	file, err := parse(name, source)
	if err != nil {
		return nil, nil, readable(errors.Unwrap(err))
	}
	program, err := starlark.FileProgram(file, isPredeclared)
	if err != nil {
		return nil, nil, readable(err)
	}
	// There is no module loader to bind anything from, so a program that asks
	// for one is asking for a language this is not.
	if program.NumLoads() > 0 {
		return nil, nil, "load is not available: a solver is one file that stands on its own"
	}
	return file, program, ""
}

// oversized says why a source is too long to read, or nothing when it is not.
// It is measured before the parser, which would hold the whole of it in memory
// first.
func oversized(source string) string {
	if len(source) > MaxSource {
		return fmt.Sprintf("the solver is longer than %d KB", MaxSource/1024)
	}
	return ""
}

// refused is the result of a program that never began to run: its source is
// refused, and the message says why.
func refused(message string) solver.Result {
	return solver.Result{Status: solver.StatusBadSource, Message: message}
}

// spent is a result of a program that did, with what it cost.
func spent(status solver.Status, letters []string, message string, thread *starlark.Thread, started time.Time) solver.Result {
	return solver.Result{
		Status:   status,
		Letters:  letters,
		Message:  message,
		Steps:    thread.ExecutionSteps(),
		Duration: time.Since(started),
	}
}
