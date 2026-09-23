package checks

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// fleschKincaid is the grade the Flesch–Kincaid formula gives an English text.
// Words and sentences are counted as the prototype's library counted them,
// because the threshold was calibrated on its numbers; the syllables are an
// estimate, since the library looked them up in a pronouncing dictionary this
// service does not carry.
func fleschKincaid(text string) float64 {
	words := fkWords(text)
	sentences := fkSentences(text)
	if len(words) == 0 || sentences == 0 {
		return 0
	}
	syllables := 0
	for _, word := range words {
		syllables += syllablesOf(word)
	}
	return 0.39*float64(len(words))/float64(sentences) + 11.8*float64(syllables)/float64(len(words)) - 15.59
}

// fkWords are the words of a text as the library counted them: the pieces
// between whitespace that hold a letter, a digit or an underscore. The library
// took the punctuation out before it split, which never joins two pieces or
// splits one: "5-litre" and "don't" are one word each, and a dash standing
// alone is none. Each piece is kept as written, so that its syllables are
// counted with its accents.
func fkWords(text string) []string {
	var words []string
	for _, piece := range strings.Fields(text) {
		if strings.IndexFunc(piece, isWordRune) >= 0 {
			words = append(words, piece)
		}
	}
	return words
}

// isWordRune is what a regular expression means by a word character: a letter,
// a number or an underscore.
func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

// fkSentences counts sentences as the library did: a stretch that starts at the
// edge of a word, runs up to the next full stop, exclamation or question mark
// and takes those marks with it. A stretch of two words or fewer is not counted
// — it is a heading or a stray, not a sentence — but a text always has one.
func fkSentences(text string) int {
	runes := []rune(text)
	if len(runes) == 0 {
		return 0
	}

	var found, ignored int
	for i := 0; i < len(runes); {
		if !atWordEdge(runes, i) || fkEnd(runes[i]) {
			i++
			continue
		}
		start := i
		i = pastSentence(runes, i)
		found++
		if len(fkWords(string(runes[start:i]))) <= 2 {
			ignored++
		}
	}
	return max(1, found-ignored)
}

// fkEnd is a mark the library ended a sentence at.
func fkEnd(r rune) bool { return r == '.' || r == '!' || r == '?' }

// atWordEdge says whether a position is the edge of a word: a word character
// with none before it, or the other way round.
func atWordEdge(runes []rune, i int) bool {
	if i == 0 {
		return isWordRune(runes[0])
	}
	return isWordRune(runes[i-1]) != isWordRune(runes[i])
}

// pastSentence is the position after the sentence that starts here: up to the
// next marks, and past them.
func pastSentence(runes []rune, i int) int {
	for i < len(runes) && !fkEnd(runes[i]) {
		i++
	}
	for i < len(runes) && fkEnd(runes[i]) {
		i++
	}
	return i
}

// syllablesOf estimates how many syllables an English word has. A word with no
// letters — a number — is one, as the library counted a number it had no
// pronunciation for.
func syllablesOf(word string) int {
	letters := lettersOf(word)
	if len(letters) == 0 {
		return 1
	}
	return max(1, vowelGroups(letters))
}

// A letter is a letter of a word as the estimate reads it: the letter it is
// written with, and what an accent on it says about how it is said.
type letter struct {
	// plain is the letter lowercased and without its accent: é, è and ë are e.
	plain rune
	// accented says it carries an accent of any kind. An accented e is said, as
	// in café, where a plain one at the end of a word would be silent.
	accented bool
	// apart says it carries a diaeresis: the ï of naïve is said apart from the
	// a before it.
	apart bool
}

// combiningDiaeresis is the two dots of ä, ë, ï, ö, ü and ÿ written as a mark
// of their own, after the letter they stand on.
const combiningDiaeresis = '\u0308'

// lettersOf reads the letters of a word and the accents on them, whether an
// accent is written into its letter or after it as a mark of its own. Digits
// and punctuation are not read.
func lettersOf(word string) []letter {
	var letters []letter
	for _, r := range norm.NFD.String(word) {
		switch {
		case unicode.IsLetter(r):
			letters = append(letters, letter{plain: unicode.ToLower(r)})
		case unicode.Is(unicode.Mn, r) && len(letters) > 0:
			last := &letters[len(letters)-1]
			last.accented = true
			last.apart = last.apart || r == combiningDiaeresis
		}
	}
	return letters
}

// vowelGroups counts the vowels of a word that begin a syllable, and takes off
// the e that is written and not said.
func vowelGroups(word []letter) int {
	groups := 0
	for i := range word {
		if vowelAt(word, i) && (i == 0 || !vowelAt(word, i-1) || word[i].apart || splits(word, i)) {
			groups++
		}
	}
	return groups - silentEndings(word)
}

// vowelAt says whether a letter of a word is said as a vowel: a, e, i, o, u,
// with or without an accent, and a y anywhere but first.
func vowelAt(word []letter, i int) bool {
	switch word[i].plain {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	case 'y':
		return i > 0
	}
	return false
}

// splits says whether a vowel is said apart from the vowel just before it: the
// i-a of "liar", the i-o of "lion", the u-a of "usual". Not in -cial, -tion,
// -sion or -xion, where the i is not said on its own, and not in the qua- of
// "equal".
func splits(word []letter, i int) bool {
	before := rune(0)
	if i >= 2 {
		before = word[i-2].plain
	}
	switch string([]rune{word[i-1].plain, word[i].plain}) {
	case "ia", "io", "iu":
		return !strings.ContainsRune("ctsx", before)
	case "ua":
		return before != 'q'
	}
	return false
}

// silentEndings is how many of a word's vowel groups its ending writes without
// saying: the final e of "make" but not of "table", "metre" or "free", the e of
// "jumped" and "makes" but not of "wanted", "boxes", "wishes", "apples" or
// "litres" — and never an accented e, as in "café" or "résumés".
func silentEndings(word []letter) int {
	n := len(word)
	plainAt := func(i int) rune { return word[i].plain }
	consonantAt := func(i int) bool { return i >= 0 && !vowelAt(word, i) }
	unsaid := func(i int) bool { return plainAt(i) == 'e' && !word[i].accented }
	// An l or r between a consonant and the last e is said as a syllable of its
	// own: ta-ble, me-tre, ap-ples, li-tres.
	syllabic := func(end int) bool {
		return end >= 2 && strings.ContainsRune("lr", plainAt(end-1)) && consonantAt(end-2)
	}
	switch {
	case n > 2 && unsaid(n-1) && plainAt(n-2) != 'e' && !syllabic(n-1):
		return 1
	case n > 3 && unsaid(n-2) && plainAt(n-1) == 'd' && consonantAt(n-3) && !strings.ContainsRune("td", plainAt(n-3)):
		return 1
	case n > 3 && unsaid(n-2) && plainAt(n-1) == 's' && consonantAt(n-3) && !strings.ContainsRune("sxzcgh", plainAt(n-3)) &&
		!syllabic(n-2):
		return 1
	}
	return 0
}
