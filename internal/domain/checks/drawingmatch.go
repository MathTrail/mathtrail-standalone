package checks

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DrawingMatch checks that a drawing, its structure and the wording name the
// same things: every label the structure declares is drawn, and every label
// the wording names is declared. Whether the picture means what the wording
// says is past what a program can tell; the model's self-check answers for it.
//
// A task without a drawing, or without the structure that describes it, has
// nothing to match, and the structure check says so where it should.
//
// A refusal names the labels themselves. They are the wording's own words and
// say nothing about which option is right.
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
	return problems
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
