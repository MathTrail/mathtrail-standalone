package starlark

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.starlark.net/starlark"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// failed turns a refusal of the interpreter into something the model can act
// on.
//
// Which limit tripped is read from the budgets and not from the text of the
// error. Both limits arrive the same way — the interpreter reaches its step
// count and cancels the thread exactly as the clock does — and the sentence it
// builds out of that belongs to the library, which may reword it; a counter
// and a deadline may not be reworded.
//
// The two contexts are told apart on purpose. The caller's own time running
// out is not a verdict about the solver, and a run that stopped because the
// caller's deadline was the shorter one must not be reported against a limit
// of ours that never fired.
func (s *sandbox) failed(
	caller, own context.Context, thread *starlark.Thread, err error, started time.Time,
) (solver.Result, error) {
	switch {
	case caller.Err() != nil:
		// Whoever asked for the run has gone, cancelled or out of time of
		// their own: the solver did nothing wrong, and there is nobody left to
		// tell about it either way.
		return solver.Result{}, fmt.Errorf("starlark: run stopped: %w", caller.Err())

	case errors.Is(own.Err(), context.DeadlineExceeded):
		return spent(solver.StatusTimeout, nil,
			fmt.Sprintf("the solver ran longer than %s", s.limits.Timeout), thread, started), nil

	case thread.ExecutionSteps() >= s.limits.Steps:
		return spent(solver.StatusTimeout, nil,
			fmt.Sprintf("the solver took more than %d steps", s.limits.Steps), thread, started), nil
	}

	// What is left is the program's own failure: a fail(), an index out of
	// range, a type that does not fit, a cap of the vocabulary, a recursive
	// call. The message is the model's own; the backtrace beneath it is the
	// interpreter's, and it stays here.
	var failure *starlark.EvalError
	if errors.As(err, &failure) {
		return spent(solver.StatusError, nil, failure.Msg, thread, started), nil
	}
	return spent(solver.StatusError, nil, readable(err), thread, started), nil
}

// readable is what the model is told about a source that will not run: the
// first complaint, without the name this service gave the file. The rest of
// the list follows from the first, and a name the model never chose only makes
// the sentence longer.
func readable(err error) string {
	first, _, _ := strings.Cut(err.Error(), "\n")
	return strings.TrimSpace(strings.TrimPrefix(first, programName+":"))
}
