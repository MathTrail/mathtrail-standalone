package starlark_test

import (
	"context"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The four solvers here are the ports written out in full: a one-liner, a
// comprehension over a helper, an iterative search and a dynamic program.
// Between them they use every helper a brute force reaches for and every
// construct the dialect allows, and they run the way a submitted solver runs —
// twice, on shifted labels, through the same limits.
//
// They are what says the vocabulary is enough. A helper that returned the
// right list in a table and the wrong answer in a real solver would pass every
// other test in this package.

// enumeration counts the pairs that can be chosen from three children.
const enumeration = `def solve(options):
    pairs = combinations(["Ann", "Ben", "Kim"], 2)
    return match(options, len(pairs))
`

// calendarDays adds up the days of three months.
const calendarDays = `def solve(options):
    days = sum([days_in_month(2023, month) for month in [3, 4, 5]])
    return match(options, days)
`

// weighings is the fewest weighings that always find the one lighter coin
// among twelve. Recursion is closed, so the recurrence is filled in from the
// bottom instead.
const weighings = `def solve(options):
    best = {0: 0, 1: 0}
    for coins in range(2, 13):
        best[coins] = min([1 + max(best[k], best[coins - 2 * k]) for k in range(1, coins // 2 + 1)])
    return match(options, best[12])
`

func TestThePorts(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		source  string
		options solver.Options
		letter  string
	}{
		{"pairs of three children", enumeration, solver.Options{"2", "3", "4", "5", "6"}, "B"},
		{"days in three months", calendarDays, solver.Options{"90", "91", "92", "93", "94"}, "C"},
		{"pouring between two jugs", pouring, pours, "B"},
		{"weighings for twelve coins", weighings, solver.Options{"1", "2", "3", "4", "5"}, "C"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			agreement, err := solver.Verdict(context.Background(), sandbox(t, limits()), test.source, test.options)
			if err != nil {
				t.Fatalf("Verdict: got error %v, want none", err)
			}
			if agreement.Letter != test.letter {
				t.Errorf("letter: got %q (%s), want %q", agreement.Letter, agreement.Explain(), test.letter)
			}
			for i, run := range agreement.Runs {
				if run.Steps == 0 {
					t.Errorf("run %d: got no steps, want what the search cost", i)
				}
			}
		})
	}
}
