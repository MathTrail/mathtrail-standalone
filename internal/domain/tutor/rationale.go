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
// says. When the model chose instead, it keeps both accounts: what the model
// asked for and what the rule would have done, so that the two can be told
// apart afterwards rather than guessed at. Where the model named a topic or a
// level and left the rest to the corridor, the task it gets is one the rule's
// account, being of another topic or level, does not name — so it is said
// here.
func rationale(goal profile.Goal, because string, corridor *rating.Corridor, choice *Choice, point rating.Point) string {
	rule := fmt.Sprintf("Rule: %s, so %s. Difficulty %d of grades %s is %s.",
		because, goal, corridor.Recommended.Difficulty, corridor.Recommended.GradeLevel, fitText[corridor.Fit])
	if !choice.made() {
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

	said := strings.TrimSpace(choice.Reason)
	if said == "" {
		said = "no reason given"
	}
	told := ended(fmt.Sprintf("Model asked for %s: %s", strings.Join(asked, " and "), said))
	filledIn := choice.GradeLevel == "" || choice.Difficulty == 0
	if filledIn && (choice.Topic != "" || choice.GradeLevel != "") {
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
