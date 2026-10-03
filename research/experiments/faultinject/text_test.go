package main

import (
	"slices"
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
