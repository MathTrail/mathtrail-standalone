package main

import (
	"cmp"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// The checks against the papers: how large a reproduction runs, and what it
// must come within.
const (
	tournamentPlayers = 100
	tournamentUrn     = 100
	tournamentGames   = 2_000_000
	fixedItemAbility  = 0.8
	fixedItemSteps    = 2_000_000
	chainRuns         = 200_000
)

// writeAll writes every file of the results into a directory, the numbers file
// opening with the run's design.
func writeAll(dir string, all []cell, results [][]vector, ms []metric, design []string) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("learnersim: make the results directory: %w", err)
	}
	names := metricNames(ms)
	summaries := summarizeCells(all, results, ms)
	comparisons, err := primaryComparisons(all, results, ms)
	if err != nil {
		return err
	}
	corridor, err := corridorTable(all, summaries, names)
	if err != nil {
		return err
	}
	chain := chainGrid()
	checked := checkLines()
	numbers := slices.Concat(
		[]string{"# Numbers of the experiment with simulated learners, computed by learnersim."},
		design, cellNumbers(all, summaries, names), comparisonNumbers(comparisons), chainNumbers(chain), checked,
	)
	writers := []struct {
		name  string
		write func(path string) error
	}{
		{"cells.csv", func(path string) error { return writeCSV(path, cellTable(all, summaries, names)) }},
		{"comparisons.csv", func(path string) error { return writeCSV(path, comparisonTable(comparisons)) }},
		{"chain.csv", func(path string) error { return writeCSV(path, chainTable(chain)) }},
		{"corridor.csv", func(path string) error { return writeCSV(path, corridor) }},
		{"checks.txt", func(path string) error { return writeLines(path, checked) }},
		{"numbers.txt", func(path string) error { return writeLines(path, numbers) }},
	}
	for _, w := range writers {
		if err := w.write(filepath.Join(dir, w.name)); err != nil {
			return fmt.Errorf("learnersim: write %s: %w", w.name, err)
		}
	}
	return nil
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
// interval.
func summarizeCells(all []cell, results [][]vector, ms []metric) [][]summary {
	summaries := make([][]summary, len(all))
	for c := range all {
		summaries[c] = summarizeCell(&all[c], results[c], ms)
	}
	return summaries
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
		summaries = append(summaries, summarize(p.read, children, seeded(cl.name()+"/"+p.name, "bootstrap")))
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

// primaryComparisons are the four primary comparisons of the protocol. A cell
// or a metric it names that the run does not have, or a comparison the run
// gives no values for, is an error, not a comparison quietly left out.
func primaryComparisons(all []cell, results [][]vector, ms []metric) ([]comparison, error) {
	cp := &comparer{results: results, ms: ms, cells: map[string]int{}, metrics: map[string]int{}}
	for c := range all {
		cp.cells[all[c].name()] = c
	}
	for i, name := range metricNames(ms) {
		cp.metrics[name] = i
	}
	service := func(g generator) string { return "shrinking/both/" + string(g) }
	for _, r := range sweepRules() {
		if r.service {
			continue
		}
		cp.add("1", service(staticChildren), r.name+"/"+string(r.shape)+"/"+string(staticChildren), "r1_rms_200", "r3_inside", "r4_false")
		cp.add("2", service(learning), r.name+"/"+string(r.shape)+"/"+string(learning), "r6_lag", "r3_inside")
	}
	cp.add("3", service(misplaced), "no_trial/both/"+string(misplaced), "r7_error_5", "r7_error_10", "r7_longest_wrong", "r7_hard_first")
	for _, floor := range []string{"floor_0.01", "floor_0.02", "floor_0.05"} {
		cp.add("4", floor+"/both/"+string(jumping), service(jumping), "r6_jump_answers")
		cp.add("4", floor+"/both/"+string(staticChildren), service(staticChildren), "r1_rms_200")
	}
	if len(cp.missing) > 0 {
		return nil, fmt.Errorf("learnersim: the primary comparisons name what the run lacks: %s", strings.Join(cp.missing, "; "))
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

// corridorTable is what the paper's figure of the sweep is drawn from: for
// every rule of the sweep on children who stay put, the share of its tasks in
// the corridor and the share of its declarations of mastery that were false,
// in percent with their intervals, the rule with the most tasks in the
// corridor first. A rule without a value or an interval for either is an error
// rather than a row drawn at zero or without its bars.
func corridorTable(all []cell, summaries [][]summary, names []string) ([][]string, error) {
	inside, wrongly := slices.Index(names, "r3_inside"), slices.Index(names, "r4_false")
	if inside < 0 || wrongly < 0 {
		return nil, fmt.Errorf("learnersim: the figure needs r3_inside and r4_false, and the metrics are %v", names)
	}
	sweep := map[string]bool{}
	for _, r := range sweepRules() {
		sweep[r.name+"/"+string(r.shape)] = true
	}
	type row struct {
		label, kind     string
		inside, wrongly summary
	}
	var rows []row
	for c := range all {
		r := all[c].rule
		if all[c].generator != staticChildren || !sweep[r.name+"/"+string(r.shape)] {
			continue
		}
		if !drawable(summaries[c][inside]) || !drawable(summaries[c][wrongly]) {
			return nil, fmt.Errorf("learnersim: %s has no value or no interval for the figure", all[c].name())
		}
		kind := "baseline"
		if r.service {
			kind = "service"
		}
		rows = append(rows, row{label: ruleLabel(r), kind: kind, inside: summaries[c][inside], wrongly: summaries[c][wrongly]})
	}
	slices.SortStableFunc(rows, func(a, b row) int { return cmp.Compare(b.inside.value, a.inside.value) })
	percent := func(v float64) string { return strconv.FormatFloat(100*v, 'f', 1, 64) }
	table := [][]string{{"label", "kind", "inside", "inside_low", "inside_high", "declared_wrongly", "declared_wrongly_low", "declared_wrongly_high"}}
	for _, r := range rows {
		table = append(table, []string{
			r.label, r.kind,
			percent(r.inside.value), percent(r.inside.low), percent(r.inside.high),
			percent(r.wrongly.value), percent(r.wrongly.low), percent(r.wrongly.high),
		})
	}
	return table, nil
}

// drawable says whether a summary has a value and both ends of its interval;
// a sample too sparse for any resample to read leaves the ends out.
func drawable(s summary) bool {
	return s.has && !math.IsNaN(s.low) && !math.IsNaN(s.high)
}

// ruleLabel names a rule of the sweep as the figure shows it, with no comma or
// quotation mark, which the reader of the figure's table would take for a new
// column or keep as text; a rule the names below do not cover keeps its own.
func ruleLabel(r *rule) string {
	if r.service {
		return "the service"
	}
	names := map[string]string{
		"shrinking": "shrinking step", "constant": "constant step", "gradient": "gradient step",
		"glicko2": "Glicko-2", "glicko2_floor": "Glicko-2 with floor", "urnings": "Urnings",
	}
	shapes := map[structure]string{general: "one level", topics: "per topic", both: "level + offsets"}
	name, known := names[r.name]
	if !known {
		name = r.name
	}
	return name + " (" + shapes[r.shape] + ")"
}

func chainTable(rows []chainRow) [][]string {
	table := [][]string{{"p", "hint_share", "attempts", "chance"}}
	for _, r := range rows {
		table = append(table, []string{number(r.p), number(r.h), strconv.Itoa(r.m), number(r.chance)})
	}
	return table
}

// checkLines are the checks of the baselines against their papers and of the
// chain against a simulation of itself, one key=value line each.
func checkLines() []string {
	rating, rd, volatility := glickmanExample()
	lines := []string{
		"# Checks of the baselines against their papers, and of the chain against itself.",
		fmt.Sprintf("glickman_example_rating=%.2f", rating),
		fmt.Sprintf("glickman_example_rd=%.2f", rd),
		fmt.Sprintf("glickman_example_volatility=%.6f", volatility),
		fmt.Sprintf("urnings_tournament_players=%d", tournamentPlayers),
		fmt.Sprintf("urnings_tournament_games=%d", tournamentGames),
		fmt.Sprintf("urnings_tournament_gap=%.4f", tournamentGap(tournamentPlayers, tournamentUrn, tournamentGames, seeded("tournament", "check"))),
		fmt.Sprintf("urnings_fixed_item_gap=%.4f", fixedItemCheck()),
		fmt.Sprintf("urnings_fixed_item_variation_near_the_learner=%.4f", fixedItemNearCheck()),
	}
	for i, gap := range chainGaps() {
		lines = append(lines, fmt.Sprintf("chain_gap_%d=%.4f", i+1, gap))
	}
	return lines
}

// chainGaps are how far a simulation of the chain lands from the computed
// chance at each of the points it is checked at.
func chainGaps() []float64 {
	gaps := make([]float64, len(chainPoints))
	for i, point := range chainPoints {
		simulated := simulatedChain(point.p, point.h, point.m, chainRuns, seeded("chain/"+strconv.Itoa(i), "check"))
		gaps[i] = simulated - chainChance(point.p, point.h, point.m)
	}
	return gaps
}

// chainPoints are the three points of the grid the chain is checked at by a
// simulation of itself.
var chainPoints = []chainRow{{p: 0.5, h: 0, m: 10}, {p: 0.7, h: 0.1, m: 20}, {p: 0.85, h: 0.2, m: 5}}

// writeLines writes key=value lines into a file made as the tables are made.
func writeLines(path string, lines []string) error {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	if _, err := file.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// numberKey is a name written as a key of the numbers file.
func numberKey(name string) string {
	return strings.ToLower(strings.NewReplacer(".", "_", "/", "_", "-", "_").Replace(name))
}

// cellNumbers are every metric of every cell, with its interval, as lines of
// the numbers file.
func cellNumbers(all []cell, summaries [][]summary, names []string) []string {
	var lines []string
	for c := range all {
		for i, name := range names {
			s := summaries[c][i]
			if !s.has {
				continue
			}
			key := numberKey(all[c].name() + "_" + name)
			lines = append(lines, key+"="+number(s.value), key+"_low="+number(s.low), key+"_high="+number(s.high))
		}
	}
	return lines
}

// designNumbers are the run's design as the paper states it: how many children
// a cell draws and how many answers each gives, how many generators and rules
// of the sweep there are, the spreads that make the base population, how fast
// a child who learns grows, and the answers the lag behind such a child is
// read over.
func designNumbers(children, answers int) []string {
	constant := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	return []string{
		"children_per_cell=" + strconv.Itoa(children),
		"answers_per_child=" + strconv.Itoa(answers),
		"generators=" + strconv.Itoa(len(allGenerators)),
		"sweep_rules=" + strconv.Itoa(len(sweepRules())),
		"sweep_alternatives=" + strconv.Itoa(len(sweepRules())-1),
		"child_level_spread=" + constant(spreadAroundStart),
		"writing_error_sd=" + constant(writingError),
		"learning_per_answer=" + constant(learningPerAnswer),
		"jump_by=" + constant(jumpBy),
		"lag_window_first=" + strconv.Itoa(lagFrom+1),
		"lag_window_last=" + strconv.Itoa(answers),
	}
}

// comparisonNumbers are the primary comparisons, with their intervals, as
// lines of the numbers file, followed by how many there are and how many found
// a difference: those whose interval leaves zero out. None is corrected for
// the others, so the count is what a reader weighs chance findings against.
func comparisonNumbers(comparisons []comparison) []string {
	var lines []string
	found := 0
	for _, c := range comparisons {
		key := numberKey("pc" + c.group + "_" + c.metric + "_" + c.first + "_vs_" + c.second)
		lines = append(lines, key+"="+number(c.result.value), key+"_low="+number(c.result.low), key+"_high="+number(c.result.high))
		if c.result.low > 0 || c.result.high < 0 {
			found++
		}
	}
	return append(lines, "primary_comparisons="+strconv.Itoa(len(comparisons)), "primary_comparisons_found="+strconv.Itoa(found))
}

// chainNumbers are the computed chain's chances as lines of the numbers file,
// keyed by the true chance and the share of hints in hundredths and by the
// attempts: chain_p70_h00_m5 is a chance of 0.7, no hints and five attempts.
func chainNumbers(rows []chainRow) []string {
	lines := make([]string, 0, len(rows))
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("chain_p%02d_h%02d_m%d=%s", hundredths(r.p), hundredths(r.h), r.m, number(r.chance)))
	}
	return lines
}

func hundredths(share float64) int { return int(math.Round(share * 100)) }
