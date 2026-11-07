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
		`{"kind":"flags","colors":{"red":"красная","blue":"синяя"},"groups":[{"color":"red","flags":[["red","blue"],` +
			`["red","?"]]},{"label":"B","flags":[["blue","red","blue"]]}]}`,
		`{"kind":"flags","colors":{"red":"\u0301x","purple":1,"blue":["b"]},"groups":[{"flags":["red",[]]}]}`,
		`{"kind":"row","items":[{"skip":true,"label":"\u202e"}],"line":null,"copies":1e400}`,
		`{"kind":"bars","bars":[{"parts":1e9,"segments":[{"size":-1}],"shaded":"3"}]}`,
		`{"kind":"lattice","width":3,"height":2,"cells":[[0,0]],"lines":[[0,0,3,2],[0,2,3,0]],"dashed":[[1,0,1,2]],` +
			`"points":[{"at":[3,2],"label":"P"}],"marks":[{"at":[2,1],"label":"S"}]}`,
		`{"kind":"lattice","width":1e9,"height":-1,"lines":[[0,0,0,0],[1,1]],"points":[{"at":[0,0],"label":"ABC"}]}`,
		`{"kind":"tiles","sides":5,"layout":"ring","tiles":[["1","","","","?"],["","","","",""]]}`,
		`{"kind":"tiles","sides":6,"layout":"row","joint":9,"tiles":[["100"],[1,2,3,4,5,6]]}`,
		`{"kind":"shapes","rows":[{"shapes":["triangle","star"],"total":"6"},{"shapes":[],"total":null}]}`,
		`{"kind":"stairs","steps":8,"marks":[{"step":8,"label":"B"},{"step":8,"label":"C"}],"jumps":[{"from":0,"to":0}]}`,
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
// a picture read without a problem labels only with labels, writes only the
// words a colour may be called by, and shows nothing empty.
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
	labels, words, shown := read.Labels(), read.Words(), read.Shown()
	if len(problems) == 0 {
		writesOnlyWhatItMay(t, labels, words, shown, decimals)
	}
}

// writesOnlyWhatItMay fails the test where a picture read without a problem
// labels with what is no label, writes a word no colour may be called by, or
// shows a value with nothing in it.
func writesOnlyWhatItMay(t *testing.T, labels, words []string, shown []Shown, decimals Decimals) {
	t.Helper()
	for _, label := range labels {
		if !decimals.isLabel(label) {
			t.Fatalf("Labels() holds %q, which is no label", label)
		}
	}
	for _, word := range words {
		if !isColorWord(word) {
			t.Fatalf("Words() holds %q, which no colour may be called by", word)
		}
	}
	for _, value := range shown {
		if value.Face == nil && value.Text == "" {
			t.Fatalf("Shown() holds a value with nothing in it")
		}
	}
}
