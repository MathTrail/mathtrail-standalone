package checks_test

import (
	"encoding/json"
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
