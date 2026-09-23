package starlark

import (
	"fmt"
	"slices"
	"strconv"

	"go.starlark.net/starlark"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// argumentsOf is the options as a solver sees them: a dictionary from letter
// to the text of that option.
//
// They arrive as an argument rather than as a global because the program is
// run twice, on two different dictionaries, and a global would let the second
// run see what the first left behind. It is frozen: a solver reads the options
// it was given and does not get to edit them.
func argumentsOf(options solver.Options) (*starlark.Dict, error) {
	dict := starlark.NewDict(solver.Count)
	for place, text := range options {
		if err := dict.SetKey(starlark.String(solver.Letter(place)), starlark.String(text)); err != nil {
			return nil, fmt.Errorf("starlark: build the options: %w", err)
		}
	}
	dict.Freeze()
	return dict, nil
}

// solveOf is the entry point out of what the program defined, or a sentence
// about why there is none to call.
func solveOf(globals starlark.StringDict) (solve *starlark.Function, problem string) {
	defined, found := globals[entryPoint]
	if !found {
		return nil, "the program defines no " + entryPoint
	}
	solve, isFunction := defined.(*starlark.Function)
	if !isFunction {
		return nil, fmt.Sprintf("%s is a %s rather than a function", entryPoint, defined.Type())
	}
	if solve.NumParams() != 1 || solve.NumKwonlyParams() != 0 || solve.HasVarargs() || solve.HasKwargs() {
		return nil, fmt.Sprintf("%s must take exactly one argument, the options", entryPoint)
	}
	return solve, ""
}

// lettersOf reads what the program returned, or says what is wrong with it.
// The letters come back sorted, so that two runs of one program can be
// compared without the order of a return statement mattering.
func lettersOf(returned starlark.Value) (letters []string, problem string) {
	var items []starlark.Value
	switch sequence := returned.(type) {
	case *starlark.List:
		items = make([]starlark.Value, 0, sequence.Len())
		for i := range sequence.Len() {
			items = append(items, sequence.Index(i))
		}
	case starlark.Tuple:
		items = sequence
	default:
		return nil, fmt.Sprintf("%s returned a %s rather than a list of letters", entryPoint, returned.Type())
	}

	if len(items) > solver.Count {
		return nil, fmt.Sprintf("%s returned %d letters and a task has %d options", entryPoint, len(items), solver.Count)
	}

	letters = make([]string, 0, len(items))
	for _, item := range items {
		letter, isString := item.(starlark.String)
		if !isString || solver.Place(string(letter)) < 0 {
			return nil, fmt.Sprintf("%s is not one of the option letters", textOf(item))
		}
		if slices.Contains(letters, string(letter)) {
			return nil, fmt.Sprintf("%s returned %s twice", entryPoint, string(letter))
		}
		letters = append(letters, string(letter))
	}
	slices.Sort(letters)
	return letters, ""
}

// textOf is how a value is written when it is compared with the text of an
// option. A float is written out in full: the shortest form the language
// itself would use turns a million into 1e+06, and an option says 1000000.
func textOf(value starlark.Value) string {
	switch typed := value.(type) {
	case starlark.String:
		return string(typed)
	case starlark.Float:
		return strconv.FormatFloat(float64(typed), 'f', -1, 64)
	default:
		return value.String()
	}
}
