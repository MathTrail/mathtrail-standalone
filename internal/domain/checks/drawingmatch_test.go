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
		{"places named by words, in any case", "Ann walks from the start to the end.", "Start ───▶ End", labelled("Start", "End")},
		{"ends named only by the segment they close", "The segment AB is 4 cm long.", pointsAB, labelled("A", "B")},
		{"cells named by their row and column", "The robot goes from cell A1 to cell B2.",
			"  1 2\nA ┼ ┼\nB ┼ ┼", labelled("A", "B", "1", "2")},
		{"rows and columns named by the first and the last", "Rows are A to D, columns 1 to 3.",
			"  1 2 3\nA\nB\nC\nD", labelled("A", "B", "C", "D", "1", "2", "3")},
		{"rows named so in another language", "Ряды от A до C.", "A\nB\nC", labelled("A", "B", "C")},
		{"a frame, which has no question", "", pointsAB, labelled("A", "B")},
		{"labels set among the characters of a script without spaces", "点A在3，点B在7。它们相距多远？", pointsAB,
			labelled("A", "B")},
		{"a label opening a question in such a script", "AとBの距離は何センチですか？", pointsAB, labelled("A", "B")},
		{"a label opening an English sentence", "A is at 3 and B is at 7. How far apart are they?", pointsAB, labelled("A", "B")},
		{"a label opening a sentence in Russian", "A находится в 3 см от B.", pointsAB, labelled("A", "B")},
		{"labels opening a sentence in Spanish", "A y B están a 4 cm.", pointsAB, labelled("A", "B")},
		{"labels before the word they name in Turkish", "A kutusunda 5 top, B kutusunda 3 top var.", pointsAB,
			labelled("A", "B")},
		{"labels with a possessive", "Box A's lid is red and box B's is blue.", pointsAB, labelled("A", "B")},
		{"a label that is the letter I", "Rows G to I are shaded.", "G\nH\nI", labelled("G", "H", "I")},
		{"columns spanned by a dash", "The columns are 1–4.", "1 2 3 4", labelled("1", "2", "3", "4")},
		{"columns spanned in Portuguese", "As colunas do 1 ao 4 estão vazias.", "1 2 3 4", labelled("1", "2", "3", "4")},
		{"cells spanned by their corners", "The grid has cells A1 to D4.", "  1 2 3 4\nA\nB\nC\nD",
			labelled("A", "B", "C", "D", "1", "2", "3", "4")},
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
		{"a declared label left out of the drawing", "Point A is at 3, point B is at 7 and point C is at 9.", pointsAB,
			labelled("A", "B", "C"), `task.drawing does not show "C"`},
		{"a label found only inside a word", "Point A is at 3.", "BAR───┤", labelled("A"), `task.drawing does not show "A"`},
		{"a number found only inside a longer one", "Box 1 is full and box 2 is empty.", "┌─┐ ┌─┐\n└─┘ └─┘\n 1   12", labelled("1", "2"),
			`task.drawing does not show "2"`},
		{"a point the wording names and the structure leaves out", "Point D is between A and B.", pointsAB,
			labelled("A", "B"), `task.question names "D"`},
		{"a triangle drawn without its third corner", "The triangle ABC has a right angle at B.", "A\n│\nB────",
			labelled("A", "B"), `task.question names "C"`},
		{"a label the wording does not name", "How far apart are the two points?", pointsAB,
			labelled("A", "B"), "task.question does not name 2 of the labels"},
		{"labels standing for names the wording spells otherwise",
			"40 sweets are shared between Anya and Boris in the ratio 2 to 3. How many sweets does Boris get?",
			"A │▒▒│▒▒│\nB │▒▒│▒▒│▒▒│", labelled("A", "B"), "task.question does not name 2 of the labels"},
		{"rows left out, with no range to span them", "Rows A and C are shaded.", "A\nB\nC\nD",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"points between two named points, with no range", "Point A is at 3 and point D is at 9.", "A\nB\nC\nD",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"a number between two named numbers, with no range", "The weights are 1 and 9.", "1 5 9",
			labelled("1", "5", "9"), "task.question does not name 1 of the labels"},
		{"points too far apart in the wording to be a range",
			"Point A stands well away from the tall tree near point D.", "A\nB\nC\nD",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"a range of numbers that are not labels", "It takes 1 to 5 minutes.", "2 3", labelled("2", "3"),
			"task.question does not name 2 of the labels"},
		{"ends in two sentences", "Row A is red. Row D is blue.", "A\nB\nC\nD",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"ends set one under the other", "Rows shaded:\nA\nD", "A\nB\nC\nD",
			labelled("A", "B", "C", "D"), "task.question does not name 2 of the labels"},
		{"a capital opening a word with an accent", "Aún quedan 5 litros en el cubo.", "A\n┼", labelled("A"),
			"task.question does not name 1 of the labels"},
		{"a range from a capital opening a sentence, with none between", "A shop has three shelves; shelf C holds 9 jars.",
			"A\nB\nC", labelled("A", "B", "C"), "task.question does not name 1 of the labels"},
		{"a capital that is the unit of a number", "The jug holds 3 L of water.", "L\n┼", labelled("L"),
			"task.question does not name 1 of the labels"},
		{"a word in capitals holding a label's letter", "The stone is NOT at P. It is at Q.", pointsPQ + "\nT",
			labelled("P", "Q", "T"), "task.question does not name 1 of the labels"},
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

// A range whose ends the wording names names every label between them, in
// each language of the cards, with the marks any language spans a range by,
// and with the ends joined to a hyphen, an apostrophe or a script's letters.
func TestARangeNamesTheLabelsBetweenItsEnds(t *testing.T) {
	t.Parallel()

	for _, question := range []string{
		"The rows run from A up to D.",
		"从A到D的行涂了颜色。",
		"A से D तक की पंक्तियाँ रंगी हैं।",
		"Las filas de la A a la D están pintadas.",
		"الصفوف من A إلى D ملونة.",
		"Les rangées de A jusqu'à D sont coloriées.",
		"A থেকে D পর্যন্ত সারি রং করা।",
		"As linhas da A à D estão pintadas.",
		"Ряды с A по D закрашены.",
		"A سے D تک قطاریں رنگی ہیں۔",
		"Baris A sampai D diwarnai.",
		"Die Reihen von A bis D sind bemalt.",
		"AからDまでの行に色がある。",
		"A'dan D'ye kadar olan satırlar boyalı.",
		"A부터 D까지의 행은 칠해져 있다.",
		"Các hàng từ A đến D được tô màu.",
		"Le righe dalla A alla D sono colorate.",
		"ردیف های A تا D رنگی است.",
		"Rzędy od A do D są pomalowane.",
		"Ряди від A до D зафарбовані.",
		"แถว A ถึง D ถูกระบายสี",
		"De rijen van A tot D zijn gekleurd.",
		"Rows A-D are shaded.",
		"Rows A\u2010D are shaded.",
		"Rows A…D are shaded.",
		"Rows A～D are shaded.",
	} {
		t.Run(question, func(t *testing.T) {
			t.Parallel()

			if problems := checks.DrawingMatch(question, "A\nB\nC\nD", labelled("A", "B", "C", "D")); len(problems) != 0 {
				t.Errorf("DrawingMatch() = %v, want B and C named by the range", problems)
			}
		})
	}
}

// Two ends the wording lists — with a word for "and", "or" or "with" in each
// language of the cards, or with a mark that lists or reckons — name those two
// and nothing between them.
func TestListedEndsNameNothingBetween(t *testing.T) {
	t.Parallel()

	for _, question := range []string{
		"Rows A and D are shaded.",
		"A和D两行涂了颜色。",
		"पंक्तियाँ A और D रंगी हैं।",
		"Las filas A y D están pintadas.",
		"الصفان A و D ملونان.",
		"Les rangées A et D sont coloriées.",
		"সারি A ও D রং করা।",
		"As linhas A e D estão pintadas.",
		"Ряды A и D закрашены.",
		"قطار A اور D رنگی ہیں۔",
		"Baris A dan D diwarnai.",
		"Die Reihen A und D sind bemalt.",
		"AとDの行に色がある。",
		"A ve D satırları boyalı.",
		"A와 D 행은 칠해져 있다.",
		"Hàng A và D được tô màu.",
		"Le righe A e D sono colorate.",
		"ردیف A یا D رنگی است.",
		"Rzędy A i D są pomalowane.",
		"Ряди A і D зафарбовані.",
		"แถว A และ D ถูกระบายสี",
		"De rijen A en D zijn gekleurd.",
		"Rows A, D are shaded.",
		"Rows A/D are shaded.",
		"A + D = 10.",
	} {
		t.Run(question, func(t *testing.T) {
			t.Parallel()

			problems := checks.DrawingMatch(question, "A\nB\nC\nD", labelled("A", "B", "C", "D"))
			if len(problems) != 1 || !strings.Contains(problems[0].Message, "does not name 2 of the labels") {
				t.Errorf("DrawingMatch() = %v, want B and C unnamed", problems)
			}
		})
	}
}

// A label the wording does not name is counted, never quoted: a label that is
// a number may be the text of an option, and the refusal reaches the model
// before the child has answered.
func TestAnUnnamedLabelIsNotQuoted(t *testing.T) {
	t.Parallel()

	problems := checks.DrawingMatch("How many boxes are empty?", "┌─┐ ┌─┐\n└─┘ └─┘\n 7   9", labelled("7", "9"))
	if len(problems) != 1 {
		t.Fatalf("DrawingMatch() = %v, want exactly one problem", problems)
	}
	if message := problems[0].Message; strings.Contains(message, `"7"`) || strings.Contains(message, `"9"`) ||
		!strings.Contains(message, "does not name 2 of the labels") {
		t.Errorf("message = %q, want the two unnamed labels counted and neither quoted", message)
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
		{"a word in capitals sharing a letter with a label", "The stone is NOT at P. It is at Q, and O is empty.", []string{"O"}},
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
// code, at most one for each of the three ways it compares them.
func FuzzDrawingMatch(f *testing.F) {
	f.Add("The triangle ABC has a right angle at B.", "A\n│\nB────", "A")
	f.Add(`Ann says: "A stone lies at P."`, pointsPQ, "P")
	f.Add("\xff'A-B' 3 L °C", "\x00", "")

	f.Fuzz(func(t *testing.T, question, drawing, label string) {
		problems := checks.DrawingMatch(question, drawing, labelled(label, "A"))
		if len(problems) > 3 {
			t.Fatalf("DrawingMatch() = %d problems, want at most three", len(problems))
		}
		for _, problem := range problems {
			if problem.Code != checks.CodeDrawingMismatch || problem.Message == "" {
				t.Fatalf("problem = %+v, want a message under %q", problem, checks.CodeDrawingMismatch)
			}
		}
	})
}
