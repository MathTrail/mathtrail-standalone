package profile

import "time"

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
// or the chat's model with a reason of its own. It is kept so that the two can
// be told apart afterwards.
type TutorMode string

const (
	// TutorRule means the model kept what the rule asked for.
	TutorRule TutorMode = "rule"
	// TutorLLM means the model chose otherwise and said why.
	TutorLLM TutorMode = "llm"
)

// Brief is what the model was asked for, exactly as it received it. It is kept
// with the request so that a task can be read against the instructions it was
// written to.
type Brief struct {
	// Constraints are the limits on the wording.
	Constraints []string `json:"constraints"`
	// Difficulty is the level asked for, from MinDifficulty to MaxDifficulty.
	Difficulty int `json:"difficulty"`
	// ExcludedSkills is what may appear neither in the wording nor in a trap.
	ExcludedSkills []string `json:"excluded_skills"`
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
type OpenRequest struct {
	// Attempts is how many times the model has handed something back.
	Attempts int `json:"attempts"`
	// Brief is what it was asked for.
	Brief Brief `json:"brief"`
	// ID must come back with the task. Anything else is a request that has
	// been superseded.
	ID string `json:"id"`
	// Language is the language of the chat, so the task is written in it.
	Language string `json:"language"`
	// OpenedAt starts the window after which an unfinished request counts as
	// abandoned.
	OpenedAt Time `json:"opened_at"`
	// TutorMode is who chose the topic and the difficulty.
	TutorMode TutorMode `json:"tutor_mode"`
}

// CurrentTask is the task the child is working on. Everything in it is open
// except one string: the answer, the explanation behind every wrong option,
// the solution and the solver are sealed together, because four explanations
// in the open would name the four wrong options and hand over the fifth.
type CurrentTask struct {
	// Difficulty is the level of the task, from MinDifficulty to MaxDifficulty.
	Difficulty int `json:"difficulty"`
	// Drawing is the picture in text, when the task has one.
	Drawing string `json:"drawing,omitempty"`
	// Fingerprint is the sketch that joins the list of past tasks once this
	// one is accepted.
	Fingerprint string `json:"fingerprint"`
	// Hint is the nudge the child may ask for.
	Hint string `json:"hint"`
	// ID keys the answer: it is recorded once, against this task.
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
	// Sealed is everything that would give the answer away, as one opaque
	// string. Nothing about it changes shape with the answer: its length says
	// nothing, and it is opened in exactly one place, after the child has
	// answered.
	Sealed string `json:"sealed"`
	// Topic is the catalog id of what is being asked.
	Topic string `json:"topic"`
	// Wording is the question as the child reads it.
	Wording string `json:"wording"`
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

// startDay moves the counters onto the day they are about to count, clearing
// them when the day has moved on. It is called by whatever raises a counter
// rather than left to the caller: a counter raised without the day looked at
// is a limit that never resets.
func (d *Daily) startDay(now time.Time) {
	today := DateOf(now)
	if d.Date.Equal(today.Time) {
		return
	}
	*d = Daily{Date: today}
}
