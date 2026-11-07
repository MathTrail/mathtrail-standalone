package picture_test

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// The facts of a picture are the members a program proving its answer has to
// read: neither its kind, nor what only lays it out, nor what holds nothing.
func TestTheFactsOfAPicture(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, description string
		want              []string
	}{
		{"no picture", ``, nil},
		{"a description that is no object", `[{"kind":"row"}]`, nil},
		{"the kind alone", `{"kind":"row"}`, nil},
		{"every member that holds something", `{"kind":"grid","rows":["A","B"],"cols":["1"],"filled":["A1"],"marks":{"B1":"X"}}`,
			[]string{"cols", "filled", "marks", "rows"}},
		{"what lays a row out", `{"kind":"row","items":[{"label":"P"}],"copies":2,"line":true}`, []string{"items"}},
		{"what lays piles out", `{"kind":"piles","piles":[{"count":3}],"box":true,"across":true}`, []string{"piles"}},
		{"what lays a row out is a fact of another kind", `{"kind":"table","copies":2,"rows":[["1"]]}`, []string{"copies", "rows"}},
		{"members of no kind the format has", `{"kind":"lattice","width":3,"lines":[[0,0,1,0]]}`, []string{"lines", "width"}},
		{"a member that is the mark of what is asked", `{"kind":"row","items":[{}],"gaps":"?"}`, []string{"items"}},
		{"members that hold nothing", `{"kind":"venn","sets":[],"both":null,"neither":"","total":{},"notes":["?",""]}`, nil},
		{"an object with a fact inside", `{"kind":"ring","count":6,"start":{"at":3}}`, []string{"count", "start"}},
		{"an object with nothing inside", `{"kind":"ring","count":6,"start":{"label":"?"}}`, []string{"count"}},
		{"a list of empty objects, which counts them", `{"kind":"row","items":[{},{},{}]}`, []string{"items"}},
		{"a zero and a flag", `{"kind":"number_line","from":0,"step":false}`, []string{"from", "step"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := picture.Facts(json.RawMessage(test.description)); !slices.Equal(got, test.want) {
				t.Errorf("Facts(%s) = %q, want %q", test.description, got, test.want)
			}
		})
	}
}

// A description comes from the chat's model and is untrusted: whatever it
// holds, its facts are members it has, each once, in order, and never its
// kind.
func FuzzFacts(f *testing.F) {
	for _, seed := range []string{
		`{"kind":"row","items":[{"label":"1"},{"skip":true},{"label":"9"}],"gaps":"2","copies":1e400}`,
		`{"kind":"piles","piles":[{"count":12,"shown":6},{"skip":true},{}],"box":"?"}`,
		`{"kind":"calendar","first":3,"days":30,"marks":{"14":"?","x":"A"}}`,
		`{"kind":7,"kind ":"row","":null}`, `[{"kind":"clock"}]`, `{`, `null`, `{"a":[[["?"]]],"b":[null,""]}`,
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		facts := picture.Facts(raw)
		// The description is read as the format reads one: its first value.
		var members map[string]json.RawMessage
		if json.NewDecoder(bytes.NewReader(raw)).Decode(&members) != nil {
			members = nil
		}
		if !slices.IsSorted(facts) || len(slices.Compact(slices.Clone(facts))) != len(facts) {
			t.Fatalf("Facts(%q) = %q, want each member once, sorted", raw, facts)
		}
		for _, name := range facts {
			if _, held := members[name]; !held || name == "kind" {
				t.Fatalf("Facts(%q) = %q, holding %q, want members the description has other than its kind", raw, facts, name)
			}
		}
	})
}
