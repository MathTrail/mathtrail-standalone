package profile

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The refusals of Record. An answer that arrives for a task nobody is working
// on is not a fault of the service, and the two cases read differently to
// whoever hears about them: there is no task on the card at all, or the
// answer is for a task that is no longer there.
var (
	// ErrNoTask means there is no task on the card to answer.
	ErrNoTask = errors.New("profile: no task on the card")
	// ErrNoSuchOption means the answer is neither a letter of the options nor
	// "I don't know".
	ErrNoSuchOption = errors.New("profile: the answer names no option of the task")
	// ErrOtherTask means the answer is for a task other than the one on the
	// card: one skipped, one left behind by the next, or one never given.
	ErrOtherTask = errors.New("profile: the answer is for another task")
)

// DontKnow is the answer "I don't know". The card shows the solution for it,
// so it is an answer, and a wrong one: a child who has seen the solution has
// not solved the task. It chose no option and took no trap, so neither is
// written, and the map of the child's mistakes does not count it.
const DontKnow = "?"

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

// Answered is one answer as it reaches the service: the task it is for, what
// the child chose — a letter of the options, or DontKnow — whether they opened
// the hint first, and when it arrived. Whether it was right is not among it:
// that is for the sealed part of the task to say.
type Answered struct {
	TaskID   string
	Choice   string
	HintUsed bool
	At       time.Time
}

// Given is the answer a task was given, kept with the task while it stays on
// the card. An answer sent twice — from a second tab, after a reply lost on its
// way, or by the model recording what the card already did — is told what was
// recorded rather than recorded again, and this, with the seal that stays with
// the task, is what tells it word for word.
type Given struct {
	// Choice is what the child chose: a letter of the options, or DontKnow.
	Choice string `json:"choice"`
	// HintUsed records that the hint was opened first.
	HintUsed bool `json:"hint_used"`
	// LevelAfter and LevelBefore are where the child stood in the task's topic
	// after the answer and before it.
	LevelAfter  float64 `json:"level_after"`
	LevelBefore float64 `json:"level_before"`
	// Trial is which answer of the trial series this was, from 1, or zero when
	// the series was already over.
	Trial int `json:"trial"`
}

// Recorded is an answer as it was recorded, for whoever has to say so: the
// result the child is shown, the words the model relays and the line written
// to the log.
type Recorded struct {
	// Topic is what was answered, GradeLevel the level the task was written
	// for and Difficulty its difficulty inside that level.
	Topic      string
	GradeLevel rating.GradeLevel
	Difficulty int
	// Choice is what the child chose: a letter of the options, or DontKnow.
	Choice string
	// Correct says whether that was the right option. DontKnow never is.
	Correct bool
	// Right is the letter of the right option.
	Right string
	// Trap is the mistake a wrong letter leads to, with what the child is told
	// about it. It is empty for a right answer, and for DontKnow, which took
	// no trap.
	Trap Distractor
	// Solution is the worked answer the child is shown.
	Solution string
	// HintUsed records that the hint was opened first.
	HintUsed bool
	// LevelBefore and LevelAfter are where the child stood in this topic
	// before the answer and after it.
	LevelBefore float64
	LevelAfter  float64
	// Trial is which answer of the trial series this was, from 1, or zero when
	// the series was already over. While it runs there is no rating to show,
	// only how many of its answers are in.
	Trial int
	// Probability is the chance of a correct answer as it stood when the task
	// was handed out, and Pace is how long the answer took.
	Probability float64
	Pace        Pace
	// Mastered is set when this answer earned mastery of the topic, and
	// Unmastered when it lost it. Both cannot be true of one answer.
	Mastered   bool
	Unmastered bool
	// Again marks an answer recorded by an earlier call and only told again
	// now. Nothing moved, and what only recording it knew — the chance, the
	// pace, whether mastery was won or lost — is not told again.
	Again bool
}

// ChoiceOf reads an answer as the child or the model typed it: a letter of the
// options in either case, or DontKnow, with any space around it dropped. What
// it returns is the choice, or the rule the text broke, in words that never
// repeat the text.
func ChoiceOf(text string) (choice, rule string) {
	choice = strings.ToUpper(strings.TrimSpace(text))
	if !choosable(choice) {
		letters := solver.Letters()
		return "", fmt.Sprintf("must be one of the letters %s to %s, or %s when the child does not know",
			letters[0], letters[len(letters)-1], DontKnow)
	}
	return choice, ""
}

// choosable reports whether a choice is one the child can make: a letter of
// the options, as it labels them, or DontKnow.
func choosable(choice string) bool { return choice == DontKnow || solver.Place(choice) >= 0 }

// Record takes the child's answer to the task on the card, once. The sealed
// part of the task says whether it was right and, for a wrong letter, the trap
// behind it; then the level moves, the summary of the topic adds it up, the
// answer goes to the end of the history window, and the task keeps what it
// was given. An answer to a task that has one already moves nothing: it is told
// what was recorded, whatever it chose itself.
//
// An answer of the trial series moves the level differently from every answer
// after it: the level is estimated afresh from the start and all the answers
// of the series at once, and the correction of the topic stays where it is.
//
// Nothing moves when the seal cannot be opened: a task whose answer cannot be
// read cannot be checked either. An answer recorded now seals the task again
// against the letter it was given, which is what lets it be told again — and
// only it: a task the file says was answered, which was never answered here,
// does not open.
//
// The caller is left with a profile to write when the answer was recorded now,
// and with nothing to write when it was only told again.
func (p *Profile) Record(answer Answered, sealer Sealer) (Recorded, error) {
	task := p.CurrentTask
	switch {
	case task == nil:
		return Recorded{}, ErrNoTask
	case task.ID != answer.TaskID:
		return Recorded{}, fmt.Errorf("%w: %s is on the card", ErrOtherTask, task.ID)
	case !choosable(answer.Choice):
		// Refused before anything moves: the history keeps the letter of a
		// wrong answer, and a profile holding one that names no option could
		// not be written back.
		return Recorded{}, ErrNoSuchOption
	}

	secret, err := p.OpenTask(sealer)
	if err != nil {
		return Recorded{}, err
	}
	if task.Answered != nil {
		told := toldOf(task, &secret)
		told.Again = true
		return told, nil
	}
	resealed, err := sealSecret(sealer, secret, answeredBinding(p.StudentID, task.ID, answer.Choice))
	if err != nil {
		return Recorded{}, err
	}
	recorded := p.apply(task, &answer, &secret)
	task.Sealed = resealed
	return recorded, nil
}

// apply records an answer the task has not had before and keeps it with the
// task. Only correctness moves the level; the hint is written beside the
// answer and changes what comes next through the runs of the topic.
func (p *Profile) apply(task *CurrentTask, answer *Answered, secret *TaskSecret) Recorded {
	correct, trap := judge(answer.Choice, secret)
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
		moved, probability = p.place(before, point, correct)
	} else {
		updated := rating.Update(before, point.Beta(), correct)
		moved, probability = updated.State, updated.Probability
	}

	p.Ratings.Theta, p.Ratings.Answers = moved.Theta, moved.Answers
	topic.Delta, topic.Answers = moved.Delta, moved.TopicAnswers
	if correct {
		p.Ratings.ConsecutiveFailures = 0
		topic.Correct++
	} else {
		p.Ratings.ConsecutiveFailures++
		if trap.Trap != "" {
			if topic.Traps == nil {
				topic.Traps = map[string]int{}
			}
			topic.Traps[trap.Trap]++
		}
	}

	task.Answered = &Given{
		Choice:      answer.Choice,
		HintUsed:    answer.HintUsed,
		LevelAfter:  moved.Level(),
		LevelBefore: before.Level(),
		Trial:       trial,
	}
	recorded := toldOf(task, secret)
	recorded.Probability = probability
	recorded.Pace = paceOf(task.IssuedAt.Time, answer.At)
	recorded.Mastered, recorded.Unmastered = topic.master(probability, task.GradeLevel, correct, answer)
	p.putTopic(task.Topic, &topic)

	p.remember(&Answer{
		AnsweredAt: At(answer.At),
		Chosen:     chosenOf(answer.Choice, correct),
		Confused:   answer.Choice == DontKnow,
		Correct:    correct,
		Difficulty: task.Difficulty,
		GradeLevel: task.GradeLevel,
		HintUsed:   answer.HintUsed,
		Pace:       recorded.Pace,
		TaskID:     task.ID,
		Topic:      task.Topic,
		Trap:       trap.Trap,
	})
	return recorded
}

// toldOf is the answer a task was given, told from what the task keeps of it
// and from its seal. Recording an answer and telling it again both end here,
// which is what makes the second telling the same as the first.
func toldOf(task *CurrentTask, secret *TaskSecret) Recorded {
	given := task.Answered
	correct, trap := judge(given.Choice, secret)
	return Recorded{
		Topic:       task.Topic,
		GradeLevel:  task.GradeLevel,
		Difficulty:  task.Difficulty,
		Choice:      given.Choice,
		Correct:     correct,
		Right:       secret.Answer,
		Trap:        trap,
		Solution:    secret.Solution,
		HintUsed:    given.HintUsed,
		LevelBefore: given.LevelBefore,
		LevelAfter:  given.LevelAfter,
		Trial:       given.Trial,
	}
}

// judge says whether a choice is the right option and, for a wrong letter, the
// trap it leads to. DontKnow is never right and took no trap.
func judge(choice string, secret *TaskSecret) (correct bool, trap Distractor) {
	switch choice {
	case DontKnow:
		return false, Distractor{}
	case secret.Answer:
		return true, Distractor{}
	}
	return false, secret.Distractors[choice]
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
// answers of the window, as many as the answers the level rests on, with the
// skipped tasks between them passed over. The window always keeps as many of
// the latest answers as the series is long, so while the series runs it holds
// every one of them. A window edited by hand into holding fewer is read as it
// stands, and the answers it no longer shows are not weighed.
func (p *Profile) trialSoFar() []rating.Answer {
	given := p.answers()
	count := min(max(p.Ratings.Answers, 0), len(given))
	// Only an answer of the series asks for them, so they and the answer on its
	// way never outnumber the series.
	answers := make([]rating.Answer, 0, rating.TrialAnswers)
	for i := len(given) - count; i < len(given); i++ {
		answers = append(answers, rating.Answer{Point: given[i].point(), Correct: given[i].Correct})
	}
	return answers
}

// answers are the entries of the window that are answers, oldest first: the
// window without its skipped tasks.
func (p *Profile) answers() []Answer {
	return slices.DeleteFunc(slices.Clone(p.Recent), func(entry Answer) bool { return entry.Skipped })
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
func (t *Topic) master(probability float64, level rating.GradeLevel, correct bool, answer *Answered) (mastered, unmastered bool) {
	switch {
	case !correct:
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

// remember puts an entry at the end of the window and, while the window is
// over full, lets one go. Nothing is lost by that: everything the rule reads
// about an entry this old has already been added into the summary of its
// topic.
func (p *Profile) remember(entry *Answer) {
	p.Recent = append(p.Recent, *entry)
	for len(p.Recent) > MaxRecent {
		gone := leaving(p.Recent)
		p.Recent = slices.Delete(p.Recent, gone, gone+1)
	}
}

// leaving is the place of the entry a full window lets go of: the oldest —
// unless it is one of the latest answers the window keeps, and then the oldest
// skipped task instead. The rule and the trial series read those answers back,
// and a run of skipped tasks must not push them out; a skipped task, new or
// old, is only ever pushed out by another entry.
func leaving(window []Answer) int {
	if window[0].Skipped || answersAfter(window, 0) >= MinRecent {
		return 0
	}
	// A full window whose oldest answer is one of the latest holds more
	// skipped tasks than answers, so there is always one to let go.
	return slices.IndexFunc(window, func(entry Answer) bool { return entry.Skipped })
}

// answersAfter counts the answers that came after the entry at this place.
func answersAfter(window []Answer, place int) int {
	count := 0
	for i := place + 1; i < len(window); i++ {
		if !window[i].Skipped {
			count++
		}
	}
	return count
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

// chosenOf is the option a wrong answer chose, and nothing for a right answer
// or for DontKnow: a right answer has no mistake to name — naming one would put
// a trap of the lesson just finished where the file says there is none — and
// "I don't know" chose no option at all.
func chosenOf(choice string, correct bool) string {
	if correct || choice == DontKnow {
		return ""
	}
	return choice
}
