package profile_test

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// The five fixtures are the prototype's own seed profiles, carried into this
// shape. They are not invented: the ratings in them were replayed from the
// same answers the prototype replayed, so the numbers below are checked
// against the vectors it exported rather than against themselves.

// students is the fixture set, named as the prototype named them so that a
// file here and a row in the reference can be put side by side.
var students = []string{"dima", "masha", "olya", "petya", "sasha"}

func fixturePath(student string) string {
	return filepath.Join("..", "..", "..", "testdata", "profiles", student+".json")
}

func readFixture(t *testing.T, student string) []byte {
	t.Helper()

	raw, err := os.ReadFile(fixturePath(student))
	if err != nil {
		t.Fatalf("read the fixture: %v", err)
	}
	return raw
}

func parseFixture(t *testing.T, student string) *profile.Profile {
	t.Helper()

	p, err := profile.Parse(readFixture(t, student))
	if err != nil {
		t.Fatalf("Parse(%s) error = %v, want nil", student, err)
	}
	return p
}

// A file written by this package and read back by it is the same file. That
// is what lets a profile pass through a tool call without the parts nobody
// touched drifting.
func TestEveryFixtureSurvivesBeingReadAndWritten(t *testing.T) {
	t.Parallel()

	for _, student := range students {
		t.Run(student, func(t *testing.T) {
			t.Parallel()

			raw := readFixture(t, student)
			written, err := profile.Marshal(parseFixture(t, student))
			if err != nil {
				t.Fatalf("Marshal() error = %v, want nil", err)
			}
			if !bytes.Equal(written, raw) {
				t.Errorf("the file changed by being read and written\n--- on disk\n%s\n--- written\n%s", raw, written)
			}
		})
	}
}

// replayedRatings is one profile as the prototype replayed it: the numbers
// its own code arrived at from the same answers.
type replayedRatings struct {
	Student string  `json:"student"`
	Theta   float64 `json:"theta"`
	Answers int     `json:"answers"`
	Topics  map[string]struct {
		Delta   float64 `json:"delta"`
		Answers int     `json:"answers"`
	} `json:"topics"`
}

// The ratings in the fixtures against the ones the prototype exported. This is
// the check that the conversion did not quietly invent a number: the same
// answers, replayed by two implementations, have to end in the same place.
func TestTheFixturesCarryTheReferenceRatings(t *testing.T) {
	t.Parallel()

	var golden struct {
		SeedReplays []replayedRatings `json:"seed_replays"`
	}
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "golden", "ratings.json"))
	if err != nil {
		t.Fatalf("read the reference: %v", err)
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse the reference: %v", err)
	}
	if len(golden.SeedReplays) != len(students) {
		t.Fatalf("the reference replays %d profiles, the fixtures are %d", len(golden.SeedReplays), len(students))
	}

	for _, want := range golden.SeedReplays {
		t.Run(want.Student, func(t *testing.T) {
			t.Parallel()

			checkRatings(t, parseFixture(t, want.Student), want)
		})
	}
}

// checkRatings holds one fixture to one row of the reference.
func checkRatings(t *testing.T, p *profile.Profile, want replayedRatings) {
	t.Helper()

	if math.Abs(p.Ratings.Theta-want.Theta) > 1e-12 {
		t.Errorf("theta = %v, want %v", p.Ratings.Theta, want.Theta)
	}
	if p.Ratings.Answers != want.Answers {
		t.Errorf("answers = %d, want %d", p.Ratings.Answers, want.Answers)
	}
	if len(p.Topics) != len(want.Topics) {
		t.Fatalf("topics = %d, want %d", len(p.Topics), len(want.Topics))
	}

	for id, topic := range want.Topics {
		got, met := p.Topics[id]
		if !met {
			t.Errorf("the fixture never met %s", id)
			continue
		}
		if math.Abs(got.Delta-topic.Delta) > 1e-12 {
			t.Errorf("%s: delta = %v, want %v", id, got.Delta, topic.Delta)
		}
		if got.Answers != topic.Answers {
			t.Errorf("%s: answers = %d, want %d", id, got.Answers, topic.Answers)
		}
	}
}

// What the fixtures are for beyond the numbers: one of them carries a task in
// flight, so the shape of the sealed block is covered, and the summary of
// every topic adds up to the answers in the window.
func TestTheFixturesAreConsistentWithThemselves(t *testing.T) {
	t.Parallel()

	sealed := 0
	for _, student := range students {
		p := parseFixture(t, student)
		if p.CurrentTask != nil {
			sealed++
		}
		checkSummaryCovers(t, student, p)
	}
	if sealed == 0 {
		t.Error("no fixture carries a task in flight, so the sealed block is never exercised")
	}
}

// checkSummaryCovers holds one fixture to the promise the summary makes: it
// counts at least what the window still shows, because the window is what gets
// dropped and the summary is what never does.
func checkSummaryCovers(t *testing.T, student string, p *profile.Profile) {
	t.Helper()

	answered := map[string]int{}
	for i := range p.Recent {
		answered[p.Recent[i].Topic]++
	}
	for topic, times := range answered {
		if p.Topics[topic].Answers < times {
			t.Errorf("%s: %s counts %d answers and the window holds %d",
				student, topic, p.Topics[topic].Answers, times)
		}
	}
	if len(p.TaskFingerprints) != len(p.Recent) {
		t.Errorf("%s: %d fingerprints against %d answers",
			student, len(p.TaskFingerprints), len(p.Recent))
	}
}
