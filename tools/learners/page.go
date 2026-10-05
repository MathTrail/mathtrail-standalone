package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The page's run computes the numbers the site's page of the research shows
// of the student model: the service, the rule before its floor and its
// cautious mastery, and the ceiling, on the cells of the page's table and no
// others, each goal read as the bench's own criterion reads it. A cell's
// children and intervals depend on its rule and its generator alone, so these
// cells give the numbers of the whole run to the last digit, in a fraction of
// its time.

// pageCommandName is the word that runs the page's cells rather than the
// bench, and pageProducer how the page's data names what wrote its numbers.
const (
	pageCommandName = "page"
	pageProducer    = "tools/learners " + pageCommandName
)

// pageChildren and pageAnswers are the size of the page's run: the paper's,
// a thousand children a cell giving two hundred answers each.
const (
	pageChildren = 1000
	pageAnswers  = 200
)

// The answers the page's labels name: the error is read after the last
// checkpoint, and the screen over the last window, so a row's measure and
// the label the page gives it come from the same number.
var (
	errorMetric = "r1_rms_" + strconv.Itoa(checkpoints[len(checkpoints)-1])
	lateWindow  = screenWindows[len(screenWindows)-1]
)

// pageGenerators are the generators the page's goals are read on: the child
// who stays put, the widest spread of topics, the main learner and the jump.
var pageGenerators = []generator{staticChildren, farTopics, learningHalf, jumping}

// ceilingGenerators are those of them the ceiling bounds a step on: the
// mastery's ceiling is perfection, which needs no run.
var ceilingGenerators = []generator{staticChildren, learningHalf, jumping}

// pageCells are the cells of the page's table and no others: the service and
// the rule before it on every generator of the goals, the ceiling on those it
// bounds.
func pageCells() []cell {
	return append(
		cellsOn([]*rule{serviceRule(), earlierServiceRule()}, pageGenerators),
		cellsOn([]*rule{ceilingRule()}, ceilingGenerators)...,
	)
}

// pageCriterion reads the page's goals with the service as the baseline every
// bound is read off, and the oracle as the ceiling.
func pageCriterion() *criterion {
	return &criterion{baseline: serviceRule().name, goals: pageGoals}
}

// pageGoals are the goals of both choices, the step's and the mastery's, each
// once: the screen's goals stand in both lists. The screen's goals of answers
// 6–20 are left out, since their bound is the service's own number of the
// same window, which no reading of the service can do anything but reach.
func pageGoals() []goal {
	late := "_" + lateWindow.name()
	var all []goal
	for _, g := range append(stepGoals(), masteryGoals()...) {
		if g.screen && !strings.HasSuffix(g.metric, late) {
			continue
		}
		if !slices.ContainsFunc(all, func(seen goal) bool { return seen.generator == g.generator && seen.metric == g.metric }) {
			all = append(all, g)
		}
	}
	return all
}

// The groups the rows of the page's table fall into, and their kinds.
const (
	stepGroup    = "step"
	masteryGroup = "mastery"
	screenGroup  = "screen"
	goalRow      = "goal"
	contextRow   = "context"
)

// Whose numbers bound a row from above: the oracle's, which knows where the
// child stands, or perfection, no false mastery and no wait.
const (
	oracleCeiling  = "oracle"
	perfectCeiling = "perfect"
)

// pageRow is one row of the page's table: a measure on a generator, the
// group of goals it belongs to, or none for a number shown for context, its
// unit and its ceiling. own marks a goal whose bound is read off the
// service's own number of the same measure, which the service can therefore
// only ever reach or miss by construction; better says which way is better
// for a number of context, which has no goal to say it.
type pageRow struct {
	id        string
	group     string
	generator generator
	metric    string
	unit      string
	ceiling   string
	own       bool
	better    string
}

// pageRows are the rows of the page's table, in the order it shows them.
var pageRows = []pageRow{
	{id: "lag", group: stepGroup, generator: learningHalf, metric: "r6_lag", unit: "logit", ceiling: oracleCeiling, own: true},
	{id: "corridor_learning", group: stepGroup, generator: learningHalf, metric: "r3_inside", unit: "share", ceiling: oracleCeiling, own: true},
	{id: "jump_unsettled", group: stepGroup, generator: jumping, metric: "r6_jump_unsettled", unit: "share", ceiling: oracleCeiling},
	{id: "false_mastery_static", group: masteryGroup, generator: staticChildren, metric: "r4_false", unit: "share", ceiling: perfectCeiling},
	{id: "false_mastery_wide", group: masteryGroup, generator: farTopics, metric: "r4_false", unit: "share", ceiling: perfectCeiling},
	{id: "late_mastery_static", group: masteryGroup, generator: staticChildren, metric: "r5_late_answers", unit: "answers", ceiling: perfectCeiling, own: true},
	{id: "late_mastery_wide", group: masteryGroup, generator: farTopics, metric: "r5_late_answers", unit: "answers", ceiling: perfectCeiling, own: true},
	{id: "card_move_static", group: screenGroup, generator: staticChildren, metric: "r8_move_p95_" + lateWindow.name(), unit: "points"},
	{id: "rank_changes_static", group: screenGroup, generator: staticChildren, metric: "r8_rank_" + lateWindow.name(), unit: "per_100_answers"},
	{id: "card_move_learning", group: screenGroup, generator: learningHalf, metric: "r8_move_p95_" + lateWindow.name(), unit: "points"},
	{id: "rank_changes_learning", group: screenGroup, generator: learningHalf, metric: "r8_rank_" + lateWindow.name(), unit: "per_100_answers"},
	{id: "error_static", generator: staticChildren, metric: errorMetric, unit: "logit", ceiling: oracleCeiling, better: lower},
	{id: "corridor_static", generator: staticChildren, metric: "r3_inside", unit: "share", ceiling: oracleCeiling, better: higher},
}

// Which way a number is better.
const (
	lower  = "lower"
	higher = "higher"
)

// The marks a rule's number of a goal gets on the page: the bench's own, and
// the baseline's where the bound is read off the service's own number.
const (
	markReached    = "reached"
	markOnTheEdge  = "on_the_edge"
	markNotReached = "not_reached"
	markBaseline   = "baseline"
)

// pageMarks are the bench's marks as the page's data writes them.
var pageMarks = map[string]string{reached: markReached, onTheEdge: markOnTheEdge, notReached: markNotReached}

// pageNumbers is what the page's run writes: the bench's block of the page's
// data and the product's.
type pageNumbers struct {
	Bench   benchBlock   `json:"bench"`
	Product productBlock `json:"product"`
}

// benchBlock is the bench's part of the page's data: what the run was, the
// goals' parameters, the rules and the rows of the table.
type benchBlock struct {
	Producer       string          `json:"producer"`
	Inputs         string          `json:"inputs"`
	Seed           uint64          `json:"seed"`
	Experiment     string          `json:"experiment"`
	Children       int             `json:"children"`
	Answers        int             `json:"answers"`
	Interval       float64         `json:"interval"`
	Resamples      int             `json:"resamples"`
	ErrorAfter     int             `json:"error_after"`
	ScreenWindows  screenWindowsOf `json:"screen_windows"`
	GoalParameters goalParameters  `json:"goal_parameters"`
	Rules          []pageRule      `json:"rules"`
	Rows           []pageRowData   `json:"rows"`
}

// screenWindowsOf are the two windows of answers the screen is read over: the
// first after the trial series, and the last of the run.
type screenWindowsOf struct {
	Early answerSpan `json:"early"`
	Late  answerSpan `json:"late"`
}

// answerSpan is a window of answers, its first and its last.
type answerSpan struct {
	First int `json:"first"`
	Last  int `json:"last"`
}

// goalParameters are the numbers the goals are written with.
type goalParameters struct {
	LagShare      float64 `json:"lag_share"`
	CorridorShare float64 `json:"corridor_share"`
	UnsettledMost float64 `json:"unsettled_most"`
	FalseMost     float64 `json:"false_most"`
	LateTimes     float64 `json:"late_times"`
}

// pageRule is a rule of the page's run and the part it plays on the page.
type pageRule struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// pageRowData is a row of the table as the page's data writes it. What a row
// does not have is null rather than left out, so every row has the same keys.
type pageRowData struct {
	ID        string     `json:"id"`
	Kind      string     `json:"kind"`
	Criterion *string    `json:"criterion"`
	Generator string     `json:"generator"`
	Metric    string     `json:"metric"`
	Unit      string     `json:"unit"`
	ReadAs    string     `json:"read_as"`
	Better    string     `json:"better"`
	Bound     *pageBound `json:"bound"`
	Values    pageValues `json:"values"`
}

// pageBound is a goal's bound, and whether it is read off the service's own
// number of the same measure.
type pageBound struct {
	Value float64 `json:"value"`
	Own   bool    `json:"own"`
}

// pageValues are a row's numbers under each rule, the ceiling's when the row
// has one.
type pageValues struct {
	Service pageValue    `json:"service"`
	Earlier pageValue    `json:"earlier"`
	Ceiling *pageCeiling `json:"ceiling"`
}

// pageValue is a rule's number with its interval, and its mark when the row
// is a goal.
type pageValue struct {
	Value float64 `json:"value"`
	Low   float64 `json:"low"`
	High  float64 `json:"high"`
	Mark  *string `json:"mark"`
}

// pageCeiling is the number a row is bounded by from above, and whose it is.
type pageCeiling struct {
	Of    string  `json:"of"`
	Value float64 `json:"value"`
	Low   float64 `json:"low"`
	High  float64 `json:"high"`
}

// productBlock is the product's part of the page's data: the counts and
// constants the page quotes, read off the product's content and code.
type productBlock struct {
	Topics         int            `json:"topics"`
	Traps          int            `json:"traps"`
	Checks         int            `json:"checks"`
	Grades         gradeSpan      `json:"grades"`
	ReferenceTasks referenceTasks `json:"reference_tasks"`
	Options        int            `json:"options"`
	Relabel        int            `json:"relabel"`
	Relabelled     []string       `json:"relabelled"`
	Guess          float64        `json:"guess"`
	Corridor       corridorOf     `json:"corridor"`
	TrialAnswers   int            `json:"trial_answers"`
}

// gradeSpan is the first school year the service is for and the last.
type gradeSpan struct {
	First int `json:"first"`
	Last  int `json:"last"`
}

// referenceTasks are how many reference tasks the product carries, in all
// and by grade level.
type referenceTasks struct {
	Total   int            `json:"total"`
	ByLevel map[string]int `json:"by_level"`
}

// corridorOf is the corridor of the chance of a right answer, its bounds and
// its middle.
type corridorOf struct {
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Middle float64 `json:"middle"`
}

// errPageRow is a row of the page's table the run has no number for.
var errPageRow = errors.New("learners page: a row has no number")

// rowsOf reads every row of the page's table off a run read by the page's
// criterion.
func rowsOf(cr *criterionRun) ([]pageRowData, error) {
	earlier := ruleOf(cr, earlierRule)
	goals := pageGoals()
	rows := make([]pageRowData, 0, len(pageRows))
	for i := range pageRows {
		row := &pageRows[i]
		var (
			data pageRowData
			err  error
		)
		at := slices.IndexFunc(goals, func(g goal) bool { return g.generator == row.generator && g.metric == row.metric })
		switch {
		case row.group == "":
			data, err = contextRowOf(cr, earlier, row)
		case at < 0:
			err = fmt.Errorf("%w: %s holds no goal of either criterion", errPageRow, row.id)
		default:
			data, err = goalRowOf(cr, earlier, row, goals[at])
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, data)
	}
	return rows, nil
}

// ruleOf is the run's rule of this name.
func ruleOf(cr *criterionRun, name string) *rule {
	for c := range cr.all {
		if cr.all[c].rule.name == name {
			return cr.all[c].rule
		}
	}
	return nil
}

// goalRowOf reads a goal's row: the service's and the earlier rule's numbers
// as the criterion reads them, each with its mark, and the ceiling.
func goalRowOf(cr *criterionRun, earlier *rule, row *pageRow, g goal) (pageRowData, error) {
	service, earlierReading := cr.readGoal(cr.baseline, g), cr.readGoal(earlier, g)
	if service.verdict == unread || earlierReading.verdict == unread {
		return pageRowData{}, fmt.Errorf("%w: %s on %s", errPageRow, row.metric, row.generator)
	}
	ceiling, err := ceilingOf(cr, row, g.size)
	if err != nil {
		return pageRowData{}, err
	}
	serviceMark := pageMarks[service.mark]
	if row.own {
		serviceMark = markBaseline
	}
	earlierMark := pageMarks[earlierReading.mark]
	better := lower
	if !g.atMost {
		better = higher
	}
	return pageRowData{
		ID: row.id, Kind: goalRow, Criterion: &row.group, Generator: string(row.generator), Metric: row.metric,
		Unit: row.unit, ReadAs: readingOf(g.size), Better: better,
		Bound: &pageBound{Value: service.bound, Own: row.own},
		Values: pageValues{
			Service: pageValue{Value: service.value, Low: service.low, High: service.high, Mark: &serviceMark},
			Earlier: pageValue{Value: earlierReading.value, Low: earlierReading.low, High: earlierReading.high, Mark: &earlierMark},
			Ceiling: ceiling,
		},
	}, nil
}

// contextRowOf reads a row shown for context: every rule's number as it is,
// with no bound and no mark.
func contextRowOf(cr *criterionRun, earlier *rule, row *pageRow) (pageRowData, error) {
	service, hasService := cr.summaryOf(cr.baseline, row.generator, row.metric)
	before, hasBefore := cr.summaryOf(earlier, row.generator, row.metric)
	if !hasService || !hasBefore {
		return pageRowData{}, fmt.Errorf("%w: %s on %s", errPageRow, row.metric, row.generator)
	}
	ceiling, err := ceilingOf(cr, row, false)
	if err != nil {
		return pageRowData{}, err
	}
	return pageRowData{
		ID: row.id, Kind: contextRow, Generator: string(row.generator), Metric: row.metric,
		Unit: row.unit, ReadAs: readingOf(false), Better: row.better,
		Values: pageValues{
			Service: pageValue{Value: service.value, Low: service.low, High: service.high},
			Earlier: pageValue{Value: before.value, Low: before.low, High: before.high},
			Ceiling: ceiling,
		},
	}, nil
}

// ceilingOf is a row's ceiling: the oracle's number, read as the goal reads
// it, perfection, or none.
func ceilingOf(cr *criterionRun, row *pageRow, size bool) (*pageCeiling, error) {
	switch row.ceiling {
	case perfectCeiling:
		return &pageCeiling{Of: perfectCeiling}, nil
	case oracleCeiling:
		s, has := cr.summaryOf(cr.ceiling, row.generator, row.metric)
		if !has {
			return nil, fmt.Errorf("%w: the ceiling's %s on %s", errPageRow, row.metric, row.generator)
		}
		value, low, high := s.value, s.low, s.high
		if size {
			value, low, high = sizeOf(s)
		}
		return &pageCeiling{Of: oracleCeiling, Value: value, Low: low, High: high}, nil
	}
	return nil, nil
}

// readingOf names how a number is read: as it is, or as its size.
func readingOf(size bool) string {
	if size {
		return "size"
	}
	return "value"
}

// benchBlockOf is the bench's block of the page's data for a run: what it
// was, and its rows.
func benchBlockOf(cr *criterionRun, children int, inputs string) (benchBlock, error) {
	rows, err := rowsOf(cr)
	if err != nil {
		return benchBlock{}, err
	}
	early, late := screenWindows[0], lateWindow
	return benchBlock{
		Producer: pageProducer, Inputs: inputs, Seed: masterSeed, Experiment: experiment,
		Children: children, Answers: pageAnswers, Interval: 1 - 2*tail, Resamples: resamples,
		ErrorAfter:    checkpoints[len(checkpoints)-1],
		ScreenWindows: screenWindowsOf{Early: answerSpan{First: early.first, Last: early.last}, Late: answerSpan{First: late.first, Last: late.last}},
		GoalParameters: goalParameters{
			LagShare: lagShare, CorridorShare: corridorShare, UnsettledMost: unsettledMost, FalseMost: falseMost, LateTimes: lateTimes,
		},
		Rules: []pageRule{
			{ID: ruleID(serviceRule()), Role: "service"},
			{ID: ruleID(earlierServiceRule()), Role: "earlier"},
			{ID: ruleID(ceilingRule()), Role: "ceiling"},
		},
		Rows: rows,
	}, nil
}

// ruleID is how a rule is named in the results: its name and its structure.
func ruleID(r *rule) string { return r.name + "/" + string(r.shape) }

// productFacts are the counts and constants of the product the page quotes,
// read off its content and its code rather than typed, so that the page
// follows the product as it changes.
func productFacts() (productBlock, error) {
	shipped, err := content.Load()
	if err != nil {
		return productBlock{}, fmt.Errorf("learners page: load the content: %w", err)
	}
	byLevel := map[string]int{}
	for _, level := range rating.GradeLevels() {
		byLevel[string(level)] = len(shipped.ReferenceQuestions(level))
	}
	letters := solver.Letters()
	relabelled := make([]string, 0, len(letters))
	for _, letter := range letters {
		relabelled = append(relabelled, solver.Relabelled(letter))
	}
	return productBlock{
		Topics: len(shipped.TopicIDs()), Traps: len(shipped.TrapIDs()), Checks: len(checks.Codes()),
		Grades:         gradeSpan{First: profile.MinGrade, Last: profile.MaxGrade},
		ReferenceTasks: referenceTasks{Total: shipped.ExampleCount(), ByLevel: byLevel},
		Options:        solver.Count,
		// The shift between the solver's two runs is where the first letter
		// of the first run stands in the second.
		Relabel: solver.Place(solver.Relabelled(solver.Letter(0))), Relabelled: relabelled,
		Guess:        rating.Guess,
		Corridor:     corridorOf{Low: rating.CorridorLow, High: rating.CorridorHigh, Middle: rating.CorridorMiddle},
		TrialAnswers: rating.TrialAnswers,
	}, nil
}

// pageNumbersOf runs the page's cells, children a cell, and reads the page's
// numbers off them, with the product's.
func pageNumbersOf(children int, inputs string) (pageNumbers, error) {
	w, err := newWorld(pageAnswers)
	if err != nil {
		return pageNumbers{}, err
	}
	all := pageCells()
	ms := metrics()
	results, err := runCells(w, all, children, ms)
	if err != nil {
		return pageNumbers{}, err
	}
	cr := newCriterionRun(all, summarizeCells(all, results, ms), results, ms, children, pageCriterion())
	bench, err := benchBlockOf(cr, children, inputs)
	if err != nil {
		return pageNumbers{}, err
	}
	product, err := productFacts()
	if err != nil {
		return pageNumbers{}, err
	}
	return pageNumbers{Bench: bench, Product: product}, nil
}

// inputsKey is the key of what the numbers are computed from, as the build
// names it: the hash of the program and of the processor it runs on.
var inputsKey = regexp.MustCompile(`^[0-9a-f]{64}$`)

// pageCommand runs the page's cells and writes their numbers, with the
// product's, to the file it is given. The numbers are those of the paper's
// seed and run, a thousand children a cell, which the command takes no flag
// to change: the key of the build is then all that names them, and numbers
// kept under it are the page's whoever computed them.
func pageCommand(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("learners page", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputs := flags.String("inputs", "", "the key of what the numbers are computed from, 64 hexadecimal digits")
	out := flags.String("out", "", "the file the numbers are written to")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch {
	case flags.NArg() > 0:
		fmt.Fprintf(stderr, "learners page: takes flags alone, and was given %q\n", flags.Args())
		return 2
	case !inputsKey.MatchString(*inputs):
		fmt.Fprintf(stderr, "learners page: -inputs is %q, want the key of the build, 64 hexadecimal digits\n", *inputs)
		return 2
	case *out == "":
		fmt.Fprintln(stderr, "learners page: -out names no file to write the numbers to")
		return 2
	}
	numbers, err := pageNumbersOf(pageChildren, *inputs)
	if err == nil {
		err = writeJSON(*out, numbers)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "learners page: %d cells of %d children, %d answers each, written to %s\n", len(pageCells()), pageChildren, pageAnswers, *out)
	return 0
}

// writeJSON writes a value as indented JSON, ending in a new line. Every
// number is written in the shortest form that reads back as the same number,
// which is also the form JavaScript writes it in.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("learners: write %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("learners: write %s: %w", path, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("learners: write %s: %w", path, err)
	}
	return nil
}
