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
	rating.FitTooHard: "the closest to the corridor, which holds no level; a little too hard",
	rating.FitTooEasy: "the closest to the corridor, which holds no level; a little too easy",
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
// apart afterwards rather than guessed at. A topic the model chose without a
// difficulty takes the one its own corridor recommends, which the rule's
// account, being of another topic, does not name — so it is said here.
func rationale(goal profile.Goal, because string, corridor *rating.Corridor, choice Choice, difficulty int) string {
	rule := fmt.Sprintf("Rule: %s, so %s. Difficulty %d is %s.",
		because, goal, corridor.Recommended, fitText[corridor.Fit])
	if !choice.made() {
		return rule
	}

	asked := make([]string, 0, 2)
	if choice.Topic != "" {
		asked = append(asked, "topic "+choice.Topic)
	}
	if choice.Difficulty != 0 {
		asked = append(asked, fmt.Sprintf("difficulty %d", choice.Difficulty))
	}

	said := strings.TrimSpace(choice.Reason)
	if said == "" {
		said = "no reason given"
	}
	told := ended(fmt.Sprintf("Model asked for %s: %s", strings.Join(asked, " and "), said))
	if choice.Topic != "" && choice.Difficulty == 0 {
		told += fmt.Sprintf(" Difficulty %d is the one its corridor recommends.", difficulty)
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
