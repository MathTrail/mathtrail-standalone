package checks

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// unnamedLabels counts the labels of a picture written in Latin capitals that
// its question, in a lesson labelled so, does not name; a number or a question
// mark is not looked for. The question is read softly, since a refusal costs
// the child an attempt: a label counts as named where the question has it as a
// word of its own, among capitals written together (the AB of a segment), in
// the name of a cell (B2), or inside a range whose ends it names — rows A to D
// name B and C in any language. An example of a kind of picture has no
// question, and nothing to count.
func unnamedLabels(question string, labels []string, by labelling) int {
	if strings.TrimSpace(question) == "" {
		return 0
	}
	declared := declaredOf(labels)
	runes := []rune(by.asLatin(question))
	ranges := rangesOf(runes, declared, by.listing)
	unnamed := 0
	for _, label := range labels {
		if capitalsOnly(label) && !named(runes, label, declared, ranges) {
			unnamed++
		}
	}
	return unnamed
}

// named says whether a question names a label, read softly: as a word of its
// own, among capitals written together, in the name of a cell, or inside a
// range whose ends it names.
func named(runes []rune, label string, declared map[string]bool, ranges []span) bool {
	return namedAsWord(runes, label) || namedInCapitals(runes, label, declared) || namedInCell(runes, label) ||
		inRange(label, ranges)
}

// declaredOf is a picture's labels as a set.
func declaredOf(labels []string) map[string]bool {
	declared := make(map[string]bool, len(labels))
	for _, label := range labels {
		declared[label] = true
	}
	return declared
}

// capitalsOnly says whether a label is written in Latin capitals alone.
func capitalsOnly(label string) bool {
	if label == "" {
		return false
	}
	for _, r := range label {
		if !latinCapital(r) {
			return false
		}
	}
	return true
}

// namedAsWord says whether a question has a label as a word of its own: with
// no Latin letter or digit joined to it, so that a script written without
// spaces — 点A在3 — names A as "point A is at 3" does, and so do rows A-D, box
// A's lid and the Turkish A'dan. A label of several letters is found whatever
// its case, as the start names Start. A lone capital names its label wherever
// it stands but straight after a number, where it is the number's unit, the L
// of a 3 L jug: the A opening a sentence is the English article only in
// English, and a label opening one in any language — A y B, A und B liegen,
// the Turkish A kutusunda, box A — is followed by a word as the article is;
// nor is the English I told from the row I by anything but its language.
func namedAsWord(runes []rune, label string) bool {
	target := []rune(label)
	for at := 0; at+len(target) <= len(runes); at++ {
		if !sameRunes(runes[at:at+len(target)], target) || joinedInLatin(runes, at, at+len(target)) {
			continue
		}
		if len(target) == 1 && followsNumber(runes, at) {
			continue
		}
		return true
	}
	return false
}

// sameRunes says whether a piece of a question is a label: the same capital,
// or the same word whatever its case.
func sameRunes(piece, label []rune) bool {
	for i := range label {
		if piece[i] != label[i] && (len(label) == 1 || unicode.ToLower(piece[i]) != unicode.ToLower(label[i])) {
			return false
		}
	}
	return true
}

// joinedInLatin says whether a Latin letter, with its accents or without, or a
// digit joins the piece of a text between two places to what stands on either
// side. A letter of another script joins nothing: those written without
// spaces set a label among their characters.
func joinedInLatin(runes []rune, start, end int) bool {
	return (start > 0 && joinsLatin(runes[start-1])) || (end < len(runes) && joinsLatin(runes[end]))
}

// joinsLatin says whether a character is part of a Latin word: a Latin letter
// or a digit. An apostrophe or a hyphen is not, though the wording is read the
// other way as if it were: there a capital is a label only where it surely is
// one, here a label counts as named wherever it may be — the A of A-D and of
// A's, and the T of a T-shirt as well.
func joinsLatin(r rune) bool {
	return unicode.IsNumber(r) || (unicode.IsLetter(r) && unicode.Is(unicode.Latin, r))
}

// namedInCapitals says whether a single capital is named among capitals
// written together, the AB of a segment: a run of two or more that stands
// apart, most of whose letters are labels — as the wording is read the other
// way, so that the capitals of a word such as NOT name nothing.
func namedInCapitals(runes []rune, label string, declared map[string]bool) bool {
	if first, _ := utf8.DecodeRuneInString(label); utf8.RuneCountInString(label) != 1 || !latinCapital(first) {
		return false
	}
	for start := 0; start < len(runes); {
		if !latinCapital(runes[start]) {
			start++
			continue
		}
		end := start + 1
		for end < len(runes) && latinCapital(runes[end]) {
			end++
		}
		run := string(runes[start:end])
		if end-start >= 2 && !joinedInLatin(runes, start, end) && strings.Contains(run, label) &&
			len(labelsOf(runes, start, end, declared)) > 0 {
			return true
		}
		start = end
	}
	return false
}

// namedInCell says whether a label is named in the name of a cell, a capital
// and the number after it as B2 is, which names the row B and the column 2.
func namedInCell(runes []rune, label string) bool {
	for start := 0; start < len(runes); start++ {
		if !latinCapital(runes[start]) {
			continue
		}
		end := start + 1
		for end < len(runes) && runes[end] >= '0' && runes[end] <= '9' {
			end++
		}
		if end-start < 2 || joinedInLatin(runes, start, end) {
			continue
		}
		if string(runes[start]) == label || string(runes[start+1:end]) == label {
			return true
		}
	}
	return false
}

// span is a range a question names: the kind of label at its ends and the
// places of the two in the order of that kind.
type span struct{ kind, low, high int }

// inRange says whether a label lies strictly inside a range a question names.
func inRange(label string, ranges []span) bool {
	kind, place := orderOf(label)
	for _, named := range ranges {
		if kind != unordered && kind == named.kind && named.low < place && place < named.high {
			return true
		}
	}
	return false
}

// rangesOf are the ranges a question names, in a language that joins two
// labels in words as listed: two declared labels of a kind, each standing
// alone or in the name of a cell, joined as the ends of a range are — rows A
// to D, von A bis D, A부터 D까지, 从A到D, cells A1 to D4 — so that a range is
// read in any language, while points A and D, or A at 3 and D at 9, name none.
func rangesOf(runes []rune, declared map[string]bool, listed listing) []span {
	var ranges []span
	ends := endsOf(runes)
	for i := 1; i < len(ends); i++ {
		first, second := ends[i-1], ends[i]
		if !joinsRange(string(runes[first.stop:second.start]), listed) {
			continue
		}
		for _, pair := range [][2]string{{first.letter, second.letter}, {first.number, second.number}} {
			if declared[pair[0]] && declared[pair[1]] {
				kind, one := orderOf(pair[0])
				_, other := orderOf(pair[1])
				ranges = append(ranges, span{kind: kind, low: min(one, other), high: max(one, other)})
			}
		}
	}
	return ranges
}

// end is a piece of a question that may end a range, with where it starts and
// stops: a Latin capital, a number, or the two in the name of a cell.
type end struct {
	start, stop    int
	letter, number string
}

// endsOf are the pieces of a question that may end a range, in order, each
// standing apart from any Latin word. Capitals written together, the AB of a
// segment, end none; they are kept all the same, since a range has nothing of
// the kind between its ends.
func endsOf(runes []rune) []end {
	var ends []end
	for at := 0; at < len(runes); {
		piece, found := endAt(runes, at)
		if found {
			ends = append(ends, piece)
		}
		at = max(piece.stop, at+1)
	}
	return ends
}

// endAt is the end a question has at a place, if it has one there: capitals
// and the digits after them, joined to no Latin word. Where there is none, it
// still stops past the characters it read.
func endAt(runes []rune, at int) (end, bool) {
	piece := end{start: at, stop: at}
	if (!latinCapital(runes[at]) && !asciiDigit(runes[at])) || (at > 0 && joinsLatin(runes[at-1])) {
		return piece, false
	}
	capitals := at
	for capitals < len(runes) && latinCapital(runes[capitals]) {
		capitals++
	}
	piece.stop = capitals
	for piece.stop < len(runes) && asciiDigit(runes[piece.stop]) {
		piece.stop++
	}
	if piece.stop < len(runes) && joinsLatin(runes[piece.stop]) {
		return piece, false
	}
	if capitals-at <= 1 {
		piece.letter = string(runes[at:capitals])
		piece.number = string(runes[capitals:piece.stop])
	}
	return piece, true
}

// asciiDigit says whether a character is one of the digits a number label is
// drawn with.
func asciiDigit(r rune) bool { return r >= '0' && r <= '9' }

// joinsRange says whether what stands between two ends joins them as the ends
// of a range, in a language that joins two labels in words as listed: a dash,
// dots, or at most three words, between spaces or the word spaces of Ethiopic
// — the "to" of rows A to D in whatever language — none of them a word or a
// mark that lists or reckons instead, as "and", a comma or a plus do, and no
// end of a sentence, but where the words are a phrase that spans a range,
// though a word of it lists, as A till och med D does.
func joinsRange(joint string, listed listing) bool {
	trimmed := strings.TrimSpace(joint)
	if slices.Contains([]string{"..", "...", "…"}, trimmed) {
		return true
	}
	words := strings.FieldsFunc(trimmed, betweenWords)
	if len(words) == 0 || len(words) > 3 || marksAList(trimmed) {
		return false
	}
	if listed.spans(words) {
		return true
	}
	for _, word := range words {
		if listed.lists(word) {
			return false
		}
	}
	return true
}

// listingMarks list or reckon two things, and never join the ends of a range:
// among them the comma and the semicolon of Ethiopic, the comma of Armenian,
// the middle dot of a product, the semicolon of Greek, which is a raised dot
// that a keyboard writes as the middle dot, and the Greek question mark, which
// looks like a semicolon.
const listingMarks = ",;:、،؛&+=×÷*<>≤≥≠−፣፤\u055d\u00b7\u0387\u037e"

// The kinds of label that come in an order.
const (
	unordered = iota
	letterLabel
	numberLabel
)

// orderOf is the kind of a label and its place in the order of its kind: a
// Latin capital by the alphabet, a number by its value.
func orderOf(label string) (kind, place int) {
	if utf8.RuneCountInString(label) == 1 && latinCapital(rune(label[0])) {
		return letterLabel, int(label[0])
	}
	if value, err := strconv.Atoi(label); err == nil && value >= 0 {
		return numberLabel, value
	}
	return unordered, 0
}

// alphanumeric says whether a character is a letter or a digit.
func alphanumeric(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }

// unpicturedLabels are the labels the wording names and the picture does not
// show, each once.
func unpicturedLabels(question string, labels []string) []string {
	declared := declaredOf(labels)
	var unpictured []string
	for _, label := range namedLabels(question, declared) {
		if !declared[label] && !slices.Contains(unpictured, label) {
			unpictured = append(unpictured, label)
		}
	}
	return unpictured
}

// namedLabels are the labels a wording names, as far as they can be told from
// its words. A label is a Latin capital standing apart from any word; numbers
// are quantities and quotations are speech, so neither is read. Where a
// capital could as well be a word, it is taken for a word, because a refusal
// for a label that was never one costs the child an attempt.
func namedLabels(question string, declared map[string]bool) []string {
	runes := []rune(question)
	var named []string
	for start := 0; start < len(runes); {
		if !latinCapital(runes[start]) {
			start++
			continue
		}
		end := start + 1
		for end < len(runes) && latinCapital(runes[end]) {
			end++
		}
		if standsAlone(runes, start, end) {
			named = append(named, labelsOf(runes, start, end, declared)...)
		}
		start = end
	}
	return named
}

// labelsOf are the labels a run of capitals standing alone names, if any.
//
// A lone capital is a label, except where it may be a word: at the start of a
// sentence, of a quotation or after a colon it may be the English "A", the
// English "I" is always a word, and a capital straight after a number is its
// unit, the L of "a 3 L jug". Capitals written together — the AB of a
// segment, the ABC of a triangle — are a label each when most of them are
// labels of the picture, and a word in capitals otherwise, like NOT.
func labelsOf(runes []rune, start, end int, declared map[string]bool) []string {
	if end-start == 1 {
		if runes[start] == 'I' || opensSentence(runes, start) || followsNumber(runes, start) {
			return nil
		}
		return []string{string(runes[start])}
	}

	letters := strings.Split(string(runes[start:end]), "")
	known := 0
	for _, letter := range letters {
		if declared[letter] {
			known++
		}
	}
	if 2*known <= len(letters) {
		return nil
	}
	return letters
}

// latinCapital says whether a character is a capital of the Latin alphabet.
func latinCapital(r rune) bool { return r >= 'A' && r <= 'Z' }

// standsAlone says whether a run of capitals is a word of its own: nothing
// joins it to what stands on either side — a letter, a digit, an apostrophe
// or a hyphen, as in "O'Neil" or "T-shirt".
func standsAlone(runes []rune, start, end int) bool {
	return (start == 0 || !joinsWord(runes[start-1])) && (end == len(runes) || !joinsWord(runes[end]))
}

// joinsWord says whether a character joins what stands on either side of it
// into one word.
func joinsWord(r rune) bool {
	return alphanumeric(r) || strings.ContainsRune("'’-‐‑", r)
}

// opensSentence says whether a position begins a sentence, a quotation or
// what follows a colon: nothing but space stands between it and the start of
// the text, a sentence's final mark, a colon or an opening quote.
func opensSentence(runes []rune, at int) bool {
	i := at - 1
	for i >= 0 && unicode.IsSpace(runes[i]) {
		i--
	}
	return i < 0 || strings.ContainsRune(sentenceEnds, runes[i]) || strings.ContainsRune(`:"“„«‘「『`, runes[i])
}

// followsNumber says whether a capital comes straight after a number, with
// nothing but spaces or a degree sign between: the L of "3 L", the C of
// "30 °C".
func followsNumber(runes []rune, at int) bool {
	i := at - 1
	for i >= 0 && (runes[i] == ' ' || runes[i] == '°') {
		i--
	}
	return i >= 0 && unicode.IsNumber(runes[i])
}

// mismatch is a problem with how a picture agrees with the wording.
func mismatch(format string, args ...any) Problem {
	return Problem{Code: CodeDrawingMismatch, Message: fmt.Sprintf(format, args...)}
}
