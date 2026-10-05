package checks

import (
	"fmt"
	"unicode"
)

// lettering is a script as the language check holds a text to it: the letters
// it is written in, what a refusal calls them, and whether the script the tag
// library guesses a tag stands for is enough to hold a task to it.
type lettering struct {
	letters       []*unicode.RangeTable
	called        string
	guessIsEnough bool
}

// The letterings of the languages written in more than one script of their
// own, whichever of them a tag names. A guess at Chinese is enough: whichever
// of Simplified and Traditional the library guesses, the characters are one
// set. Japanese is written in kanji and kana together, and Korean in Hangul
// with the Han characters it borrows.
var (
	chinese = lettering{
		letters: []*unicode.RangeTable{unicode.Han}, called: "Chinese characters", guessIsEnough: true,
	}
	japanese = lettering{
		letters: []*unicode.RangeTable{unicode.Han, unicode.Hiragana, unicode.Katakana}, called: "kanji and kana",
	}
	korean = lettering{letters: []*unicode.RangeTable{unicode.Hangul, unicode.Han}, called: "Hangul"}
)

// letterings are the scripts the texts of a task are held to, by the codes a
// language tag names them with: the scripts of every language the cards speak,
// under every code a tag of one of them may carry. A task in a language written
// in a script left out is held to none.
var letterings = map[string]lettering{
	"Latn": {letters: []*unicode.RangeTable{unicode.Latin}, called: "Latin letters"},
	"Cyrl": {letters: []*unicode.RangeTable{unicode.Cyrillic}, called: "Cyrillic letters"},
	"Arab": {letters: []*unicode.RangeTable{unicode.Arabic}, called: "Arabic letters"},
	"Deva": {letters: []*unicode.RangeTable{unicode.Devanagari}, called: "Devanagari letters"},
	"Beng": {letters: []*unicode.RangeTable{unicode.Bengali}, called: "Bengali letters"},
	"Thai": {letters: []*unicode.RangeTable{unicode.Thai}, called: "Thai letters"},
	"Hans": chinese, "Hant": chinese, "Hani": chinese,
	"Jpan": japanese, "Hira": japanese, "Kana": japanese, "Hrkt": japanese,
	"Kore": korean, "Hang": korean,
}

// readText is one text of a task the child reads, or several held to the
// letters together, under the name a refusal points at it by.
type readText struct {
	field  string
	plural bool
	texts  []string
}

// Language checks that the texts a child reads are written in the letters of
// the lesson's language: the question, the hint and the solution each on its
// own, so that one left untranslated is found, and the explanations behind the
// wrong options together, since each is a few words that two names in other
// letters could outweigh. The options are not held to it: an option may be a
// label, a name or a number in any language.
//
// A text is counted as a child reads it — a word at a time, or a character at
// a time in a script written without spaces — and is refused when less than
// half of what is counted is in the lesson's letters. A model that wrote in
// another language wrote next to none of them, and a task that names its
// points in Latin capitals or its children Tom and Mary still wrote most of
// it in them; half is let through, and a text with nothing to count says
// nothing. Letters are counted rather than a language guessed at, because a
// guess about twenty words with names and numbers in them is sometimes wrong,
// and a wrong refusal costs the child an attempt. So two languages written in
// the same letters are not told apart.
//
// A task in a language written in no script the check knows, or in several
// with none named by its tag, is held to nothing.
func Language(task *Task, language string) []Problem {
	if task == nil {
		return nil
	}
	script, held := heldScript(language)
	if !held {
		return nil
	}
	lesson, known := letterings[script]
	if !known {
		return nil
	}

	var astray []readText
	for _, read := range textsRead(task) {
		if mostlyElsewhere(read.texts, lesson.letters) {
			astray = append(astray, read)
		}
	}
	if len(astray) == 0 {
		return nil
	}
	return []Problem{writtenElsewhere(astray, lesson.called, language)}
}

// writtenElsewhere is the refusal of texts written mostly in other letters
// than the lesson's: it names them, the language and its letters, and says
// what to write, quoting nothing of the task.
func writtenElsewhere(astray []readText, letters, language string) Problem {
	fields := make([]string, len(astray))
	for i, read := range astray {
		fields[i] = read.field
	}
	verb := "is"
	if len(astray) > 1 || astray[0].plural {
		verb = "are"
	}
	return Problem{Code: CodeWrongLanguage, Message: fmt.Sprintf(
		"%s %s mostly not in %s, which %s, the language of the lesson, is written in. "+
			"Write every text the child reads in %s: the question, the options, the hint, the solution and the "+
			"explanations, and the names in them. The reference tasks are in English whatever the language of the "+
			"lesson; the task is not. Labels in Latin capitals, numbers and short symbols in small Latin letters, "+
			"such as x, ab or cm, are not counted.",
		inWords(fields, 0), verb, letters, language, language)}
}

// textsRead are the texts of a task the child reads that are held to the
// lesson's letters, in the order a refusal names them.
func textsRead(task *Task) []readText {
	explanations := make([]string, 0, len(task.Distractors))
	for _, distractor := range task.Distractors {
		explanations = append(explanations, distractor.Text)
	}
	return []readText{
		{field: "task.question", texts: []string{task.Question}},
		{field: "task.hint", texts: []string{task.Hint}},
		{field: "task.solution", texts: []string{task.Solution}},
		{field: "the explanations in task.distractors", plural: true, texts: explanations},
	}
}

// mostlyElsewhere says whether less than half of what is counted in these
// texts, taken together, is in these letters. Each text is counted on its
// own, so that the end of one and the start of the next are never read as one
// word.
func mostlyElsewhere(texts []string, letters []*unicode.RangeTable) bool {
	var own, all int
	for _, text := range texts {
		inLetters, read := counted(text, letters)
		own, all = own+inLetters, all+read
	}
	return 2*own < all
}

// counted is how much of a text a child reads, and how much of that is in these
// letters, counted as a child reads it: a character of a script written without
// spaces, a word of any other. A word is cut where its letters change script,
// so that a Latin name with a Cyrillic ending is a name and an ending. What is
// read in no language is left out: numbers and signs, the letters every script
// shares, such as the Japanese mark of a long vowel, words of Latin capitals
// alone, which are labels such as AB, short words in small Latin letters, which
// are symbols such as x, ab and cm, and single Greek letters such as π. A mark
// is read with the letter it is written over.
func counted(text string, letters []*unicode.RangeTable) (own, all int) {
	reading := tally{letters: letters}
	for _, r := range text {
		reading.read(r)
	}
	reading.endWord()
	return reading.own, reading.all
}

// tally is a text being counted, a character at a time.
type tally struct {
	letters  []*unicode.RangeTable
	word     []rune
	own, all int
}

// read takes the next character of the text.
func (t *tally) read(r rune) {
	switch {
	case unicode.IsMark(r):
		// A mark belongs to the letter before it, and neither counts nor ends
		// the word that letter is in.
	case !unicode.IsLetter(r) || unicode.In(r, unicode.Common, unicode.Inherited):
		t.endWord()
	case unicode.In(r, spaceless...):
		t.endWord()
		t.count(unicode.In(r, t.letters...))
	default:
		if len(t.word) > 0 && letterKind(t.word[0], t.letters) != letterKind(r, t.letters) {
			t.endWord()
		}
		t.word = append(t.word, r)
	}
}

// endWord counts the word read so far, unless it is a symbol, and starts the
// next one.
func (t *tally) endWord() {
	if len(t.word) > 0 && !symbol(t.word) {
		t.count(unicode.In(t.word[0], t.letters...))
	}
	t.word = t.word[:0]
}

// count counts one unit read, in the lesson's letters or not.
func (t *tally) count(inLetters bool) {
	t.all++
	if inLetters {
		t.own++
	}
}

// The kinds of letter a word is cut between.
const (
	kindLesson = iota
	kindLatin
	kindGreek
	kindOther
)

// letterKind is the kind of letter a character is, as words are cut between
// them: the lesson's own, Latin, Greek, or any other.
func letterKind(r rune, letters []*unicode.RangeTable) int {
	switch {
	case unicode.In(r, letters...):
		return kindLesson
	case unicode.Is(unicode.Latin, r):
		return kindLatin
	case unicode.Is(unicode.Greek, r):
		return kindGreek
	}
	return kindOther
}

// shortSymbol is the longest a word in small Latin letters is and still a
// symbol: a variable such as x, a product such as ab, the order abc, a unit
// such as cm or min. A word of another language that a child reads is longer,
// or opens its sentence with a capital.
const shortSymbol = 3

// symbol says whether a word is read in no language: Latin capitals alone, a
// label such as AB; a short word in small Latin letters, a symbol such as x,
// ab or cm; or a single Greek letter, a symbol such as π. Every letter of a
// word is of one kind, so its first tells which.
func symbol(word []rune) bool {
	switch {
	case unicode.Is(unicode.Latin, word[0]):
		return every(word, unicode.IsUpper) || len(word) <= shortSymbol && every(word, unicode.IsLower)
	case unicode.Is(unicode.Greek, word[0]):
		return len(word) == 1
	}
	return false
}

// every says whether every letter of a word is of a kind.
func every(word []rune, is func(rune) bool) bool {
	for _, r := range word {
		if !is(r) {
			return false
		}
	}
	return true
}
