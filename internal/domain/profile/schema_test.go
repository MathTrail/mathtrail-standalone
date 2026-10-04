package profile_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The schema describes the file for everyone who is not this package. Nothing
// validates against it at runtime, so what keeps it true is this: the fields
// the Go types name and the fields the schema names have to be the same set,
// at every level. A description that drifts from the thing it describes is
// worse than none.

// schemaObject is as much of a JSON Schema as this check reads.
type schemaObject struct {
	Type       json.RawMessage          `json:"type"`
	Properties map[string]*schemaObject `json:"properties"`
	Items      *schemaObject            `json:"items"`
	Additional *schemaObject            `json:"additionalProperties"`
	Ref        string                   `json:"$ref"`
	OneOf      []*schemaObject          `json:"oneOf"`
	Defs       map[string]*schemaObject `json:"$defs"`
}

func TestTheSchemaAndTheTypesNameTheSameFields(t *testing.T) {
	t.Parallel()

	var root schemaObject
	if err := json.Unmarshal(profile.Schema(), &root); err != nil {
		t.Fatalf("the schema is not a JSON document: %v", err)
	}
	compare(t, "profile", reflect.TypeOf(profile.Profile{}), &root, root.Defs)
}

// compare walks a Go type and the piece of schema that describes it together.
func compare(t *testing.T, path string, goType reflect.Type, schema *schemaObject, defs map[string]*schemaObject) {
	t.Helper()

	schema = resolve(t, path, schema, defs)
	if schema == nil {
		return
	}

	switch goType.Kind() {
	case reflect.Pointer:
		compare(t, path, goType.Elem(), schema, defs)
	case reflect.Slice:
		if schema.Items == nil {
			t.Errorf("%s: the schema does not say what the array holds", path)
			return
		}
		compare(t, path+"[]", goType.Elem(), schema.Items, defs)
	case reflect.Map:
		if schema.Additional == nil {
			t.Errorf("%s: the schema does not say what the object holds", path)
			return
		}
		compare(t, path+"{}", goType.Elem(), schema.Additional, defs)
	case reflect.Struct:
		compareStruct(t, path, goType, schema, defs)
	default:
		// A string, a number or a boolean: the schema says its type and there
		// are no field names below it to compare.
	}
}

func compareStruct(t *testing.T, path string, goType reflect.Type, schema *schemaObject, defs map[string]*schemaObject) {
	t.Helper()

	// A type that writes itself — a day, a moment — is a string in the file
	// and has no fields of its own to describe.
	if goType == reflect.TypeOf(profile.Time{}) || goType == reflect.TypeOf(profile.Date{}) {
		return
	}

	named := map[string]reflect.Type{}
	for i := range goType.NumField() {
		field := goType.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		named[name] = field.Type
	}

	for name, fieldType := range named {
		below, described := schema.Properties[name]
		if !described {
			t.Errorf("%s: the types have %q and the schema does not", path, name)
			continue
		}
		compare(t, path+"."+name, fieldType, below, defs)
	}
	for name := range schema.Properties {
		if _, exists := named[name]; !exists {
			t.Errorf("%s: the schema has %q and the types do not", path, name)
		}
	}
}

// resolve follows a reference, and picks the described half of a "this or
// null" pair.
func resolve(t *testing.T, path string, schema *schemaObject, defs map[string]*schemaObject) *schemaObject {
	t.Helper()

	for range 4 { // a reference to a reference is not a shape this file has
		switch {
		case schema.Ref != "":
			name := strings.TrimPrefix(schema.Ref, "#/$defs/")
			found, defined := defs[name]
			if !defined {
				t.Errorf("%s: the schema refers to %q, which it does not define", path, schema.Ref)
				return nil
			}
			schema = found
		case len(schema.OneOf) > 0:
			schema = describedHalf(schema.OneOf)
			if schema == nil {
				t.Errorf("%s: the schema offers a choice with nothing described in it", path)
				return nil
			}
		default:
			return schema
		}
	}
	return schema
}

// describedHalf is the branch of a choice that is not simply null.
func describedHalf(choice []*schemaObject) *schemaObject {
	for _, branch := range choice {
		if !slices.Equal(branch.Type, []byte(`"null"`)) {
			return branch
		}
	}
	return nil
}

// Whatever the schema says, it has to be a schema a reader can follow.
func TestTheSchemaIsOneAReaderCanFollow(t *testing.T) {
	t.Parallel()

	var document map[string]json.RawMessage
	if err := json.Unmarshal(profile.Schema(), &document); err != nil {
		t.Fatalf("the schema is not a JSON document: %v", err)
	}
	for _, key := range []string{"$schema", "title", "description", "type", "required", "properties", "$defs"} {
		if _, said := document[key]; !said {
			t.Errorf("a schema says %q, and this one does not", key)
		}
	}
}

// The limits the schema tells a reader are the limits Validate holds: every
// difficulty runs from MinDifficulty to MaxDifficulty, and every option letter
// is one the solver labels an option with. A limit enforced in one place and
// announced in another drifts apart.
func TestTheSchemaStatesTheLimitsValidateHolds(t *testing.T) {
	t.Parallel()

	var document any
	if err := json.Unmarshal(profile.Schema(), &document); err != nil {
		t.Fatalf("the schema is not a JSON document: %v", err)
	}

	difficulties, letterLists := 0, 0
	walkSchema(document, func(name string, node map[string]any) {
		switch name {
		case "difficulty":
			difficulties++
			if node["minimum"] != float64(profile.MinDifficulty) || node["maximum"] != float64(profile.MaxDifficulty) {
				t.Errorf("a difficulty runs from %v to %v, want %d to %d",
					node["minimum"], node["maximum"], profile.MinDifficulty, profile.MaxDifficulty)
			}
		case "chosen":
			letterLists++
			sameLetters(t, "chosen", node["enum"])
		case "options":
			if required, listed := node["required"]; listed {
				letterLists++
				sameLetters(t, "options", required)
			}
		}
	})
	if difficulties == 0 || letterLists == 0 {
		t.Fatalf("the schema states %d difficulties and %d lists of letters, want some of each", difficulties, letterLists)
	}
}

// The caps the schema tells a reader of the child's details are the caps the
// details are held to. The skills were once capped at fifteen in both, and the
// catalog grew past it with neither noticing.
func TestTheSchemaStatesTheCapsOfTheDetails(t *testing.T) {
	t.Parallel()

	var document struct {
		Defs struct {
			Student struct {
				Properties map[string]map[string]any `json:"properties"`
			} `json:"student"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(profile.Schema(), &document); err != nil {
		t.Fatalf("the schema is not a JSON document: %v", err)
	}
	details := document.Defs.Student.Properties
	interestItems, _ := details["interests"]["items"].(map[string]any)

	for _, tc := range []struct {
		where string
		got   any
		want  int
	}{
		{"excluded_skills.maxItems", details["excluded_skills"]["maxItems"], profile.MaxExcludedSkills},
		{"grade.minimum", details["grade"]["minimum"], profile.MinGrade},
		{"grade.maximum", details["grade"]["maximum"], profile.MaxGrade},
		{"interests.maxItems", details["interests"]["maxItems"], profile.MaxInterests},
		{"interests.items.maxLength", interestItems["maxLength"], profile.MaxInterest},
		{"notes.maxLength", details["notes"]["maxLength"], profile.MaxNotes},
		{"pseudonym.maxLength", details["pseudonym"]["maxLength"], profile.MaxPseudonym},
		{"ui_language.maxLength", details["ui_language"]["maxLength"], profile.MaxLanguageTag},
	} {
		if tc.got != float64(tc.want) {
			t.Errorf("the schema states %s as %v, want %d", tc.where, tc.got, tc.want)
		}
	}
}

// How much history the schema tells a reader the file keeps is how much
// Validate lets it keep: the entries of the window, and the days the week is
// told from.
func TestTheSchemaStatesHowMuchHistoryIsKept(t *testing.T) {
	t.Parallel()

	var document struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(profile.Schema(), &document); err != nil {
		t.Fatalf("the schema is not a JSON document: %v", err)
	}
	for _, tc := range []struct {
		where string
		want  int
	}{
		{"rating_days", profile.RatingDaysKept},
		{"recent", profile.MaxRecent},
	} {
		if got := document.Properties[tc.where]["maxItems"]; got != float64(tc.want) {
			t.Errorf("the schema states %s.maxItems as %v, want %d", tc.where, got, tc.want)
		}
	}
}

// walkSchema calls visit with every named property of a schema, at any depth.
func walkSchema(node any, visit func(name string, node map[string]any)) {
	switch node := node.(type) {
	case map[string]any:
		visitProperties(node, visit)
		for _, child := range node {
			walkSchema(child, visit)
		}
	case []any:
		for _, child := range node {
			walkSchema(child, visit)
		}
	}
}

// visitProperties calls visit with each property one object of a schema names.
func visitProperties(node map[string]any, visit func(name string, node map[string]any)) {
	properties, isObject := node["properties"].(map[string]any)
	if !isObject {
		return
	}
	for name, property := range properties {
		if described, isObject := property.(map[string]any); isObject {
			visit(name, described)
		}
	}
}

// sameLetters fails the test unless a list of the schema holds exactly the
// solver's letters, in order.
func sameLetters(t *testing.T, where string, list any) {
	t.Helper()

	items, isList := list.([]any)
	letters := make([]string, 0, len(items))
	for _, item := range items {
		if letter, isString := item.(string); isString {
			letters = append(letters, letter)
		}
	}
	if !isList || !slices.Equal(letters, solver.Letters()) {
		t.Errorf("%s lists %v, want the letters %v", where, list, solver.Letters())
	}
}

// Every fixture holds to the schema. The fixtures are what the tests of the
// whole service stand on, so a fixture breaking what the file must be would
// let a test pass on a file no service writes.
func TestEveryFixtureHoldsToTheSchema(t *testing.T) {
	t.Parallel()

	for _, student := range students {
		t.Run(student, func(t *testing.T) {
			t.Parallel()

			if err := againstTheSchema(t, readFixture(t, student)); err != nil {
				t.Errorf("the fixture breaks the schema: %v", err)
			}
		})
	}
}

// The schema tells a skipped task from an answer by the fields the file holds:
// an answer has its outcome, and a skipped task has none of it. Either one
// written the other way breaks the schema, and each one as this package writes
// it keeps to it. The service reads a little more than the schema allows — an
// outcome field of false on a skip, or an answer's false left out, decodes the
// same — and writes it back in the schema's shape.
func TestTheSchemaTellsASkippedTaskFromAnAnswer(t *testing.T) {
	t.Parallel()

	where := `"answered_at": "2026-09-20T18:00:00Z", "difficulty": 2, "grade_level": "1-2", "task_id": "tsk_1", "topic": "logic.ordering"`
	outcome := `"confused": false, "correct": true, "hint_used": false, "pace": "fast"`
	for _, tc := range []struct {
		name  string
		entry string
		holds bool
	}{
		{"an answer", "{" + where + ", " + outcome + "}", true},
		{"a skipped task", "{" + where + `, "skipped": true}`, true},
		{"an answer with no outcome", "{" + where + "}", false},
		{"an answer with no correct", "{" + where + `, "confused": false, "hint_used": false, "pace": "fast"}`, false},
		{"a skipped task with an outcome", "{" + where + `, "skipped": true, ` + outcome + "}", false},
		{"a skipped task with a pace", "{" + where + `, "skipped": true, "pace": "slow"}`, false},
		{"a skipped task with a trap", "{" + where + `, "skipped": true, "trap": "missed_case"}`, false},
		{"an entry saying it was not skipped", "{" + where + `, "skipped": false, ` + outcome + "}", false},
		{"an answer with the levels before it", "{" + where + ", " + outcome + `, "before": {"delta": 0.1, "theta": 2.5}}`, true},
		{"a skipped task with the levels before it", "{" + where + `, "skipped": true, "before": {"delta": 0, "theta": 2.5}}`, false},
		{"an answer with half the levels before it", "{" + where + ", " + outcome + `, "before": {"theta": 2.5}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := withWindow(t, parseFixture(t, "dima"), tc.entry)
			if err := againstTheSchema(t, file); (err == nil) != tc.holds {
				t.Errorf("the schema says %v about %s, want it to hold: %v", err, tc.entry, tc.holds)
			}
		})
	}
}

// withWindow is a profile as the file holds it, with one entry for its whole
// window.
func withWindow(t *testing.T, p *profile.Profile, entry string) []byte {
	t.Helper()
	return withPart(t, p, "recent", "["+entry+"]")
}

// withPart is a profile as the file holds it, with one part of it, by its key,
// written as raw says.
func withPart(t *testing.T, p *profile.Profile, key, raw string) []byte {
	t.Helper()

	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	var file map[string]json.RawMessage
	if unread := json.Unmarshal(written, &file); unread != nil {
		t.Fatalf("the file is not a JSON object: %v", unread)
	}
	file[key] = json.RawMessage(raw)
	edited, err := json.Marshal(file)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}
	return edited
}

// The schema and the service hold the history of the ratings to one contract:
// whatever the schema allows, the service reads, and whatever the service
// reads, it writes back in a shape the schema allows. The service reads a
// little more — a null where a field may be left out reads as left out — and
// refuses a number left out as a rule broken, never as bytes that are no
// profile: a store rolls those back, and a file edited by hand would be lost.
func TestTheSchemaAndTheServiceHoldTheHistoryAlike(t *testing.T) {
	t.Parallel()

	answer := `"answered_at": "2026-09-20T18:00:00Z", "confused": false, "correct": true, "difficulty": 2, ` +
		`"grade_level": "1-2", "hint_used": false, "pace": "fast", "task_id": "tsk_1", "topic": "logic.ordering"`
	for _, tc := range []struct {
		name    string
		key     string // the part of the file the case writes
		raw     string
		allowed bool  // whether the schema allows the file
		refused error // what the service refuses it as, or nil when it reads it
	}{
		{"a day", "rating_days", `[{"date": "2026-09-20", "deltas": {"counting.gaps": 0.1}, "theta": 2.5}]`, true, nil},
		{"a day before any topic was answered", "rating_days", `[{"date": "2026-09-20", "deltas": {}, "theta": 2.5}]`, true, nil},
		{"the first day after an answer the history missed", "rating_days",
			`[{"date": "2026-09-20", "deltas": {}, "theta": 2.5, "unkept": "2026-09-18"}]`, true, nil},
		{"a day whose missed answer is null", "rating_days",
			`[{"date": "2026-09-20", "deltas": {}, "theta": 2.5, "unkept": null}]`, false, nil},
		{"a day with no overall level", "rating_days", `[{"date": "2026-09-20", "deltas": {}}]`, false, profile.ErrInvalid},
		{"a day whose corrections are null", "rating_days", `[{"date": "2026-09-20", "deltas": null, "theta": 2.5}]`, false, profile.ErrInvalid},
		{"a day with no corrections", "rating_days", `[{"date": "2026-09-20", "theta": 2.5}]`, false, profile.ErrInvalid},
		{"a correction of no topic", "rating_days", `[{"date": "2026-09-20", "deltas": {"": 0.1}, "theta": 2.5}]`, false, profile.ErrInvalid},
		{"a day with no date", "rating_days", `[{"deltas": {}, "theta": 2.5}]`, false, profile.ErrInvalid},
		{"no day at all", "rating_days", `[null]`, false, profile.ErrInvalid},
		{"a correction that is no number", "rating_days",
			`[{"date": "2026-09-20", "deltas": {"counting.gaps": "0.1"}, "theta": 2.5}]`, false, profile.ErrMalformed},
		{"an answer with the levels before it", "recent", `[{` + answer + `, "before": {"delta": 0.1, "theta": 2.5}}]`, true, nil},
		{"an answer whose levels before it are null", "recent", `[{` + answer + `, "before": null}]`, false, nil},
		{"an answer with half the levels before it", "recent", `[{` + answer + `, "before": {"theta": 2.5}}]`, false, profile.ErrInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := withPart(t, parseFixture(t, "dima"), tc.key, tc.raw)
			if err := againstTheSchema(t, file); (err == nil) != tc.allowed {
				t.Errorf("the schema says %v about %s, want it allowed: %v", err, tc.raw, tc.allowed)
			}
			read, err := profile.Parse(file)
			if tc.refused != nil {
				if !errors.Is(err, tc.refused) {
					t.Errorf("Parse() error = %v about %s, want %v", err, tc.raw, tc.refused)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v about %s, want it read", err, tc.raw)
			}
			wantWrittenAsTheSchemaAllows(t, read)
		})
	}
}

// wantWrittenAsTheSchemaAllows fails the test unless the profile is written
// as a file the schema allows.
func wantWrittenAsTheSchemaAllows(t *testing.T, p *profile.Profile) {
	t.Helper()

	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want the profile written", err)
	}
	if broken := againstTheSchema(t, written); broken != nil {
		t.Errorf("the profile is written as a file the schema refuses: %v", broken)
	}
}

// againstTheSchema holds a file to the schema, read the way any reader of JSON
// Schema reads it: by the library the protocol already brings, which knows
// every keyword the schema uses — the condition that tells a skipped task from
// an answer among them.
func againstTheSchema(t *testing.T, file []byte) error {
	t.Helper()

	var schema jsonschema.Schema
	if err := json.Unmarshal(profile.Schema(), &schema); err != nil {
		t.Fatalf("the schema is not one: %v", err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want a schema that resolves", err)
	}
	var document any
	if err := json.Unmarshal(file, &document); err != nil {
		t.Fatalf("the file is not a JSON document: %v", err)
	}
	return resolved.Validate(document)
}

// "I don't know" is a wrong answer that chose nothing: the schema holds an
// entry marked confused to no right answer, no chosen option and no trap, as
// the service writes it, and a wrong letter keeps both.
func TestTheSchemaHoldsIDontKnowToAWrongAnswerThatChoseNothing(t *testing.T) {
	t.Parallel()

	where := `"answered_at": "2026-09-20T18:00:00Z", "difficulty": 2, "grade_level": "1-2", "task_id": "tsk_1", "topic": "logic.ordering"`
	for _, tc := range []struct {
		name    string
		outcome string
		holds   bool
	}{
		{"I don't know", `"confused": true, "correct": false, "hint_used": true, "pace": "slow"`, true},
		{"a wrong letter", `"chosen": "B", "confused": false, "correct": false, "hint_used": false, "pace": "fast", "trap": "missed_case"`, true},
		{"I don't know that was right", `"confused": true, "correct": true, "hint_used": false, "pace": "fast"`, false},
		{"I don't know that chose an option", `"chosen": "B", "confused": true, "correct": false, "hint_used": false, "pace": "fast"`, false},
		{"I don't know that fell for a trap", `"confused": true, "correct": false, "hint_used": false, "pace": "fast", "trap": "missed_case"`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			entry := "{" + where + ", " + tc.outcome + "}"
			file := withWindow(t, parseFixture(t, "dima"), entry)
			if err := againstTheSchema(t, file); (err == nil) != tc.holds {
				t.Errorf("the schema says %v about %s, want it to hold: %v", err, entry, tc.holds)
			}
		})
	}
}

// The answer a task keeps is described as the service writes it, and nothing
// that is not an answer the child could give holds to it.
func TestTheSchemaDescribesTheAnswerATaskKeeps(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		answered string
		holds    bool
	}{
		{"a letter", `{"choice": "B", "hint_used": false, "level_after": 2.1, "level_before": 2.5, "trial": 0}`, true},
		{"I don't know", `{"choice": "?", "hint_used": true, "level_after": 2.1, "level_before": 2.5, "trial": 4}`, true},
		{"a letter no option has", `{"choice": "F", "hint_used": false, "level_after": 2.1, "level_before": 2.5, "trial": 0}`, false},
		{"a trial past the series", `{"choice": "B", "hint_used": false, "level_after": 2.1, "level_before": 2.5, "trial": 6}`, false},
		{"no level before", `{"choice": "B", "hint_used": false, "level_after": 2.1, "trial": 0}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := withAnswered(t, parseFixture(t, "masha"), tc.answered)
			if err := againstTheSchema(t, file); (err == nil) != tc.holds {
				t.Errorf("the schema says %v about %s, want it to hold: %v", err, tc.answered, tc.holds)
			}
		})
	}
}

// The choices the schema names for the answer a task keeps are the choices a
// child has: the letters of the options, then "I don't know"; and the trial
// series is as long as the rating says it is.
func TestTheSchemaStatesTheChoicesOfAnAnswer(t *testing.T) {
	t.Parallel()

	var document struct {
		Defs struct {
			Given struct {
				Properties map[string]map[string]any `json:"properties"`
			} `json:"given"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(profile.Schema(), &document); err != nil {
		t.Fatalf("the schema is not a JSON document: %v", err)
	}
	given := document.Defs.Given.Properties

	listed, _ := given["choice"]["enum"].([]any)
	want := make([]any, 0, solver.Count+1)
	for _, letter := range solver.Letters() {
		want = append(want, letter)
	}
	want = append(want, profile.DontKnow)
	if !reflect.DeepEqual(listed, want) {
		t.Errorf("the schema lists the choices %v, want %v", listed, want)
	}
	if got := given["trial"]["maximum"]; got != float64(rating.TrialAnswers) {
		t.Errorf("the schema says a trial answer is at most number %v, want %d", got, rating.TrialAnswers)
	}
}

// withAnswered is a profile as the file holds it, with its task on the card
// keeping this answer.
func withAnswered(t *testing.T, p *profile.Profile, answered string) []byte {
	t.Helper()

	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	var file map[string]json.RawMessage
	if unread := json.Unmarshal(written, &file); unread != nil {
		t.Fatalf("the file is not a JSON object: %v", unread)
	}
	var task map[string]json.RawMessage
	if unread := json.Unmarshal(file["current_task"], &task); unread != nil || task == nil {
		t.Fatalf("the file holds no task on the card: %v", unread)
	}
	task["answered"] = json.RawMessage(answered)
	if file["current_task"], err = json.Marshal(task); err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}
	edited, err := json.Marshal(file)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}
	return edited
}
