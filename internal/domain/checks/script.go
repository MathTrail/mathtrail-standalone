package checks

import (
	"strings"
	"unicode"
)

// spaceless are the scripts written without spaces between words. A text in
// them is measured in characters rather than words: by the readability check
// for its sentences, by the explanation check for its length, and by the
// duplicate check for its bigrams.
var spaceless = []*unicode.RangeTable{
	unicode.Han, unicode.Hiragana, unicode.Katakana,
	unicode.Thai, unicode.Lao, unicode.Khmer, unicode.Myanmar, unicode.Tibetan,
}

// Spaceless says whether most of the letters of a text belong to the scripts
// written without spaces. The two kinds of script are counted against each
// other rather than script by script, so that Japanese — kanji, hiragana and
// katakana together — is one kind of text however its letters divide.
func Spaceless(text string) bool {
	var letters, without int
	for _, r := range text {
		if !unicode.IsLetter(r) {
			continue
		}
		letters++
		if unicode.In(r, spaceless...) {
			without++
		}
	}
	return 2*without > letters
}

// lengthOf is how long a text is, in the unit its script calls for — its words,
// or its letters and digits — and which unit that was. Punctuation and spacing
// count for nothing in either, so a text is as long as what it says.
func lengthOf(text string) (count int, unit string) {
	if Spaceless(text) {
		for _, r := range text {
			if !boundary(r) {
				count++
			}
		}
		return count, "characters"
	}
	return len(strings.FieldsFunc(text, boundary)), "words"
}

// reading is a text as two texts are compared by what they say: lowercased,
// and everything that is neither a letter, a digit nor a mark reduced to a
// single space between words.
func reading(text string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(text), boundary), " ")
}

// startsWith says whether a text begins with another, as they read: word for
// word where words are separated by spaces, character for character where they
// are not. "Four" does not start "Fourteen ships", and 小明 does start 小明有.
func startsWith(text, start string) bool {
	text, start = reading(text), reading(start)
	if start == "" || !strings.HasPrefix(text, start) {
		return false
	}
	return len(text) == len(start) || Spaceless(start) || text[len(start)] == ' '
}
