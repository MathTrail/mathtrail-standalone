package picture

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Paint is a colour of the card's palette. A picture paints with these alone,
// each drawn in its theme's tone, and writes beside each the word the lesson
// calls it by: a colour drawn with no word would tell nothing to a child who
// cannot tell it apart, and nothing to a screen reader.
type Paint string

// The colours of the palette, in the order the format lists them.
const (
	Red    Paint = "red"
	Yellow Paint = "yellow"
	Green  Paint = "green"
	Blue   Paint = "blue"
	White  Paint = "white"
	Black  Paint = "black"
)

// Paints are the colours of the palette, in the order the format lists them.
func Paints() []Paint {
	return []Paint{Red, Yellow, Green, Blue, White, Black}
}

// unknownPaint is what a part of a picture is painted where its colour is the
// unknown.
const unknownPaint = "?"

// colors are the colours a picture paints with, each with the word the lesson
// calls it by. A colour whose word could not be read stands with none, so
// that what it paints is not refused a second time.
type colors map[Paint]string

// The rules of a picture's colours, as a refusal says them.
var (
	colorWordRule = fmt.Sprintf("must be the word the question calls the colour by: letters of any script, at most "+
		"%d words and %d characters", maxColorWords, maxColorCharacters)
	paletteRule = "is no colour of the palette: name " + oneOf(paintNames())
)

// readColors reads the colours a picture paints with: each a colour of the
// palette, keyed by its name, with the word the lesson calls it by, and no two
// by one word. A key that is no colour of the palette is the model's own, and
// a refusal never names it.
func readColors(o *object) colors {
	entries := o.keyed("colors")
	read := colors{}
	words := map[string]bool{}
	stray := false
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		paint := Paint(key)
		if !slices.Contains(Paints(), paint) {
			if !stray {
				o.fault(o.at("colors")+".*", "%s", paletteRule)
			}
			stray = true
			continue
		}
		read[paint] = ""
		word, isText := entries[key].(string)
		if !isText || !isColorWord(word) {
			o.fault(o.at("colors")+"."+key, "%s", colorWordRule)
			continue
		}
		folded := strings.ToLower(word)
		if words[folded] {
			o.fault(o.at("colors"), "calls two colours by one word: give each its own")
			continue
		}
		words[folded] = true
		read[paint] = word
	}
	return read
}

// isColorWord says whether a text is a word a colour may be called by: letters
// of any script, with the marks some scripts write their vowels with, in at
// most maxColorWords words a space apart, and no longer than
// maxColorCharacters characters.
func isColorWord(text string) bool {
	words := strings.Split(text, " ")
	if utf8.RuneCountInString(text) > maxColorCharacters || len(words) > maxColorWords {
		return false
	}
	for _, word := range words {
		first, _ := utf8.DecodeRuneInString(word)
		if word == "" || !unicode.IsLetter(first) {
			return false
		}
		for _, r := range word {
			if !unicode.IsLetter(r) && !unicode.IsMark(r) {
				return false
			}
		}
	}
	return true
}

// paint reads a member that paints a part of a picture: a colour the
// picture's colors name, or ? where the unknown may stand. The colour is
// marked used, so that one that paints nothing can be told.
func (c colors) paint(r *reading, path string, value any, unknown bool, used map[Paint]bool) {
	text, _ := value.(string)
	if unknown && text == unknownPaint {
		return
	}
	if _, named := c[Paint(text)]; !named {
		rule := "must be a colour colors names"
		if unknown {
			rule += ", or ?"
		}
		r.fault(path, "%s", rule)
		return
	}
	used[Paint(text)] = true
}

// unused says of each colour colors names that it paints nothing: a word in
// the key under the picture beside no paint in it.
func (c colors) unused(o *object, used map[Paint]bool) {
	for _, paint := range Paints() {
		if _, named := c[paint]; named && !used[paint] {
			o.fault(o.at("colors")+"."+string(paint), "paints nothing in the picture; leave it out")
		}
	}
}

// words are the words of the colours, in the order the palette lists them.
func (c colors) words() []string {
	var words []string
	for _, paint := range Paints() {
		if word := c[paint]; word != "" {
			words = append(words, word)
		}
	}
	return words
}

// paintNames are the names of the palette's colours.
func paintNames() []string {
	names := make([]string, 0, len(Paints()))
	for _, paint := range Paints() {
		names = append(names, string(paint))
	}
	return names
}
