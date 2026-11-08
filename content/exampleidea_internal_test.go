package content

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

// editedTask rewrites one reference task of counting.gaps in a copy of the
// content.
func editedTask(t *testing.T, src fstest.MapFS, id string, edit func(task map[string]any)) {
	t.Helper()

	name := examplesDir + "/counting.gaps" + examplesSuffix
	var tasks []map[string]any
	if err := json.Unmarshal(src[name].Data, &tasks); err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	for _, task := range tasks {
		if task["id"] == id {
			edit(task)
		}
	}
	src[name] = asFile(t, tasks)
}

// A reference task names an idea only where a package of that idea can show
// it: an idea its list holds as solved from the picture, in a topic told in
// words, by a task that has a picture to solve it from, and by no other task.
func TestAReferenceTaskOfAnIdeaIsHeldToIt(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		edit func(t *testing.T, src fstest.MapFS)
		want string
	}{
		{"an idea told in words", func(t *testing.T, src fstest.MapFS) {
			editedTask(t, src, "gaps-12-d2-7", func(task map[string]any) { task["idea"] = 1 })
		}, "gaps-12-d2-7: names idea 1 of counting.gaps at 1-2, which its list does not hold as one solved from the picture"},
		{"a place the list does not have", func(t *testing.T, src fstest.MapFS) {
			editedTask(t, src, "gaps-12-d2-7", func(task map[string]any) { task["idea"] = 99 })
		}, "gaps-12-d2-7: names idea 99 of counting.gaps at 1-2"},
		{"a place before the first", func(t *testing.T, src fstest.MapFS) {
			editedTask(t, src, "gaps-12-d2-7", func(task map[string]any) { task["idea"] = -1 })
		}, `"gaps-12-d2-7": idea -1 is no place on a list of ideas, which counts from 1`},
		{"no picture", func(t *testing.T, src fstest.MapFS) {
			editedTask(t, src, "gaps-12-d2-7", func(task map[string]any) { delete(task, "picture") })
		}, "gaps-12-d2-7: names an idea solved from the picture, and has no picture to solve it from"},
		{"a second task of one idea", func(t *testing.T, src fstest.MapFS) {
			editedTask(t, src, "gaps-12-d2-6", func(task map[string]any) { task["idea"] = 15 })
		}, "names idea 15 of counting.gaps at 1-2, as "},
		{"a topic solved from the picture", func(t *testing.T, src fstest.MapFS) {
			solvedFromPicture(t, src, "counting.gaps")
		}, "every task of counting.gaps is solved from the picture, so its reference tasks name none"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			test.edit(t, src)
			wantProblem(t, src, test.want)
		})
	}
}
