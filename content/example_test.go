package content

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

const (
	gapsFile   = examplesDir + "/counting.gaps.json"
	knightFile = examplesDir + "/logic.knights_liars.json"
)

// withExamples puts one file of reference tasks in place of the real one, so
// that a test is looking at the tasks it wrote and nothing else. The solvers of
// the tasks it replaces go with them: they prove answers to questions that are
// no longer there.
func withExamples(t *testing.T, src fstest.MapFS, file string, tasks ...Example) {
	t.Helper()

	var replaced []Example
	if existing, there := src[file]; there {
		if err := json.Unmarshal(existing.Data, &replaced); err != nil {
			t.Fatalf("read the reference tasks of %s: %v", file, err)
		}
	}
	for i := range replaced {
		delete(src, solverFile(replaced[i].ID))
	}
	src[file] = asFile(t, tasks)
}

func TestAReferenceTaskWithNothingWrongIsAccepted(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withExamples(t, src, gapsFile, validExample())
	if _, err := load(src); err != nil {
		t.Fatalf("load: %v", err)
	}
}

func TestBrokenReferenceTaskStopsTheService(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		change func(*Example)
		want   string
	}{
		{
			name:   "an id that is not lowercase words joined by dashes",
			change: func(task *Example) { task.ID = "Gaps Posts 1" },
			want:   "an id is lowercase words joined by dashes",
		},
		{
			name:   "a difficulty outside the five",
			change: func(task *Example) { task.Difficulty = 6 },
			want:   "difficulty 6 is outside 1-5",
		},
		{
			name:   "no question",
			change: func(task *Example) { task.Question = "" },
			want:   "the question is empty",
		},
		{
			name:   "no solution",
			change: func(task *Example) { task.Solution = "" },
			want:   "the solution is empty",
		},
		{
			name:   "two options a child could not tell apart",
			change: func(task *Example) { task.Options["E"] = " 3 METRES " },
			want:   "options B and E read the same",
		},
		{
			name:   "an option that is missing",
			change: func(task *Example) { delete(task.Options, "E") },
			want:   "option E is missing",
		},
		{
			name:   "an option with nothing in it",
			change: func(task *Example) { task.Options["E"] = "  " },
			want:   "option E is empty",
		},
		{
			name:   "a correct answer that is not one of the five",
			change: func(task *Example) { task.CorrectAnswer = "F" },
			want:   `the correct answer is "F"`,
		},
		{
			name:   "an explanation attached to the correct answer",
			change: func(task *Example) { task.Distractors["B"] = Distractor{Trap: "off_by_one", Text: "It is right."} },
			want:   "option B is the correct answer and needs no explanation",
		},
		{
			name:   "a wrong option nobody explains",
			change: func(task *Example) { delete(task.Distractors, "E") },
			want:   "wrong option E has no explanation",
		},
		{
			name: "an explanation with no words in it",
			change: func(task *Example) {
				task.Distractors["A"] = Distractor{Trap: "off_by_one", Text: ""}
			},
			want: "option A explains nothing",
		},
		{
			name: "a mistake that is in no catalog",
			change: func(task *Example) {
				task.Distractors["A"] = Distractor{Trap: "forgot_a_case", Text: "One gap is not the fence."}
			},
			want: `trap "forgot_a_case", which is not in the catalog`,
		},
		{
			name:   "a drawing nothing describes",
			change: func(task *Example) { task.Drawing = "o--o--o--o" },
			want:   "a drawing without its structure cannot be checked against the wording",
		},
		{
			name: "a description of a drawing that is not there",
			change: func(task *Example) {
				task.DrawingStructure = &DrawingStructure{
					Kind:    "number_line",
					Objects: []DrawingObject{{ID: "A", Label: "A"}},
				}
			},
			want: "a drawing structure describes a drawing, and there is none",
		},
		{
			name: "a drawing that does not say what it draws",
			change: func(task *Example) {
				task.Drawing = "o--o"
				task.DrawingStructure = &DrawingStructure{Objects: []DrawingObject{{ID: "A", Label: "A"}}}
			},
			want: "the drawing structure does not say what is drawn",
		},
		{
			name: "a drawing that names nothing in it",
			change: func(task *Example) {
				task.Drawing = "o--o"
				task.DrawingStructure = &DrawingStructure{Kind: "number_line"}
			},
			want: "the drawing structure names nothing that is drawn",
		},
		{
			name: "the same drawn object twice",
			change: func(task *Example) {
				task.Drawing = "o--o"
				task.DrawingStructure = &DrawingStructure{
					Kind:    "number_line",
					Objects: []DrawingObject{{ID: "A", Label: "A"}, {ID: "A", Label: "B"}},
				}
			},
			want: `drawn object "A" appears twice`,
		},
		{
			name: "a drawn object with no label to look for",
			change: func(task *Example) {
				task.Drawing = "o--o"
				task.DrawingStructure = &DrawingStructure{
					Kind:    "number_line",
					Objects: []DrawingObject{{ID: "A"}},
				}
			},
			want: "drawn object 1 has no label",
		},
		{
			name: "a relation that is not a triple",
			change: func(task *Example) {
				task.Drawing = "o--o"
				task.DrawingStructure = &DrawingStructure{
					Kind:      "number_line",
					Objects:   []DrawingObject{{ID: "A", Label: "A"}},
					Relations: []DrawingRelation{{Type: "left_of", From: "A"}},
				}
			},
			want: "relation 1 is not a type, a from and a to",
		},
		{
			name: "a relation that starts at something nobody drew",
			change: func(task *Example) {
				task.Drawing = "o--o"
				task.DrawingStructure = &DrawingStructure{
					Kind:      "number_line",
					Objects:   []DrawingObject{{ID: "A", Label: "A"}},
					Relations: []DrawingRelation{{Type: "left_of", From: "Z", To: "A"}},
				}
			},
			want: `relation 1 starts at "Z", which is not drawn`,
		},
		{
			name: "a relation that ends at something nobody drew",
			change: func(task *Example) {
				task.Drawing = "o--o"
				task.DrawingStructure = &DrawingStructure{
					Kind:      "number_line",
					Objects:   []DrawingObject{{ID: "A", Label: "A"}},
					Relations: []DrawingRelation{{Type: "left_of", From: "A", To: "B"}},
				}
			},
			want: `relation 1 ends at "B", which is not drawn`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := validExample()
			tc.change(&task)
			src := contentCopy(t)
			withExamples(t, src, gapsFile, task)
			wantProblem(t, src, tc.want)
		})
	}
}

func TestTheSameTaskTwiceStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withExamples(t, src, gapsFile, validExample(), validExample())
	wantProblem(t, src, `task "gaps-posts-test": the id is used twice`)
}

func TestATaskInTheWrongFileStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withExamples(t, src, knightFile, validExample())
	wantProblem(t, src, `the topic is "counting.gaps", and this file holds "logic.knights_liars"`)
}

// A topic is taught at the levels its catalog entry lists, and a task at any
// other level would be handed to a child it was not written for.
func TestATaskAtALevelItsTopicIsNotTaughtAtStopsTheService(t *testing.T) {
	t.Parallel()

	task := validExample()
	task.ID = "knights-12-d3-1"
	task.Topic = "logic.knights_liars"
	task.GradeLevel = Level12

	src := contentCopy(t)
	withExamples(t, src, knightFile, task)
	wantProblem(t, src, `grade level "1-2" is not one topic "logic.knights_liars" is offered at`)
}

func TestTasksOfATopicNobodyHasStopTheService(t *testing.T) {
	t.Parallel()

	task := validExample()
	task.Topic = "patterns.sequences"

	src := contentCopy(t)
	withExamples(t, src, examplesDir+"/patterns.sequences.json", task)
	wantProblem(t, src, `"patterns.sequences" is not a topic in the catalog`)
}

func TestAFileOfTasksThatIsNotAListStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[gapsFile] = asFile(t, validExample())
	wantProblem(t, src, "content: parse "+gapsFile)
}

func TestATaskWithAFieldTheFormatDoesNotHaveStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[gapsFile] = &fstest.MapFile{Data: []byte(`[{"id":"gaps-posts-test","answer":"B"}]`)}
	wantProblem(t, src, `unknown field "answer"`)
}

// No reference task carries a drawing yet, and the branch that copies one is
// the deepest thing in the content: a pointer, two slices and a value behind a
// pointer of its own.
func TestCloningATaskCopiesItsDrawing(t *testing.T) {
	t.Parallel()

	value := 3
	original := validExample()
	original.Drawing = "A--B"
	original.DrawingStructure = &DrawingStructure{
		Kind:      "number_line",
		Objects:   []DrawingObject{{ID: "A", Label: "A", Value: &value}},
		Relations: []DrawingRelation{{Type: "left_of", From: "A", To: "B"}},
	}

	copied := original.clone()
	copied.DrawingStructure.Kind = "grid"
	copied.DrawingStructure.Objects[0].Label = "Z"
	*copied.DrawingStructure.Objects[0].Value = 7
	copied.DrawingStructure.Relations[0].Type = "right_of"

	structure := original.DrawingStructure
	if structure.Kind != "number_line" {
		t.Errorf("kind = %q, want it untouched at %q", structure.Kind, "number_line")
	}
	if got := structure.Objects[0].Label; got != "A" {
		t.Errorf("label = %q, want it untouched at %q", got, "A")
	}
	if got := *structure.Objects[0].Value; got != 3 {
		t.Errorf("value = %d, want it untouched at %d", got, 3)
	}
	if got := structure.Relations[0].Type; got != "left_of" {
		t.Errorf("relation = %q, want it untouched at %q", got, "left_of")
	}
}

// The reference tasks sit in one file per topic, beside the directory of their
// solvers. Anything else there is a mistake nobody would otherwise notice: tasks
// in a directory of their own, or in a file that is never read as tasks.
func TestSomethingElseAmongTheReferenceTasksStopsTheService(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, file, want string
	}{
		{"tasks in a directory of their own", examplesDir + "/counting/counting.gaps.json",
			"reference tasks live in files, one per topic, not in directories"},
		{"a file that is not a topic's", examplesDir + "/notes.md",
			"a file of reference tasks is named after its topic and ends in .json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			src[tc.file] = &fstest.MapFile{Data: []byte("[]\n")}
			wantProblem(t, src, tc.want)
		})
	}
}

// solverFile is where the program that proves one task's answer lives.
func solverFile(id string) string {
	return examplesDir + "/" + solversDir + "/" + id + solverSuffix
}

func TestATaskIsGivenTheSolverStoredBesideIt(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withExamples(t, src, gapsFile, validExample())
	src[solverFile("gaps-posts-test")] = &fstest.MapFile{
		Data: []byte("def solve(options):\n    return match(options, 3)\n"),
	}

	loaded, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, task := range loaded.Examples() {
		if task.ID != "gaps-posts-test" {
			continue
		}
		if want := "match(options, 3)"; !strings.Contains(task.Solver, want) {
			t.Fatalf("solver = %q, want one containing %q", task.Solver, want)
		}
		return
	}
	t.Fatal("the task was not loaded at all")
}

// A solver naming no task is how a rename half-done looks: the program is still
// there, and nothing would ever run it again.
func TestASolverBelongingToNoTaskStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[solverFile("gaps-posts-renamed")] = &fstest.MapFile{Data: []byte("def solve(options):\n    return []\n")}
	wantProblem(t, src, `no reference task is called "gaps-posts-renamed"`)
}

func TestAnEmptySolverStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withExamples(t, src, gapsFile, validExample())
	src[solverFile("gaps-posts-test")] = &fstest.MapFile{Data: []byte("\n   \n")}
	wantProblem(t, src, "gaps-posts-test.star: the solver is empty")
}

func TestASolverNotNamedAfterATaskStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[examplesDir+"/"+solversDir+"/notes.txt"] = &fstest.MapFile{Data: []byte("a note\n")}
	wantProblem(t, src, "notes.txt: a solver is named after its task and ends in .star")
}

func TestSolversInDirectoriesOfTheirOwnStopTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[examplesDir+"/"+solversDir+"/counting.gaps/gaps-posts-test.star"] = &fstest.MapFile{
		Data: []byte("def solve(options):\n    return []\n"),
	}
	wantProblem(t, src, "counting.gaps: a solver is a file, not a directory")
}

// The solvers are part of the content: without their directory no reference task
// can prove its answer, and a binary like that is not one to serve from.
func TestContentWithoutItsSolversStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	for name := range src {
		if strings.HasPrefix(name, examplesDir+"/"+solversDir+"/") {
			delete(src, name)
		}
	}
	wantProblem(t, src, "content: read "+examplesDir+"/"+solversDir)
}

// unreadable is the content with one file that is listed and cannot be read, as
// a file the process has no permission for would be — the one failure a copy
// in memory cannot show by itself.
type unreadable struct {
	fstest.MapFS
	name string
}

func (u unreadable) Open(name string) (fs.File, error) {
	if name == u.name {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}
	return u.MapFS.Open(name)
}

func (u unreadable) ReadFile(name string) ([]byte, error) {
	if name == u.name {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrPermission}
	}
	return u.MapFS.ReadFile(name)
}

// A solver that is there and cannot be read stops the service, rather than
// leaving its task without the program that proves it.
func TestASolverThatCannotBeReadStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withExamples(t, src, gapsFile, validExample())
	src[solverFile("gaps-posts-test")] = &fstest.MapFile{Data: []byte("def solve(options):\n    return []\n")}
	wantProblem(t, unreadable{MapFS: src, name: solverFile("gaps-posts-test")},
		"content: read "+solverFile("gaps-posts-test"))
}

// The program used to be a field of the task, and a copy left there would be a
// second source of truth that nothing reads.
func TestASolverWrittenIntoTheTaskStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[gapsFile] = &fstest.MapFile{Data: []byte(`[{"id":"gaps-posts-test","solver":"def solve(options):\n    return []\n"}]`)}
	wantProblem(t, src, `unknown field "solver"`)
}
