package learner_test

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/learner"
)

// child is the identifier a profile gives its child: a random UUID.
const child = "3b241101-e2bb-4255-8caf-4136c566a962"

// secretNamed is a key that differs from every other name's, in the encoding a
// deployment carries it in.
func secretNamed(name string) string {
	secret := make([]byte, learner.KeySize)
	copy(secret, name)
	return base64.StdEncoding.EncodeToString(secret)
}

func newKey(t *testing.T, encoded string) *learner.Key {
	t.Helper()

	key, err := learner.NewKey(encoded)
	if err != nil {
		t.Fatalf("NewKey() error = %v, want none", err)
	}
	return key
}

// A key is the standard base64 of the right number of bytes, with whatever
// whitespace a file leaves around it, and a refusal never repeats the value.
func TestAKeyIsReadFromItsEncoding(t *testing.T) {
	t.Parallel()

	short := base64.StdEncoding.EncodeToString(make([]byte, learner.KeySize-1))
	for _, tc := range []struct {
		name    string
		encoded string
		wantErr bool
	}{
		{"a key", secretNamed("one"), false},
		{"with a newline after it", secretNamed("one") + "\n", false},
		{"too short", short, true},
		{"not base64", "not-a-key-at-all!", true},
		{"base64url rather than base64", base64.URLEncoding.EncodeToString(make([]byte, learner.KeySize)) + "_", true},
		{"nothing", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := learner.NewKey(tc.encoded)
			if (err != nil) != tc.wantErr {
				t.Fatalf("NewKey() error = %v, want an error: %v", err, tc.wantErr)
			}
			if err == nil {
				return
			}
			if !errors.Is(err, learner.ErrKey) {
				t.Errorf("NewKey() error = %v, want it to wrap ErrKey", err)
			}
			if trimmed := strings.TrimSpace(tc.encoded); trimmed != "" && strings.Contains(err.Error(), trimmed) {
				t.Errorf("NewKey() error = %q, which repeats the value it refused", err)
			}
		})
	}
}

// A child is counted under one name all month, whoever signs in for it, and
// under another from the first second of the next month; another child, or
// another secret, is another name.
func TestANameLastsACalendarMonth(t *testing.T) {
	t.Parallel()

	key := newKey(t, secretNamed("one"))
	first := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	name := key.Of(child, first)
	tokyo := time.FixedZone("UTC+9", 9*60*60)

	for _, tc := range []struct {
		name  string
		got   string
		same  bool
		about string
	}{
		{"the last second of the month", key.Of(child, first.AddDate(0, 1, 0).Add(-time.Second)), true, "the same month"},
		{"a moment written in another zone", key.Of(child, first.Add(time.Hour).In(tokyo)), true, "the same month in UTC"},
		{"the next month", key.Of(child, first.AddDate(0, 1, 0)), false, "another month"},
		{"the month before", key.Of(child, first.Add(-time.Second)), false, "another month"},
		{"the same month a year on", key.Of(child, first.AddDate(1, 0, 0)), false, "another month"},
		{"another child", key.Of("another child", first), false, "another child"},
		{"another secret", newKey(t, secretNamed("two")).Of(child, first), false, "another secret"},
		{"a key of its own", learner.RandomKey().Of(child, first), false, "another secret"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if (tc.got == name) != tc.same {
				t.Errorf("Of() = %q beside %q for %s; want the same: %v", tc.got, name, tc.about, tc.same)
			}
		})
	}
}

// Midnight of the first of a month in UTC is still the last month a few hours
// to the west, and the name follows UTC rather than the zone a time was
// written in.
func TestANameFollowsTheMonthInUTC(t *testing.T) {
	t.Parallel()

	key := newKey(t, secretNamed("one"))
	newYork := time.FixedZone("UTC-5", -5*60*60)
	lastEvening := time.Date(2026, time.September, 30, 21, 0, 0, 0, newYork) // 1 October, 02:00 UTC

	if got, want := key.Of(child, lastEvening), key.Of(child, time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)); got != want {
		t.Errorf("Of() = %q at 21:00 on 30 September in UTC-5, want %q, the name of October in UTC", got, want)
	}
}

func TestNamesHoldTheirProperties(t *testing.T) {
	t.Parallel()

	key := newKey(t, secretNamed("one"))
	properties := gopter.NewProperties(nil)

	// Any moment from 1970 to 2100, and any zone within fourteen hours of UTC.
	moments := gen.Int64Range(0, 4_102_444_800).Map(func(seconds int64) time.Time { return time.Unix(seconds, 0).UTC() })
	zones := gen.IntRange(-14*60, 14*60).Map(func(minutes int) *time.Location {
		return time.FixedZone("offset", minutes*60)
	})

	properties.Property("a name is 16 characters of base64url that show nothing of the child", prop.ForAll(
		func(id string, at time.Time) bool {
			name := key.Of(id, at)
			_, err := base64.RawURLEncoding.DecodeString(name)
			return len(name) == 16 && err == nil && (id == "" || !strings.Contains(name, id))
		},
		gen.AnyString(), moments,
	))

	properties.Property("a moment is named by its month in UTC, whatever zone it is written in", prop.ForAll(
		func(id string, at time.Time, zone *time.Location) bool {
			return key.Of(id, at) == key.Of(id, at.In(zone))
		},
		gen.AnyString(), moments, zones,
	))

	properties.Property("every moment of a month gives one name", prop.ForAll(
		func(id string, at time.Time) bool {
			start := time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
			return key.Of(id, start) == key.Of(id, at) && key.Of(id, at) == key.Of(id, end)
		},
		gen.AnyString(), moments,
	))

	properties.Property("the next month gives another name", prop.ForAll(
		func(id string, at time.Time) bool {
			next := time.Date(at.Year(), at.Month()+1, 1, 0, 0, 0, 0, time.UTC)
			return key.Of(id, at) != key.Of(id, next)
		},
		gen.AnyString(), moments,
	))

	properties.Property("two children are two names", prop.ForAll(
		func(one, other string, at time.Time) bool {
			return one == other || key.Of(one, at) != key.Of(other, at)
		},
		gen.AnyString(), gen.Identifier(), moments,
	))

	properties.TestingRun(t)
}
