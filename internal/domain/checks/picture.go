package checks

import (
	"encoding/json"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// A task's picture is held to two checks. Its format: a description of one of
// the kinds of picture, every member of its kind within the kind's limits. And
// its match with the wording: the labels the wording names are the picture's,
// the picture's labels in capitals are named in the wording, and what it shows
// does not hold the right answer the wording does not give. Whether the
// picture means what the wording says is past what a program can tell; the
// model's self-check answers for it.

// pictured is a task's picture as it was read, once, for both its checks.
type pictured struct {
	read     picture.Picture
	broken   []picture.Problem
	decimals picture.Decimals
}

// readPicture reads the picture a task describes, its numbers written with the
// decimal mark of the lesson's language. A task with no picture has nothing to
// read, and neither has one whose picture is no object: the structure check
// says what is wrong with that.
func readPicture(task *Task, lesson string) *pictured {
	var members map[string]json.RawMessage
	if task == nil || task.Picture == nil || json.Unmarshal(task.Picture, &members) != nil {
		return nil
	}
	decimals := decimalsOf(lesson)
	read, broken := picture.Parse(task.Picture, decimals)
	return &pictured{read: read, broken: broken, decimals: decimals}
}

// PictureFormat checks that the picture of a task written in a lesson's
// language is a description of one of the kinds of picture: every rule it
// breaks is a problem that names the member and the rule, never the value.
func PictureFormat(task *Task, lesson string) []Problem {
	return readPicture(task, lesson).format()
}

// PictureMatch checks the picture of a task written in a lesson's language
// against its wording: the labels the wording names, the labels the picture
// shows, and the right answer.
func PictureMatch(task *Task, lesson string) []Problem {
	problems, _ := readPicture(task, lesson).match(task)
	return problems
}

// format is every rule the picture breaks, as refusals of its format.
func (p *pictured) format() []Problem {
	if p == nil {
		return nil
	}
	problems := make([]Problem, 0, len(p.broken))
	for _, broken := range p.broken {
		problems = append(problems, Problem{Code: CodeDrawingFormat, Message: "task." + broken.Path + " " + broken.Rule})
	}
	return problems
}

// match holds the picture to the wording, as far as it could be read.
//
// The labels the wording names are looked for only in a picture read whole,
// since a label left out for breaking a rule would be reported missing. The
// labels the picture shows, and the answer, are held to the wording in any
// picture that names its kind, so that one more attempt mends all of it.
func (p *pictured) match(task *Task) (problems []Problem, unchecked string) {
	switch {
	case p == nil:
		return nil, ""
	case p.read == nil:
		return nil, "the picture was not held to the wording: that needs a description of one of the kinds of picture"
	}
	labels := p.read.Labels()
	if len(p.broken) == 0 {
		if unpictured := unpicturedLabels(task.Question, labels); len(unpictured) > 0 {
			problems = append(problems, mismatch("task.question names %s, which task.picture does not label; "+
				"label each in the picture, or leave it out of the question", listed(unpictured)))
		}
	} else {
		unchecked = "the labels the question names were not looked for in the picture: that needs task.picture " +
			"in the format"
	}
	if unnamed := unnamedLabels(task.Question, labels); unnamed > 0 {
		problems = append(problems, mismatch("task.question does not name %d of the labels task.picture shows in "+
			"Latin capitals; name each in the question — the football club (F), segment PQ, cell K2, rows P to S — "+
			"or leave it out of the picture", unnamed))
	}
	if right, named := rightOption(task); named && showsTheAnswer(p.read, task.Question, right, p.decimals) {
		problems = append(problems, mismatch("task.picture shows the right answer, which the question does not "+
			"give: show what the question gives and nothing it asks for, and mark the unknown with ?"))
	}
	return problems, unchecked
}

// rightOption is the text of the task's right option, and whether the task
// names one. A task that does not is the structure check's to refuse.
func rightOption(task *Task) (string, bool) {
	text, held := task.Options[task.CorrectAnswer]
	return text, held && !solver.Blank(text)
}
