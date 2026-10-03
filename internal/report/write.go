package report

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
)

// toolOutcomes are the ways a tool call ends, in the order the report gives
// them.
var toolOutcomes = []string{"ok", "refused", "failed", "invalid"}

// write writes the report: what the lines span, then a table for each thing
// they count, each under a sentence that says how to read it.
func write(out io.Writer, c *counts) error {
	var report strings.Builder
	report.WriteString("# What the log adds up to\n\n" + c.spanned() + "\n")
	for _, section := range []struct {
		title, about string
		table        *table
	}{
		{"Tasks", "Asked for counts the requests opened; handed in, the attempts judged, of which the checks " +
			"refused some; out of attempts, the requests whose last attempt was refused.", c.tasksTable()},
		{"Accepted tasks", "The attempts an accepted task took, and the seconds from its request to its " +
			"acceptance: the time the chat's model took to write it.", c.acceptedTable()},
		{"Why attempts were refused", "An attempt is counted by one check, the first its refusal names, and " +
			"fails that check and any others beside it.", c.refusalsTable()},
		{"Chances and what came of them", c.promisesAbout(), c.promisesTable()},
		{"How the estimate keeps up", "The same answers by which of the child's answers each was: a right answer " +
			"as 1 and a wrong one as 0, less the chance promised, on average, with the standard error of that " +
			"mean. Above zero the child did better than the estimate promised, which is how an estimate that " +
			"falls behind a learning child shows; below zero, worse.", c.keptUpTable()},
		{"Limits reached", "A pace writes one line for a flood of refusals, and a day's ceiling one for every " +
			"call it refused.", c.limitsTable()},
		{"Tool calls", "Milliseconds are the service's own time for a call, its calls to Drive included.",
			c.toolsTable()},
	} {
		fmt.Fprintf(&report, "\n## %s\n\n%s\n\n", section.title, section.about)
		section.table.writeTo(&report)
	}
	_, err := io.WriteString(out, report.String())
	return err
}

// spanned says how many lines were read, over what time, how many were left
// out as not the service's, and how many of the service's could not be read.
func (c *counts) spanned() string {
	var said string
	switch {
	case c.lines == 0:
		said = "No lines of the service's were read."
	case c.from.IsZero():
		said = fmt.Sprintf("%d lines of the service's, with no time on them.", c.lines)
	default:
		said = fmt.Sprintf("%d lines of the service's, from %s to %s.", c.lines, moment(c.from), moment(c.to))
	}
	if c.others > 0 {
		said += fmt.Sprintf(" %d more lines, not the service's, were left out.", c.others)
	}
	if c.unreadable > 0 {
		said += fmt.Sprintf(" %d lines of the service's did not read as this report expects, and are not counted.",
			c.unreadable)
	}
	return said
}

// tasksTable is what became of the tasks of each group.
func (c *counts) tasksTable() *table {
	t := &table{columns: []string{"Instructions", "Host", "Asked for", "Handed in", "Refused", "Out of attempts"}, named: 2}
	for _, g := range c.groups() {
		if counted := c.tasks[g]; counted.asked+counted.handedIn > 0 {
			t.add(g.version, g.host, number(counted.asked), number(counted.handedIn), number(counted.refused),
				number(counted.outOfAttempts))
		}
	}
	return t
}

// acceptedTable is, for each group, the tasks accepted, the attempts they took
// and how long they took to write.
func (c *counts) acceptedTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Accepted", "At the first attempt", "Attempts, mean",
		"Seconds, median", "Seconds, 90th percentile", "Seconds, longest",
	}, named: 2}
	for _, g := range c.groups() {
		counted := c.tasks[g]
		if len(counted.attempts) == 0 {
			continue
		}
		seconds := slices.Sorted(slices.Values(counted.seconds))
		t.add(g.version, g.host, number(len(counted.attempts)), number(countOf(counted.attempts, 1)),
			strconv.FormatFloat(mean(counted.attempts), 'f', 1, 64),
			number(atRank(seconds, 50)), number(atRank(seconds, 90)), number(atRank(seconds, 100)))
	}
	return t
}

// refusalsTable is the checks that refused attempts under each version, the
// one that refused the most first.
func (c *counts) refusalsTable() *table {
	t := &table{columns: []string{"Instructions", "Check", "Counted by it", "Failed it"}, named: 2}
	keys := slices.Collect(maps.Keys(c.refusals))
	slices.SortFunc(keys, func(a, b refusedBy) int {
		return cmp.Or(
			cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
			cmp.Compare(c.refusals[b].counted, c.refusals[a].counted),
			cmp.Compare(a.check, b.check),
		)
	})
	for _, key := range keys {
		t.add(key.version, key.check, number(c.refusals[key].counted), number(c.refusals[key].failed))
	}
	return t
}

// limitsTable is how many lines each limit reached left.
func (c *counts) limitsTable() *table {
	t := &table{columns: []string{"Limit", "Lines"}, named: 1}
	for _, limit := range slices.Sorted(maps.Keys(c.limits)) {
		t.add(limit, number(c.limits[limit]))
	}
	return t
}

// toolsTable is how the calls of each tool ended, and how long they took, by
// the host that made them.
func (c *counts) toolsTable() *table {
	t := &table{columns: []string{
		"Host", "Tool", "Calls", "Answered", "Refused", "Failed", "Invalid",
		"Milliseconds, median", "Milliseconds, 95th percentile",
	}, named: 2}
	keys := slices.Collect(maps.Keys(c.tools))
	slices.SortFunc(keys, func(a, b toolOf) int { return cmp.Or(cmp.Compare(a.host, b.host), cmp.Compare(a.tool, b.tool)) })
	for _, key := range keys {
		called := c.tools[key]
		cells := []string{key.host, key.tool, number(len(called.milliseconds))}
		for _, outcome := range toolOutcomes {
			cells = append(cells, number(called.outcomes[outcome]))
		}
		took := slices.Sorted(slices.Values(called.milliseconds))
		t.add(append(cells, number(atRank(took, 50)), number(atRank(took, 95)))...)
	}
	return t
}

// groups are the groups tasks were counted for, in the order compareGroups
// puts them.
func (c *counts) groups() []group {
	groups := slices.Collect(maps.Keys(c.tasks))
	slices.SortFunc(groups, c.compareGroups)
	return groups
}

// compareGroups orders groups by version, in the order the versions first
// appear, and by host within a version.
func (c *counts) compareGroups(a, b group) int {
	return cmp.Or(
		cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
		cmp.Compare(a.host, b.host),
	)
}

// table is one table of the report: a heading for each column, and its rows.
// The first columns, as many as named says, name what a row is about; the rest
// hold numbers, which stand to the right.
type table struct {
	columns []string
	named   int
	rows    [][]string
}

// add adds a row of cells.
func (t *table) add(cells ...string) { t.rows = append(t.rows, cells) }

// writeTo writes the table in Markdown, or says there is nothing to put in it.
func (t *table) writeTo(b *strings.Builder) {
	if len(t.rows) == 0 {
		b.WriteString("None in these lines.\n")
		return
	}
	alignments := make([]string, len(t.columns))
	for i := range alignments {
		alignments[i] = "---:"
		if i < t.named {
			alignments[i] = "---"
		}
	}
	writeRow(b, t.columns)
	b.WriteString("|" + strings.Join(alignments, "|") + "|\n")
	for _, row := range t.rows {
		writeRow(b, row)
	}
}

// writeRow writes one row of a table. A cell with nothing in it says so, and
// a bar in a cell stays in the cell.
func writeRow(b *strings.Builder, cells []string) {
	written := make([]string, len(cells))
	for i, cell := range cells {
		written[i] = cmp.Or(strings.ReplaceAll(cell, "|", `\|`), "(none)")
	}
	b.WriteString("| " + strings.Join(written, " | ") + " |\n")
}

// number is a whole number as the report writes it.
func number[T int | int64](n T) string { return strconv.FormatInt(int64(n), 10) }

// moment is a moment as the report writes it, in UTC to the second.
func moment(at time.Time) string { return at.UTC().Format("2006-01-02 15:04:05 UTC") }

// countOf is how many of values are the value given.
func countOf(values []int, value int) int {
	counted := 0
	for _, each := range values {
		if each == value {
			counted++
		}
	}
	return counted
}
