package report

import (
	"cmp"
	"maps"
	"slices"
	"time"
)

// A task written while the child waits comes to the card that waits for it
// when the card's question finds it. These count how long after the hand-in
// that accepted a task the card's question brought it.

// toolSubmitTask is the tool a task is handed in with, and toolReadTask the
// one a waiting card asks how its task stands with.
const (
	toolSubmitTask = "submit_task"
	toolReadTask   = "read_task"
)

// shownWithin is the longest after its hand-in a task counts as shown: past the
// longest a task is waited for, a card that asks was drawn again with its chat
// rather than waiting.
const shownWithin = 15 * time.Minute

// shownAfterHandIn are the milliseconds from each hand-in that accepted a task
// — a call of submit_task that answered with the task — to the first question
// of a card that brought it once the hand-in had begun — a call of read_task
// that answered with the task of the same request —, by the version of the
// instructions and the host of the hand-in. A call's line is written as the
// call answers, so both are counted from the moment the answer left; a
// question the hand-in woke may answer before the hand-in's own line is
// written, and counts as nothing. A task no card asked for after its hand-in,
// in a chat with no card, is not counted, nor one a card asked for only past
// shownWithin.
func shownAfterHandIn(lines []line) map[group][]int64 {
	brought := map[string][]time.Time{}
	for i := range lines {
		if l := &lines[i]; l.Message == eventToolCall && l.Tool == toolReadTask && l.Screen == screenTask && l.TaskRequest != "" {
			brought[l.TaskRequest] = append(brought[l.TaskRequest], l.Time)
		}
	}
	shown := map[group][]int64{}
	for i := range lines {
		l := &lines[i]
		if l.Message != eventToolCall || l.Tool != toolSubmitTask || l.Outcome != "ok" || l.Screen != screenTask ||
			l.TaskRequest == "" {
			continue
		}
		began := l.Time.Add(-time.Duration(l.DurationMS) * time.Millisecond)
		if at, found := firstFrom(brought[l.TaskRequest], began); found && at.Sub(l.Time) <= shownWithin {
			g := group{version: l.InstructionsVersion, host: l.Client}
			shown[g] = append(shown[g], max(at.Sub(l.Time).Milliseconds(), 0))
		}
	}
	return shown
}

// firstFrom is the earliest of the moments that is not before from, if any is.
func firstFrom(moments []time.Time, from time.Time) (time.Time, bool) {
	var first time.Time
	found := false
	for _, at := range moments {
		if !at.Before(from) && (!found || at.Before(first)) {
			first, found = at, true
		}
	}
	return first, found
}

// shownTable is how long after the hand-in that accepted a task the waiting
// card's question brought it, by the version of the instructions and the host.
func (c *counts) shownTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Shown", "Milliseconds, median", "Milliseconds, 95th percentile", "Longest",
	}, named: 2}
	keys := slices.Collect(maps.Keys(c.shown))
	slices.SortFunc(keys, func(a, b group) int {
		return cmp.Or(cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
			cmp.Compare(a.version, b.version), cmp.Compare(a.host, b.host))
	})
	for _, key := range keys {
		took := slices.Sorted(slices.Values(c.shown[key]))
		t.add(key.version, key.host, number(len(took)), number(atRank(took, 50)), number(atRank(took, 95)),
			number(took[len(took)-1]))
	}
	return t
}
