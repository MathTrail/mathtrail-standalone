package starlark_test

import (
	"context"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// benchmarkRun is what a run costs from source to letters: the parse, the
// compilation and the call, every time, because a program is never kept.
func benchmarkRun(b *testing.B, source string, options solver.Options) {
	b.Helper()
	runner, err := starlark.New(starlark.Limits{
		Steps:       25_000_000,
		Timeout:     2 * time.Second,
		Concurrency: 4,
	})
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	b.ReportAllocs()
	for b.Loop() {
		result, err := runner.Run(ctx, source, options)
		if err != nil {
			b.Fatalf("Run: %v", err)
		}
		// One letter rather than only a status: a solver that computed the
		// wrong number would still run to the end, and a benchmark of the
		// wrong computation is worth nothing.
		if !result.One() {
			b.Fatalf("got %q %v (%s), want %q and one letter",
				result.Status, result.Letters, result.Message, solver.StatusOK)
		}
	}
}

// BenchmarkLookup is the floor: everything a run does around a program that
// does nothing.
func BenchmarkLookup(b *testing.B) {
	benchmarkRun(b, "def solve(options):\n    return match(options, 2 * 3)\n", options)
}

// BenchmarkSearch is a real one: the breadth-first search over jug states.
func BenchmarkSearch(b *testing.B) {
	benchmarkRun(b, pouring, pours)
}

// BenchmarkOrdering is a brute force of the shape most of them have: every
// ordering of a handful of things, each one held against the conditions of the
// task.
func BenchmarkOrdering(b *testing.B) {
	const ordering = `def solve(options):
    count = 0
    for order in permutations(range(6)):
        if order[0] < order[1] and order[2] > order[3]:
            count = count + 1
    return match(options, count)
`
	benchmarkRun(b, ordering, solver.Options{"60", "120", "180", "240", "360"})
}
