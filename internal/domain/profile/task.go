package profile

import (
	"encoding/json"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Goal is why this task is being set now. The rule decides it and the model
// never changes it.
type Goal string

const (
	// GoalReinforce works the same ground again after a failure.
	GoalReinforce Goal = "reinforce"
	// GoalNewTopic moves on to a topic the child has not met.
	GoalNewTopic Goal = "new_topic"
)

// TutorMode says who chose the topic and the difficulty of a task: the rule,
// the chat's model with a reason of its own, or the child or the adult, who
// chose the topic of the lessons. It is kept so that they can be told apart
// afterwards.
type TutorMode string

const (
	// TutorRule means the model kept what the rule asked for.
	TutorRule TutorMode = "rule"
	// TutorLLM means the model chose otherwise and said why.
	TutorLLM TutorMode = "llm"
	// TutorPerson means a person chose the task's topic — the child or the
	// adult, who keep the lessons to it — and the rule set the level and the
	// difficulty on it.
	TutorPerson TutorMode = "person"
)

// Known reports whether the mode is one of the three that choose a task: the
// rule, the model or a person. A file edited by hand can name anybody else.
func (m TutorMode) Known() bool { return m == TutorRule || m == TutorLLM || m == TutorPerson }

// Brief is what the model was asked for, exactly as it received it. It is kept
// with the request so that a task can be read against the instructions it was
// written to.
type Brief struct {
	// Constraints are the limits on the wording.
	Constraints []string `json:"constraints"`
	// Difficulty is the difficulty asked for inside the level, from
	// MinDifficulty to MaxDifficulty.
	Difficulty int `json:"difficulty"`
	// ExcludedSkills is what may appear neither in the wording nor in a trap.
	ExcludedSkills []string `json:"excluded_skills"`
	// GradeLevel is the level the task is to be written for: where on the
	// ladder the child stands in this topic, whatever their grade.
	GradeLevel rating.GradeLevel `json:"grade_level"`
	// PedagogicalGoal is why this task now.
	PedagogicalGoal Goal `json:"pedagogical_goal"`
	// Rationale is the rule's own account of the choice, in words.
	Rationale string `json:"rationale"`
	// Setting is what the task is dressed in, taken from the child's
	// interests.
	Setting string `json:"setting"`
	// TargetConcept is the topic, by catalog id.
	TargetConcept string `json:"target_concept"`
	// TrapsToUse are the mistakes the wrong options should lead to.
	TrapsToUse []string `json:"traps_to_use"`
}

// OpenRequest is a task being written right now. It lives in the file rather
// than in an instance's memory because it is what makes "three attempts and no
// more" survive a restart, a second instance and a model that forgets.
//
// Most requests are for the task the child is waiting for. One written ahead
// is for the task after the one on the card: the model writes it while the
// child works, and it is kept until the next task is asked for — unless it is
// asked for while it is still being written, and then it is waited for like
// any other.
type OpenRequest struct {
	// Ahead says the task is written ahead, to be kept until it is asked for,
	// rather than handed to a child who is waiting for it.
	Ahead bool `json:"ahead,omitempty"`
	// Asked says a person asked for the topic, the level or the difficulty
	// the task is set at, rather than the rule choosing them.
	Asked bool `json:"asked,omitempty"`
	// Attempts is how many times the model has handed something back.
	Attempts int `json:"attempts"`
	// Brief is what it was asked for.
	Brief Brief `json:"brief"`
	// ID must come back with the task. Anything else is a request that has
	// been superseded.
	ID string `json:"id"`
	// Language is the lesson's language, which the task is written in: the one
	// the parent chose when the request was opened, or else the chat's.
	Language string `json:"language"`
	// LessonTopic is the topic the lessons were kept to when a request ahead
	// was opened, or none: a task written ahead is handed out only while the
	// lessons are still kept to the same.
	LessonTopic string `json:"lesson_topic,omitempty"`
	// OpenedAt starts the window after which an unfinished request counts as
	// abandoned.
	OpenedAt Time `json:"opened_at"`
	// TutorMode is who chose the topic and the difficulty.
	TutorMode TutorMode `json:"tutor_mode"`
}

// CurrentTask is the task on the child's card. Everything in it is open except
// one string: the answer, the explanation behind every wrong option, the
// solution and the solver are sealed together, because four explanations in
// the open would name the four wrong options and hand over the fifth.
//
// It stays on the card after it is answered, with the answer it was given,
// until the next task is asked for: the same answer sent again is told what
// was recorded, and telling it takes the seal, which leaves with the task.
type CurrentTask struct {
	// Answered is the answer the task was given, or nil while the child is
	// still working on it.
	Answered *Given `json:"answered,omitempty"`
	// Asked says a person asked for the topic, the level or the difficulty
	// the task is set at: the task written ahead after it is set the same, in
	// case one more like it is wanted.
	Asked bool `json:"asked,omitempty"`
	// Difficulty is the difficulty of the task inside its level, from
	// MinDifficulty to MaxDifficulty.
	Difficulty int `json:"difficulty"`
	// Fingerprint is the sketch that joins the list of past tasks once this
	// one is accepted.
	Fingerprint string `json:"fingerprint"`
	// GradeLevel is the level the task was written for. With the difficulty it
	// is where the task stands on the ladder, which the answer is measured
	// against.
	GradeLevel rating.GradeLevel `json:"grade_level"`
	// Hint is the nudge the child may ask for.
	Hint string `json:"hint"`
	// ID keys the answer: it is recorded once, against this task, and the
	// seal is tied to it.
	ID string `json:"id"`
	// InstructionsVersion is which version of the instructions produced the
	// task, so the line logged about the result can carry it.
	InstructionsVersion string `json:"instructions_version"`
	// IssuedAt is when the task was handed out, and the pace is measured from
	// it.
	IssuedAt Time `json:"issued_at"`
	// Language is what the task is written in.
	Language string `json:"language"`
	// Options are the ones the child chooses between, keyed by letter.
	Options map[string]string `json:"options"`
	// Picture is the description of the picture the card draws, as the checks
	// accepted it, its keys in order, and none for a task with none. The
	// profile keeps it and does not read it: a file edited by hand never fails
	// to read for a picture.
	Picture json.RawMessage `json:"picture,omitempty"`
	// Sealed is everything that would give the answer away, as one opaque
	// string. Nothing about it changes shape with the answer: its length says
	// nothing, and it is opened only by an answer to this task.
	Sealed string `json:"sealed"`
	// Topic is the catalog id of what is being asked.
	Topic string `json:"topic"`
	// TutorMode is who chose the task's topic and difficulty, as its request
	// had it: the rule, the model with a reason of its own, or a person, who
	// chose the topic of the lessons. A task handed out before the card kept it
	// has none.
	TutorMode TutorMode `json:"tutor_mode,omitempty"`
	// Wording is the question as the child reads it.
	Wording string `json:"wording"`
}

// InFlight is the task the child is still working on, or nil: none is on the
// card, or the one there has had its answer. Whatever would leave a task
// behind asks this, because only an unanswered task is left behind — an
// answered one already has its answer in the window.
func (p *Profile) InFlight() *CurrentTask {
	if p.CurrentTask == nil || p.CurrentTask.Answered != nil {
		return nil
	}
	return p.CurrentTask
}

// Daily is what the limits count, and the day they are counting. It is in the
// file because every instance of the service has to see the same numbers.
type Daily struct {
	// Accepted is how many tasks were accepted today. Its unit is an accepted
	// task, so a refusal costs the child nothing.
	Accepted int `json:"accepted"`
	// Date is the UTC day these numbers belong to. UTC costs little here: the
	// product shows no rhythm of practice at all, so an early rollover makes
	// the limit looser for one evening and never stricter.
	Date Date `json:"date"`
	// Failed is how many requests ended with the model out of attempts. It
	// exists to stop a loop, not to ration a lesson.
	Failed int `json:"failed"`
}

// CountAccepted records that a task was accepted, which is the unit the daily
// limit is counted in: a refusal costs the child nothing, so only a task that
// reached them does.
func (p *Profile) CountAccepted(now time.Time) {
	p.Daily.startDay(now)
	p.Daily.Accepted++
}

// CountFailed records that a request ended with the model out of attempts. It
// is counted apart from the tasks because it exists to stop a loop rather than
// to ration a lesson.
func (p *Profile) CountFailed(now time.Time) {
	p.Daily.startDay(now)
	p.Daily.Failed++
}

// Today is what the counters hold for the day now falls on: the counters as
// they stand when they are that day's, and nothing counted when they are
// another day's. A limit reads the counters through it, so that yesterday's
// tasks never hold back today's.
func (d *Daily) Today(now time.Time) Daily {
	today := DateOf(now)
	if d.Date.Equal(today.Time) {
		return *d
	}
	return Daily{Date: today}
}

// startDay moves the counters onto the day they are about to count, clearing
// them when the day has moved on. It is called by whatever raises a counter
// rather than left to the caller: a counter raised without the day looked at
// is a limit that never resets.
func (d *Daily) startDay(now time.Time) {
	*d = d.Today(now)
}
