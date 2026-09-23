package starlark_test

import (
	"context"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// checkLetters holds an answer to the shape the contract promises whoever
// reads it: option letters, each of them once.
func checkLetters(t *testing.T, letters []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, letter := range letters {
		if solver.Place(letter) < 0 {
			t.Fatalf("letters: got %v, and %q labels no option", letters, letter)
		}
		if seen[letter] {
			t.Fatalf("letters: got %v, with %q twice", letters, letter)
		}
		seen[letter] = true
	}
}

// FuzzRun feeds the sandbox what it is actually given: a source written
// somewhere else, by something that was never held to a grammar. Whatever
// arrives, a run answers with one of the six statuses and the process carries
// on — a panic here would take down the request handler around it.
func FuzzRun(f *testing.F) {
	for _, seed := range []string{
		"",
		"def solve(options):\n    return match(options, 6)\n",
		"def solve(options):\n    return ['A']\n",
		"def solve(options):\n    fail('no')\n",
		"def solve(options):\n    while True:\n        pass\n",
		"def solve(options)\n",
		"load('x.star', 'y')\n",
		"solve = 6\n",
		"\x00\xff",
		"def solve(options):\n    return [len(list(range(9999999)))]\n",
	} {
		f.Add(seed)
	}

	// Small budgets: a fuzz run is thousands of programs, and none of them has
	// anything to say that takes a second to say it.
	runner, err := starlark.New(starlark.Limits{
		Steps:       100_000,
		Timeout:     200 * time.Millisecond,
		Concurrency: 2,
	})
	if err != nil {
		f.Fatalf("New: %v", err)
	}

	f.Fuzz(func(t *testing.T, source string) {
		result, err := runner.Run(context.Background(), source, options)
		if err != nil {
			t.Fatalf("Run: got error %v, want none", err)
		}

		switch result.Status {
		case solver.StatusOK:
			checkLetters(t, result.Letters)
		case solver.StatusBadSource, solver.StatusNoEntryPoint,
			solver.StatusError, solver.StatusTimeout, solver.StatusBadOutput:
			if result.Message == "" {
				t.Fatalf("status %q: got no message, want a sentence for the model", result.Status)
			}
		default:
			t.Fatalf("status: got %q, want one the contract names", result.Status)
		}
	})
}
