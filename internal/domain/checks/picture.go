package checks

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// A task's picture is held to two checks. Its format: a description of one of
// the kinds of picture, every member of its kind within the kind's limits. And
// its match with the wording: the labels the wording names are the picture's,
// the picture's labels in capitals and the words of its colours are named in
// the wording, and what it shows does not hold the right answer the wording
// does not give. Whether the picture means what the wording says is past what
// a program can tell; the model's self-check answers for it.
//
// The picture of the solution is shown with the solution, after the answer:
// it is held to the same format, its labels in capitals and the words of its
// colours named in the question or the solution, and it may show the answer.
// The total under it is the equality the solution comes to, and ends in the
// right answer.

// The members of a task a picture is described in.
const (
	pictureField  = "picture"
	solutionField = "solution_picture"
)

// pictured is a picture of a task as it was read, once, for both its checks,
// the member of the task it was read from, and how the lesson's wording names
// its labels.
type pictured struct {
	field    string
	read     picture.Picture
	broken   []picture.Problem
	decimals picture.Decimals
	by       labelling
}

// readPicture reads the picture a task describes, its numbers written with the
// decimal mark of the lesson's language. A task with no picture has nothing to
// read, and neither has one whose picture is no object: the structure check
// says what is wrong with that.
func readPicture(task *Task, lesson string) *pictured {
	if task == nil {
		return nil
	}
	return readDescribed(task.Picture, pictureField, lesson)
}

// readSolutionPicture reads the picture of the solution a task describes, as
// readPicture reads its picture.
func readSolutionPicture(task *Task, lesson string) *pictured {
	if task == nil {
		return nil
	}
	return readDescribed(task.SolutionPicture, solutionField, lesson)
}

// readDescribed reads a description kept in a member of a task, or nothing
// when it holds no object.
func readDescribed(described json.RawMessage, field, lesson string) *pictured {
	var members map[string]json.RawMessage
	if described == nil || json.Unmarshal(described, &members) != nil {
		return nil
	}
	decimals := decimalsOf(lesson)
	read, broken := picture.Parse(described, decimals)
	return &pictured{field: field, read: read, broken: broken, decimals: decimals, by: labellingOf(lesson)}
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
	problems, _ := readPicture(task, lesson).match(task, false)
	return problems
}

// PictureMatchFromPicture checks the picture of a task solved from the
// picture against its wording, as PictureMatch does a word task's, by the
// rules such a task is held to.
func PictureMatchFromPicture(task *Task, lesson string) []Problem {
	problems, _ := readPicture(task, lesson).match(task, true)
	return problems
}

// SolutionPicture checks the picture of the solution of a task written in a
// lesson's language, and the total under it, as the review checks them: their
// format, and their match with the wording and the solution.
func SolutionPicture(task *Task, lesson string) []Problem {
	read := readSolutionPicture(task, lesson)
	matched, _ := read.matchSolution(task)
	return append(read.format(), matched...)
}

// format is every rule the picture breaks, as refusals of its format, each
// named by its path in the member the picture was read from.
func (p *pictured) format() []Problem {
	if p == nil {
		return nil
	}
	problems := make([]Problem, 0, len(p.broken))
	for _, broken := range p.broken {
		path := p.field + strings.TrimPrefix(broken.Path, pictureField)
		problems = append(problems, Problem{Code: CodeDrawingFormat, Message: "task." + path + " " + broken.Rule})
	}
	return problems
}

// match holds the picture to the wording, as far as it could be read.
//
// The labels the wording names are looked for only in a picture read whole,
// since a label left out for breaking a rule would be reported missing. The
// labels the picture shows, and the answer, are held to the wording in any
// picture that names its kind, so that one more attempt mends all of it.
//
// The picture of a task solved from the picture holds the task's facts, and
// draws among them the candidates the child chooses from: it shows the answer
// only where it draws the right value without the values of at least three
// wrong options beside it, and a picture of tiles or of shapes marks the place
// it asks about with ?.
func (p *pictured) match(task *Task, fromPicture bool) (problems []Problem, unchecked string) {
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
	if unnamed := unnamedLabels(task.Question, labels, p.by); unnamed > 0 {
		problems = append(problems, mismatch("task.question does not name %d of the labels task.picture shows in "+
			"Latin capitals; name each in the question — the football club (F), segment PQ, cell K2, rows P to S — "+
			"or leave it out of the picture", unnamed))
	}
	if unnamed := unnamedColors(task.Question, p.read.Words()); unnamed > 0 {
		problems = append(problems, mismatch("task.question does not name %d of the colours task.picture calls by "+
			"a word in its colors; give each the word the question writes, in its form there, or leave the colour "+
			"out of the picture", unnamed))
	}
	if p.showsTheAnswer(task, fromPicture) {
		problems = append(problems, mismatch("task.picture shows the right answer, which the question does not "+
			"give: show what the question gives and nothing it asks for, and mark the unknown with ?"))
	}
	if fromPicture && len(p.broken) == 0 && marksTheAsked(p.read.Kind()) && !slices.Contains(labels, "?") {
		problems = append(problems, mismatch("task.picture marks no place with ?: mark the triangle or the "+
			"total the question asks about with ?"))
	}
	return problems, unchecked
}

// showsTheAnswer says whether the picture shows the right option the question
// does not give. A picture of a task solved from the picture that draws, beside
// it, the values of at least three of the wrong options draws candidates the
// child chooses among, and shows no answer by drawing them.
func (p *pictured) showsTheAnswer(task *Task, fromPicture bool) bool {
	right, named := rightOption(task)
	if !named || !showsTheAnswer(p.read, task.Question, p.by.asLatin(right), p.decimals, p.by) {
		return false
	}
	if !fromPicture {
		return true
	}
	drawn := 0
	for letter, wrong := range task.Options {
		if letter != task.CorrectAnswer && !solver.Blank(wrong) &&
			showsTheAnswer(p.read, task.Question, p.by.asLatin(wrong), p.decimals, p.by) {
			drawn++
		}
	}
	return drawn < candidatesDrawn
}

// candidatesDrawn is how many of the four wrong options' values a picture of a
// task solved from the picture draws beside the right one for the five to be
// candidates rather than the answer shown: two let pictures of two dice, two
// ice creams and a fold that show their answer through, and four refuse the
// crowns, which draw three.
const candidatesDrawn = 3

// marksTheAsked says whether a picture of a kind marks the place its task asks
// about with ?: a tile's triangle or a row's total, which the picture would
// otherwise leave the child to guess.
func marksTheAsked(kind picture.Kind) bool {
	return kind == picture.Tiles || kind == picture.Shapes
}

// matchSolution holds the picture of the solution to the wording and the
// solution: its labels in capitals, and those of the total under it, are named
// in the question or the solution, and the total ends in the right answer. A
// total that breaks its rule is a refusal of its format.
func (p *pictured) matchSolution(task *Task) (problems []Problem, unchecked string) {
	switch {
	case p == nil:
		return nil, ""
	case p.read == nil:
		return nil, "the picture of the solution was not held to the wording: that needs task.solution_picture in " +
			"the format of one of the kinds of picture"
	}
	labels := p.read.Labels()
	if task.SolutionTotal != "" {
		total, broken := picture.ReadTotal(task.SolutionTotal, p.decimals)
		for _, rule := range broken {
			problems = append(problems, Problem{Code: CodeDrawingFormat, Message: "task." + rule.Path + " " + rule.Rule})
		}
		labels = append(labels, total.Labels...)
		if right, named := rightOption(task); named && len(broken) == 0 &&
			!endsInTheAnswer(total, p.by.asLatin(right), p.decimals) {
			problems = append(problems, mismatch("task.solution_total does not end in the right answer: end it in "+
				"the right option's number or label, or leave it out"))
		}
	}
	if unnamed := unnamedLabels(task.Question+" "+task.Solution, labels, p.by); unnamed > 0 {
		problems = append(problems, mismatch("neither task.question nor task.solution names %d of the labels "+
			"task.solution_picture and its total show in Latin capitals; name each, or leave it out of the picture",
			unnamed))
	}
	if unnamed := unnamedColors(task.Question+" "+task.Solution, p.read.Words()); unnamed > 0 {
		problems = append(problems, mismatch("neither task.question nor task.solution names %d of the colours "+
			"task.solution_picture calls by a word in its colors; give each the word they write, in its form there, "+
			"or leave the colour out of the picture", unnamed))
	}
	return problems, unchecked
}

// endsInTheAnswer says whether a total ends in the right option: the one
// number the option holds, or the option itself, a label, as two texts are
// compared. An option that holds no number and is no label — a number written
// in words — cannot be held to the number a total ends in, and is not.
func endsInTheAnswer(total picture.Total, right string, decimals picture.Decimals) bool {
	option, result := answerOf(right, decimals), keyOf(total.Result, decimals)
	switch {
	case option.number != "":
		return result == option.number
	case capitalsOnly(option.text):
		return result == option.key
	}
	return true
}

// rightOption is the text of the task's right option, and whether the task
// names one. A task that does not is the structure check's to refuse.
func rightOption(task *Task) (string, bool) {
	text, held := task.Options[task.CorrectAnswer]
	return text, held && !solver.Blank(text)
}
