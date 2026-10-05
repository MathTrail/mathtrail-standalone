package main

import (
	"math"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/report"
)

// The lines the measures are read against.
const (
	corridorLow     = 0.70
	corridorHigh    = 0.85
	masteredAt      = 0.775
	masteredLoosely = 0.70
	tooHard         = 0.50
	tooEasy         = 0.95
	settled         = 0.5
	settledFor      = 10
	lagFrom         = 100
	firstTasks      = 10
	firstAnswers    = 15
	calibrationBins = 10
)

// checkpoints are the answers after which the error of the level is taken.
var checkpoints = []int{5, 10, 20, 50, 100, 200}

// placements are the answers after which the overall level's error is taken,
// for how well a child placed a level off is placed.
var placements = []int{5, 10}

// longestChain is the most eligible attempts the chance of a false "mastered"
// is read within.
const longestChain = 50

// chainLengths are the numbers of eligible attempts the chance of a false
// "mastered" is read within.
var chainLengths = []int{3, 5, 10, 20, longestChain}

// bin is one bin of calibration: the predicted chances and the answers in it.
type bin struct {
	predicted, observed float64
	n                   int
}

// topicLevel is a topic at a grade level.
type topicLevel struct {
	topic string
	level rating.GradeLevel
}

// lateItem is a topic and level the child truly masters: how many answers in
// the topic it took the rule to declare it, or that it never did.
type lateItem struct {
	at       topicLevel
	answers  int
	declared bool
}

// topicSet is a set of topics that keeps the order they came in, so that a sum
// over them is taken in the same order, and comes to the same bits, every run.
type topicSet struct {
	in    map[string]bool
	order []string
}

func (t *topicSet) add(topic string) {
	if t.in == nil {
		t.in = map[string]bool{}
	}
	if !t.in[topic] {
		t.in[topic] = true
		t.order = append(t.order, topic)
	}
}

// pairTrack follows one topic of one child for the chance of a false
// "mastered": its eligible attempts while the child is not truly mastered at
// their tasks' level, and the attempt the rule declared mastery at.
type pairTrack struct {
	attempts   int
	clean      bool
	declaredAt int
	chances    []float64
}

// childResult is everything measured on one child under one rule.
type childResult struct {
	errorRMS, errorMean     map[int]float64
	brier                   float64
	predicted, observed     float64
	answers                 int
	bins                    [calibrationBins]bin
	inside, below, over     int
	declared, wrongly       int
	looselyWrong, heldBelow int
	atTask, wronglyAtTask   int
	late                    map[topicLevel]*lateItem
	lateItems               []*lateItem
	lag                     float64
	lagged                  int
	answeredTopics          topicSet
	sinceJump               topicSet
	jumpFollowed            int
	jumpSettledFor          int
	jumpDone                bool
	placed                  map[int]float64
	wrongRun, longest       int
	hardFirst               int
	pairs                   map[string]*pairTrack
	reachable               int
	screen                  screenRecord
	keptUp                  keptUp
	shown                   []*report.Followed
	following               map[string]*report.Followed
}

func newChildResult(*child) *childResult {
	return &childResult{
		errorRMS: map[int]float64{}, errorMean: map[int]float64{}, placed: map[int]float64{},
		late: map[topicLevel]*lateItem{}, pairs: map[string]*pairTrack{}, following: map[string]*report.Followed{},
	}
}

// observedAnswer is what is known of an answer before the rule takes it in —
// the rule's level in the topic and overall among it — whether the profile
// held the task's topic mastered at its level or above before it, and whether
// the progress showed the topic mastered.
type observedAnswer struct {
	estimate, overall, truth, predicted, chance float64
	correct, heldMastered, shownBefore          bool
}

// observeBefore measures an answer at the moment the task is set: the rule's
// prediction against the outcome, where the task's true chance falls, whether
// the topic's ladder held a task in the corridor at all, and the error of the
// rule's level in the topic.
func (s *session) observeBefore(brief *profile.Brief, beta, truth float64, correct bool, k int) observedAnswer {
	r := s.result
	topic := brief.TargetConcept
	o := observedAnswer{
		estimate: s.est.level(topic), overall: s.est.overall(), truth: s.c.level(topic),
		predicted: s.est.chance(topic, beta), chance: truth,
		correct: correct, heldMastered: masteredAtOrAbove(s.p, topic, brief.GradeLevel),
		shownBefore: s.showsMastered(topic),
	}
	if s.reachable(topic) {
		r.reachable++
	}
	score := 0.0
	if correct {
		score = 1
	}
	r.brier += (o.predicted - score) * (o.predicted - score)
	r.predicted += o.predicted
	r.observed += score
	r.answers++
	b := &r.bins[min(calibrationBins-1, int(o.predicted*calibrationBins))]
	b.predicted, b.observed, b.n = b.predicted+o.predicted, b.observed+score, b.n+1
	switch {
	case truth < corridorLow:
		if truth < tooHard {
			r.below++
		}
	case truth <= corridorHigh:
		r.inside++
	case truth > tooEasy:
		r.over++
	}
	if k >= lagFrom {
		r.lag += o.estimate - o.truth
		r.lagged++
	}
	r.firstTasks(k, truth, correct)
	r.answeredTopics.add(topic)
	return o
}

// firstTasks keeps what placement is read from: how many of the first tasks
// were too hard, and the longest run of wrong answers among the first ones.
func (r *childResult) firstTasks(k int, truth float64, correct bool) {
	if k < firstTasks && truth < tooHard {
		r.hardFirst++
	}
	if k >= firstAnswers {
		return
	}
	if correct {
		r.wrongRun = 0
		return
	}
	r.wrongRun++
	r.longest = max(r.longest, r.wrongRun)
}

// after measures what the answer did: whether it declared a topic mastered,
// and whether rightly, read at the level the topic is held mastered at, which
// a rule may set below the task's; how long a mastery the child has was left
// undeclared; the eligible attempts a false "mastered" is counted over; and
// what the report would read of it in the service's log.
func (r *childResult) after(s *session, brief *profile.Brief, recorded *profile.Recorded, o *observedAnswer, k int) {
	topic, level := brief.TargetConcept, brief.GradeLevel
	truth := s.c.truthAt(topic, level)
	if recorded.Mastered {
		r.declared++
		held := level
		if at := s.p.Topics[topic].MasteredLevel; at != nil {
			held = *at
		}
		heldTruth := s.c.truthAt(topic, held)
		if held.Shift() < level.Shift() {
			r.heldBelow++
		} else {
			r.atTask++
			if heldTruth < masteredAt {
				r.wronglyAtTask++
			}
		}
		if heldTruth < masteredAt {
			r.wrongly++
		}
		if heldTruth < masteredLoosely {
			r.looselyWrong++
		}
	}
	r.followLate(s, topicLevel{topic: topic, level: level}, truth, o.heldMastered, recorded.Mastered)
	r.followPair(topic, recorded, truth, o.chance)
	r.followJump(s, topic, k)
	r.screen.after(s, topic, o, k)
	r.live(s, recorded, o)
}

// reachable says whether the topic's ladder holds a task whose true chance,
// for the child as they stand, lies in the corridor: whether any rule could
// have handed out a task in it, whatever it knew of the child.
func (s *session) reachable(topic string) bool {
	for _, point := range s.w.ladders[topic] {
		if chance := s.c.chance(topic, point.Beta()); chance >= corridorLow && chance <= corridorHigh {
			return true
		}
	}
	return false
}

// followLate counts, for every topic and level the child truly masters, the
// answers in the topic from the first at which it is truly mastered until the
// rule declares it at that level or above. Whether the rule had declared it is
// read as the profile stood before the answer, so a mastery declared on the
// very answer the child is first truly mastered at counts as declared at once.
func (r *childResult) followLate(s *session, at topicLevel, truth float64, heldBefore, mastered bool) {
	for _, item := range r.lateItems {
		if !item.declared && item.at.topic == at.topic {
			item.answers++
		}
	}
	if _, started := r.late[at]; !started && truth >= masteredAt && !heldBefore {
		item := &lateItem{at: at}
		r.late[at], r.lateItems = item, append(r.lateItems, item)
	}
	if !mastered {
		return
	}
	for _, item := range r.lateItems {
		if !item.declared && item.at.topic == at.topic && masteredAtOrAbove(s.p, at.topic, item.at.level) {
			item.declared = true
		}
	}
}

// masteredAtOrAbove says whether the profile holds a topic mastered at a level
// or a higher one.
func masteredAtOrAbove(p *profile.Profile, topic string, level rating.GradeLevel) bool {
	summary := p.Topics[topic]
	return summary.MasteredLevel != nil && summary.MasteredLevel.Shift() >= level.Shift()
}

// followPair counts a topic's eligible attempts — tasks whose chance, as the
// service predicted it, was at most the corridor's middle — while the child is
// not truly mastered at their level, and the attempt mastery was declared at.
func (r *childResult) followPair(topic string, recorded *profile.Recorded, truth, chance float64) {
	pair, seen := r.pairs[topic]
	if !seen {
		pair = &pairTrack{clean: true}
		r.pairs[topic] = pair
	}
	if !pair.clean || pair.declaredAt > 0 {
		return
	}
	// A child truly mastered at the task's level, on an eligible attempt or
	// when mastery is declared, takes the topic out of the count: what is
	// counted is a declaration the child had not earned.
	eligible := recorded.Probability <= rating.CorridorMiddle
	if (eligible || recorded.Mastered) && truth >= masteredAt {
		pair.clean = false
		return
	}
	if eligible {
		pair.attempts++
		pair.chances = append(pair.chances, chance)
	}
	// Mastery can be declared on an answer that is not eligible, once the run
	// it needs is there; it counts at the eligible attempts made by then.
	if recorded.Mastered {
		pair.declaredAt = max(pair.attempts, 1)
	}
}

// followJump counts the answers from a child's jump until the error of the
// rule, averaged over the topics answered since the jump, stays within a bound
// for a run of answers. The errors are averaged with their signs: what is
// followed is the lag the jump left, which every topic shares, not the noise
// each topic's level carries with or without a jump.
func (r *childResult) followJump(s *session, topic string, k int) {
	// The jump comes after answer jumpAt, so the first answer after it is the
	// one at index jumpAt.
	if s.c.jumpAt < 0 || k < s.c.jumpAt || r.jumpDone {
		return
	}
	r.sinceJump.add(topic)
	sum := 0.0
	for _, answered := range r.sinceJump.order {
		sum += s.est.level(answered) - s.c.level(answered)
	}
	r.jumpFollowed++
	if math.Abs(sum/float64(len(r.sinceJump.order))) >= settled {
		r.jumpSettledFor = 0
		return
	}
	r.jumpSettledFor++
	if r.jumpSettledFor >= settledFor {
		r.jumpDone = true
		r.jumpFollowed -= settledFor
	}
}

// checkpoint takes the error of the level after the answers it is read at,
// over the topics the child has answered, and the overall level's error after
// five and ten answers.
func (r *childResult) checkpoint(s *session, k int) {
	answers := k + 1
	if slices.Contains(placements, answers) {
		r.placed[answers] = s.est.overall() - s.c.theta
	}
	if !slices.Contains(checkpoints, answers) {
		return
	}
	squares, sum := 0.0, 0.0
	for _, topic := range r.answeredTopics.order {
		e := s.est.level(topic) - s.c.level(topic)
		squares += e * e
		sum += e
	}
	n := float64(len(r.answeredTopics.order))
	r.errorRMS[answers] = math.Sqrt(squares / n)
	r.errorMean[answers] = sum / n
}
