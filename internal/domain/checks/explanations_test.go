package checks_test

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

func TestWellExplainedWrongOptionsPass(t *testing.T) {
	t.Parallel()

	if problems := checks.Explanations(validDraft(), "", testCatalog); len(problems) != 0 {
		t.Fatalf("Explanations() = %v, want none", problems)
	}
}

// explanationBreakages are the ways an explanation fails to tell a child what
// went wrong, each on its own, and a fragment of what the refusal must say.
var explanationBreakages = []breakage{
	{"two explanations the same", func(d *checks.Draft) {
		d.Task.Distractors["B"] = checks.Distractor{Trap: "missed_case", Text: "4 is the number of ships."}
	}, `2 explanations for traps "missed_case" and "number_from_text" say the same thing`},
	{"the same but for case, spacing and punctuation", func(d *checks.Draft) {
		d.Task.Distractors["B"] = checks.Distractor{Trap: "missed_case", Text: "  4 IS the number   of ships!!"}
	}, "say the same thing"},
	{"two explanations of one trap the same", func(d *checks.Draft) {
		d.Task.Distractors["B"] = checks.Distractor{Trap: "number_from_text", Text: "4 is the number of ships."}
	}, `2 explanations for trap "number_from_text" say the same thing`},
	{"three the same, two of them of one trap", func(d *checks.Draft) {
		d.Task.Distractors["B"] = checks.Distractor{Trap: "number_from_text", Text: "4 is the number of ships."}
		d.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: "4 is the number of ships."}
	}, `3 explanations for traps "number_from_text" and "wrong_operation" say the same thing`},
	{"two the same with no trap", func(d *checks.Draft) {
		d.Task.Distractors["A"] = checks.Distractor{Text: "You missed one pair."}
		d.Task.Distractors["B"] = checks.Distractor{Text: "You missed one pair."}
	}, "2 explanations with no trap say the same thing"},
	{"two the same, one of them with no trap", func(d *checks.Draft) {
		d.Task.Distractors["B"] = checks.Distractor{Text: "4 is the number of ships."}
	}, `2 explanations for trap "number_from_text" and with no trap say the same thing`},
	{"the solution itself", func(d *checks.Draft) {
		d.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: "List the pairs: 6."}
	}, `an explanation for trap "wrong_operation" repeats the start of task.solution`},
	{"the opening words of the solution", func(d *checks.Draft) {
		d.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: "List the pairs."}
	}, "repeats the start of task.solution"},
	{"the opening words of the hint", func(d *checks.Draft) {
		d.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: "How many ships can the first ship"}
	}, "repeats the start of task.hint"},
	{"the catalog's description of its trap", func(d *checks.Draft) {
		d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "counts the same thing twice"}
	}, `an explanation for trap "double_count" is the catalog's description of its trap`},
	{"one word", func(d *checks.Draft) {
		d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "Wrong."}
	}, `an explanation for trap "double_count" is 1 word long, and it takes at least 3 words`},
	{"punctuation and nothing else", func(d *checks.Draft) {
		d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "?!"}
	}, "is 0 words long"},
	{"two words in Russian", func(d *checks.Draft) {
		d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "Неверно, подумай."}
	}, "is 2 words long"},
	{"two characters in Chinese", func(d *checks.Draft) {
		d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "错了。"}
	}, "is 2 characters long, and it takes at least 6 characters"},
	{"the opening characters of a Chinese solution", func(d *checks.Draft) {
		d.Task.Solution = "小明有三个苹果和两个梨，一共五个。"
		d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "小明有三个苹果"}
	}, "repeats the start of task.solution"},
	{"too short with no trap to point at", func(d *checks.Draft) {
		d.Task.Distractors["E"] = checks.Distractor{Text: "No."}
	}, "an explanation with no trap is 1 word long"},
}

func TestAPoorExplanationIsRefusedByItsTrap(t *testing.T) {
	t.Parallel()

	for _, test := range explanationBreakages {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			test.change(&draft)
			problems := checks.Explanations(draft, "", testCatalog)
			if !mentions(problems, test.want) {
				t.Fatalf("Explanations() = %v, want a problem mentioning %q", problems, test.want)
			}
			for _, problem := range problems {
				if problem.Code != checks.CodeDistractorExplanations {
					t.Errorf("code = %q, want %q", problem.Code, checks.CodeDistractorExplanations)
				}
			}
		})
	}
}

// What the conditions do not forbid passes: they judge nothing about meaning,
// and a wrong guess would cost the child another attempt.
func TestWhatTheConditionsDoNotForbidPasses(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*checks.Draft)
	}{
		{"three words", func(d *checks.Draft) {
			d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "Monday is today."}
		}},
		{"three words in Russian", func(d *checks.Draft) {
			d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "Ты забыл промежуток."}
		}},
		{"six characters in Chinese", func(d *checks.Draft) {
			d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "你多数了一次。"}
		}},
		{"the solution's words, but not its start", func(d *checks.Draft) {
			d.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: "There are 6 pairs, not 8."}
		}},
		{"a start that stops inside a word", func(d *checks.Draft) {
			d.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: "List the pair"}
		}},
		{"another trap's description", func(d *checks.Draft) {
			d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "Off by one when counting gaps."}
		}},
		{"no text at all, which is the structure check's", func(d *checks.Draft) {
			d.Task.Distractors["E"] = checks.Distractor{Trap: "double_count", Text: "   "}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			test.change(&draft)
			if problems := checks.Explanations(draft, "", testCatalog); len(problems) != 0 {
				t.Fatalf("Explanations() = %v, want none", problems)
			}
		})
	}
}

// A refusal points at an explanation by its trap: the letters of the options
// explained would name the wrong options, and so the right one.
func TestNoExplanationRefusalQuotesAnOption(t *testing.T) {
	t.Parallel()

	options := slices.Collect(maps.Values(validDraft().Task.Options))
	for _, test := range explanationBreakages {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			test.change(&draft)
			for _, problem := range checks.Explanations(draft, "", testCatalog) {
				if quoted := quotedOption(problem.Message, options); quoted != "" {
					t.Errorf("%q quotes %q", problem.Message, quoted)
				}
			}
		})
	}
}

func TestAnUnreadTaskIsNotCheckedAgain(t *testing.T) {
	t.Parallel()

	if problems := checks.Explanations(checks.Draft{}, "", testCatalog); len(problems) != 0 {
		t.Fatalf("Explanations(empty draft) = %v, want none", problems)
	}
}

// The verdict is about the explanations, not about where they sit: moving the
// same four under other letters changes nothing a refusal says. If it did, the
// refusal would carry something of the letters it must not name.
func TestTheVerdictDoesNotDependOnTheLetters(t *testing.T) {
	t.Parallel()

	text := gen.OneConstOf("Wrong.", "4 is the number of ships.", "List the pairs.", "You counted every pair twice.",
		"Counts the same thing twice.", "错了。", "You missed one pair.")
	trap := gen.OneConstOf("number_from_text", "missed_case", "wrong_operation", "double_count", "")

	properties := gopter.NewProperties(nil)
	properties.Property("a refusal is the same whichever letters the explanations sit under", prop.ForAll(
		func(texts, traps []string, shift int) bool {
			explained := make([]checks.Distractor, len(texts))
			for i := range explained {
				explained[i] = checks.Distractor{Trap: traps[i], Text: texts[i]}
			}
			return slices.Equal(verdict(explained, 0), verdict(explained, shift))
		},
		gen.SliceOfN(4, text), gen.SliceOfN(4, trap), gen.IntRange(1, 4),
	))
	properties.TestingRun(t)
}

// verdict is what the check says about four explanations placed under the wrong
// options starting from one letter, sorted so that order does not count.
func verdict(explained []checks.Distractor, shift int) []string {
	draft := validDraft()
	wrong := []string{"A", "B", "D", "E"}
	draft.Task.Distractors = map[string]checks.Distractor{}
	for i, distractor := range explained {
		draft.Task.Distractors[wrong[(i+shift)%len(wrong)]] = distractor
	}
	var messages []string
	for _, problem := range checks.Explanations(draft, "", testCatalog) {
		messages = append(messages, problem.Message)
	}
	slices.Sort(messages)
	return messages
}

// The explanations come from the chat's model and are untrusted: whatever they
// hold, the check ends in problems of its own code, never in a crash.
func FuzzExplanations(f *testing.F) {
	f.Add("4 is the number of ships.", "You missed one pair.", "List the pairs.", "Wrong.", "List the pairs: 6.", "How many?")
	f.Add("错了。", "小明有三个苹果", "", "?!", "小明有三个苹果和两个梨。", "")
	f.Add("\xff\xfe", "a  a", "A A", "a", "a a a", "\x00")

	f.Fuzz(func(t *testing.T, first, second, third, fourth, solution, hint string) {
		draft := validDraft()
		draft.Task.Solution, draft.Task.Hint = solution, hint
		draft.Task.Distractors = map[string]checks.Distractor{
			"A": {Trap: "number_from_text", Text: first},
			"B": {Trap: "missed_case", Text: second},
			"D": {Trap: "wrong_operation", Text: third},
			"E": {Trap: "double_count", Text: fourth},
		}
		for _, problem := range checks.Explanations(draft, "", testCatalog) {
			if problem.Code != checks.CodeDistractorExplanations || problem.Message == "" {
				t.Fatalf("problem = %+v, want a message under %q", problem, checks.CodeDistractorExplanations)
			}
		}
	})
}

// An explanation's length is counted in the unit of the task's language, as the
// question's is. A Chinese explanation that names its children in Latin
// letters has more of those than Chinese ones, and counted by its letters it
// would be one word, too short to say anything; in Chinese it is eleven
// characters and says enough.
func TestAnExplanationIsCountedInTheUnitOfItsLanguage(t *testing.T) {
	t.Parallel()

	draft := validDraft()
	draft.Task.Distractors["B"] = checks.Distractor{Trap: "missed_case", Text: "Tom把Mary算漏了"}

	for _, test := range []struct {
		language string
		refused  bool
	}{
		{language: "zh", refused: false},
		{language: "", refused: true},
	} {
		tooShort := false
		for _, problem := range checks.Explanations(draft, test.language, testCatalog) {
			tooShort = tooShort || strings.Contains(problem.Message, "it takes at least")
		}
		if tooShort != test.refused {
			t.Errorf("Explanations(language %q) refuses it as too short: %v, want %v", test.language, tooShort, test.refused)
		}
	}
}
