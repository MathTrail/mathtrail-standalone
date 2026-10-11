package checks_test

import (
	"encoding/json"
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// catalog is a catalog written out in the test, so that what the checks are
// measured against is in one place and small enough to read.
type catalog struct {
	traps map[string]string // id to description
}

func (c catalog) TrapDescription(id string) (string, bool) {
	description, known := c.traps[id]
	return description, known
}

var testCatalog = catalog{
	traps: map[string]string{
		"number_from_text": "Takes a number from the question as the answer.",
		"missed_case":      "Leaves out one of the cases.",
		"wrong_operation":  "Uses the wrong operation.",
		"double_count":     "Counts the same thing twice.",
		"off_by_one":       "Off by one when counting gaps.",
	},
}

// asked is the brief the request was opened with.
func asked() *profile.Brief {
	return &profile.Brief{
		PedagogicalGoal: profile.GoalReinforce,
		TargetConcept:   "combinatorics.enumeration",
		GradeLevel:      rating.Grades34,
		Difficulty:      3,
		Setting:         "space",
		TrapsToUse:      []string{"missed_case", "double_count"},
		ExcludedSkills:  []string{"division_with_remainder"},
		Constraints:     []string{},
		Rationale:       "The last task on this topic failed, so the same ground again.",
	}
}

// validDraft is the prototype's own well-formed task, with the brief it was
// asked for and a self-check that found nothing, and the picture of its
// solution every task draws: the six pairs, one to a row. Its options are
// words rather than numbers, so that a refusal quoting one would be caught by
// a search.
func validDraft() checks.Draft {
	return checks.Draft{
		Task: &checks.Task{
			CoreIdea: "Unordered pairs among 4 objects: 6.",
			Question: "Four spaceships dock in pairs. How many different pairs can they make?",
			Options: map[string]string{
				"A": "four pairs", "B": "five pairs", "C": "six pairs", "D": "eight pairs", "E": "twelve pairs",
			},
			CorrectAnswer: "C",
			Hint:          "How many ships can the first ship dock with?",
			Solution:      "List the pairs: 6.",
			SolutionPicture: json.RawMessage(
				`{"kind":"table","rows":[["1","2"],["1","3"],["1","4"],["2","3"],["2","4"],["3","4"]]}`),
			SolutionTotal: "3 + 2 + 1 = 6",
			Distractors: map[string]checks.Distractor{
				"A": {Trap: "number_from_text", Text: "4 is the number of ships."},
				"B": {Trap: "missed_case", Text: "You missed one pair."},
				"D": {Trap: "wrong_operation", Text: "You doubled the number of ships."},
				"E": {Trap: "double_count", Text: "You counted every pair twice."},
			},
		},
		SelfCheck: &checks.SelfCheck{
			Issues: []checks.Issue{},
			OptionCheck: map[string]string{
				"A": "The number of ships.", "B": "Misses a pair.", "C": "The six pairs.",
				"D": "Doubles the ships.", "E": "Counts each pair twice.",
			},
			FinalAnswer: "C",
		},
	}
}

func TestAWellFormedDraftPasses(t *testing.T) {
	t.Parallel()

	if problems := checks.Structure(validDraft(), testCatalog); len(problems) != 0 {
		t.Fatalf("Structure() = %v, want no problems", problems)
	}
}

// breakage is one way a draft can be out of the format, and a fragment of what
// the refusal must say about it.
type breakage struct {
	name   string
	change func(*checks.Draft)
	want   string
}

// prototypeBreakages are the twelve of the prototype's tests/test_filters.py,
// on the same task, broken the same way.
var prototypeBreakages = []breakage{
	{"same after strip", func(d *checks.Draft) { d.Task.Options["B"] = " six pairs " }, "says the same as another"},
	{"same ignoring case", func(d *checks.Draft) { d.Task.Options["A"] = "SIX PAIRS" }, "says the same as another"},
	{"empty option", func(d *checks.Draft) { d.Task.Options["E"] = "  " }, "task.options has an empty option"},
	{"four options", func(d *checks.Draft) { delete(d.Task.Options, "E") }, "exactly five options"},
	{"unknown correct letter", func(d *checks.Draft) { d.Task.CorrectAnswer = "F" }, "task.correct_answer"},
	{"missing distractor", func(d *checks.Draft) { delete(d.Task.Distractors, "A") }, "the four wrong options"},
	{"distractor for the correct option", func(d *checks.Draft) {
		d.Task.Distractors["C"] = checks.Distractor{Trap: "off_by_one", Text: "Right answer."}
	}, "the four wrong options"},
	{"correct option has a distractor", func(d *checks.Draft) { d.Task.CorrectAnswer = "B" }, "the four wrong options"},
	{"blank hint", func(d *checks.Draft) { d.Task.Hint = "   " }, "task.hint"},
	{"no hint", func(d *checks.Draft) { d.Task.Hint = "" }, "task.hint"},
	{"unknown trap", func(d *checks.Draft) {
		d.Task.Distractors["B"] = checks.Distractor{Trap: "forgot_a_case", Text: "You missed one pair."}
	}, `"forgot_a_case"`},
	{"empty distractor text", func(d *checks.Draft) {
		d.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: ""}
	}, "an explanation with no text"},
}

// formatBreakages break the rest of the format, beyond what the prototype's
// tests did: the ids exist, and every part is complete.
var formatBreakages = []breakage{
	{"an option that shows nothing", func(d *checks.Draft) { d.Task.Options["E"] = "\u200c\u200b" },
		"task.options has an empty option"},
	{"options that are one number written two ways", func(d *checks.Draft) {
		d.Task.Options["B"], d.Task.Options["D"] = "6", "6.0"
	}, "says the same as another"},
	{"options apart only by the spaces between words", func(d *checks.Draft) { d.Task.Options["B"] = "six  pairs" },
		"says the same as another"},
	{"options apart only by a no-break space", func(d *checks.Draft) { d.Task.Options["B"] = "six\u00a0pairs" },
		"says the same as another"},
	{"options apart only by a character that takes no room", func(d *checks.Draft) {
		d.Task.Options["B"] = "six\u200b pairs"
	}, "says the same as another"},
	{"options apart only by how an accent is put together", func(d *checks.Draft) {
		d.Task.Options["A"], d.Task.Options["B"] = "six caf\u00e9s", "six cafe\u0301s"
	}, "says the same as another"},
	{"no question", func(d *checks.Draft) { d.Task.Question = "" }, "task.question"},
	{"a question that shows nothing", func(d *checks.Draft) { d.Task.Question = "\u200b\u2060" }, "task.question"},
	{"no core idea", func(d *checks.Draft) { d.Task.CoreIdea = "" }, "task.core_idea"},
	{"no solution", func(d *checks.Draft) { d.Task.Solution = "\n" }, "task.solution"},
	{"no explanations at all", func(d *checks.Draft) { d.Task.Distractors = nil }, "task.distractors is missing"},
	{"an explanation with no trap", func(d *checks.Draft) {
		d.Task.Distractors["E"] = checks.Distractor{Text: "You counted every pair twice."}
	}, "names no trap"},
	{"a picture that is no object", func(d *checks.Draft) { d.Task.Picture = json.RawMessage(`"a clock at 4:30"`) },
		"task.picture must be an object"},
	{"a picture of the solution that is no object", func(d *checks.Draft) {
		d.Task.SolutionPicture = json.RawMessage(`["six pairs"]`)
	}, "task.solution_picture must be an object"},
	{"no picture of the solution", func(d *checks.Draft) {
		d.Task.SolutionPicture, d.Task.SolutionTotal = nil, ""
	}, "task.solution_picture is missing, or holds nothing: every task draws its solution"},
	{"a picture written as a list", func(d *checks.Draft) { d.Task.Picture = json.RawMessage(`[{"kind":"clock"}]`) },
		"task.picture must be an object"},
	{"issues left out", func(d *checks.Draft) { d.SelfCheck.Issues = nil }, "self_check.issues is missing"},
	{"an issue of no known type", func(d *checks.Draft) {
		d.SelfCheck.Issues = []checks.Issue{{Type: "boring", Severity: "minor", Comment: "Too easy."}}
	}, `type "boring"`},
	{"an issue of no known severity", func(d *checks.Draft) {
		d.SelfCheck.Issues = []checks.Issue{{Type: "ambiguous", Severity: "fatal", Comment: "Two readings."}}
	}, `severity "fatal"`},
	{"an issue with no comment", func(d *checks.Draft) {
		d.SelfCheck.Issues = []checks.Issue{{Type: "ambiguous", Severity: "minor"}}
	}, "has no comment"},
	{"an option nobody checked", func(d *checks.Draft) { delete(d.SelfCheck.OptionCheck, "D") }, "self_check.option_check"},
	{"an empty option check", func(d *checks.Draft) { d.SelfCheck.OptionCheck["D"] = "" }, "self_check.option_check has an empty entry"},
	{"a final answer that is no option", func(d *checks.Draft) { d.SelfCheck.FinalAnswer = "maybe" }, "self_check.final_answer"},
}

func TestABrokenDraftIsRefusedByName(t *testing.T) {
	t.Parallel()

	for _, test := range slices.Concat(prototypeBreakages, formatBreakages) {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			test.change(&draft)
			problems := checks.Structure(draft, testCatalog)
			if !mentions(problems, test.want) {
				t.Fatalf("Structure() = %v, want a problem mentioning %q", problems, test.want)
			}
			for _, problem := range problems {
				if problem.Code != checks.CodeBadStructure {
					t.Errorf("code = %q, want %q", problem.Code, checks.CodeBadStructure)
				}
			}
		})
	}
}

// A letter is a standalone capital A to E: a refusal naming one says which
// option is right or wrong.
var letter = regexp.MustCompile(`\b[A-E]\b`)

// A refusal reaches the widget with everything else submit_task returns, so it
// must say what to fix without saying which option is which: no letter and no
// option text, whatever was broken.
func TestNoRefusalQuotesAnOption(t *testing.T) {
	t.Parallel()

	options := slices.Collect(maps.Values(validDraft().Task.Options))
	for _, test := range slices.Concat(prototypeBreakages, formatBreakages) {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			test.change(&draft)
			for _, problem := range checks.Structure(draft, testCatalog) {
				if quoted := quotedOption(problem.Message, options); quoted != "" {
					t.Errorf("%q quotes %q", problem.Message, quoted)
				}
			}
		})
	}
}

// quotedOption is the letter or the option text a message quotes, or nothing.
func quotedOption(message string, options []string) string {
	if found := letter.FindString(message); found != "" {
		return found
	}
	for _, text := range options {
		if strings.Contains(strings.ToLower(message), strings.ToLower(strings.TrimSpace(text))) {
			return text
		}
	}
	return ""
}

// Every fault is reported at once, and that holds inside a list as well: each
// faulty entry of a list is named by its place, not only the first one found.
// An entry keyed by an option's letter cannot be named without the letter, so
// those are counted instead.
func TestEveryFaultyEntryIsReported(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*checks.Draft)
		want   []string
	}{
		{"two issues with no comment", func(d *checks.Draft) {
			d.SelfCheck.Issues = []checks.Issue{
				{Type: "ambiguous", Severity: "minor"},
				{Type: "missing_data", Severity: "minor"},
			}
		}, []string{"self_check.issues.0 has no comment", "self_check.issues.1 has no comment"}},
		{"two empty option checks", func(d *checks.Draft) {
			d.SelfCheck.OptionCheck["A"], d.SelfCheck.OptionCheck["B"] = "", " "
		}, []string{"self_check.option_check has 2 empty entries"}},
		{"two empty options", func(d *checks.Draft) {
			d.Task.Options["A"], d.Task.Options["B"] = "", " "
		}, []string{"task.options has 2 empty options"}},
		{"two explanations with no text", func(d *checks.Draft) {
			d.Task.Distractors["A"] = checks.Distractor{Trap: "number_from_text"}
			d.Task.Distractors["B"] = checks.Distractor{Trap: "missed_case"}
		}, []string{"task.distractors has 2 explanations with no text"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			test.change(&draft)
			problems := checks.Structure(draft, testCatalog)
			for _, want := range test.want {
				if !mentions(problems, want) {
					t.Errorf("Structure() = %v, want a problem mentioning %q", problems, want)
				}
			}
		})
	}
}

// A trap nobody has is one thing to fix however many explanations name it.
func TestAnUnknownTrapIsNamedOnce(t *testing.T) {
	t.Parallel()

	draft := validDraft()
	draft.Task.Distractors["A"] = checks.Distractor{Trap: "guessed", Text: "A guess."}
	draft.Task.Distractors["B"] = checks.Distractor{Trap: "guessed", Text: "Another guess."}
	named := 0
	for _, problem := range checks.Structure(draft, testCatalog) {
		if strings.Contains(problem.Message, `"guessed"`) {
			named++
		}
	}
	if named != 1 {
		t.Fatalf("the unknown trap is named %d times, want once", named)
	}
}

// What the format allows is not refused: a self-check that could not solve the
// task, a minor issue, and a picture, whose members its own check reads.
func TestWhatTheFormatAllowsPasses(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*checks.Draft)
	}{
		{"an unsolvable verdict", func(d *checks.Draft) { d.SelfCheck.FinalAnswer = "UNSOLVABLE" }},
		{"a minor issue", func(d *checks.Draft) {
			d.SelfCheck.Issues = []checks.Issue{{Type: "too_hard_for_grade", Severity: "minor", Comment: "Close to the edge."}}
		}},
		{"a picture", func(d *checks.Draft) { d.Task.Picture = json.RawMessage(`{"kind":"clock","time":"4:30"}`) }},
		{"a picture its own check refuses", func(d *checks.Draft) { d.Task.Picture = json.RawMessage(`{"kind":"star"}`) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			test.change(&draft)
			if problems := checks.Structure(draft, testCatalog); len(problems) != 0 {
				t.Fatalf("Structure() = %v, want no problems", problems)
			}
		})
	}
}

// A part that could not be read at all has already been refused by Decode;
// checking its empty fields as well would bury that refusal under a dozen
// about fields that were never there.
func TestAPartThatCouldNotBeReadIsNotCheckedAgain(t *testing.T) {
	t.Parallel()

	if problems := checks.Structure(checks.Draft{}, testCatalog); len(problems) != 0 {
		t.Fatalf("Structure(empty draft) = %v, want no problems", problems)
	}
}

// mentions says whether any problem's message contains the fragment.
func mentions(problems []checks.Problem, fragment string) bool {
	return slices.ContainsFunc(problems, func(p checks.Problem) bool { return strings.Contains(p.Message, fragment) })
}

// A text the child reads is no longer than a card holds: one at its limit
// passes, and one character more is refused, naming the text and the limit.
// The limit counts characters, so a text in another script — here Cyrillic, of
// two bytes a character — is held to the same length.
func TestATextLongerThanACardHoldsIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		set     func(d *checks.Draft, text string)
		longest int
		want    string
	}{
		{"the question", func(d *checks.Draft, text string) { d.Task.Question = text },
			1500, "task.question is 1501 characters long, and it is at most 1500"},
		{"the hint", func(d *checks.Draft, text string) { d.Task.Hint = text },
			500, "task.hint is 501 characters long, and it is at most 500"},
		{"the solution", func(d *checks.Draft, text string) { d.Task.Solution = text },
			2000, "task.solution is 2001 characters long, and it is at most 2000"},
		{"an option", func(d *checks.Draft, text string) { d.Task.Options["E"] = text },
			200, "an option is at most 200 characters"},
		{"an explanation", func(d *checks.Draft, text string) {
			d.Task.Distractors["B"] = checks.Distractor{Trap: "missed_case", Text: text}
		}, 500, "an explanation is at most 500 characters"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			atTheLimit := validDraft()
			tc.set(&atTheLimit, strings.Repeat("я", tc.longest))
			if problems := checks.Structure(atTheLimit, testCatalog); len(problems) != 0 {
				t.Errorf("Structure() at %d characters = %v, want no problems", tc.longest, problems)
			}
			past := validDraft()
			tc.set(&past, strings.Repeat("я", tc.longest+1))
			if problems := checks.Structure(past, testCatalog); !mentions(problems, tc.want) {
				t.Errorf("Structure() at %d characters = %v, want a problem mentioning %q", tc.longest+1, problems, tc.want)
			}
		})
	}
}
