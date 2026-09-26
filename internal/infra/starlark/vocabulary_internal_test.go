package starlark

import (
	"slices"
	"strings"
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

// Only a built-in of the language can be given a cap. A name the language does
// not have, and a name it has for something that is no built-in, are refused
// by name rather than wrapped into something that fails when a solver calls it.
func TestOnlyABuiltinCanBeBounded(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"walk_everything", "None"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if capped, err := bounded(name, 1_000_000); err == nil || !strings.Contains(err.Error(), name) {
				t.Errorf("bounded(%q) = %v, %v; want a refusal naming it", name, capped, err)
			}
		})
	}
}

// A walker that cannot be bounded stops the vocabulary from being built at
// all. Built without it, the vocabulary would leave the language's own in
// reach of a solver, which is the one with no cap on it.
func TestAWalkerThatCannotBeBoundedStopsTheVocabulary(t *testing.T) {
	// Not parallel: it changes the list every vocabulary is built from, and
	// puts it back before any test that runs in parallel starts.
	kept := walkers
	walkers = append(slices.Clip(kept), "None")
	t.Cleanup(func() { walkers = kept })

	if declared, err := vocabulary(1_000_000); err == nil || !strings.Contains(err.Error(), "None") {
		t.Errorf("vocabulary() = %d names, %v; want a refusal naming None", len(declared), err)
	}
}

// A thread no run belongs to has no clock of its own, and a helper called on
// one works to its end rather than failing for the want of one: it is never
// told to stop, because nothing is waiting for it.
func TestAThreadOfNoRunIsNeverStopped(t *testing.T) {
	t.Parallel()

	// Long enough that the walk looks at the clock more than once, as a long
	// sum does.
	numbers := rangeOf(4 * clockEvery)
	total, err := sum(1_000_000).CallInternal(&starlark.Thread{}, starlark.Tuple{numbers}, nil)
	if err != nil {
		t.Fatalf("sum on a thread of no run: got error %v, want none", err)
	}
	if want := starlark.MakeInt(numbers.Len() * (numbers.Len() - 1) / 2); !same(t, total, want) {
		t.Errorf("sum on a thread of no run = %v, want %v", total, want)
	}
}
