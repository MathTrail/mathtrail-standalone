package checks_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// shipped is the test catalog, with the reference tasks a question is compared
// against.
type shipped struct {
	catalog
	references []string
}

func (s shipped) ReferenceQuestions(int) []string { return s.references }

// working is a solver that works its answer out: it finds the option saying
// what it computed, wherever the letters put it, the way a solver ending in
// match(options, value) does.
type working struct {
	value string
	runs  int
}

func (w *working) Run(_ context.Context, _ string, options solver.Options) (solver.Result, error) {
	w.runs++
	return solver.Result{
		Status: solver.StatusOK, Letters: options.Match(w.value), Steps: 1000, Duration: time.Millisecond,
	}, nil
}

// scripted is a solver that comes to the same result on every run, whatever
// the options: a letter written out by hand, or a failure.
type scripted struct {
	result solver.Result
	err    error
	runs   int
}

func (s *scripted) Run(context.Context, string, solver.Options) (solver.Result, error) {
	s.runs++
	return s.result, s.err
}

// program is the source handed in with every draft here. The fakes above run
// nothing, so what it says only has to be kept out of every refusal.
const program = "def solve(options):\n    return match(options, 6)\n"

// scenario is one submission and what it is judged against.
type scenario struct {
	draft        checks.Draft
	runner       solver.Runner
	references   []string
	fingerprints []string
	grade        int
	language     string
}

// accepted is the well-formed draft of the structure tests with a solver that
// finds its answer: a task that passes every check.
func accepted() scenario {
	return scenario{draft: validDraft(), runner: &working{value: "six pairs"}, grade: 4, language: "en"}
}

// review hands the scenario's draft in as the JSON a model would send, and
// judges it.
func (s *scenario) review(t *testing.T) checks.Outcome {
	t.Helper()

	reviewer := checks.NewReviewer(shipped{catalog: testCatalog, references: s.references}, s.runner,
		checks.DefaultDrawingLimits())
	examined, err := reviewer.Examine(t.Context(), submissionOf(t, s.draft))
	if err != nil {
		t.Fatalf("Examine() error = %v", err)
	}
	return reviewer.Judge(examined, checks.Against{
		Asked: asked(), Language: s.language, Grade: s.grade, Fingerprints: s.fingerprints,
	})
}

// submissionOf is a draft as the JSON a model would send.
func submissionOf(t *testing.T, draft checks.Draft) *checks.Submission {
	t.Helper()
	return &checks.Submission{
		Brief: jsonOf(t, draft.Brief), Task: jsonOf(t, draft.Task), SelfCheck: jsonOf(t, draft.SelfCheck),
		Solver: program,
	}
}

func jsonOf(t testing.TB, part any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(part)
	if err != nil {
		t.Fatalf("marshal %T: %v", part, err)
	}
	return raw
}

// codesOf are the codes an outcome reports, each once, in its order.
func codesOf(outcome *checks.Outcome) []checks.Code {
	var codes []checks.Code
	for _, reason := range outcome.Reasons() {
		codes = append(codes, reason.Code)
	}
	return codes
}

// longSentence is one sentence of twenty-four words, too long for a child of
// grades one and two.
const longSentence = " The ships fly past the moon and the stars and the comets and the planets " +
	"and the rockets and the satellites and home again."

func TestAGoodTaskIsAccepted(t *testing.T) {
	t.Parallel()

	runner := &working{value: "six pairs"}
	good := accepted()
	good.runner = runner
	outcome := good.review(t)

	if !outcome.Accepted() {
		t.Fatalf("problems %v and unchecked %v, want the task accepted", outcome.Problems, outcome.Unchecked)
	}
	if runner.runs != 2 {
		t.Errorf("the solver ran %d times, want twice: as handed in, and relabelled", runner.runs)
	}
	event := outcome.Event()
	if event.Outcome != "accepted" || event.Primary != "" || len(event.Failed) != 0 {
		t.Errorf("Event() = %+v, want an acceptance with nothing failed", event)
	}
	if event.SolverSteps != 2000 || event.SolverTime != 2*time.Millisecond {
		t.Errorf("Event() cost %d steps in %v, want the two runs together: 2000 in 2ms",
			event.SolverSteps, event.SolverTime)
	}
}

// refusals are the prototype's refusals, one reason each, and the ones this
// service added: each breaks one thing about a task that passes, and is
// refused for that thing alone.
var refusals = []struct {
	name   string
	change func(*scenario)
	code   checks.Code
	want   string
}{
	{"a brief for another topic", func(s *scenario) { s.draft.Brief.TargetConcept = "counting.gaps" },
		checks.CodeBadStructure, "not the topic this task was asked for"},
	{"a brief that drops an excluded skill", func(s *scenario) { s.draft.Brief.ExcludedSkills = []string{} },
		checks.CodeBadStructure, `dropped "division_with_remainder"`},
	{"two explanations saying the same", func(s *scenario) {
		s.draft.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: "You missed one pair."}
	}, checks.CodeDistractorExplanations, "say the same thing"},
	{"a drawing with a space at the end of a line", func(s *scenario) { draw(s, "P───Q \n│", "P", "Q") },
		checks.CodeDrawingFormat, "spaces at the end of line 1"},
	{"a drawing without a label it declares", func(s *scenario) { draw(s, "P───Q", "P", "Q", "R") },
		checks.CodeDrawingMismatch, `does not show "R"`},
	{"a sentence too long for the grade", func(s *scenario) { s.draft.Task.Question += longSentence; s.grade = 1 },
		checks.CodeReadability, "sentence 3 of task.question is 24 words long"},
	{"a solver that crashes", func(s *scenario) {
		s.runner = &scripted{result: solver.Result{Status: solver.StatusError, Message: "fail: six pairs is C"}}
	}, checks.CodeSolverError, "the solver failed while it ran"},
	{"a solver that finds another answer", func(s *scenario) { s.runner = &working{value: "four pairs"} },
		checks.CodeSolverDisagrees, "a different option from task.correct_answer"},
	{"a self-check with another answer", func(s *scenario) { s.draft.SelfCheck.FinalAnswer = "B" },
		checks.CodeSolverDisagrees, "self_check.final_answer is not task.correct_answer"},
	{"a self-check that finds the task unsolvable", func(s *scenario) { s.draft.SelfCheck.FinalAnswer = "UNSOLVABLE" },
		checks.CodeSolverDisagrees, "cannot be solved as written"},
	{"a blocking issue", func(s *scenario) {
		s.draft.SelfCheck.Issues = []checks.Issue{{Type: "ambiguous", Severity: "blocking", Comment: "C reads two ways."}}
	}, checks.CodeSelfCheckBlocking, "self_check.issues.0 (ambiguous) is blocking"},
	{"a copy of a reference task", func(s *scenario) { s.references = []string{s.draft.Task.Question} },
		checks.CodeNearDuplicate, "near-copy of one of the reference tasks"},
	{"a repeat of a task the child has had", func(s *scenario) {
		s.fingerprints = []string{checks.Fingerprint(s.draft.Task.Question)}
	}, checks.CodeNearDuplicate, "already been given"},
}

// draw gives a scenario's task a drawing, and a structure declaring these
// labels.
func draw(s *scenario, drawing string, labels ...string) {
	s.draft.Task.Drawing = drawing
	s.draft.Task.DrawingStructure = &checks.DrawingStructure{Kind: "number_line"}
	for _, label := range labels {
		s.draft.Task.DrawingStructure.Objects = append(s.draft.Task.DrawingStructure.Objects,
			checks.DrawingObject{ID: label, Label: label})
	}
}

func TestEachRefusalHasItsCode(t *testing.T) {
	t.Parallel()

	for _, test := range refusals {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			broken := accepted()
			test.change(&broken)
			outcome := broken.review(t)

			if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{test.code}) {
				t.Fatalf("codes = %v, want only %q: %v", codes, test.code, outcome.Problems)
			}
			if !mentions(outcome.Problems, test.want) {
				t.Errorf("problems = %v, want one saying %q", outcome.Problems, test.want)
			}
			if outcome.Accepted() || outcome.Primary() != test.code || outcome.Event().Outcome != "rejected" {
				t.Errorf("accepted %v, primary %q, event %+v, want a refusal counted by %q",
					outcome.Accepted(), outcome.Primary(), outcome.Event(), test.code)
			}
		})
	}
}

// All nine refusals at once come back all nine, in the order the checks run:
// the first is the one an attempt is counted by.
func TestEveryFailedCheckIsReportedInItsOrder(t *testing.T) {
	t.Parallel()

	broken := accepted()
	for _, test := range refusals {
		if test.name != "a solver that finds another answer" && test.name != "a self-check that finds the task unsolvable" &&
			test.name != "a repeat of a task the child has had" {
			test.change(&broken)
		}
	}
	draw(&broken, "P───Q \n│", "P", "Q", "R")
	broken.references = []string{broken.draft.Task.Question}
	outcome := broken.review(t)

	want := []checks.Code{
		checks.CodeBadStructure, checks.CodeDistractorExplanations, checks.CodeDrawingFormat, checks.CodeDrawingMismatch,
		checks.CodeReadability, checks.CodeSolverError, checks.CodeSolverDisagrees, checks.CodeSelfCheckBlocking,
		checks.CodeNearDuplicate,
	}
	if codes := codesOf(&outcome); !slices.Equal(codes, want) {
		t.Fatalf("codes = %v, want %v", codes, want)
	}
	if event := outcome.Event(); event.Primary != checks.CodeBadStructure || !slices.Equal(event.Failed, want) {
		t.Errorf("Event() = %+v, want every code, counted by the first", event)
	}
}

// A fault of the structure keeps from running only the checks it took the
// input of. Five options missing one is nothing to run a solver on, so the
// solver is not run and the model is told so; a brief asking for another topic
// leaves the task itself whole, and everything else about it is checked in
// the same attempt.
func TestAStructuralFaultBlocksOnlyWhatItBroke(t *testing.T) {
	t.Parallel()

	crash := &scripted{result: solver.Result{Status: solver.StatusError, Message: "never run"}}
	noE := accepted()
	noE.runner = crash
	delete(noE.draft.Task.Options, "E")
	outcome := noE.review(t)
	if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{checks.CodeBadStructure}) {
		t.Errorf("an option missing: codes = %v, want only %q", codes, checks.CodeBadStructure)
	}
	if crash.runs != 0 || !slices.ContainsFunc(outcome.Unchecked, func(note string) bool {
		return strings.Contains(note, "the solver was not run")
	}) {
		t.Errorf("an option missing: the solver ran %d times, unchecked %v, want it not run and said so",
			crash.runs, outcome.Unchecked)
	}

	crash = &scripted{result: solver.Result{Status: solver.StatusError}}
	elsewhere := accepted()
	elsewhere.runner = crash
	elsewhere.draft.Brief.TargetConcept = "counting.gaps"
	elsewhere.draft.Task.Question += longSentence
	elsewhere.grade = 1
	outcome = elsewhere.review(t)
	want := []checks.Code{checks.CodeBadStructure, checks.CodeReadability, checks.CodeSolverError}
	if codes := codesOf(&outcome); !slices.Equal(codes, want) || crash.runs != 1 {
		t.Errorf("another topic: codes = %v after %d runs, want %v after one", codes, crash.runs, want)
	}
}

// What could not be read is not checked, and the model is told what each
// check that did not run needs.
func TestWhatCouldNotBeReadIsNotChecked(t *testing.T) {
	t.Parallel()

	missing := accepted()
	missing.draft.Task, missing.draft.SelfCheck = nil, nil
	outcome := missing.review(t)

	for _, want := range []string{
		"the explanations behind the wrong options were not checked", "readability was not checked",
		"the solver was not run", "the answers were not compared", "the self-check was not checked",
		"the question was not compared with earlier tasks",
	} {
		if !slices.ContainsFunc(outcome.Unchecked, func(note string) bool { return strings.Contains(note, want) }) {
			t.Errorf("unchecked = %v, want a note saying %q", outcome.Unchecked, want)
		}
	}
	if outcome.Accepted() || outcome.Primary() != checks.CodeBadStructure {
		t.Errorf("accepted %v, primary %q, want a refusal of the structure", outcome.Accepted(), outcome.Primary())
	}
}

// A drawing handed in without the structure that describes it cannot be held
// against the wording: the structure check refuses the missing part, and the
// match is noted as not checked rather than passed.
func TestADrawingWithoutItsStructureIsNotMatched(t *testing.T) {
	t.Parallel()

	undescribed := accepted()
	undescribed.draft.Task.Drawing = "P───Q"
	outcome := undescribed.review(t)

	if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{checks.CodeBadStructure}) {
		t.Errorf("codes = %v, want only %q", codes, checks.CodeBadStructure)
	}
	if !slices.ContainsFunc(outcome.Unchecked, func(note string) bool {
		return strings.Contains(note, "the drawing was not checked against the wording")
	}) {
		t.Errorf("unchecked = %v, want the match noted as not checked", outcome.Unchecked)
	}
}

// What is out of the format is the structure check's to refuse, and it is
// refused once. A final answer that is no answer at all is not also a
// disagreement; a blocking issue of a type the format does not have still
// blocks, and is pointed at without its type, which is the model's own words.
func TestWhatIsOutOfTheFormatIsRefusedOnce(t *testing.T) {
	t.Parallel()

	noAnswer := accepted()
	noAnswer.draft.SelfCheck.FinalAnswer = "maybe"
	outcome := noAnswer.review(t)
	if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{checks.CodeBadStructure}) {
		t.Errorf("a final answer that is no answer: codes = %v, want only %q", codes, checks.CodeBadStructure)
	}

	strange := accepted()
	strange.draft.SelfCheck.Issues = []checks.Issue{{Type: "C is wrong", Severity: "blocking", Comment: "It is."}}
	outcome = strange.review(t)
	want := []checks.Code{checks.CodeBadStructure, checks.CodeSelfCheckBlocking}
	if codes := codesOf(&outcome); !slices.Equal(codes, want) {
		t.Fatalf("a blocking issue of no known type: codes = %v, want %v", codes, want)
	}
	blocking := outcome.Reasons()[1].Messages
	if !slices.Equal(blocking, []string{
		"self_check.issues.0 is blocking; resolve what it describes before handing the task in",
	}) {
		t.Errorf("the blocking issue: %q, want it pointed at by its place alone", blocking)
	}
}

// A solver that does not run to an answer is told which way it failed and, for
// a limit, which limit — never what the interpreter or the program said.
func TestEachWayASolverFailsIsToldApart(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		status  solver.Status
		message string
		want    string
	}{
		{solver.StatusBadSource, "got string literal \"six pairs\"", "does not parse"},
		{solver.StatusNoEntryPoint, "solve is a string rather than a function", "defines no solve to call"},
		{solver.StatusError, "fail: the answer is C", "failed while it ran"},
		{solver.StatusBadOutput, "solve returned C twice", "something other than a list of distinct option letters"},
		{solver.StatusTimeout, "the solver took more than 25000000 steps", "the solver took more than 25000000 steps"},
	} {
		t.Run(string(test.status), func(t *testing.T) {
			t.Parallel()

			failing := accepted()
			failing.runner = &scripted{result: solver.Result{Status: test.status, Message: test.message}}
			outcome := failing.review(t)

			if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{checks.CodeSolverError}) {
				t.Fatalf("codes = %v, want only %q", codes, checks.CodeSolverError)
			}
			if !mentions(outcome.Problems, test.want) {
				t.Errorf("problems = %v, want one saying %q", outcome.Problems, test.want)
			}
			if test.status != solver.StatusTimeout && mentions(outcome.Problems, test.message) {
				t.Errorf("problems = %v, want the run's own words left out", outcome.Problems)
			}
		})
	}
}

// A run that failed in a way the review has no sentence for is still refused
// in words: no refusal leaves the model with nothing to act on.
func TestAFailureOfNoKnownKindIsStillTold(t *testing.T) {
	t.Parallel()

	odd := accepted()
	odd.runner = &scripted{result: solver.Result{Status: solver.Status("exploded")}}
	outcome := odd.review(t)

	if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{checks.CodeSolverError}) {
		t.Fatalf("codes = %v, want only %q", codes, checks.CodeSolverError)
	}
	if !mentions(outcome.Problems, "the solver did not run to an answer") {
		t.Errorf("problems = %v, want the failure told in words", outcome.Problems)
	}
}

// A solver that ran and disagrees is told how it disagreed, without a letter.
func TestEachWayASolverDisagreesIsToldApart(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		runner solver.Runner
		want   string
	}{
		{"no option fits", &scripted{result: solver.Result{Status: solver.StatusOK, Letters: []string{}}},
			"found no option that fits"},
		{"two options fit", &scripted{result: solver.Result{Status: solver.StatusOK, Letters: []string{"A", "C"}}},
			"found more than one option that fits"},
		{"the letter is written out by hand", &scripted{result: solver.Result{Status: solver.StatusOK, Letters: []string{"C"}}},
			"pointed elsewhere once the options were relabelled"},
		{"another answer is worked out", &working{value: "twelve pairs"},
			"a different option from task.correct_answer"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			disagreeing := accepted()
			disagreeing.runner = test.runner
			outcome := disagreeing.review(t)

			if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{checks.CodeSolverDisagrees}) {
				t.Fatalf("codes = %v, want only %q", codes, checks.CodeSolverDisagrees)
			}
			if !mentions(outcome.Problems, test.want) {
				t.Errorf("problems = %v, want one saying %q", outcome.Problems, test.want)
			}
		})
	}
}

// A minor issue lets the task through. It is counted by its type, and what the
// self-check said about it is dropped: the service keeps no text of a task.
// An issue of a type the format does not have is the structure's to refuse,
// and its words are not counted either.
func TestAMinorIssueIsCountedByItsType(t *testing.T) {
	t.Parallel()

	minor := accepted()
	minor.draft.SelfCheck.Issues = []checks.Issue{
		{Type: "too_hard_for_grade", Severity: "minor", Comment: "Maybe long for grade 3."},
	}
	outcome := minor.review(t)
	if !outcome.Accepted() || !slices.Equal(outcome.MinorIssues, []string{"too_hard_for_grade"}) {
		t.Fatalf("accepted %v with minor issues %v, want the task accepted and the type kept",
			outcome.Accepted(), outcome.MinorIssues)
	}
	if event := outcome.Event(); !slices.Equal(event.MinorIssues, []string{"too_hard_for_grade"}) {
		t.Errorf("Event().MinorIssues = %v, want the type", event.MinorIssues)
	}

	minor.draft.SelfCheck.Issues[0].Type = "C is the right one"
	outcome = minor.review(t)
	if len(outcome.MinorIssues) != 0 || len(outcome.Event().MinorIssues) != 0 {
		t.Errorf("an issue of no known type: minor issues %v, want none counted", outcome.MinorIssues)
	}
}

// Whether the Flesch–Kincaid grade applies is the language's to say: English
// words in a task declared Russian are held to sentence length alone, as the
// prototype held its tasks in Russian.
func TestATaskInRussianIsNotHeldToFleschKincaid(t *testing.T) {
	t.Parallel()

	hard := accepted()
	hard.grade = 1
	hard.draft.Task.Question = "Every afternoon seven classmates exchange colourful postcards. Each classmate " +
		"sends exactly one postcard to every other classmate. How many postcards are delivered altogether?"

	english := hard.review(t)
	if codes := codesOf(&english); !slices.Equal(codes, []checks.Code{checks.CodeReadability}) {
		t.Errorf("in English: codes = %v, want only %q", codes, checks.CodeReadability)
	}
	hard.language = "ru"
	if russian := hard.review(t); !russian.Accepted() {
		t.Errorf("in Russian: problems %v, want the task accepted", russian.Problems)
	}
}

// The sandbox failing is not the task's fault: the review stops with an error
// and no outcome, and the model has no attempt counted against it.
func TestTheSandboxFailingIsNotTheTasks(t *testing.T) {
	t.Parallel()

	down := errors.New("no slot in the sandbox")
	reviewer := checks.NewReviewer(shipped{catalog: testCatalog}, &scripted{err: down}, checks.DefaultDrawingLimits())
	if _, err := reviewer.Examine(t.Context(), submissionOf(t, validDraft())); !errors.Is(err, down) {
		t.Fatalf("Examine() error = %v, want it to wrap %v", err, down)
	}
}

// Problems are grouped by the check that found them, so that a tool can list
// each reason once with all it found.
func TestReasonsGroupTheProblemsByCheck(t *testing.T) {
	t.Parallel()

	copied := accepted()
	copied.references = []string{copied.draft.Task.Question}
	copied.fingerprints = []string{checks.Fingerprint(copied.draft.Task.Question)}
	outcome := copied.review(t)

	reasons := outcome.Reasons()
	if len(reasons) != 1 || reasons[0].Code != checks.CodeNearDuplicate || len(reasons[0].Messages) != 2 {
		t.Errorf("Reasons() = %+v, want one reason of %q with both of its messages", reasons, checks.CodeNearDuplicate)
	}
}

// Each code is one reason however its problems are spread through the list,
// so that the log names every failed check once.
func TestEachCodeIsOneReasonWhereverItsProblemsStand(t *testing.T) {
	t.Parallel()

	outcome := checks.Outcome{Problems: []checks.Problem{
		{Code: checks.CodeBadStructure, Message: "first"},
		{Code: checks.CodeReadability, Message: "second"},
		{Code: checks.CodeBadStructure, Message: "third"},
	}}
	want := []checks.Reason{
		{Code: checks.CodeBadStructure, Messages: []string{"first", "third"}},
		{Code: checks.CodeReadability, Messages: []string{"second"}},
	}
	if got := outcome.Reasons(); !slices.EqualFunc(got, want, func(a, b checks.Reason) bool {
		return a.Code == b.Code && slices.Equal(a.Messages, b.Messages)
	}) {
		t.Errorf("Reasons() = %+v, want %+v", got, want)
	}
	if failed := outcome.Event().Failed; !slices.Equal(failed, []checks.Code{checks.CodeBadStructure, checks.CodeReadability}) {
		t.Errorf("Event().Failed = %v, want each code once, in the order it first appears", failed)
	}
}

// Two minor issues of one type count as two: the log counts the issues a
// self-check let through by their type, one for each.
func TestTwoMinorIssuesOfOneTypeCountAsTwo(t *testing.T) {
	t.Parallel()

	minor := accepted()
	minor.draft.SelfCheck.Issues = []checks.Issue{
		{Type: "ambiguous", Severity: "minor", Comment: "One reading of it."},
		{Type: "ambiguous", Severity: "minor", Comment: "Another reading."},
	}
	outcome := minor.review(t)
	if !outcome.Accepted() || !slices.Equal(outcome.MinorIssues, []string{"ambiguous", "ambiguous"}) {
		t.Errorf("accepted %v with minor issues %v, want the task accepted and the type counted twice",
			outcome.Accepted(), outcome.MinorIssues)
	}
}

// Nothing the review says, and nothing the log keeps of it, repeats a word the
// model wrote that could give the answer away: no option letter, no option's
// text, no comment of the self-check, nothing the program said. A refusal
// reaches the widget, and the log keeps no text of a task at all.
func TestNothingOfTheTaskLeavesTheReview(t *testing.T) {
	t.Parallel()

	written := []string{"four pairs", "five pairs", "six pairs", "eight pairs", "twelve pairs",
		"C reads two ways", "the answer is", "six pairs is C", program}
	for _, test := range refusals {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			broken := accepted()
			test.change(&broken)
			outcome := broken.review(t)

			repeatsNone(t, append(slices.Clone(outcome.Unchecked), messagesOf(outcome.Problems)...), written)
			logged, err := json.Marshal(outcome.Event())
			if err != nil {
				t.Fatalf("marshal the event: %v", err)
			}
			repeatsNone(t, []string{string(logged)},
				append(slices.Clone(written), broken.draft.Task.Question, broken.draft.Task.Solution))
		})
	}
}

// repeatsNone fails the test for every sentence that names an option letter or
// repeats any of these words.
func repeatsNone(t *testing.T, sentences, words []string) {
	t.Helper()

	for _, sentence := range sentences {
		if letter.MatchString(sentence) {
			t.Errorf("%q names a letter", sentence)
		}
		for _, repeated := range words {
			if strings.Contains(sentence, repeated) {
				t.Errorf("%q repeats %q", sentence, repeated)
			}
		}
	}
}

func messagesOf(problems []checks.Problem) []string {
	messages := make([]string, len(problems))
	for i, problem := range problems {
		messages[i] = problem.Message
	}
	return messages
}
