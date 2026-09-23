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
		Steps:       10_000_000,
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
		if result.Status != solver.StatusOK {
			b.Fatalf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusOK)
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
