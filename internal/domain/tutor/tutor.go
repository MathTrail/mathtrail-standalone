// Package tutor decides what the child is asked next.
//
// It reads one profile and writes one brief — the terms of reference the
// chat's model writes a task to: what the task is about, the level it is
// written for and how hard it is inside that level, what it is dressed in,
// which mistakes its wrong options should lead to, and what may not appear in
// it at all.
//
// Nothing here is random and nothing here is a model: the same profile always
// produces the same brief. That is what makes the rule a baseline the model
// can be measured against, and it is why this package needs no mocks to be
// tested — a profile in, a brief out.
//
// The rule reads the summary of each topic and the short window of recent
// answers, never a full history, because a full history is not kept. It never
// reads the parent's notes about the child, which travel to the model as tone
// and are not allowed to change what the child is set, and it never reads the
// child's grade either: the grade decided where the child started, and from
// there the topics within reach, the level and the difficulty all follow what
// the child's answers say.
package tutor

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Traps is how many mistakes a brief names for the wrong options to lead to.
const Traps = 2

// Catalog is what the rule has to know about the content the service ships.
// It is declared here, by the side that needs it, and it speaks in
// identifiers and levels: which topics exist, at which levels each is taught,
// how two traps are told apart when nothing else separates them, and what
// usually goes wrong in a topic at a level.
type Catalog interface {
	// TopicIDs lists every topic in catalog order.
	TopicIDs() []string
	// LevelsOf lists the grade levels a topic is taught at, and none for a
	// topic the catalog does not have.
	LevelsOf(topic string) []rating.GradeLevel
	// TrapIDs lists every trap in catalog order.
	TrapIDs() []string
	// ExampleTraps lists the traps of the reference tasks of this topic at
	// this level, the most frequent first.
	ExampleTraps(topic string, level rating.GradeLevel) []string
}

// Choice is what the model asked for instead of what the rule suggested. Its
// zero value means the model did not ask for anything.
type Choice struct {
	// Topic, GradeLevel and Difficulty are what to set instead. Any of them
	// may be left out, and the corridor of the topic fills in what is.
	Topic      string
	GradeLevel rating.GradeLevel
	Difficulty int
	// Reason is why, in the model's own words. It is kept in the brief beside
	// the rule's own suggestion.
	Reason string
}

// made reports whether the model asked for anything at all.
func (c *Choice) made() bool { return c.Topic != "" || c.GradeLevel != "" || c.Difficulty != 0 }

// Next builds the brief for this child, and says who chose it.
//
// The goal is always the rule's, even when the model picks the topic: a brief
// that says "reinforce" about a topic the child has never practised records
// that a failure just happened and something else was asked for, which is a
// fact about the child rather than a defect.
func Next(p *profile.Profile, catalog Catalog, choice Choice) (profile.Brief, profile.TutorMode, error) {
	open := withinReach(p, catalog)
	if len(open) == 0 {
		return profile.Brief{}, "", errors.New("tutor: the catalog has no topic taught at any level")
	}

	// The rule's own corridor, of the topic it suggested: what is set unless
	// the model chose otherwise, and the account the rationale keeps whatever
	// was set.
	suggested, goal, because := choose(p, catalog, open)
	ruled := corridorIn(p, catalog, suggested)
	topic, point, mode, err := apply(p, catalog, &choice, suggested, &ruled)
	if err != nil {
		return profile.Brief{}, "", err
	}

	return profile.Brief{
		Constraints: []string{},
		Difficulty:  point.Difficulty,
		// An empty list rather than none: the model hands the brief back as it
		// received it, and a list left out is refused where an empty one says
		// there are no skills to keep out.
		ExcludedSkills:  append([]string{}, p.Student.ExcludedSkills...),
		GradeLevel:      point.GradeLevel,
		PedagogicalGoal: goal,
		Rationale:       rationale(goal, because, &ruled, &choice, point),
		Setting:         setting(p),
		TargetConcept:   topic,
		TrapsToUse:      traps(p, topic, point.GradeLevel, catalog),
	}, mode, nil
}

// withinReach keeps the topics a child can be set now, in catalog order: those
// whose easiest task — difficulty 1 of the lowest level the topic is taught
// at — the child solves often enough to be set it. A child below every one of
// them is given the topics whose easiest task is the easiest the catalog has,
// since nothing easier exists; a topic taught at no level is never given.
func withinReach(p *profile.Profile, catalog Catalog) []string {
	var open, bottom []string
	lowest := math.Inf(1)
	for _, topic := range catalog.TopicIDs() {
		points := pointsOf(catalog, topic)
		if len(points) == 0 {
			continue
		}
		easiest := points[0]
		if rating.InReach(levelIn(p, topic), easiest) {
			open = append(open, topic)
		}
		switch beta := easiest.Beta(); {
		case beta < lowest:
			lowest, bottom = beta, []string{topic}
		case beta == lowest:
			bottom = append(bottom, topic)
		}
	}
	if len(open) > 0 {
		return open
	}
	return bottom
}

// choose picks the topic and the goal, and says in words what the choice
// rested on.
func choose(p *profile.Profile, catalog Catalog, open []string) (topic string, goal profile.Goal, because string) {
	// A failure is worked over again, on the topic it happened in — unless
	// the child is in the trial series, which is finding where they stand and
	// moves to a new topic each time, or the topic has gone out of reach. The
	// window is read rather than trusted: a profile restored from an older
	// revision can arrive with nothing in it. A failure on a topic out of
	// reach is not worked over at all: a task there would be failed again, and
	// the run the rule waited to end could never end.
	now := "no failure now"
	switch {
	case p.Ratings.InTrial():
		now = fmt.Sprintf("trial series, task %d of %d", max(p.Ratings.Answers, 0)+1, rating.TrialAnswers)
		if p.Ratings.ConsecutiveFailures > 0 {
			now += ", which does not go over a failure"
		}
	case p.Ratings.ConsecutiveFailures > 0:
		now = "a failure the window no longer shows"
		if len(p.Recent) > 0 {
			failed := p.Recent[len(p.Recent)-1].Topic
			if slices.Contains(open, failed) {
				return failed, profile.GoalReinforce, failureBehind(p.Ratings.ConsecutiveFailures, failed)
			}
			now = fmt.Sprintf("a failure in %s, which is out of reach now", failed)
		}
	}

	// Otherwise something new: a topic within reach and not mastered at the
	// level its tasks now come from, the ones never given first and then the
	// one unseen for the longest.
	candidates := unmastered(p, catalog, open)
	allMastered := len(candidates) == 0
	if allMastered {
		// Every topic within reach is mastered, so all of them come back into
		// the rotation rather than the child being left with nothing.
		candidates = open
	}

	topic = oldest(p, candidates)
	when := "never given yet"
	if !p.Topics[topic].LastIssued.IsZero() {
		when = "given longest ago"
	}
	if allMastered {
		return topic, profile.GoalNewTopic,
			fmt.Sprintf("%s; every topic within reach is mastered, and %s is the one %s", now, topic, when)
	}
	return topic, profile.GoalNewTopic, fmt.Sprintf("%s; %s is the unmastered topic within reach %s", now, topic, when)
}

// unmastered keeps the topics the child has still to master where they stand,
// in catalog order.
func unmastered(p *profile.Profile, catalog Catalog, open []string) []string {
	kept := make([]string, 0, len(open))
	for _, topic := range open {
		if !masteredWhereItStands(p, catalog, topic) {
			kept = append(kept, topic)
		}
	}
	return kept
}

// masteredWhereItStands reports whether a topic counts as mastered for the
// child now: mastered at the level its recommended point comes from, or at a
// higher one. Mastery at a lower level says nothing about the tasks the child
// would be set there now, so such a topic is back in the rotation.
func masteredWhereItStands(p *profile.Profile, catalog Catalog, topic string) bool {
	summary := p.Topics[topic]
	if summary.MasteredSince == nil || summary.MasteredLevel == nil {
		return false
	}
	recommended := corridorIn(p, catalog, topic).Recommended
	return recommended.GradeLevel.Shift() <= summary.MasteredLevel.Shift()
}

// oldest is the topic that has waited longest: one never given at all, or the
// one given furthest back. Candidates arrive in catalog order and the walk
// keeps the first of any tie, so the order of a map never reaches the answer.
func oldest(p *profile.Profile, candidates []string) string {
	best := candidates[0]
	for _, topic := range candidates[1:] {
		if waitedLonger(p, topic, best) {
			best = topic
		}
	}
	return best
}

func waitedLonger(p *profile.Profile, topic, than string) bool {
	mine, theirs := p.Topics[topic].LastIssued, p.Topics[than].LastIssued
	switch {
	case mine.IsZero() && theirs.IsZero():
		return false // both untouched: catalog order already settled it
	case mine.IsZero():
		return true // never given beats given at any time
	case theirs.IsZero():
		return false
	default:
		return mine.Before(theirs.Time)
	}
}

// apply lets the model's choice override the rule's, and refuses a choice no
// brief could be built from. A topic the catalog does not have, or a level
// the topic is not taught at, would leave the examples, the traps and the
// limits with nothing to stand on. What the model leaves out, the corridor of
// the topic fills in: a topic alone is set at its recommended point, a level
// alone at the recommended difficulty of that level, and a difficulty alone
// at the level of the recommended point.
func apply(p *profile.Profile, catalog Catalog, choice *Choice, suggested string, ruled *rating.Corridor) (
	topic string, point rating.Point, mode profile.TutorMode, err error,
) {
	if !choice.made() {
		return suggested, ruled.Recommended, profile.TutorRule, nil
	}

	topic = suggested
	if choice.Topic != "" {
		if len(catalog.LevelsOf(choice.Topic)) == 0 {
			return "", rating.Point{}, "", fmt.Errorf("tutor: %q is not a topic in the catalog", choice.Topic)
		}
		topic = choice.Topic
	}
	if choice.GradeLevel != "" && !slices.Contains(catalog.LevelsOf(topic), choice.GradeLevel) {
		return "", rating.Point{}, "", fmt.Errorf("tutor: %s is not taught at level %q", topic, choice.GradeLevel)
	}

	// A difficulty of the model's own is used as given: the corridor is a
	// recommendation, and a deliberate step outside it is what a tutor
	// sometimes does.
	if choice.Difficulty != 0 &&
		(choice.Difficulty < profile.MinDifficulty || choice.Difficulty > profile.MaxDifficulty) {
		return "", rating.Point{}, "", fmt.Errorf("tutor: difficulty %d is outside %d to %d",
			choice.Difficulty, profile.MinDifficulty, profile.MaxDifficulty)
	}

	point = ruled.Recommended
	if topic != suggested {
		point = corridorIn(p, catalog, topic).Recommended
	}
	if choice.GradeLevel != "" {
		point = rating.NewCorridor(levelIn(p, topic), rating.Points(choice.GradeLevel)).Recommended
	}
	if choice.Difficulty != 0 {
		point.Difficulty = choice.Difficulty
	}
	return topic, point, profile.TutorLLM, nil
}

// pointsOf are the points of the ladder a topic can be set at: every
// difficulty of every level it is taught at, the easiest first.
func pointsOf(catalog Catalog, topic string) []rating.Point {
	return rating.Points(catalog.LevelsOf(topic)...)
}

// corridorIn is where the child's chances lie in one topic, over the points
// the topic can be set at.
func corridorIn(p *profile.Profile, catalog Catalog, topic string) rating.Corridor {
	return rating.NewCorridor(levelIn(p, topic), pointsOf(catalog, topic))
}

// levelIn is where the child stands in one topic: their level overall with the
// topic's own correction. A topic never met has no correction, so what decides
// is what the child can do generally — a cold start that is right far more
// often than beginning everybody at the middle.
func levelIn(p *profile.Profile, topic string) float64 {
	return p.Ratings.Theta + p.Topics[topic].Delta
}

// setting is what the task is dressed in: the interests taken in turn, by the
// number of answers behind the child. The count never goes backwards, which is
// what the window of recent answers cannot promise — it is pruned, and a
// rotation over it would quietly start repeating.
//
// No interests, no setting: the model picks something itself.
func setting(p *profile.Profile) string {
	interests := p.Student.Interests
	if len(interests) == 0 {
		return ""
	}
	// The count comes from a file a person can open and edit, so it can be
	// anything at all. A remainder of a negative number is negative in Go, and
	// a rule that reached past the front of the list over that would take the
	// request down: nonsense in is a setting out, not a panic.
	turn := p.Ratings.Answers % len(interests)
	if turn < 0 {
		turn += len(interests)
	}
	return interests[turn]
}
