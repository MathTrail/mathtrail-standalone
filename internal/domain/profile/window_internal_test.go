package profile

import (
	"fmt"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// entries is a window of entries told apart by their order: an s makes a
// skipped task, anything else an answer.
func entries(kinds string) []Answer {
	window := make([]Answer, 0, len(kinds))
	for i, kind := range kinds {
		window = append(window, Answer{TaskID: fmt.Sprintf("tsk_%03d", i), Skipped: kind == 's'})
	}
	return window
}

// The window lets its oldest entry go, as it always has — unless that entry is
// one of the latest answers it keeps, when the oldest skipped task goes
// instead.
func TestTheWindowLetsTheRightEntryGo(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		window string
		next   string
		gone   string
	}{
		{"a full window of answers, and an answer", "aaaaaaaaaaaaaaaaaaaa", "a", "tsk_000"},
		{"a full window of answers, and a skip", "aaaaaaaaaaaaaaaaaaaa", "s", "tsk_000"},
		{"a skip at the front", "saaaaaaaaaaaaaaaaaaa", "a", "tsk_000"},
		{"five answers at the front, then skips", "aaaaasssssssssssssss", "s", "tsk_005"},
		{"six answers at the front, then skips", "aaaaaassssssssssssss", "s", "tsk_000"},
		{"the latest answers behind a skip", "asaaaasssssssssssssss"[:20], "s", "tsk_001"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Profile{Recent: entries(tc.window)}
			next := Answer{TaskID: "tsk_next", Skipped: tc.next == "s"}
			p.remember(&next)

			if len(p.Recent) != MaxRecent {
				t.Fatalf("the window holds %d entries, want %d", len(p.Recent), MaxRecent)
			}
			if slices.ContainsFunc(p.Recent, func(entry Answer) bool { return entry.TaskID == tc.gone }) {
				t.Errorf("%s is still in the window, want it the one let go", tc.gone)
			}
			if last := p.Recent[len(p.Recent)-1]; last.TaskID != "tsk_next" {
				t.Errorf("the window ends with %s, want the new entry", last.TaskID)
			}
		})
	}
}

// Whatever mix of answers and skipped tasks a child leaves behind, the window
// holds what the rule, the trial series and the progress screen need of it.
func TestTheWindowHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("full, in order, with the newest entry and the latest answers", prop.ForAll(
		func(skips []bool) string {
			p := &Profile{}
			var answers []string
			for i, skipped := range skips {
				entry := Answer{TaskID: fmt.Sprintf("tsk_%04d", i), Skipped: skipped}
				p.remember(&entry)
				if !skipped {
					answers = append(answers, entry.TaskID)
				}
				if broken := windowBroken(p.Recent, i+1, answers); broken != "" {
					return fmt.Sprintf("after %d entries: %s", i+1, broken)
				}
			}
			return ""
		},
		gen.SliceOf(gen.Weighted([]gen.WeightedGen{
			{Weight: 3, Gen: gen.Const(true)},
			{Weight: 1, Gen: gen.Const(false)},
		})),
	))

	properties.TestingRun(t)
}

// windowBroken says what about a window breaks its properties, after this
// many entries were remembered with these answers among them, or nothing.
func windowBroken(window []Answer, remembered int, answers []string) string {
	holds := func(id string) bool {
		return slices.ContainsFunc(window, func(entry Answer) bool { return entry.TaskID == id })
	}
	latest := answers[max(0, len(answers)-MinRecent):]
	switch {
	case len(window) != min(remembered, MaxRecent):
		return fmt.Sprintf("the window holds %d entries, want %d", len(window), min(remembered, MaxRecent))
	case window[len(window)-1].TaskID != fmt.Sprintf("tsk_%04d", remembered-1):
		return "the newest entry is not at the end"
	case !slices.IsSortedFunc(window, func(a, b Answer) int { return compareIDs(a.TaskID, b.TaskID) }):
		return "the entries are out of order"
	case !allOf(latest, holds):
		return fmt.Sprintf("one of the latest answers %v was let go", latest)
	}
	return ""
}

func compareIDs(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func allOf(ids []string, holds func(string) bool) bool {
	for _, id := range ids {
		if !holds(id) {
			return false
		}
	}
	return true
}
