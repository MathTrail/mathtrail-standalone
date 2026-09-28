package lesson_test

import (
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// costlyTask is the task of the costly solver named, sized to the ceiling
// given.
func costlyTask(t *testing.T, name string, steps uint64) lesson.Task {
	t.Helper()

	task, err := lesson.CostlyTask(name, steps)
	if err != nil {
		t.Fatalf("CostlyTask(%q, %d) error = %v", name, steps, err)
	}
	return task
}

// optionsOf are the options of a task, as the sandbox hands them to a solver.
func optionsOf(task *lesson.Task) solver.Options {
	var options solver.Options
	for place, letter := range solver.Letters() {
		options[place] = task.Body.Options[letter]
	}
	return options
}

// sandbox is the service's own sandbox, with a slot for every run a case
// makes at once.
func sandbox(t *testing.T, steps uint64, clock time.Duration) solver.Runner {
	t.Helper()

	runner, err := starlark.New(starlark.Limits{Steps: steps, Timeout: clock, Concurrency: 2, Wait: time.Minute})
	if err != nil {
		t.Fatalf("starlark.New() error = %v", err)
	}
	return runner
}

// Every costly solver spends nine to ten tenths of the ceiling it is sized to,
// in both of its runs, and still proves the task's answer: the budget is what
// it spends, not what stops it.
func TestEveryCostlySolverSpendsNearlyTheWholeBudget(t *testing.T) {
	t.Parallel()

	for _, name := range lesson.Costly() {
		if name == lesson.Product {
			continue
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			task := costlyTask(t, name, lesson.MinSteps)
			agreement, err := solver.Verdict(t.Context(), sandbox(t, lesson.MinSteps, time.Minute), task.Solver, optionsOf(&task))
			if err != nil {
				t.Fatalf("Verdict() error = %v", err)
			}
			if !agreement.Agreed() || agreement.Letter != task.Body.Correct || len(agreement.Runs) != 2 {
				t.Fatalf("Verdict() = %+v (%s), want both runs pointing at %s", agreement, agreement.Explain(), task.Body.Correct)
			}
			for i := range agreement.Runs {
				spentItsShare(t, i+1, &agreement.Runs[i])
			}
		})
	}
}

// spentItsShare is a case that fails unless the run spent nine to ten tenths
// of the ceiling the costly solvers are sized to in their tests.
func spentItsShare(t *testing.T, which int, run *solver.Result) {
	t.Helper()

	if share := float64(run.Steps) / lesson.MinSteps; share < 0.9 || share >= 1 {
		t.Errorf("run %d spent %d steps, %.1f%% of %d, want nine to ten tenths of it",
			which, run.Steps, 100*share, lesson.MinSteps)
	}
}

// The product is stopped by the clock alone: it costs the budget a hundred
// thousand steps, and its multiplying goes on past any clock of the sandbox.
func TestOnlyTheClockStopsTheProduct(t *testing.T) {
	t.Parallel()

	const clock = 300 * time.Millisecond
	task := costlyTask(t, lesson.Product, lesson.MinSteps)
	result, err := sandbox(t, lesson.MinSteps, clock).Run(t.Context(), task.Solver, optionsOf(&task))
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != solver.StatusTimeout || !strings.Contains(result.Message, clock.String()) || result.Steps >= lesson.MinSteps {
		t.Errorf("Run() = %s after %d steps (%s), want the run stopped by the clock of %v, short of the budget",
			result.Status, result.Steps, result.Message, clock)
	}
}

// A costly solver is one of those there are, sized to a ceiling it can be
// sized to.
func TestACostlySolverNobodyCouldRunIsRefused(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		steps uint64
	}{
		{"nothing", lesson.MinSteps},
		{lesson.Loop, lesson.MinSteps - 1},
	} {
		if task, err := lesson.CostlyTask(test.name, test.steps); err == nil {
			t.Errorf("CostlyTask(%q, %d) = a task of %d bytes, want an error", test.name, test.steps, len(task.Solver))
		}
	}
}

// A task handed in for a request nobody opened is turned away as stale, by a
// child with no profile at all.
func TestATaskHandedInForNoRequestIsStale(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t)
	service := session.Open(target, time.Minute)
	defer service.Close()
	child := service.Child("stale")
	defer func() { _ = child.Close() }()
	task := costlyTask(t, lesson.Loop, lesson.MinSteps)

	if answer := lesson.HandInStale(t.Context(), child, &task); answer.Kind.String() != "stale:stale_request" {
		t.Errorf("HandInStale() = %q (%v), want the task turned away as stale", answer.Kind, answer.Err)
	}
}
