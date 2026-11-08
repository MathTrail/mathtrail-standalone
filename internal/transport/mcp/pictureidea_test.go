package mcpserver_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// gapsChoice asks for a task on gaps at 3-4, as the model's own choice.
var gapsChoice = map[string]any{
	"language": "en", "topic": "counting.gaps", "grade_level": "3-4", "reason": "More practice with gaps.",
}

// wordsOfGapsGiven leaves a child whose round of the ideas of gaps at 3-4 has
// given every idea told in words, so that what is left of it is the ideas the
// shipped list marks as solved from the picture.
func wordsOfGapsGiven(t *testing.T, shipped *content.Content) (change func(*profile.Profile), pictured []int) {
	t.Helper()

	var words []int
	for number := 1; number <= shipped.IdeaCount("counting.gaps", rating.Grades34); number++ {
		if shipped.IdeaFromPicture("counting.gaps", rating.Grades34, number) {
			pictured = append(pictured, number)
		} else {
			words = append(words, number)
		}
	}
	if len(pictured) == 0 {
		t.Fatal("the list of gaps at 3-4 marks no idea as solved from the picture")
	}
	return func(p *profile.Profile) {
		topic := p.Topics["counting.gaps"]
		topic.Ideas = map[rating.GradeLevel]profile.IdeaRound{rating.Grades34: {Given: words, Round: 1}}
		p.Topics["counting.gaps"] = topic
	}, pictured
}

// The ideas of today's topics marked as solved from the picture are drawn in a
// chat that draws the card, once the trial series is over: the brief asks for
// such a task, and its package carries the guide's section on it and the
// reference task of its idea. A chat of words passes them over and opens the
// next round of the list on an idea told in words, with the guide alone.
func TestAnIdeaSolvedFromThePictureIsDrawnOnlyWhereTheCardIsDrawn(t *testing.T) {
	t.Parallel()

	shipped, _ := shipped()
	for _, tc := range []struct {
		name        string
		host        chatHost
		fromPicture bool
	}{
		{"in a chat with cards", drawingCards, true},
		{"in a chat of words", wordsAlone, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			given, pictured := wordsOfGapsGiven(t, shipped)
			kept := keptAs(t, "petya", given)
			session := serveLesson(t, kept, &clock{at: lessonDay}, nil).connectChat(t, "", tc.host)

			call(t, session, "next_task", gapsChoice)
			p, _ := loadKept(t, kept)
			if p.OpenRequest == nil {
				t.Fatal("no request is open, want one on gaps")
			}
			request := p.OpenRequest
			onPicture := slices.Contains(pictured, request.Idea.Number)
			if request.Brief.FromPicture() != tc.fromPicture || onPicture != tc.fromPicture ||
				(!tc.fromPicture && request.Idea.Round != 2) {
				t.Errorf("the request is on idea %+v, solved from the picture %v; want it %v, in round 2 when not",
					request.Idea, request.Brief.FromPicture(), tc.fromPicture)
			}
			wantPackageOf(t, fetchPackage(t, session, request.ID), shipped, request.Idea.Number, tc.fromPicture)
		})
	}
}

// wantPackageOf holds the package of a request on an idea of gaps at 3-4 to
// carrying the guide's section on a task solved from the picture and the
// reference task of its idea when the idea is one, and the section not when
// it is not.
func wantPackageOf(t *testing.T, pkg string, shipped *content.Content, idea int, fromPicture bool) {
	t.Helper()

	if strings.Contains(pkg, "## A task solved from the picture") != fromPicture {
		t.Errorf("the package carries the guide's section on a task solved from the picture: %v, want %v",
			!fromPicture, fromPicture)
	}
	if fromPicture && !strings.Contains(pkg, referenceQuestionOf(t, shipped, idea)) {
		t.Errorf("the package of idea %d does not show its reference task", idea)
	}
}

// referenceQuestionOf is the question of the reference task built on an idea
// of gaps at 3-4, as a package writes it out.
func referenceQuestionOf(t *testing.T, shipped *content.Content, idea int) string {
	t.Helper()

	tasks := shipped.Examples()
	for i := range tasks {
		if task := &tasks[i]; task.Topic == "counting.gaps" && task.GradeLevel == rating.Grades34 && task.Idea == idea {
			quoted, err := json.Marshal(task.Question)
			if err != nil {
				t.Fatalf("write the question out: %v", err)
			}
			return strings.Trim(string(quoted), `"`)
		}
	}
	t.Fatalf("no reference task of gaps at 3-4 is built on idea %d", idea)
	return ""
}
