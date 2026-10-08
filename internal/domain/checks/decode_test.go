package checks_test

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// parts is a draft as the JSON it arrives as.
type parts struct{ task, selfCheck json.RawMessage }

// validParts is the well-formed draft of the structure tests, written out.
func validParts(t testing.TB) parts {
	t.Helper()

	draft := validDraft()
	return parts{
		task:      marshal(t, draft.Task),
		selfCheck: marshal(t, draft.SelfCheck),
	}
}

func marshal(t testing.TB, value any) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %T: %v", value, err)
	}
	return raw
}

func TestAWellFormedSubmissionDecodesWhole(t *testing.T) {
	t.Parallel()

	submitted := validParts(t)
	draft, problems := checks.Decode(submitted.task, submitted.selfCheck)
	if len(problems) != 0 {
		t.Fatalf("Decode() problems = %v, want none", problems)
	}
	if draft.Task == nil || draft.SelfCheck == nil || draft.Retired != nil {
		t.Fatalf("Decode() = %+v, want every part read and nothing retired", draft)
	}
	if problems := checks.Structure(draft, testCatalog); len(problems) != 0 {
		t.Fatalf("Structure() after Decode() = %v, want no problems", problems)
	}
}

// A member the format has retired is read past, and named, so that a chat
// begun with the guide that asked for it loses no attempt over it.
func TestARetiredMemberIsReadPast(t *testing.T) {
	t.Parallel()

	submitted := validParts(t)
	task := withMember(t, submitted.task, "design_thought_process", "Plot: docking.")
	draft, problems := checks.Decode(task, submitted.selfCheck)
	if len(problems) != 0 {
		t.Fatalf("Decode() problems = %v, want none", problems)
	}
	if !slices.Equal(draft.Retired, []string{"task.design_thought_process"}) {
		t.Errorf("Retired = %v, want the plan named by its path", draft.Retired)
	}
}

// A retired member is read past where the format had it, and nowhere else: by
// any other path, the same name is a stray like any other.
func TestARetiredMemberIsAStrayAnywhereElse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*parts)
		want   string
	}{
		{"in the self-check", func(p *parts) {
			p.selfCheck = withMember(t, p.selfCheck, "design_thought_process", "Plot: docking.")
		}, `self_check has a member the format does not have: "design_thought_process"`},
		{"in an explanation", func(p *parts) {
			p.task = json.RawMessage(strings.Replace(string(p.task), `"trap":"missed_case"`,
				`"trap":"missed_case","design_thought_process":"why"`, 1))
		}, `task.distractors.* has a member the format does not have: "design_thought_process"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			submitted := validParts(t)
			test.change(&submitted)
			draft, problems := checks.Decode(submitted.task, submitted.selfCheck)
			if !mentions(problems, test.want) || draft.Retired != nil {
				t.Errorf("Decode() = %v, retired %v, want a stray and nothing retired", problems, draft.Retired)
			}
		})
	}
}

func TestASubmissionOutOfTheFormatIsRefusedByName(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*parts)
		want   []string
	}{
		{"a part left out", func(p *parts) { p.task = nil }, []string{"task is missing"}},
		{"a part sent as null", func(p *parts) { p.selfCheck = json.RawMessage("null") }, []string{"self_check is missing"}},
		{"a part that is not an object", func(p *parts) { p.task = json.RawMessage(`["question"]`) }, []string{"task must be a JSON object"}},
		{"a part that is not JSON", func(p *parts) { p.task = json.RawMessage(`{"question": `) }, []string{"task is not valid JSON"}},
		{"a member the format does not have", func(p *parts) {
			p.task = withMember(t, p.task, "answer", "six pairs")
		}, []string{`task has a member the format does not have: "answer"`}},
		{"every stray member, not only the first", func(p *parts) {
			p.task = withMember(t, withMember(t, p.task, "mood", "happy"), "grade", 3)
		}, []string{`"mood"`, `"grade"`}},
		{"a stray member deep inside", func(p *parts) {
			p.task = json.RawMessage(strings.Replace(string(p.task), `"trap":"missed_case"`, `"trap":"missed_case","why":"it is wrong"`, 1))
		}, []string{`task.distractors.* has a member the format does not have: "why"`}},
		{"a number where a string belongs", func(p *parts) {
			p.task = json.RawMessage(strings.Replace(string(p.task), `"trap":"missed_case"`, `"trap":5`, 1))
		}, []string{"task.distractors.*.trap must be a string"}},
		{"a string where a number belongs", func(p *parts) {
			p.task = withDrawnValue(t, p.task, "3")
		}, []string{"task.drawing_structure.objects.0.value must be a whole number"}},
		{"a fraction where a whole number belongs", func(p *parts) {
			p.task = withDrawnValue(t, p.task, 3.5)
		}, []string{"task.drawing_structure.objects.0.value must be a whole number"}},
		{"an object where a list belongs", func(p *parts) {
			p.selfCheck = json.RawMessage(strings.Replace(string(p.selfCheck), `"issues":[]`, `"issues":{}`, 1))
		}, []string{"self_check.issues must be a list"}},
		{"text where an object belongs", func(p *parts) {
			p.task = explanationAsText(p.task)
		}, []string{"task.distractors.* must be an object"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			submitted := validParts(t)
			test.change(&submitted)
			_, problems := checks.Decode(submitted.task, submitted.selfCheck)
			for _, want := range test.want {
				if !mentions(problems, want) {
					t.Errorf("Decode() = %v, want a problem mentioning %q", problems, want)
				}
			}
			for _, problem := range problems {
				if letter.MatchString(problem.Message) {
					t.Errorf("%q quotes an option's letter", problem.Message)
				}
			}
		})
	}
}

// A value of the wrong shape where an object belongs is refused once, by what
// it should hold, and never also as a member the format does not have: a value
// that is no object has no members to judge, and two refusals of one slip
// would send the model looking for a second.
func TestAWrongShapeWhereAnObjectBelongsIsRefusedOnce(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*parts)
		want   string
	}{
		{"an explanation written as text", func(p *parts) {
			p.task = explanationAsText(p.task)
		}, "task.distractors.* must be an object"},
		{"an issue written as a number", func(p *parts) {
			p.selfCheck = json.RawMessage(strings.Replace(string(p.selfCheck), `"issues":[]`, `"issues":[5]`, 1))
		}, "self_check.issues.0 must be an object"},
		{"a drawn object written as its name", func(p *parts) {
			p.task = withMember(t, p.task, "drawing_structure", map[string]any{"kind": "row", "objects": []any{"ship"}})
		}, "task.drawing_structure.objects.0 must be an object"},
		{"a drawing's structure written in words", func(p *parts) {
			p.task = withMember(t, p.task, "drawing_structure", "a row of four ships")
		}, "task.drawing_structure must be an object"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			submitted := validParts(t)
			test.change(&submitted)
			_, problems := checks.Decode(submitted.task, submitted.selfCheck)
			if want := []checks.Problem{{Code: checks.CodeBadStructure, Message: test.want}}; !slices.Equal(problems, want) {
				t.Errorf("Decode() = %v, want %v", problems, want)
			}
		})
	}
}

// explanationAsText is a task whose last explanation is written as the words
// the child is told alone, with no trap.
func explanationAsText(raw json.RawMessage) json.RawMessage {
	return json.RawMessage(strings.Replace(string(raw),
		`{"trap":"double_count","text":"You counted every pair twice."}`, `"You counted every pair twice."`, 1))
}

// The same stray member in all four explanations is one thing to fix.
func TestAStrayRepeatedAcrossEntriesIsNamedOnce(t *testing.T) {
	t.Parallel()

	submitted := validParts(t)
	task := strings.ReplaceAll(string(submitted.task), `"text":`, `"why":"so","text":`)
	_, problems := checks.Decode(json.RawMessage(task), submitted.selfCheck)
	strays := slices.DeleteFunc(slices.Clone(problems), func(p checks.Problem) bool {
		return !strings.Contains(p.Message, `"why"`)
	})
	if len(strays) != 1 {
		t.Fatalf("got %d problems about the stray member, want 1: %v", len(strays), problems)
	}
}

// A part with a mistyped member is still read, so that what else is wrong with
// it is found in the same attempt rather than the next.
func TestAMistypedPartIsStillChecked(t *testing.T) {
	t.Parallel()

	submitted := validParts(t)
	task := withDrawnValue(t, withMember(t, submitted.task, "hint", ""), "3")
	draft, problems := checks.Decode(task, submitted.selfCheck)
	if draft.Task == nil {
		t.Fatal("the task was dropped over one mistyped member")
	}
	problems = append(problems, checks.Structure(draft, testCatalog)...)
	if !mentions(problems, "task.hint") {
		t.Errorf("got %v, want the empty hint found in the same pass", problems)
	}
}

// withDrawnValue gives a task a drawing structure whose one object holds value.
func withDrawnValue(t *testing.T, raw json.RawMessage, value any) json.RawMessage {
	t.Helper()

	return withMember(t, raw, "drawing_structure", map[string]any{
		"kind": "row", "objects": []any{map[string]any{"id": "ship", "label": "S", "value": value}},
	})
}

// withMember adds one member to a JSON object.
func withMember(t *testing.T, raw json.RawMessage, name string, value any) json.RawMessage {
	t.Helper()

	var members map[string]any
	if err := json.Unmarshal(raw, &members); err != nil {
		t.Fatalf("read %s: %v", raw, err)
	}
	members[name] = value
	return marshal(t, members)
}

// What the model hands in is untrusted input, and a panic reading it takes the
// request handler down with it. Whatever arrives, reading and checking it ends
// in problems, never in a crash — and a part that could not be read is always
// named, and what was read past or mended only by the names of the format.
func FuzzDecode(f *testing.F) {
	valid := validParts(f)
	f.Add([]byte(valid.task), []byte(valid.selfCheck))
	f.Add([]byte(`{"options":{"A":1}}`), []byte(`null`))
	f.Add([]byte(`{"distractors":{"B":{"trap":[]}}}`), []byte(`{"issues":[{"type":5}]}`))
	f.Add([]byte(`{"drawing_structure":{"objects":[{"value":"x"}]}}`), []byte(`"`))
	f.Add([]byte(`{"design_thought_process":"a plan","question":"?"}`), []byte(`{"design_thought_process":1}`))
	f.Add([]byte(`{"options":{"a":4," B ":5.5,"c":"-1","D":1e3},"correct_answer":" e"}`), []byte(`{"issues":null,"final_answer":"unsolvable","option_check":{"a":"x","A":"y"}}`))

	f.Fuzz(func(t *testing.T, task, selfCheck []byte) {
		draft, problems := checks.Decode(task, selfCheck)
		problems = append(problems, checks.Structure(draft, testCatalog)...)
		wantProblemsOfTheFormat(t, problems)
		wantEveryUnreadPartNamed(t, &draft, problems)
		wantOnlyTheFormatsNames(t, &draft)
	})
}

// wantProblemsOfTheFormat holds what reading a draft found to problems of its
// format, each saying something.
func wantProblemsOfTheFormat(t *testing.T, problems []checks.Problem) {
	t.Helper()

	for _, problem := range problems {
		if problem.Code != checks.CodeBadStructure {
			t.Errorf("code = %q, want %q", problem.Code, checks.CodeBadStructure)
		}
		if problem.Message == "" {
			t.Error("a problem says nothing")
		}
	}
}

// wantEveryUnreadPartNamed holds a part that could not be read to a problem
// that names it.
func wantEveryUnreadPartNamed(t *testing.T, draft *checks.Draft, problems []checks.Problem) {
	t.Helper()

	unread := map[string]bool{"task": draft.Task == nil, "self_check": draft.SelfCheck == nil}
	for name, missing := range unread {
		if missing && !mentions(problems, name) {
			t.Errorf("%s was not read, and no problem names it: %v", name, problems)
		}
	}
}

// wantOnlyTheFormatsNames holds what a draft was read past and mended in to
// the names of the format, each once.
func wantOnlyTheFormatsNames(t *testing.T, draft *checks.Draft) {
	t.Helper()

	for _, path := range draft.Retired {
		if path != "task.design_thought_process" {
			t.Errorf("Retired names %q, which the format never retired", path)
		}
	}
	mendable := []string{
		"task.correct_answer", "task.options", "task.distractors", "self_check.option_check",
		"self_check.final_answer", "self_check.issues",
	}
	for i, field := range draft.Mended {
		if !slices.Contains(mendable, field) || slices.Contains(draft.Mended[:i], field) {
			t.Errorf("Mended = %v, want fields of the format a mend names, each once", draft.Mended)
		}
	}
}

// The stray check reads a struct's fields one level at a time, which is how
// encoding/json reads them only while a format type embeds nothing and keeps
// nothing unexported. Every type a draft is made of is held to that here, so
// that the day one of them needs more, this says what the stray check has to
// learn first rather than a child's task being refused for a field it had.
func TestTheFormatIsWhatTheStrayCheckCanRead(t *testing.T) {
	t.Parallel()

	seen := map[reflect.Type]bool{}
	var walk func(reflect.Type)
	walk = func(format reflect.Type) {
		for format.Kind() == reflect.Pointer || format.Kind() == reflect.Slice || format.Kind() == reflect.Map {
			format = format.Elem()
		}
		if format.Kind() != reflect.Struct || seen[format] {
			return
		}
		seen[format] = true
		for field := range format.Fields() {
			switch {
			case field.Anonymous:
				t.Errorf("%s embeds %s: the stray check reads one level of fields and would call the embedded "+
					"ones strays — teach it to flatten them first", format, field.Type)
			case !field.IsExported():
				t.Errorf("%s has the unexported field %s: encoding/json never reads it, "+
					"and the stray check would count it as known", format, field.Name)
			}
			walk(field.Type)
		}
	}
	walk(reflect.TypeFor[checks.Draft]())
}
