package progress

import (
	"cmp"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

const (
	// JudgedAnswers is how many answers a topic must rest on before the review
	// says how it stands: as many as mastering it takes. A topic with fewer is
	// too early to judge, which is no verdict on it.
	JudgedAnswers = profile.MasteryAnswers
	// ReviewedTopics is how many topics the review names on either side at
	// most: the strong ones, and the ones to develop, each of which is given
	// a step.
	ReviewedTopics = 3
	// FailuresInARow is how many wrong answers in a row in a topic make it one
	// to develop: as many as it takes to lose its mastery.
	FailuresInARow = profile.MasteryLostAfter
	// FewestHints is how many of a topic's latest answers must have used the
	// hint, and at least half of them, before the hint is a reason to develop
	// the topic.
	FewestHints = 2
)

// apart is how far a topic's level stands from the overall one, or moved over
// the week, before the review says so: half the corridor, half the step one
// rank takes. A move of a point or two is the corridor's own swing, and says
// nothing of what the child knows.
func apart() float64 { return rating.CorridorWidth() / 2 }

// Reason is why a topic is named as strong or as one to develop.
type Reason string

const (
	// ReasonMastered is a topic mastered at the level its tasks come from.
	ReasonMastered Reason = "mastered"
	// ReasonHigh is a topic half a rank or more above the overall level.
	ReasonHigh Reason = "high"
	// ReasonRose is a strong topic whose own level rose half a rank or more
	// over the week.
	ReasonRose Reason = "rose"
	// ReasonLow is a topic half a rank or more below the overall level.
	ReasonLow Reason = "low"
	// ReasonFailures is a topic answered wrongly FailuresInARow times in a row
	// or more.
	ReasonFailures Reason = "failures"
	// ReasonHints is a topic whose hint was used in at least half of its
	// latest answers, and in FewestHints of them.
	ReasonHints Reason = "hints"
	// ReasonTrap is a topic whose own latest answers fell for one trap as
	// often as makes a mistake one that repeats.
	ReasonTrap Reason = "trap"
	// ReasonFell is a topic whose own level fell half a rank or more over the
	// week.
	ReasonFell Reason = "fell"
)

// StepKind is what a step of the review advises.
type StepKind string

const (
	// StepTrap is the advice the catalog gives for a trap: in a topic, for
	// the trap its answers keep falling for; for no topic, for the mistake
	// that repeats most across them.
	StepTrap StepKind = "trap"
	// StepRhythm is a few short tasks a day in a topic that fell, until it
	// comes back.
	StepRhythm StepKind = "rhythm"
	// StepPractice is a few more tasks in a topic below the overall level or
	// answered wrongly in a row, set where the child stands.
	StepPractice StepKind = "practice"
	// StepUnaided is trying each task of a topic without the hint first.
	StepUnaided StepKind = "unaided"
)

// Review is the review of where the child stands, for the adult: the topics
// that are strong, the ones to develop with what keeps them back, the ones
// too early to judge, and the steps to take, each a piece of advice.
type Review struct {
	Strong  []Judged
	Develop []Judged
	Early   []string
	Steps   []Step
}

// Judged is a topic the review names: why, the trap its answers keep falling
// for when that is a reason, and, for one to develop, whether its last answer
// was right — a move has begun.
type Judged struct {
	Topic   string
	Reasons []Reason
	Trap    string
	Moving  bool
}

// Step is one step of the review: what it advises, the topic it is for — none
// for the step that holds for every topic — and the trap whose advice it is.
type Step struct {
	Kind  StepKind
	Topic string
	Trap  string
}

// ReviewOf is the review of where the child stands, worked out from the
// profile and its summary — the topics as listed, with mastery and reach, and
// what the rule sets next — and from the mistakes that repeat across the
// window at repeats, as the map of them is drawn. Nothing it reads is the
// pace of an answer, a task left without one, or the task on the card. During
// the trial series there is none: the series is still finding where the child
// stands.
//
// A topic is judged once it rests on JudgedAnswers answers, and before that is
// too early to judge. A topic with a reason to develop it is one to develop
// and never a strong one: the review is there to help. The ones to develop
// are those the rule could set now, the topic it sets next first, so that the
// first step agrees with it, and the lowest after it; the strong ones are the
// mastered first and the highest after them; ties keep the catalog's order.
func ReviewOf(p *profile.Profile, catalog tutor.Catalog, summary *Summary, mistakes []Mistake, repeats int, today profile.Date) *Review {
	if summary.Trial != nil {
		return nil
	}
	weekStart, _ := p.WeekStart(today)
	review := &Review{Strong: []Judged{}, Develop: []Judged{}, Early: []string{}, Steps: []Step{}}
	var strong, develop []weighed
	for i := range summary.Topics {
		topic := &summary.Topics[i]
		if topic.Answers == 0 {
			continue
		}
		if topic.Answers < JudgedAnswers {
			review.Early = append(review.Early, topic.ID)
			continue
		}
		kept := p.Topics[topic.ID]
		moved := movedOverTheWeek(weekStart, topic.ID, kept.Delta)
		if judged := toDevelop(p.Recent, catalog, topic.ID, &kept, moved, repeats); judged != nil {
			if topic.InReach {
				develop = append(develop, weighed{judged: *judged, delta: kept.Delta, first: topic.ID == summary.Next.Topic})
			}
			continue
		}
		if judged := strengthOf(topic, kept.Delta, moved); judged != nil {
			strong = append(strong, weighed{judged: *judged, delta: kept.Delta, first: topic.Mastered})
		}
	}
	review.Develop = firstOf(develop, func(a, b weighed) int { return cmp.Compare(a.delta, b.delta) })
	review.Strong = firstOf(strong, func(a, b weighed) int { return cmp.Compare(b.delta, a.delta) })
	review.Steps = stepsOf(review.Develop, mistakes)
	return review
}

// weighed is a judged topic with what orders it among the others: whether it
// goes before the rest, and its correction.
type weighed struct {
	judged Judged
	delta  float64
	first  bool
}

// firstOf are the first ReviewedTopics of the topics: those that go first,
// then the rest in the order by gives them, ties kept in the order they came,
// which is the catalog's.
func firstOf(topics []weighed, by func(a, b weighed) int) []Judged {
	slices.SortStableFunc(topics, func(a, b weighed) int {
		if a.first != b.first {
			if a.first {
				return -1
			}
			return 1
		}
		return by(a, b)
	})
	named := []Judged{}
	for _, topic := range topics[:min(len(topics), ReviewedTopics)] {
		named = append(named, topic.judged)
	}
	return named
}

// movedOverTheWeek is how far a topic's correction moved since the week
// began: nothing when the week cannot be told or nothing moved in it, and
// nothing for a topic first answered in it, which stood nowhere before.
func movedOverTheWeek(start *profile.RatingDay, id string, delta float64) float64 {
	if start == nil {
		return 0
	}
	before, met := start.Deltas[id]
	if !met {
		return 0
	}
	return delta - before
}

// toDevelop is a topic judged as one to develop, with every reason it is one,
// or nil when it has none.
func toDevelop(window []profile.Answer, catalog tutor.Catalog, id string, kept *profile.Topic, moved float64, repeats int) *Judged {
	answers := answersIn(window, id)
	trap := repeatingTrap(answers, catalog, repeats)
	var reasons []Reason
	if kept.Delta <= -apart() {
		reasons = append(reasons, ReasonLow)
	}
	if kept.WrongStreak >= FailuresInARow {
		reasons = append(reasons, ReasonFailures)
	}
	if hinted := hintsIn(answers); hinted >= FewestHints && 2*hinted >= len(answers) {
		reasons = append(reasons, ReasonHints)
	}
	if trap != "" {
		reasons = append(reasons, ReasonTrap)
	}
	if moved <= -apart() {
		reasons = append(reasons, ReasonFell)
	}
	if len(reasons) == 0 {
		return nil
	}
	return &Judged{Topic: id, Reasons: reasons, Trap: trap, Moving: kept.WrongStreak == 0}
}

// strengthOf is a topic judged as a strong one, with every reason it is one,
// or nil when it is not: mastered, or half a rank above the overall level —
// and, for one that is, a rise over the week.
func strengthOf(topic *Topic, delta, moved float64) *Judged {
	var reasons []Reason
	if topic.Mastered {
		reasons = append(reasons, ReasonMastered)
	}
	if delta >= apart() {
		reasons = append(reasons, ReasonHigh)
	}
	if len(reasons) == 0 {
		return nil
	}
	if moved >= apart() {
		reasons = append(reasons, ReasonRose)
	}
	return &Judged{Topic: topic.ID, Reasons: reasons}
}

// answersIn are the answers of the window given in a topic: a task left
// without one is no answer.
func answersIn(window []profile.Answer, topic string) []profile.Answer {
	var answers []profile.Answer
	for i := range window {
		if window[i].Topic == topic && !window[i].Skipped {
			answers = append(answers, window[i])
		}
	}
	return answers
}

// hintsIn is how many of the answers used the hint.
func hintsIn(answers []profile.Answer) int {
	hinted := 0
	for i := range answers {
		if answers[i].HintUsed {
			hinted++
		}
	}
	return hinted
}

// repeatingTrap is the trap the answers keep falling for, as often as makes a
// mistake one that repeats, or none: the most frequent, and of two as
// frequent, the one fallen for last, as the map of mistakes orders them.
func repeatingTrap(answers []profile.Answer, catalog tutor.Catalog, repeats int) string {
	var (
		chosen string
		most   found
	)
	for trap, counted := range trapsIn(answers, catalog) {
		if counted.times >= repeats && (counted.times > most.times || counted.times == most.times && counted.last > most.last) {
			chosen, most = trap, counted
		}
	}
	return chosen
}

// stepsOf are the steps of the review: one for each topic to develop, by what
// keeps it back — the advice of the trap it keeps falling for, a few short
// tasks a day when it fell, a few more tasks when it stands low or was
// answered wrongly in a row, and the tasks tried without the hint when the
// hint was its only reason — and one for every topic, the advice of the
// mistake that repeats most across them, when no step has given it already.
func stepsOf(develop []Judged, mistakes []Mistake) []Step {
	steps := []Step{}
	advised := map[string]bool{}
	for _, topic := range develop {
		switch {
		case topic.Trap != "":
			steps = append(steps, Step{Kind: StepTrap, Topic: topic.Topic, Trap: topic.Trap})
			advised[topic.Trap] = true
		case slices.Contains(topic.Reasons, ReasonFell):
			steps = append(steps, Step{Kind: StepRhythm, Topic: topic.Topic})
		case slices.Contains(topic.Reasons, ReasonLow) || slices.Contains(topic.Reasons, ReasonFailures):
			steps = append(steps, Step{Kind: StepPractice, Topic: topic.Topic})
		default:
			steps = append(steps, Step{Kind: StepUnaided, Topic: topic.Topic})
		}
	}
	for _, mistake := range mistakes {
		if !advised[mistake.Trap] {
			return append(steps, Step{Kind: StepTrap, Trap: mistake.Trap})
		}
	}
	return steps
}
