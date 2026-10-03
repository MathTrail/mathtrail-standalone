package reviewing_test

import (
	"context"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
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
