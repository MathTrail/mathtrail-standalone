package main

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// The sizes a drawing is made to pass, from the limits the service holds a
// drawing to: one line more than a card shows, one cell wider, and one space
// more in a row than a drawing may hold.
var (
	tooManyLines  = checks.DefaultDrawingLimits().Height + 1
	tooManyCells  = checks.DefaultDrawingLimits().Width + 1
	tooManySpaces = checks.DefaultDrawingLimits().SpaceRun + 1
)

// The labels a defect of the drawing adds: a point the wording names and the
// structure does not, and an object the structure declares and nobody draws.
const (
	strayLabel    = "Q"
	undrawnLabel  = "Z"
	strayWording  = "Point Q is marked."
	undrawnObject = "injected"
)

// drawingOperators are the defects of a drawing: classes D16 to D18.
func drawingOperators() []operator {
	format := []checks.Code{checks.CodeDrawingFormat}
	mismatch := []checks.Code{checks.CodeDrawingMismatch}
	return []operator{
		{ID: "D16a", Class: "D16", Kind: mechanism, Expected: format, Apply: redrawn(tooTall)},
		{ID: "D16b", Class: "D16", Kind: mechanism, Expected: format, Apply: redrawn(tooWide)},
		{ID: "D16c", Class: "D16", Kind: mechanism, Expected: format, Apply: redrawn(insertedAfterFirst(strings.Repeat(" ", tooManySpaces)))},
		{ID: "D16d", Class: "D16", Kind: mechanism, Expected: format, Apply: redrawn(trailingSpaces)},
		{ID: "D17a", Class: "D17", Kind: mechanism, Expected: format, Apply: redrawn(insertedAfterFirst("\u200b"))},
		{ID: "D17b", Class: "D17", Kind: mechanism, Expected: format, Apply: redrawn(insertedAfterFirst("\t"))},
		{ID: "D17c", Class: "D17", Kind: mechanism, Expected: format, Apply: redrawn(insertedAfterFirst("\u00a0"))},
		{ID: "D17d", Class: "D17", Kind: mechanism, Expected: format, Apply: redrawn(insertedAfterFirst("╭"))},
		{ID: "D18a", Class: "D18", Kind: mechanism, Expected: mismatch, Apply: strayPoint},
		{ID: "D18b", Class: "D18", Kind: mechanism, Expected: mismatch, Apply: undrawnAdded},
		{ID: "D18c", Class: "D18", Kind: mechanism, Expected: mismatch, Apply: labelRenamed},
	}
}

// redrawn changes the lines of a host's drawing; a host without one is left
// out.
func redrawn(change func(lines []string) []string) func(m *maker) (mutant, bool) {
	return func(m *maker) (mutant, bool) {
		if !m.host.HasDrawing() {
			return mutant{}, false
		}
		sub := m.start()
		lines := strings.Split(strings.TrimSuffix(sub.Task.Drawing, "\n"), "\n")
		sub.Task.Drawing = strings.Join(change(lines), "\n")
		return mutant{Sub: sub}, true
	}
}

// tooTall adds lines of dashes until the drawing is a line taller than allowed.
func tooTall(lines []string) []string {
	for len(lines) < tooManyLines {
		lines = append(lines, "-")
	}
	return lines
}

// tooWide extends the longest line with dashes until it is a cell wider than
// allowed.
func tooWide(lines []string) []string {
	longest := 0
	for i, line := range lines {
		if utf8.RuneCountInString(line) > utf8.RuneCountInString(lines[longest]) {
			longest = i
		}
	}
	for utf8.RuneCountInString(lines[longest]) < tooManyCells {
		lines[longest] += "-"
	}
	return lines
}

// trailingSpaces ends the first line with two spaces.
func trailingSpaces(lines []string) []string {
	lines[0] += "  "
	return lines
}

// insertedAfterFirst puts a text after the first character of the first line.
func insertedAfterFirst(text string) func(lines []string) []string {
	return func(lines []string) []string {
		_, size := utf8.DecodeRuneInString(lines[0])
		lines[0] = lines[0][:size] + text + lines[0][size:]
		return lines
	}
}

// strayPoint has the wording name a point the structure does not declare.
func strayPoint(m *maker) (mutant, bool) {
	if !m.host.HasDrawing() || m.labelInUse(strayLabel) {
		return mutant{}, false
	}
	sub := m.start()
	question, placed := beforeLastSentence(sub.Task.Question, strayWording)
	if !placed {
		return mutant{}, false
	}
	sub.Task.Question = question
	return mutant{Sub: sub}, true
}

// undrawnAdded declares an object nobody draws or names.
func undrawnAdded(m *maker) (mutant, bool) {
	if !m.host.HasDrawing() || m.labelInUse(undrawnLabel) {
		return mutant{}, false
	}
	sub := m.start()
	sub.Task.DrawingStructure.Objects = append(sub.Task.DrawingStructure.Objects, checks.DrawingObject{ID: undrawnObject, Label: undrawnLabel})
	return mutant{Sub: sub}, true
}

// labelRenamed renames one declared object in the structure only, so that the
// drawing and the wording keep the old label.
func labelRenamed(m *maker) (mutant, bool) {
	if !m.host.HasDrawing() || m.labelInUse(undrawnLabel) {
		return mutant{}, false
	}
	sub := m.start()
	objects := sub.Task.DrawingStructure.Objects
	objects[m.rand.IntN(len(objects))].Label = undrawnLabel
	return mutant{Sub: sub}, true
}

// labelInUse says whether a label already appears in the host's drawing, its
// wording or its structure, where adding it would make no defect.
func (m *maker) labelInUse(label string) bool {
	task := &m.host.Base.Task
	declared := slices.ContainsFunc(task.DrawingStructure.Objects, func(object checks.DrawingObject) bool { return object.Label == label })
	return declared || strings.Contains(task.Drawing, label) || strings.Contains(task.Question, label)
}
