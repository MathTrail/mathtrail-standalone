package checks_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// The drawings a frame gives: a number line, a grid, a balance, two jugs and a
// clock face, each within the limits and made of what a drawing may use.
var wellMadeDrawings = map[string]string{
	"a number line": "" +
		"0   A       B      10\n" +
		"├───┼───────┼──────┤",
	"a grid": "" +
		"┌───┬───┬───┐\n" +
		"│ A │   │ B │\n" +
		"├───┼───┼───┤\n" +
		"│   │ C │   │\n" +
		"└───┴───┴───┘",
	"a balance": "" +
		"    ▼\n" +
		" ▄▄▄▄▄▄▄▄▄\n" +
		" A   ▲   B",
	"two jugs": "" +
		" A       B\n" +
		"│   │ → │   │\n" +
		"│▒▒▒│   │   │\n" +
		"└───┘   └───┘",
	"a clock face": "" +
		"   12\n" +
		" 9  ●→ 3\n" +
		"    6",
	"a line as wide as a line may be":           strings.Repeat("─", 30),
	"as many lines as a drawing may have":       strings.TrimSuffix(strings.Repeat("│\n", 12), "\n"),
	"as many spaces in a row as allowed":        "A" + strings.Repeat(" ", 20) + "B",
	"a newline at the end, which opens no line": strings.Repeat("│\n", 12),
	"every character a drawing may use":         linesOf(checks.DrawingCharacters, 30),
}

// linesOf is the characters of text in lines of at most width characters.
func linesOf(text string, width int) string {
	var lines []string
	for runes := []rune(text); len(runes) > 0; {
		cut := min(width, len(runes))
		lines = append(lines, string(runes[:cut]))
		runes = runes[cut:]
	}
	return strings.Join(lines, "\n")
}

func TestAWellMadeDrawingPasses(t *testing.T) {
	t.Parallel()

	for name, drawing := range wellMadeDrawings {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if problems := checks.DrawingFormat(drawing, checks.DefaultDrawingLimits()); len(problems) != 0 {
				t.Errorf("DrawingFormat() = %v, want no problems", problems)
			}
		})
	}
}

// Every rule refuses a drawing on its own, with a problem of its own that says
// which rule and where.
func TestEachFormatRuleRefusesOnItsOwn(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, drawing, want string
	}{
		{"a line too many", strings.Repeat("│\n", 13), "task.drawing has 13 lines, and a drawing may have at most 12"},
		{"a line a cell too wide", strings.Repeat("─", 31), "wider than 30 cells on line 1"},
		{"a space too many in a row", "A" + strings.Repeat(" ", 21) + "B", "more than 20 spaces in a row on line 1"},
		{"a space at the end of a line", "A───B \n│", "spaces at the end of line 1"},
		{"a tab", "A\t───B", "tabs on line 1 (U+0009)"},
		{"a carriage return", "A───B\r\n│", "control characters on line 1 (U+000D)"},
		{"a bell", "A───B\u0007", "control characters on line 1 (U+0007)"},
		{"a right-to-left override", "\u202EA───B", "bidi controls on line 1 (U+202E)"},
		{"a left-to-right mark", "A\u200E───B", "bidi controls on line 1 (U+200E)"},
		{"a directional isolate", "A\u2066───B", "bidi controls on line 1 (U+2066)"},
		{"a zero-width space", "A───\u200BB", "invisible characters on line 1 (U+200B)"},
		{"a byte order mark", "\uFEFFA───B", "invisible characters on line 1 (U+FEFF)"},
		{"a non-breaking space", "A\u00A0───B", "spaces other than the plain space on line 1 (U+00A0)"},
		{"an ideographic space", "A\u3000───B", "spaces other than the plain space on line 1 (U+3000)"},
		{"a combining mark", "A\u0301───B", "combining marks on line 1 (U+0301)"},
		{"an emoji", "A───🍎", "characters a drawing may not use on line 1 (U+1F34E)"},
		{"a Cyrillic letter that looks Latin", "А───B", "characters a drawing may not use on line 1 (U+0410)"},
		{"a Chinese character", "A───猫", "characters a drawing may not use on line 1 (U+732B)"},
		// Box drawing, blocks, shapes and arrows that a common monospaced font
		// lacks, and draws from another font out of line with the rest.
		{"a double arrow", "A───⇒B", "characters a drawing may not use on line 1 (U+21D2)"},
		{"a rounded corner", "╭───B", "characters a drawing may not use on line 1 (U+256D)"},
		{"a diagonal", "A╲B", "characters a drawing may not use on line 1 (U+2572)"},
		{"a heavy line", "A━━━B", "characters a drawing may not use on line 1 (U+2501)"},
		{"a diamond", "A─◆─B", "characters a drawing may not use on line 1 (U+25C6)"},
		{"a lower block", "A▁▁B", "characters a drawing may not use on line 1 (U+2581)"},
		{"a wide square", "A◽B", "characters a drawing may not use on line 1 (U+25FD)"},
		{"an arrow asked to be an emoji", "A→\uFE0FB", "combining marks on line 1 (U+FE0F)"},
		{"a double-headed arrow, which is also an emoji", "A↔B", "characters a drawing may not use on line 1 (U+2194)"},
		{"a small square, which is also an emoji", "A▪B", "characters a drawing may not use on line 1 (U+25AA)"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.DrawingFormat(test.drawing, checks.DefaultDrawingLimits())
			if len(problems) != 1 {
				t.Fatalf("DrawingFormat() = %v, want exactly one problem", problems)
			}
			if problems[0].Code != checks.CodeDrawingFormat {
				t.Errorf("code = %q, want %q", problems[0].Code, checks.CodeDrawingFormat)
			}
			if !strings.Contains(problems[0].Message, test.want) {
				t.Errorf("message = %q, want it to say %q", problems[0].Message, test.want)
			}
		})
	}
}

// A line is as wide as the cells it takes on a screen, not as long as its
// characters: a Chinese character takes two, a box-drawing line one.
func TestTheWidthIsCountedInScreenCells(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, line string
		tooWide    bool
	}{
		{"30 box-drawing lines take 30 cells", strings.Repeat("─", 30), false},
		{"15 Chinese characters take 30 cells", strings.Repeat("猫", 15), false},
		{"16 Chinese characters take 32 cells", strings.Repeat("猫", 16), true},
		{"15 fullwidth letters take 30 cells", strings.Repeat("Ａ", 15), false},
		{"16 fullwidth letters take 32 cells", strings.Repeat("Ａ", 16), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.DrawingFormat(test.line, checks.DefaultDrawingLimits())
			if mentions(problems, "wider than 30 cells") != test.tooWide {
				t.Errorf("DrawingFormat() = %v, want too wide %v", problems, test.tooWide)
			}
		})
	}
}

// Every character a drawing may use takes one cell, as ASCII does: a drawing
// of them is as wide as it is long, and lines of it stand under one another.
func TestEveryDrawingCharacterTakesOneCell(t *testing.T) {
	t.Parallel()

	for _, r := range checks.DrawingCharacters {
		line := strings.Repeat(string(r), 30)
		if problems := checks.DrawingFormat(line, checks.DefaultDrawingLimits()); len(problems) != 0 {
			t.Errorf("thirty of U+%04X = %v, want them thirty cells and allowed", r, problems)
		}
		if !mentions(checks.DrawingFormat(line+string(r), checks.DefaultDrawingLimits()), "wider than 30 cells") {
			t.Errorf("thirty-one of U+%04X fit thirty cells, want each one cell", r)
		}
	}
}

// However many lines break a rule and however many characters of a kind a
// drawing uses, the rule is one problem, naming the first few of each: the
// model needs to know where to look, not a refusal as long as the drawing.
func TestEachRuleIsOneProblemNamingTheFirstFew(t *testing.T) {
	t.Parallel()

	tabs := checks.DrawingFormat(strings.Repeat("A\tB\n", 8), checks.DefaultDrawingLimits())
	if len(tabs) != 1 || !strings.Contains(tabs[0].Message, "tabs on lines 1, 2, 3, 4, 5 and 3 more (U+0009)") {
		t.Errorf("eight lines with a tab: DrawingFormat() = %v, want one problem naming five lines and three more", tabs)
	}

	fruit := checks.DrawingFormat("🍎🍐🍊🍋🍌🍉🍇", checks.DefaultDrawingLimits())
	want := "on line 1 (U+1F34E, U+1F350, U+1F34A, U+1F34B, U+1F34C and 2 more)"
	if len(fruit) != 1 || !strings.Contains(fruit[0].Message, want) {
		t.Errorf("seven kinds of fruit: DrawingFormat() = %v, want one problem saying %q", fruit, want)
	}
}

// The limits are the caller's to give: the check holds a drawing to whatever
// it is handed, so that a calibration can try other numbers.
func TestTheLimitsAreTheOnesGiven(t *testing.T) {
	t.Parallel()

	drawing := "A" + strings.Repeat(" ", 10) + "B\n" + strings.Repeat("─", 25) + "\n│\n│"
	if problems := checks.DrawingFormat(drawing, checks.DefaultDrawingLimits()); len(problems) != 0 {
		t.Fatalf("DrawingFormat() at the default limits = %v, want no problems", problems)
	}

	tight := checks.DrawingLimits{Width: 20, Height: 3, SpaceRun: 5}
	problems := checks.DrawingFormat(drawing, tight)
	for _, want := range []string{"has 4 lines", "wider than 20 cells on line 2", "more than 5 spaces in a row on line 1"} {
		if !mentions(problems, want) {
			t.Errorf("DrawingFormat() at %+v = %v, want a problem saying %q", tight, problems, want)
		}
	}
}

func TestAMissingDrawingHasNoFormat(t *testing.T) {
	t.Parallel()

	for _, drawing := range []string{"", " \n \n"} {
		if problems := checks.DrawingFormat(drawing, checks.DefaultDrawingLimits()); len(problems) != 0 {
			t.Errorf("DrawingFormat(%q) = %v, want no problems", drawing, problems)
		}
	}
}

// A drawing comes from the chat's model and is untrusted: whatever it holds,
// the check ends in problems of its own code, at most one for each rule.
func FuzzDrawingFormat(f *testing.F) {
	for _, drawing := range wellMadeDrawings {
		f.Add(drawing, 30, 12, 20)
	}
	f.Add("A\t\u202E\r\n🍎 ", 1, 1, 1)
	f.Add("\xff\xfe\n\n\n", 0, -1, 0)

	f.Fuzz(func(t *testing.T, drawing string, width, height, spaceRun int) {
		limits := checks.DrawingLimits{Width: width, Height: height, SpaceRun: spaceRun}
		problems := checks.DrawingFormat(drawing, limits)
		if len(problems) > 11 {
			t.Fatalf("DrawingFormat() = %d problems, want at most one for each of the 11 rules", len(problems))
		}
		for _, problem := range problems {
			if problem.Code != checks.CodeDrawingFormat || problem.Message == "" {
				t.Fatalf("problem = %+v, want a message under %q", problem, checks.CodeDrawingFormat)
			}
		}
	})
}
