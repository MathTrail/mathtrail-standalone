// Package report turns the calls of a run into the numbers the run is judged
// by: how each kind of answer was spread and how long it took, what one unit
// of the run's work cost under the platform's billing and how much of the free
// tier that leaves, and the facts that fail a run outright.
package report

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// Call is one call of a run, as the report counts it.
type Call struct {
	Tool   string
	Kind   session.Kind
	Status int
	// Due is when the call was meant to go out. A call that went out late
	// counts its lateness in its time, so that a service slow to take calls
	// cannot hide behind the calls it made wait.
	Due time.Time
	// Started and Ended are when the call went out and when it came back.
	Started, Ended time.Time
}

// CallOf is an answer as the report counts it, due at the moment given.
func CallOf(answer *session.Answer, due time.Time) Call {
	return Call{
		Tool: answer.Tool, Kind: answer.Kind, Status: answer.Status,
		Due: due, Started: answer.Started, Ended: answer.Ended,
	}
}

// Took is how long the call took from the moment it was due.
func (c *Call) Took() time.Duration { return c.Ended.Sub(c.Due) }

// Run is one run of a scenario: what it did and what it came to.
type Run struct {
	Scenario     string
	Began, Ended time.Time
	Calls        []Call
	// Unsent is how many calls were due and could not be sent, because the
	// runner already had as many in flight as it may.
	Unsent int
	// Requests is how many requests the run sent the service, the ones the
	// protocol library sent on its own among them.
	Requests int
	// Reconnects is how many sessions had to be opened again.
	Reconnects int
	// Unit is what one unit of the run's work is, such as "accepted task", and
	// Units is how many the run did.
	Unit  string
	Units int
	// Broken is what the scenario expected and did not see.
	Broken []string
	// Stopped is a run asked to stop before its end: its numbers are of the
	// part that ran.
	Stopped bool
}

// Instance is what the platform gives an instance, and bills for.
type Instance struct {
	VCPU float64
	GiB  float64
}

// The free tier of the platform, for a month.
const (
	FreeVCPUSeconds = 180_000
	FreeGiBSeconds  = 360_000
	FreeRequests    = 2_000_000
)

// billingTick is the step the platform bills an instance's time in.
const billingTick = 100 * time.Millisecond

// Billed is the time the platform bills an instance for over the calls given:
// every moment at least one of them was under way, each unbroken stretch of it
// rounded up to the step the platform bills in. An instance that is serving
// no request is billed nothing.
func Billed(calls []Call) time.Duration {
	spans := make([]Call, len(calls))
	copy(spans, calls)
	slices.SortFunc(spans, func(a, b Call) int { return a.Started.Compare(b.Started) })

	var billed time.Duration
	var from, to time.Time
	for i := range spans {
		switch {
		case i == 0:
			from, to = spans[i].Started, spans[i].Ended
		case spans[i].Started.After(to):
			billed += roundUp(to.Sub(from))
			from, to = spans[i].Started, spans[i].Ended
		case spans[i].Ended.After(to):
			to = spans[i].Ended
		}
	}
	if len(spans) > 0 {
		billed += roundUp(to.Sub(from))
	}
	return billed
}

func roundUp(stretch time.Duration) time.Duration {
	return time.Duration(math.Ceil(float64(stretch)/float64(billingTick))) * billingTick
}

// Cost is what one unit of a run's work cost under the platform's billing.
type Cost struct {
	VCPUSeconds float64
	GiBSeconds  float64
	Requests    float64
}

// Cost is what one unit of the run's work cost on the instance given, and
// false when the run did no work to share it out over.
func (r *Run) Cost(instance Instance) (Cost, bool) {
	if r.Units == 0 {
		return Cost{}, false
	}
	seconds := Billed(r.Calls).Seconds() / float64(r.Units)
	return Cost{
		VCPUSeconds: seconds * instance.VCPU,
		GiBSeconds:  seconds * instance.GiB,
		Requests:    float64(r.Requests) / float64(r.Units),
	}, true
}

// Fit is how many units of work a month of the free tier holds at the cost
// given, and which of its three allowances runs out first.
func Fit(cost Cost) (units float64, scarcest string) {
	units = math.Inf(1)
	for _, allowance := range []struct {
		name       string
		free, cost float64
	}{
		{"vCPU-seconds", FreeVCPUSeconds, cost.VCPUSeconds},
		{"GiB-seconds", FreeGiBSeconds, cost.GiBSeconds},
		{"requests", FreeRequests, cost.Requests},
	} {
		if allowance.cost <= 0 {
			continue
		}
		if fits := allowance.free / allowance.cost; fits < units {
			units, scarcest = fits, allowance.name
		}
	}
	return units, scarcest
}

// Hard is a fact that fails a run outright, and how many times it happened.
type Hard struct {
	What  string
	Count int
}

// Verdict is the facts that fail the run, most frequent first: every answer
// the service should never give, and everything the scenario expected and did
// not see.
func (r *Run) Verdict() []Hard {
	counts := map[string]int{}
	for i := range r.Calls {
		if hard(r.Calls[i].Kind) {
			counts[r.Calls[i].Kind.String()]++
		}
	}
	hards := make([]Hard, 0, len(counts)+len(r.Broken))
	for what, count := range counts {
		hards = append(hards, Hard{What: what, Count: count})
	}
	slices.SortFunc(hards, func(a, b Hard) int {
		return cmp.Or(cmp.Compare(b.Count, a.Count), strings.Compare(a.What, b.What))
	})
	for _, broken := range r.Broken {
		hards = append(hards, Hard{What: broken, Count: 1})
	}
	return hards
}

// hard says whether an answer is one the service should never give, whatever
// load it is under. What it is meant to say under load is named: its answers,
// its three refusals — a message a pace held back among them — and a task
// turned away for want of a slot; a call the run itself abandoned says nothing
// about the service. Everything else is hard — a failure of its own, a request
// it did not answer, an error of the protocol, an answer that belongs to
// another call, a protocol it did not speak, and a kind of answer nobody has
// seen before.
func hard(kind session.Kind) bool {
	switch kind.Class {
	case session.OK, session.Rejected, session.Stale, session.Limited, session.Stopped:
		return false
	case session.Failure:
		return kind != session.Busy
	}
	return true
}

// Write writes the runs as Markdown, and is the facts that fail any of them.
func Write(w io.Writer, runs []Run, instance Instance) ([]Hard, error) {
	var all []Hard
	var page strings.Builder
	for i := range runs {
		hards := runs[i].Verdict()
		all = append(all, hards...)
		runs[i].write(&page, hards, instance)
	}
	if _, err := io.WriteString(w, page.String()); err != nil {
		return all, fmt.Errorf("report: write: %w", err)
	}
	return all, nil
}

func (r *Run) write(page *strings.Builder, hards []Hard, instance Instance) {
	fmt.Fprintf(page, "# Load: %s\n\n", r.Scenario)
	switch {
	case r.Stopped:
		// A run cut short has no verdict: what it saw of the service is told
		// all the same.
		page.WriteString("**Stopped** before its end: the numbers are of the part that ran, and there is no verdict.\n\n")
	case len(hards) == 0:
		page.WriteString("**Verdict:** clean — nothing the service should never do.\n\n")
	default:
		page.WriteString("**Verdict:** failed —\n\n")
	}
	if len(hards) > 0 {
		for _, h := range hards {
			fmt.Fprintf(page, "- %s × %d\n", cell(h.What), h.Count)
		}
		page.WriteString("\n")
	}

	fmt.Fprintf(page, "Ran %s for %s: %d calls sent", r.Began.UTC().Format(time.DateTime+" UTC"),
		r.Ended.Sub(r.Began).Round(time.Second), len(r.Calls))
	if r.Unsent > 0 {
		fmt.Fprintf(page, ", %d due and not sent", r.Unsent)
	}
	fmt.Fprintf(page, ", %d requests, %d sessions opened again.\n\n", r.Requests, r.Reconnects)

	r.writeKinds(page)
	r.writeCost(page, instance)
}

// cell is a text as a cell of a Markdown table, or an item of a list, can hold
// it: on one line, and with no bar to split the cell.
func cell(text string) string {
	return strings.NewReplacer("|", `\|`, "\r", " ", "\n", " ").Replace(text)
}

// writeKinds writes a line for every kind of answer: how many, what share, and
// how long they took from the moment each was due.
func (r *Run) writeKinds(page *strings.Builder) {
	page.WriteString("| Kind | Count | Share | p50 | p95 | p99 | Max |\n|---|---:|---:|---:|---:|---:|---:|\n")
	for _, kind := range Kinds(r.Calls) {
		metrics := latencies(r.Calls, kind)
		fmt.Fprintf(page, "| %s | %d | %.1f%% | %s | %s | %s | %s |\n",
			cell(kind.String()), metrics.Requests, 100*float64(metrics.Requests)/float64(len(r.Calls)),
			rounded(metrics.Latencies.P50), rounded(metrics.Latencies.P95),
			rounded(metrics.Latencies.P99), rounded(metrics.Latencies.Max))
	}
	page.WriteString("\n")
}

// writeCost writes what one unit of the run's work cost as the platform bills
// it, and how many a month the free tier holds.
func (r *Run) writeCost(page *strings.Builder, instance Instance) {
	cost, done := r.Cost(instance)
	if !done {
		fmt.Fprintf(page, "No %s was done, so there is no cost of one to tell.\n\n", r.Unit)
		return
	}
	fmt.Fprintf(page, "Billed for %s — every moment at least one call was under way, in steps of %s — "+
		"on an instance of %g vCPU and %g GiB, over %d × %s:\n\n",
		Billed(r.Calls), billingTick, instance.VCPU, instance.GiB, r.Units, r.Unit)
	fmt.Fprintf(page, "| One %s | Costs | A month of the free tier holds |\n|---|---:|---:|\n", r.Unit)
	for _, line := range []struct {
		name       string
		cost, free float64
	}{
		{"vCPU-seconds", cost.VCPUSeconds, FreeVCPUSeconds},
		{"GiB-seconds", cost.GiBSeconds, FreeGiBSeconds},
		{"requests", cost.Requests, FreeRequests},
	} {
		fmt.Fprintf(page, "| %s | %.3f | %.0f |\n", line.name, line.cost, line.free/line.cost)
	}
	units, scarcest := Fit(cost)
	fmt.Fprintf(page, "\nThe free tier holds %.0f × %s a month; %s run out first. "+
		"The requests are the tool's own: a chat host sends more for every call it makes.\n\n", units, r.Unit, scarcest)
}

// Kinds are the kinds of answer among the calls: answers as asked first, then
// every other in the order of its name.
func Kinds(calls []Call) []session.Kind {
	seen := map[string]session.Kind{}
	for i := range calls {
		seen[calls[i].Kind.String()] = calls[i].Kind
	}
	names := slices.Collect(maps.Keys(seen))
	sort.Slice(names, func(i, j int) bool {
		if (names[i] == session.Answered.String()) != (names[j] == session.Answered.String()) {
			return names[i] == session.Answered.String()
		}
		return names[i] < names[j]
	})
	kinds := make([]session.Kind, len(names))
	for i, name := range names {
		kinds[i] = seen[name]
	}
	return kinds
}

// latencies are the times the calls of one kind took, as the load library
// counts them.
func latencies(calls []Call, kind session.Kind) *vegeta.Metrics {
	var metrics vegeta.Metrics
	for i := range calls {
		if calls[i].Kind != kind {
			continue
		}
		metrics.Add(&vegeta.Result{
			Code:      uint16(max(calls[i].Status, 0)), //nolint:gosec // a status is under 1000
			Timestamp: calls[i].Due,
			Latency:   calls[i].Took(),
		})
	}
	metrics.Close()
	return &metrics
}

func rounded(d time.Duration) time.Duration { return d.Round(time.Millisecond) }
