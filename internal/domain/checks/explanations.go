package checks

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// The shortest an explanation may be, in the unit its script is counted in.
// The reference tasks say what went wrong in as few as three words — "Monday
// is today." — so three is where the check refuses; the instructions ask the
// model for more. Six characters for scripts written without spaces follows
// the two-to-one of the readability limits.
const (
	minimumWords      = 3
	minimumCharacters = 6
)

// TrapDescriptions is what the explanation check has to know about the trap
// catalog. It is declared here, by the side that needs it, and it speaks in
// identifiers.
type TrapDescriptions interface {
	// TrapDescription is how the catalog describes a trap, and whether it has
	// one by this id.
	TrapDescription(id string) (string, bool)
}

// Explanations checks the four explanations behind the wrong options. Each is
// what a child who chose that option is told, and in text mode it is most of
// what the product teaches. The conditions are mechanical, and nothing here
// judges whether an explanation is a good one — a wrong guess at that would
// cost the child another attempt:
//
//   - no two explanations say the same thing;
//   - none is the solution or the hint, or their opening words;
//   - none is its trap's description from the catalog, copied;
//   - none is too short to say what went wrong.
//
// Texts are compared as they read — case, punctuation and spacing aside. An
// explanation is pointed at by its trap, never by its option's letter: the
// letters of the explanations name the wrong options, and so the right one.
// An explanation with no text at all is the structure check's to report.
func Explanations(draft Draft, traps TrapDescriptions) []Problem {
	if draft.Task == nil {
		return nil
	}
	task := draft.Task
	problems := sameTexts(task)
	for _, key := range slices.Sorted(maps.Keys(task.Distractors)) {
		distractor := task.Distractors[key]
		if strings.TrimSpace(distractor.Text) == "" {
			continue
		}
		problems = append(problems, checkExplanation(task, distractor, traps)...)
	}
	return distinct(problems)
}

// sameTexts reports explanations that say the same thing as another. Each wrong
// option is a different mistake, and a child who made it is owed the
// explanation of that one.
func sameTexts(task *Task) []Problem {
	traps := map[string][]string{}
	for _, distractor := range task.Distractors {
		if said := reading(distractor.Text); said != "" {
			traps[said] = append(traps[said], distractor.Trap)
		}
	}

	var problems []Problem
	for _, said := range slices.Sorted(maps.Keys(traps)) {
		if group := traps[said]; len(group) > 1 {
			problems = append(problems, explanation("%s say the same thing; each wrong option needs its own",
				explanationsFor(group)))
		}
	}
	return problems
}

// checkExplanation checks one explanation on its own: that it is not the way
// to the answer, not the catalog's words, and long enough to say something.
func checkExplanation(task *Task, distractor Distractor, traps TrapDescriptions) []Problem {
	which := explanationFor(distractor.Trap)

	var problems []Problem
	for _, told := range []struct{ field, text string }{
		{"task.solution", task.Solution},
		{"task.hint", task.Hint},
	} {
		if startsWith(told.text, distractor.Text) {
			problems = append(problems, explanation("%s repeats the start of %s; it should say what went wrong "+
				"for a child who chose that option, not how to solve the task", which, told.field))
		}
	}

	if description, known := traps.TrapDescription(distractor.Trap); known && reading(description) == reading(distractor.Text) {
		problems = append(problems, explanation("%s is the catalog's description of its trap; it should say "+
			"what went wrong in this task", which))
	}

	count, unit := lengthOf(distractor.Text)
	minimum := minimumWords
	if unit == "characters" {
		minimum = minimumCharacters
	}
	if count < minimum {
		problems = append(problems, explanation("%s is %s long, and it takes at least %d %s to say what went wrong",
			which, quantity(count, unit), minimum, unit))
	}
	return problems
}

// explanationFor is how a refusal points at an explanation: by its trap, which
// the model chose and can find, rather than by its option's letter.
func explanationFor(trap string) string {
	if trap == "" {
		return "an explanation with no trap"
	}
	return fmt.Sprintf("an explanation for trap %q", trap)
}

// explanationsFor is how a refusal points at several explanations at once: how
// many there are, then each trap among them once, and those with no trap
// apart — "3 explanations for traps "a" and "b"", "2 explanations with no
// trap". The count is of explanations, so it stays true when two share a trap.
func explanationsFor(traps []string) string {
	var named []string
	anonymous := false
	for _, trap := range traps {
		if trap == "" {
			anonymous = true
			continue
		}
		named = append(named, trap)
	}
	slices.Sort(named)
	named = slices.Compact(named)

	var whose []string
	if len(named) > 0 {
		noun := "traps"
		if len(named) == 1 {
			noun = "trap"
		}
		whose = append(whose, "for "+noun+" "+listed(named))
	}
	if anonymous {
		whose = append(whose, "with no trap")
	}
	return fmt.Sprintf("%d explanations %s", len(traps), strings.Join(whose, " and "))
}

// explanation is a problem with the explanations behind the wrong options.
func explanation(format string, args ...any) Problem {
	return Problem{Code: CodeDistractorExplanations, Message: fmt.Sprintf(format, args...)}
}
