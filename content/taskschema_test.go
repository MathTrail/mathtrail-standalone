package content_test

import (
	"encoding/json"
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// The schema of a task writes the format of a picture down for everyone who
// is not the service. The service reads a picture with Go, not with the
// schema, so what keeps the two one format is this: the schema names the
// kinds the format has, in its order, and every member of each, no more; it
// accepts every picture the content holds; and it refuses what the format
// refuses, as far as a schema can say it.

// taskSchema is the schema of a task as the content holds it, resolved for a
// validator to read, and its part that describes a picture, resolved alone.
func taskSchema(t *testing.T) (root *jsonschema.Schema, pictures *jsonschema.Resolved) {
	t.Helper()

	raw, found := loaded(t).Schema("task.json")
	if !found {
		t.Fatal("the content holds no schema of a task")
	}
	root = new(jsonschema.Schema)
	if err := json.Unmarshal(raw, root); err != nil {
		t.Fatalf("the schema of a task is not a schema: %v", err)
	}
	if _, err := root.Resolve(nil); err != nil {
		t.Fatalf("the schema of a task does not resolve: %v", err)
	}
	pictures, err := (&jsonschema.Schema{Schema: root.Schema, Ref: "#/$defs/picture", Defs: root.Defs}).Resolve(nil)
	if err != nil {
		t.Fatalf("the schema's picture does not resolve: %v", err)
	}
	return root, pictures
}

// valueOf is a description as the validator reads it.
func valueOf(t *testing.T, description []byte) any {
	t.Helper()

	var value any
	if err := json.Unmarshal(description, &value); err != nil {
		t.Fatalf("the description is no JSON: %v", err)
	}
	return value
}

func TestTheTaskSchemaNamesTheKindsOfPictureInTheFormatsOrder(t *testing.T) {
	t.Parallel()

	root, _ := taskSchema(t)
	described := root.Defs["picture"]
	if described == nil || described.Properties["kind"] == nil {
		t.Fatal("the schema of a task describes no kind of picture")
	}
	var named, chosen []picture.Kind
	for _, kind := range described.Properties["kind"].Enum {
		text, _ := kind.(string)
		named = append(named, picture.Kind(text))
	}
	for _, branch := range described.OneOf {
		chosen = append(chosen, picture.Kind(strings.TrimPrefix(branch.Ref, "#/$defs/")))
	}
	if !slices.Equal(named, picture.Kinds()) {
		t.Errorf("the schema names the kinds %v, want the format's %v", named, picture.Kinds())
	}
	if !slices.Equal(chosen, picture.Kinds()) {
		t.Errorf("the schema describes the kinds %v, want the format's %v", chosen, picture.Kinds())
	}
}

// Every picture the content holds meets the schema: the example of every
// kind, and every reference task's picture.
func TestEveryPictureTheContentHoldsMeetsTheTaskSchema(t *testing.T) {
	t.Parallel()

	_, pictures := taskSchema(t)
	shipped := loaded(t)
	for _, example := range shipped.PictureExamples() {
		if err := pictures.Validate(valueOf(t, example.Picture)); err != nil {
			t.Errorf("the example of a %s does not meet the schema: %v", example.Kind, err)
		}
	}
	for _, task := range shipped.Examples() {
		if task.Picture == nil {
			continue
		}
		if err := pictures.Validate(valueOf(t, task.Picture)); err != nil {
			t.Errorf("%s: the picture does not meet the schema: %v", task.ID, err)
		}
	}
}

// everyMember describes a picture of each kind with every member the kind
// has, down to the members of its items, each within its limits.
var everyMember = []string{
	`{"kind":"clock","time":"16:45","hour_label":"H","minute_label":"M"}`,
	`{"kind":"table","header":["","A","B"],"rows":[["P","9:15","…"],["Q","12","?"]]}`,
	`{"kind":"number_line","from":0,"to":20,"step":2,"marks":[{"at":4,"label":"A"},{"at":10}]}`,
	`{"kind":"row","items":[{"label":"A","below":"1","mark":"ring"},{"skip":true},{"label":"B","mark":"square"}],` +
		`"line":false,"gaps":"3","span":"?","copies":2}`,
	`{"kind":"ring","count":12,"start":{"at":3,"label":"S"}}`,
	`{"kind":"grid","rows":["A","B","C"],"cols":["1","2","3"],"filled":["A1","B2"],"marks":{"C3":"T"}}`,
	`{"kind":"bars","bars":[{"label":"A","parts":4,"shaded":1,"length":40,"value":"?",` +
		`"braces":[{"from":0,"to":2,"label":"12"}],"span":"B"},` +
		`{"label":"C","segments":[{"size":10,"label":"5"},{"size":20}]}],"notes":["A + C = 30"]}`,
	`{"kind":"venn","sets":[{"label":"F","count":"12"},{"label":"S","count":"9"}],"both":"?","neither":"2",` +
		`"total":"25"}`,
	`{"kind":"balance","left":["A","B"],"right":["5"]}`,
	`{"kind":"containers","items":[{"capacity":5,"amount":5,"label":"A"},{"capacity":3,"amount":0,"label":"B"}]}`,
	`{"kind":"piles","piles":[{"label":"A","value":"?","count":30,"shown":6,"group":5,"fill":"light",` +
		`"shape":"square","boxed":true},{"skip":true},{"count":3}],"box":true,"across":true}`,
	`{"kind":"calendar","first":3,"days":30,"week_starts":"sunday","marks":{"14":"?"}}`,
}

// A picture with every member of its kind is one the format reads with no
// problem and the schema accepts, and it holds exactly the members the
// schema names for its kind: a member the schema names and the format does
// not have, or one the format has and the schema leaves out, is found here.
func TestTheTaskSchemaNamesEveryMemberOfEveryKindOfPicture(t *testing.T) {
	t.Parallel()

	_, pictures := taskSchema(t)
	defs := schemaDefs(t)
	described := map[picture.Kind]bool{}
	for _, description := range everyMember {
		kind := picture.Kind(picture.KindOf(json.RawMessage(description)))
		described[kind] = true
		if _, broken := picture.Parse(json.RawMessage(description), picture.Point); len(broken) != 0 {
			t.Errorf("the %s with every member breaks %v", kind, broken)
		}
		value := valueOf(t, []byte(description))
		if err := pictures.Validate(value); err != nil {
			t.Errorf("the %s with every member does not meet the schema: %v", kind, err)
		}
		held, named := map[string]bool{}, map[string]bool{}
		membersHeld(value, string(kind), held)
		membersNamed(defs, defs[string(kind)], string(kind), named)
		if !maps.Equal(held, named) {
			t.Errorf("the %s with every member holds %v, and the schema names %v",
				kind, slices.Sorted(maps.Keys(held)), slices.Sorted(maps.Keys(named)))
		}
	}
	for _, kind := range picture.Kinds() {
		if !described[kind] {
			t.Errorf("no description of a %s with every member, so its members were not held to the schema", kind)
		}
	}
}

// schemaDefs are the definitions of the schema of a task, as much of each as
// the walk over its members reads.
func schemaDefs(t *testing.T) map[string]*schemaNode {
	t.Helper()

	raw, _ := loaded(t).Schema("task.json")
	var document struct {
		Defs map[string]*schemaNode `json:"$defs"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("read the schema's definitions: %v", err)
	}
	return document.Defs
}

// schemaNode is as much of a JSON Schema as the walk over its members reads.
type schemaNode struct {
	Ref        string                 `json:"$ref"`
	Properties map[string]*schemaNode `json:"properties"`
	Items      *schemaNode            `json:"items"`
	Additional json.RawMessage        `json:"additionalProperties"`
	OneOf      []*schemaNode          `json:"oneOf"`
	AnyOf      []*schemaNode          `json:"anyOf"`
}

// membersNamed adds the path of every member a schema names below a path: an
// item of a list adds [] to the path, and a member whose name is the
// description's own to choose, as a cell is, adds *.
func membersNamed(defs map[string]*schemaNode, node *schemaNode, path string, into map[string]bool) {
	if node == nil {
		return
	}
	if node.Ref != "" {
		membersNamed(defs, defs[strings.TrimPrefix(node.Ref, "#/$defs/")], path, into)
	}
	for name, member := range node.Properties {
		into[path+"."+name] = true
		membersNamed(defs, member, path+"."+name, into)
	}
	var keyed schemaNode
	if json.Unmarshal(node.Additional, &keyed) == nil {
		into[path+".*"] = true
		membersNamed(defs, &keyed, path+".*", into)
	}
	membersNamed(defs, node.Items, path+"[]", into)
	for _, branch := range slices.Concat(node.OneOf, node.AnyOf) {
		membersNamed(defs, branch, path, into)
	}
}

// memberName is the name of a member of the format, as against a key the
// description chose, which is a cell or a day.
var memberName = regexp.MustCompile(`^[a-z_]+$`)

// membersHeld adds the path of every member a description holds below a
// path, written as membersNamed writes them.
func membersHeld(value any, path string, into map[string]bool) {
	switch held := value.(type) {
	case map[string]any:
		for name, member := range held {
			if !memberName.MatchString(name) {
				name = "*"
			}
			into[path+"."+name] = true
			membersHeld(member, path+"."+name, into)
		}
	case []any:
		for _, item := range held {
			membersHeld(item, path+"[]", into)
		}
	}
}

// What the format refuses, the schema refuses too, wherever a schema can say
// it: a picture of no kind, a member no kind has, a value of the wrong kind
// or past its limits, and the members a kind takes only together, or never
// together.
func TestThePictureTheFormatRefusesTheTaskSchemaRefuses(t *testing.T) {
	t.Parallel()

	_, pictures := taskSchema(t)
	for _, test := range []struct{ name, description string }{
		{"a kind the format does not have", `{"kind":"pie","slices":3}`},
		{"no kind", `{"time":"4:00"}`},
		{"a member its kind does not have", `{"kind":"clock","time":"4:00","colour":"red"}`},
		{"a label that is a word", `{"kind":"balance","left":["apple"],"right":[]}`},
		{"a label past its length", `{"kind":"venn","sets":[{"label":"ABCDEF"},{"label":"B"}]}`},
		{"a time past the day's end", `{"kind":"clock","time":"24:00"}`},
		{"a number past its limit", `{"kind":"ring","count":25}`},
		{"a number written as a text", `{"kind":"ring","count":"9"}`},
		{"a list past its limit", `{"kind":"bars","bars":[{},{},{},{},{}]}`},
		{"a note that compares nothing", `{"kind":"bars","bars":[{}],"notes":["A + B"]}`},
		{"shaded parts of a bar cut into none", `{"kind":"bars","bars":[{"shaded":1}]}`},
		{"a bar cut into parts and segments", `{"kind":"bars","bars":[{"parts":2,"segments":[{"size":1}]}]}`},
		{"a pile cut short with no count", `{"kind":"piles","piles":[{"shown":2}]}`},
		{"a skip that holds something", `{"kind":"row","items":[{},{"skip":true,"label":"A"},{}]}`},
		{"a mark on no day of the month", `{"kind":"calendar","first":1,"days":30,"marks":{"32":"A"}}`},
		{"a line of a grid named twice", `{"kind":"grid","rows":["A","A"],"cols":["1"]}`},
		{"a cell of a table that is a word", `{"kind":"table","rows":[["cat"]]}`},
		{"a mark that is no dot, ring or square", `{"kind":"row","items":[{"mark":"star"},{}]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if _, broken := picture.Parse(json.RawMessage(test.description), picture.Point); len(broken) == 0 {
				t.Fatal("the format reads it with no problem, so the case shows nothing")
			}
			if err := pictures.Validate(valueOf(t, []byte(test.description))); err == nil {
				t.Error("the schema accepts a picture the format refuses")
			}
		})
	}
}
