package content_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// notYetPorted are the topics whose reference tasks still carry no solver.
//
// It shrinks with every batch and is empty once the last one lands, at which
// point this variable goes away with it. A topic is removed from here in the
// same change that gives its tasks their solvers — the bench refuses either
// half on its own, so the two cannot drift apart.
var notYetPorted = []string{
	"algorithms.weighing_pouring",
	"arithmetic.tricks",
	"logic.knights_liars",
	"parity.alternation",
	"pigeonhole.basic",
	"time.calendar",
	"time.clocks",
}

// Every reference task claims an answer, and until it carries a solver that is
// all the claim is: a letter somebody typed. This is the bench that turns it
// into something proved.
//
// It runs each solver the way a submitted one is run — the same sandbox, the
// same dialect and helpers, the same two runs under rotated labels, and the
// limits the deployed service uses. That makes the reference tasks the
// regression suite of the sandbox as well: a limit moved by a hand that did not
// mean to move it is caught here, against four hundred and fifty real programs,
// rather than by a child.
func TestEveryReferenceTaskProvesItsOwnAnswer(t *testing.T) {
	t.Parallel()

	embedded, err := content.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	sandbox, err := starlark.New(starlark.Limits{
		Steps:       config.DefaultSolverSteps,
		Timeout:     config.DefaultSolverTimeout,
		Concurrency: config.DefaultSolverConcurrency,
	})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	var spent costs
	for _, task := range embedded.Examples() {
		t.Run(task.ID, func(t *testing.T) {
			if slices.Contains(notYetPorted, task.Topic) {
				refuseSolver(t, &task)
				return
			}
			spent.add(task.ID, prove(t, sandbox, &task))
		})
	}

	spent.report(t)
}

// refuseSolver is the other half of the list of topics not yet ported. Without
// it a topic could be ported and left named there, and the bench would quietly
// stop checking the very tasks that had just gained solvers.
func refuseSolver(t *testing.T, task *content.Example) {
	t.Helper()

	if task.Solver != "" {
		t.Errorf("%s has a solver, but %s is still listed as not yet ported", task.ID, task.Topic)
	}
}

// prove runs one task's solver the way a submitted one is run, and holds it to
// the answer the task claims.
func prove(t *testing.T, sandbox solver.Runner, task *content.Example) solver.Agreement {
	t.Helper()

	if task.Solver == "" {
		t.Fatalf("%s carries no solver, and %s is not listed as not yet ported", task.ID, task.Topic)
	}

	agreement, err := solver.Verdict(t.Context(), sandbox, task.Solver, optionsOf(t, task))
	if err != nil {
		t.Fatalf("Verdict() error = %v, want the run to happen", err)
	}
	if !agreement.Agreed() {
		t.Fatalf("the solver did not prove an answer: %s", agreement.Explain())
	}
	if agreement.Letter != task.CorrectAnswer {
		t.Errorf("the solver proves %s, the task claims %s", agreement.Letter, task.CorrectAnswer)
	}
	return agreement
}

// optionsOf is the task's five option texts in the order the solver contract
// expects them.
func optionsOf(t *testing.T, task *content.Example) solver.Options {
	t.Helper()

	var options solver.Options
	for letter, text := range task.Options {
		place := solver.Place(letter)
		if place < 0 {
			t.Fatalf("option %q is not a letter from A to E", letter)
		}
		options[place] = text
	}
	return options
}

// costs is what the run of every solver spent, which is what says whether the
// sandbox's ceilings are the right ones.
type costs struct {
	runs int

	mostSteps   uint64
	stepsBy     string
	longest     time.Duration
	longestBy   string
	nearCeiling []string
}

// nearCeiling is the share of the step budget past which a solver is worth
// naming: a program spending a tenth of what it is allowed is not in danger,
// but it is the one that will be when a limit moves.
const nearCeiling = config.DefaultSolverSteps / 10

func (c *costs) add(id string, agreement solver.Agreement) {
	for _, run := range agreement.Runs {
		c.runs++
		if run.Steps > c.mostSteps {
			c.mostSteps, c.stepsBy = run.Steps, id
		}
		if run.Duration > c.longest {
			c.longest, c.longestBy = run.Duration, id
		}
		if run.Steps >= nearCeiling && !slices.Contains(c.nearCeiling, id) {
			c.nearCeiling = append(c.nearCeiling, id)
		}
	}
}

// report prints what the bench measured. The numbers are read by a person
// deciding what the limits should be, so they are logged rather than asserted:
// a ceiling that a test enforces is a ceiling nobody ever revisits.
func (c *costs) report(t *testing.T) {
	t.Helper()

	if c.runs == 0 {
		t.Log("no solver ran: every topic is still listed as not yet ported")
		return
	}
	t.Logf("%d runs; most steps %d (%s) of %d allowed; longest %v (%s) of %v allowed; %d at or past a tenth of the budget%s",
		c.runs,
		c.mostSteps, c.stepsBy, uint64(config.DefaultSolverSteps),
		c.longest.Round(time.Microsecond), c.longestBy, config.DefaultSolverTimeout,
		len(c.nearCeiling), named(c.nearCeiling))
}

// named lists the tasks worth looking at, and says nothing when there are none.
func named(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return fmt.Sprintf(": %v", ids)
}
