package mcpserver

import (
	"encoding/json"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// A task solved from the picture is kept by the sketch of its picture, and a
// word task by the sketch of its question.
func TestATaskIsKeptByTheSketchOfWhatMakesItNew(t *testing.T) {
	t.Parallel()
	task := &checks.Task{
		Question: "How many triangles does the figure hold?",
		Picture:  json.RawMessage(`{"kind":"lattice","width":2,"height":2,"lines":[[0,0,2,2],[0,2,2,0]]}`),
	}
	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	s := &Service{content: loaded}
	words := &profile.OpenRequest{Language: "en"}
	drawn := &profile.OpenRequest{Language: "en", Brief: profile.Brief{Constraints: []string{profile.FactsOnThePicture}}}

	if got, want := s.writtenOf(task, words).Fingerprint, checks.Fingerprint(task.Question, "en"); got != want {
		t.Errorf("a word task is kept by %q, want the sketch of its question", got)
	}
	if got, want := s.writtenOf(task, drawn).Fingerprint, checks.PictureFingerprint(task.Picture); got != want {
		t.Errorf("a task solved from the picture is kept by %q, want the sketch of its picture", got)
	}
}
