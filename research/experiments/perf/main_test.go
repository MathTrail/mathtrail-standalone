package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain has the benchmark of a review count its passes as the command has
// it count them. In a test binary a benchmark otherwise runs for a second, so
// what a review allocates would be averaged over however many passes fit in
// it, and the count the numbers record would not be the command's.
func TestMain(m *testing.M) {
	if err := flag.Set("test.benchtime", benchmarkPasses); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// wantNeverMade fails if there is anything at a path.
func wantNeverMade(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Lstat(%s) error = %v, want it never made", path, err)
	}
}

// Flags the command cannot read stop it before it measures anything, with
// the exit code of a wrong command line, and leave the directory it was
// given as it was.
func TestTheCommandRefusesFlagsItCannotRead(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "out")
	var stdout, stderr bytes.Buffer
	args := []string{"-out", out, "-passes", "x"}
	if code := run(args, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), `invalid value "x" for flag -passes`) || stdout.Len() != 0 {
		t.Errorf("run(%q) = %d, printing %q and %q; want 2, nothing on stdout and the flag's error on stderr", args, code, stdout.String(), stderr.String())
	}
	wantNeverMade(t, out)
}

// A command with nowhere to write fails with the usage, and the exit code of
// a measurement that did not happen.
func TestTheCommandWithNowhereToWriteFails(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	args := []string{"-out", ""}
	if code := run(args, &stdout, &stderr); code != 1 || !strings.HasPrefix(stderr.String(), "usage: perf ") || stdout.Len() != 0 {
		t.Errorf("run(%q) = %d, printing %q and %q; want 1, nothing on stdout and the usage on stderr", args, code, stdout.String(), stderr.String())
	}
}

// A measurement needs a directory to go to and at least one pass, and says
// so before it reviews anything or makes the directory. The context is
// cancelled, so that a measurement that went ahead regardless fails with
// the cancellation instead of the usage.
func TestAMeasurementNeedsADirectoryAndAPass(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "out")
	// The cases run one after the other, so that the directory is looked
	// for once both have been refused.
	for _, tc := range []struct {
		name   string
		out    string
		passes int
	}{{"no directory", "", 1}, {"no pass", out, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			err := measureInto(cancelled(), tc.out, tc.passes, &stdout)
			if err == nil || !strings.HasPrefix(err.Error(), "usage: perf ") || !strings.Contains(err.Error(), "at least 1") || stdout.Len() != 0 {
				t.Errorf("measureInto(%q, %d) error = %v, printing %q; want the usage, which asks for at least 1 pass, and nothing printed", tc.out, tc.passes, err, stdout.String())
			}
		})
	}
	wantNeverMade(t, out)
}

// A measurement whose caller has gone stops with the cancellation before it
// writes anything: no directory of results is made that a reader could take
// for a measurement.
func TestACancelledMeasurementMakesNoDirectory(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "out")
	var stdout bytes.Buffer
	if err := measureInto(cancelled(), out, 1, &stdout); !errors.Is(err, context.Canceled) || stdout.Len() != 0 {
		t.Errorf("measureInto() error = %v, printing %q; want the cancellation and nothing printed", err, stdout.String())
	}
	wantNeverMade(t, out)
}
