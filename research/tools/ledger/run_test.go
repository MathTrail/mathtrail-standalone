package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The claims, facts and files a run of the command reads: a claim proved by a
// line of a file, which states a number, and a claim proved by a fact.
const (
	runClaims = `[
		{"id":"C001","group":"Model","claim":"The catalog has 17 topics","status":"built","evidence":[{"file":"catalog.go","match":"^const Topics"}],"numbers":{"catalog_topics":"17"},"used_in":"A"},
		{"id":"C002","group":"Checks","claim":"The stats count 17 topics","status":"measured","evidence":[{"stat":"topics","expect":"17"}],"used_in":"A"}
	]`
	runStats = "repository=github.com/example/product\ncommit=e1c315303bf6\ncommit_date=2026-09-25\ntopics=17\n"
	// runBefore and runAfter stand around the section the command rewrites.
	runBefore = "# Ledger\n\nWhat the claims rest on.\n<!-- ledger:product:begin -->"
	runAfter  = "<!-- ledger:product:end -->\n\nWritten by hand after the table.\n"
)

// runFiles are the files at the pinned commit, with the line that proves C001.
var runFiles = map[string]string{"catalog.go": "package catalog\n\nconst Topics = 17\n"}

// ledgerRun is one fixture of the command: its files on disk, and the
// repositories and commits the command asked to read the proofs at.
type ledgerRun struct {
	claims, stats, ledger, numbers string
	shown                          [][2]string
}

func newLedgerRun(t *testing.T) *ledgerRun {
	t.Helper()
	dir := t.TempDir()
	r := &ledgerRun{
		claims:  filepath.Join(dir, "claims.json"),
		stats:   filepath.Join(dir, "stats.txt"),
		ledger:  filepath.Join(dir, "ledger.md"),
		numbers: filepath.Join(dir, "numbers.txt"),
	}
	writeFile(t, r.claims, runClaims, 0o600)
	writeFile(t, r.stats, runStats, 0o600)
	writeFile(t, r.ledger, runBefore+"\nan old table\n"+runAfter, 0o640)
	return r
}

// args are the arguments that name every file of the fixture, with the
// repository at /clone, followed by the ones given.
func (r *ledgerRun) args(more ...string) []string {
	return append([]string{"-claims", r.claims, "-stats", r.stats, "-ledger", r.ledger, "-section", "product", "-numbers", r.numbers, "-repo", "/clone"}, more...)
}

// show stands in for reading a repository: it serves the files given,
// whatever the commit, and notes what it was asked for.
func (r *ledgerRun) show(contents map[string]string) func(repo, commit string) Source {
	return func(repo, commit string) Source {
		r.shown = append(r.shown, [2]string{repo, commit})
		return files(contents)
	}
}

// run runs the command against the files given and returns its exit status and
// what it printed.
func (r *ledgerRun) run(t *testing.T, contents map[string]string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errs bytes.Buffer
	code = run(args, r.show(contents), &out, &errs)
	return code, out.String(), errs.String()
}

func writeFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatal(err)
	}
	// The process's umask must not decide the mode a test starts from.
	if err := os.Chmod(path, perm); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

// namesIn are the names of a directory's entries, in order.
func namesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names
}

func modeOf(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

// rendered is what the fixture's claims render at its commit: the table and
// the numbers.
func rendered(t *testing.T) (table, numbers string) {
	t.Helper()
	claims, err := ParseClaims([]byte(runClaims))
	if err != nil {
		t.Fatal(err)
	}
	stats, err := ParseStats(runStats)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := Resolve(claims, files(runFiles), stats)
	if err != nil {
		t.Fatal(err)
	}
	return Render(rows, stats), RenderNumbers(rows, stats)
}

// The command reads the proofs at the commit the stats name, from the
// repository it is given, and rewrites only the ledger's section: the text a
// person wrote around it stays as it was, byte for byte. The numbers the paper
// reads come out beside it, and a check right after finds both up to date.
func TestRunWritesTheSectionAndTheNumbersThenFindsThemUpToDate(t *testing.T) {
	t.Parallel()
	r := newLedgerRun(t)
	table, numbers := rendered(t)

	code, stdout, stderr := r.run(t, runFiles, r.args()...)

	if code != 0 {
		t.Fatalf("run() = %d, want 0; stderr: %s", code, stderr)
	}
	if want := [][2]string{{"/clone", "e1c315303bf6"}}; !slices.Equal(r.shown, want) {
		t.Errorf("proofs read at %v, want %v", r.shown, want)
	}
	ledger := readFile(t, r.ledger)
	if section, err := Section(ledger, "product"); err != nil || section != table {
		t.Errorf("section = %q, %v; want the rendered table %q", section, err, table)
	}
	if !strings.HasPrefix(ledger, runBefore) || !strings.HasSuffix(ledger, runAfter) {
		t.Errorf("ledger =\n%s\nwant the text around the section kept as it was", ledger)
	}
	if mode := modeOf(t, r.ledger); mode != 0o640 {
		t.Errorf("ledger mode = %v, want -rw-r----- as it was", mode)
	}
	if got := readFile(t, r.numbers); got != numbers {
		t.Errorf("numbers =\n%s\nwant\n%s", got, numbers)
	}
	if mode := modeOf(t, r.numbers); mode != 0o644 {
		t.Errorf("numbers mode = %v, want -rw-r--r--", mode)
	}
	if want := "2 claims written to " + r.ledger + "\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	if got, want := namesIn(t, filepath.Dir(r.ledger)), []string{"claims.json", "ledger.md", "numbers.txt", "stats.txt"}; !slices.Equal(got, want) {
		t.Errorf("the directory holds %q, want %q and nothing left behind", got, want)
	}

	code, stdout, stderr = r.run(t, runFiles, r.args("-check")...)

	if code != 0 {
		t.Fatalf("run(-check) = %d, want 0; stderr: %s", code, stderr)
	}
	if want := "2 claims, every proof found at e1c315303bf6, and " + r.ledger + " is up to date\n"; stdout != want {
		t.Errorf("stdout of the check = %q, want %q", stdout, want)
	}
}

// A check fails on a section or a numbers file that no longer says what the
// claims render, names the file to regenerate, and writes neither: a check
// that rewrote what it found stale would pass the second time it ran.
func TestRunCheckRefusesWhatIsStaleAndWritesNothing(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		stale func(t *testing.T, r *ledgerRun) string // makes one file stale and returns its path
	}{
		{"a stale section", func(t *testing.T, r *ledgerRun) string {
			t.Helper()
			writeFile(t, r.ledger, runBefore+"\nan old table\n"+runAfter, 0o640)
			return r.ledger
		}},
		{"stale numbers", func(t *testing.T, r *ledgerRun) string {
			t.Helper()
			writeFile(t, r.numbers, "catalog_topics=16\n", 0o644)
			return r.numbers
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newLedgerRun(t)
			if code, _, stderr := r.run(t, runFiles, r.args()...); code != 0 {
				t.Fatalf("run() = %d, want 0; stderr: %s", code, stderr)
			}
			stale := tc.stale(t, r)
			ledger, numbers := readFile(t, r.ledger), readFile(t, r.numbers)

			code, _, stderr := r.run(t, runFiles, r.args("-check")...)

			if code != 1 || !strings.Contains(stderr, stale+" is out of date") {
				t.Errorf("run(-check) = %d, stderr %q; want 1 naming %s", code, stderr, stale)
			}
			if readFile(t, r.ledger) != ledger || readFile(t, r.numbers) != numbers {
				t.Error("a check changed the ledger or the numbers")
			}
		})
	}
}

// A claim whose line is gone from the commit stops the run before anything is
// written: a table with the claim's old proof would vouch for a line that is
// no longer there.
func TestRunRefusesAProofThatIsGoneAndWritesNothing(t *testing.T) {
	t.Parallel()
	r := newLedgerRun(t)
	before := readFile(t, r.ledger)

	code, _, stderr := r.run(t, map[string]string{"catalog.go": "package catalog\n"}, r.args()...)

	if code != 1 || !strings.Contains(stderr, "ledger: C001:") {
		t.Errorf("run() = %d, stderr %q; want 1 naming C001", code, stderr)
	}
	if readFile(t, r.ledger) != before {
		t.Error("the ledger changed though a proof was gone")
	}
	if _, err := os.Stat(r.numbers); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("numbers file: %v, want none written", err)
	}
}

// The claims, the stats, the ledger and its section must each be named: a run
// that guessed one would rewrite or check the wrong thing. The refusal comes
// before any repository is read.
func TestRunNeedsEveryRequiredFlag(t *testing.T) {
	t.Parallel()
	for _, omitted := range []string{"-claims", "-stats", "-ledger", "-section"} {
		t.Run(omitted, func(t *testing.T) {
			t.Parallel()
			r := newLedgerRun(t)
			args := r.args()
			i := slices.Index(args, omitted)
			args = slices.Delete(args, i, i+2)

			code, _, stderr := r.run(t, runFiles, args...)

			if code != 1 || !strings.Contains(stderr, "-claims, -stats, -ledger and -section are all required") {
				t.Errorf("run() without %s = %d, stderr %q; want 1 and the required flags named", omitted, code, stderr)
			}
			if len(r.shown) != 0 {
				t.Errorf("proofs read at %v, want none", r.shown)
			}
		})
	}
}

// A flag the command does not know is a mistake in the call, and its exit
// status tells it from a claim that failed; asking for the usage is no failure
// at all.
func TestRunTellsABadFlagFromAFailureAndFromHelp(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		args []string
		code int
		want string // what stderr says
	}{
		{"an unknown flag", []string{"-verbose"}, 2, "flag provided but not defined: -verbose"},
		{"-h", []string{"-h"}, 0, "-claims"},
		{"-help", []string{"-help"}, 0, "-claims"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newLedgerRun(t)

			code, stdout, stderr := r.run(t, runFiles, tc.args...)

			if code != tc.code || !strings.Contains(stderr, tc.want) {
				t.Errorf("run(%q) = %d, stderr %q; want %d and %q", tc.args, code, stderr, tc.code, tc.want)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want nothing", stdout)
			}
		})
	}
}

// A file the run needs and cannot read stops it, naming what it was reading.
func TestRunStopsAtAFileItCannotRead(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		missing func(r *ledgerRun) string
		want    string
	}{
		{"the claims", func(r *ledgerRun) string { return r.claims }, "read claims"},
		{"the stats", func(r *ledgerRun) string { return r.stats }, "read stats"},
		{"the ledger", func(r *ledgerRun) string { return r.ledger }, "read ledger"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newLedgerRun(t)
			if err := os.Remove(tc.missing(r)); err != nil {
				t.Fatal(err)
			}

			code, _, stderr := r.run(t, runFiles, r.args()...)

			if code != 1 || !strings.Contains(stderr, tc.want) {
				t.Errorf("run() = %d, stderr %q; want 1 and %q", code, stderr, tc.want)
			}
		})
	}
}

// Numbers that cannot be written fail the run, which names them: the paper
// would otherwise print numbers older than the table that proves them.
func TestRunFailsWhenTheNumbersCannotBeWritten(t *testing.T) {
	t.Parallel()
	r := newLedgerRun(t)
	r.numbers = filepath.Join(filepath.Dir(r.numbers), "missing", "numbers.txt")

	code, _, stderr := r.run(t, runFiles, r.args()...)

	if code != 1 || !strings.Contains(stderr, "ledger: write numbers:") {
		t.Errorf("run() = %d, stderr %q; want 1 and the numbers named", code, stderr)
	}
}
