package reviewing

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// SandboxLimits are the service's own limits on a solver's run at the pin,
// written out here rather than read from its configuration, so that a later
// change of the service's defaults is not silently taken up: a test holds the
// two together and fails when they part.
var SandboxLimits = starlark.Limits{
	Steps:       25_000_000,
	Timeout:     2 * time.Second,
	Concurrency: 1,
	Wait:        3 * time.Second,
}

// NewRunner is the sandbox every solver of an experiment runs in.
func NewRunner() (solver.Runner, error) {
	runner, err := starlark.New(SandboxLimits)
	if err != nil {
		return nil, fmt.Errorf("reviewing: make the sandbox: %w", err)
	}
	return runner, nil
}

// References is the shipped content with some reference questions left out of
// the near-duplicate comparison: those a case was built from, which would
// otherwise refuse it as a copy of itself.
type References struct {
	*content.Content
	leftOut map[string]bool
}

// Without is the content with these questions left out of the comparison.
func Without(shipped *content.Content, questions ...string) References {
	leftOut := make(map[string]bool, len(questions))
	for _, question := range questions {
		leftOut[question] = true
	}
	return References{Content: shipped, leftOut: leftOut}
}

// ReferenceQuestions are the reference questions of a level, less those left
// out.
func (r References) ReferenceQuestions(level rating.GradeLevel) []string {
	return slices.DeleteFunc(slices.Clone(r.Content.ReferenceQuestions(level)), func(question string) bool {
		return r.leftOut[question]
	})
}

// Verdict is what a review said of one submission: every check that refused
// it, the first of them, the checks that could not run, and the solver's runs.
type Verdict struct {
	Codes     []checks.Code
	Primary   checks.Code
	Unchecked []string
	Runs      []checks.Run
}

// Refused says whether any check refused the submission.
func (v *Verdict) Refused() bool { return len(v.Codes) > 0 }

// Has says whether this check refused the submission.
func (v *Verdict) Has(code checks.Code) bool { return slices.Contains(v.Codes, code) }

// Review runs a submission through the service's reviewer, both halves, as the
// service does when a task is handed in.
func Review(ctx context.Context, runner solver.Runner, against checks.Content, sub *Submission) (Verdict, error) {
	outcome, runs, err := Judged(ctx, runner, against, sub)
	if err != nil {
		return Verdict{}, err
	}
	return VerdictOf(&outcome, runs), nil
}

// VerdictOf is what a review's outcome and its solver's runs say of a
// submission.
func VerdictOf(outcome *checks.Outcome, runs []checks.Run) Verdict {
	var codes []checks.Code
	for _, problem := range outcome.Problems {
		if !slices.Contains(codes, problem.Code) {
			codes = append(codes, problem.Code)
		}
	}
	return Verdict{Codes: codes, Primary: outcome.Primary(), Unchecked: outcome.Unchecked, Runs: runs}
}

// Judged is the reviewer's whole outcome for a submission, with the solver's
// runs.
func Judged(ctx context.Context, runner solver.Runner, against checks.Content, sub *Submission) (checks.Outcome, []checks.Run, error) {
	prepared, err := Prepare(runner, against, sub)
	if err != nil {
		return checks.Outcome{}, nil, err
	}
	examined, err := prepared.Examine(ctx)
	if err != nil {
		return checks.Outcome{}, nil, err
	}
	outcome, err := prepared.Judge(examined)
	if err != nil {
		return checks.Outcome{}, nil, err
	}
	return outcome, examined.Runs(), nil
}

// Prepared is a submission ready for the service's reviewer: as the service
// receives it, with what it is judged against and a reviewer over the content
// it is compared with. Its two halves run apart, so that each can be timed.
type Prepared struct {
	handed   checks.Submission
	against  checks.Against
	reviewer checks.Reviewer
}

// Prepare makes a submission ready to be reviewed against this content, in
// this sandbox.
func Prepare(runner solver.Runner, against checks.Content, sub *Submission) (*Prepared, error) {
	handed, err := sub.HandedIn()
	if err != nil {
		return nil, err
	}
	asked := sub.Asked
	return &Prepared{
		handed:   handed,
		against:  checks.Against{Asked: &asked, Language: Language, Fingerprints: sub.Fingerprints},
		reviewer: checks.NewReviewer(against, runner, checks.DefaultDrawingLimits()),
	}, nil
}

// Examine is the half of a review that reads the submission and runs its
// solver.
func (p *Prepared) Examine(ctx context.Context) (checks.Examined, error) {
	examined, err := p.reviewer.Examine(ctx, &p.handed)
	if err != nil {
		return checks.Examined{}, fmt.Errorf("examine: %w", err)
	}
	return examined, nil
}

// Judge is the half that runs every check on what Examine found, against the
// request the submission answers and the child's earlier tasks.
func (p *Prepared) Judge(examined checks.Examined) (checks.Outcome, error) {
	outcome, err := p.reviewer.Judge(examined, p.against)
	if err != nil {
		return checks.Outcome{}, fmt.Errorf("judge: %w", err)
	}
	return outcome, nil
}
