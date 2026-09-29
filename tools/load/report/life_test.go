package report_test

import (
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// mib is a size of that many MiB, in bytes.
func mib(n uint64) uint64 { return n << 20 }

// Whatever the load, an instance of the service never panics, is never killed
// for its memory, never ends by itself and never holds more memory than the
// ceiling allows: each of those fails the run it ended with, as often as it
// happened.
func TestEveryFactOfAnInstancesLifeFailsTheRun(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		life  report.Life
		what  string
		count int
	}{
		{"panics", report.Life{Measured: true, Panics: 3}, "lines of a panic in the service's log", 3},
		{"killed, by the container's word", report.Life{Measured: true, OOMKilled: true, Exited: true, ExitCode: 137},
			"the service was killed for want of memory", 1},
		{"killed, by the cgroup's count", report.Life{Measured: true, OOMKills: 1}, "the service was killed for want of memory", 1},
		{"ended by itself", report.Life{Measured: true, Exited: true, ExitCode: 2}, "the service ended by itself, with code 2", 1},
		{"above the ceiling", report.Life{Measured: true, Peak: mib(401), Ceiling: mib(400)},
			"the service held more memory than the ceiling of 400 MiB", 1},
		{"a ceiling nothing measured", report.Life{Measured: false, Ceiling: mib(400)},
			"the instance went unmeasured, and its memory ceiling unchecked", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			run := report.Run{Calls: []report.Call{between(0, 10, session.Answered)}, Lives: []report.Life{test.life}}
			if hards := run.Verdict(); len(hards) != 1 || hards[0].What != test.what || hards[0].Count != test.count {
				t.Errorf("Verdict() = %+v, want %q × %d alone", hards, test.what, test.count)
			}
		})
	}

	t.Run("a life that is none of those", func(t *testing.T) {
		t.Parallel()

		run := report.Run{Lives: []report.Life{{Measured: true, Peak: mib(400), Ceiling: mib(400), CPU: time.Second}}}
		if hards := run.Verdict(); len(hards) != 0 {
			t.Errorf("Verdict() = %+v, want nothing", hards)
		}
	})
}

// A run on an instance of its own says what the instance spent while it went
// on, beside what the platform bills, and how every instance that ended with
// it lived: what it spent in all, its peak, how it ended, its panics and the
// runs of its sandbox.
func TestTheReportShowsWhatTheInstanceSpentAndHowItLived(t *testing.T) {
	t.Parallel()

	run := report.Run{
		Scenario: "lesson", Unit: "accepted task", Units: 2,
		Calls: []report.Call{between(0, 150, session.Answered)},
		Spent: &report.Spent{CPU: 1500 * time.Millisecond, Memory: mib(45)},
		Lives: []report.Life{{
			Measured: true, CPU: 12300 * time.Millisecond, Peak: mib(402),
			SolverRuns: []report.SolverRun{
				{Status: "ok", Steps: 100, Took: 10 * time.Millisecond},
				{Status: "ok", Steps: 300, Took: 30 * time.Millisecond},
				{Status: "timeout", Steps: 25_000_000, Took: 2 * time.Second},
			},
		}},
	}
	var page strings.Builder
	if _, err := report.Write(&page, []report.Run{run}, report.Instance{VCPU: 1, GiB: 0.5}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	for _, want := range []string{
		"Spent, as the instance's cgroup counted it: 1.500 CPU-seconds while the run went on, " +
			"0.750 for one accepted task, and at most 45 MiB held at once.",
		"| 1 | 12.300 | 402 MiB | by the run | 0 | 3 |",
		"| ok | 2 | 100 | 300 | 10ms | 30ms |",
		"| timeout | 1 | 25000000 | 25000000 | 2s | 2s |",
	} {
		if !strings.Contains(page.String(), want) {
			t.Errorf("the report has no %q:\n%s", want, page.String())
		}
	}
}

// The starts of a cold run are told one by one, and their middle: how long each
// took to answer its probe from the moment it was asked for and from the moment
// its process began, how long to its first call, and the memory it held then.
// Its work comes in no unit, so there is no cost of one to tell.
func TestTheReportShowsEveryStartAndTheirMiddle(t *testing.T) {
	t.Parallel()

	at := func(ms int) time.Time { return start.Add(time.Duration(ms) * time.Millisecond) }
	run := report.Run{Scenario: "cold", Starts: []report.Start{
		{Asked: at(0), Began: at(900), Healthy: at(1000), Answered: at(1200), Memory: mib(36)},
		{Asked: at(0), Began: at(700), Healthy: at(800), Answered: at(1000), Memory: mib(38)},
		{Asked: at(0), Began: at(500), Memory: 0},
	}}
	var page strings.Builder
	if _, err := report.Write(&page, []report.Run{run}, report.Instance{VCPU: 1, GiB: 0.5}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	for _, want := range []string{
		"| 1 | 1s | 100ms | 1.2s | 36 MiB |",
		"| 3 | never | never | never | — |",
		"| middle | 800ms | 100ms | 1s | 36 MiB |",
	} {
		if !strings.Contains(page.String(), want) {
			t.Errorf("the report has no %q:\n%s", want, page.String())
		}
	}
	for _, unwanted := range []string{"Billed", "was done"} {
		if strings.Contains(page.String(), unwanted) {
			t.Errorf("the report of a run of no unit tells a cost, %q:\n%s", unwanted, page.String())
		}
	}
}

// An instance that ended by itself is told with the last lines it wrote,
// indented, so that nothing a line says can end the block early.
func TestTheReportTellsTheLastLinesOfAnInstanceThatEnded(t *testing.T) {
	t.Parallel()

	run := report.Run{Scenario: "lesson: start", Lives: []report.Life{{
		Exited: true, ExitCode: 1, LastLines: "mathtrail: config: invalid\n```",
	}}}
	var page strings.Builder
	if _, err := report.Write(&page, []report.Run{run}, report.Instance{VCPU: 1, GiB: 0.5}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if want := "Instance 1 ended by itself; its last lines:\n\n    mathtrail: config: invalid\n    ```\n"; !strings.Contains(page.String(), want) {
		t.Errorf("the report has no %q:\n%s", want, page.String())
	}
}
