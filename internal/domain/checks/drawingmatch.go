package checks

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DrawingMatch checks that a drawing, its structure and the wording name the
// same things: every label the structure declares is drawn, every label the
// wording names is declared, and every label declared is named in the
// wording, so that no mark on the picture is left for the child to guess at.
// Whether the picture means what the wording says is past what a program can
// tell; the model's self-check answers for it.
//
// A task without a drawing, or without the structure that describes it, has
// nothing to match, and the structure check says so where it should.
//
// A refusal quotes the labels the drawing leaves out and the ones the wording
// names; the labels the wording leaves unnamed it only counts, since a label
// that is a number may be the text of an option.
func DrawingMatch(question, drawing string, structure *DrawingStructure) []Problem {
	if structure == nil || strings.TrimSpace(drawing) == "" {
		return nil
	}

	var problems []Problem
	if undrawn := undrawnLabels(drawing, structure); len(undrawn) > 0 {
		problems = append(problems, mismatch("task.drawing does not show %s, which task.drawing_structure declares; "+
			"draw every object with its label", listed(undrawn)))
	}
	if undeclared := undeclaredLabels(question, structure); len(undeclared) > 0 {
		problems = append(problems, mismatch("task.question names %s, which task.drawing_structure does not declare; "+
			"describe and draw each, or leave it out of the wording", listed(undeclared)))
	}
	if unnamed := unnamedLabels(question, structure); unnamed > 0 {
		problems = append(problems, mismatch("task.question does not name %d of the labels task.drawing_structure "+
			"declares; name each in the question — the football club (F), segment AB, cell B2, rows A to D — "+
			"or leave it out of the drawing and of task.drawing_structure", unnamed))
	}
	return problems
}

// unnamedLabels counts the labels a structure declares that its question does
// not name. The question is read softly, since a refusal costs the child an
// attempt: a label counts as named where the question has it as a word of its
// own, among capitals written together (the AB of a segment), in the name of a
// cell (B2), or inside a range whose ends it names — rows A to D name B and C
// in any language. A frame's drawing has no question, and nothing to count.
func unnamedLabels(question string, structure *DrawingStructure) int {
	if strings.TrimSpace(question) == "" {
		return 0
	}
	labels := declaredLabels(structure)
	declared := make(map[string]bool, len(labels))
	for _, label := range labels {
		declared[label] = true
	}
	runes := []rune(question)
	ranges := rangesOf(runes, declared)
	unnamed := 0
	for _, label := range labels {
		if !namedAsWord(runes, label) && !namedInCapitals(runes, label, declared) && !namedInCell(runes, label) &&
			!inRange(label, ranges) {
			unnamed++
		}
	}
	return unnamed
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

// rangesOf are the ranges a question names: two declared labels of a kind,
// each standing alone or in the name of a cell, joined as the ends of a range
// are — rows A to D, von A bis D, A부터 D까지, 从A到D, cells A1 to D4 — so
// that a range is read in any language, while points A and D, or A at 3 and D
// at 9, name none.
func rangesOf(runes []rune, declared map[string]bool) []span {
	var ranges []span
	ends := endsOf(runes)
	for i := 1; i < len(ends); i++ {
		first, second := ends[i-1], ends[i]
		if !joinsRange(string(runes[first.stop:second.start])) {
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
// of a range: a dash, dots, or at most three words — the "to" of rows A to D
// in whatever language — none of them a word or a mark that lists or reckons
// instead, as "and", a comma or a plus do, and no end of a sentence.
func joinsRange(joint string) bool {
	trimmed := strings.TrimSpace(joint)
	if slices.Contains([]string{"..", "...", "…"}, trimmed) {
		return true
	}
	words := strings.Fields(trimmed)
	if len(words) == 0 || len(words) > 3 || strings.ContainsAny(trimmed, listingMarks+sentenceEnds) {
		return false
	}
	for _, word := range words {
		if slices.Contains(listingWords, strings.ToLower(word)) {
			return false
		}
	}
	return true
}

// listingMarks list or reckon two things, and never join the ends of a range.
const listingMarks = ",;:、،؛&+=×÷*<>≤≥≠−"

// listingWords list two things, or put one with the other, in the languages
// of the cards — and, or, with — and so does a slash standing alone, the one
// of A/D. Between two ends they name those two alone, not a range.
var listingWords = []string{
	"/",
	"and", "or", "nor", "plus", "with", // English
	"和", "与", "及", "或", "或者", "跟", "同", "以及", "还有", // Chinese
	"और", "या", "तथा", "एवं", "व", // Hindi
	"y", "e", "o", "u", "ni", "con", // Spanish
	"و", "أو", "او", "أم", // Arabic
	"et", "ou", "avec", // French
	"ও", "এবং", "বা", "আর", "অথবা", // Bengali
	"nem", "com", // Portuguese, with e and ou above
	"и", "или", "либо", "с", // Russian
	"اور", "یا", // Urdu and Persian
	"dan", "atau", "serta", "dengan", // Indonesian
	"und", "oder", "sowie", "mit", // German
	"と", "や", "か", "または", "および", "及び", "又は", // Japanese
	"ve", "veya", "ya", "ile", "yahut", // Turkish
	"와", "과", "및", "또는", "이나", "나", "하고", "랑", "이랑", // Korean
	"và", "hoặc", "với", "cùng", // Vietnamese
	"ed", "od", "oppure", // Italian, with e, o and con above
	"i", "oraz", "lub", "albo", "czy", "z", // Polish
	"і", "й", "та", "або", "чи", "з", // Ukrainian
	"และ", "หรือ", "กับ", // Thai
	"en", "of", "met", // Dutch
}

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

// declaredLabels are the labels the objects of a structure carry, each once.
// An object with no label is the structure check's to report.
func declaredLabels(structure *DrawingStructure) []string {
	var labels []string
	for _, object := range structure.Objects {
		if label := strings.TrimSpace(object.Label); label != "" && !slices.Contains(labels, label) {
			labels = append(labels, label)
		}
	}
	return labels
}

// undrawnLabels are the labels a structure declares and its drawing does not
// show.
func undrawnLabels(drawing string, structure *DrawingStructure) []string {
	var undrawn []string
	for _, label := range declaredLabels(structure) {
		if !drawn(drawing, label) {
			undrawn = append(undrawn, label)
		}
	}
	return undrawn
}

// drawn says whether a drawing shows a label whole: with no letter or digit
// joined to it, so that "A" is drawn in "A───B" but not in "BAR", and "1" is
// not drawn in "12". The lines, arrows and spaces around a label are what a
// drawing is made of, and they join it to nothing.
func drawn(drawing, label string) bool {
	for from := 0; from < len(drawing); {
		at := strings.Index(drawing[from:], label)
		if at < 0 {
			return false
		}
		start := from + at
		if standsApart(drawing, start, start+len(label)) {
			return true
		}
		from = start + 1
	}
	return false
}

// standsApart says whether no letter or digit is joined to either end of the
// piece of a text between two positions.
func standsApart(text string, start, end int) bool {
	first, _ := utf8.DecodeRuneInString(text[start:end])
	last, _ := utf8.DecodeLastRuneInString(text[start:end])
	before, _ := utf8.DecodeLastRuneInString(text[:start])
	after, _ := utf8.DecodeRuneInString(text[end:])
	joinedBefore := alphanumeric(first) && alphanumeric(before)
	joinedAfter := alphanumeric(last) && alphanumeric(after)
	return !joinedBefore && !joinedAfter
}

// alphanumeric says whether a character is a letter or a digit.
func alphanumeric(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }

// undeclaredLabels are the labels the wording names and the structure does not
// declare, each once.
func undeclaredLabels(question string, structure *DrawingStructure) []string {
	declared := map[string]bool{}
	for _, label := range declaredLabels(structure) {
		declared[label] = true
	}
	var undeclared []string
	for _, label := range namedLabels(question, declared) {
		if !declared[label] && !slices.Contains(undeclared, label) {
			undeclared = append(undeclared, label)
		}
	}
	return undeclared
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
// labels the structure declares, and a word in capitals otherwise, like NOT.
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

// mismatch is a problem with how a drawing agrees with its structure and with
// the wording.
func mismatch(format string, args ...any) Problem {
	return Problem{Code: CodeDrawingMismatch, Message: fmt.Sprintf(format, args...)}
}
