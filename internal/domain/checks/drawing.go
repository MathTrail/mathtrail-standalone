package checks

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/width"
)

// DrawingLimits are how large a drawing may be.
type DrawingLimits struct {
	// Width is the widest a line may be, in screen cells.
	Width int
	// Height is how many lines a drawing may have.
	Height int
	// SpaceRun is the most spaces a line may hold in a row.
	SpaceRun int
}

// DefaultDrawingLimits are the limits every drawing is held to, set here and
// nowhere else. Thirty cells is what a phone card 353 pixels wide showed
// legibly without scrolling sideways, twelve lines fit the card without
// pushing its buttons off the screen, and a drawing held together by a longer
// run of spaces falls apart in any font that is not monospaced. All three are
// conservative until a live run on real widgets calibrates them.
func DefaultDrawingLimits() DrawingLimits {
	return DrawingLimits{Width: 30, Height: 12, SpaceRun: 20}
}

// DrawingFormat checks that a drawing shows on a phone as it was drawn: no
// wider and no taller than the limits, and made only of characters that every
// monospaced font draws alike. Each rule a drawing breaks is one problem that
// names the lines it breaks it on, so that the refusal stays short however
// broken the drawing is.
//
// An absent or blank drawing has no format to check.
func DrawingFormat(drawing string, limits DrawingLimits) []Problem {
	if strings.TrimSpace(drawing) == "" {
		return nil
	}
	lines := linesOf(drawing)

	var problems []Problem
	if len(lines) > limits.Height {
		problems = append(problems, drawingFormat("task.drawing has %d lines, and a drawing may have at most %d",
			len(lines), limits.Height))
	}
	if wide := linesWhere(lines, func(line string) bool { return cellsOf(line) > limits.Width }); len(wide) > 0 {
		problems = append(problems, drawingFormat("task.drawing is wider than %d cells on %s; "+
			"a wide character takes two cells", limits.Width, onLines(wide)))
	}
	if spaced := linesWhere(lines, func(line string) bool { return longestSpaceRun(line) > limits.SpaceRun }); len(spaced) > 0 {
		problems = append(problems, drawingFormat("task.drawing has more than %d spaces in a row on %s; "+
			"a drawing held together by long runs of spaces falls apart outside a monospaced font",
			limits.SpaceRun, onLines(spaced)))
	}
	if trailing := linesWhere(lines, func(line string) bool { return strings.HasSuffix(line, " ") }); len(trailing) > 0 {
		problems = append(problems, drawingFormat("task.drawing has spaces at the end of %s; remove them, "+
			"since they do not survive being copied and the drawing loses its alignment", onLines(trailing)))
	}
	return append(problems, unfitCharacters(lines)...)
}

// linesOf splits a drawing into its lines. A newline at the very end closes
// the last line rather than opening another.
func linesOf(drawing string) []string {
	return strings.Split(strings.TrimSuffix(drawing, "\n"), "\n")
}

// linesWhere are the numbers of the lines that break a rule, counted from one
// as a person counts them.
func linesWhere(lines []string, breaks func(line string) bool) []int {
	var numbers []int
	for i, line := range lines {
		if breaks(line) {
			numbers = append(numbers, i+1)
		}
	}
	return numbers
}

// cellsOf is how many screen cells a line takes in a monospaced font: two for
// a wide or fullwidth character, one for any other. The characters East Asian
// fonts may draw either way — box drawing among them — count as one, the way
// monospaced fonts draw them everywhere else.
func cellsOf(line string) int {
	cells := 0
	for _, r := range line {
		switch width.LookupRune(r).Kind() {
		case width.EastAsianWide, width.EastAsianFullwidth:
			cells += 2
		default:
			cells++
		}
	}
	return cells
}

// longestSpaceRun is the most spaces a line holds in a row.
func longestSpaceRun(line string) int {
	longest, run := 0, 0
	for _, r := range line {
		if r != ' ' {
			run = 0
			continue
		}
		run++
		longest = max(longest, run)
	}
	return longest
}

// drawable says whether a drawing may use a character: printable ASCII, and
// the arrows, box drawing, block elements and geometric shapes a frame is made
// of. Anything else renders differently from one font to the next.
func drawable(r rune) bool {
	switch {
	case r >= 0x0020 && r <= 0x007E: // printable ASCII
	case r >= 0x2190 && r <= 0x21FF: // arrows
	case r >= 0x2500 && r <= 0x257F: // box drawing
	case r >= 0x2580 && r <= 0x259F: // block elements
	case r >= 0x25A0 && r <= 0x25FF: // geometric shapes
	default:
		return false
	}
	return true
}

// A characterFault is a kind of character a drawing may not use: what a
// refusal calls it, how to tell one, and what to do instead.
type characterFault struct {
	name, instead string
	is            func(r rune) bool
}

// characterFaults are the kinds of character a drawing may not use. A
// character is reported under the first kind it fits, so the most particular
// name is the one the model reads: every one of them makes the drawing look
// different to the check and to the child, which is the whole attack surface
// of a picture made of text.
var characterFaults = []characterFault{
	{"tabs", "use spaces", func(r rune) bool { return r == '\t' }},
	{"control characters", `separate the lines with \n and nothing else`, unicode.IsControl},
	{"bidi controls", "a drawing is laid out left to right as it stands",
		func(r rune) bool { return unicode.Is(unicode.Bidi_Control, r) }},
	{"invisible characters", "remove them", func(r rune) bool { return unicode.Is(unicode.Cf, r) }},
	{"spaces other than the plain space", "use the plain space", func(r rune) bool { return unicode.Is(unicode.Zs, r) }},
	{"combining marks", "use characters that stand on their own", func(r rune) bool { return unicode.Is(unicode.M, r) }},
	{"characters a drawing may not use", "draw with ASCII, box drawing, block elements, geometric shapes and arrows",
		func(rune) bool { return true }},
}

// unfitCharacters are the problems with the characters of a drawing: one for
// each kind it uses and may not, naming the lines and the characters.
func unfitCharacters(lines []string) []Problem {
	found := make([]foundCharacters, len(characterFaults))
	for i, line := range lines {
		for _, r := range line {
			if kind := faultOf(r); kind >= 0 {
				found[kind].add(i+1, r)
			}
		}
	}

	var problems []Problem
	for kind, fault := range characterFaults {
		if where := found[kind]; len(where.lines) > 0 {
			problems = append(problems, drawingFormat("task.drawing has %s on %s (%s); %s",
				fault.name, onLines(where.lines), where.codePoints(), fault.instead))
		}
	}
	return problems
}

// faultOf is the kind of character a drawing may not use that a character
// is, as its place in characterFaults, or -1 for a character a drawing may use.
func faultOf(r rune) int {
	if drawable(r) {
		return -1
	}
	return slices.IndexFunc(characterFaults, func(fault characterFault) bool { return fault.is(r) })
}

// foundCharacters are where the characters of one kind were found: on which
// lines, and which characters, each counted once.
type foundCharacters struct {
	lines []int
	runes []rune
	seen  map[rune]bool
}

// add records one character found on a line.
func (f *foundCharacters) add(line int, r rune) {
	if len(f.lines) == 0 || f.lines[len(f.lines)-1] != line {
		f.lines = append(f.lines, line)
	}
	if f.seen == nil {
		f.seen = map[rune]bool{}
	}
	if !f.seen[r] {
		f.seen[r] = true
		f.runes = append(f.runes, r)
	}
}

// codePoints names the characters found by their code points, the only name an
// invisible character has, and stops after the first few.
func (f *foundCharacters) codePoints() string {
	shown := f.runes[:min(len(f.runes), mostNamed)]
	names := make([]string, len(shown))
	for i, r := range shown {
		names[i] = fmt.Sprintf("%U", r)
	}
	return inWords(names, len(f.runes)-len(shown))
}

// onLines names the lines of a drawing by number and stops after the first
// few: the model needs to know where to look, not a list as long as the
// drawing.
func onLines(numbers []int) string {
	shown := numbers[:min(len(numbers), mostNamed)]
	names := make([]string, len(shown))
	for i, number := range shown {
		names[i] = strconv.Itoa(number)
	}
	if len(numbers) == 1 {
		return "line " + names[0]
	}
	return "lines " + inWords(names, len(numbers)-len(shown))
}

// mostNamed is how many lines or characters a refusal names before it only
// counts the rest.
const mostNamed = 5

// drawingFormat is a problem with the format of a drawing.
func drawingFormat(format string, args ...any) Problem {
	return Problem{Code: CodeDrawingFormat, Message: fmt.Sprintf(format, args...)}
}
