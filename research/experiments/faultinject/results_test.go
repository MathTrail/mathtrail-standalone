package main

import (
	"encoding/csv"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// resultFiles are the files a run writes.
var resultFiles = []string{"cases.csv", "checks.csv", "classes.csv", "numbers.txt", "operators.csv", "reading.md", "sanity.csv"}

// smallRun is what a run of three operators on three hosts might find: cases
// of every kind of operator, one of them found not to be its defect.
func smallRun() (*results, []operator) {
	ops := []operator{
		{ID: "D01a", Class: "D01", Kind: mechanism, Expected: []checks.Code{checks.CodeSolverDisagrees}},
		{ID: "D02b", Class: "D02", Kind: discovery, Expected: []checks.Code{checks.CodeBadStructure}},
		{ID: "X02a", Class: "X02", Kind: outOfScope},
	}
	hosts := []*host{
		readingHost("cal-56-d2-1"), readingHost("cal-56-d2-2"), readingHost("cal-56-d3-1"),
	}
	found := &results{}
	for _, h := range hosts {
		h.Topic, h.Level = "time.calendar", rating.Grades56
		found.Sanity = append(found.Sanity, sanityRow{Host: h})
		for i := range ops {
			found.Cases = append(found.Cases, caseRow{
				Operator: &ops[i], Host: h, Mutant: mutant{Sub: h.Base.Clone()}, Valid: true,
				Verdict: refusedBy(ops[i].Expected...), Similarity: -1, Sketch: -1,
			})
		}
	}
	found.Cases[1].Valid = false
	return found, ops
}

// readCSV reads a table of the results as a CSV reader takes it, every row
// with as many fields as the first.
func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("%s does not read as CSV: %v", filepath.Base(path), err)
	}
	return rows
}

// columnSum adds up one column of a table, below its header.
func columnSum(t *testing.T, rows [][]string, name string) int {
	t.Helper()
	at := slices.Index(rows[0], name)
	if at < 0 {
		t.Fatalf("no column %s in %q", name, rows[0])
	}
	sum := 0
	for _, line := range rows[1:] {
		n, err := strconv.Atoi(line[at])
		if err != nil {
			t.Fatalf("%s %q is not a count: %v", name, line[at], err)
		}
		sum += n
	}
	return sum
}

// A run writes its results whole: the seven files and nothing else, every
// table a CSV file with a row per entry, and the counts of the tables adding
// up to the cases of the run.
func TestTheResultsAreWrittenWhole(t *testing.T) {
	t.Parallel()
	found, ops := smallRun()
	dir := filepath.Join(t.TempDir(), "results")
	if err := writeAll(dir, found, ops); err != nil {
		t.Fatalf("writeAll: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if !slices.Equal(names, resultFiles) {
		t.Fatalf("writeAll wrote %q, want %q", names, resultFiles)
	}
	tables := map[string][][]string{}
	for _, name := range resultFiles {
		if strings.HasSuffix(name, ".csv") {
			tables[name] = readCSV(t, filepath.Join(dir, name))
		}
	}
	valid := 0
	for i := range found.Cases {
		if found.Cases[i].Valid {
			valid++
		}
	}
	if got, want := len(tables["cases.csv"]), 1+len(found.Cases); got != want {
		t.Errorf("cases.csv has %d rows, want %d", got, want)
	}
	if got, want := len(tables["sanity.csv"]), 1+len(found.Sanity); got != want {
		t.Errorf("sanity.csv has %d rows, want %d", got, want)
	}
	if got, want := columnSum(t, tables["operators.csv"], "applied"), len(found.Cases); got != want {
		t.Errorf("operators.csv applied adds up to %d, want the %d cases", got, want)
	}
	if got, want := columnSum(t, tables["classes.csv"], "valid"), valid; got != want {
		t.Errorf("classes.csv valid adds up to %d, want the %d valid cases", got, want)
	}
}

// Verdicts kept from a sample that has since changed stop the run before any
// table is written, and the cases of the new sample are written first, so that
// a person can read them and give their verdicts again.
func TestVerdictsFromAnotherSampleStopTheTables(t *testing.T) {
	t.Parallel()
	found, ops := smallRun()
	dir := writeReadingFile(t, "operator,host,verdict,note\nX02a,cal-12-d1-1,real,\n")
	err := writeAll(dir, found, ops)
	if err == nil || !strings.Contains(err.Error(), "is not a case of this run's sample") {
		t.Fatalf("writeAll = %v, want the verdict outside the sample named", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "reading.md")); err != nil {
		t.Errorf("reading.md was not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sanity.csv")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("sanity.csv after the refusal: %v, want none written", err)
	}
}

// A directory that cannot be made is an error that says so.
func TestAResultsDirectoryThatCannotBeMadeIsAnError(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	found, ops := smallRun()
	if err := writeAll(filepath.Join(file, "results"), found, ops); err == nil || !strings.Contains(err.Error(), "make the results directory") {
		t.Errorf("writeAll under a regular file = %v, want an error making the results directory", err)
	}
}

// A file of the results that cannot be written is an error naming it.
func TestAResultFileThatCannotBeWrittenIsNamed(t *testing.T) {
	t.Parallel()
	for _, name := range resultFiles {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, name), 0o750); err != nil {
				t.Fatal(err)
			}
			found, ops := smallRun()
			if err := writeAll(dir, found, ops); err == nil || !strings.Contains(err.Error(), "write "+name) {
				t.Errorf("writeAll with a directory for %s = %v, want an error naming it", name, err)
			}
		})
	}
}
