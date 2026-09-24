package content

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// PackageBudget is the most a package may weigh as the model receives it. It
// is paid out of the family's own message allowance, a few thousand tokens of
// the chat's context each time, so it is kept small.
const PackageBudget = 16 * 1024

// examplesPerPackage is how many reference tasks a package shows.
const examplesPerPackage = 3

// guideName is the page every package carries: how a task is written, handed
// in and checked.
const guideName = "task_writing.md"

// Request is what a package is built from: the brief the task is to be
// written to, the corridor its difficulty came from and the child it is for.
type Request struct {
	// Language is the language the task is to be written in.
	Language string
	// Brief is what the task is to be.
	Brief profile.Brief
	// Corridor is where the child's chances lie on the brief's topic.
	Corridor rating.Corridor
	// Grade is the child's school year.
	Grade int
	// Interests are what the child likes: the settings tasks are dressed in.
	Interests []string
	// Notes are what the parent wrote about the child.
	Notes string
	// Answers is how many answers the child has given. It rotates the
	// reference tasks a package shows, so that a child asking for the same
	// topic twice does not get the same examples twice.
	Answers int
}

// Package is everything the model is handed to write one task from, as the
// JSON it receives: the brief, the corridor, the topic, every trap, what the
// task may not use, the child, three reference tasks, the limits it is held
// to, the page on how to write it and the version of that page.
//
// There is no pseudonym in it, because the request has none to give: a task
// has no use for the child's name, and the package is the one place it is easy
// to leave out. The parent's notes travel as a string among the rest, which
// the guide introduces as information about the child and not as
// instructions; being a string of JSON, they cannot close themselves off and
// speak as anything else.
//
// A package heavier than the budget shows two reference tasks instead of
// three. One still heavier after that is sent as it is: what is left is what
// a task cannot be written without — the brief, the catalogs, the guide and
// two reference tasks — and only a parent's own words, at the limits of the
// profile and in a script of wide characters, carry a package that far.
//
// A request naming a topic or a skill the catalogs do not have makes no
// package: its description would be empty, and no task written to it could
// pass the checks.
func (c *Content) Package(request *Request) ([]byte, error) {
	contents, err := c.contentsFor(request)
	if err != nil {
		return nil, err
	}
	encoded, err := encode(&contents)
	if err != nil {
		return nil, err
	}
	if len(encoded) > PackageBudget && len(contents.Examples) == examplesPerPackage {
		contents.Examples = contents.Examples[:examplesPerPackage-1]
		return encode(&contents)
	}
	return encoded, nil
}

// packageContents is a package as it is encoded, part by part. It borrows
// what it holds — the catalogs, the reference tasks, the request's own lists —
// rather than copying it: it is encoded as soon as it is gathered, nothing
// writes to it, and only its encoding leaves.
type packageContents struct {
	Language            string           `json:"language"`
	Brief               profile.Brief    `json:"brief"`
	Corridor            packageCorridor  `json:"corridor"`
	Topic               packageTopic     `json:"topic"`
	Traps               []Trap           `json:"traps"`
	Prohibitions        []Skill          `json:"prohibitions"`
	Child               packageChild     `json:"child"`
	Examples            []packageExample `json:"examples"`
	Limits              packageLimits    `json:"limits"`
	Guide               string           `json:"guide"`
	InstructionsVersion string           `json:"instructions_version"`
}

// packageCorridor is the difficulty the rule recommends for the child on this
// topic, and the chances behind it.
type packageCorridor struct {
	RecommendedDifficulty int                `json:"recommended_difficulty"`
	Fit                   rating.Fit         `json:"fit"`
	BetaMin               float64            `json:"beta_min"`
	BetaMax               float64            `json:"beta_max"`
	SuccessChance         map[string]float64 `json:"success_chance_by_difficulty"`
}

// packageTopic is the topic of the brief, as the catalog describes it.
type packageTopic struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// packageChild is what the task may be pitched at: the school year, the
// interests and the parent's notes.
type packageChild struct {
	Grade     int      `json:"grade"`
	Interests []string `json:"interests"`
	Notes     string   `json:"notes"`
}

// packageExample is a reference task as the model is shown it: without its
// id, its topic and level, which the package says already, and its solver,
// which the model is not shown.
type packageExample struct {
	Difficulty       int                   `json:"difficulty"`
	Question         string                `json:"question"`
	Drawing          string                `json:"drawing,omitempty"`
	DrawingStructure *DrawingStructure     `json:"drawing_structure,omitempty"`
	Options          map[string]string     `json:"options"`
	CorrectAnswer    string                `json:"correct_answer"`
	Hint             string                `json:"hint,omitempty"`
	Solution         string                `json:"solution"`
	Distractors      map[string]Distractor `json:"distractors"`
}

// packageLimits are what the task is held to when it is handed in.
type packageLimits struct {
	SentenceWords      int           `json:"sentence_words"`
	SentenceCharacters int           `json:"sentence_characters"`
	FleschKincaidGrade int           `json:"flesch_kincaid_grade"`
	Drawing            drawingLimits `json:"drawing"`
}

// drawingLimits are how large a drawing may be.
type drawingLimits struct {
	Width    int `json:"width"`
	Height   int `json:"height"`
	SpaceRun int `json:"space_run"`
}

// contentsFor gathers every part of a package for one request.
func (c *Content) contentsFor(request *Request) (packageContents, error) {
	topic, known := c.topicByID[request.Brief.TargetConcept]
	if !known {
		return packageContents{}, fmt.Errorf("content: build the package: no topic %q in the catalog",
			request.Brief.TargetConcept)
	}
	prohibitions, err := c.prohibitionsFor(request.Brief.ExcludedSkills)
	if err != nil {
		return packageContents{}, err
	}
	readable, drawn := checks.ReadabilityLimitsFor(request.Grade), checks.DefaultDrawingLimits()

	contents := packageContents{
		Language:     request.Language,
		Brief:        request.Brief,
		Corridor:     corridorOf(&request.Corridor),
		Topic:        packageTopic{ID: topic.ID, Name: topic.Name, Description: topic.Description},
		Traps:        c.traps,
		Prohibitions: prohibitions,
		Child:        packageChild{Grade: request.Grade, Interests: request.Interests, Notes: request.Notes},
		Limits: packageLimits{
			SentenceWords:      readable.SentenceWords,
			SentenceCharacters: readable.SentenceCharacters,
			FleschKincaidGrade: readable.FleschKincaid,
			Drawing:            drawingLimits{Width: drawn.Width, Height: drawn.Height, SpaceRun: drawn.SpaceRun},
		},
		Guide:               c.instructions[guideName],
		InstructionsVersion: c.instructionsVersion,
	}
	examples := c.examplesFor(request.Brief.TargetConcept, request.Grade, request.Brief.Difficulty, request.Answers)
	for i := range examples {
		task := &examples[i]
		contents.Examples = append(contents.Examples, packageExample{
			Difficulty: task.Difficulty, Question: task.Question, Drawing: task.Drawing,
			DrawingStructure: task.DrawingStructure, Options: task.Options, CorrectAnswer: task.CorrectAnswer,
			Hint: task.Hint, Solution: task.Solution, Distractors: task.Distractors,
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

// corridorOf is the corridor as a package states it, its chances to two
// places: the model reads which way to lean, not the fourth decimal.
func corridorOf(corridor *rating.Corridor) packageCorridor {
	chances := make(map[string]float64, len(corridor.Probabilities))
	for i, chance := range corridor.Probabilities {
		chances[strconv.Itoa(i+1)] = math.Round(chance*100) / 100
	}
	return packageCorridor{
		RecommendedDifficulty: corridor.Recommended, Fit: corridor.Fit,
		BetaMin: math.Round(corridor.BetaMin*100) / 100, BetaMax: math.Round(corridor.BetaMax*100) / 100,
		SuccessChance: chances,
	}
}

// examplesFor are the reference tasks a package shows for a topic at a child's
// grade: those of the requested difficulty first, then of the nearest, then of
// the next nearest, taken from the child's level — or, where the topic has
// nothing at that level, from the level below. Within one difficulty the
// tasks take turns by the child's answer count.
func (c *Content) examplesFor(topic string, grade, difficulty, answers int) []Example {
	for level, known := LevelOf(grade); known; level, known = levelBelow(level) {
		var pool []Example
		for i := range c.examples {
			if c.examples[i].Topic == topic && c.examples[i].GradeLevel == level {
				pool = append(pool, c.examples[i])
			}
		}
		if len(pool) > 0 {
			return nearest(pool, difficulty, answers)
		}
	}
	return nil
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
func levelBelow(level string) (string, bool) {
	levels := Levels()
	if at := slices.Index(levels, level); at > 0 {
		return levels[at-1], true
	}
	return "", false
}

// encode is a package as the model receives it. Text keeps its own characters:
// an arrow in a drawing or a quote in a question reads as itself, not as an
// escape.
func encode(contents *packageContents) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(contents); err != nil {
		return nil, fmt.Errorf("content: encode the package: %w", err)
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}
