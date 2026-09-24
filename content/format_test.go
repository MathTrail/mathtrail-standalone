package content_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// A message a program builds with % is only as good as its format: a stray %,
// as in "a rise of 10% on %d", or a format that asks for more values than it
// is given, stops the program with an error about the format in place of the
// message, and exactly when the message was needed. A template is copied by
// the model, mistakes and all. Every string the shipped programs format with %
// has to be one Starlark can fill, whether a run of it reaches that string or
// not, and whether it is written where it is used or kept in a constant at the
// top of the program.
func TestEveryFormatInTheProgramsIsOneStarlarkCanFill(t *testing.T) {
	t.Parallel()

	embedded := loaded(t)
	programs := map[string]string{}
	examples := embedded.Examples()
	for i := range examples {
		programs["examples/solvers/"+examples[i].ID+".star"] = examples[i].Solver
	}
	for _, topic := range embedded.Topics() {
		for _, template := range embedded.Templates(topic.ID) {
			programs["solvers/"+topic.ID+"/"+template.Name+".star"] = template.Program
		}
	}
	if len(programs) == 0 {
		t.Fatal("the content ships no programs, so nothing here was tested")
	}

	for _, name := range slices.Sorted(maps.Keys(programs)) {
		file, err := starlark.Parse(name, programs[name])
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		if problem := starlark.UnfillableFormat(file); problem != "" {
			t.Errorf("%s: %s, want every format one Starlark can fill with the values it is given", name, problem)
		}
	}
}
