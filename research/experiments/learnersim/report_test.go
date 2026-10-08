package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// numbersKey is what a key of a numbers file may be: a numbers file is read
// by key, and its reader refuses a key with anything but lower-case letters,
// digits and underscores, or one given twice.
var numbersKey = regexp.MustCompile(`^[a-z0-9_]+$`)

// committed reads a table of the committed results.
func committed(t *testing.T, name string) [][]string {
	t.Helper()
	text, err := os.ReadFile(filepath.Join("results", name))
	if err != nil {
		t.Fatalf("read the committed %s: %v", name, err)
	}
	rows, err := csv.NewReader(bytes.NewReader(text)).ReadAll()
	if err != nil {
		t.Fatalf("read the committed %s as CSV: %v", name, err)
	}
	return rows
}

// everyMetricRead gives each child of every cell a value for every metric,
// over a weight of one, an answer in a calibration bin and an eligible
// attempt, so that every number of a run can be read off it.
func everyMetricRead(all []cell, ms []metric, children int) [][]vector {
	results := make([][]vector, len(all))
	for c := range all {
		for i := range children {
			v := vector{parts: make([]part, len(ms))}
			for m := range ms {
				v.parts[m] = part{a: float64(c+i+m) / 100, b: 1}
			}
			v.bins[0] = bin{predicted: 0.05, n: 1}
			v.attempts[0] = atAttempt{atRisk: 1}
			results[c] = append(results[c], v)
		}
	}
	return results
}

// checkKeys fails the test at a line that is not key=value, at a key the
// paper's macros would refuse, and at a key given twice, by the rule the
// macros read a file of numbers by.
func checkKeys(t *testing.T, lines []string) {
	t.Helper()
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		key, _, found := strings.Cut(line, "=")
		switch {
		case !found || !numbersKey.MatchString(key):
			t.Errorf("line %q has no key a numbers file takes", line)
		case seen[key]:
			t.Errorf("the key %s is given twice", key)
		}
		seen[key] = true
	}
}

// Every number a run writes into its numbers file — its design, every metric
// of every cell with both ends of its interval, the primary comparisons and
// the chain — is written under a key of its own that a numbers file takes,
// three lines to a number of a cell.
func TestEveryNumberOfARunHasAKeyOfItsOwn(t *testing.T) {
	t.Parallel()
	all, ms := cells(), metrics()
	names := metricNames(ms)
	summaries := make([][]summary, len(all))
	for c := range all {
		summaries[c] = make([]summary, len(names))
		for i := range names {
			summaries[c][i] = summary{value: 0.5, low: 0.25, high: 0.75, has: true}
		}
	}
	cellLines := cellNumbers(all, summaries, names)
	if got, want := len(cellLines), 3*len(all)*len(names); got != want {
		t.Errorf("%d lines of the cells' numbers, want %d: three for each of %d metrics of %d cells", got, want, len(names), len(all))
	}
	comparisons, err := primaryComparisons(all, everyMetricRead(all, ms, 1), ms)
	if err != nil {
		t.Fatal(err)
	}
	checkKeys(t, slices.Concat(designNumbers(1, 1), cellLines, comparisonNumbers(comparisons), chainNumbers(chainGrid())))
}

// A number a cell does not have — a metric that means nothing under its rule,
// or one its children give no value for — is left out of the table of cells
// and out of the numbers file alike, rather than written as zero.
func TestANumberACellLacksIsWrittenNowhere(t *testing.T) {
	t.Parallel()
	all := pickCells(t, "shrinking/both/G0", "urnings/topics/G1")
	names := []string{"r3_inside", "r4_false"}
	summaries := [][]summary{
		{{value: 0.42, low: 0.4, high: 0.44, has: true}, {}},
		{{}, {value: 0.25, low: 0.2, high: 0.3, has: true}},
	}
	wantTable := [][]string{
		committed(t, "cells.csv")[0],
		{"shrinking", "both", "G0", "r3_inside", "0.4200", "0.4000", "0.4400"},
		{"urnings", "topics", "G1", "r4_false", "0.2500", "0.2000", "0.3000"},
	}
	if got := cellTable(all, summaries, names); !reflect.DeepEqual(got, wantTable) {
		t.Errorf("cellTable() = %q, want %q", got, wantTable)
	}
	wantLines := []string{
		"shrinking_both_g0_r3_inside=0.4200", "shrinking_both_g0_r3_inside_low=0.4000", "shrinking_both_g0_r3_inside_high=0.4400",
		"urnings_topics_g1_r4_false=0.2500", "urnings_topics_g1_r4_false_low=0.2000", "urnings_topics_g1_r4_false_high=0.3000",
	}
	if got := cellNumbers(all, summaries, names); !slices.Equal(got, wantLines) {
		t.Errorf("cellNumbers() = %q, want %q", got, wantLines)
	}
}

// The primary comparisons are the committed ones, row for row, whatever the
// number of children: they are fixed before any child is run, so a run of any
// size compares what the committed run compared, each under a key of its own.
func TestThePrimaryComparisonsAreTheCommittedOnes(t *testing.T) {
	t.Parallel()
	all, ms := cells(), metrics()
	want := committed(t, "comparisons.csv")
	for _, children := range []int{1, 2} {
		t.Run(fmt.Sprintf("%d children", children), func(t *testing.T) {
			t.Parallel()
			comparisons, err := primaryComparisons(all, everyMetricRead(all, ms, children), ms)
			if err != nil {
				t.Fatalf("primaryComparisons() error = %v", err)
			}
			wantCommittedComparisons(t, comparisonTable(comparisons), want)
			lines := comparisonNumbers(comparisons)
			if got, wantLines := len(lines), 3*len(comparisons)+2; got != wantLines {
				t.Errorf("%d lines of comparisons, want %d", got, wantLines)
			}
			checkKeys(t, lines)
		})
	}
}

// wantCommittedComparisons fails unless a table of comparisons has the
// committed header and compares what the committed one compares, row for row:
// the same group, metric and cells, whatever the differences found.
func wantCommittedComparisons(t *testing.T, table, want [][]string) {
	t.Helper()
	if len(table) != len(want) {
		t.Fatalf("%d rows, want the committed %d", len(table), len(want))
	}
	if !slices.Equal(table[0], want[0]) {
		t.Errorf("header %q, want the committed %q", table[0], want[0])
	}
	for i, row := range table[1:] {
		if !slices.Equal(row[:4], want[i+1][:4]) {
			t.Errorf("row %d: %q, want the committed %q", i+1, row[:4], want[i+1][:4])
		}
	}
}

// A primary comparison of a metric the run does not measure fails the run,
// naming the metric, rather than being left out of the comparisons.
func TestAComparisonOfAMetricTheRunDoesNotMeasureFails(t *testing.T) {
	t.Parallel()
	all := cells()
	ms := slices.DeleteFunc(metrics(), func(m metric) bool { return m.name == "r6_lag" })
	_, err := primaryComparisons(all, everyMetricRead(all, ms, 1), ms)
	if want := "the metric r6_lag"; err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("primaryComparisons() without r6_lag error = %v, want one naming %q", err, want)
	}
}

// The figure of the sweep is drawn from the share of tasks in the corridor and
// the share of false declarations; a run that lacks either cannot draw it and
// says what it has instead, and one that has both draws the committed rows.
func TestTheCorridorTableNeedsItsMetricsAndHoldsTheCommittedRows(t *testing.T) {
	t.Parallel()
	all, names := cells(), metricNames(metrics())
	summaries := figureSummaries(all, names)
	for _, lacking := range []string{"r3_inside", "r4_false"} {
		t.Run("without "+lacking, func(t *testing.T) {
			t.Parallel()
			without := slices.DeleteFunc(slices.Clone(names), func(name string) bool { return name == lacking })
			_, err := corridorTable(all, summaries, without)
			if want := "the figure needs r3_inside and r4_false"; err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("corridorTable() error = %v, want %q", err, want)
			}
		})
	}
	table, err := corridorTable(all, summaries, names)
	if err != nil {
		t.Fatal(err)
	}
	want := committed(t, "corridor.csv")
	if !slices.Equal(table[0], want[0]) {
		t.Errorf("header %q, want the committed %q", table[0], want[0])
	}
	rowsOf := func(table [][]string) []string {
		var rows []string
		for _, row := range table[1:] {
			rows = append(rows, row[0]+" / "+row[1])
		}
		slices.Sort(rows)
		return rows
	}
	if got, wantRows := rowsOf(table), rowsOf(want); !slices.Equal(got, wantRows) {
		t.Errorf("rows %q, want the committed %q", got, wantRows)
	}
}

// A rule the figure has no name for keeps its own, with its structure in the
// figure's words.
func TestARuleWithoutAFigureNameKeepsItsOwn(t *testing.T) {
	t.Parallel()
	if got, want := ruleLabel(&rule{name: "floor_0.05", shape: both}), "floor_0.05 (level + offsets)"; got != want {
		t.Errorf("ruleLabel() = %q, want %q", got, want)
	}
}

// The chain of mastery is arithmetic alone, so the table and the numbers this
// code writes of it are the committed ones to the byte: a change to the chain
// or to how it is written shows here before it reaches the paper.
func TestTheChainIsWrittenAsCommitted(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "chain.csv")
	grid := chainGrid()
	if err := writeCSV(path, chainTable(grid)); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join("results", "chain.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		gotLines, wantLines := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
		for i := range min(len(gotLines), len(wantLines)) {
			if gotLines[i] != wantLines[i] {
				t.Fatalf("chain.csv line %d is %q, the committed %q", i+1, gotLines[i], wantLines[i])
			}
		}
		t.Fatalf("chain.csv has %d lines, the committed %d", len(gotLines), len(wantLines))
	}
	numbers, err := os.ReadFile(filepath.Join("results", "numbers.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if lines := chainNumbers(grid); !strings.Contains(string(numbers), "\n"+strings.Join(lines, "\n")+"\n") {
		t.Errorf("the committed numbers lack the chain's %d lines in order, from %q", len(lines), lines[0])
	}
}

// Mastery takes three qualifying answers in a row: within three attempts it
// comes only from three moves up in a row, in fewer it never comes, and once
// it has come it stays, so the chance never falls as attempts are added.
func TestTheChainNeedsThreeInARowAndKeepsWhatItReached(t *testing.T) {
	t.Parallel()
	rng := seeded("chain", "test")
	for range 1000 {
		p, h := rng.Float64(), rng.Float64()/2
		up := p * (1 - h)
		if got, want := chainChance(p, h, 3), up*up*up; math.Abs(got-want) > 1e-15 {
			t.Fatalf("p %v, h %v: the chance within three attempts is %v, want (p(1 − h))³ = %v", p, h, got, want)
		}
		previous := 0.0
		for m := range 60 {
			chance := chainChance(p, h, m)
			if m < streakNeeded && chance != 0 {
				t.Fatalf("p %v, h %v: the chance within %d attempts is %v, want 0", p, h, m, chance)
			}
			if chance < previous {
				t.Fatalf("p %v, h %v: the chance falls from %v to %v at %d attempts", p, h, previous, chance, m)
			}
			previous = chance
		}
	}
}

// A table is written as CSV that reads back as the same rows, whatever commas,
// quotation marks or line breaks its cells hold, and a numbers file as the
// lines it was given, each ended.
func TestTablesAndLinesReadBackAsTheyWereWritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rows := [][]string{{"label", "value"}, {"Glicko-2, with floor", `the "service"`}, {"two\nlines", ""}}
	table := filepath.Join(dir, "table.csv")
	if err := writeCSV(table, rows); err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(table)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := csv.NewReader(bytes.NewReader(text)).ReadAll(); err != nil || !reflect.DeepEqual(got, rows) {
		t.Errorf("the table reads back as %q (error %v), want %q", got, err, rows)
	}
	numbers := filepath.Join(dir, "numbers.txt")
	if err := writeLines(numbers, []string{"a=1", "b=2"}); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(numbers); err != nil || string(got) != "a=1\nb=2\n" {
		t.Errorf("the numbers read back as %q (error %v), want %q", got, err, "a=1\nb=2\n")
	}
}

// A file whose directory is a file cannot be made: writing it is an error, and
// it leaves nothing behind, neither a file of its own nor a change to the file
// in its way.
func TestAResultThatCannotBeMadeIsAnError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	blocker := filepath.Join(dir, "taken")
	if err := os.WriteFile(blocker, []byte("kept"), 0o600); err != nil {
		t.Fatal(err)
	}
	writers := map[string]func(path string) error{
		"table.csv":   func(path string) error { return writeCSV(path, [][]string{{"a"}}) },
		"numbers.txt": func(path string) error { return writeLines(path, []string{"a=1"}) },
	}
	for name, write := range writers {
		if err := write(filepath.Join(blocker, name)); err == nil {
			t.Errorf("writing %s under a file = nil error, want one", name)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Errorf("the directory holds %d entries (error %v), want only the file in the way", len(entries), err)
	}
	if kept, err := os.ReadFile(blocker); err != nil || string(kept) != "kept" {
		t.Errorf("the file in the way holds %q (error %v), want it as it was", kept, err)
	}
}
