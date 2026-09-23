package tutor

import (
	"fmt"
	"strings"

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
// apart afterwards rather than guessed at.
func rationale(goal profile.Goal, because string, corridor *rating.Corridor, choice Choice) string {
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

	said := choice.Reason
	if said == "" {
		said = "no reason given"
	}
	return fmt.Sprintf("Model asked for %s: %s. %s", strings.Join(asked, " and "), said, rule)
}
