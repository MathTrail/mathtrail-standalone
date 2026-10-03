// Package reviewing makes the service's reference tasks into submissions the
// service's checks accept, and reviews submissions with the service's own
// reviewer and sandbox: what the experiments that hand tasks to the checks
// share, so that each hands in the same tasks and reviews them the same way.
package reviewing

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The texts a reference task lacks and a submission needs. They are the same
// for every host, so that whatever a host is refused for, it is not for what
// the harness gave it.
const (
	fillCoreIdea = "Reference task %s."
	fillDesign   = "Taken from the reference set without change."
	fillHint     = "Read the question again and check each option against it."
	fillBrief    = "A reference task."
	fillOption   = "Checked against the solution."
)

// Language is the language every reference task is written in, and what a
// submission made of one is judged as.
const Language = "en"

// Submission is a task as it is handed in, with what it is judged against: the
// request it answers and the sketches of the child's earlier tasks.
type Submission struct {
	Brief        profile.Brief
	Task         checks.Task
	SelfCheck    checks.SelfCheck
	Solver       string
	Asked        profile.Brief
	Fingerprints []string
}

// Host is a reference task made into a submission.
type Host struct {
	ID         string
	Topic      string
	Level      rating.GradeLevel
	Difficulty int
	Base       Submission
}

// HasDrawing says whether the host's task comes with a drawing.
func (h *Host) HasDrawing() bool { return h.Base.Task.Drawing != "" }

// Complete makes a reference task into a submission: the task as written, its
// solver, a brief that asks for exactly it, and a self-check that found nothing
// wrong and agrees with the key.
func Complete(example *content.Example) *Host {
	task := checks.Task{
		CoreIdea:             fmt.Sprintf(fillCoreIdea, example.ID),
		DesignThoughtProcess: fillDesign,
		Question:             example.Question,
		Drawing:              example.Drawing,
		DrawingStructure:     structureOf(example.DrawingStructure),
		Options:              maps.Clone(example.Options),
		CorrectAnswer:        example.CorrectAnswer,
		Hint:                 cmp.Or(example.Hint, fillHint),
		Solution:             example.Solution,
		Distractors:          make(map[string]checks.Distractor, len(example.Distractors)),
	}
	var traps []string
	for _, letter := range solver.Letters() {
		distractor, wrong := example.Distractors[letter]
		if !wrong {
			continue
		}
		task.Distractors[letter] = checks.Distractor{Trap: distractor.Trap, Text: distractor.Text}
		if !slices.Contains(traps, distractor.Trap) {
			traps = append(traps, distractor.Trap)
		}
	}
	brief := profile.Brief{
		Constraints:     []string{},
		Difficulty:      example.Difficulty,
		ExcludedSkills:  []string{},
		GradeLevel:      example.GradeLevel,
		PedagogicalGoal: profile.GoalNewTopic,
		Rationale:       fillBrief,
		Setting:         fillBrief,
		TargetConcept:   example.Topic,
		TrapsToUse:      traps,
	}
	optionCheck := make(map[string]string, solver.Count)
	for _, letter := range solver.Letters() {
		optionCheck[letter] = fillOption
	}
	return &Host{
		ID:         example.ID,
		Topic:      example.Topic,
		Level:      example.GradeLevel,
		Difficulty: example.Difficulty,
		Base: Submission{
			Brief:        brief,
			Task:         task,
			SelfCheck:    checks.SelfCheck{Issues: []checks.Issue{}, OptionCheck: optionCheck, FinalAnswer: example.CorrectAnswer},
			Solver:       example.Solver,
			Asked:        cloneBrief(&brief),
			Fingerprints: []string{},
		},
	}
}

// Hosts makes every reference task into a submission, in the order of their
// ids.
func Hosts(shipped *content.Content) []*Host {
	examples := shipped.Examples()
	hosts := make([]*Host, 0, len(examples))
	for i := range examples {
		hosts = append(hosts, Complete(&examples[i]))
	}
	slices.SortFunc(hosts, func(a, b *Host) int { return cmp.Compare(a.ID, b.ID) })
	return hosts
}

// structureOf is a drawing's structure as the checks read it.
func structureOf(written *content.DrawingStructure) *checks.DrawingStructure {
	if written == nil {
		return nil
	}
	structure := &checks.DrawingStructure{Kind: written.Kind}
	for _, object := range written.Objects {
		structure.Objects = append(structure.Objects, checks.DrawingObject{ID: object.ID, Label: object.Label, Value: object.Value})
	}
	for _, relation := range written.Relations {
		structure.Relations = append(structure.Relations, checks.DrawingRelation{Type: relation.Type, From: relation.From, To: relation.To})
	}
	return structure
}

// Clone copies a submission deeply enough that a defect put into the copy
// leaves the original as it was.
func (s *Submission) Clone() Submission {
	copied := *s
	copied.Brief = cloneBrief(&s.Brief)
	copied.Asked = cloneBrief(&s.Asked)
	copied.Task = CloneTask(&s.Task)
	copied.SelfCheck = checks.SelfCheck{
		Issues:      slices.Clone(s.SelfCheck.Issues),
		OptionCheck: maps.Clone(s.SelfCheck.OptionCheck),
		FinalAnswer: s.SelfCheck.FinalAnswer,
	}
	copied.Fingerprints = slices.Clone(s.Fingerprints)
	return copied
}

func cloneBrief(brief *profile.Brief) profile.Brief {
	copied := *brief
	copied.Constraints = slices.Clone(brief.Constraints)
	copied.ExcludedSkills = slices.Clone(brief.ExcludedSkills)
	copied.TrapsToUse = slices.Clone(brief.TrapsToUse)
	return copied
}

// CloneTask copies a task deeply enough that a change to the copy leaves the
// original as it was.
func CloneTask(task *checks.Task) checks.Task {
	copied := *task
	copied.Options = maps.Clone(task.Options)
	copied.Distractors = maps.Clone(task.Distractors)
	if task.DrawingStructure != nil {
		structure := *task.DrawingStructure
		structure.Objects = slices.Clone(structure.Objects)
		structure.Relations = slices.Clone(structure.Relations)
		copied.DrawingStructure = &structure
	}
	return copied
}

// HandedIn is the submission as the service receives it: three parts of JSON
// and the solver's source.
func (s *Submission) HandedIn() (checks.Submission, error) {
	brief, err := json.Marshal(&s.Brief)
	if err != nil {
		return checks.Submission{}, fmt.Errorf("write the brief: %w", err)
	}
	task, err := json.Marshal(&s.Task)
	if err != nil {
		return checks.Submission{}, fmt.Errorf("write the task: %w", err)
	}
	selfCheck, err := json.Marshal(&s.SelfCheck)
	if err != nil {
		return checks.Submission{}, fmt.Errorf("write the self-check: %w", err)
	}
	return checks.Submission{Brief: brief, Task: task, SelfCheck: selfCheck, Solver: s.Solver}, nil
}

// SameAs says whether two submissions hand in the same thing and are judged
// against the same request and history.
func (s *Submission) SameAs(other *Submission) bool {
	mine, err := json.Marshal(s)
	if err != nil {
		return false
	}
	theirs, err := json.Marshal(other)
	return err == nil && bytes.Equal(mine, theirs)
}

// Options are the five option texts in the order of their letters, as a
// solver takes them.
func (s *Submission) Options() solver.Options {
	var options solver.Options
	for place, letter := range solver.Letters() {
		options[place] = s.Task.Options[letter]
	}
	return options
}
