package mcpserver

import (
	"slices"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

// The input schema names a type for every value it describes, as deep as the
// arguments go, and is otherwise the one the library would serve: an object,
// whether the arguments arrive as a struct or a pointer to one.
func TestTheInputSchemaNamesATypeForEveryValue(t *testing.T) {
	t.Parallel()

	type nested struct {
		Extra any `json:"extra"`
	}
	type arguments struct {
		Part   any            `json:"part"`
		Name   string         `json:"name"`
		Lists  []nested       `json:"lists"`
		Values map[string]any `json:"values"`
	}
	for name, derive := range map[string]func() (*jsonschema.Schema, error){
		"a struct":         inputSchema[arguments],
		"a pointer to one": inputSchema[*arguments],
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schema, err := derive()
			if err != nil {
				t.Fatalf("inputSchema() error = %v, want nil", err)
			}
			if schema.Type != "object" {
				t.Errorf("type = %q, want an object", schema.Type)
			}
			wantAnyJSON(t, "part", schema.Properties["part"])
			wantAnyJSON(t, "lists[].extra", schema.Properties["lists"].Items.Properties["extra"])
			wantAnyJSON(t, "values' values", schema.Properties["values"].AdditionalProperties)
			if name := schema.Properties["name"]; name.Type != "string" || name.Types != nil {
				t.Errorf("name is typed %q %v, want a string as the library has it", name.Type, name.Types)
			}
		})
	}
}

// wantAnyJSON holds a value's schema to taking any JSON: every kind of JSON,
// and, should the value be a list, items that may be anything.
func wantAnyJSON(t *testing.T, path string, got *jsonschema.Schema) {
	t.Helper()

	if want := []string{"object", "string", "array", "number", "boolean", "null"}; !slices.Equal(got.Types, want) {
		t.Errorf("%s is typed %v, want %v", path, got.Types, want)
	}
	if got.Items == nil || got.Items.Type != "" || got.Items.Types != nil {
		t.Errorf("%s holds items %+v, want them open to anything", path, got.Items)
	}
}

// Arguments of any kind are an object of anything, as the library serves them.
func TestTheInputSchemaOfAnyArgumentsIsAnObject(t *testing.T) {
	t.Parallel()

	schema, err := inputSchema[any]()
	if err != nil || schema.Type != "object" || schema.Properties != nil {
		t.Errorf("inputSchema[any]() = %+v, %v, want an object of anything", schema, err)
	}
}
