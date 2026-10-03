package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// issueTypes are the types of issue the self-check's format has.
var issueTypes = []string{
	"ambiguous", "missing_data", "multiple_correct", "no_correct",
	"too_hard_for_grade", "needs_picture", "factual_error",
}

// answerOperators are the defects of the answer, the solver and the
// self-check: classes D01 to D08.
func answerOperators() []operator {
	disagrees := []checks.Code{checks.CodeSolverDisagrees}
	fails := []checks.Code{checks.CodeSolverError}
	return []operator{
		{ID: "D01a", Class: "D01", Kind: mechanism, Expected: disagrees, Apply: wrongKey, Valid: solverMissesKey},
		{ID: "D02a", Class: "D02", Kind: mechanism, Expected: []checks.Code{checks.CodeBadStructure, checks.CodeSolverDisagrees}, Apply: rightAnswerTwice},
		{ID: "D02b", Class: "D02", Kind: discovery, Expected: []checks.Code{checks.CodeBadStructure, checks.CodeSolverDisagrees}, Apply: rightAnswerInWords},
		{ID: "D03a", Class: "D03", Kind: mechanism, Expected: disagrees, Apply: noRightOption, Valid: solverFindsNothing},
		{ID: "D04a", Class: "D04", Kind: mechanism, Expected: fails, Apply: intoSolve("injected = 1 // 0")},
		{ID: "D04b", Class: "D04", Kind: mechanism, Expected: fails, Apply: unparsedSolver},
		{ID: "D04c", Class: "D04", Kind: mechanism, Expected: fails, Apply: unnamedSolver},
		{ID: "D04d", Class: "D04", Kind: mechanism, Expected: fails, Apply: intoSolve(`return "A"`)},
		{ID: "D04e", Class: "D04", Kind: mechanism, Expected: fails, Apply: intoSolve(`injected = "%5d" % 1`)},
		{ID: "D05a", Class: "D05", Kind: mechanism, Expected: fails, Apply: intoSolve("for injected in range(100000000):", "    pass")},
		{ID: "D06a", Class: "D06", Kind: mechanism, Expected: disagrees, Apply: letterWrittenOut},
		{ID: "D07a", Class: "D07", Kind: mechanism, Expected: disagrees, Apply: selfCheckElsewhere},
		{ID: "D07b", Class: "D07", Kind: mechanism, Expected: disagrees, Apply: selfCheckUnsolvable},
		{ID: "D08a", Class: "D08", Kind: mechanism, Expected: []checks.Code{checks.CodeSelfCheckBlocking}, Apply: blockingIssue},
	}
}

// wrongKey moves the key to a wrong option, and the explanation that stood
// there to the old key's letter, so that the task agrees with itself and only
// the solver can tell.
func wrongKey(m *maker) (mutant, bool) {
	sub := m.start()
	old, key := sub.Task.CorrectAnswer, m.wrongLetter()
	sub.Task.Distractors[old] = sub.Task.Distractors[key]
	delete(sub.Task.Distractors, key)
	sub.Task.CorrectAnswer = key
	sub.SelfCheck.FinalAnswer = key
	return mutant{Sub: sub}, true
}

// rightAnswerTwice writes the right answer a second time, in place of a wrong
// option, in another form the service reads as the same answer.
func rightAnswerTwice(m *maker) (mutant, bool) {
	sub := m.start()
	sub.Task.Options[m.wrongLetter()] = sameAnswer(sub.Task.Options[sub.Task.CorrectAnswer])
	return mutant{Sub: sub}, true
}

var (
	wholeKey   = regexp.MustCompile(`^-?\d+$`)
	decimalKey = regexp.MustCompile(`^-?\d+\.\d+$`)
)

// sameAnswer writes an answer another way that the service reads as the same
// one: a whole number with ".0", a decimal with one more zero, a text with the
// case of its first letter swapped, and anything else with a space after it.
func sameAnswer(text string) string {
	key := solver.Key(text)
	switch {
	case wholeKey.MatchString(key):
		return text + ".0"
	case decimalKey.MatchString(key):
		return text + "0"
	}
	at := strings.IndexFunc(text, unicode.IsLetter)
	if at < 0 {
		return text + " "
	}
	letter, size := utf8.DecodeRuneInString(text[at:])
	swapped := unicode.ToUpper(letter)
	if unicode.IsUpper(letter) {
		swapped = unicode.ToLower(letter)
	}
	return text[:at] + string(swapped) + text[at+size:]
}

// rightAnswerInWords writes the right answer, a small whole number, in words in
// place of a wrong option: the same answer to a child, another one to a solver
// that compares texts.
func rightAnswerInWords(m *maker) (mutant, bool) {
	sub := m.start()
	words, small := inWords(sub.Task.Options[sub.Task.CorrectAnswer])
	if !small {
		return mutant{}, false
	}
	sub.Task.Options[m.wrongLetter()] = words
	return mutant{Sub: sub}, true
}

// noRightOption replaces the right option with a wrong one and keeps the key.
func noRightOption(m *maker) (mutant, bool) {
	sub := m.start()
	key := sub.Task.CorrectAnswer
	replacement, found := m.missingAnswer(&sub)
	if !found {
		return mutant{}, false
	}
	sub.Task.Options[key] = replacement
	return mutant{Sub: sub}, true
}

// missingAnswer is a text to put where the right answer stood: for a number,
// the nearest of n + 1, n − 1, n + 2, n − 2 and n + 3 that no option says
// already; for anything else, a wrong option of another task of the topic
// that no option says already.
func (m *maker) missingAnswer(sub *submission) (string, bool) {
	said := make(map[string]bool, solver.Count)
	for _, text := range sub.Task.Options {
		said[solver.Key(text)] = true
	}
	right := solver.Key(sub.Task.Options[sub.Task.CorrectAnswer])
	if wholeKey.MatchString(right) || decimalKey.MatchString(right) {
		return nearbyNumber(right, said)
	}
	for _, other := range m.pool.donors(m.host, func(other *host) bool { return other.Topic == m.host.Topic }) {
		if text, found := unsaidWrongOption(other, said); found {
			return text, true
		}
	}
	return "", false
}

// nearbyNumber is the nearest of n + 1, n − 1, n + 2, n − 2 and n + 3 that no
// option says already.
func nearbyNumber(number string, said map[string]bool) (string, bool) {
	value, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return "", false
	}
	for _, step := range []float64{1, -1, 2, -2, 3} {
		candidate := strconv.FormatFloat(value+step, 'f', -1, 64)
		if !said[solver.Key(candidate)] {
			return candidate, true
		}
	}
	return "", false
}

// unsaidWrongOption is a wrong option of another task that no option says
// already.
func unsaidWrongOption(other *host, said map[string]bool) (string, bool) {
	for _, letter := range solver.Letters() {
		text := other.Base.Task.Options[letter]
		if letter != other.Base.Task.CorrectAnswer && !said[solver.Key(text)] {
			return text, true
		}
	}
	return "", false
}

// solverMissesKey confirms a moved key with the host's own solver: it still
// arrives at the old answer, not at the key.
func solverMissesKey(ctx context.Context, runner solver.Runner, m *mutant) (bool, error) {
	agreement, err := solver.Verdict(ctx, runner, m.Sub.Solver, m.Sub.Options())
	if err != nil {
		return false, fmt.Errorf("confirm a moved key: %w", err)
	}
	return agreement.Letter != m.Sub.Task.CorrectAnswer, nil
}

// solverFindsNothing confirms a removed right option with the host's own
// solver: it finds no option at all. A replacement it accepts is a second
// right answer rather than a missing one.
func solverFindsNothing(ctx context.Context, runner solver.Runner, m *mutant) (bool, error) {
	agreement, err := solver.Verdict(ctx, runner, m.Sub.Solver, m.Sub.Options())
	if err != nil {
		return false, fmt.Errorf("confirm a removed answer: %w", err)
	}
	first := agreement.Runs[0]
	return first.Status == solver.StatusOK && len(first.Letters) == 0, nil
}

// intoSolve inserts lines as the first statement of the solver's solve, each
// indented as the body is and by its own indentation besides.
func intoSolve(lines ...string) func(m *maker) (mutant, bool) {
	return func(m *maker) (mutant, bool) {
		sub := m.start()
		source, inserted := insertIntoSolve(sub.Solver, lines)
		if !inserted {
			return mutant{}, false
		}
		sub.Solver = source
		return mutant{Sub: sub}, true
	}
}

// insertIntoSolve puts lines at the top of the body of solve, and says whether
// it found one.
func insertIntoSolve(source string, lines []string) (string, bool) {
	all := strings.Split(source, "\n")
	for i, line := range all {
		if !strings.HasPrefix(line, "def solve(") {
			continue
		}
		indent, found := bodyIndent(all[i+1:])
		if !found {
			return source, false
		}
		inserted := make([]string, 0, len(all)+len(lines))
		inserted = append(inserted, all[:i+1]...)
		for _, added := range lines {
			inserted = append(inserted, indent+added)
		}
		inserted = append(inserted, all[i+1:]...)
		return strings.Join(inserted, "\n"), true
	}
	return source, false
}

// bodyIndent is the indentation of the first line of a body that holds code.
func bodyIndent(lines []string) (string, bool) {
	for _, line := range lines {
		code := strings.TrimLeft(line, " \t")
		if code == "" || strings.HasPrefix(code, "#") {
			continue
		}
		return line[:len(line)-len(code)], len(code) < len(line)
	}
	return "", false
}

// unparsedSolver ends the solver with a line that does not parse.
func unparsedSolver(m *maker) (mutant, bool) {
	sub := m.start()
	sub.Solver += "\nsolve(\n"
	return mutant{Sub: sub}, true
}

// unnamedSolver renames solve, so that there is nothing to call.
func unnamedSolver(m *maker) (mutant, bool) {
	sub := m.start()
	if !strings.Contains(sub.Solver, "def solve(") {
		return mutant{}, false
	}
	sub.Solver = strings.Replace(sub.Solver, "def solve(", "def solve_injected(", 1)
	return mutant{Sub: sub}, true
}

// letterWrittenOut makes the solver return the key's letter as written,
// rather than work the answer out.
func letterWrittenOut(m *maker) (mutant, bool) {
	return intoSolve(fmt.Sprintf("return [%q]", m.host.Base.Task.CorrectAnswer))(m)
}

// selfCheckElsewhere has the self-check arrive at a wrong option.
func selfCheckElsewhere(m *maker) (mutant, bool) {
	sub := m.start()
	sub.SelfCheck.FinalAnswer = m.wrongLetter()
	return mutant{Sub: sub}, true
}

// selfCheckUnsolvable has the self-check find the task unsolvable.
func selfCheckUnsolvable(m *maker) (mutant, bool) {
	sub := m.start()
	sub.SelfCheck.FinalAnswer = "UNSOLVABLE"
	return mutant{Sub: sub}, true
}

// blockingIssue has the self-check report an issue it calls blocking.
func blockingIssue(m *maker) (mutant, bool) {
	sub := m.start()
	sub.SelfCheck.Issues = append(sub.SelfCheck.Issues, checks.Issue{Type: choose(m, issueTypes), Severity: "blocking", Comment: "Injected."})
	return mutant{Sub: sub}, true
}
