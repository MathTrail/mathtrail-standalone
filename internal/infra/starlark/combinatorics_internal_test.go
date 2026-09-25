package starlark

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.starlark.net/starlark"
)

// What a helper charges the budget is what it builds: every element of the
// sequences it is handed, and every value of every tuple it returns. A helper
// that built more than it paid for would spend what the budget promised to
// keep, and one that paid for more than it built would refuse a search the
// budget could afford.
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
				built, steps := builtBy(t, helper, starlark.Tuple{pool, starlark.MakeInt(width)}, nil)
				return steps == uint64(size+built)
			},
			gen.IntRange(0, 7), gen.IntRange(0, 8),
		))
	}

	properties.Property("product pays for its tuples", prop.ForAll(
		func(first, second, repeat int) bool {
			built, steps := builtBy(t, product(roomy), starlark.Tuple{rangeOf(first), rangeOf(second)},
				[]starlark.Tuple{{starlark.String("repeat"), starlark.MakeInt(repeat)}})
			return steps == uint64(first+second+built)
		},
		gen.IntRange(0, 4), gen.IntRange(0, 5), gen.IntRange(0, 3),
	))

	properties.TestingRun(t)
}

// rangeOf is a list of the whole numbers below size.
func rangeOf(size int) *starlark.List {
	elements := make([]starlark.Value, size)
	for i := range elements {
		elements[i] = starlark.MakeInt(i)
	}
	return starlark.NewList(elements)
}

// builtBy calls a helper on a thread of its own, and says how many values the
// tuples it returned hold between them and what the call charged the budget.
func builtBy(t *testing.T, helper *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (values int, steps uint64) {
	t.Helper()

	thread := &starlark.Thread{}
	result, err := helper.CallInternal(thread, args, kwargs)
	if err != nil {
		t.Fatalf("%s: got error %v, want none", helper.Name(), err)
	}
	tuples, isList := result.(*starlark.List)
	if !isList {
		t.Fatalf("%s: got a %s, want a list", helper.Name(), result.Type())
	}
	for i := range tuples.Len() {
		tuple, isTuple := tuples.Index(i).(starlark.Tuple)
		if !isTuple {
			t.Fatalf("%s: element %d is a %s, want a tuple", helper.Name(), i, tuples.Index(i).Type())
		}
		values += tuple.Len()
	}
	return values, thread.Steps
}
