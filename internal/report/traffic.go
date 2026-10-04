package report

import (
	"strings"
	"time"
)

// The events the report reads to tell how the requests went, the route of the
// MCP endpoint, and the words a delivery that ran out of time fails with.
const (
	eventHTTPRequest      = "http_request"
	eventDriveCall        = "drive_call"
	eventTelemetryBuilt   = "telemetry built"
	eventFlushFailed      = "telemetry_flush_failed"
	eventTelemetryFailed  = "telemetry_failed"
	eventFlushPanicked    = "telemetry_flush_panicked"
	mcpRoute              = "/mcp"
	deadlineExceededWords = "deadline exceeded"
)

// driveTimes is how long the calls to Drive of each request took together, in
// milliseconds, by the request they were made in. A call whose line names no
// request is tied to no call.
func driveTimes(lines []line) map[string]int64 {
	spent := map[string]int64{}
	for i := range lines {
		if request := lines[i].request(); lines[i].Message == eventDriveCall && request != "" {
			spent[request] += int64(lines[i].DurationMS)
		}
	}
	return spent
}

// traces is what became of the traces of the requests, and of the deliveries
// that carried them.
type traces struct {
	// built says whether a build of the telemetry is among the lines, builtAt
	// when the newest was built, and configured the share it keeps — none when
	// it exports nothing.
	built      bool
	builtAt    time.Time
	configured *float64
	// kept and dropped count the requests whose trace was kept and those
	// whose trace was dropped.
	kept, dropped int
	// failed counts the deliveries that failed, by whose they were: a request
	// whose trace was kept, one whose trace was dropped, or one its own line
	// does not tell — a line not read, or an id two requests of both kinds
	// came with. late counts those of them that ran out of time the same way.
	failed, late map[string]int
	// broke counts the other failures of the telemetry: in its setup, in a
	// delivery of its own, or a delivery that panicked.
	broke int
}

// Whose a delivery was, as traces counts failed ones.
const (
	ofKept    = "kept"
	ofDropped = "dropped"
	ofUntold  = "untold"
)

// decisions is whether each request kept its trace, by the id a delivery that
// failed names it by, and the ids requests of both kinds came with, which a
// client that sends its own may send twice.
type decisions struct {
	kept  map[string]bool
	mixed map[string]bool
}

// tracesOf is what became of the traces of the requests the lines name.
func tracesOf(lines []line) traces {
	t := traces{failed: map[string]int{}, late: map[string]int{}}
	decided := decisions{kept: map[string]bool{}, mixed: map[string]bool{}}
	for i := range lines {
		t.read(&lines[i], &decided)
	}
	for i := range lines {
		if lines[i].Message == eventFlushFailed {
			t.failedDelivery(&lines[i], &decided)
		}
	}
	return t
}

// read counts a line that says how a request's trace went, how the telemetry
// was built, or that it failed on its own, and keeps whether each request kept
// its trace. Of several builds, the newest is the one in force.
func (t *traces) read(l *line, decided *decisions) {
	switch {
	case l.Message == eventHTTPRequest && l.Sampled != nil:
		if before, seen := decided.kept[l.RequestID]; seen && before != *l.Sampled {
			decided.mixed[l.RequestID] = true
		}
		decided.kept[l.RequestID] = *l.Sampled
		if *l.Sampled {
			t.kept++
		} else {
			t.dropped++
		}
	case l.Message == eventTelemetryBuilt:
		if !t.built || l.Time.After(t.builtAt) {
			t.built, t.builtAt, t.configured = true, l.Time, nil
			if l.Export {
				t.configured = l.SampleRatio
			}
		}
	case l.Message == eventTelemetryFailed || l.Message == eventFlushPanicked:
		t.broke++
	}
}

// failedDelivery counts a delivery that failed for the request it belongs to,
// and as late when it ran out of time.
func (t *traces) failedDelivery(l *line, decided *decisions) {
	whose := ofUntold
	if wasKept, read := decided.kept[l.RequestID]; read && !decided.mixed[l.RequestID] {
		whose = ofDropped
		if wasKept {
			whose = ofKept
		}
	}
	t.failed[whose]++
	if strings.Contains(l.Error, deadlineExceededWords) {
		t.late[whose]++
	}
}

// told says whether the lines say anything of the traces at all.
func (t *traces) told() bool {
	return t.kept+t.dropped+t.broke > 0 || len(t.failed) > 0
}

// busy is how busy the busiest minute by the clock was: the minute a pace is
// counted over.
type busy struct {
	// ofAccount counts the tool calls of each account in each minute.
	ofAccount map[accountMinute]int
	// mcp and others count the requests each instance was sent in each
	// minute, at the MCP endpoint and anywhere else.
	mcp, others map[instanceMinute]int
	// toolCalls and mcpRequests are all the tool calls and all the requests
	// to the MCP endpoint, which say how many requests a call comes with.
	toolCalls, mcpRequests int
}

// accountMinute is one account in one minute.
type accountMinute struct {
	account string
	minute  time.Time
}

// instanceMinute is one instance in one minute. A line no reader named an
// instance for is the one process of a run on this machine.
type instanceMinute struct {
	instance string
	minute   time.Time
}

// busyOf is how busy each minute of the lines was.
func busyOf(lines []line) busy {
	b := busy{ofAccount: map[accountMinute]int{}, mcp: map[instanceMinute]int{}, others: map[instanceMinute]int{}}
	for i := range lines {
		switch l := &lines[i]; l.Message {
		case eventToolCall:
			b.toolCall(l)
		case eventHTTPRequest:
			b.request(l)
		}
	}
	return b
}

// toolCall counts a tool call, and into its account's minute when the line
// names the account and the time.
func (b *busy) toolCall(l *line) {
	b.toolCalls++
	if l.User != "" && !l.Time.IsZero() {
		b.ofAccount[accountMinute{account: l.User, minute: minuteOf(arrived(l))}]++
	}
}

// request counts a request, at the MCP endpoint or elsewhere, and into its
// instance's minute when the line names the time.
func (b *busy) request(l *line) {
	counted := b.others
	if l.Route == mcpRoute {
		b.mcpRequests++
		counted = b.mcp
	}
	if !l.Time.IsZero() {
		counted[instanceMinute{instance: l.Instance, minute: minuteOf(arrived(l))}]++
	}
}

// arrived is when the call or the request a line is about arrived, which is
// when a pace counts it: the line is written as it ends, so its time less how
// long it took.
func arrived(l *line) time.Time {
	switch l.Message {
	case eventToolCall:
		return l.Time.Add(-time.Duration(l.DurationMS) * time.Millisecond)
	case eventHTTPRequest:
		return l.Time.Add(-time.Duration(l.Duration * float64(time.Second)))
	}
	return l.Time
}

// minuteOf is the minute by the clock a moment falls in.
func minuteOf(at time.Time) time.Time { return at.UTC().Truncate(time.Minute) }

// instances is how many instances the lines of requests name. A line no reader
// named an instance for is the one process of a run on this machine, and is
// none of them.
func (b *busy) instances() int {
	seen := map[string]bool{}
	for _, counted := range []map[instanceMinute]int{b.mcp, b.others} {
		for at := range counted {
			if at.instance != "" {
				seen[at.instance] = true
			}
		}
	}
	return len(seen)
}

// mostOf is the most any one key counted.
func mostOf[K comparable](counted map[K]int) int {
	most := 0
	for _, n := range counted {
		most = max(most, n)
	}
	return most
}

// acrossInstances is the requests of all instances together, by minute.
func acrossInstances(counted map[instanceMinute]int) map[time.Time]int {
	together := map[time.Time]int{}
	for at, n := range counted {
		together[at.minute] += n
	}
	return together
}
