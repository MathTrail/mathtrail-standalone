package checks

import "github.com/MathTrail/mathtrail-standalone/internal/domain/profile"

// Draft is a submission read into the format: the brief handed back, the task
// and the self-check. Each is nil when it could not be read at all, and then
// nothing that needs it is checked — a refusal about a field of a task that is
// not there would send the model looking in the wrong place.
type Draft struct {
	Brief     *profile.Brief
	Task      *Task
	SelfCheck *SelfCheck
}

// Task is the task as the model writes it.
type Task struct {
	// CoreIdea is the mathematics before the plot.
	CoreIdea string `json:"core_idea"`
	// DesignThoughtProcess is how the plot and the traps were chosen. Nobody
	// reads it; writing it is what makes the model plan before committing.
	DesignThoughtProcess string `json:"design_thought_process"`
	// Question is the wording the child reads.
	Question string `json:"question"`
	// Drawing and DrawingStructure come together or not at all.
	Drawing          string            `json:"drawing,omitempty"`
	DrawingStructure *DrawingStructure `json:"drawing_structure,omitempty"`
	// Options are the five answers, keyed by letter.
	Options map[string]string `json:"options"`
	// CorrectAnswer is the letter of the one right option.
	CorrectAnswer string `json:"correct_answer"`
	// Hint is a nudge that does not give the answer away.
	Hint string `json:"hint"`
	// Solution is the reasoning the child is shown after answering.
	Solution string `json:"solution"`
	// Distractors explain the four wrong options, keyed by letter.
	Distractors map[string]Distractor `json:"distractors"`
}

// Distractor is one wrong option: the mistake it comes from, and what the child
// is told after choosing it.
type Distractor struct {
	Trap string `json:"trap"`
	Text string `json:"text"`
}

// DrawingStructure is the drawing written out as data, so that the drawing and
// the wording can be compared without either being understood.
type DrawingStructure struct {
	Kind      string            `json:"kind"`
	Objects   []DrawingObject   `json:"objects"`
	Relations []DrawingRelation `json:"relations,omitempty"`
}

// DrawingObject is one thing the drawing shows.
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

// SelfCheck is the model's own pass over the task it wrote.
type SelfCheck struct {
	// Issues is every defect the model found, and an empty list when there is
	// none — which is not the same as a list left out.
	Issues []Issue `json:"issues"`
	// OptionCheck says why each option is right or wrong, keyed by letter.
	OptionCheck map[string]string `json:"option_check"`
	// FinalAnswer is the letter the model arrives at solving its own task
	// afresh, or unsolvable.
	FinalAnswer string `json:"final_answer"`
}

// Issue is one defect the model found in its own task.
type Issue struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Comment  string `json:"comment"`
}

// unsolvable is the final answer of a self-check that could not solve the task
// as written.
const unsolvable = "UNSOLVABLE"

// issueTypes are the defects a self-check can name.
var issueTypes = []string{
	"ambiguous", "missing_data", "multiple_correct", "no_correct",
	"too_hard_for_grade", "needs_picture", "factual_error",
}

// How much an issue matters: a blocking one refuses the task, and a minor one
// is counted and let through.
const (
	severityBlocking = "blocking"
	severityMinor    = "minor"
)

// severities are every severity an issue may have.
var severities = []string{severityBlocking, severityMinor}
