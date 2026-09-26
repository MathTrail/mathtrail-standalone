package solver

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// The short way a text of plain ASCII is folded is only a shortcut: it gives
// what the whole way gives, or two options would be one answer or two
// depending on which way their letters happened to take.
func TestTheShortcutForASCIIFoldsAsTheWholeWayDoes(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("a text of ASCII folds the same either way", prop.ForAll(
		func(text string) bool { return foldASCII(text) == foldAny(text) },
		gen.SliceOf(gen.IntRange(0, 127)).Map(func(codes []int) string {
			runes := make([]rune, len(codes))
			for i, code := range codes {
				runes[i] = rune(code)
			}
			return string(runes)
		}),
	))
	properties.TestingRun(t)
}
