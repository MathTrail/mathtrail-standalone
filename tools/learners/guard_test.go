package main

import (
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A number is within its band when it is within two half-widths of the
// recorded one, the edges included; a number the run did not read is not.
func TestABandHoldsWhatIsWithinTwoHalfWidths(t *testing.T) {
	t.Parallel()
	b := band{guardedMeasure: guardedMeasure{staticChildren, "r4_false"}, value: 0.5, halfWidth: 0.125}
	for _, tc := range []struct {
		got    float64
		has    bool
		inside bool
	}{
		{0.5, true, true},
		{0.25, true, true},
		{0.75, true, true},
		{0.2499, true, false},
		{0.7501, true, false},
		{0.5, false, false},
	} {
		r := guardReading{band: b, got: tc.got, has: tc.has}
		if got := r.inside(); got != tc.inside {
			t.Errorf("%v, read %v, against 0.5 ± 2 × 0.125: inside %v, want %v", tc.got, tc.has, got, tc.inside)
		}
	}
}

// A number outside its band is named by its generator and measure, with the
// number read, both ends of the band, and what it was recorded as.
func TestANumberOutsideItsBandIsNamedWithTheBand(t *testing.T) {
	t.Parallel()
	r := guardReading{
		band: band{guardedMeasure: guardedMeasure{staticChildren, "r4_false"}, value: 0.036, halfWidth: 0.01},
		got:  0.635, has: true,
	}
	why := r.why()
	for _, part := range []string{"G0", "r4_false", "0.6350", "0.0160", "0.0560", "0.0360", "0.0100"} {
		if !strings.Contains(why, part) {
			t.Errorf("the guard says %q, want it to name %s", why, part)
		}
	}
}

// The bands kept are the guarded numbers, every one of them, in the guard's
// order, each with an interval of some width.
func TestTheBandsKeptAreTheGuardedNumbers(t *testing.T) {
	t.Parallel()
	bands, err := readBands(guardSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	kept := make([]guardedMeasure, 0, len(bands))
	for _, b := range bands {
		kept = append(kept, b.guardedMeasure)
		if b.halfWidth <= 0 {
			t.Errorf("%s %s is kept with a half-width of %v, want one above zero", b.generator, b.metric, b.halfWidth)
		}
	}
	if !slices.Equal(kept, guarded) {
		t.Errorf("the bands kept are %v, want %v", kept, guarded)
	}
}

// The bands are kept to their last digit, so a band reads back as it was
// recorded: a number of no short decimal, a half-width of nothing, a value
// below zero.
func TestTheBandsAreKeptToTheLastDigit(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "guard.csv")
	bands := []band{
		{guardedMeasure: guardedMeasure{staticChildren, "r4_false"}, value: 1.0 / 3, halfWidth: 0},
		{guardedMeasure: guardedMeasure{learning, "r6_lag"}, value: -0.9983999999999999, halfWidth: 0.033271428571428574},
	}
	if err := writeBands(path, bands); err != nil {
		t.Fatal(err)
	}
	read, err := readBands(path)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(read, bands) {
		t.Errorf("the bands read back are %v, want %v", read, bands)
	}
}

// The guard takes nothing but -update, no flag of the bench's among it: its
// bands mean something only on the run they were recorded on. It refuses
// before it runs anything.
func TestTheGuardTakesNothingButUpdate(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"-children", "10"}, {"-rules", "bench"}, {"-seed", "7"}, {"-out", "elsewhere"}, {"more"}, {"-update", "more"}} {
		if code := guardCommand(args, io.Discard, io.Discard); code != 2 {
			t.Errorf("learners guard %q exited %d, want 2", args, code)
		}
	}
}

// What the guard records it reads again, every number inside its band; and
// the service's earlier rule, held to the same bands, is caught by its
// masteries, most of them false where the service's are not.
func TestTheGuardReadsWhatItRecordedAndCatchesTheEarlierRule(t *testing.T) {
	t.Parallel()
	w, err := newWorld(guardAnswers)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := filepath.Join(t.TempDir(), "guard.csv")
	recorded := guard{w: w, r: serviceRule(), children: 10, snapshot: snapshot}
	bands, err := recorded.record()
	if err != nil {
		t.Fatal(err)
	}
	if failed := writeBands(snapshot, bands); failed != nil {
		t.Fatal(failed)
	}
	readings, err := recorded.judge()
	if err != nil {
		t.Fatal(err)
	}
	for i := range readings {
		if !readings[i].inside() {
			t.Errorf("the service read against the bands just recorded from it: %s", readings[i].why())
		}
	}

	earlier := guard{w: w, r: earlierServiceRule(), children: 10, snapshot: snapshot}
	theirs, err := earlier.judge()
	if err != nil {
		t.Fatal(err)
	}
	falseMasteries := slices.IndexFunc(theirs, func(r guardReading) bool {
		return r.generator == staticChildren && r.metric == "r4_false"
	})
	if falseMasteries < 0 || theirs[falseMasteries].inside() {
		t.Error("the earlier rule's false masteries are inside the service's band, want them outside")
	}

	// A file of bands that has lost a guarded number is refused, rather than
	// read as if that number were not guarded.
	if failed := writeBands(snapshot, bands[1:]); failed != nil {
		t.Fatal(failed)
	}
	if _, failed := recorded.judge(); !errors.Is(failed, errUnguarded) {
		t.Errorf("judging against bands with one lost: error %v, want %v", failed, errUnguarded)
	}
}
