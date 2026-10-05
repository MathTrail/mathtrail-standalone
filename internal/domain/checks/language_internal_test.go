package checks

import (
	"testing"
	"unicode"
)

// A text is counted as a child reads it: a word at a time, a character at a
// time where a script has no spaces, a mark with its letter, and nothing of
// what is read in no language.
func TestATextIsCountedAsAChildReadsIt(t *testing.T) {
	t.Parallel()

	cyrillic := letterings["Cyrl"].letters
	for _, test := range []struct {
		name     string
		text     string
		letters  []*unicode.RangeTable
		own, all int
	}{
		{"Chinese naming two children in Latin letters", "Tom和Mary谁高？", chinese.letters, 3, 5},
		{"points named in Latin capitals", "AB = 3, BC = 4. Найди AC.", cyrillic, 1, 1},
		{"an accent written as a mark of its own, inside its word", "nai\u0308ve", letterings["Latn"].letters, 1, 1},
		{"a vowel sign and a virama inside a word", "नमस्ते", letterings["Deva"].letters, 1, 1},
		{"the mark of a long vowel, which every script shares", "ケーキ", letterings["Jpan"].letters, 2, 2},
		{"full-width capitals", "ＡＢ", cyrillic, 0, 0},
		{"a number, a Greek letter and a Latin one", "2πr", cyrillic, 0, 0},
		{"a Latin name with a Cyrillic ending", "Tomу", cyrillic, 1, 2},
		{"a Persian word held apart without a space", "می\u200cسازند", letterings["Arab"].letters, 2, 2},
		{"English, its short words left out", "Three robots sit", cyrillic, 0, 2},
		{"a product and the orders of three letters", "Площадь равна ab. Порядки: abc, acb, bac.", cyrillic, 3, 3},
		{"single letters", "x I a", cyrillic, 0, 0},
		{"units of measure among Japanese characters", "5cmと3cm", letterings["Jpan"].letters, 1, 1},
		{"nothing", "", cyrillic, 0, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if own, all := counted(test.text, test.letters); own != test.own || all != test.all {
				t.Errorf("counted(%q) = %d of %d, want %d of %d", test.text, own, all, test.own, test.all)
			}
		})
	}
}

// A tag holds a task to a script only where the tag library is sure of it: a
// script the tag names, a language written in one script as a rule, or
// Chinese, whose two likely scripts are one set of characters.
func TestATagHoldsATaskToAScriptOnlyWhereTheLibraryIsSure(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		tag, script string
		held        bool
	}{
		{"ru", "Cyrl", true}, {"ru-RU", "Cyrl", true}, {"en-GB", "Latn", true}, {"ja", "Jpan", true},
		{"ko", "Kore", true}, {"zh-Hans", "Hans", true}, {"zh", "Hans", true}, {"zh-TW", "Hant", true},
		{"yue", "Hant", true}, {"cmn", "Hans", true}, {"sr-Latn", "Latn", true}, {"kk-Cyrl", "Cyrl", true},
		{"kk", "", false}, {"ms", "", false}, {"pa", "", false}, {"bs", "", false}, {"sr", "", false}, {"uz", "", false}, {"az", "", false}, {"mn", "", false}, {"pa-PK", "", false},
		{"kk-CN", "", false}, {"und", "", false}, {"mul", "", false}, {"x-abc", "", false}, {"", "", false},
		{"not a tag", "", false},
	} {
		t.Run(test.tag, func(t *testing.T) {
			t.Parallel()

			if script, held := heldScript(test.tag); script != test.script || held != test.held {
				t.Errorf("heldScript(%q) = %q, %v, want %q, %v", test.tag, script, held, test.script, test.held)
			}
		})
	}
}

// Every script a task can be held to has letters to hold it to and a name a
// refusal calls them by: a script with no letters would refuse every text.
func TestEveryLetteringHasLettersAndAName(t *testing.T) {
	t.Parallel()

	for script, lettering := range letterings {
		if len(lettering.letters) == 0 || lettering.called == "" {
			t.Errorf("letterings[%q] = %+v, want letters and a name", script, lettering)
		}
	}
}
