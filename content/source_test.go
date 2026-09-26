package content

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// contentCopy returns the content of the binary as a set of files a test can
// edit. Breaking a copy is how a check is shown a fault the real content does
// not have, and the only way to prove that the fault would in fact stop the
// service.
func contentCopy(t *testing.T) fstest.MapFS {
	t.Helper()

	copied := fstest.MapFS{}
	err := fs.WalkDir(files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		raw, err := files.ReadFile(name)
		if err != nil {
			return err
		}
		copied[name] = &fstest.MapFile{Data: raw}
		return nil
	})
	if err != nil {
		t.Fatalf("copy the content of the binary: %v", err)
	}
	return copied
}

// asFile renders a value as the bytes a content file would hold.
func asFile(t *testing.T, value any) *fstest.MapFile {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("render %T as a content file: %v", value, err)
	}
	return &fstest.MapFile{Data: raw}
}

// wantProblem fails unless loading refused the content and said what is wrong
// in words that name the fault the test introduced.
func wantProblem(t *testing.T, src fs.FS, want string) {
	t.Helper()

	_, err := load(src)
	if err == nil {
		t.Fatalf("got no error, want one naming %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("got error %q, want one naming %q", err, want)
	}
}

// validExample is a reference task with nothing wrong with it, so that a test
// breaking one thing is testing that one thing. It is a task of counting.gaps,
// because that topic exists at the level the task claims.
func validExample() Example {
	return Example{
		ID:         "gaps-posts-test",
		Topic:      "counting.gaps",
		GradeLevel: rating.Grades12,
		Difficulty: 3,
		Question:   "A fence has 4 posts in a row, 1 metre apart. How long is the fence?",
		Options: map[string]string{
			"A": "1 metre", "B": "3 metres", "C": "4 metres", "D": "5 metres", "E": "8 metres",
		},
		CorrectAnswer: "B",
		Hint:          "How many gaps are there between 4 posts?",
		Solution:      "4 posts have 3 gaps of 1 metre each, so the fence is 3 metres long.",
		Distractors: map[string]Distractor{
			"A": {Trap: "answered_other_question", Text: "1 metre is one gap, not the whole fence."},
			"C": {Trap: "off_by_one", Text: "4 posts have 3 gaps between them, not 4."},
			"D": {Trap: "off_by_one", Text: "Counting a gap after the last post adds one that is not there."},
			"E": {Trap: "double_count", Text: "Each gap is counted once, not twice."},
		},
	}
}
