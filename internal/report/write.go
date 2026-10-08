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

// notLogged stands for a count no line of a group says anything about, such
// as the drawings of tasks accepted before the service said whether there was
// one. A blank cell would read as none.
const notLogged = "(not logged)"

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
			"acceptance: the time the chat's model took to write it — for a task written ahead, the time its " +
			"writing took, though the child had it at once. With a drawing counts the tasks that came " +
			"with one; where only some lines of a group say whether they did, it counts among those, as 2 of 3, and " +
			"where none says, it is not logged.", c.acceptedTable()},
		{"Tasks written ahead", "Asked ahead counts the requests opened for the next task while the child worked " +
			"on the one on the card, and kept the tasks written for them and kept. Handed out ready counts the " +
			"tasks handed out that had been kept, out of all handed out; taken by the card, the tasks a card took " +
			"into its own place, kept or waited for, which only an earlier release let a card do; let go, the " +
			"tasks written ahead, kept or still being written, that the lesson moved away from before they were " +
			"handed out.", c.aheadTable()},
		{"Why tasks written ahead were let go", "The reason the lesson moved away from a task: another language " +
			"of the lessons, another topic the lessons are kept to, a skill kept out since, another place a person " +
			"asked for, or another version of the instructions. Written says whether the task had been written and " +
			"kept, or was still being written.", c.letGoTable()},
		{"Tasks written ahead, handed out", "How long next_task took when it handed out a task written ahead, " +
			"which the card shows the moment the call answers: the whole call, and the same less its calls to Drive " +
			"made before it answered. The hand-out is written after the answer and is part of neither; a call is " +
			"counted whether that write went through or not, since the card drew the task all the same.",
			c.handOutsTable()},
		{"Tasks accepted, shown on the card", "How long after the hand-in that accepted a task written while the " +
			"child waited the waiting card's question brought it: from the answer of submit_task to the answer of " +
			"the first read_task with the task of the same request once the hand-in had begun, by the version of " +
			"the instructions and the host of the hand-in. A question held for news is answered as the hand-in " +
			"lands on its instance, and counts as nothing when it answered before the hand-in did; one on another " +
			"instance, or one not held, finds the task at the next question. A card that asked first only " +
			"more than a quarter of an hour after, one drawn again with its chat, is left out.", c.shownTable()},
		{"Hand-ins by their parts", "How large the tasks handed in were, in bytes of the JSON the checks read: " +
			"the task, the self-check, the solver, the core idea inside the task, and the whole, the brief among " +
			"it when one came. Format now is a hand-in of the form the guide asks for; before, one that also " +
			"brought what the form no longer reads, from a chat begun with an earlier guide. Mended counts the " +
			"hand-ins with a field read as it was meant rather than as it was written.", c.handInsTable()},
		{"What was mended", "The fields read as they were meant rather than as they were written — a letter in " +
			"another case, an option written as a number, no issues written as null — by how many hand-ins.",
			c.mendsTable()},
		{"Drawings by topic", "The tasks accepted on each topic whose line says whether they came with a " +
			"drawing, and how many of them did. A line written before the service said so is left out.",
			c.drawingsTable()},
		{"Why attempts were refused", "An attempt is counted by one check, the first its refusal names, and " +
			"fails that check and any others beside it.", c.refusalsTable()},
		{"Chances and what came of them", c.promisesAbout(), c.promisesTable()},
		{"How the estimate keeps up", "The same answers by which of the child's answers each was: a right answer " +
			"as 1 and a wrong one as 0, less the chance promised, on average, with the standard error of that " +
			"mean. Above zero the child did better than the estimate promised, which is how an estimate that " +
			"falls behind a learning child shows; below zero, worse. The numbers are to three places, so that the rows " +
			"of two versions can be set side by side.", c.keptUpTable()},
		{"Later answers against earlier", c.laterAbout(), c.laterTable()},
		{"Masteries taken back", c.masteriesAbout(), c.masteriesTable()},
		{"Limits reached", "A pace writes one line for a flood of refusals, and a day's ceiling one for every " +
			"call it refused.", c.limitsTable()},
		{"Tool calls", "Milliseconds are the service's own time for a call, from its start to its answer, its calls " +
			"to Drive included; without Drive, the same less the time its calls to Drive took before it answered, " +
			"tied to the call by the request they were made in. A write made after the answer is in neither. A " +
			"card's question held for news, read_task's, takes a few seconds more while it waits.",
			c.toolsTable()},
		{"Writes after the answer", "The writes the child need not wait for — a task written ahead handed out, " +
			"an answer recorded — are made once the call has answered, while its request is held open; here by " +
			"the version of the instructions and the tool. Written counts those written as they were made; remade, " +
			"those made again on a fresh read, because another writer had changed the file meanwhile or it could " +
			"not be reached; already, those another writer had made; lost, those the file no longer allowed, so " +
			"that what the call answered is not what the file holds; failed, those that could not be written. " +
			"Milliseconds run from the hand-over, as the call answers, to the write's end. Calls " +
			"elsewhere meanwhile counts the calls of the same account that began on another instance while a " +
			"write was under way: such a call read the file as it was before the write.", c.lateTable()},
		{"Traces", c.tracesAbout(), c.tracesTable()},
		{"The busiest minute", c.busyAbout(), c.busyTable()},
		{"Children, a day at a time", "The children the counts for grant applications are taken from, by the rules " +
			"those counts follow: a child is counted under the name it has that month, on each day a task was handed " +
			"to it; a child the load tool or MCP Inspector handed a task to is left out, and so is every line of its; " +
			"a line the log repeated is counted once; a day is a day in UTC. The same days of the view " +
			"`impact_private.daily` hold the same numbers.", c.childrenTable()},
		{"The rules of the log", "Every line of the service's is held to an event the service is decided to " +
			"write, to the fields decided for that event, to the form each field of the lines the children are " +
			"counted from may hold, and to carrying nothing shaped like an email address. A " +
			"line that breaks one is named by its event and its field, never by what it held: an event the table " +
			"does not name is not named, since its words may be anything.", c.breachesTable()},
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

// aheadTable is, for each group, what became of the tasks written ahead.
func (c *counts) aheadTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Asked ahead", "Kept", "Handed out ready", "Taken by the card", "Let go",
	}, named: 2}
	for _, g := range c.groups() {
		counted := c.tasks[g]
		if counted.ahead+counted.kept+counted.ready+counted.byCard+counted.letGo == 0 {
			continue
		}
		t.add(g.version, g.host, number(counted.ahead), number(counted.kept),
			fmt.Sprintf("%d of %d", counted.ready, len(counted.attempts)), number(counted.byCard),
			number(counted.letGo))
	}
	return t
}

// letGoTable is, under each version, why the tasks written ahead were let go,
// the reason that let go the most first.
func (c *counts) letGoTable() *table {
	t := &table{columns: []string{"Instructions", "Reason", "Written", "Let go"}, named: 3}
	keys := slices.Collect(maps.Keys(c.letGo))
	slices.SortFunc(keys, func(a, b letGoFor) int {
		return cmp.Or(
			cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
			cmp.Compare(c.letGo[b], c.letGo[a]),
			cmp.Compare(a.reason, b.reason),
			cmp.Compare(strconv.FormatBool(a.written), strconv.FormatBool(b.written)),
		)
	})
	for _, key := range keys {
		written := "being written"
		if key.written {
			written = "kept"
		}
		t.add(key.version, key.reason, written, number(c.letGo[key]))
	}
	return t
}

// acceptedTable is, for each group, the tasks accepted, the attempts they took
// and how long they took to write.
func (c *counts) acceptedTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Accepted", "With a drawing", "At the first attempt", "Attempts, mean",
		"Seconds, median", "Seconds, 90th percentile", "Seconds, longest",
	}, named: 2}
	for _, g := range c.groups() {
		counted := c.tasks[g]
		if len(counted.attempts) == 0 {
			continue
		}
		drawn := notLogged
		switch {
		case counted.told == len(counted.attempts):
			drawn = number(counted.drawn)
		case counted.told > 0:
			drawn = fmt.Sprintf("%d of %d", counted.drawn, counted.told)
		}
		seconds := slices.Sorted(slices.Values(counted.seconds))
		t.add(g.version, g.host, number(len(counted.attempts)), drawn, number(countOf(counted.attempts, 1)),
			strconv.FormatFloat(mean(counted.attempts), 'f', 1, 64),
			number(atRank(seconds, 50)), number(atRank(seconds, 90)), number(atRank(seconds, 100)))
	}
	return t
}

// handInsTable is, for each group and format, how large its hand-ins were:
// the median of every part, and the whole's 90th percentile.
func (c *counts) handInsTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Format", "Hand-ins", "Mended", "Task, median", "Self-check, median",
		"Solver, median", "Core idea, median", "Whole, median", "Whole, 90th percentile",
	}, named: 3}
	keys := slices.Collect(maps.Keys(c.handIns))
	slices.SortFunc(keys, func(a, b handInOf) int {
		return cmp.Or(
			cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
			cmp.Compare(a.host, b.host),
			cmp.Compare(a.format, b.format),
		)
	})
	for _, key := range keys {
		counted := c.handIns[key]
		total := slices.Sorted(slices.Values(counted.total))
		t.add(key.version, key.host, key.format, number(counted.count), number(counted.mended),
			medianOf(counted.task), medianOf(counted.selfCheck), medianOf(counted.solver), medianOf(counted.coreIdea),
			medianOf(total), rankOf(total, 90))
	}
	return t
}

// medianOf is the middle of sizes, or not logged when there are none.
func medianOf(sizes []int64) string {
	return rankOf(slices.Sorted(slices.Values(sizes)), 50)
}

// rankOf is the size at a rank of sorted sizes, or not logged when there are
// none.
func rankOf(sorted []int64, rank int) string {
	if len(sorted) == 0 {
		return notLogged
	}
	return number(atRank(sorted, rank))
}

// mendsTable is, under each version, the fields read as they were meant, the
// field read so most often first.
func (c *counts) mendsTable() *table {
	t := &table{columns: []string{"Instructions", "Field", "Hand-ins"}, named: 2}
	keys := slices.Collect(maps.Keys(c.mends))
	slices.SortFunc(keys, func(a, b mendOf) int {
		return cmp.Or(
			cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
			cmp.Compare(c.mends[b], c.mends[a]),
			cmp.Compare(a.field, b.field),
		)
	})
	for _, key := range keys {
		t.add(key.version, key.field, number(c.mends[key]))
	}
	return t
}

// drawingsTable is, under each version and on each topic, the tasks accepted
// whose line says whether they came with a drawing, and how many of them did.
func (c *counts) drawingsTable() *table {
	t := &table{columns: []string{"Instructions", "Topic", "Accepted", "With a drawing"}, named: 2}
	keys := slices.Collect(maps.Keys(c.drawings))
	slices.SortFunc(keys, func(a, b drawnOn) int {
		return cmp.Or(
			cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
			cmp.Compare(a.topic, b.topic),
		)
	})
	for _, key := range keys {
		t.add(key.version, key.topic, number(c.drawings[key].accepted), number(c.drawings[key].drawn))
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
		"Without Drive, median", "Without Drive, 95th percentile",
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
		own := slices.Sorted(slices.Values(called.own))
		t.add(append(cells, number(atRank(took, 50)), number(atRank(took, 95)),
			number(atRank(own, 50)), number(atRank(own, 95)))...)
	}
	return t
}

// tracesAbout says how to read the traces, and what share of them was kept
// beside the share the telemetry was built to keep.
func (c *counts) tracesAbout() string {
	about := "A request's trace is kept or dropped as the request arrives. The spans of a kept trace are " +
		"delivered before the request ends, and the measurements with whichever request finds them due, so a " +
		"delivery that failed lost the spans of a kept trace, or the measurements. A delivery that ran out of " +
		"time is past the deadline it is given."
	t := &c.traces
	if requests := t.kept + t.dropped; requests > 0 {
		about += fmt.Sprintf(" %s of the %d requests that carried a trace kept it.",
			strconv.FormatFloat(float64(t.kept)/float64(requests), 'f', 3, 64), requests)
	}
	switch {
	case t.configured != nil:
		about += fmt.Sprintf(" The telemetry was built to keep %s.", strconv.FormatFloat(*t.configured, 'f', -1, 64))
	case t.built:
		about += " The newest build of the telemetry among the lines exports nothing."
	}
	if untold := t.failed[ofUntold]; untold > 0 {
		about += fmt.Sprintf(" Deliveries that failed for a request whose own line does not tell whether it kept its "+
			"trace — a line not read, or an id requests of both kinds came with: %d, past the deadline %d.",
			untold, t.late[ofUntold])
	}
	if t.broke > 0 {
		about += fmt.Sprintf(" Lines that say the telemetry failed on its own, beside a delivery: %d.", t.broke)
	}
	return about
}

// tracesTable is how many requests kept their trace and how many dropped it,
// and how many of their deliveries failed.
func (c *counts) tracesTable() *table {
	t := &table{columns: []string{"Trace", "Requests", "Deliveries failed", "Of them past the deadline"}, named: 1}
	traced := &c.traces
	if !traced.told() {
		return t
	}
	t.add("kept", number(traced.kept), number(traced.failed[ofKept]), number(traced.late[ofKept]))
	t.add("dropped", number(traced.dropped), number(traced.failed[ofDropped]), number(traced.late[ofDropped]))
	return t
}

// busyAbout says how to read the busiest minute, how many instances the lines
// of requests name, and how many requests the MCP endpoint was sent for each
// tool call.
func (c *counts) busyAbout() string {
	about := "The most a minute by the clock held, the minute a pace is counted over. An account's pace counts " +
		"every message it sends the MCP endpoint, and a tool call is the one a line names the account on; an " +
		"instance's counts every request it is sent, at the MCP endpoint and at the sign-in apart."
	b := &c.busy
	switch instances := b.instances(); {
	case instances == 1:
		about += " The lines of requests name one instance."
	case instances > 1:
		about += fmt.Sprintf(" The lines of requests name %d instances.", instances)
	case len(b.mcp)+len(b.others) > 0:
		about += " The lines of requests name no instance: one process of a run on this machine wrote them."
	}
	if b.toolCalls > 0 && b.mcpRequests > 0 {
		about += fmt.Sprintf(" The MCP endpoint was sent %s requests for each tool call.",
			strconv.FormatFloat(float64(b.mcpRequests)/float64(b.toolCalls), 'f', 2, 64))
	}
	return about
}

// busyTable is the most a minute held of each thing a pace counts.
func (c *counts) busyTable() *table {
	t := &table{columns: []string{"What", "Most in a minute"}, named: 1}
	b := &c.busy
	if b.toolCalls+len(b.mcp)+len(b.others) == 0 {
		return t
	}
	t.add("Tool calls of one account", number(mostOf(b.ofAccount)))
	t.add("Requests to the MCP endpoint on one instance", number(mostOf(b.mcp)))
	t.add("Other requests on one instance", number(mostOf(b.others)))
	t.add("Requests to the MCP endpoint on all instances", number(mostOf(acrossInstances(b.mcp))))
	return t
}

// breachesTable is every way the lines broke the rules of the log, by event,
// field and rule.
func (c *counts) breachesTable() *table {
	t := &table{columns: []string{"Event", "Field", "Rule", "Lines"}, named: 3}
	keys := slices.Collect(maps.Keys(c.breaches))
	slices.SortFunc(keys, compareBreaches)
	for _, key := range keys {
		t.add(key.event, key.field, key.rule, number(c.breaches[key]))
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
// a bar or a line break in a cell stays in the cell.
func writeRow(b *strings.Builder, cells []string) {
	written := make([]string, len(cells))
	for i, cell := range cells {
		written[i] = cmp.Or(inACell.Replace(cell), "(none)")
	}
	b.WriteString("| " + strings.Join(written, " | ") + " |\n")
}

// inACell keeps a text inside the cell of a table: a bar is written as a bar,
// not a border, and a line break as a space, not the end of the row.
var inACell = strings.NewReplacer("|", `\|`, "\r\n", " ", "\n", " ", "\r", " ")

// number is a whole number as the report writes it.
func number[T int | int64](n T) string { return strconv.FormatInt(int64(n), 10) }

// ordinal is a whole number as English counts with it: 1st, 2nd, 3rd, 4th,
// 11th, 101st.
func ordinal(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return strconv.Itoa(n) + suffix
}

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
