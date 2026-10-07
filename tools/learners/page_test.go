package main

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/sys/cpu"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The page's table reads eleven cells: the service and the rule before it on
// the four generators its goals are read on, and the ceiling on the three it
// bounds a step on. A cell more costs the build time for nothing; a cell less
// leaves a row unread.
func TestThePageRunsTheCellsOfItsTableAndNoOthers(t *testing.T) {
	t.Parallel()
	var got []string
	for _, c := range pageCells() {
		got = append(got, c.name())
	}
	want := []string{
		"shrinking/both/G0", "shrinking/both/G0-topics1", "shrinking/both/G2-half", "shrinking/both/G3",
		"earlier/both/G0", "earlier/both/G0-topics1", "earlier/both/G2-half", "earlier/both/G3",
		"oracle/both/G0", "oracle/both/G2-half", "oracle/both/G3",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("pageCells() = %q, want %q", got, want)
	}
	for _, row := range pageRows {
		for _, r := range []string{"shrinking/both/", "earlier/both/"} {
			if !slices.Contains(got, r+string(row.generator)) {
				t.Errorf("row %s is read on %s, which the page's run has no cell %s%s for", row.id, row.generator, r, row.generator)
			}
		}
		if row.ceiling == oracleCeiling && !slices.Contains(got, "oracle/both/"+string(row.generator)) {
			t.Errorf("row %s takes its ceiling from the oracle on %s, which the page's run does not run", row.id, row.generator)
		}
	}
}

// The page holds the model to the goals it was chosen by, all of them and
// none besides: every goal of the step and of mastery once, but the screen's
// in answers 6–20, whose bound is the service's own number of the same window.
func TestThePageHoldsEveryGoalOfBothCriteria(t *testing.T) {
	t.Parallel()
	type key struct {
		group     string
		generator generator
		metric    string
	}
	early := "_" + screenWindows[0].name()
	var want []key
	add := func(group string, goals []goal) {
		for _, g := range goals {
			k := key{group: group, generator: g.generator, metric: g.metric}
			if g.screen {
				k.group = screenGroup
			}
			if !strings.HasSuffix(g.metric, early) && !slices.Contains(want, k) {
				want = append(want, k)
			}
		}
	}
	add(stepGroup, stepGoals())
	add(masteryGroup, masteryGoals())
	var got []key
	for _, row := range pageRows {
		if row.group != "" {
			got = append(got, key{group: row.group, generator: row.generator, metric: row.metric})
		}
	}
	byName := func(a, b key) int {
		return strings.Compare(a.group+string(a.generator)+a.metric, b.group+string(b.generator)+b.metric)
	}
	slices.SortFunc(got, byName)
	slices.SortFunc(want, byName)
	if !slices.Equal(got, want) {
		t.Errorf("the page's goal rows are %v, want every goal of both criteria once: %v", got, want)
	}
}

// pageNumbersBy is a run of the page's cells, every number of which is what
// numbers gives for its rule, generator and metric, read by the page's
// criterion.
func pageNumbersBy(numbers func(r *rule, g generator, metric string) summary) *criterionRun {
	ms := metrics()
	names := metricNames(ms)
	all := pageCells()
	summaries := make([][]summary, len(all))
	for c := range all {
		summaries[c] = make([]summary, len(names))
		for i, name := range names {
			summaries[c][i] = numbers(all[c].rule, all[c].generator, name)
		}
	}
	return newCriterionRun(all, summaries, nil, ms, pageChildren, pageCriterion())
}

// spread is a number with an interval around it, which the oracle has well
// off zero for every measure, the false masteries and the wait included, and
// each rule at its own distance, the lag behind the child.
func spread(r *rule, g generator, metric string) summary {
	v := 0.3 + 0.01*float64(len(metric)+len(g))
	switch {
	case r.ceiling:
		v += 0.2
	case r.name == earlierRule:
		v += 0.1
	}
	if metric == "r6_lag" {
		v = -v
	}
	return summary{value: v, low: v - 0.05, high: v + 0.05, has: true}
}

// rowsBy reads the page's rows off a run by these numbers, failing the test
// on a row it cannot read.
func rowsBy(t *testing.T, numbers func(r *rule, g generator, metric string) summary) []pageRowData {
	t.Helper()
	rows, err := rowsOf(pageNumbersBy(numbers))
	if err != nil {
		t.Fatalf("rowsOf() error = %v", err)
	}
	return rows
}

// A bound is the service's own when it moves with the service's number of the
// same measure on the same generator. There the service's mark is decided by
// how the bound is drawn, not by the service, and the page says "baseline".
func TestThePageCallsABoundItsOwnExactlyWhenTheServicesNumberMovesIt(t *testing.T) {
	t.Parallel()
	before := rowsBy(t, spread)
	for i, row := range pageRows {
		if row.group == "" {
			continue
		}
		moved := rowsBy(t, func(r *rule, g generator, metric string) summary {
			s := spread(r, g, metric)
			if r.service && g == row.generator && metric == row.metric {
				s.value, s.low, s.high = 2*s.value, 2*s.low, 2*s.high
			}
			return s
		})
		if moves := moved[i].Bound.Value != before[i].Bound.Value; moves != row.own {
			t.Errorf("row %s: the bound moves with the service's own number %v, and own is %v", row.id, moves, row.own)
		}
	}
}

// The service's mark is "baseline" on a bound of its own and a mark of the
// bench's everywhere else; the rule before it is held to every bound as a
// candidate would be.
func TestThePageMarksTheServiceAsTheBaselineWhereTheBoundIsItsOwn(t *testing.T) {
	t.Parallel()
	for i, row := range rowsBy(t, spread) {
		if pageRows[i].group == "" {
			if row.Values.Service.Mark != nil || row.Values.Earlier.Mark != nil || row.Bound != nil {
				t.Errorf("row %s is shown for context, and has a bound or a mark: %+v", row.ID, row)
			}
			continue
		}
		if service := *row.Values.Service.Mark; (service == markBaseline) != pageRows[i].own {
			t.Errorf("row %s: the service's mark is %q, and the bound's own is %v", row.ID, service, pageRows[i].own)
		}
		if earlier := *row.Values.Earlier.Mark; earlier == markBaseline || earlier == "" {
			t.Errorf("row %s: the earlier rule's mark is %q, want one of the bench's", row.ID, earlier)
		}
	}
}

// A row of a goal shows each rule's number as the bench's criterion reads it,
// the lag as its size, and the mark the criterion gives it.
func TestThePageReadsEveryGoalAsTheCriterionReadsIt(t *testing.T) {
	t.Parallel()
	cr := pageNumbersBy(spread)
	rows, err := rowsOf(cr)
	if err != nil {
		t.Fatalf("rowsOf() error = %v", err)
	}
	goals := pageGoals()
	for i := range rows {
		if pageRows[i].group == "" {
			continue
		}
		at := slices.IndexFunc(goals, func(g goal) bool { return g.generator == pageRows[i].generator && g.metric == pageRows[i].metric })
		wantReadAsTheCriterion(t, cr, &rows[i], goals[at])
	}
	lag := rows[slices.IndexFunc(rows, func(r pageRowData) bool { return r.ID == "lag" })]
	if lag.ReadAs != "size" || lag.Values.Service.Value <= 0 || lag.Values.Ceiling.Value <= 0 {
		t.Errorf("the lag reads as %s, the service's %v and the ceiling's %v, want sizes", lag.ReadAs, lag.Values.Service.Value, lag.Values.Ceiling.Value)
	}
}

// wantReadAsTheCriterion holds a goal's row to the criterion's reading of each
// rule's number: its value and interval, the bound, and the mark, which on the
// service may be the baseline instead.
func wantReadAsTheCriterion(t *testing.T, cr *criterionRun, row *pageRowData, g goal) {
	t.Helper()
	for _, tc := range []struct {
		name string
		r    *rule
		got  pageValue
	}{
		{"the service", cr.baseline, row.Values.Service},
		{"the rule before it", ruleOf(cr, earlierRule), row.Values.Earlier},
	} {
		want := cr.readGoal(tc.r, g)
		if tc.got.Value != want.value || tc.got.Low != want.low || tc.got.High != want.high || row.Bound.Value != want.bound {
			t.Errorf("row %s, %s: %v [%v, %v] against %v, want %v [%v, %v] against %v",
				row.ID, tc.name, tc.got.Value, tc.got.Low, tc.got.High, row.Bound.Value, want.value, want.low, want.high, want.bound)
		}
		if *tc.got.Mark != markBaseline && *tc.got.Mark != pageMarks[want.mark] {
			t.Errorf("row %s, %s: marked %q, want %q", row.ID, tc.name, *tc.got.Mark, pageMarks[want.mark])
		}
	}
}

// The ceiling of a step's goal is the oracle's, which knows where the child
// stands. The oracle keeps the service's mastery, so it is no ceiling of
// mastery's goals: theirs is perfection, no false mastery and no wait. The
// oracle shows no screen, so the screen's rows have none.
func TestThePageTakesTheCeilingOfTheStepFromTheOracleAndOfMasteryFromPerfection(t *testing.T) {
	t.Parallel()
	for i, row := range rowsBy(t, spread) {
		c := row.Values.Ceiling
		switch pageRows[i].ceiling {
		case perfectCeiling:
			if c == nil || c.Of != perfectCeiling || c.Value != 0 || c.Low != 0 || c.High != 0 {
				t.Errorf("row %s: ceiling %+v, want perfection, nothing at all", row.ID, c)
			}
		case oracleCeiling:
			if c == nil || c.Of != oracleCeiling || c.Value == 0 {
				t.Errorf("row %s: ceiling %+v, want the oracle's number", row.ID, c)
			}
		default:
			if c != nil {
				t.Errorf("row %s: ceiling %+v, want none", row.ID, c)
			}
		}
	}
}

// A row the run has no number for is refused, rather than written as zero.
func TestThePageRefusesARowItCannotRead(t *testing.T) {
	t.Parallel()
	for _, row := range pageRows {
		_, err := rowsOf(pageNumbersBy(func(r *rule, g generator, metric string) summary {
			if r.name == earlierRule && g == row.generator && metric == row.metric {
				return summary{}
			}
			return spread(r, g, metric)
		}))
		if !errors.Is(err, errPageRow) {
			t.Errorf("row %s with no number of the rule before it: error %v, want %v", row.id, err, errPageRow)
		}
	}
}

// A cell's numbers come from its rule and its generator alone: the same cell
// in a run of other cells, in another order, gives the same numbers, which is
// what lets the page's few cells stand for the whole run's.
func TestACellsNumbersDependOnItsRuleAndGeneratorAlone(t *testing.T) {
	t.Parallel()
	w, err := newWorld(pageAnswers)
	if err != nil {
		t.Fatal(err)
	}
	ms := metrics()
	summariesOf := func(all []cell) map[string][]summary {
		t.Helper()
		results, err := runCells(w, all, 3, ms)
		if err != nil {
			t.Fatal(err)
		}
		byName := map[string][]summary{}
		for c, s := range summarizeCells(all, results, ms) {
			byName[all[c].name()] = s
		}
		return byName
	}
	page := summariesOf(pageCells())
	other := slices.Clone(pageCells())
	slices.Reverse(other)
	other = append([]cell{{rule: slowConstantRule(), generator: misplaced}}, other...)
	for name, s := range summariesOf(other) {
		if want, kept := page[name]; kept && !slices.Equal(s, want) {
			t.Errorf("%s gives other numbers in another run of cells", name)
		}
	}
}

// The shift the page names is the solver's own: every letter of the first
// run comes back that many places on in the second, around the five.
func TestThePagesShiftIsTheSolversRelabelling(t *testing.T) {
	t.Parallel()
	facts, err := productFacts()
	if err != nil {
		t.Fatal(err)
	}
	for place, letter := range solver.Letters() {
		want := solver.Letter((place + facts.Relabel) % solver.Count)
		if got := solver.Relabelled(letter); got != want || facts.Relabelled[place] != want {
			t.Errorf("%s comes back as %s, and the page says %s; a shift of %d says %s", letter, got, facts.Relabelled[place], facts.Relabel, want)
		}
	}
}

// Every reference task is counted under its level, so the counts by level
// add up to them all.
func TestThePageCountsEveryReferenceTaskUnderItsLevel(t *testing.T) {
	t.Parallel()
	facts, err := productFacts()
	if err != nil {
		t.Fatal(err)
	}
	sum := 0
	for _, n := range facts.ReferenceTasks.ByLevel {
		sum += n
	}
	if sum != facts.ReferenceTasks.Total || len(facts.ReferenceTasks.ByLevel) != len(rating.GradeLevels()) {
		t.Errorf("the reference tasks by level %v add up to %d, want all %d", facts.ReferenceTasks.ByLevel, sum, facts.ReferenceTasks.Total)
	}
}

// The page labels the error "after so many answers" and the screen "in
// answers so many to so many" from the data, its words holding no digit, so a
// row must read its measure at the very answers its label names.
func TestThePageReadsItsMeasuresAtTheAnswersItsLabelsName(t *testing.T) {
	t.Parallel()
	bench, err := benchBlockOf(pageNumbersBy(spread), 2, keyOfBuild)
	if err != nil {
		t.Fatal(err)
	}
	after := "r1_rms_" + strconv.Itoa(bench.ErrorAfter)
	late := "_" + strconv.Itoa(bench.ScreenWindows.Late.First) + "_" + strconv.Itoa(bench.ScreenWindows.Late.Last)
	for _, row := range bench.Rows {
		if strings.HasPrefix(row.Metric, "r1_rms_") && row.Metric != after {
			t.Errorf("row %s reads %s, and its label names the error %s", row.ID, row.Metric, after)
		}
		if strings.HasPrefix(row.Metric, "r8_") && !strings.HasSuffix(row.Metric, late) {
			t.Errorf("row %s reads %s, and its label names the answers %s", row.ID, row.Metric, late)
		}
	}
}

// A number that is none cannot be written: the data would carry a hole the
// page would draw as a number.
func TestThePageRefusesANumberThatIsNone(t *testing.T) {
	t.Parallel()
	if err := writeJSON(filepath.Join(t.TempDir(), "numbers.json"), pageValue{Value: math.NaN()}); err == nil {
		t.Error("writeJSON wrote NaN, want an error")
	}
}

// keyOfBuild is a key of a build, as the page's run is given it.
var keyOfBuild = strings.Repeat("0123456789abcdef", 4)

// The page's command takes the key of its build and the file it writes, and
// refuses what it cannot run before it runs anything.
func TestThePageCommandRefusesWhatItCannotRun(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "numbers.json")
	for _, tc := range []struct {
		name string
		args []string
		says string
	}{
		{"words besides its flags", []string{"-inputs", keyOfBuild, "-out", out, "now"}, "flags alone"},
		{"a run of other children", []string{"-children", "10", "-inputs", keyOfBuild, "-out", out}, "-children"},
		{"no key", []string{"-out", out}, "-inputs"},
		{"a key of another shape", []string{"-inputs", "abc", "-out", out}, "-inputs"},
		{"no file", []string{"-inputs", keyOfBuild}, "-out"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr strings.Builder
			if code := pageCommand(tc.args, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), tc.says) {
				t.Errorf("pageCommand(%q) = %d, %q; want 2 and a word on %s", tc.args, code, stderr.String(), tc.says)
			}
		})
	}
	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a refused run wrote %s", out)
	}
}

// fixtureCommit and fixtureDate are the commit the fixture of the page's data
// is made from, fixed by the fixture rather than read off the repository.
const (
	fixtureCommit = "0123456789abcdef0123456789abcdef01234567"
	fixtureDate   = "2026-10-05T00:00:00Z"
	paperCommitOf = "52ce86908135"
)

// fixturePaper is the fixture of the paper's facts.
var fixturePaper = filepath.Join("testdata", "paper.json")

// fixtureLive is the snapshot of the live numbers the fixture is made with:
// a month whose numbers are shown, made up.
var fixtureLive = filepath.Join("testdata", "live.json")

// numbersFile writes a run's numbers where the page's file is made from.
func numbersFile(t *testing.T, numbers *pageNumbers) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "numbers.json")
	if err := writeJSON(path, numbers); err != nil {
		t.Fatal(err)
	}
	return path
}

// smallNumbers are the page's numbers as a few children of each cell give
// them, read by a fake run rather than a real one.
func smallNumbers(t *testing.T) *pageNumbers {
	t.Helper()
	bench, err := benchBlockOf(pageNumbersBy(spread), 2, keyOfBuild)
	if err != nil {
		t.Fatal(err)
	}
	product, err := productFacts()
	if err != nil {
		t.Fatal(err)
	}
	return &pageNumbers{Bench: bench, Product: product}
}

// fileOfText writes a file of a test and gives its path.
func fileOfText(t *testing.T, dir, name, text string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// The page's file vouches for what it says, so what it is made from is held
// to its shape first: the numbers and the build they are of, the commit, its
// date, the paper's commit and facts.
func TestThePageFileRefusesWhatItCannotVouchFor(t *testing.T) {
	t.Parallel()
	small := smallNumbers(t)
	numbers := numbersFile(t, small)
	dir := t.TempDir()
	good := pageFileArgs{numbers: numbers, inputs: keyOfBuild, commit: fixtureCommit, date: fixtureDate, paperCommit: paperCommitOf, out: "unused"}
	file := `{"lang": "en", "path": "/assets/paper-a.en.pdf", "pages": 15, "bytes": 350493, "sha256": "` + strings.Repeat("ab", 32) + `"}`
	facts := func(files string) string { return `{"commit": "52ce86908135", "files": [` + files + `]}` }
	for _, tc := range []struct {
		name string
		edit func(t *testing.T, a *pageFileArgs)
	}{
		{"a short commit", func(_ *testing.T, a *pageFileArgs) { a.commit = "0123456" }},
		{"a date of another form", func(_ *testing.T, a *pageFileArgs) { a.date = "2026-10-05" }},
		{"no paper's commit", func(_ *testing.T, a *pageFileArgs) { a.paperCommit = "" }},
		{"no key", func(_ *testing.T, a *pageFileArgs) { a.inputs = "" }},
		{"numbers of another build", func(_ *testing.T, a *pageFileArgs) { a.inputs = strings.Repeat("f", 64) }},
		{"no numbers", func(_ *testing.T, a *pageFileArgs) { a.numbers = filepath.Join(dir, "none.json") }},
		{"numbers of another shape", func(t *testing.T, a *pageFileArgs) {
			a.numbers = fileOfText(t, dir, "other.json", `{"bench": {}, "product": {}, "more": 1}`)
		}},
		{"numbers with no product", func(t *testing.T, a *pageFileArgs) { a.numbers = numbersFile(t, &pageNumbers{Bench: small.Bench}) }},
		{"numbers with a row missing", func(t *testing.T, a *pageFileArgs) {
			cut := *small
			cut.Bench.Rows = cut.Bench.Rows[1:]
			a.numbers = numbersFile(t, &cut)
		}},
		{"facts of another shape", func(t *testing.T, a *pageFileArgs) {
			a.paper = fileOfText(t, dir, "shape.json", `{"commit": "52ce86908135", "files": [`+file+`], "title": "x"}`)
		}},
		{"facts of no file", func(t *testing.T, a *pageFileArgs) { a.paper = fileOfText(t, dir, "none-listed.json", facts("")) }},
		{"a file outside the assets", func(t *testing.T, a *pageFileArgs) {
			a.paper = fileOfText(t, dir, "path.json", facts(strings.Replace(file, "/assets/", "/research/", 1)))
		}},
		{"a file of no pages", func(t *testing.T, a *pageFileArgs) {
			a.paper = fileOfText(t, dir, "pages.json", facts(strings.Replace(file, `"pages": 15`, `"pages": 0`, 1)))
		}},
		{"a file of no hash", func(t *testing.T, a *pageFileArgs) {
			a.paper = fileOfText(t, dir, "hash.json", facts(strings.Replace(file, strings.Repeat("ab", 32), "ab", 1)))
		}},
		{"a snapshot of the live numbers that is not there", func(_ *testing.T, a *pageFileArgs) { a.live = filepath.Join(dir, "no-live.json") }},
		{"a snapshot of the live numbers of another shape", func(t *testing.T, a *pageFileArgs) {
			a.live = fileOfText(t, dir, "live-shape.json", `{"month": null, "total": null, "chances": [], "kept_up": [], "more": 1}`)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := good
			tc.edit(t, &a)
			if _, _, err := pageFileOf(&a); err == nil {
				t.Error("made the file, want it refused")
			}
		})
	}
	if _, _, err := pageFileOf(&good); err != nil {
		t.Errorf("the good arguments: error %v, want the file made", err)
	}
	for _, tc := range []struct {
		name string
		args []string
		code int
	}{
		{"words besides its flags", []string{"-numbers", numbers, "-out", filepath.Join(dir, "x.json"), "now"}, 2},
		{"a flag of the wrong shape", []string{"-numbers", numbers, "-inputs", keyOfBuild, "-commit", "abc", "-date", fixtureDate,
			"-paper-commit", paperCommitOf, "-out", filepath.Join(dir, "y.json")}, 2},
		{"numbers of another build", []string{"-numbers", numbers, "-inputs", strings.Repeat("f", 64), "-commit", fixtureCommit,
			"-date", fixtureDate, "-paper-commit", paperCommitOf, "-out", filepath.Join(dir, "z.json")}, 1},
		{"a snapshot of the live numbers that is not there", []string{"-numbers", numbers, "-inputs", keyOfBuild, "-commit", fixtureCommit,
			"-date", fixtureDate, "-paper-commit", paperCommitOf, "-live", filepath.Join(dir, "no-live.json"), "-out", filepath.Join(dir, "w.json")}, 1},
	} {
		t.Run("the command, given "+tc.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr strings.Builder
			if code := pageFileCommand(tc.args, &stdout, &stderr); code != tc.code {
				t.Errorf("exited %d, want %d: %s", code, tc.code, stderr.String())
			}
		})
	}
}

// The page names the commit of the paper it links: the PDF's own when the
// site ships one, the paper's numbers' otherwise, and the job's summary says
// which it is, and when the two part.
func TestThePageFileTakesThePapersCommitFromThePDFItShips(t *testing.T) {
	t.Parallel()
	numbers := numbersFile(t, smallNumbers(t))
	for _, tc := range []struct {
		name, paper, paperCommit, want, says string
		files                                int
		warns                                bool
	}{
		{"no PDF", "", "abcdef123456", "abcdef123456", "The paper's numbers are of the commit abcdef123456.", 0, false},
		{"a PDF of the paper's commit", fixturePaper, paperCommitOf, paperCommitOf, "PDF in en, whose numbers are of the commit " + paperCommitOf + ".", 1, false},
		{"a PDF of an older commit", fixturePaper, "abcdef123456", paperCommitOf, "PDF in en, whose numbers are of the commit " + paperCommitOf + ".", 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := pageFileArgs{numbers: numbers, inputs: keyOfBuild, commit: fixtureCommit, date: fixtureDate, paperCommit: tc.paperCommit, paper: tc.paper}
			file, warning, err := pageFileOf(&a)
			if err != nil {
				t.Fatalf("error %v", err)
			}
			if file.Paper.Commit != tc.want || len(file.Paper.Files) != tc.files || (warning != "") != tc.warns {
				t.Errorf("the paper's commit %s with %d files, warning %q; want %s with %d, warning %v",
					file.Paper.Commit, len(file.Paper.Files), warning, tc.want, tc.files, tc.warns)
			}
			if report := goalsReport(&file, warning); !strings.Contains(report, tc.says) {
				t.Errorf("the summary does not say %q:\n%s", tc.says, report)
			}
		})
	}
}

// The table of goals in the job's summary shows each row under both rules,
// with their marks, and what it was read against.
func TestTheGoalsTableShowsBothRulesWithTheirMarks(t *testing.T) {
	t.Parallel()
	numbers := numbersFile(t, smallNumbers(t))
	file, _, err := pageFileOf(&pageFileArgs{numbers: numbers, inputs: keyOfBuild, commit: fixtureCommit, date: fixtureDate, paperCommit: paperCommitOf})
	if err != nil {
		t.Fatal(err)
	}
	report := goalsReport(&file, "the PDF is old")
	for _, row := range file.Bench.Rows {
		line := lineOf(report, row.ID)
		if line == "" {
			t.Errorf("the table has no row %s:\n%s", row.ID, report)
			continue
		}
		for _, v := range []pageValue{row.Values.Service, row.Values.Earlier} {
			if v.Mark != nil && !strings.Contains(line, strings.ReplaceAll(*v.Mark, "_", " ")) {
				t.Errorf("row %s does not show the mark %s: %s", row.ID, *v.Mark, line)
			}
		}
	}
	for _, want := range []string{"shrinking/both", "earlier/both", "oracle/both", "the PDF is old", short(fixtureCommit)} {
		if !strings.Contains(report, want) {
			t.Errorf("the table does not say %q:\n%s", want, report)
		}
	}
}

// lineOf is the line of a Markdown table that holds this row, or none.
func lineOf(table, id string) string {
	for _, line := range strings.Split(table, "\n") {
		if strings.HasPrefix(line, "| "+id+" |") {
			return line
		}
	}
	return ""
}

// fixtureOfThePage is the fixture of the page's data file, which the site's
// reader of it reads too: one file holds both sides to one shape.
var fixtureOfThePage = filepath.Join("..", "..", "site", "research", "testdata", "research.json")

// errRewriteElsewhere is a fixture asked to be rewritten where its bytes do
// not hold.
var errRewriteElsewhere = errors.New("the page's fixture holds its bytes on amd64 with FMA, built at GOAMD64 below v3, alone, and is rewritten there")

// fixtureHolds says whether the fixture's bytes hold on a machine: on amd64,
// built at a level at which Go fuses no multiplication with an addition, and
// on a processor with FMA, by which math.Exp takes the path it takes on every
// runner. Elsewhere the last digits of its numbers part, and a fixture
// rewritten there would fail on every machine that keeps it.
func fixtureHolds(arch string, fma, fused bool) error {
	if arch != "amd64" || !fma || fused {
		return errRewriteElsewhere
	}
	return nil
}

// The fixture is held and rewritten on amd64 with FMA, built unfused, alone.
func TestTheFixtureHoldsOnAmd64WithFMAAlone(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		arch       string
		fma, fused bool
		holds      bool
	}{
		{"arm64", true, false, false},
		{"amd64", false, false, false},
		{"amd64", true, true, false},
		{"amd64", true, false, true},
	} {
		if err := fixtureHolds(tc.arch, tc.fma, tc.fused); (err == nil) != tc.holds {
			t.Errorf("fixtureHolds(%s, FMA %v, fused %v) = %v, want it to hold %v", tc.arch, tc.fma, tc.fused, err, tc.holds)
		}
	}
}

// A run of ten children a cell, the fixtures of the paper's facts and of the
// live numbers' snapshot and a fixed commit make the page's file of the
// fixture, byte for byte, and every mark in it follows from its numbers. A
// change of the model or of the content moves its numbers, which go test
// -update rewrites from the run, on amd64 with FMA, to be read before it is
// kept.
func TestThePageFileIsItsFixture(t *testing.T) {
	t.Parallel()
	here := fixtureHolds(runtime.GOARCH, cpu.X86.HasAVX && cpu.X86.HasFMA, fusedMultiplyAdd)
	if *update && here != nil {
		t.Fatalf("%v, and this is %s", here, runtime.GOARCH)
	}
	if here != nil {
		t.Skipf("%v, and this is %s", here, runtime.GOARCH)
	}
	file := smallPageFile(t)
	for _, row := range file.Bench.Rows {
		for _, problem := range marksThatDoNotFollow(&row) {
			t.Errorf("row %s: %s", row.ID, problem)
		}
	}
	made := encoded(t, &file)
	if *update {
		keepFixture(t, made)
	}
	if want := readBytes(t, fixtureOfThePage); !bytes.Equal(made, want) {
		t.Errorf("the page's file of a small run is not its fixture %s; if the change is meant, rewrite it with "+
			"go test -run TestThePageFileIsItsFixture -update on amd64 and read the diff:\n%s", fixtureOfThePage, made)
	}
	var back researchFile
	if err := readStrictJSON(fixtureOfThePage, &back); err != nil || back.Schema != researchSchema {
		t.Errorf("the fixture reads back as schema %d, error %v; want schema %d", back.Schema, err, researchSchema)
	}
}

// smallPageFile is the page's file of a run of ten children a cell, with the
// fixtures of the paper's facts and of the live numbers' snapshot, and the
// fixture's commit.
func smallPageFile(t *testing.T) researchFile {
	t.Helper()
	numbers, err := pageNumbersOf(10, keyOfBuild)
	if err != nil {
		t.Fatal(err)
	}
	file, _, err := pageFileOf(&pageFileArgs{
		numbers: numbersFile(t, &numbers), inputs: keyOfBuild, commit: fixtureCommit, date: fixtureDate, paperCommit: paperCommitOf,
		paper: fixturePaper, live: fixtureLive,
	})
	if err != nil {
		t.Fatal(err)
	}
	return file
}

// encoded is the bytes the page's file is written as.
func encoded(t *testing.T, file *researchFile) []byte {
	t.Helper()
	path := filepath.Join(t.TempDir(), "research.json")
	if err := writeJSON(path, file); err != nil {
		t.Fatal(err)
	}
	return readBytes(t, path)
}

// keepFixture rewrites the fixture of the page's file with what a run made.
func keepFixture(t *testing.T, made []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(fixtureOfThePage), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixtureOfThePage, made, 0o600); err != nil {
		t.Fatal(err)
	}
}

// marksThatDoNotFollow are the marks of a row that its numbers do not give:
// a goal's mark is read on the whole interval against the bound, on the side
// the row calls better, and the baseline stands only on the service, against
// a bound of its own.
func marksThatDoNotFollow(row *pageRowData) []string {
	if row.Bound == nil {
		return nil
	}
	var problems []string
	for _, rv := range []struct {
		whose   string
		v       pageValue
		service bool
	}{{"the service", row.Values.Service, true}, {"the rule before it", row.Values.Earlier, false}} {
		mark := *rv.v.Mark
		if mark == markBaseline {
			if !rv.service || !row.Bound.Own {
				problems = append(problems, rv.whose+" is marked the baseline against a bound not its own")
			}
			continue
		}
		low, high, bound := rv.v.Low, rv.v.High, row.Bound.Value
		if row.Better == higher {
			low, high, bound = -rv.v.High, -rv.v.Low, -bound
		}
		want := markOnTheEdge
		switch {
		case high <= bound:
			want = markReached
		case low > bound:
			want = markNotReached
		}
		if mark != want {
			problems = append(problems, rv.whose+" is marked "+mark+", and its numbers give "+want)
		}
	}
	return problems
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
