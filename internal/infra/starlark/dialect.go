package starlark

import (
	"fmt"

	"go.starlark.net/syntax"
)

// dialect is how much of Python this Starlark is allowed to be.
//
// Models write Python, so every switch here is a question of whether the
// difference is worth an error the model cannot see coming. Four are: sets are
// how a search remembers where it has been; a breadth-first search needs a
// loop whose length is not known in advance, and iterating a list while
// appending to it is an error in this language; a model writes a script rather
// than a module, so control flow at the top level should not be a parse error;
// and reassigning a top-level name is ordinary Python that forbidding buys
// nothing.
//
// One is not. Recursion reads backwards — the field switches off the check for
// it, not the recursion — and the check stays on. A recursive Starlark
// function recurses on the Go stack, and a stack overflow there cannot be
// recovered: it would take the whole process down rather than the one request.
// Search is written iteratively instead, and a recurrence is filled in from
// the bottom.
//
// The loader is switched off for completeness only: there is no module loader
// to bind names from, so a program that asks for one is refused before it runs.
var dialect = &syntax.FileOptions{
	Set:               true,
	While:             true,
	TopLevelControl:   true,
	GlobalReassign:    true,
	Recursion:         false,
	LoadBindsGlobally: false,
}

// parse reads a program in the dialect a solver runs in, without binding its
// names or running it. A program is read the way a run reads it by compile,
// which begins here.
func parse(name, source string) (*syntax.File, error) {
	file, err := dialect.Parse(name, source, 0)
	if err != nil {
		return nil, fmt.Errorf("starlark: parse %s: %w", name, err)
	}
	return file, nil
}
