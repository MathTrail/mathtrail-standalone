package instance

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
)

// solverRun is the line the service writes for every run of its sandbox.
const solverRun = "solver_run"

// The words the runtime begins with when a panic nothing caught, or a fault
// nothing can catch, ends the process: it writes them itself, outside the
// service's own lines.
var crashes = []string{"panic:", "fatal error:"}

// readLog is what the service's log tells a load run: how many of its lines
// tell of a panic — a line of the service's own that carries what it panicked
// with, or the first line the runtime writes when one ends the process — and
// every run of its sandbox, with what each came to, the steps it took and how
// long.
func readLog(written string) (panics int, runs []report.SolverRun) {
	for line := range strings.Lines(written) {
		var entry struct {
			Message    string          `json:"message"`
			Panic      json.RawMessage `json:"panic"`
			Status     string          `json:"status"`
			Steps      uint64          `json:"steps"`
			DurationMS int64           `json:"duration_ms"`
		}
		if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &entry); err != nil {
			// The runtime starts the first line of its words at the margin, and
			// indents what follows it — a panic raised again among them.
			if crashed(line) {
				panics++
			}
			continue
		}
		if entry.Panic != nil {
			panics++
		}
		if entry.Message == solverRun {
			runs = append(runs, report.SolverRun{
				Status: entry.Status, Steps: entry.Steps, Took: time.Duration(entry.DurationMS) * time.Millisecond,
			})
		}
	}
	return panics, runs
}

// crashed says whether a line is the first the runtime writes as a panic or a
// fault ends the process.
func crashed(line string) bool {
	for _, words := range crashes {
		if strings.HasPrefix(line, words) {
			return true
		}
	}
	return false
}
