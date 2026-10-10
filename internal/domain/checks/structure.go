package checks

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// Structure checks a draft against the format: every text the child or the
// model relies on is there, the ids it names exist, and the self-check says
// all it has to. Where the task stands — its topic, level and difficulty — is
// not the draft's to say: the request it is handed in for keeps that.
//
// A part the draft could not read is not checked; Decode has already said why.
func Structure(draft Draft, catalog TrapDescriber) []Problem {
	var problems []Problem
	if draft.Task != nil {
		problems = append(problems, checkTask(draft.Task, catalog)...)
	}
	if draft.SelfCheck != nil {
		problems = append(problems, checkSelfCheck(draft.SelfCheck)...)
	}
	return problems
}

// The longest each text of a task the child reads may be, in characters. Each
// is several times the longest of the reference tasks: past it, a text is no
// longer one a child reads on a card, and it would swell the file the task is
// kept in, which is read back whole on every call.
const (
	longestQuestion    = 1500
	longestOption      = 200
	longestHint        = 500
	longestSolution    = 2000
	longestExplanation = 500
)

// checkTask checks the task itself: every text the child or the model relies
// on is there and no longer than a card holds, the child is offered five
// answers they can tell apart, and every wrong one, and only a wrong one,
// carries the mistake it comes from.
func checkTask(task *Task, catalog TrapDescriber) []Problem {
	var problems []Problem
	problems = append(problems, required("task.core_idea", task.CoreIdea)...)
	problems = append(problems, required("task.question", task.Question)...)
	problems = append(problems, required("task.hint", task.Hint)...)
	problems = append(problems, required("task.solution", task.Solution)...)
	problems = append(problems, bounded("task.question", task.Question, longestQuestion)...)
	problems = append(problems, bounded("task.hint", task.Hint, longestHint)...)
	problems = append(problems, bounded("task.solution", task.Solution, longestSolution)...)
	problems = append(problems, checkOptions(task)...)
	problems = append(problems, checkDistractors(task, catalog)...)
	problems = append(problems, checkPicture(task)...)
	problems = append(problems, checkSolutionPicture(task)...)
	return problems
}

// checkOptions checks that the child is offered five answers they can tell
// apart, with exactly one of them marked correct. Two options are one answer
// when the solver would take them for one, which is also when a child reads
// them as one.
func checkOptions(task *Task) []Problem {
	var problems []Problem
	if !sameKeys(task.Options, optionKeys) {
		problems = append(problems, structural("task.options must hold exactly five options, "+
			"under the five keys of the format"))
	}

	said := make(map[string]bool, len(task.Options))
	var empty, repeated, long int
	for _, text := range task.Options {
		key := solver.Key(text)
		switch {
		case solver.Blank(text):
			empty++
		case said[key]:
			repeated++
		}
		if utf8.RuneCountInString(text) > longestOption {
			long++
		}
		said[key] = true
	}
	if empty > 0 {
		problems = append(problems, structural("task.options has %s", several(empty, "an empty option", "empty options")))
	}
	if long > 0 {
		problems = append(problems, structural("task.options has %s; an option is at most %d characters",
			several(long, "an option longer than a card holds", "options longer than a card holds"), longestOption))
	}
	if repeated > 0 {
		problems = append(problems, structural("task.options has %s: the same number, or the same words once "+
			"letter case, the width of the gaps between them, the width of the characters themselves and "+
			"characters that take no room are set aside",
			several(repeated, "an option that says the same as another", "options that say the same as others")))
	}

	if solver.Place(task.CorrectAnswer) < 0 {
		problems = append(problems, structural("task.correct_answer must be one of the five option keys"))
	}
	return problems
}

// checkDistractors checks that every wrong option, and only a wrong option, is
// explained, and that each explanation names a mistake the catalog knows. This
// is what turns a wrong answer into a diagnosis.
func checkDistractors(task *Task, catalog TrapDescriber) []Problem {
	var problems []Problem
	switch {
	case task.Distractors == nil:
		problems = append(problems, structural("task.distractors is missing"))
	case solver.Place(task.CorrectAnswer) < 0:
		// Which options are the wrong ones is not known; the correct answer's
		// own refusal comes first.
	case !sameKeys(task.Distractors, wrongKeys(task.CorrectAnswer)):
		problems = append(problems, structural("task.distractors must explain exactly the four wrong options, "+
			"each once, and not the correct one"))
	}

	// The entries are keyed by an option's letter and cannot be pointed at, so
	// what is wrong with them is counted; a trap nobody has is named once,
	// however many explanations name it.
	var untold, unnamed, long int
	unknown := map[string]bool{}
	for _, distractor := range task.Distractors {
		if solver.Blank(distractor.Text) {
			untold++
		}
		if utf8.RuneCountInString(distractor.Text) > longestExplanation {
			long++
		}
		switch _, known := catalog.TrapDescription(distractor.Trap); {
		case distractor.Trap == "":
			unnamed++
		case !known:
			unknown[distractor.Trap] = true
		}
	}
	for _, trap := range slices.Sorted(maps.Keys(unknown)) {
		problems = append(problems, structural("task.distractors names trap %q, which is not in the catalog", trap))
	}
	if untold > 0 {
		problems = append(problems, structural("task.distractors has %s",
			several(untold, "an explanation with no text", "explanations with no text")))
	}
	if unnamed > 0 {
		problems = append(problems, structural("task.distractors has %s",
			several(unnamed, "an explanation that names no trap", "explanations that name no trap")))
	}
	if long > 0 {
		problems = append(problems, structural("task.distractors has %s; an explanation is at most %d characters",
			several(long, "an explanation longer than a card holds", "explanations longer than a card holds"),
			longestExplanation))
	}
	return problems
}

// wrongKeys are the keys of the four options that are not the correct one.
func wrongKeys(correct string) []string {
	return slices.DeleteFunc(slices.Clone(optionKeys), func(key string) bool { return key == correct })
}

// checkPicture checks that a picture, when the task has one, is an object:
// what the object holds is the picture's own check to judge, member by member.
func checkPicture(task *Task) []Problem {
	var members map[string]json.RawMessage
	if task.Picture == nil || json.Unmarshal(task.Picture, &members) == nil {
		return nil
	}
	return []Problem{structural("task.picture must be an object: the description of one of the kinds of picture, " +
		"or left out")}
}

// checkSolutionPicture checks that the task draws its solution, in an object:
// the card of how an answer went shows every solution with its picture, so a
// task without one is refused for it. The total under the picture needs no
// word of its own here: it stands under the picture the refusal asks for.
func checkSolutionPicture(task *Task) []Problem {
	var members map[string]json.RawMessage
	switch {
	case task.SolutionPicture == nil:
		return []Problem{structural("task.solution_picture is missing, or holds nothing: every task draws its " +
			"solution — what the solution works out, in the format of a picture, which the card shows once the " +
			"child has answered. The guide in the package says how; get_package with the same request_id hands it " +
			"back")}
	case json.Unmarshal(task.SolutionPicture, &members) != nil:
		return []Problem{structural("task.solution_picture must be an object: the description of one of the kinds " +
			"of picture")}
	}
	return nil
}

// checkSelfCheck checks that the model's own pass over its task is complete:
// every issue is one the format names, every option has its reason, and the
// answer it arrived at is one it could have arrived at.
func checkSelfCheck(check *SelfCheck) []Problem {
	var problems []Problem
	if check.Issues == nil {
		problems = append(problems, structural("self_check.issues is missing; an empty list says there are none"))
	}
	for i, issue := range check.Issues {
		if !slices.Contains(issueTypes, issue.Type) {
			problems = append(problems, structural("self_check.issues.%d has type %q, and the format's types are %s",
				i, issue.Type, strings.Join(issueTypes, ", ")))
		}
		if !slices.Contains(severities, issue.Severity) {
			problems = append(problems, structural("self_check.issues.%d has severity %q, and the format's are %s",
				i, issue.Severity, strings.Join(severities, " and ")))
		}
		if solver.Blank(issue.Comment) {
			problems = append(problems, structural("self_check.issues.%d has no comment", i))
		}
	}

	if !sameKeys(check.OptionCheck, optionKeys) {
		problems = append(problems, structural("self_check.option_check must say why each of the five options "+
			"is right or wrong, under the five keys of the format"))
	}
	empty := 0
	for _, reason := range check.OptionCheck {
		if solver.Blank(reason) {
			empty++
		}
	}
	if empty > 0 {
		problems = append(problems, structural("self_check.option_check has %s", several(empty, "an empty entry", "empty entries")))
	}

	if solver.Place(check.FinalAnswer) < 0 && check.FinalAnswer != unsolvable {
		problems = append(problems, structural("self_check.final_answer must be one of the five option keys, or %s",
			unsolvable))
	}
	return problems
}

// required reports a text the format requires and the draft left empty.
func required(field, text string) []Problem {
	if solver.Blank(text) {
		return []Problem{structural("%s is missing or empty", field)}
	}
	return nil
}

// bounded reports a text longer than a card holds.
func bounded(field, text string, longest int) []Problem {
	if length := utf8.RuneCountInString(text); length > longest {
		return []Problem{structural("%s is %d characters long, and it is at most %d", field, length, longest)}
	}
	return nil
}

// sameKeys says whether a map is keyed by exactly these keys.
func sameKeys[V any](entries map[string]V, keys []string) bool {
	if len(entries) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, present := entries[key]; !present {
			return false
		}
	}
	return true
}

// several says how many of something are wrong, in words that stay right for
// one. An entry keyed by an option's letter cannot be pointed at without the
// letter, so the refusal counts rather than names.
func several(count int, one, many string) string {
	if count == 1 {
		return one
	}
	return fmt.Sprintf("%d %s", count, many)
}
