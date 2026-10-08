package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// acceptedInTheExperiment are the reference tasks the experiment with
// injected defects found the checks accept as they are, in the order it
// reviewed them, as its committed results record them.
func acceptedInTheExperiment(t *testing.T) []string {
	t.Helper()
	path := filepath.Join("..", "faultinject", "results", "sanity.csv")
	rows := readCSV(t, path)
	hostAt, acceptedAt := slices.Index(rows[0], "host"), slices.Index(rows[0], "accepted")
	if hostAt < 0 || acceptedAt < 0 {
		t.Fatalf("%s has the columns %v, want host and accepted among them", path, rows[0])
	}
	var hosts []string
	for _, row := range rows[1:] {
		yes, err := strconv.ParseBool(row[acceptedAt])
		if err != nil {
			t.Fatalf("%s: %s accepted = %q, want true or false", path, row[hostAt], row[acceptedAt])
		}
		if yes {
			hosts = append(hosts, row[hostAt])
		}
	}
	return hosts
}

// missingFrom are the ids of one list that another does not hold.
func missingFrom(ids, other []string) []string {
	var missing []string
	for _, id := range ids {
		if !slices.Contains(other, id) {
			missing = append(missing, id)
		}
	}
	return missing
}

// The tasks perf times are the ones the experiment with injected defects
// found the checks accept as they are, all of them and no other, and as many
// as perf's committed numbers say it timed: the paper gives the cost of
// reviewing exactly those tasks.
func TestPerfTimesTheTasksTheExperimentWithInjectedDefectsAccepts(t *testing.T) {
	t.Parallel()
	tasks, err := accepted(context.Background(), loaded(t), unhurriedRunner(t))
	if err != nil {
		t.Fatalf("accepted() error = %v", err)
	}
	got := make([]string, 0, len(tasks))
	for i := range tasks {
		got = append(got, tasks[i].id)
	}
	want := acceptedInTheExperiment(t)
	if !slices.Equal(got, want) {
		t.Errorf("perf times %d tasks, the experiment accepted %d; perf adds %v and leaves out %v",
			len(got), len(want), missingFrom(got, want), missingFrom(want, got))
	}
	if timed := numbersOf(readLines(t, committed("numbers.txt")))["tasks"]; strconv.Itoa(len(got)) != timed {
		t.Errorf("perf times %d tasks, want the %s its committed numbers say it timed", len(got), timed)
	}
}

// Perf refuses to measure when no reference task is accepted, as when the
// content holds none, rather than write numbers of nothing for the paper to
// cite.
func TestPerfRefusesToMeasureNoTask(t *testing.T) {
	t.Parallel()
	tasks, err := accepted(context.Background(), &content.Content{}, newRunner(t))
	if err == nil || err.Error() != "perf: the checks accept none of the reference tasks" || tasks != nil {
		t.Errorf("accepted() over no reference tasks = %d tasks, error = %v; want none and the refusal", len(tasks), err)
	}
}

// wantCancelledAt fails unless the error is the cancellation, met at the
// review of the task named.
func wantCancelledAt(t *testing.T, err error, id string) {
	t.Helper()
	if !errors.Is(err, context.Canceled) || !strings.HasPrefix(fmt.Sprint(err), "perf: "+id+": ") {
		t.Errorf("error = %v, want the cancellation, at perf: %s: …", err, id)
	}
}

// Each stage of a measurement whose caller has gone stops at its first
// review, with that review's error naming the task, and hands back nothing
// measured; the benchmark gives up at once rather than run on. The test is
// not run in parallel: the benchmark machinery holds one lock for the whole
// test binary, and another test's benchmark would count in the time this
// one takes to give up.
func TestACancelledReviewStopsEveryStage(t *testing.T) {
	shipped, runner := loaded(t), newRunner(t)
	first := reviewing.Hosts(shipped)[0]
	prepared, err := reviewing.Prepare(runner, reviewing.Without(shipped, first.Base.Task.Question), &first.Base)
	if err != nil {
		t.Fatalf("reviewing.Prepare(%s) error = %v", first.ID, err)
	}
	tasks := []task{{id: first.ID, prepared: prepared}}
	ctx := cancelled()

	t.Run("accepted", func(t *testing.T) {
		found, err := accepted(ctx, shipped, runner)
		wantCancelledAt(t, err, first.ID)
		if found != nil {
			t.Errorf("accepted() = %d tasks, want none", len(found))
		}
	})
	t.Run("timeReviews", func(t *testing.T) {
		timings, runs, err := timeReviews(ctx, tasks, 2)
		wantCancelledAt(t, err, first.ID)
		if timings != nil || runs != nil {
			t.Errorf("timeReviews() = %d timings and %d runs, want none", len(timings), len(runs))
		}
	})
	t.Run("reviewCost", func(t *testing.T) {
		// The bound tells a benchmark that gives up from one that runs on or
		// hangs, with room to spare for a machine the race detector slows
		// down: one that gives up takes milliseconds.
		const bound = 10 * time.Second
		start := time.Now()
		spent, err := reviewCost(ctx, tasks)
		took := time.Since(start)
		wantCancelledAt(t, err, first.ID)
		if spent != (cost{}) || took > bound {
			t.Errorf("reviewCost() = %+v after %v, want nothing, within %v", spent, took, bound)
		}
	})
}
