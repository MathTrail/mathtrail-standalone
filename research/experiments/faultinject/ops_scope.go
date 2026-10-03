package main

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// pseudonym is a name a child might choose, put where the task addresses them.
const pseudonym = "Bright-Owl, "

// opposites are the words a misreading turns into their opposite, and the
// negations it drops.
var opposites = map[string]string{
	"smallest": "largest", "largest": "smallest",
	"least": "most", "most": "least", "fewest": "most",
	"fewer": "more", "more": "fewer",
	"minimum": "maximum", "maximum": "minimum",
	"first": "last", "last": "first",
	"before": "after", "after": "before",
	"left": "right", "right": "left",
	"cannot": "can", "can": "cannot",
	"not": "",
}

var (
	oppositeWords = wordPattern(keysOf(opposites))
	exactly       = regexp.MustCompile(`(?i)\bexactly\b`)
	digit         = regexp.MustCompile(`\d`)
)

// scopeOperators are the defects the service is not built to see: classes X01
// to X08.
func scopeOperators() []operator {
	return []operator{
		{ID: "X01a", Class: "X01", Kind: outOfScope, Apply: reworded(madeVague)},
		{ID: "X01b", Class: "X01", Kind: outOfScope, Apply: reworded(madeRange)},
		{ID: "X02a", Class: "X02", Kind: outOfScope, Apply: reworded(conditionRemoved)},
		{ID: "X03a", Class: "X03", Kind: outOfScope, Apply: reworded(askedOpposite)},
		{ID: "X04a", Class: "X04", Kind: outOfScope, Apply: trapsSwapped},
		{ID: "X04b", Class: "X04", Kind: outOfScope, Apply: trapReplaced},
		{ID: "X05a", Class: "X05", Kind: outOfScope, Apply: hintGivesAway},
		{ID: "X06a", Class: "X06", Kind: outOfScope, Apply: reworded(addressed)},
		{ID: "X07a", Class: "X07", Kind: outOfScope, Apply: offTopic},
		{ID: "X08a", Class: "X08", Kind: outOfScope, Apply: offDifficulty},
	}
}

// reworded changes the host's question; a question the change finds nothing
// in is left out.
func reworded(change func(question string) (string, bool)) func(m *maker) (mutant, bool) {
	return func(m *maker) (mutant, bool) {
		sub := m.start()
		question, changed := change(sub.Task.Question)
		if !changed {
			return mutant{}, false
		}
		sub.Task.Question = question
		return mutant{Sub: sub}, true
	}
}

// madeVague turns the first "exactly" into "about" or, failing one, puts
// "about" before the first whole number.
func madeVague(question string) (string, bool) {
	if at := exactly.FindStringIndex(question); at != nil {
		return question[:at[0]] + withCaseOf("about", question[at[0]:at[1]]) + question[at[1]:], true
	}
	return replaceFirstNumber(question, func(n int) string { return fmt.Sprintf("about %d", n) })
}

// madeRange turns the first whole number n into "n or n + 1".
func madeRange(question string) (string, bool) {
	return replaceFirstNumber(question, func(n int) string { return fmt.Sprintf("%d or %d", n, n+1) })
}

// conditionRemoved drops the first sentence other than the last that holds a
// digit.
func conditionRemoved(question string) (string, bool) {
	sentences := sentencesOf(question)
	for i := range len(sentences) - 1 {
		if digit.MatchString(sentences[i]) {
			return rejoin(slices.Delete(sentences, i, i+1)), true
		}
	}
	return question, false
}

// askedOpposite turns the first word of the list into its opposite, or drops
// a negation, so that the question asks something else.
func askedOpposite(question string) (string, bool) {
	return swapFirstWord(question, oppositeWords, opposites)
}

// addressed puts the child's pseudonym before the question.
func addressed(question string) (string, bool) {
	return pseudonym + lowerFirst(question), true
}

// trapsSwapped swaps the traps of two explanations that name different ones.
func trapsSwapped(m *maker) (mutant, bool) {
	sub := m.start()
	letters := m.wrongLetters()
	var pairs [][2]string
	for i, first := range letters {
		for _, second := range letters[i+1:] {
			if sub.Task.Distractors[first].Trap != sub.Task.Distractors[second].Trap {
				pairs = append(pairs, [2]string{first, second})
			}
		}
	}
	if len(pairs) == 0 {
		return mutant{}, false
	}
	pair := choose(m, pairs)
	first, second := sub.Task.Distractors[pair[0]], sub.Task.Distractors[pair[1]]
	first.Trap, second.Trap = second.Trap, first.Trap
	sub.Task.Distractors[pair[0]], sub.Task.Distractors[pair[1]] = first, second
	return mutant{Sub: sub}, true
}

// trapReplaced names another trap of the catalog behind one explanation.
func trapReplaced(m *maker) (mutant, bool) {
	sub := m.start()
	letter := m.wrongLetter()
	distractor := sub.Task.Distractors[letter]
	others := slices.DeleteFunc(m.pool.content.TrapIDs(), func(trap string) bool { return trap == distractor.Trap })
	distractor.Trap = choose(m, others)
	sub.Task.Distractors[letter] = distractor
	return mutant{Sub: sub}, true
}

// hintGivesAway has the hint say the answer.
func hintGivesAway(m *maker) (mutant, bool) {
	sub := m.start()
	sub.Task.Hint = "The answer is " + strings.TrimSpace(sub.Task.Options[sub.Task.CorrectAnswer]) + "."
	return mutant{Sub: sub}, true
}

// offTopic sets the whole task of another topic of the level under the host's
// brief.
func offTopic(m *maker) (mutant, bool) {
	return m.swapTask(func(other *host) bool { return other.Level == m.host.Level && other.Topic != m.host.Topic })
}

// offDifficulty sets the whole task of the same topic and level, two or more
// difficulties away, under the host's brief.
func offDifficulty(m *maker) (mutant, bool) {
	return m.swapTask(func(other *host) bool {
		gap := other.Difficulty - m.host.Difficulty
		return other.Level == m.host.Level && other.Topic == m.host.Topic && (gap >= 2 || gap <= -2)
	})
}

// swapTask replaces the host's task, solver, key and traps with a donor's,
// keeping the host's brief and request, and leaves the donor's question out of
// the comparison: the case is about the brief, not about a copy.
func (m *maker) swapTask(keep func(other *host) bool) (mutant, bool) {
	donors := m.pool.donors(m.host, keep)
	if len(donors) == 0 {
		return mutant{}, false
	}
	donor := choose(m, donors)
	sub := m.start()
	sub.Task = reviewing.CloneTask(&donor.Base.Task)
	sub.Solver = donor.Base.Solver
	sub.SelfCheck.FinalAnswer = donor.Base.SelfCheck.FinalAnswer
	sub.Brief.TrapsToUse = slices.Clone(donor.Base.Brief.TrapsToUse)
	return mutant{Sub: sub, LeaveOut: []string{donor.Base.Task.Question}, Donor: donor.ID}, true
}
