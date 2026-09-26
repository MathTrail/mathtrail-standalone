package checks_test

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// The properties here name what the measure has to hold for any question, not
// only for the pairs Postgres was asked about: it does not care which side a
// question is on, a question is wholly alike itself, the case of its letters
// and what stands between its words change nothing, and two alphabets that
// share no letter share nothing.

// genWords is a question as its words, from a small alphabet so that two
// questions share some of them.
func genWords() gopter.Gen {
	return gen.SliceOfN(12, gen.RegexMatch(`[a-f]{1,6}`)).SuchThat(func(words []string) bool { return len(words) > 0 })
}

// genSeparator is what a writer might put between two words.
func genSeparator() gopter.Gen { return gen.OneConstOf(" ", "  ", ", ", " - ", "; ", "!\n", "'") }

func TestTheMeasureHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	// Two generators rather than one used twice, so that each argument reads as
	// the side it stands for.
	one, other := genWords(), genWords()

	properties.Property("alike is the same whichever side a question is on", prop.ForAll(
		func(a, b []string) bool {
			first, second := strings.Join(a, " "), strings.Join(b, " ")
			return checks.Similarity(first, second) == checks.Similarity(second, first)
		},
		one, other,
	))

	properties.Property("a question is wholly alike itself", prop.ForAll(
		func(words []string) bool {
			question := strings.Join(words, " ")
			return checks.Similarity(question, question) == 1
		},
		genWords(),
	))

	properties.Property("the case of the letters and what separates the words change nothing", prop.ForAll(
		func(words []string, separator string, shout bool) bool {
			plain := strings.Join(words, " ")
			written := strings.Join(words, separator)
			if shout {
				written = strings.ToUpper(written)
			}
			return checks.Similarity(plain, written) == 1
		},
		genWords(), genSeparator(), gen.Bool(),
	))

	properties.Property("alphabets that share no letter share nothing", prop.ForAll(
		func(latin, other []string) bool {
			return checks.Similarity(strings.Join(latin, " "), strings.Join(cyrillic(other), " ")) == 0
		},
		one, other,
	))

	properties.Property("a question copies itself and repeats itself", prop.ForAll(
		func(words []string) bool {
			question := strings.Join(words, " ")
			return len(checks.NearDuplicate(question, "", []string{question}, []string{checks.Fingerprint(question, "")})) == 2
		},
		genWords(),
	))

	properties.TestingRun(t)
}

// cyrillic writes words of the first letters of the Latin alphabet in the
// first letters of the Cyrillic one.
func cyrillic(words []string) []string {
	written := make([]string, len(words))
	for i, word := range words {
		written[i] = strings.Map(func(r rune) rune { return 'а' + (r - 'a') }, word)
	}
	return written
}
