package checks_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// words is a sentence of n plain words, easy enough that only its length can
// be held against it.
func words(n int) string {
	return strings.TrimSpace(strings.Repeat("a cat ", n/2)+strings.Repeat("a ", n%2)) + "."
}

// characters is a Chinese sentence of n characters.
func characters(n int) string { return strings.Repeat("猫", n) + "。" }

// The limits of each level, in one place for the check and for the package
// that tells the model them: a level that is none of the three is held to
// the youngest's.
func TestTheLimitsOfEachLevel(t *testing.T) {
	t.Parallel()

	for level, want := range map[rating.GradeLevel]checks.ReadabilityLimits{
		rating.Grades12: {SentenceWords: 20, SentenceCharacters: 40, FleschKincaid: 4},
		rating.Grades34: {SentenceWords: 25, SentenceCharacters: 50, FleschKincaid: 6},
		rating.Grades56: {SentenceWords: 30, SentenceCharacters: 60, FleschKincaid: 8},
		"":              {SentenceWords: 20, SentenceCharacters: 40, FleschKincaid: 4},
		"7-8":           {SentenceWords: 20, SentenceCharacters: 40, FleschKincaid: 4},
	} {
		if got := checks.ReadabilityLimitsFor(level); got != want {
			t.Errorf("ReadabilityLimitsFor(%q) = %+v, want %+v", level, got, want)
		}
	}
}

func TestTheLongestSentenceIsHeldToTheLevel(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		question string
		language string
		level    rating.GradeLevel
		refused  bool
	}{
		{"20 words at 1-2", words(20), "ru", rating.Grades12, false},
		{"21 words at 1-2", words(21), "ru", rating.Grades12, true},
		{"25 words at 3-4", words(25), "ru", rating.Grades34, false},
		{"26 words at 3-4", words(26), "ru", rating.Grades34, true},
		{"30 words at 5-6", words(30), "ru", rating.Grades56, false},
		{"31 words at 5-6", words(31), "ru", rating.Grades56, true},
		{"40 characters at 1-2", characters(40), "zh", rating.Grades12, false},
		{"41 characters at 1-2", characters(41), "zh", rating.Grades12, true},
		{"60 characters at 5-6", characters(60), "zh", rating.Grades56, false},
		{"61 characters at 5-6", characters(61), "zh", rating.Grades56, true},
		{"a level that is none is held to 1-2", words(21), "ru", "7-8", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.Readability(test.question, test.language, test.level)
			if (len(problems) > 0) != test.refused {
				t.Fatalf("Readability() = %v, want refused %v", problems, test.refused)
			}
			for _, problem := range problems {
				if problem.Code != checks.CodeReadability {
					t.Errorf("code = %q, want %q", problem.Code, checks.CodeReadability)
				}
			}
		})
	}
}

// Every sentence is counted in the unit of the language the task is written
// in, and held to the limit in that same unit: a sign quoted in Chinese in a
// Russian question is one word, a sentence of words in a Chinese question is
// its letters, and a Chinese sentence stays Chinese whether it names its points
// in Latin letters or its children by Latin names — however many of its
// letters they come to. Without a language, the letters of the question decide.
func TestEverySentenceIsCountedInTheUnitOfItsLanguage(t *testing.T) {
	t.Parallel()

	var heights []string
	for next := 'B'; next <= 'N'; next++ {
		heights = append(heights, string(next-1)+"比"+string(next)+"高")
	}
	for _, test := range []struct {
		name, question, language, want string
	}{
		{"a Chinese sign in a Russian question", words(20) + " " + words(20) + " " + characters(41), "ru", ""},
		{"a sentence of words in a Chinese question", characters(40) + characters(40) + words(21), "zh",
			"sentence 3 of task.question is 41 characters long, and a task of grades 1-2 has sentences of at most " +
				"40 characters"},
		{"a Chinese sentence naming its points in Latin letters", "谁最高？" + strings.Join(heights, "，") + "。", "zh",
			"sentence 2 of task.question is 52 characters long"},
		{"a Chinese sentence naming its children by Latin names", strings.Repeat("Tom给Mary一个苹果，", 4) + "谁多？", "zh",
			"sentence 1 of task.question is 50 characters long"},
		{"a Chinese sentence tagged by its dialect", strings.Repeat("Tom给Mary一个苹果，", 4) + "谁多？", "cmn-Hans",
			"sentence 1 of task.question is 50 characters long"},
		{"a question with no language, decided by its letters", characters(40) + characters(40) + words(21), "",
			"sentence 3 of task.question is 41 characters long"},
		{"a word that is no tag, decided by its letters", characters(40) + characters(40) + words(21), "Chinese",
			"sentence 3 of task.question is 41 characters long"},
		// A private tag is well formed and names nothing, whatever it spells:
		// letters of Chinese are read as Chinese, and words as words.
		{"a tag that names no language, decided by its letters", characters(40) + characters(40) + words(21), "x-zh",
			"sentence 3 of task.question is 41 characters long"},
		{"a tag that names no language, over a question in words", strings.Repeat("apples ", 10) + "apples.", "x-zh",
			""},
		// "und" says the language is not known, not that it is English.
		{"the language not determined, decided by its letters", characters(40) + characters(40) + words(21), "und",
			"sentence 3 of task.question is 41 characters long"},
		{"a tag that names only its script", strings.Repeat("Tom给Mary一个苹果，", 4) + "谁多？", "und-Hans",
			"sentence 1 of task.question is 50 characters long"},
		{"a tag that names only its place", strings.Repeat("Tom给Mary一个苹果，", 4) + "谁多？", "und-CN",
			"sentence 1 of task.question is 50 characters long"},
		{"several languages at once, decided by the letters", characters(40) + characters(40) + words(21), "mul",
			"sentence 3 of task.question is 41 characters long"},
		{"a script code that names none, decided by the letters", characters(40) + characters(40) + words(21), "und-Zyyy",
			"sentence 3 of task.question is 41 characters long"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.Readability(test.question, test.language, rating.Grades12)
			switch {
			case test.want == "" && len(problems) > 0:
				t.Errorf("Readability() = %v, want no problems", problems)
			case test.want != "" && !mentions(problems, test.want):
				t.Errorf("Readability() = %v, want a problem saying %q", problems, test.want)
			}
		})
	}
}

// A sentence ends at a mark that ends one, and at any closing quote or bracket
// after it — where a space follows, or at once in the scripts written without
// spaces — in every script a task may be written in.
func TestSentencesEndWhereTheirMarksSay(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question, language string
		level                    rating.GradeLevel
		refused                  bool
	}{
		{"a question mark ends a sentence", words(20)[:len(words(20))-1] + "? " + words(20), "ru", rating.Grades12, false},
		{"a decimal point does not", "Cut 2.5 metres of " + words(18), "ru", rating.Grades12, true},
		{"a mark and a closing quote end one", "Ann says: '" + words(16) + "' " + words(20), "ru", rating.Grades12, false},
		{"a full stop and a closing bracket end one", "(" + words(18) + ") " + words(20), "ru", rating.Grades12, false},
		{"an ellipsis ends one", words(20)[:len(words(20))-1] + "… " + words(20), "ru", rating.Grades12, false},
		{"Chinese sentences end without a space", characters(40) + characters(40), "zh", rating.Grades12, false},
		{"Chinese full-width question marks too", strings.Repeat("猫", 40) + "？" + characters(40), "zh", rating.Grades12, false},
		{"Japanese in three scripts is counted in characters", strings.Repeat("ネコがいます", 7) + "。", "ja", rating.Grades12, true},
		{"Arabic ends a question with its own mark", strings.Repeat("قط ", 20) + "؟ " + strings.Repeat("قط ", 20), "ar", rating.Grades12, false},
		{"French sets a question mark apart, and it is no word", strings.Repeat("chat ", 20) + "? " + strings.Repeat("chat ", 20), "fr", rating.Grades12, false},
		{"Hindi ends a sentence with a danda", strings.Repeat("बिल्ली ", 20) + "। " + strings.Repeat("बिल्ली ", 20), "hi", rating.Grades12, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.Readability(test.question, test.language, test.level)
			if (len(problems) > 0) != test.refused {
				t.Fatalf("Readability() = %v, want refused %v", problems, test.refused)
			}
		})
	}
}

// Every sentence that is too long is named by its place, not only the first,
// and nothing of the draft is quoted back.
func TestEveryLongSentenceIsNamedByItsPlace(t *testing.T) {
	t.Parallel()

	question := words(21) + " " + words(5) + " " + words(22)
	problems := checks.Readability(question, "ru", rating.Grades12)
	for _, want := range []string{"sentence 1 of task.question is 21 words long", "sentence 3 of task.question is 22 words long"} {
		if !mentions(problems, want) {
			t.Errorf("Readability() = %v, want a problem mentioning %q", problems, want)
		}
	}
	for _, problem := range problems {
		if strings.Contains(problem.Message, "cat") || letter.MatchString(problem.Message) {
			t.Errorf("%q quotes the draft", problem.Message)
		}
	}
}

// Flesch–Kincaid counts syllables the way English spells them, so it applies to
// a question in English — whatever its region — and to nothing else.
func TestFleschKincaidAppliesToEnglishAlone(t *testing.T) {
	t.Parallel()

	hard := "Every afternoon seven classmates exchange colourful postcards. Each classmate sends exactly one " +
		"postcard to every other classmate. How many postcards are delivered altogether?"
	for _, test := range []struct {
		language string
		refused  bool
	}{
		{"en", true}, {"en-GB", true}, {"EN_us", true},
		{"ru", false}, {"eng", false}, {"", false},
	} {
		t.Run(test.language, func(t *testing.T) {
			t.Parallel()

			problems := checks.Readability(hard, test.language, rating.Grades12)
			if mentions(problems, "Flesch–Kincaid") != test.refused {
				t.Fatalf("Readability(%q) = %v, want a Flesch–Kincaid refusal %v", test.language, problems, test.refused)
			}
		})
	}
}

// The prototype's own tests set eight questions against grades one to four, in
// English and in Russian. A task is held to its level at the limits of the
// level's youngest grade, so the verdicts compared are the prototype's for
// grades 1 and 3, one by one; those for grades 2 and 4 were for limits a grade
// looser, which no task is held to any more.
func TestTheReadabilityOfThePrototypesAnchors(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "golden", "readability.json"))
	if err != nil {
		t.Fatalf("read the golden vectors: %v", err)
	}
	var golden struct {
		Anchors []struct {
			Name    string          `json:"name"`
			Text    string          `json:"text"`
			English map[string]bool `json:"ok_by_grade_en"`
			Russian map[string]bool `json:"ok_by_grade_ru"`
		} `json:"anchors"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse the golden vectors: %v", err)
	}
	if len(golden.Anchors) == 0 {
		t.Fatal("the golden vectors hold no anchors")
	}

	for _, anchor := range golden.Anchors {
		for language, verdicts := range map[string]map[string]bool{"en": anchor.English, "ru": anchor.Russian} {
			for _, level := range []rating.GradeLevel{rating.Grades12, rating.Grades34} {
				grade := level.FirstGrade()
				want := verdicts[strconv.Itoa(grade)]
				problems := checks.Readability(anchor.Text, language, level)
				if passes := len(problems) == 0; passes != want {
					t.Errorf("%s in %s at %s: passes %v, the prototype said %v at grade %d (%v)",
						anchor.Name, language, level, passes, want, grade, problems)
				}
			}
		}
	}
}

// The question comes from the chat's model and is untrusted: whatever it holds
// and whatever level it is measured at — one of the three or none of them —
// the check ends in problems of its own code, never in a crash.
func FuzzReadability(f *testing.F) {
	f.Add("Four spaceships dock in pairs. How many different pairs can they make?", "en", "1-2")
	f.Add("小明有三个苹果。小红有五个梨？", "zh", "3-4")
	f.Add("'?!.'. . .\x00", "en-GB", "")
	f.Add("Ann says: 'Ben is a liar.' Ben says: 'Kim is a liar.'", "EN", "5-6")

	f.Fuzz(func(t *testing.T, question, language, level string) {
		for _, problem := range checks.Readability(question, language, rating.GradeLevel(level)) {
			if problem.Code != checks.CodeReadability || problem.Message == "" {
				t.Fatalf("problem = %+v, want a message under %q", problem, checks.CodeReadability)
			}
		}
	})
}

// A sentence ends where its own script ends one: at the marks of Khmer, Myanmar
// and Tibetan whether or not a space follows, and in Thai and Lao — which put
// no space between words — at a space between two letters. Its vowels and tones
// are marks written over a letter and count with it, not beside it. Two
// sentences of thirty letters are within a first-grader's forty; one of
// forty-one is not.
func TestSentencesEndWhereTheirScriptEndsThem(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question, language string
		refused                  bool
	}{
		{"Khmer", strings.Repeat("ក", 30) + "។" + strings.Repeat("ក", 30) + "។", "km", false},
		{"Myanmar", strings.Repeat("က", 30) + "။ " + strings.Repeat("က", 30) + "။", "my", false},
		{"Tibetan", strings.Repeat("ཀ", 30) + "།" + strings.Repeat("ཀ", 30) + "།", "bo", false},
		{"Thai", strings.Repeat("ก", 30) + " " + strings.Repeat("ก", 30), "th", false},
		{"Lao", strings.Repeat("ກ", 30) + " " + strings.Repeat("ກ", 30), "lo", false},
		{"Thai with its vowels and tones", strings.Repeat("กี่", 35), "th", false},
		{"Thai of one sentence too long", strings.Repeat("ก", 41), "th", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if problems := checks.Readability(test.question, test.language, rating.Grades12); (len(problems) > 0) != test.refused {
				t.Errorf("Readability() = %v, want refused %v", problems, test.refused)
			}
		})
	}
}

// A run of marks that ends no sentence is passed over once. Looked at again
// from each of its marks it would cost its length squared, and a question the
// model wrote can hold as long a run as it likes.
func TestALongRunOfMarksIsReadInOnePass(t *testing.T) {
	t.Parallel()

	question := strings.Repeat("!", 100_000) + "x"
	started := time.Now()
	checks.Readability(question, "ru", rating.Grades12)
	if took := time.Since(started); took > time.Second {
		t.Errorf("reading a run of %d marks took %v, want it read in one pass", len(question)-1, took)
	}
}
