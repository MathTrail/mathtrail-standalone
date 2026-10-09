package checks_test

import (
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// faulty is a task that passes, being broken one fault at a time. What more
// than one fault touches — the picture, the question — is settled once they
// have all been applied.
type faulty struct {
	scenario
	outOfFormat, strayLabel, copied bool
}

// faults are one way each to fail every check, none of them in the way of
// another: an idea left unsaid leaves the task itself whole.
var faults = []struct {
	code  checks.Code
	apply func(*faulty)
}{
	{checks.CodeBadStructure, func(f *faulty) { f.draft.Task.CoreIdea = "" }},
	{checks.CodeWrongLanguage, func(f *faulty) { f.language = "ru" }},
	{checks.CodeDistractorExplanations, func(f *faulty) {
		f.draft.Task.Distractors["D"] = checks.Distractor{Trap: "wrong_operation", Text: "You missed one pair."}
	}},
	{checks.CodeDrawingFormat, func(f *faulty) { f.outOfFormat = true }},
	{checks.CodeDrawingMismatch, func(f *faulty) { f.strayLabel = true }},
	{checks.CodeReadability, func(f *faulty) { f.draft.Task.Question += longSentence }},
	{checks.CodeSolverError, func(f *faulty) { f.runner = &scripted{result: solver.Result{Status: solver.StatusError}} }},
	{checks.CodeSolverDisagrees, func(f *faulty) { f.draft.SelfCheck.FinalAnswer = "B" }},
	{checks.CodeSelfCheckBlocking, func(f *faulty) {
		f.draft.SelfCheck.Issues = []checks.Issue{{Type: "ambiguous", Severity: "blocking", Comment: "Two readings."}}
	}},
	{checks.CodeNearDuplicate, func(f *faulty) { f.copied = true }},
}

// broken is a task that passes, with the faults the bits of mask choose. It is
// asked for in no language, which holds it neither to letters nor to the
// reading grade of English, so that a lesson in Russian is a fault of its own.
func broken(mask int) scenario {
	f := faulty{scenario: accepted()}
	f.askedFor(rating.Grades12)
	f.language = ""
	for i, fault := range faults {
		if mask&(1<<i) != 0 {
			fault.apply(&f)
		}
	}
	if f.outOfFormat || f.strayLabel {
		copies, labels := 1, []string{"P", "Q"}
		if f.outOfFormat {
			copies = 3
		}
		if f.strayLabel {
			labels = append(labels, "R")
		}
		picture(&f.scenario, pictureOf(copies, labels...), "P", "Q")
	}
	if f.copied {
		f.references = []string{f.draft.Task.Question}
	}
	return f.scenario
}

func TestTheReviewHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("any set of faults is reported whole, in the order of the checks, the first counted",
		prop.ForAll(
			func(mask int) bool {
				task := broken(mask)
				outcome := task.review(t)
				var want []checks.Code
				for i, fault := range faults {
					if mask&(1<<i) != 0 {
						want = append(want, fault.code)
					}
				}
				first := checks.Code("")
				if len(want) > 0 {
					first = want[0]
				}
				return slices.Equal(codesOf(&outcome), want) && outcome.Primary() == first &&
					outcome.Accepted() == (len(want) == 0)
			},
			gen.IntRange(0, 1<<len(faults)-1),
		))

	properties.TestingRun(t)
}

// A submission comes from the chat's model and is untrusted in every part:
// whatever its JSON and its program hold, and whatever level and language it
// is judged against — one of the three or none of them — the review ends in an outcome that keeps its own rules —
// codes it knows, in the order the checks run, the first one counted, a task
// accepted only when it failed nothing, and a check left unrun only for a task
// out of the format.
func FuzzReview(f *testing.F) {
	good := validDraft()
	f.Add([]byte(jsonOf(f, good.Task)), []byte(jsonOf(f, good.SelfCheck)), program, "3-4", "en")
	f.Add([]byte(jsonOf(f, good.Task)), []byte(jsonOf(f, good.SelfCheck)), program, "1-2", "ru")
	f.Add([]byte(`{"options":{"A":"x"},"question":" "}`), []byte("{"), "", "", "")
	f.Add([]byte(`{"question":"P is left of R.","drawing":"P───Q \n","picture":`+
		`{"kind":"row","items":[{"label":"R"},{"label":"7"}],"copies":3}}`),
		[]byte(`{"issues":[{"type":"ambiguous","severity":"blocking","comment":"C"}],"final_answer":"B"}`),
		"x", "9-10", "en-GB")

	f.Fuzz(func(t *testing.T, task, selfCheck []byte, source, level, language string) {
		reviewer := checks.NewReviewer(shipped{catalog: testCatalog, references: []string{good.Task.Question}},
			&working{value: "six pairs"})
		examined, err := reviewer.Examine(t.Context(),
			&checks.Submission{Task: task, SelfCheck: selfCheck, Solver: source})
		if err != nil {
			t.Fatalf("Examine() error = %v with a sandbox that never fails", err)
		}
		request := asked()
		request.GradeLevel = rating.GradeLevel(level)
		outcome, err := reviewer.Judge(examined, checks.Against{Asked: request, Language: language})
		if err != nil {
			t.Fatalf("Judge() error = %v for what Examine returned, against an open request", err)
		}
		keepsItsRules(t, &outcome)
	})
}

// keepsItsRules fails the test for an outcome that breaks a rule of its own:
// a code it does not know or one out of the order of the checks, a check left
// unrun with no fault of the format to account for it, an acceptance of
// anything that failed, a first code that is not the one counted.
func keepsItsRules(t *testing.T, outcome *checks.Outcome) {
	t.Helper()

	order := checks.Codes()
	rank := -1
	for _, problem := range outcome.Problems {
		at := slices.Index(order, problem.Code)
		if at < 0 || at < rank || problem.Message == "" {
			t.Fatalf("problems = %v, want messages under known codes in the order of the checks", outcome.Problems)
		}
		rank = at
	}
	if len(outcome.Unchecked) > 0 && len(outcome.Problems) == 0 {
		t.Fatalf("unchecked %v with nothing refused, want a check left unrun only for a task out of the format",
			outcome.Unchecked)
	}
	if accepted := len(outcome.Problems) == 0; outcome.Accepted() != accepted {
		t.Fatalf("Accepted() = %v, want %v", outcome.Accepted(), accepted)
	}
	if len(outcome.Problems) > 0 && outcome.Primary() != outcome.Problems[0].Code {
		t.Fatalf("Primary() = %q, want the first code, %q", outcome.Primary(), outcome.Problems[0].Code)
	}
	if event := outcome.Event(); !slices.Equal(event.Failed, codesOf(outcome)) {
		t.Fatalf("Event().Failed = %v, want %v", event.Failed, codesOf(outcome))
	}
}
