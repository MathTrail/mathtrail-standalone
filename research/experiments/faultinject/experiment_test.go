package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// operatorNamed is the operator of the experiment with this id.
func operatorNamed(t *testing.T, id string) *operator {
	t.Helper()
	ops := operators()
	at := slices.IndexFunc(ops, func(op operator) bool { return op.ID == id })
	if at < 0 {
		t.Fatalf("no operator %s", id)
	}
	return &ops[at]
}

// committedCases are the rows of the committed table of cases, by operator and
// host.
var committedCases = sync.OnceValues(func() (map[string][]string, error) {
	file, err := os.Open(filepath.Join("results", "cases.csv"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}
	operatorAt, hostAt := slices.Index(rows[0], "operator"), slices.Index(rows[0], "host")
	if operatorAt < 0 || hostAt < 0 {
		return nil, fmt.Errorf("no operator or host among the columns %q", rows[0])
	}
	byCase := make(map[string][]string, len(rows))
	for _, row := range rows[1:] {
		byCase[row[operatorAt]+" on "+row[hostAt]] = row
	}
	return byCase, nil
})

// committedCase is the committed row of one operator's case on one host.
func committedCase(t *testing.T, operatorID, hostID string) []string {
	t.Helper()
	byCase, err := committedCases()
	if err != nil {
		t.Fatalf("read the committed cases: %v", err)
	}
	row, found := byCase[operatorID+" on "+hostID]
	if !found {
		t.Fatalf("the committed cases hold no case of %s on %s", operatorID, hostID)
	}
	return row
}

// cancelled is the context of a run stopped before it began.
func cancelled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// A case run on its own is the case the whole run made of it: its row in the
// table of cases is the committed one. That holds for an operator that draws
// nothing from the other hosts, whose case is the same whatever hosts a run
// has. A case made from a question is measured against it, a repeat of the
// very question being alike in full; a case made from none has no measure.
func TestACaseRunAloneIsTheOneTheRunMade(t *testing.T) {
	t.Parallel()
	b := testBench(t)
	for _, c := range []struct {
		id                 string
		similarity, sketch float64
	}{
		{"D01a", -1, -1}, {"D07a", -1, -1}, {"D12e", -1, -1}, {"D15a", 1, 1}, {"D16d", -1, -1}, {"X03a", -1, -1},
	} {
		t.Run(c.id, func(t *testing.T) {
			t.Parallel()
			op := operatorNamed(t, c.id)
			h, _ := firstCase(t, b, op)
			made, applies, err := runCase(context.Background(), b.shipped, b.runner, b.pool, op, h)
			if err != nil || !applies {
				t.Fatalf("runCase(%s on %s) = %v, %v; want a case", c.id, h.ID, applies, err)
			}
			if made.Similarity != c.similarity || made.Sketch != c.sketch {
				t.Errorf("%s on %s measured %v and %v, want %v and %v", c.id, h.ID, made.Similarity, made.Sketch, c.similarity, c.sketch)
			}
			if got, want := caseTable([]caseRow{made})[1], committedCase(t, c.id, h.ID); !slices.Equal(got, want) {
				t.Errorf("the case of %s on %s = %q, want the committed %q", c.id, h.ID, got, want)
			}
		})
	}
}

// A run stopped by whoever asked for it stops at the case it is on, at the
// step it is at: confirming the defect with its host's solver, for an operator
// that needs that, or reviewing the case.
func TestACaseStopsWithTheRun(t *testing.T) {
	t.Parallel()
	b := testBench(t)
	for _, c := range []struct{ id, step string }{
		{"D01a", "confirm a moved key: "},
		{"D07a", "examine: "},
	} {
		t.Run(c.id, func(t *testing.T) {
			t.Parallel()
			op := operatorNamed(t, c.id)
			h, _ := firstCase(t, b, op)
			_, _, err := runCase(cancelled(), b.shipped, b.runner, b.pool, op, h)
			if !errors.Is(err, context.Canceled) || !strings.HasPrefix(err.Error(), c.step) {
				t.Errorf("runCase(%s) with the run stopped = %v, want %q and %v", c.id, err, c.step, context.Canceled)
			}
		})
	}
}

// An operator that does not apply to a host makes no case of it, and asks
// nothing of the sandbox to find that out: not even a stopped run is an error.
func TestAnOperatorThatDoesNotApplyAsksNothingOfTheSandbox(t *testing.T) {
	t.Parallel()
	b := testBench(t)
	drawingOnly := operatorNamed(t, "D16d")
	at := slices.IndexFunc(b.pool.hosts, func(h *host) bool { return !h.HasDrawing() })
	if at < 0 {
		t.Fatal("every host of the bench has a drawing, want one without")
	}
	made, applies, err := runCase(cancelled(), b.shipped, b.runner, b.pool, drawingOnly, b.pool.hosts[at])
	if err != nil || applies || made.Operator != nil {
		t.Errorf("runCase(D16d) on a host without a drawing = %+v, %v, %v; want no case and no error", made, applies, err)
	}
}

// The review of every reference task as it is, made afresh under the limits an
// experiment puts on a solver, is the one the committed results hold, byte for
// byte. Every operator is then offered the tasks the checks accept as they
// are, and only those, in the order of their ids, and its cases are the
// committed ones.
func TestTheReviewOfTheTasksAsTheyAreIsTheCommittedOne(t *testing.T) {
	t.Parallel()
	shipped, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	// Under the race detector and beside the other tests a solver runs many
	// times slower than in the experiment, so the time a run may take, and
	// wait for its slot, is no part of what is tested here; the steps it may
	// take, which decide every verdict, stay the experiment's.
	limits := reviewing.SandboxLimits
	limits.Timeout, limits.Wait = time.Minute, time.Minute
	runner, err := starlark.New(limits)
	if err != nil {
		t.Fatal(err)
	}
	// One operator of the experiment, offered every host but making its case
	// of the first alone, so that the run reviews one case beside the tasks
	// themselves.
	ops := slices.DeleteFunc(operators(), func(op operator) bool { return op.ID != "D07a" })
	if len(ops) != 1 {
		t.Fatalf("%d operators D07a, want one", len(ops))
	}
	var offered []string
	apply := ops[0].Apply
	ops[0].Apply = func(m *maker) (mutant, bool) {
		offered = append(offered, m.host.ID)
		if len(offered) > 1 {
			return mutant{}, false
		}
		return apply(m)
	}
	found, err := runExperiment(context.Background(), shipped, runner, ops)
	if err != nil {
		t.Fatalf("runExperiment: %v", err)
	}
	assertSameFile(t, sanityTable(found.Sanity), filepath.Join("results", "sanity.csv"))
	var eligible []string
	for i := range found.Sanity {
		if found.Sanity[i].eligible() {
			eligible = append(eligible, found.Sanity[i].Host.ID)
		}
	}
	if !slices.Equal(offered, eligible) {
		t.Fatalf("the operator was offered %d hosts, want the %d accepted as they are", len(offered), len(eligible))
	}
	if len(found.Cases) != 1 {
		t.Fatalf("runExperiment made %d cases, want 1", len(found.Cases))
	}
	if got, want := caseTable(found.Cases)[1], committedCase(t, "D07a", eligible[0]); !slices.Equal(got, want) {
		t.Errorf("the case of D07a on %s = %q, want the committed %q", eligible[0], got, want)
	}
}

// assertSameFile writes a table as the results are written and compares it
// with a committed file, byte for byte.
func assertSameFile(t *testing.T, table [][]string, committed string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), filepath.Base(committed))
	if err := writeCSV(path, table); err != nil {
		t.Fatalf("writeCSV: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(committed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s written afresh differs from the committed one", filepath.Base(committed))
	}
}

// An experiment without a directory to write to does not start, and says how
// it is used: the run is stopped already, so that an experiment that started
// would say so instead.
func TestAnExperimentNeedsADirectory(t *testing.T) {
	t.Parallel()
	err := experimentInto(cancelled(), "", io.Discard)
	if want := "usage: faultinject [-out <directory>]"; err == nil || err.Error() != want {
		t.Errorf("experimentInto with no directory = %v, want %q", err, want)
	}
}

// An experiment stopped before its results are in names the reference task it
// stopped at, writes nothing, not even its directory, and reports nothing done.
func TestAStoppedExperimentWritesNothing(t *testing.T) {
	t.Parallel()
	shipped, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	var stdout strings.Builder
	err = experimentInto(cancelled(), out, &stdout)
	first := "faultinject: review host " + reviewing.Hosts(shipped)[0].ID + ":"
	if !errors.Is(err, context.Canceled) || !strings.HasPrefix(err.Error(), first) {
		t.Errorf("experimentInto with the run stopped = %v, want %q and %v", err, first, context.Canceled)
	}
	if _, err := os.Stat(out); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("the directory of a stopped run: %v, want none made", err)
	}
	if stdout.Len() > 0 {
		t.Errorf("a stopped run reported %q, want nothing", stdout.String())
	}
}

// The command refuses a flag it does not know as a usage error and lists the
// flags there are; a directory left empty is a failed run, told how the
// command is used.
func TestTheCommandSaysHowItIsUsed(t *testing.T) {
	t.Parallel()
	var stderr strings.Builder
	if code := run([]string{"-nope"}, io.Discard, &stderr); code != 2 || !strings.Contains(stderr.String(), "-out") {
		t.Errorf("run(-nope) = %d with %q, want 2 and the flags listed", code, stderr.String())
	}
	stderr.Reset()
	if code := run([]string{"-out", ""}, io.Discard, &stderr); code != 1 || !strings.Contains(stderr.String(), "usage: faultinject [-out <directory>]") {
		t.Errorf("run(-out \"\") = %d with %q, want 1 and the usage", code, stderr.String())
	}
}
