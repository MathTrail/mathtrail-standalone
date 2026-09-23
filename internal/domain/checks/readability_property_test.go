package checks

import (
	"strings"
	"testing"
	"unicode"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"golang.org/x/text/unicode/norm"
)

// The properties here name what measuring a question has to hold for any
// text, not only for the reference tasks. Splitting it into sentences loses
// nothing and doubles nothing, and a sentence added never makes the longest
// one shorter — the check could otherwise be passed by writing more. An accent
// never takes a syllable away: it tells how a letter is said, and never makes
// a vowel something else.

// genText is a question made of words, the marks that end and close sentences,
// and the spaces between them.
func genText() gopter.Gen {
	piece := gen.OneConstOf("cat", "dog", "3.5", "2", "+", "-", "…", ".", "!", "?", "'", ")", "。", "猫", "؟", "।")
	space := gen.OneConstOf(" ", "", "  ", "\n")
	return gen.SliceOfN(30, gopter.CombineGens(piece, space).Map(func(parts []any) string {
		word, _ := parts[0].(string)
		gap, _ := parts[1].(string)
		return word + gap
	})).Map(func(pieces []string) string { return strings.Join(pieces, "") })
}

// withoutSpaces is a text with its whitespace taken out.
func withoutSpaces(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, text)
}

// longestOf is the length of the longest sentence of a text, in its unit.
func longestOf(text string) int {
	longest := 0
	for _, sentence := range sentences(text) {
		length, _ := lengthOf(sentence)
		longest = max(longest, length)
	}
	return longest
}

func TestSplittingHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	first, second := genText(), genText()

	properties.Property("nothing of a text is lost or doubled by splitting it", prop.ForAll(
		func(text string) bool {
			return withoutSpaces(strings.Join(sentences(text), "")) == withoutSpaces(text)
		},
		first,
	))

	properties.Property("a sentence added never makes the longest one shorter", prop.ForAll(
		func(text, more string) bool {
			return longestOf(text+" "+more) >= longestOf(text)
		},
		first, second,
	))

	properties.TestingRun(t)
}

// accentedWord is a word written plain, and the same word with accents on
// some of its letters.
type accentedWord struct{ plain, accented string }

// genAccentedWord is a word of the letters the syllable rules look at, with an
// acute, a grave, a circumflex or a diaeresis on some of them — written into
// the letter wherever Unicode has the two as one character.
func genAccentedWord() gopter.Gen {
	letter := gen.OneConstOf("a", "e", "i", "o", "u", "y", "b", "c", "d", "h", "l", "q", "r", "s", "t", "x")
	accent := gen.OneConstOf("", "", "", "\u0301", "\u0300", "\u0302", "\u0308")
	return gen.SliceOf(gopter.CombineGens(letter, accent)).Map(func(pairs [][]any) accentedWord {
		var plain, accented strings.Builder
		for _, pair := range pairs {
			letter, _ := pair[0].(string)
			mark, _ := pair[1].(string)
			plain.WriteString(letter)
			accented.WriteString(letter + mark)
		}
		return accentedWord{plain: plain.String(), accented: norm.NFC.String(accented.String())}
	})
}

func TestSyllablesHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("an accent never takes a syllable away", prop.ForAll(
		func(word accentedWord) bool {
			return syllablesOf(word.accented) >= syllablesOf(word.plain)
		},
		genAccentedWord(),
	))

	properties.TestingRun(t)
}
