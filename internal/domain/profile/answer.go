package profile

import (
	"errors"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The refusals of Record. An answer that arrives for a task nobody is working
// on is not a fault of the service, and the two cases read differently to
// whoever hears about them: one child answered twice, the other answered
// something that has already been replaced.
var (
	// ErrNoTask means there is no task in flight to answer.
	ErrNoTask = errors.New("profile: no task in flight")
	// ErrNoSuchOption means the answer names an option the task does not have.
	ErrNoSuchOption = errors.New("profile: the answer names no option of the task")
	// ErrOtherTask means the answer is for a task other than the one in
	// flight — an answer sent twice, or one that arrived after the child
	// moved on.
	ErrOtherTask = errors.New("profile: the answer is for another task")
)

// How long an answer may take before it is called something other than
// normal. The service measures this itself, from the moment it handed the task
// out: a clock on the other side of a chat is not one this can trust.
const (
	fastAnswer = time.Minute
	slowAnswer = 3 * time.Minute
)

// What earns mastery of a topic, and what loses it.
const (
	// MasteryAnswers is how many answers a topic needs behind it before it can
	// be mastered at all: it keeps a lucky start from mastering a topic on the
	// third task.
	MasteryAnswers = 5
	// MasteryStreak is the run of correct answers that earns it — short enough
	// to be reachable, long enough to rule out guessing.
	MasteryStreak = 3
	// MasteryLostAfter is the run of wrong answers that loses it. One wrong
	// answer is a slip and undoes nothing.
	MasteryLostAfter = 2
)

// Answered is one answer as it reaches the service.
type Answered struct {
	// TaskID says what is being answered. It has to be the task in flight.
	TaskID string
	// Chosen is the option the child picked.
	Chosen string
	// Correct says whether that was the right one. It is the only thing the
	// rating formula reads.
	Correct bool
	// Trap is the mistake the chosen option leads to, known only for a wrong
	// answer and taken from the sealed part of the task.
	Trap string
	// HintUsed and Confused are what the child reached for before answering.
	// They change what comes next and never the rating.
	HintUsed bool
	Confused bool
	// At is when the answer arrived.
	At time.Time
}

// Recorded is what one answer changed, for whoever has to say so: the result
// the child is shown and the line written to the log.
type Recorded struct {
	// Topic is what was answered, GradeLevel the level the task was written
	// for and Difficulty its difficulty inside that level.
	Topic      string
	GradeLevel rating.GradeLevel
	Difficulty int
	// Probability is the chance of a correct answer as it stood when the task
	// was handed out.
	Probability float64
	// LevelBefore and LevelAfter are where the child stood in this topic
	// before the answer and after it.
	LevelBefore float64
	LevelAfter  float64
	// Trial is which answer of the trial series this was, from 1, or zero when
	// the series was already over. While it runs there is no rating to show,
	// only how many of its answers are in.
	Trial int
	// Pace is how long the answer took.
	Pace Pace
	// Mastered is set when this answer earned mastery of the topic, and
	// Unmastered when it lost it. Both cannot be true of one answer.
	Mastered   bool
	Unmastered bool
}

// Record applies an answer to the profile: it moves the level, adds to the
// summary of the topic, puts the answer at the end of the history window, and
// takes the task out of flight. The task itself is not read here — what the
// answer was worth comes from the levels and where the task stood on the
// ladder, which is all the formula has ever read.
//
// An answer of the trial series moves the level differently from every answer
// after it: the level is estimated afresh from the start and all the answers
// of the series at once, and the correction of the topic stays where it is.
//
// The caller is left with a profile to write, and nothing else to remember.
//
//nolint:gocritic // hugeParam: an answer is handed over by value, because it describes something that already happened and this must not look like a struct Record fills in
func (p *Profile) Record(answer Answered) (Recorded, error) {
	task := p.CurrentTask
	switch {
	case task == nil:
		return Recorded{}, ErrNoTask
	case task.ID != answer.TaskID:
		return Recorded{}, fmt.Errorf("%w: %s is in flight", ErrOtherTask, task.ID)
	case answer.Chosen != "" && solver.Place(answer.Chosen) < 0:
		// Refused before anything moves: the history keeps the letter, and a
		// profile holding one that names no option could not be written back.
		return Recorded{}, fmt.Errorf("%w: %q", ErrNoSuchOption, answer.Chosen)
	}

	topic := p.Topics[task.Topic]
	before := rating.State{
		Theta:        p.Ratings.Theta,
		Delta:        topic.Delta,
		Answers:      p.Ratings.Answers,
		TopicAnswers: topic.Answers,
	}
	point := rating.Point{GradeLevel: task.GradeLevel, Difficulty: task.Difficulty}

	var (
		trial       int
		moved       rating.State
		probability float64
	)
	if p.Ratings.InTrial() {
		trial = max(p.Ratings.Answers, 0) + 1
		moved, probability = p.place(before, point, answer.Correct)
	} else {
		updated := rating.Update(before, point.Beta(), answer.Correct)
		moved, probability = updated.State, updated.Probability
	}

	p.Ratings.Theta, p.Ratings.Answers = moved.Theta, moved.Answers
	topic.Delta, topic.Answers = moved.Delta, moved.TopicAnswers
	if answer.Correct {
		p.Ratings.ConsecutiveFailures = 0
	} else {
		p.Ratings.ConsecutiveFailures++
		if answer.Trap != "" {
			if topic.Traps == nil {
				topic.Traps = map[string]int{}
			}
			topic.Traps[answer.Trap]++
		}
	}
	if answer.Correct {
		topic.Correct++
	}

	recorded := Recorded{
		Topic:       task.Topic,
		GradeLevel:  task.GradeLevel,
		Difficulty:  task.Difficulty,
		Probability: probability,
		LevelBefore: before.Level(),
		LevelAfter:  moved.Level(),
		Trial:       trial,
		Pace:        paceOf(task.IssuedAt.Time, answer.At),
	}
	recorded.Mastered, recorded.Unmastered = topic.master(probability, task.GradeLevel, &answer)
	p.Topics[task.Topic] = topic

	p.remember(&Answer{
		AnsweredAt: At(answer.At),
		Chosen:     wrongOnly(answer.Chosen, answer.Correct),
		Confused:   answer.Confused,
		Correct:    answer.Correct,
		Difficulty: task.Difficulty,
		GradeLevel: task.GradeLevel,
		HintUsed:   answer.HintUsed,
		Pace:       recorded.Pace,
		TaskID:     task.ID,
		Topic:      task.Topic,
		Trap:       wrongOnly(answer.Trap, answer.Correct),
	})

	// The lesson is over: the task in flight is what the next answer would be
	// measured against, and there is no next answer to this one.
	p.CurrentTask = nil
	return recorded, nil
}

// place is an answer of the trial series: the level estimated afresh from the
// start and every answer of the series so far, this one included, while the
// correction of the topic stays where it is — a trial answer says where the
// child stands, not what they know of one topic. The counts move as they
// always do, and the chance is the one the task was handed out at.
func (p *Profile) place(before rating.State, point rating.Point, correct bool) (moved rating.State, probability float64) {
	answers := append(p.trialSoFar(), rating.Answer{Point: point, Correct: correct})
	return rating.State{
		Theta:        rating.Estimate(p.Ratings.Start, answers),
		Delta:        before.Delta,
		Answers:      before.Answers + 1,
		TopicAnswers: before.TopicAnswers + 1,
	}, rating.Probability(before.Level(), point.Beta())
}

// trialSoFar are the answers of the trial series already given: the last
// entries of the window, as many as the answers the level rests on. The window
// is never pruned below the length of the series, so while the series runs it
// holds every one of them. A window edited by hand into holding fewer is read
// as it stands, and the answers it no longer shows are not weighed.
func (p *Profile) trialSoFar() []rating.Answer {
	count := min(max(p.Ratings.Answers, 0), len(p.Recent))
	answers := make([]rating.Answer, 0, count+1)
	for i := len(p.Recent) - count; i < len(p.Recent); i++ {
		answers = append(answers, rating.Answer{Point: p.Recent[i].point(), Correct: p.Recent[i].Correct})
	}
	return answers
}

// master moves the two runs this topic keeps and reports whether the answer
// earned mastery or lost it.
//
// The run that earns it counts only answers that prove something: correct, at
// the middle of the corridor or harder, and unaided. A correct answer to an
// easy task leaves the run where it is rather than adding to it — it is not
// evidence against the child, and it is not evidence for them either.
//
// Mastery is held at the level of the task that completed the run, and a run
// completed on a task of a higher level masters the topic again, there: once
// the child's tasks in a topic come from the level above, the topic is back
// in the rotation until they master it at that level too. A run is spent once
// it completes, whether it earns mastery or finds the topic mastered already,
// so mastering the topic at a higher level takes a run of its own rather than
// one answer added to a run of the tasks below.
func (t *Topic) master(probability float64, level rating.GradeLevel, answer *Answered) (mastered, unmastered bool) {
	switch {
	case !answer.Correct:
		t.TopStreak = 0
		t.WrongStreak++
	case probability <= rating.CorridorMiddle && !answer.HintUsed:
		t.TopStreak++
		t.WrongStreak = 0
	default:
		t.WrongStreak = 0
	}

	switch {
	case t.MasteredSince != nil && t.WrongStreak >= MasteryLostAfter:
		t.MasteredSince, t.MasteredLevel = nil, nil
		return false, true
	case t.Answers >= MasteryAnswers && t.TopStreak >= MasteryStreak:
		t.TopStreak = 0
		if t.masteredAtOrAbove(level) {
			return false, false
		}
		since := DateOf(answer.At)
		t.MasteredSince, t.MasteredLevel = &since, &level
		return true, false
	}
	return false, false
}

// masteredAtOrAbove reports whether the topic is mastered at this level or at
// a higher one, which a run completed at this level adds nothing to.
func (t *Topic) masteredAtOrAbove(level rating.GradeLevel) bool {
	return t.MasteredSince != nil && t.MasteredLevel != nil && t.MasteredLevel.Shift() >= level.Shift()
}

// remember puts an answer at the end of the window and drops the oldest when
// the window is full. Nothing is lost by that: everything the rule reads about
// an answer this old has already been added into the summary of its topic.
func (p *Profile) remember(answer *Answer) {
	p.Recent = append(p.Recent, *answer)
	if len(p.Recent) > MaxRecent {
		p.Recent = append([]Answer{}, p.Recent[len(p.Recent)-MaxRecent:]...)
	}
}

// paceOf is how long the child took, measured from the moment the task was
// handed out. A clock that runs backwards — a task issued in the future by a
// profile edited elsewhere — is read as normal rather than as impossibly fast.
func paceOf(issued, answered time.Time) Pace {
	took := answered.Sub(issued)
	switch {
	case took <= 0:
		return PaceNormal
	case took < fastAnswer:
		return PaceFast
	case took > slowAnswer:
		return PaceSlow
	default:
		return PaceNormal
	}
}

// wrongOnly keeps what only a wrong answer carries. A right answer has no
// mistake to name, and naming one would put a trap of the lesson just finished
// where the file says there is none.
func wrongOnly(value string, correct bool) string {
	if correct {
		return ""
	}
	return value
}
