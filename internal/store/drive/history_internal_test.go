package drivestore

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
)

// historyOf is a file's history of as many revisions as kept says, made a
// minute apart, the ones kept says kept forever — listed in the order the
// shuffle gives, since Drive promises none.
func historyOf(kept []bool, shuffle []int) []drive.Revision {
	made := time.Date(2026, time.September, 28, 9, 0, 0, 0, time.UTC)
	history := make([]drive.Revision, len(kept))
	for i, keep := range kept {
		history[i] = drive.Revision{ID: strconv.Itoa(i), ModifiedTime: made.Add(time.Duration(i) * time.Minute), KeepForever: keep}
	}
	if len(history) == 0 {
		return history
	}
	for i, j := range shuffle {
		a, b := i%len(history), j%len(history)
		history[a], history[b] = history[b], history[a]
	}
	return history
}

// The revisions a recovery tries are at most five, the latest first, never the
// one the file holds now, the latest four before it, and the latest one kept
// forever whenever there is one — in whatever order Drive listed them.
func TestTheCandidatesOfARecoveryHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	kept := gen.SliceOf(gen.Bool())
	shuffle := gen.SliceOf(gen.IntRange(0, 1<<10))

	properties.Property("the candidates are the latest before the damage, and the latest kept", prop.ForAll(
		func(kept []bool, shuffle []int) bool {
			tried := candidates(historyOf(kept, shuffle))
			if len(kept) < 2 {
				return len(tried) == 0
			}
			head := strconv.Itoa(len(kept) - 1)
			latestKept := ""
			for i := len(kept) - 2; i >= 0 && latestKept == ""; i-- {
				if kept[i] {
					latestKept = strconv.Itoa(i)
				}
			}
			ids := make([]string, 0, len(tried))
			for _, revision := range tried {
				ids = append(ids, revision.ID)
			}
			wantRecent := make([]string, 0, recent)
			for i := len(kept) - 2; i >= 0 && len(wantRecent) < recent; i-- {
				wantRecent = append(wantRecent, strconv.Itoa(i))
			}
			latestFirst := slices.IsSortedFunc(tried, func(a, b drive.Revision) int { return b.ModifiedTime.Compare(a.ModifiedTime) })
			return len(tried) <= recent+1 &&
				latestFirst &&
				!slices.Contains(ids, head) &&
				slices.Equal(ids[:len(wantRecent)], wantRecent) &&
				(latestKept == "" || slices.Contains(ids, latestKept))
		},
		kept, shuffle,
	))
	properties.TestingRun(t)
}

// The day's first write keeps its revision forever. Whatever run of writes
// comes — some lost to another, which leave the file as it was — no day after
// the one the profile was made on passes without one kept, if a write landed
// on it. The day the profile was made keeps nothing forever: its first state
// is the new profile, and the next day's first write keeps where it ended.
func TestNoDayOfUsePassesWithoutAStateKept(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	// Each write comes some hours after the one before, and lands or is lost.
	hours := gen.SliceOf(gen.IntRange(0, 30))
	landed := gen.SliceOf(gen.Bool())

	properties.Property("every day after the first that a write landed on has one kept", prop.ForAll(
		func(hours []int, landed []bool) bool {
			at := time.Date(2026, time.September, 1, 7, 0, 0, 0, time.UTC)
			made := at.Format(time.DateOnly)
			file := fileWrittenAt(at)
			keptOn, landedOn := map[string]bool{}, map[string]bool{}
			for i, gap := range hours {
				at = at.Add(time.Duration(gap) * time.Hour)
				if i < len(landed) && !landed[i] {
					continue
				}
				day := at.Format(time.DateOnly)
				landedOn[day] = true
				if firstOfTheDay(file, writtenAt(at)) {
					keptOn[day] = true
				}
				file = fileWrittenAt(at)
			}
			for day := range landedOn {
				if day != made && !keptOn[day] {
					return false
				}
			}
			return true
		},
		hours, landed,
	))
	properties.TestingRun(t)
}

// A write is kept when the file does not say when it was written.
func TestAFileThatDoesNotSayWhenItWasWrittenIsKept(t *testing.T) {
	t.Parallel()

	now := writtenAt(time.Date(2026, time.September, 28, 9, 0, 0, 0, time.UTC))
	for _, raw := range [][]byte{nil, []byte("not json"), []byte(`{"updated_at": "yesterday"}`), []byte(`{}`)} {
		if !firstOfTheDay(raw, now) {
			t.Errorf("firstOfTheDay(%q) = false, want the write kept", raw)
		}
	}
}

// writtenAt is a profile touched at the moment given.
func writtenAt(at time.Time) *profile.Profile {
	p := profile.New(profile.Student{Pseudonym: "Mia", Grade: 3}, "drivestore_test", at)
	p.Touch("drivestore_test", at)
	return p
}

// fileWrittenAt is the file of a profile last written at the moment given.
func fileWrittenAt(at time.Time) []byte {
	raw, err := profile.Marshal(writtenAt(at))
	if err != nil {
		panic(err)
	}
	return raw
}

// A write dated before the day the file was last written on — an instance
// whose clock runs behind another's — is no new day, and keeps nothing.
func TestAWriteDatedBeforeTheFilesDayIsNoNewDay(t *testing.T) {
	t.Parallel()

	written := time.Date(2026, time.September, 28, 0, 0, 5, 0, time.UTC)
	for _, tc := range []struct {
		at   time.Time
		keep bool
	}{
		{written.Add(-10 * time.Second), false},
		{written.Add(time.Hour), false},
		{written.Add(24 * time.Hour), true},
	} {
		if got := firstOfTheDay(fileWrittenAt(written), writtenAt(tc.at)); got != tc.keep {
			t.Errorf("firstOfTheDay(file of %v, write at %v) = %t, want %t", written, tc.at, got, tc.keep)
		}
	}
}
