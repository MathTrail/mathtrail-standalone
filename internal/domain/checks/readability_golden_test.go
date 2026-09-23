package checks

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goldenReadability is what the prototype measured of its 450 reference
// questions, with its own code (T16): the Flesch–Kincaid grade its library
// gave each, and the sentences its splitter found.
type goldenReadability struct {
	Examples []struct {
		ID         string  `json:"id"`
		GradeLevel string  `json:"grade_level"`
		FK         float64 `json:"fk_grade"`
		Longest    int     `json:"longest_sentence_words"`
		Sentences  int     `json:"sentences"`
	} `json:"examples"`
}

// measuredQuestion is one reference question together with what the prototype
// measured of it.
type measuredQuestion struct {
	id, level, question string
	fk                  float64
	longest, sentences  int
}

func readGoldenReadability(t *testing.T) []measuredQuestion {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "golden", "readability.json"))
	if err != nil {
		t.Fatalf("read the golden vectors: %v", err)
	}
	var golden goldenReadability
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse the golden vectors: %v", err)
	}
	questions := map[string]string{}
	for _, reference := range readReferenceQuestions(t) {
		questions[reference.ID] = reference.Question
	}

	measured := make([]measuredQuestion, 0, len(golden.Examples))
	for _, example := range golden.Examples {
		question, known := questions[example.ID]
		if !known {
			t.Fatalf("the golden vectors measure %s, and no reference task has that id", example.ID)
		}
		measured = append(measured, measuredQuestion{
			id: example.ID, level: example.GradeLevel, question: question,
			fk: example.FK, longest: example.Longest, sentences: example.Sentences,
		})
	}
	return measured
}

// The words and the sentences of the formula are counted exactly as the
// prototype's library counted them — only the syllables are estimated — and
// that is checkable: with the same words and sentences, the syllables the
// library must have counted come out of its grade as a whole number.
func TestTheFormulaCountsWordsAndSentencesAsThePrototype(t *testing.T) {
	t.Parallel()

	for _, reference := range readGoldenReadability(t) {
		words := float64(len(fkWords(reference.question)))
		sentences := float64(fkSentences(reference.question))
		syllables := (reference.fk + 15.59 - 0.39*words/sentences) * words / 11.8
		if math.Abs(syllables-math.Round(syllables)) > 1e-6 {
			t.Errorf("%s: %v words in %v sentences leave %.4f syllables, and a count is a whole number",
				reference.id, words, sentences, syllables)
		}
	}
}

// The syllables are an estimate, since the prototype's library looked them up
// in a pronouncing dictionary this service does not carry. What the estimate
// is held to is what the threshold was calibrated for: over the 450 reference
// questions the grade misses the prototype's by at most a quarter on average
// and leans neither way by more than a sixth, the verdict for each grade of a
// task's level agrees at least 97 times in 100, and the share of tasks passing
// at each grade moves by at most four points (SPEC 5.5).
func TestFleschKincaidStaysWithinItsTolerance(t *testing.T) {
	t.Parallel()

	compared := compareWithPrototype(readGoldenReadability(t))
	if average := compared.missed / float64(compared.questions); average > 0.25 {
		t.Errorf("the grade misses the prototype's by %.3f on average, want at most 0.25", average)
	}
	if lean := compared.leaned / float64(compared.questions); math.Abs(lean) > 0.15 {
		t.Errorf("the grade leans %+.3f from the prototype's on average, want at most 0.15 either way", lean)
	}
	if agreement := 1 - float64(compared.disagreements)/float64(compared.verdicts); agreement < 0.97 {
		t.Errorf("the verdict agrees with the prototype's in %.1f%% of %d, want at least 97%%",
			100*agreement, compared.verdicts)
	}
	for grade, count := range compared.passes {
		theirs, ours := percent(count[0], compared.tasks[grade]), percent(count[1], compared.tasks[grade])
		if math.Abs(ours-theirs) > 4 {
			t.Errorf("grade %d: %.1f%% of tasks pass, and %.1f%% passed for the prototype", grade, ours, theirs)
		}
	}
}

// comparison is how the estimated grade compares with the prototype's over the
// reference questions: how far it misses and which way it leans, and how often
// the verdict at each grade of a question's level comes out the same.
type comparison struct {
	questions, verdicts, disagreements int
	missed, leaned                     float64
	passes                             map[int][2]int // grade: the prototype's passes, ours
	tasks                              map[int]int
}

func compareWithPrototype(references []measuredQuestion) comparison {
	compared := comparison{passes: map[int][2]int{}, tasks: map[int]int{}}
	for _, reference := range references {
		ours := fleschKincaid(reference.question)
		compared.questions++
		compared.missed += math.Abs(ours - reference.fk)
		compared.leaned += ours - reference.fk
		for _, grade := range gradesOf(reference.level) {
			limit := float64(grade + gradeMargin)
			theirsPass, oursPass := reference.fk <= limit, ours <= limit
			compared.verdicts++
			if theirsPass != oursPass {
				compared.disagreements++
			}
			count := compared.passes[grade]
			count[0] += passed(theirsPass)
			count[1] += passed(oursPass)
			compared.passes[grade] = count
			compared.tasks[grade]++
		}
	}
	return compared
}

// passed counts a verdict that passed.
func passed(verdict bool) int {
	if verdict {
		return 1
	}
	return 0
}

// Where a question is split into sentences as the prototype split it, its
// longest sentence has as many tokens between spaces as the prototype counted,
// and as many words less only the tokens of punctuation alone — an ellipsis in
// "1 + 3 + ... + 99", which the prototype counted and a child does not read.
// SPEC 5.5 splits in one place the prototype did not, after a mark and its
// closing quote — "Ann says: 'Ben is a liar.' Ben says: …" — and there it
// only ever finds more sentences, and never a longer one.
func TestSentencesAreMeasuredAsThePrototypeMeasuredThem(t *testing.T) {
	t.Parallel()

	for _, reference := range readGoldenReadability(t) {
		found := sentences(reference.question)
		tokens, words := 0, 0
		for _, sentence := range found {
			tokens = max(tokens, len(strings.Fields(sentence)))
			length, _ := lengthOf(sentence)
			words = max(words, length)
		}

		sameSplit := len(found) == reference.sentences
		switch {
		case sameSplit && tokens != reference.longest:
			t.Errorf("%s: the longest sentence has %d tokens, and the prototype counted %d", reference.id, tokens, reference.longest)
		case words > tokens:
			t.Errorf("%s: %d words in a sentence of %d tokens", reference.id, words, tokens)
		case !sameSplit && !closesAfterAMark(reference.question):
			t.Errorf("%s: %d sentences, the prototype found %d, and nothing in it is where SPEC 5.5 splits differently",
				reference.id, len(found), reference.sentences)
		case !sameSplit && (len(found) < reference.sentences || tokens > reference.longest):
			t.Errorf("%s: %d sentences and the longest %d tokens, against the prototype's %d and %d",
				reference.id, len(found), tokens, reference.sentences, reference.longest)
		}
	}
}

// closesAfterAMark says whether a text has a sentence-ending mark followed by a
// closing quote or bracket and then a space — the one place SPEC 5.5 ends a
// sentence and the prototype did not.
func closesAfterAMark(text string) bool {
	runes := []rune(text)
	for i := 0; i+2 < len(runes); i++ {
		if strings.ContainsRune(sentenceEnds, runes[i]) && strings.ContainsRune(closers, runes[i+1]) && runes[i+2] == ' ' {
			return true
		}
	}
	return false
}

// gradesOf are the grades a level is taught at.
func gradesOf(level string) []int {
	switch level {
	case "1-2":
		return []int{1, 2}
	case "3-4":
		return []int{3, 4}
	default:
		return []int{5, 6}
	}
}

func percent(part, whole int) float64 { return 100 * float64(part) / float64(whole) }
