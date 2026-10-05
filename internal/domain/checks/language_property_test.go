package checks_test

import (
	"slices"
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

// genLatinWords are words in Latin letters that are read as words: four letters
// or more, none of them capitals, so that none is a symbol such as ab or cm.
func genLatinWords() gopter.Gen { return gen.SliceOf(gen.RegexMatch(`[a-z]{4,8}`)) }

// genCyrillicWords are words in Cyrillic letters.
func genCyrillicWords() gopter.Gen { return gen.SliceOf(gen.RegexMatch(`[a-h]{1,7}`)).Map(cyrillic) }

// some are words of a generator, at least one of them.
func some(words gopter.Gen) gopter.Gen {
	return words.SuchThat(func(words []string) bool { return len(words) > 0 })
}

// genSymbols are what a task writes among its words and a child reads in no
// language: labels in Latin capitals, numbers, single Latin and Greek letters,
// units of measure, signs, full-width capitals and the mark of a long vowel.
// They come from one list rather than from several generators, since a slice
// of gopter's is held to the sieve of its first element, and the sieves of
// several generators would throw most slices away.
func genSymbols() gopter.Gen {
	return gen.SliceOf(gen.OneConstOf(
		"AB", "XYZ", "Q", "C1", "12", "7", "2026", "x", "n", "I", "π", "α", "cm", "kg", "min", "+", "=", "·", "?",
		"ＡＢ", "ー",
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
			genCyrillicWords(), genSymbols(),
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
			some(genLatinWords()), some(genCyrillicWords()), genSymbols(),
		))

	properties.Property("a word in the lesson's letters never turns a pass into a refusal, "+
		"nor one in other letters a refusal into a pass",
		prop.ForAll(
			func(own, other, oneOwn, oneOther []string) bool {
				text := strings.Join(append(slices.Clone(own), other...), " ")
				before := refusedIn(asking(text), "ru")
				return (before || !refusedIn(asking(text+" "+oneOwn[0]), "ru")) &&
					(!before || refusedIn(asking(text+" "+oneOther[0]), "ru"))
			},
			genCyrillicWords(), genLatinWords(), some(genCyrillicWords()), some(genLatinWords()),
		))

	properties.TestingRun(t)
}
