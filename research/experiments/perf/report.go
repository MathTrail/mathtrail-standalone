package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// charsPerToken is the rough rule a package's tokens are estimated by: a
// quarter of its characters. It is a rule of thumb for English text and JSON,
// not a tokenizer's count, and the paper calls it an estimate.
const charsPerToken = 4

// measurements are everything one run of the harness found.
type measurements struct {
	tasks, passes int
	timings       []timing
	solverRuns    []solverRun
	cost          cost
	packages      []packageAt
}

// writeAll writes the tables and the numbers the paper cites into a
// directory.
func writeAll(dir string, m *measurements) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("perf: make the results directory: %w", err)
	}
	writers := []struct {
		name  string
		write func(path string) error
	}{
		{"timings.csv", func(path string) error { return writeCSV(path, timingTable(m.timings)) }},
		{"solver_runs.csv", func(path string) error { return writeCSV(path, solverRunTable(m.solverRuns)) }},
		{"packages.csv", func(path string) error { return writeCSV(path, packageTable(m.packages)) }},
		{"numbers.txt", func(path string) error { return writeLines(path, numberLines(m)) }},
	}
	for _, w := range writers {
		if err := w.write(filepath.Join(dir, w.name)); err != nil {
			return fmt.Errorf("perf: write %s: %w", w.name, err)
		}
	}
	return nil
}

func writeCSV(path string, rows [][]string) error {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	out := csv.NewWriter(file)
	if err := out.WriteAll(rows); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// writeLines writes key=value lines into a file made as the tables are made.
func writeLines(path string, lines []string) error {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	if _, err := file.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func timingTable(timings []timing) [][]string {
	table := [][]string{{"task", "pass", "examine_ns", "judge_ns"}}
	for _, t := range timings {
		table = append(table, []string{t.task, strconv.Itoa(t.pass), strconv.FormatInt(t.examine.Nanoseconds(), 10), strconv.FormatInt(t.judge.Nanoseconds(), 10)})
	}
	return table
}

func solverRunTable(runs []solverRun) [][]string {
	table := [][]string{{"task", "pass", "solver_run", "status", "steps", "duration_ns"}}
	for _, r := range runs {
		table = append(table, []string{
			r.task, strconv.Itoa(r.pass), strconv.Itoa(r.index), string(r.status),
			strconv.FormatUint(r.steps, 10), strconv.FormatInt(r.took.Nanoseconds(), 10),
		})
	}
	return table
}

func packageTable(packs []packageAt) [][]string {
	table := [][]string{{"topic", "grade_level", "difficulty", "turns", "smallest_bytes", "largest_bytes", "largest_characters", "largest_tokens_estimated"}}
	for _, p := range packs {
		table = append(table, []string{
			p.topic, string(p.point.GradeLevel), strconv.Itoa(p.point.Difficulty), strconv.Itoa(p.turns),
			strconv.Itoa(p.smallest), strconv.Itoa(p.largest), strconv.Itoa(p.largestChars), strconv.Itoa(p.largestChars / charsPerToken),
		})
	}
	return table
}

// numberLines are the numbers the paper cites, one key=value line each, with
// the machine they were measured on.
func numberLines(m *measurements) []string {
	lines := []string{"# Numbers of the measurement of what a review costs, computed by perf."}
	lines = append(lines, machineLines()...)
	lines = append(lines, "tasks="+strconv.Itoa(m.tasks), "passes="+strconv.Itoa(m.passes))
	lines = append(lines, timingLines(m.timings)...)
	lines = append(lines, solverLines(m.solverRuns)...)
	lines = append(lines,
		"benchmark_passes="+strconv.Itoa(m.cost.passes),
		"review_mib="+oneDecimal(float64(m.cost.bytes)/(1024*1024)),
		"review_allocations="+strconv.FormatInt(m.cost.allocs, 10),
	)
	return append(lines, packageLines(m.packages)...)
}

// timingLines are the two halves of a review and the whole, in milliseconds:
// each with the interval of its median over tasks, and the whole with the
// least and the most the median of one pass over every task came to, which
// shows how much the machine itself moved while it measured.
func timingLines(timings []timing) []string {
	parts := []struct {
		name string
		of   func(t *timing) time.Duration
	}{
		{"examine", func(t *timing) time.Duration { return t.examine }},
		{"judge", func(t *timing) time.Duration { return t.judge }},
		{"review", func(t *timing) time.Duration { return t.examine + t.judge }},
	}
	var lines []string
	for _, part := range parts {
		var all []float64
		byTask := map[string]int{}
		var groups [][]float64
		for i := range timings {
			ms := float64(part.of(&timings[i])) / float64(time.Millisecond)
			all = append(all, ms)
			at, seen := byTask[timings[i].task]
			if !seen {
				at = len(groups)
				byTask[timings[i].task] = at
				groups = append(groups, nil)
			}
			groups[at] = append(groups[at], ms)
		}
		low, high := medianInterval(groups, part.name)
		lines = append(lines, spreadLines(part.name, spreadOf(all))...)
		lines = append(lines, part.name+"_median_low_ms="+twoDecimals(low), part.name+"_median_high_ms="+twoDecimals(high))
	}
	least, most := passMedians(timings)
	return append(lines, "review_pass_median_least_ms="+twoDecimals(least), "review_pass_median_most_ms="+twoDecimals(most))
}

// passMedians are the least and the most the median of a whole review came to
// in one pass over every task.
func passMedians(timings []timing) (least, most float64) {
	byPass := map[int][]float64{}
	for _, t := range timings {
		byPass[t.pass] = append(byPass[t.pass], float64(t.examine+t.judge)/float64(time.Millisecond))
	}
	least, most = math.Inf(1), math.Inf(-1)
	for _, reviews := range byPass {
		median := quantile(slices.Sorted(slices.Values(reviews)), 0.5)
		least, most = min(least, median), max(most, median)
	}
	return least, most
}

func spreadLines(name string, s spread) []string {
	return []string{
		name + "_median_ms=" + twoDecimals(s.median),
		name + "_p95_ms=" + twoDecimals(s.p95),
		name + "_p99_ms=" + twoDecimals(s.p99),
		name + "_max_ms=" + twoDecimals(s.max),
	}
}

// solverLines are the solver's runs: how many, their steps against the
// sandbox's limit, and their durations.
func solverLines(runs []solverRun) []string {
	steps := make([]float64, 0, len(runs))
	took := make([]time.Duration, 0, len(runs))
	ok := 0
	for _, r := range runs {
		steps = append(steps, float64(r.steps))
		took = append(took, r.took)
		if r.status == solver.StatusOK {
			ok++
		}
	}
	s, d := spreadOf(steps), spreadOf(milliseconds(took))
	return []string{
		"solver_runs=" + strconv.Itoa(len(runs)),
		"solver_runs_ok=" + strconv.Itoa(ok),
		"solver_steps_median=" + whole(s.median),
		"solver_steps_p99=" + whole(s.p99),
		"solver_steps_max=" + whole(s.max),
		"solver_steps_limit=" + strconv.FormatUint(reviewing.SandboxLimits.Steps, 10),
		"solver_steps_max_percent_of_limit=" + oneDecimal(100*s.max/float64(reviewing.SandboxLimits.Steps)),
		"solver_run_median_ms=" + twoDecimals(d.median),
		"solver_run_p99_ms=" + twoDecimals(d.p99),
		"solver_run_max_ms=" + twoDecimals(d.max),
		"solver_timeout_ms=" + strconv.FormatInt(reviewing.SandboxLimits.Timeout.Milliseconds(), 10),
	}
}

// packageLines are the packages' sizes, in KiB and in estimated tokens, over
// every point and every turn of its reference tasks, each turn counted once,
// against the budget the service holds them to.
func packageLines(packs []packageAt) []string {
	var kib, tokens []float64
	built := 0
	for _, p := range packs {
		for i, size := range p.sizes {
			kib = append(kib, float64(size)/1024)
			tokens = append(tokens, float64(p.chars[i])/charsPerToken)
		}
		built += len(p.sizes)
	}
	sizes := slices.Sorted(slices.Values(kib))
	t := spreadOf(tokens)
	turns := 0
	if len(packs) > 0 {
		turns = packs[0].turns
	}
	return []string{
		"package_points=" + strconv.Itoa(len(packs)),
		"package_turns=" + strconv.Itoa(turns),
		"packages_built=" + strconv.Itoa(built),
		"package_kib_min=" + oneDecimal(quantile(sizes, 0)),
		"package_kib_median=" + oneDecimal(quantile(sizes, 0.5)),
		"package_kib_max=" + oneDecimal(quantile(sizes, 1)),
		"package_budget_kib=" + strconv.Itoa(content.PackageBudget/1024),
		"package_max_percent_of_budget=" + oneDecimal(100*quantile(sizes, 1)*1024/content.PackageBudget),
		"package_tokens_median=" + hundreds(t.median),
		"package_tokens_max=" + hundreds(t.max),
	}
}

// machineLines say what the numbers were measured on.
func machineLines() []string {
	return []string{
		"machine_processor=" + processor(),
		"machine_logical_processors=" + strconv.Itoa(runtime.NumCPU()),
		"machine_gomaxprocs=" + strconv.Itoa(runtime.GOMAXPROCS(0)),
		"machine_go=" + runtime.Version(),
	}
}

// processor is the processor's name as the system gives it, or the
// architecture where it gives none.
func processor() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return runtime.GOARCH
	}
	defer func() { _ = file.Close() }()
	lines := bufio.NewScanner(file)
	for lines.Scan() {
		key, value, found := strings.Cut(lines.Text(), ":")
		if found && strings.TrimSpace(key) == "model name" {
			return strings.Join(strings.Fields(value), " ")
		}
	}
	return runtime.GOARCH
}

func twoDecimals(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }
func oneDecimal(v float64) string  { return strconv.FormatFloat(v, 'f', 1, 64) }
func whole(v float64) string       { return strconv.FormatFloat(math.Round(v), 'f', 0, 64) }

// hundreds rounds an estimate to the nearest hundred, as precise as the rule
// it comes from.
func hundreds(v float64) string { return strconv.FormatFloat(100*math.Round(v/100), 'f', 0, 64) }
