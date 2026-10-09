package picture

import (
	"encoding/json"
	"strings"
	"testing"
)

// A description comes from the chat's model and is untrusted: whatever it
// holds, reading it ends in a picture or in problems, every problem within the
// picture, and what the picture says it labels and shows can be asked of it.
func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		`{"kind":"clock","time":"4:30","hour_label":"H"}`,
		`{"kind":"table","header":["","A"],"rows":[["1","9:15"],["…","?"]]}`,
		`{"kind":"number_line","from":0,"to":10,"step":2,"marks":[{"at":4,"label":"A"}]}`,
		`{"kind":"row","items":[{"label":"1"},{"skip":true},{"label":"9"}],"gaps":"2"}`,
		`{"kind":"ring","count":9,"start":{"at":3,"label":"S"}}`,
		`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"filled":["A1"],"marks":{"B2":"X"}}`,
		`{"kind":"bars","bars":[{"parts":4,"shaded":1,"braces":[{"from":0,"to":2,"label":"6"}]}],"notes":["A = 2,5"]}`,
		`{"kind":"venn","sets":[{"label":"F"},{"label":"C"}],"both":"?"}`,
		`{"kind":"balance","left":["A"],"right":[]}`,
		`{"kind":"containers","items":[{"capacity":5,"amount":6}]}`,
		`{"kind":"piles","piles":[{"count":12,"shown":6},{"skip":true},{}]}`,
		`{"kind":"calendar","first":3,"days":30,"marks":{"14":"?","x":"A"}}`,
		`{"kind":"row","items":[{"skip":true,"label":"\u202e"}],"line":null,"copies":1e400}`,
		`{"kind":"bars","bars":[{"parts":1e9,"segments":[{"size":-1}],"shaded":"3"}]}`,
		`[{"kind":"clock"}]`, `{"kind":"clock","time":"9:99"}`, `{`,
	} {
		f.Add([]byte(seed), false)
		f.Add([]byte(seed), true)
	}

	f.Fuzz(func(t *testing.T, raw []byte, comma bool) {
		decimals := Point
		if comma {
			decimals = Comma
		}
		read, problems := Parse(json.RawMessage(raw), decimals)
		holdsItsRules(t, read, problems, decimals)
	})
}

// holdsItsRules fails the test where a reading breaks what every reading holds:
// a picture or a reason there is none, every problem within the picture, and
// a picture read without a problem labels only with labels and shows nothing
// empty.
func holdsItsRules(t *testing.T, read Picture, problems []Problem, decimals Decimals) {
	t.Helper()
	if read == nil && len(problems) == 0 {
		t.Fatalf("Parse() = no picture and no problem, want a reason there is none")
	}
	for _, problem := range problems {
		if !strings.HasPrefix(problem.Path, "picture") || problem.Rule == "" {
			t.Fatalf("problem = %+v, want a rule at a member of the picture", problem)
		}
	}
	if read == nil {
		return
	}
	labels, shown := read.Labels(), read.Shown()
	if len(problems) > 0 {
		return
	}
	for _, label := range labels {
		if !decimals.isLabel(label) {
			t.Fatalf("Labels() holds %q, which is no label", label)
		}
	}
	for _, value := range shown {
		if value.Face == nil && value.Text == "" {
			t.Fatalf("Shown() holds a value with nothing in it")
		}
	}
}
