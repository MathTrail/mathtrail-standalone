package content

import (
	"maps"
	"strings"
	"testing"
	"testing/fstest"
)

// picturePath is where the example of a kind of picture sits.
func picturePath(kind string) string { return picturesDir + "/" + kind + pictureSuffix }

// withPicture puts the example of a kind into a copy of the content.
func withPicture(src fstest.MapFS, kind, data string) {
	src[picturePath(kind)] = &fstest.MapFile{Data: []byte(data)}
}

// An example the model could not start from stops the service, and so does a
// kind with no example: a model shown none describes a picture from memory.
func TestAnExampleOfAPictureTheModelCouldNotStartFromStopsTheService(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(src fstest.MapFS)
		want   string
	}{
		{"a kind with no example", func(src fstest.MapFS) { delete(src, picturePath("ring")) },
			"there is no example of a ring"},
		{"a file named by no kind", func(src fstest.MapFS) {
			withPicture(src, "star", `{"purpose":"A star.","topics":["counting.gaps"],"picture":{"kind":"star"}}`)
		}, "star.json: an example is a file named by the kind of picture it shows"},
		{"a file of another kind", func(src fstest.MapFS) {
			src[picturesDir+"/notes.txt"] = &fstest.MapFile{Data: []byte("a note")}
		}, "notes.txt: an example is a file named by the kind"},
		{"an example that does not say what it is for", func(src fstest.MapFS) {
			withPicture(src, "ring", `{"purpose":" ","topics":["counting.gaps"],"picture":{"kind":"ring","count":5}}`)
		}, "the purpose is empty"},
		{"an example no package carries", func(src fstest.MapFS) {
			withPicture(src, "ring", `{"purpose":"Places.","topics":[],"picture":{"kind":"ring","count":5}}`)
		}, "the example names no topic"},
		{"an example for a topic the catalog lacks", func(src fstest.MapFS) {
			withPicture(src, "ring", `{"purpose":"Places.","topics":["astronomy"],"picture":{"kind":"ring","count":5}}`)
		}, `topic "astronomy" is not in the catalog`},
		{"an example for a topic twice", func(src fstest.MapFS) {
			withPicture(src, "ring",
				`{"purpose":"Places.","topics":["counting.gaps","counting.gaps"],"picture":{"kind":"ring","count":5}}`)
		}, `topic "counting.gaps" is named twice`},
		{"an example of another kind than its file", func(src fstest.MapFS) {
			withPicture(src, "ring", `{"purpose":"Places.","topics":["counting.gaps"],"picture":{"kind":"clock","time":"4:30"}}`)
		}, "the picture is a clock, and the file is the example of a ring"},
		{"an example with no picture", func(src fstest.MapFS) {
			withPicture(src, "ring", `{"purpose":"Places.","topics":["counting.gaps"]}`)
		}, "there is no picture to show the kind by"},
		{"an example out of its kind's limits", func(src fstest.MapFS) {
			withPicture(src, "ring", `{"purpose":"Places.","topics":["counting.gaps"],"picture":{"kind":"ring","count":2}}`)
		}, "picture.count must be a whole number from 3 to 24"},
		{"a member the format does not have", func(src fstest.MapFS) {
			withPicture(src, "ring", `{"purpose":"Places.","colour":"red","topics":["counting.gaps"],`+
				`"picture":{"kind":"ring","count":5}}`)
		}, `unknown field "colour"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			test.change(src)
			wantProblem(t, src, test.want)
		})
	}
}

// The examples' directory, or an example in it, that cannot be read stops the
// service, and so does content without the directory at all.
func TestExamplesOfPicturesThatCannotBeReadStopTheService(t *testing.T) {
	t.Parallel()

	for _, name := range []string{picturePath("ring"), picturesDir} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wantProblem(t, unreadable{MapFS: contentCopy(t), name: name}, "content: read "+name)
		})
	}
	t.Run("no directory", func(t *testing.T) {
		t.Parallel()

		src := contentCopy(t)
		for name := range src {
			if strings.HasPrefix(name, picturesDir+"/") {
				delete(src, name)
			}
		}
		wantProblem(t, src, "content: read "+picturesDir)
	})
}

// An example of a picture is part of what the model is told, so the version
// every log line carries follows it: drawn otherwise, or given to other
// topics, it is a different version.
func TestAChangedExampleOfAPictureChangesTheInstructionsVersion(t *testing.T) {
	t.Parallel()

	base := contentCopy(t)
	before, err := load(base)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, test := range []struct {
		name, data string
	}{
		{"it is drawn otherwise", `{"purpose":"Places in a ring.","topics":["counting.gaps"],` +
			`"picture":{"kind":"ring","count":7}}`},
		{"it is given to other topics", `{"purpose":"Places in a ring.","topics":["counting.gaps","logic.ordering"],` +
			`"picture":{"kind":"ring","count":12}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := maps.Clone(base)
			withPicture(src, "ring", test.data)
			after, err := load(src)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if after.InstructionsVersion() == before.InstructionsVersion() {
				t.Errorf("version = %q both times, want a different one once the example changed",
					after.InstructionsVersion())
			}
		})
	}
}
