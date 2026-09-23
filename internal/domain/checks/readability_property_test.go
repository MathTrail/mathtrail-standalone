package checks

import (
	"strings"
	"testing"
	"unicode"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// The properties here name what splitting a question into sentences has to
// hold for any text, not only for the reference tasks: nothing of the text is
// lost or doubled on the way, and a sentence added never makes the longest one
// shorter — the check could otherwise be passed by writing more.

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
