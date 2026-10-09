package profile_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// drawn is the written task with a picture, its keys out of order as a model
// may write them.
func drawn(picture string) *profile.Written {
	task := written()
	task.Picture = json.RawMessage(picture)
	return task
}

// A task's picture is kept as the checks accepted it, with the keys of every
// object in order as the rest of the file is written, and read back the same.
func TestAPictureIsKeptWithItsKeysInOrder(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "en", asked)
	task, err := p.Issue(drawn(`{"time":"4:30","kind":"clock","minute_label":"M"}`), secret(), newSealer(t), asked)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
	if want := `{"kind":"clock","minute_label":"M","time":"4:30"}`; string(task.Picture) != want {
		t.Errorf("the picture kept is %s, want %s", task.Picture, want)
	}

	file, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	read, err := profile.Parse(file)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if !sameJSON(t, read.CurrentTask.Picture, task.Picture) {
		t.Errorf("the picture read back is %s, want %s", read.CurrentTask.Picture, task.Picture)
	}
}

// A task written ahead keeps its picture, and takes it to the card when it is
// handed out.
func TestAKeptTaskTakesItsPictureToTheCard(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "masha")
	brief := briefOn("counting.gaps")
	if _, err := p.AskAhead(&brief, profile.TutorRule, "en", "", asked); err != nil {
		t.Fatalf("AskAhead() error = %v, want nil", err)
	}
	picture := `{"items":[{"label":"1"},{"skip":true},{"label":"9"}],"kind":"row"}`
	kept, err := p.Keep(drawn(picture), secret(), newSealer(t), keptAt)
	if err != nil {
		t.Fatalf("Keep() error = %v, want nil", err)
	}
	task, err := p.HandOutReady(keptAt)
	if err != nil {
		t.Fatalf("HandOutReady() error = %v, want nil", err)
	}
	if string(task.Picture) != picture || string(kept.Picture) != picture {
		t.Errorf("the picture on the card is %s, kept %s, want %s", task.Picture, kept.Picture, picture)
	}
}

// A task with no picture keeps none, and the file says nothing of one.
func TestATaskWithNoPictureKeepsNone(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "en", asked)
	task, err := p.Issue(written(), secret(), newSealer(t), asked)
	if err != nil || task.Picture != nil {
		t.Fatalf("Issue() = %s, %v, want a task with no picture", task.Picture, err)
	}
	file, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	if bytes.Contains(file, []byte(`"picture"`)) {
		t.Error("the file names a picture for a task with none")
	}
}

// A picture that is no JSON is not kept, and nothing changes: the request
// stays open for a task that can be.
func TestAPictureThatIsNoJSONIsNotKept(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	brief := briefOn("counting.gaps")
	p.Ask(&brief, profile.TutorRule, "en", asked)
	if _, err := p.Issue(drawn(`{"kind":`), secret(), newSealer(t), asked); err == nil {
		t.Fatal("Issue() error = nil, want the picture refused")
	}
	if p.OpenRequest == nil {
		t.Error("the request was closed for a task that was not handed out")
	}
}

// A file written while tasks were drawn in characters still reads. The text
// drawing of the task on the card, and of the task kept, is read past and gone
// from the file once it is written again: no card and no word shows it.
func TestAFileFromBeforePicturesIsReadPastItsDrawings(t *testing.T) {
	t.Parallel()

	sealer := newSealer(t)
	p := parseFixture(t, "masha")
	keepOne(t, p, sealer, "counting.gaps")
	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	before := leftIn(t, written, map[string]map[string]any{
		"current_task": {"drawing": "A  B\n●──●"},
		"ready_task":   {"drawing": "|--3--|"},
	})

	read, err := profile.Parse(before)
	if err != nil {
		t.Fatalf("Parse() error = %v, want the text drawings read past", err)
	}
	again, err := profile.Marshal(read)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	if !bytes.Equal(again, written) {
		t.Errorf("the file written again is\n%s\nwant it as it was without the drawings:\n%s", again, written)
	}
}

// sameJSON says whether two documents hold the same values, whatever their
// spacing.
func sameJSON(t *testing.T, one, other json.RawMessage) bool {
	t.Helper()

	var first, second any
	if json.Unmarshal(one, &first) != nil || json.Unmarshal(other, &second) != nil {
		return false
	}
	return fmt.Sprint(first) == fmt.Sprint(second)
}

func TestAPictureHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	keys := gen.SliceOf(gen.OneConstOf("kind", "time", "rows", "label", "zeta", "alpha", "Middle", "m2")).
		Map(func(drawn []string) []string { return slices.Compact(slices.Sorted(slices.Values(drawn))) })

	properties.Property("a picture is written with its keys in order whatever order it came in, and read back "+
		"the same", prop.ForAll(
		func(names []string, reverse bool) bool {
			order := slices.Clone(names)
			if reverse {
				slices.Reverse(order)
			}
			members := make([]string, 0, len(order))
			for i, name := range order {
				members = append(members, fmt.Sprintf(`%q:{"b":%d,"a":[%d,"x<y"]}`, name, i, i))
			}
			came := "{" + strings.Join(members, ",") + "}"

			p := parseFixture(t, "dima")
			brief := briefOn("counting.gaps")
			p.Ask(&brief, profile.TutorRule, "en", asked)
			task, err := p.Issue(drawn(came), secret(), newSealer(t), asked)
			if err != nil {
				return false
			}
			inOrder := encoded(t, came)
			file, err := profile.Marshal(p)
			if err != nil {
				return false
			}
			read, err := profile.Parse(file)
			return err == nil && bytes.Equal(task.Picture, inOrder) &&
				bytes.Equal(compacted(t, read.CurrentTask.Picture), inOrder)
		},
		keys, gen.Bool(),
	))

	properties.TestingRun(t)
}

// encoded is a document written with the keys of every object in order and
// nothing escaped, as encoding/json writes a value it has read.
func encoded(t *testing.T, document string) []byte {
	t.Helper()

	var value any
	if err := json.Unmarshal([]byte(document), &value); err != nil {
		t.Fatalf("read %s: %v", document, err)
	}
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		t.Fatalf("write %s: %v", document, err)
	}
	return bytes.TrimRight(out.Bytes(), "\n")
}

// compacted is a document with no spacing.
func compacted(t *testing.T, document []byte) []byte {
	t.Helper()

	var out bytes.Buffer
	if err := json.Compact(&out, document); err != nil {
		t.Fatalf("compact %s: %v", document, err)
	}
	return out.Bytes()
}
