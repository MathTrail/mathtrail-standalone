package checks_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// words is a sentence of n plain words, easy enough that only its length can
// be held against it.
func words(n int) string {
	return strings.TrimSpace(strings.Repeat("a cat ", n/2)+strings.Repeat("a ", n%2)) + "."
}

// characters is a Chinese sentence of n characters.
func characters(n int) string { return strings.Repeat("猫", n) + "。" }

func TestTheLongestSentenceIsHeldToTheLevel(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		question string
		grade    int
		refused  bool
	}{
		{"20 words at grade 1", words(20), 1, false},
		{"21 words at grade 2", words(21), 2, true},
		{"25 words at grade 3", words(25), 3, false},
		{"26 words at grade 4", words(26), 4, true},
		{"30 words at grade 5", words(30), 5, false},
		{"31 words at grade 6", words(31), 6, true},
		{"40 characters at grade 1", characters(40), 1, false},
		{"41 characters at grade 2", characters(41), 2, true},
		{"60 characters at grade 6", characters(60), 6, false},
		{"61 characters at grade 6", characters(61), 6, true},
		{"a grade below one is held to grade 1", words(21), 0, true},
		{"a grade above six is held to grade 6", words(30), 9, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.Readability(test.question, "ru", test.grade)
			if refused := len(problems) > 0; refused != test.refused {
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

// Sentences end where SPEC 5.5 says, in every script a task may be written in.
func TestSentencesEndWhereTheirMarksSay(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question, language string
		grade                    int
		refused                  bool
	}{
		{"a question mark ends a sentence", words(20)[:len(words(20))-1] + "? " + words(20), "ru", 1, false},
		{"a decimal point does not", "Cut 2.5 metres of " + words(18), "ru", 1, true},
		{"a mark and a closing quote end one", "Ann says: '" + words(16) + "' " + words(20), "ru", 1, false},
		{"a full stop and a closing bracket end one", "(" + words(18) + ") " + words(20), "ru", 1, false},
		{"an ellipsis ends one", words(20)[:len(words(20))-1] + "… " + words(20), "ru", 1, false},
		{"Chinese sentences end without a space", characters(40) + characters(40), "zh", 1, false},
		{"Chinese full-width question marks too", strings.Repeat("猫", 40) + "？" + characters(40), "zh", 1, false},
		{"Japanese in three scripts is counted in characters", strings.Repeat("ネコがいます", 7) + "。", "ja", 1, true},
		{"Arabic ends a question with its own mark", strings.Repeat("قط ", 20) + "؟ " + strings.Repeat("قط ", 20), "ar", 1, false},
		{"French sets a question mark apart, and it is no word", strings.Repeat("chat ", 20) + "? " + strings.Repeat("chat ", 20), "fr", 1, false},
		{"Hindi ends a sentence with a danda", strings.Repeat("बिल्ली ", 20) + "। " + strings.Repeat("बिल्ली ", 20), "hi", 1, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := checks.Readability(test.question, test.language, test.grade)
			if refused := len(problems) > 0; refused != test.refused {
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
	problems := checks.Readability(question, "ru", 1)
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

			problems := checks.Readability(hard, test.language, 1)
			if refused := mentions(problems, "Flesch–Kincaid"); refused != test.refused {
				t.Fatalf("Readability(%q) = %v, want a Flesch–Kincaid refusal %v", test.language, problems, test.refused)
			}
		})
	}
}

// The prototype's own tests set eight questions against grades one to four, in
// English and in Russian. The verdicts here are theirs, one by one.
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
			for grade := 1; grade <= 4; grade++ {
				want := verdicts[strconv.Itoa(grade)]
				problems := checks.Readability(anchor.Text, language, grade)
				if passes := len(problems) == 0; passes != want {
					t.Errorf("%s in %s at grade %d: passes %v, the prototype said %v (%v)",
						anchor.Name, language, grade, passes, want, problems)
				}
			}
		}
	}
}

// The question comes from the chat's model and is untrusted: whatever it holds
// and whatever grade it is measured for, the check ends in problems of its own
// code, never in a crash.
func FuzzReadability(f *testing.F) {
	f.Add("Four spaceships dock in pairs. How many different pairs can they make?", "en", 2)
	f.Add("小明有三个苹果。小红有五个梨？", "zh", 1)
	f.Add("'?!.'. . .\x00", "en-GB", 0)
	f.Add("Ann says: 'Ben is a liar.' Ben says: 'Kim is a liar.'", "EN", 99)

	f.Fuzz(func(t *testing.T, question, language string, grade int) {
		for _, problem := range checks.Readability(question, language, grade) {
			if problem.Code != checks.CodeReadability || problem.Message == "" {
				t.Fatalf("problem = %+v, want a message under %q", problem, checks.CodeReadability)
			}
		}
	})
}
