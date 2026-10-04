package tutor

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// fitText says in words where the recommended difficulty landed against the
// corridor, because "too_hard" in a sentence meant for a person is not a
// sentence.
var fitText = map[rating.Fit]string{
	rating.FitInside:  "inside the corridor",
	rating.FitTooHard: "the closest to the corridor, which holds no task of the topic; a little too hard",
	rating.FitTooEasy: "the closest to the corridor, which holds no task of the topic; a little too easy",
	rating.FitUnknown: "the middle one, standing in for a level that could not be read",
}

// failureBehind says how much of a run is behind this choice, and where.
func failureBehind(failures int, topic string) string {
	run := "a failure"
	if failures > 1 {
		run = fmt.Sprintf("%d failures in a row", failures)
	}
	return fmt.Sprintf("%s, the last one in %s", run, topic)
}

// rationale is the one sentence the brief carries about why it says what it
// says. When the model or the person chose instead, it keeps every account:
// what the model asked for, the topic the lessons are kept to, and what the
// rule would have done, so that they can be told apart afterwards rather than
// guessed at. Where the model named a topic or a level and left the rest to
// the corridor, the task it gets is one the rule's account, being of another
// topic or level, does not name — so it is said here, and so is the point a
// topic chosen for the lessons is set at.
func rationale(goal profile.Goal, because string, corridor *rating.Corridor, choice *Choice, chosen string, point rating.Point) string {
	rule := fmt.Sprintf("Rule: %s, so %s. Difficulty %d of grades %s is %s.",
		because, goal, corridor.Recommended.Difficulty, corridor.Recommended.GradeLevel, fitText[corridor.Fit])
	// Where the model named a topic or a level and left the rest to the
	// corridor, its own sentence says the point the task came to.
	filledIn := (choice.GradeLevel == "" || choice.Difficulty == 0) && (choice.Topic != "" || choice.GradeLevel != "")
	if chosen != "" {
		// The rule's account below is of its own topic, so the point the task
		// stands at on the chosen one is said here, unless the model's sentence
		// says it.
		kept := fmt.Sprintf("The lessons are kept to topic %s, which the child or the adult chose.", chosen)
		switch {
		case !choice.Made():
			kept += fmt.Sprintf(" That is difficulty %d of grades %s, its corridor's.", point.Difficulty, point.GradeLevel)
		case !filledIn:
			kept += fmt.Sprintf(" The task is difficulty %d of grades %s.", point.Difficulty, point.GradeLevel)
		}
		rule = kept + " " + rule
	}
	if !choice.Made() {
		return rule
	}

	asked := make([]string, 0, 3)
	if choice.Topic != "" {
		asked = append(asked, "topic "+choice.Topic)
	}
	if choice.GradeLevel != "" {
		asked = append(asked, fmt.Sprintf("grades %s", choice.GradeLevel))
	}
	if choice.Difficulty != 0 {
		asked = append(asked, fmt.Sprintf("difficulty %d", choice.Difficulty))
	}

	told := ended(fmt.Sprintf("Model asked for %s: %s", strings.Join(asked, " and "), profile.Typed(choice.Reason)))
	if filledIn {
		told += fmt.Sprintf(" That is difficulty %d of grades %s, its corridor filling in the rest.",
			point.Difficulty, point.GradeLevel)
	}
	return told + " " + rule
}

// ended is a sentence with the mark that ends it: its own, when the model wrote
// one in whatever script — 。 and ؟ and ។ as much as a full stop — or a full
// stop. Unicode names the marks that end a sentence, all but the ellipsis and
// the Tibetan shad, which end one too.
func ended(sentence string) string {
	last, _ := utf8.DecodeLastRuneInString(sentence)
	if unicode.Is(unicode.Sentence_Terminal, last) || last == '…' || last == '།' {
		return sentence
	}
	return sentence + "."
}
