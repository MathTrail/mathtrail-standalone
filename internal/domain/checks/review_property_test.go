package checks_test

import (
	"reflect"
	"slices"
	"strings"
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
	{checks.CodeWrongLanguage, func(f *faulty) {
		// A question in English unlike the reference task a copy is made
		// against, so that a lesson in Russian copies nothing of its own.
		f.language, f.draft.Task.QuestionInEnglish = "ru", robots
	}},
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
// reading grade of English and asks it for no question in English, so that a
// lesson in Russian is a fault of its own.
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
				var copied []checks.Earlier
				if slices.Contains(want, checks.CodeNearDuplicate) {
					copied = []checks.Earlier{checks.EarlierReference}
				}
				return slices.Equal(codesOf(&outcome), want) && outcome.Primary() == first &&
					outcome.Accepted() == (len(want) == 0) && slices.Equal(outcome.Event().NearDuplicateOf, copied)
			},
			gen.IntRange(0, 1<<len(faults)-1),
		))

	properties.Property("in a lesson in English, the question in English is not read", prop.ForAll(
		func(inEnglish string) bool {
			plain, sent := accepted(), accepted()
			plain.references, sent.references = []string{clock}, []string{clock}
			sent.draft.Task.QuestionInEnglish = inEnglish
			without, with := plain.review(t), sent.review(t)
			return reflect.DeepEqual(without.Problems, with.Problems) && slices.Equal(without.Unchecked, with.Unchecked) &&
				reflect.DeepEqual(without.Event(), with.Event())
		},
		gen.OneGenOf(gen.AnyString(), gen.Const(clock), gen.Const("")),
	))

	properties.Property("in a lesson in Russian, a question in English that is a reference task is a copy of it",
		prop.ForAll(
			func(words []string) bool {
				task := scenario{runner: &working{value: "шесть пар"}}
				russianLesson(&task)
				task.draft.Task.QuestionInEnglish = strings.Join(words, " ")
				task.references = []string{task.draft.Task.QuestionInEnglish}
				outcome := task.review(t)
				return slices.Equal(outcome.Event().NearDuplicateOf, []checks.Earlier{checks.EarlierReferenceInEnglish}) &&
					slices.Equal(codesOf(&outcome), []checks.Code{checks.CodeNearDuplicate})
			},
			genWords(),
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
	russian := russianDraft()
	f.Add([]byte(jsonOf(f, russian.Task)), []byte(jsonOf(f, russian.SelfCheck)), program, "3-4", "ru")
	f.Add([]byte(`{"question":"Сколько пар?","question_en":" "}`), []byte(`{}`), "", "1-2", "uk")
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
		outcome, err := reviewer.Judge(examined, checks.Against{
			Asked: request, Language: language, Fingerprints: []string{checks.Fingerprint(good.Task.Question, language)},
		})
		if err != nil {
			t.Fatalf("Judge() error = %v for what Examine returned, against an open request", err)
		}
		keepsItsRules(t, &outcome)
		keepsTheRulesOfItsLanguage(t, &outcome, language)
	})
}

// keepsTheRulesOfItsLanguage fails the test for an outcome that breaks a rule
// of the question in English: a log that says whether one came where none was
// asked for, or does not say it of a task that could be read where one was; a
// task of such a lesson that came without one and was not refused for it; and
// a copy of a reference task found by a question in English no lesson asked
// for.
func keepsTheRulesOfItsLanguage(t *testing.T, outcome *checks.Outcome, language string) {
	t.Helper()

	asked := checks.AsksQuestionInEnglish(language)
	given := outcome.Event().QuestionInEnglishGiven
	if (given != nil) != (asked && outcome.Draft.Task != nil) {
		t.Fatalf("QuestionInEnglishGiven = %v in a lesson in %q, want it said exactly of a task read where it is asked for",
			given, language)
	}
	refused := slices.ContainsFunc(outcome.Problems, func(problem checks.Problem) bool {
		return problem.Code == checks.CodeBadStructure && strings.Contains(problem.Message, "task.question_en is missing")
	})
	if given != nil && refused == *given {
		t.Fatalf("a question in English given: %v, and refused for its want: %v; want one refused exactly when none came",
			*given, refused)
	}
	if !asked && slices.Contains(outcome.Event().NearDuplicateOf, checks.EarlierReferenceInEnglish) {
		t.Fatalf("NearDuplicateOf = %v in a lesson in %q, which asks for no question in English",
			outcome.Event().NearDuplicateOf, language)
	}
}

// keepsItsRules fails the test for an outcome that breaks a rule of its own:
// a code it does not know or one out of the order of the checks, a kind of
// earlier task named by anything but a near-duplicate or a near-duplicate that
// names none, a check left unrun with no fault of the format to account for
// it, an acceptance of anything that failed, a first code that is not the one
// counted, and a log that names other checks or other kinds of earlier task
// than the problems do.
func keepsItsRules(t *testing.T, outcome *checks.Outcome) {
	t.Helper()

	order := checks.Codes()
	rank := -1
	var repeated []checks.Earlier
	for _, problem := range outcome.Problems {
		at := slices.Index(order, problem.Code)
		if at < 0 || at < rank || problem.Message == "" {
			t.Fatalf("problems = %v, want messages under known codes in the order of the checks", outcome.Problems)
		}
		if (problem.Code == checks.CodeNearDuplicate) != (problem.NearDuplicateOf != "") {
			t.Fatalf("problem %+v, want a kind of earlier task named by a near-duplicate, and by it alone", problem)
		}
		if problem.NearDuplicateOf != "" {
			repeated = append(repeated, problem.NearDuplicateOf)
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
	if event := outcome.Event(); !slices.Equal(event.NearDuplicateOf, repeated) {
		t.Fatalf("Event().NearDuplicateOf = %v, want the kinds its problems name, %v", event.NearDuplicateOf, repeated)
	}
}
