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
			"words, whatever the gaps between them",
			solver.Options{"1", "2", "it  is\u00a0impossible", "3", "4"},
			"It is impossible",
			[]string{"C"},
		},
		{
			"words, whatever takes no room between them",
			solver.Options{"1", "2", "it is\u200b impossible", "3", "4"},
			"It is impossible",
			[]string{"C"},
		},
		{
			"an accent, whichever way it was put together",
			solver.Options{"1", "2", "caf\u00e9", "3", "4"},
			"cafe\u0301",
			[]string{"C"},
		},
		{
			"a final sigma, whatever the case of the word",
			solver.Options{"1", "2", "\u03ad\u03be\u03b9\u03c2", "3", "4"},
			"\u0388\u039e\u0399\u03a3",
			[]string{"C"},
		},
		{
			"not a letter that joins its neighbours differently",
			solver.Options{"\u0645\u06cc\u200c\u0631\u0648\u0645", "\u0645\u06cc\u0631\u0648\u0645", "1", "2", "3"},
			"\u0645\u06cc\u0631\u0648\u0645",
			[]string{"B"},
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

// Two option texts are one answer when they are one number, or one text once
// folded, and never because one of them could be read as a number: they have
// one key.
func TestTwoTextsAreOneAnswerWhenTheyHaveOneKey(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		a, b string
		same bool
	}{
		{"a number with a point and without", "6", "6.0", true},
		{"a number with a sign and with zeros in front", "+6", "006", true},
		{"a zero with a sign and with a fraction", "-0", "0.00", true},
		{"a fraction with a zero behind", "6.10", "6.1", true},
		{"words in any case, with any gaps", "Six pairs", "  six   PAIRS ", true},
		{"a sigma at the end of a word and inside it", "\u03a3\u03c3", "\u03c3\u03c2", true},
		{"a Cherokee capital and its small letter", "\u13c8", "\uab98", true},
		{"a number with a zero-width space", "6.0\u200b", "6", true},
		{"words with a filler that draws nothing", "six\u3164 pairs", "six pairs", true},
		{"a number with a mark of direction", "\u200f6", "6", true},
		{"a number written wide", "\uff16", "6", true},
		{"a word written wide", "\uff33\uff29\uff38", "six", true},
		{"numbers an embedding shows in another order", "\u202b12 34\u202c", "12 34", false},
		{"a number shown reversed by an override", "\u202e12\u202c", "12", false},
		{"a dotless i of Turkish and an i", "s\u0131k", "sik", false},
		{"a number and a number of apples", "6", "6 apples", false},
		{"a fraction written with a slash and as a decimal", "1/2", "0.5", false},
		{"a word and a longer word", "six", "sixth", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := solver.Key(test.a) == solver.Key(test.b); got != test.same {
				t.Errorf("Key(%+q) = %+q and Key(%+q) = %+q, want one key %v",
					test.a, solver.Key(test.a), test.b, solver.Key(test.b), test.same)
			}
		})
	}
}

func TestLettersAreEveryLabelInOrder(t *testing.T) {
	t.Parallel()
	letters := solver.Letters()
	for place := range solver.Count {
		if letters[place] != solver.Letter(place) {
			t.Errorf("Letters()[%d] = %q, want %q", place, letters[place], solver.Letter(place))
		}
	}
	letters[0] = "Z"
	if solver.Letters()[0] != "A" {
		t.Errorf("Letters() handed out a slice another caller changed")
	}
}

// Two texts have one key when a card shows them alike, and only then: whatever
// the case of the letters, however an accent was put together, whatever takes
// no room between them — and not when a character changes how the letters
// around it join.
func TestAKeyReadsATextAsACardShowsIt(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		a, b string
		same bool
	}{
		{"an iota after a letter, whatever its case", "\u03b1\u03b9", "\u0391\u0399", true},
		{"a long s with a dot below and an s with one", "\u017f\u0323", "\u1e63", true},
		{"a capital and a small alpha with a breathing", "\u1f08", "\u1f00", true},
		{"a Cherokee capital and its small letter", "\u13a0", "\uab70", true},
		{"an accent kept apart by a grapheme joiner", "e\u034f\u0301", "\u00e9", true},
		{"a product written with an invisible times", "2\u20623", "23", true},
		{"a number with the marks of direction around it", "\u200f6\u200e", "6", true},
		{"an iota written under its letter or beside it", "\u1fb3", "\u03b1\u03b9", false},
		{"a dotless i of Turkish and an i", "s\u0131k", "sik", false},
		{"a dotted capital I of Turkish and a capital I", "\u0130", "I", false},
		{"letters kept apart by a joiner of Persian", "\u0645\u06cc\u200c\u0631", "\u0645\u06cc\u0631", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := solver.Key(test.a) == solver.Key(test.b); got != test.same {
				t.Errorf("Key(%+q) = %+q, Key(%+q) = %+q, want alike %v",
					test.a, solver.Key(test.a), test.b, solver.Key(test.b), test.same)
			}
			if again := solver.Key(solver.Key(test.a)); again != solver.Key(test.a) {
				t.Errorf("Key(Key(%+q)) = %+q, want %+q: a key is its own key",
					test.a, again, solver.Key(test.a))
			}
		})
	}
}

// A text is blank when a card would show nothing for it: an option of spaces
// and of characters that only steer the letters around them is an empty one,
// and one of the signs Arabic writes in front of a number is not.
func TestBlank(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		text  string
		blank bool
	}{
		{"nothing", "", true},
		{"spaces of two kinds", " \u00a0 ", true},
		{"a joiner of Persian", "\u200c", true},
		{"characters that take no room", "\u200b\u00ad\u200e", true},
		{"fillers that draw nothing", "\u3164\u115f\u1160\uffa0\u2800", true},
		{"an accent on nothing", "\u0301", false},
		{"the sign of the end of a verse", "\u06dd", false},
		{"a digit", "0", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := solver.Blank(test.text); got != test.blank {
				t.Errorf("Blank(%+q) = %v, want %v", test.text, got, test.blank)
			}
		})
	}
}

// A key is how the options of a task written by the model are told apart, and
// how the value its solver computed is found among them. Whatever a text
// holds, its key is its own key: a key read again, or a text written in the
// form its key has, gives the same answer as the text it came from.
func FuzzKey(f *testing.F) {
	for _, seed := range []string{
		"6", "6.0", "6.0\u200b", "-0.00", "+007", "\u03b1\u03b9", "\u1fb3", "\u017f\u0323", "S\u0323",
		"It is  impossible", "\u202b12 34\u202c", "s\u0131k", "",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, text string) {
		if key := solver.Key(text); solver.Key(key) != key {
			t.Fatalf("Key(Key(%+q)) = %+q, want %+q", text, solver.Key(key), key)
		}
	})
}
