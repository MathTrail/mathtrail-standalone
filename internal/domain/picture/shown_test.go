package picture_test

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// read is a well-made description read, or the test stops.
func read(t *testing.T, description string, decimals picture.Decimals) picture.Picture {
	t.Helper()
	read, problems := picture.Parse(json.RawMessage(description), decimals)
	if len(problems) != 0 || read == nil {
		t.Fatalf("Parse(%s) = %v, %v, want a picture and no problems", description, read, problems)
	}
	return read
}

// shownTexts are the texts a picture shows, and the faces of its clocks
// written as H:MM.
func shownTexts(shown []picture.Shown) []string {
	var texts []string
	for _, value := range shown {
		if value.Face != nil {
			texts = append(texts, "face "+faceText(*value.Face))
			continue
		}
		texts = append(texts, value.Text)
	}
	return texts
}

func faceText(face picture.Time) string {
	return fmt.Sprintf("%d:%02d", face.Hour, face.Minute)
}

func TestEachKindSaysWhatItLabelsAndWhatItShows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		description string
		labels      []string
		shown       []string
	}{
		{"a clock shows its time as its face does", `{"kind":"clock","time":"16:30","hour_label":"H"}`,
			[]string{"H"}, []string{"H", "face 4:30"}},
		{"a clock at midnight", `{"kind":"clock","time":"0:05"}`, nil, []string{"face 12:05"}},
		{"a table shows its labels and its times", `{"kind":"table","header":["","A"],` +
			`"rows":[["1","9:15"],["…","?"]]}`, []string{"A", "1", "?"}, []string{"A", "1", "9:15", "?"}},
		{"a number line shows its ends and the ticks its marks stand on",
			`{"kind":"number_line","from":0,"to":10,"step":2,"marks":[{"at":4,"label":"A"},{"at":8}]}`,
			[]string{"A"}, []string{"0", "10", "4", "A", "8"}},
		{"a row counting its items shows the last count alone",
			`{"kind":"row","items":[{"label":"1"},{"label":"2"},{"skip":true},{"label":"9"},{"label":"10","below":"B"}]}`,
			[]string{"1", "2", "9", "10", "B"}, []string{"10", "B"}},
		{"a row counting with no skip", `{"kind":"row","items":[{"label":"1"},{"label":"2"},{"label":"3"}]}`,
			[]string{"1", "2", "3"}, []string{"3"}},
		{"a row that jumps where nothing is cut counts nothing",
			`{"kind":"row","items":[{"label":"1"},{"label":"3"}]}`, []string{"1", "3"}, []string{"1", "3"}},
		{"a row that starts past one counts nothing",
			`{"kind":"row","items":[{"label":"2"},{"label":"3"}]}`, []string{"2", "3"}, []string{"2", "3"}},
		{"a row with an item unlabelled counts nothing",
			`{"kind":"row","items":[{"label":"1"},{}]}`, []string{"1"}, []string{"1"}},
		{"a row of names shows every name and its gaps",
			`{"kind":"row","items":[{"label":"A"},{"label":"B"}],"gaps":"3","span":"?"}`,
			[]string{"3", "?", "A", "B"}, []string{"A", "B", "3", "?"}},
		{"a ring shows its count and where it starts", `{"kind":"ring","count":9,"start":{"at":3,"label":"S"}}`,
			[]string{"S"}, []string{"9", "3", "S"}},
		{"a grid shows the cells it fills and marks, not its lines' names",
			`{"kind":"grid","rows":["A","B"],"cols":["1","2"],"filled":["A1"],"marks":{"B2":"X"}}`,
			[]string{"A", "B", "1", "2", "X"}, []string{"A1", "B2", "X"}},
		{"bars show their parts, shaded parts, labels and notes",
			`{"kind":"bars","bars":[{"label":"A","parts":4,"shaded":1,"length":8,` +
				`"braces":[{"from":0,"to":2,"label":"6"}]},{"label":"B","segments":[{"size":5,"label":"C"}]}],` +
				`"notes":["A + 2.5 = ?"]}`,
			[]string{"A", "6", "B", "C"}, []string{"A", "6", "4", "1", "B", "C", "A", "2.5", "?"}},
		{"two groups show every count", `{"kind":"venn","sets":[{"label":"F","count":"12"},{"label":"C"}],` +
			`"both":"?","total":"30"}`, []string{"F", "12", "C", "?", "30"}, []string{"F", "12", "C", "?", "30"}},
		{"a balance shows what stands on it", `{"kind":"balance","left":["A","5"],"right":["?"]}`,
			[]string{"A", "5", "?"}, []string{"A", "5", "?"}},
		{"containers show what they hold and can hold",
			`{"kind":"containers","items":[{"capacity":5,"amount":0,"label":"A"}]}`,
			[]string{"A"}, []string{"5", "0", "A"}},
		{"piles show their counts, not what they show of them",
			`{"kind":"piles","piles":[{"label":"A","count":12,"shown":6,"group":3},{"skip":true},{"value":"7"}]}`,
			[]string{"A", "7"}, []string{"A", "12", "7"}},
		{"a calendar shows its last day and the days it marks, not its first weekday",
			`{"kind":"calendar","first":3,"days":30,"marks":{"14":"?","2":"A"}}`,
			[]string{"A", "?"}, []string{"30", "2", "A", "14", "?"}},
		{"flags show their groups' labels, how many flags each holds and how many in all",
			`{"kind":"flags","colors":{"red":"red","blue":"blue"},"groups":[{"label":"A","flags":[["red","blue"],` +
				`["blue","red"]]},{"color":"red","flags":[["red","?"],["?","red"],["red","blue","red"]]}]}`,
			[]string{"A"}, []string{"A", "2", "3", "5"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			read := read(t, test.description, picture.Point)
			if got := read.Labels(); !slices.Equal(got, test.labels) {
				t.Errorf("Labels() = %q, want %q", got, test.labels)
			}
			if got := shownTexts(read.Shown()); !slices.Equal(got, test.shown) {
				t.Errorf("Shown() = %q, want %q", got, test.shown)
			}
		})
	}
}

// The words a picture writes are those of the colours it paints with, each
// once, in the order of the palette; a kind that paints nothing writes none.
func TestAPictureWritesTheWordsOfItsColoursAlone(t *testing.T) {
	t.Parallel()
	flags := read(t, `{"kind":"flags","colors":{"yellow":"жёлтая","red":"красная","black":"чёрная"},`+
		`"groups":[{"color":"black","flags":[["red","yellow"],["yellow","red"]]}]}`, picture.Point)
	if got, want := flags.Words(), []string{"красная", "жёлтая", "чёрная"}; !slices.Equal(got, want) {
		t.Errorf("Words() = %q, want %q", got, want)
	}
	for _, test := range wellMade {
		if test.kind == picture.Flags {
			continue
		}
		if got := read(t, test.description, picture.Point).Words(); got != nil {
			t.Errorf("%s: Words() = %q, want none", test.name, got)
		}
	}
}

// A minus that is a number's sign is part of the number a note shows, and one
// that subtracts stands between two of them.
func TestANoteShowsANumberBelowZeroWithItsSign(t *testing.T) {
	t.Parallel()
	read := read(t, `{"kind":"bars","bars":[{"label":"A"}],`+
		`"notes":["A = −3","B − 2 = (-1)","C - 4 > -5"]}`, picture.Point)
	want := []string{"A", "A", "−3", "B", "2", "-1", "C", "4", "-5"}
	if got := shownTexts(read.Shown()); !slices.Equal(got, want) {
		t.Errorf("Shown() = %q, want %q", got, want)
	}
}

func TestANoteIsReadWithTheLessonsDecimalMark(t *testing.T) {
	t.Parallel()
	read := read(t, `{"kind":"bars","bars":[{"label":"A"}],"notes":["A = 2,5"]}`, picture.Comma)
	if got, want := shownTexts(read.Shown()), []string{"A", "A", "2,5"}; !slices.Equal(got, want) {
		t.Errorf("Shown() = %q, want %q", got, want)
	}
}

func TestAPictureReadInPartShowsWhatCouldBeRead(t *testing.T) {
	t.Parallel()
	description := `{"kind":"containers","items":[{"capacity":30,"amount":2,"label":"A"},{"capacity":4,"amount":4}]}`
	read, problems := picture.Parse(json.RawMessage(description), picture.Point)
	if read == nil || len(problems) != 1 {
		t.Fatalf("Parse() = %v, %v, want a picture and one problem", read, problems)
	}
	if got, want := shownTexts(read.Shown()), []string{"2", "A", "4", "4"}; !slices.Equal(got, want) {
		t.Errorf("Shown() = %q, want %q: a capacity that could not be read is not shown", got, want)
	}
}
