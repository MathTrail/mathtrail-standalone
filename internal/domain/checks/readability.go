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
// sentence of it is longer than the child's level allows, and — for a question
// in English, and only then — that its Flesch–Kincaid grade is no more than
// three above the child's. The language of the task decides the unit — the
// script of the whole question, when the task came with none — and every
// sentence is counted and held to the limit in that one unit: a Chinese
// sentence naming its points in Latin letters is still Chinese.
//
// Every sentence that is too long is named by its place, so that the model can
// find each without being shown the draft back.
func Readability(question, language string, grade int) []Problem {
	var problems []Problem

	limits := ReadabilityLimitsFor(grade)
	unit, allowed := unitFor(question, language), limits.SentenceWords
	if unit == unitCharacters {
		allowed = limits.SentenceCharacters
	}
	for i, sentence := range sentences(question) {
		if length := lengthIn(sentence, unit); length > allowed {
			problems = append(problems, Problem{Code: CodeReadability, Message: fmt.Sprintf(
				"sentence %d of task.question is %s long, and a child in grade %d reads sentences of at most %d %s; "+
					"split it or say it in fewer words", i+1, quantity(length, unit), grade, allowed, unit)})
		}
	}

	if english(language) {
		if index, most := fleschKincaid(question), limits.FleschKincaid; index > float64(most) {
			problems = append(problems, Problem{Code: CodeReadability, Message: fmt.Sprintf(
				"task.question reads at grade %.1f by Flesch–Kincaid, and a child in grade %d is given at most "+
					"grade %d; use shorter words and shorter sentences", index, grade, most)})
		}
	}
	return problems
}

// ReadabilityLimits are what the question of a task for a child of one grade
// is held to.
type ReadabilityLimits struct {
	// SentenceWords is the longest a sentence may be where words are separated
	// by spaces.
	SentenceWords int
	// SentenceCharacters is the longest a sentence may be where they are not.
	SentenceCharacters int
	// FleschKincaid is the highest Flesch–Kincaid grade a question in English
	// may read at.
	FleschKincaid int
}

// ReadabilityLimitsFor are the limits a question for a child of this grade is
// held to, in one place for the check and for whoever tells the model them. A
// grade outside one to six is held to the sentences of the nearest level:
// letting a sentence through for want of a limit would be the worse mistake.
func ReadabilityLimitsFor(grade int) ReadabilityLimits {
	level := sentenceLimits[(min(max(grade, 1), 6)-1)/2]
	return ReadabilityLimits{
		SentenceWords:      level.words,
		SentenceCharacters: level.characters,
		FleschKincaid:      grade + gradeMargin,
	}
}

// english says whether a language tag's primary subtag is English: "en",
// "en-GB", "EN_us". The formula counts syllables the way English spells them,
// and in any other language it measures nothing.
func english(language string) bool { return primarySubtag(language) == "en" }

// sentenceEnds are the marks that end a sentence, in the scripts a task may be
// written in.
const sentenceEnds = ".!?…。！？؟۔।॥።፨។៕။།༎"

// closers are what may stand between a sentence's last mark and the space
// after it: closing quotes and brackets.
const closers = "\"'”’»)]}」』）"

// wideEnds are the marks that end a sentence alone, whatever follows them: the
// marks of scripts whose punctuation is not followed by a space, and those of
// Khmer, Myanmar and Tibetan, which end a sentence whether or not a space is
// written after them.
const wideEnds = "。！？។៕။།༎"

// spacedEnds are the scripts that end a sentence with a space rather than a
// mark. Thai and Lao put no space between words, so a space between two of
// their letters is where a sentence, or a clause read as one, ends.
var spacedEnds = []*unicode.RangeTable{unicode.Thai, unicode.Lao}

// sentences splits a text into its sentences. A run of ending marks, with any
// closing quotes or brackets after it, ends a sentence when whitespace or the
// end of the text follows — or at once, for the marks that end one alone. In
// Thai and Lao, a space between two letters ends one too.
func sentences(text string) []string {
	runes := []rune(text)
	var found []string
	start := 0
	for i := 0; i < len(runes); i++ {
		end, next, ends := sentenceEnd(runes, i)
		if !ends {
			// Past a run of marks that ends nothing at once: looked at again
			// from each of its marks, a long run would cost its length squared.
			i = max(i, next-1)
			continue
		}
		if sentence := strings.TrimSpace(string(runes[start:end])); sentence != "" {
			found = append(found, sentence)
		}
		start, i = next, next-1
	}
	if rest := strings.TrimSpace(string(runes[start:])); rest != "" {
		found = append(found, rest)
	}
	return found
}

// sentenceEnd says whether a sentence ends at the character at i: where it
// ends, where the next one begins, and whether it ends there at all. Where it
// does not, next is where to look again.
func sentenceEnd(runes []rune, i int) (end, next int, ends bool) {
	if unicode.IsSpace(runes[i]) {
		if next, ends = spacedEnd(runes, i); ends {
			return i, next, true
		}
		return 0, i + 1, false
	}
	if !strings.ContainsRune(sentenceEnds, runes[i]) {
		return 0, i + 1, false
	}
	end = i + 1
	for end < len(runes) && strings.ContainsRune(sentenceEnds, runes[end]) {
		end++
	}
	for end < len(runes) && strings.ContainsRune(closers, runes[end]) {
		end++
	}
	alone := strings.ContainsRune(wideEnds, runes[i])
	if !alone && end < len(runes) && !unicode.IsSpace(runes[end]) {
		return 0, end, false
	}
	return end, end, true
}

// spacedEnd says whether the whitespace at i stands between two letters of a
// script that ends a sentence with a space, and where the next sentence
// begins.
func spacedEnd(runes []rune, i int) (next int, ends bool) {
	if i == 0 || !unicode.In(runes[i-1], spacedEnds...) {
		return 0, false
	}
	next = i
	for next < len(runes) && unicode.IsSpace(runes[next]) {
		next++
	}
	return next, next < len(runes) && unicode.In(runes[next], spacedEnds...)
}
