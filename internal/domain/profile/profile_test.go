package profile_test

import (
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

func newStudent() profile.Student {
	return profile.Student{
		ExcludedSkills: []string{"fractions"},
		Grade:          2,
		Interests:      []string{"space", "cats"},
		Notes:          "Likes counting out loud.",
		Pseudonym:      "Otter",
	}
}

// A profile that has just been made is one the service can already work with:
// a child who has answered nothing has a level of zero, which is exactly what
// the ratings say about them.
func TestANewProfileIsUsableAtOnce(t *testing.T) {
	t.Parallel()

	made := time.Date(2026, 9, 23, 7, 30, 15, 500, time.UTC)
	p := profile.New(newStudent(), "1.2.3", made)

	if err := p.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want a new profile to be usable", err)
	}
	if p.SchemaVersion != profile.Version {
		t.Errorf("schema_version = %d, want %d", p.SchemaVersion, profile.Version)
	}
	if p.Revision != 1 {
		t.Errorf("revision = %d, want the first", p.Revision)
	}
	if p.StudentID == "" {
		t.Error("a new profile has no identifier")
	}
	if p.Ratings.Theta != 0 || p.Ratings.Answers != 0 {
		t.Errorf("a new profile starts at %v over %d answers, want zero over none",
			p.Ratings.Theta, p.Ratings.Answers)
	}
	if p.CurrentTask != nil || p.OpenRequest != nil {
		t.Error("a new profile already has a task or a request")
	}

	// The moment is kept as the file keeps it: UTC, whole seconds.
	if want := made.Truncate(time.Second); !p.CreatedAt.Equal(want) || !p.UpdatedAt.Equal(want) {
		t.Errorf("created %v and updated %v, want both at %v", p.CreatedAt, p.UpdatedAt, want)
	}
	if p.Daily.Date.Format("2006-01-02") != "2026-09-23" {
		t.Errorf("the counters belong to %v, want the day the profile was made", p.Daily.Date)
	}

	// Empty rather than missing: a file that says "no topics yet" reads
	// better than one that says nothing at all.
	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	for _, want := range []string{`"recent": []`, `"task_fingerprints": []`, `"topics": {}`} {
		if !strings.Contains(string(written), want) {
			t.Errorf("a new profile does not write %s:\n%s", want, written)
		}
	}
}

// Two children are two children, whatever they are called.
func TestEveryProfileGetsItsOwnIdentifier(t *testing.T) {
	t.Parallel()

	now := time.Now()
	first := profile.New(newStudent(), "1.2.3", now)
	second := profile.New(newStudent(), "1.2.3", now)

	if first.StudentID == second.StudentID {
		t.Errorf("two profiles share the identifier %q", first.StudentID)
	}
}

// Every write moves the revision and the moment, and stamps the build that
// did it. They go together because a write that moved only one of them is a
// write nobody can reconstruct afterwards.
func TestAWriteLeavesItsMark(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	before := p.Revision
	at := time.Date(2026, 9, 23, 8, 0, 0, 0, time.UTC)

	p.Touch("9.9.9", at)

	if p.Revision != before+1 {
		t.Errorf("revision = %d, want %d", p.Revision, before+1)
	}
	if !p.UpdatedAt.Equal(at) {
		t.Errorf("updated_at = %v, want %v", p.UpdatedAt, at)
	}
	if p.AppVersion != "9.9.9" {
		t.Errorf("app_version = %q, want the build that wrote it", p.AppVersion)
	}
	if p.SchemaVersion != profile.Version {
		t.Errorf("schema_version = %d, want a write to bring it up to %d", p.SchemaVersion, profile.Version)
	}
}
