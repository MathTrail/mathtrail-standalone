package content_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// tasksOfIdeas are the reference tasks that name an idea solved from the
// picture.
func tasksOfIdeas(shipped *content.Content) []content.Example {
	return slices.DeleteFunc(shipped.Examples(), func(task content.Example) bool { return task.Idea == 0 })
}

// questionsShown are the questions of the reference tasks a package shows.
func questionsShown(t *testing.T, examples []map[string]json.RawMessage) []string {
	t.Helper()

	questions := make([]string, 0, len(examples))
	for _, example := range examples {
		var question string
		if err := json.Unmarshal(example["question"], &question); err != nil {
			t.Fatalf("read the question of an example: %v", err)
		}
		questions = append(questions, question)
	}
	return questions
}

// ideaPlace is an idea of a topic's list at a level, by its place from 1.
type ideaPlace struct {
	topic  string
	level  rating.GradeLevel
	number int
}

// ideasSolvedFromThePicture are the ideas the lists of topics told in words
// mark as solved from the picture.
func ideasSolvedFromThePicture(shipped *content.Content) []ideaPlace {
	var marked []ideaPlace
	for _, topic := range shipped.Topics() {
		for _, level := range topic.GradeLevels {
			for number := 1; !topic.SolvedFromPicture && number <= shipped.IdeaCount(topic.ID, level); number++ {
				if shipped.IdeaFromPicture(topic.ID, level, number) {
					marked = append(marked, ideaPlace{topic.ID, level, number})
				}
			}
		}
	}
	return marked
}

// Every idea a list of a topic told in words marks as solved from the picture
// has a reference task built on it: the package of such a task shows the genre
// its idea asks for, which none of the topic's other reference tasks shows.
func TestEveryIdeaSolvedFromThePictureHasItsReferenceTask(t *testing.T) {
	t.Parallel()

	shipped := loaded(t)
	named := map[ideaPlace]bool{}
	for _, task := range tasksOfIdeas(shipped) {
		named[ideaPlace{task.Topic, task.GradeLevel, task.Idea}] = true
	}
	marked := ideasSolvedFromThePicture(shipped)
	if len(marked) == 0 {
		t.Fatal("no list of a topic told in words marks an idea as solved from the picture, so nothing here was tested")
	}
	for _, idea := range marked {
		if !named[idea] {
			t.Errorf("idea %d of %s at %s is solved from the picture, and no reference task is built on it",
				idea.number, idea.topic, idea.level)
		}
	}
}

// The package of a task solved from the picture shows the reference task of
// its idea in its last place, whatever the difficulty and the turn, and the
// package of a task in words, or of another idea, never shows it.
func TestAReferenceTaskOfAnIdeaIsShownInThePackageOfThatIdeaAlone(t *testing.T) {
	t.Parallel()

	shipped := loaded(t)
	for _, task := range tasksOfIdeas(shipped) {
		t.Run(task.ID, func(t *testing.T) {
			t.Parallel()

			for difficulty := profile.MinDifficulty; difficulty <= profile.MaxDifficulty; difficulty++ {
				wantShownInItsPackageAlone(t, shipped, &task, difficulty)
			}
		})
	}
}

// wantShownInItsPackageAlone holds the packages at one difficulty to showing
// a reference task of an idea last in the package of its idea, and in no
// package of a task in words or of another idea.
func wantShownInItsPackageAlone(t *testing.T, shipped *content.Content, task *content.Example, difficulty int) {
	t.Helper()

	asked := request(task.Topic, task.GradeLevel, difficulty, difficulty)
	asked.Idea = profile.Idea{Number: task.Idea, Round: 1}
	if _, inWords := packageFor(t, shipped, asked); slices.Contains(questionsShown(t, inWords.Examples), task.Question) {
		t.Errorf("difficulty %d: the package of a task in words shows %s", difficulty, task.ID)
	}

	asked.Brief.Constraints = []string{profile.FactsOnThePicture}
	_, pictured := packageFor(t, shipped, asked)
	if shown := questionsShown(t, pictured.Examples); len(shown) == 0 || shown[len(shown)-1] != task.Question {
		t.Errorf("difficulty %d: the package of its idea shows %d tasks, want %s last", difficulty, len(shown), task.ID)
	}

	asked.Idea.Number = 1
	if _, other := packageFor(t, shipped, asked); slices.Contains(questionsShown(t, other.Examples), task.Question) {
		t.Errorf("difficulty %d: the package of idea 1 shows %s", difficulty, task.ID)
	}
}
