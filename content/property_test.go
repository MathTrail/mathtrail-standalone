package content

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The properties here name what has to hold for every input rather than for the
// handful of inputs somebody thought of. Four things in this package have that
// shape: the version of the instructions, which must depend on the content and
// on nothing else; the copies handed out of the content, which must share no
// memory with it; the catalog of traps, which is read only when every trap has
// its advice; and the ideas of a topic, a list of which is read only when it
// holds none twice, and which a package must name the whole list of before it
// comes back.

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
	Value    int
	Pictured bool
}

// picture is a description made of the seed's raw material: it need not be a
// picture the format accepts, since a copy carries over whatever it holds.
func (s *exampleSeed) picture() json.RawMessage {
	raw, err := json.Marshal(map[string]any{"kind": s.Kind, "label": s.Label, "value": s.Value})
	if err != nil {
		panic(err)
	}
	return raw
}

// build assembles a reference task holding everything a copy has to carry over:
// the two maps, and a picture held as the bytes it was written in.
func (s *exampleSeed) build() Example {
	task := Example{
		ID:            "gaps-posts-test",
		Topic:         "counting.gaps",
		GradeLevel:    rating.Grades12,
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
	if s.Pictured {
		task.Picture = s.picture()
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
		"Value":    gen.Int(),
		"Pictured": gen.Bool(),
	}).Map(func(seed exampleSeed) *exampleSeed { return &seed })
}

// example assembles an example of a kind of picture from the same raw
// material: topics in a list and a picture held as bytes, which is everything
// a copy of an example has to carry over.
func (s *exampleSeed) example() PictureExample {
	return PictureExample{
		Kind: picture.Kind(s.Kind), Purpose: s.Question, Topics: []string{s.Kind, s.Label}, Picture: s.picture(),
	}
}

// editExample rewrites every part of an example of a kind that a caller could
// reach through a reference rather than through a copy.
func editExample(example *PictureExample) {
	example.Topics[0] = "edited by a caller"
	example.Picture[0] = 'X'
}

// editEverythingMutable rewrites every part of a task that a caller could reach
// through a reference rather than through a copy.
func editEverythingMutable(task *Example) {
	task.Options["A"] = "edited by a caller"
	task.Distractors["C"] = Distractor{Trap: "off_by_one", Text: "edited by a caller"}
	if len(task.Picture) > 0 {
		task.Picture[0] = 'X'
	}
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

	properties.Property("a copy of an example of a kind is equal to the example", prop.ForAll(
		func(seed *exampleSeed) bool {
			example := seed.example()
			return reflect.DeepEqual(example.clone(), example)
		},
		genExampleSeed(),
	))

	properties.Property("editing a copy of an example of a kind leaves the example it came from untouched",
		prop.ForAll(
			func(seed *exampleSeed) bool {
				example, witness := seed.example(), seed.example()
				copied := example.clone()
				editExample(&copied)
				return reflect.DeepEqual(example, witness)
			},
			genExampleSeed(),
		))

	properties.Property("editing the levels or the bases of a copied topic leaves the topic untouched", prop.ForAll(
		func(id string, levels []rating.GradeLevel, bases []string) bool {
			topic := Topic{ID: id, GradeLevels: slices.Clone(levels), BuildsOn: slices.Clone(bases)}
			witness := Topic{ID: id, GradeLevels: slices.Clone(levels), BuildsOn: slices.Clone(bases)}
			copied := topic.clone()
			if len(copied.GradeLevels) > 0 {
				copied.GradeLevels[0] = "edited by a caller"
			}
			if len(copied.BuildsOn) > 0 {
				copied.BuildsOn[0] = "edited.by.a.caller"
			}
			return reflect.DeepEqual(topic, witness)
		},
		gen.AnyString(),
		gen.SliceOf(gen.AnyString().Map(func(level string) rating.GradeLevel { return rating.GradeLevel(level) })),
		gen.SliceOf(gen.AnyString()),
	))

	properties.Property("a copy of a topic is equal to the topic", prop.ForAll(
		func(id, slug string, levels []rating.GradeLevel, bases []string, sitePage bool) bool {
			topic := Topic{ID: id, GradeLevels: levels, Slug: slug, BuildsOn: bases, SitePage: sitePage}
			return reflect.DeepEqual(topic.clone(), topic)
		},
		gen.AnyString(),
		gen.AlphaString(),
		gen.SliceOf(gen.AnyString().Map(func(level string) rating.GradeLevel { return rating.GradeLevel(level) })),
		gen.SliceOf(gen.AnyString()),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// A trap catalog reads exactly when every trap in it carries its advice: a
// trap without one would leave the review of the progress with nothing to tell
// an adult about the mistake, and the service with a gap it found too late.
func TestATrapCatalogReadsOnlyWithEveryTrapsAdvice(t *testing.T) {
	t.Parallel()

	traps, err := loadTraps(files)
	if err != nil {
		t.Fatalf("loadTraps() error = %v, want the shipped catalog", err)
	}
	properties := gopter.NewProperties(nil)
	properties.Property("read exactly when every trap advises", prop.ForAll(
		func(advised []bool) bool {
			edited := slices.Clone(traps)
			every := true
			for i := range edited {
				if !advised[i] {
					edited[i].Advice = ""
					every = false
				}
			}
			raw, err := json.Marshal(edited)
			if err != nil {
				return false
			}
			_, err = loadTraps(fstest.MapFS{trapsFile: {Data: raw}})
			return (err == nil) == every
		},
		gen.SliceOfN(len(traps), gen.Bool()),
	))
	properties.TestingRun(t)
}

// The idea a package names is what moves a model with no memory of the tasks
// before along a topic's ideas. Whatever the count it is given — a file can
// hold any — it names an idea of the list by its place on it, the tasks of one
// round come to every idea once, and the next round comes to them again in the
// same order.
func TestTheIdeaHoldsItsProperties(t *testing.T) {
	t.Parallel()

	list := sampleIdeas(rating.Grades12)
	properties := gopter.NewProperties(nil)
	properties.Property("the idea named is the one at its place on the list", prop.ForAll(
		func(tasks int) bool {
			idea := ideaOf(list, tasks)
			return idea.Number >= 1 && idea.Number <= len(list) && idea.Text == list[idea.Number-1] && idea.Round >= 1
		},
		gen.Int(),
	))
	properties.Property("a list's worth of tasks in a row comes to every idea once", prop.ForAll(
		func(start int) bool {
			met := map[string]bool{}
			for tasks := start; tasks < start+ideasPerList; tasks++ {
				met[ideaOf(list, tasks).Text] = true
			}
			return len(met) == ideasPerList
		},
		gen.IntRange(0, 1<<30),
	))
	properties.Property("a list's worth of tasks later comes the same idea, a round on", prop.ForAll(
		func(tasks int) bool {
			now, later := ideaOf(list, tasks), ideaOf(list, tasks+ideasPerList)
			return later.Number == now.Number && later.Text == now.Text && later.Round == now.Round+1
		},
		gen.IntRange(0, 1<<30),
	))
	properties.TestingRun(t)
}

// A list is what a package names an idea from, so it is taken only when it
// holds as many ideas as a list does and none of them twice — the same idea
// in other capitals or with other spaces is said twice too.
func TestAListOfIdeasIsTakenOnlyAsIdeasEachOfItsOwn(t *testing.T) {
	t.Parallel()

	const pool = 100 // ideas a list is drawn from, so that most lists hold none twice
	// spelling writes an idea of the pool in one of three ways a reader takes
	// for the same idea.
	spelling := func(pick int) string {
		idea := fmt.Sprintf("Idea number %d", pick/3)
		switch pick % 3 {
		case 1:
			return strings.ToUpper(idea)
		case 2:
			return "  " + strings.ReplaceAll(idea, " ", "   ") + " "
		}
		return idea
	}
	properties := gopter.NewProperties(nil)
	properties.Property("a list is taken when it holds a list's worth of ideas and none twice", prop.ForAll(
		func(picks []int, length int) bool {
			list, ideas := make([]string, length), map[int]bool{}
			for i, pick := range picks[:length] {
				list[i], ideas[pick/3] = spelling(pick), true
			}
			p := &problems{file: ideasDir}
			checkIdeas(p, rating.Grades12, list)
			return (p.err() == nil) == (length == ideasPerList && len(ideas) == length)
		},
		gen.SliceOfN(ideasPerList+1, gen.IntRange(0, 3*pool-1)),
		gen.IntRange(ideasPerList-1, ideasPerList+1),
	))
	properties.TestingRun(t)
}
