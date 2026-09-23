package checks

import (
	"testing"

	"golang.org/x/text/unicode/norm"
)

// English spells a few words with accents, and an accent says how its letter
// is said: the vowel under it is still a vowel, a diaeresis says it apart from
// the vowel before it, and an accented e is said where a plain one would be
// silent. An accent written after its letter, as a mark of its own, is the
// same accent.
func TestAnAccentSaysHowALetterIsSaid(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		word      string
		syllables int
		because   string
	}{
		{"déjà", 2, "an accented vowel is a vowel"},
		{"über", 2, "an accented vowel is a vowel"},
		{"café", 2, "an accented e is said"},
		{"José", 2, "an accented e is said"},
		{"résumé", 3, "an accented e is said"},
		{"résumés", 3, "an accented e is said before an s"},
		{"clichéd", 2, "an accented e is said before a d"},
		{"naïve", 2, "a diaeresis says a vowel apart"},
		{"coöperate", 4, "a diaeresis says a vowel apart"},
		{"Zoë", 2, "a diaeresis says a vowel apart, and the e it stands on is said"},
		{"cafe\u0301", 2, "an accent written after its letter is the same accent"},
		{"nai\u0308ve", 2, "a diaeresis written after its letter is the same diaeresis"},
	} {
		t.Run(test.word, func(t *testing.T) {
			t.Parallel()

			if got := syllablesOf(test.word); got != test.syllables {
				t.Errorf("syllablesOf(%q) = %d, want %d: %s", test.word, got, test.syllables, test.because)
			}
		})
	}
}

// A text reads at the same grade however it is written: in capitals, where
// the apostrophe of a contraction stands before a capital letter, or with its
// accents written after their letters.
func TestTheGradeDoesNotDependOnHowTheTextIsWritten(t *testing.T) {
	t.Parallel()

	accented := "José buys a café crème and a crêpe."
	for _, test := range []struct{ name, text, same string }{
		{"in capitals", "DON'T COUNT THE BUTTERFLIES TWICE, IT'S EASY.", "don't count the butterflies twice, it's easy."},
		{"accents after their letters", norm.NFD.String(accented), norm.NFC.String(accented)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got, want := fleschKincaid(test.text), fleschKincaid(test.same); got != want {
				t.Errorf("fleschKincaid(%q) = %.2f, want %.2f as for %q", test.text, got, want, test.same)
			}
		})
	}
}
