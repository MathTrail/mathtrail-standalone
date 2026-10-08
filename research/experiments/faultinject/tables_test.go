package main

import (
	"cmp"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// refusedBy is the verdict of a review in which these checks refused, the
// first of them reported first, as a review reports them.
func refusedBy(codes ...checks.Code) verdict {
	if len(codes) == 0 {
		return verdict{}
	}
	return verdict{Codes: codes, Primary: codes[0]}
}

// isRefused is the condition a share of refused cases counts.
func isRefused(row *caseRow) bool { return row.Verdict.Refused() }

// row is a line of a table written as the CSV file writes it.
func row(line string) []string { return strings.Split(line, ",") }

// assertTable compares a table with the one wanted, row by row.
func assertTable(t *testing.T, name string, got, want [][]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s has %d rows, want %d: %q", name, len(got), len(want), got)
	}
	for i := range want {
		if !slices.Equal(got[i], want[i]) {
			t.Errorf("%s row %d = %q, want %q", name, i, got[i], want[i])
		}
	}
}

// A class's row gives each share twice over: with the interval of the
// bootstrap over its hosts, and with the exact interval of its pooled cases.
// Where every host has the same share, as at 0 and 100 %, the bootstrap has no
// width, and only the exact interval says how few cases the share rests on. A
// case the run found not to be its defect counts in neither.
func TestAClassRowCarriesTheExactIntervalWhereTheBootstrapHasNoWidth(t *testing.T) {
	t.Parallel()
	d01 := &operator{ID: "D01a", Class: "D01", Kind: mechanism, Expected: []checks.Code{checks.CodeSolverDisagrees}}
	x01 := &operator{ID: "X01a", Class: "X01", Kind: outOfScope}
	a, b, c := &host{ID: "a"}, &host{ID: "b"}, &host{ID: "c"}
	caught := refusedBy(checks.CodeSolverDisagrees)
	rows := []caseRow{
		{Operator: x01, Host: c, Valid: true},
		{Operator: d01, Host: a, Valid: true, Verdict: caught},
		{Operator: d01, Host: a, Valid: true, Verdict: caught},
		{Operator: d01, Host: b},
		{Operator: d01, Host: b, Valid: true, Verdict: caught},
		{Operator: d01, Host: b, Valid: true, Verdict: caught},
		{Operator: x01, Host: c, Valid: true, Verdict: verdict{Unchecked: []string{"the drawing was not checked"}}},
	}
	assertTable(t, "classTable", classTable(rows), [][]string{
		row("class,valid,refused_percent,refused_low,refused_high,refused_exact_low,refused_exact_high," +
			"expected_percent,expected_low,expected_high,expected_exact_low,expected_exact_high,unchecked_percent"),
		row("D01,4,100.0,100.0,100.0,39.8,100.0,100.0,100.0,100.0,39.8,100.0,0.0"),
		row("X01,2,0.0,0.0,0.0,0.0,84.2,0.0,0.0,0.0,0.0,84.2,50.0"),
	})
}

// casesOn are the valid cases of an operator on hosts h0, h1, … with these
// tallies: so many cases on each host, the first so many of them refused.
func casesOn(op *operator, tallies ...share) []caseRow {
	var rows []caseRow
	for i, tally := range tallies {
		on := &host{ID: "h" + strconv.Itoa(i)}
		for j := range tally.total {
			refused := verdict{}
			if j < tally.hits {
				refused = refusedBy(checks.CodeBadStructure)
			}
			rows = append(rows, caseRow{Operator: op, Host: on, Valid: true, Verdict: refused})
		}
	}
	return rows
}

// mixedClass is a class whose five hosts came out differently: 7 of its 10
// valid cases refused in all.
func mixedClass(op *operator) []caseRow {
	return casesOn(op, share{1, 2}, share{2, 2}, share{0, 1}, share{3, 4}, share{1, 1})
}

// The bootstrap draws hosts, each with all its cases: where every host has the
// same share, so has every draw, and the interval is that share alone.
func TestTheBootstrapOfHostsThatAgreeIsAPoint(t *testing.T) {
	t.Parallel()
	rows := casesOn(&operator{ID: "D02a", Class: "D02"}, share{1, 2}, share{2, 4}, share{3, 6})
	if got, want := pooledInterval(rows, "D02", isRefused), []string{"50.0", "50.0"}; !slices.Equal(got, want) {
		t.Errorf("pooledInterval of three hosts at 50 %% = %q, want %q", got, want)
	}
}

// Only the valid cases of the class count: cases of another class, and cases
// the run found not to be their defect, leave the interval as it was to the
// last digit, and so does asking for it again.
func TestTheBootstrapCountsOnlyTheValidCasesOfItsClass(t *testing.T) {
	t.Parallel()
	op, other := &operator{ID: "D02a", Class: "D02"}, &operator{ID: "D03a", Class: "D03"}
	rows := mixedClass(op)
	want := pooledInterval(rows, "D02", isRefused)
	if again := pooledInterval(rows, "D02", isRefused); !slices.Equal(again, want) {
		t.Errorf("pooledInterval asked twice = %q and %q, want one answer", want, again)
	}
	// The cases that do not count come first, on the class's hosts and on a
	// host of their own: had they been counted, they would change the hosts
	// drawn as well as the shares.
	noise := []caseRow{
		{Operator: op, Host: &host{ID: "h4"}, Verdict: refusedBy(checks.CodeBadStructure)},
		{Operator: op, Host: &host{ID: "stray"}},
		{Operator: other, Host: &host{ID: "h0"}, Valid: true, Verdict: refusedBy(checks.CodeSolverDisagrees)},
		{Operator: other, Host: &host{ID: "stray"}, Valid: true},
	}
	if got := pooledInterval(slices.Concat(noise, rows), "D02", isRefused); !slices.Equal(got, want) {
		t.Errorf("pooledInterval with cases that do not count = %q, want %q", got, want)
	}
}

// The interval drawn for a class holds the share of its pooled cases.
func TestTheBootstrapIntervalHoldsThePooledShare(t *testing.T) {
	t.Parallel()
	rows := mixedClass(&operator{ID: "D02a", Class: "D02"})
	got := pooledInterval(rows, "D02", isRefused)
	low, lowErr := strconv.ParseFloat(got[0], 64)
	high, highErr := strconv.ParseFloat(got[1], 64)
	if lowErr != nil || highErr != nil {
		t.Fatalf("pooledInterval = %q, want two percentages", got)
	}
	// Rounding to a tenth keeps the order, so the pooled share is compared as
	// it would be written.
	pooled, _ := strconv.ParseFloat(percent(7.0/10), 64)
	if low > pooled || pooled > high || low == high {
		t.Errorf("pooledInterval = %q, want an interval around the pooled %.1f %%", got, pooled)
	}
}

// A class with no valid case has nothing to draw: its interval is the whole
// range of shares, from 0 to 100 %.
func TestTheBootstrapOfAClassWithoutCasesIsTheWholeRange(t *testing.T) {
	t.Parallel()
	rows := mixedClass(&operator{ID: "D02a", Class: "D02"})
	if got, want := pooledInterval(rows, "D09", isRefused), []string{"0.0", "100.0"}; !slices.Equal(got, want) {
		t.Errorf("pooledInterval of a class without cases = %q, want %q", got, want)
	}
}

// The checks table has a row for every check of every class, in the order the
// service reports its checks in, and in each the shares of the class's valid
// cases the check refused, refused first, and refused alone. A class with no
// valid case has no shares, rather than shares of 0 %: nothing was measured.
func TestTheChecksTableSharesOutTheRefusalsOfAClass(t *testing.T) {
	t.Parallel()
	d02, x01 := &operator{ID: "D02a", Class: "D02"}, &operator{ID: "X01a", Class: "X01"}
	on := &host{ID: "a"}
	rows := []caseRow{
		{Operator: d02, Host: on, Valid: true, Verdict: refusedBy(checks.CodeBadStructure)},
		{Operator: d02, Host: on, Valid: true, Verdict: refusedBy(checks.CodeBadStructure, checks.CodeSolverDisagrees)},
		{Operator: d02, Host: on, Valid: true, Verdict: refusedBy(checks.CodeSolverDisagrees)},
		{Operator: d02, Host: on, Valid: true},
		{Operator: d02, Host: on, Verdict: refusedBy(checks.CodeReadability)},
		{Operator: x01, Host: on, Verdict: refusedBy(checks.CodeReadability)},
	}
	want := [][]string{row("class,check,fired_percent,first_percent,only_percent")}
	for _, code := range codeOrder {
		shares := map[checks.Code]string{
			checks.CodeBadStructure:    "50.0,50.0,25.0",
			checks.CodeSolverDisagrees: "50.0,25.0,25.0",
		}[code]
		want = append(want, row("D02,"+string(code)+","+cmp.Or(shares, "0.0,0.0,0.0")))
	}
	for _, code := range codeOrder {
		want = append(want, row("X01,"+string(code)+",,,"))
	}
	assertTable(t, "checkTable", checkTable(rows), want)
}

// randomClass is a class of eight valid cases, each refused by up to three
// checks, the first of them reported first, beside up to two cases that do not
// count; and the share of its valid cases refused, in percent.
func randomClass(rng *rand.Rand, op *operator) (rows []caseRow, refused float64) {
	for i := range 8 + rng.IntN(3) {
		codes := slices.Clone(codeOrder)
		rng.Shuffle(len(codes), func(x, y int) { codes[x], codes[y] = codes[y], codes[x] })
		valid, judged := i < 8, refusedBy(codes[:rng.IntN(4)]...)
		if valid && judged.Refused() {
			refused += 100.0 / 8
		}
		rows = append(rows, caseRow{Operator: op, Host: &host{ID: "h" + strconv.Itoa(i)}, Valid: valid, Verdict: judged})
	}
	return rows, refused
}

// Whatever the checks a review reports, a check that refused a case alone
// refused it first, and one that refused it first refused it; and as every
// refused case has exactly one first check, the shares refused first add up to
// the share of the class refused at all.
func TestTheChecksTableAddsUpToTheShareRefused(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(1, 2))
	op := &operator{ID: "D02a", Class: "D02"}
	for range 500 {
		rows, refused := randomClass(rng, op)
		sum := 0.0
		for _, line := range checkTable(rows)[1:] {
			fired, first, only := parsePercent(t, line[2]), parsePercent(t, line[3]), parsePercent(t, line[4])
			if only > first || first > fired {
				t.Fatalf("checkTable row %q: want only ≤ first ≤ fired", line)
			}
			sum += first
		}
		// Eight cases make every share a whole number of eighths, which a
		// tenth of a percent writes exactly.
		if sum != refused {
			t.Fatalf("the shares refused first add up to %.1f %%, want the %.1f %% refused", sum, refused)
		}
	}
}

// parsePercent reads a share as a table writes it.
func parsePercent(t *testing.T, text string) float64 {
	t.Helper()
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		t.Fatalf("%q is not a percentage: %v", text, err)
	}
	return value
}

// The operators table has one row per operator, in the order of the operators
// and whether or not an operator made a case: one that made none has its counts
// at zero and its shares empty rather than at 0 %. It counts every case an
// operator applied to, valid or not, and gives its shares over the valid ones;
// the counts of a reading appear only for an operator a person read.
func TestTheOperatorsTableHasARowForEveryOperator(t *testing.T) {
	t.Parallel()
	disagrees := []checks.Code{checks.CodeSolverDisagrees}
	ops := []operator{
		{ID: "D01a", Class: "D01", Kind: mechanism, Expected: disagrees},
		{ID: "D02a", Class: "D02", Kind: mechanism, Expected: []checks.Code{checks.CodeBadStructure, checks.CodeSolverDisagrees}},
		{ID: "D03a", Class: "D03", Kind: mechanism, Expected: disagrees},
		{ID: "X02a", Class: "X02", Kind: outOfScope},
	}
	a, b, c := &host{ID: "a"}, &host{ID: "b"}, &host{ID: "c"}
	rows := []caseRow{
		{Operator: &ops[3], Host: a, Valid: true, Verdict: refusedBy(checks.CodeReadability)},
		{Operator: &ops[3], Host: b, Valid: true},
		{Operator: &ops[1], Host: a, Valid: true, Verdict: refusedBy(checks.CodeBadStructure)},
		{Operator: &ops[0], Host: a, Valid: true, Verdict: refusedBy(checks.CodeSolverDisagrees)},
		{Operator: &ops[0], Host: b, Valid: true, Verdict: refusedBy(checks.CodeSolverDisagrees)},
		{Operator: &ops[0], Host: c},
	}
	read := map[string]share{"X02a": {hits: 1, total: 2}}
	assertTable(t, "operatorTable", operatorTable(rows, ops, read), [][]string{
		row("operator,class,kind,expected_codes,applied,valid,refused_percent,refused_low,refused_high," +
			"expected_percent,expected_low,expected_high,read,equivalent"),
		row("D01a,D01,mechanism,solver_disagrees,3,2,100.0,15.8,100.0,100.0,15.8,100.0,,"),
		row("D02a,D02,mechanism,bad_structure;solver_disagrees,1,1,100.0,2.5,100.0,100.0,2.5,100.0,,"),
		row("D03a,D03,mechanism,solver_disagrees,0,0,,,,,,,,"),
		row("X02a,X02,out_of_scope,,2,2,50.0,1.3,98.7,0.0,0.0,84.2,2,1"),
	})
}
