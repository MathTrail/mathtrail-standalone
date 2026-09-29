package starlark

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.starlark.net/starlark"
)

// What a helper charges the budget is the price applied to what it does: a
// step for every element of the sequences it is handed, and for the list of
// tuples it returns a step for every sixteen bytes with a quarter on top — the
// list itself, and for every tuple its place in the list, its header and its
// values. A helper that built more than it paid for would keep what the budget
// promised nobody would, and one that paid for more than it built would refuse
// a search the budget could afford.
func TestEveryHelperPaysForWhatItBuilds(t *testing.T) {
	t.Parallel()

	const roomy = 1 << 40
	properties := gopter.NewProperties(nil)

	for _, helper := range []*starlark.Builtin{
		permutations(roomy), combinations(roomy), combinationsWithReplacement(roomy),
	} {
		properties.Property(helper.Name()+" pays for its tuples", prop.ForAll(
			func(size, width int) bool {
				pool := rangeOf(size)
				tuples, steps := builtBy(t, helper, starlark.Tuple{pool, starlark.MakeInt(width)}, nil, width)
				return steps == uint64(size)+priceOfTuples(tuples, width)
			},
			gen.IntRange(0, 7), gen.IntRange(0, 8),
		))
	}

	properties.Property("product pays for its tuples and for every sequence it opens", prop.ForAll(
		func(first, second, repeat int) bool {
			tuples, steps := builtBy(t, product(roomy), starlark.Tuple{rangeOf(first), rangeOf(second)},
				[]starlark.Tuple{{starlark.String("repeat"), starlark.MakeInt(repeat)}}, 2*repeat)
			return steps == priceOf(2*64)+uint64(first+second)+priceOfTuples(tuples, 2*repeat)
		},
		gen.IntRange(0, 4), gen.IntRange(0, 5), gen.IntRange(0, 3),
	))

	properties.TestingRun(t)
}

// priceOf is the price of what a call builds, written out: a step for every
// sixteen bytes and a quarter on top, rounded up.
func priceOf(bytes int) uint64 {
	return uint64((5*bytes + 63) / 64)
}

// priceOfTuples is the price of a list of that many tuples of that width: the
// list's own thirty-two bytes, and for every tuple sixteen for its place,
// twenty-four for its header and sixteen for each of its values.
func priceOfTuples(count, width int) uint64 {
	return priceOf(32 + count*(16+24+16*width))
}

// rangeOf is a list of the whole numbers below size.
func rangeOf(size int) *starlark.List {
	elements := make([]starlark.Value, size)
	for i := range elements {
		elements[i] = starlark.MakeInt(i)
	}
	return starlark.NewList(elements)
}

// builtBy calls a helper on a thread of its own, holds every tuple it returns
// to the width given, and says how many tuples it returned and what the call
// charged the budget.
func builtBy(
	t *testing.T, helper *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple, width int,
) (tuples int, steps uint64) {
	t.Helper()

	thread := &starlark.Thread{}
	result, err := helper.CallInternal(thread, args, kwargs)
	if err != nil {
		t.Fatalf("%s: got error %v, want none", helper.Name(), err)
	}
	list, isList := result.(*starlark.List)
	if !isList {
		t.Fatalf("%s: got a %s, want a list", helper.Name(), result.Type())
	}
	for i := range list.Len() {
		tuple, isTuple := list.Index(i).(starlark.Tuple)
		if !isTuple {
			t.Fatalf("%s: element %d is a %s, want a tuple", helper.Name(), i, list.Index(i).Type())
		}
		if tuple.Len() != width {
			t.Fatalf("%s: tuple %d holds %d values, want %d", helper.Name(), i, tuple.Len(), width)
		}
	}
	return list.Len(), thread.Steps
}
