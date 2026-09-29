package report

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

// Spent is what an instance spent while a run went on, as its cgroup counted
// it.
type Spent struct {
	// CPU is the processor time it used.
	CPU time.Duration
	// Memory is the most memory it was seen to hold at once, in bytes.
	Memory uint64
}

// Life is the life of one instance a run started, from its start to its end:
// what it spent, how it ended, and what its log told.
type Life struct {
	// Measured says whether the instance's cgroup could be read at all. The
	// numbers that come from it — the processor time, the peak and the kills —
	// are zero when it could not.
	Measured bool
	// CPU is the processor time it used over its whole life.
	CPU time.Duration
	// Peak is the most memory it ever held, in bytes.
	Peak uint64
	// OOMKills is how many processes the kernel killed in it for want of
	// memory, and OOMKilled is the container's own word that its service was
	// one of them.
	OOMKills  int
	OOMKilled bool
	// Exited is an instance that ended before the run ended it, ExitCode how
	// it ended, and LastLines the last lines it wrote: what it said as it went.
	Exited    bool
	ExitCode  int
	LastLines string
	// Panics is how many lines of its log tell of a panic.
	Panics int
	// SolverRuns are the runs of its sandbox, as its log tells them.
	SolverRuns []SolverRun
	// Ceiling is the most memory it may have held without failing the run, in
	// bytes; zero is no ceiling.
	Ceiling uint64
}

// SolverRun is one run of the sandbox, as the service's log tells it.
type SolverRun struct {
	Status string
	Steps  uint64
	Took   time.Duration
}

// Start is one start of the service, as a cold run measures it.
type Start struct {
	// Asked is when the service was asked to start, Began when its process
	// began, Healthy when its probe first answered and Answered when its first
	// tool call did. A zero moment is one it never came to.
	Asked, Began, Healthy, Answered time.Time
	// Memory is the most memory it had held once its first call was answered,
	// in bytes, and zero when nothing measured it.
	Memory uint64
}

// facts are the facts of the instance's life that fail a run, and how many
// times each happened: whatever the load, the service never panics, is never
// killed for its memory, never ends by itself, and never holds more memory
// than the ceiling allows — which a ceiling nothing measured cannot say.
func (l *Life) facts() map[string]int {
	facts := map[string]int{}
	if l.Panics > 0 {
		facts["lines of a panic in the service's log"] = l.Panics
	}
	switch {
	case l.OOMKilled || l.OOMKills > 0:
		facts["the service was killed for want of memory"]++
	case l.Exited:
		facts[fmt.Sprintf("the service ended by itself, with code %d", l.ExitCode)]++
	}
	switch {
	case l.Ceiling > 0 && !l.Measured:
		// A ceiling the run was asked to hold the instance to and could not is
		// said, rather than taken as kept.
		facts["the instance went unmeasured, and its memory ceiling unchecked"]++
	case l.Ceiling > 0 && l.Peak > l.Ceiling:
		facts[fmt.Sprintf("the service held more memory than the ceiling of %s", mebibytes(l.Ceiling))]++
	}
	return facts
}

// writeSpent writes what the instance spent while the run went on, when
// anything measured it: beside what the platform bills, what the instance
// used of it.
func (r *Run) writeSpent(page *strings.Builder) {
	if r.Spent == nil {
		return
	}
	fmt.Fprintf(page, "Spent, as the instance's cgroup counted it: %.3f CPU-seconds while the run went on",
		r.Spent.CPU.Seconds())
	if r.Units > 0 {
		fmt.Fprintf(page, ", %.3f for one %s", r.Spent.CPU.Seconds()/float64(r.Units), r.Unit)
	}
	fmt.Fprintf(page, ", and at most %s held at once.\n\n", mebibytes(r.Spent.Memory))
}

// writeLives writes the life of every instance that ended with the run, and
// the runs of their sandboxes.
func (r *Run) writeLives(page *strings.Builder) {
	if len(r.Lives) == 0 {
		return
	}
	page.WriteString("The instances, as their cgroups counted them and their logs told:\n\n" +
		"| Instance | CPU-seconds | Memory peak | Ended | Lines of a panic | Solver runs |\n" +
		"|---:|---:|---:|---|---:|---:|\n")
	var runs []SolverRun
	for i := range r.Lives {
		life := &r.Lives[i]
		cpu, peak := "—", "—"
		if life.Measured {
			cpu, peak = fmt.Sprintf("%.3f", life.CPU.Seconds()), mebibytes(life.Peak)
		}
		fmt.Fprintf(page, "| %d | %s | %s | %s | %d | %d |\n",
			i+1, cpu, peak, ending(life), life.Panics, len(life.SolverRuns))
		runs = append(runs, life.SolverRuns...)
	}
	page.WriteString("\n")
	for i := range r.Lives {
		if last := r.Lives[i].LastLines; last != "" {
			// Indented rather than fenced: nothing a line says can close it.
			fmt.Fprintf(page, "Instance %d ended by itself; its last lines:\n\n    %s\n\n",
				i+1, strings.ReplaceAll(last, "\n", "\n    "))
		}
	}
	writeSolverRuns(page, runs)
}

// ending is how an instance ended, in words.
func ending(life *Life) string {
	switch {
	case life.OOMKilled || life.OOMKills > 0:
		return "killed for want of memory"
	case life.Exited:
		return fmt.Sprintf("by itself, code %d", life.ExitCode)
	}
	return "by the run"
}

// writeSolverRuns writes the runs of the sandboxes, status by status: how many
// ended so, and the steps and the time they took.
func writeSolverRuns(page *strings.Builder, runs []SolverRun) {
	if len(runs) == 0 {
		return
	}
	byStatus := map[string][]SolverRun{}
	for _, run := range runs {
		byStatus[run.Status] = append(byStatus[run.Status], run)
	}
	page.WriteString("| Solver runs | Count | Steps p50 | Steps max | Took p50 | Took max |\n" +
		"|---|---:|---:|---:|---:|---:|\n")
	for _, status := range slices.Sorted(maps.Keys(byStatus)) {
		these := byStatus[status]
		steps := make([]uint64, len(these))
		took := make([]time.Duration, len(these))
		for i, run := range these {
			steps[i], took[i] = run.Steps, run.Took
		}
		slices.Sort(steps)
		slices.Sort(took)
		fmt.Fprintf(page, "| %s | %d | %d | %d | %s | %s |\n", cell(status), len(these),
			median(steps), steps[len(steps)-1], rounded(median(took)), rounded(took[len(took)-1]))
	}
	page.WriteString("\n")
}

// writeStarts writes the starts a cold run measured, one by one and their
// middle: how long each took from the moment it was asked for, and from the
// moment its process began, to its first answers.
func (r *Run) writeStarts(page *strings.Builder) {
	if len(r.Starts) == 0 {
		return
	}
	page.WriteString("| Start | Asked → probe | Began → probe | Asked → first call | Memory after it |\n" +
		"|---:|---:|---:|---:|---:|\n")
	var toProbe, fromBegan, toCall []time.Duration
	var memory []uint64
	for i, start := range r.Starts {
		fmt.Fprintf(page, "| %d | %s | %s | %s | %s |\n", i+1, between(start.Asked, start.Healthy),
			between(start.Began, start.Healthy), between(start.Asked, start.Answered), memoryOf(start.Memory))
		toProbe = gather(toProbe, start.Asked, start.Healthy)
		fromBegan = gather(fromBegan, start.Began, start.Healthy)
		toCall = gather(toCall, start.Asked, start.Answered)
		if start.Memory > 0 {
			memory = append(memory, start.Memory)
		}
	}
	fmt.Fprintf(page, "| middle | %s | %s | %s | %s |\n\n",
		middle(toProbe), middle(fromBegan), middle(toCall), memoryOf(medianOf(memory)))
}

// between is how long it was from one moment to a later one, or "never" when
// the later never came.
func between(from, to time.Time) string {
	if from.IsZero() || to.IsZero() {
		return "never"
	}
	return rounded(to.Sub(from)).String()
}

// gather adds how long it was between two moments to the lengths given, when
// both came.
func gather(lengths []time.Duration, from, to time.Time) []time.Duration {
	if from.IsZero() || to.IsZero() {
		return lengths
	}
	return append(lengths, to.Sub(from))
}

// middle is the middle of the lengths, or "never" when there is none.
func middle(lengths []time.Duration) string {
	if len(lengths) == 0 {
		return "never"
	}
	slices.Sort(lengths)
	return rounded(median(lengths)).String()
}

// medianOf is the middle of the sizes, zero when there is none.
func medianOf(sizes []uint64) uint64 {
	if len(sizes) == 0 {
		return 0
	}
	slices.Sort(sizes)
	return median(sizes)
}

// median is the middle value of values already in order: the lower of the
// two middle ones when there is an even number of them.
func median[T cmp.Ordered](sorted []T) T { return sorted[(len(sorted)-1)/2] }

// memoryOf is a size as a report writes it, or a dash when nothing measured it.
func memoryOf(bytes uint64) string {
	if bytes == 0 {
		return "—"
	}
	return mebibytes(bytes)
}

// mebibytes is a size in whole MiB.
func mebibytes(bytes uint64) string { return fmt.Sprintf("%d MiB", bytes>>20) }
