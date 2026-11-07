package picture

import (
	"fmt"
	"slices"
	"unicode/utf8"
)

// lattice is a figure on squared paper: a field of cells, some of them filled,
// lines drawn from corner to corner, lines drawn dashed — a fold, a cut, an
// axis — and labels at corners and in cells. It is the kind whose facts a task
// may hold on the picture alone: the walls of a maze, the sides of the
// triangles to count, the faces of a net.
type lattice struct {
	noWords
	points, marks []latticeLabel
}

// latticeLabel is a label set at a corner of a lattice, or in a cell of it,
// and where it stands.
type latticeLabel struct {
	at    [2]int
	label string
}

// segment is a line of a lattice, from one corner to another.
type segment [4]int

func readLattice(o *object) Picture {
	var l lattice
	width, _ := o.whole("width", minLatticeSide, maxLatticeSide, true)
	height, _ := o.whole("height", minLatticeSide, maxLatticeSide, true)
	// A lattice whose size could not be read is judged against the largest
	// one, so that only what is wrong in itself is said to be.
	across, down := width, height
	if across == 0 {
		across = maxLatticeSide
	}
	if down == 0 {
		down = maxLatticeSide
	}
	readFilledCells(o, across, down)
	var drawn []segment
	lines := readLines(o, "lines", maxLatticeLines, across, down, &drawn)
	readLines(o, "dashed", maxLatticeDashed, across, down, &drawn)
	crossOnlyAtCorners(o, lines)
	l.points = readLatticeLabels(o, "points", maxLatticePoints, [2]int{across, down}, false)
	l.marks = readLatticeLabels(o, "marks", maxLatticeMarks, [2]int{across - 1, down - 1}, true)
	if width > 0 && height > 0 && pointLabelsOf(width, height, l.points, l.marks) == nil {
		o.fault(o.at("points"), "must leave each point's label a cell of its own among the four round its corner, "+
			"one no mark holds")
	}
	o.close("a lattice")
	return &l
}

// The rules of what a cell, a corner and a line of a lattice must be, as a
// refusal says it, and what each number of one must be: never the lattice's
// own size, which may be the answer.
const (
	cellShapeRule   = "must be [x, y]: two whole numbers that name a cell of the lattice, from [0, 0] at the top left"
	cornerShapeRule = "must be [x, y]: two whole numbers that name a corner of the lattice, from [0, 0] at the top left"
	lineShapeRule   = "must be [x1, y1, x2, y2]: two corners of the lattice, from [0, 0] at the top left to " +
		"[width, height]"
	acrossRule = "must be a whole number from 0 that lies within the lattice's width"
	downRule   = "must be a whole number from 0 that lies within the lattice's height"
)

// numbersOf reads a list of whole numbers, as many as ends has, the nth from 0
// to ends[n], saying what is wrong with each number at its own place, and
// whether all of them could be read.
func numbersOf(o *object, path string, item any, ends []int, shapeRule string) ([]int, bool) {
	numbers, isList := item.([]any)
	if !isList || len(numbers) != len(ends) {
		o.fault(path, "%s", shapeRule)
		return nil, false
	}
	read := make([]int, len(ends))
	valid := true
	for j, end := range ends {
		number, isWhole := wholeOf(numbers[j])
		if !isWhole || number < 0 || number > end {
			rule := acrossRule
			if j%2 == 1 {
				rule = downRule
			}
			o.fault(fmt.Sprintf("%s.%d", path, j), "%s", rule)
			valid = false
			continue
		}
		read[j] = number
	}
	return read, valid
}

// readFilledCells reads the cells a lattice fills, each a cell of it, once.
func readFilledCells(o *object, across, down int) {
	var read [][2]int
	for i, item := range o.list("cells", 0, maxLatticeCells, "cells", false) {
		path := fmt.Sprintf("%s.%d", o.at("cells"), i)
		numbers, valid := numbersOf(o, path, item, []int{across - 1, down - 1}, cellShapeRule)
		if !valid {
			continue
		}
		cell := [2]int{numbers[0], numbers[1]}
		if slices.Contains(read, cell) {
			o.fault(path, "names a cell already named")
			continue
		}
		read = append(read, cell)
	}
}

// placedLine is a line of a lattice and its place in the list it was read
// from, so that what is wrong with it is said at that place.
type placedLine struct {
	segment segment
	at      int
}

// readLines reads a list of lines, each from one corner of the lattice to
// another, straight, and drawn once among every line read so far, plain or
// dashed, from either end.
func readLines(o *object, name string, most, across, down int, drawn *[]segment) []placedLine {
	var read []placedLine
	for i, item := range o.list(name, 0, most, "lines", false) {
		path := fmt.Sprintf("%s.%d", o.at(name), i)
		numbers, valid := numbersOf(o, path, item, []int{across, down, across, down}, lineShapeRule)
		if !valid {
			continue
		}
		line := segment{numbers[0], numbers[1], numbers[2], numbers[3]}
		reversed := segment{line[2], line[3], line[0], line[1]}
		switch {
		case line[0] == line[2] && line[1] == line[3]:
			o.fault(path, "must join two different corners")
		case slices.Contains(*drawn, line) || slices.Contains(*drawn, reversed):
			o.fault(path, "draws a line already drawn")
		default:
			*drawn = append(*drawn, line)
			read = append(read, placedLine{segment: line, at: i})
		}
	}
	return read
}

// crossOnlyAtCorners refuses a line that meets another anywhere but at a
// corner of the lattice, or runs along it: a program that counts a figure by
// its corners would miss a corner that lies between them.
func crossOnlyAtCorners(o *object, lines []placedLine) {
	for i, line := range lines {
		for _, earlier := range lines[:i] {
			if meetsBetweenCorners(line.segment, earlier.segment) {
				o.fault(fmt.Sprintf("%s.%d", o.at("lines"), line.at), "meets another line between corners of the "+
					"lattice: lines may meet only at its corners, and never run along each other")
				break
			}
		}
	}
}

// meetsBetweenCorners says whether two lines share a point that is no corner
// of the lattice, or share a stretch of their length.
func meetsBetweenCorners(one, other segment) bool {
	px, py := one[0], one[1]
	rx, ry := one[2]-one[0], one[3]-one[1]
	qx, qy := other[0], other[1]
	sx, sy := other[2]-other[0], other[3]-other[1]
	d := rx*sy - ry*sx
	wx, wy := qx-px, qy-py
	if d == 0 {
		if wx*ry-wy*rx != 0 {
			return false
		}
		// On one straight line: they run along each other when the stretch
		// they share, measured along the first, is longer than nothing.
		t0, t1 := wx*rx+wy*ry, (wx+sx)*rx+(wy+sy)*ry
		return max(min(t0, t1), 0) < min(max(t0, t1), rx*rx+ry*ry)
	}
	t, u := wx*sy-wy*sx, wx*ry-wy*rx
	if d < 0 {
		d, t, u = -d, -t, -u
	}
	if t < 0 || t > d || u < 0 || u > d {
		return false
	}
	// They cross at the point one reaches t/d of the way along: a corner when
	// both of its numbers are whole.
	return (px*d+t*rx)%d != 0 || (py*d+t*ry)%d != 0
}

// shortLabelRule is what a label of a lattice must be, as a refusal says it.
var shortLabelRule = fmt.Sprintf("must be a label of at most %d characters: a Latin capital or two, a number "+
	"of one or two digits, or ?", maxShortLabelCharacters)

// readLatticeLabels reads the labels set at corners, or in cells, each
// within ends: each where it stands, once, with a label that a corner may
// leave out.
func readLatticeLabels(o *object, name string, most int, ends [2]int, needLabel bool) []latticeLabel {
	shapeRule := cornerShapeRule
	if needLabel {
		shapeRule = cellShapeRule
	}
	var read []latticeLabel
	var places [][2]int
	for _, item := range o.objects(name, 0, most, "labels", false) {
		if item == nil {
			continue
		}
		at, placed := readPlace(item, ends, shapeRule, places)
		if placed {
			places = append(places, at)
		}
		label := item.label("label", needLabel)
		if utf8.RuneCountInString(label) > maxShortLabelCharacters {
			item.fault(item.at("label"), "%s", shortLabelRule)
			label = ""
		}
		item.close("a label of a lattice")
		if placed {
			read = append(read, latticeLabel{at: at, label: label})
		}
	}
	return read
}

// readPlace reads where a label of a lattice stands, within ends and where
// no label taken stands, and whether it could be read.
func readPlace(item *object, ends [2]int, shapeRule string, taken [][2]int) ([2]int, bool) {
	value, held := item.member("at")
	if !held {
		item.missing("at", true)
		return [2]int{}, false
	}
	numbers, valid := numbersOf(item, item.at("at"), value, ends[:], shapeRule)
	if !valid {
		return [2]int{}, false
	}
	at := [2]int{numbers[0], numbers[1]}
	if slices.Contains(taken, at) {
		item.fault(item.at("at"), "is where another label already stands")
		return [2]int{}, false
	}
	return at, true
}

// quarters are the four cells round a corner, as steps from the corner's own
// cell, the one below and to the right of it, in the order a point's label is
// given them: above and to the right, above and to the left, below and to the
// right, below and to the left.
var quarters = [4][2]int{{0, -1}, {-1, -1}, {0, 0}, {-1, 0}}

// pointLabelsOf gives each labelled point a cell of its own among the four
// round its corner, one no mark holds, a cell of the paper before one past its
// edge, and moves a label already given a cell to another of its own where a
// label beside it has none left; it is the card's own way of setting them,
// and nil when some label has no cell.
func pointLabelsOf(width, height int, points, marks []latticeLabel) [][2]int {
	seats := seating{choices: cellsRound(width, height, points, marks), holder: map[[2]int]int{}}
	seats.chosen = make([][2]int, len(seats.choices))
	for i := range seats.choices {
		if !seats.seat(i, map[[2]int]bool{}) {
			return nil
		}
	}
	return seats.chosen
}

// cellsRound are the cells each labelled point's label may stand in, those
// of the paper first, each in the order of quarters, and none a mark holds.
func cellsRound(width, height int, points, marks []latticeLabel) [][][2]int {
	marked := map[[2]int]bool{}
	for _, mark := range marks {
		marked[mark.at] = true
	}
	var choices [][][2]int
	for _, point := range points {
		if point.label == "" {
			continue
		}
		var on, off [][2]int
		for _, quarter := range quarters {
			cell := [2]int{point.at[0] + quarter[0], point.at[1] + quarter[1]}
			onThePaper := cell[0] >= 0 && cell[1] >= 0 && cell[0] < width && cell[1] < height
			switch {
			case marked[cell]:
			case onThePaper:
				on = append(on, cell)
			default:
				off = append(off, cell)
			}
		}
		choices = append(choices, slices.Concat(on, off))
	}
	return choices
}

// seating is the cells given to labels so far: whom each cell is given to,
// and the cell each label holds.
type seating struct {
	choices [][][2]int
	holder  map[[2]int]int
	chosen  [][2]int
}

// seat gives label i a cell, moving a label that holds one to another of its
// own where it can be moved, and says whether it could.
func (s *seating) seat(i int, tried map[[2]int]bool) bool {
	for _, cell := range s.choices[i] {
		if tried[cell] {
			continue
		}
		tried[cell] = true
		held, taken := s.holder[cell]
		if !taken || s.seat(held, tried) {
			s.holder[cell] = i
			s.chosen[i] = cell
			return true
		}
	}
	return false
}

func (*lattice) Kind() Kind { return Lattice }

func (l *lattice) Labels() []string {
	var labels []string
	for _, set := range [][]latticeLabel{l.points, l.marks} {
		for _, item := range set {
			labels = append(labels, item.label)
		}
	}
	return present(labels...)
}

// Shown are the lattice's labels. Its cells, lines and corners are drawn, not
// written: a count of them is the child's to make.
func (l *lattice) Shown() []Shown {
	return texts(l.Labels()...)
}
