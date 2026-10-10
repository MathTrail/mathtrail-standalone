package checks_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// showsTheAnswer says whether holding a picture to its wording found that it
// shows the right answer.
func showsTheAnswer(t *testing.T, question, picture, right, lesson string) bool {
	t.Helper()

	task := &checks.Task{
		Question: question, Picture: json.RawMessage(picture), CorrectAnswer: "A",
		Options: map[string]string{"A": right, "B": "w1", "C": "w2", "D": "w3", "E": "w4"},
	}
	if broken := checks.PictureFormat(task, lesson); len(broken) != 0 {
		t.Fatalf("PictureFormat() = %v, want a picture in the format", broken)
	}
	return slices.ContainsFunc(checks.PictureMatch(task, lesson), func(problem checks.Problem) bool {
		return strings.Contains(problem.Message, "shows the right answer")
	})
}

// Each of these pictures shows its task's right answer, which its wording
// does not give, in one of the ways a picture can.
func TestAPictureThatShowsTheAnswerIsRefused(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question, picture, right, lesson string
	}{
		{"a clock at the time asked for", "A film starts at 3:00 and lasts 90 minutes. When does it end?",
			`{"kind":"clock","time":"4:30"}`, "4:30", "en"},
		{"a clock at the time asked for, on the afternoon's twelve hours",
			"Lunch is at 13:00 and lasts three and a half hours. When does it end?",
			`{"kind":"clock","time":"4:30"}`, "16:30", "en"},
		{"a clock at a time asked for with a nought before its hour", "The bus leaves half an hour after 7:45. When?",
			`{"kind":"clock","time":"8:15"}`, "08:15", "en"},
		{"a clock at a time asked for in the afternoon", "Tea starts at 3 pm and lasts 90 minutes. When does it end?",
			`{"kind":"clock","time":"16:30"}`, "4:30 pm", "en"},
		{"a clock on the hour asked for", "A film starts at 3 and lasts two hours. When does it end?",
			`{"kind":"clock","time":"5:00"}`, "5 o'clock", "en"},
		{"a row that ends on the count asked for", "Posts stand every 3 metres along a 12-metre fence. How many posts?",
			`{"kind":"row","items":[{"label":"1"},{"label":"2"},{"skip":true},{"label":"4"},{"label":"5"}]}`, "5", "en"},
		{"a grid that fills the cell asked for", "Rows A and B, columns 1 and 2: which cell holds the treasure?",
			`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"filled":["B2"]}`, "B2", "en"},
		{"a grid that marks the cell asked for", "Rows A and B, columns 1 and 2: in which cell is the treasure T?",
			`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"marks":{"B2":"T"}}`, "B2", "en"},
		{"a cell asked for, which the wording holds only inside a longer name",
			"Rows A and B, columns 1 and 2. Cell AB2 is off the map. In which cell is the treasure T?",
			`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"marks":{"B2":"T"}}`, "B2", "en"},
		{"a mark on the tick asked for", "Ann is 3 steps left of 10. Where is Ann?",
			`{"kind":"number_line","from":0,"to":10,"marks":[{"at":7,"label":"?"}]}`, "7", "en"},
		{"a container holding the amount asked for", "Pour to get exactly two litres. How much is in the big jug?",
			`{"kind":"containers","items":[{"capacity":5,"amount":2},{"capacity":3,"amount":0}]}`, "2", "en"},
		{"a bar's value, the answer with its unit", "Each part is 12 cm. How long is the whole?",
			`{"kind":"bars","bars":[{"parts":3,"value":"36"}]}`, "36 cm", "en"},
		{"a note that works the answer out", "A is 5 and B is 7. What is A + B?",
			`{"kind":"bars","bars":[{"label":"A"},{"label":"B"}],"notes":["A + B = 12"]}`, "12", "en"},
		{"a ring of the count asked for", "Children stand in a circle. How many children are there?",
			`{"kind":"ring","count":9}`, "9", "en"},
		{"a month's page marking the day asked for", "The fair is on the second Saturday. What date is it?",
			`{"kind":"calendar","first":3,"days":30,"marks":{"14":"?"}}`, "14", "en"},
		{"a table holding the time asked for", "Train A leaves at the time the table gives. When does it leave?",
			`{"kind":"table","rows":[["A","9:15"]]}`, "9:15", "en"},
		{"a table holding the time asked for, written with a nought before its hour",
			"Train A leaves at the time the table gives. When does it leave?",
			`{"kind":"table","rows":[["A","9:05"]]}`, "09:05", "en"},
		{"a number with a decimal comma", "Половина — 1,25. Сколько всего?",
			`{"kind":"bars","bars":[{"parts":2,"value":"2,5"}]}`, "2,5", "ru"},
		{"a note that works out a number below zero", "A number A is 5 less than 2. What is A?",
			`{"kind":"bars","bars":[{"label":"A"}],"notes":["A = −3"]}`, "−3", "en"},
		{"a number with a decimal comma, the answer with its unit", "Половина мешка весит 1,25 кг. Сколько весит мешок?",
			`{"kind":"bars","bars":[{"parts":2,"value":"2,5"}]}`, "2,5 кг", "ru"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if !showsTheAnswer(t, test.question, test.picture, test.right, test.lesson) {
				t.Errorf("PictureMatch() lets a picture through that shows the answer %q", test.right)
			}
		})
	}
}

// These pictures show no answer the wording does not give: the wording holds
// it itself, or the picture holds it only among names that number places
// alike, or among numbers that only lay it out.
func TestAPictureThatShowsNoAnswerIsNotRefused(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question, picture, right, lesson string
	}{
		{"a number the wording gives", "Ann stands at 7 and Ben at 3. Where does Ann stand?",
			`{"kind":"number_line","from":0,"to":10,"marks":[{"at":7,"label":"A"},{"at":3,"label":"B"}]}`, "7", "en"},
		{"a time the wording gives", "The film starts at 4:30. When does it start?",
			`{"kind":"clock","time":"4:30"}`, "4:30", "en"},
		{"a time the wording gives on another twelve hours", "Tea is at 16:30. When is tea?",
			`{"kind":"clock","time":"16:30"}`, "4:30 pm", "en"},
		{"a slot among slots numbered alike", "Seven bikes stand in slots numbered 1 to 7. In which slot is Ivo's bike?",
			`{"kind":"row","items":[{"label":"1"},{"label":"2"},{"label":"3"},{"label":"4"},{"label":"5"},` +
				`{"label":"6"},{"label":"7"}]}`, "5", "en"},
		{"islanders numbered alike", "Six islanders stand in a row. How many knights are there?",
			`{"kind":"row","items":[{"label":"1"},{"label":"2"},{"label":"3"},{"label":"4"},{"label":"5"},` +
				`{"label":"6"}]}`, "3", "en"},
		{"a column of a grid", "Rows A to C and columns 1 to 3. By how much does the perimeter grow?",
			`{"kind":"grid","rows":["A","B","C"],"cols":["1","2","3"],"filled":["A1","A2"]}`, "2", "en"},
		{"a row drawn twice", "Posts stand on each of the two sides of a path. How many sides?",
			`{"kind":"row","items":[{},{}],"copies":2}`, "2", "en"},
		{"counters in groups", "There are 20 stones. How many should the first player take?",
			`{"kind":"piles","piles":[{"count":20,"group":5}]}`, "5", "en"},
		{"the step of a line", "Point A is at 4. Each jump is two long. How long is a jump?",
			`{"kind":"number_line","from":0,"to":10,"step":2,"marks":[{"at":4,"label":"A"}]}`, "2", "en"},
		{"a point's letter inside a range the wording names", "Points A to E stand on a line. Which is in the middle?",
			`{"kind":"row","items":[{"label":"A"},{"label":"B"},{"label":"C"},{"label":"D"},{"label":"E"}]}`, "C", "en"},
		{"a number with its unit the wording gives", "A 2 kg weight balances the bag. What does the bag weigh?",
			`{"kind":"balance","left":["2"],"right":["?"]}`, "2 kg", "en"},
		{"a time the wording gives straight after another", "The lesson runs 9:00-10:30. When does it end?",
			`{"kind":"clock","time":"10:30"}`, "10:30", "en"},
		{"a time in a table the wording gives, written another way", "Train A leaves at 9:05. When does it leave?",
			`{"kind":"table","rows":[["A","9:05"]]}`, "09:05", "en"},
		{"a number the wording gives at the end of a range",
			"The fair runs on days 3-5 of the month. On which day does it end?",
			`{"kind":"calendar","first":1,"days":30,"marks":{"5":"?"}}`, "5", "en"},
		{"a number the wording joins to words written without spaces", "现在是3:00。时针指向几？",
			`{"kind":"clock","time":"3:00"}`, "3", "zh"},
		{"a cell the wording names among words written without spaces", "行A和B，列1和2。宝藏T在B2里。宝藏在哪个格子？",
			`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"marks":{"B2":"T"}}`, "B2", "zh"},
		{"a note's number below zero, the answer its size", "A is three less than nought. How far is A from nought?",
			`{"kind":"bars","bars":[{"label":"A"}],"notes":["A = −3"]}`, "3", "en"},
		{"a number the wording gives with its decimal comma", "Гиря в 2,5 кг уравновешивает мешок. Сколько весит мешок?",
			`{"kind":"balance","left":["2,5"],"right":["?"]}`, "2,5", "ru"},
		{"a cell whose name holds the answer's digit", "Rows A and B, columns 1 and 2: which column is B2 in?",
			`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"marks":{"B2":"T"}}`, "2", "en"},
		{"a cell the wording names, after a longer name that holds it",
			"Rows A and B, columns 1 and 2. Cell AB2 is off the map, and cell B2 holds the treasure T. Which cell holds it?",
			`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"marks":{"B2":"T"}}`, "B2", "en"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if showsTheAnswer(t, test.question, test.picture, test.right, test.lesson) {
				t.Errorf("PictureMatch() refuses a picture that shows no answer %q the wording does not give", test.right)
			}
		})
	}
}

// The refusal says the picture shows the answer, and quotes nothing: the
// right option's text is what the child must not read before answering.
func TestTheAnswerShownIsNotQuoted(t *testing.T) {
	t.Parallel()

	task := &checks.Task{
		Question: "Children stand in a circle. How many children are there?", CorrectAnswer: "A",
		Picture: json.RawMessage(`{"kind":"ring","count":9}`),
		Options: map[string]string{"A": "9", "B": "8", "C": "10", "D": "7", "E": "11"},
	}
	for _, problem := range checks.PictureMatch(task, "en") {
		if strings.ContainsAny(problem.Message, "0123456789") {
			t.Errorf("message = %q, want no number of the task in it", problem.Message)
		}
	}
}

// A lesson written in Greek names its points with Greek capitals, and a label
// the wording names so is given by the wording, whichever of the two alphabets
// the picture and the right option write it in: a picture of point B, in a
// task whose wording names it, shows no answer. In a lesson in any other
// language a Greek capital names no Latin label.
func TestAGreekWordingGivesTheLatinLabelItLooksLike(t *testing.T) {
	t.Parallel()

	question := "Το σημείο \u0392 απέχει 5 cm από το \u0391. Ποιο σημείο απέχει 5 cm από το \u0391;"
	picture := `{"kind":"table","rows":[["A"],["B"]]}`
	if showsTheAnswer(t, question, picture, "B", "el") {
		t.Errorf("PictureMatch() refuses a picture of B in a Greek lesson whose wording names it, want it let through")
	}
	if !showsTheAnswer(t, question, picture, "B", "ru") {
		t.Errorf("PictureMatch() lets a picture of B through in a lesson in Russian whose wording names a Greek capital, want it refused")
	}
}

// A Greek lesson's right option written in a Greek capital drawn as a Latin
// one is the Latin label a picture shows: a picture of B, in a task whose
// right option is Β and whose wording does not give it, shows the answer.
func TestAGreekOptionIsTheLatinLabelItLooksLike(t *testing.T) {
	t.Parallel()

	if !showsTheAnswer(t, "Ποιο σημείο απέχει 5 cm από το \u0391;", `{"kind":"table","rows":[["A"],["B"]]}`,
		"\u0392", "el") {
		t.Errorf("PictureMatch() lets a picture of B through in a Greek lesson whose right option is Greek Beta, want it refused")
	}
}

// A Greek lesson's wording that names a cell Β2 in Greek gives the cell B2 a
// picture marks: every reading of the wording takes the Greek cell for the
// Latin one, and the right option too.
func TestAGreekCellGivesTheLatinCellItLooksLike(t *testing.T) {
	t.Parallel()

	picture := `{"kind":"grid","rows":["A","B"],"cols":["1","2"],"filled":["B2"]}`
	if showsTheAnswer(t, "Το κελί \u03922 είναι γεμάτο. Ποιο κελί είναι γεμάτο;", picture, "\u03922", "el") {
		t.Errorf("PictureMatch() refuses a grid with B2 filled in a Greek lesson whose wording names that cell, want it let through")
	}
}
