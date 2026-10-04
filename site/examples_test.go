// Package site_test proves what the site's pages claim that a program can
// check: the answer of every example a topic's page works through.
package site_test

import (
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// example is what the site's data says of one example on a topic's page that a
// program can check: the solver that searches it, and the answer the page
// works it through to.
type example struct {
	Solver string `json:"solver"`
	Answer string `json:"answer"`
}

// strangers stand beside an example's answer as the other four options: texts
// no answer is, so that a solver passes only by finding the answer itself.
var strangers = [solver.Count - 1]string{"no answer 1", "no answer 2", "no answer 3", "no answer 4"}

// An example on a topic's page is worked through to its answer in words, in
// every language of the site. The answer is data as well, and a solver written
// for the example from its topic's template searches every case the example
// leaves open. This runs each solver the way a submitted one is run — the same
// sandbox, the same dialect and helpers, two runs under rotated labels, the
// limits the deployed service uses — and holds it to finding that answer: an
// answer that was never right, or a solver changed without it, stops here
// rather than on a parent's screen. The words are not read here: that they
// state the task the solver searches, and its answer, is for whoever reads
// them to hold.
func TestEveryExampleOnTheSiteProvesItsAnswer(t *testing.T) {
	t.Parallel()

	sandbox := serviceSandbox(t)
	examples := examplesOf(t)
	for _, topic := range slices.Sorted(maps.Keys(examples)) {
		for _, example := range examples[topic] {
			t.Run(topic+"/"+example.Solver, func(t *testing.T) {
				prove(t, sandbox, solverOf(topic, example), example.Answer)
			})
		}
	}
}

// A solver that no example names proves nothing a page shows: it was left
// behind by an example taken off its page, or the data spells its example's
// name another way, and that example went unproved.
func TestEverySolverOnTheSiteBelongsToAnExample(t *testing.T) {
	t.Parallel()

	named := map[string]bool{}
	for topic, examples := range examplesOf(t) {
		for _, example := range examples {
			named[solverOf(topic, example)] = true
		}
	}
	err := filepath.WalkDir("solvers", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if !named[path] {
			t.Errorf("%s: no example of the site's data names this solver", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(solvers) error = %v, want every solver read", err)
	}
}

// examplesOf is the examples of the site's data, by the topic their page is
// about.
func examplesOf(t *testing.T) map[string][]example {
	t.Helper()

	file, err := os.ReadFile("data.json")
	if err != nil {
		t.Fatalf("ReadFile(data.json) error = %v, want the site's data", err)
	}
	var data struct {
		Examples map[string][]example `json:"examples"`
	}
	if err := json.Unmarshal(file, &data); err != nil {
		t.Fatalf("data.json: %v", err)
	}
	if len(data.Examples) == 0 {
		t.Fatal("data.json names no example, so nothing here was tested")
	}
	return data.Examples
}

// solverOf is where the solver of an example of a topic is kept.
func solverOf(topic string, example example) string {
	return filepath.Join("solvers", topic, example.Solver+".star")
}

// prove runs the solver at path the way a submitted one is run, on the answer
// beside four strangers, and holds it to the answer.
func prove(t *testing.T, sandbox solver.Runner, path, answer string) {
	t.Helper()

	program, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v, want the example's solver", path, err)
	}
	options := solver.Options{answer}
	copy(options[1:], strangers[:])
	agreement, err := solver.Verdict(t.Context(), sandbox, string(program), options)
	if err != nil {
		t.Fatalf("Verdict() error = %v, want the run to happen", err)
	}
	if !agreement.Agreed() {
		t.Fatalf("%s did not prove an answer: %s", path, agreement.Explain())
	}
	if agreement.Letter != solver.Letter(0) {
		t.Errorf("%s proves %q, the page says %q", path, options[solver.Place(agreement.Letter)], answer)
	}
}

// serviceSandbox is the sandbox a submitted solver runs in, with the limits
// the deployed service uses — all but the wait for a slot, which no example
// is about.
func serviceSandbox(t *testing.T) solver.Runner {
	t.Helper()

	sandbox, err := starlark.New(starlark.Limits{
		Steps:       config.DefaultSolverSteps,
		Timeout:     config.DefaultSolverTimeout,
		Concurrency: config.DefaultSolverConcurrency,
		Wait:        time.Minute,
	})
	if err != nil {
		t.Fatalf("starlark.New() error = %v, want nil", err)
	}
	return sandbox
}
