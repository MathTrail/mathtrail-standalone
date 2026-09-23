package starlark_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// TestArithmetic covers the three helpers that are not about lists at all: two
// walks over a sequence of numbers and the divisor that makes exact fractions
// possible without dividing anything.
func TestArithmetic(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		expression string
		want       string
	}{
		{`sum([1, 2, 3])`, `6`},
		{`sum([])`, `0`},
		{`sum([1, 2], 10)`, `13`},
		{`sum([1, 2], start=10)`, `13`},
		{`sum(range(101))`, `5050`},
		{`sum([1.5, 2])`, `3.5`},
		// Whole numbers stay exact however long they get, which is what a
		// count of something big needs them to do.
		{`sum([100000000000000000000, 1])`, `100000000000000000001`},
		{`sum([days_in_month(2023, month) for month in [3, 4, 5]])`, `92`},

		{`prod([2, 3, 4])`, `24`},
		{`prod([])`, `1`},
		{`prod([2, 3], 10)`, `60`},
		{`prod(range(1, 11))`, `3628800`},
		{`prod([2, 2.5])`, `5.0`},

		{`gcd(12, 18)`, `6`},
		{`gcd(18, 12)`, `6`},
		{`gcd(-12, 18)`, `6`},
		{`gcd(12, -18)`, `6`},
		{`gcd(0, 5)`, `5`},
		{`gcd(0, 0)`, `0`},
		{`gcd(7, 13)`, `1`},
		{`gcd(3000000021, 5000000035)`, `1000000007`},
	} {
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()
			if got := repr(t, test.expression); got != test.want {
				t.Errorf("%s = %s, want %s", test.expression, got, test.want)
			}
		})
	}
}

func TestArithmeticRefusals(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		expression string
		contains   string
	}{
		{"words do not add up", `sum(["a", "b"])`, "not a number"},
		{"nor does a start of words", `sum([1], "x")`, "not a number"},
		{"a string is not a sequence", `sum("123")`, "write out its characters as a list"},
		{"a fraction has no divisor", `gcd(1.5, 2)`, "not a whole number"},
		{"gcd wants two", `gcd(12)`, "want 2"},
		{"more to walk than will be walked", `sum(range(1000000001))`, "elements"},
		{"nothing to add up", `sum()`, "missing argument"},
		{"a fraction as the second", `gcd(12, 1.5)`, "not a whole number"},
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
