package content

import (
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

	file := ideaFile{}
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
	src, file := contentCopy(t), ideaFile{}
	for level, length := range lengths {
		file[level] = ideasOfLength(level, length)
	}
	src[ideaPath("counting.gaps")] = asFile(t, file)
	loaded, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for level, length := range lengths {
		if got, want := loaded.ideas["counting.gaps"][level], ideasOfLength(level, length); !slices.Equal(got, want) {
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
		spoil func(ideaFile)
		want  string
	}{
		{"nine ideas", func(f ideaFile) { f[rating.Grades12] = f[rating.Grades12][:9] },
			`level "1-2": 9 ideas, and a list holds from 10 to 40`},
		{"forty-one ideas", func(f ideaFile) { f[rating.Grades12] = ideasOfLength(rating.Grades12, 41) },
			`level "1-2": 41 ideas, and a list holds from 10 to 40`},
		{"an idea of spaces", func(f ideaFile) { f[rating.Grades12][3] = " \t " },
			`level "1-2": idea 4 is empty`},
		{"an idea over two lines", func(f ideaFile) { f[rating.Grades12][3] = "One line.\nAnother." },
			`level "1-2": idea 4 runs over more than a line`},
		{"an idea longer than a line", func(f ideaFile) { f[rating.Grades12][3] = strings.Repeat("ä", longestIdea+1) },
			`level "1-2": idea 4 is 201 characters long, and an idea is at most 200`},
		{"an idea said twice", func(f ideaFile) {
			f[rating.Grades12][5] = "  " + strings.ToUpper(strings.ReplaceAll(f[rating.Grades12][2], " ", "  "))
		}, `level "1-2": idea 6 says what idea 3 says`},
		{"a level left out", func(f ideaFile) { delete(f, rating.Grades34) },
			`level "3-4": counting.gaps is taught at it, and it has no ideas`},
		{"a level no topic is taught at", func(f ideaFile) { f["7-8"] = sampleIdeas("7-8") },
			`level "7-8": counting.gaps is not taught at it`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			file := ideaFile{}
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
			file := ideaFile{}
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
