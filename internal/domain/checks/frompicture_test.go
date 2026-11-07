package checks_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// pictureTask is a task solved from the picture: a question that gives no
// facts, a picture that holds them, and five options of which A is right.
func pictureTask(picture string, options ...string) *checks.Task {
	task := &checks.Task{
		Question: "Which number goes where the question mark is?", Picture: json.RawMessage(picture),
		CorrectAnswer: "A", Options: map[string]string{},
	}
	for at, letter := range []string{"A", "B", "C", "D", "E"} {
		task.Options[letter] = options[at]
	}
	return task
}

// refusesFor says whether holding a picture to its wording, as a task solved
// from the picture is held, found a problem whose message says so.
func refusesFor(t *testing.T, task *checks.Task, says string) bool {
	t.Helper()
	if broken := checks.PictureFormat(task, "en"); len(broken) != 0 {
		t.Fatalf("PictureFormat() = %v, want a picture in the format", broken)
	}
	return slices.ContainsFunc(checks.PictureMatchFromPicture(task, "en"), func(problem checks.Problem) bool {
		return strings.Contains(problem.Message, says)
	})
}

// A crown of tiles draws the numbers the child chooses among: drawn beside
// three or four wrong options, the right one is a candidate, not the answer
// shown; beside fewer, a child could read it off the picture.
func TestAPictureThatDrawsTheAnswerAmongThreeWrongOnesShowsNone(t *testing.T) {
	t.Parallel()
	const crown = `{"kind":"tiles","sides":4,"layout":"row","tiles":[["1","2","3","4"],["5","","","?"]]}`

	for _, test := range []struct {
		name    string
		options []string
		shows   bool
	}{
		{"the right value and all four wrong ones", []string{"1", "2", "3", "4", "5"}, false},
		{"the right value and three wrong ones", []string{"1", "2", "3", "4", "9"}, false},
		{"the right value and two wrong ones", []string{"1", "2", "3", "8", "9"}, true},
		{"the right value alone", []string{"1", "6", "7", "8", "9"}, true},
		{"no value an option holds", []string{"6", "7", "8", "9", "10"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := refusesFor(t, pictureTask(crown, test.options...), "shows the right answer"); got != test.shows {
				t.Errorf("shows the right answer = %v, want %v", got, test.shows)
			}
		})
	}
}

// The exception is a task solved from the picture's alone: a word task whose
// picture draws its answer among three wrong values still shows it.
func TestAWordTaskThatDrawsTheAnswerAmongWrongOnesShowsIt(t *testing.T) {
	t.Parallel()
	task := pictureTask(`{"kind":"table","rows":[["1","2","3","4"]]}`, "1", "2", "3", "4", "9")
	shown := slices.ContainsFunc(checks.PictureMatch(task, "en"), func(problem checks.Problem) bool {
		return strings.Contains(problem.Message, "shows the right answer")
	})
	if !shown {
		t.Error("PictureMatch() lets the answer drawn among three wrong values through, want it refused")
	}
}

// Tiles and rows of shapes mark the place they ask about with ?, and a picture
// that does not is refused in words that name no option; a lattice, a row and
// a staircase ask for counts and routes, and need none.
func TestAPictureOfTilesOrShapesMarksThePlaceItAsksAbout(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, picture string
		refused       bool
	}{
		{"tiles with ?", `{"kind":"tiles","sides":4,"layout":"row","tiles":[["1","","",""],["","","?",""]]}`, false},
		{"tiles with none", `{"kind":"tiles","sides":4,"layout":"row","tiles":[["1","","",""],["","","",""]]}`, true},
		{"shapes with ?", `{"kind":"shapes","rows":[{"shapes":["circle"],"total":"4"},{"shapes":["circle","square"],"total":"?"}]}`, false},
		{"shapes with none", `{"kind":"shapes","rows":[{"shapes":["circle"],"total":"4"}]}`, true},
		{"a lattice with none", `{"kind":"lattice","width":2,"height":2,"lines":[[0,0,2,0]]}`, false},
		{"a staircase with none", `{"kind":"stairs","steps":5,"marks":[{"step":0,"label":"A"}]}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			task := pictureTask(test.picture, "6", "7", "8", "9", "10")
			if got := refusesFor(t, task, "marks no place with ?"); got != test.refused {
				t.Errorf("refused for no ? = %v, want %v", got, test.refused)
			}
			if test.refused && len(checks.PictureMatch(task, "en")) != 0 {
				t.Errorf("PictureMatch() of a word task refuses %v, want the ? asked of a task solved from the picture alone",
					checks.PictureMatch(task, "en"))
			}
		})
	}
}

// A task asked for as one solved from the picture comes with a picture: one
// without is refused by its structure, and a word task without one is not.
func TestATaskSolvedFromThePictureHasAPicture(t *testing.T) {
	t.Parallel()
	fromPicture := &profile.Brief{Constraints: []string{profile.FactsOnThePicture}}
	words := &profile.Brief{}
	bare := &checks.Task{Question: "How many triangles are there?"}
	drawn := &checks.Task{Question: "How many triangles are there?", Picture: json.RawMessage(`{"kind":"lattice"}`)}

	if got := checks.PictureRequired(bare, fromPicture); len(got) != 1 || got[0].Code != checks.CodeBadStructure {
		t.Errorf("PictureRequired(no picture, from the picture) = %v, want one bad_structure", got)
	}
	if got := checks.PictureRequired(drawn, fromPicture); len(got) != 0 {
		t.Errorf("PictureRequired(a picture, from the picture) = %v, want none", got)
	}
	if got := checks.PictureRequired(bare, words); len(got) != 0 {
		t.Errorf("PictureRequired(no picture, a word task) = %v, want none", got)
	}
}

// The picture of a task solved from the picture repeats one the child was
// given when it is the same figure, whatever the order its lists are written
// in and whatever its question asks; another figure does not.
func TestAPictureThatRepeatsAPastOneIsRefused(t *testing.T) {
	t.Parallel()
	const posts = `{"kind":"row","items":[{},{},{},{},{},{},{}]}`
	const maze = `{"kind":"lattice","width":4,"height":4,"lines":[[1,0,1,3],[2,1,2,4],[3,0,3,3]],` +
		`"marks":[{"at":[0,0],"label":"S"},{"at":[3,3],"label":"E"}]}`
	past := []string{checks.PictureFingerprint(json.RawMessage(maze)), checks.PictureFingerprint(json.RawMessage(posts))}

	for _, test := range []struct {
		name, picture string
		repeats       bool
	}{
		{"the same figure", maze, true},
		{"the same figure, its lists in another order",
			`{"marks":[{"label":"E","at":[3,3]},{"at":[0,0],"label":"S"}],"kind":"lattice","height":4,"width":4,` +
				`"lines":[[3,0,3,3],[1,0,1,3],[2,1,2,4]]}`, true},
		{"a row of the same posts, more of them", `{"kind":"row","items":[{},{},{},{},{},{},{},{},{},{},{},{}]}`, false},
		{"another figure", `{"kind":"lattice","width":5,"height":3,"lines":[[1,1,4,1],[2,0,2,2]],` +
			`"marks":[{"at":[0,2],"label":"S"},{"at":[4,0],"label":"E"}]}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := checks.RepeatsPicture(json.RawMessage(test.picture), past)
			if repeats := len(got) == 1 && got[0].NearDuplicateOf == checks.EarlierPastPicture; repeats != test.repeats {
				t.Errorf("RepeatsPicture() = %v, want a repeat: %v", got, test.repeats)
			}
		})
	}
}

// A picture that cannot be read, and a sketch damaged in the file, repeat
// nothing: neither is a reason to refuse a task.
func TestNothingRepeatsWhatCannotBeRead(t *testing.T) {
	t.Parallel()
	if got := checks.PictureFingerprint(json.RawMessage(`[1,2]`)); got != "" {
		t.Errorf("PictureFingerprint(a list) = %q, want none", got)
	}
	if got := checks.RepeatsPicture(json.RawMessage(`{"kind":"stairs","steps":3}`), []string{"damaged"}); len(got) != 0 {
		t.Errorf("RepeatsPicture(against a damaged sketch) = %v, want none", got)
	}
}

// A task solved from the picture keeps its words and changes its picture: its
// question is held to no question the child was given, and its picture to the
// pictures of the child's past such tasks.
func TestATaskSolvedFromThePictureIsNewByItsPicture(t *testing.T) {
	t.Parallel()
	words := func(s *scenario) {
		s.fingerprints = []string{checks.Fingerprint(s.draft.Task.Question, "en")}
	}
	drawn := func(s *scenario) {
		s.pictures = []string{checks.PictureFingerprint(s.draft.Task.Picture)}
	}
	for _, test := range []struct {
		name        string
		fromPicture bool
		past        func(*scenario)
		repeats     []checks.Earlier
	}{
		{"its question asked before", true, words, nil},
		{"its picture drawn before", true, drawn, []checks.Earlier{checks.EarlierPastPicture}},
		{"a word task's question asked before", false, words, []checks.Earlier{checks.EarlierPast}},
		{"a word task's picture drawn before", false, drawn, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			s := accepted()
			solvedFromPicture(&s)
			s.fromPicture = test.fromPicture
			s.runner = &working{value: "six pairs", read: []string{"items"}}
			test.past(&s)
			outcome := s.review(t)
			if got := kindsOf(outcome.Problems); !slices.Equal(got, test.repeats) {
				t.Errorf("repeats = %v, want %v (problems %v)", got, test.repeats, outcome.Problems)
			}
		})
	}
}

// A self-check may say that a fact stands in the words, and saying so with a
// blocking issue refuses the task, under the issue's type.
func TestAFactInTheWordsIsAnIssueTheSelfCheckNames(t *testing.T) {
	t.Parallel()
	s := accepted()
	s.draft.SelfCheck.Issues = []checks.Issue{{Type: "facts_in_words", Severity: "blocking", Comment: "The 6 is in the question."}}
	outcome := s.review(t)
	if got := codesOf(&outcome); !slices.Equal(got, []checks.Code{checks.CodeSelfCheckBlocking}) {
		t.Fatalf("codes = %v, want only self_check_blocking", got)
	}
	if !strings.Contains(outcome.Problems[0].Message, "(facts_in_words)") {
		t.Errorf("the refusal %q does not name the issue's type", outcome.Problems[0].Message)
	}
}
