package starlark

import (
	"slices"
	"testing"

	"go.starlark.net/starlark"
)

// TestTheVocabularyIsWhatItSays holds the table of predeclared names to the
// list a solver is promised, in both directions: nothing a written page
// mentions is missing, and nothing is there that no page mentions. It also
// catches the one mistake a table of literals invites — a name on the left
// that does not match the name the built-in answers to, which would leave a
// helper nobody could call by the name it was given.
func TestTheVocabularyIsWhatItSays(t *testing.T) {
	t.Parallel()
	promised := []string{
		// What a brute force is made of.
		"permutations", "combinations", "combinations_with_replacement", "product",
		"sum", "prod", "gcd",
		// What a task about dates needs.
		"is_leap", "days_in_month", "weekday", "add_days", "days_between",
		// What turns a computed answer into the letters of the task.
		"match",
		// The language's own, declared again with a ceiling on them.
		"all", "any", "dict", "enumerate", "list",
		"max", "min", "reversed", "set", "sorted", "tuple", "zip",
	}

	declared, err := vocabulary(1_000_000)
	if err != nil {
		t.Fatalf("vocabulary: %v", err)
	}

	for _, name := range promised {
		value, found := declared[name]
		if !found {
			t.Errorf("%s is promised and not declared", name)
			continue
		}
		builtin, isBuiltin := value.(*starlark.Builtin)
		if !isBuiltin {
			t.Errorf("%s is a %s, want something a solver can call", name, value.Type())
			continue
		}
		if builtin.Name() != name {
			t.Errorf("%s answers to %q, want its own name", name, builtin.Name())
		}
	}

	for name := range declared {
		if !slices.Contains(promised, name) {
			t.Errorf("%s is declared and promised to nobody", name)
		}
	}
}
