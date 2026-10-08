package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// readingHost is a host to read a case of: a question, a hint of its own, and
// a trap behind each of two wrong options.
func readingHost(id string) *host {
	return &host{ID: id, Base: submission{Task: checks.Task{
		Question:      "Ann has 3 coins and gets 4 more. How many coins does she have?",
		Hint:          "Count on from 3.",
		Options:       map[string]string{"A": "6", "B": "7", "C": "8", "D": "1", "E": "12"},
		CorrectAnswer: "B",
		Distractors:   map[string]checks.Distractor{"A": {Trap: "t1"}, "C": {Trap: "t3"}},
	}}}
}

// A case to read shows what a person needs to judge it: the question before
// and after, the question it was made from where that is another, the key, the
// hint and the traps where they changed, and what refused it, or that nothing
// did.
func TestACaseToReadShowsWhatChanged(t *testing.T) {
	t.Parallel()
	before := readingHost("cal-56-d2-1")
	after := before.Base.Clone()
	after.Task.Question = "Ben has 4 balls and gets 5 more. How many balls does he have?"
	after.Task.Hint = "The answer is 7."
	after.Task.Distractors["A"] = checks.Distractor{Trap: "t2"}
	made := &caseRow{
		Operator: &operator{ID: "X09b"},
		Host:     before,
		Mutant:   mutant{Sub: after, Source: "Kim has 4 balls and gets 5 more. How many balls does she have?"},
	}
	var out strings.Builder
	writeReadingCase(&out, made)
	want := "\n### X09b on cal-56-d2-1\n\n" +
		"- Before: Ann has 3 coins and gets 4 more. How many coins does she have?\n" +
		"- Made from: Kim has 4 balls and gets 5 more. How many balls does she have?\n" +
		"- After: Ben has 4 balls and gets 5 more. How many balls does he have?\n" +
		"- Key: B, \"7\"\n" +
		"- Hint after: The answer is 7.\n" +
		"- Trap of A: t1 → t2\n" +
		"- Refused by: nothing\n"
	if got := out.String(); got != want {
		t.Errorf("writeReadingCase =\n%s\nwant\n%s", got, want)
	}
}

// What did not change is not shown: a question made from itself, the hint and
// the traps as they were. The checks that refused a case are named in the
// order the review reported them.
func TestACaseToReadLeavesOutWhatDidNotChange(t *testing.T) {
	t.Parallel()
	before := readingHost("cal-56-d2-1")
	made := &caseRow{
		Operator: &operator{ID: "X06a"},
		Host:     before,
		Mutant:   mutant{Sub: before.Base.Clone(), Source: before.Base.Task.Question},
		Verdict:  refusedBy(checks.CodeReadability, checks.CodeNearDuplicate),
	}
	var out strings.Builder
	writeReadingCase(&out, made)
	want := "\n### X06a on cal-56-d2-1\n\n" +
		"- Before: Ann has 3 coins and gets 4 more. How many coins does she have?\n" +
		"- After: Ann has 3 coins and gets 4 more. How many coins does she have?\n" +
		"- Key: B, \"7\"\n" +
		"- Refused by: readability;near_duplicate\n"
	if got := out.String(); got != want {
		t.Errorf("writeReadingCase =\n%s\nwant\n%s", got, want)
	}
}

// The file to read has a section for every operator in the order it is given,
// each with the cases drawn for it in the order they were drawn.
func TestTheCasesToReadFollowTheOrderOfTheOperators(t *testing.T) {
	t.Parallel()
	caseOf := func(op, on string) *caseRow {
		h := readingHost(on)
		return &caseRow{Operator: &operator{ID: op}, Host: h, Mutant: mutant{Sub: h.Base.Clone()}}
	}
	sample := map[string][]*caseRow{
		"X01a": {caseOf("X01a", "b")},
		"X09a": {caseOf("X09a", "c"), caseOf("X09a", "a")},
	}
	path := filepath.Join(t.TempDir(), "reading.md")
	if err := writeReading(path, []string{"X09a", "X01a"}, sample); err != nil {
		t.Fatalf("writeReading: %v", err)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var headings []string
	for _, line := range strings.Split(string(written), "\n") {
		if strings.HasPrefix(line, "#") {
			headings = append(headings, line)
		}
	}
	want := []string{"# Cases of the out-of-scope operators to read", "## X09a", "### X09a on c", "### X09a on a", "## X01a", "### X01a on b"}
	if !slices.Equal(headings, want) {
		t.Errorf("writeReading headings = %q, want %q", headings, want)
	}
	if intro := "For each case: is it the defect its operator is named after? Record the verdict in reading.csv.\n"; !strings.Contains(string(written), intro) {
		t.Errorf("writeReading does not say what to record and where: %q", written)
	}
}

// A file of verdicts that is there but cannot be read is an error, not a
// reading not done yet: counting no verdicts would quietly drop what a person
// found from the numbers.
func TestVerdictsThatCannotBeReadAreAnError(t *testing.T) {
	t.Parallel()
	// A regular file where the directory of the results should be.
	notADirectory := filepath.Join(t.TempDir(), "results")
	if err := os.WriteFile(notADirectory, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readVerdicts(notADirectory, readSample()); err == nil || !strings.Contains(err.Error(), "open the verdicts") {
		t.Errorf("readVerdicts under a regular file = %v, want an error opening the verdicts", err)
	}
	// A directory where the file of verdicts should be.
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "reading.csv"), 0o750); err != nil {
		t.Fatal(err)
	}
	if _, err := readVerdicts(dir, readSample()); err == nil || !strings.Contains(err.Error(), "read the verdicts") {
		t.Errorf("readVerdicts of a directory = %v, want an error reading the verdicts", err)
	}
}
