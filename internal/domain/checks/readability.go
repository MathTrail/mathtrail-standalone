package checks

import (
	"fmt"
	"strings"
	"unicode"
)

// The longest sentence a child of each level is given, in words where words
// are separated by spaces and in characters where they are not. The word
// limits are the prototype's, measured on its reference tasks; the character
// limits are twice them by analogy, because nobody has measured them.
var sentenceLimits = [...]struct{ words, characters int }{
	{words: 20, characters: 40}, // grades 1–2
	{words: 25, characters: 50}, // grades 3–4
	{words: 30, characters: 60}, // grades 5–6
}

// gradeMargin is how many grades above the child's the Flesch–Kincaid index of
// an English question may reach: at one, fewer than half the reference tasks
// for grades 1–2 passed for a first-grader, and the index is noisy on texts
// this short.
const gradeMargin = 3

// Readability checks that a child of this grade can read the question: that no
// sentence of it is longer than the child's level allows, in the unit its
// script is counted in, and — for a question in English, and only then — that
// its Flesch–Kincaid grade is no more than three above the child's.
//
// Every sentence that is too long is named by its place, so that the model can
// find each without being shown the draft back.
func Readability(question, language string, grade int) []Problem {
	var problems []Problem

	allowed, characters := limitsAt(grade)
	if Spaceless(question) {
		allowed = characters
	}
	for i, sentence := range sentences(question) {
		length, unit := lengthOf(sentence)
		if length > allowed {
			problems = append(problems, Problem{Code: CodeReadability, Message: fmt.Sprintf(
				"sentence %d of task.question is %s long, and a child in grade %d reads sentences of at most %d %s; "+
					"split it or say it in fewer words", i+1, quantity(length, unit), grade, allowed, unit)})
		}
	}

	if english(language) {
		if index, most := fleschKincaid(question), grade+gradeMargin; index > float64(most) {
			problems = append(problems, Problem{Code: CodeReadability, Message: fmt.Sprintf(
				"task.question reads at grade %.1f by Flesch–Kincaid, and a child in grade %d is given at most "+
					"grade %d; use shorter words and shorter sentences", index, grade, most)})
		}
	}
	return problems
}

// limitsAt is the longest sentence a child of this grade is given, in words and
// in characters. A grade outside one to six is held to the nearest level:
// letting a sentence through for want of a limit would be the worse mistake.
func limitsAt(grade int) (words, characters int) {
	level := sentenceLimits[(min(max(grade, 1), 6)-1)/2]
	return level.words, level.characters
}

// english says whether a language tag's primary subtag is English: "en",
// "en-GB", "EN_us". The formula counts syllables the way English spells them,
// and in any other language it measures nothing.
func english(language string) bool {
	primary, _, _ := strings.Cut(strings.ToLower(language), "-")
	primary, _, _ = strings.Cut(primary, "_")
	return primary == "en"
}

// sentenceEnds are the marks that end a sentence, in the scripts a task may be
// written in.
const sentenceEnds = ".!?…。！？؟۔।॥።፨"

// closers are what may stand between a sentence's last mark and the space
// after it: closing quotes and brackets.
const closers = "\"'”’»)]}」』）"

// wideEnds are the marks of scripts whose punctuation is not followed by a
// space, so the mark alone ends the sentence.
const wideEnds = "。！？"

// sentences splits a text into its sentences. A run of ending marks, with any
// closing quotes or brackets after it, ends a sentence when whitespace or the
// end of the text follows — or at once, for the marks of scripts written
// without spaces.
func sentences(text string) []string {
	runes := []rune(text)
	var found []string
	start := 0
	for i := 0; i < len(runes); i++ {
		if !strings.ContainsRune(sentenceEnds, runes[i]) {
			continue
		}
		wide := strings.ContainsRune(wideEnds, runes[i])
		end := i + 1
		for end < len(runes) && strings.ContainsRune(sentenceEnds, runes[end]) {
			end++
		}
		for end < len(runes) && strings.ContainsRune(closers, runes[end]) {
			end++
		}
		if !wide && end < len(runes) && !unicode.IsSpace(runes[end]) {
			i = end - 1
			continue
		}
		if sentence := strings.TrimSpace(string(runes[start:end])); sentence != "" {
			found = append(found, sentence)
		}
		start, i = end, end-1
	}
	if rest := strings.TrimSpace(string(runes[start:])); rest != "" {
		found = append(found, rest)
	}
	return found
}
