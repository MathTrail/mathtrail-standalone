package profile_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
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

// Every fixture holds to the schema: every field the schema requires is there,
// at every depth, and every value the schema lists the choices of is one of
// them. The fixtures are what the tests of the whole service stand on, so a
// fixture missing a field the file must carry would let a test pass on a file
// no service writes.
func TestEveryFixtureHoldsToTheSchema(t *testing.T) {
	t.Parallel()

	var schema map[string]any
	if err := json.Unmarshal(profile.Schema(), &schema); err != nil {
		t.Fatalf("the schema is not a JSON document: %v", err)
	}
	defs, _ := schema["$defs"].(map[string]any)

	for _, student := range students {
		t.Run(student, func(t *testing.T) {
			t.Parallel()

			var fixture any
			if err := json.Unmarshal(readFixture(t, student), &fixture); err != nil {
				t.Fatalf("the fixture is not a JSON document: %v", err)
			}
			holds(t, student, fixture, schema, defs)
		})
	}
}

// holds fails the test for every place a value breaks what its piece of the
// schema requires and lists.
func holds(t *testing.T, path string, value any, schema, defs map[string]any) {
	t.Helper()

	if ref, isRef := schema["$ref"].(string); isRef {
		schema, _ = defs[strings.TrimPrefix(ref, "#/$defs/")].(map[string]any)
	}
	if choice, isChoice := schema["oneOf"].([]any); isChoice {
		schema = branchFor(value, choice)
		holds(t, path, value, schema, defs)
		return
	}
	if choices, listed := schema["enum"].([]any); listed && !slices.Contains(choices, value) {
		t.Errorf("%s is %v, want one of %v", path, value, choices)
	}

	switch value := value.(type) {
	case map[string]any:
		holdsObject(t, path, value, schema, defs)
	case []any:
		items, _ := schema["items"].(map[string]any)
		for i, item := range value {
			holds(t, fmt.Sprintf("%s[%d]", path, i), item, items, defs)
		}
	}
}

// holdsObject is holds for an object: the fields its piece of the schema
// requires are there, and each field it has holds to its own piece.
func holdsObject(t *testing.T, path string, object, schema, defs map[string]any) {
	t.Helper()

	required, _ := schema["required"].([]any)
	for _, name := range required {
		if key, _ := name.(string); !hasKey(object, key) {
			t.Errorf("%s has no %s, which the schema requires", path, name)
		}
	}
	properties, _ := schema["properties"].(map[string]any)
	additional, _ := schema["additionalProperties"].(map[string]any)
	for name, below := range object {
		if described, isDescribed := properties[name].(map[string]any); isDescribed {
			holds(t, path+"."+name, below, described, defs)
		} else if additional != nil {
			holds(t, path+"."+name, below, additional, defs)
		}
	}
}

// hasKey reports whether an object names this key at all, even as null.
func hasKey(object map[string]any, key string) bool {
	_, present := object[key]
	return present
}

// branchFor is the branch of a "this or null" choice a value is held to.
func branchFor(value any, choice []any) map[string]any {
	for _, branch := range choice {
		described, _ := branch.(map[string]any)
		if isNull := described["type"] == "null"; isNull == (value == nil) {
			return described
		}
	}
	return map[string]any{}
}
