package picture

import (
	"fmt"
	"slices"
)

// shapes are rows of shapes with a total beside each: the same shape stands for
// the same number wherever it is drawn.
type shapes struct {
	noWords
	totals []string
}

// shapeNames are the shapes a row may hold.
var shapeNames = []string{"triangle", "square", "circle", "diamond"}

func readShapes(o *object) Picture {
	var s shapes
	for _, row := range o.objects("rows", minShapeRows, maxShapeRows, "rows", true) {
		if row == nil {
			continue
		}
		for i, item := range row.list("shapes", 1, maxShapesInRow, "shapes", true) {
			name, isText := item.(string)
			if !isText || !slices.Contains(shapeNames, name) {
				row.fault(fmt.Sprintf("%s.%d", row.at("shapes"), i), "must be %s", oneOf(shapeNames))
			}
		}
		s.totals = append(s.totals, row.label("total", true))
		row.close("a row of shapes")
	}
	o.close("shapes")
	return &s
}

func (*shapes) Kind() Kind { return Shapes }

func (s *shapes) Labels() []string { return present(s.totals...) }

func (s *shapes) Shown() []Shown { return texts(s.totals...) }
