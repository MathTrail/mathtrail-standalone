package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// usage is the line a run of no shape is refused with.
const usage = "usage: learnersim [-out <directory>] [-children <n>] [-answers <n>]"

// errCannotSeal is the failure of a sealer that cannot seal.
var errCannotSeal = errors.New("the sealer cannot seal")

// failingSealer is a sealer whose every call fails, as one with a broken key
// would.
type failingSealer struct{}

func (failingSealer) Seal([]byte, ...string) (string, error) { return "", errCannotSeal }
func (failingSealer) Open(string, ...string) ([]byte, error) { return nil, errCannotSeal }

// pickCells are the cells of the experiment with these names, in this order.
func pickCells(t *testing.T, names ...string) []cell {
	t.Helper()
	all := cells()
	picked := make([]cell, 0, len(names))
	for _, name := range names {
		at := slices.IndexFunc(all, func(c cell) bool { return c.name() == name })
		if at < 0 {
			t.Fatalf("the experiment has no cell %s", name)
		}
		picked = append(picked, all[at])
	}
	return picked
}

// A run whose shape is refused — a count that is not a number, no children, no
// answers, nowhere to write — stops before it computes anything, and leaves no
// results directory behind that could be taken for a run's.
func TestARunOfNoShapeIsRefusedBeforeAnythingIsMade(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		code int
		says string
	}{
		{"a count that is not a number", []string{"-children", "x"}, 2, `invalid value "x" for flag -children`},
		{"no children", []string{"-children", "0"}, 1, usage},
		{"no answers", []string{"-answers", "0"}, 1, usage},
		{"no directory", []string{"-out", ""}, 1, usage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := filepath.Join(t.TempDir(), "results")
			var stdout, stderr bytes.Buffer
			if code := runCommand(append([]string{"-out", dir}, tc.args...), &stdout, &stderr); code != tc.code {
				t.Errorf("runCommand() = %d, want %d", code, tc.code)
			}
			if !strings.Contains(stderr.String(), tc.says) {
				t.Errorf("stderr = %q, want it to say %q", stderr.String(), tc.says)
			}
			if stdout.Len() > 0 {
				t.Errorf("stdout = %q, want nothing", stdout.String())
			}
			if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("os.Stat(the results directory) error = %v, want it not to exist", err)
			}
		})
	}
}

// A run too short for the primary comparisons is refused, naming what it
// lacks, before any file of its results is written: a directory holding some
// files of a run would read as a finished one.
func TestARunTooShortForItsComparisonsWritesNoResults(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "results")
	var stdout, stderr bytes.Buffer
	if code := runCommand([]string{"-out", dir, "-children", "1", "-answers", "20"}, &stdout, &stderr); code != 1 {
		t.Fatalf("runCommand() = %d, want 1; stderr %q", code, stderr.String())
	}
	for _, want := range []string{
		"learnersim: the primary comparisons name what the run lacks: ",
		// No child gives 200 answers, nor 101, so neither the error after 200
		// answers nor the lag read from the hundred and first can be read.
		"r1_rms_200 of shrinking/both/G0 against shrinking/general/G0",
		"r6_lag of shrinking/both/G2 against shrinking/general/G2",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr = %q, want it to say %q", stderr.String(), want)
		}
	}
	if stdout.Len() > 0 {
		t.Errorf("stdout = %q, want nothing", stdout.String())
	}
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) || len(entries) > 0 {
		t.Errorf("the results directory holds %d files (error %v), want none", len(entries), err)
	}
}

// A run that cannot make its results directory fails, saying so, and leaves
// the file in its way as it was.
func TestARunThatCannotMakeItsDirectoryFailsSayingSo(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "taken")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := runCommand([]string{"-out", filepath.Join(blocker, "results"), "-children", "1", "-answers", "1"}, &stdout, &stderr); code != 1 {
		t.Fatalf("runCommand() = %d, want 1; stderr %q", code, stderr.String())
	}
	if want := "learnersim: make the results directory: "; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr = %q, want it to say %q", stderr.String(), want)
	}
	if kept, err := os.ReadFile(blocker); err != nil || string(kept) != "not a directory" {
		t.Errorf("the file in the way holds %q (error %v), want it as it was", kept, err)
	}
}

// The children of the cells run in parallel, each into its own slot, and what
// each comes to depends only on its seeds: not on the worker that ran it, nor
// on the sealing key each run makes for itself.
func TestRunCellsFillsEverySlotWithItsOwnChild(t *testing.T) {
	t.Parallel()
	inParallel, err := newWorld(30)
	if err != nil {
		t.Fatal(err)
	}
	oneByOne, err := newWorld(30)
	if err != nil {
		t.Fatal(err)
	}
	picked := pickCells(t, "shrinking/both/G0", "glicko2/topics/G2", "urnings/general/G3", "floor_0.05/both/G3")
	ms := metrics()
	const children = 3
	got, err := runCells(inParallel, picked, children, ms)
	if err != nil {
		t.Fatalf("runCells() error = %v", err)
	}
	for c := range picked {
		for i := range children {
			r, failed := run(oneByOne, picked[c].rule, newChild(picked[c].generator, i, oneByOne.topics))
			if failed != nil {
				t.Fatal(failed)
			}
			if want := vectorOf(r, ms); !reflect.DeepEqual(got[c][i], want) {
				t.Errorf("%s, child %d: runCells gave %v, a run of its own %v", picked[c].name(), i, got[c][i].parts, want.parts)
			}
		}
	}
}

// A child whose task cannot be sealed fails the whole run with the cause and
// the answer it failed at. Every child fails here, and there are more of them
// than workers, so a worker that stopped at its first failure would leave the
// rest of the children with nobody to run them.
func TestRunCellsFailsWithTheCauseOfAChildsFailure(t *testing.T) {
	t.Parallel()
	w, err := newWorld(30)
	if err != nil {
		t.Fatal(err)
	}
	w.sealer = failingSealer{}
	picked := pickCells(t, "shrinking/both/G0", "glicko2/topics/G2")
	_, err = runCells(w, picked, runtime.GOMAXPROCS(0)+1, metrics())
	if !errors.Is(err, errCannotSeal) {
		t.Fatalf("runCells() error = %v, want %v", err, errCannotSeal)
	}
	if want := ", answer 1: "; !strings.Contains(err.Error(), want) {
		t.Errorf("runCells() error = %q, want it to name the answer: %q", err, want)
	}
}
