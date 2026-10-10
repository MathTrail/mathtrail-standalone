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

	cyrillic, greek := letterings["Cyrl"].letters, letterings["Grek"].letters
	for _, test := range []struct {
		name     string
		text     string
		letters  []*unicode.RangeTable
		own, all int
	}{
		{"Chinese naming two children in Latin letters", "Tom和Mary谁高？", chinese.letters, 3, 5},
		{"Greek articles and its or, words of a single letter", "Ο Νίκος ή η Μαρία", greek, 5, 5},
		{"a single Greek letter among the words of another lesson", "Найди π и α.", cyrillic, 2, 2},
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

// A tag holds a task to the letters its language is written in: those of the
// script it names, of the one its language is written in as a rule, of any a
// language of the cards is written in when it is written in several, with the
// one its place adds — and to none where the tag library only guesses at the
// script of a language the cards do not speak, or the tag names nothing.
// Chinese is held at a guess, since its two likely scripts are one set of
// characters.
func TestATagHoldsATaskToTheLettersItsLanguageIsWrittenIn(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ tag, called string }{
		{"ru", "Cyrillic letters"}, {"ru-RU", "Cyrillic letters"}, {"en-GB", "Latin letters"},
		{"ja", "kanji and kana"}, {"ko", "Hangul"}, {"zh-Hans", "Chinese characters"}, {"zh", "Chinese characters"},
		{"zh-TW", "Chinese characters"}, {"yue", "Chinese characters"}, {"cmn", "Chinese characters"},
		{"el", "Greek letters"}, {"he", "Hebrew letters"}, {"iw", "Hebrew letters"}, {"am", "Ethiopic letters"},
		{"hy", "Armenian letters"}, {"ka", "Georgian letters"}, {"te", "Telugu letters"}, {"ta", "Tamil letters"},
		{"gu", "Gujarati letters"}, {"kn", "Kannada letters"}, {"ml", "Malayalam letters"}, {"or", "Odia letters"},
		{"mr", "Devanagari letters"}, {"mai", "Devanagari letters"}, {"bg", "Cyrillic letters"}, {"ps", "Arabic letters"},
		{"zh-Hant", "Chinese characters"}, {"sw", "Latin letters"}, {"zu", "Latin letters"}, {"nb", "Latin letters"},
		{"no", "Latin letters"},
		{"pa", "Gurmukhi or Arabic letters"}, {"pa-PK", "Gurmukhi or Arabic letters"},
		{"ms", "Latin or Arabic letters"}, {"ha", "Latin or Arabic letters"}, {"ku", "Latin or Arabic letters"},
		{"sr", "Cyrillic or Latin letters"}, {"kk", "Cyrillic or Latin letters"}, {"bs", "Latin or Cyrillic letters"},
		{"uz", "Latin or Cyrillic letters"}, {"az", "Latin or Cyrillic letters"},
		{"kk-CN", "Cyrillic, Latin or Arabic letters"}, {"uz-AF", "Latin, Cyrillic or Arabic letters"},
		{"az-IR", "Latin, Cyrillic or Arabic letters"},
		{"ckb", "Arabic letters"}, {"fil", "Latin letters"}, {"tl", "Latin letters"}, {"yo", "Latin letters"},
		{"ig", "Latin letters"},
		{"sr-Latn", "Latin letters"}, {"sr-Cyrl", "Cyrillic letters"}, {"kk-Cyrl", "Cyrillic letters"},
		{"pa-Arab", "Arabic letters"}, {"ku-Arab", "Arabic letters"}, {"el-Latn", "Latin letters"},
		{"mn", ""}, {"und", ""}, {"mul", ""}, {"x-abc", ""}, {"", ""}, {"not a tag", ""},
	} {
		t.Run(test.tag, func(t *testing.T) {
			t.Parallel()

			lesson, held := lessonLetters(test.tag)
			if called := lesson.called; held != (test.called != "") || called != test.called {
				t.Errorf("lessonLetters(%q) = %q, held %v, want %q", test.tag, called, held, test.called)
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

// Every script a language of the cards is usually written in is one the check
// has letters for: one it lacked would narrow the language's letters, and a
// task written in that script would be refused.
func TestEveryUsualScriptHasLetters(t *testing.T) {
	t.Parallel()

	for language, scripts := range usualScripts {
		for _, script := range scripts {
			if _, known := letterings[script]; !known {
				t.Errorf("usualScripts[%q] holds %q, which letterings lacks", language, script)
			}
		}
	}
}
