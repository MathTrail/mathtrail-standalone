package reviewing_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

func loaded(t *testing.T) *content.Content {
	t.Helper()
	shipped, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	return shipped
}

// A reference task, made into a submission with the fixed texts, is accepted
// by the service's checks as it is: at each level, some are.
func TestTheCompletedTasksOfEveryLevelAreAccepted(t *testing.T) {
	t.Parallel()
	shipped := loaded(t)
	runner, err := reviewing.NewRunner()
	if err != nil {
		t.Fatal(err)
	}
	accepted := map[rating.GradeLevel]bool{}
	for _, h := range reviewing.Hosts(shipped) {
		if accepted[h.Level] {
			continue
		}
		v, err := reviewing.Review(context.Background(), runner, reviewing.Without(shipped, h.Base.Task.Question), &h.Base)
		if err != nil {
			t.Fatalf("review %s: %v", h.ID, err)
		}
		accepted[h.Level] = !v.Refused() && len(v.Unchecked) == 0
	}
	for _, level := range rating.GradeLevels() {
		if !accepted[level] {
			t.Errorf("no completed reference task of grades %s is accepted", level)
		}
	}
}

// The comparison leaves out the questions a case was built from, and nothing
// else, and leaves the shipped content as it was.
func TestTheComparisonLeavesOutOnlyTheSources(t *testing.T) {
	t.Parallel()
	shipped := loaded(t)
	all := shipped.ReferenceQuestions(rating.Grades34)
	if len(all) < 3 {
		t.Fatalf("grades 3–4 have %d reference questions, want at least 3", len(all))
	}
	got := reviewing.Without(shipped, all[0], all[2]).ReferenceQuestions(rating.Grades34)
	want := slices.Concat(all[1:2], all[3:])
	if !slices.Equal(got, want) {
		t.Errorf("without the first and the third: %d questions, want the other %d", len(got), len(want))
	}
	if nothing := reviewing.Without(shipped).ReferenceQuestions(rating.Grades34); !slices.Equal(nothing, all) {
		t.Errorf("leaving nothing out: %d questions, want all %d", len(nothing), len(all))
	}
	if again := shipped.ReferenceQuestions(rating.Grades34); !slices.Equal(again, all) {
		t.Error("leaving questions out changed the shipped content")
	}
}

// Solvers run under the service's own limits as they were at the pin. When the
// service's defaults move, this fails, so that the change is reported rather
// than silently left out of the experiments or taken into them.
func TestTheSandboxLimitsAreTheServiceDefaults(t *testing.T) {
	t.Parallel()
	service := starlark.Limits{
		Steps:       config.DefaultSolverSteps,
		Timeout:     config.DefaultSolverTimeout,
		Concurrency: config.DefaultSolverConcurrency,
		Wait:        config.DefaultSolverWait,
	}
	if reviewing.SandboxLimits != service {
		t.Errorf("the experiments' limits are %+v, and the service's defaults are now %+v", reviewing.SandboxLimits, service)
	}
}

// A clone is a copy: a change to it leaves the original as it was.
func TestACloneLeavesTheOriginalAlone(t *testing.T) {
	t.Parallel()
	h := reviewing.Hosts(loaded(t))[0]
	clone := h.Base.Clone()
	clone.Task.Options["A"] = "changed"
	clone.Brief.TrapsToUse[0] = "changed"
	clone.SelfCheck.OptionCheck["A"] = "changed"
	if h.Base.SameAs(&clone) {
		t.Error("a changed clone hands in what the original does")
	}
	for part, value := range map[string]string{
		"an option":                h.Base.Task.Options["A"],
		"a trap of the brief":      h.Base.Brief.TrapsToUse[0],
		"a line of the self-check": h.Base.SelfCheck.OptionCheck["A"],
	} {
		if value == "changed" {
			t.Errorf("changing %s of a clone changed the original's", part)
		}
	}
}

// A host with a drawing says so and one without says not, and a clone of a
// task is a copy down to its drawing's structure: a defect put into the
// clone's drawing leaves the original's as it was.
func TestACloneOfATaskCopiesItsDrawing(t *testing.T) {
	t.Parallel()
	hosts := reviewing.Hosts(loaded(t))
	drawn := slices.IndexFunc(hosts, func(h *reviewing.Host) bool {
		structure := h.Base.Task.DrawingStructure
		return structure != nil && len(structure.Objects) > 0 && len(structure.Relations) > 0
	})
	plain := slices.IndexFunc(hosts, func(h *reviewing.Host) bool { return h.Base.Task.Drawing == "" })
	if drawn < 0 || plain < 0 {
		t.Fatalf("no reference task with a drawing of objects and relations (%d), or none without a drawing (%d)", drawn, plain)
	}
	if !hosts[drawn].HasDrawing() || hosts[plain].HasDrawing() {
		t.Errorf("HasDrawing() = %v with a drawing and %v without, want true and false", hosts[drawn].HasDrawing(), hosts[plain].HasDrawing())
	}
	original := hosts[drawn].Base.Task.DrawingStructure
	kind, label, relation := original.Kind, original.Objects[0].Label, original.Relations[0].Type
	clone := reviewing.CloneTask(&hosts[drawn].Base.Task)
	clone.DrawingStructure.Kind = "changed"
	clone.DrawingStructure.Objects[0].Label = "changed"
	clone.DrawingStructure.Relations[0].Type = "changed"
	if original.Kind != kind || original.Objects[0].Label != label || original.Relations[0].Type != relation {
		t.Errorf("changing a clone's drawing made the original's kind %q, label %q and relation %q, want %q, %q and %q",
			original.Kind, original.Objects[0].Label, original.Relations[0].Type, kind, label, relation)
	}
}

// A submission's options reach a solver in the order of their letters, so that
// the place a solver names is the letter the key names.
func TestOptionsComeInTheOrderOfTheirLetters(t *testing.T) {
	t.Parallel()
	for _, h := range reviewing.Hosts(loaded(t)) {
		options := h.Base.Options()
		for place, letter := range solver.Letters() {
			if options[place] != h.Base.Task.Options[letter] {
				t.Fatalf("%s: option %d is %q, want the text of %s, %q", h.ID, place, options[place], letter, h.Base.Task.Options[letter])
			}
		}
	}
}

// A verdict names each check that refused a submission once, in the order the
// reviewer first reported it, keeps the reviewer's first refusal as its
// primary one, and passes on what could not be checked and the solver's runs.
func TestAVerdictNamesEachRefusalOnceInTheOrderItCame(t *testing.T) {
	t.Parallel()
	outcome := &checks.Outcome{
		Problems: []checks.Problem{
			{Code: checks.CodeReadability, Message: "too long"},
			{Code: checks.CodeNearDuplicate, Message: "a copy"},
			{Code: checks.CodeReadability, Message: "too hard"},
		},
		Unchecked: []string{"the drawing was not checked"},
	}
	runs := []checks.Run{{Status: solver.StatusOK, Steps: 120}, {Status: solver.StatusOK, Steps: 118}}
	v := reviewing.VerdictOf(outcome, runs)
	if want := []checks.Code{checks.CodeReadability, checks.CodeNearDuplicate}; !slices.Equal(v.Codes, want) {
		t.Errorf("Codes = %v, want %v", v.Codes, want)
	}
	if v.Primary != outcome.Primary() || v.Primary != checks.CodeReadability {
		t.Errorf("Primary = %q, want the outcome's %q", v.Primary, outcome.Primary())
	}
	if !slices.Equal(v.Unchecked, outcome.Unchecked) || !slices.Equal(v.Runs, runs) {
		t.Errorf("Unchecked = %q and Runs = %v, want %q and %v", v.Unchecked, v.Runs, outcome.Unchecked, runs)
	}
	if !v.Refused() || !v.Has(checks.CodeNearDuplicate) || v.Has(checks.CodeSolverError) {
		t.Errorf("Refused() = %v, Has(near duplicate) = %v, Has(solver error) = %v; want true, true and false",
			v.Refused(), v.Has(checks.CodeNearDuplicate), v.Has(checks.CodeSolverError))
	}
	if accepted := reviewing.VerdictOf(&checks.Outcome{}, nil); accepted.Refused() || accepted.Primary != "" {
		t.Errorf("a verdict of no problems: Refused() = %v, Primary = %q; want false and none", accepted.Refused(), accepted.Primary)
	}
}

// A review whose caller has gone fails with the caller's reason rather than
// coming to a verdict, so that a stopped experiment counts no task as refused
// or accepted.
func TestAReviewWhoseCallerHasGoneFails(t *testing.T) {
	t.Parallel()
	shipped := loaded(t)
	runner, err := reviewing.NewRunner()
	if err != nil {
		t.Fatal(err)
	}
	h := reviewing.Hosts(shipped)[0]
	against := reviewing.Without(shipped, h.Base.Task.Question)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, failed := reviewing.Review(ctx, runner, against, &h.Base); !errors.Is(failed, context.Canceled) {
		t.Errorf("Review() error = %v, want %v", failed, context.Canceled)
	}
	if _, _, failed := reviewing.Judged(ctx, runner, against, &h.Base); !errors.Is(failed, context.Canceled) {
		t.Errorf("Judged() error = %v, want %v", failed, context.Canceled)
	}
}

// The half of a review that judges refuses what the other half never
// examined, rather than judging an empty submission as one with nothing wrong.
func TestJudgingWhatWasNeverExaminedFails(t *testing.T) {
	t.Parallel()
	shipped := loaded(t)
	runner, err := reviewing.NewRunner()
	if err != nil {
		t.Fatal(err)
	}
	h := reviewing.Hosts(shipped)[0]
	prepared, err := reviewing.Prepare(runner, reviewing.Without(shipped, h.Base.Task.Question), &h.Base)
	if err != nil {
		t.Fatal(err)
	}
	_, err = prepared.Judge(checks.Examined{})
	if want := "judge: checks: judge: the submission was not examined"; err == nil || err.Error() != want {
		t.Errorf("Judge(nothing examined) error = %v, want %q", err, want)
	}
}
