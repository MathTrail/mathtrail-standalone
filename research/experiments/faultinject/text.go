package main

import (
	"cmp"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// sentenceMarks end a sentence when a space or the end of the text follows
// them, after any closing quotes or brackets.
const (
	sentenceMarks = ".!?"
	closingMarks  = "\"'”’»)]"
)

// sentencesOf splits a question into its sentences, each with its own end. A
// mark ends a sentence only where a space or the end of the text comes after
// it, so "1.5 litres" stays whole.
func sentencesOf(text string) []string {
	runes := []rune(text)
	var found []string
	start := 0
	for i := 0; i < len(runes); i++ {
		if !strings.ContainsRune(sentenceMarks, runes[i]) {
			continue
		}
		end := i + 1
		for end < len(runes) && strings.ContainsRune(closingMarks, runes[end]) {
			end++
		}
		if end < len(runes) && !unicode.IsSpace(runes[end]) {
			continue
		}
		if sentence := strings.TrimSpace(string(runes[start:end])); sentence != "" {
			found = append(found, sentence)
		}
		start, i = end, end-1
	}
	if rest := strings.TrimSpace(string(runes[start:])); rest != "" {
		found = append(found, rest)
	}
	return found
}

// wordsIn counts the words of a text: the pieces between spaces that hold a
// letter or a digit, so that a dash or an ellipsis standing alone is none.
func wordsIn(text string) int {
	count := 0
	for _, piece := range strings.Fields(text) {
		if strings.IndexFunc(piece, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) >= 0 {
			count++
		}
	}
	return count
}

// firstWords is the first n words of a text, or all of it when it has fewer.
func firstWords(text string, n int) string {
	words := strings.Fields(text)
	return strings.Join(words[:min(n, len(words))], " ")
}

// lowerFirst lowers the first letter of a text.
func lowerFirst(text string) string {
	first, size := utf8.DecodeRuneInString(text)
	if first == utf8.RuneError {
		return text
	}
	return string(unicode.ToLower(first)) + text[size:]
}

// withCaseOf writes a replacement with the capital of the word it replaces: in
// capitals when that word is all capitals, with a capital first letter when it
// has one.
func withCaseOf(replacement, original string) string {
	switch {
	case len(original) > 1 && original == strings.ToUpper(original):
		return strings.ToUpper(replacement)
	case startsUpper(original):
		first, size := utf8.DecodeRuneInString(replacement)
		return string(unicode.ToUpper(first)) + replacement[size:]
	default:
		return replacement
	}
}

// startsUpper says whether a text begins with a capital.
func startsUpper(text string) bool {
	first, _ := utf8.DecodeRuneInString(text)
	return unicode.IsUpper(first)
}

// wholeNumber finds the runs of digits in a text. A run is a whole number
// unless a point or a comma joins it to more digits, as in 1.5 or 2,000.
var wholeNumber = regexp.MustCompile(`\d+`)

// wholeNumbers are where the whole numbers of a text stand, in order.
func wholeNumbers(text string) [][2]int {
	var spans [][2]int
	for _, at := range wholeNumber.FindAllStringIndex(text, -1) {
		if joinedToDigits(text, at[0], at[1]) {
			continue
		}
		spans = append(spans, [2]int{at[0], at[1]})
	}
	return spans
}

// joinedToDigits says whether a run of digits is part of a longer number: a
// point or a comma stands between it and more digits on either side.
func joinedToDigits(text string, start, end int) bool {
	before := start >= 2 && strings.ContainsRune(".,", rune(text[start-1])) && isDigit(text[start-2])
	after := end+1 < len(text) && strings.ContainsRune(".,", rune(text[end])) && isDigit(text[end+1])
	return before || after
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// bumpNumbers writes every whole number n of a text as n + 1, and says whether
// there was one.
func bumpNumbers(text string) (string, bool) {
	spans := wholeNumbers(text)
	if len(spans) == 0 {
		return text, false
	}
	var out strings.Builder
	last := 0
	for _, span := range spans {
		n, err := strconv.Atoi(text[span[0]:span[1]])
		if err != nil {
			return text, false
		}
		out.WriteString(text[last:span[0]])
		out.WriteString(strconv.Itoa(n + 1))
		last = span[1]
	}
	out.WriteString(text[last:])
	return out.String(), true
}

// replaceFirstNumber rewrites the first whole number of a text, and says
// whether there was one.
func replaceFirstNumber(text string, rewrite func(n int) string) (string, bool) {
	spans := wholeNumbers(text)
	if len(spans) == 0 {
		return text, false
	}
	n, err := strconv.Atoi(text[spans[0][0]:spans[0][1]])
	if err != nil {
		return text, false
	}
	return text[:spans[0][0]] + rewrite(n) + text[spans[0][1]:], true
}

// numberWords are the whole numbers from zero to twenty in English words.
var numberWords = [...]string{
	"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten",
	"eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen", "twenty",
}

// inWords is a whole number from zero to twenty written in English words, and
// whether the text was such a number written in digits and nothing else.
func inWords(text string) (string, bool) {
	n, err := strconv.Atoi(text)
	if err != nil || n < 0 || n >= len(numberWords) || strconv.Itoa(n) != text {
		return "", false
	}
	return numberWords[n], true
}

// swapWords rewrites the words of a text that a map names, all at once, each
// with the capital of the word it replaces; it says whether any was found.
func swapWords(text string, pattern *regexp.Regexp, to map[string]string) (string, bool) {
	found := false
	out := pattern.ReplaceAllStringFunc(text, func(word string) string {
		replacement, known := to[strings.ToLower(word)]
		if !known {
			return word
		}
		found = true
		return withCaseOf(replacement, word)
	})
	return out, found
}

// wordPattern matches any of these words standing whole, in any case.
func wordPattern(words []string) *regexp.Regexp {
	quoted := make([]string, 0, len(words))
	for _, word := range words {
		quoted = append(quoted, regexp.QuoteMeta(word))
	}
	return regexp.MustCompile(`(?i)\b(` + strings.Join(quoted, "|") + `)\b`)
}

// cyclic maps every word of a list to the next one, the last to the first.
func cyclic(words []string) map[string]string {
	next := make(map[string]string, len(words))
	for i, word := range words {
		next[strings.ToLower(word)] = words[(i+1)%len(words)]
	}
	return next
}

// keysOf lists the keys of a map of words, the longest first and then in
// alphabetical order, so that a pattern built from them is the same every run.
func keysOf(words map[string]string) []string {
	keys := slices.Collect(maps.Keys(words))
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Or(cmp.Compare(len(b), len(a)), cmp.Compare(a, b))
	})
	return keys
}

// swapFirstWord rewrites the first word of a text that a map names, with the
// capital of the word it replaces; a word mapped to "" is removed with the
// space before it. It says whether any was found.
func swapFirstWord(text string, pattern *regexp.Regexp, to map[string]string) (string, bool) {
	at := pattern.FindStringIndex(text)
	if at == nil {
		return text, false
	}
	word := text[at[0]:at[1]]
	replacement := to[strings.ToLower(word)]
	if replacement == "" {
		start := at[0]
		if start > 0 && text[start-1] == ' ' {
			start--
		}
		return text[:start] + text[at[1]:], true
	}
	return text[:at[0]] + withCaseOf(replacement, word) + text[at[1]:], true
}

// rejoin puts sentences back together with one space between them.
func rejoin(sentences []string) string { return strings.Join(sentences, " ") }
