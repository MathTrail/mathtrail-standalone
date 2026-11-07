package picture

import "fmt"

// The limits a picture is held to, every one of them in this one place. Each
// keeps a picture one a card of 320 pixels draws with its smallest label still
// legible, and none of them is a setting: changing one changes the format.
const (
	// MaxLabelCharacters is the longest a label may be.
	MaxLabelCharacters = 5
	// MaxNoteCharacters is the longest a note under bars may be.
	MaxNoteCharacters = 24
	// MaxTotalCharacters is the longest the total under a picture of a
	// solution may be: written larger than a note, it holds the whole
	// equality the solution comes to.
	MaxTotalCharacters = 32

	minTableRows, maxTableRows   = 1, 8
	minTableCells, maxTableCells = 1, 6

	minTicks, maxTicks = 2, 16
	// A tick's number is drawn by the card, and is held to the length of a
	// label: from -9999 to 99999.
	lowestTick, highestTick = -9999, 99999

	minRowItems, maxRowItems = 2, 12
	maxCopies                = 2

	minRingPlaces, maxRingPlaces = 3, 24

	minGridSide, maxGridSide = 1, 8

	minBars, maxBars     = 1, 4
	minParts, maxParts   = 1, 12
	minLength, maxLength = 1, 60
	maxBraces            = 4
	maxNotes             = 3

	vennSets = 2

	maxOnAPan = 4

	minContainers, maxContainers = 1, 4
	minCapacity, maxCapacity     = 1, 20

	minPiles, maxPiles = 1, 6
	maxPileCount       = 40
	fewestShown        = 2
	minGroup, maxGroup = 1, 40

	firstWeekday, lastWeekday = 1, 7
	fewestDays, mostDays      = 28, 31

	minFlagGroups, maxFlagGroups = 1, 4
	minFlags, maxFlagsInGroup    = 1, 6
	maxFlags                     = 12
	minStripes, maxStripes       = 2, 3

	minLatticeSide, maxLatticeSide = 1, 8
	maxLatticeCells                = 64
	maxLatticeLines                = 48
	maxLatticeDashed               = 8
	maxLatticePoints               = 16
	maxLatticeMarks                = 36
	// A label of a lattice stands in a cell or beside a corner, and two
	// characters are the most a cell of a lattice 8 cells across holds at the
	// card's size; a tile's triangle holds a number of as many digits.
	maxShortLabelCharacters = 2

	minTileSides, maxTileSides = 3, 6
	minTiles, maxTiles         = 2, 12
	// A row runs straight, and its tiles' triangles narrow as it grows longer:
	// a row of hexagons, whose triangles are the narrowest, is held shorter.
	maxTilesInRow    = 6
	maxHexagonsInRow = 4
	maxTileDigits    = 2

	minShapeRows, maxShapeRows = 1, 5
	maxShapesInRow             = 6

	minSteps, maxSteps = 3, 16
	maxStairMarks      = 4
	maxStairJumps      = 4

	// A colour is called by a word or two of the lesson's language, written
	// beside its paint in the key under the picture.
	maxColorWords      = 2
	maxColorCharacters = 16
)

// The limits of each kind, in the words the model is shown them in.

func clockLimits() string {
	return "a time from 0:00 to 23:59, written H:MM; a time after 12:59 is drawn as a twelve-hour face shows it"
}

func tableLimits() string {
	return fmt.Sprintf("%d to %d rows of %d to %d cells, every row as long as the header or the first row; "+
		"a cell is a label, a time, %s for cells left out, or empty", minTableRows, maxTableRows,
		minTableCells, maxTableCells, Ellipsis)
}

func numberLineLimits() string {
	return fmt.Sprintf("%d to %d ticks, from from to to by step, each number from %d to %d; "+
		"a mark stands on a tick, one to a tick", minTicks, maxTicks, lowestTick, highestTick)
}

func rowLimits() string {
	return fmt.Sprintf("%d to %d items; a skip holds nothing, is never first or last, and is never next to "+
		"another skip; copies 1 or %d", minRowItems, maxRowItems, maxCopies)
}

func ringLimits() string {
	return fmt.Sprintf("%d to %d places, numbered from 1 by the card; start at one of them", minRingPlaces, maxRingPlaces)
}

func gridLimits() string {
	return fmt.Sprintf("%d to %d rows and %d to %d columns, each named once; a cell is its row and its column, B2, "+
		"and lies in the grid", minGridSide, maxGridSide, minGridSide, maxGridSide)
}

func barsLimits() string {
	return fmt.Sprintf("%d to %d bars, drawn to one scale; %d to %d parts or segments to a bar, no more shaded than "+
		"parts; a length or a segment's size from %d to %d; at most %d braces to a bar, each within it; "+
		"at most %d notes", minBars, maxBars, minParts, maxParts, minLength, maxLength, maxBraces, maxNotes)
}

func vennLimits() string {
	return fmt.Sprintf("exactly %d sets; every count is a label", vennSets)
}

func balanceLimits() string {
	return fmt.Sprintf("at most %d labels on a pan", maxOnAPan)
}

func containersLimits() string {
	return fmt.Sprintf("%d to %d containers; a capacity from %d to %d, an amount from 0 up to the capacity",
		minContainers, maxContainers, minCapacity, maxCapacity)
}

func pilesLimits() string {
	return fmt.Sprintf("%d to %d piles, a skip among them as in a row; a count from 0 to %d; shown from %d to one "+
		"less than the count; group from %d to %d", minPiles, maxPiles, maxPileCount, fewestShown, minGroup, maxGroup)
}

func flagsLimits() string {
	return fmt.Sprintf("%d to %d groups of %d to %d flags, at most %d flags in all; a flag is %d or %d stripes from "+
		"top to bottom, each a colour colors names or ?; a group is named under it by a colour or a label; colors "+
		"gives each colour it paints with, of %s, the word the question calls it by, at most %d words and %d "+
		"characters", minFlagGroups, maxFlagGroups, minFlags, maxFlagsInGroup, maxFlags, minStripes, maxStripes,
		oneOf(paintNames()), maxColorWords, maxColorCharacters)
}

func calendarLimits() string {
	return fmt.Sprintf("first from %d for Monday to %d for Sunday; %d to %d days; a mark stands on a day of the month",
		firstWeekday, lastWeekday, fewestDays, mostDays)
}

func latticeLimits() string {
	return fmt.Sprintf("%d to %d cells across and down; a cell is [x, y], from [0, 0] at the top left; a line, plain "+
		"or dashed, is [x1, y1, x2, y2], two corners from [0, 0] to [width, height] joined straight, drawn once; two "+
		"lines meet only at a corner; at most %d cells filled, %d lines, %d dashed lines, %d points at corners and %d "+
		"marks in cells; a label of at most %d characters; each point's label needs a cell of its own among the four "+
		"round its corner, held by no mark and by no other point's label", minLatticeSide, maxLatticeSide, maxLatticeCells, maxLatticeLines, maxLatticeDashed,
		maxLatticePoints, maxLatticeMarks, maxShortLabelCharacters)
}

func tilesLimits() string {
	return fmt.Sprintf("%d to %d tiles of %d to %d sides each; each tile lists, for each of its triangles, a whole "+
		"number of at most %d digits, ? or an empty text, clockwise from the side it shares with the tile before it; "+
		"joint, the side it shares with the tile after it, counted the same way, half its sides unless given; a row "+
		"runs straight, of tiles of 4 or 6 sides joined at the side opposite the first, at most %d tiles and %d "+
		"hexagons; a ring goes round once and closes: 4 squares joint 1 or 3, 3 hexagons joint 1 or 5, 6 hexagons "+
		"joint 2 or 4, 6 triangles, or 10 pentagons joint 2 or 3",
		minTiles, maxTiles, minTileSides, maxTileSides, maxTileDigits, maxTilesInRow, maxHexagonsInRow)
}

func shapesLimits() string {
	return fmt.Sprintf("%d to %d rows of 1 to %d shapes, each %s, with a total, a label, beside each row",
		minShapeRows, maxShapeRows, maxShapesInRow, oneOf(shapeNames))
}

func stairsLimits() string {
	return fmt.Sprintf("%d to %d steps from the floor, 0, to the top; at most %d marks, each a label on a step of "+
		"its own, and %d jumps, each from one step to another, drawn once", minSteps, maxSteps, maxStairMarks,
		maxStairJumps)
}
