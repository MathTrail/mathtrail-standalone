package starlark_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// forever is a program the budgets have to stop, because nothing else will.
const forever = "def solve(options):\n    while True:\n        pass\n    return []\n"

func TestSourceLongerThanTheLimit(t *testing.T) {
	t.Parallel()
	result := run(t, "def solve(options):\n    # "+strings.Repeat("x", 9*1024)+"\n    return []\n")
	if result.Status != solver.StatusBadSource {
		t.Errorf("status: got %q, want %q", result.Status, solver.StatusBadSource)
	}
	if !strings.Contains(result.Message, "longer than 8 KB") {
		t.Errorf("message: got %q, want it to give the limit", result.Message)
	}
	if result.Steps != 0 {
		t.Errorf("steps: got %d, want none spent on a source that was never parsed", result.Steps)
	}
}

func TestTheStepBudgetStopsALoop(t *testing.T) {
	t.Parallel()
	const budget = 200_000
	runner := sandbox(t, starlark.Limits{Steps: budget, Timeout: time.Minute, Concurrency: 1})

	// The clock is set far out of reach, and the message is read rather than
	// the status: both limits end a run the same way, and this test is about
	// which of the two ended this one.
	started := time.Now()
	result, err := runner.Run(context.Background(), forever, options)
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	if result.Status != solver.StatusTimeout {
		t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusTimeout)
	}
	if !strings.Contains(result.Message, "more than 200000 steps") {
		t.Errorf("message: got %q, want it to name the budget that stopped the loop", result.Message)
	}
	if result.Steps < budget {
		t.Errorf("steps: got %d, want at least the budget of %d", result.Steps, budget)
	}
	if elapsed > 5*time.Second {
		t.Errorf("elapsed: got %v, want the budget to stop the loop long before the clock", elapsed)
	}
}

func TestTheClockStopsWhatTheBudgetWouldNot(t *testing.T) {
	t.Parallel()
	// A budget no loop can spend in the time given, so that what stops the
	// program is the clock and only the clock.
	runner := sandbox(t, starlark.Limits{Steps: 1 << 62, Timeout: 50 * time.Millisecond, Concurrency: 1})

	started := time.Now()
	result, err := runner.Run(context.Background(), forever, options)
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	if result.Status != solver.StatusTimeout {
		t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusTimeout)
	}
	if !strings.Contains(result.Message, "50ms") {
		t.Errorf("message: got %q, want it to give the clock", result.Message)
	}
	if elapsed > time.Second {
		t.Errorf("elapsed: got %v, want the run stopped soon after 50ms", elapsed)
	}
}

func TestACallerThatHasGone(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := sandbox(t, limits()).Run(ctx, forever, options); err == nil {
		t.Fatal("Run: got no error, want the run refused")
	}
}

// TestASequenceTooLongToWalk covers the one thing the interpreter does not:
// memory. It counts its own instructions and looks for a cancellation only
// between them, so a single call that walks a billion-element range spends the
// whole instance with neither limit watching. The durations below are the
// point of the test — a refusal that arrives after the walk is no refusal.
func TestASequenceTooLongToWalk(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"len(list(range(1000000000)))",
		"len(tuple(range(1000000000)))",
		"len(sorted(range(1000000000)))",
		"len(set(range(1000000000)))",
		"len(reversed(range(1000000000)))",
		"len(list(enumerate(range(1000000000))))",
		"len(list(zip(range(1000000000), [1, 2])))",
		"min(range(1000000000))",
		"max(range(1000000000))",
		"1 if any(range(1000000000)) else 0",
		"1 if all(range(1000000000)) else 0",
		// The same sequence handed over by name rather than by position: some
		// of these built-ins take it either way, and the cap has to weigh a
		// keyword argument exactly as it weighs a positional one.
		"len(sorted(iterable=range(1000000000)))",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			result := run(t, "def solve(options):\n    return match(options, "+source+")\n")
			if result.Status != solver.StatusError {
				t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusError)
			}
			if !strings.Contains(result.Message, "elements") {
				t.Errorf("message: got %q, want it to say how much is too much", result.Message)
			}
			if result.Duration > 100*time.Millisecond {
				t.Errorf("duration: got %v, want the refusal before the walk rather than after it", result.Duration)
			}
		})
	}
}

// TestABuiltinCannotSpendWhatIsLeft is the other half of the same rule: a
// sequence within the cap still has to be paid for out of the step budget, and
// paid for before it is built.
func TestABuiltinCannotSpendWhatIsLeft(t *testing.T) {
	t.Parallel()
	const budget = 10_000
	runner := sandbox(t, starlark.Limits{Steps: budget, Timeout: time.Minute, Concurrency: 1})

	// A hundred thousand elements: well under the cap, well over the budget.
	source := "def solve(options):\n    return match(options, len(list(range(100000))))\n"
	result, err := runner.Run(context.Background(), source, options)
	if err != nil {
		t.Fatalf("Run: got error %v, want none", err)
	}
	if result.Status != solver.StatusTimeout {
		t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusTimeout)
	}
	if result.Steps < budget {
		t.Errorf("steps: got %d, want the budget of %d charged for what was asked", result.Steps, budget)
	}
	if result.Duration > 100*time.Millisecond {
		t.Errorf("duration: got %v, want the refusal before the list was built", result.Duration)
	}
}

// TestOnlySoManyRunsAtOnce holds the sandbox to its slots. Each run below
// spends its whole clock, so with one slot three of them take three times as
// long as one, and with no slots at all they would take as long as one.
func TestOnlySoManyRunsAtOnce(t *testing.T) {
	t.Parallel()
	const spell = 100 * time.Millisecond
	runner := sandbox(t, starlark.Limits{Steps: 1 << 62, Timeout: spell, Concurrency: 1})

	var waiting sync.WaitGroup
	started := time.Now()
	for range 3 {
		waiting.Add(1)
		go func() {
			defer waiting.Done()
			if _, err := runner.Run(context.Background(), forever, options); err != nil {
				t.Errorf("Run: got error %v, want none", err)
			}
		}()
	}
	waiting.Wait()

	if elapsed := time.Since(started); elapsed < 2*spell {
		t.Errorf("elapsed: got %v for three runs of %v through one slot, want them queued", elapsed, spell)
	}
}

// TestSlotsComeBack is the other side of it: a sandbox that forgot to release
// a slot would run exactly as many programs as it has slots and then stop.
func TestSlotsComeBack(t *testing.T) {
	t.Parallel()
	runner := sandbox(t, starlark.Limits{Steps: 1_000_000, Timeout: time.Minute, Concurrency: 1})

	for i := range 4 {
		result, err := runner.Run(context.Background(), "def solve(options):\n    return match(options, 6)\n", options)
		if err != nil {
			t.Fatalf("run %d: got error %v, want none", i, err)
		}
		if result.Status != solver.StatusOK {
			t.Fatalf("run %d: got %q (%s), want %q", i, result.Status, result.Message, solver.StatusOK)
		}
	}
}

// TestTheCapIsWhereItSays walks either side of the cap, because a limit that
// is never approached in a test is a limit nobody has checked.
func TestTheCapIsWhereItSays(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		elements string
		status   solver.Status
	}{
		{"one element under the cap", "999999", solver.StatusOK},
		{"the cap itself", "1000000", solver.StatusOK},
		{"one element over", "1000001", solver.StatusError},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "def solve(options):\n    return match(options, len(list(range(" + test.elements + "))))\n"
			result := run(t, source)
			if result.Status != test.status {
				t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, test.status)
			}
		})
	}
}

// TestACallerThatGoesAwayMidRun is the other half of a caller leaving: the run
// has already started, and what comes back is the sandbox saying it could not
// finish rather than a verdict about the solver.
func TestACallerThatGoesAwayMidRun(t *testing.T) {
	t.Parallel()
	runner := sandbox(t, starlark.Limits{Steps: 1 << 62, Timeout: time.Minute, Concurrency: 1})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	defer cancel()

	result, err := runner.Run(ctx, forever, options)
	if err == nil {
		t.Fatalf("Run: got %q (%s), want the run reported as stopped", result.Status, result.Message)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error: got %v, want it to carry the cancellation", err)
	}
}

// TestACallerWhoseTimeRanOut is the same thing arriving as a deadline rather
// than as a cancellation, and it is the ordinary case: a request carries a
// budget of its own, and when that budget is shorter than the sandbox's own
// clock it is the one that ends the run. What comes back then is the sandbox
// saying it could not finish — not a verdict, and above all not a verdict
// blamed on a limit of ours that never fired.
func TestACallerWhoseTimeRanOut(t *testing.T) {
	t.Parallel()
	runner := sandbox(t, starlark.Limits{Steps: 1 << 62, Timeout: time.Minute, Concurrency: 1})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := runner.Run(ctx, forever, options)
	if err == nil {
		t.Fatalf("Run: got %q (%s), want the run reported as stopped", result.Status, result.Message)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error: got %v, want it to carry the caller's deadline", err)
	}
	if strings.Contains(err.Error(), "1m0s") {
		t.Errorf("error: got %v, want nothing said about a limit that never fired", err)
	}
}

// TestASlotIsWorthWaitingForButNotForEver holds the queue to the caller's own
// deadline: a request that will not be answered in time should not be holding
// a place in the queue when its time runs out.
func TestASlotIsWorthWaitingForButNotForEver(t *testing.T) {
	t.Parallel()
	runner := sandbox(t, starlark.Limits{Steps: 1 << 62, Timeout: time.Second, Concurrency: 1})

	// The blocker takes the only slot and holds it for its whole clock, and
	// what stands between the call and the slot is parsing three lines. The
	// head start is that, a thousand times over; there is no way to watch a
	// slot being taken from outside, and adding one to the sandbox for the
	// sake of a test would be the test writing the code.
	go func() {
		_, _ = runner.Run(context.Background(), forever, options)
	}()
	time.Sleep(250 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := runner.Run(ctx, forever, options)
	if err == nil {
		t.Fatalf("Run: got %q (%s), want the wait given up on", result.Status, result.Message)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error: got %v, want it to carry the deadline", err)
	}
}
