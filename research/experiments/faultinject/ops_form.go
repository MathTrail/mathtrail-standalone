package main

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The limits a defect is made to cross: a solution longer than a card holds,
// and a sentence at least this many words past the level's limit.
const (
	longestSolution = 2000
	wordsPastLimit  = 3
)

// longWords is a sentence of long words, which raises a question's
// Flesch–Kincaid grade without making any sentence long.
const longWords = "Consider every possibility systematically and comprehensively."

// formOperators are the defects of the format, the brief, the traps, the
// explanations and the wording: classes D09 to D13.
func formOperators() []operator {
	structure := []checks.Code{checks.CodeBadStructure}
	explanations := []checks.Code{checks.CodeDistractorExplanations}
	ops := make([]operator, 0, 32)
	for _, field := range requiredTexts {
		ops = append(ops, operator{ID: "D09a-" + field.name, Class: "D09", Kind: mechanism, Expected: structure, Apply: emptied(field.empty)})
	}
	ops = append(ops,
		operator{ID: "D09b", Class: "D09", Kind: mechanism, Expected: structure, Apply: optionRemoved},
		operator{ID: "D09c", Class: "D09", Kind: mechanism, Expected: structure, Apply: keyNotALetter},
		operator{ID: "D09d", Class: "D09", Kind: mechanism, Expected: structure, Apply: keyExplained},
		operator{ID: "D09e", Class: "D09", Kind: mechanism, Expected: structure, Apply: solutionTooLong},
		operator{ID: "D09f", Class: "D09", Kind: mechanism, Expected: structure, Apply: optionCheckGap},
		operator{ID: "D09g", Class: "D09", Kind: mechanism, Expected: structure, Apply: unknownIssueType},
		operator{ID: "D10a", Class: "D10", Kind: mechanism, Expected: structure, Apply: otherTopic},
		operator{ID: "D10b", Class: "D10", Kind: mechanism, Expected: structure, Apply: otherLevel},
		operator{ID: "D10c", Class: "D10", Kind: mechanism, Expected: structure, Apply: otherDifficulty},
		operator{ID: "D10d", Class: "D10", Kind: mechanism, Expected: structure, Apply: skillDropped},
		operator{ID: "D11a", Class: "D11", Kind: mechanism, Expected: structure, Apply: unknownTrap},
		operator{ID: "D12a", Class: "D12", Kind: mechanism, Expected: explanations, Apply: explained(fromSolution)},
		operator{ID: "D12b", Class: "D12", Kind: mechanism, Expected: explanations, Apply: explained(fromHint)},
		operator{ID: "D12c", Class: "D12", Kind: mechanism, Expected: explanations, Apply: explained(fromAnother)},
		operator{ID: "D12d", Class: "D12", Kind: mechanism, Expected: explanations, Apply: explained(fromCatalog)},
		operator{ID: "D12e", Class: "D12", Kind: mechanism, Expected: explanations, Apply: explained(cutShort)},
		operator{ID: "D12f", Class: "D12", Kind: mechanism, Expected: structure, Apply: explained(nothing)},
		operator{ID: "D13a", Class: "D13", Kind: mechanism, Expected: []checks.Code{checks.CodeReadability}, Apply: longSentence},
		operator{ID: "D13b", Class: "D13", Kind: mechanism, Expected: []checks.Code{checks.CodeReadability}, Apply: hardWords},
	)
	return ops
}

// requiredTexts are the texts the format requires, each with a way to empty it.
var requiredTexts = []struct {
	name  string
	empty func(s *submission)
}{
	{"question", func(s *submission) { s.Task.Question = "" }},
	{"hint", func(s *submission) { s.Task.Hint = "" }},
	{"solution", func(s *submission) { s.Task.Solution = "" }},
	{"core_idea", func(s *submission) { s.Task.CoreIdea = "" }},
	{"design_thought_process", func(s *submission) { s.Task.DesignThoughtProcess = "" }},
	{"setting", func(s *submission) { s.Brief.Setting = "" }},
	{"rationale", func(s *submission) { s.Brief.Rationale = "" }},
}

// emptied leaves one required text empty.
func emptied(empty func(s *submission)) func(m *maker) (mutant, bool) {
	return func(m *maker) (mutant, bool) {
		sub := m.start()
		empty(&sub)
		return mutant{Sub: sub}, true
	}
}

// optionRemoved takes away one wrong option and its explanation, leaving four.
func optionRemoved(m *maker) (mutant, bool) {
	sub := m.start()
	letter := m.wrongLetter()
	delete(sub.Task.Options, letter)
	delete(sub.Task.Distractors, letter)
	return mutant{Sub: sub}, true
}

// keyNotALetter names a key no option has.
func keyNotALetter(m *maker) (mutant, bool) {
	sub := m.start()
	sub.Task.CorrectAnswer = "F"
	return mutant{Sub: sub}, true
}

// keyExplained explains the right option as if it were wrong, with a wrong
// option's explanation.
func keyExplained(m *maker) (mutant, bool) {
	sub := m.start()
	sub.Task.Distractors[sub.Task.CorrectAnswer] = sub.Task.Distractors[m.wrongLetter()]
	return mutant{Sub: sub}, true
}

// solutionTooLong repeats the solution until it is longer than a card holds.
func solutionTooLong(m *maker) (mutant, bool) {
	sub := m.start()
	solution := sub.Task.Solution
	for utf8.RuneCountInString(sub.Task.Solution) <= longestSolution {
		sub.Task.Solution += " " + solution
	}
	return mutant{Sub: sub}, true
}

// optionCheckGap leaves one option without the self-check's reason for it.
func optionCheckGap(m *maker) (mutant, bool) {
	sub := m.start()
	delete(sub.SelfCheck.OptionCheck, choose(m, solver.Letters()))
	return mutant{Sub: sub}, true
}

// unknownIssueType has the self-check report an issue of a type the format
// does not have.
func unknownIssueType(m *maker) (mutant, bool) {
	sub := m.start()
	sub.SelfCheck.Issues = append(sub.SelfCheck.Issues, checks.Issue{Type: "injected", Severity: "minor", Comment: "Injected."})
	return mutant{Sub: sub}, true
}

// otherTopic hands back a brief for another topic of the catalog.
func otherTopic(m *maker) (mutant, bool) {
	sub := m.start()
	topics := slices.DeleteFunc(m.pool.content.TopicIDs(), func(topic string) bool { return topic == m.host.Topic })
	sub.Brief.TargetConcept = choose(m, topics)
	return mutant{Sub: sub}, true
}

// otherLevel hands back a brief for another level.
func otherLevel(m *maker) (mutant, bool) {
	sub := m.start()
	levels := slices.DeleteFunc(rating.GradeLevels(), func(level rating.GradeLevel) bool { return level == m.host.Level })
	sub.Brief.GradeLevel = choose(m, levels)
	return mutant{Sub: sub}, true
}

// otherDifficulty hands back a brief for another difficulty.
func otherDifficulty(m *maker) (mutant, bool) {
	sub := m.start()
	var difficulties []int
	for difficulty := 1; difficulty <= rating.Difficulties; difficulty++ {
		if difficulty != m.host.Difficulty {
			difficulties = append(difficulties, difficulty)
		}
	}
	sub.Brief.Difficulty = choose(m, difficulties)
	return mutant{Sub: sub}, true
}

// skillDropped has the request exclude a skill the brief handed back leaves
// out.
func skillDropped(m *maker) (mutant, bool) {
	sub := m.start()
	skills := m.pool.content.Skills()
	sub.Asked.ExcludedSkills = []string{choose(m, skills).ID}
	return mutant{Sub: sub}, true
}

// unknownTrap names a trap the catalog does not have.
func unknownTrap(m *maker) (mutant, bool) {
	sub := m.start()
	letter := m.wrongLetter()
	distractor := sub.Task.Distractors[letter]
	distractor.Trap = "injected_trap"
	sub.Task.Distractors[letter] = distractor
	return mutant{Sub: sub}, true
}

// explained replaces one explanation of a wrong option with what text makes of
// it.
func explained(text func(m *maker, sub *submission, letter string) string) func(m *maker) (mutant, bool) {
	return func(m *maker) (mutant, bool) {
		sub := m.start()
		letter := m.wrongLetter()
		distractor := sub.Task.Distractors[letter]
		distractor.Text = text(m, &sub, letter)
		sub.Task.Distractors[letter] = distractor
		return mutant{Sub: sub}, true
	}
}

func fromSolution(_ *maker, sub *submission, _ string) string {
	return firstWords(sub.Task.Solution, 6)
}

func fromHint(_ *maker, sub *submission, _ string) string { return firstWords(sub.Task.Hint, 6) }

func fromAnother(m *maker, sub *submission, letter string) string {
	others := slices.DeleteFunc(m.wrongLetters(), func(other string) bool { return other == letter })
	return sub.Task.Distractors[choose(m, others)].Text
}

func fromCatalog(m *maker, sub *submission, letter string) string {
	description, _ := m.pool.content.TrapDescription(sub.Task.Distractors[letter].Trap)
	return description
}

func cutShort(_ *maker, sub *submission, letter string) string {
	return firstWords(sub.Task.Distractors[letter].Text, 2)
}

func nothing(*maker, *submission, string) string { return "" }

// longSentence joins the first sentences of the question until one is at
// least a few words longer than the level allows, never joining the last.
func longSentence(m *maker) (mutant, bool) {
	sub := m.start()
	sentences := sentencesOf(sub.Task.Question)
	if len(sentences) < 3 {
		return mutant{}, false
	}
	allowed := checks.ReadabilityLimitsFor(m.host.Level).SentenceWords + wordsPastLimit
	joined, used := sentences[0], 1
	for used < len(sentences)-1 && wordsIn(joined) < allowed {
		next, joinable := joinSentences(joined, sentences[used])
		if !joinable {
			return mutant{}, false
		}
		joined, used = next, used+1
	}
	if wordsIn(joined) < allowed {
		return mutant{}, false
	}
	sub.Task.Question = strings.Join(append([]string{joined}, sentences[used:]...), " ")
	return mutant{Sub: sub}, true
}

// joinSentences makes two sentences one with ", and": the first loses its
// final mark and the second its capital. A sentence that ends inside quotes is
// not joined.
func joinSentences(first, second string) (string, bool) {
	last, size := utf8.DecodeLastRuneInString(first)
	if !strings.ContainsRune(sentenceMarks, last) {
		return "", false
	}
	return first[:len(first)-size] + ", and " + lowerFirst(second), true
}

// hardWords puts a sentence of long words before the question's last
// sentence. A case counts only where that takes the question past its level's
// Flesch–Kincaid limit, as the service measures it.
func hardWords(m *maker) (mutant, bool) {
	sub := m.start()
	question, placed := beforeLastSentence(sub.Task.Question, longWords)
	if !placed {
		return mutant{}, false
	}
	sub.Task.Question = question
	// The host passed every sentence limit, and the sentence added is six words
	// long, so any problem the check finds now is the Flesch–Kincaid grade.
	if len(checks.Readability(sub.Task.Question, language, m.host.Level)) == 0 {
		return mutant{}, false
	}
	return mutant{Sub: sub}, true
}

// beforeLastSentence puts a sentence of its own before a text's last sentence,
// and says whether the text had one.
func beforeLastSentence(text, sentence string) (string, bool) {
	sentences := sentencesOf(text)
	if len(sentences) == 0 {
		return text, false
	}
	last := len(sentences) - 1
	return strings.Join(slices.Concat(sentences[:last], []string{sentence}, sentences[last:]), " "), true
}
