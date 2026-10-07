package report

import (
	"cmp"
	"maps"
	"slices"
	"time"
)

// A tool makes the write the child need not wait for after it has answered: a
// task written ahead, handed out on the card next_task draws, and an answer
// recorded. These count how soon such a task reached the card, and how the
// writes after the answer went.

// eventWriteAfterAnswer is the line a write after the answer leaves,
// toolNextTask the tool that hands out a task written ahead, and screenTask
// the screen its answer draws when it does: the task, which next_task draws
// for no other reason.
const (
	eventWriteAfterAnswer = "write_after_answer"
	toolNextTask          = "next_task"
	screenTask            = "task"
)

// lateOutcomes are how a write after the answer may go, in the order its
// table shows them.
var lateOutcomes = []string{"written", "remade", "already", "lost", "failed"}

// lateOf is the writes after the answer one tool made under one version of the
// instructions.
type lateOf struct{ version, tool string }

// lateWrites is how the writes after the answer of one tool went: how many
// went each way, how long each took from the answer to its end, and how many
// calls of the same account began on another instance while one was under way.
type lateWrites struct {
	outcomes     map[string]int
	milliseconds []int64
	elsewhere    int
}

// handOuts are how long next_task took to hand out a task written ahead: the
// whole call, and the call less its calls to Drive made before it answered.
type handOuts struct {
	milliseconds, own []int64
}

// keptHandedOut are the requests whose line about the task accepted says it
// had been written ahead and kept ready.
func keptHandedOut(lines []line) map[string]bool {
	handed := map[string]bool{}
	for i := range lines {
		if request := lines[i].request(); lines[i].Message == eventTaskAccepted && lines[i].Ready && request != "" {
			handed[request] = true
		}
	}
	return handed
}

// handedOutKept says whether a call handed out a task written ahead: a call of
// next_task whose answer drew the task, or, for a line written before the
// service named the screen, whose request accepted a task kept ready. A task
// whose write after the answer failed was handed out all the same: the card
// drew it.
func handedOutKept(l *line, kept map[string]bool) bool {
	return l.Tool == toolNextTask && (l.Screen == screenTask || kept[l.request()])
}

// handOut counts a call of next_task that handed out a task written ahead.
func (c *counts) handOut(l *line, own int64) {
	g := group{version: l.InstructionsVersion, host: l.Client}
	counted, found := c.handOuts[g]
	if !found {
		counted = &handOuts{}
		c.handOuts[g] = counted
	}
	counted.milliseconds = append(counted.milliseconds, int64(l.DurationMS))
	counted.own = append(counted.own, own)
}

// writesAfterAnswer counts the writes after the answer by the version of the
// instructions and the tool that made them, with the calls of the same account that began on another instance
// while one was under way: a call there read the file as it was before the
// write.
func writesAfterAnswer(lines []line) map[lateOf]*lateWrites {
	callsOf := map[string][]*line{}
	for i := range lines {
		if l := &lines[i]; l.Message == eventToolCall && l.User != "" && l.Instance != "" {
			callsOf[l.User] = append(callsOf[l.User], l)
		}
	}
	late := map[lateOf]*lateWrites{}
	for i := range lines {
		l := &lines[i]
		if l.Message != eventWriteAfterAnswer {
			continue
		}
		key := lateOf{version: l.InstructionsVersion, tool: l.Tool}
		counted, found := late[key]
		if !found {
			counted = &lateWrites{outcomes: map[string]int{}}
			late[key] = counted
		}
		counted.outcomes[l.Outcome]++
		counted.milliseconds = append(counted.milliseconds, int64(l.DurationMS))
		counted.elsewhere += startedElsewhere(callsOf[l.User], l)
	}
	return late
}

// startedElsewhere counts the calls that began on another instance than the
// write's while the write was under way.
func startedElsewhere(calls []*line, write *line) int {
	if write.Instance == "" {
		return 0
	}
	began := write.Time.Add(-time.Duration(write.DurationMS) * time.Millisecond)
	elsewhere := 0
	for _, call := range calls {
		start := call.Time.Add(-time.Duration(call.DurationMS) * time.Millisecond)
		if call.Instance != write.Instance && !start.Before(began) && !start.After(write.Time) {
			elsewhere++
		}
	}
	return elsewhere
}

// handOutsTable is how long next_task took to hand out a task written ahead,
// by the version of the instructions and the chat host.
func (c *counts) handOutsTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Handed out", "Milliseconds, median", "Milliseconds, 95th percentile",
		"Without Drive, median", "Without Drive, 95th percentile",
	}, named: 2}
	keys := slices.Collect(maps.Keys(c.handOuts))
	slices.SortFunc(keys, func(a, b group) int {
		return cmp.Or(cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
			cmp.Compare(a.version, b.version), cmp.Compare(a.host, b.host))
	})
	for _, key := range keys {
		counted := c.handOuts[key]
		took := slices.Sorted(slices.Values(counted.milliseconds))
		own := slices.Sorted(slices.Values(counted.own))
		t.add(key.version, key.host, number(len(took)), number(atRank(took, 50)), number(atRank(took, 95)),
			number(atRank(own, 50)), number(atRank(own, 95)))
	}
	return t
}

// lateTable is how the writes after the answer went, by the version of the
// instructions and the tool that made them.
func (c *counts) lateTable() *table {
	t := &table{columns: []string{
		"Instructions", "Tool", "Writes", "Written", "Remade", "Already", "Lost", "Failed",
		"Milliseconds, median", "Milliseconds, 95th percentile", "Longest", "Calls elsewhere meanwhile",
	}, named: 2}
	keys := slices.Collect(maps.Keys(c.late))
	slices.SortFunc(keys, func(a, b lateOf) int {
		return cmp.Or(cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
			cmp.Compare(a.version, b.version), cmp.Compare(a.tool, b.tool))
	})
	for _, key := range keys {
		counted := c.late[key]
		took := slices.Sorted(slices.Values(counted.milliseconds))
		cells := []string{key.version, key.tool, number(len(took))}
		for _, outcome := range lateOutcomes {
			cells = append(cells, number(counted.outcomes[outcome]))
		}
		t.add(append(cells, number(atRank(took, 50)), number(atRank(took, 95)), number(took[len(took)-1]),
			number(counted.elsewhere))...)
	}
	return t
}
