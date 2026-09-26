// Package tutor decides what the child is asked next.
//
// It reads one profile and writes one brief — the terms of reference the
// chat's model writes a task to: what the task is about, how hard it should
// be, what it is dressed in, which mistakes its wrong options should lead to,
// and what may not appear in it at all.
//
// Nothing here is random and nothing here is a model: the same profile always
// produces the same brief. That is what makes the rule a baseline the model
// can be measured against, and it is why this package needs no mocks to be
// tested — a profile in, a brief out.
//
// The rule reads the summary of each topic and the short window of recent
// answers, never a full history, because a full history is not kept. It never
// reads the parent's notes about the child: those travel to the model as tone
// and are not allowed to change what the child is set.
package tutor

import (
	"fmt"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Traps is how many mistakes a brief names for the wrong options to lead to.
const Traps = 2

// Catalog is what the rule has to know about the content the service ships.
// It is declared here, by the side that needs it, and it speaks in
// identifiers: which topics exist for a child of this grade, how two traps are
// told apart when nothing else separates them, and what usually goes wrong in
// a topic.
//
// Grades cross this line and levels do not. The rule knows what school year
// the child is in; what years are grouped into is the catalog's business.
type Catalog interface {
	// TopicsAt lists the topics a child of this grade can be given, in
	// catalog order.
	TopicsAt(grade int) []string
	// TrapIDs lists every trap in catalog order.
	TrapIDs() []string
	// ExampleTraps lists the traps of the reference tasks of this topic and
	// grade, the most frequent first.
	ExampleTraps(topic string, grade int) []string
}

// Choice is what the model asked for instead of what the rule suggested. Its
// zero value means the model did not ask for anything.
type Choice struct {
	// Topic and Difficulty are what to set instead. Either may be left out.
	Topic      string
	Difficulty int
	// Reason is why, in the model's own words. It is kept in the brief beside
	// the rule's own suggestion.
	Reason string
}

// made reports whether the model asked for anything at all.
func (c Choice) made() bool { return c.Topic != "" || c.Difficulty != 0 }

// Next builds the brief for this child, and says who chose it.
//
// The goal is always the rule's, even when the model picks the topic: a brief
// that says "reinforce" about a topic the child has never practised records
// that a failure just happened and something else was asked for, which is a
// fact about the child rather than a defect.
func Next(p *profile.Profile, catalog Catalog, choice Choice) (profile.Brief, profile.TutorMode, error) {
	available := catalog.TopicsAt(p.Student.Grade)
	if len(available) == 0 {
		return profile.Brief{}, "", fmt.Errorf("tutor: grade %d has no topics", p.Student.Grade)
	}

	suggested, goal, because := choose(p, available)
	topic, difficulty, mode, err := apply(choice, suggested, available)
	if err != nil {
		return profile.Brief{}, "", err
	}

	// The difficulty is the corridor's of the topic set; the rationale keeps
	// the rule's own account, of the topic it suggested, whatever was set.
	if difficulty == 0 {
		difficulty = rating.NewCorridor(levelIn(p, topic)).Recommended
	}
	ruled := rating.NewCorridor(levelIn(p, suggested))

	return profile.Brief{
		Constraints: []string{},
		Difficulty:  difficulty,
		// An empty list rather than none: the model hands the brief back as it
		// received it, and a list left out is refused where an empty one says
		// there are no skills to keep out.
		ExcludedSkills:  append([]string{}, p.Student.ExcludedSkills...),
		PedagogicalGoal: goal,
		Rationale:       rationale(goal, because, &ruled, choice, difficulty),
		Setting:         setting(p),
		TargetConcept:   topic,
		TrapsToUse:      traps(p, topic, catalog),
	}, mode, nil
}

// choose picks the topic and the goal, and says in words what the choice
// rested on.
func choose(p *profile.Profile, available []string) (topic string, goal profile.Goal, because string) {
	// A failure is worked over again, on the topic it happened in — while that
	// topic is still one of the child's grade. The window is read rather than
	// trusted: a profile restored from an older revision can arrive with
	// nothing in it, and a grade changed since, or a topic withdrawn from the
	// catalog, leaves a failure on a topic no task can now be written for. A
	// rule that kept asking for one would refuse every task the model wrote,
	// and the run it waited to end could never end.
	now := "no failure now"
	if p.Ratings.ConsecutiveFailures > 0 {
		now = "a failure the window no longer shows"
		if len(p.Recent) > 0 {
			failed := p.Recent[len(p.Recent)-1].Topic
			if slices.Contains(available, failed) {
				return failed, profile.GoalReinforce, failureBehind(p.Ratings.ConsecutiveFailures, failed)
			}
			now = fmt.Sprintf("a failure in %s, which is not a topic of this grade", failed)
		}
	}

	// Otherwise something new: a topic of this grade that is not mastered,
	// the ones never given first and then the one unseen for the longest.
	candidates := unmastered(p, available)
	allMastered := len(candidates) == 0
	if allMastered {
		// Every topic of the grade is mastered, so all of them come back into
		// the rotation rather than the child being left with nothing.
		candidates = available
	}

	topic = oldest(p, candidates)
	when := "never given yet"
	if !p.Topics[topic].LastIssued.IsZero() {
		when = "given longest ago"
	}
	if allMastered {
		return topic, profile.GoalNewTopic,
			fmt.Sprintf("%s; every topic of this grade is mastered, and %s is the one %s", now, topic, when)
	}
	return topic, profile.GoalNewTopic, fmt.Sprintf("%s; %s is the unmastered topic %s", now, topic, when)
}

// unmastered keeps the topics the child has still to master, in catalog order.
func unmastered(p *profile.Profile, available []string) []string {
	kept := make([]string, 0, len(available))
	for _, topic := range available {
		if p.Topics[topic].MasteredSince == nil {
			kept = append(kept, topic)
		}
	}
	return kept
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
// brief could be built from. A topic outside the catalog would leave the
// examples, the traps and the corridor with nothing to stand on.
func apply(choice Choice, suggested string, available []string) (topic string, difficulty int, mode profile.TutorMode, err error) {
	if !choice.made() {
		return suggested, 0, profile.TutorRule, nil
	}

	topic = suggested
	if choice.Topic != "" {
		if !slices.Contains(available, choice.Topic) {
			return "", 0, "", fmt.Errorf("tutor: %q is not a topic of this grade", choice.Topic)
		}
		topic = choice.Topic
	}

	// A difficulty of the model's own is used as given: the corridor is a
	// recommendation, and a deliberate step outside it is what a tutor
	// sometimes does.
	if choice.Difficulty != 0 &&
		(choice.Difficulty < profile.MinDifficulty || choice.Difficulty > profile.MaxDifficulty) {
		return "", 0, "", fmt.Errorf("tutor: difficulty %d is outside %d to %d",
			choice.Difficulty, profile.MinDifficulty, profile.MaxDifficulty)
	}
	return topic, choice.Difficulty, profile.TutorLLM, nil
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
