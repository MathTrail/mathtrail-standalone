package picture

import (
	"fmt"
	"regexp"
)

// tiles is a row or a ring of identical tiles, each a regular polygon cut into
// triangles from its middle to its sides, with a number, ? or nothing in each
// triangle: a tile shares one of its sides with the tile after it.
type tiles struct {
	noWords
	labels []string
}

// tileNumber is what a tile's triangle may hold besides ? and nothing.
var tileNumber = regexp.MustCompile(fmt.Sprintf(`^[0-9]{1,%d}$`, maxTileDigits))

func readTiles(o *object) Picture {
	var t tiles
	sides, _ := o.whole("sides", minTileSides, maxTileSides, true)
	layout := readLayout(o)
	// A tile whose sides could not be read is joined at a side of the largest
	// tile, so that only what is wrong in itself is said to be.
	most := sides - 1
	if sides == 0 {
		most = maxTileSides - 1
	}
	joint, jointRead := sides/2, true
	if _, given := o.member("joint"); given {
		joint, jointRead = o.whole("joint", 1, most, false)
	}
	items := o.list("tiles", minTiles, maxTiles, "tiles", true)
	t.labels = readTriangles(o, items, sides)
	if sides > 0 && jointRead && len(items) > 0 && layout != "" &&
		!tilesLaidOut(sides, joint, len(items), layout == "ring") {
		o.fault(o.at("layout"), "%s", layoutRule(layout))
	}
	o.close("tiles")
	return &t
}

// readLayout reads how tiles are laid, row or ring, and the empty text when
// it is left out or is neither.
func readLayout(o *object) string {
	value, held := o.member("layout")
	layout, _ := value.(string)
	switch {
	case !held:
		o.missing("layout", true)
	case layout != "row" && layout != "ring":
		o.fault(o.at("layout"), "must be %s", oneOf([]string{"row", "ring"}))
		return ""
	}
	return layout
}

// readTriangles reads the triangles of every tile, each a whole number of at
// most so many digits, ? or the empty text, as many to a tile as it has sides.
func readTriangles(o *object, items []any, sides int) []string {
	var labels []string
	for i, item := range items {
		path := fmt.Sprintf("%s.%d", o.at("tiles"), i)
		triangles, isList := item.([]any)
		if !isList || (sides > 0 && len(triangles) != sides) {
			o.fault(path, "must be a list of as many triangles as a tile has sides, from the side it shares with "+
				"the tile before it, clockwise; a triangle left blank is the empty text")
			continue
		}
		for j, triangle := range triangles {
			text, isText := triangle.(string)
			if !isText || (text != "" && text != "?" && !tileNumber.MatchString(text)) {
				o.fault(fmt.Sprintf("%s.%d", path, j), "must be a whole number of at most %d digits, ? or the "+
					"empty text", maxTileDigits)
				continue
			}
			labels = append(labels, text)
		}
	}
	return labels
}

// turnOf is how far each tile is turned from the one before it, in sides of a
// tile, from less than half its sides back to half of them on: a tile laid
// against side joint of the one before faces it with its own first side, so
// it is turned half round and joint sides on. It is counted in halves of a
// side, since a tile of an odd number of sides turns by halves.
func turnOf(sides, joint int) int {
	turn := (2*joint + sides) % (2 * sides)
	if turn > sides {
		turn -= 2 * sides
	}
	return turn
}

// tilesLaidOut says whether tiles lie as their layout says: a row runs
// straight on and holds at most so many tiles, fewer of hexagons, whose
// triangles are narrower; a ring goes round once and closes, its last tile
// against its first.
func tilesLaidOut(sides, joint, count int, ring bool) bool {
	turn := turnOf(sides, joint)
	if ring {
		return count*max(turn, -turn) == 2*sides
	}
	most := maxTilesInRow
	if sides == 6 {
		most = maxHexagonsInRow
	}
	return turn == 0 && count <= most
}

// layoutRule is what tiles of a layout must be, as a refusal says it, and
// never how many there are: the count may be the answer.
func layoutRule(layout string) string {
	if layout == "row" {
		return fmt.Sprintf("must run straight: a row of tiles of 4 or 6 sides, each against the side opposite its "+
			"first, at most %d tiles and %d hexagons", maxTilesInRow, maxHexagonsInRow)
	}
	return "must go round once and close, its last tile against its first: 4 squares, 3 or 6 hexagons, 6 " +
		"triangles or 10 pentagons, with a joint that closes them"
}

func (*tiles) Kind() Kind { return Tiles }

func (t *tiles) Labels() []string { return present(t.labels...) }

func (t *tiles) Shown() []Shown { return texts(t.Labels()...) }
