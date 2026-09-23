package starlark_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// pouring is the shape of a real brute force: a breadth-first search over the
// states of two jugs, with a list and a head index for the queue and a
// dictionary for what has been seen. It is here because it uses the whole of
// what the dialect allows — a while loop, tuples as dictionary keys, slicing,
// top-level names — and none of what it does not.
const pouring = `CAPACITIES = [3, 5]
GOAL = 2

def next_states(state):
    following = []
    for i in range(len(CAPACITIES)):
        following.append(state[:i] + (CAPACITIES[i],) + state[i + 1:])
        following.append(state[:i] + (0,) + state[i + 1:])
        for j in range(len(CAPACITIES)):
            if i != j:
                amount = min(state[i], CAPACITIES[j] - state[j])
                poured = list(state)
                poured[i] = poured[i] - amount
                poured[j] = poured[j] + amount
                following.append(tuple(poured))
    return following

def solve(options):
    start = (0, 0)
    steps = {start: 0}
    queue = [start]
    head = 0
    while head < len(queue):
        state = queue[head]
        head = head + 1
        if GOAL in state:
            return match(options, steps[state])
        for following in next_states(state):
            if following not in steps:
                steps[following] = steps[state] + 1
                queue.append(following)
    return match(options, "It is impossible")
`

// pours are the options of the task the search above belongs to.
var pours = solver.Options{"1", "2", "3", "4", "It is impossible"}

func TestASearchRunsToItsAnswer(t *testing.T) {
	t.Parallel()
	result, err := sandbox(t, limits()).Run(context.Background(), pouring, pours)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	if result.Status != solver.StatusOK {
		t.Fatalf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusOK)
	}
	if !equal(result.Letters, []string{"B"}) {
		t.Errorf("letters: got %v, want [B]", result.Letters)
	}
	if result.Steps == 0 {
		t.Error("steps: got none, want what the search cost")
	}
}

// TestTwoRunsOfOneProgram is the rule the whole check rests on: the same
// program against the same options, the second time under shifted labels.
func TestTwoRunsOfOneProgram(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		source   string
		options  solver.Options
		letter   string
		contains string
	}{
		{
			name:    "a program that works the answer out",
			source:  "def solve(options):\n    return match(options, 2 * 3)\n",
			options: options,
			letter:  "C",
		},
		{
			name:    "a search",
			source:  pouring,
			options: pours,
			letter:  "B",
		},
		{
			name:     "a program that writes the letter out by hand",
			source:   "def solve(options):\n    return ['C']\n",
			options:  options,
			contains: "reading the letters",
		},
		{
			name:     "a program that fails",
			source:   "def solve(options):\n    fail('no')\n",
			options:  options,
			contains: "no",
		},
		{
			name:     "an answer no option carries",
			source:   "def solve(options):\n    return match(options, 7)\n",
			options:  options,
			contains: "no option that fits",
		},
		{
			name:     "two options saying the same thing",
			source:   "def solve(options):\n    return match(options, 6)\n",
			options:  solver.Options{"6", "5", "6", "8", "12"},
			contains: "more than one option",
		},
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
			if test.contains != "" && !strings.Contains(agreement.Explain(), test.contains) {
				t.Errorf("explanation: got %q, want it to mention %q", agreement.Explain(), test.contains)
			}
		})
	}
}

// TestTheSameProgramTwiceIsTheSameResult holds the sandbox to the property the
// whole verdict rests on: nothing in a run depends on anything but the source
// and the options.
func TestTheSameProgramTwiceIsTheSameResult(t *testing.T) {
	t.Parallel()
	runner := sandbox(t, limits())

	first, err := runner.Run(context.Background(), pouring, pours)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	for i := range 9 {
		again, err := runner.Run(context.Background(), pouring, pours)
		if err != nil {
			t.Fatalf("run %d: got error %v, want none", i, err)
		}
		if again.Status != first.Status || !equal(again.Letters, first.Letters) || again.Steps != first.Steps {
			t.Fatalf("run %d: got %q %v in %d steps, want %q %v in %d",
				i, again.Status, again.Letters, again.Steps, first.Status, first.Letters, first.Steps)
		}
	}
}

// TestPrintReachesNobody keeps what a program prints out of this service's own
// streams: a solver is written elsewhere, and what it has to say about a task
// is nobody's business here.
//
// It swaps a file the whole process shares, so it does not run beside anything.
func TestPrintReachesNobody(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	stderr := os.Stderr
	os.Stderr = write
	t.Cleanup(func() { os.Stderr = stderr })

	result := run(t, "def solve(options):\n    print('the answer is 6')\n    return match(options, 6)\n")

	if closed := write.Close(); closed != nil {
		t.Fatalf("Close: %v", closed)
	}
	printed, err := io.ReadAll(read)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(printed) != 0 {
		t.Errorf("stderr: got %q, want nothing", printed)
	}
	if result.Status != solver.StatusOK {
		t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusOK)
	}
}

// TestTheOptionsCannotBeEdited keeps a program from rewriting what it was
// given: an option it edited would be matched against a text the task never
// carried, and the two runs would be arguing about different tasks.
func TestTheOptionsCannotBeEdited(t *testing.T) {
	t.Parallel()
	result := run(t, "def solve(options):\n    options['A'] = '6'\n    return match(options, 6)\n")

	if result.Status != solver.StatusError {
		t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusError)
	}
	if !strings.Contains(result.Message, "frozen") {
		t.Errorf("message: got %q, want it to say the options do not change", result.Message)
	}
}

// TestALargeNumberIsWrittenOutInFull keeps a computed value in the shape an
// option is written in. The language's own way of putting a float into words
// turns a million into 1e+06, and no task offers that.
func TestALargeNumberIsWrittenOutInFull(t *testing.T) {
	t.Parallel()
	millions := solver.Options{"1000000", "100000", "10000", "1000", "100"}

	result, err := sandbox(t, limits()).Run(
		context.Background(), "def solve(options):\n    return match(options, 1000.0 * 1000.0)\n", millions)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	if result.Status != solver.StatusOK || !equal(result.Letters, []string{"A"}) {
		t.Errorf("got %q %v (%s), want ok and [A]", result.Status, result.Letters, result.Message)
	}
}
