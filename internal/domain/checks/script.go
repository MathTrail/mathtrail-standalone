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

// lengthOf is how long a text is, in the unit its script calls for, and which
// unit that was. Where words are separated by spaces, they are counted as a
// child reads them; where they are not, the letters and digits are, and the
// punctuation between them is not.
func lengthOf(text string) (count int, unit string) {
	if Spaceless(text) {
		for _, r := range text {
			if !boundary(r) {
				count++
			}
		}
		return count, "characters"
	}
	for _, token := range strings.Fields(text) {
		if strings.IndexFunc(token, readAloud) >= 0 {
			count++
		}
	}
	return count, "words"
}

// readAloud says whether a character is read out, which is what makes a token
// between spaces a word: "5-litre" is one word, and so are the "+", "=" and
// "-" of "2 + 3 - 1 = 4". Punctuation alone is not — the "?" French sets apart
// with a space, a dash, a guillemet — except the hyphen-minus, which standing
// alone in a task is the minus sign.
func readAloud(r rune) bool { return !unicode.IsPunct(r) || r == '-' }

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
