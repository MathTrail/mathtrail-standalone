package content

import (
	"maps"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

// framePath is where the frame of this name lives.
func framePath(name string) string { return drawingsDir + "/" + name + frameSuffix }

// sampleFrame is a frame with nothing wrong with it: two posts and the gap
// between them.
func sampleFrame() frameFile {
	return frameFile{
		Purpose: "Two posts and the gap between them.",
		Topics:  []string{"counting.gaps"},
		Drawing: []string{"A     B", "●─────●", "   ##"},
		Structure: &DrawingStructure{
			Kind:      "row",
			Objects:   []DrawingObject{{ID: "A", Label: "A"}, {ID: "B", Label: "B"}},
			Relations: []DrawingRelation{{Type: "next_to", From: "A", To: "B"}},
		},
	}
}

// withFrame puts a frame, by name, among the real ones.
func withFrame(t *testing.T, src fstest.MapFS, name string, frame *frameFile) {
	t.Helper()
	src[framePath(name)] = asFile(t, frame)
}

func TestAFrameIsReadAsItsFileHoldsIt(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	written := sampleFrame()
	withFrame(t, src, "posts", &written)
	loaded, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	want := Frame{
		Name: "posts", Purpose: written.Purpose, Topics: written.Topics,
		Drawing: "A     B\n●─────●\n   ##", Structure: written.Structure,
	}
	for _, frame := range loaded.Frames() {
		if frame.Name == "posts" {
			if !reflect.DeepEqual(frame, want) {
				t.Errorf("frame = %+v, want %+v", frame, want)
			}
			return
		}
	}
	t.Error("the frame was not read at all")
}

// A frame the loader cannot vouch for stops the service: one that is not a
// frame by its name, one that does not say what it is for, one no package
// would carry, one that draws nothing, and one whose structure cannot be read
// as data or holds the values only a task's drawing has.
func TestABrokenFrameStopsTheService(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		spoil func(*frameFile)
		want  string
	}{
		{"an empty purpose", func(f *frameFile) { f.Purpose = " " }, "the purpose is empty"},
		{"no topic", func(f *frameFile) { f.Topics = nil }, "the frame names no topic"},
		{"a topic nobody has", func(f *frameFile) { f.Topics = []string{"counting.everything"} },
			`topic "counting.everything" is not in the catalog`},
		{"a topic named twice", func(f *frameFile) { f.Topics = []string{"counting.gaps", "counting.gaps"} },
			`topic "counting.gaps" is named twice`},
		{"no drawing", func(f *frameFile) { f.Drawing = []string{"   "} }, "the drawing is empty"},
		{"a line that breaks in two", func(f *frameFile) { f.Drawing[0] = "A\n     B" },
			"line 1 of the drawing breaks in two"},
		{"no structure", func(f *frameFile) { f.Structure = nil }, "the frame has no drawing_structure"},
		{"a structure that says nothing is drawn", func(f *frameFile) { f.Structure.Kind = "" },
			"drawing_structure: the drawing structure does not say what is drawn"},
		{"a value, which only a task has", func(f *frameFile) {
			value := 5
			f.Structure.Objects[0].Value = &value
		}, `drawing_structure: drawn object "A" has a value`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			frame := sampleFrame()
			test.spoil(&frame)
			withFrame(t, src, "posts", &frame)
			wantProblem(t, src, framePath("posts")+":\n  - "+test.want)
		})
	}
}

// Something among the frames that is not a frame by its name stops the
// service, and so does a member the format does not have: a misspelt key
// would otherwise be dropped on the way in.
func TestSomethingThatIsNotAFrameStopsTheService(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, file, data, want string
	}{
		{"a directory named like a frame", drawingsDir + "/old.json/posts.json", "{}",
			"old.json: a frame is a file named in lowercase words"},
		{"a file of another kind", drawingsDir + "/notes.txt", "a note", "notes.txt: a frame is a file named"},
		{"a name that is not lowercase words", drawingsDir + "/Posts.json", "{}", "Posts.json: a frame is a file named"},
		{"a member the format does not have", framePath("posts"), `{"purpose": "Posts.", "colour": "red"}`,
			`unknown field "colour"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			src[test.file] = &fstest.MapFile{Data: []byte(test.data)}
			wantProblem(t, src, test.want)
		})
	}
}

// A frame or the frames' directory that is there and cannot be read stops the
// service, and so does content without the directory at all.
func TestFramesThatCannotBeReadStopTheService(t *testing.T) {
	t.Parallel()

	for _, name := range []string{framePath("row"), drawingsDir} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wantProblem(t, unreadable{MapFS: contentCopy(t), name: name}, "content: read "+name)
		})
	}
	t.Run("no directory", func(t *testing.T) {
		t.Parallel()

		src := contentCopy(t)
		for name := range src {
			if strings.HasPrefix(name, drawingsDir+"/") {
				delete(src, name)
			}
		}
		wantProblem(t, src, "content: read "+drawingsDir)
	})
}

// A frame is part of what the model is told, so the version every log line
// carries follows it: redrawn, renamed or given to other topics, it is a
// different version.
func TestAChangedFrameChangesTheInstructionsVersion(t *testing.T) {
	t.Parallel()

	base := contentCopy(t)
	written := sampleFrame()
	withFrame(t, base, "posts", &written)
	before, err := load(base)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	for _, test := range []struct {
		name   string
		change func(t *testing.T, src fstest.MapFS)
	}{
		{"it is redrawn", func(t *testing.T, src fstest.MapFS) {
			frame := sampleFrame()
			frame.Drawing[1] = "●───────●"
			withFrame(t, src, "posts", &frame)
		}},
		{"it is renamed", func(t *testing.T, src fstest.MapFS) {
			delete(src, framePath("posts"))
			frame := sampleFrame()
			withFrame(t, src, "fence", &frame)
		}},
		{"it is given to another topic too", func(t *testing.T, src fstest.MapFS) {
			frame := sampleFrame()
			frame.Topics = append(frame.Topics, "logic.ordering")
			withFrame(t, src, "posts", &frame)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := maps.Clone(base)
			test.change(t, src)
			after, err := load(src)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if after.InstructionsVersion() == before.InstructionsVersion() {
				t.Errorf("version = %q both times, want a different one once the frame changed",
					after.InstructionsVersion())
			}
		})
	}
}
