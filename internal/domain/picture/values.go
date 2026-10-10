package picture

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Decimals is the mark a number is written with before its decimals, as the
// lesson's language writes it: a picture writes its numbers as the wording
// does, so that a child reads 2,5 on the card where the question says 2,5.
type Decimals int

// The two decimal marks a lesson writes.
const (
	// Point writes two and a half as 2.5.
	Point Decimals = iota
	// Comma writes it as 2,5.
	Comma
)

// mark is the character the decimal mark is written with.
func (d Decimals) mark() rune {
	if d == Comma {
		return ','
	}
	return '.'
}

// Time is a time of day, from 0:00 to 23:59.
type Time struct {
	Hour, Minute int
}

// OnTheFace is the time as a face of twelve hours shows it: the same hands
// stand for 4:30 and 16:30, and for 0:15 and 12:15, so the hour runs from 1 to
// 12.
func (t Time) OnTheFace() Time {
	hour := t.Hour % 12
	if hour == 0 {
		hour = 12
	}
	return Time{Hour: hour, Minute: t.Minute}
}

// timeWritten is a time as the format writes it, H:MM.
var timeWritten = regexp.MustCompile(`^([01]?\d|2[0-3]):([0-5]\d)$`)

// ReadTime reads a time written as the format writes it, H:MM.
func ReadTime(text string) (Time, bool) {
	parts := timeWritten.FindStringSubmatch(text)
	if parts == nil {
		return Time{}, false
	}
	hour, _ := strconv.Atoi(parts[1])
	minute, _ := strconv.Atoi(parts[2])
	return Time{Hour: hour, Minute: minute}, true
}

// isLabel says whether a text is a label: a question mark, a Latin capital or
// a run of them, or a number written as the wording writes it, and never past
// MaxLabelCharacters characters.
func (d Decimals) isLabel(text string) bool {
	if utf8.RuneCountInString(text) > MaxLabelCharacters {
		return false
	}
	return text == "?" || capitals(text) || d.isNumber(text)
}

// capitals says whether a text is a run of Latin capitals and nothing else.
func capitals(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// isNumber says whether a text is a number as a label writes it: digits, a
// minus sign before them if it is below zero — the hyphen or the sign of
// mathematics, which are one sign to a child — and decimals after the
// lesson's decimal mark. Nothing groups its thousands: 1,000 would read as
// one in a language that writes its decimals with a comma.
func (d Decimals) isNumber(text string) bool {
	if unsigned, signed := strings.CutPrefix(text, "-"); signed {
		text = unsigned
	} else {
		text = strings.TrimPrefix(text, "−")
	}
	whole, decimals, marked := strings.Cut(text, string(d.mark()))
	return digits(whole) && (!marked || digits(decimals))
}

// digits says whether a text is one or more of the digits 0 to 9.
func digits(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isCell says whether a text may stand in a cell of a table: a label, a time,
// an ellipsis for cells left out, or nothing.
func (d Decimals) isCell(text string) bool {
	_, isTime := ReadTime(text)
	return text == "" || text == Ellipsis || isTime || d.isLabel(text)
}

// Ellipsis is what a table writes in a cell for the cells it leaves out.
const Ellipsis = "…"

// noteSigns are the signs a note may join its labels and numbers with.
const noteSigns = "+-−×÷=<>() ?"

// isNote says whether a text is a note: an equality or an inequality of
// labels and numbers, short enough to stand under a picture.
func (d Decimals) isNote(text string) bool {
	return utf8.RuneCountInString(text) <= MaxNoteCharacters && strings.ContainsAny(text, "=<>") && d.joined(text)
}

// joined says whether a text is labels, numbers and ? joined by the signs a
// note may hold, and spaces, with no run of capitals longer than a label.
func (d Decimals) joined(text string) bool {
	run := 0
	for _, r := range text {
		switch {
		case r >= 'A' && r <= 'Z':
			run++
			if run > MaxLabelCharacters {
				return false
			}
			continue
		case (r >= '0' && r <= '9') || r == d.mark() || strings.ContainsRune(noteSigns, r):
		default:
			return false
		}
		run = 0
	}
	return true
}

// noteTokens are the labels and numbers a note holds, in its order: each run
// of capitals, each number with its decimals and, below zero, its sign, and
// each question mark.
func noteTokens(note string, decimals Decimals) []string {
	var tokens []string
	runes := []rune(note)
	for at := 0; at < len(runes); {
		end := at + 1
		switch {
		case runes[at] >= 'A' && runes[at] <= 'Z':
			for end < len(runes) && runes[end] >= 'A' && runes[end] <= 'Z' {
				end++
			}
		case runes[at] >= '0' && runes[at] <= '9', signed(runes, at):
			for end < len(runes) && (runes[end] >= '0' && runes[end] <= '9' || runes[end] == decimals.mark()) {
				end++
			}
		case runes[at] != '?':
			at++
			continue
		}
		tokens = append(tokens, strings.TrimRight(string(runes[at:end]), string(decimals.mark())))
		at = end
	}
	return tokens
}

// signed says whether the minus at a place in a note is the sign of the
// number after it, as in A = −3, rather than a subtraction, as in A − 3: a
// digit follows it, and no label, number, ? or closing bracket stands before
// it.
func signed(runes []rune, at int) bool {
	if (runes[at] != '-' && runes[at] != '−') || at+1 == len(runes) || runes[at+1] < '0' || runes[at+1] > '9' {
		return false
	}
	for before := at - 1; before >= 0; before-- {
		switch r := runes[before]; {
		case r == ' ':
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '?', r == ')':
			return false
		default:
			return true
		}
	}
	return true
}
