package content_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// Every kind of picture has its example, in the order the format lists the
// kinds, each in the format.
func TestEveryKindOfPictureHasItsExample(t *testing.T) {
	t.Parallel()

	examples := loaded(t).PictureExamples()
	kinds := make([]picture.Kind, 0, len(examples))
	for _, example := range examples {
		kinds = append(kinds, example.Kind)
		if _, broken := picture.Parse(example.Picture, picture.Point); len(broken) != 0 {
			t.Errorf("the example of a %s breaks %v", example.Kind, broken)
		}
	}
	if !slices.Equal(kinds, picture.Kinds()) {
		t.Errorf("examples of %v, want one of each kind in the format's order: %v", kinds, picture.Kinds())
	}
}

// Every reference task's picture is held to what a task the model writes is:
// its format, its labels against its question, and the answer kept off it.
// The reference tasks show the model how a topic is drawn, and a reference
// picture the checks refused would teach it to be refused.
func TestEveryReferencePicturePassesTheChecksOfATaskHandedIn(t *testing.T) {
	t.Parallel()

	drawn := 0
	for _, example := range loaded(t).Examples() {
		if example.Picture == nil {
			continue
		}
		drawn++
		task := &checks.Task{
			Question: example.Question, Picture: example.Picture, Options: example.Options,
			CorrectAnswer: example.CorrectAnswer,
		}
		for _, problem := range append(checks.PictureFormat(task, "en"), checks.PictureMatch(task, "en")...) {
			t.Errorf("%s: %s", example.ID, problem.Message)
		}
	}
	if drawn == 0 {
		t.Fatal("no reference task has a picture, so nothing here was tested")
	}
	t.Logf("%d reference pictures held to the checks", drawn)
}
