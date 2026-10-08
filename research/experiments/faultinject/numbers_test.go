package main

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// numberKeyPattern is what a key of the numbers may be: it ends up in the
// name of a macro, where only these characters are safe.
var numberKeyPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// numbersIn reads a file of numbers by the rule the paper's macros read every
// such file by: blank lines and comments aside, every line is key=value, with
// a key of lower-case letters, digits and underscores, a value that is not
// empty, and no key given twice, so that a key the paper cites names one
// number.
func numbersIn(t *testing.T, text string) map[string]string {
	t.Helper()
	values := map[string]string{}
	for n, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		_, seen := values[key]
		switch {
		case !found:
			t.Errorf("line %d, %q, has no '='", n+1, line)
		case !numberKeyPattern.MatchString(key):
			t.Errorf("line %d: %q is not a key of lower-case letters, digits and underscores", n+1, key)
		case value == "":
			t.Errorf("line %d: %q has no value", n+1, key)
		case seen:
			t.Errorf("line %d: %q is given twice", n+1, key)
		}
		values[key] = value
	}
	return values
}

// casesOfEveryOperator are valid cases of every operator of the experiment on
// hosts of every shipped topic: refused as expected for an operator that
// breaks a rule a check states, and for any other missed once in every topic
// and refused besides in every other one.
func casesOfEveryOperator(ops []operator, topics []string) []caseRow {
	var rows []caseRow
	for i := range ops {
		op := &ops[i]
		for j, topic := range topics {
			on := &host{ID: op.ID + "/" + strconv.Itoa(j), Topic: topic, Level: rating.Grades56}
			caught := caseRow{Operator: op, Host: on, Valid: true, Verdict: refusedBy(checks.CodeReadability)}
			if len(op.Expected) > 0 {
				caught.Verdict = refusedBy(op.Expected[0])
			}
			missed := caseRow{Operator: op, Host: on, Valid: true}
			switch {
			case op.Kind == mechanism:
				rows = append(rows, caught)
			case j%2 == 0:
				rows = append(rows, caught, missed)
			default:
				rows = append(rows, missed)
			}
		}
	}
	return rows
}

// The numbers of a run whose every operator made valid cases are all of them
// numbers the paper can cite: one key=value line each, every key a name a
// macro can carry and none given twice, whatever the ids of the operators and
// topics are made of.
func TestEveryNumberOfARunIsOneThePaperCanCite(t *testing.T) {
	t.Parallel()
	shipped, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	ops := operators()
	found := &results{
		Sanity: []sanityRow{
			{Host: &host{ID: "a", Level: rating.Grades12}},
			{Host: &host{ID: "b", Level: rating.Grades34}, Verdict: refusedBy(checks.CodeReadability)},
			{Host: &host{ID: "c", Level: rating.Grades56}},
		},
		Cases: casesOfEveryOperator(ops, shipped.TopicIDs()),
	}
	read := map[string]share{}
	for _, op := range ops {
		if op.Kind == outOfScope {
			read[op.ID] = share{hits: 1, total: 2}
		}
	}
	path := filepath.Join(t.TempDir(), "numbers.txt")
	if err = writeNumbers(path, found, ops, read); err != nil {
		t.Fatalf("writeNumbers: %v", err)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	values := numbersIn(t, string(written))
	for key, want := range map[string]string{
		"hosts": "3", "eligible": "2", "operators": strconv.Itoa(len(ops)), "x09b_read": "2", "x09b_equivalent": "1",
		"d14b_arithmetic_tricks_missed": "1", "d09a_design_thought_process_cases": strconv.Itoa(len(shipped.TopicIDs())),
	} {
		if got := values[key]; got != want {
			t.Errorf("numbers.txt %s = %q, want %q", key, got, want)
		}
	}
}

// The numbers of the operators that break a rule a check states: an operator
// with no case is counted apart and given no interval, an operator one check
// refused alone every time is told from one with a case it missed, and the
// lowest start of an interval is found apart for the operators only the tasks
// with a drawing take. The others count for nothing here.
func TestMechanismNumbersSumUpTheOperatorsThatBreakARule(t *testing.T) {
	t.Parallel()
	ops := []operator{
		{ID: "D01a", Class: "D01", Kind: mechanism, Expected: []checks.Code{checks.CodeSolverDisagrees}},
		{ID: "D16a", Class: "D16", Kind: mechanism, Expected: []checks.Code{checks.CodeDrawingFormat}},
		{ID: "D17a", Class: "D17", Kind: mechanism, Expected: []checks.Code{checks.CodeDrawingFormat}},
		{ID: "D18a", Class: "D18", Kind: mechanism, Expected: []checks.Code{checks.CodeDrawingMismatch}},
		{ID: "D02b", Class: "D02", Kind: discovery, Expected: []checks.Code{checks.CodeBadStructure}},
		{ID: "X01a", Class: "X01", Kind: outOfScope},
	}
	valid := func(op *operator, v verdict) caseRow {
		return caseRow{Operator: op, Host: &host{ID: "a"}, Valid: true, Verdict: v}
	}
	disagrees, format := refusedBy(checks.CodeSolverDisagrees), refusedBy(checks.CodeDrawingFormat)
	rows := []caseRow{
		valid(&ops[0], disagrees), valid(&ops[0], disagrees), valid(&ops[0], disagrees),
		valid(&ops[1], format), valid(&ops[1], verdict{}),
		valid(&ops[2], format), valid(&ops[2], format), valid(&ops[2], format),
		valid(&ops[4], verdict{}), valid(&ops[5], verdict{}),
		{Operator: &ops[3], Host: &host{ID: "b"}, Verdict: refusedBy(checks.CodeDrawingMismatch)},
	}
	want := []string{
		"mechanism_operators=4",
		"mechanism_operators_without_cases=1",
		"mechanism_operators_one_check=2",
		"mechanism_operators_more_checks=1",
		"mechanism_cases=8",
		"mechanism_expected_percent=87.5",
		"mechanism_missed=1",
		"mechanism_lowest_low=29.2",
		"drawing_lowest_low=1.3",
		"drawing_hosts_fewest=2",
		"drawing_hosts_most=3",
	}
	if got := mechanismNumbers(rows, ops); !slices.Equal(got, want) {
		t.Errorf("mechanismNumbers =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// Where no operator only the tasks with a drawing take made a case, the fewest
// and the most of their cases are both zero, not the largest number there is.
func TestWithoutCasesOfADrawingTheirCountsAreZero(t *testing.T) {
	t.Parallel()
	ops := []operator{{ID: "D01a", Class: "D01", Kind: mechanism, Expected: []checks.Code{checks.CodeSolverDisagrees}}}
	rows := []caseRow{{Operator: &ops[0], Host: &host{ID: "a"}, Valid: true, Verdict: refusedBy(checks.CodeSolverDisagrees)}}
	lines := mechanismNumbers(rows, ops)
	for _, want := range []string{"drawing_hosts_fewest=0", "drawing_hosts_most=0"} {
		if !slices.Contains(lines, want) {
			t.Errorf("mechanismNumbers lack %q: %v", want, lines)
		}
	}
}

// The numbers of the hosts count the reference tasks at every level and the
// tasks the checks accept as they are, with every check run; the cases are the
// valid ones, and the classes those any case came from.
func TestHostNumbersCountTheTasksTheCasesAndTheClasses(t *testing.T) {
	t.Parallel()
	d01, x01 := &operator{ID: "D01a", Class: "D01"}, &operator{ID: "X01a", Class: "X01"}
	found := &results{
		Sanity: []sanityRow{
			{Host: &host{ID: "a", Level: rating.Grades12}},
			{Host: &host{ID: "b", Level: rating.Grades12}, Verdict: refusedBy(checks.CodeReadability)},
			{Host: &host{ID: "c", Level: rating.Grades34}, Verdict: verdict{Unchecked: []string{"the solver was not run"}}},
			{Host: &host{ID: "d", Level: rating.Grades56}},
		},
		Cases: []caseRow{
			{Operator: d01, Host: &host{ID: "a"}, Valid: true},
			{Operator: d01, Host: &host{ID: "d"}, Valid: true},
			{Operator: d01, Host: &host{ID: "d"}},
			{Operator: x01, Host: &host{ID: "a"}, Valid: true},
		},
	}
	want := []string{
		"hosts=4", "hosts_12=2", "eligible_12=1", "hosts_34=1", "eligible_34=0", "hosts_56=1", "eligible_56=1",
		"eligible=2", "cases=3", "classes=2",
	}
	if got := hostNumbers(found); !slices.Equal(got, want) {
		t.Errorf("hostNumbers = %q, want %q", got, want)
	}
}
