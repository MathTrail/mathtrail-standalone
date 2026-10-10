package checks

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// A picture that paints writes the word of each colour beside its paint, in
// the key under the picture, and the word is held to the wording as a label
// is: a key whose word the question never says leaves the child matching a
// colour to nothing. A word is found where the wording holds its stem — its
// first letters, all but its last two and never fewer than three — with case
// and ё read past, so that the word written in another form is found:
// красная in красной, rot in roten. In the scripts that set their words apart
// with spaces and seldom join one to the word before it — Latin, Greek and
// Cyrillic — the stem counts only where a word starts, so that red is not
// found in coloured, nor красн in прекрасный; elsewhere it counts wherever it
// stands, since Chinese, Japanese and Thai write no spaces and Arabic joins
// its article to the word. A word whose stem changes as it is declined, أحمر
// and حمراء, is found only in the form the wording writes, which is the one
// the guide asks for.

// unnamedColors counts the words of colours a picture writes that a text
// does not name: a word of two is named where the text holds both.
func unnamedColors(text string, words []string) int {
	folded := foldedForColors(text)
	unnamed := 0
	for _, word := range words {
		for _, part := range strings.Fields(foldedForColors(word)) {
			if !holdsStem(folded, stemOf(part)) {
				unnamed++
				break
			}
		}
	}
	return unnamed
}

// holdsStem says whether a text holds a stem where it may name a word: where a
// word starts, for a stem of Latin, Greek or Cyrillic letters, and anywhere for
// one of another script.
func holdsStem(text, stem string) bool {
	first, size := utf8.DecodeRuneInString(stem)
	if !unicode.In(first, unicode.Latin, unicode.Greek, unicode.Cyrillic) {
		return strings.Contains(text, stem)
	}
	for at := 0; ; {
		found := strings.Index(text[at:], stem)
		if found < 0 {
			return false
		}
		start := at + found
		before, _ := utf8.DecodeLastRuneInString(text[:start])
		if start == 0 || !unicode.IsLetter(before) {
			return true
		}
		at = start + size
	}
}

// stemOf is the part of a word a text must hold to name it: its first letters,
// all but its last two and never fewer than three, or the whole word where it
// is no longer than that.
func stemOf(word string) string {
	letters := []rune(word)
	return string(letters[:min(len(letters), max(3, len(letters)-2))])
}

// foldedForColors is a text with its case folded and ё read as е, as the
// wording and the words of colours are compared.
func foldedForColors(text string) string {
	return strings.ReplaceAll(strings.ToLower(text), "ё", "е")
}
