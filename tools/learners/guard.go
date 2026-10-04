package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// The guard holds the service's path to the numbers its student model was
// chosen by. It runs the first children of the paper's run of four
// generators, reads each number once, with no interval, and holds it within
// two half-widths of the interval recorded with it. A change that moves the
// model further than that, such as a step without its floor or mastery by
// another rule, turns it red. On code that moved nothing it reads the
// recorded numbers to the last digit, so a band is a tolerance for changes,
// not for chance.

// guardCommandName is the word that runs the guard rather than the bench.
const guardCommandName = "guard"

// guardChildren and guardAnswers are the size of the guard's run: enough
// children for a band of a few standard errors to catch what the model was
// chosen for, and few enough to run on every change.
const (
	guardChildren = 300
	guardAnswers  = 200
)

// guardSnapshot is where the bands are kept.
var guardSnapshot = filepath.Join("testdata", "guard.csv")

// guardedMeasure is one number the guard holds: a measure on a generator.
type guardedMeasure struct {
	generator generator
	metric    string
}

// guarded are the numbers the student model was chosen by. On children who
// stay put: the error after 200 answers, the corridor, the masteries declared
// falsely, the answers until a mastery, the masteries never declared, and how
// often the rank changes late in a run. On children placed a level off: the
// error after ten answers. On children who learn: the lag and the corridor.
// On children who jump once: those not caught up.
var guarded = []guardedMeasure{
	{staticChildren, "r1_rms_200"},
	{staticChildren, "r3_inside"},
	{staticChildren, "r4_false"},
	{staticChildren, "r5_late_answers"},
	{staticChildren, "r5_never"},
	{staticChildren, "r8_rank_150_200"},
	{misplaced, "r7_error_10"},
	{learning, "r6_lag"},
	{learning, "r3_inside"},
	{jumping, "r6_jump_unsettled"},
}

// band is a guarded number as it was recorded: its value and the half-width
// of its interval.
type band struct {
	guardedMeasure
	value, halfWidth float64
}

// holds says whether a number is within two half-widths of the recorded one.
// The edge is within.
func (b *band) holds(v float64) bool { return math.Abs(v-b.value) <= 2*b.halfWidth }

// ends are the band's lower and upper ends.
func (b *band) ends() (low, high float64) { return b.value - 2*b.halfWidth, b.value + 2*b.halfWidth }

// guardReading is a guarded number as a run read it, held to its band.
type guardReading struct {
	band
	got float64
	has bool
}

// inside says whether the run read the number and it is within its band.
func (r *guardReading) inside() bool { return r.has && r.holds(r.got) }

// why says how a reading falls outside its band, naming the measure, the
// number read and the band.
func (r *guardReading) why() string {
	low, high := r.ends()
	read := "has no number"
	if r.has {
		read = number(r.got) + " is outside its band"
	}
	return fmt.Sprintf("learners guard: %s %s %s [%s, %s] (recorded %s, half-width %s)",
		r.generator, r.metric, read, number(low), number(high), number(r.value), number(r.halfWidth))
}

// guardReadings holds the numbers of a run to their bands, in the bands'
// order. valueOf gives the run's number of a guarded measure, and whether it
// has one.
func guardReadings(bands []band, valueOf func(guardedMeasure) (float64, bool)) []guardReading {
	readings := make([]guardReading, 0, len(bands))
	for _, b := range bands {
		got, has := valueOf(b.guardedMeasure)
		readings = append(readings, guardReading{band: b, got: got, has: has})
	}
	return readings
}

// valuesOf reads a run's number of a guarded measure under a rule: its value
// over every child of the rule's cell on the measure's generator, with no
// interval. A measure the run has no cell or no metric for has no number.
func valuesOf(all []cell, results [][]vector, ms []metric, r *rule) func(guardedMeasure) (float64, bool) {
	return func(m guardedMeasure) (float64, bool) {
		c, i, has := guardedCell(all, ms, r, m)
		if !has {
			return 0, false
		}
		return readerOf(i, ms)(results[c], everyone(len(results[c])))
	}
}

// guardedCell is where a run keeps a guarded number under a rule: the index
// of the rule's cell on the measure's generator, that of the measure among
// the run's metrics, and whether the run has both.
func guardedCell(all []cell, ms []metric, r *rule, m guardedMeasure) (c, i int, has bool) {
	c = slices.IndexFunc(all, func(cl cell) bool { return sameRule(cl.rule, r) && cl.generator == m.generator })
	i = slices.Index(metricNames(ms), m.metric)
	return c, i, c >= 0 && i >= 0
}

// guard is a run of the guard: the world it runs in, the rule it holds, how
// many children of each generator it runs, and where its bands are kept.
type guard struct {
	w        *world
	r        *rule
	children int
	snapshot string
}

// run gives the guard's children every answer under its rule, on each
// generator its numbers are read on.
func (g *guard) run() ([]cell, [][]vector, []metric, error) {
	all := cellsOn([]*rule{g.r}, guardedGenerators())
	ms := metrics()
	results, err := runCells(g.w, all, g.children, ms)
	return all, results, ms, err
}

// guardedGenerators are the generators the guard reads its numbers on, each
// once, in the order its measures name them.
func guardedGenerators() []generator {
	var gens []generator
	for _, m := range guarded {
		if !slices.Contains(gens, m.generator) {
			gens = append(gens, m.generator)
		}
	}
	return gens
}

// record reads every guarded number off a run with its interval, drawn as a
// run of the bench draws the interval of the same cell, and gives the bands:
// each number with half the width of its interval.
func (g *guard) record() ([]band, error) {
	all, results, ms, err := g.run()
	if err != nil {
		return nil, err
	}
	bands := make([]band, 0, len(guarded))
	for _, m := range guarded {
		c, i, has := guardedCell(all, ms, g.r, m)
		if !has {
			return nil, fmt.Errorf("learners guard: %s %s is read by no cell or metric of the run", m.generator, m.metric)
		}
		s := summarize(readerOf(i, ms), results[c], seeded(all[c].name()+"/"+m.metric, "bootstrap"))
		if !s.has {
			return nil, fmt.Errorf("learners guard: %s %s has no number to record", m.generator, m.metric)
		}
		bands = append(bands, band{guardedMeasure: m, value: s.value, halfWidth: (s.high - s.low) / 2})
	}
	return bands, nil
}

// errUnguarded is a file of bands that does not hold every guarded number, in
// the guard's order: a number with no band would not be held at all.
var errUnguarded = errors.New("learners guard: the bands kept are not the guarded numbers; record them again")

// judge reads the guarded numbers off a run, with no interval, and holds each
// to the band kept for it.
func (g *guard) judge() ([]guardReading, error) {
	bands, err := readBands(g.snapshot)
	if err != nil {
		return nil, err
	}
	if !slices.EqualFunc(bands, guarded, func(b band, m guardedMeasure) bool { return b.guardedMeasure == m }) {
		return nil, fmt.Errorf("%w, from %s", errUnguarded, g.snapshot)
	}
	all, results, ms, err := g.run()
	if err != nil {
		return nil, err
	}
	return guardReadings(bands, valuesOf(all, results, ms, g.r)), nil
}

// bandHeader is the first row of the file the bands are kept in.
var bandHeader = []string{"generator", "metric", "value", "half_width"}

// writeBands keeps the bands, a row each, in the guard's order, every number
// to its last digit: a number rounded would part from the run's own, and a
// half-width rounded to nothing would leave no room even for that.
func writeBands(path string, bands []band) error {
	rows := [][]string{bandHeader}
	exact := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	for _, b := range bands {
		rows = append(rows, []string{string(b.generator), b.metric, exact(b.value), exact(b.halfWidth)})
	}
	return writeCSV(path, rows)
}

// errBands is a file of bands the guard cannot read.
var errBands = errors.New("learners guard: the bands cannot be read")

// readBands reads the bands kept in a file.
func readBands(path string) ([]band, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errBands, err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errBands, err)
	}
	if len(rows) == 0 || !slices.Equal(rows[0], bandHeader) {
		return nil, fmt.Errorf("%w: %s does not begin with %s", errBands, path, strings.Join(bandHeader, ","))
	}
	bands := make([]band, 0, len(rows)-1)
	for _, row := range rows[1:] {
		value, errValue := strconv.ParseFloat(row[2], 64)
		halfWidth, errWidth := strconv.ParseFloat(row[3], 64)
		if errValue != nil || errWidth != nil {
			return nil, fmt.Errorf("%w: %s has a row of no numbers, %q", errBands, path, row)
		}
		bands = append(bands, band{guardedMeasure: guardedMeasure{generator(row[0]), row[1]}, value: value, halfWidth: halfWidth})
	}
	return bands, nil
}

// guardCommand runs the guard: it holds the numbers of its run to their
// bands, or, given -update, records the bands again from its run. It takes
// no flag of the bench's, since its bands mean something only on the run
// they were recorded on.
func guardCommand(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("learners guard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	update := flags.Bool("update", false, "record the bands again from the guard's run, into "+guardSnapshot)
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "learners guard: takes -update alone, and was given %q\n", flags.Args())
		return 2
	}
	w, err := newWorld(guardAnswers)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	g := guard{w: w, r: serviceRule(), children: guardChildren, snapshot: guardSnapshot}
	if *update {
		return recordInto(&g, stdout, stderr)
	}
	readings, err := g.judge()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprint(stdout, guardReport(readings))
	code := 0
	for i := range readings {
		if !readings[i].inside() {
			fmt.Fprintln(stderr, readings[i].why())
			code = 1
		}
	}
	return code
}

// recordInto records the bands from the guard's run and keeps them.
func recordInto(g *guard, stdout, stderr io.Writer) int {
	bands, err := g.record()
	if err == nil {
		err = writeBands(g.snapshot, bands)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "learners guard: %d bands recorded into %s\n", len(bands), g.snapshot)
	return 0
}

// guardReport is what the guard read, as Markdown: a line on its run, every
// number against its band, and what came out of it.
func guardReport(readings []guardReading) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# The guard of the student model\n\n")
	fmt.Fprintf(&b, "The service's path on the first %d children of each of G0, G1, G2 and G3 of the paper's run "+
		"(seed %d, experiment %s), %d answers each. Every number is read once, with no interval, "+
		"and held within two half-widths of the interval recorded with it.\n\n", guardChildren, masterSeed, experiment, guardAnswers)
	b.WriteString("| Generator | Measure | Value | Recorded | Band | |\n|---|---|---:|---:|---|---|\n")
	var outside []string
	for i := range readings {
		r := &readings[i]
		low, high := r.ends()
		got, verdict := noNumber, "inside"
		if r.has {
			got = number(r.got)
		}
		if !r.inside() {
			verdict = "**outside**"
			outside = append(outside, string(r.generator)+" "+r.metric)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s to %s | %s |\n",
			r.generator, r.metric, got, number(r.value), number(low), number(high), verdict)
	}
	if len(outside) == 0 {
		fmt.Fprintf(&b, "\nAll %d numbers are within their bands.\n", len(readings))
	} else {
		fmt.Fprintf(&b, "\n%d of %d numbers are outside their bands: %s.\n", len(outside), len(readings), strings.Join(outside, ", "))
	}
	return b.String()
}
