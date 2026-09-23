package profile_test

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
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
