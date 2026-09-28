package report_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// start is the moment the calls of these cases are counted from.
var start = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

// between is a call that went out and came back at the milliseconds given,
// counted from start, and was due when it went out.
func between(from, to int, kind session.Kind) report.Call {
	return report.Call{
		Tool: "get_profile", Kind: kind, Status: 200,
		Due:     start.Add(time.Duration(from) * time.Millisecond),
		Started: start.Add(time.Duration(from) * time.Millisecond),
		Ended:   start.Add(time.Duration(to) * time.Millisecond),
	}
}

// The platform bills an instance for the time it is serving at least one
// request, in steps of a tenth of a second: calls that overlap are billed
// once, and each unbroken stretch is rounded up.
func TestBilledTimeIsTheUnionOfTheCallsRoundedUp(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		calls []report.Call
		want  time.Duration
	}{
		{"no call", nil, 0},
		{"one call", []report.Call{between(0, 250, session.Answered)}, 300 * time.Millisecond},
		{"one call of a whole step", []report.Call{between(0, 200, session.Answered)}, 200 * time.Millisecond},
		{"two calls apart", []report.Call{between(0, 50, session.Answered), between(1000, 1050, session.Answered)}, 200 * time.Millisecond},
		{"two calls overlapping", []report.Call{between(0, 150, session.Answered), between(100, 250, session.Answered)}, 300 * time.Millisecond},
		{"a call inside another", []report.Call{between(0, 350, session.Answered), between(100, 150, session.Answered)}, 400 * time.Millisecond},
		{"calls in no order", []report.Call{between(1000, 1050, session.Answered), between(0, 50, session.Answered)}, 200 * time.Millisecond},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := report.Billed(test.calls); got != test.want {
				t.Errorf("Billed() = %v, want %v", got, test.want)
			}
		})
	}
}

// Whatever the calls, the billed time is a whole number of steps, no shorter
// than the longest call, and no longer than every call billed on its own.
func TestBilledTimeKeepsToItsBounds(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(gopter.DefaultTestParameters())
	properties.Property("billed time is steps, at least the longest call, at most the calls apart", prop.ForAll(
		func(froms, lengths []int) bool {
			calls := make([]report.Call, min(len(froms), len(lengths)))
			var longest, apart time.Duration
			for i := range calls {
				calls[i] = between(froms[i], froms[i]+lengths[i], session.Answered)
				took := calls[i].Ended.Sub(calls[i].Started)
				longest = max(longest, took)
				apart += time.Duration(math.Ceil(float64(took)/float64(100*time.Millisecond))) * 100 * time.Millisecond
			}
			billed := report.Billed(calls)
			return billed%(100*time.Millisecond) == 0 && billed >= longest && billed <= apart
		},
		gen.SliceOf(gen.IntRange(0, 10_000)),
		gen.SliceOf(gen.IntRange(0, 3_000)),
	))
	properties.TestingRun(t)
}

// The free tier holds as many units of work a month as its scarcest
// allowance lets through.
func TestTheFreeTierIsBoundByItsScarcestAllowance(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		cost      report.Cost
		wantUnits float64
		wantFirst string
	}{
		{"the processor", report.Cost{VCPUSeconds: 1, GiBSeconds: 0.5, Requests: 1}, 180_000, "vCPU-seconds"},
		{"the memory", report.Cost{VCPUSeconds: 0.1, GiBSeconds: 4, Requests: 1}, 90_000, "GiB-seconds"},
		{"the requests", report.Cost{VCPUSeconds: 0.1, GiBSeconds: 0.05, Requests: 40}, 50_000, "requests"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			units, first := report.Fit(test.cost)
			if math.Abs(units-test.wantUnits) > 1e-6 || first != test.wantFirst {
				t.Errorf("Fit(%+v) = %v, %q, want %v, %q", test.cost, units, first, test.wantUnits, test.wantFirst)
			}
		})
	}
}

// What a unit of work cost is the billed time of the run, on the instance
// given, and the requests of the run, shared out over the units it did.
func TestAUnitCostsItsShareOfTheBilledTime(t *testing.T) {
	t.Parallel()

	run := report.Run{
		Calls:    []report.Call{between(0, 400, session.Answered), between(1000, 1400, session.Answered)},
		Requests: 6,
		Unit:     "accepted task",
		Units:    2,
	}
	cost, done := run.Cost(report.Instance{VCPU: 1, GiB: 0.5})
	if !done {
		t.Fatal("Cost() says no work was done, want the cost of one of two units")
	}
	if want := (report.Cost{VCPUSeconds: 0.4, GiBSeconds: 0.2, Requests: 3}); math.Abs(cost.VCPUSeconds-want.VCPUSeconds) > 1e-9 ||
		math.Abs(cost.GiBSeconds-want.GiBSeconds) > 1e-9 || cost.Requests != want.Requests {
		t.Errorf("Cost() = %+v, want %+v", cost, want)
	}
	if _, done := (&report.Run{Units: 0}).Cost(report.Instance{VCPU: 1, GiB: 0.5}); done {
		t.Error("Cost() of a run that did no work says it did, want no cost")
	}
}

// Every kind of answer keeps its own times: a slow refusal does not make the
// answers look slow, nor the other way round.
func TestPercentilesAreKeptKindByKind(t *testing.T) {
	t.Parallel()

	busy := session.Busy
	run := report.Run{Scenario: "saturation", Unit: "examined hand-in", Calls: []report.Call{
		between(0, 20, session.Answered), between(0, 30, session.Answered),
		between(0, 3000, busy), between(0, 3100, busy),
	}}
	var page strings.Builder
	if _, err := report.Write(&page, []report.Run{run}, report.Instance{VCPU: 1, GiB: 0.5}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	// The estimator draws the middle percentiles between the calls of a kind, so
	// what is held exactly is the longest: each kind's own, and no other's.
	for _, want := range []struct{ begins, ends string }{
		{"| ok | 2 | 50.0% |", "| 30ms |"},
		{"| failure:busy | 2 | 50.0% |", "| 3.1s |"},
	} {
		found := false
		for line := range strings.Lines(page.String()) {
			line = strings.TrimSuffix(line, "\n")
			found = found || strings.HasPrefix(line, want.begins) && strings.HasSuffix(line, want.ends)
		}
		if !found {
			t.Errorf("the report has no line beginning %q and ending %q:\n%s", want.begins, want.ends, page.String())
		}
	}
}

// Every answer the service should never give fails the run: a failure of its
// own, a request it did not answer, an answer of another call, a protocol it
// does not speak — and whatever the scenario expected and did not see.
func TestEveryHardFactFailsTheRun(t *testing.T) {
	t.Parallel()

	for _, kind := range []session.Kind{
		{Class: session.HTTP, Detail: "500"},
		{Class: session.HTTP, Detail: "404"},
		{Class: session.NoAnswer, Detail: "timeout"},
		{Class: session.NoAnswer, Detail: "reset"},
		session.Mismatched,
		{Class: session.Handshake, Detail: "2025-11-25"},
		{Class: session.Failure, Detail: "Something went wrong inside MathTrail."},
		{Class: session.JSONRPC, Detail: "-32603"},
		// The first code a server may use for errors of its own says nothing by
		// itself: a pace's refusal is told by its message, and is paced.
		{Class: session.JSONRPC, Detail: "-32000"},
		// A pace's error of the protocol carried on a failing status is told by
		// the status, and fails the run as every such status does.
		{Class: session.HTTP, Detail: "429(-32000)"},
		// A status of a payload nobody has seen before is nothing the service is
		// meant to say.
		{Class: session.Class("error"), Detail: "unknown"},
	} {
		t.Run(kind.String(), func(t *testing.T) {
			t.Parallel()

			run := report.Run{Calls: []report.Call{between(0, 10, session.Answered), between(0, 10, kind)}}
			if hards := run.Verdict(); len(hards) != 1 || hards[0].What != kind.String() {
				t.Errorf("Verdict() = %+v, want %q alone", hards, kind)
			}
		})
	}

	t.Run("an expectation the run did not meet", func(t *testing.T) {
		t.Parallel()

		run := report.Run{Calls: []report.Call{between(0, 10, session.Answered)}, Broken: []string{"3 of 5 tasks accepted"}}
		if hards := run.Verdict(); len(hards) != 1 || hards[0].What != "3 of 5 tasks accepted" {
			t.Errorf("Verdict() = %+v, want the expectation alone", hards)
		}
	})
}

// What the service is meant to say — its answers, its refusals, a task turned
// away for want of a slot, a message a pace held back — fails nothing, however
// often it says it, and neither does a call the run itself abandoned.
func TestWhatTheServiceIsMeantToSayFailsNothing(t *testing.T) {
	t.Parallel()

	run := report.Run{}
	for _, kind := range []session.Kind{
		session.Answered,
		{Class: session.Rejected, Detail: "solver_error"},
		{Class: session.Stale, Detail: "stale_request"},
		{Class: session.Limited, Detail: "limit_reached"},
		session.Paced,
		session.Busy,
		{Class: session.Stopped},
	} {
		run.Calls = append(run.Calls, between(0, 10, kind))
	}
	if hards := run.Verdict(); len(hards) != 0 {
		t.Errorf("Verdict() = %+v, want nothing", hards)
	}
}

// The report says what the run came to: the verdict first, every kind of
// answer, and what a unit of work cost against the free tier.
func TestTheReportShowsTheVerdictEveryKindAndTheCost(t *testing.T) {
	t.Parallel()

	run := report.Run{
		Scenario: "lesson",
		Began:    start, Ended: start.Add(90 * time.Second),
		Calls:    []report.Call{between(0, 150, session.Answered), between(5000, 5100, session.Kind{Class: session.Stale})},
		Requests: 3,
		Unit:     "accepted task",
		Units:    1,
	}
	var page strings.Builder
	hards, err := report.Write(&page, []report.Run{run}, report.Instance{VCPU: 1, GiB: 0.5})
	if err != nil || len(hards) != 0 {
		t.Fatalf("Write() = %v, %v, want no hard fact and no error", hards, err)
	}
	for _, want := range []string{
		"# Load: lesson",
		"**Verdict:** clean",
		"| ok | 1 |",
		"| stale | 1 |",
		"| vCPU-seconds | 0.300 | 600000 |",
		"| requests | 3.000 | 666667 |",
		"vCPU-seconds run out first",
	} {
		if !strings.Contains(page.String(), want) {
			t.Errorf("the report has no %q:\n%s", want, page.String())
		}
	}
}

// The calls due and not sent are told when there were any, and only then.
func TestTheReportTellsTheCallsNotSentOnlyWhenThereWereAny(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		unsent int
		want   bool
	}{{0, false}, {3, true}} {
		t.Run(strconv.Itoa(test.unsent), func(t *testing.T) {
			t.Parallel()

			run := report.Run{Scenario: "saturation", Calls: []report.Call{between(0, 10, session.Answered)}, Unsent: test.unsent}
			var page strings.Builder
			if _, err := report.Write(&page, []report.Run{run}, report.Instance{VCPU: 1, GiB: 0.5}); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if got := strings.Contains(page.String(), fmt.Sprintf(", %d due and not sent,", test.unsent)); got != test.want {
				t.Errorf("the report tells %d calls not sent: %v, want %v:\n%s", test.unsent, got, test.want, page.String())
			}
		})
	}
}

// A kind of answer is written as one cell of its table, and as one item of
// the list of hard facts, whatever its words: a bar or a line break in a
// sentence of the service's would otherwise split the one or the other.
func TestAKindKeepsToItsCell(t *testing.T) {
	t.Parallel()

	kind := session.Kind{Class: session.Failure, Detail: "Left | right.\nBelow."}
	run := report.Run{Scenario: "lesson", Calls: []report.Call{between(0, 10, kind)}}
	var page strings.Builder
	if _, err := report.Write(&page, []report.Run{run}, report.Instance{VCPU: 1, GiB: 0.5}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	for _, want := range []string{
		`- failure:Left \| right. Below. × 1` + "\n",
		`| failure:Left \| right. Below. | 1 | 100.0% |`,
	} {
		if !strings.Contains(page.String(), want) {
			t.Errorf("the report has no %q:\n%s", want, page.String())
		}
	}
}

// A run stopped before its end says so, before its numbers.
func TestTheReportSaysARunWasStopped(t *testing.T) {
	t.Parallel()

	run := report.Run{Scenario: "lesson", Calls: []report.Call{between(0, 10, session.Answered)}, Stopped: true}
	var page strings.Builder
	if _, err := report.Write(&page, []report.Run{run}, report.Instance{VCPU: 1, GiB: 0.5}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if !strings.Contains(page.String(), "**Stopped** before its end") {
		t.Errorf("the report does not say the run was stopped:\n%s", page.String())
	}
}
