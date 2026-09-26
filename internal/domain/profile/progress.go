package profile

import "github.com/MathTrail/mathtrail-standalone/internal/domain/rating"

// Ratings is where the child stands across every topic.
type Ratings struct {
	// Answers is how many answers the level rests on. It narrows the step
	// every new answer may move the level by, and it is also what the setting
	// rotation counts: a number that never goes backwards, unlike the length
	// of a window that is pruned. While it is below the length of the trial
	// series, the child is in it.
	Answers int `json:"answers"`
	// ConsecutiveFailures is how many wrong answers came in a row. It is the
	// switch between going over the same ground again and moving on. A task
	// the child skipped leaves it alone: there was no answer to learn from.
	ConsecutiveFailures int `json:"consecutive_failures"`
	// Start is where the child began on the ladder: the shift of the level of
	// the grade the profile was made with. It is written once and never moves
	// — a grade changed later is a label — and the trial series estimates the
	// level from it.
	Start float64 `json:"start"`
	// Theta is the level itself: a place on the one ladder of every grade.
	Theta float64 `json:"theta"`
}

// InTrial reports whether the child is still in the trial series, where the
// level is estimated from all the answers so far at once rather than moved by
// one answer at a time. The next answer is then one of the series.
func (r Ratings) InTrial() bool { return r.Answers < rating.TrialAnswers }

// Topic is the summary of one topic the child has met. It is what lets the
// history window stay short: everything the rule needs about the past is
// added up here, so the answers themselves can be dropped.
type Topic struct {
	// Answers and Correct are how many answers this topic rests on and how
	// many of them were right.
	Answers int `json:"answers"`
	Correct int `json:"correct"`
	// Delta corrects the overall level for this topic — a child may count
	// well and reason about parity badly.
	Delta float64 `json:"delta"`
	// LastIssued is the day a task of this topic was last accepted, and is
	// absent for a topic never issued, which the rule puts first. It moves
	// when a task is accepted rather than when the topic is chosen: a
	// generation that produced nothing must not push its topic away.
	LastIssued Date `json:"last_issued,omitzero"`
	// MasteredLevel is the level of the task whose answer earned mastery, set
	// and cleared together with MasteredSince. A topic counts as mastered only
	// while its tasks come from this level or below: mastered among the tasks
	// of grades 1–2 says nothing about those of grades 3–4.
	MasteredLevel *rating.GradeLevel `json:"mastered_level"`
	// MasteredSince is the day the topic was mastered, or null. A mastered
	// topic steps out of the rotation until every topic within the child's
	// reach has been mastered.
	MasteredSince *Date `json:"mastered_since"`
	// TopStreak is the run of correct answers at the harder half of the
	// corridor with no hint — the run that earns mastery.
	TopStreak int `json:"top_streak"`
	// Traps is how often this child fell for each trap in this topic. The
	// rule picks the two most frequent, and the progress screen draws the map
	// of misconceptions from the same numbers.
	Traps map[string]int `json:"traps"`
	// WrongStreak is the run of wrong answers in this topic. It is what mastery
	// is lost by, and it is counted rather than read back out of the history
	// window for the same reason TopStreak is: the window is pruned, and a run
	// that has to be exact cannot be read from something that forgets.
	WrongStreak int `json:"wrong_streak"`
}

// Pace is how long the child took, measured by the service from the moment the
// task was handed out — never by the clock of whoever answered.
type Pace string

const (
	// PaceFast is an answer inside a minute.
	PaceFast Pace = "fast"
	// PaceNormal is anything between the two.
	PaceNormal Pace = "normal"
	// PaceSlow is an answer after three minutes.
	PaceSlow Pace = "slow"
)

// Answer is one entry of the history window.
type Answer struct {
	// AnsweredAt is when the answer arrived.
	AnsweredAt Time `json:"answered_at"`
	// Chosen is the option the child picked, and Trap the trap behind it.
	// Both are written only for a wrong answer: on a right one there is no
	// mistake to name.
	Chosen string `json:"chosen,omitempty"`
	// Confused records that "I don't understand" was pressed before
	// answering. Like HintUsed it changes what comes next and never the
	// rating.
	Confused bool `json:"confused"`
	// Correct is the only thing the rating formula reads. It is never absent:
	// "I don't understand" is a button that asks for a simpler explanation,
	// not a third kind of answer.
	Correct bool `json:"correct"`
	// Difficulty is the difficulty of the task that was answered, inside its
	// level.
	Difficulty int `json:"difficulty"`
	// GradeLevel is the level the task was written for. With the difficulty
	// it is where the task stood on the ladder, which the trial series reads
	// back to weigh every one of its answers again.
	GradeLevel rating.GradeLevel `json:"grade_level"`
	// HintUsed records that the hint was opened.
	HintUsed bool `json:"hint_used"`
	// Pace is how long it took.
	Pace Pace `json:"pace"`
	// TaskID ties this entry to the fingerprint of the task and to the line
	// about it in the log.
	TaskID string `json:"task_id"`
	// Topic is the catalog id of what was asked.
	Topic string `json:"topic"`
	// Trap is the trap the wrong option led to.
	Trap string `json:"trap,omitempty"`
}

// point is where the answered task stood on the ladder.
func (a *Answer) point() rating.Point {
	return rating.Point{GradeLevel: a.GradeLevel, Difficulty: a.Difficulty}
}
