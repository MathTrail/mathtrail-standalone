package starlark

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.starlark.net/starlark"
)

// TestTheBuiltinsPayForTheMemoryTheyTake holds the price to what calls really
// take, measured rather than worked out. Whatever a call keeps is no more than
// sixteen bytes for every step it was charged, and whatever it allocates on
// the way, kept or let go, no more than twice that. The first is the bound the
// budget promises on the memory of a run; the second keeps what the price
// leaves to the collector within reach of the price.
//
// Not parallel: what it reads is the memory of the whole process, and a test
// running beside it would be counted as the call's. The parallel tests of a
// package start only once every test that is not parallel has finished.
func TestTheBuiltinsPayForTheMemoryTheyTake(t *testing.T) {
	declared, err := vocabulary(1 << 40)
	if err != nil {
		t.Fatalf("vocabulary: %v", err)
	}

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 20
	properties := gopter.NewProperties(parameters)
	for _, c := range pricedCalls(t) {
		builtin, isBuiltin := declared[c.builtin].(*starlark.Builtin)
		if !isBuiltin {
			t.Fatalf("%s: nothing declared by that name", c.builtin)
		}
		properties.Property(c.name, prop.ForAll(
			func(size, width int) *gopter.PropResult {
				args, kwargs := c.arguments(size, width)
				took := measure(t, builtin, args, kwargs)
				return gopter.NewPropResult(
					took.kept <= int64(bytesPerStep*took.steps)+keptSlack &&
						took.allocated <= int64(2*bytesPerStep*took.steps)+allocatedSlack,
					took.String())
			},
			gen.IntRange(c.sizes[0], c.sizes[1]), gen.IntRange(c.widths[0], c.widths[1]),
		))
	}
	properties.TestingRun(t)
}

// What a measurement allows for besides the price: the few words the runtime
// itself may keep or take while a call is measured.
const (
	keptSlack      = 256
	allocatedSlack = 512
)

// pricedCall is a call of one built-in, of a shape its arguments are made to,
// in sizes from the first of a range to the last.
type pricedCall struct {
	name, builtin string
	sizes, widths [2]int
	arguments     func(size, width int) (starlark.Tuple, []starlark.Tuple)
}

// pricedCalls are the shapes of call that build the most for their price, or
// that each rule of the price is there for: tables of keys hashing alike,
// numbers too large for 32 bits, tuples of every width, and sequences opened in
// their thousands with nothing in them.
func pricedCalls(t *testing.T) []pricedCall {
	t.Helper()
	upTo := [2]int{0, 3000}
	none := [2]int{0, 0}
	only := func(args ...starlark.Value) (starlark.Tuple, []starlark.Tuple) { return args, nil }
	return []pricedCall{
		{"a list of a range", "list", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, size, 1))
		}},
		{"a list of numbers past 32 bits", "list", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 1<<40, 1<<40+size, 1))
		}},
		{"a list of characters", "list", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(characters(t, size))
		}},
		{"a tuple of a list", "tuple", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(rangeOf(size))
		}},
		{"a range reversed", "reversed", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, size, 1))
		}},
		{"a list sorted by a key", "sorted", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return starlark.Tuple{rangeOf(size)}, []starlark.Tuple{{starlark.String("key"), starlark.Universe["abs"]}}
		}},
		{"a set of a range", "set", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, size, 1))
		}},
		{"a set of keys that hash alike", "set", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, 2048*size, 2048))
		}},
		{"a dictionary of pairs", "dict", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(pairsOf(size))
		}},
		{"a dictionary of a dictionary", "dict", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(dictionaryOf(t, size))
		}},
		{"a dictionary of keywords", "dict", [2]int{0, 300}, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return nil, keywords(size)
		}},
		{"a range numbered", "enumerate", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, size, 1))
		}},
		{"a range numbered from past 32 bits", "enumerate", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, size, 1), starlark.MakeInt64(1<<40))
		}},
		{"two ranges zipped", "zip", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, size, 1), numbers(t, 0, size+7, 1))
		}},
		{"empty lists zipped", "zip", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(emptyLists(size)...)
		}},
		{"the least of a range", "min", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, size+1, 1))
		}},
		{"whether any of a list", "any", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(rangeOf(size))
		}},
		{"a sum of a range", "sum", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 0, size, 1))
		}},
		{"a product of the numbers to twelve", "prod", [2]int{1, 13}, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(numbers(t, 1, size, 1))
		}},
		{"orderings of a pool", "permutations", [2]int{0, 8}, [2]int{0, 8}, func(size, width int) (starlark.Tuple, []starlark.Tuple) {
			return only(rangeOf(size), starlark.MakeInt(min(width, size+1)))
		}},
		{"choices from a pool", "combinations", [2]int{0, 16}, [2]int{0, 16}, func(size, width int) (starlark.Tuple, []starlark.Tuple) {
			return only(rangeOf(size), starlark.MakeInt(width))
		}},
		{"choices from a pool drawn again", "combinations_with_replacement", [2]int{0, 9}, [2]int{0, 9},
			func(size, width int) (starlark.Tuple, []starlark.Tuple) {
				return only(rangeOf(size), starlark.MakeInt(width))
			}},
		{"a product of two, repeated", "product", [2]int{0, 10}, [2]int{0, 3}, func(size, repeat int) (starlark.Tuple, []starlark.Tuple) {
			return starlark.Tuple{rangeOf(size), rangeOf(size/2 + 1)},
				[]starlark.Tuple{{starlark.String("repeat"), starlark.MakeInt(repeat)}}
		}},
		{"a product of one, repeated wide", "product", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return starlark.Tuple{rangeOf(1)}, []starlark.Tuple{{starlark.String("repeat"), starlark.MakeInt(size)}}
		}},
		{"a product of empty lists", "product", upTo, none, func(size, _ int) (starlark.Tuple, []starlark.Tuple) {
			return only(emptyLists(size)...)
		}},
	}
}

// took is what one call took of memory, and what it was charged.
type took struct {
	kept, allocated int64
	steps           uint64
}

func (t took) String() string {
	return fmt.Sprintf("kept %d bytes and allocated %d for %d steps, which allow %d and %d",
		t.kept, t.allocated, t.steps, bytesPerStep*t.steps+keptSlack, 2*bytesPerStep*t.steps+allocatedSlack)
}

// measure calls a built-in on a thread of its own and says what the call took:
// the bytes the heap holds once it has returned and the collector has run, the
// bytes it allocated on the way, and the steps it was charged. A call that
// seems to take more than it may is measured twice more and the least of each
// is kept, since what runs beside it can only add to what it seems to take.
func measure(t *testing.T, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) took {
	t.Helper()
	least := took{kept: 1 << 62, allocated: 1 << 62}
	for range 3 {
		var before, after runtime.MemStats
		thread := &starlark.Thread{}
		runtime.GC()
		runtime.ReadMemStats(&before)
		result, err := builtin.CallInternal(thread, args, kwargs)
		runtime.GC()
		runtime.ReadMemStats(&after)
		runtime.KeepAlive(result)
		if err != nil {
			t.Fatalf("%s: got error %v, want none", builtin.Name(), err)
		}
		least.kept = min(least.kept, int64(after.HeapAlloc)-int64(before.HeapAlloc))
		least.allocated = min(least.allocated, int64(after.TotalAlloc-before.TotalAlloc))
		least.steps = thread.Steps
		if least.kept <= int64(bytesPerStep*least.steps)+keptSlack &&
			least.allocated <= int64(2*bytesPerStep*least.steps)+allocatedSlack {
			break
		}
	}
	return least
}

// numbers is range(from, to, step) as the language makes it.
func numbers(t *testing.T, from, to, step int) starlark.Value {
	t.Helper()
	made, err := starlark.Call(&starlark.Thread{}, starlark.Universe["range"],
		starlark.Tuple{starlark.MakeInt(from), starlark.MakeInt(to), starlark.MakeInt(step)}, nil)
	if err != nil {
		t.Fatalf("range(%d, %d, %d): %v", from, to, step, err)
	}
	return made
}

// characters is the characters of a text of that many bytes, as the language
// walks them one by one.
func characters(t *testing.T, size int) starlark.Value {
	t.Helper()
	elems, err := starlark.String(strings.Repeat("a", size)).Attr("elems")
	if err != nil {
		t.Fatalf("elems: %v", err)
	}
	walked, err := starlark.Call(&starlark.Thread{}, elems, nil, nil)
	if err != nil {
		t.Fatalf("elems(): %v", err)
	}
	return walked
}

// pairsOf is a list of that many pairs of a number and itself.
func pairsOf(size int) *starlark.List {
	pairs := make([]starlark.Value, size)
	for i := range pairs {
		pairs[i] = starlark.Tuple{starlark.MakeInt(i), starlark.MakeInt(i)}
	}
	return starlark.NewList(pairs)
}

// dictionaryOf is a dictionary of that many numbers, each to itself.
func dictionaryOf(t *testing.T, size int) *starlark.Dict {
	t.Helper()
	dictionary := starlark.NewDict(size)
	for i := range size {
		if err := dictionary.SetKey(starlark.MakeInt(i), starlark.MakeInt(i)); err != nil {
			t.Fatalf("SetKey: %v", err)
		}
	}
	return dictionary
}

// keywords are that many arguments by name, each a number.
func keywords(size int) []starlark.Tuple {
	pairs := make([]starlark.Tuple, size)
	for i := range pairs {
		pairs[i] = starlark.Tuple{starlark.String(fmt.Sprintf("k%d", i)), starlark.MakeInt(i)}
	}
	return pairs
}

// emptyLists are that many lists with nothing in them.
func emptyLists(size int) []starlark.Value {
	lists := make([]starlark.Value, size)
	for i := range lists {
		lists[i] = starlark.NewList(nil)
	}
	return lists
}

// TestEveryWalkerPaysItsPrice pins what each built-in of the language pays for
// a call of a known size, worked out by hand from the price: a step for every
// element walked, and for what is built a step for every sixteen bytes with a
// quarter on top. The sums are written out so that a change to any size the
// price is made of shows here as a number, next to the call it moved.
func TestEveryWalkerPaysItsPrice(t *testing.T) {
	t.Parallel()
	declared, err := vocabulary(1 << 40)
	if err != nil {
		t.Fatalf("vocabulary: %v", err)
	}

	thousand := numbers(t, 0, 1000, 1)
	for _, test := range []struct {
		name    string
		builtin string
		args    starlark.Tuple
		kwargs  []starlark.Tuple
		steps   uint64
	}{
		// 1,000 walked; 32 + 16·1,000 bytes built, 5/64 of which is 1,252.5.
		{"a list of a thousand", "list", starlark.Tuple{thousand}, nil, 1000 + 1253},
		{"a tuple of a thousand", "tuple", starlark.Tuple{thousand}, nil, 1000 + 1253},
		{"a thousand sorted", "sorted", starlark.Tuple{thousand}, nil, 1000 + 1253},
		{"a thousand sorted, handed over by name", "sorted", nil,
			[]starlark.Tuple{{starlark.String("iterable"), thousand}}, 1000 + 1253},
		{"a thousand reversed", "reversed", starlark.Tuple{thousand}, nil, 1000 + 1253},
		// 512 + 256·1,000 bytes: 20,040 steps.
		{"a set of a thousand", "set", starlark.Tuple{thousand}, nil, 1000 + 20040},
		{"an empty set", "set", nil, nil, 40},
		// No sequence walked, and a table of two: 512 + 512 bytes.
		{"a dictionary of two keywords", "dict", nil, []starlark.Tuple{
			{starlark.String("a"), starlark.MakeInt(1)}, {starlark.String("b"), starlark.MakeInt(2)},
		}, 80},
		// 32 + 72·1,000 bytes: 5,627.5 steps.
		{"a thousand numbered", "enumerate", starlark.Tuple{thousand}, nil, 1000 + 5628},
		// Ten pairs, 752 bytes, and ten numbers past 32 bits, 480: 1,232.
		{"ten numbered from the last number of 32 bits", "enumerate",
			starlark.Tuple{numbers(t, 0, 10, 1), starlark.MakeInt(1<<31 - 1)}, nil, 10 + 97},
		// Both walked; 500 pairs of 72 bytes, the list's 32 and 64 for each of
		// the two sequences opened: 36,160 bytes.
		{"a thousand zipped with five hundred", "zip",
			starlark.Tuple{thousand, numbers(t, 0, 500, 1)}, nil, 1500 + 2825},
		{"the least of a thousand", "min", starlark.Tuple{thousand}, nil, 1000},
		{"whether all of a thousand", "all", starlark.Tuple{thousand}, nil, 1000},
		// The walk makes ten numbers past 32 bits, 480 bytes: 10 + 37.5. The
		// list of them is 192 bytes more: 15.
		{"a list of ten numbers past 32 bits", "list",
			starlark.Tuple{numbers(t, 1<<31-1, 1<<31+9, 1)}, nil, 10 + 38 + 15},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			builtin, isBuiltin := declared[test.builtin].(*starlark.Builtin)
			if !isBuiltin {
				t.Fatalf("%s: nothing declared by that name", test.builtin)
			}
			thread := &starlark.Thread{}
			if _, err := builtin.CallInternal(thread, test.args, test.kwargs); err != nil {
				t.Fatalf("%s: got error %v, want none", test.builtin, err)
			}
			if thread.Steps != test.steps {
				t.Errorf("%s: charged %d steps, want %d", test.builtin, thread.Steps, test.steps)
			}
		})
	}
}

// TestTheNumbersPastThirtyTwoBitsArePaidFor walks the edge the price of a
// number turns on: a whole number the interpreter keeps inside the value costs
// nothing to make, and one a bit past it is a value of its own, 48 bytes, for
// every element that reaches it. A price that moved the edge by one would
// either charge every walk for numbers it never makes or let a walk of large
// numbers through at the price of small ones.
func TestTheNumbersPastThirtyTwoBitsArePaidFor(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		value starlark.Value
		small bool
	}{
		{"the largest of 32 bits", starlark.MakeInt(math.MaxInt32), true},
		{"the smallest of 32 bits", starlark.MakeInt(math.MinInt32), true},
		{"one past the largest", starlark.MakeInt64(math.MaxInt32 + 1), false},
		{"one past the smallest", starlark.MakeInt64(math.MinInt32 - 1), false},
		{"a word, which is no number", starlark.String("12"), false},
	} {
		if got := small(test.value); got != test.small {
			t.Errorf("small(%s) = %v, want %v", test.name, got, test.small)
		}
	}

	for _, test := range []struct {
		name  string
		start starlark.Value
		n     int
		built uint64
	}{
		{"nothing numbered", starlark.MakeInt(math.MaxInt32), 0, 0},
		{"ten numbered up to the last of 32 bits", starlark.MakeInt(math.MaxInt32 - 9), 10, 0},
		{"ten numbered one past it", starlark.MakeInt(math.MaxInt32 - 8), 10, 10 * numberBytes},
		{"ten numbered from a word", starlark.String("a"), 10, 10 * numberBytes},
	} {
		if got := numbering(test.start, test.n); got != test.built {
			t.Errorf("numbering(%s) = %d bytes, want %d", test.name, got, test.built)
		}
	}

	for _, test := range []struct {
		name     string
		from, to int
		built    uint64
	}{
		{"a range that ends on the last of 32 bits", math.MaxInt32 - 9, math.MaxInt32 + 1, 0},
		{"a range that ends one past it", math.MaxInt32 - 8, math.MaxInt32 + 2, 10 * numberBytes},
		{"a range that starts one below the first", math.MinInt32 - 1, math.MinInt32 + 9, 10 * numberBytes},
	} {
		walked := numbers(t, test.from, test.to, 1)
		if got := walking(walked, starlark.Len(walked)); got.built != test.built || got.walked != 10 {
			t.Errorf("walking(%s) = %+v, want 10 walked and %d bytes built", test.name, got, test.built)
		}
	}
}
