package content_test

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
	"unicode"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// filledDir holds every drawing frame filled for one reference task, the way a
// model fills it: the same frame with the numbers that task's question gives.
const filledDir = "testdata/drawings"

// filled is a frame as a model fills it for one task.
type filled struct {
	From      string                   `json:"from"`
	Drawing   []string                 `json:"drawing"`
	Structure *checks.DrawingStructure `json:"drawing_structure"`
}

// emojiLike are the characters a drawing may hold that some platforms draw as
// a colour emoji two cells wide, while the check counts each as one cell. A
// frame keeps away from them, since it is copied into every drawing made
// from it.
var emojiLike = []rune{'↔', '↕', '↖', '↗', '↘', '↙', '↩', '↪', '▪', '▫', '▶', '◀', '◻', '◼', '◽', '◾'}

// Every frame is a drawing the checks accept as it stands, with the limits
// every submitted drawing is held to: a frame that failed them would teach the
// model to fail them too.
func TestEveryDrawingFramePassesTheDrawingChecks(t *testing.T) {
	t.Parallel()

	frames := loaded(t).Frames()
	if len(frames) == 0 {
		t.Fatal("the content has no drawing frames, so nothing here was tested")
	}
	for _, frame := range frames {
		t.Run(frame.Name, func(t *testing.T) {
			t.Parallel()

			holdsToTheDrawingChecks(t, "", frame.Drawing, structureOf(frame.Structure))
		})
	}
}

// Every frame, filled with what a reference task of one of its topics gives, is
// still a drawing the checks accept, held against that task's own question. A
// frame that broke once filled — a label run into its number, a number too
// wide for its place — would break every task drawn from it. The child sees a
// drawing before answering, so a filled one leaves no place for a number
// unfilled.
func TestEveryDrawingFrameFilledForAReferenceTaskPassesTheDrawingChecks(t *testing.T) {
	t.Parallel()

	shipped := loaded(t)
	tasks := make(map[string]content.Example, shipped.ExampleCount())
	for _, task := range shipped.Examples() {
		tasks[task.ID] = task
	}
	for _, frame := range shipped.Frames() {
		t.Run(frame.Name, func(t *testing.T) {
			t.Parallel()

			fill := readFilled(t, frame.Name)
			task, known := tasks[fill.From]
			switch {
			case !known:
				t.Fatalf("filled for %q, and no reference task is called that", fill.From)
			case !slices.Contains(frame.Topics, task.Topic):
				t.Fatalf("filled for %s, a task of %s, want a task of one of %v", fill.From, task.Topic, frame.Topics)
			}
			drawing := strings.Join(fill.Drawing, "\n")
			if strings.Contains(drawing, "#") {
				t.Errorf("the filled drawing leaves a place for a number unfilled:\n%s", drawing)
			}
			holdsToTheDrawingChecks(t, task.Question, drawing, fill.Structure)
		})
	}
}

// Every reference task that carries a drawing carries one the checks accept,
// against its own question: the model is shown these drawings beside the
// frames, and a drawing that failed a check would teach it to fail. A
// reference task is a filled drawing, so it leaves no place for a number
// unfilled.
func TestEveryReferenceDrawingPassesTheDrawingChecks(t *testing.T) {
	t.Parallel()

	drawn := 0
	for _, task := range loaded(t).Examples() {
		if task.Drawing == "" {
			continue
		}
		drawn++
		t.Run(task.ID, func(t *testing.T) {
			t.Parallel()

			if strings.Contains(task.Drawing, "#") {
				t.Errorf("the drawing leaves a place for a number unfilled:\n%s", task.Drawing)
			}
			holdsToTheDrawingChecks(t, task.Question, task.Drawing, structureOf(task.DrawingStructure))
		})
	}
	if drawn == 0 {
		t.Fatal("no reference task carries a drawing, so nothing here was tested")
	}
}

// Every filled drawing fills a frame the content has, so that a frame renamed
// or taken out leaves nothing behind that no test reads.
func TestEveryFilledDrawingFillsAFrame(t *testing.T) {
	t.Parallel()

	names := map[string]bool{}
	for _, frame := range loaded(t).Frames() {
		names[frame.Name+".json"] = true
	}
	entries, err := os.ReadDir(filledDir)
	if err != nil {
		t.Fatalf("read %s: %v", filledDir, err)
	}
	for _, entry := range entries {
		if !names[entry.Name()] {
			t.Errorf("%s/%s fills no frame the content has", filledDir, entry.Name())
		}
	}
}

// holdsToTheDrawingChecks fails unless a drawing passes the checks a submitted
// drawing passes and its labels agree with its structure both ways round:
// every label the structure declares is drawn, and everything drawn in letters
// is a declared label, so that no word or unit comes along into a task in
// another language. A drawing and its structure are both required here,
// because the checks let a missing one pass without a word.
func holdsToTheDrawingChecks(t *testing.T, question, drawing string, structure *checks.DrawingStructure) {
	t.Helper()

	if strings.TrimSpace(drawing) == "" || structure == nil {
		t.Fatalf("drawing %q with structure %v, want both", drawing, structure)
	}
	for _, problem := range checks.DrawingFormat(drawing, checks.DefaultDrawingLimits()) {
		t.Errorf("format: %s", problem.Message)
	}
	for _, problem := range checks.DrawingMatch(question, drawing, structure) {
		t.Errorf("labels: %s", problem.Message)
	}

	declared := map[string]bool{}
	for _, object := range structure.Objects {
		declared[object.Label] = true
	}
	for _, written := range lettered(drawing) {
		if !declared[written] {
			t.Errorf("%q is drawn and not declared: a drawing holds labels, and no words or units", written)
		}
	}
	for _, r := range drawing {
		if slices.Contains(emojiLike, r) {
			t.Errorf("%c (U+%04X) is drawn, and a platform may draw it as a wide colour emoji", r, r)
		}
	}
}

// lettered are the runs of letters and digits in a drawing that hold a letter:
// its labels, and anything written in words.
func lettered(drawing string) []string {
	var found []string
	for _, run := range strings.FieldsFunc(drawing, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) }) {
		if strings.IndexFunc(run, unicode.IsLetter) >= 0 {
			found = append(found, run)
		}
	}
	return found
}

// structureOf is a frame's structure in the form the checks read a task's.
func structureOf(structure *content.DrawingStructure) *checks.DrawingStructure {
	if structure == nil {
		return nil
	}
	converted := &checks.DrawingStructure{Kind: structure.Kind}
	for _, object := range structure.Objects {
		converted.Objects = append(converted.Objects, checks.DrawingObject(object))
	}
	for _, relation := range structure.Relations {
		converted.Relations = append(converted.Relations, checks.DrawingRelation(relation))
	}
	return converted
}

// readFilled reads one frame as it was filled for a reference task, refusing
// any member a filled drawing does not have.
func readFilled(t *testing.T, name string) filled {
	t.Helper()

	raw, err := os.ReadFile(filledDir + "/" + name + ".json")
	if err != nil {
		t.Fatalf("frame %s is filled for no reference task: %v", name, err)
	}
	var fill filled
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fill); err != nil {
		t.Fatalf("read the filled %s: %v", name, err)
	}
	return fill
}
