package starlark_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// options are the five texts of a task whose answer is six.
var options = solver.Options{"4", "5", "6", "8", "12"}

// limits are what a run gets unless a test is about a limit itself.
func limits() starlark.Limits {
	return starlark.Limits{Steps: 25_000_000, Timeout: 2 * time.Second, Concurrency: 4}
}

func sandbox(t *testing.T, limits starlark.Limits) solver.Runner {
	t.Helper()
	runner, err := starlark.New(limits)
	if err != nil {
		t.Fatalf("New(%+v): %v", limits, err)
	}
	return runner
}

// run executes one source against the five options above.
func run(t *testing.T, source string) solver.Result {
	t.Helper()
	result, err := sandbox(t, limits()).Run(context.Background(), source, options)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	return result
}

func TestNewRefusesLimitsNothingCouldRunUnder(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		limits starlark.Limits
	}{
		{"no steps", starlark.Limits{Steps: 0, Timeout: time.Second, Concurrency: 1}},
		{"no clock", starlark.Limits{Steps: 1, Timeout: 0, Concurrency: 1}},
		{"no slots", starlark.Limits{Steps: 1, Timeout: time.Second, Concurrency: 0}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := starlark.New(test.limits); err == nil {
				t.Fatalf("New(%+v): got no error, want one", test.limits)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()
	for _, test := range []runCase{
		{
			name:    "a computed value looked up among the options",
			source:  "def solve(options):\n    return match(options, 2 * 3)\n",
			status:  solver.StatusOK,
			letters: []string{"C"},
		},
		{
			name:    "a value no option carries",
			source:  "def solve(options):\n    return match(options, 7)\n",
			status:  solver.StatusOK,
			letters: []string{},
		},
		{
			name:    "a tuple of letters",
			source:  "def solve(options):\n    return ('B', 'A')\n",
			status:  solver.StatusOK,
			letters: []string{"A", "B"},
		},
		{
			name:    "the top level runs before the call",
			source:  "ANSWER = 6\n\ndef solve(options):\n    return match(options, ANSWER)\n",
			status:  solver.StatusOK,
			letters: []string{"C"},
		},
		{
			name:     "a source that does not parse",
			source:   "def solve(options)\n    return []\n",
			status:   solver.StatusBadSource,
			contains: "got newline",
		},
		{
			name:     "an import",
			source:   "import math\n\ndef solve(options):\n    return []\n",
			status:   solver.StatusBadSource,
			contains: "got import",
		},
		{
			name:     "a load",
			source:   "load('helpers.star', 'helper')\n\ndef solve(options):\n    return []\n",
			status:   solver.StatusBadSource,
			contains: "load is not available",
		},
		{
			name:     "no solve at all",
			source:   "def answer(options):\n    return []\n",
			status:   solver.StatusNoEntryPoint,
			contains: "defines no solve",
		},
		{
			name:     "solve is not a function",
			source:   "solve = 6\n",
			status:   solver.StatusNoEntryPoint,
			contains: "rather than a function",
		},
		{
			name:     "solve takes nothing",
			source:   "def solve():\n    return []\n",
			status:   solver.StatusNoEntryPoint,
			contains: "exactly one argument",
		},
		{
			name:     "solve takes two",
			source:   "def solve(options, extra):\n    return []\n",
			status:   solver.StatusNoEntryPoint,
			contains: "exactly one argument",
		},
		{
			name:     "solve takes anything",
			source:   "def solve(*args):\n    return []\n",
			status:   solver.StatusNoEntryPoint,
			contains: "exactly one argument",
		},
		{
			name:     "solve takes keywords",
			source:   "def solve(**kwargs):\n    return []\n",
			status:   solver.StatusNoEntryPoint,
			contains: "exactly one argument",
		},
		{
			name:     "the program gives up",
			source:   "def solve(options):\n    fail('the wording is ambiguous')\n",
			status:   solver.StatusError,
			contains: "the wording is ambiguous",
		},
		{
			name:     "an index past the end",
			source:   "def solve(options):\n    return [1, 2][5]\n",
			status:   solver.StatusError,
			contains: "out of range",
		},
		{
			name:     "types that do not add up",
			source:   "def solve(options):\n    return match(options, 1 + 'one')\n",
			status:   solver.StatusError,
			contains: "unknown binary op",
		},
		{
			name:     "a recursive function",
			source:   "def down(n):\n    return 0 if n == 0 else down(n - 1)\n\ndef solve(options):\n    return match(options, down(3))\n",
			status:   solver.StatusError,
			contains: "called recursively",
		},
		{
			// Which of the two limits ends it depends on how fast the machine
			// is: the same twenty-five million steps take a fraction of the clock on
			// one and more than all of it on another. Both are a timeout, and
			// which one says so is settled where each is tested on its own.
			name:   "a loop with no end",
			source: "def solve(options):\n    while True:\n        pass\n    return []\n",
			status: solver.StatusTimeout,
		},
		{
			name:    "a float is written out in full",
			source:  "def solve(options):\n    return match(options, 2.0 * 3)\n",
			status:  solver.StatusOK,
			letters: []string{"C"},
		},
		{
			name:     "the options are not the ones the program was given",
			source:   "def solve(options):\n    return match([1, 2], 3)\n",
			status:   solver.StatusError,
			contains: "not a list",
		},
		{
			name:     "an options dictionary of its own making",
			source:   "def solve(options):\n    return match({'A': '6'}, 6)\n",
			status:   solver.StatusError,
			contains: "B is not among these",
		},
		{
			name:    "an empty sequence costs nothing",
			source:  "def solve(options):\n    return match(options, len(list([])) + 6)\n",
			status:  solver.StatusOK,
			letters: []string{"C"},
		},
		{
			name:     "a letter instead of a list",
			source:   "def solve(options):\n    return 'A'\n",
			status:   solver.StatusBadOutput,
			contains: "rather than a list",
		},
		{
			name:     "one letter twice",
			source:   "def solve(options):\n    return ['A', 'A']\n",
			status:   solver.StatusBadOutput,
			contains: "twice",
		},
		{
			name:     "a letter no option has",
			source:   "def solve(options):\n    return ['F']\n",
			status:   solver.StatusBadOutput,
			contains: "not one of the option letters",
		},
		{
			name:     "a letter in the wrong case",
			source:   "def solve(options):\n    return ['a']\n",
			status:   solver.StatusBadOutput,
			contains: "not one of the option letters",
		},
		{
			name:     "a number instead of a letter",
			source:   "def solve(options):\n    return [6]\n",
			status:   solver.StatusBadOutput,
			contains: "not one of the option letters",
		},
		{
			name:     "more letters than a task has options",
			source:   "def solve(options):\n    return ['A', 'B', 'C', 'D', 'E', 'A']\n",
			status:   solver.StatusBadOutput,
			contains: "6 letters",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.check(t, run(t, test.source))
		})
	}
}

// runCase is one source and what running it has to come to. The letters are
// checked only when a case names them, and the message only has to mention
// what the case says it must — the whole sentence is the sandbox's to word.
type runCase struct {
	name     string
	source   string
	status   solver.Status
	letters  []string
	contains string
}

func (c *runCase) check(t *testing.T, result solver.Result) {
	t.Helper()
	if result.Status != c.status {
		t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, c.status)
	}
	if c.letters != nil && !equal(result.Letters, c.letters) {
		t.Errorf("letters: got %v, want %v", result.Letters, c.letters)
	}
	if c.contains != "" && !strings.Contains(result.Message, c.contains) {
		t.Errorf("message: got %q, want it to mention %q", result.Message, c.contains)
	}
	if c.status != solver.StatusOK && result.Message == "" {
		t.Error("message: got nothing, want a sentence for the model")
	}
}

func equal(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// repr runs one expression and gives back what the language prints for it.
//
// The program fails on purpose: a solver can hand back nothing but letters,
// and a refusal is the one thing that carries a text of its own. It is how a
// test says what a helper returned rather than only whether it liked it.
func repr(t *testing.T, expression string) string {
	t.Helper()
	result := run(t, "def solve(options):\n    fail(str("+expression+"))\n")
	if result.Status != solver.StatusError {
		t.Fatalf("%s: got %q (%s), want the value itself", expression, result.Status, result.Message)
	}
	return strings.TrimPrefix(result.Message, "fail: ")
}
