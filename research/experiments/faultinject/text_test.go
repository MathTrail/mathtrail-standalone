package main

import (
	"slices"
	"strconv"
	"testing"
)

func TestSentencesEndAtAMarkFollowedBySpace(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, text string
		want       []string
	}{
		{"plain", "Ann has 3 cats. Ben has 1.5 litres! How many?", []string{"Ann has 3 cats.", "Ben has 1.5 litres!", "How many?"}},
		{"closing quote", "Ann says: 'Ben lies.' Ben says: 'No.' Who lies?", []string{"Ann says: 'Ben lies.'", "Ben says: 'No.'", "Who lies?"}},
		{"no mark at the end", "Pick one. Then stop", []string{"Pick one.", "Then stop"}},
		{"a point inside a number", "It costs 2.50 dollars.", []string{"It costs 2.50 dollars."}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := sentencesOf(test.text); !slices.Equal(got, test.want) {
				t.Errorf("sentencesOf(%q) = %q, want %q", test.text, got, test.want)
			}
		})
	}
}

func TestWordsHoldALetterOrADigit(t *testing.T) {
	t.Parallel()
	if got, want := wordsIn("1 + 3 + ... + 99 - done — yes"), 5; got != want {
		t.Errorf("wordsIn = %d, want %d", got, want)
	}
}

func TestBumpNumbersChangesWholeNumbersOnly(t *testing.T) {
	t.Parallel()
	got, changed := bumpNumbers("9 cats, 1.5 litres, 2,000 steps and 3:15.")
	if want := "10 cats, 1.5 litres, 2,000 steps and 4:16."; got != want || !changed {
		t.Errorf("bumpNumbers = %q, %v; want %q, true", got, changed, want)
	}
	if _, changed := bumpNumbers("No numbers here."); changed {
		t.Error("bumpNumbers changed a text with no number")
	}
}

func TestNumbersInWords(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text, want string
		ok         bool
	}{
		{"6", "six", true}, {"0", "zero", true}, {"20", "twenty", true},
		{"21", "", false}, {"06", "", false}, {"6 cm", "", false}, {"-1", "", false},
	} {
		if got, ok := inWords(test.text); got != test.want || ok != test.ok {
			t.Errorf("inWords(%q) = %q, %v; want %q, %v", test.text, got, ok, test.want, test.ok)
		}
	}
}

func TestSwapFirstWordKeepsTheCapitalAndDropsANegation(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text, want string
	}{
		{"Which is the smallest? Not the largest.", "Which is the largest? Not the largest."},
		{"Fewer than 3 coins.", "More than 3 coins."},
		{"He can NOT weigh it.", "He cannot NOT weigh it."},
		{"Which can he not weigh?", "Which cannot he not weigh?"},
		{"Which load is not possible?", "Which load is possible?"},
	} {
		if got, _ := swapFirstWord(test.text, oppositeWords, opposites); got != test.want {
			t.Errorf("swapFirstWord(%q) = %q, want %q", test.text, got, test.want)
		}
	}
}

func TestANewSettingRenamesEveryNameAndObjectAtOnce(t *testing.T) {
	t.Parallel()
	got, changed := inNewSetting("Ben gives Ann 3 coins. Ann keeps a coin.")
	if want := "Ann gives Kim 3 balls. Kim keeps a ball."; got != want || !changed {
		t.Errorf("inNewSetting = %q, %v; want %q, true", got, changed, want)
	}
}

func TestLinesGoIntoTheBodyOfSolve(t *testing.T) {
	t.Parallel()
	source := "X = 1\n\ndef solve(options):\n    # why\n    return match(options, X)\n"
	got, inserted := insertIntoSolve(source, []string{"for i in range(3):", "    pass"})
	want := "X = 1\n\ndef solve(options):\n    for i in range(3):\n        pass\n    # why\n    return match(options, X)\n"
	if got != want || !inserted {
		t.Errorf("insertIntoSolve = %q, %v; want %q, true", got, inserted, want)
	}
	if _, inserted := insertIntoSolve("def other(options):\n    return []\n", []string{"pass"}); inserted {
		t.Error("insertIntoSolve found a solve in a program without one")
	}
}

// A first letter is lowered whatever its script, and a text that does not
// begin with a character it can read is left as it is.
func TestLowerFirstLowersOnlyAFirstLetterItCanRead(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, text, want string }{
		{"a letter of the Latin script", "Ann has 3 coins.", "ann has 3 coins."},
		{"a letter with an accent", "Éva has 3 coins.", "éva has 3 coins."},
		{"no text", "", ""},
		{"a byte that is no character", "\xffAnn", "\xffAnn"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := lowerFirst(test.text); got != test.want {
				t.Errorf("lowerFirst(%q) = %q, want %q", test.text, got, test.want)
			}
		})
	}
}

// A replacement takes the capitals of the word it replaces: all of them for a
// word in capitals, the first for a word that has a capital first letter, and
// none otherwise; a word of one capital letter has a capital first letter.
func TestAReplacementTakesTheCapitalsOfTheWordItReplaces(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, replacement, original, want string }{
		{"a word in capitals", "about", "EXACTLY", "ABOUT"},
		{"a capital first letter", "about", "Exactly", "About"},
		{"no capital", "about", "exactly", "about"},
		{"one capital letter", "kim", "A", "Kim"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := withCaseOf(test.replacement, test.original); got != test.want {
				t.Errorf("withCaseOf(%q, %q) = %q, want %q", test.replacement, test.original, got, test.want)
			}
		})
	}
}

// A run of digits too long to be a whole number the operators can count with
// leaves the whole text alone, rather than changing the numbers before it and
// not the rest; the first whole number is the first that stands alone.
func TestANumberPastTheRangeLeavesTheTextAlone(t *testing.T) {
	t.Parallel()
	about := func(n int) string { return "about " + strconv.Itoa(n) }
	text := "3 cats and 99999999999999999999 dogs."
	if got, changed := bumpNumbers(text); got != text || changed {
		t.Errorf("bumpNumbers(%q) = %q, %v; want it unchanged, false", text, got, changed)
	}
	for _, test := range []struct {
		name, text, want string
		changed          bool
	}{
		{"a number past the range first", "99999999999999999999 dogs and 3 cats.", "99999999999999999999 dogs and 3 cats.", false},
		{"no number", "No number here.", "No number here.", false},
		{"a decimal before a whole number", "It takes 1.5 litres or 3 cups.", "It takes 1.5 litres or about 3 cups.", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got, changed := replaceFirstNumber(test.text, about); got != test.want || changed != test.changed {
				t.Errorf("replaceFirstNumber(%q) = %q, %v; want %q, %v", test.text, got, changed, test.want, test.changed)
			}
		})
	}
}

// A word the pattern finds and the map does not name is left as it was, and
// does not count as found.
func TestSwapWordsChangesOnlyTheWordsItKnows(t *testing.T) {
	t.Parallel()
	pattern, to := wordPattern([]string{"ann", "ben"}), map[string]string{"ann": "kim"}
	for _, test := range []struct {
		name, text, want string
		found            bool
	}{
		{"a word the map names", "Ann gives Ben a coin.", "Kim gives Ben a coin.", true},
		{"only a word it does not", "Ben keeps the coin.", "Ben keeps the coin.", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got, found := swapWords(test.text, pattern, to); got != test.want || found != test.found {
				t.Errorf("swapWords(%q) = %q, %v; want %q, %v", test.text, got, found, test.want, test.found)
			}
		})
	}
}
