package picture

import "fmt"

// table is a table of cells: a header row, if it has one, and its rows. A
// cell is a label, a time, an ellipsis for the cells left out, or nothing.
type table struct {
	noWords
	cells []string
}

func readTable(o *object) Picture {
	var t table
	width := 0
	if value, held := o.member("header"); held {
		header := readCells(o, o.at("header"), value)
		t.cells = append(t.cells, header...)
		width = len(header)
	}
	for i, value := range o.list("rows", minTableRows, maxTableRows, "rows", true) {
		path := fmt.Sprintf("%s.%d", o.at("rows"), i)
		row := readCells(o, path, value)
		if width == 0 {
			width = len(row)
		}
		if len(row) != width {
			o.fault(path, "must be as long as the header, or as the first row where there is no header")
		}
		t.cells = append(t.cells, row...)
	}
	o.close("a table")
	return t
}

// readCells reads one row of a table, header or not, at its path.
func readCells(o *object, path string, value any) []string {
	items, isList := value.([]any)
	if !isList || len(items) < minTableCells || len(items) > maxTableCells {
		o.fault(path, "must be a list of %d to %d cells", minTableCells, maxTableCells)
		if !isList || len(items) > maxTableCells {
			return nil
		}
	}
	cells := make([]string, len(items))
	for i, item := range items {
		text, isText := item.(string)
		if !isText || !o.decimals.isCell(text) {
			o.fault(fmt.Sprintf("%s.%d", path, i), "must be a cell: a label, a time, %s for cells left out, or empty",
				Ellipsis)
			continue
		}
		cells[i] = text
	}
	return cells
}

func (table) Kind() Kind { return Table }

// Labels are the cells that hold a label.
func (t table) Labels() []string {
	var labels []string
	for _, cell := range t.cells {
		if _, isTime := ReadTime(cell); !isTime && cell != Ellipsis {
			labels = append(labels, cell)
		}
	}
	return present(labels...)
}

// Shown are every cell that holds a label or a time.
func (t table) Shown() []Shown {
	var shown []string
	for _, cell := range t.cells {
		if cell != Ellipsis {
			shown = append(shown, cell)
		}
	}
	return texts(shown...)
}
