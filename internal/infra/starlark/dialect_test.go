package starlark_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// A program is read in the dialect a solver runs in, so whatever inspects a
// program reads what the sandbox would: the search that uses the whole of the
// dialect parses, and a program that does not parse is refused by its name.
func TestParseReadsAProgramInTheDialectASolverRunsIn(t *testing.T) {
	t.Parallel()

	file, err := starlark.Parse("pouring.star", pouring)
	if err != nil {
		t.Fatalf("Parse(pouring) error = %v, want the program read", err)
	}
	if len(file.Stmts) == 0 {
		t.Errorf("Parse(pouring) read no statements, want the program's own")
	}
	if _, err := starlark.Parse("broken.star", "def solve(:\n"); err == nil || !strings.Contains(err.Error(), "broken.star") {
		t.Errorf("Parse(broken) error = %v, want one naming broken.star", err)
	}
}
