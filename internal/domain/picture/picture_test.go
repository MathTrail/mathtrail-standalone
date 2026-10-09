package picture_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// wellMade are descriptions of every kind the format accepts, the edges of
// each limit among them.
var wellMade = []struct {
	name        string
	kind        picture.Kind
	description string
}{
	{"a clock", picture.Clock, `{"kind":"clock","time":"4:30"}`},
	{"a clock at the last minute, its hands labelled", picture.Clock,
		`{"kind":"clock","time":"23:59","hour_label":"H","minute_label":"M"}`},
	{"a table with a header", picture.Table,
		`{"kind":"table","header":["","A","B"],"rows":[["1","9:15","…"],["2","?",""]]}`},
	{"a table of one cell", picture.Table, `{"kind":"table","rows":[["?"]]}`},
	{"a table of eight rows of six cells", picture.Table, `{"kind":"table","rows":[` +
		strings.Repeat(`["A","B","C","D","E","F"],`, 7) + `["1","2","3","4","5","6"]]}`},
	{"a number line of sixteen ticks", picture.NumberLine,
		`{"kind":"number_line","from":0,"to":15,"marks":[{"at":0,"label":"A"},{"at":15}]}`},
	{"a number line of two ticks at its farthest", picture.NumberLine,
		`{"kind":"number_line","from":-9999,"to":99999,"step":109998}`},
	{"a number line by threes", picture.NumberLine,
		`{"kind":"number_line","from":-6,"to":12,"step":3,"marks":[{"at":-3,"label":"P"},{"at":9,"label":"?"}]}`},
	{"a row cut short", picture.Row,
		`{"kind":"row","items":[{"label":"A","mark":"ring"},{"skip":true},{"label":"B","below":"3"}],` +
			`"gaps":"2","span":"12","line":false,"copies":2}`},
	{"a row of twelve items", picture.Row, `{"kind":"row","items":[` + strings.Repeat(`{},`, 11) + `{}]}`},
	{"a skip with its item's members left out", picture.Row,
		`{"kind":"row","items":[{},{"skip":true,"label":null,"below":"","mark":null},{}]}`},
	{"a ring", picture.Ring, `{"kind":"ring","count":24,"start":{"at":24,"label":"S"}}`},
	{"a ring of three places, starting at one with no label", picture.Ring,
		`{"kind":"ring","count":3,"start":{"at":1}}`},
	{"a grid", picture.Grid,
		`{"kind":"grid","rows":["A","B","C"],"cols":["1","2","3"],"filled":["A1","C3"],"marks":{"B2":"X"}}`},
	{"a grid of one cell", picture.Grid, `{"kind":"grid","rows":["A"],"cols":["1"]}`},
	{"bars of parts and segments", picture.Bars,
		`{"kind":"bars","bars":[{"label":"A","parts":3,"shaded":3,"braces":[{"from":0,"to":3,"label":"12"}]},` +
			`{"label":"B","segments":[{"size":2,"label":"?"},{"size":60}],"value":"40","span":"?"}],` +
			`"notes":["A + B = 52","A < B","2 × A − 1 > B"]}`},
	{"a bar of twelve parts, none shaded", picture.Bars,
		`{"kind":"bars","bars":[{"parts":12,"shaded":0,"length":60}]}`},
	{"two groups", picture.Venn,
		`{"kind":"venn","sets":[{"label":"F","count":"12"},{"label":"C"}],"both":"?","neither":"3","total":"30"}`},
	{"a balance", picture.Balance, `{"kind":"balance","left":["A","B","5","?"],"right":[]}`},
	{"containers", picture.Containers,
		`{"kind":"containers","items":[{"capacity":5,"amount":0,"label":"A"},{"capacity":20,"amount":20}]}`},
	{"piles", picture.Piles,
		`{"kind":"piles","piles":[{"label":"A","count":40,"shown":39,"group":5,"fill":"light","shape":"square",` +
			`"boxed":true},{"skip":true},{"value":"?"}],"box":true,"across":false}`},
	{"a heap of no stated size", picture.Piles, `{"kind":"piles","piles":[{"label":"H","count":0},{"label":"K"}]}`},
	{"a calendar", picture.Calendar,
		`{"kind":"calendar","first":7,"days":31,"week_starts":"sunday","marks":{"1":"A","31":"?"}}`},
	{"a calendar of February", picture.Calendar, `{"kind":"calendar","first":1,"days":28}`},
	{"members left out as null or empty", picture.Clock,
		`{"kind":"clock","time":"0:00","hour_label":null,"minute_label":""}`},
	{"a whole number written with a decimal point", picture.Ring, `{"kind":"ring","count":9.0}`},
}

func TestEveryKindReadsAWellMadeDescription(t *testing.T) {
	t.Parallel()
	for _, test := range wellMade {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			read, problems := picture.Parse(json.RawMessage(test.description), picture.Point)
			if len(problems) != 0 {
				t.Fatalf("Parse() problems = %v, want none", problems)
			}
			if read == nil || read.Kind() != test.kind {
				t.Fatalf("Parse() = %v, want a picture of kind %q", read, test.kind)
			}
		})
	}
}

func TestEveryKindOfTheFormatIsWellMadeAbove(t *testing.T) {
	t.Parallel()
	for _, kind := range picture.Kinds() {
		found := slices.ContainsFunc(wellMade, func(test struct {
			name        string
			kind        picture.Kind
			description string
		}) bool {
			return test.kind == kind
		})
		if !found {
			t.Errorf("no well-made description of a %s", kind)
		}
	}
}

// breakages are descriptions that break one rule each, and the member that
// breaks it.
var breakages = []struct {
	name        string
	description string
	path        string
	rule        string
}{
	{"no kind", `{"time":"4:30"}`, "picture.kind", "is missing"},
	{"a kind the format does not have", `{"kind":"clocks"}`, "picture.kind", "is no kind of picture"},
	{"a kind that is no text", `{"kind":3}`, "picture.kind", "is no kind of picture"},

	{"a clock with no time", `{"kind":"clock"}`, "picture.time", "is missing"},
	{"a time past midnight", `{"kind":"clock","time":"24:00"}`, "picture.time", "must be a time"},
	{"a time with a lone minute", `{"kind":"clock","time":"4:5"}`, "picture.time", "must be a time"},
	{"a time as a number", `{"kind":"clock","time":430}`, "picture.time", "must be a time"},
	{"a hand's label in words", `{"kind":"clock","time":"4:30","hour_label":"hour"}`, "picture.hour_label",
		"must be a label"},
	{"a member a clock does not have", `{"kind":"clock","time":"4:30","colour":"red"}`, "picture.colour",
		"is no member of a clock"},
	{"a member named as no member could be", `{"kind":"clock","time":"4:30","Ben":1}`, "picture.*",
		"is no member of a clock"},

	{"a table with no rows", `{"kind":"table"}`, "picture.rows", "is missing"},
	{"a table of nine rows", `{"kind":"table","rows":[` + strings.Repeat(`["1"],`, 8) + `["1"]]}`,
		"picture.rows", "must hold 1 to 8 rows"},
	{"a row of seven cells", `{"kind":"table","rows":[["1","2","3","4","5","6","7"]]}`, "picture.rows.0",
		"must be a list of 1 to 6 cells"},
	{"a row shorter than the header", `{"kind":"table","header":["A","B"],"rows":[["1","2"],["3"]]}`,
		"picture.rows.1", "must be as long as the header"},
	{"a row shorter than the first", `{"kind":"table","rows":[["1","2"],["3"]]}`, "picture.rows.1",
		"must be as long as the header"},
	{"a cell in words", `{"kind":"table","rows":[["1","nine"]]}`, "picture.rows.0.1", "must be a cell"},

	{"a line ending before it starts", `{"kind":"number_line","from":5,"to":5}`, "picture.to",
		"must be greater than from"},
	{"a line its step does not reach the end of", `{"kind":"number_line","from":0,"to":10,"step":3}`,
		"picture.step", "must reach to from from in whole steps"},
	{"a line of seventeen ticks", `{"kind":"number_line","from":0,"to":16}`, "picture.step",
		"must make 2 to 16 ticks"},
	{"a tick's number past five characters", `{"kind":"number_line","from":0,"to":100000,"step":100000}`,
		"picture.to", "must be a whole number from -9999 to 99999"},
	{"a mark between ticks", `{"kind":"number_line","from":0,"to":10,"step":2,"marks":[{"at":3}]}`,
		"picture.marks.0.at", "must stand on a tick of the line"},
	{"a mark past the end", `{"kind":"number_line","from":0,"to":10,"marks":[{"at":11}]}`,
		"picture.marks.0.at", "must stand on a tick of the line"},
	{"two marks on a tick", `{"kind":"number_line","from":0,"to":10,"marks":[{"at":3},{"at":3,"label":"B"}]}`,
		"picture.marks.1.at", "must stand on a tick no other mark stands on"},
	{"a mark standing nowhere", `{"kind":"number_line","from":0,"to":10,"marks":[{"label":"A"}]}`,
		"picture.marks.0.at", "is missing"},

	{"a row of one item", `{"kind":"row","items":[{}]}`, "picture.items", "must hold 2 to 12 items"},
	{"a row of thirteen items", `{"kind":"row","items":[` + strings.Repeat(`{},`, 12) + `{}]}`,
		"picture.items", "must hold 2 to 12 items"},
	{"a flood of items, each out of the format", `{"kind":"row","items":[` + strings.Repeat(`7,`, 99) + `7]}`,
		"picture.items", "must hold 2 to 12 items"},
	{"a skip with a member no item has", `{"kind":"row","items":[{},{"skip":true,"colour":null},{}]}`,
		"picture.items.1.colour", "is no member of an item of a row"},
	{"a skip among piles with a member the model named", `{"kind":"piles","piles":[{},{"skip":true,"Ben":""},{}]}`,
		"picture.piles.1.*", "is no member of a pile"},
	{"a number grouped by thousands", `{"kind":"balance","left":["1,000"],"right":[]}`, "picture.left.0",
		"must be a label: a Latin capital or a run of them, a number written with the question's decimal mark and " +
			"no separator between thousands"},
	{"a row that starts with a skip", `{"kind":"row","items":[{"skip":true},{},{}]}`, "picture.items.0",
		"is a skip, which is never first or last"},
	{"a row that ends with a skip", `{"kind":"row","items":[{},{},{"skip":true}]}`, "picture.items.2",
		"is a skip, which is never first or last"},
	{"two skips running", `{"kind":"row","items":[{},{"skip":true},{"skip":true},{}]}`, "picture.items.2",
		"is a skip next to another skip"},
	{"a skip with a label", `{"kind":"row","items":[{},{"skip":true,"label":"A"},{}]}`, "picture.items.1",
		"is a skip, which holds nothing else"},
	{"a mark of no shape a row has", `{"kind":"row","items":[{"mark":"star"},{}]}`, "picture.items.0.mark",
		"must be dot, ring or square"},
	{"a row drawn three times", `{"kind":"row","items":[{},{}],"copies":3}`, "picture.copies",
		"must be a whole number from 1 to 2"},
	{"a line that is neither true nor false", `{"kind":"row","items":[{},{}],"line":"yes"}`, "picture.line",
		"must be true or false"},
	{"an item that is no object", `{"kind":"row","items":[{},"A"]}`, "picture.items.1", "must be an object"},

	{"a ring of two places", `{"kind":"ring","count":2}`, "picture.count", "must be a whole number from 3 to 24"},
	{"a ring starting at no place", `{"kind":"ring","count":5,"start":{"at":0}}`, "picture.start.at",
		"must be a place of the ring"},
	{"a ring starting past its last place", `{"kind":"ring","count":5,"start":{"at":6}}`, "picture.start.at",
		"must be a place of the ring"},
	{"a start that is no object", `{"kind":"ring","count":5,"start":3}`, "picture.start", "must be an object"},

	{"a grid of nine rows", `{"kind":"grid","rows":["A","B","C","D","E","F","G","H","I"],"cols":["1"]}`,
		"picture.rows", "must hold 1 to 8 labels"},
	{"a row named twice", `{"kind":"grid","rows":["A","A"],"cols":["1"]}`, "picture.rows.1",
		"names a line of the grid another name already names"},
	{"two cells of one name", `{"kind":"grid","rows":["1","12"],"cols":["22","2"],"filled":["122"]}`, "picture.cols",
		"must name the columns apart from the rows"},
	{"two cells of one name in capitals", `{"kind":"grid","rows":["A","AB"],"cols":["BC","C"]}`, "picture.cols",
		"must name the columns apart from the rows"},
	{"a column in words", `{"kind":"grid","rows":["A"],"cols":["one"]}`, "picture.cols.0", "must be a label"},
	{"a filled cell off the grid", `{"kind":"grid","rows":["A"],"cols":["1"],"filled":["B1"]}`,
		"picture.filled.0", "must be a cell of the grid"},
	{"a cell filled twice", `{"kind":"grid","rows":["A"],"cols":["1","2"],"filled":["A1","A1"]}`,
		"picture.filled.1", "fills a cell already filled"},
	{"a mark off the grid", `{"kind":"grid","rows":["A"],"cols":["1"],"marks":{"Z9":"X"}}`, "picture.marks.*",
		"must be a cell of the grid"},
	{"marks as a list", `{"kind":"grid","rows":["A"],"cols":["1"],"marks":["A1"]}`, "picture.marks",
		"must be an object"},
	{"a mark in words", `{"kind":"grid","rows":["A"],"cols":["1"],"marks":{"A1":"start"}}`, "picture.marks.*",
		"must be a label"},

	{"bars with no bar", `{"kind":"bars"}`, "picture.bars", "is missing"},
	{"five bars", `{"kind":"bars","bars":[{},{},{},{},{}]}`, "picture.bars", "must hold 1 to 4 bars"},
	{"a bar of thirteen parts", `{"kind":"bars","bars":[{"parts":13}]}`, "picture.bars.0.parts",
		"must be a whole number from 1 to 12"},
	{"more parts shaded than there are", `{"kind":"bars","bars":[{"parts":3,"shaded":4}]}`,
		"picture.bars.0.shaded", "must be a whole number from 0 up to the bar's parts"},
	{"parts shaded on a bar of no parts", `{"kind":"bars","bars":[{"shaded":1}]}`, "picture.bars.0.shaded",
		"comes with parts"},
	{"a bar of parts and segments", `{"kind":"bars","bars":[{"parts":2,"segments":[{"size":1}]}]}`,
		"picture.bars.0.segments", "come where parts are"},
	{"a segment too long", `{"kind":"bars","bars":[{"segments":[{"size":61}]}]}`,
		"picture.bars.0.segments.0.size", "must be a whole number from 1 to 60"},
	{"a bar too long", `{"kind":"bars","bars":[{"length":0}]}`, "picture.bars.0.length",
		"must be a whole number from 1 to 60"},
	{"a brace past its bar", `{"kind":"bars","bars":[{"parts":3,"braces":[{"from":1,"to":4,"label":"A"}]}]}`,
		"picture.bars.0.braces.0", "must lie within its bar"},
	{"a brace over nothing", `{"kind":"bars","bars":[{"parts":3,"braces":[{"from":2,"to":2,"label":"A"}]}]}`,
		"picture.bars.0.braces.0", "must lie within its bar"},
	{"a brace past the segments", `{"kind":"bars","bars":[{"segments":[{"size":4}],` +
		`"braces":[{"from":0,"to":2,"label":"A"}]}]}`, "picture.bars.0.braces.0", "must lie within its bar"},
	{"a brace with no label", `{"kind":"bars","bars":[{"parts":3,"braces":[{"from":0,"to":3}]}]}`,
		"picture.bars.0.braces.0.label", "is missing"},
	{"five braces", `{"kind":"bars","bars":[{"parts":5,"braces":[` +
		strings.Repeat(`{"from":0,"to":1,"label":"A"},`, 4) + `{"from":0,"to":1,"label":"A"}]}]}`,
		"picture.bars.0.braces", "must hold at most 4 braces"},
	{"four notes", `{"kind":"bars","bars":[{}],"notes":["A = 1","A = 1","A = 1","A = 1"]}`, "picture.notes",
		"must hold at most 3 notes"},
	{"a note that equals nothing", `{"kind":"bars","bars":[{}],"notes":["A + B"]}`, "picture.notes.0",
		"must be a note"},
	{"a note in words", `{"kind":"bars","bars":[{}],"notes":["A is 3 = B"]}`, "picture.notes.0", "must be a note"},
	{"a note past its length", `{"kind":"bars","bars":[{}],"notes":["A + B + C + D + E + F = 21"]}`,
		"picture.notes.0", "must be a note"},

	{"one group", `{"kind":"venn","sets":[{"label":"A"}]}`, "picture.sets", "must hold exactly 2 sets"},
	{"a group with no label", `{"kind":"venn","sets":[{"label":"A"},{"count":"3"}]}`, "picture.sets.1.label",
		"is missing"},
	{"a count in words", `{"kind":"venn","sets":[{"label":"A"},{"label":"B"}],"total":"many"}`,
		"picture.total", "must be a label"},

	{"five on a pan", `{"kind":"balance","left":["A","B","C","D","E"],"right":[]}`, "picture.left",
		"must hold at most 4 labels"},
	{"a balance with one pan", `{"kind":"balance","left":["A"]}`, "picture.right", "is missing"},

	{"a container too big", `{"kind":"containers","items":[{"capacity":21,"amount":0}]}`,
		"picture.items.0.capacity", "must be a whole number from 1 to 20"},
	{"a container holding more than it can", `{"kind":"containers","items":[{"capacity":3,"amount":4}]}`,
		"picture.items.0.amount", "must be a whole number from 0 up to the container's capacity"},
	{"a container holding nothing said", `{"kind":"containers","items":[{"capacity":3}]}`,
		"picture.items.0.amount", "is missing"},
	{"five containers", `{"kind":"containers","items":[` + strings.Repeat(`{"capacity":1,"amount":0},`, 4) +
		`{"capacity":1,"amount":0}]}`, "picture.items", "must hold 1 to 4 containers"},

	{"seven piles", `{"kind":"piles","piles":[{},{},{},{},{},{},{}]}`, "picture.piles", "must hold 1 to 6 piles"},
	{"a pile too big", `{"kind":"piles","piles":[{"count":41}]}`, "picture.piles.0.count",
		"must be a whole number from 0 to 40"},
	{"a pile cut short of no stated size", `{"kind":"piles","piles":[{"shown":4}]}`, "picture.piles.0.shown",
		"comes with count"},
	{"a pile cut short that shows all", `{"kind":"piles","piles":[{"count":5,"shown":5}]}`,
		"picture.piles.0.shown", "must be a whole number from 2 to one less than the pile's count"},
	{"a pile cut short to one counter", `{"kind":"piles","piles":[{"count":5,"shown":1}]}`,
		"picture.piles.0.shown", "must be a whole number from 2 to one less than the pile's count"},
	{"a pile in groups of none", `{"kind":"piles","piles":[{"count":5,"group":0}]}`, "picture.piles.0.group",
		"must be a whole number from 1 to 40"},
	{"a pile of a colour", `{"kind":"piles","piles":[{"fill":"red"}]}`, "picture.piles.0.fill",
		"must be dark or light"},
	{"piles that start with a skip", `{"kind":"piles","piles":[{"skip":true},{}]}`, "picture.piles.0",
		"is a skip, which is never first or last"},

	{"a month starting on no weekday", `{"kind":"calendar","first":0,"days":30}`, "picture.first",
		"must be a whole number from 1 to 7"},
	{"a month of thirty-two days", `{"kind":"calendar","first":1,"days":32}`, "picture.days",
		"must be a whole number from 28 to 31"},
	{"weeks starting on a Friday", `{"kind":"calendar","first":1,"days":30,"week_starts":"friday"}`,
		"picture.week_starts", "must be monday or sunday"},
	{"a mark past the month's end", `{"kind":"calendar","first":1,"days":30,"marks":{"31":"A"}}`,
		"picture.marks.*", "must be a day of the month"},
	{"a mark on no day", `{"kind":"calendar","first":1,"days":30,"marks":{"0":"A"}}`, "picture.marks.*",
		"must be a day of the month"},
	{"a day written with a nought before it", `{"kind":"calendar","first":1,"days":30,"marks":{"1":"A","01":"B"}}`,
		"picture.marks.*", "must be a day of the month"},

	{"a count written as text", `{"kind":"ring","count":"9"}`, "picture.count", "must be a whole number"},
	{"a count with decimals", `{"kind":"ring","count":3.5}`, "picture.count", "must be a whole number"},
	{"a label written as a number", `{"kind":"row","items":[{},{}],"gaps":2}`, "picture.gaps", "must be a label"},
	{"a label of six capitals", `{"kind":"row","items":[{},{}],"span":"ABCDEF"}`, "picture.span",
		"must be a label"},
	{"a label in small letters", `{"kind":"row","items":[{"label":"a"},{}]}`, "picture.items.0.label",
		"must be a label"},
	{"a label in another alphabet", `{"kind":"row","items":[{"label":"А"},{}]}`, "picture.items.0.label",
		"must be a label"},
	{"a number grouping its thousands", `{"kind":"row","items":[{},{}],"span":"1,000"}`, "picture.span",
		"must be a label"},
	{"a decimal comma where decimals take a point", `{"kind":"row","items":[{},{}],"span":"2,5"}`,
		"picture.span", "must be a label"},
}

func TestEachRuleIsOneProblemAtItsMember(t *testing.T) {
	t.Parallel()
	for _, test := range breakages {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, problems := picture.Parse(json.RawMessage(test.description), picture.Point)
			if len(problems) != 1 {
				t.Fatalf("Parse() problems = %v, want one at %s", problems, test.path)
			}
			if problems[0].Path != test.path || !strings.HasPrefix(problems[0].Rule, test.rule) {
				t.Errorf("Parse() problem = %+v, want %s %s…", problems[0], test.path, test.rule)
			}
		})
	}
}

func TestADescriptionThatIsNoObjectHasNoPicture(t *testing.T) {
	t.Parallel()
	for _, description := range []string{`"clock 4:30"`, `["clock"]`, `4`, `null`, `{"kind":`, `{} {}`, ``} {
		read, problems := picture.Parse(json.RawMessage(description), picture.Point)
		if read != nil || len(problems) != 1 || problems[0].Path != "picture" {
			t.Errorf("Parse(%s) = %v, %v, want no picture and one problem at picture", description, read, problems)
		}
	}
}

func TestAKindTheFormatDoesNotHaveIsNotReadFurther(t *testing.T) {
	t.Parallel()
	read, problems := picture.Parse(json.RawMessage(`{"kind":"star","points":"many","colour":"gold"}`), picture.Point)
	if read != nil || len(problems) != 1 || problems[0].Path != "picture.kind" {
		t.Errorf("Parse() = %v, %v, want no picture and one problem at picture.kind", read, problems)
	}
}

func TestANumberIsWrittenWithTheLessonsDecimalMark(t *testing.T) {
	t.Parallel()
	tests := []struct {
		span     string
		decimals picture.Decimals
		want     bool
	}{
		{"2.5", picture.Point, true},
		{"2,5", picture.Comma, true},
		{"2,5", picture.Point, false},
		{"2.5", picture.Comma, false},
		{"-3", picture.Point, true},
		{"−3", picture.Comma, true},
		{"007", picture.Point, true},
		{"-0", picture.Point, true},
		{"1,000", picture.Point, false},
		{"2.", picture.Point, false},
		{".5", picture.Point, false},
	}
	for _, test := range tests {
		description := `{"kind":"row","items":[{},{}],"span":"` + test.span + `"}`
		_, problems := picture.Parse(json.RawMessage(description), test.decimals)
		if got := len(problems) == 0; got != test.want {
			t.Errorf("Parse(span %q, %v) accepted = %v, want %v", test.span, test.decimals, got, test.want)
		}
	}
}

func TestEveryProblemOfAPictureStaysWithinIt(t *testing.T) {
	t.Parallel()
	for _, test := range breakages {
		_, problems := picture.Parse(json.RawMessage(test.description), picture.Point)
		for _, problem := range problems {
			if !strings.HasPrefix(problem.Path, "picture") {
				t.Errorf("%s: problem path = %q, want one within picture", test.name, problem.Path)
			}
		}
	}
}

func TestKindsAreTheTwelveTheFormatLists(t *testing.T) {
	t.Parallel()
	want := []picture.Kind{
		picture.Clock, picture.Table, picture.NumberLine, picture.Row, picture.Ring, picture.Grid,
		picture.Bars, picture.Venn, picture.Balance, picture.Containers, picture.Piles, picture.Calendar,
	}
	if got := picture.Kinds(); !slices.Equal(got, want) {
		t.Errorf("Kinds() = %v, want %v", got, want)
	}
}

func TestEveryKindSaysItsLimits(t *testing.T) {
	t.Parallel()
	for _, kind := range picture.Kinds() {
		if picture.LimitsOf(kind) == "" {
			t.Errorf("LimitsOf(%q) is empty, want a sentence", kind)
		}
	}
	if limits := picture.LimitsOf("star"); limits != "" {
		t.Errorf("LimitsOf(star) = %q, want nothing for a kind the format does not have", limits)
	}
}

func TestKindOfNamesWhatALogLineMayCarry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		description string
		want        string
	}{
		{``, picture.None},
		{`null`, picture.None},
		{`  `, picture.None},
		{`{"kind":"clock","time":"25:99"}`, "clock"},
		{`{"kind":"number_line"}`, "number_line"},
		{`{"kind":"star"}`, picture.Other},
		{`{"time":"4:30"}`, picture.Other},
		{`{"kind":3}`, picture.Other},
		{`["clock"]`, picture.Other},
		{`"clock"`, picture.Other},
		{`{"kind":`, picture.Other},
	}
	for _, test := range tests {
		if got := picture.KindOf(json.RawMessage(test.description)); got != test.want {
			t.Errorf("KindOf(%s) = %q, want %q", test.description, got, test.want)
		}
	}
}

func TestATimeIsReadAsAFaceOfTwelveHoursShowsIt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		written string
		want    picture.Time
	}{
		{"0:00", picture.Time{Hour: 12, Minute: 0}},
		{"00:15", picture.Time{Hour: 12, Minute: 15}},
		{"4:30", picture.Time{Hour: 4, Minute: 30}},
		{"16:30", picture.Time{Hour: 4, Minute: 30}},
		{"12:05", picture.Time{Hour: 12, Minute: 5}},
		{"23:59", picture.Time{Hour: 11, Minute: 59}},
	}
	for _, test := range tests {
		read, isTime := picture.ReadTime(test.written)
		if !isTime || read.OnTheFace() != test.want {
			t.Errorf("ReadTime(%q).OnTheFace() = %v, %v, want %v", test.written, read.OnTheFace(), isTime, test.want)
		}
	}
	for _, written := range []string{"24:00", "4:60", "4:5", "430", "4.30", "", " 4:30", "4:30 pm", "-1:30"} {
		if _, isTime := picture.ReadTime(written); isTime {
			t.Errorf("ReadTime(%q) is a time, want none", written)
		}
	}
}
