package solver_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// task is a set of options whose answer is six, written three ways so that the
// tests can say which way they mean.
var task = solver.Options{"4", "5", "6", "8", "12"}

func TestLetterAndPlace(t *testing.T) {
	t.Parallel()
	for place := range solver.Count {
		if back := solver.Place(solver.Letter(place)); back != place {
			t.Errorf("Place(Letter(%d)) = %d, want %d", place, back, place)
		}
	}
	for _, letter := range []string{"F", "a", "", "AB"} {
		if place := solver.Place(letter); place != -1 {
			t.Errorf("Place(%q) = %d, want -1", letter, place)
		}
	}
}

func TestRotatedMovesEveryTextAndLeavesNoneBehind(t *testing.T) {
	t.Parallel()
	rotated := task.Rotated()

	for place := range solver.Count {
		if rotated[place] == task[place] {
			t.Errorf("place %d: got %q under the same letter, want every option moved", place, task[place])
		}
		letter := solver.Letter(place)
		if moved := solver.Relabelled(letter); rotated[solver.Place(moved)] != task[place] {
			t.Errorf("%s: its text is not under %s after the shift", letter, moved)
		}
	}
}

func TestRotatedFiveTimesIsWhereItStarted(t *testing.T) {
	t.Parallel()
	back := task
	for range solver.Count {
		back = back.Rotated()
	}
	if back != task {
		t.Errorf("five shifts: got %v, want %v", back, task)
	}
}

func TestRelabelledOfNothing(t *testing.T) {
	t.Parallel()
	if got := solver.Relabelled("F"); got != "" {
		t.Errorf("Relabelled(F) = %q, want an empty string", got)
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		options solver.Options
		value   string
		want    []string
	}{
		{"a number as it is written", task, "6", []string{"C"}},
		{"the same number with a decimal point", task, "6.0", []string{"C"}},
		{"the same number padded", task, "  6  ", []string{"C"}},
		{"the same number with a leading zero", task, "006", []string{"C"}},
		{"a number no option carries", task, "7", []string{}},
		{"a negative number", solver.Options{"-3", "3", "0", "1", "2"}, "-3.00", []string{"A"}},
		{
			"a number longer than a float can hold",
			solver.Options{"9007199254740993", "9007199254740992", "1", "2", "3"},
			"9007199254740993",
			[]string{"A"},
		},
		{
			"words, whatever their case and spacing",
			solver.Options{"1", "2", "  it IS impossible ", "3", "4"},
			"It is impossible",
			[]string{"C"},
		},
		{
			"a number does not match an option carrying a unit",
			solver.Options{"12 cm", "12", "13", "14", "15"},
			"12",
			[]string{"B"},
		},
		{
			"what looks like a fraction stays text",
			solver.Options{"24/7", "3", "4", "5", "6"},
			"3.4285714285714284",
			[]string{},
		},
		{
			"a fraction written the same way matches as text",
			solver.Options{"24/7", "3", "4", "5", "6"},
			" 24/7 ",
			[]string{"A"},
		},
		{"two options saying one thing", solver.Options{"6", "5", "6.0", "8", "12"}, "6", []string{"A", "C"}},
		{"nothing at all", task, "", []string{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.options.Match(test.value)
			if !slices.Equal(got, test.want) {
				t.Errorf("Match(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}

// FuzzMatch feeds the comparison what the model computed and what the model
// wrote: two strings from outside, neither of which has been held to a shape.
func FuzzMatch(f *testing.F) {
	f.Add("6", "6.0")
	f.Add("It is impossible", "it is impossible")
	f.Add("", "")
	f.Add("9007199254740993", "9007199254740992")
	f.Add("\x00", "0")

	f.Fuzz(func(t *testing.T, text, value string) {
		options := solver.Options{text, "", "x", text, value}
		matched := options.Match(value)

		if len(matched) > solver.Count {
			t.Fatalf("Match returned %d letters, and a task has %d options", len(matched), solver.Count)
		}
		if !slices.IsSorted(matched) {
			t.Fatalf("Match returned %v, want the letters in order", matched)
		}
		for i, letter := range matched {
			if solver.Place(letter) < 0 {
				t.Fatalf("Match returned %v, and %q labels no option", matched, letter)
			}
			if i > 0 && matched[i-1] == letter {
				t.Fatalf("Match returned %v, with %q twice", matched, letter)
			}
		}
		// The option that holds the value itself always matches it.
		if !slices.Contains(matched, solver.Letter(solver.Count-1)) {
			t.Fatalf("Match(%q) = %v, and the option is that very text", value, matched)
		}
	})
}
