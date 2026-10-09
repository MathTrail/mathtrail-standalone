package picture

import (
	"fmt"
	"maps"
	"slices"
)

// grid is a square grid whose rows and columns have names, so that a cell is
// named by its row and its column, B2: some cells filled, and some marked
// with a label.
type grid struct {
	rows, cols []string
	filled     []string
	marks      []gridMark
}

// gridMark is a label set in a cell of a grid.
type gridMark struct {
	cell, label string
}

// cellRule is what a cell of a grid must be, as a refusal says it.
const cellRule = "must be a cell of the grid: its row and its column, as B2 is"

func readGrid(o *object) Picture {
	var g grid
	g.rows = readNames(o, "rows")
	g.cols = readNames(o, "cols")
	cells := cellsOf(o, g.rows, g.cols)
	g.filled = readFilled(o, cells)
	g.marks = readGridMarks(o, cells)
	o.close("a grid")
	return &g
}

// readNames reads the names of a grid's rows or of its columns: labels, each
// once.
func readNames(o *object, name string) []string {
	var names []string
	for i, label := range o.labels(name, minGridSide, maxGridSide, true) {
		if slices.Contains(names, label) {
			o.fault(fmt.Sprintf("%s.%d", o.at(name), i), "names a line of the grid another name already names")
			continue
		}
		names = append(names, label)
	}
	return names
}

// cellsOf are the names of every cell of a grid, and nil when its rows or its
// columns could not be read, or name two cells alike: then no cell can be
// judged. A cell's name is its row's run into its column's, so rows 1 and 12
// with columns 22 and 2 would name two cells 122, and the card and the checks
// could take different ones.
func cellsOf(o *object, rows, cols []string) map[string]bool {
	if len(rows) == 0 || len(cols) == 0 {
		return nil
	}
	cells := make(map[string]bool, len(rows)*len(cols))
	for _, row := range rows {
		for _, col := range cols {
			if cells[row+col] {
				o.fault(o.at("cols"), "must name the columns apart from the rows: run together, two cells would "+
					"have one name")
				return nil
			}
			cells[row+col] = true
		}
	}
	return cells
}

// readFilled reads the cells a grid fills, each a cell of the grid, once.
func readFilled(o *object, cells map[string]bool) []string {
	var filled []string
	for i, item := range o.list("filled", 0, maxGridSide*maxGridSide, "cells", false) {
		path := fmt.Sprintf("%s.%d", o.at("filled"), i)
		cell, isText := item.(string)
		switch {
		case !isText || (cells != nil && !cells[cell]):
			o.fault(path, cellRule)
		case slices.Contains(filled, cell):
			o.fault(path, "fills a cell already filled")
		default:
			filled = append(filled, cell)
		}
	}
	return filled
}

// readGridMarks reads the labels a grid sets in its cells, keyed by the cell.
func readGridMarks(o *object, cells map[string]bool) []gridMark {
	entries := o.keyed("marks")
	var marks []gridMark
	for _, cell := range slices.Sorted(maps.Keys(entries)) {
		path := o.at("marks") + ".*"
		label, isText := entries[cell].(string)
		switch {
		case cells != nil && !cells[cell]:
			o.fault(path, cellRule)
		case !isText || !o.decimals.isLabel(label):
			o.fault(path, "%s", labelRule)
		default:
			marks = append(marks, gridMark{cell: cell, label: label})
		}
	}
	return marks
}

func (*grid) Kind() Kind { return Grid }

func (g *grid) Labels() []string {
	labels := slices.Concat(g.rows, g.cols)
	for _, mark := range g.marks {
		labels = append(labels, mark.label)
	}
	return present(labels...)
}

// Shown are the cells the grid fills and marks, and the labels it marks them
// with. The names of its rows and columns name every line alike, and single
// nothing out.
func (g *grid) Shown() []Shown {
	shown := slices.Clone(g.filled)
	for _, mark := range g.marks {
		shown = append(shown, mark.cell, mark.label)
	}
	return texts(shown...)
}
