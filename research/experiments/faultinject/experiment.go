package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// sanityRow is what the review of one host, as it is, came to.
type sanityRow struct {
	Host    *host
	Verdict verdict
}

// eligible says whether the checks accept the host as it is, with every check
// run: only such a host can carry a defect whose refusal is the defect's.
func (r *sanityRow) eligible() bool { return !r.Verdict.Refused() && len(r.Verdict.Unchecked) == 0 }

// caseRow is one case: an operator on a host, and what the review said.
type caseRow struct {
	Operator *operator
	Host     *host
	Mutant   mutant
	Valid    bool
	Verdict  verdict
	// Similarity is how alike the case's question and its source are, and
	// Sketch the same as the child's history estimates it; both are negative
	// where there is no source.
	Similarity float64
	Sketch     float64
}

// expected says whether a check the operator's class expects refused the case.
func (c *caseRow) expected() bool {
	return slices.ContainsFunc(c.Operator.Expected, c.Verdict.Has)
}

// results are everything the experiment found.
type results struct {
	Sanity []sanityRow
	Cases  []caseRow
}

// runExperiment reviews every host as it is, then every case of every
// operator on every eligible host.
func runExperiment(ctx context.Context, shipped *content.Content, runner solver.Runner, ops []operator) (results, error) {
	var found results
	var eligible []*host
	for _, h := range reviewing.Hosts(shipped) {
		v, err := reviewing.Review(ctx, runner, reviewing.Without(shipped, h.Base.Task.Question), &h.Base)
		if err != nil {
			return results{}, fmt.Errorf("faultinject: review host %s: %w", h.ID, err)
		}
		row := sanityRow{Host: h, Verdict: v}
		found.Sanity = append(found.Sanity, row)
		if row.eligible() {
			eligible = append(eligible, h)
		}
	}
	p := &pool{content: shipped, hosts: eligible}
	for i := range ops {
		for _, h := range eligible {
			row, applies, err := runCase(ctx, shipped, runner, p, &ops[i], h)
			if err != nil {
				return results{}, fmt.Errorf("faultinject: run %s on %s: %w", ops[i].ID, h.ID, err)
			}
			if applies {
				found.Cases = append(found.Cases, row)
			}
		}
	}
	return found, nil
}

// runCase applies one operator to one host and reviews the case, and says
// whether the operator applied.
func runCase(ctx context.Context, shipped *content.Content, runner solver.Runner, p *pool, op *operator, h *host) (caseRow, bool, error) {
	made, applies := op.Apply(newMaker(h, p, op.ID))
	if !applies {
		return caseRow{}, false, nil
	}
	row := caseRow{Operator: op, Host: h, Mutant: made, Valid: !made.Sub.SameAs(&h.Base), Similarity: -1, Sketch: -1}
	if row.Valid && op.Valid != nil {
		confirmed, err := op.Valid(ctx, runner, &made)
		if err != nil {
			return caseRow{}, false, err
		}
		row.Valid = confirmed
	}
	leftOut := append([]string{h.Base.Task.Question}, made.LeaveOut...)
	v, err := reviewing.Review(ctx, runner, reviewing.Without(shipped, leftOut...), &made.Sub)
	if err != nil {
		return caseRow{}, false, err
	}
	row.Verdict = v
	if made.Source != "" {
		row.Similarity = checks.Similarity(made.Sub.Task.Question, made.Source)
	}
	if len(made.Sub.Fingerprints) > 0 {
		if estimate, readable := sketchSimilarity(checks.Fingerprint(made.Sub.Task.Question, language), made.Sub.Fingerprints[0]); readable {
			row.Sketch = estimate
		}
	}
	return row, true, nil
}
