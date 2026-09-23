package starlark

import (
	"fmt"
	"math/big"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// sum adds up a sequence of numbers. The language has no sum of its own, and a
// brute force that counts anything needs one.
func sum(steps uint64) *starlark.Builtin {
	return starlark.NewBuiltin("sum", func(
		thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		return fold(thread, steps, builtin, args, kwargs, syntax.PLUS, starlark.MakeInt(0))
	})
}

// prod multiplies a sequence of numbers together.
func prod(steps uint64) *starlark.Builtin {
	return starlark.NewBuiltin("prod", func(
		thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		return fold(thread, steps, builtin, args, kwargs, syntax.STAR, starlark.MakeInt(1))
	})
}

// fold is what sum and prod both are: a walk over the numbers of a sequence,
// from a starting value the caller may choose.
//
// The arithmetic is the language's own, so whole numbers stay exact however
// large they grow and a float anywhere makes the answer a float — which is
// what a solver would get writing the same loop out by hand.
func fold(
	thread *starlark.Thread, steps uint64, builtin *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple, operation syntax.Token, from starlark.Value,
) (starlark.Value, error) {
	name := builtin.Name()
	sequence := starlark.Value(nil)
	start := from
	if err := starlark.UnpackArgs(name, args, kwargs, "seq", &sequence, "start?", &start); err != nil {
		return nil, err
	}
	elements, err := elementsOf(thread, steps, name, sequence)
	if err != nil {
		return nil, err
	}
	return accumulate(name, operation, start, elements)
}

// accumulate is the walk itself, over numbers and nothing else. The language
// would happily join two strings with the same operator, and a sum that
// returned a sentence would be a puzzle rather than an answer.
func accumulate(
	name string, operation syntax.Token, start starlark.Value, elements []starlark.Value,
) (starlark.Value, error) {
	total := start
	if err := mustBeNumber(name, total); err != nil {
		return nil, err
	}
	for _, element := range elements {
		if err := mustBeNumber(name, element); err != nil {
			return nil, err
		}
		added, err := starlark.Binary(operation, total, element)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		total = added
	}
	return total, nil
}

// gcd is the greatest common divisor of two whole numbers, and it is the
// building block of exact fractions: a solver comparing a over b with c over d
// multiplies through instead of dividing, and reaches for this when the
// numbers have to be brought back down.
//
// It is the divisor of the numbers themselves, sign ignored, and of nothing
// and nothing it is nothing — the same answers Python gives.
func gcd(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (
	starlark.Value, error,
) {
	name := builtin.Name()
	var first, second starlark.Value
	if err := starlark.UnpackPositionalArgs(name, args, kwargs, 2, &first, &second); err != nil {
		return nil, err
	}
	a, err := wholeNumber(name, first)
	if err != nil {
		return nil, err
	}
	b, err := wholeNumber(name, second)
	if err != nil {
		return nil, err
	}
	return starlark.MakeBigInt(new(big.Int).GCD(nil, nil, a, b)), nil
}

// mustBeNumber refuses what cannot be added up.
func mustBeNumber(name string, value starlark.Value) error {
	switch value.(type) {
	case starlark.Int, starlark.Float:
		return nil
	}
	return fmt.Errorf("%s: %s is not a number", name, value.Type())
}

// wholeNumber is a value as an exact integer, however many digits it has.
func wholeNumber(name string, value starlark.Value) (*big.Int, error) {
	number, isInt := value.(starlark.Int)
	if !isInt {
		return nil, fmt.Errorf("%s: %s is not a whole number", name, value.Type())
	}
	return number.BigInt(), nil
}
