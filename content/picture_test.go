package content_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
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

// Every picture of a solution a reference task draws is held to what a task
// handed in is held to: its format and its total's, and its labels, the words
// of its colours and its total against the question and the solution. A
// reference task shows the model how its topic draws a solution, and one the
// checks refused would teach it to be refused.
func TestEveryReferencePictureOfASolutionPassesTheChecksOfATaskHandedIn(t *testing.T) {
	t.Parallel()

	drawn := 0
	for _, example := range loaded(t).Examples() {
		if example.SolutionPicture == nil {
			continue
		}
		drawn++
		task := &checks.Task{
			Question: example.Question, Solution: example.Solution, Options: example.Options,
			CorrectAnswer: example.CorrectAnswer, SolutionPicture: example.SolutionPicture,
			SolutionTotal: example.SolutionTotal,
		}
		for _, problem := range checks.SolutionPicture(task, "en") {
			t.Errorf("%s: %s", example.ID, problem.Message)
		}
	}
	if drawn == 0 {
		t.Fatal("no reference task draws its solution, so nothing here was tested")
	}
	t.Logf("%d reference pictures of a solution held to the checks", drawn)
}

// Every topic has, at every level it is taught at, a reference task that draws
// its solution: every task the model writes draws its own, and the package of
// each topic and level shows the model how.
func TestEveryTopicDrawsASolutionAtEveryLevel(t *testing.T) {
	t.Parallel()

	shipped := loaded(t)
	type place struct {
		topic string
		level rating.GradeLevel
	}
	drawsAt := map[place]bool{}
	for _, task := range shipped.Examples() {
		if task.SolutionPicture != nil {
			drawsAt[place{task.Topic, task.GradeLevel}] = true
		}
	}
	for _, topic := range shipped.Topics() {
		for _, level := range topic.GradeLevels {
			if !drawsAt[place{topic.ID, level}] {
				t.Errorf("%s at %s: no reference task draws its solution", topic.ID, level)
			}
		}
	}
}
