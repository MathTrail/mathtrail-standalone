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

func calendarLimits() string {
	return fmt.Sprintf("first from %d for Monday to %d for Sunday; %d to %d days; a mark stands on a day of the month",
		firstWeekday, lastWeekday, fewestDays, mostDays)
}
