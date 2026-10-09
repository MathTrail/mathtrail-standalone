// Package picture is the format of a task's picture: a description the card
// draws, of one of twelve kinds, and what can be read from it.
//
// The chat's model writes the description and the card draws it, so the
// picture comes out the same on every device: there is nothing in it for a
// font to draw differently. The package reads a description strictly, saying
// what is wrong with it member by member, and tells the checks what the
// picture shows — its labels, and the values a child reads off it — so that a
// picture can be held to the wording and kept from showing the answer.
//
// It is pure computation. Nothing here knows the task the picture belongs to.
package picture

import (
	"bytes"
	"encoding/json"
	"slices"
)

// Kind is what a picture draws: a clock, a row, a balance and so on.
type Kind string

// The twelve kinds of picture, in the order the format lists them.
const (
	Clock      Kind = "clock"
	Table      Kind = "table"
	NumberLine Kind = "number_line"
	Row        Kind = "row"
	Ring       Kind = "ring"
	Grid       Kind = "grid"
	Bars       Kind = "bars"
	Venn       Kind = "venn"
	Balance    Kind = "balance"
	Containers Kind = "containers"
	Piles      Kind = "piles"
	Calendar   Kind = "calendar"
)

// What KindOf says of a task with no picture, and of a picture of no kind the
// format has, which only a file edited by hand holds.
const (
	None  = "none"
	Other = "other"
)

// Picture is a description read into the format.
type Picture interface {
	// Kind is what the picture draws.
	Kind() Kind
	// Labels are every label the picture shows, each once, in the order the
	// description gives them: the names the wording is held to.
	Labels() []string
	// Shown are the values a child reads off the picture that could be the
	// answer to its task. Names that number places alike, and numbers that
	// only lay the picture out, are not among them: they single nothing out.
	Shown() []Shown
}

// Shown is one value a picture shows.
type Shown struct {
	// Text is the value as the description writes it: a label, a time in a
	// table, a cell, a label or a number of a note, or a number the card draws
	// for one thing, written in digits.
	Text string
	// Face is set for the time a clock shows instead of a text: where its
	// hands stand on a face of twelve hours.
	Face *Time
}

// Problem is one rule a description breaks.
type Problem struct {
	// Path names the member that breaks it, from the picture down —
	// picture.items.2.amount — and a key the model chose as "*", since a key
	// may be the answer's own cell.
	Path string
	// Rule says what the member must be, and never what it held: a value of
	// the picture may be the text of the right option.
	Rule string
}

// kinds are the kinds a description may name, each with how it is read and
// how its limits are told to the model, in the order the format lists them.
// A kind is added with a reader of its own and a line here.
var kinds = []struct {
	kind   Kind
	read   func(*object) Picture
	limits func() string
}{
	{Clock, readClock, clockLimits},
	{Table, readTable, tableLimits},
	{NumberLine, readNumberLine, numberLineLimits},
	{Row, readRow, rowLimits},
	{Ring, readRing, ringLimits},
	{Grid, readGrid, gridLimits},
	{Bars, readBars, barsLimits},
	{Venn, readVenn, vennLimits},
	{Balance, readBalance, balanceLimits},
	{Containers, readContainers, containersLimits},
	{Piles, readPiles, pilesLimits},
	{Calendar, readCalendar, calendarLimits},
}

// Kinds are every kind of picture, in the order the format lists them.
func Kinds() []Kind {
	all := make([]Kind, 0, len(kinds))
	for _, entry := range kinds {
		all = append(all, entry.kind)
	}
	return all
}

// LimitsOf says in a sentence what a picture of a kind is held to, in the
// numbers the checks hold it to, and nothing for a kind the format does not
// have. It is what the model is shown beside the kind's example.
func LimitsOf(kind Kind) string {
	for _, entry := range kinds {
		if entry.kind == kind {
			return entry.limits()
		}
	}
	return ""
}

// Parse reads a description into the format, written with the decimal mark of
// the lesson's language. Every rule it breaks is a problem, each named by the
// member that breaks it, so that one more attempt can mend them all.
//
// A description that is no object, or names no kind the format has, is not
// read further and has no picture. Otherwise the picture holds every member
// that could be read, even when others could not: what the picture shows can
// still be held to the wording in the same pass.
func Parse(raw json.RawMessage, decimals Decimals) (Picture, []Problem) {
	reading := &reading{decimals: decimals}
	whole, isObject := decoded(raw)
	if !isObject {
		reading.fault("picture", "must be an object: a description of one of the kinds of picture")
		return nil, reading.problems
	}
	description := reading.object("picture", whole)
	read := readerOf(description)
	if read == nil {
		return nil, reading.problems
	}
	return read(description), reading.problems
}

// readerOf is how a description of the kind it names is read, and nil, with
// the problem said, when it names none the format has.
func readerOf(description *object) func(*object) Picture {
	value, held := description.member("kind")
	text, isText := value.(string)
	for _, entry := range kinds {
		if held && isText && string(entry.kind) == text {
			return entry.read
		}
	}
	rule := "is no kind of picture"
	if !held {
		rule = "is missing"
	}
	description.fault(description.at("kind"), "%s: name one of %s", rule, kindNames())
	return nil
}

// kindNames lists every kind the way a sentence lists things.
func kindNames() string {
	names := make([]string, 0, len(kinds))
	for _, entry := range kinds {
		names = append(names, string(entry.kind))
	}
	return oneOf(names)
}

// KindOf is the kind a description names, as a log line may carry it: the
// kind, None for no description, or Other for one that names no kind the
// format has. Nothing else of the description is judged, since a description
// that could not be read is never kept.
func KindOf(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return None
	}
	var head struct {
		Kind json.RawMessage `json:"kind"`
	}
	var kind string
	if json.Unmarshal(trimmed, &head) != nil || json.Unmarshal(head.Kind, &kind) != nil {
		return Other
	}
	if slices.Contains(Kinds(), Kind(kind)) {
		return kind
	}
	return Other
}

// decoded is a description as JSON values, numbers kept as they were written,
// and whether it is an object at all.
func decoded(raw json.RawMessage) (map[string]any, bool) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil || decoder.More() {
		return nil, false
	}
	members, isObject := value.(map[string]any)
	return members, isObject
}
