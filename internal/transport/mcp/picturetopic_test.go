package mcpserver_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The catalog has topics all of whose tasks are solved from the picture.
// Named by the model or kept to by a person, such a topic is the child's only
// in a chat known to draw the card; in any other the rule chooses.

// figures is a topic of the shipped catalog all of whose tasks are solved
// from the picture.
const figures = "counting.figures"

// figuresChoice asks for a task on figures, as the model's own choice.
var figuresChoice = map[string]any{
	"language": "en", "topic": figures, "reason": "Counting the triangles of a figure.",
}

// A topic solved from the picture named by the model is refused under
// needs_card in a chat that does not draw the card, by next_task and by
// prepare_task alike, and nothing is written. Where the card is drawn the same
// choice opens a request for a task solved from the picture, in the trial
// series too, since the series holds only the rule's own choice to words.
func TestATopicSolvedFromThePictureNamedInAChatOfWordsIsRefused(t *testing.T) {
	t.Parallel()

	for _, tool := range []string{"next_task", "prepare_task"} {
		t.Run(tool, func(t *testing.T) {
			t.Parallel()

			kept := racer(t)
			h := serveLesson(t, kept, &clock{at: lessonDay}, nil)
			cards, words := h.connectChat(t, "", drawingCards), h.connectChat(t, "", wordsAlone)
			if tool == "prepare_task" {
				raceHandedOut(t, cards, kept)
			}
			_, revision := loadKept(t, kept)

			refused := payloadOf[requestPayload](t, call(t, words, tool, figuresChoice))
			if refused.Status != "rejected" || refused.Code != "invalid_arguments" || len(refused.Problems) != 1 ||
				refused.Problems[0].Field != "topic" || refused.Problems[0].Code != profile.CodeNeedsCard {
				t.Errorf("%s in a chat of words = %+v, want the topic refused under needs_card", tool, refused)
			}
			if _, now := loadKept(t, kept); now != revision {
				t.Fatalf("%s in a chat of words wrote the profile, want nothing written", tool)
			}

			call(t, cards, tool, figuresChoice)
			if p, _ := loadKept(t, kept); p.OpenRequest == nil || p.OpenRequest.Brief.TargetConcept != figures ||
				!p.OpenRequest.Brief.FromPicture() {
				t.Errorf("%s in a chat with cards opened %+v, want a request on %s solved from the picture",
					tool, p.OpenRequest, figures)
			}
		})
	}
}

// Lessons kept to a topic solved from the picture keep to it where the card
// is drawn, and the model is told so. In a chat that does not draw the card the
// choice counts as none: the rule chooses a topic told in words, and the model
// is told why the tasks leave the topic chosen.
func TestLessonsKeptToATopicSolvedFromThePictureKeepToItOnlyWhereTheCardIsDrawn(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		host  chatHost
		said  string
		keeps bool
	}{
		{"in a chat with cards", drawingCards, "keeps the lessons to Counting figures: every task is on it", true},
		{"in a chat of words", wordsAlone, "this chat is not known to draw the card: here the rule chooses", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := keptTo(t, "petya", figures)
			session := serveLesson(t, kept, &clock{at: lessonDay}, nil).connectChat(t, "", tc.host)

			words := textOf(t, call(t, session, "next_task", map[string]any{"language": "en"}))
			wantSaid(t, words, tc.said)
			wantBriefOn(t, kept, tc.keeps)
		})
	}
}

// wantBriefOn holds the request the profile keeps open to being on figures
// and solved from the picture, or to being on neither.
func wantBriefOn(t *testing.T, kept store.Storage, onFigures bool) {
	t.Helper()

	p, _ := loadKept(t, kept)
	if p.OpenRequest == nil {
		t.Fatal("no request is open, want one")
	}
	brief := &p.OpenRequest.Brief
	if (brief.TargetConcept == figures) != onFigures || brief.FromPicture() != onFigures {
		t.Errorf("the request open is on %s, solved from the picture %v; want it on %s %v",
			brief.TargetConcept, brief.FromPicture(), figures, onFigures)
	}
}

// The package of a topic solved from the picture shows the model the example
// of the kind its tasks are drawn in, the lattice, though the topic has no
// reference task of its own to show one.
func TestAPackageOfATopicSolvedFromThePictureShowsItsKindOfPicture(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	h := serveLesson(t, kept, &clock{at: lessonDay}, nil)
	cards := h.connectChat(t, "", drawingCards)
	call(t, cards, "next_task", figuresChoice)
	p, _ := loadKept(t, kept)
	if text := fetchPackage(t, cards, p.OpenRequest.ID); !strings.Contains(text, `"kind":"lattice"`) {
		t.Errorf("the package of %s shows no lattice:\n%s", figures, text)
	}
}
