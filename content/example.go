package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"regexp"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// Example is one reference task. Reference tasks are what the model is shown
// before it writes its own: they carry the level, how hard a task of it is, the
// kind of reasoning the topic asks for and above all the way a trap is
// attached to every wrong option — not a task to be told again in a new story.
// They are in English whatever the language of the chat, because what they
// carry is not the wording.
type Example struct {
	ID         string            `json:"id"`
	Topic      string            `json:"topic"`
	GradeLevel rating.GradeLevel `json:"grade_level"`
	Difficulty int               `json:"difficulty"`
	Question   string            `json:"question"`
	// Picture is the description of the task's picture, which the card draws,
	// set only where the task has something to see.
	Picture       json.RawMessage   `json:"picture,omitempty"`
	Options       map[string]string `json:"options"`
	CorrectAnswer string            `json:"correct_answer"`
	// Hint is a nudge that does not give the answer away. The tasks ported from
	// the prototype have none; everything written since does.
	Hint     string `json:"hint,omitempty"`
	Solution string `json:"solution"`
	// SolutionPicture is the description of the picture of the solution, which
	// the card shows with the solution once the child has answered, and
	// SolutionTotal the equality written large under it. Every task the model
	// writes draws its solution, and each topic at each level has a reference
	// task that shows it how.
	SolutionPicture json.RawMessage       `json:"solution_picture,omitempty"`
	SolutionTotal   string                `json:"solution_total,omitempty"`
	Distractors     map[string]Distractor `json:"distractors"`
	// Solver is the Starlark program that brute-forces this very task, which is
	// what makes the example re-checkable rather than merely plausible.
	//
	// It is read from a file of its own beside the tasks rather than from this
	// one, so that a program is stored as a program: readable, diffable a line
	// at a time, and written without escaping every newline. The field is
	// therefore not part of the task's JSON — a `solver` written there is
	// ignored, which is why it is refused outright.
	Solver string `json:"-"`
}

// Distractor is one wrong option: the mistake it comes from, and what the child
// is told after choosing it.
type Distractor struct {
	Trap string `json:"trap"`
	Text string `json:"text"`
}

const (
	examplesDir        = "examples"
	examplesSuffix     = ".json"
	solversDir         = "solvers"
	solverSuffix       = ".star"
	distractorsPerTask = 4
)

// A reference task id is lowercase and dash-separated, as in ord-12-d1-1.
var exampleIDPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// loadExamples reads every file of reference tasks, one file per topic named
// after it, gives each task the solver stored beside it, and refuses a task that
// could not be used as an example: one whose options a child could not tell
// apart, one whose trap is in no catalog, one offered at a level its topic is
// not taught at.
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
		if entry.IsDir() && entry.Name() == solversDir {
			continue // the solvers of these tasks, attached once every task is read
		}

		tasks, err := loadTopicExamples(src, entry, topics, traps, seen)
		if err != nil {
			faults = append(faults, err)
		}
		all = append(all, tasks...)
	}

	if err := attachSolvers(src, all); err != nil {
		faults = append(faults, err)
	}
	return all, errors.Join(faults...)
}

// loadTopicExamples reads the reference tasks of one topic from the file named
// after it, and reports everything wrong with them at once. Tasks that were read
// are returned even when some of them are faulty, so that an id taken by a
// broken task is still an id the next file cannot take.
func loadTopicExamples(
	src fs.FS,
	entry fs.DirEntry,
	topics map[string]Topic,
	traps map[string]Trap,
	seen map[string]bool,
) ([]Example, error) {
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
		return nil, err
	}

	var tasks []Example
	if err := decode(src, file, &tasks); err != nil {
		return nil, err
	}
	check := exampleCheck{topic: topic, topics: topics, traps: traps, seen: seen}
	for i := range tasks {
		check.run(p, i, &tasks[i])
	}
	return tasks, p.err()
}

// attachSolvers gives each reference task the Starlark program that brute-forces
// it, read from the file named after the task.
//
// The two halves hold each other together from both ends. A task whose file is
// missing simply has no solver, and the bench that runs them all refuses that; a
// file belonging to no task is refused here, because nothing would ever run it
// and nobody would notice.
func attachSolvers(src fs.FS, tasks []Example) error {
	dir := examplesDir + "/" + solversDir
	entries, err := fs.ReadDir(src, dir)
	if err != nil {
		return fmt.Errorf("content: read %s: %w", dir, err)
	}

	byID := make(map[string]*Example, len(tasks))
	for i := range tasks {
		byID[tasks[i].ID] = &tasks[i]
	}

	p := &problems{file: dir}
	for _, entry := range entries {
		id, named := strings.CutSuffix(entry.Name(), solverSuffix)
		switch {
		case entry.IsDir():
			p.addf("%s: a solver is a file, not a directory", entry.Name())
			continue
		case !named:
			p.addf("%s: a solver is named after its task and ends in %s", entry.Name(), solverSuffix)
			continue
		}

		task, known := byID[id]
		if !known {
			p.addf("%s: no reference task is called %q, so this solver would never run", entry.Name(), id)
			continue
		}
		program, err := fs.ReadFile(src, dir+"/"+entry.Name())
		if err != nil {
			return fmt.Errorf("content: read %s/%s: %w", dir, entry.Name(), err)
		}
		if strings.TrimSpace(string(program)) == "" {
			p.addf("%s: the solver is empty", entry.Name())
			continue
		}
		task.Solver = string(program)
	}
	return p.err()
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

	if task.Difficulty < profile.MinDifficulty || task.Difficulty > profile.MaxDifficulty {
		p.addf("%s: difficulty %d is outside %d-%d", where, task.Difficulty, profile.MinDifficulty, profile.MaxDifficulty)
	}
	if solver.Blank(task.Question) {
		p.addf("%s: the question is empty", where)
	}
	if solver.Blank(task.Solution) {
		p.addf("%s: the solution is empty", where)
	}

	c.checkOptions(p, where, task)
	c.checkDistractors(p, where, task)
	checkPicture(p, where, "picture", task.Picture)
	checkPicture(p, where, "solution_picture", task.SolutionPicture)
	checkTotal(p, where, task)
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
	if len(task.Options) != solver.Count {
		p.addf("%s: %d options, and a task offers exactly %v", where, len(task.Options), solver.Letters())
	}
	firstSaying := make(map[string]string, solver.Count)
	for _, letter := range solver.Letters() {
		text, ok := task.Options[letter]
		switch {
		case !ok:
			p.addf("%s: option %s is missing", where, letter)
			continue
		case solver.Blank(text):
			p.addf("%s: option %s is empty", where, letter)
			continue
		}
		key := solver.Key(text)
		if first, repeated := firstSaying[key]; repeated {
			p.addf("%s: options %s and %s say the same", where, first, letter)
			continue
		}
		firstSaying[key] = letter
	}

	if solver.Place(task.CorrectAnswer) < 0 {
		p.addf("%s: the correct answer is %q, want one of %v", where, task.CorrectAnswer, solver.Letters())
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
	for _, letter := range solver.Letters() {
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
		if solver.Blank(distractor.Text) {
			p.addf("%s: option %s explains nothing", where, letter)
		}
	}
}

// checkPicture checks that a picture a reference task describes in a member —
// its own, or its solution's — is, where it has one, a description of one of
// the kinds of picture, in the format a task the model writes is held to. A
// reference task is in English, whose numbers a point writes. What the
// picture says of the wording is the business of the tests of the content,
// which hold every reference picture to the checks a task the model writes
// passes.
func checkPicture(p *problems, where, member string, description json.RawMessage) {
	if description == nil {
		return
	}
	_, broken := picture.Parse(description, picture.Point)
	for _, problem := range broken {
		p.addf("%s: %s%s %s", where, member, strings.TrimPrefix(problem.Path, "picture"), problem.Rule)
	}
}

// checkTotal checks that the total of a reference task's solution, where it
// has one, stands under a picture of the solution and is the equality the
// format reads.
func checkTotal(p *problems, where string, task *Example) {
	if task.SolutionTotal == "" {
		return
	}
	if task.SolutionPicture == nil {
		p.addf("%s: a total stands under the picture of the solution, and the task draws none", where)
	}
	_, broken := picture.ReadTotal(task.SolutionTotal, picture.Point)
	for _, problem := range broken {
		p.addf("%s: %s %s", where, problem.Path, problem.Rule)
	}
}

// clone copies a reference task together with everything it points at: the
// options, the explanations of the wrong ones and the pictures. A task holds
// maps, and a map is a reference however many times the value around it is
// copied, so without this a caller could rewrite an option inside the binary's
// own content for every request that follows.
func (e *Example) clone() Example {
	copied := *e
	copied.Options = maps.Clone(e.Options)
	copied.Distractors = maps.Clone(e.Distractors)
	copied.Picture = bytes.Clone(e.Picture)
	copied.SolutionPicture = bytes.Clone(e.SolutionPicture)
	return copied
}

// draws says whether a reference task carries a picture.
func (e *Example) draws() bool { return len(e.Picture) > 0 }

// drawsItsSolution says whether a reference task carries a picture of its
// solution.
func (e *Example) drawsItsSolution() bool { return len(e.SolutionPicture) > 0 }
