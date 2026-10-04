package main

import (
	"errors"
	"math"
	"slices"
	"strconv"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// candidateKind is what a rule is a candidate for: the step, worked out from
// an uncertainty counted by answers or from one the profile keeps, or the
// exit, a floor under the service's step, taken when no step is chosen.
type candidateKind string

// The kinds of candidate.
const (
	mainCandidate   candidateKind = "main"
	backupCandidate candidateKind = "backup"
	exitCandidate   candidateKind = "exit"
)

// candidacy is what a choice reads of a candidate beside its numbers: the
// choice it is put forward in, what it is a candidate for there, how many
// numbers it adds to the service's rule, whether it adds fields to the
// profile, and how far it stands from the service's rule.
type candidacy struct {
	choice   choiceKind
	kind     candidateKind
	added    int
	fields   bool
	distance float64
}

// uncertainStep is a step worked out from an uncertainty counted by answers,
// by its numbers: the overall level's uncertainty after the trial series, the
// spread of topics a topic's uncertainty starts at, the uncertainty an answer
// adds to the overall level and to the topic answered, the most one answer may
// move the level in a topic, and whether the gains are the service's own
// rather than the model's, which only the step with its parts taken away
// takes.
type uncertainStep struct {
	afterSeries, topicSpread float64
	addedOverall, addedTopic float64
	limit                    float64
	serviceGains             bool
}

// curves are the step's curves for the overall level, counted from the end of
// the trial series, and for a topic, counted from its first answer.
func (u uncertainStep) curves() (overall, topic stepCurve) {
	overallGain, topicGain := gainPerUncertainty, gainPerUncertainty
	if u.serviceGains {
		overallGain, topicGain = serviceGain(serviceOverallStep), serviceGain(serviceTopicStep)
	}
	return uncertain(u.afterSeries, rating.TrialAnswers, u.addedOverall, overallGain),
		uncertain(u.topicSpread*u.topicSpread, 0, u.addedTopic, topicGain)
}

// serviceGain is the gain a step of the service's has when it is read as one
// from an uncertainty narrowed by answers in the middle of the corridor: its
// first step over the uncertainty its decay stands for.
func serviceGain(c stepCurve) float64 { return c.first * answerInformation / c.decay }

// added is how many numbers the step adds to the service's rule: the
// uncertainty an answer adds to the overall level, the one it adds to a topic,
// and the limit, each where it is used. The uncertainty after the series and
// the spread of topics stand where the service's first steps and its decay
// stand, and add none.
func (u uncertainStep) added() int {
	return countOf(u.addedOverall > 0, u.addedTopic > 0, !math.IsInf(u.limit, 1))
}

// name is how the step's numbers name its rule.
func (u uncertainStep) name() string {
	name := "uncertain_" + u.numbers()
	if u.serviceGains {
		name += "_kservice"
	}
	return name
}

func (u uncertainStep) numbers() string {
	return "v" + decimal(u.afterSeries) + "_s" + decimal(u.topicSpread) + "_" + addedAndLimit(u.addedOverall, u.addedTopic, u.limit)
}

func (u uncertainStep) rule() *rule {
	overall, topic := u.curves()
	return &rule{
		name: u.name(), shape: both, trial: true,
		make:      steps(func(s *stepRule) { s.overallStep, s.topicStep, s.limit = overall, topic, u.limit }),
		candidacy: &candidacy{kind: mainCandidate, added: u.added(), distance: stepDistance(overall, topic, 0)},
	}
}

// withoutParts is the step with each of its parts taken away in turn, as far
// as it has them: the limit, the uncertainty added to the overall level, the
// one added to a topic, all three at once, and the model's gains, given the
// service's instead.
func (u uncertainStep) withoutParts() []uncertainStep {
	unlimited, noOverall, noTopic, bare, serviceGains := u, u, u, u, u
	unlimited.limit = noLimit
	noOverall.addedOverall = 0
	noTopic.addedTopic = 0
	bare.limit, bare.addedOverall, bare.addedTopic = noLimit, 0, 0
	serviceGains.serviceGains = true
	var all []uncertainStep
	for _, part := range []uncertainStep{unlimited, noOverall, noTopic, bare, serviceGains} {
		if part != u && !slices.Contains(all, part) {
			all = append(all, part)
		}
	}
	return all
}

// numbers are values of each number of a step from counted answers: the
// sweep's grid, or the ladders its refinement steps along.
type numbers struct {
	afterSeries, topicSpread, addedOverall, addedTopic, limit []float64
}

// sweepGrid is the grid the sweep runs every step of: values chosen to stand
// on both sides of where each constraint gives way — the uncertainty after
// the series from well under the service's first step after it, 0.17, to the
// uncertainty five answers leave of the start's; the spread of topics around
// the children's; the uncertainty an answer adds from none to as much as
// takes the steady step past the constant step's; and the limit from none to
// a move of the card's no larger than the service's at the start.
var sweepGrid = numbers{
	afterSeries:  []float64{0.1, 0.17, 0.4, 1},
	topicSpread:  []float64{0.35, 0.5, 0.7},
	addedOverall: []float64{0, 0.003, 0.01},
	addedTopic:   []float64{0, 0.006, 0.03},
	limit:        []float64{noLimit, 0.42, 0.3},
}

// refineLadders are finer values of every number, the sweep's among them, that
// the refinement steps along from the sweep's best step, one number at a
// time, to the values on either side.
var refineLadders = numbers{
	afterSeries:  []float64{0.05, 0.1, 0.13, 0.17, 0.25, 0.4, 0.6, 0.8, 1, 1.2},
	topicSpread:  []float64{0.25, 0.35, 0.42, 0.5, 0.6, 0.7, 0.85, 1},
	addedOverall: []float64{0, 0.0015, 0.003, 0.006, 0.01, 0.02, 0.04},
	addedTopic:   []float64{0, 0.003, 0.006, 0.015, 0.03, 0.06},
	limit:        []float64{0.22, 0.3, 0.36, 0.42, 0.5, noLimit},
}

// every is every step of a grid, in a fixed order.
func (g *numbers) every() []uncertainStep {
	var all []uncertainStep
	for _, v := range g.afterSeries {
		for _, s := range g.topicSpread {
			for _, qt := range g.addedOverall {
				for _, qd := range g.addedTopic {
					for _, limit := range g.limit {
						all = append(all, uncertainStep{afterSeries: v, topicSpread: s, addedOverall: qt, addedTopic: qd, limit: limit})
					}
				}
			}
		}
	}
	return all
}

// holds says whether a step's every number is among a grid's values.
func (g *numbers) holds(u uncertainStep) bool {
	return slices.Contains(g.afterSeries, u.afterSeries) && slices.Contains(g.topicSpread, u.topicSpread) &&
		slices.Contains(g.addedOverall, u.addedOverall) && slices.Contains(g.addedTopic, u.addedTopic) &&
		slices.Contains(g.limit, u.limit) && !u.serviceGains
}

// neighbours are the steps one number away from a step on the ladders: in
// each of its numbers in turn, the values next to its own below and above.
func (u uncertainStep) neighbours(l *numbers) []uncertainStep {
	sides := []struct {
		ladder []float64
		own    float64
		set    func(s *uncertainStep, v float64)
	}{
		{l.afterSeries, u.afterSeries, func(s *uncertainStep, v float64) { s.afterSeries = v }},
		{l.topicSpread, u.topicSpread, func(s *uncertainStep, v float64) { s.topicSpread = v }},
		{l.addedOverall, u.addedOverall, func(s *uncertainStep, v float64) { s.addedOverall = v }},
		{l.addedTopic, u.addedTopic, func(s *uncertainStep, v float64) { s.addedTopic = v }},
		{l.limit, u.limit, func(s *uncertainStep, v float64) { s.limit = v }},
	}
	var all []uncertainStep
	for _, side := range sides {
		at := slices.Index(side.ladder, side.own)
		if at < 0 {
			continue
		}
		for _, next := range []int{at - 1, at + 1} {
			if next >= 0 && next < len(side.ladder) {
				near := u
				side.set(&near, side.ladder[next])
				all = append(all, near)
			}
		}
	}
	return all
}

// storedStep is the backup candidate by its numbers: the spread of topics,
// the uncertainty an answer adds to the overall level and to a topic, and the
// limit.
type storedStep struct {
	topicSpread, addedOverall, addedTopic, limit float64
}

// storedLike is the backup with a step's numbers, but for the overall level's
// uncertainty after the series, which the backup reads off each child's own
// series.
func storedLike(u uncertainStep) storedStep {
	return storedStep{topicSpread: u.topicSpread, addedOverall: u.addedOverall, addedTopic: u.addedTopic, limit: u.limit}
}

func (b storedStep) name() string {
	return "stored_s" + decimal(b.topicSpread) + "_" + addedAndLimit(b.addedOverall, b.addedTopic, b.limit)
}

func (b storedStep) rule() *rule {
	return &rule{
		name: b.name(), shape: both, trial: true,
		make: func(_ *child, start float64) estimator {
			return newStoredUncertainty(start, b.topicSpread, b.addedOverall, b.addedTopic, b.limit)
		},
		candidacy: &candidacy{kind: backupCandidate, added: countOf(b.addedOverall > 0, b.addedTopic > 0, !math.IsInf(b.limit, 1)), fields: true},
	}
}

// storedAround are the backup with the step's numbers and with each of the
// uncertainties it adds moved to the values next to its own on the ladders.
func (u uncertainStep) storedAround(l *numbers) []storedStep {
	all := []storedStep{storedLike(u)}
	for _, near := range u.neighbours(&numbers{addedOverall: l.addedOverall, addedTopic: l.addedTopic}) {
		all = append(all, storedLike(near))
	}
	return all
}

// filterRule is the full filter on a step's numbers: what the step's
// simplicity costs.
func filterRule(u uncertainStep) *rule {
	return &rule{
		name: "filter_" + u.numbers(), shape: both, trial: true,
		make: func(_ *child, start float64) estimator {
			return newFullFilter(start, u.afterSeries, u.topicSpread, u.addedOverall, u.addedTopic, u.limit)
		},
	}
}

// historyRule is the estimate over the whole history with a spread of topics.
func historyRule(topicSpread float64) *rule {
	return &rule{
		name: "history_s" + decimal(topicSpread), shape: both, trial: true,
		make: func(_ *child, start float64) estimator { return newWholeHistory(start, topicSpread) },
	}
}

// switchThresholds are the thresholds the sweep runs the two speeds at.
var switchThresholds = []float64{2, 4, 8}

// switchRule is the two speeds with a threshold.
func switchRule(threshold float64) *rule {
	return &rule{
		name: "switch_h" + decimal(threshold), shape: both, trial: true,
		make: func(_ *child, start float64) estimator { return newTwoSpeeds(start, threshold) },
	}
}

// floorRule is the service's step, as the choice of the step found it, with a
// floor under the overall level's: an exit, which adds the floor. The floor of
// serviceFloor is the step the service took.
func floorRule(floor float64) *rule {
	return &rule{
		name: "floor_" + decimal(floor), shape: both, trial: true,
		make:      steps(func(s *stepRule) { s.floor = floor }),
		candidacy: &candidacy{kind: exitCandidate, added: 1, distance: stepDistance(serviceOverallStep, serviceTopicStep, floor)},
	}
}

// exitFloors are the floors the exit is chosen among.
var exitFloors = []float64{0.02, 0.05, 0.1}

// The answers the distance of a step from the service's, as the choice of the
// step found it, is read over: the overall level's from the first after the
// trial series to the 200th, a topic's over its first forty.
const (
	distanceAnswers      = 200
	distanceTopicAnswers = 40
)

// stepDistance is how far a step stands from the service's, as the choice of
// the step found it — before its floor — for choosing the nearest of steps
// alike in all else: the mean difference of the overall level's steps over
// the answers after the trial series, plus that of a topic's over its first
// answers. The limit, which only a large surprise meets, is left out.
func stepDistance(overall, topic stepCurve, floor float64) float64 {
	overallSum := 0.0
	for n := rating.TrialAnswers; n < distanceAnswers; n++ {
		overallSum += math.Abs(max(overall.at(n), floor) - serviceOverallStep.at(n))
	}
	topicSum := 0.0
	for m := range distanceTopicAnswers {
		topicSum += math.Abs(topic.at(m) - serviceTopicStep.at(m))
	}
	return overallSum/float64(distanceAnswers-rating.TrialAnswers) + topicSum/distanceTopicAnswers
}

// The rules every set of a choice holds beside its candidates: the service,
// the slow constant step the bench's resolution is measured with, and the
// ceiling.
func serviceRule() *rule { return &rule{name: "shrinking", shape: both, service: true} }

func slowConstantRule() *rule {
	return &rule{name: measuringRule, shape: both, trial: true, make: steps(func(s *stepRule) {
		s.overallStep, s.topicStep = stepCurve{first: 0.1}, stepCurve{first: 0.2}
	})}
}

func ceilingRule() *rule {
	return &rule{name: "oracle", shape: both, ceiling: true, make: func(c *child, _ float64) estimator { return oracle{c: c} }}
}

// withCalibration is a set's own rules between the service and the slow
// constant step before them and the ceiling after them, each rule once.
func withCalibration(own []*rule) []*rule {
	all := []*rule{serviceRule(), slowConstantRule()}
	for _, r := range own {
		if !containsRule(all, r) {
			all = append(all, r)
		}
	}
	return append(all, ceilingRule())
}

// sweepExperiment names the children the sweep and its refinement draw: their
// own, so that the candidates put forward are measured in the decision run on
// children they were not picked on.
const sweepExperiment = "sweep"

// sweepRules are the sweep's: every step of the grid and the two speeds at
// every threshold.
func sweepRules() ([]*rule, error) {
	var own []*rule
	for _, u := range sweepGrid.every() {
		own = append(own, u.rule())
	}
	for _, threshold := range switchThresholds {
		own = append(own, switchRule(threshold))
	}
	return withCalibration(own), nil
}

// sweepBest is the step the sweep found best by the rule the bench's document
// gives, which the refinement looks around. No step met every constraint as
// the sweep read them; this one misses three — the corridor of the child who
// learns at half speed, and the overall rank's changes late in a run, for that
// child and for one who stays put — which none missed fewer of, and of those
// it scores highest.
var sweepBest = &uncertainStep{afterSeries: 0.1, topicSpread: 0.7, addedOverall: 0.003, addedTopic: 0.006, limit: 0.3}

var errNoSweepBest = errors.New("learners: the refinement looks around the sweep's best step, which is not set yet")

// refineRules are the refinement's: the steps one number away from the
// sweep's best on the ladders but off the sweep's grid, and the backup with
// the best's numbers and around them.
func refineRules() ([]*rule, error) {
	if sweepBest == nil {
		return nil, errNoSweepBest
	}
	var own []*rule
	for _, u := range sweepBest.neighbours(&refineLadders) {
		if !sweepGrid.holds(u) {
			own = append(own, u.rule())
		}
	}
	for _, b := range sweepBest.storedAround(&refineLadders) {
		own = append(own, b.rule())
	}
	return withCalibration(own), nil
}

// nomination is what the sweep and its refinement put forward for the
// decision run, by the rule the bench's document gives: the steps from counted
// answers, the best of them, the backup, and the threshold of the two speeds.
type nomination struct {
	steps     []uncertainStep
	best      *uncertainStep
	backup    *storedStep
	threshold float64
}

// nominated is the decision run's nomination, frozen before that run. No step
// of the sweep or its refinement met every constraint or missed one alone, so
// the order the sweep's best was found by — the fewest constraints missed, then
// the score — stood in for meeting them: the five first, each missing three,
// and the first that adds no number and the first that adds one; the backup
// and the two speeds that came first in the same order; and the first of all
// as the best.
var nominated = nomination{
	steps: []uncertainStep{
		{afterSeries: 0.13, topicSpread: 0.7, addedOverall: 0.003, addedTopic: 0.006, limit: 0.3},
		{afterSeries: 0.1, topicSpread: 0.7, addedOverall: 0.003, addedTopic: 0.006, limit: 0.3},
		{afterSeries: 0.1, topicSpread: 0.7, addedOverall: 0.003, addedTopic: 0.003, limit: 0.3},
		{afterSeries: 0.1, topicSpread: 0.7, addedOverall: 0.003, limit: 0.3},
		{afterSeries: 0.05, topicSpread: 0.7, addedOverall: 0.003, addedTopic: 0.006, limit: 0.3},
		{afterSeries: 0.17, topicSpread: 0.7, limit: 0.3},
		{afterSeries: 0.1, topicSpread: 0.7, limit: noLimit},
	},
	best:      &uncertainStep{afterSeries: 0.13, topicSpread: 0.7, addedOverall: 0.003, addedTopic: 0.006, limit: 0.3},
	backup:    &storedStep{topicSpread: 0.7, addedOverall: 0.0015, addedTopic: 0.006, limit: 0.3},
	threshold: 2,
}

var errNoNomination = errors.New("learners: the decision run takes the candidates the sweep put forward, which are not set yet")

// decisionRules are the decision run's: the steps put forward, the best one's
// parts taken away in turn, the backup, the full filter and the estimate over
// the whole history on the best one's numbers, the two speeds, and the floors
// of the exit.
func decisionRules() ([]*rule, error) {
	if nominated.best == nil || nominated.backup == nil {
		return nil, errNoNomination
	}
	best := *nominated.best
	var own []*rule
	for _, u := range nominated.steps {
		own = append(own, u.rule())
	}
	for _, u := range best.withoutParts() {
		own = append(own, u.rule())
	}
	own = append(own, nominated.backup.rule(), filterRule(best), historyRule(best.topicSpread), switchRule(nominated.threshold))
	for _, floor := range exitFloors {
		own = append(own, floorRule(floor))
	}
	return withCalibration(own), nil
}

// chosenStep names the rule the decision run chose, as its criterion reads
// the choice. No candidate met every constraint and was better than the
// service, so the choice came to the exit: the floor of the highest score
// among those that meet the constraints of not worse and of the screen.
const chosenStep = "floor_0.05"

// The refusals of the runs that follow the choice: there is no step chosen
// yet, or the one chosen is the exit, which is taken without a confirmation
// and has no parts to take away, so the held-out children stay held out for a
// candidate.
var (
	errNothingChosen = errors.New("learners: no step is chosen yet, so there is nothing to confirm or take apart")
	errExitTaken     = errors.New("learners: the step chosen is the exit, which is taken without a confirmation and has no parts to take apart; the held-out children stay held out")
)

// chosenCandidate is the rule the decision run chose, when it is a candidate
// rather than the exit.
func chosenCandidate() (*rule, error) {
	chosen, err := chosenRule()
	if err != nil {
		return nil, err
	}
	return confirmable(chosen)
}

// confirmable is a chosen rule as the runs that follow the choice take it: a
// candidate as it is, the exit refused.
func confirmable(r *rule) (*rule, error) {
	if !isCandidateIn(r, forStep) {
		return nil, errExitTaken
	}
	return r, nil
}

// chosenRule is the rule the decision run chose, among its own.
func chosenRule() (*rule, error) {
	all, err := decisionRules()
	if err != nil || chosenStep == "" {
		return nil, errNothingChosen
	}
	for _, r := range all {
		if r.name == chosenStep {
			return r, nil
		}
	}
	return nil, errNothingChosen
}

// partsRules are the chosen step and its parts taken away in turn, run on the
// working seeds when the chosen step is not the one whose parts the decision
// run took away.
func partsRules() ([]*rule, error) {
	chosen, err := chosenCandidate()
	if err != nil {
		return nil, err
	}
	own := []*rule{chosen}
	for _, u := range slices.Concat(nominated.steps, nominated.best.withoutParts()) {
		if u.name() == chosen.name {
			for _, part := range u.withoutParts() {
				own = append(own, part.rule())
			}
		}
	}
	return withCalibration(own), nil
}

// confirmationRules are the last choice's candidate alone beside the service,
// the slow constant step and the ceiling — and, for a model of mastery, the
// baseline it was chosen against: a set held out is confirmed on once, by one
// candidate. The step chosen is the exit, which is taken without one, so it is
// the model of mastery chosen over it that is confirmed, once one is.
func confirmationRules() ([]*rule, error) {
	if chosenMastery != "" {
		model, err := chosenModel()
		if err != nil {
			return nil, err
		}
		step, err := chosenRule()
		if err != nil {
			return nil, err
		}
		return withCalibration([]*rule{step, model}), nil
	}
	chosen, err := chosenCandidate()
	if err != nil {
		return nil, err
	}
	return withCalibration([]*rule{chosen}), nil
}

// confirmationCriterion is the criterion the confirmation reads: the choice
// of mastery's, once a model of it is chosen, or else the step's.
func confirmationCriterion() *criterion {
	if chosenMastery != "" {
		return masteryCriterion()
	}
	return stepCriterion()
}

// benchRules are the bench's own: the service's path — the shrinking step in
// the structure of both — and the rules it is compared with: the step kept
// constant, at the service's first step and at half of it; the shrinking step
// with a floor under the overall level's; the service's step with no trial
// series; Glicko-2 that knows a child can guess, with one level and with a
// level per topic; the step chosen, once one is; and, last, the oracle, which
// every rule is measured against as its ceiling. A rule's name and structure
// name its cells, and so every seed its numbers draw.
func benchRules() []*rule {
	glicko := func(shape structure) *rule {
		return &rule{name: "glicko2_floor", shape: shape, make: func(_ *child, start float64) estimator { return newGlicko(shape, start) }}
	}
	all := []*rule{
		serviceRule(),
		{name: "constant", shape: both, trial: true, make: steps(func(s *stepRule) { s.overallStep.decay, s.topicStep.decay = 0, 0 })},
		slowConstantRule(),
		floorRule(0.05),
		{name: "no_trial", shape: both, make: steps(nil)},
		glicko(general),
		glicko(topics),
	}
	if chosen, err := chosenRule(); err == nil && !containsRule(all, chosen) {
		all = append(all, chosen)
	}
	if model, err := chosenModel(); err == nil {
		all = append(all, model)
	}
	return append(all, ceilingRule())
}

// ruleSet is a set of rules a run can be given: its name, its rules, the seed
// and the name of the children it draws when it has children of its own, and
// the directory its results go to within the one a run writes to.
type ruleSet struct {
	name       string
	rules      func() ([]*rule, error)
	seed       uint64
	experiment string
	directory  string
	// criterion is the criterion a run of the set reads; nil for the step's.
	criterion func() *criterion
}

// ownSeeds says whether a set draws children of its own whatever the command
// line says.
func (s ruleSet) ownSeeds() bool { return s.experiment != "" }

// The names of the sets the command line names.
const (
	benchSet        = "bench"
	confirmationSet = "confirmation"
)

// ruleSets are the sets a run can be given: the bench's own, the sweep and
// its refinement on children of their own, the decision run, the parts of the
// chosen step, and the confirmation on the held-out children.
func ruleSets() []ruleSet {
	return []ruleSet{
		{name: benchSet, rules: func() ([]*rule, error) { return benchRules(), nil }},
		{name: "sweep", rules: sweepRules, seed: paperSeed, experiment: sweepExperiment, directory: "sweep"},
		{name: "refine", rules: refineRules, seed: paperSeed, experiment: sweepExperiment, directory: "refine"},
		{name: "decision", rules: decisionRules, directory: "decision"},
		{name: "parts", rules: partsRules, directory: "parts"},
		{name: "mastery-pilot", rules: masteryRules, seed: paperSeed, experiment: sweepExperiment, directory: "mastery-pilot", criterion: masteryCriterion},
		{name: "mastery", rules: masteryRules, directory: "mastery", criterion: masteryCriterion},
		{
			name: confirmationSet, rules: confirmationRules, seed: heldOutSeed, experiment: heldOutExperiment, directory: heldOutDirectory,
			criterion: confirmationCriterion,
		},
	}
}

// criterionOf is the criterion a run of a set reads.
func criterionOf(setName string) *criterion {
	if set, known := ruleSetNamed(setName); known && set.criterion != nil {
		return set.criterion()
	}
	return stepCriterion()
}

// ruleSetNamed is the set of that name.
func ruleSetNamed(name string) (ruleSet, bool) {
	for _, s := range ruleSets() {
		if s.name == name {
			return s, true
		}
	}
	return ruleSet{}, false
}

// countOf is how many of the conditions hold.
func countOf(conditions ...bool) int {
	n := 0
	for _, c := range conditions {
		if c {
			n++
		}
	}
	return n
}

// addedAndLimit names the uncertainties a step adds and its limit.
func addedAndLimit(addedOverall, addedTopic, limit float64) string {
	limitName := "none"
	if !math.IsInf(limit, 1) {
		limitName = decimal(limit)
	}
	return "qt" + decimal(addedOverall) + "_qd" + decimal(addedTopic) + "_L" + limitName
}

// decimal is a number as a rule's name writes it: as short as it reads back.
func decimal(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
