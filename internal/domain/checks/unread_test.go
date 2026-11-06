package checks_test

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// firstOnly is a solver that reads the picture on its first run and nothing of
// it on the second, as one that looks at the options before the picture
// might: each run is held to the picture on its own.
type firstOnly struct {
	working
	first []string
}

func (f *firstOnly) Run(ctx context.Context, source string, options solver.Options, picture json.RawMessage) (solver.Result, error) {
	result, err := f.working.Run(ctx, source, options, picture)
	if f.runs == 1 {
		result.Read = f.first
	}
	return result, err
}

// The solver of a task solved from its picture is held to every member of the
// picture that carries facts, on both runs, and to nothing else; the solver of
// a word task is held to none of it.
func TestTheSolverOfATaskSolvedFromThePictureReadsIt(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		change  func(*scenario)
		refused string
	}{
		{"a solver that reads every fact", func(s *scenario) {
			solvedFromPicture(s)
			s.runner = &working{value: "six pairs", read: []string{"items"}}
		}, ""},
		{"a solver that leaves only what lays the picture out", func(s *scenario) {
			solvedFromPicture(s)
			s.runner = &working{value: "six pairs", read: []string{"items"}}
			s.draft.Task.Picture = json.RawMessage(`{"kind":"row","items":[{"label":"P"},{"label":"Q"}],"copies":1,"line":true}`)
		}, ""},
		{"a solver that leaves only the mark of what is asked", func(s *scenario) {
			solvedFromPicture(s)
			s.runner = &working{value: "six pairs", read: []string{"items"}}
			s.draft.Task.Picture = json.RawMessage(`{"kind":"row","items":[{"label":"P"},{"label":"Q"}],"gaps":"?"}`)
		}, ""},
		{"a solver that reads one fact of two", func(s *scenario) {
			solvedFromPicture(s)
			s.runner = &working{value: "six pairs", read: []string{"items"}}
			s.draft.Task.Picture = json.RawMessage(`{"kind":"row","items":[{"label":"P"},{"label":"Q"}],"gaps":"3"}`)
		}, "the solver left task.picture.gaps unread"},
		{"a solver that reads the picture on one run of two", func(s *scenario) {
			solvedFromPicture(s)
			s.runner = &firstOnly{working: working{value: "six pairs"}, first: []string{"items"}}
		}, "the solver left task.picture.items unread"},
		{"a member the model made up, named without its name", func(s *scenario) {
			solvedFromPicture(s)
			s.draft.Task.Picture = json.RawMessage(`{"kind":"row","items":[{"label":"P"},{"label":"Q"}],"Six pairs":6}`)
			s.runner = &working{value: "six pairs", read: []string{"items"}}
		}, "the solver left task.picture.* unread"},
		{"a word task whose solver reads nothing of its picture", func(s *scenario) {
			solvedFromPicture(s)
			s.fromPicture = false
		}, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			task := accepted()
			test.change(&task)
			outcome := task.review(t)
			refusal := slices.IndexFunc(outcome.Problems, func(problem checks.Problem) bool {
				return problem.Code == checks.CodeSolverUnread
			})
			switch {
			case test.refused == "" && refusal >= 0:
				t.Errorf("problems = %v, want no %s", outcome.Problems, checks.CodeSolverUnread)
			case test.refused != "" && (refusal < 0 || !strings.HasPrefix(outcome.Problems[refusal].Message, test.refused)):
				t.Errorf("problems = %v, want %s saying %q", outcome.Problems, checks.CodeSolverUnread, test.refused)
			case test.refused != "" && strings.Contains(outcome.Problems[refusal].Message, "Six"):
				t.Errorf("message %q quotes what the model wrote", outcome.Problems[refusal].Message)
			}
		})
	}
}

// A solver that crashed is refused for the crash, and what it read of the
// picture is noted as not checked rather than judged: it may have stopped
// before it read what it would have.
func TestASolverThatCrashedIsNotJudgedOnWhatItRead(t *testing.T) {
	t.Parallel()

	task := accepted()
	solvedFromPicture(&task)
	task.runner = &scripted{result: solver.Result{Status: solver.StatusError, Message: "fail"}}
	outcome := task.review(t)

	if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{checks.CodeSolverError}) {
		t.Errorf("codes = %v, want only %s", codes, checks.CodeSolverError)
	}
	if !slices.ContainsFunc(outcome.Unchecked, func(note string) bool { return strings.Contains(note, "read of the picture") }) {
		t.Errorf("unchecked = %v, want what the solver read of the picture noted as not checked", outcome.Unchecked)
	}
}

// A task that could not be read leaves the picture nothing to be held to, and
// the note says so rather than sending the model to its solver.
func TestATaskThatCannotBeReadIsNotJudgedOnItsPicture(t *testing.T) {
	t.Parallel()

	task := accepted()
	solvedFromPicture(&task)
	task.draft.Task = nil
	outcome := task.review(t)

	want := "what the solver read of the picture was not checked: that needs a task that can be read"
	if !slices.Contains(outcome.Unchecked, want) {
		t.Errorf("unchecked = %v, want %q", outcome.Unchecked, want)
	}
}
