package starlark

import (
	"math"

	"go.starlark.net/starlark"
)

// What a built-in costs the step budget. The interpreter counts its own
// instructions and nothing else, and a single call to a built-in can walk a
// million elements or build a million tuples inside one of them. So
// everything declared here pays before it works: a step for every element it
// walks, and a step for every sixteen bytes of what it builds and hands back,
// worked out from the sizes it was handed. The bytes carry a quarter on top,
// because the allocator rounds every block it hands out up to a size of its
// own, and a price that left the rounding out would let a run keep more than
// it paid for.
//
// That makes the budget a bound on memory as well as on time: whatever the
// built-ins of a run build, it keeps no more of it than sixteen bytes a step.
// What a call takes and lets go again — its iterators, a copy of a pool, the
// tables a set outgrows as it fills — is not priced. It is the collector's,
// and the tests hold it to twice the price.
//
// What the language builds by itself is outside all of this: a literal, an
// operator, a method of a value. The interpreter charges those a step an
// instruction whatever they take.

// bytesPerStep is what a step buys of what a built-in builds.
const bytesPerStep = 16

// The sizes the estimates are made of, as the language lays its values out on a
// machine of 64 bits.
const (
	// valueBytes is one value held in a list or a tuple.
	valueBytes = 16
	// boxBytes is the header of a tuple held as a value.
	boxBytes = 24
	// listBytes is a list itself, before its elements.
	listBytes = 32
	// tableBytes is a set or a dictionary before its first entry, the first
	// bucket of its table included.
	tableBytes = 512
	// entryBytes is what an entry of a set or a dictionary holds at worst: its
	// share of a table that has just doubled, and of the buckets that keys
	// hashing alike spill into.
	entryBytes = 256
	// numberBytes is a whole number too large for 32 bits, which the
	// interpreter keeps outside the value that names it.
	numberBytes = 48
	// openBytes is what a call takes to start walking one of its arguments.
	openBytes = 64
)

// cost is what a call is charged: the elements it walks and the bytes it
// builds.
type cost struct{ walked, built uint64 }

// steps is a cost in steps: one for every element walked, and one for every
// sixteen bytes built with a quarter on top — five for every sixty-four,
// rounded up.
func (c cost) steps() uint64 {
	return c.walked + (5*c.built+4*bytesPerStep-1)/(4*bytesPerStep)
}

// listOf is what a list of n values takes.
func listOf(n int) uint64 {
	return listBytes + valueBytes*clamped(n)
}

// tuplesOf is what a list of count tuples takes, each holding width values.
func tuplesOf(count, width int) uint64 {
	return listBytes + clamped(count)*(valueBytes+boxBytes+valueBytes*clamped(width))
}

// tableOf is what a set or a dictionary of n entries takes.
func tableOf(n int) uint64 {
	return tableBytes + entryBytes*clamped(n)
}

// clamped is a size as the estimates read it: nothing below none, and never
// more than one past the ceiling. A size past the ceiling is refused anyway,
// and holding it there keeps the arithmetic of an estimate from overflowing
// before the refusal is reached.
func clamped(n int) uint64 {
	return uint64(min(max(n, 0), maxElements+1))
}

// walking is what walking a value costs: a step for every element, and the
// whole numbers a range makes as it goes when they are too large for the
// interpreter to keep inside the value — the walk makes them, and whatever it
// walks for may keep them.
func walking(value starlark.Value, length int) cost {
	price := cost{walked: clamped(length)}
	if length > 0 && value.Type() == "range" {
		// The numbers of a range run one way, so when both ends fit in 32 bits
		// every number between them does. Asking for the two ends makes the
		// two numbers, as the walk would; nothing else is made to ask.
		if indexed, isIndexed := value.(starlark.Indexable); isIndexed &&
			(!small(indexed.Index(0)) || !small(indexed.Index(length-1))) {
			price.built = numberBytes * clamped(length)
		}
	}
	return price
}

// numbering is what the numbers of a numbered walk take: nothing while every
// number from the first to the last fits in 32 bits, and a whole number each
// once they do not.
func numbering(start starlark.Value, n int) uint64 {
	if n <= 0 {
		return 0
	}
	if first, isInt := start.(starlark.Int); isInt && small(first) {
		if from, _ := first.Int64(); from+int64(n)-1 <= math.MaxInt32 {
			return 0
		}
	}
	return numberBytes * clamped(n)
}

// small says whether a value is a whole number that fits in 32 bits, which the
// interpreter keeps inside the value itself. It asks the number for its value
// rather than for a conversion, whose refusal of a number that does not fit
// would build a sentence nobody reads, at every walk of a range of large ones.
func small(value starlark.Value) bool {
	number, isInt := value.(starlark.Int)
	if !isInt {
		return false
	}
	n, exact := number.Int64()
	return exact && n >= math.MinInt32 && n <= math.MaxInt32
}

// nothing is what all, any, max and min build: they walk, and hand back
// something they found or a truth.
func nothing(starlark.Tuple, []starlark.Tuple) (count, width int, built uint64) {
	return 0, 0, 0
}

// flat is what list, tuple, reversed and sorted build: a list of the values of
// the sequence they are handed, by position or by name.
func flat(args starlark.Tuple, kwargs []starlark.Tuple) (count, width int, built uint64) {
	n := longest(args, kwargs)
	return n, 1, listOf(n)
}

// table is what set and dict build: a table with an entry for every element of
// the sequence they are handed, and one for every keyword besides.
func table(args starlark.Tuple, kwargs []starlark.Tuple) (count, width int, built uint64) {
	n := len(kwargs)
	if len(args) > 0 {
		n += max(starlark.Len(args[0]), 0)
	}
	return n, 1, tableOf(n)
}

// numbered is what enumerate builds: a pair for every element, the element and
// its number.
func numbered(args starlark.Tuple, _ []starlark.Tuple) (count, width int, built uint64) {
	var n int
	start := starlark.Value(starlark.MakeInt(0))
	if len(args) > 0 {
		n = max(starlark.Len(args[0]), 0)
	}
	if len(args) > 1 {
		start = args[1]
	}
	return n, 2, tuplesOf(n, 2) + numbering(start, n)
}

// rows is what zip builds: a tuple for every row the shortest of its sequences
// has, as wide as the sequences are many — and a walk begun on every one of
// them.
func rows(args starlark.Tuple, _ []starlark.Tuple) (count, width int, built uint64) {
	n := 0
	for i, argument := range args {
		if length := max(starlark.Len(argument), 0); i == 0 || length < n {
			n = length
		}
	}
	return n, len(args), tuplesOf(n, len(args)) + openBytes*clamped(len(args))
}

// longest is the length of the longest sequence among the arguments of a call,
// by position or by name.
func longest(args starlark.Tuple, kwargs []starlark.Tuple) int {
	n := 0
	for _, argument := range args {
		n = max(n, starlark.Len(argument))
	}
	for _, pair := range kwargs {
		n = max(n, starlark.Len(pair[1]))
	}
	return n
}
