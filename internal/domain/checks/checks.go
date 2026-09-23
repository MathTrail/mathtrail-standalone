// Package checks decides whether a task the chat's model handed in can be
// given to a child: whether it is in the format and agrees with what was asked
// for, and whether it repeats a task the child has had or copies an example
// the model was shown.
//
// It is pure computation. The caller brings the submission, the catalogs and
// what the profile remembers; the package says what is wrong, in words the
// model can act on, and never quotes an option's letter or its text. A refusal
// reaches the widget with everything else the tool returns, and the model
// holds its own draft: it needs to be told what to fix, not shown it.
package checks

import "fmt"

// Code is a reason a submission is refused, as the model and the log see it.
type Code string

const (
	// CodeBadStructure is a submission out of the format, or one that does
	// not agree with what the open request asked for.
	CodeBadStructure Code = "bad_structure"
	// CodeDistractorExplanations is an explanation behind a wrong option that
	// cannot tell a child what went wrong: a copy of another, of the solution
	// or the hint, of the catalog, or too short to say anything.
	CodeDistractorExplanations Code = "distractor_explanations"
	// CodeDrawingFormat is a drawing a phone would not show as it was drawn:
	// too wide, too tall, or made of characters that fonts draw differently.
	CodeDrawingFormat Code = "drawing_format"
	// CodeDrawingMismatch is a drawing that does not show what its structure
	// declares, or a structure that leaves out a label the wording names.
	CodeDrawingMismatch Code = "drawing_mismatch"
	// CodeReadability is a question a child of the grade cannot read: a sentence
	// too long for the level, or, in English, words and sentences too hard for
	// the grade.
	CodeReadability Code = "readability"
	// CodeNearDuplicate is a question that repeats a task the child has had,
	// or copies a reference task the model was shown.
	CodeNearDuplicate Code = "near_duplicate"
)

// Problem is one thing wrong with a submission. Every problem found is
// reported at once, so that the model can fix all of them in one more
// attempt rather than one per attempt.
type Problem struct {
	// Code is which check failed.
	Code Code
	// Message says what to fix. It names fields, catalog ids and limits, and
	// never an option's letter or text.
	Message string
}

// structural is a problem with the format of the submission.
func structural(format string, args ...any) Problem {
	return Problem{Code: CodeBadStructure, Message: fmt.Sprintf(format, args...)}
}

// optionKeys are the five keys a task's options are held under, in the order a
// child reads them.
var optionKeys = []string{"A", "B", "C", "D", "E"}
