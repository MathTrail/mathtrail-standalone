package progress_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
)

// wrong is a wrong answer to an option of topic, with trap behind it.
func wrong(topic, trap string) profile.Answer {
	return profile.Answer{Topic: topic, Chosen: "B", Trap: trap, TaskID: "task_" + trap}
}

// The map is drawn from the answers of each seed profile's window: at the
// threshold of a repeat there is nothing on it, since no seed fell for one trap
// twice, and at one it lists every trap the seed fell for, the last first.
func TestTheMapOfTheSeedProfiles(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		student string
		once    []progress.Mistake
	}{
		{"dima", []progress.Mistake{{Trap: "off_by_one", Times: 1}}},
		{"masha", []progress.Mistake{{Trap: "missed_case", Times: 1}}},
		{"olya", []progress.Mistake{
			{Trap: "double_count", Times: 1},
			{Trap: "answered_other_question", Times: 1},
		}},
		{"petya", []progress.Mistake{{Trap: "stopped_early", Times: 1}}},
		{"sasha", []progress.Mistake{}},
	} {
		t.Run(tc.student, func(t *testing.T) {
			t.Parallel()
			window, catalog := fixture(t, tc.student).Recent, embedded(t)

			if got := progress.Mistakes(window, catalog, 1); !slices.Equal(got, tc.once) {
				t.Errorf("Mistakes(%s, 1) = %v, want %v", tc.student, got, tc.once)
			}
			if got := progress.Mistakes(window, catalog, 2); got == nil || len(got) != 0 {
				t.Errorf("Mistakes(%s, 2) = %#v, want an empty map", tc.student, got)
			}
		})
	}
}

// A trap is on the map from the threshold on, and not a time before it; so is
// the word that it has come up before.
func TestATrapRepeatsFromTheThresholdOn(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	window := fixture(t, "petya").Recent
	window = append(window, wrong("counting.gaps", "off_by_one"))

	if got := progress.Mistakes(window, catalog, 2); len(got) != 0 {
		t.Errorf("once: Mistakes() = %v, want none", got)
	}
	if progress.Repeats(window, catalog, "off_by_one", 2) {
		t.Error("once: Repeats() = true, want false")
	}

	window = append(window, wrong("time.clocks", "off_by_one"))

	want := []progress.Mistake{{Trap: "off_by_one", Times: 2}}
	if got := progress.Mistakes(window, catalog, 2); !slices.Equal(got, want) {
		t.Errorf("twice: Mistakes() = %v, want %v", got, want)
	}
	if !progress.Repeats(window, catalog, "off_by_one", 2) {
		t.Error("twice: Repeats() = false, want true")
	}
	if got := progress.Mistakes(window, catalog, 3); len(got) != 0 {
		t.Errorf("twice at a threshold of three: Mistakes() = %v, want none", got)
	}
	if progress.Repeats(window, catalog, "off_by_one", 3) {
		t.Error("twice at a threshold of three: Repeats() = true, want false")
	}
}

// Only a wrong answer to an option has a trap behind it: a right answer, "I
// don't know" and a task left without an answer take none.
func TestOnlyAWrongOptionCountsForATrap(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	window := []profile.Answer{
		{Topic: "counting.gaps", Correct: true},
		{Topic: "counting.gaps", Confused: true},
		{Topic: "counting.gaps", Skipped: true},
		wrong("counting.gaps", "off_by_one"),
	}

	want := []progress.Mistake{{Trap: "off_by_one", Times: 1}}
	if got := progress.Mistakes(window, catalog, 1); !slices.Equal(got, want) {
		t.Errorf("Mistakes() = %v, want %v", got, want)
	}
	if progress.Repeats(window, catalog, "", 1) {
		t.Error(`Repeats("") = true, want false: no trap is no mistake`)
	}
}

// The most frequent mistake comes first, and of two as frequent the one the
// child fell for last: the map reads as what to look at first.
func TestTheMapShowsTheMostFrequentAndTheLatestFirst(t *testing.T) {
	t.Parallel()

	window := []profile.Answer{
		wrong("counting.gaps", "off_by_one"),
		wrong("combinatorics.enumeration", "missed_case"),
		wrong("combinatorics.enumeration", "double_count"),
		wrong("combinatorics.enumeration", "missed_case"),
		wrong("counting.gaps", "off_by_one"),
		wrong("combinatorics.enumeration", "missed_case"),
		wrong("combinatorics.enumeration", "double_count"),
	}

	want := []progress.Mistake{
		{Trap: "missed_case", Times: 3},
		{Trap: "double_count", Times: 2},
		{Trap: "off_by_one", Times: 2},
	}
	if got := progress.Mistakes(window, embedded(t), 2); !slices.Equal(got, want) {
		t.Errorf("Mistakes() = %v, want %v", got, want)
	}
}

// A trap the catalog does not have — a file edited by hand, a trap since
// retired — is neither on the map nor said to repeat, however often the window
// holds it: the answer and the map never disagree.
func TestATrapTheCatalogDoesNotHaveNeverRepeats(t *testing.T) {
	t.Parallel()

	catalog := embedded(t)
	window := []profile.Answer{
		wrong("counting.gaps", "counted_the_cat"),
		wrong("counting.gaps", "counted_the_cat"),
		wrong("counting.gaps", "off_by_one"),
		wrong("counting.gaps", "off_by_one"),
	}

	want := []progress.Mistake{{Trap: "off_by_one", Times: 2}}
	if got := progress.Mistakes(window, catalog, 2); !slices.Equal(got, want) {
		t.Errorf("Mistakes() = %v, want %v", got, want)
	}
	if progress.Repeats(window, catalog, "counted_the_cat", 2) {
		t.Error("Repeats(counted_the_cat) = true, want false: the catalog has no such trap")
	}
}
