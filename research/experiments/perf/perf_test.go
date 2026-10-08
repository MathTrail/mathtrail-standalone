package main

import (
	"context"
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

func loaded(t *testing.T) *content.Content {
	t.Helper()
	shipped, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	return shipped
}

func newRunner(t *testing.T) solver.Runner {
	t.Helper()
	runner, err := reviewing.NewRunner()
	if err != nil {
		t.Fatalf("reviewing.NewRunner() error = %v", err)
	}
	return runner
}

// unhurriedRunner is the experiment's sandbox with time to spare, for a test
// of what the reviews decide over every reference task. Under the race
// detector and beside the other tests a solver runs many times slower than in
// the experiment, so the time a run may take, and wait for its slot, is no
// part of what such a test checks; the steps a run may take, which decide
// every verdict, stay the experiment's.
func unhurriedRunner(t *testing.T) solver.Runner {
	t.Helper()
	limits := reviewing.SandboxLimits
	limits.Timeout, limits.Wait = time.Minute, time.Minute
	runner, err := starlark.New(limits)
	if err != nil {
		t.Fatalf("starlark.New() error = %v", err)
	}
	return runner
}

// ladderPackages are the packages at every point of the ladder, built once
// for every test that reads them: building every one of them takes seconds
// under the race detector.
var ladderPackages = sync.OnceValues(func() ([]packageAt, error) {
	shipped, err := content.Load()
	if err != nil {
		return nil, err
	}
	return packages(shipped)
})

func builtPackages(t *testing.T) []packageAt {
	t.Helper()
	packs, err := ladderPackages()
	if err != nil {
		t.Fatalf("packages() error = %v", err)
	}
	return packs
}

// cancelled is a context whose caller has already gone.
func cancelled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%s) error = %v", path, err)
	}
	return strings.Split(string(data), "\n")
}

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("os.Open(%s) error = %v", path, err)
	}
	defer func() { _ = file.Close() }()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil || len(rows) == 0 {
		t.Fatalf("%s holds %d rows, error = %v; want a header at least", path, len(rows), err)
	}
	return rows
}

// committed is the path of a file of the results the paper's numbers come
// from, as they are committed beside the harness.
func committed(name string) string { return filepath.Join("results", name) }

// sameReview reviews a reference task as the harness times it and as the
// experiment with injected defects did, reports where the two part, and says
// whether the experiment refused it.
func sameReview(t *testing.T, shipped *content.Content, runner solver.Runner, h *reviewing.Host) (refused bool) {
	t.Helper()
	ctx := context.Background()
	want, err := reviewing.Review(ctx, runner, reviewing.Without(shipped, h.Base.Task.Question), &h.Base)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := reviewing.Prepare(runner, reviewing.Without(shipped, h.Base.Task.Question), &h.Base)
	if err != nil {
		t.Fatal(err)
	}
	r, err := review(ctx, &task{id: h.ID, prepared: prepared})
	if err != nil {
		t.Fatal(err)
	}
	got := reviewing.VerdictOf(&r.outcome, r.runs)
	if !slices.Equal(got.Codes, want.Codes) || !slices.Equal(got.Unchecked, want.Unchecked) || len(got.Runs) != len(want.Runs) {
		t.Errorf("%s: the timed review refuses with %v, leaving %v unrun, after %d runs; the experiment's with %v, leaving %v, after %d",
			h.ID, got.Codes, got.Unchecked, len(got.Runs), want.Codes, want.Unchecked, len(want.Runs))
	}
	return want.Refused()
}

// evenSample is a sample of hosts taken at even steps through all of them, so that
// it reaches every topic and level, with the first host that has a drawing
// added if none of them has one.
func evenSample(hosts []*reviewing.Host, size int) []*reviewing.Host {
	size = min(size, len(hosts))
	sample := make([]*reviewing.Host, 0, size+1)
	for i := range size {
		sample = append(sample, hosts[i*len(hosts)/size])
	}
	if !slices.ContainsFunc(sample, (*reviewing.Host).HasDrawing) {
		if at := slices.IndexFunc(hosts, (*reviewing.Host).HasDrawing); at >= 0 {
			sample = append(sample, hosts[at])
		}
	}
	return sample
}

// The review the harness times and benchmarks is the review of the experiment
// with injected defects: on the same task it refuses with the same checks, or
// accepts, with the same checks left unrun, over tasks of every level, with
// and without a drawing.
func TestTheTimedReviewIsTheExperimentsReview(t *testing.T) {
	t.Parallel()
	shipped := loaded(t)
	runner, err := reviewing.NewRunner()
	if err != nil {
		t.Fatal(err)
	}
	sample := evenSample(reviewing.Hosts(shipped), 40)
	refused, levels := 0, map[rating.GradeLevel]bool{}
	for _, h := range sample {
		if sameReview(t, shipped, runner, h) {
			refused++
		}
		levels[h.Level] = true
	}
	if refused == 0 || refused == len(sample) || len(levels) != len(rating.GradeLevels()) || !slices.ContainsFunc(sample, (*reviewing.Host).HasDrawing) {
		t.Errorf("the sample of %d tasks has %d refused, %d levels and a drawing %v; want some refused and some accepted, every level and a drawing",
			len(sample), refused, len(levels), slices.ContainsFunc(sample, (*reviewing.Host).HasDrawing))
	}
}

// A quantile is a value measured: the value at index ⌊q·n⌋, the last at q = 1.
func TestAQuantileIsAValueMeasured(t *testing.T) {
	t.Parallel()
	sorted := []float64{1, 2, 3, 4}
	for q, want := range map[float64]float64{0: 1, 0.25: 2, 0.5: 3, 0.74: 3, 0.95: 4, 1: 4} {
		if got := quantile(sorted, q); got != want {
			t.Errorf("quantile(%v, %v) = %v, want %v", sorted, q, got, want)
		}
	}
}

// A quantile of no values is no number, at every q, rather than an index
// outside them: a table with nothing in it reads as NaN instead of stopping
// the harness halfway through what it writes.
func TestAQuantileOfNothingIsNoNumber(t *testing.T) {
	t.Parallel()
	for _, q := range []float64{0, 0.5, 0.99, 1} {
		t.Run(strconv.FormatFloat(q, 'g', -1, 64), func(t *testing.T) {
			t.Parallel()
			if got := quantile(nil, q); !math.IsNaN(got) {
				t.Errorf("quantile(nil, %v) = %v, want NaN", q, got)
			}
		})
	}
}

// The interval of a median draws tasks with all their reviews, and holds the
// median: tasks whose reviews all take the same time give an interval of that
// time alone.
func TestTheMediansIntervalDrawsWholeTasks(t *testing.T) {
	t.Parallel()
	if low, high := medianInterval([][]float64{{2, 2}, {2, 2, 2}}, "test"); low != 2 || high != 2 {
		t.Errorf("interval of tasks all at 2 = %v–%v, want 2–2", low, high)
	}
	groups := [][]float64{{1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 4}, {5, 5, 5}}
	low, high := medianInterval(groups, "test")
	if low > 3 || high < 3 || low < 1 || high > 5 {
		t.Errorf("interval of tasks at 1 to 5 = %v–%v, want it to hold 3 within 1–5", low, high)
	}
}

// Every point of the ladder every topic is taught at has a package at every
// turn of its reference tasks, within the budget the service holds packages
// to.
func TestEveryPointOfTheLadderHasAPackageAtEveryTurn(t *testing.T) {
	t.Parallel()
	shipped := loaded(t)
	packs := builtPackages(t)
	points := 0
	for _, topic := range shipped.TopicIDs() {
		points += rating.Difficulties * len(shipped.LevelsOf(topic))
	}
	if len(packs) != points {
		t.Errorf("%d points with packages, want all %d", len(packs), points)
	}
	for _, p := range packs {
		if len(p.sizes) != p.turns || p.smallest == 0 || p.largest > content.PackageBudget || p.largestChars > p.largest {
			t.Errorf("%s at %s/%d: %d of %d turns built, %d–%d bytes; want every turn, within %d", p.topic, p.point.GradeLevel, p.point.Difficulty, len(p.sizes), p.turns, p.smallest, p.largest, content.PackageBudget)
		}
	}
}

// A cycle of turns is a count of answers after which every group of reference
// tasks that take turns has gone round a whole number of times, and the least
// such count.
func TestTheCycleOfTurnsTakesEveryGroupRound(t *testing.T) {
	t.Parallel()
	shipped := loaded(t)
	cycle := turnCycle(shipped)
	groups := map[string]int{}
	examples := shipped.Examples()
	for i := range examples {
		groups[examples[i].Topic+"/"+string(examples[i].GradeLevel)+"/"+strconv.Itoa(examples[i].Difficulty)]++
	}
	sizes := map[int]bool{}
	for group, size := range groups {
		if cycle%size != 0 {
			t.Errorf("%s has %d tasks, which a cycle of %d does not go round", group, size, cycle)
		}
		sizes[size] = true
	}
	for divisor := 1; divisor < cycle; divisor++ {
		if cycle%divisor != 0 {
			continue
		}
		roundsAll := true
		for size := range sizes {
			roundsAll = roundsAll && divisor%size == 0
		}
		if roundsAll {
			t.Errorf("a cycle of %d answers goes round every group, shorter than %d", divisor, cycle)
		}
	}
}

// numbersOf reads key=value lines into a map.
func numbersOf(lines []string) map[string]string {
	values := map[string]string{}
	for _, line := range lines {
		if key, value, found := strings.Cut(line, "="); found {
			values[key] = value
		}
	}
	return values
}

// wantNumbers fails for every key whose value is not the one wanted.
func wantNumbers(t *testing.T, got, want map[string]string) {
	t.Helper()
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s = %q, want %q", key, got[key], value)
		}
	}
}

// The timings are read half by half and whole, a median's interval draws
// whole tasks, and each pass gives a median of its own. Task a always takes
// 1 ms and task b 9 ms: drawn by task, a resample can hold a alone, and the
// interval reaches down to 1 ms; drawn by pass, every resample would hold both
// and give 9 ms.
func TestTimingsAreReadByTaskAndByPass(t *testing.T) {
	t.Parallel()
	ms := time.Millisecond
	timings := []timing{
		{task: "a", pass: 0, examine: ms / 4, judge: 3 * ms / 4},
		{task: "b", pass: 0, examine: 4 * ms, judge: 5 * ms},
		{task: "a", pass: 1, examine: ms / 2, judge: ms / 2},
		{task: "b", pass: 1, examine: 3 * ms, judge: 6 * ms},
	}
	got := numbersOf(timingLines(timings))
	wantNumbers(t, got, map[string]string{
		"examine_median_ms": "3.00", "examine_max_ms": "4.00",
		"judge_median_ms": "5.00", "judge_max_ms": "6.00",
		"review_median_ms": "9.00", "review_max_ms": "9.00",
		"review_median_low_ms": "1.00", "review_median_high_ms": "9.00",
		"review_pass_median_least_ms": "9.00", "review_pass_median_most_ms": "9.00",
	})
}

// The solver's runs are read for their steps against the sandbox's limit and
// for their durations, and those that ended normally are counted.
func TestSolverRunsAreReadAgainstTheLimit(t *testing.T) {
	t.Parallel()
	runs := []solverRun{
		{status: solver.StatusOK, steps: 1_000, took: time.Millisecond},
		{status: solver.StatusOK, steps: 2_000, took: 2 * time.Millisecond},
		{status: solver.StatusTimeout, steps: 2_500_000, took: 4 * time.Millisecond},
	}
	wantNumbers(t, numbersOf(solverLines(runs)), map[string]string{
		"solver_runs": "3", "solver_runs_ok": "2",
		"solver_steps_median": "2000", "solver_steps_max": "2500000",
		"solver_steps_max_percent_of_limit": "10.0",
		"solver_run_median_ms":              "2.00", "solver_run_max_ms": "4.00",
	})
}

// The packages are read over every point and every turn, each turn once, and
// against the budget.
func TestPackagesAreReadOverEveryTurn(t *testing.T) {
	t.Parallel()
	packs := []packageAt{
		{turns: 2, sizes: []int{1024, 2048}, chars: []int{1000, 2000}},
		{turns: 2, sizes: []int{3072, 3072}, chars: []int{2600, 2600}},
	}
	wantNumbers(t, numbersOf(packageLines(packs)), map[string]string{
		"package_points": "2", "package_turns": "2", "packages_built": "4",
		"package_kib_min": "1.0", "package_kib_median": "3.0", "package_kib_max": "3.0",
		"package_max_percent_of_budget": "4.7",
		"package_tokens_median":         "700", "package_tokens_max": "700",
	})
}

// What one review allocates is a pass's allocations shared out over the
// reviews of the pass, not over the passes.
func TestAReviewsAllocationsAreAPassSharedOut(t *testing.T) {
	t.Parallel()
	result := testing.BenchmarkResult{N: 3, MemBytes: 3 * 447 * 1000, MemAllocs: 3 * 447 * 10}
	if got := perReview(&result, 447); got.bytes != 1000 || got.allocs != 10 || got.passes != 3 {
		t.Errorf("perReview = %+v, want 1000 bytes and 10 allocations a review over 3 passes", got)
	}
}
