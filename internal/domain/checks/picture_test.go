package checks_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// A picture out of its format is refused at the member that breaks a rule,
// named from the task down, under the code of the picture's format.
func TestAPictureOutOfItsFormatIsRefusedAtItsMember(t *testing.T) {
	t.Parallel()

	task := &checks.Task{Picture: json.RawMessage(`{"kind":"containers","items":[{"capacity":3,"amount":4}]}`)}
	problems := checks.PictureFormat(task, "en")
	want := "task.picture.items.0.amount must be a whole number from 0 up to the container's capacity"
	if len(problems) != 1 || problems[0].Code != checks.CodeDrawingFormat || problems[0].Message != want {
		t.Errorf("PictureFormat() = %v, want one %q: %s", problems, checks.CodeDrawingFormat, want)
	}
}

// A picture writes its numbers as its wording does, with the decimal mark of
// the lesson's language.
func TestAPicturesNumbersAreWrittenAsTheLessonsLanguageWritesThem(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		value, lesson string
		accepted      bool
	}{
		{"2,5", "de", true},
		{"2.5", "de", false},
		{"2.5", "en", true},
		{"2,5", "en", false},
		{"2,5", "pt-BR", true},
		{"2.5", "es-MX", true},
	} {
		task := &checks.Task{Picture: json.RawMessage(`{"kind":"bars","bars":[{"value":"` + test.value + `"}]}`)}
		if got := len(checks.PictureFormat(task, test.lesson)) == 0; got != test.accepted {
			t.Errorf("PictureFormat(%s in %s) accepted = %v, want %v", test.value, test.lesson, got, test.accepted)
		}
	}
}

// A task with no picture has no format to check, and neither has one whose
// picture is no object, which the structure check refuses.
func TestATaskWithNoPictureHasNoFormatToCheck(t *testing.T) {
	t.Parallel()

	for _, picture := range []json.RawMessage{nil, json.RawMessage(`"a clock"`), json.RawMessage(`[1]`)} {
		if problems := checks.PictureFormat(&checks.Task{Picture: picture}, "en"); len(problems) != 0 {
			t.Errorf("PictureFormat(%s) = %v, want none", picture, problems)
		}
	}
	if problems := checks.PictureFormat(nil, "en"); len(problems) != 0 {
		t.Errorf("PictureFormat(nil) = %v, want none", problems)
	}
}

// A total of a task whose answer is a label ends in that label: an option
// that holds no number is held to the total's end as the text it is.
func TestATotalEndsInTheLabelItsAnswerIs(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		total string
		codes []checks.Code
	}{
		{"AB + BC = AC", nil},
		{"AB + BC = AD", []checks.Code{checks.CodeDrawingMismatch}},
	} {
		held := &checks.Task{
			Question: "Points A, B, C and D lie on a line in that order. AB is 2 cm, BC is 3 cm, CD is 4 cm " +
				"and AD is 9 cm. Which segment is 5 cm long?",
			Solution:        "AB and BC together are 2 + 3 = 5 cm, and they make AC.",
			SolutionPicture: json.RawMessage(`{"kind":"row","items":[{"label":"A"},{"label":"B"},{"label":"C"},{"label":"D"}]}`),
			SolutionTotal:   test.total, CorrectAnswer: "C",
			Options: map[string]string{"A": "AB", "B": "BC", "C": "AC", "D": "AD", "E": "CD"},
		}
		var codes []checks.Code
		for _, problem := range checks.SolutionPicture(held, "en") {
			codes = append(codes, problem.Code)
		}
		if !slices.Equal(codes, test.codes) {
			t.Errorf("SolutionPicture() with the total %q: codes = %v, want %v", test.total, codes, test.codes)
		}
	}
}

// The picture of a solution is checked alone as the review checks it: in its
// format, its total in its own, and against the wording and the solution;
// a task with none has nothing to check.
func TestThePictureOfASolutionIsCheckedAloneAsTheReviewChecksIt(t *testing.T) {
	t.Parallel()

	task := func(picture, total string) *checks.Task {
		return &checks.Task{
			Question: "How many posts stand along a fence of 12 metres, one every 3 metres?",
			Solution: "12 ÷ 3 = 4 gaps, and 4 + 1 = 5 posts.", SolutionPicture: json.RawMessage(picture),
			SolutionTotal: total, CorrectAnswer: "C",
			Options: map[string]string{"A": "3", "B": "4", "C": "5", "D": "6", "E": "12"},
		}
	}
	posts := `{"kind":"row","items":[{},{},{},{},{}],"gaps":"3","span":"12"}`
	for _, test := range []struct {
		name, picture, total string
		codes                []checks.Code
	}{
		{"a picture and a total that ends in the answer", posts, "12 ÷ 3 + 1 = 5", nil},
		{"no picture at all", "", "", nil},
		{"a picture out of its format", `{"kind":"row","items":[{}]}`, "", []checks.Code{checks.CodeDrawingFormat}},
		{"a total out of its format", posts, "12 ÷ 3 + 1 = ?", []checks.Code{checks.CodeDrawingFormat}},
		{"a total that ends in a wrong option", posts, "12 ÷ 3 = 4", []checks.Code{checks.CodeDrawingMismatch}},
		{"a label neither the question nor the solution names", `{"kind":"row","items":[{"label":"Q"},{}]}`, "",
			[]checks.Code{checks.CodeDrawingMismatch}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			held := task(test.picture, test.total)
			if test.picture == "" {
				held.SolutionPicture = nil
			}
			var codes []checks.Code
			for _, problem := range checks.SolutionPicture(held, "en") {
				codes = append(codes, problem.Code)
			}
			if !slices.Equal(codes, test.codes) {
				t.Errorf("SolutionPicture() codes = %v, want %v", codes, test.codes)
			}
		})
	}
}
