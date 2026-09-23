package starlark

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// maxElements is the longest sequence anything declared here will build or
// walk. A brute force that needs more than a million of anything is the wrong
// shape for a task a child is meant to reason about.
const maxElements = 1_000_000

// walkers are the built-ins of the language that can be handed a whole
// sequence at once, and therefore the ones that can spend an instance without
// spending a single step of the budget.
//
// A range is the way in: it is lazy and costs nothing to make, so range of a
// billion is free, iterating it costs a step an element and stops at the
// budget — and handing it to any of these turns it into a billion values
// inside one call that the interpreter is not watching.
var walkers = []string{
	"all", "any", "dict", "enumerate", "list",
	"max", "min", "reversed", "set", "sorted", "tuple", "zip",
}

// vocabulary is what a solver can see on top of the language itself: the
// built-ins above with a cap on them, and the one helper that turns a computed
// answer back into the letters of the task.
func vocabulary(steps uint64) (starlark.StringDict, error) {
	declared := make(starlark.StringDict, len(walkers)+1)
	for _, name := range walkers {
		capped, err := bounded(name, steps)
		if err != nil {
			return nil, err
		}
		declared[name] = capped
	}
	declared["match"] = starlark.NewBuiltin("match", match)
	return declared, nil
}

// bounded wraps one built-in so that it refuses a sequence past the cap and
// charges the step budget for the elements it is about to walk.
//
// The interpreter counts its own instructions and nothing else, and it looks
// for a cancellation only between them, so neither limit can reach inside a
// single call. The length of a sequence is known before the walk begins, which
// is what makes the refusal possible before anything is allocated rather than
// after.
func bounded(name string, steps uint64) (*starlark.Builtin, error) {
	inner, isBuiltin := starlark.Universe[name].(*starlark.Builtin)
	if !isBuiltin {
		return nil, fmt.Errorf("starlark: the language has no built-in %s to bound", name)
	}
	return starlark.NewBuiltin(name, func(
		thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		for _, argument := range args {
			length := starlark.Len(argument)
			if length < 0 {
				continue // a number, a function, anything with no length to bound
			}
			if length > maxElements {
				return nil, fmt.Errorf("%s of %d elements, and %d is the most that will be walked at once",
					name, length, maxElements)
			}
			if err := charge(thread, steps, length); err != nil {
				return nil, err
			}
		}
		return inner.CallInternal(thread, args, kwargs)
	}), nil
}

// errTooManySteps stops a built-in whose work the budget cannot pay for. The
// model never reads it: what tripped is worked out from the budget itself, and
// the sentence the model gets is written there.
var errTooManySteps = errors.New("too many steps")

// charge puts the elements a built-in is about to walk onto the step budget.
// The interpreter would notice the overrun between its own instructions, but
// only once the work was done and the memory for it taken, so the charge comes
// first and the work does not happen at all.
func charge(thread *starlark.Thread, steps uint64, elements int) error {
	if elements <= 0 {
		return nil
	}
	thread.Steps += uint64(elements)
	if thread.Steps >= steps {
		return errTooManySteps
	}
	return nil
}

// match returns the letters whose option text matches a computed value, and it
// is the intended last line of a solver: work the answer out, then let the
// options say which letter that is. A solver that writes the letter itself
// passes the first run of a program and fails the second.
//
// It is the one thing here that walks nothing: five options and five
// comparisons cost the budget less than the call that reaches them.
func match(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var given, value starlark.Value
	if err := starlark.UnpackPositionalArgs("match", args, kwargs, 2, &given, &value); err != nil {
		return nil, err
	}
	dict, isDict := given.(*starlark.Dict)
	if !isDict {
		return nil, fmt.Errorf("match takes the options it was given, not a %s", given.Type())
	}

	var options solver.Options
	for place := range solver.Count {
		letter := solver.Letter(place)
		text, found, err := dict.Get(starlark.String(letter))
		if err != nil {
			return nil, fmt.Errorf("match: read option %s: %w", letter, err)
		}
		if !found {
			return nil, fmt.Errorf("match takes the options it was given, and %s is not among these", letter)
		}
		options[place] = textOf(text)
	}

	matched := options.Match(textOf(value))
	letters := make([]starlark.Value, len(matched))
	for i, letter := range matched {
		letters[i] = starlark.String(letter)
	}
	return starlark.NewList(letters), nil
}
