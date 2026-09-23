package checks

import "github.com/MathTrail/mathtrail-standalone/internal/domain/solver"

// runFailures are what the model is told of a solver that did not run to an
// answer: which way it failed, and never what the interpreter or the program
// said, which can quote an option's text or name a letter. A run stopped by a
// limit is told by the sandbox's own sentence instead, which names the limit.
var runFailures = map[solver.Status]string{
	solver.StatusBadSource: "the solver does not parse, or asks for something the dialect does not have; " +
		"keep to the dialect the instructions describe",
	solver.StatusNoEntryPoint: "the solver defines no solve to call; it needs one function, solve, " +
		"that takes the options and nothing else",
	solver.StatusError: "the solver failed while it ran; walk it through the task's own options " +
		"and fix what stops it",
	solver.StatusBadOutput: "solve returned something other than a list of distinct option letters; " +
		"end it with return match(options, value)",
}

// unknownFailure is what the model is told of a run that failed in a way the
// sentences above do not name, so that no refusal is ever left without words.
const unknownFailure = "the solver did not run to an answer; walk it through the task's own options " +
	"and fix what stops it"

// solverRuns reports a solver that did not run to an answer. An agreement
// with no run in it at all is refused rather than trusted: a task whose
// solver never ran has not been proven.
func (r *reviewer) solverRuns(answer *solver.Agreement) (problems []Problem, unchecked string) {
	if answer == nil {
		return nil, "the solver was not run: that needs the five options of task.options"
	}
	if len(answer.Runs) == 0 {
		return []Problem{{Code: CodeSolverError, Message: "the solver did not run; hand the task in again"}}, ""
	}
	failed := answer.Failed()
	if failed == nil {
		return nil, ""
	}
	message, named := runFailures[failed.Status]
	switch {
	case failed.Status == solver.StatusTimeout:
		message = failed.Message + "; count the cases rather than list them, or search a smaller space"
	case !named:
		message = unknownFailure
	}
	return []Problem{{Code: CodeSolverError, Message: message}}, ""
}

// answers holds the three witnesses to a task's answer against each other: the
// option the task names, the one its solver arrives at and the one its
// self-check does. The model is told which of them disagreed, never with the
// letters: a refusal that named them would name the answer.
func (r *reviewer) answers(answer *solver.Agreement, draft Draft) (problems []Problem, unchecked string) {
	task := draft.Task
	if task == nil || solver.Place(task.CorrectAnswer) < 0 {
		return nil, "the answers were not compared: that needs task.correct_answer to be one of the option letters"
	}

	if ran(answer) {
		if message := solverVerdict(answer, task.CorrectAnswer); message != "" {
			problems = append(problems, disagrees(message))
		}
	}
	if message := selfCheckVerdict(draft.SelfCheck, task.CorrectAnswer); message != "" {
		problems = append(problems, disagrees(message))
	}
	return problems, ""
}

// ran says whether a solver ran to an answer of its own: at least once, and
// never into a failure. Only then is there a verdict to hold against the task.
func ran(answer *solver.Agreement) bool {
	return answer != nil && len(answer.Runs) > 0 && answer.Failed() == nil
}

// solverVerdict says how a solver that ran disagrees with the task, or nothing
// when it arrived at the option the task names.
func solverVerdict(answer *solver.Agreement, correct string) string {
	switch {
	case answer.Agreed() && answer.Letter == correct:
		return ""
	case answer.Agreed():
		return "the solver arrived at a different option from task.correct_answer; " +
			"find which of the two is wrong and fix it"
	case len(answer.Runs) == 2:
		return "the solver pointed elsewhere once the options were relabelled: it reads the letters " +
			"rather than working the answer out; compute the value and end with return match(options, value)"
	case len(answer.Runs[0].Letters) == 0:
		return "the solver found no option that fits what it computed; " +
			"check the options against the answer, and the solver against the wording"
	default:
		return "the solver found more than one option that fits; a task has exactly one right option"
	}
}

// selfCheckVerdict says how the self-check disagrees with the task, or nothing
// when it arrived at the option the task names. An answer that is not an
// answer at all is the structure check's to report.
func selfCheckVerdict(check *SelfCheck, correct string) string {
	switch {
	case check == nil || check.FinalAnswer == correct:
		return ""
	case check.FinalAnswer == unsolvable:
		return "self_check.final_answer says the task cannot be solved as written; fix the wording until it can"
	case solver.Place(check.FinalAnswer) < 0:
		return ""
	default:
		return "self_check.final_answer is not task.correct_answer; solve the task afresh and fix whichever is wrong"
	}
}

// disagrees is a problem with how the witnesses to a task's answer agree.
func disagrees(message string) Problem {
	return Problem{Code: CodeSolverDisagrees, Message: message}
}
