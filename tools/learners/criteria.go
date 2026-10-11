package main

import (
	"math"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// choiceKind is which choice a candidate is put forward in: the step, the
// rule of mastery over the step chosen, or a debt over the model chosen. The
// zero kind is the step's, which every candidate of the step was put forward
// in before mastery had a choice.
type choiceKind int

// The choices.
const (
	forStep choiceKind = iota
	forMastery
	forDebt
)

// criterion is what one choice is made by, written down before any of its
// candidates ran: which candidates it weighs; the baseline every comparison
// is with; the goals, the measures of not worse and the score's generators
// and measures; whether the score's way ends at the ceiling or at perfect;
// whether the exit is the baseline itself, kept as it is; and how the
// criterion names the baseline, its main candidates, the way its score
// closes and the rule its candidates' distances are from.
type criterion struct {
	choice         choiceKind
	baseline       string
	goals          func() []goal
	notWorse       []notWorse
	scored         []scoredGenerator
	toPerfect      bool
	exitIsBaseline bool
	against        string
	main           string
	way            string
	distanceFrom   string
}

// endOfWay names where the criterion's score ends its way.
func (c *criterion) endOfWay() string {
	if c.toPerfect {
		return "perfect"
	}
	return "the ceiling"
}

// checks are the criterion's checks of not worse: every measure of it on
// every generator it is read on.
func (c *criterion) checks() []notWorse {
	var all []notWorse
	for _, g := range allGenerators {
		for _, nw := range c.notWorse {
			if nw.only != nil && !slices.Contains(nw.only, g) {
				continue
			}
			nw.generator = g
			all = append(all, nw)
		}
	}
	return all
}

// stepCriterion is the criterion the step is chosen by.
func stepCriterion() *criterion {
	return &criterion{
		choice: forStep, baseline: serviceRule().name, goals: stepGoals, notWorse: stepNotWorse, scored: scoredGenerators,
		against: "the service", main: "main step", distanceFrom: "the service's step",
		way: "the share of the way from the service to the ceiling a rule closes, with its 95 % interval; " +
			"the ceiling is no candidate, and is read to show the way",
	}
}

// masteryCriterion is the criterion the rule of mastery is chosen by, over
// the step chosen: every comparison is with the step chosen under the
// service's rule of mastery, which is also the exit, kept as it is when no
// candidate meets the criterion; the goals hold false masteries and the wait
// for mastery, and the score closes the way to no false mastery and no wait.
func masteryCriterion() *criterion {
	return &criterion{
		choice: forMastery, baseline: masteryBaseline, goals: masteryGoals, notWorse: masteryNotWorse, scored: masteryScored(),
		toPerfect: true, exitIsBaseline: true,
		against: "the baseline", main: "main rule of mastery", distanceFrom: "the service's rule of mastery",
		way: "the share of the way from the baseline, " + masteryBaseline + "/both — the step chosen under the service's rule of mastery — " +
			"to perfect, no false mastery and no wait, a rule closes, with its 95 % interval",
	}
}

// stepGoals are the step's goals: the main learner's lag and corridor, the
// jump, and the screen.
func stepGoals() []goal {
	all := []goal{
		{
			name: "the lag at most " + percentOf(lagShare) + " of the service's", generator: learningHalf, metric: "r6_lag", size: true, atMost: true,
			bound: func(cr *criterionRun) (float64, bool) {
				lag, has := cr.baselineValue(learningHalf, "r6_lag")
				return lagShare * math.Abs(lag), has
			},
		},
		{
			name: "the corridor a share of the way to the ceiling", generator: learningHalf, metric: "r3_inside",
			bound: func(cr *criterionRun) (float64, bool) {
				service, hasService := cr.baselineValue(learningHalf, "r3_inside")
				ceiling, hasCeiling := cr.ceilingValue(learningHalf, "r3_inside")
				return service + corridorShare*(ceiling-service), hasService && hasCeiling
			},
		},
		{
			name: "children not caught up after the jump at most " + percentOf(unsettledMost), generator: jumping, metric: "r6_jump_unsettled", atMost: true,
			bound: func(*criterionRun) (float64, bool) { return unsettledMost, true },
		},
	}
	return append(all, screenGoals("the service's")...)
}

// screenGoals are the goals of what the child is shown, on the child who
// stays put and on the main learner: in both windows, the card's move and the
// overall rank's changes no more than those of the rule compared with in
// answers 6–20.
func screenGoals(whose string) []goal {
	var all []goal
	early, earlyLabel := screenWindows[0].name(), screenWindows[0].label()
	for _, g := range screenGenerators {
		for _, w := range screenWindows {
			all = append(all,
				goal{
					name: "the card's move in answers " + w.label() + " no more than " + whose + " in answers " + earlyLabel, generator: g,
					metric: "r8_move_p95_" + w.name(), atMost: true, screen: true,
					bound: func(cr *criterionRun) (float64, bool) { return cr.baselineValue(g, "r8_move_p95_"+early) },
				},
				goal{
					name: "the rank's changes in answers " + w.label() + " no more than " + whose + " in answers " + earlyLabel, generator: g,
					metric: "r8_rank_" + w.name(), atMost: true, screen: true,
					bound: func(cr *criterionRun) (float64, bool) { return cr.baselineValue(g, "r8_rank_"+early) },
				},
			)
		}
	}
	return all
}

// stepNotWorse are the measures a step is held to no worse than the
// service's on: the error after 200 answers and the corridor. The share of
// masteries declared falsely is not among them: under the service's rule of
// mastery it grows the closer an estimate follows the child, the ceiling's
// most of all, so it would hold back every step that follows a child better;
// the choice of the rule of mastery, read over the step chosen, holds it.
var stepNotWorse = []notWorse{
	{metric: "r1_rms_200", tolerance: errorTolerance},
	{metric: "r3_inside", higherIsBetter: true, tolerance: corridorTolerance},
}

// The goals of mastery: at most this share of the masteries declared false,
// and the answers until a mastery the child has is declared at most this many
// times the baseline's, both on the child who stays put and on the widest
// spread of topics, on which a margin of mastery is chosen.
const (
	falseMost      = 0.20
	lateTimes      = 1.5
	falseTolerance = 0.02
)

// masteryGenerators are the generators the goals of mastery are read on.
var masteryGenerators = []generator{staticChildren, farTopics}

// masteryGoals are the goals of mastery and the screen's.
func masteryGoals() []goal {
	var all []goal
	for _, g := range masteryGenerators {
		all = append(all,
			goal{
				name: "masteries declared falsely at most " + percentOf(falseMost), generator: g, metric: "r4_false", atMost: true,
				bound: func(*criterionRun) (float64, bool) { return falseMost, true },
			},
			goal{
				name: "the answers until mastery at most 1.5 times the baseline's", generator: g, metric: "r5_late_answers", atMost: true,
				bound: func(cr *criterionRun) (float64, bool) {
					late, has := cr.baselineValue(g, "r5_late_answers")
					return lateTimes * late, has
				},
			},
		)
	}
	return append(all, screenGoals("the baseline's")...)
}

// The price a rule of mastery that declares only what it is sure of may cost
// against the baseline, for the false masteries it saves. The error after 200
// answers may be up to masteryErrorTolerance worse: a topic held mastered
// leaves the rotation, and for a child who moves on its estimate stays where
// it was. On the widest spread of topics, up to wideNeverTolerance more of the
// masteries a child has may go undeclared within a run: a topic far above the
// child's level clears a margin late. Both were set after the decision run of
// mastery, at what the cautious estimate costs there plus resolutionWidths of
// its standard errors, rounded up — the width every tolerance of not worse
// keeps, so that a rule costing that much passes a run of new children 99
// times in 100 — and a run on children no run has drawn is what tests them.
const (
	masteryErrorTolerance = 0.03
	wideNeverTolerance    = 0.08
)

// masteryNotWorse are the measures a rule of mastery is held to no worse than
// the baseline's on: the step's, which a rule of mastery moves by the topics
// it takes out of the rotation, the error within its price; the share of
// masteries declared falsely; and the share of masteries the child has that
// are never declared, since a rule could otherwise wait less only by
// declaring less. That last is read where the goals of mastery are, on
// children who stay put, where whatever is left undeclared is the rule's own
// doing, and on the widest spread of topics within its price: for a child who
// learns, the step chosen trails the child, and a margin on top of the trail
// leaves masteries reached late in a run undeclared by its end, which is the
// step's to answer for.
var masteryNotWorse = []notWorse{
	{metric: "r1_rms_200", tolerance: masteryErrorTolerance},
	{metric: "r3_inside", higherIsBetter: true, tolerance: corridorTolerance},
	{metric: "r4_false", tolerance: falseTolerance},
	{metric: "r5_never", tolerance: falseTolerance, only: []generator{staticChildren}},
	{metric: "r5_never", tolerance: wideNeverTolerance, only: []generator{farTopics}},
}

// The measures the score of mastery reads: the share of masteries declared
// falsely, and the answers until a mastery the child has is declared.
var (
	falseMasteries = badness{metric: "r4_false", size: same}
	lateMastery    = badness{metric: "r5_late_answers", size: same}
)

// masteryScored are the generators the score of mastery is read on, with the
// step's weights, each on false masteries and the wait for mastery.
func masteryScored() []scoredGenerator {
	all := make([]scoredGenerator, 0, len(scoredGenerators))
	for _, sg := range scoredGenerators {
		all = append(all, scoredGenerator{generator: sg.generator, weight: sg.weight, measures: []badness{falseMasteries, lateMastery}})
	}
	return all
}

// debtBaseline is the model every debt is compared with: the step chosen
// under the rule of mastery chosen, which is the model the service runs. It is
// also the exit of the choice of a debt, kept as it is when no candidate meets
// the criterion.
const debtBaseline = chosenMastery

// debtCriterion is the criterion a debt over the model chosen is chosen by:
// every comparison is with that model, which is also the exit; the goals hold
// what a wrong answer does to the card, what a right answer leaves it at and
// how far a child who stays put is shown above where they stand; and the score
// closes the way to a card no wrong answer moves.
func debtCriterion() *criterion {
	return &criterion{
		choice: forDebt, baseline: debtBaseline, goals: debtGoals, notWorse: debtNotWorse, scored: debtScored(),
		toPerfect: true, exitIsBaseline: true,
		against: "the baseline", main: "main debt", distanceFrom: "the baseline",
		way: "the share of the way from the baseline, " + debtBaseline + "/both — the model the service runs — " +
			"to a card no wrong answer moves, a rule closes, with its 95 % interval",
	}
}

// The goals of a debt, as the author set them before any of its candidates
// ran: the card's fall after a wrong answer at most this share of the
// baseline's; and at most this share of right answers leaving the card where
// it stood, since a debt paid with the whole of a right answer's gain would
// show the child nothing for half their right answers.
const (
	fallShare = 0.5
	stillMost = 0.10
)

// biasMost is how far above where they stand a child who stays put may be
// shown after 200 answers — the price the author accepts while a debt is open:
// 0.15 of a rank, in logits, a rank being one corridor wide.
var biasMost = 0.15 * rating.CorridorWidth()

// debtGoals are the goals of a debt and the screen's: on the child who stays
// put and on the main learner, in both windows, the card's fall after a wrong
// answer and the right answers that leave it still; on the child who stays
// put, the bias after 200 answers.
func debtGoals() []goal {
	var all []goal
	for _, g := range screenGenerators {
		for _, w := range screenWindows {
			fall := "r8_wrong_fall_" + w.name()
			all = append(all,
				goal{
					name: "the card's fall after a wrong answer in answers " + w.label() + " at most " + percentOf(fallShare) + " of the baseline's", generator: g,
					metric: fall, atMost: true, screen: true,
					bound: func(cr *criterionRun) (float64, bool) {
						v, has := cr.baselineValue(g, fall)
						return fallShare * v, has
					},
				},
				goal{
					name: "right answers that leave the card still in answers " + w.label() + " at most " + percentOf(stillMost), generator: g,
					metric: "r8_right_still_" + w.name(), atMost: true, screen: true,
					bound: func(*criterionRun) (float64, bool) { return stillMost, true },
				},
			)
		}
	}
	all = append(all, goal{
		name: "the bias after 200 answers at most 0.15 of a rank", generator: staticChildren, metric: "r1_mean_200", atMost: true,
		bound: func(*criterionRun) (float64, bool) { return biasMost, true },
	})
	return append(all, screenGoals("the baseline's")...)
}

// The tolerances of the checks a debt adds beyond the step's and mastery's:
// the answers until the estimate catches up after a jump, and the longest run
// of wrong answers among the first fifteen, each in answers. Every tolerance is
// widened to the bench's resolution as the others are.
const (
	jumpTolerance = 1
	runTolerance  = 0.1
)

// debtNotWorse are the measures a debt is held to no worse than the
// baseline's on: the step's, the error after 200 answers and the corridor, on
// every generator; false masteries on every generator, since the mastery
// chosen reads the level a debt keeps higher; the lag of every child who
// learns, by the size of its mean, since a debt that keeps the level high could
// pass for following a learner by overshooting; the answers until the estimate
// catches up after a jump or a drop; and on the child placed a level off and
// the one who stays put, the error of the overall level after ten answers and
// the longest run of wrong answers among the first fifteen — a debt follows a
// child placed too high down more slowly.
var debtNotWorse = []notWorse{
	{metric: "r1_rms_200", tolerance: errorTolerance},
	{metric: "r3_inside", higherIsBetter: true, tolerance: corridorTolerance},
	{metric: "r4_false", tolerance: falseTolerance},
	{metric: "r6_lag", size: true, tolerance: errorTolerance, only: []generator{learning, learningHalf, learningFades}},
	{metric: "r6_jump_answers", tolerance: jumpTolerance, only: []generator{jumping, dropping}},
	{metric: "r7_error_10", tolerance: errorTolerance, only: []generator{misplaced, staticChildren}},
	{metric: "r7_longest_wrong", tolerance: runTolerance, only: []generator{misplaced, staticChildren}},
}

// debtScored are the generators the score of a debt is read on, the child who
// stays put and the main learner alike, each on the card's fall after a wrong
// answer in both windows.
func debtScored() []scoredGenerator {
	var falls []badness
	for _, w := range screenWindows {
		falls = append(falls, badness{metric: "r8_wrong_fall_" + w.name(), size: same})
	}
	all := make([]scoredGenerator, 0, len(screenGenerators))
	for _, g := range screenGenerators {
		all = append(all, scoredGenerator{generator: g, weight: 1, measures: falls})
	}
	return all
}
