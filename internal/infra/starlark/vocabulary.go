package starlark

import (
	"context"
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
// built-ins above with a cap on them, the counting and calendar helpers a
// brute force is mostly made of, and the one that turns a computed answer back
// into the letters of the task.
//
// They are few on purpose. Every one of them is code to write, test, document
// and explain to a model that has one shot at using it correctly, and a
// vocabulary nobody can hold in mind is one that gets used wrong.
func vocabulary(steps uint64) (starlark.StringDict, error) {
	declared := starlark.StringDict{
		"permutations":                  permutations(steps),
		"combinations":                  combinations(steps),
		"combinations_with_replacement": combinationsWithReplacement(steps),
		"product":                       product(steps),
		"sum":                           sum(steps),
		"prod":                          prod(steps),
		"gcd":                           starlark.NewBuiltin("gcd", gcd),
		"is_leap":                       starlark.NewBuiltin("is_leap", isLeap),
		"days_in_month":                 starlark.NewBuiltin("days_in_month", daysInMonth),
		"weekday":                       starlark.NewBuiltin("weekday", weekday),
		"add_days":                      starlark.NewBuiltin("add_days", addDays),
		"days_between":                  starlark.NewBuiltin("days_between", daysBetween),
		"match":                         match(steps),
	}
	for _, name := range walkers {
		capped, err := bounded(name, steps)
		if err != nil {
			return nil, err
		}
		declared[name] = capped
	}
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
			if err := weigh(thread, steps, name, argument); err != nil {
				return nil, err
			}
		}
		// A keyword argument is a pair of name and value, and the value is
		// weighed like any other: some of these built-ins take their sequence
		// by name as readily as by position, and a cap that looked only at the
		// position would be one anybody could walk around by typing the name.
		for _, pair := range kwargs {
			if err := weigh(thread, steps, name, pair[1]); err != nil {
				return nil, err
			}
		}
		return inner.CallInternal(thread, args, kwargs)
	}), nil
}

// weigh puts one argument of a built-in on the budget before the built-in
// walks it.
func weigh(thread *starlark.Thread, steps uint64, name string, value starlark.Value) error {
	length, err := lengthOf(value)
	if err == nil {
		err = charge(thread, steps, length, 1)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// lengthOf is how many elements a value holds, for the values that can say.
//
// Something that is not a sequence at all — a number, a function — has no
// length and costs nothing to hand over. Something that can be walked but will
// not say how far is another matter: taking it for free would be one call that
// goes past both limits at once, and this language has such a value. The
// characters of a string as codepoints are iterable and have no length, so a
// list of them would be built out of a string of any size at no charge.
func lengthOf(value starlark.Value) (int, error) {
	if length := starlark.Len(value); length >= 0 {
		return length, nil
	}
	if _, walkable := value.(starlark.Iterable); walkable {
		return 0, fmt.Errorf(
			"%s will not say how long it is, and what cannot be measured cannot be bounded;"+
				" the characters of a string are .elems()", value.Type())
	}
	return 0, nil
}

// runContext is where a run keeps its own context on the thread it runs on.
const runContext = "run context"

// clockOf is the context of the run a thread belongs to, which says when the
// run has been told to stop. The interpreter asks only between its own
// instructions, so a helper that works for long inside one of them asks it as
// it goes, and is no harder to stop than the loop it stands for. A thread no
// run belongs to is never stopped.
func clockOf(thread *starlark.Thread) context.Context {
	if ctx, isContext := thread.Local(runContext).(context.Context); isContext {
		return ctx
	}
	return context.Background()
}

// errTooManySteps stops a built-in whose work the budget cannot pay for. The
// model never reads it: what tripped is worked out from the budget itself, and
// the sentence the model gets is written there.
var errTooManySteps = errors.New("too many steps")

// charge answers the whole question about what a built-in is about to walk or
// build: whether it may at all, and what it costs the step budget.
//
// Both halves live here rather than at the call sites. They are one question,
// and a caller that remembered the price but forgot the ceiling would leave a
// hole nobody reading it would see. The charge comes before the work, because
// the interpreter would notice the overrun only once the memory had been
// taken. A value with no length of its own — a number, a function — costs
// nothing and passes.
//
// The two numbers do different jobs. The count is how many elements the list
// ends up holding, and that is what the ceiling is about: a million of
// anything is already more than a task at this level could need. The width is
// how many values sit inside each element, and it is what the budget is
// charged for, because a hundred thousand tuples of twenty is two million
// values allocated and the length of the list says nothing about it. A flat
// sequence has a width of one. Both ceilings hold even when nothing will be
// built, because a caller may lay out room for a tuple before it learns that
// there is none to fill.
func charge(thread *starlark.Thread, steps uint64, count, width int) error {
	if count > maxElements {
		return fmt.Errorf("%d elements, and %d is the most that will be walked at once", count, maxElements)
	}
	if width > maxElements {
		return fmt.Errorf("a tuple of %d values, and %d is the most that will be walked at once", width, maxElements)
	}
	if count <= 0 || width <= 0 {
		return nil
	}
	thread.Steps += uint64(count) * uint64(width)
	if thread.Steps >= steps {
		return errTooManySteps
	}
	return nil
}

// elementsOf reads a sequence a helper was handed, charging the budget for its
// length before any of it is held.
//
// A string is refused by name. Strings are not sequences in this language and
// nothing else here iterates one either, but a model porting Python will reach
// for one — so the refusal says what to write instead rather than what went
// wrong.
func elementsOf(thread *starlark.Thread, steps uint64, name string, value starlark.Value) ([]starlark.Value, error) {
	if _, isString := value.(starlark.String); isString {
		return nil, fmt.Errorf("%s: a string is not a sequence here; write out its characters as a list", name)
	}
	walkable, isWalkable := value.(starlark.Iterable)
	if !isWalkable {
		return nil, fmt.Errorf("%s: %s is not a sequence", name, value.Type())
	}
	length, err := lengthOf(value)
	if err == nil {
		err = charge(thread, steps, length, 1)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}

	elements := make([]starlark.Value, 0, length)
	iterator := walkable.Iterate()
	defer iterator.Done()
	var element starlark.Value
	for iterator.Next(&element) {
		elements = append(elements, element)
	}
	return elements, nil
}

// match returns the letters whose option text matches a computed value, and it
// is the intended last line of a solver: work the answer out, then let the
// options say which letter that is. A solver that writes the letter itself
// passes the first run of a program and fails the second.
//
// Matching folds every text it reads, which is work in proportion to the
// text's length inside one call, and a solver builds a text of any length in
// one step. So each text is paid for a step a byte before any of it is folded,
// under the same ceiling as a sequence: a text that could stand on a card is
// nowhere near it.
func match(steps uint64) *starlark.Builtin {
	return starlark.NewBuiltin("match", func(
		thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		var given, value starlark.Value
		if err := starlark.UnpackPositionalArgs("match", args, kwargs, 2, &given, &value); err != nil {
			return nil, err
		}
		options, err := optionsIn(given)
		if err != nil {
			return nil, err
		}
		computed := textOf(value)
		for _, text := range append([]string{computed}, options[:]...) {
			// Refused in its own words before it is charged: a text is not a
			// sequence to walk, and a model told it was would look for one.
			if len(text) > maxElements {
				return nil, fmt.Errorf("match: a text of %d bytes, and %d is the longest that will be compared", len(text), maxElements)
			}
			if err := charge(thread, steps, len(text), 1); err != nil {
				return nil, fmt.Errorf("match: %w", err)
			}
		}

		matched := options.Match(computed)
		letters := make([]starlark.Value, len(matched))
		for i, letter := range matched {
			letters[i] = starlark.String(letter)
		}
		return starlark.NewList(letters), nil
	})
}

// optionsIn reads the options back out of the dictionary a solver handed over,
// which has to be the one it was given: every letter, each with its text.
func optionsIn(given starlark.Value) (solver.Options, error) {
	var options solver.Options
	dict, isDict := given.(*starlark.Dict)
	if !isDict {
		return options, fmt.Errorf("match takes the options it was given, not a %s", given.Type())
	}
	for place := range solver.Count {
		letter := solver.Letter(place)
		text, found, err := dict.Get(starlark.String(letter))
		if err != nil {
			return options, fmt.Errorf("match: read option %s: %w", letter, err)
		}
		if !found {
			return options, fmt.Errorf("match takes the options it was given, and %s is not among these", letter)
		}
		options[place] = textOf(text)
	}
	return options, nil
}
