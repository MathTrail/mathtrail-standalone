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
	eventTaskKept       = "task_kept"
	eventTaskDropped    = "task_dropped"
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
	// told counts the tasks accepted whose line says whether they came with a
	// drawing, and drawn those of them that did.
	told, drawn int
	// ahead counts the requests opened for a task written ahead, kept the
	// tasks written ahead and kept, and letGo those let go, kept or still being
	// written, once the lesson moved away from them. ready counts the tasks
	// handed out that had been kept, and byCard the tasks a card took, kept
	// or waited for, while a release let a card take one into its own place.
	ahead, kept, letGo, ready, byCard int
}

// handInOf is whose hand-ins are measured together: a group, and the format
// they came in — the one asked for now, the one before with what it retired,
// or not told, in a line written before hand-ins were measured.
type handInOf struct {
	group
	format string
}

// The formats a hand-in is counted under.
const (
	formatNow    = "now"
	formatBefore = "before"
)

// handIns are how large the hand-ins of one kind were, part by part, in bytes,
// and how many of them had a field read as it was meant.
type handIns struct {
	task, selfCheck, solver, coreIdea, total []int64
	count, mended                            int
}

// mendOf is a field read as it was meant, under a version of the
// instructions.
type mendOf struct{ version, field string }

// letGoFor is why tasks written ahead were let go, under a version of the
// instructions, and whether they had been written and kept.
type letGoFor struct {
	version, reason string
	written         bool
}

// drawnOn is a topic as the tasks of one version of the instructions were
// written on it.
type drawnOn struct{ version, topic string }

// drawings is how many tasks of a topic were accepted with a line that says
// whether they came with a drawing, and how many of them did.
type drawings struct{ accepted, drawn int }

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
	// own is the milliseconds of each call less those of its calls to Drive.
	own []int64
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
	handIns  map[handInOf]*handIns
	mends    map[mendOf]int
	drawings map[drawnOn]*drawings
	refusals map[refusedBy]*refusals
	letGo    map[letGoFor]int
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
	// testing are the children a host of notChildren handed a task to, and
	// masteries the topics the other children were shown as mastered, as
	// their answers after followed them.
	testing   map[string]bool
	masteries map[group][]shownMastery
	// traces is what became of the traces and their deliveries, busy how busy
	// the busiest minute was, and breaches every way the lines broke the rules
	// of the log.
	traces   traces
	busy     busy
	breaches map[breach]int
	// children is what each day adds up to in the counts kept for years.
	children map[time.Time]*childrenOn
}

// tally adds the lines up. A task or an answer is counted for the chat host of
// the tool call it came in, whose line is the one line of a call that names
// the host.
func tally(in *input) *counts {
	lines := in.lines
	c := &counts{
		lines: len(lines), others: in.others, unreadable: in.unreadable,
		tasks: map[group]*tasks{}, handIns: map[handInOf]*handIns{}, mends: map[mendOf]int{},
		drawings: map[drawnOn]*drawings{}, refusals: map[refusedBy]*refusals{},
		letGo:  map[letGoFor]int{},
		limits: map[string]int{}, tools: map[toolOf]*calls{},
		promises: map[promisedIn]*cameTrue{}, keptUp: map[keptUpIn]*keptUp{}, answersLeftOut: map[string]int{},
	}
	hosts := hostsOf(lines)
	drive := driveTimes(lines)
	firstSeen := map[string]time.Time{}
	for i := range lines {
		l := &lines[i]
		c.within(l.Time)
		switch l.Message {
		case eventToolCall:
			c.toolCall(l, drive)
		case eventLimitHit:
			c.limits[l.Limit]++
		case eventTaskRequested, eventTaskSubmitted, eventTaskAccepted, eventTaskKept, eventTaskDropped:
			seenAt(firstSeen, l.InstructionsVersion, l.Time)
			g := group{version: l.InstructionsVersion, host: hostOf(hosts, l.call())}
			c.task(l, c.tasksOf(g))
			if l.Message == eventTaskSubmitted {
				c.handIn(l, g)
			}
		case eventAnswerRecorded:
			seenAt(firstSeen, l.InstructionsVersion, l.Time)
			c.answer(l, group{version: l.InstructionsVersion, host: hostOf(hosts, l.call())})
		case eventTaskSkipped:
			c.skipped++
		}
	}
	c.versions = inOrder(firstSeen)
	testedUsers, testedLearners := testChildren(lines)
	c.testing = testedUsers
	c.masteries = shownMasteries(lines, hosts, testedUsers)
	c.traces, c.busy, c.breaches = tracesOf(lines), busyOf(lines), in.breaches
	c.children = childrenByDay(lines, testedLearners)
	return c
}

// handIn counts how large a hand-in was, and what in it was read as meant, for
// its group and the format it came in.
func (c *counts) handIn(l *line, g group) {
	key := handInOf{group: g, format: formatNow}
	switch {
	case l.TotalBytes == nil:
		key.format = notLogged
	case len(l.Retired) > 0:
		key.format = formatBefore
	}
	counted, found := c.handIns[key]
	if !found {
		counted = &handIns{}
		c.handIns[key] = counted
	}
	counted.count++
	if len(l.Mended) > 0 {
		counted.mended++
	}
	for _, field := range l.Mended {
		c.mends[mendOf{version: g.version, field: field}]++
	}
	if l.TotalBytes == nil {
		return
	}
	for _, part := range []struct {
		into *[]int64
		size *whole
	}{
		{&counted.task, l.TaskBytes}, {&counted.selfCheck, l.SelfCheckBytes}, {&counted.solver, l.SolverBytes},
		{&counted.coreIdea, l.CoreIdeaBytes}, {&counted.total, l.TotalBytes},
	} {
		if part.size != nil {
			*part.into = append(*part.into, int64(*part.size))
		}
	}
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

// toolCall counts a tool call: how it ended, how long it took, and how long
// less the calls to Drive made in its request. Calls to Drive made at once
// could add up to more than the call took, and leave it no time of its own.
func (c *counts) toolCall(l *line, drive map[string]int64) {
	key := toolOf{host: l.Client, tool: l.Tool}
	called, found := c.tools[key]
	if !found {
		called = &calls{outcomes: map[string]int{}}
		c.tools[key] = called
	}
	called.outcomes[l.Outcome]++
	called.milliseconds = append(called.milliseconds, int64(l.DurationMS))
	called.own = append(called.own, max(int64(l.DurationMS)-drive[l.request()], 0))
}

// task counts a line about a task into the tasks of its group.
func (c *counts) task(l *line, counted *tasks) {
	switch l.Message {
	case eventTaskRequested:
		if l.AlreadyOpen {
			return
		}
		counted.asked++
		if l.Ahead {
			counted.ahead++
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
		c.drawing(l, counted)
		if l.Ready {
			counted.ready++
		}
		if l.ByCard {
			counted.byCard++
		}
	case eventTaskKept:
		counted.kept++
	case eventTaskDropped:
		counted.letGo++
		c.letGo[letGoFor{version: l.InstructionsVersion, reason: l.Reason, written: l.Written}]++
	}
}

// drawing counts whether an accepted task came with a drawing, for its group
// and for its topic. A line that does not say is counted neither way, so that
// the tasks accepted before the service said so read as unknown rather than
// as drawn by none.
func (c *counts) drawing(l *line, counted *tasks) {
	if l.Drawing == nil {
		return
	}
	key := drawnOn{version: l.InstructionsVersion, topic: l.Topic}
	onTopic, found := c.drawings[key]
	if !found {
		onTopic = &drawings{}
		c.drawings[key] = onTopic
	}
	counted.told++
	onTopic.accepted++
	if *l.Drawing {
		counted.drawn++
		onTopic.drawn++
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
