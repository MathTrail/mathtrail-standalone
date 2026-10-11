package checks

import "encoding/json"

// Draft is a submission read into the format: the task and the self-check.
// Each is nil when it could not be read at all, and then nothing that needs
// it is checked — a refusal about a field of a task that is not there would
// send the model looking in the wrong place.
type Draft struct {
	Task      *Task
	SelfCheck *SelfCheck
	// Retired names the members the format once had and no longer reads that
	// came all the same, by their paths: a chat begun with the guide that
	// asked for them still writes them.
	Retired []string
	// Mended names the fields read as the model meant them rather than as it
	// wrote them, each once: a letter in another case, an option written as a
	// number, no issues or no picture written as null.
	Mended []string
}

// Task is the task as the model writes it.
type Task struct {
	// CoreIdea is the mathematics before the plot, in a sentence or two: the
	// idea the package names, and why the answer is what it is.
	CoreIdea string `json:"core_idea"`
	// Question is the wording the child reads.
	Question string `json:"question"`
	// Picture is the description of a picture the card draws, as it came: its
	// own check reads it member by member, and a task with none has none.
	Picture json.RawMessage `json:"picture,omitempty"`
	// Options are the five answers, keyed by letter.
	Options map[string]string `json:"options"`
	// CorrectAnswer is the letter of the one right option.
	CorrectAnswer string `json:"correct_answer"`
	// Hint is a nudge that does not give the answer away.
	Hint string `json:"hint"`
	// Solution is the reasoning the child is shown after answering.
	Solution string `json:"solution"`
	// SolutionPicture is the description of the picture of the solution, as it
	// came: shown with the solution, after the answer, so it may show the
	// answer. Every task has one, and the structure check refuses a task that
	// has none.
	SolutionPicture json.RawMessage `json:"solution_picture"`
	// SolutionTotal is the equality the solution comes to, written large under
	// its picture, and nothing without one.
	SolutionTotal string `json:"solution_total,omitempty"`
	// Distractors explain the four wrong options, keyed by letter.
	Distractors map[string]Distractor `json:"distractors"`
}

// Distractor is one wrong option: the mistake it comes from, and what the child
// is told after choosing it.
type Distractor struct {
	Trap string `json:"trap"`
	Text string `json:"text"`
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
