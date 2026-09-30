package main

import (
	"fmt"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// sample is one drawing of the calibration set: what it tests, what to look at
// on the card, and the picture with its size in screen cells and lines.
type sample struct {
	name    string
	look    string
	width   int
	height  int
	drawing string
}

// ladders are the widths and the heights the set tries: from a little under
// the limits a drawing is held to now to well past them, so that the limits
// the ratings choose are among them.
func ladders() (widths, heights []int) {
	limits := checks.DefaultDrawingLimits()
	for _, step := range []int{-6, -2, 0, 2, 4, 6, 10, 14} {
		widths = append(widths, limits.Width+step)
	}
	for _, step := range []int{0, 4, 8} {
		heights = append(heights, limits.Height+step)
	}
	return widths, heights
}

// classWidth is how wide the drawings of the character classes are: a row of
// four cells of a grid, each six characters inside.
const classWidth = 29

// calibrationSet is every drawing the server shows, in the order they are
// numbered: the widths, the heights, the classes of characters a drawing may
// use besides ASCII, and two frames the model is given, as it is given them.
func calibrationSet(frames []content.Frame) []sample {
	widths, heights := ladders()
	set := make([]sample, 0, len(widths)+len(heights)+8)
	for _, width := range widths {
		set = append(set, sample{
			name:    fmt.Sprintf("a box %d cells wide", width),
			look:    "Is the whole box on the card, with its right edge straight, or does it scroll sideways?",
			width:   width,
			height:  4,
			drawing: ruled(width),
		})
	}
	for _, height := range heights {
		set = append(set, sample{
			name:    fmt.Sprintf("a box %d lines tall", height),
			look:    "How much of the card is left for the options below it on this screen?",
			width:   tallWidth,
			height:  height,
			drawing: tall(height),
		})
	}
	set = append(set, characterClasses()...)
	for _, name := range []string{"grid-4x4", "venn"} {
		for _, frame := range frames {
			if frame.Name == name {
				set = append(set, framed(frame))
			}
		}
	}
	return set
}

// ruled is a box width cells wide with its columns numbered inside: the units
// on one line and the tens under them, so that where the card cuts it off, or
// where its right edge drifts, can be read off the numbers.
func ruled(width int) string {
	inner := width - 2
	var units, tens strings.Builder
	for column := 1; column <= inner; column++ {
		units.WriteByte(byte('0' + column%10))
		if column%10 == 0 {
			tens.WriteByte(byte('0' + column/10%10))
		} else {
			tens.WriteByte('.')
		}
	}
	edge := "+" + strings.Repeat("-", inner) + "+"
	return strings.Join([]string{edge, "|" + units.String() + "|", "|" + tens.String() + "|", edge}, "\n")
}

// tallWidth is how wide the tall boxes are: narrow enough to fit any card, so
// that only their height is under test.
const tallWidth = 24

// tall is a box height lines tall with each line numbered, so that how much
// of it a card shows before its options, or the screen, run out can be read
// off the numbers.
func tall(height int) string {
	inner := tallWidth - 2
	edge := "+" + strings.Repeat("-", inner) + "+"
	lines := []string{edge}
	for line := 2; line < height; line++ {
		label := fmt.Sprintf("line %d", line)
		lines = append(lines, "|"+label+strings.Repeat(".", inner-len(label))+"|")
	}
	return strings.Join(append(lines, edge), "\n")
}

// characterClasses are the characters a drawing may use besides ASCII, each
// class under a row of ASCII digits as long as its lines: a class a font draws
// wider or narrower than a digit leaves the digits' right edge.
func characterClasses() []sample {
	const look = "Do the right ends of all lines stand exactly under the last digit of the top line?"
	classes := []struct {
		name  string
		lines []string
	}{
		{"box drawing", []string{
			"┌──────┬──────┬──────┬──────┐",
			"│ A    │ B    │ C    │ D    │",
			"├──────┼──────┼──────┼──────┤",
			"│ E    │ F    │ G    │ H    │",
			"└──────┴──────┴──────┴──────┘",
		}},
		{"block elements", []string{
			pattern("█", classWidth),
			pattern("▓", 10) + pattern("▒", 10) + pattern("░", classWidth-20),
			pattern("▀", 14) + pattern("▄", classWidth-14),
			pattern("▌▐", classWidth),
		}},
		{"geometric shapes", []string{
			pattern("■□", classWidth),
			pattern("▲△", classWidth),
			pattern("●○", classWidth),
			pattern("◆◇", classWidth),
		}},
		{"arrows", []string{
			pattern("←→", classWidth),
			pattern("↑↓", classWidth),
			pattern("↔↕", classWidth),
			pattern("⇐⇒", classWidth),
		}},
	}
	samples := make([]sample, 0, len(classes))
	for _, class := range classes {
		lines := append([]string{digits(classWidth)}, class.lines...)
		samples = append(samples, sample{
			name:    class.name + " against ASCII digits",
			look:    look,
			width:   classWidth,
			height:  len(lines),
			drawing: strings.Join(lines, "\n"),
		})
	}
	return samples
}

// pattern is the characters of cycle repeated until the line is count
// characters long.
func pattern(cycle string, count int) string {
	runes := []rune(cycle)
	var line strings.Builder
	for at := range count {
		line.WriteRune(runes[at%len(runes)])
	}
	return line.String()
}

// digits is a row of ASCII digits count long: 1 to 9, then 0, over and over.
func digits(count int) string {
	return pattern("1234567890", count)
}

// framed is a frame the model is given for its drawings, as it is given it.
func framed(frame content.Frame) sample {
	lines := strings.Split(frame.Drawing, "\n")
	width := 0
	for _, line := range lines {
		width = max(width, len([]rune(line)))
	}
	return sample{
		name:    "the frame " + frame.Name + ", as the model is given it",
		look:    "Is it straight and easy to read, as a child would see it on a task?",
		width:   width,
		height:  len(lines),
		drawing: frame.Drawing,
	}
}
