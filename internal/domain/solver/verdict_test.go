package solver_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// scripted is a runner that answers from a list rather than by computing:
// each run takes the next answer down. It stands in for the sandbox because
// what is being tested is what two answers mean together, and a real
// interpreter would only make it harder to write down the pair that matters.
type scripted struct {
	answers []solver.Result
	err     error
	// errAfter is how many runs go through before the error is returned, so
	// that a test can break the second run rather than the first.
	errAfter int
	// asked records the options of every run, in order, so that a test can
	// check the second run was given the shifted ones.
	asked []solver.Options
}

func (s *scripted) Run(_ context.Context, _ string, options solver.Options) (solver.Result, error) {
	s.asked = append(s.asked, options)
	if s.err != nil && len(s.asked) > s.errAfter {
		return solver.Result{}, s.err
	}
	if len(s.answers) == 0 {
		return solver.Result{}, errors.New("the test scripted fewer answers than the rule asked for")
	}
	answer := s.answers[0]
	s.answers = s.answers[1:]
	return answer, nil
}

// ok is a run that reached the letters given.
func ok(letters ...string) solver.Result {
	return solver.Result{Status: solver.StatusOK, Letters: letters}
}

func TestVerdict(t *testing.T) {
	t.Parallel()
	for _, test := range []verdictCase{
		{
			name:    "both runs follow the option and agree",
			answers: []solver.Result{ok("A"), ok("C")},
			letter:  "A",
			runs:    2,
		},
		{
			name:    "the last letter of the round agrees with the first",
			answers: []solver.Result{ok("E"), ok("B")},
			letter:  "E",
			runs:    2,
		},
		{
			name:     "a letter written out by hand comes back unmoved",
			answers:  []solver.Result{ok("C"), ok("C")},
			runs:     2,
			contains: "reading the letters",
		},
		{
			name:     "the second run finds nothing",
			answers:  []solver.Result{ok("A"), ok()},
			runs:     2,
			contains: "at nothing",
		},
		{
			name:     "the second run finds two",
			answers:  []solver.Result{ok("A"), ok("C", "D")},
			runs:     2,
			contains: "at C and D",
		},
		{
			name:     "the second run fails",
			answers:  []solver.Result{ok("A"), {Status: solver.StatusError, Message: "list index out of range"}},
			runs:     2,
			contains: "list index out of range",
		},
		{
			name:     "the first run fails, so there is no second",
			answers:  []solver.Result{{Status: solver.StatusTimeout, Message: "the solver took more than 10 steps"}},
			runs:     1,
			contains: "more than 10 steps",
		},
		{
			name:     "the first run finds no option that fits",
			answers:  []solver.Result{ok()},
			runs:     1,
			contains: "no option that fits",
		},
		{
			name:     "the first run finds two",
			answers:  []solver.Result{ok("A", "B")},
			runs:     1,
			contains: "more than one option that fits: A, B",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			runner := &scripted{answers: test.answers}

			agreement, err := solver.Verdict(context.Background(), runner, "irrelevant", task)
			if err != nil {
				t.Fatalf("Verdict: got error %v, want none", err)
			}
			test.check(t, agreement, len(runner.asked))
		})
	}
}

// verdictCase is a pair of scripted runs and what the rule has to make of
// them: which letter, how many runs it took to get there, and what the
// explanation has to mention when it did not.
type verdictCase struct {
	name     string
	answers  []solver.Result
	letter   string
	runs     int
	contains string
}

func (c *verdictCase) check(t *testing.T, agreement solver.Agreement, asked int) {
	t.Helper()
	if agreement.Letter != c.letter {
		t.Errorf("letter: got %q, want %q", agreement.Letter, c.letter)
	}
	if agreement.Agreed() != (c.letter != "") {
		t.Errorf("Agreed: got %v for letter %q", agreement.Agreed(), agreement.Letter)
	}
	if len(agreement.Runs) != c.runs {
		t.Errorf("runs: got %d, want %d", len(agreement.Runs), c.runs)
	}
	if asked != c.runs {
		t.Errorf("runs asked for: got %d, want %d", asked, c.runs)
	}
	if c.contains != "" && !strings.Contains(agreement.Explain(), c.contains) {
		t.Errorf("explanation: got %q, want it to mention %q", agreement.Explain(), c.contains)
	}
	if agreement.Agreed() && agreement.Explain() != "" {
		t.Errorf("explanation: got %q for an agreement, want none", agreement.Explain())
	}
}

func TestTheSecondRunSeesTheShiftedOptions(t *testing.T) {
	t.Parallel()
	runner := &scripted{answers: []solver.Result{ok("A"), ok("C")}}

	if _, err := solver.Verdict(context.Background(), runner, "irrelevant", task); err != nil {
		t.Fatalf("Verdict: got error %v, want none", err)
	}
	if len(runner.asked) != 2 {
		t.Fatalf("runs asked for: got %d, want 2", len(runner.asked))
	}
	if runner.asked[0] != task {
		t.Errorf("first run: got %v, want the options as they are", runner.asked[0])
	}
	if runner.asked[1] != task.Rotated() {
		t.Errorf("second run: got %v, want %v", runner.asked[1], task.Rotated())
	}
}

func TestVerdictPassesOnWhatTheRunnerCouldNotDo(t *testing.T) {
	t.Parallel()
	broken := errors.New("no free slot")

	if _, err := solver.Verdict(context.Background(), &scripted{err: broken}, "irrelevant", task); !errors.Is(err, broken) {
		t.Fatalf("Verdict: got error %v, want %v", err, broken)
	}
}

func TestVerdictPassesOnWhatBrokeTheSecondRun(t *testing.T) {
	t.Parallel()
	broken := errors.New("no free slot")
	runner := &scripted{answers: []solver.Result{ok("A")}, err: broken, errAfter: 1}

	if _, err := solver.Verdict(context.Background(), runner, "irrelevant", task); !errors.Is(err, broken) {
		t.Fatalf("Verdict: got error %v, want %v", err, broken)
	}
}

func TestFailedNamesTheRunThatDidNotGetThere(t *testing.T) {
	t.Parallel()
	timedOut := solver.Result{Status: solver.StatusTimeout, Message: "the solver ran longer than 2s"}
	runner := &scripted{answers: []solver.Result{ok("A"), timedOut}}

	agreement, err := solver.Verdict(context.Background(), runner, "irrelevant", task)
	if err != nil {
		t.Fatalf("Verdict: got error %v, want none", err)
	}
	failed := agreement.Failed()
	if failed == nil {
		t.Fatal("Failed: got nil, want the second run")
	}
	if failed.Status != solver.StatusTimeout {
		t.Errorf("Failed: got %q, want %q", failed.Status, solver.StatusTimeout)
	}

	agreed := solver.Agreement{Letter: "A", Runs: []solver.Result{ok("A"), ok("C")}}
	if failed := agreed.Failed(); failed != nil {
		t.Errorf("Failed: got %+v for two runs that finished, want nil", failed)
	}
}

func TestAnAgreementWithNoRuns(t *testing.T) {
	t.Parallel()
	var nothing solver.Agreement

	if nothing.Agreed() {
		t.Error("Agreed: got true for an agreement with no runs")
	}
	if got := nothing.Explain(); got != "the solver did not run" {
		t.Errorf("Explain: got %q, want it to say the solver did not run", got)
	}
}

func TestOne(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		result solver.Result
		want   bool
	}{
		{"one letter", ok("A"), true},
		{"none", ok(), false},
		{"two", ok("A", "B"), false},
		{"one letter of a run that failed", solver.Result{Status: solver.StatusError, Letters: []string{"A"}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.result.One(); got != test.want {
				t.Errorf("One() = %v, want %v", got, test.want)
			}
		})
	}
}
