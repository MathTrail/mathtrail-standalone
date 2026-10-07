package checks_test

import (
	"encoding/json"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// read is a draft as the JSON it arrives as, read and held to the format.
func read(t *testing.T, submitted parts) (checks.Draft, []checks.Problem) {
	t.Helper()

	draft, problems := checks.Decode(submitted.task, submitted.selfCheck)
	return draft, append(problems, checks.Structure(draft, testCatalog)...)
}

// lowerKeys writes the keys of one object of a part in lower case, between
// spaces.
func lowerKeys(t *testing.T, raw json.RawMessage, member string) json.RawMessage {
	t.Helper()

	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		t.Fatalf("read %s: %v", raw, err)
	}
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(members[member], &entries); err != nil {
		t.Fatalf("read %s of %s: %v", member, raw, err)
	}
	lowered := make(map[string]json.RawMessage, len(entries))
	for key, value := range entries {
		lowered[" "+strings.ToLower(key)+" "] = value
	}
	members[member] = marshal(t, lowered)
	return marshal(t, members)
}

// A letter written in another case or between spaces is read as the letter it
// is, wherever the format holds a letter: the draft passes the format and
// names the field it was read in, and never the letter.
func TestALetterInAnotherCaseOrWithSpacesIsMended(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*parts)
		field  string
	}{
		{"the right answer", func(p *parts) { p.task = withMember(t, p.task, "correct_answer", " c ") },
			"task.correct_answer"},
		{"the keys of the options", func(p *parts) { p.task = lowerKeys(t, p.task, "options") }, "task.options"},
		{"the keys of the explanations", func(p *parts) { p.task = lowerKeys(t, p.task, "distractors") },
			"task.distractors"},
		{"the keys of the self-check's options", func(p *parts) {
			p.selfCheck = lowerKeys(t, p.selfCheck, "option_check")
		}, "self_check.option_check"},
		{"the self-check's verdict", func(p *parts) { p.selfCheck = withMember(t, p.selfCheck, "final_answer", "c") },
			"self_check.final_answer"},
		{"the self-check's verdict that it is unsolvable", func(p *parts) {
			p.selfCheck = withMember(t, p.selfCheck, "final_answer", "unsolvable ")
		}, "self_check.final_answer"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			submitted := validParts(t)
			test.change(&submitted)
			draft, problems := read(t, submitted)
			if len(problems) != 0 {
				t.Fatalf("problems = %v, want the letter read as the one it is", problems)
			}
			if !slices.Equal(draft.Mended, []string{test.field}) {
				t.Errorf("Mended = %v, want %q alone", draft.Mended, test.field)
			}
		})
	}
}

// An option written as a plain number is read as its text, digit for digit.
func TestAnOptionWrittenAsANumberIsMended(t *testing.T) {
	t.Parallel()

	submitted := validParts(t)
	submitted.task = withMember(t, submitted.task, "options", map[string]any{
		"A": 4, "B": 5, "C": 6, "D": 8.5, "E": -12,
	})
	draft, problems := read(t, submitted)
	if len(problems) != 0 {
		t.Fatalf("problems = %v, want the numbers read as their text", problems)
	}
	if want := map[string]string{"A": "4", "B": "5", "C": "6", "D": "8.5", "E": "-12"}; !maps.Equal(draft.Task.Options, want) {
		t.Errorf("options = %v, want %v", draft.Task.Options, want)
	}
	if !slices.Equal(draft.Mended, []string{"task.options"}) {
		t.Errorf("Mended = %v, want the options named once", draft.Mended)
	}
}

// No issues written as null are read as an empty list; issues left out are
// still missing, since a member left out is the model's to write.
func TestNullForTheIssuesIsMendedAndLeavingThemOutIsNot(t *testing.T) {
	t.Parallel()

	submitted := validParts(t)
	submitted.selfCheck = withMember(t, submitted.selfCheck, "issues", nil)
	draft, problems := read(t, submitted)
	if len(problems) != 0 || !slices.Equal(draft.Mended, []string{"self_check.issues"}) {
		t.Errorf("issues as null: problems %v, mended %v, want them read as none", problems, draft.Mended)
	}

	var members map[string]json.RawMessage
	if err := json.Unmarshal(validParts(t).selfCheck, &members); err != nil {
		t.Fatalf("read the self-check: %v", err)
	}
	delete(members, "issues")
	left := validParts(t)
	left.selfCheck = marshal(t, members)
	draft, problems = read(t, left)
	if !mentions(problems, "self_check.issues is missing") || draft.Mended != nil {
		t.Errorf("issues left out: problems %v, mended %v, want them missing", problems, draft.Mended)
	}
}

// What could mean two things, or is no slip of how a value is written, is
// still refused, and nothing is named as mended.
func TestWhatIsNotBeyondDoubtIsStillRefused(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*parts)
		want   string
	}{
		{"an answer written with its bracket", func(p *parts) { p.task = withMember(t, p.task, "correct_answer", "C)") },
			"task.correct_answer"},
		{"an answer written as the option's text", func(p *parts) {
			p.task = withMember(t, p.task, "correct_answer", "six pairs")
		}, "task.correct_answer"},
		{"an answer written as a number", func(p *parts) { p.task = withMember(t, p.task, "correct_answer", "3") },
			"task.correct_answer"},
		{"two keys that come to one letter", func(p *parts) {
			p.task = withMember(t, p.task, "options", map[string]string{
				"A": "four pairs", "a": "five pairs", "C": "six pairs", "D": "eight pairs", "E": "twelve pairs",
			})
		}, "task.options must hold exactly five options"},
		{"six keys, two of which come to one letter", func(p *parts) {
			p.task = withMember(t, p.task, "options", map[string]string{
				"A": "four pairs", "a": "five pairs", "B": "five pairs", "C": "six pairs", "D": "eight pairs",
				"E": "twelve pairs",
			})
		}, "task.options must hold exactly five options"},
		{"a number written with an exponent", func(p *parts) {
			p.task = json.RawMessage(strings.Replace(string(p.task), `"A":"four pairs"`, `"A":4e0`, 1))
		}, "task.options.* must be a string"},
		{"a number past what a float holds exactly", func(p *parts) {
			p.task = json.RawMessage(strings.Replace(string(p.task), `"A":"four pairs"`, `"A":9007199254740993`, 1))
		}, "task.options.* must be a string"},
		{"options written as a list", func(p *parts) {
			p.task = withMember(t, p.task, "options", []string{"4", "5", "6", "8", "12"})
		}, "task.options must be an object"},
		{"a member whose name is in another case", func(p *parts) { p.task = withMember(t, p.task, "Hint", "Look.") },
			`task has a member the format does not have: "Hint"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			submitted := validParts(t)
			test.change(&submitted)
			draft, problems := read(t, submitted)
			if !mentions(problems, test.want) {
				t.Errorf("problems = %v, want one mentioning %q", problems, test.want)
			}
			if draft.Mended != nil {
				t.Errorf("Mended = %v, want it refused rather than read", draft.Mended)
			}
		})
	}
}

// Mending reads a draft as it was meant and changes nothing else: a draft
// written right is read with nothing mended; any case and spacing of its
// letters is read as that same draft, naming exactly the fields written
// otherwise; and a draft read once mends nothing more when read again.
func TestMendingHoldsItsProperties(t *testing.T) {
	t.Parallel()

	right := validParts(t)
	meant, problems := read(t, right)
	if len(problems) != 0 || meant.Mended != nil {
		t.Fatalf("the draft written right: problems %v, mended %v, want neither", problems, meant.Mended)
	}

	properties := gopter.NewProperties(nil)
	properties.Property("any case and spacing of the letters reads as the draft meant", prop.ForAll(
		func(written []bool, spaces string) bool {
			answer, keys, verdict := written[0], written[1], written[2]
			submitted := validParts(t)
			var named []string
			if answer {
				submitted.task = withMember(t, submitted.task, "correct_answer", spaces+"c"+spaces)
				named = append(named, "task.correct_answer")
			}
			if keys {
				submitted.task = lowerKeys(t, submitted.task, "options")
				named = append(named, "task.options")
			}
			if verdict {
				submitted.selfCheck = withMember(t, submitted.selfCheck, "final_answer", spaces+"c")
				named = append(named, "self_check.final_answer")
			}
			draft, problems := read(t, submitted)
			if len(problems) != 0 || !slices.Equal(draft.Mended, named) {
				return false
			}
			again, _ := read(t, parts{task: marshal(t, draft.Task), selfCheck: marshal(t, draft.SelfCheck)})
			if again.Mended != nil {
				return false
			}
			draft.Mended = nil
			return reflect.DeepEqual(draft, meant) && reflect.DeepEqual(again, meant)
		},
		gen.SliceOfN(3, gen.Bool()),
		gen.OneConstOf("", " ", "  ", "\t", "\n"),
	))
	properties.TestingRun(t)
}
