package content

import (
	"testing"
	"testing/fstest"
)

func TestInstructionsVersionDependsOnTheContentAndNotOnTheOrder(t *testing.T) {
	t.Parallel()

	written := []instructionFile{
		{name: "mcp_instructions.md", text: "How the tools are used.\n"},
		{name: "task_writing.md", text: "How a task is written.\n"},
		{name: "drawing.md", text: "How a picture is drawn.\n"},
	}
	want := instructionsVersion(written)

	shuffled := []instructionFile{written[2], written[0], written[1]}
	if got := instructionsVersion(shuffled); got != want {
		t.Errorf("read in another order: got %q, want %q", got, want)
	}
}

func TestInstructionsVersionFollowsEveryChange(t *testing.T) {
	t.Parallel()

	base := []instructionFile{
		{name: "mcp_instructions.md", text: "How the tools are used.\n"},
		{name: "task_writing.md", text: "How a task is written.\n"},
	}
	version := instructionsVersion(base)

	for _, tc := range []struct {
		name    string
		changed []instructionFile
	}{
		{
			name: "a word is rewritten",
			changed: []instructionFile{
				{name: "mcp_instructions.md", text: "How the tools are used well.\n"},
				{name: "task_writing.md", text: "How a task is written.\n"},
			},
		},
		{
			// The two files hold the same words as before, split differently. A
			// version that only hashed the text would call this the same
			// instructions, and the log would then point at the wrong wording.
			name: "a sentence moves from one file to the other",
			changed: []instructionFile{
				{name: "mcp_instructions.md", text: "How the tools are used.\nHow a task is written.\n"},
				{name: "task_writing.md", text: ""},
			},
		},
		{
			name: "a file is added",
			changed: []instructionFile{
				{name: "mcp_instructions.md", text: "How the tools are used.\n"},
				{name: "task_writing.md", text: "How a task is written.\n"},
				{name: "drawing.md", text: "How a picture is drawn.\n"},
			},
		},
		{
			name: "a file is renamed",
			changed: []instructionFile{
				{name: "tools.md", text: "How the tools are used.\n"},
				{name: "task_writing.md", text: "How a task is written.\n"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := instructionsVersion(tc.changed); got == version {
				t.Errorf("got the same version %q, want a different one", got)
			}
		})
	}
}

func TestInstructionsThatSayNothingStopTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[instructionsDir+"/task_writing.md"] = &fstest.MapFile{Data: []byte("   \n")}
	wantProblem(t, src, "task_writing.md: the file is empty")
}

func TestAStrayFileAmongTheInstructionsStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[instructionsDir+"/notes.txt"] = &fstest.MapFile{Data: []byte("a note to self")}
	wantProblem(t, src, "notes.txt: the instructions are files ending in .md")
}
