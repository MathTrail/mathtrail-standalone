package checks

import (
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// decimalsOf is the mark the lesson's language writes a number's decimals
// with, as the common data of the world's locales has it: a point in English,
// a comma in Russian or German, and the region's own where it has one — a
// point in Swiss German, a comma in South African English. A picture writes
// its numbers as its wording does, so this is the mark a label is read with.
// A language the data does not know writes them with a point.
func decimalsOf(lesson string) picture.Decimals {
	tag, err := language.Parse(lesson)
	if err != nil || tag == language.Und {
		return picture.Point
	}
	if strings.ContainsRune(message.NewPrinter(tag).Sprint(0.5), ',') {
		return picture.Comma
	}
	return picture.Point
}
