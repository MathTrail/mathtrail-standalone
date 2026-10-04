package content

import (
	"bytes"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

const (
	topicsFile = catalogsDir + "/topics.json"
	trapsFile  = catalogsDir + "/traps.json"
	skillsFile = catalogsDir + "/skills.json"
)

// withEntry puts one more entry at the front of a catalog and leaves the rest
// of it alone, so that a test breaks one entry rather than the whole file.
func withEntry(t *testing.T, src fstest.MapFS, file, entry string) {
	t.Helper()

	raw := src[file]
	if raw == nil {
		t.Fatalf("%s is not in the content", file)
	}
	_, rest, found := bytes.Cut(raw.Data, []byte("["))
	if !found {
		t.Fatalf("%s does not hold a JSON array", file)
	}
	src[file] = &fstest.MapFile{Data: append([]byte("["+entry+","), rest...)}
}

func TestBrokenCatalogStopsTheService(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		file  string
		entry string
		want  string
	}{
		{
			name:  "a topic id that is not an area and a topic",
			file:  topicsFile,
			entry: `{"id":"ordering","name":"Ordering","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":[],"site_page":false}`,
			want:  "an id is an area and a topic in snake case",
		},
		{
			name:  "a topic id with a space in it",
			file:  topicsFile,
			entry: `{"id":"logic.simple ordering","name":"Ordering","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":[],"site_page":false}`,
			want:  "an id is an area and a topic in snake case",
		},
		{
			name:  "the same topic twice",
			file:  topicsFile,
			entry: `{"id":"logic.ordering","name":"Ordering","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":[],"site_page":false}`,
			want:  `topic "logic.ordering": the id is used twice`,
		},
		{
			name:  "a topic with no name",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":[],"site_page":false}`,
			want:  "the name is empty",
		},
		{
			name:  "a topic with no description",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","grade_levels":["1-2"],"slug":"seating","builds_on":[],"site_page":false}`,
			want:  "the description is empty",
		},
		{
			name:  "a topic at no level at all",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":[],"slug":"seating","builds_on":[],"site_page":false}`,
			want:  "no grade levels",
		},
		{
			name:  "a topic at a level that does not exist",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["7-8"],"slug":"seating","builds_on":[],"site_page":false}`,
			want:  `grade level "7-8" is not one of`,
		},
		{
			name:  "a topic at the same level twice",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2","1-2"],"slug":"seating","builds_on":[],"site_page":false}`,
			want:  `grade level "1-2" is listed twice`,
		},
		{
			name:  "a topic with a field the format does not have",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":[],"site_page":false,"note":"free text"}`,
			want:  `unknown field "note"`,
		},
		{
			name:  "a topic whose page has no address",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"","builds_on":[],"site_page":false}`,
			want:  `topic "logic.seating": a slug is lowercase words and numbers joined by dashes`,
		},
		{
			name:  "a slug that could lead an address elsewhere",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"../Seating","builds_on":[],"site_page":false}`,
			want:  `topic "logic.seating": a slug is lowercase words and numbers joined by dashes`,
		},
		{
			name:  "two topics whose pages would share an address",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"ordering","builds_on":[],"site_page":false}`,
			want:  `topic "logic.ordering": the slug "ordering" is already "logic.seating"'s`,
		},
		{
			name:  "a topic that does not say what it builds on",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","site_page":false}`,
			want:  `topic "logic.seating": builds_on is missing`,
		},
		{
			name:  "a topic that builds on itself",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":["logic.seating"],"site_page":false}`,
			want:  `topic "logic.seating": a topic cannot build on itself`,
		},
		{
			name:  "a topic that names one base twice",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":["logic.ordering","logic.ordering"],"site_page":false}`,
			want:  `topic "logic.seating": it builds on "logic.ordering" twice`,
		},
		{
			name:  "a topic that names its bases out of the catalog's order",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":["combinatorics.enumeration","logic.ordering"],"site_page":false}`,
			want:  `topic "logic.seating": it names "logic.ordering" after "combinatorics.enumeration", and a topic's bases follow the catalog's order`,
		},
		{
			name:  "a topic that builds on one the catalog does not have",
			file:  topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":["logic.tables"],"site_page":false}`,
			want:  `topic "logic.seating": it builds on "logic.tables", which is not in the catalog`,
		},
		{
			name: "topics that build on one another in a circle",
			file: topicsFile,
			entry: `{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":["logic.queues"],"site_page":false},` +
				`{"id":"logic.queues","name":"Queues","description":"Who comes first.","grade_levels":["1-2"],"slug":"queues","builds_on":["logic.ordering","logic.seating"],"site_page":false}`,
			want: "topics build on one another in a circle: logic.seating → logic.queues → logic.seating",
		},
		{
			name:  "a trap id that is not one snake case name",
			file:  trapsFile,
			entry: `{"id":"off by one","description":"Counted the gaps wrongly.","advice":"Draw the gaps."}`,
			want:  "an id is one name in snake case",
		},
		{
			name:  "the same trap twice",
			file:  trapsFile,
			entry: `{"id":"off_by_one","description":"Counted the gaps wrongly.","advice":"Draw the gaps."}`,
			want:  `trap "off_by_one": the id is used twice`,
		},
		{
			name:  "a trap with no description",
			file:  trapsFile,
			entry: `{"id":"forgot_a_case","advice":"List the cases."}`,
			want:  `trap "forgot_a_case": the description is empty`,
		},
		{
			name:  "a trap with no advice",
			file:  trapsFile,
			entry: `{"id":"forgot_a_case","description":"Missed one."}`,
			want:  `trap "forgot_a_case": the advice is empty`,
		},
		{
			name:  "a trap with a field the format does not have",
			file:  trapsFile,
			entry: `{"id":"forgot_a_case","description":"Missed one.","text":"free text"}`,
			want:  `unknown field "text"`,
		},
		{
			name:  "a skill id that is not one snake case name",
			file:  skillsFile,
			entry: `{"id":"simple fractions","description":"Halves and quarters."}`,
			want:  "an id is one name in snake case",
		},
		{
			name:  "the same skill twice",
			file:  skillsFile,
			entry: `{"id":"addition","description":"Adding whole numbers."}`,
			want:  `skill "addition": the id is used twice`,
		},
		{
			name:  "a skill with no description",
			file:  skillsFile,
			entry: `{"id":"long_division"}`,
			want:  `skill "long_division": the description is empty`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			withEntry(t, src, tc.file, tc.entry)
			wantProblem(t, src, tc.want)
		})
	}
}

func TestEmptyCatalogStopsTheService(t *testing.T) {
	t.Parallel()

	for _, file := range []string{topicsFile, trapsFile, skillsFile} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			src[file] = &fstest.MapFile{Data: []byte("[]")}
			wantProblem(t, src, "the catalog is empty")
		})
	}
}

func TestCatalogThatIsNotAListStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[topicsFile] = &fstest.MapFile{Data: []byte(`{"id":"logic.ordering"}`)}
	wantProblem(t, src, "content: parse "+topicsFile)
}

func TestTopicKnowsWhichLevelsItIsOfferedAt(t *testing.T) {
	t.Parallel()

	topic := Topic{ID: "logic.ordering", GradeLevels: []rating.GradeLevel{rating.Grades12, rating.Grades34}}
	for _, tc := range []struct {
		level rating.GradeLevel
		want  bool
	}{
		{rating.Grades12, true},
		{rating.Grades34, true},
		{rating.Grades56, false},
		{"", false},
	} {
		if got := topic.HasLevel(tc.level); got != tc.want {
			t.Errorf("HasLevel(%q) = %v, want %v", tc.level, got, tc.want)
		}
	}
}

// A digit belongs in an id the same way it does in a trap or a skill id, and a
// topic named after one — shapes on a 2d grid, say — must not be turned away by
// a rule nobody meant to write.
func TestATopicIDWithADigitIsAccepted(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withEntry(t, src, topicsFile,
		`{"id":"geometry.shapes_2d","name":"Shapes on a grid","description":"Figures made of cells.","grade_levels":["5-6"],"slug":"shapes-2d","builds_on":[],"site_page":false}`)
	if _, err := load(src); err != nil {
		t.Fatalf("load: %v", err)
	}
}

// The links are read once every topic is, so a topic may build on one the
// catalog lists after it: the order of the catalog is the rule's, not the map's.
func TestATopicMayBuildOnOneListedAfterIt(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withEntry(t, src, topicsFile,
		`{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":["logic.ordering"],"site_page":false}`)
	c, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	topic, _ := c.Topic("logic.seating")
	if got, want := topic.BuildsOn, []string{"logic.ordering"}; !slices.Equal(got, want) {
		t.Errorf("BuildsOn = %v, want %v", got, want)
	}
}

func TestACircleIsNamedOnceThoughABaseIsNamedTwice(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withEntry(t, src, topicsFile,
		`{"id":"logic.seating","name":"Seating","description":"Who stands where.","grade_levels":["1-2"],"slug":"seating","builds_on":["logic.queues"],"site_page":false},`+
			`{"id":"logic.queues","name":"Queues","description":"Who comes first.","grade_levels":["1-2"],"slug":"queues","builds_on":["logic.seating","logic.seating"],"site_page":false}`)
	_, err := load(src)
	if err == nil {
		t.Fatal("got no error, want the circle and the base named twice")
	}
	const circle = "topics build on one another in a circle: logic.seating → logic.queues → logic.seating"
	if got := strings.Count(err.Error(), circle); got != 1 {
		t.Errorf("the circle is named %d times in %q, want once", got, err)
	}
}
