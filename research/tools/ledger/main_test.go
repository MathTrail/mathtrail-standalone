package main

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceFileKeepsPermissionsAndLeavesNothingBehind(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.md")
	if err := os.WriteFile(path, []byte("old"), 0o640); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(path, []byte("new"), 0o640); err != nil {
		t.Fatalf("replaceFile: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("content = %q, want \"new\"", content)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Errorf("permissions = %v, want -rw-r-----", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries, want only the ledger", len(entries))
	}
}

func TestReplaceFileLeavesNothingBehindWhenItFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// A directory where the file should be: the final rename cannot succeed.
	path := filepath.Join(dir, "ledger.md")
	if err := os.Mkdir(path, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := replaceFile(path, []byte("new"), 0o640); err == nil {
		t.Fatal("replaceFile renamed a file over a directory, want an error")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries after a failed write, want only what was there", len(entries))
	}
}

// A missing directory fails the write before anything is made: there is no
// half-written file for a later run to read.
func TestReplaceFileIntoAMissingDirectoryCreatesNothing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := replaceFile(filepath.Join(dir, "missing", "numbers.txt"), []byte("new"), 0o644); err == nil {
		t.Fatal("replaceFile wrote into a directory that is not there, want an error")
	}

	if got := namesIn(t, dir); len(got) != 0 {
		t.Errorf("the directory holds %q after a failed write, want nothing", got)
	}
}

// The tool rewrites only its section of a ledger a person also edits: the text
// around the markers and the file's permissions stay as they were.
func TestWriteLedgerKeepsTheModeAndTheTextAroundTheSection(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "ledger.md")
	writeFile(t, path, runBefore+"\nold\n"+runAfter, 0o640)

	if err := writeLedger(path, "product", "new\n"); err != nil {
		t.Fatalf("writeLedger: %v", err)
	}

	if got, want := readFile(t, path), runBefore+"\nnew\n"+runAfter; got != want {
		t.Errorf("ledger =\n%s\nwant\n%s", got, want)
	}
	if mode := modeOf(t, path); mode != 0o640 {
		t.Errorf("mode = %v, want -rw-r----- as it was", mode)
	}
}

// A ledger the tool cannot read, or one without its section's markers, is
// refused and left as it was: the tool never decides where a section goes.
func TestWriteLedgerRefusesALedgerItCannotRewrite(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		ledger func(t *testing.T, dir string) string // makes the ledger in dir and returns its path
		want   string
	}{
		{"a ledger that is not there", func(t *testing.T, dir string) string {
			t.Helper()
			return filepath.Join(dir, "ledger.md")
		}, "read ledger"},
		{"a directory", func(t *testing.T, dir string) string {
			t.Helper()
			return dir
		}, "read ledger"},
		{"a ledger with no markers", func(t *testing.T, dir string) string {
			t.Helper()
			path := filepath.Join(dir, "ledger.md")
			writeFile(t, path, "a ledger with no section\n", 0o640)
			return path
		}, "the ledger needs exactly one"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := tc.ledger(t, dir)
			before := snapshot(t, dir)

			err := writeLedger(path, "product", "new\n")

			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("writeLedger error = %v, want one containing %q", err, tc.want)
			}
			if after := snapshot(t, dir); !maps.Equal(after, before) {
				t.Errorf("the directory holds %v after the refusal, want %v", after, before)
			}
		})
	}
}

// The check reads the ledger's section and says whether the table the claims
// render now is what stands there, naming the ledger when it is not; a ledger
// it cannot read, or one with no section, is no ledger that is up to date.
func TestCheckLedgerComparesTheSectionWithTheTable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	current := filepath.Join(dir, "current.md")
	writeFile(t, current, runBefore+"\nthe table\n"+runAfter, 0o640)
	plain := filepath.Join(dir, "plain.md")
	writeFile(t, plain, "no section here\n", 0o640)
	for _, tc := range []struct {
		name, path, table string
		want              string // what the error says, or nothing for a section that is current
	}{
		{"a current section", current, "the table\n", ""},
		{"a stale section", current, "another table\n", current + ` is out of date: its "product" section differs`},
		{"a ledger with no markers", plain, "the table\n", "the ledger needs exactly one"},
		{"a ledger that is not there", filepath.Join(dir, "missing.md"), "the table\n", "read ledger"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := checkLedger(tc.path, "product", tc.table)

			wantError(t, "checkLedger", err, tc.want)
		})
	}
}

// The numbers are checked only when a file is named for them, and then they
// must be exactly what the claims state.
func TestCheckNumbersComparesTheFileWithTheNumbers(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "numbers.txt")
	writeFile(t, path, "pairs=33\n", 0o644)
	for _, tc := range []struct {
		name, path, numbers string
		want                string
	}{
		{"no numbers file named", "", "pairs=34\n", ""},
		{"current numbers", path, "pairs=33\n", ""},
		{"stale numbers", path, "pairs=34\n", path + " is out of date"},
		{"a numbers file that is not there", path + ".missing", "pairs=33\n", "read numbers"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := checkNumbers(tc.path, tc.numbers)

			wantError(t, "checkNumbers", err, tc.want)
		})
	}
}

// wantError fails the test when err is not what is wanted: no error for an
// empty want, and otherwise an error containing it.
func wantError(t *testing.T, call string, err error, want string) {
	t.Helper()
	switch {
	case want == "" && err != nil:
		t.Errorf("%s error = %v, want none", call, err)
	case want != "" && (err == nil || !strings.Contains(err.Error(), want)):
		t.Errorf("%s error = %v, want one containing %q", call, err, want)
	}
}

// snapshot is every regular file under a directory with what it holds.
func snapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	held := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := os.ReadFile(path)
		held[path] = string(content)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return held
}
