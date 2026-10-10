package checks

import (
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Whatever word a colour is called by, a text names it where it holds the
// word's stem, whatever ending follows and whatever case it is written in,
// and a text that holds none of its letters never does.
func TestNamingAColourHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("a text holding the stem names the word, in any case and with any ending", prop.ForAll(
		func(word, ending string) bool {
			text := "Flags " + strings.ToUpper(stemOf(word)) + ending + " come first."
			return unnamedColors(text, []string{word}) == 0
		},
		gen.RegexMatch(`^[a-zа-яё]{1,16}$`), gen.RegexMatch(`^[a-zа-я]{0,3}$`),
	))
	properties.Property("a stem the end of another word holds does not name the word", prop.ForAll(
		func(word string) bool {
			stem := stemOf(word)
			joined := "q"
			if strings.HasPrefix(stem, joined) {
				joined = "z"
			}
			return unnamedColors(joined+stem, []string{word}) == 1
		},
		gen.RegexMatch(`^[a-zа-яё]{1,16}$`),
	))
	properties.Property("a text holding none of the word's letters does not name it", prop.ForAll(
		func(word string) bool {
			return unnamedColors("12 + 30 = 42?", []string{word}) == 1
		},
		gen.RegexMatch(`^[a-zа-яё]{1,16}$`),
	))
	properties.TestingRun(t)
}
