package checks_test

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// catalog is a catalog written out in the test, so that what the checks are
// measured against is in one place and small enough to read.
type catalog struct {
	topics, skills []string
	traps          map[string]string // id to description
}

func (c catalog) HasTopic(id string) bool { return slices.Contains(c.topics, id) }
func (c catalog) HasSkill(id string) bool { return slices.Contains(c.skills, id) }

func (c catalog) HasTrap(id string) bool {
	_, known := c.traps[id]
	return known
}

func (c catalog) TrapDescription(id string) (string, bool) {
	description, known := c.traps[id]
	return description, known
}

var testCatalog = catalog{
	topics: []string{"combinatorics.enumeration", "counting.gaps"},
	skills: []string{"division_with_remainder", "fractions"},
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
		Difficulty:      3,
		Setting:         "space",
		TrapsToUse:      []string{"missed_case", "double_count"},
		ExcludedSkills:  []string{"division_with_remainder"},
		Constraints:     []string{},
		Rationale:       "The last task on this topic failed, so the same ground again.",
	}
}

// validDraft is the prototype's own well-formed task, with the brief it was
// asked for and a self-check that found nothing. Its options are words rather
// than numbers, so that a refusal quoting one would be caught by a search.
func validDraft() checks.Draft {
	return checks.Draft{
		Brief: asked(),
		Task: &checks.Task{
			CoreIdea:             "Unordered pairs among 4 objects: 6.",
			DesignThoughtProcess: "Plot: docking.",
			Question:             "Four spaceships dock in pairs. How many different pairs can they make?",
			Options: map[string]string{
				"A": "four pairs", "B": "five pairs", "C": "six pairs", "D": "eight pairs", "E": "twelve pairs",
			},
			CorrectAnswer: "C",
			Hint:          "How many ships can the first ship dock with?",
			Solution:      "List the pairs: 6.",
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

	if problems := checks.Structure(validDraft(), asked(), testCatalog); len(problems) != 0 {
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
	{"same after strip", func(d *checks.Draft) { d.Task.Options["B"] = " six pairs " }, "reads the same as another"},
	{"same ignoring case", func(d *checks.Draft) { d.Task.Options["A"] = "SIX PAIRS" }, "reads the same as another"},
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
// tests did: the brief agrees with the request, the ids exist, and every part
// is complete.
var formatBreakages = []breakage{
	{"a goal the format does not have", func(d *checks.Draft) { d.Brief.PedagogicalGoal = "revise" }, "brief.pedagogical_goal"},
	{"a topic nobody has", func(d *checks.Draft) { d.Brief.TargetConcept = "geometry.spheres" }, "not a topic in the catalog"},
	{"a topic the request was not for", func(d *checks.Draft) { d.Brief.TargetConcept = "counting.gaps" }, "not the topic this task was asked for"},
	{"a difficulty off the scale", func(d *checks.Draft) { d.Brief.Difficulty = 6 }, "from 1 to 5"},
	{"a difficulty the request was not for", func(d *checks.Draft) { d.Brief.Difficulty = 4 }, "not the difficulty this task was asked for"},
	{"no setting", func(d *checks.Draft) { d.Brief.Setting = " " }, "brief.setting"},
	{"no rationale", func(d *checks.Draft) { d.Brief.Rationale = "" }, "brief.rationale"},
	{"traps left out", func(d *checks.Draft) { d.Brief.TrapsToUse = nil }, "brief.traps_to_use is missing"},
	{"no trap at all", func(d *checks.Draft) { d.Brief.TrapsToUse = []string{} }, "at least one trap"},
	{"a trap nobody has", func(d *checks.Draft) { d.Brief.TrapsToUse = []string{"guessed"} }, `names "guessed", which is not a trap`},
	{"a trap named twice", func(d *checks.Draft) { d.Brief.TrapsToUse = []string{"missed_case", "missed_case"} }, "twice"},
	{"excluded skills left out", func(d *checks.Draft) { d.Brief.ExcludedSkills = nil }, "brief.excluded_skills is missing"},
	{"an excluded skill dropped", func(d *checks.Draft) { d.Brief.ExcludedSkills = []string{} }, `dropped "division_with_remainder"`},
	{"a skill nobody has", func(d *checks.Draft) {
		d.Brief.ExcludedSkills = []string{"division_with_remainder", "juggling"}
	}, `names "juggling", which is not a skill`},
	{"constraints left out", func(d *checks.Draft) { d.Brief.Constraints = nil }, "brief.constraints is missing"},
	{"an empty constraint", func(d *checks.Draft) { d.Brief.Constraints = []string{"short", " "} }, "brief.constraints.1 is empty"},
	{"no question", func(d *checks.Draft) { d.Task.Question = "" }, "task.question"},
	{"no core idea", func(d *checks.Draft) { d.Task.CoreIdea = "" }, "task.core_idea"},
	{"no design thought", func(d *checks.Draft) { d.Task.DesignThoughtProcess = "" }, "task.design_thought_process"},
	{"no solution", func(d *checks.Draft) { d.Task.Solution = "\n" }, "task.solution"},
	{"no explanations at all", func(d *checks.Draft) { d.Task.Distractors = nil }, "task.distractors is missing"},
	{"an explanation with no trap", func(d *checks.Draft) {
		d.Task.Distractors["E"] = checks.Distractor{Text: "You counted every pair twice."}
	}, "names no trap"},
	{"a drawing without its structure", func(d *checks.Draft) { d.Task.Drawing = "o--o" }, "task.drawing_structure, and it is missing"},
	{"a structure without its drawing", func(d *checks.Draft) {
		d.Task.DrawingStructure = &checks.DrawingStructure{Kind: "number_line", Objects: []checks.DrawingObject{{ID: "A", Label: "A"}}}
	}, "task.drawing, and it is missing"},
	{"a drawing that does not say what it draws", func(d *checks.Draft) {
		d.Task.Drawing = "o--o"
		d.Task.DrawingStructure = &checks.DrawingStructure{Objects: []checks.DrawingObject{{ID: "A", Label: "A"}}}
	}, "task.drawing_structure.kind"},
	{"a drawing that draws nothing", func(d *checks.Draft) {
		d.Task.Drawing = "o--o"
		d.Task.DrawingStructure = &checks.DrawingStructure{Kind: "number_line"}
	}, "at least one object"},
	{"a drawn object with no label", func(d *checks.Draft) {
		d.Task.Drawing = "o--o"
		d.Task.DrawingStructure = &checks.DrawingStructure{Kind: "number_line", Objects: []checks.DrawingObject{{ID: "A"}}}
	}, "objects.0 has no label"},
	{"a relation that is not a triple", func(d *checks.Draft) {
		d.Task.Drawing = "o--o"
		d.Task.DrawingStructure = &checks.DrawingStructure{
			Kind:      "number_line",
			Objects:   []checks.DrawingObject{{ID: "A", Label: "A"}},
			Relations: []checks.DrawingRelation{{Type: "left_of", From: "A"}},
		}
	}, "a type, a from and a to"},
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
			problems := checks.Structure(draft, asked(), testCatalog)
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
			for _, problem := range checks.Structure(draft, asked(), testCatalog) {
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
		{"two empty constraints", func(d *checks.Draft) {
			d.Brief.Constraints = []string{" ", "short", ""}
		}, []string{"brief.constraints.0 is empty", "brief.constraints.2 is empty"}},
		{"two drawn objects with no label", func(d *checks.Draft) {
			d.Task.Drawing = "A--B--C"
			d.Task.DrawingStructure = &checks.DrawingStructure{
				Kind:    "number_line",
				Objects: []checks.DrawingObject{{ID: "A"}, {ID: "B", Label: "B"}, {ID: "C"}},
			}
		}, []string{"objects.0 has no label", "objects.2 has no label"}},
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
			problems := checks.Structure(draft, asked(), testCatalog)
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
	for _, problem := range checks.Structure(draft, asked(), testCatalog) {
		if strings.Contains(problem.Message, `"guessed"`) {
			named++
		}
	}
	if named != 1 {
		t.Fatalf("the unknown trap is named %d times, want once", named)
	}
}

// What the format allows is not refused: a skill the model adds, traps swapped
// for others from the catalog, a self-check that could not solve the task, and
// a drawing that comes with its structure.
func TestWhatTheFormatAllowsPasses(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*checks.Draft)
	}{
		{"a skill added to the excluded ones", func(d *checks.Draft) {
			d.Brief.ExcludedSkills = []string{"division_with_remainder", "fractions"}
		}},
		{"traps swapped for others", func(d *checks.Draft) { d.Brief.TrapsToUse = []string{"off_by_one"} }},
		{"a constraint", func(d *checks.Draft) { d.Brief.Constraints = []string{"at most three sentences"} }},
		{"an unsolvable verdict", func(d *checks.Draft) { d.SelfCheck.FinalAnswer = "UNSOLVABLE" }},
		{"a minor issue", func(d *checks.Draft) {
			d.SelfCheck.Issues = []checks.Issue{{Type: "too_hard_for_grade", Severity: "minor", Comment: "Close to the edge."}}
		}},
		{"a drawing with its structure", func(d *checks.Draft) {
			d.Task.Drawing = "A--B"
			d.Task.DrawingStructure = &checks.DrawingStructure{
				Kind:      "number_line",
				Objects:   []checks.DrawingObject{{ID: "A", Label: "A"}, {ID: "B", Label: "B"}},
				Relations: []checks.DrawingRelation{{Type: "left_of", From: "A", To: "B"}},
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			draft := validDraft()
			test.change(&draft)
			if problems := checks.Structure(draft, asked(), testCatalog); len(problems) != 0 {
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

	if problems := checks.Structure(checks.Draft{}, asked(), testCatalog); len(problems) != 0 {
		t.Fatalf("Structure(empty draft) = %v, want no problems", problems)
	}
}

// mentions says whether any problem's message contains the fragment.
func mentions(problems []checks.Problem, fragment string) bool {
	return slices.ContainsFunc(problems, func(p checks.Problem) bool { return strings.Contains(p.Message, fragment) })
}
