package content

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// PackageBudget is the most a package may weigh as the model receives it. It is
// a ceiling against a package growing unnoticed, not a target to fill: the
// heaviest package a profile allows — a child at every limit, written in the
// characters JSON takes the most bytes to carry — is about half of it. A
// test holds every such package to it, so nothing is ever left out of a
// package to make it fit.
const PackageBudget = 64 * 1024

// examplesPerPackage is how many reference tasks a package shows.
const examplesPerPackage = 3

// ideasPerList is how many ideas of a topic its list holds at each level the
// topic is taught at, and so how many tasks of the topic go by before an idea
// comes round again.
const ideasPerList = 10

// guideName is the page every package carries: how a task is written, handed
// in and checked.
const guideName = "task_writing.md"

// Request is what a package is built from: the brief the task is to be
// written to, the corridor its level and difficulty came from and the child it
// is for.
type Request struct {
	// Language is the language the task is to be written in.
	Language string
	// Brief is what the task is to be. Its level decides the reference tasks
	// the model is shown and the limits the task is held to.
	Brief profile.Brief
	// Corridor is where the child's chances lie on the brief's topic.
	Corridor rating.Corridor
	// Grade is the child's school year. The model is told it as the child's
	// age, for the words and the plot, and it decides nothing else: the task
	// is of the level the brief names, whatever the grade.
	Grade int
	// Interests are what the child likes: the settings tasks are dressed in.
	Interests []string
	// Notes are what the parent wrote about the child.
	Notes string
	// Answers is how many answers the child has given. It rotates the
	// reference tasks a package shows, so that a child asking for the same
	// topic twice does not get the same examples twice.
	Answers int
	// TopicTasks is how many tasks of the brief's topic the child has left
	// behind, answered or skipped. It picks the idea of the topic the task is
	// built on, so that tasks of one topic go through its ideas one after
	// another instead of coming back to the same one.
	TopicTasks int
}

// Package is everything the model is handed to write one task from, as the
// JSON it receives: the brief, the idea of the topic the task is built on, the
// corridor, the topic, every trap, what the task may not use, the child, three
// reference tasks, the solver templates and the examples of the pictures of
// the topic, the limits it is held to and the page on how to write it. The version of
// what it is told is not among them: the model has no use for it, and the
// service, which records it with every task, knows it already.
//
// There is no pseudonym in it, because the request has none to give: a task
// has no use for the child's name, and the package is the one place it is easy
// to leave out. The parent's notes travel as a string among the rest, which
// the guide introduces as information about the child and not as
// instructions; being a string of JSON, they cannot close themselves off and
// speak as anything else.
//
// A request naming a topic or a skill the catalogs do not have, or a level the
// topic is not taught at, makes no package: its description would be empty,
// and no task written to it could pass the checks.
func (c *Content) Package(request *Request) ([]byte, error) {
	contents, err := c.contentsFor(request)
	if err != nil {
		return nil, err
	}
	return encode(&contents)
}

// packageContents is a package as it is encoded, part by part. It borrows
// what it holds — the catalogs, the reference tasks, the request's own lists —
// rather than copying it: it is encoded as soon as it is gathered, nothing
// writes to it, and only its encoding leaves.
type packageContents struct {
	Language     string           `json:"language"`
	Brief        profile.Brief    `json:"brief"`
	Idea         packageIdea      `json:"idea"`
	Corridor     packageCorridor  `json:"corridor"`
	Topic        packageTopic     `json:"topic"`
	Traps        []packageTrap    `json:"traps"`
	Prohibitions []Skill          `json:"prohibitions"`
	Child        packageChild     `json:"child"`
	Examples     []packageExample `json:"examples"`
	Templates    []string         `json:"solver_templates"`
	Pictures     []packagePicture `json:"pictures"`
	Limits       packageLimits    `json:"limits"`
	Guide        string           `json:"guide"`
}

// packageIdea is the idea of the topic the task is built on: Text says it, and
// Number is its place on the topic's list at the brief's level. Round is which
// pass of the list the number is on for this child, the first being 1: from
// the second, the task is a variant of its idea, so that an idea that comes
// round again is not the same task again.
type packageIdea struct {
	Number int    `json:"number"`
	Round  int    `json:"round"`
	Text   string `json:"text"`
}

// ideaOf is the idea on a topic's list the next task is built on, after the
// tasks of the topic the child has left behind: the next on the list each
// time, and round again after the last. A model in a new chat sees none of the
// tasks written before, and left to itself reaches for the same favourite idea
// every time; an idea named for it, a step further with every task, is what
// moves it along the topic.
func ideaOf(list []string, tasks int) packageIdea {
	tasks = max(tasks, 0) // a count read from a file is not trusted to be positive
	place := tasks % len(list)
	return packageIdea{Number: place + 1, Round: tasks/len(list) + 1, Text: list[place]}
}

// packageCorridor is the point of the ladder the rule recommends for the child
// on this topic, and the chances behind it: the chance of a correct answer at
// every point the topic can be set at, so that a model stepping away from the
// recommendation sees how far it steps.
type packageCorridor struct {
	RecommendedGradeLevel rating.GradeLevel `json:"recommended_grade_level"`
	RecommendedDifficulty int               `json:"recommended_difficulty"`
	Fit                   rating.Fit        `json:"fit"`
	BetaMin               float64           `json:"beta_min"`
	BetaMax               float64           `json:"beta_max"`
	SuccessChances        []packageChance   `json:"success_chances"`
}

// packageChance is the chance of a correct answer at one point of the topic.
type packageChance struct {
	GradeLevel rating.GradeLevel `json:"grade_level"`
	Difficulty int               `json:"difficulty"`
	Chance     float64           `json:"chance"`
}

// packageTrap is a trap of the catalog as the model writing a task is shown
// it: its id and what the mistake is. What an adult can do about it is not for
// writing a task, and would weigh on every package.
type packageTrap struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// packageTopic is the topic of the brief, as the catalog describes it.
type packageTopic struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// packageChild is what the wording may be pitched at: the school year, as the
// child's age, the interests and the parent's notes.
type packageChild struct {
	Grade     int      `json:"grade"`
	Interests []string `json:"interests"`
	Notes     string   `json:"notes"`
}

// packageExample is a reference task as the model is shown it: without its
// id, its topic and level, which the package says already, and its solver,
// which the model is not shown.
type packageExample struct {
	Difficulty    int                   `json:"difficulty"`
	Question      string                `json:"question"`
	Picture       json.RawMessage       `json:"picture,omitempty"`
	Options       map[string]string     `json:"options"`
	CorrectAnswer string                `json:"correct_answer"`
	Hint          string                `json:"hint,omitempty"`
	Solution      string                `json:"solution"`
	Distractors   map[string]Distractor `json:"distractors"`
}

// packageLimits are what the task is held to when it is handed in: how long a
// sentence may be, how long a label and a note of its picture, and how long
// the total under the picture of its solution.
type packageLimits struct {
	SentenceWords      int `json:"sentence_words"`
	SentenceCharacters int `json:"sentence_characters"`
	FleschKincaidGrade int `json:"flesch_kincaid_grade"`
	LabelCharacters    int `json:"label_characters"`
	NoteCharacters     int `json:"note_characters"`
	TotalCharacters    int `json:"total_characters"`
}

// contentsFor gathers every part of a package for one request.
func (c *Content) contentsFor(request *Request) (packageContents, error) {
	topic, known := c.topicByID[request.Brief.TargetConcept]
	if !known {
		return packageContents{}, fmt.Errorf("content: build the package: no topic %q in the catalog",
			request.Brief.TargetConcept)
	}
	level := request.Brief.GradeLevel
	if !topic.HasLevel(level) {
		return packageContents{}, fmt.Errorf("content: build the package: %s is not taught at level %q", topic.ID, level)
	}
	prohibitions, err := c.prohibitionsFor(request.Brief.ExcludedSkills)
	if err != nil {
		return packageContents{}, err
	}
	readable := checks.ReadabilityLimitsFor(level)

	contents := packageContents{
		Language:     request.Language,
		Brief:        request.Brief,
		Idea:         ideaOf(c.ideas[topic.ID][level], request.TopicTasks),
		Corridor:     corridorOf(&request.Corridor),
		Topic:        packageTopic{ID: topic.ID, Name: topic.Name, Description: topic.Description},
		Traps:        packageTraps(c.traps),
		Prohibitions: prohibitions,
		Child:        packageChild{Grade: request.Grade, Interests: request.Interests, Notes: request.Notes},
		Limits: packageLimits{
			SentenceWords:      readable.SentenceWords,
			SentenceCharacters: readable.SentenceCharacters,
			FleschKincaidGrade: readable.FleschKincaid,
			LabelCharacters:    picture.MaxLabelCharacters,
			NoteCharacters:     picture.MaxNoteCharacters,
			TotalCharacters:    picture.MaxTotalCharacters,
		},
		Templates: c.templatePrograms(topic.ID),
		Pictures:  c.picturesFor(topic.ID),
		Guide:     c.instructions[guideName],
	}
	examples := c.examplesFor(request.Brief.TargetConcept, level, request.Brief.Difficulty, request.Answers)
	contents.Examples = make([]packageExample, 0, len(examples)) // a topic with none is shown an empty list, not null
	for i := range examples {
		task := &examples[i]
		contents.Examples = append(contents.Examples, packageExample{
			Difficulty: task.Difficulty, Question: task.Question, Picture: task.Picture, Options: task.Options,
			CorrectAnswer: task.CorrectAnswer, Hint: task.Hint, Solution: task.Solution, Distractors: task.Distractors,
		})
	}
	return contents, nil
}

// prohibitionsFor are the skills a task may not use, as the catalog describes
// them. The list is never null: a child with none has an empty one.
func (c *Content) prohibitionsFor(ids []string) ([]Skill, error) {
	prohibitions := make([]Skill, 0, len(ids))
	for _, id := range ids {
		skill, known := c.skillByID[id]
		if !known {
			return nil, fmt.Errorf("content: build the package: no skill %q in the catalog", id)
		}
		prohibitions = append(prohibitions, skill)
	}
	return prohibitions, nil
}

// corridorOf is the corridor as a package states it, its numbers to two
// places: the model reads which way to lean, not the fourth decimal. The
// chances are listed the easiest point first, and the list is never null.
func corridorOf(corridor *rating.Corridor) packageCorridor {
	chances := make([]packageChance, 0, len(corridor.Chances))
	for _, chance := range corridor.Chances {
		chances = append(chances, packageChance{
			GradeLevel: chance.GradeLevel, Difficulty: chance.Difficulty, Chance: twoPlaces(chance.Probability),
		})
	}
	return packageCorridor{
		RecommendedGradeLevel: corridor.Recommended.GradeLevel,
		RecommendedDifficulty: corridor.Recommended.Difficulty,
		Fit:                   corridor.Fit,
		BetaMin:               twoPlaces(corridor.BetaMin),
		BetaMax:               twoPlaces(corridor.BetaMax),
		SuccessChances:        chances,
	}
}

// twoPlaces is a number rounded to two decimal places.
func twoPlaces(x float64) float64 { return math.Round(x*100) / 100 }

// examplesFor are the reference tasks a package shows for a topic at a level:
// those of the requested difficulty first, then of the nearest, then of the
// next nearest, taken from that level — or, where the topic has nothing at
// that level, from the level below — with one that draws among them wherever
// the level has one. Within one difficulty the tasks take turns by the child's
// answer count.
func (c *Content) examplesFor(topic string, level rating.GradeLevel, difficulty, answers int) []Example {
	for known := level.Known(); known; level, known = levelBelow(level) {
		var pool []Example
		for i := range c.examples {
			if c.examples[i].Topic == topic && c.examples[i].GradeLevel == level {
				pool = append(pool, c.examples[i])
			}
		}
		if len(pool) > 0 {
			return withADrawnTask(nearest(pool, difficulty, answers), pool, difficulty, answers)
		}
	}
	return nil
}

// withADrawnTask is the tasks picked for a package with one that draws among
// them, where the pool has one: a model shown no picture describes none, even
// in a topic with examples of pictures. When none of those picked draws, the
// last of them gives way to the task that draws nearest the difficulty, and
// tasks that draw at one difficulty take turns by the answer count, as all
// tasks of a difficulty do.
func withADrawnTask(picked, pool []Example, difficulty, answers int) []Example {
	if len(picked) == 0 || anyDraws(picked) {
		return picked
	}
	var drawn []Example
	for i := range pool {
		if pool[i].draws() {
			drawn = append(drawn, pool[i])
		}
	}
	if len(drawn) == 0 {
		return picked
	}
	shown := slices.Clone(picked)
	shown[len(shown)-1] = nearest(drawn, difficulty, answers)[0]
	return shown
}

// anyDraws says whether any of some reference tasks carries a picture.
func anyDraws(tasks []Example) bool {
	for i := range tasks {
		if tasks[i].draws() {
			return true
		}
	}
	return false
}

// nearest are the three tasks of a pool nearest to a difficulty. Tasks of one
// difficulty are rotated by the answer count before they are taken, so that
// over any run of answers each of them is shown as often as the others; of two
// difficulties equally near, the easier comes first.
func nearest(pool []Example, difficulty, answers int) []Example {
	byDifficulty := map[int][]Example{}
	for i := range pool {
		byDifficulty[pool[i].Difficulty] = append(byDifficulty[pool[i].Difficulty], pool[i])
	}
	difficulties := slices.Collect(maps.Keys(byDifficulty))
	slices.SortFunc(difficulties, func(a, b int) int {
		return cmp.Or(cmp.Compare(distance(a, difficulty), distance(b, difficulty)), cmp.Compare(a, b))
	})

	var picked []Example
	for _, at := range difficulties {
		tasks := byDifficulty[at]
		slices.SortFunc(tasks, func(a, b Example) int { return cmp.Compare(a.ID, b.ID) })
		start := (answers%len(tasks) + len(tasks)) % len(tasks) // a count read from a file is not trusted to be positive
		picked = append(picked, slices.Concat(tasks[start:], tasks[:start])...)
		if len(picked) >= examplesPerPackage {
			return picked[:examplesPerPackage]
		}
	}
	return picked
}

// distance is how far apart two difficulties are.
func distance(a, b int) int { return max(a-b, b-a) }

// levelBelow is the level before this one, if there is one.
func levelBelow(level rating.GradeLevel) (rating.GradeLevel, bool) {
	levels := rating.GradeLevels()
	if at := slices.Index(levels, level); at > 0 {
		return levels[at-1], true
	}
	return "", false
}

// encode is a package as the model receives it. Text keeps its own characters:
// a sign in a picture's note or a quote in a question reads as itself, not as
// an escape.
func encode(contents *packageContents) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(contents); err != nil {
		return nil, fmt.Errorf("content: encode the package: %w", err)
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

// packageTraps are the traps of the catalog as a package shows them, in
// catalog order.
func packageTraps(traps []Trap) []packageTrap {
	shown := make([]packageTrap, 0, len(traps))
	for _, trap := range traps {
		shown = append(shown, packageTrap{ID: trap.ID, Description: trap.Description})
	}
	return shown
}
