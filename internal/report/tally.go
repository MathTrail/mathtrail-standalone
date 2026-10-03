package report

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// The events the report reads, as their lines are named.
const (
	eventToolCall       = "tool_call"
	eventTaskRequested  = "task_requested"
	eventTaskSubmitted  = "task_submitted"
	eventTaskAccepted   = "task_accepted"
	eventTaskSkipped    = "task_skipped"
	eventAnswerRecorded = "answer_recorded"
	eventLimitHit       = "limit_hit"
)

// outcomeRejected is an attempt at a task that its checks refused.
const outcomeRejected = "rejected"

// callNotRead stands for the host of a task whose tool call's own line is not
// among the lines read. It is no word the service writes, so that it is never
// taken for a client that gave no name, which the service calls unknown.
const callNotRead = "(call not read)"

// group is whom the tasks are counted for: the version of the instructions a
// task was written to, and the chat host whose model wrote it.
type group struct{ version, host string }

// tasks is what became of the tasks of one group.
type tasks struct {
	// asked counts the requests opened. A request handed back while it is
	// still open is the same request, and is not counted again.
	asked int
	// handedIn counts the attempts judged, and refused those of them the
	// checks refused.
	handedIn, refused int
	// outOfAttempts counts the requests whose last attempt was refused.
	outOfAttempts int
	// attempts and seconds are, for every task accepted, the attempts it took
	// and the seconds from its request to it: the time the chat's model took
	// to write it.
	attempts []int
	seconds  []int64
}

// refusedBy is a check that refused attempts, under a version of the
// instructions.
type refusedBy struct{ version, check string }

// refusals is how many attempts a check refused: those counted by it, the
// first check their refusal names, and all that failed it.
type refusals struct{ counted, failed int }

// toolOf is a tool as one chat host called it.
type toolOf struct{ host, tool string }

// calls is how the calls of a tool ended, and how long each took.
type calls struct {
	outcomes     map[string]int
	milliseconds []int64
}

// counts is everything the lines add up to.
type counts struct {
	// lines, others and unreadable are how many lines of the service's were
	// read, how many lines were another's, and how many of the service's did
	// not read as the report expects.
	lines, others, unreadable int
	from, to                  time.Time
	// versions are the versions of the instructions, in the order they first
	// appear: a newer version comes after an older one.
	versions []string
	tasks    map[group]*tasks
	refusals map[refusedBy]*refusals
	limits   map[string]int
	tools    map[toolOf]*calls
	// promises and keptUp are the answers weighed against the chance their
	// tasks were handed out at, by that chance and by how many answers the
	// child had given; answersLeftOut counts the answers neither takes, by
	// why, and skipped the tasks left without an answer.
	promises       map[promisedIn]*cameTrue
	keptUp         map[keptUpIn]*keptUp
	answersLeftOut map[string]int
	skipped        int
}

// tally adds the lines up. A task or an answer is counted for the chat host of
// the tool call it came in, whose line is the one line of a call that names
// the host.
func tally(in *input) *counts {
	lines := in.lines
	c := &counts{
		lines: len(lines), others: in.others, unreadable: in.unreadable,
		tasks: map[group]*tasks{}, refusals: map[refusedBy]*refusals{},
		limits: map[string]int{}, tools: map[toolOf]*calls{},
		promises: map[promisedIn]*cameTrue{}, keptUp: map[keptUpIn]*keptUp{}, answersLeftOut: map[string]int{},
	}
	hosts := hostsOf(lines)
	firstSeen := map[string]time.Time{}
	for i := range lines {
		l := &lines[i]
		c.within(l.Time)
		switch l.Message {
		case eventToolCall:
			c.toolCall(l)
		case eventLimitHit:
			c.limits[l.Limit]++
		case eventTaskRequested, eventTaskSubmitted, eventTaskAccepted:
			seenAt(firstSeen, l.InstructionsVersion, l.Time)
			c.task(l, c.tasksOf(group{version: l.InstructionsVersion, host: hostOf(hosts, l.call())}))
		case eventAnswerRecorded:
			seenAt(firstSeen, l.InstructionsVersion, l.Time)
			c.answer(l, group{version: l.InstructionsVersion, host: hostOf(hosts, l.call())})
		case eventTaskSkipped:
			c.skipped++
		}
	}
	c.versions = inOrder(firstSeen)
	return c
}

// seenAt keeps the earliest moment a version is seen at. A line with no time
// says only that the version was seen.
func seenAt(firstSeen map[string]time.Time, version string, at time.Time) {
	first, found := firstSeen[version]
	if !found || (!at.IsZero() && (first.IsZero() || at.Before(first))) {
		firstSeen[version] = at
	}
}

// within widens the time the lines span to take in a moment. A line with no
// time leaves it as it is.
func (c *counts) within(at time.Time) {
	if at.IsZero() {
		return
	}
	if c.from.IsZero() || at.Before(c.from) {
		c.from = at
	}
	if at.After(c.to) {
		c.to = at
	}
}

// toolCall counts a tool call: how it ended and how long it took.
func (c *counts) toolCall(l *line) {
	key := toolOf{host: l.Client, tool: l.Tool}
	called, found := c.tools[key]
	if !found {
		called = &calls{outcomes: map[string]int{}}
		c.tools[key] = called
	}
	called.outcomes[l.Outcome]++
	called.milliseconds = append(called.milliseconds, int64(l.DurationMS))
}

// task counts a line about a task into the tasks of its group.
func (c *counts) task(l *line, counted *tasks) {
	switch l.Message {
	case eventTaskRequested:
		if !l.AlreadyOpen {
			counted.asked++
		}
	case eventTaskSubmitted:
		counted.handedIn++
		if l.Outcome != outcomeRejected {
			return
		}
		counted.refused++
		if l.Attempt >= whole(profile.MaxAttempts) {
			counted.outOfAttempts++
		}
		c.refusedAttempt(l)
	case eventTaskAccepted:
		counted.attempts = append(counted.attempts, int(l.Attempts))
		counted.seconds = append(counted.seconds, int64(l.SecondsSinceRequest))
	}
}

// refusedAttempt counts a refused attempt against the check it is counted by,
// and against every check it failed.
func (c *counts) refusedAttempt(l *line) {
	checks := append([]string{l.Primary}, l.Failed...)
	slices.Sort(checks)
	for _, check := range slices.Compact(checks) {
		if check == "" {
			continue
		}
		key := refusedBy{version: l.InstructionsVersion, check: check}
		refused, found := c.refusals[key]
		if !found {
			refused = &refusals{}
			c.refusals[key] = refused
		}
		if check == l.Primary {
			refused.counted++
		}
		if slices.Contains(l.Failed, check) {
			refused.failed++
		}
	}
}

// tasksOf is the tasks of a group, counted from nothing the first time the
// group is seen.
func (c *counts) tasksOf(g group) *tasks {
	counted, found := c.tasks[g]
	if !found {
		counted = &tasks{}
		c.tasks[g] = counted
	}
	return counted
}

// hostsOf is the chat host of every tool call among the lines.
func hostsOf(lines []line) map[string]string {
	hosts := map[string]string{}
	for i := range lines {
		if call := lines[i].call(); lines[i].Message == eventToolCall && call != "" {
			hosts[call] = lines[i].Client
		}
	}
	return hosts
}

// hostOf is the chat host of a call, or callNotRead when the call's own line
// is not among the lines read.
func hostOf(hosts map[string]string, call string) string {
	if host := hosts[call]; host != "" {
		return host
	}
	return callNotRead
}

// inOrder is the versions by the moment each first appears, and by name where
// two appear together or a line has no time.
func inOrder(firstSeen map[string]time.Time) []string {
	versions := make([]string, 0, len(firstSeen))
	for version := range firstSeen {
		versions = append(versions, version)
	}
	slices.SortFunc(versions, func(a, b string) int {
		return cmp.Or(firstSeen[a].Compare(firstSeen[b]), cmp.Compare(a, b))
	})
	return versions
}

// atRank is the value at a rank, given in hundredths, among values sorted from
// the smallest, by the nearest-rank method: the smallest of the values with at
// least that share of them at or below it. It is always one of the values, of
// which there is at least one.
func atRank[T cmp.Ordered](sorted []T, rank int) T {
	at := int(math.Ceil(float64(rank)/100*float64(len(sorted)))) - 1
	return sorted[max(at, 0)]
}

// mean is the average of values, of which there is at least one.
func mean(values []int) float64 {
	sum := 0
	for _, value := range values {
		sum += value
	}
	return float64(sum) / float64(len(values))
}
