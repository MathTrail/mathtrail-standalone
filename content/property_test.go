package content

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The properties here name what has to hold for every input rather than for the
// handful of inputs somebody thought of. Two things in this package have that
// shape: the version of the instructions, which must depend on the content and
// on nothing else, and the copies handed out of the content, which must share
// no memory with it.

// instructionFilesFrom turns generated texts into a set of files whose names
// differ, the way a directory listing always hands them over. Two files of one
// name could not come from one directory, and the version promises nothing
// about a set that could not exist.
func instructionFilesFrom(texts []string) []instructionFile {
	files := make([]instructionFile, 0, len(texts))
	for i, text := range texts {
		files = append(files, instructionFile{name: fmt.Sprintf("%d.md", i), text: text})
	}
	return files
}

func TestInstructionsVersionHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("the order the files were read in does not reach the version", prop.ForAll(
		func(texts []string, seed int64) bool {
			files := instructionFilesFrom(texts)
			shuffled := slices.Clone(files)
			shuffle := rand.New(rand.NewSource(seed))
			shuffle.Shuffle(len(shuffled), func(i, j int) {
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			})
			return instructionsVersion(files) == instructionsVersion(shuffled)
		},
		gen.SliceOf(gen.AnyString()),
		gen.Int64(),
	))

	properties.Property("a version is always twelve hexadecimal characters", prop.ForAll(
		func(texts []string) bool {
			return regexp.MustCompile(`^[0-9a-f]{12}$`).MatchString(
				instructionsVersion(instructionFilesFrom(texts)))
		},
		gen.SliceOf(gen.AnyString()),
	))

	// What the length in front of each text buys: the same words split
	// differently are a different set of instructions, and a version that
	// hashed the text alone would call the two the same.
	properties.Property("the same words in one file are not the same version as in several", prop.ForAll(
		func(texts []string) bool {
			if len(texts) < 2 {
				return true
			}
			several := instructionFilesFrom(texts)
			one := instructionFilesFrom([]string{strings.Join(texts, "")})
			return instructionsVersion(several) != instructionsVersion(one)
		},
		gen.SliceOf(gen.AnyString()),
	))

	properties.TestingRun(t)
}

// exampleSeed is the raw material of a reference task: plain values a generator
// can produce, and a build that turns them into the same task every time. Two
// tasks built from one seed are what lets a property say "the original is
// untouched" without trusting the copy it is testing.
type exampleSeed struct {
	Question string
	Options  []string
	Traps    []string
	Texts    []string
	Kind     string
	Label    string
	Relation string
	Value    int
	Drawn    bool
}

// build assembles a reference task holding everything a copy has to carry over:
// the two maps, and a drawing whose object keeps its value behind a pointer.
func (s *exampleSeed) build() Example {
	task := Example{
		ID:            "gaps-posts-test",
		Topic:         "counting.gaps",
		GradeLevel:    Level12,
		Difficulty:    3,
		Question:      s.Question,
		Options:       map[string]string{},
		CorrectAnswer: "B",
		Solution:      s.Question,
		Distractors:   map[string]Distractor{},
	}
	for i, letter := range solver.Letters() {
		task.Options[letter] = s.Options[i]
		if letter == task.CorrectAnswer {
			continue
		}
		task.Distractors[letter] = Distractor{Trap: s.Traps[i], Text: s.Texts[i]}
	}
	if s.Drawn {
		value := s.Value
		task.Drawing = "A--B"
		task.DrawingStructure = &DrawingStructure{
			Kind:      s.Kind,
			Objects:   []DrawingObject{{ID: "A", Label: s.Label, Value: &value}},
			Relations: []DrawingRelation{{Type: s.Relation, From: "A", To: "B"}},
		}
	}
	return task
}

// genExampleSeed produces the raw material of a task. The slices are as long as
// there are options, because build reads one entry per letter, and the seed is
// handed over as a pointer because a property is given what the generator makes.
func genExampleSeed() gopter.Gen {
	return gen.Struct(reflect.TypeOf(exampleSeed{}), map[string]gopter.Gen{
		"Question": gen.AnyString(),
		"Options":  gen.SliceOfN(solver.Count, gen.AnyString()),
		"Traps":    gen.SliceOfN(solver.Count, gen.AnyString()),
		"Texts":    gen.SliceOfN(solver.Count, gen.AnyString()),
		"Kind":     gen.AnyString(),
		"Label":    gen.AnyString(),
		"Relation": gen.AnyString(),
		"Value":    gen.Int(),
		"Drawn":    gen.Bool(),
	}).Map(func(seed exampleSeed) *exampleSeed { return &seed })
}

// frame assembles a drawing frame from the same raw material: topics in a list
// and a structure whose object keeps its value behind a pointer, which is
// everything a copy of a frame has to carry over.
func (s *exampleSeed) frame() Frame {
	value := s.Value
	return Frame{
		Name:    "posts",
		Purpose: s.Question,
		Topics:  []string{s.Kind, s.Label},
		Drawing: "A--B",
		Structure: &DrawingStructure{
			Kind:      s.Kind,
			Objects:   []DrawingObject{{ID: "A", Label: s.Label, Value: &value}},
			Relations: []DrawingRelation{{Type: s.Relation, From: "A", To: "B"}},
		},
	}
}

// editFrame rewrites every part of a frame that a caller could reach through a
// reference rather than through a copy.
func editFrame(frame *Frame) {
	frame.Topics[0] = "edited by a caller"
	frame.Structure.Kind = "edited by a caller"
	frame.Structure.Objects[0].Label = "edited by a caller"
	*frame.Structure.Objects[0].Value = 999
	frame.Structure.Relations[0].Type = "edited by a caller"
}

// editEverythingMutable rewrites every part of a task that a caller could reach
// through a reference rather than through a copy.
func editEverythingMutable(task *Example) {
	task.Options["A"] = "edited by a caller"
	task.Distractors["C"] = Distractor{Trap: "off_by_one", Text: "edited by a caller"}
	if task.DrawingStructure == nil {
		return
	}
	task.DrawingStructure.Kind = "edited by a caller"
	task.DrawingStructure.Objects[0].Label = "edited by a caller"
	*task.DrawingStructure.Objects[0].Value = 999
	task.DrawingStructure.Relations[0].Type = "edited by a caller"
}

func TestCopiesHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("a copy of a task is equal to the task", prop.ForAll(
		func(seed *exampleSeed) bool {
			task := seed.build()
			return reflect.DeepEqual(task.clone(), task)
		},
		genExampleSeed(),
	))

	// The witness is built from the same seed rather than copied from the
	// original, so the property does not lean on the copy it is testing.
	properties.Property("editing a copy leaves the task it came from untouched", prop.ForAll(
		func(seed *exampleSeed) bool {
			task, witness := seed.build(), seed.build()
			copied := task.clone()
			editEverythingMutable(&copied)
			return reflect.DeepEqual(task, witness)
		},
		genExampleSeed(),
	))

	properties.Property("a copy of a frame is equal to the frame", prop.ForAll(
		func(seed *exampleSeed) bool {
			frame := seed.frame()
			return reflect.DeepEqual(frame.clone(), frame)
		},
		genExampleSeed(),
	))

	properties.Property("editing a copy of a frame leaves the frame it came from untouched", prop.ForAll(
		func(seed *exampleSeed) bool {
			frame, witness := seed.frame(), seed.frame()
			copied := frame.clone()
			editFrame(&copied)
			return reflect.DeepEqual(frame, witness)
		},
		genExampleSeed(),
	))

	properties.Property("editing the levels of a copied topic leaves the topic untouched", prop.ForAll(
		func(id string, levels []string) bool {
			if len(levels) == 0 {
				return true
			}
			topic := Topic{ID: id, GradeLevels: slices.Clone(levels)}
			copied := topic.clone()
			copied.GradeLevels[0] = "edited by a caller"
			return reflect.DeepEqual(topic.GradeLevels, levels)
		},
		gen.AnyString(),
		gen.SliceOf(gen.AnyString()),
	))

	properties.TestingRun(t)
}
