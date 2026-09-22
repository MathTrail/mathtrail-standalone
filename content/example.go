package content

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// Example is one reference task. Reference tasks are what the model is shown
// before it writes its own: they carry the shape, the style and above all the
// way a trap is attached to every wrong option. They are in English whatever
// the language of the chat, because what they carry is the structure and not
// the wording.
type Example struct {
	ID         string `json:"id"`
	Topic      string `json:"topic"`
	GradeLevel string `json:"grade_level"`
	Difficulty int    `json:"difficulty"`
	Question   string `json:"question"`
	// Drawing and DrawingStructure are set only where the topic needs a
	// picture, and they are set together: a drawing with no structure cannot be
	// checked against the wording.
	Drawing          string            `json:"drawing,omitempty"`
	DrawingStructure *DrawingStructure `json:"drawing_structure,omitempty"`
	Options          map[string]string `json:"options"`
	CorrectAnswer    string            `json:"correct_answer"`
	// Hint is a nudge that does not give the answer away. The tasks ported from
	// the prototype have none; everything written since does.
	Hint        string                `json:"hint,omitempty"`
	Solution    string                `json:"solution"`
	Distractors map[string]Distractor `json:"distractors"`
	// Solver is the Starlark program that brute-forces this very task, which is
	// what makes the example re-checkable rather than merely plausible.
	Solver string `json:"solver,omitempty"`
}

// Distractor is one wrong option: the mistake it comes from, and what the child
// is told after choosing it.
type Distractor struct {
	Trap string `json:"trap"`
	Text string `json:"text"`
}

// DrawingStructure is a text drawing written out as data, so that the drawing
// and the wording can be compared without either being understood.
type DrawingStructure struct {
	Kind      string            `json:"kind"`
	Objects   []DrawingObject   `json:"objects"`
	Relations []DrawingRelation `json:"relations,omitempty"`
}

// DrawingObject is one thing the drawing shows: its id, the label exactly as it
// appears in the picture, and a value where it has one.
type DrawingObject struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value *int   `json:"value,omitempty"`
}

// DrawingRelation is how two objects of a drawing stand to each other.
type DrawingRelation struct {
	Type string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
}

const (
	examplesDir        = "examples"
	examplesSuffix     = ".json"
	minDifficulty      = 1
	maxDifficulty      = 5
	distractorsPerTask = 4
)

// optionLetters are the five options every task offers, in the order a child
// reads them.
var optionLetters = []string{"A", "B", "C", "D", "E"}

// A reference task id is lowercase and dash-separated, as in ord-12-d1-1.
var exampleIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// loadExamples reads every file of reference tasks, one file per topic named
// after it, and refuses a task that could not be used as an example: one whose
// options a child could not tell apart, one whose trap is in no catalog, one
// offered at a level its topic is not taught at.
func loadExamples(src fs.FS, topics map[string]Topic, traps map[string]Trap) ([]Example, error) {
	entries, err := fs.ReadDir(src, examplesDir)
	if err != nil {
		return nil, fmt.Errorf("content: read %s: %w", examplesDir, err)
	}

	var (
		all    []Example
		faults []error
		seen   = make(map[string]bool)
	)
	for _, entry := range entries {
		file := examplesDir + "/" + entry.Name()
		p := &problems{file: file}

		topic, named := strings.CutSuffix(entry.Name(), examplesSuffix)
		switch {
		case entry.IsDir():
			p.addf("reference tasks live in files, one per topic, not in directories")
		case !named:
			p.addf("a file of reference tasks is named after its topic and ends in %s", examplesSuffix)
		default:
			if _, known := topics[topic]; !known {
				p.addf("%q is not a topic in the catalog, so no task here can be used", topic)
			}
		}
		if err := p.err(); err != nil {
			faults = append(faults, err)
			continue
		}

		var tasks []Example
		if err := decode(src, file, &tasks); err != nil {
			faults = append(faults, err)
			continue
		}
		check := exampleCheck{topic: topic, topics: topics, traps: traps, seen: seen}
		for i := range tasks {
			check.run(p, i, &tasks[i])
		}
		if err := p.err(); err != nil {
			faults = append(faults, err)
		}
		all = append(all, tasks...)
	}
	return all, errors.Join(faults...)
}

// exampleCheck is what one reference task is measured against: the topic of the
// file it was found in, the catalogs it may refer to, and the ids already taken
// by the tasks read before it.
type exampleCheck struct {
	topic  string
	topics map[string]Topic
	traps  map[string]Trap
	seen   map[string]bool
}

// run records every fault of one task.
func (c exampleCheck) run(p *problems, i int, task *Example) {
	where := entryName("task", i, task.ID)

	if !exampleIDPattern.MatchString(task.ID) {
		p.addf("%s: an id is lowercase words joined by dashes", where)
	} else if c.seen[task.ID] {
		p.addf("%s: the id is used twice", where)
	}
	c.seen[task.ID] = true

	c.checkPlacement(p, where, task)

	if task.Difficulty < minDifficulty || task.Difficulty > maxDifficulty {
		p.addf("%s: difficulty %d is outside %d-%d", where, task.Difficulty, minDifficulty, maxDifficulty)
	}
	if task.Question == "" {
		p.addf("%s: the question is empty", where)
	}
	if task.Solution == "" {
		p.addf("%s: the solution is empty", where)
	}

	c.checkOptions(p, where, task)
	c.checkDistractors(p, where, task)
	c.checkDrawing(p, where, task)
}

// checkPlacement checks that the task is in the file of its own topic and is
// offered at a level that topic is taught at.
func (c exampleCheck) checkPlacement(p *problems, where string, task *Example) {
	if task.Topic != c.topic {
		p.addf("%s: the topic is %q, and this file holds %q", where, task.Topic, c.topic)
		return
	}
	topic, known := c.topics[task.Topic]
	if !known {
		p.addf("%s: topic %q is not in the catalog", where, task.Topic)
		return
	}
	if !topic.HasLevel(task.GradeLevel) {
		p.addf("%s: grade level %q is not one topic %q is offered at (%v)",
			where, task.GradeLevel, topic.ID, topic.GradeLevels)
	}
}

// checkOptions checks that the child is offered five answers they can tell
// apart, with exactly one of them marked correct.
func (c exampleCheck) checkOptions(p *problems, where string, task *Example) {
	if len(task.Options) != len(optionLetters) {
		p.addf("%s: %d options, and a task offers exactly %v", where, len(task.Options), optionLetters)
	}
	texts := make(map[string]string, len(task.Options))
	for _, letter := range optionLetters {
		text, ok := task.Options[letter]
		switch {
		case !ok:
			p.addf("%s: option %s is missing", where, letter)
			continue
		case strings.TrimSpace(text) == "":
			p.addf("%s: option %s is empty", where, letter)
			continue
		}
		same := strings.ToLower(strings.TrimSpace(text))
		if first, repeated := texts[same]; repeated {
			p.addf("%s: options %s and %s read the same", where, first, letter)
		}
		texts[same] = letter
	}

	if !isOptionLetter(task.CorrectAnswer) {
		p.addf("%s: the correct answer is %q, and it is one of %v", where, task.CorrectAnswer, optionLetters)
	}
}

// checkDistractors checks that every wrong option, and only a wrong option,
// carries the mistake it comes from and the words the child is told. This is
// what turns a wrong answer into a diagnosis, so a missing entry is not a
// formality.
func (c exampleCheck) checkDistractors(p *problems, where string, task *Example) {
	if len(task.Distractors) != distractorsPerTask {
		p.addf("%s: %d explanations of wrong options, and every task has %d",
			where, len(task.Distractors), distractorsPerTask)
	}
	for _, letter := range optionLetters {
		distractor, ok := task.Distractors[letter]
		if letter == task.CorrectAnswer {
			if ok {
				p.addf("%s: option %s is the correct answer and needs no explanation", where, letter)
			}
			continue
		}
		if !ok {
			p.addf("%s: wrong option %s has no explanation", where, letter)
			continue
		}
		if _, known := c.traps[distractor.Trap]; !known {
			p.addf("%s: option %s names trap %q, which is not in the catalog", where, letter, distractor.Trap)
		}
		if strings.TrimSpace(distractor.Text) == "" {
			p.addf("%s: option %s explains nothing", where, letter)
		}
	}
}

// checkDrawing checks that a picture and its description arrive together, that
// the description is well formed, and that it holds together on its own: a
// relation names objects the same drawing draws. What the picture says about
// the wording is a matter for the checks a submitted task passes; here it only
// has to be readable as data.
func (c exampleCheck) checkDrawing(p *problems, where string, task *Example) {
	hasDrawing, hasStructure := task.Drawing != "", task.DrawingStructure != nil
	switch {
	case hasDrawing && !hasStructure:
		p.addf("%s: a drawing without its structure cannot be checked against the wording", where)
		return
	case hasStructure && !hasDrawing:
		p.addf("%s: a drawing structure describes a drawing, and there is none", where)
		return
	case !hasStructure:
		return
	}

	structure := task.DrawingStructure
	if structure.Kind == "" {
		p.addf("%s: the drawing structure does not say what is drawn", where)
	}
	if len(structure.Objects) == 0 {
		p.addf("%s: the drawing structure names nothing that is drawn", where)
	}
	ids := make(map[string]bool, len(structure.Objects))
	for i, object := range structure.Objects {
		switch {
		case object.ID == "":
			p.addf("%s: drawn object %d has no id", where, i+1)
		case ids[object.ID]:
			p.addf("%s: drawn object %q appears twice", where, object.ID)
		}
		ids[object.ID] = true
		if object.Label == "" {
			p.addf("%s: drawn object %d has no label to look for in the drawing", where, i+1)
		}
	}
	for i, relation := range structure.Relations {
		if relation.Type == "" || relation.From == "" || relation.To == "" {
			p.addf("%s: relation %d is not a type, a from and a to", where, i+1)
			continue
		}
		if !ids[relation.From] {
			p.addf("%s: relation %d starts at %q, which is not drawn", where, i+1, relation.From)
		}
		if !ids[relation.To] {
			p.addf("%s: relation %d ends at %q, which is not drawn", where, i+1, relation.To)
		}
	}
}

// isOptionLetter reports whether this names one of the five options.
func isOptionLetter(letter string) bool {
	return slices.Contains(optionLetters, letter)
}

// clone copies a reference task together with everything it points at: the
// options, the explanations of the wrong ones and the drawing. A task holds
// maps, and a map is a reference however many times the value around it is
// copied, so without this a caller could rewrite an option inside the binary's
// own content for every request that follows.
func (e *Example) clone() Example {
	copied := *e
	copied.Options = maps.Clone(e.Options)
	copied.Distractors = maps.Clone(e.Distractors)
	if e.DrawingStructure != nil {
		structure := *e.DrawingStructure
		structure.Objects = slices.Clone(structure.Objects)
		for i, object := range structure.Objects {
			if object.Value != nil {
				value := *object.Value
				structure.Objects[i].Value = &value
			}
		}
		structure.Relations = slices.Clone(structure.Relations)
		copied.DrawingStructure = &structure
	}
	return copied
}
