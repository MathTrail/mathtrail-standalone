package checks_test

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// The properties of the language check: what is read in no language never
// turns a text in the lesson's letters into a refusal, a text with none of
// them is refused in every lesson written in others, and a word added in the
// lesson's letters never turns a pass into a refusal, nor one in other letters
// a refusal into a pass.

// genLatinWord is a word in Latin letters that is read as a word: two letters
// or more, and not in capitals.
func genLatinWord() gopter.Gen { return gen.RegexMatch(`[a-z]{2,8}`) }

// genCyrillicWord is a word in Cyrillic letters.
func genCyrillicWord() gopter.Gen {
	return gen.RegexMatch(`[a-h]{1,7}`).Map(func(word string) string { return cyrillic([]string{word})[0] })
}

// genWordsIn are a few words of one generator, at least one.
func genWordsIn(word gopter.Gen) gopter.Gen {
	return gen.SliceOfN(8, word).SuchThat(func(words []string) bool { return len(words) > 0 })
}

// genSymbols are what a task writes among its words and a child reads in no
// language: labels in Latin capitals, numbers, single Latin and Greek letters,
// signs, full-width capitals and the mark of a long vowel.
func genSymbols() gopter.Gen {
	return gen.SliceOfN(8, gen.OneGenOf(
		gen.RegexMatch(`[A-Z]{1,4}`), gen.RegexMatch(`[0-9]{1,3}`), gen.RegexMatch(`[a-z]`),
		gen.OneConstOf("π", "α", "+", "=", "·", "?", "ＡＢ", "ー"),
	))
}

// among writes symbols among words, each apart from the words beside it, so
// that none of them joins a word.
func among(words, symbols []string) string {
	var written []string
	for i, word := range words {
		written = append(written, word)
		if i < len(symbols) {
			written = append(written, symbols[i])
		}
	}
	return strings.Join(written, " ")
}

// lessonsInOtherLetters are languages of the cards written in letters other
// than Latin.
var lessonsInOtherLetters = []string{"ru", "uk", "ar", "fa", "ur", "hi", "bn", "th", "ja", "zh-Hans", "ko"}

func TestTheLanguageCheckHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("what is read in no language never turns a text in the lesson's letters into a refusal",
		prop.ForAll(
			func(words, symbols []string) bool {
				return !refusedIn(asking(among(words, symbols)), "ru")
			},
			genWordsIn(genCyrillicWord()), genSymbols(),
		))

	properties.Property("a text with none of the lesson's letters is refused in every lesson written in others",
		prop.ForAll(
			func(latin, cyrillicWords, symbols []string) bool {
				for _, language := range lessonsInOtherLetters {
					if !refusedIn(asking(among(latin, symbols)), language) {
						return false
					}
				}
				return refusedIn(asking(among(cyrillicWords, symbols)), "en")
			},
			genWordsIn(genLatinWord()), genWordsIn(genCyrillicWord()), genSymbols(),
		))

	properties.Property("a word in the lesson's letters never turns a pass into a refusal, "+
		"nor one in other letters a refusal into a pass",
		prop.ForAll(
			func(words []string, own, other string) bool {
				text := strings.Join(words, " ")
				before := refusedIn(asking(text), "ru")
				return (before || !refusedIn(asking(text+" "+own), "ru")) &&
					(!before || refusedIn(asking(text+" "+other), "ru"))
			},
			genWordsIn(gen.OneGenOf(genCyrillicWord(), genLatinWord())), genCyrillicWord(), genLatinWord(),
		))

	properties.TestingRun(t)
}
