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
type parts struct{ brief, task, selfCheck json.RawMessage }

// validParts is the well-formed draft of the structure tests, written out.
func validParts(t testing.TB) parts {
	t.Helper()

	draft := validDraft()
	return parts{
		brief:     marshal(t, draft.Brief),
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
	draft, problems := checks.Decode(submitted.brief, submitted.task, submitted.selfCheck)
	if len(problems) != 0 {
		t.Fatalf("Decode() problems = %v, want none", problems)
	}
	if draft.Brief == nil || draft.Task == nil || draft.SelfCheck == nil {
		t.Fatalf("Decode() = %+v, want every part read", draft)
	}
	if problems := checks.Structure(draft, asked(), testCatalog); len(problems) != 0 {
		t.Fatalf("Structure() after Decode() = %v, want no problems", problems)
	}
}

func TestASubmissionOutOfTheFormatIsRefusedByName(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*parts)
		want   []string
	}{
		{"a part left out", func(p *parts) { p.brief = nil }, []string{"brief is missing"}},
		{"a part sent as null", func(p *parts) { p.selfCheck = json.RawMessage("null") }, []string{"self_check is missing"}},
		{"a part that is not an object", func(p *parts) { p.task = json.RawMessage(`["question"]`) }, []string{"task must be a JSON object"}},
		{"a part that is not JSON", func(p *parts) { p.task = json.RawMessage(`{"question": `) }, []string{"task is not valid JSON"}},
		{"a member the format does not have", func(p *parts) {
			p.task = withMember(t, p.task, "answer", "six pairs")
		}, []string{`task has a member the format does not have: "answer"`}},
		{"every stray member, not only the first", func(p *parts) {
			p.brief = withMember(t, withMember(t, p.brief, "mood", "happy"), "grade", 3)
		}, []string{`"mood"`, `"grade"`}},
		{"a stray member deep inside", func(p *parts) {
			p.task = json.RawMessage(strings.Replace(string(p.task), `"trap":"missed_case"`, `"trap":"missed_case","why":"it is wrong"`, 1))
		}, []string{`task.distractors.* has a member the format does not have: "why"`}},
		{"a number where a string belongs", func(p *parts) {
			p.task = json.RawMessage(strings.Replace(string(p.task), `"trap":"missed_case"`, `"trap":5`, 1))
		}, []string{"task.distractors.*.trap must be a string"}},
		{"a string where a number belongs", func(p *parts) {
			p.brief = json.RawMessage(strings.Replace(string(p.brief), `"difficulty":3`, `"difficulty":"3"`, 1))
		}, []string{"brief.difficulty must be a whole number"}},
		{"a fraction where a whole number belongs", func(p *parts) {
			p.brief = json.RawMessage(strings.Replace(string(p.brief), `"difficulty":3`, `"difficulty":3.5`, 1))
		}, []string{"brief.difficulty must be a whole number"}},
		{"an object where a list belongs", func(p *parts) {
			p.selfCheck = json.RawMessage(strings.Replace(string(p.selfCheck), `"issues":[]`, `"issues":{}`, 1))
		}, []string{"self_check.issues must be a list"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			submitted := validParts(t)
			test.change(&submitted)
			_, problems := checks.Decode(submitted.brief, submitted.task, submitted.selfCheck)
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

// The same stray member in all four explanations is one thing to fix.
func TestAStrayRepeatedAcrossEntriesIsNamedOnce(t *testing.T) {
	t.Parallel()

	submitted := validParts(t)
	task := strings.ReplaceAll(string(submitted.task), `"text":`, `"why":"so","text":`)
	_, problems := checks.Decode(submitted.brief, json.RawMessage(task), submitted.selfCheck)
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
	brief := strings.Replace(string(submitted.brief), `"difficulty":3`, `"difficulty":"3"`, 1)
	brief = strings.Replace(brief, `"setting":"space"`, `"setting":""`, 1)
	draft, problems := checks.Decode(json.RawMessage(brief), submitted.task, submitted.selfCheck)
	if draft.Brief == nil {
		t.Fatal("the brief was dropped over one mistyped member")
	}
	problems = append(problems, checks.Structure(draft, asked(), testCatalog)...)
	if !mentions(problems, "brief.setting") {
		t.Errorf("got %v, want the empty setting found in the same pass", problems)
	}
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
// named.
func FuzzDecode(f *testing.F) {
	valid := validParts(f)
	f.Add([]byte(valid.brief), []byte(valid.task), []byte(valid.selfCheck))
	f.Add([]byte(`{}`), []byte(`{"options":{"A":1}}`), []byte(`null`))
	f.Add([]byte(`[]`), []byte(`{"distractors":{"B":{"trap":[]}}}`), []byte(`{"issues":[{"type":5}]}`))
	f.Add([]byte(`{"difficulty":1e400}`), []byte(`{"drawing_structure":{"objects":[{"value":"x"}]}}`), []byte(`"`))

	f.Fuzz(func(t *testing.T, brief, task, selfCheck []byte) {
		draft, problems := checks.Decode(brief, task, selfCheck)
		problems = append(problems, checks.Structure(draft, asked(), testCatalog)...)
		for _, problem := range problems {
			if problem.Code != checks.CodeBadStructure {
				t.Errorf("code = %q, want %q", problem.Code, checks.CodeBadStructure)
			}
			if problem.Message == "" {
				t.Error("a problem says nothing")
			}
		}
		unread := map[string]bool{"brief": draft.Brief == nil, "task": draft.Task == nil, "self_check": draft.SelfCheck == nil}
		for name, missing := range unread {
			if missing && !mentions(problems, name) {
				t.Errorf("%s was not read, and no problem names it: %v", name, problems)
			}
		}
	})
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
