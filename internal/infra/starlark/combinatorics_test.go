package starlark_test

import (
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// TestCombinatorics holds the four helpers to the lists Python produces, in
// the order Python produces them. The order is the point: a solver that walks
// these lists and takes the first answer it likes would quietly take a
// different one if they came out shuffled.
func TestCombinatorics(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		expression string
		want       string
	}{
		{`permutations([1, 2, 3])`, `[(1, 2, 3), (1, 3, 2), (2, 1, 3), (2, 3, 1), (3, 1, 2), (3, 2, 1)]`},
		{`permutations([1, 2, 3], 2)`, `[(1, 2), (1, 3), (2, 1), (2, 3), (3, 1), (3, 2)]`},
		{`permutations([1, 2, 3], r=2)`, `[(1, 2), (1, 3), (2, 1), (2, 3), (3, 1), (3, 2)]`},
		{`permutations([1, 2], 0)`, `[()]`},
		{`permutations([1, 2], 3)`, `[]`},
		{`permutations([], 0)`, `[()]`},
		// Elements are taken by their place and not by what they are, so a
		// pool holding one thing twice holds two things.
		{`permutations([1, 1], 2)`, `[(1, 1), (1, 1)]`},
		{`permutations(range(3), 2)`, `[(0, 1), (0, 2), (1, 0), (1, 2), (2, 0), (2, 1)]`},

		{`combinations([1, 2, 3, 4], 2)`, `[(1, 2), (1, 3), (1, 4), (2, 3), (2, 4), (3, 4)]`},
		{`combinations([1, 2, 3], 3)`, `[(1, 2, 3)]`},
		{`combinations([1, 2], 0)`, `[()]`},
		{`combinations([1], 2)`, `[]`},
		{`combinations(["Ann", "Ben", "Kim"], 2)`,
			`[("Ann", "Ben"), ("Ann", "Kim"), ("Ben", "Kim")]`},

		{`combinations_with_replacement([1, 2, 3], 2)`, `[(1, 1), (1, 2), (1, 3), (2, 2), (2, 3), (3, 3)]`},
		{`combinations_with_replacement([1, 2], 3)`, `[(1, 1, 1), (1, 1, 2), (1, 2, 2), (2, 2, 2)]`},
		{`combinations_with_replacement([1, 2], 0)`, `[()]`},
		{`combinations_with_replacement([], 1)`, `[]`},
		// A tuple of nothing is one tuple out of any pool, including none.
		{`combinations_with_replacement([], 0)`, `[()]`},

		{`product([1, 2], ["x", "y"])`, `[(1, "x"), (1, "y"), (2, "x"), (2, "y")]`},
		{`product([1, 2], repeat=2)`, `[(1, 1), (1, 2), (2, 1), (2, 2)]`},
		{`product([1], [2], [3])`, `[(1, 2, 3)]`},
		{`product()`, `[()]`},
		{`product([1, 2], repeat=0)`, `[()]`},
		{`product([], [1])`, `[]`},
		{`product([1, 2], ["x"], repeat=2)`, `[(1, "x", 1, "x"), (1, "x", 2, "x"), (2, "x", 1, "x"), (2, "x", 2, "x")]`},
		{`len(product(*[[0, 1]] * 10))`, `1024`},
		// A keyword has to come before the unpacking, which is the one place
		// this language asks for an order Python does not.
		{`len(product(repeat=2, *[[0, 1], [2, 3]]))`, `16`},
	} {
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()
			if got := repr(t, test.expression); got != test.want {
				t.Errorf("%s\n got %s\nwant %s", test.expression, got, test.want)
			}
		})
	}
}

// TestCombinatoricsRefusals covers what a solver gets wrong: the message has to
// say what to write instead, because the model reading it has one more attempt.
func TestCombinatoricsRefusals(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		expression string
		contains   string
	}{
		{"a string as a sequence", `product("HT", repeat=3)`, "write out its characters as a list"},
		{"a number as a sequence", `permutations(7, 2)`, "not a sequence"},
		{"a tuple of less than nothing", `permutations([1, 2], -1)`, "shorter than empty"},
		{"a repeat of less than nothing", `product([1, 2], repeat=-1)`, "must be between 0"},
		{"no r at all", `combinations([1, 2])`, "missing argument"},
		{"more orderings than will be built", `permutations(range(1000), 3)`, "elements"},
		{"more choices than will be built", `combinations(range(100), 10)`, "elements"},
		{"more of the same than will be built", `combinations_with_replacement(range(100), 10)`, "elements"},
		{"a product of more than will be built", `product(*[[0, 1]] * 20)`, "elements"},
		{"a tuple wider than will be built", `product([1], repeat=2000000)`, "repeat"},
		{"a tuple wider than will be built, the long way", `product(repeat=2000, *[[1]] * 1000)`, "a tuple of 2000000 values"},
		{"a product of a pair a hundred times over", `product([0, 1], repeat=100)`, "elements"},
		{"a tuple wider than will be built, out of nothing", `product([], [], repeat=600000)`, "a tuple of 1200000 values"},
		{"an r that is not a number", `combinations([1, 2], "two")`, "r must be a whole number"},
		{"an r wider than will be built", `combinations([1], 2000000)`, "a tuple of 2000000 values"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := run(t, "def solve(options):\n    return match(options, "+test.expression+")\n")
			if result.Status != solver.StatusError {
				t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusError)
			}
			if !strings.Contains(result.Message, test.contains) {
				t.Errorf("message: got %q, want it to mention %q", result.Message, test.contains)
			}
		})
	}
}

// TestTheWidthOfATupleIsPaidFor is the half of the rule a list's length cannot
// show. A thousand tuples of ten is ten thousand values allocated, and a
// budget that counted only the tuples would let a solver take a hundred times
// what it was given.
func TestTheWidthOfATupleIsPaidFor(t *testing.T) {
	t.Parallel()
	// A thousand and twenty-four tuples of ten: well inside the ceiling, and
	// ten thousand two hundred and forty values.
	const wide = "def solve(options):\n    return match(options, len(product(*[[0, 1]] * 10)))\n"

	roomy := sandbox(t, starlark.Limits{Steps: 20_000, Timeout: time.Minute, Concurrency: 1})
	result, err := roomy.Run(t.Context(), wide, options)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	if result.Status != solver.StatusOK {
		t.Errorf("with room: got %q (%s), want %q", result.Status, result.Message, solver.StatusOK)
	}

	tight := sandbox(t, starlark.Limits{Steps: 5_000, Timeout: time.Minute, Concurrency: 1})
	result, err = tight.Run(t.Context(), wide, options)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	if result.Status != solver.StatusTimeout {
		t.Errorf("with less: got %q (%s), want %q", result.Status, result.Message, solver.StatusTimeout)
	}

	// The same number of values in a flat list is paid for once each, so the
	// budget that would not cover the tuples covers these.
	flat := "def solve(options):\n    return match(options, len(list(range(1024))))\n"
	result, err = tight.Run(t.Context(), flat, options)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	if result.Status != solver.StatusOK {
		t.Errorf("flat: got %q (%s), want %q", result.Status, result.Message, solver.StatusOK)
	}
}
