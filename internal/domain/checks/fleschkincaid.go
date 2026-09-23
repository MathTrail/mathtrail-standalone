package checks

import (
	"strings"
	"unicode"
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
		syllables += syllablesOf(strings.ToLower(word))
	}
	return 0.39*float64(len(words))/float64(sentences) + 11.8*float64(syllables)/float64(len(words)) - 15.59
}

// fkWords are the words of a text as the library counted them: punctuation
// removed without leaving a gap — "5-litre" is one word — except an apostrophe
// that begins a contraction, then split on whitespace.
func fkWords(text string) []string {
	runes := []rune(text)
	var kept strings.Builder
	for i, r := range runes {
		switch {
		case r == '\'':
			if contractionFollows(runes[i+1:]) {
				kept.WriteRune(r)
			}
		case isWordRune(r) || unicode.IsSpace(r):
			kept.WriteRune(r)
		}
	}
	return strings.Fields(kept.String())
}

// contractionFollows says whether what follows an apostrophe makes it the
// apostrophe of an English contraction: 't, 's, 'd, 've, 'll or 're.
func contractionFollows(rest []rune) bool {
	if len(rest) == 0 {
		return false
	}
	switch rest[0] {
	case 't', 's', 'd':
		return true
	}
	if len(rest) < 2 {
		return false
	}
	switch string(rest[:2]) {
	case "ve", "ll", "re":
		return true
	}
	return false
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

// syllablesOf estimates how many syllables an English word has. A word of
// digits is one, as the library counted a number it had no pronunciation for.
func syllablesOf(word string) int {
	letters := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, word)
	if letters == "" {
		return 1
	}
	return max(1, vowelGroups(letters))
}

// vowelGroups counts the vowels of a word that begin a syllable, and takes off
// the e that is written and not said.
func vowelGroups(word string) int {
	groups := 0
	for i := range len(word) {
		if vowelAt(word, i) && (i == 0 || !vowelAt(word, i-1) || splits(word, i)) {
			groups++
		}
	}
	return groups - silentEndings(word)
}

// vowelAt says whether a letter of a word is said as a vowel: a, e, i, o, u,
// and a y anywhere but first.
func vowelAt(word string, i int) bool {
	switch word[i] {
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
func splits(word string, i int) bool {
	before := byte(0)
	if i >= 2 {
		before = word[i-2]
	}
	switch word[i-1 : i+1] {
	case "ia", "io", "iu":
		return !strings.ContainsRune("ctsx", rune(before))
	case "ua":
		return before != 'q'
	}
	return false
}

// silentEndings is how many of a word's vowel groups its ending writes without
// saying: the final e of "make" but not of "table", "metre" or "free", the e of
// "jumped" and "makes" but not of "wanted", "boxes", "apples" or "litres".
func silentEndings(word string) int {
	n := len(word)
	consonantAt := func(i int) bool { return i >= 0 && !vowelAt(word, i) }
	// An l or r between a consonant and the last e is said as a syllable of its
	// own: ta-ble, me-tre, ap-ples, li-tres.
	syllabic := func(end int) bool {
		return end >= 2 && strings.ContainsRune("lr", rune(word[end-1])) && consonantAt(end-2)
	}
	switch {
	case n > 2 && word[n-1] == 'e' && !strings.HasSuffix(word, "ee") && !syllabic(n-1):
		return 1
	case n > 3 && strings.HasSuffix(word, "ed") && consonantAt(n-3) && !strings.ContainsRune("td", rune(word[n-3])):
		return 1
	case n > 3 && strings.HasSuffix(word, "es") && consonantAt(n-3) && !strings.ContainsRune("sxzcg", rune(word[n-3])) &&
		!strings.HasSuffix(word, "hes") && !syllabic(n-2):
		return 1
	}
	return 0
}
