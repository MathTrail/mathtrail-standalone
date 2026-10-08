package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// acceptedTasks are the first reference tasks the experiment with injected
// defects accepts as they are, prepared for review as the harness prepares
// them.
func acceptedTasks(t *testing.T, shipped *content.Content, runner solver.Runner, n int) []task {
	t.Helper()
	var tasks []task
	for _, h := range reviewing.Hosts(shipped) {
		against := reviewing.Without(shipped, h.Base.Task.Question)
		verdict, err := reviewing.Review(context.Background(), runner, against, &h.Base)
		if err != nil {
			t.Fatalf("reviewing.Review(%s) error = %v", h.ID, err)
		}
		if verdict.Refused() || len(verdict.Unchecked) > 0 {
			continue
		}
		prepared, err := reviewing.Prepare(runner, against, &h.Base)
		if err != nil {
			t.Fatalf("reviewing.Prepare(%s) error = %v", h.ID, err)
		}
		if tasks = append(tasks, task{id: h.ID, prepared: prepared}); len(tasks) == n {
			return tasks
		}
	}
	t.Fatalf("the experiment with injected defects accepts %d reference tasks, want %d", len(tasks), n)
	return nil
}

// wantFiles fails unless a directory holds exactly these entries, in the
// order of their names.
func wantFiles(t *testing.T, dir string, want ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("os.ReadDir(%s) error = %v", dir, err)
	}
	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		got = append(got, entry.Name())
	}
	if !slices.Equal(got, want) {
		t.Errorf("%s holds %v, want %v", dir, got, want)
	}
}

// wantSameBytes fails unless a file the run wrote holds the bytes of the one
// committed, and names the first line where the two part.
func wantSameBytes(t *testing.T, written, shipped string) {
	t.Helper()
	got, want := readLines(t, written), readLines(t, shipped)
	at := 0
	for at < min(len(got), len(want)) && got[at] == want[at] {
		at++
	}
	if at < len(got) || at < len(want) {
		t.Errorf("%s parts from %s at line %d: %q, want %q", written, shipped, at+1, lineAt(got, at), lineAt(want, at))
	}
}

// lineAt is a line of a file, or nothing past its end.
func lineAt(lines []string, at int) string {
	if at < len(lines) {
		return lines[at]
	}
	return ""
}

// wantTable fails unless a table the run wrote holds these rows.
func wantTable(t *testing.T, path string, want [][]string) {
	t.Helper()
	if got := readCSV(t, path); !slices.EqualFunc(got, want, slices.Equal) {
		t.Errorf("%s holds %v, want %v", path, got, want)
	}
}

func nanoseconds(d time.Duration) string { return strconv.FormatInt(d.Nanoseconds(), 10) }

// wantTimingsInPassOrder fails unless every task was timed once in each pass,
// a pass after the one before, both halves of every review taking time, and
// the table holds the reviews in that order.
func wantTimingsInPassOrder(t *testing.T, path string, timings []timing, tasks []task, passes int) {
	t.Helper()
	if len(timings) != passes*len(tasks) {
		t.Fatalf("%d timings, want each of %d tasks once in each of %d passes", len(timings), len(tasks), passes)
	}
	table := [][]string{{"task", "pass", "examine_ns", "judge_ns"}}
	for i, got := range timings {
		pass, id := i/len(tasks), tasks[i%len(tasks)].id
		if got.task != id || got.pass != pass || got.examine <= 0 || got.judge <= 0 {
			t.Errorf("timing %d = %s in pass %d taking %v and %v, want %s in pass %d, both halves taking time", i, got.task, got.pass, got.examine, got.judge, id, pass)
		}
		table = append(table, []string{id, strconv.Itoa(pass), nanoseconds(got.examine), nanoseconds(got.judge)})
	}
	wantTable(t, path, table)
}

// wantTwoRunsAReview fails unless every review ran the solver twice, as
// handed in and with the labels shifted, both runs ending normally, and the
// table holds the runs in that order.
func wantTwoRunsAReview(t *testing.T, path string, runs []solverRun, tasks []task, passes int) {
	t.Helper()
	if len(runs) != 2*passes*len(tasks) {
		t.Fatalf("%d runs of the solver, want 2 in each of %d reviews", len(runs), passes*len(tasks))
	}
	table := [][]string{{"task", "pass", "solver_run", "status", "steps", "duration_ns"}}
	for i, got := range runs {
		review := i / 2
		pass, id, index := review/len(tasks), tasks[review%len(tasks)].id, i%2
		if got.task != id || got.pass != pass || got.index != index || got.status != solver.StatusOK {
			t.Errorf("run %d = run %d of %s in pass %d, %s; want run %d of %s in pass %d, %s", i, got.index, got.task, got.pass, got.status, index, id, pass, solver.StatusOK)
		}
		table = append(table, []string{id, strconv.Itoa(pass), strconv.Itoa(index), string(got.status), strconv.FormatUint(got.steps, 10), nanoseconds(got.took)})
	}
	wantTable(t, path, table)
}

// macroKey is what a key of the numbers may be for the macros the paper
// cites them by.
var macroKey = regexp.MustCompile(`^[a-z0-9_]+$`)

// wantMacroLines fails for every line the macros the paper cites the numbers
// by would refuse: a line that is neither blank nor a comment is key=value,
// with a key of lower-case letters, digits and underscores, a value, and a
// key no line before it gave.
func wantMacroLines(t *testing.T, lines []string) {
	t.Helper()
	given := map[string]bool{}
	for n, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !found || !macroKey.MatchString(key) || value == "" {
			t.Errorf("line %d = %q, want key=value, the key of lower-case letters, digits and underscores", n+1, line)
		}
		if given[key] {
			t.Errorf("line %d gives %s again", n+1, key)
		}
		given[key] = true
	}
}

// startingWith are the lines that start with a prefix.
func startingWith(lines []string, prefix string) []string {
	var found []string
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			found = append(found, line)
		}
	}
	return found
}

// wantProcessorLine fails unless the processor is named by one line of text
// that a reader of key=value lines cannot misread.
func wantProcessorLine(t *testing.T, lines []string) {
	t.Helper()
	name := processor()
	if name == "" || strings.ContainsAny(name, "=\n") {
		t.Errorf("processor() = %q, want one line of text with no '='", name)
	}
	if got, want := startingWith(lines, "machine_processor="), []string{"machine_processor=" + name}; !slices.Equal(got, want) {
		t.Errorf("the processor's lines = %q, want %q", got, want)
	}
}

// A small measurement is written whole: its four files and no others, the
// package sizes exactly as the committed results hold them, every review
// timed in pass order with the solver run twice, the benchmark over as many
// passes as the committed numbers count, and every number a line the macros
// the paper cites it by accept. What a review allocates is not checked: the
// benchmark counts what the whole test binary allocates meanwhile.
func TestASmallMeasurementIsWrittenWhole(t *testing.T) {
	t.Parallel()
	shipped, runner := loaded(t), newRunner(t)
	tasks := acceptedTasks(t, shipped, runner, 3)
	const passes = 2
	ctx := context.Background()
	timings, solverRuns, err := timeReviews(ctx, tasks, passes)
	if err != nil {
		t.Fatalf("timeReviews() error = %v", err)
	}
	spent, err := reviewCost(ctx, tasks)
	if err != nil {
		t.Fatalf("reviewCost() error = %v", err)
	}
	m := measurements{tasks: len(tasks), passes: passes, timings: timings, solverRuns: solverRuns, cost: spent, packages: builtPackages(t)}
	dir := t.TempDir()
	if err := writeAll(dir, &m); err != nil {
		t.Fatalf("writeAll() error = %v", err)
	}

	wantFiles(t, dir, "numbers.txt", "packages.csv", "solver_runs.csv", "timings.csv")
	wantSameBytes(t, filepath.Join(dir, "packages.csv"), committed("packages.csv"))
	wantTimingsInPassOrder(t, filepath.Join(dir, "timings.csv"), timings, tasks, passes)
	wantTwoRunsAReview(t, filepath.Join(dir, "solver_runs.csv"), solverRuns, tasks, passes)

	lines, committedLines := readLines(t, filepath.Join(dir, "numbers.txt")), readLines(t, committed("numbers.txt"))
	wantMacroLines(t, lines)
	wantProcessorLine(t, lines)
	sizes := startingWith(lines, "package")
	if want := startingWith(committedLines, "package"); !slices.Equal(sizes, want) {
		t.Errorf("the package sizes = %q, want the committed %q", sizes, want)
	}
	if want := packageLines(m.packages); !slices.Equal(sizes, want) {
		t.Errorf("the package sizes = %q, want %q", sizes, want)
	}
	committedNumbers := numbersOf(committedLines)
	if strconv.Itoa(spent.passes) != committedNumbers["benchmark_passes"] {
		t.Errorf("reviewCost() over %d passes, want the committed %s", spent.passes, committedNumbers["benchmark_passes"])
	}
	wantNumbers(t, numbersOf(lines), map[string]string{
		"tasks": "3", "passes": "2", "solver_runs": "12", "solver_runs_ok": "12",
		"benchmark_passes":   committedNumbers["benchmark_passes"],
		"solver_steps_limit": committedNumbers["solver_steps_limit"],
		"solver_timeout_ms":  committedNumbers["solver_timeout_ms"],
	})
}

// tinyMeasurement is one review of one task and one package: enough to put
// something in every file.
func tinyMeasurement() *measurements {
	ms := time.Millisecond
	return &measurements{
		tasks: 1, passes: 1,
		timings: []timing{{task: "a", examine: ms, judge: 2 * ms}},
		solverRuns: []solverRun{
			{task: "a", status: solver.StatusOK, steps: 1_000, took: ms},
			{task: "a", index: 1, status: solver.StatusOK, steps: 1_000, took: ms},
		},
		cost: cost{passes: 3, bytes: 1 << 20, allocs: 1_000},
		packages: []packageAt{{
			topic: "a", point: rating.Point{GradeLevel: rating.Grades12, Difficulty: 1}, turns: 1,
			sizes: []int{1024}, chars: []int{1024}, smallest: 1024, largest: 1024, largestChars: 1024,
		}},
	}
}

// The results go into a directory, and a file where it should be is refused
// rather than taken for one.
func TestAResultsDirectoryThatIsAFileIsRefused(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "results")
	if err := os.WriteFile(path, []byte("a file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeAll(path, tinyMeasurement()); err == nil || !strings.HasPrefix(err.Error(), "perf: make the results directory: ") {
		t.Errorf("writeAll() error = %v, want perf: make the results directory: …", err)
	}
}

// Writing stops at the first file that cannot be made and names it, and
// makes none of the files after it: a directory of results is whole or
// visibly short.
func TestWritingStopsAtTheFirstFileThatCannotBeMade(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		blocked string
		holds   []string
	}{
		{"timings.csv", []string{"timings.csv"}},
		{"solver_runs.csv", []string{"solver_runs.csv", "timings.csv"}},
		{"packages.csv", []string{"packages.csv", "solver_runs.csv", "timings.csv"}},
		{"numbers.txt", []string{"numbers.txt", "packages.csv", "solver_runs.csv", "timings.csv"}},
	} {
		t.Run(tc.blocked, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, tc.blocked), 0o750); err != nil {
				t.Fatal(err)
			}
			if err := writeAll(dir, tinyMeasurement()); err == nil || !strings.HasPrefix(err.Error(), "perf: write "+tc.blocked+": ") {
				t.Errorf("writeAll() error = %v, want perf: write %s: …", err, tc.blocked)
			}
			wantFiles(t, dir, tc.holds...)
		})
	}
}

// fullDisk is a file every write to fails as on a disk with no space left.
const fullDisk = "/dev/full"

// A file the disk cannot take whole is an error that names it, rather than a
// short file left to be read as the whole one.
func TestAFileTheDiskCannotTakeIsAnError(t *testing.T) {
	t.Parallel()
	if _, err := os.Stat(fullDisk); err != nil {
		t.Skipf("no %s to stand for a full disk: %v", fullDisk, err)
	}
	for _, name := range []string{"timings.csv", "numbers.txt"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if err := os.Symlink(fullDisk, filepath.Join(dir, name)); err != nil {
				t.Fatal(err)
			}
			err := writeAll(dir, tinyMeasurement())
			if !errors.Is(err, syscall.ENOSPC) || !strings.HasPrefix(err.Error(), "perf: write "+name+": ") {
				t.Errorf("writeAll() error = %v, want perf: write %s: …, with no space left", err, name)
			}
		})
	}
}
