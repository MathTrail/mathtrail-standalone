package main

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// readSample is a sample of three cases: two of X02a, on hosts a and b, and
// one of X03a, on host c.
func readSample() map[string][]*caseRow {
	return map[string][]*caseRow{
		"X02a": {{Host: &host{ID: "a"}}, {Host: &host{ID: "b"}}},
		"X03a": {{Host: &host{ID: "c"}}},
	}
}

func writeReadingFile(t *testing.T, written string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "reading.csv"), []byte(written), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestReadingVerdictsAreCountedPerOperator(t *testing.T) {
	t.Parallel()
	if verdicts, err := readVerdicts(t.TempDir(), readSample()); err != nil || len(verdicts) != 0 {
		t.Fatalf("readVerdicts with no file = %v, %v; want none, nil", verdicts, err)
	}
	dir := writeReadingFile(t, "operator,host,verdict,note\nX02a,a,real,\nX02a,b,equivalent,\"why, with a comma\"\nX03a,c,real,\n")
	verdicts, err := readVerdicts(dir, readSample())
	if err != nil {
		t.Fatalf("readVerdicts: %v", err)
	}
	if got, want := verdicts["X02a"], (share{hits: 1, total: 2}); got != want {
		t.Errorf("X02a = %+v, want %+v", got, want)
	}
	if got, want := verdicts["X03a"], (share{hits: 0, total: 1}); got != want {
		t.Errorf("X03a = %+v, want %+v", got, want)
	}
}

// A verdict counts only when it is real or equivalent, on a case of this
// run's sample, and given once.
func TestAReadingVerdictBelongsToTheSample(t *testing.T) {
	t.Parallel()
	for name, written := range map[string]string{
		"an unknown verdict":         "operator,host,verdict,note\nX02a,a,maybe,\n",
		"a short row":                "operator,host\nX02a,a\n",
		"a case outside the sample":  "operator,host,verdict,note\nX02a,z,real,\n",
		"an operator not sampled":    "operator,host,verdict,note\nX05a,a,real,\n",
		"two verdicts on one case":   "operator,host,verdict,note\nX02a,a,real,\nX02a,a,equivalent,\n",
		"the operator of a neighbor": "operator,host,verdict,note\nX03a,a,real,\n",
	} {
		if _, err := readVerdicts(writeReadingFile(t, written), readSample()); err == nil {
			t.Errorf("%s: readVerdicts = nil error, want one", name)
		}
	}
}

// An operator counts as refused by one check alone only when the same check,
// alone, refused every valid case of it; a case the run left out, or a case
// of another operator, does not count against it.
func TestAnOperatorIsRefusedByOneCheckWhenOneCheckRefusedEveryCase(t *testing.T) {
	t.Parallel()
	op, other := &operator{ID: "D01a"}, &operator{ID: "D02a"}
	one := verdict{Codes: []checks.Code{checks.CodeSolverDisagrees}}
	another := verdict{Codes: []checks.Code{checks.CodeBadStructure}}
	two := verdict{Codes: []checks.Code{checks.CodeBadStructure, checks.CodeSolverDisagrees}}
	cases := []struct {
		name string
		rows []caseRow
		want bool
	}{
		{"one check every time", []caseRow{{Operator: op, Valid: true, Verdict: one}, {Operator: op, Valid: true, Verdict: one}}, true},
		{"a second check once", []caseRow{{Operator: op, Valid: true, Verdict: one}, {Operator: op, Valid: true, Verdict: two}}, false},
		{"another check alone once", []caseRow{{Operator: op, Valid: true, Verdict: one}, {Operator: op, Valid: true, Verdict: another}}, false},
		{"a case no check refused", []caseRow{{Operator: op, Valid: true, Verdict: one}, {Operator: op, Valid: true}}, false},
		{"no valid case", []caseRow{{Operator: op, Verdict: one}}, false},
		{"a second check on a case left out", []caseRow{{Operator: op, Valid: true, Verdict: one}, {Operator: op, Verdict: two}}, true},
		{"a second check on another operator", []caseRow{{Operator: op, Valid: true, Verdict: one}, {Operator: other, Valid: true, Verdict: two}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := refusedByOneCheck(c.rows, op); got != c.want {
				t.Errorf("refusedByOneCheck = %v, want %v", got, c.want)
			}
		})
	}
}

// A class's numbers and an out-of-scope operator's carry the share refused by
// any check with its exact interval: one of two cases refused is 50.0 % with
// a Clopper–Pearson interval of 1.3–98.7 %.
func TestRefusedSharesCarryTheirExactInterval(t *testing.T) {
	t.Parallel()
	op := &operator{ID: "X01a", Class: "X01", Kind: outOfScope}
	refused := verdict{Codes: []checks.Code{checks.CodeReadability}}
	rows := []caseRow{{Operator: op, Valid: true, Verdict: refused}, {Operator: op, Valid: true}}
	for name, lines := range map[string][]string{"class": classNumbers(rows), "operator": operatorNumbers(rows, op, nil)} {
		key := map[string]string{"class": "x01", "operator": "x01a"}[name]
		for _, want := range []string{key + "_refused_percent=50.0", key + "_refused_low=1.3", key + "_refused_high=98.7"} {
			if !slices.Contains(lines, want) {
				t.Errorf("%s numbers lack %q: %v", name, want, lines)
			}
		}
	}
}

// The numbers of the hosts count the reference tasks the checks accept as
// they are, level by level and in all.
func TestHostNumbersCountTheAcceptedTasksInAll(t *testing.T) {
	t.Parallel()
	refused := verdict{Codes: []checks.Code{checks.CodeReadability}}
	found := &results{Sanity: []sanityRow{
		{Host: &host{ID: "a", Level: rating.Grades12}},
		{Host: &host{ID: "b", Level: rating.Grades12}, Verdict: refused},
		{Host: &host{ID: "c", Level: rating.Grades56}},
	}}
	lines := hostNumbers(found)
	for _, want := range []string{"eligible_12=1", "eligible_56=1", "eligible=2"} {
		if !slices.Contains(lines, want) {
			t.Errorf("hostNumbers lack %q: %v", want, lines)
		}
	}
}

// The numbers of the operators that break a rule a check states count the
// cases the expected check missed, so that "every case was refused" rests on a
// count rather than on a rounded share.
func TestMechanismNumbersCountTheCasesTheExpectedCheckMissed(t *testing.T) {
	t.Parallel()
	op := &operator{ID: "D01a", Class: "D01", Kind: mechanism, Expected: []checks.Code{checks.CodeSolverDisagrees}}
	caught := verdict{Codes: []checks.Code{checks.CodeSolverDisagrees}}
	rows := []caseRow{{Operator: op, Valid: true, Verdict: caught}, {Operator: op, Valid: true, Verdict: caught}, {Operator: op, Valid: true}}
	lines := mechanismNumbers(rows, []operator{*op})
	for _, want := range []string{"mechanism_cases=3", "mechanism_missed=1", "mechanism_operators_one_check=0", "mechanism_operators_more_checks=1"} {
		if !slices.Contains(lines, want) {
			t.Errorf("mechanismNumbers lack %q: %v", want, lines)
		}
	}
}

// A discovery operator's numbers count the cases any check refused and give
// that share's exact interval, beside the share the expected check refused.
func TestDiscoveryNumbersCountWhatAnyCheckRefused(t *testing.T) {
	t.Parallel()
	op := &operator{ID: "D14b", Class: "D14", Kind: discovery, Expected: []checks.Code{checks.CodeNearDuplicate}}
	other := verdict{Codes: []checks.Code{checks.CodeReadability}}
	topic := &host{ID: "a", Topic: "arithmetic.tricks"}
	rows := []caseRow{{Operator: op, Host: topic, Valid: true, Verdict: other}, {Operator: op, Host: topic, Valid: true}}
	lines := operatorNumbers(rows, op, nil)
	for _, want := range []string{"d14b_expected_percent=0.0", "d14b_refused=1", "d14b_refused_percent=50.0", "d14b_refused_low=1.3", "d14b_refused_high=98.7"} {
		if !slices.Contains(lines, want) {
			t.Errorf("operatorNumbers lack %q: %v", want, lines)
		}
	}
}

// The cases drawn for a person to read are valid ones, which every share the
// paper reports is counted over.
func TestTheReadingSampleDrawsOnlyValidCases(t *testing.T) {
	t.Parallel()
	op := &operator{ID: "X02a", Class: "X02", Kind: outOfScope}
	var rows []caseRow
	for i := range 30 {
		rows = append(rows, caseRow{Operator: op, Host: &host{ID: strconv.Itoa(i)}, Valid: i%2 == 0})
	}
	_, sample := readingSample(rows)
	if len(sample["X02a"]) == 0 {
		t.Fatal("readingSample drew nothing, want a sample")
	}
	for _, row := range sample["X02a"] {
		if !row.Valid {
			t.Errorf("readingSample drew the invalid case on host %s", row.Host.ID)
		}
	}
}
