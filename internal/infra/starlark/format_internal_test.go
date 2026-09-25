package starlark

import (
	"errors"
	"strings"
	"testing"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// The check finds each kind of format Starlark refuses, and none that it
// fills: a check that could never fail would pass every program it reads.
func TestAFormatStarlarkCannotFillIsFound(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		use     formatUse
		refused bool
	}{
		{"a stray percent sign", formatUse{text: "a rise of 10% on %d", values: 1}, true},
		{"fewer values than conversions", formatUse{text: "%d moves from %d", values: 1}, true},
		{"more values than conversions", formatUse{text: "%d moves", values: 2}, true},
		{"a key never closed", formatUse{text: "%(share d", values: -1}, true},
		{"a percent sign at the end", formatUse{text: "a rise of 10%", values: 0}, true},
		{"a letter that is no conversion", formatUse{text: "%q", values: 1}, true},
		{"a percent sign written twice", formatUse{text: "%d%%", values: 1}, false},
		{"conversions named by keys", formatUse{text: "%(share)d%(sign)%", values: -1}, false},
		{"values known only when it runs", formatUse{text: "%d from the row of %d", values: -1}, false},
		{"a key and values that are no dictionary", formatUse{text: "%(share)d", values: 1}, true},
		{"a key among conversions of the next value", formatUse{text: "%X%()X", values: 0}, true},
		{"a key, then a value by its place", formatUse{text: "%(share)d of %d", values: -1}, true},
		{"a number by its place beside a key", formatUse{text: "%d and %(share)d", values: -1}, true},
		{"two values by their places beside a key", formatUse{text: "%s, %s and %(share)d", values: -1}, true},
		{"the dictionary itself, then a key", formatUse{text: "%s holds %(share)d", values: -1}, false},
		{"as many values as conversions", formatUse{text: "%d : %d", values: 2}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problem := formatProblem(test.use)
			if (problem != "") != test.refused {
				t.Errorf("formatProblem(%q, %d values) = %q, want refused %v", test.use.text, test.use.values, problem, test.refused)
			}
		})
	}
}

// The check and the interpreter agree on every format: what the check refuses,
// Starlark cannot fill with the values it is given, and what the check lets
// through with a tuple of values, Starlark fills. A refusal the interpreter
// would not make would blame a run's failure on a format that did not cause
// it. A format with a key is filled from a dictionary, so the values are
// sometimes one that has every key, the way a name holding a dictionary is
// given, and then the check does not count values it cannot see: its refusals
// are held to the interpreter, and so is what it lets through with a key in it,
// since only a dictionary fills a key and this one fills every key there is.
func FuzzTheFormatCheckAgreesWithStarlark(f *testing.F) {
	for _, seed := range []string{
		"%d", "a rise of 10% on %d", "%d%%", "%3d", "%.1f", "%", "%(share)d", "%(share)%", "%(share",
		"%d : %d", "", "%c", "%%%", "%é", "%%(share)d", "%(share)s of %s", "%s holds %(share)d", "%d and %(share)d",
	} {
		f.Add(seed, uint8(1))
		f.Add(seed, uint8(4))
	}
	f.Fuzz(func(t *testing.T, text string, count uint8) {
		values, given := int(count%5), starlark.Value(everyKey{})
		if values == 4 {
			values = -1 // a name, whose values show only when the program runs
		} else {
			tuple := make(starlark.Tuple, values)
			for i := range tuple {
				tuple[i] = starlark.MakeInt(1)
			}
			given = tuple
		}
		problem := formatProblem(formatUse{text: text, values: values})
		_, err := starlark.Binary(syntax.PERCENT, starlark.String(text), given)

		switch {
		case problem != "" && err == nil:
			t.Fatalf("formatProblem(%q, %d values) = %q, and Starlark fills it", text, values, problem)
		case problem == "" && err != nil && values >= 0:
			t.Fatalf("formatProblem(%q, %d values) lets it through, and Starlark refuses it: %v", text, values, err)
		case problem == "" && err != nil && values < 0 && keyed(text):
			t.Fatalf("formatProblem(%q, a dictionary) lets it through, and Starlark refuses it: %v", text, err)
		}
	})
}

// everyKey is a dictionary that holds 1 under every key: what a format with
// keys is filled from when none of its keys is missing.
type everyKey struct{}

func (everyKey) String() string        { return "everyKey" }
func (everyKey) Type() string          { return "everyKey" }
func (everyKey) Freeze()               {}
func (everyKey) Truth() starlark.Bool  { return starlark.True }
func (everyKey) Hash() (uint32, error) { return 0, errors.New("everyKey is not hashable") }

func (everyKey) Get(starlark.Value) (starlark.Value, bool, error) {
	return starlark.MakeInt(1), true, nil
}

// A problem names the conversion, never the key in front of it: the key is the
// program's own words, and may hold an option's.
func TestAProblemWithAKeyLeavesTheKeyOut(t *testing.T) {
	t.Parallel()

	problem := formatProblem(formatUse{text: "%(the answer is 36)q", values: -1})
	if problem == "" || strings.Contains(problem, "the answer is 36") {
		t.Errorf("formatProblem() = %q, want a refusal that leaves the key out", problem)
	}
}

// keyed says whether a format names a value by a key anywhere, which is when
// only a dictionary can fill it.
func keyed(text string) bool {
	for i := 0; i < len(text); i++ {
		if text[i] != '%' {
			continue
		}
		end, kind, problem := conversionAt(text, i)
		if problem != "" {
			return false
		}
		if kind == keyedConversion {
			return true
		}
		i = end
	}
	return false
}
