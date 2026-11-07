package content

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// ideaPath is where the ideas of this topic live.
func ideaPath(topic string) string { return ideasDir + "/" + topic + ideaSuffix }

// wordIdeas is a topic's file of ideas as a test writes it: every idea in
// words, a text of its own.
type wordIdeas map[rating.GradeLevel][]string

// sampleIdeas is a list with nothing wrong with it: as few ideas as a list may
// hold, each of its own.
func sampleIdeas(level rating.GradeLevel) []string { return ideasOfLength(level, fewestIdeas) }

// ideasOfLength is a list of that many ideas, each of its own.
func ideasOfLength(level rating.GradeLevel, length int) []string {
	list := make([]string, length)
	for i := range list {
		list[i] = fmt.Sprintf("Idea %d of the topic at %s.", i+1, level)
	}
	return list
}

// withIdeas gives a topic a list of ideas at each of the levels named, so that
// a test that adds a topic to the catalog adds one the content would take.
func withIdeas(t *testing.T, src fstest.MapFS, topic string, levels ...rating.GradeLevel) {
	t.Helper()

	file := wordIdeas{}
	for _, level := range levels {
		file[level] = sampleIdeas(level)
	}
	src[ideaPath(topic)] = asFile(t, file)
}

var gapsLevels = []rating.GradeLevel{rating.Grades12, rating.Grades34, rating.Grades56}

// A list is read as its file holds it, whatever its length from the fewest
// ideas a list may hold to the most: ten at 1-2, twenty at 3-4 and forty at
// 5-6. The catalog says how many ideas each list holds, and none for a topic
// it does not have or a level the topic is not taught at.
func TestIdeasAreReadAsTheirFileHoldsThem(t *testing.T) {
	t.Parallel()

	lengths := map[rating.GradeLevel]int{
		rating.Grades12: fewestIdeas, rating.Grades34: 20, rating.Grades56: profile.MaxIdeasPerList,
	}
	src, file := contentCopy(t), wordIdeas{}
	for level, length := range lengths {
		file[level] = ideasOfLength(level, length)
	}
	src[ideaPath("counting.gaps")] = asFile(t, file)
	loaded, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for level, length := range lengths {
		if got, want := textsOf(loaded.ideas["counting.gaps"][level]), ideasOfLength(level, length); !slices.Equal(got, want) {
			t.Errorf("ideas at %s = %q, want %q", level, got, want)
		}
		if got := loaded.IdeaCount("counting.gaps", level); got != length {
			t.Errorf("IdeaCount(counting.gaps, %s) = %d, want %d", level, got, length)
		}
	}
	for _, none := range []struct {
		topic string
		level rating.GradeLevel
	}{{"percent.basic", rating.Grades12}, {"no.such.topic", rating.Grades12}} {
		if got := loaded.IdeaCount(none.topic, none.level); got != 0 {
			t.Errorf("IdeaCount(%s, %s) = %d, want 0 for a list the catalog does not have", none.topic, none.level, got)
		}
	}
}

// Every topic of the catalog has a list at every level it is taught at, of a
// length a list may have: a package of any topic at any of its levels names an
// idea, and every place a family's file may name is on its list.
func TestEveryTopicHasAListOfIdeasAtEachOfItsLevels(t *testing.T) {
	t.Parallel()

	shipped, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, topic := range shipped.topics {
		for _, level := range topic.GradeLevels {
			if got := len(shipped.ideas[topic.ID][level]); got < fewestIdeas || got > profile.MaxIdeasPerList {
				t.Errorf("%s at %s has %d ideas, want from %d to %d",
					topic.ID, level, got, fewestIdeas, profile.MaxIdeasPerList)
			}
		}
	}
}

// A list a package could not name an idea from stops the service: a list
// shorter or longer than a list may be, an idea that says nothing, runs over a
// line or past the length of one, an idea said twice however it is spelt, and
// a level the topic is not taught at or one it is taught at left out.
func TestABrokenListOfIdeasStopsTheService(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		spoil func(wordIdeas)
		want  string
	}{
		{"nine ideas", func(f wordIdeas) { f[rating.Grades12] = f[rating.Grades12][:9] },
			`level "1-2": 9 ideas, and a list holds from 10 to 40`},
		{"forty-one ideas", func(f wordIdeas) { f[rating.Grades12] = ideasOfLength(rating.Grades12, 41) },
			`level "1-2": 41 ideas, and a list holds from 10 to 40`},
		{"an idea of spaces", func(f wordIdeas) { f[rating.Grades12][3] = " \t " },
			`level "1-2": idea 4 is empty`},
		{"an idea over two lines", func(f wordIdeas) { f[rating.Grades12][3] = "One line.\nAnother." },
			`level "1-2": idea 4 runs over more than a line`},
		{"an idea longer than a line", func(f wordIdeas) { f[rating.Grades12][3] = strings.Repeat("ä", longestIdea+1) },
			`level "1-2": idea 4 is 201 characters long, and an idea is at most 200`},
		{"an idea said twice", func(f wordIdeas) {
			f[rating.Grades12][5] = "  " + strings.ToUpper(strings.ReplaceAll(f[rating.Grades12][2], " ", "  "))
		}, `level "1-2": idea 6 says what idea 3 says`},
		{"a level left out", func(f wordIdeas) { delete(f, rating.Grades34) },
			`level "3-4": counting.gaps is taught at it, and it has no ideas`},
		{"a level no topic is taught at", func(f wordIdeas) { f["7-8"] = sampleIdeas("7-8") },
			`level "7-8": counting.gaps is not taught at it`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			file := wordIdeas{}
			for _, level := range gapsLevels {
				file[level] = sampleIdeas(level)
			}
			test.spoil(file)
			src[ideaPath("counting.gaps")] = asFile(t, file)
			wantProblem(t, src, ideaPath("counting.gaps")+":\n  - "+test.want)
		})
	}
}

// Something among the ideas that is not a topic's list stops the service, and
// so does a topic of the catalog with no list at all: its package would name
// no idea.
func TestSomethingThatIsNotAListOfIdeasStopsTheService(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, file, data, want string
	}{
		{"a list of no topic", ideaPath("counting.everything"), "{}",
			"counting.everything.json: an idea list is a file named by a topic of the catalog"},
		{"a directory named like a list", ideasDir + "/old.json/counting.gaps.json", "{}",
			"old.json: an idea list is a file named by a topic of the catalog"},
		{"a file of another kind", ideasDir + "/notes.txt", "a note",
			"notes.txt: an idea list is a file named by a topic of the catalog"},
		{"a level's ideas that are not a list", ideaPath("counting.gaps"), `{"1-2": "One idea."}`,
			"content: parse " + ideaPath("counting.gaps")},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			src[test.file] = &fstest.MapFile{Data: []byte(test.data)}
			wantProblem(t, src, test.want)
		})
	}
	t.Run("a topic with no list", func(t *testing.T) {
		t.Parallel()

		src := contentCopy(t)
		delete(src, ideaPath("counting.gaps"))
		wantProblem(t, src, `topic "counting.gaps" has no file of ideas`)
	})
}

// A list or the lists' directory that is there and cannot be read stops the
// service, and so does content without the directory at all.
func TestIdeasThatCannotBeReadStopTheService(t *testing.T) {
	t.Parallel()

	for _, name := range []string{ideaPath("counting.gaps"), ideasDir} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wantProblem(t, unreadable{MapFS: contentCopy(t), name: name}, "content: read "+name)
		})
	}
	t.Run("no directory", func(t *testing.T) {
		t.Parallel()

		src := contentCopy(t)
		for name := range src {
			if strings.HasPrefix(name, ideasDir+"/") {
				delete(src, name)
			}
		}
		wantProblem(t, src, "content: read "+ideasDir)
	})
}

// An idea is part of what the model is told, so the version every log line
// carries follows it: reworded, or moved to another place on its list, it is
// a different version.
func TestAChangedIdeaChangesTheInstructionsVersion(t *testing.T) {
	t.Parallel()

	base := contentCopy(t)
	withIdeas(t, base, "counting.gaps", gapsLevels...)
	before, err := load(base)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	for _, test := range []struct {
		name  string
		spoil func(list []string)
	}{
		{"it is reworded", func(list []string) { list[0] = "Another idea altogether." }},
		{"two ideas change places", func(list []string) { list[0], list[1] = list[1], list[0] }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := maps.Clone(base)
			file := wordIdeas{}
			for _, level := range gapsLevels {
				file[level] = sampleIdeas(level)
			}
			test.spoil(file[rating.Grades34])
			src[ideaPath("counting.gaps")] = asFile(t, file)
			after, err := load(src)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if after.InstructionsVersion() == before.InstructionsVersion() {
				t.Errorf("version = %q both times, want a different one once an idea changed",
					after.InstructionsVersion())
			}
		})
	}
}

// markedIdea is an idea solved from the picture as its file writes it.
func markedIdea(text string) map[string]any {
	return map[string]any{"idea": text, "solved_from_picture": true}
}

// gapsIdeas is the file of counting.gaps with the ideas listed at 1-2 after
// the sample list of every level.
func gapsIdeas(more ...any) map[rating.GradeLevel][]any {
	file := map[rating.GradeLevel][]any{}
	for _, level := range gapsLevels {
		for _, idea := range sampleIdeas(level) {
			file[level] = append(file[level], idea)
		}
	}
	file[rating.Grades12] = append(file[rating.Grades12], more...)
	return file
}

// solvedFromPicture marks a topic of the catalog as one all of whose tasks are
// solved from the picture.
func solvedFromPicture(t *testing.T, src fstest.MapFS, topic string) {
	t.Helper()

	var topics []map[string]any
	if err := json.Unmarshal(src[topicsFile].Data, &topics); err != nil {
		t.Fatalf("read %s: %v", topicsFile, err)
	}
	for _, entry := range topics {
		if entry["id"] == topic {
			entry["solved_from_picture"] = true
		}
	}
	src[topicsFile] = asFile(t, topics)
}

// A list marks an idea solved from the picture on the idea's own line, and the
// content says so of that place alone: not of the ideas in words around it, of
// a place the list does not have, or of the topic, whose other tasks are told
// in words. The package names the idea by its words, as any.
func TestAnIdeaMarkedSolvedFromThePictureIsReadSo(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[ideaPath("counting.gaps")] = asFile(t, gapsIdeas(markedIdea("Posts in a row drawn on the card.")))
	loaded, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, test := range []struct {
		level  rating.GradeLevel
		number int
		want   bool
	}{
		{rating.Grades12, fewestIdeas + 1, true},
		{rating.Grades12, 1, false},
		{rating.Grades12, fewestIdeas, false},
		{rating.Grades12, fewestIdeas + 2, false},
		{rating.Grades12, 0, false},
		{rating.Grades34, fewestIdeas + 1, false},
	} {
		if got := loaded.IdeaFromPicture("counting.gaps", test.level, test.number); got != test.want {
			t.Errorf("IdeaFromPicture(counting.gaps, %s, %d) = %v, want %v", test.level, test.number, got, test.want)
		}
	}
	if loaded.SolvedFromPicture("counting.gaps") {
		t.Error("SolvedFromPicture(counting.gaps) = true for a topic with ideas in words, want false")
	}
	if got := textsOf(loaded.ideas["counting.gaps"][rating.Grades12])[fewestIdeas]; got != "Posts in a row drawn on the card." {
		t.Errorf("the idea marked reads %q, want its words", got)
	}
}

// A topic the catalog marks as one all of whose tasks are solved from the
// picture makes every idea of its lists one, and needs no idea in words: none
// of its tasks goes to a chat that is given those alone.
func TestEveryIdeaOfATopicSolvedFromThePictureIsOne(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	solvedFromPicture(t, src, "counting.gaps")
	marked := make([]any, 0, fewestIdeas)
	for _, idea := range sampleIdeas(rating.Grades12) {
		marked = append(marked, markedIdea(idea))
	}
	file := gapsIdeas()
	file[rating.Grades12] = marked
	src[ideaPath("counting.gaps")] = asFile(t, file)
	loaded, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !loaded.SolvedFromPicture("counting.gaps") {
		t.Error("SolvedFromPicture(counting.gaps) = false, want the catalog's mark")
	}
	for _, level := range gapsLevels {
		for number := 1; number <= fewestIdeas; number++ {
			if !loaded.IdeaFromPicture("counting.gaps", level, number) {
				t.Errorf("IdeaFromPicture(counting.gaps, %s, %d) = false, want every idea of the topic", level, number)
			}
		}
	}
	if loaded.SolvedFromPicture("time.clocks") || loaded.IdeaFromPicture("time.clocks", rating.Grades34, 1) {
		t.Error("another topic reads as solved from the picture, want only the one the catalog marks")
	}
}

// A mark is written one way, and a list of a topic told in words keeps a full
// list of ideas in words, which is all a chat not known to draw the card and
// the trial series are given: an object that marks nothing, one that holds
// anything else, and too few ideas in words stop the service.
func TestABrokenMarkOfAnIdeaStopsTheService(t *testing.T) {
	t.Parallel()

	nineWords := gapsIdeas(markedIdea("Posts in a row drawn on the card."))
	nineWords[rating.Grades12] = nineWords[rating.Grades12][1:]
	for _, test := range []struct {
		name string
		file map[rating.GradeLevel][]any
		want string
	}{
		{"a mark of nothing", gapsIdeas(map[string]any{"idea": "Posts.", "solved_from_picture": false}),
			"an idea in words is a text"},
		{"a member of no idea", gapsIdeas(map[string]any{"idea": "Posts.", "solved_from_picture": true, "kind": "row"}),
			"an idea is a text, or an object of idea and solved_from_picture"},
		{"too few ideas in words", nineWords,
			`level "1-2": 9 ideas in words, and a list of a topic not all of whose tasks are solved from the picture ` +
				"holds at least 10"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			src[ideaPath("counting.gaps")] = asFile(t, test.file)
			wantProblem(t, src, test.want)
		})
	}
}
