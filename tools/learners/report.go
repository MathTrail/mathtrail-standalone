package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// writeAll writes every file of the results into a directory: the numbers of
// every cell, the comparisons, the summary, the criterion as every rule meets
// it, and what the run was. It returns the summary, for the command to print.
func writeAll(dir string, all []cell, results [][]vector, ms []metric, d design) (string, error) {
	if err := os.MkdirAll(filepath.Clean(dir), 0o750); err != nil {
		return "", fmt.Errorf("learners: make the results directory: %w", err)
	}
	names := metricNames(ms)
	summaries := summarizeCells(all, results, ms)
	comparisons, err := primaryComparisons(all, results, ms)
	if err != nil {
		return "", err
	}
	summary, err := summaryOf(all, summaries, names, d)
	if err != nil {
		return "", err
	}
	read, resolutions := readCriterion(newCriterionRun(all, summaries, results, ms, d.children))
	writers := []struct {
		name  string
		write func(path string) error
	}{
		{"cells.csv", func(path string) error { return writeCSV(path, cellTable(all, summaries, names)) }},
		{"comparisons.csv", func(path string) error { return writeCSV(path, comparisonTable(comparisons)) }},
		{"summary.md", func(path string) error { return writeText(path, summary) }},
		{"criterion.md", func(path string) error { return writeText(path, criterionText(read, resolutions, d)) }},
		{"criterion.csv", func(path string) error { return writeCSV(path, criterionTable(read)) }},
		{"run.txt", func(path string) error { return writeText(path, d.lines()) }},
	}
	for _, w := range writers {
		if err := w.write(filepath.Join(dir, w.name)); err != nil {
			return "", fmt.Errorf("learners: write %s: %w", w.name, err)
		}
	}
	return summary, nil
}

func writeCSV(path string, rows [][]string) error {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	out := csv.NewWriter(file)
	if err := out.WriteAll(rows); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// writeText writes a text into a file made as the tables are made.
func writeText(path, text string) error {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	if _, err := file.WriteString(text); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// metricNames are the names of the metrics, and of the pooled numbers after
// them.
func metricNames(ms []metric) []string {
	names := make([]string, 0, len(ms))
	for _, m := range ms {
		names = append(names, m.name)
	}
	for _, p := range pooledMetrics() {
		names = append(names, p.name)
	}
	return names
}

// summarizeCells reads every metric off every cell it applies to, with its
// interval, many cells at once: each number draws on a stream of its own, so
// the order the cells are read in changes none of them.
func summarizeCells(all []cell, results [][]vector, ms []metric) [][]summary {
	summaries := make([][]summary, len(all))
	eachAtOnce(len(all), func(c int) { summaries[c] = summarizeCell(&all[c], results[c], ms) })
	return summaries
}

// eachAtOnce calls do for every index below n, as many at once as there are
// processors, and returns when all are done. do may write only what belongs
// to its own index.
func eachAtOnce(n int, do func(i int)) {
	indices := make(chan int)
	var wg sync.WaitGroup
	for range runtime.GOMAXPROCS(0) {
		wg.Go(func() {
			for i := range indices {
				do(i)
			}
		})
	}
	for i := range n {
		indices <- i
	}
	close(indices)
	wg.Wait()
}

// summarizeCell reads every metric off one cell's children, leaving empty the
// ones that do not apply to its rule, and the pooled numbers after them. Each
// interval draws on a stream named after its cell and metric, so that adding a
// metric leaves the others' intervals as they were.
func summarizeCell(cl *cell, children []vector, ms []metric) []summary {
	summaries := make([]summary, 0, len(ms))
	for i, m := range ms {
		var s summary
		if m.applies == nil || m.applies(cl.rule) {
			s = summarize(readerOf(i, ms), children, seeded(cl.name()+"/"+m.name, "bootstrap"))
		}
		summaries = append(summaries, s)
	}
	for _, p := range pooledMetrics() {
		var s summary
		if p.applies == nil || p.applies(cl.rule) {
			s = summarize(p.read, children, seeded(cl.name()+"/"+p.name, "bootstrap"))
		}
		summaries = append(summaries, s)
	}
	return summaries
}

func cellTable(all []cell, summaries [][]summary, names []string) [][]string {
	table := [][]string{{"rule", "structure", "generator", "metric", "value", "low", "high"}}
	for c := range all {
		for i, name := range names {
			s := summaries[c][i]
			if !s.has {
				continue
			}
			table = append(table, []string{
				all[c].rule.name, string(all[c].rule.shape), string(all[c].generator), name,
				number(s.value), number(s.low), number(s.high),
			})
		}
	}
	return table
}

func number(v float64) string { return strconv.FormatFloat(v, 'f', 4, 64) }

// comparison is a primary comparison: one metric, the first cell less the
// second, over the same children.
type comparison struct {
	group, metric, first, second string
	result                       summary
}

// primaryComparisons are the comparisons every run makes, in four groups: the
// service against every other rule but the ceiling on children who stay put
// (1) and on children who learn (2), the service against no trial series on
// children placed far off (3), and the floor under the step against the
// service on children who jump and on children who stay put (4). A cell or a
// metric it names that the run does not have, or a comparison the run gives no
// values for, is an error, not a comparison quietly left out.
func primaryComparisons(all []cell, results [][]vector, ms []metric) ([]comparison, error) {
	cp := &comparer{results: results, ms: ms, cells: map[string]int{}, metrics: map[string]int{}}
	for c := range all {
		cp.cells[all[c].name()] = c
	}
	for i, name := range metricNames(ms) {
		cp.metrics[name] = i
	}
	service := func(g generator) string { return "shrinking/both/" + string(g) }
	for _, r := range rules() {
		if r.service || r.ceiling {
			continue
		}
		cp.add("1", service(staticChildren), r.name+"/"+string(r.shape)+"/"+string(staticChildren), "r1_rms_200", "r3_inside", "r4_false")
		cp.add("2", service(learning), r.name+"/"+string(r.shape)+"/"+string(learning), "r6_lag", "r3_inside")
	}
	cp.add("3", service(misplaced), "no_trial/both/"+string(misplaced), "r7_error_5", "r7_error_10", "r7_longest_wrong", "r7_hard_first")
	cp.add("4", "floor_0.05/both/"+string(jumping), service(jumping), "r6_jump_answers")
	cp.add("4", "floor_0.05/both/"+string(staticChildren), service(staticChildren), "r1_rms_200")
	if len(cp.missing) > 0 {
		return nil, fmt.Errorf("learners: the primary comparisons name what the run lacks: %s", strings.Join(cp.missing, "; "))
	}
	return cp.out, nil
}

// comparer makes the comparisons of a run, and keeps what they named that the
// run lacks.
type comparer struct {
	results        [][]vector
	ms             []metric
	cells, metrics map[string]int
	out            []comparison
	missing        []string
}

// add compares two cells, the first less the second, on each of the metrics.
func (cp *comparer) add(group, first, second string, metrics ...string) {
	a, okA := cp.cells[first]
	b, okB := cp.cells[second]
	if !okA || !okB {
		cp.missing = append(cp.missing, first+" against "+second)
		return
	}
	for _, name := range metrics {
		i, known := cp.metrics[name]
		if !known {
			cp.missing = append(cp.missing, "the metric "+name)
			continue
		}
		result := compare(readerOf(i, cp.ms), cp.results[a], cp.results[b], seeded(group+"/"+first+"/"+second+"/"+name, "bootstrap"))
		if !result.has {
			cp.missing = append(cp.missing, name+" of "+first+" against "+second)
			continue
		}
		cp.out = append(cp.out, comparison{group: group, metric: name, first: first, second: second, result: result})
	}
}

func comparisonTable(comparisons []comparison) [][]string {
	table := [][]string{{"comparison", "metric", "first", "second", "difference", "low", "high"}}
	for _, c := range comparisons {
		table = append(table, []string{c.group, c.metric, c.first, c.second, number(c.result.value), number(c.result.low), number(c.result.high)})
	}
	return table
}
