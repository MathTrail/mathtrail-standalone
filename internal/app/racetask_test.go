package app_test

import "strings"

// The race is the task the tests of this package hand in, written the way a
// model is asked to write one: a topic and traps from the catalogs, five
// options, a solver in Starlark and a self-check. Its words are kept apart,
// because everything the child reads and everything that gives the answer away
// is what a line of the log must never carry.

// raceSolver proves the answer to the race the way a model is asked to: it
// tries every order, keeps the ones the wording allows, and looks up who came
// first among the options.
const raceSolver = `def solve(options):
    firsts = []
    for order in permutations(["Ann", "Ben", "Kim"]):
        place = {name: i for i, name in enumerate(order)}
        if place["Ben"] < place["Kim"] and place["Kim"] < place["Ann"]:
            firsts.append(order[0])
    return match(options, firsts[0])
`

// What the child reads of the race.
const (
	raceQuestion = "Ann, Ben and Kim ran a race. Ben finished before Kim. Ann finished after Kim. Who finished first?"
	raceHint     = "Who finished before Kim?"
)

// raceOptions are the race's five options, the right one at C.
var raceOptions = map[string]string{"A": "Ann", "B": "Kim", "C": "Ben", "D": "Nobody", "E": "All at once"}

// What gives the race away: the letter of its right option, its solution, what
// the child who chose each wrong option is told, and the model's own notes.
const (
	raceAnswer   = "C"
	raceSolution = "Ben is before Kim, and Kim is before Ann. So Ben is first."
	raceIdea     = "Order three runners from two comparisons."
)

// raceExplained is what the child who chose each wrong option is told, and the
// trap of the catalog it is for.
var raceExplained = map[string]map[string]string{
	"A": {"trap": "reversed_relation", "text": "Ann finished after Kim, so she is last."},
	"B": {"trap": "stopped_early", "text": "Kim is in the middle: Ben beat her."},
	"D": {"trap": "ignored_condition", "text": "Someone always finishes first."},
	"E": {"trap": "answered_other_question", "text": "They finished one after another."},
}

// raceChecked is what the model's self-check says of each option.
var raceChecked = map[string]string{
	"A": "Ann is last.", "B": "Kim is second.", "C": "Ben is first.", "D": "Someone was first.", "E": "Nobody tied.",
}

// raceTask is the race as the model hands it in.
func raceTask() map[string]any {
	return map[string]any{
		"core_idea":      raceIdea,
		"question":       raceQuestion,
		"options":        raceOptions,
		"correct_answer": raceAnswer,
		"hint":           raceHint,
		"solution":       raceSolution,
		"distractors":    raceExplained,
	}
}

// raceSelfCheck is the model's check of its own race.
func raceSelfCheck() map[string]any {
	return map[string]any{"issues": []any{}, "option_check": raceChecked, "final_answer": raceAnswer}
}

// raceWords are the race's texts, every one a line must never carry anywhere
// in it: what the child reads but the options, what gives the answer away, the
// model's notes, and each line of its program that says anything of the race.
// The options are a word or two each, too short to look for inside other
// text; a line must not carry one as a value of its own.
func raceWords() []string {
	words := []string{raceQuestion, raceHint, raceSolution, raceIdea}
	for line := range strings.Lines(raceSolver) {
		if said := strings.TrimSpace(line); strings.ContainsAny(said, `"[`) {
			words = append(words, said)
		}
	}
	for _, wrong := range raceExplained {
		words = append(words, wrong["text"])
	}
	for _, checked := range raceChecked {
		words = append(words, checked)
	}
	return words
}
