package main

import (
	"errors"
	"strings"
)

// debtStep is a debt over the step chosen, by its numbers: the share of a wrong
// answer's charge taken at once at its first place in a run of wrong answers
// in its topic and at its second — the third and every later place take the
// whole —, the share of a right answer's gain that pays what is owed, and how
// much one unit of payment clears.
type debtStep struct {
	first, second, repayment, burn float64
}

// The debts put forward, written down before any of them ran. The author's: a
// first wrong answer in a run takes a tenth of its charge, the second two
// fifths, and a right answer pays what is owed with four fifths of its gain,
// so that a fifth of it always raises the level. The backup takes a fifth of
// the first wrong answer's charge, as the simulations of 2026-10-11 first ran
// it, and is chosen only when the author's is not.
var (
	authorsDebt = debtStep{first: 0.1, second: 0.4, repayment: 0.8, burn: 1}
	backupDebt  = debtStep{first: 0.2, second: 0.4, repayment: 0.8, burn: 1}
)

// The comparisons, which no choice takes: the author's debt paid with the
// whole of a right answer's gain, which shows the child nothing for a right
// answer while anything is owed; and paid at twice the pace — the pace the debt
// was first put forward with, then over a first share of a fifth and no second
// one — which clears a debt with half of it charged and forgives the rest.
var (
	wholeGainDebt = debtStep{first: 0.1, second: 0.4, repayment: 1, burn: 1}
	twiceDebt     = debtStep{first: 0.1, second: 0.4, repayment: 0.8, burn: 2}
)

func (d debtStep) shares() []float64 { return []float64{d.first, d.second, 1} }

// name is how the debt's numbers name it: debt_0.1_0.4_p0.8, with _x2 for a
// payment that clears twice what it charges.
func (d debtStep) name() string {
	name := "debt_" + decimal(d.first) + "_" + decimal(d.second) + "_p" + decimal(d.repayment)
	if d.burn != 1 {
		name += "_x" + decimal(d.burn)
	}
	return name
}

var errNotAStep = errors.New("learners: a debt goes over a step rule, and the step chosen keeps its level another way")

// chosenStepAndModel are the step chosen and the model chosen over it, which
// every debt goes over; either missing is said as it is.
func chosenStepAndModel() (step, model *rule, err error) {
	if step, err = chosenRule(); err != nil {
		return nil, nil, err
	}
	if model, err = chosenModel(); err != nil {
		return nil, nil, err
	}
	return step, model, nil
}

// debtOver is a model with a debt over its step: the step, keeping the debt,
// under the model's rule of mastery, named with the debt between the two. A
// step whose estimate is no step rule has nowhere to keep a debt, and is
// refused rather than run as if it kept one; the estimate it is told by is
// made for a child of no generator, which a step rule never reads.
func debtOver(step, model *rule, d debtStep) (*rule, error) {
	if _, isStep := step.make(&child{}, 0).(*stepRule); !isStep {
		return nil, errNotAStep
	}
	stepMake := step.make
	return &rule{
		name: step.name + "+" + d.name() + strings.TrimPrefix(model.name, step.name), shape: step.shape, trial: step.trial,
		make: func(c *child, start float64) estimator {
			s, _ := stepMake(c, start).(*stepRule)
			s.debt = newOwing(d.shares(), d.repayment, d.burn)
			return s
		},
		mastery: model.mastery,
	}, nil
}

// debtCandidate is a debt over the model chosen, put forward in the choice of
// a debt: it adds its three numbers — the two shares and the payment — and the
// two debts to the profile, and stands from the baseline, which charges every
// wrong answer whole, by how much of a first wrong answer's charge it defers.
func debtCandidate(step, model *rule, d debtStep, kind candidateKind) (*rule, error) {
	r, err := debtOver(step, model, d)
	if err != nil {
		return nil, err
	}
	r.candidacy = &candidacy{choice: forDebt, kind: kind, added: 3, fields: true, distance: 1 - d.first}
	return r, nil
}

// debtRules are a run of the choice of a debt: the baseline, the author's debt
// and the backup, and the comparisons — the debt paid with the whole gain, paid
// at twice the pace, and the author's with mastery read off the level less
// what it owes — between the calibration's rules.
func debtRules() ([]*rule, error) {
	step, model, err := chosenStepAndModel()
	if err != nil {
		return nil, err
	}
	authors, err := debtCandidate(step, model, authorsDebt, mainCandidate)
	if err != nil {
		return nil, err
	}
	backup, err := debtCandidate(step, model, backupDebt, backupCandidate)
	if err != nil {
		return nil, err
	}
	own := []*rule{model, authors, backup}
	for _, d := range []debtStep{wholeGainDebt, twiceDebt} {
		r, overErr := debtOver(step, model, d)
		if overErr != nil {
			return nil, overErr
		}
		own = append(own, r)
	}
	settled, err := debtOver(step, model, authorsDebt)
	if err != nil {
		return nil, err
	}
	settled.name += "_settled"
	settled.masteryOnSettled = true
	return withCalibration(append(own, settled)), nil
}

// chosenDebt names the debt the decision run of a debt chose, as its criterion
// reads the choice; empty until that run, or when the choice keeps the model
// as it is.
const chosenDebt = ""

var errNoDebtChosen = errors.New("learners: no debt is chosen yet, so there is nothing to confirm")

// debtConfirmationRules are the debt chosen alone beside the baseline, the
// service, the slow constant step and the ceiling, on children held out for
// it: a set held out is confirmed on once, by one candidate.
func debtConfirmationRules() ([]*rule, error) {
	if chosenDebt == "" {
		return nil, errNoDebtChosen
	}
	all, err := debtRules()
	if err != nil {
		return nil, err
	}
	model, debt := ruleIn(all, debtBaseline), ruleIn(all, chosenDebt)
	if model == nil || debt == nil {
		return nil, errNoDebtChosen
	}
	return withCalibration([]*rule{model, debt}), nil
}

// ruleIn is the rule of that name among these, or nil.
func ruleIn(all []*rule, name string) *rule {
	for _, r := range all {
		if r.name == name {
			return r
		}
	}
	return nil
}
