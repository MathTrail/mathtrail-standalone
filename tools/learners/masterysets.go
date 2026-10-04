package main

import (
	"errors"
	"math"
)

// masteryBaseline is the rule every rule of mastery is compared with: the step
// chosen, under the service's own rule of mastery. It is also the exit of the
// choice of mastery, kept as it is when no candidate meets the criterion.
const masteryBaseline = chosenStep

// masteryCandidate is a rule of mastery put forward in its choice: its name,
// its test, and what the choice reads of it — what it is a candidate for, how
// many numbers it adds to the service's rule, whether it adds fields to the
// profile, and how near the service's rule it stands, for choosing the
// nearest of rules alike in all else.
type masteryCandidate struct {
	name     string
	test     func() masteryTest
	kind     candidateKind
	added    int
	fields   bool
	distance float64
}

// masteryCandidates are the rules of mastery put forward, each by one variant
// and none tuned: the service's run made five long, which adds no number; the
// cautious estimate at three margins, which adds the margin; and Wald's test,
// which adds its two error rates and keeps its sums in the profile, and so is
// a backup, chosen over the main one only by the margin. The run of five
// stands nearest the service's rule, then the cautious estimate, the smaller
// margin the nearer, and Wald's test, which weighs answers in a way of its
// own, farthest.
func masteryCandidates() []masteryCandidate {
	all := []masteryCandidate{{name: "run5", test: func() masteryTest { return newRunOf(5) }, kind: mainCandidate}}
	for _, z := range cautiousMargins {
		all = append(all, masteryCandidate{
			name: "cautious_z" + decimal(z), test: func() masteryTest { return cautious{z: z} }, kind: mainCandidate, added: 1, distance: z,
		})
	}
	return append(all, masteryCandidate{
		name: "wald", test: func() masteryTest { return newWald() }, kind: backupCandidate, added: 2, fields: true, distance: math.Inf(1),
	})
}

// cautiousMargins are the margins of the cautious estimate put forward.
var cautiousMargins = []float64{1, 1.28, 1.64}

// modelOf is the step chosen with a rule of mastery of its own in the
// service's place: the model the choice of mastery weighs.
func modelOf(step *rule, mc masteryCandidate) *rule {
	return &rule{
		name: step.name + "+" + mc.name, shape: step.shape, trial: step.trial, make: step.make, mastery: mc.test,
		candidacy: &candidacy{choice: forMastery, kind: mc.kind, added: mc.added, fields: mc.fields, distance: mc.distance},
	}
}

// masteryModels are the step chosen under every rule of mastery put forward,
// and, to show what reading at the middle task costs without a margin, under
// the cautious estimate at no margin, which is no candidate.
func masteryModels() ([]*rule, error) {
	step, err := chosenRule()
	if err != nil {
		return nil, err
	}
	var all []*rule
	for _, mc := range masteryCandidates() {
		all = append(all, modelOf(step, mc))
	}
	noMargin := modelOf(step, masteryCandidate{name: "cautious_z0", test: func() masteryTest { return cautious{z: 0} }})
	noMargin.candidacy = nil
	return append(all, noMargin), nil
}

// masteryRules are a run of the choice of mastery: the baseline and every
// model, between the calibration's rules.
func masteryRules() ([]*rule, error) {
	step, err := chosenRule()
	if err != nil {
		return nil, err
	}
	models, err := masteryModels()
	if err != nil {
		return nil, err
	}
	return withCalibration(append([]*rule{step}, models...)), nil
}

// chosenMastery names the model the decision run of mastery chose, as its
// criterion reads the choice; empty until that run, or when the choice keeps
// the service's rule of mastery.
const chosenMastery = ""

var errNoMasteryChosen = errors.New("learners: no rule of mastery is chosen yet")

// chosenModel is the model the decision run of mastery chose, among its own.
func chosenModel() (*rule, error) {
	models, err := masteryModels()
	if err != nil || chosenMastery == "" {
		return nil, errNoMasteryChosen
	}
	for _, r := range models {
		if r.name == chosenMastery {
			return r, nil
		}
	}
	return nil, errNoMasteryChosen
}
