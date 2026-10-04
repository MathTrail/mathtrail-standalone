package store_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// students are the seed profiles, each a file exactly as a store would keep it.
var students = []string{"dima", "masha", "olya", "petya", "sasha"}

func fixture(t testing.TB, student string) []byte {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "profiles", student+".json"))
	if err != nil {
		t.Fatalf("read the fixture: %v", err)
	}
	return raw
}

// A store reads the file as the file is: what it hands back is written out
// again byte for byte.
func TestParseReadsTheFileAsItIs(t *testing.T) {
	t.Parallel()

	for _, student := range students {
		t.Run(student, func(t *testing.T) {
			t.Parallel()

			raw := fixture(t, student)
			p, err := store.Parse(raw)
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}
			written, err := profile.Marshal(p)
			if err != nil {
				t.Fatalf("Marshal() error = %v, want nil", err)
			}
			if !bytes.Equal(written, raw) {
				t.Errorf("Parse() then Marshal() =\n%s\nwant the file as it was:\n%s", written, raw)
			}
		})
	}
}

// A file nothing can read is damage, and why it could not be read stays
// visible behind it. A file from a newer build is not damage: it is readable,
// only not by this build.
func TestParseTellsDamageFromANewerBuild(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		raw   string
		want  error
		cause error
	}{
		{"an empty file", "", store.ErrCorrupted, profile.ErrMalformed},
		{"a write cut short", `{"schema_version": 1, "student": {"pseud`, store.ErrCorrupted, profile.ErrMalformed},
		{"a file that is not a profile", fmt.Sprintf(`{"schema_version": %d}`, profile.Version),
			store.ErrCorrupted, profile.ErrInvalid},
		{"a shape older than anything this build reads", `{"schema_version": 0}`, store.ErrCorrupted, profile.ErrOlder},
		{"a file from a newer build", fmt.Sprintf(`{"schema_version": %d}`, profile.Version+1),
			profile.ErrNewer, profile.ErrNewer},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			p, err := store.Parse([]byte(test.raw))
			if !errors.Is(err, test.want) || !errors.Is(err, test.cause) {
				t.Errorf("Parse() error = %v, want %v, because of %v", err, test.want, test.cause)
			}
			if errors.Is(err, store.ErrCorrupted) && errors.Is(err, profile.ErrNewer) {
				t.Errorf("Parse() error = %v, want damage or newer, not both", err)
			}
			if p != nil {
				t.Error("Parse() returned a profile alongside the refusal")
			}
		})
	}
}

// What a store holds may have been edited by hand or cut short on its way, so
// whatever the bytes, reading them ends in a profile the service can work with
// or in exactly one of the two refusals — never in a crash, and never in a
// refusal nobody could branch on.
func FuzzParse(f *testing.F) {
	for _, student := range students {
		f.Add(fixture(f, student))
	}
	for _, seed := range []string{
		"", "{", "null", "[]", `{"schema_version": 0}`, `{"schema_version": 1}`, `{"schema_version": 2}`,
		`{"schema_version": 1e400}`, `{"schema_version": -1}`,
		`{"schema_version": 1, "rating_days": [{"date": "2026-09-20", "deltas": {"counting.gaps": 0.1}, "theta": 2.5}]}`,
		`{"schema_version": 1, "recent": [{"before": {"delta": 0.1, "theta": 2.5}}]}`,
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		p, err := store.Parse(raw)
		switch {
		case err == nil && p == nil:
			t.Fatal("Parse() = nil, nil, want a profile or a refusal")
		case err == nil:
			if invalid := p.Validate(); invalid != nil {
				t.Errorf("Parse() handed back a profile Validate() refuses: %v", invalid)
			}
		case p != nil:
			t.Error("Parse() returned a profile alongside the refusal")
		case errors.Is(err, store.ErrCorrupted) == errors.Is(err, profile.ErrNewer):
			t.Errorf("Parse() error = %v, want exactly one of %v and %v", err, store.ErrCorrupted, profile.ErrNewer)
		}
	})
}

// At a moment, a file of a newer build is newer while a rollout could explain
// it — written within the window of that moment, on either side — and
// unsupported beside it otherwise, or when it gives no moment this build
// reads. A profile this build reads, and damage, are what Parse makes them.
func TestParseAtTellsARolloutFromAFileNoRolloutExplains(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 29, 12, 0, 0, 0, time.UTC)
	newerAt := func(updatedAt string) []byte {
		return fmt.Appendf(nil, `{"schema_version": %d, "updated_at": %s}`, profile.Version+1, updatedAt)
	}
	moment := func(at time.Time) string { return strconv.Quote(at.Format(time.RFC3339)) }
	for _, test := range []struct {
		name                          string
		raw                           []byte
		newer, unsupported, corrupted bool
	}{
		{"this build's profile", fixture(t, "dima"), false, false, false},
		{"damage", []byte("{"), false, false, true},
		{"a newer file written now", newerAt(moment(now)), true, false, false},
		{"one written as long ago as a rollout takes", newerAt(moment(now.Add(-store.RolloutWindow))), true, false, false},
		{"one dated as far ahead as a clock could run", newerAt(moment(now.Add(store.RolloutWindow))), true, false, false},
		{"one written a second longer ago", newerAt(moment(now.Add(-store.RolloutWindow - time.Second))), true, true, false},
		{"one dated a second further ahead", newerAt(moment(now.Add(store.RolloutWindow + time.Second))), true, true, false},
		{"one that gives no moment", fmt.Appendf(nil, `{"schema_version": %d}`, profile.Version+1), true, true, false},
		{"one that gives a moment in no form this build reads", newerAt(`"yesterday"`), true, true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			p, err := store.ParseAt(test.raw, now)
			if errors.Is(err, profile.ErrNewer) != test.newer || errors.Is(err, store.ErrUnsupported) != test.unsupported ||
				errors.Is(err, store.ErrCorrupted) != test.corrupted {
				t.Errorf("ParseAt() error = %v, want newer: %v, unsupported: %v, corrupted: %v",
					err, test.newer, test.unsupported, test.corrupted)
			}
			if (p == nil) == (err == nil) {
				t.Errorf("ParseAt() = %v, %v; want a profile or a refusal, and never both", p, err)
			}
		})
	}
}
