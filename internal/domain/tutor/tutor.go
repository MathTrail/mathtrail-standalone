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
// the child's answers say. One thing a person says does steer it: a topic the
// child or the adult chose for the lessons, which every task is set on once
// the trial series is over, at the level and the difficulty the answers say.
package tutor

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Traps is how many mistakes a brief names for the wrong options to lead to.
const Traps = 2

// InterestEvery is how often a task is dressed in one of the child's
// interests: one task in this many, the first of each run of them, while the
// others are dressed in a setting the model chooses. Met in every task, an
// interest wears thin, and the tasks blur into one another.
const InterestEvery = 3

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

// MaxReason is how long the model's reason for a choice of its own may be, in
// characters: a sentence or two. It travels in the brief, and with the brief
// in every package of the request.
const MaxReason = 300

// Choice is what the model asked for instead of what the rule suggested. Its
// zero value means the model did not ask for anything.
type Choice struct {
	// Topic, GradeLevel and Difficulty are what to set instead. Any of them
	// may be left out, and the corridor of the topic fills in what is.
	Topic      string
	GradeLevel rating.GradeLevel
	Difficulty int
	// Reason is why, in the model's own words. A choice of any of the three
	// needs one, and it is kept in the brief beside the rule's own suggestion.
	// A reason alone chooses nothing.
	Reason string
}

// Made reports whether the model asked for anything at all: a topic, a level
// or a difficulty of its own. A reason alone asks for nothing.
func (c *Choice) Made() bool { return c.Topic != "" || c.GradeLevel != "" || c.Difficulty != 0 }

// Beside is the choice as it stands beside the topic the lessons are kept to,
// lesson: naming that topic asks for nothing, since it is set anyway, and a
// reason left with nothing beside it explains nothing.
func (c *Choice) Beside(lesson string) Choice {
	beside := *c
	if lesson != "" && beside.Topic == lesson {
		beside.Topic = ""
		if !beside.Made() {
			beside.Reason = ""
		}
	}
	return beside
}

// ChoiceError is a choice of the model's that no brief can be built from, with
// every rule it breaks. A caller finds it with errors.As, to tell the model
// what to change.
type ChoiceError struct {
	// Problems name the part of the choice and the rule it broke. They never
	// repeat what the model wrote.
	Problems []profile.Problem
}

// Error lists every rule the choice breaks.
func (e *ChoiceError) Error() string {
	broken := make([]string, 0, len(e.Problems))
	for _, problem := range e.Problems {
		broken = append(broken, problem.String())
	}
	return "tutor: the model's choice breaks a rule: " + strings.Join(broken, "; ")
}

// problems are the rules a choice breaks. A level named alone is set on the
// topic the lessons are kept to, when the child or the adult chose one, and on
// the topic the rule suggested otherwise. A topic has to be one of the
// catalog, and none other than the one chosen for the lessons, and a level one
// it is taught at — the examples, the traps and the limits of a brief stand on
// both — and a difficulty one of a level. And a choice of any of them says
// why.
func (c *Choice) problems(catalog Catalog, suggested, chosen string) []profile.Problem {
	var found []profile.Problem
	topic := suggested
	if chosen != "" {
		topic = chosen
	}
	if c.Topic != "" {
		switch {
		case len(catalog.LevelsOf(c.Topic)) == 0:
			topic = c.Topic
			found = append(found, profile.Problem{Field: "topic", Broken: profile.Broken{
				Code: profile.CodeNotInCatalog, Rule: "must be a topic of the catalog, by its id"}})
		case chosen != "":
			// The chosen topic is what will be set, so a level is held to it.
			found = append(found, profile.Problem{Field: "topic", Broken: profile.Broken{
				Code: profile.CodeNotOneOf, Rule: fmt.Sprintf("must be %s, the topic the child or the adult chose for "+
					"the lessons, or left out; save_profile changes the choice when they ask", chosen)}})
		default:
			topic = c.Topic
		}
	}
	if broken := c.levelRule(catalog, topic, suggested, chosen); broken.Code != "" {
		found = append(found, profile.Problem{Field: "grade_level", Broken: broken})
	}
	if c.Difficulty != 0 && (c.Difficulty < profile.MinDifficulty || c.Difficulty > profile.MaxDifficulty) {
		found = append(found, profile.Problem{Field: "difficulty", Broken: profile.Broken{
			Code: profile.CodeOutOfRange, Rule: fmt.Sprintf("must be from %d to %d", profile.MinDifficulty, profile.MaxDifficulty)}})
	}
	if broken := c.reasonRule(); broken.Code != "" {
		found = append(found, profile.Problem{Field: "reason", Broken: broken})
	}
	return found
}

// levelRule is the rule a level of the model's breaks, or nothing. A level is
// held to the topic it will be set on: the model's own; the one chosen for the
// lessons, which no level of the model's moves it off; or the rule's when the
// model named none, which the rule names back so that the model can pick a
// topic that fits instead.
func (c *Choice) levelRule(catalog Catalog, topic, suggested, chosen string) profile.Broken {
	levels := catalog.LevelsOf(topic)
	switch {
	case c.GradeLevel == "":
		return profile.Broken{}
	case !c.GradeLevel.Known():
		return profile.Broken{Code: profile.CodeNotOneOf, Rule: fmt.Sprintf("must be one of %s", joined(rating.GradeLevels()))}
	case len(levels) == 0 || slices.Contains(levels, c.GradeLevel):
		// An unknown topic is refused as a topic; its levels are nobody's.
		return profile.Broken{}
	case chosen != "" && topic == chosen:
		return profile.Broken{Code: profile.CodeNotTaught, Rule: fmt.Sprintf(
			"must be a level the topic chosen for the lessons, %s, is taught at: %s", chosen, joined(levels))}
	case c.Topic == "":
		return profile.Broken{Code: profile.CodeNotTaught, Rule: fmt.Sprintf(
			"must be a level the rule's topic, %s, is taught at: %s; or name a topic taught at this level",
			suggested, joined(levels))}
	}
	return profile.Broken{Code: profile.CodeNotTaught, Rule: fmt.Sprintf("must be a level the topic is taught at: %s", joined(levels))}
}

// reasonRule is the rule the model's reason breaks, or nothing. A reason with
// no choice beside it breaks none: it chooses nothing, and it is not kept. The
// reason is counted as it will be kept — cleaned as any text typed into the
// file is — so that nothing unseen pads it past its limit or reorders it where
// it is read.
func (c *Choice) reasonRule() profile.Broken {
	if !c.Made() {
		return profile.Broken{}
	}
	switch count := utf8.RuneCountInString(profile.Typed(c.Reason)); {
	case count == 0:
		return profile.Broken{Code: profile.CodeRequired, Rule: "is required with a topic, a level or a difficulty of " +
			"your own: say in a sentence why the child needs it now"}
	case count > MaxReason:
		return profile.Broken{Code: profile.CodeTooLong, Rule: fmt.Sprintf("must be at most %d characters, not %d", MaxReason, count)}
	}
	return profile.Broken{}
}

// joined lists levels the way a sentence does.
func joined(levels []rating.GradeLevel) string {
	names := make([]string, 0, len(levels))
	for _, level := range levels {
		names = append(names, string(level))
	}
	return strings.Join(names, ", ")
}

// Next builds the brief for this child, and says who chose it. A choice of the
// model's that no brief can be built from is refused with a ChoiceError.
//
// Once the trial series is over, a topic the child or the adult chose for the
// lessons is set in place of the rule's. The model may still name a level or
// a difficulty with a reason, and asking for that same topic asks for nothing;
// another topic it is refused, since the choice is theirs to change.
//
// The goal is always the rule's, even when the model or the person picks the
// topic: a brief that says "reinforce" about a topic the child has never
// practised records that a failure just happened and something else was asked
// for, which is a fact about the child rather than a defect.
//
// While a task on the card still waits for its answer, the brief is for the
// task asked for in its place, which leaves that one behind: it is dressed as
// the task after it.
func Next(p *profile.Profile, catalog Catalog, choice Choice) (profile.Brief, profile.TutorMode, error) {
	return brief(p, catalog, choice, false)
}

// Ahead builds the brief for the task after the one on the card, which the
// model writes ahead, and says who chose it. It is the brief Next builds now.
// While the child still works on the task on the card it is chosen before
// that answer, which the next brief ahead then takes in, and dressed as the
// task after that one: which task an interest dresses is counted by the tasks
// behind the child, and the task on the card is not behind them yet, so two
// tasks in a row would otherwise be dressed alike. Once that task has its
// answer, the task written ahead is simply the next one.
func Ahead(p *profile.Profile, catalog Catalog, choice Choice) (profile.Brief, profile.TutorMode, error) {
	return brief(p, catalog, choice, true)
}

// brief is the brief of Next, or of Ahead when ahead is set.
func brief(p *profile.Profile, catalog Catalog, choice Choice, ahead bool) (profile.Brief, profile.TutorMode, error) {
	open := WithinReach(p, catalog)
	if len(open) == 0 {
		return profile.Brief{}, "", errors.New("tutor: the catalog has no topic taught at any level")
	}

	awaited := awaits(p, ahead)
	suggested, goal, because := choose(p, catalog, open, awaited)
	chosen := LessonTopic(p, catalog)
	choice = choice.Beside(chosen)
	if problems := choice.problems(catalog, suggested, chosen); len(problems) > 0 {
		return profile.Brief{}, "", &ChoiceError{Problems: problems}
	}
	// The rule's own corridor, of the topic it suggested: what is set unless
	// the model or the person chose otherwise, and the account the rationale
	// keeps whatever was set.
	ruled := CorridorIn(p, catalog, suggested)
	topic, point, mode := apply(p, catalog, &choice, suggested, chosen, &ruled)

	said := rationale(goal, because, &ruled, &choice, chosen, point)
	if ahead {
		said = writtenAhead(awaited) + said
	}
	return profile.Brief{
		Constraints: []string{},
		Difficulty:  point.Difficulty,
		// An empty list rather than none, so that the brief in the package says
		// there are no skills to keep out rather than leaving a null to guess at.
		ExcludedSkills:  append([]string{}, p.Student.ExcludedSkills...),
		GradeLevel:      point.GradeLevel,
		PedagogicalGoal: goal,
		Rationale:       said,
		Setting:         setting(p),
		TargetConcept:   topic,
		TrapsToUse:      traps(p, topic, point.GradeLevel, catalog),
	}, mode, nil
}

// awaits says a task is written ahead of the answer to the task on the card.
// That task comes before it, whatever answer it gets, so whatever counts what
// the child has left behind counts the task on the card among it already: its
// place in the trial series, and the idea and the reference tasks of its
// package are what they will be once that answer is in.
func awaits(p *profile.Profile, ahead bool) bool {
	return ahead && p.InFlight() != nil
}

// Behind is how many answers the child has given and how many tasks of a
// topic they have left behind, answered or skipped, as a task of that topic
// written now counts them: with the task on the card among them when the task
// is written ahead of its answer. A package built from them names a task
// written ahead the idea after the one the task on the card was built on, when
// the two share a topic, and the same idea and reference tasks before that
// answer as after it. A skip of the task on the card moves no answer, so after
// one the reference tasks are those of the answers before it, while the idea
// stays: a task skipped is among the tasks of its topic.
func Behind(p *profile.Profile, topic string, ahead bool) (answers, topicTasks int) {
	stats := p.Topics[topic]
	answers, topicTasks = p.Ratings.Answers, stats.Answers+stats.Skipped
	if awaits(p, ahead) {
		answers++
		if p.InFlight().Topic == topic {
			topicTasks++
		}
	}
	return answers, topicTasks
}

// writtenAhead is what the rationale of a task written ahead opens with: that
// it was written before the answer to the task on the card, while that answer
// is awaited, or else before it was asked for.
func writtenAhead(awaited bool) string {
	if awaited {
		return "Written ahead, before the answer to the task on the card. "
	}
	return "Written ahead, before it is asked for. "
}

// ChosenTopic is the topic the child or the adult chose to keep the lessons
// to, while the catalog has it, or none. The file is the parent's to edit, so
// a topic the catalog does not have, or has since lost, counts as no choice.
func ChosenTopic(p *profile.Profile, catalog Catalog) string {
	topic := p.Student.LessonTopic
	if topic == "" || len(catalog.LevelsOf(topic)) == 0 {
		return ""
	}
	return topic
}

// LessonTopic is the topic the lessons are kept to now: the one chosen, once
// the trial series is over, or none. The series finds where the child stands
// by moving to a new topic each time, so a choice made during it waits for its
// end.
func LessonTopic(p *profile.Profile, catalog Catalog) string {
	if p.Ratings.InTrial() {
		return ""
	}
	return ChosenTopic(p, catalog)
}

// WithinReach keeps the topics a child can be set now, in catalog order: those
// whose easiest task — difficulty 1 of the lowest level the topic is taught
// at — the child solves often enough to be set it. A child below every one of
// them is given the topics whose easiest task is the easiest the catalog has,
// since nothing easier exists; a topic taught at no level is never given.
func WithinReach(p *profile.Profile, catalog Catalog) []string {
	var open, bottom []string
	lowest := math.Inf(1)
	for _, topic := range catalog.TopicIDs() {
		points := pointsOf(catalog, topic)
		if len(points) == 0 {
			continue
		}
		easiest := points[0]
		if rating.InReach(p.LevelIn(topic), easiest) {
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
// rested on: for the next task, or for the one after the task on the card when
// its answer is awaited.
func choose(p *profile.Profile, catalog Catalog, open []string, awaited bool) (topic string, goal profile.Goal, because string) {
	// A failure is worked over again, on the topic of the last answer — unless
	// the child is in the trial series, which is finding where they stand and
	// moves to a new topic each time, or the topic has gone out of reach. A
	// task skipped since is no answer and does not move the failure. The
	// window is read rather than trusted: a profile restored from an older
	// revision can arrive with no answer in it. A failure on a topic out of
	// reach is not worked over at all: a task there would be failed again, and
	// the run the rule waited to end could never end.
	now := "no failure now"
	switch {
	case p.Ratings.InTrial():
		now = trialText(p.Ratings.Answers, awaited)
		if p.Ratings.ConsecutiveFailures > 0 {
			now += ", which does not go over a failure"
		}
	case p.Ratings.ConsecutiveFailures > 0:
		now = "a failure the window no longer shows"
		if last, answered := p.LastAnswer(); answered {
			if slices.Contains(open, last.Topic) {
				return last.Topic, profile.GoalReinforce, failureBehind(p.Ratings.ConsecutiveFailures, last.Topic)
			}
			now = fmt.Sprintf("a failure in %s, which is out of reach now", last.Topic)
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

// trialText says which task of the trial series is being chosen, answered
// being how many the child has answered: the next one, or the one after the
// task on the card when its answer is awaited — which the series may not
// reach, its last task being the one on the card.
func trialText(answered int, awaited bool) string {
	task := max(answered, 0) + 1
	if awaited {
		task++
	}
	if task > rating.TrialAnswers {
		return "the last task of the trial series is on the card"
	}
	return fmt.Sprintf("trial series, task %d of %d", task, rating.TrialAnswers)
}

// unmastered keeps the topics the child has still to master where they stand,
// in catalog order.
func unmastered(p *profile.Profile, catalog Catalog, open []string) []string {
	kept := make([]string, 0, len(open))
	for _, topic := range open {
		if !Mastered(p, catalog, topic) {
			kept = append(kept, topic)
		}
	}
	return kept
}

// Mastered reports whether a topic counts as mastered for the child now:
// mastered at the level its recommended point comes from, or at a higher one.
// Mastery at a lower level says nothing about the tasks the child would be set
// there now, so such a topic is back in the rotation — and a screen that said
// "mastered" about it would be saying something the rule no longer believes.
// A file whose masteries the earlier rule declared counts none of them.
func Mastered(p *profile.Profile, catalog Catalog, topic string) bool {
	summary := p.Topics[topic]
	if !p.MasteriesStand() || summary.MasteredSince == nil || summary.MasteredLevel == nil {
		return false
	}
	recommended := CorridorIn(p, catalog, topic).Recommended
	return recommended.GradeLevel.Shift() <= summary.MasteredLevel.Shift()
}

// MasteredTopics is how many topics of the catalog count as mastered for the
// child now, by Mastered: the topics the progress calls mastered, counted. A
// topic mastered below the level its tasks come from now is not among them.
func MasteredTopics(p *profile.Profile, catalog Catalog) int {
	count := 0
	for _, topic := range catalog.TopicIDs() {
		if Mastered(p, catalog, topic) {
			count++
		}
	}
	return count
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

// apply lets the topic chosen for the lessons, and then the model's choice,
// override the rule's. The choice has been held to its rules already, so it
// names nothing a brief cannot be built from, and no topic other than the one
// chosen. What the model leaves out, the corridor of the topic fills in: a
// topic alone — the model's or the person's — is set at its recommended point,
// a level alone at the recommended difficulty of that level, and a difficulty
// alone at the level of the recommended point.
//
// A difficulty of the model's own is used as given: the corridor is a
// recommendation, and a deliberate step outside it is what a tutor sometimes
// does.
func apply(p *profile.Profile, catalog Catalog, choice *Choice, suggested, chosen string, ruled *rating.Corridor) (
	topic string, point rating.Point, mode profile.TutorMode,
) {
	topic, point, mode = suggested, ruled.Recommended, profile.TutorRule
	if chosen != "" {
		topic, point, mode = chosen, CorridorIn(p, catalog, chosen).Recommended, profile.TutorPerson
	}
	if !choice.Made() {
		return topic, point, mode
	}

	if choice.Topic != "" {
		topic = choice.Topic
		point = CorridorIn(p, catalog, topic).Recommended
	}
	if choice.GradeLevel != "" {
		point = rating.NewCorridor(p.LevelIn(topic), rating.Points(choice.GradeLevel)).Recommended
	}
	if choice.Difficulty != 0 {
		point.Difficulty = choice.Difficulty
	}
	return topic, point, profile.TutorLLM
}

// pointsOf are the points of the ladder a topic can be set at: every
// difficulty of every level it is taught at, the easiest first.
func pointsOf(catalog Catalog, topic string) []rating.Point {
	return rating.Points(catalog.LevelsOf(topic)...)
}

// CorridorIn is where the child's chances lie in one topic, over the points
// the topic can be set at: what a brief on it recommends, and what the model
// is shown of the chances around that point.
func CorridorIn(p *profile.Profile, catalog Catalog, topic string) rating.Corridor {
	return rating.NewCorridor(p.LevelIn(topic), pointsOf(catalog, topic))
}

// setting is what the task is dressed in: one of the child's interests on the
// first of every InterestEvery tasks, the interests taken in turn, and nothing
// on the others, which the model dresses in a setting of its own. The tasks
// are counted as the child leaves them behind, answered or skipped, so that a
// task leafed past for its interest is not followed by another dressed in
// one. The count never goes backwards, which is what the window of recent
// answers cannot promise — it is pruned, and a rotation over it would quietly
// start repeating.
//
// No interests, no setting: the model picks something itself.
func setting(p *profile.Profile) string {
	interests := p.Student.Interests
	if len(interests) == 0 {
		return ""
	}
	tasks := tasksBehind(p)
	if tasks%InterestEvery != 0 {
		return ""
	}
	return interests[(tasks/InterestEvery)%len(interests)]
}

// tasksBehind is how many tasks the child has left behind, answered or
// skipped, with the task on the card among them while it waits for its
// answer: a brief built then is for a task after it, whether written ahead of
// that answer or asked for in its place, which leaves it behind, skipped — and
// a brief is built before that skip is recorded.
//
// The counts come from a file a person can open and edit, so they can be
// anything at all. A remainder of a negative number is negative in Go, and a
// rule that reached past the front of the list over that would take the
// request down: a count below zero is taken as none, and nonsense in is a
// setting out, not a panic.
func tasksBehind(p *profile.Profile) int {
	tasks := p.Ratings.Answers
	for _, topic := range p.Topics {
		tasks += topic.Skipped
	}
	if p.InFlight() != nil {
		tasks++
	}
	return max(tasks, 0)
}
