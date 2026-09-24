package content_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// Every reference task claims an answer, and until it carries a solver that is
// all the claim is: a letter somebody typed. This is the bench that turns it
// into something proved, and a task with no solver fails it.
//
// It runs each solver the way a submitted one is run — the same sandbox, the
// same dialect and helpers, the same two runs under rotated labels, and the
// limits the deployed service uses. That makes the reference tasks the
// regression suite of the sandbox as well: a limit moved by a hand that did not
// mean to move it is caught here, against hundreds of real programs, rather
// than by a child.
func TestEveryReferenceTaskProvesItsOwnAnswer(t *testing.T) {
	t.Parallel()

	embedded := loaded(t)
	sandbox := serviceSandbox(t)

	var spent costs
	for _, task := range embedded.Examples() {
		t.Run(task.ID, func(t *testing.T) {
			spent.add(task.ID, prove(t, sandbox, &task))
		})
	}

	spent.report(t)
}

// A solver template is shown to the model as a program to start its own from,
// so it has to be a correct one, not merely a plausible one. Each still holds
// the numbers of the reference task it was generalised from, and is run on
// that task the way a submitted solver is run — the same sandbox, the same two
// runs, the limits the service uses — and has to prove that task's answer.
func TestEverySolverTemplateProvesTheTaskItWasWrittenFor(t *testing.T) {
	t.Parallel()

	embedded := loaded(t)
	sandbox := serviceSandbox(t)
	tasks := make(map[string]content.Example, embedded.ExampleCount())
	for _, task := range embedded.Examples() {
		tasks[task.ID] = task
	}

	var spent costs
	for _, topic := range embedded.Topics() {
		for _, template := range embedded.Templates(topic.ID) {
			name := topic.ID + "/" + template.Name
			t.Run(name, func(t *testing.T) {
				task := tasks[template.Source]
				task.Solver = template.Program
				spent.add(name, prove(t, sandbox, &task))
			})
		}
	}
	if spent.runs == 0 {
		t.Fatal("no template ran, so nothing here was tested")
	}

	spent.report(t)
}

// A template is where a model starts a solver of its own, and a slip made while
// filling one in has to stop the program rather than quietly prove the answer
// of another task. The template of what remains after shares knows the two
// things a share is taken of, and that a whole cannot give away more than
// itself.
func TestTheTemplateOfWhatRemainsStopsOnASlipInFillingItIn(t *testing.T) {
	t.Parallel()

	embedded := loaded(t)
	sandbox := serviceSandbox(t)
	template := templateNamed(t, embedded, "fractions.parts", "what-is-left")
	source := exampleNamed(t, embedded, template.Source)

	for _, test := range []struct{ name, steps, says string }{
		{"a share of something it does not know", `STEPS = [(1, 2, "rest"), (1, 4, "Rest")]`, `not of "Rest"`},
		{"shares that give away more than the whole", `STEPS = [(1, 2, "whole"), (2, 3, "whole")]`, "more than the whole"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			program := withLine(t, template.Program, "STEPS = ", test.steps)
			result, err := sandbox.Run(t.Context(), program, optionsOf(t, &source))
			if err != nil {
				t.Fatalf("Run() error = %v, want the run to happen", err)
			}
			if result.Status != solver.StatusError || !strings.Contains(result.Message, test.says) {
				t.Errorf("Run() = %s %q, want %s saying %q", result.Status, result.Message, solver.StatusError, test.says)
			}
		})
	}
}

// The search template trusts the largest number it found only when its search
// went on at least as far again past it, since a larger one could lie beyond
// where the search stops; a limit the question sets itself is an answer like
// any other. Filled for the task of the most nuts, whose answer is 80.
func TestTheSearchTemplateTellsTheQuestionsLimitFromWhereItStops(t *testing.T) {
	t.Parallel()

	embedded := loaded(t)
	sandbox := serviceSandbox(t)
	template := templateNamed(t, embedded, "number.divisibility", "search")
	task := exampleNamed(t, embedded, "div-56-d4-1")

	filled := template.Program
	for _, line := range [][2]string{
		{"SMALLEST = ", "SMALLEST = 1"},
		{"    return all(", "    return n // 9 == n % 9"},
		{"    return match(options, found[0])", "    return match(options, largest(found))"},
	} {
		filled = withLine(t, filled, line[0], line[1])
	}

	for _, test := range []struct {
		name, largest, capped string
		refused               bool
	}{
		{"a limit the question sets", "LARGEST = 80", "CAPPED = False", false},
		{"a search that stops at the answer", "LARGEST = 80", "CAPPED = True", true},
		{"a search that stops short of the answer", "LARGEST = 55", "CAPPED = True", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			program := withLine(t, withLine(t, filled, "LARGEST = ", test.largest), "CAPPED = ", test.capped)
			result, err := sandbox.Run(t.Context(), program, optionsOf(t, &task))
			switch {
			case err != nil:
				t.Fatalf("Run() error = %v, want the run to happen", err)
			case test.refused && (result.Status != solver.StatusError || !strings.Contains(result.Message, "where the search stops")):
				t.Errorf("Run() = %s %v %q, want %s saying the search stopped too soon", result.Status, result.Letters, result.Message, solver.StatusError)
			case !test.refused && (!result.One() || result.Letters[0] != task.CorrectAnswer):
				t.Errorf("Run() = %s %v %q, want %s", result.Status, result.Letters, result.Message, task.CorrectAnswer)
			}
		})
	}
}

// templateNamed is one solver template of a topic, found by its name.
func templateNamed(t *testing.T, embedded *content.Content, topic, name string) content.Template {
	t.Helper()

	for _, template := range embedded.Templates(topic) {
		if template.Name == name {
			return template
		}
	}
	t.Fatalf("%s has no template named %s", topic, name)
	return content.Template{}
}

// exampleNamed is the reference task with this id.
func exampleNamed(t *testing.T, embedded *content.Content, id string) content.Example {
	t.Helper()

	examples := embedded.Examples()
	for i := range examples {
		if examples[i].ID == id {
			return examples[i]
		}
	}
	t.Fatalf("no reference task has the id %s", id)
	return content.Example{}
}

// withLine is a program with its one line that starts with prefix replaced.
func withLine(t *testing.T, program, prefix, line string) string {
	t.Helper()

	lines := strings.Split(program, "\n")
	found := 0
	for i := range lines {
		if strings.HasPrefix(lines[i], prefix) {
			lines[i] = line
			found++
		}
	}
	if found != 1 {
		t.Fatalf("the program has %d lines starting with %q, want one", found, prefix)
	}
	return strings.Join(lines, "\n")
}

// A template is generalised from the solver of a reference task and never
// written from nothing. Every topic has its reference tasks, so every topic
// has at least one template: a package without one would leave the model to
// write its solver from nothing.
func TestEveryTopicHasASolverTemplate(t *testing.T) {
	t.Parallel()

	embedded := loaded(t)
	for _, topic := range embedded.Topics() {
		if len(embedded.Templates(topic.ID)) == 0 {
			t.Errorf("%s: got no solver template, want one generalised from the solvers of its "+
				"reference tasks, in solvers/%s/", topic.ID, topic.ID)
		}
	}
}

// serviceSandbox is the sandbox a submitted solver runs in, with the limits
// the deployed service uses.
func serviceSandbox(t *testing.T) solver.Runner {
	t.Helper()

	sandbox, err := starlark.New(starlark.Limits{
		Steps:       config.DefaultSolverSteps,
		Timeout:     config.DefaultSolverTimeout,
		Concurrency: config.DefaultSolverConcurrency,
	})
	if err != nil {
		t.Fatalf("starlark.New() error = %v, want nil", err)
	}
	return sandbox
}

// prove runs one task's solver the way a submitted one is run, and holds it to
// the answer the task claims.
func prove(t *testing.T, sandbox solver.Runner, task *content.Example) solver.Agreement {
	t.Helper()

	if task.Solver == "" {
		t.Fatalf("%s carries no solver: it belongs in examples/solvers/%s.star", task.ID, task.ID)
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
