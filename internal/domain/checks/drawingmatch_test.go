package checks_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// labelled is a drawing structure whose objects carry these labels.
func labelled(labels ...string) *checks.DrawingStructure {
	structure := &checks.DrawingStructure{Kind: "number_line"}
	for _, label := range labels {
		structure.Objects = append(structure.Objects, checks.DrawingObject{ID: label, Label: label})
	}
	return structure
}

// pointsAB and pointsPQ are number lines showing two labelled points.
const (
	pointsAB = "0   A       B      10\n├───┼───────┼──────┤"
	pointsPQ = "0   P       Q      10\n├───┼───────┼──────┤"
)

func TestADrawingThatShowsWhatItDeclaresMatches(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question, drawing string
		structure               *checks.DrawingStructure
	}{
		{"points on a number line", "Point A is at 3 and point B is at 7. How far apart are they?", pointsAB, labelled("A", "B")},
		{"a segment named by its ends", "How long is AB, if A is at 3 and B is at 7?", pointsAB, labelled("A", "B")},
		{"boxes numbered in the drawing", "Box 1 holds 4 balls and box 2 holds 6.", "┌─┐ ┌─┐\n└─┘ └─┘\n 1   2", labelled("1", "2")},
		{"places named by words", "Ann walks from the start to the end.", "Start ───▶ End", labelled("Start", "End")},
		{"a label the wording does not name", "How far apart are the two points?", pointsAB, labelled("A", "B")},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if problems := checks.DrawingMatch(test.question, test.drawing, test.structure); len(problems) != 0 {
				t.Errorf("DrawingMatch() = %v, want no problems", problems)
			}
		})
	}
}

func TestEachMismatchIsRefused(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question, drawing string
		structure               *checks.DrawingStructure
		want                    string
	}{
		{"a declared label left out of the drawing", "Point A is at 3 and point B is at 7.", pointsAB,
			labelled("A", "B", "C"), `task.drawing does not show "C"`},
		{"a label found only inside a word", "Point A is at 3.", "BAR───┤", labelled("A"), `task.drawing does not show "A"`},
		{"a number found only inside a longer one", "Box 2 is empty.", "┌─┐ ┌─┐\n└─┘ └─┘\n 1   12", labelled("1", "2"),
			`task.drawing does not show "2"`},
		{"a point the wording names and the structure leaves out", "Point D is between A and B.", pointsAB,
			labelled("A", "B"), `task.question names "D"`},
		{"a triangle drawn without its third corner", "The triangle ABC has a right angle at B.", "A\n│\nB────",
			labelled("A", "B"), `task.question names "C"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.DrawingMatch(test.question, test.drawing, test.structure)
			if len(problems) != 1 {
				t.Fatalf("DrawingMatch() = %v, want exactly one problem", problems)
			}
			if problems[0].Code != checks.CodeDrawingMismatch {
				t.Errorf("code = %q, want %q", problems[0].Code, checks.CodeDrawingMismatch)
			}
			if !strings.Contains(problems[0].Message, test.want) {
				t.Errorf("message = %q, want it to say %q", problems[0].Message, test.want)
			}
		})
	}
}

// A capital that could as well be a word is taken for one, and numbers and
// quotations are not read at all: a refusal for a label that was never one
// would cost the child an attempt. The structure here declares P and Q only,
// so any other letter read as a label would be refused.
func TestWordsAreNotTakenForLabels(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question string
		labels         []string
	}{
		{"the English article starting a sentence", "A farmer puts a stone at P and another at Q.", nil},
		{"the article starting a quotation", `Ann says: "A stone lies at P." Ben says: "It lies at Q."`, nil},
		{"the article after a colon", "Look at the line: A stone lies at P, a shell at Q.", nil},
		{"the English I", "Then I put a stone at P and a shell at Q.", nil},
		{"a unit after a number", "A 3 L jug stands at P and a 5 L jug at Q.", nil},
		{"a unit after a degree sign", "It is 30 °C at P and 25 °C at Q.", nil},
		{"a word in capitals", "The stone is NOT at P. It is at Q.", nil},
		{"a word in capitals sharing a letter with a label", "The stone is NOT at P. It is at Q.", []string{"O"}},
		{"a capital joined by a hyphen", "Ann wears a T-shirt at P and makes a U-turn at Q.", nil},
		{"a capital joined by an apostrophe", "Mr O'Neil stands at P, and Q is empty.", nil},
		{"numbers", "P is at 3 and Q is at 7, and 12 is the end.", nil},
		{"a quotation", `Ann says: "Ben is at P." Ben is at Q.`, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			labels := append([]string{"P", "Q"}, test.labels...)
			drawing := pointsPQ + "\n" + strings.Join(test.labels, " ")
			if problems := checks.DrawingMatch(test.question, drawing, labelled(labels...)); len(problems) != 0 {
				t.Errorf("DrawingMatch() = %v, want no problems", problems)
			}
		})
	}
}

// A label the wording names over and over is one thing to fix, and it is
// named once.
func TestALabelIsNamedOnce(t *testing.T) {
	t.Parallel()

	problems := checks.DrawingMatch("Point D is right of A, and D is left of B.", pointsAB, labelled("A", "B"))
	if len(problems) != 1 || !strings.Contains(problems[0].Message, `task.question names "D", which`) {
		t.Errorf("DrawingMatch() = %v, want one problem naming D once", problems)
	}
}

func TestNothingIsMatchedWithoutADrawingAndItsStructure(t *testing.T) {
	t.Parallel()

	if problems := checks.DrawingMatch("Point D is left of A.", pointsAB, nil); len(problems) != 0 {
		t.Errorf("DrawingMatch() without a structure = %v, want no problems", problems)
	}
	if problems := checks.DrawingMatch("Point D is left of A.", " \n", labelled("A")); len(problems) != 0 {
		t.Errorf("DrawingMatch() without a drawing = %v, want no problems", problems)
	}
}

// The wording, the drawing and the labels all come from the chat's model and
// are untrusted: whatever they hold, the check ends in problems of its own
// code, at most one for each direction it compares.
func FuzzDrawingMatch(f *testing.F) {
	f.Add("The triangle ABC has a right angle at B.", "A\n│\nB────", "A")
	f.Add(`Ann says: "A stone lies at P."`, pointsPQ, "P")
	f.Add("\xff'A-B' 3 L °C", "\x00", "")

	f.Fuzz(func(t *testing.T, question, drawing, label string) {
		problems := checks.DrawingMatch(question, drawing, labelled(label, "A"))
		if len(problems) > 2 {
			t.Fatalf("DrawingMatch() = %d problems, want at most two", len(problems))
		}
		for _, problem := range problems {
			if problem.Code != checks.CodeDrawingMismatch || problem.Message == "" {
				t.Fatalf("problem = %+v, want a message under %q", problem, checks.CodeDrawingMismatch)
			}
		}
	})
}
