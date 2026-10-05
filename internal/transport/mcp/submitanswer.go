package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// codeStaleTask is the refusal of an answer to a task that is not the one on
// the child's card: one skipped, one left for a newer task, or one never
// given. Nothing is recorded.
const codeStaleTask = "stale_task"

// submitAnswerIn is what submit_answer takes: the task the answer is for, what
// the child answered, and whether they opened the hint first. The answer is
// read by the tool rather than held to a list in the schema, so that a refusal
// says what an answer is without repeating what arrived.
type submitAnswerIn struct {
	TaskID   string `json:"task_id" jsonschema:"the id of the task on the child's card"`
	Answer   string `json:"answer" jsonschema:"the letter of the option the child chose, A to E, or ? when the child says they do not know"`
	HintUsed bool   `json:"hint_used,omitempty" jsonschema:"whether the child opened the hint before answering; false when left out"`
}

// answeredOut is what submit_answer hands back: how the answer went, or why
// none was recorded, and the last answer the child gave. It draws no card of
// its own: the card that sent the answer turns to its result, and the model
// reads the same.
type answeredOut struct {
	Screen     string       `json:"screen"`
	Status     string       `json:"status,omitempty"`
	Code       string       `json:"code,omitempty"`
	Problems   []problemOut `json:"problems,omitempty"`
	LastAnswer *answerLine  `json:"last_answer"`
	Result     *resultOut   `json:"result"`
}

// resultOut is how an answer went, as the result side of the card shows it:
// the child's choice and the right option, whether they are the same, the trap
// behind a wrong letter with what the child is told about it, the solution,
// and the rating in the topic before and after — or, while the trial series
// runs, how many of its answers are in. AlreadyAnswered marks an answer
// recorded by an earlier call and only told again.
type resultOut struct {
	TaskID          string    `json:"task_id"`
	Topic           string    `json:"topic"`
	Choice          string    `json:"choice"`
	Correct         bool      `json:"correct"`
	CorrectAnswer   string    `json:"correct_answer"`
	Trap            *trapOut  `json:"trap"`
	Solution        string    `json:"solution"`
	HintUsed        bool      `json:"hint_used"`
	Rating          *movedOut `json:"rating"`
	Trial           *trialOut `json:"trial"`
	AlreadyAnswered bool      `json:"already_answered"`
}

// trapOut is the mistake a wrong letter leads to: its catalog id, what the
// child is told about it, and whether it has come up before among the latest
// answers — a mistake that repeats, which the model gives a short reminder of
// in its own words.
type trapOut struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Repeated bool   `json:"repeated"`
}

// movedOut is the rating in a topic before an answer and after it, as the
// child is shown it.
type movedOut struct {
	Before int `json:"before"`
	After  int `json:"after"`
}

func (s *Service) submitAnswerTool() Tool {
	return Define(Spec{
		Name:  "submit_answer",
		Title: "Record the child's answer",
		Description: "Records an answer the child gives in the chat to the task on the card, with the task's id: the " +
			"letter of the option the child chose, A to E, or ? when the child says they do not know, which counts " +
			"as a wrong answer. Where cards are shown, an answer given on the card is recorded by the card itself: do " +
			"not ask for one in the chat. Pass hint_used when the child opened the hint first. " +
			"Record the answer before you explain anything. " +
			"The result says whether the answer was right, which option is, what went wrong on the way " +
			"to a wrong one and whether that mistake has come up before, the solution, and how the child's rating in the topic moved — during the trial series, " +
			"how many of its tasks are done instead. An answer is recorded once: the same task answered again, on " +
			"the card or by you, changes nothing and is told what was recorded the first time — the whole result " +
			"while the task is on the card, and whether it was right once the next task has been asked for. The " +
			"answer is written to the profile's file in the adult's Google Drive, with the ratings it moves.",
		Effect:     Adds,
		Idempotent: true,
	}, s.submitAnswer)
}

// submitAnswer records the child's answer to the task on the card, once, and
// tells how it went: to the card that sent it, which turns to its result, and
// to the model, which explains it in words. An answer sent again for the same
// task is told what was recorded, and nothing is written.
func (s *Service) submitAnswer(ctx context.Context, account store.Account, in submitAnswerIn) (Reply[answeredOut], error) {
	return afresh(ctx, func() (Reply[answeredOut], error) { return s.answer(ctx, account, in) })
}

// answer is one read of the profile and the write of the answer recorded in
// it, or what was recorded already.
func (s *Service) answer(ctx context.Context, account store.Account, in submitAnswerIn) (Reply[answeredOut], error) {
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[answeredOut]{
			Text:    "No answer can be recorded yet: there is no task to answer. " + firstRunText,
			Payload: answeredOut{Screen: screenFirstRun, Status: statusStale, Code: codeStaleTask},
		}, nil
	case err != nil:
		return Reply[answeredOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	choice, broken := profile.ChoiceOf(in.Answer)
	if broken.Code != "" {
		return answerRefused(p, broken), nil
	}
	now := s.now()
	masteredBefore := p.CurrentTask != nil && tutor.Mastered(p, s.content, p.CurrentTask.Topic)
	recorded, err := s.recordAnswer(ctx, p, profile.Answered{TaskID: in.TaskID, Choice: choice, HintUsed: in.HintUsed, At: now})
	switch {
	case errors.Is(err, profile.ErrNoTask), errors.Is(err, profile.ErrOtherTask):
		return s.notOnTheCard(p, in.TaskID), nil
	case errors.Is(err, profile.ErrSealed):
		return Reply[answeredOut]{}, s.discard(ctx, account, p, revision, err)
	case err != nil:
		return Reply[answeredOut]{}, fmt.Errorf("mcp: record the answer: %w", err)
	case recorded.Again:
		return s.told(p, &recorded), nil
	}

	// The line is written once the file holds the answer, so that an answer
	// whose write lost to another is not counted when it is sent again.
	p.Touch(s.version, now)
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return Reply[answeredOut]{}, fmt.Errorf("mcp: save the profile: %w", err)
	}
	version := p.CurrentTask.InstructionsVersion
	s.events.writeFor(ctx, account, version, eventAnswerRecorded, s.recordedFields(p, account, &recorded, now)...)
	if recorded.Mastered && !masteredBefore && tutor.Mastered(p, s.content, recorded.Topic) {
		s.events.writeFor(ctx, account, version, eventTopicMastered, append(s.learnerFields(p, account, now),
			zap.String("topic", s.topicLabel(recorded.Topic)),
			zap.Int("grade", p.Student.Grade),
		)...)
	}
	return s.told(p, &recorded), nil
}

// recordedFields are what the line about an answer carries: how it went, the
// child it counts, and how many topics of the catalog the child has mastered
// once it is in, as the progress counts them. The number is the profile's
// rather than one added up from the lines about topics mastered, since a topic
// is lost as well as won, and a topic can count again without an answer when
// the child's level in it falls back to where it was mastered.
func (s *Service) recordedFields(p *profile.Profile, account store.Account, recorded *profile.Recorded, now time.Time) []zap.Field {
	return slices.Concat(
		s.answerFields(recorded, p.Ratings.Answers),
		s.countedFields(p, account, now),
		[]zap.Field{zap.Int("topics_mastered", tutor.MasteredTopics(p, s.content))},
	)
}

// taughtOf is the levels the topic of the task on the card is taught at, which
// its mastery is judged among; none when there is no task, which Record turns
// away before it judges anything.
func (s *Service) taughtOf(p *profile.Profile) []rating.GradeLevel {
	if p.CurrentTask == nil {
		return nil
	}
	return s.content.LevelsOf(p.CurrentTask.Topic)
}

// recordAnswer applies the answer inside a span of its own, which says how it
// went in the words a line about it uses and in no others: whether it was
// right, and the trap of the catalog it fell for. Neither letter reaches it, nor
// the pace or the rating, and an answer only told again says nothing at all.
func (s *Service) recordAnswer(ctx context.Context, p *profile.Profile, answer profile.Answered) (profile.Recorded, error) {
	_, span := s.tracer.Start(ctx, "record_answer")
	defer span.End()

	recorded, err := p.Record(answer, s.sealer, s.taughtOf(p))
	switch {
	case errors.Is(err, profile.ErrSealed):
		span.SetStatus(codes.Error, "the task could not be opened")
		return profile.Recorded{}, err
	case err != nil:
		return profile.Recorded{}, err
	case recorded.Again:
		return recorded, nil
	}
	span.SetAttributes(attribute.Bool("mathtrail.answer.correct", recorded.Correct))
	if recorded.Trap.Trap != "" {
		span.SetAttributes(attribute.String("mathtrail.answer.trap", s.trapLabel(recorded.Trap.Trap)))
	}
	return recorded, nil
}

// discard takes a task whose seal will not open off the card and writes that,
// and is the failure the model is told. Left there, every answer to it would
// fail the same way, and the next ask for a task would count it as one the
// child leafed past. A task that had its answer before the seal was lost is
// told as one: the answer counts, and only telling it again is gone.
func (s *Service) discard(ctx context.Context, account store.Account, p *profile.Profile, revision store.Revision, cause error) error {
	if p.CurrentTask.Answered != nil {
		cause = fmt.Errorf("%w: %w", errToldNoMore, cause)
	}
	p.DiscardTask()
	p.Touch(s.version, s.now())
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return fmt.Errorf("mcp: save the profile: %w", err)
	}
	return fmt.Errorf("mcp: open the task: %w", cause)
}

// told is an answer as the card that sent it shows it, and as the words tell
// it to the model.
func (s *Service) told(p *profile.Profile, recorded *profile.Recorded) Reply[answeredOut] {
	task := p.CurrentTask
	repeated := progress.Repeats(p.Recent, s.content, recorded.Trap.Trap, s.repeats)
	return Reply[answeredOut]{
		Text: s.toldText(task, recorded, repeated),
		Payload: answeredOut{
			Screen: screenResult, LastAnswer: lastAnswerOf(p), Result: resultOf(task.ID, recorded, repeated),
		},
	}
}

// resultOf is how an answer went, as the result side of the card shows it;
// repeated says the trap it fell for has come up before.
func resultOf(taskID string, recorded *profile.Recorded, repeated bool) *resultOut {
	result := &resultOut{
		TaskID:          taskID,
		Topic:           recorded.Topic,
		Choice:          recorded.Choice,
		Correct:         recorded.Correct,
		CorrectAnswer:   recorded.Right,
		Solution:        recorded.Solution,
		HintUsed:        recorded.HintUsed,
		AlreadyAnswered: recorded.Again,
	}
	if recorded.Trap != (profile.Distractor{}) {
		result.Trap = &trapOut{ID: recorded.Trap.Trap, Text: recorded.Trap.Text, Repeated: repeated}
	}
	// While the trial series runs there is no rating to show, only how far it
	// has got; the answer that ends it is the last to show that.
	if recorded.Trial > 0 {
		result.Trial = &trialOut{Answered: recorded.Trial, Of: rating.TrialAnswers}
	} else {
		result.Rating = &movedOut{Before: shownRating(recorded.LevelBefore), After: shownRating(recorded.LevelAfter)}
	}
	return result
}

// shownRating is a level as a rating is shown: a whole number on the chess
// scale.
func shownRating(level float64) int { return rating.Shown(rating.Elo(level)) }

// toldText tells how an answer went, for the model to explain it by: the
// child's choice and the right option with their texts, what went wrong on the
// way to a wrong one — and whether it has come up before — the solution, and
// where the child stands now. The texts of the task stand in quotes, in the
// task's own language, around sentences of the service's.
func (s *Service) toldText(task *profile.CurrentTask, recorded *profile.Recorded, repeated bool) string {
	lead := fmt.Sprintf("The answer to task %s is recorded.", task.ID)
	if recorded.Again {
		lead = fmt.Sprintf("Task %s was answered before, and nothing was recorded now: this is what was recorded then.", task.ID)
	}
	right := fmt.Sprintf("%s) %s", recorded.Right, quoted(task.Options[recorded.Right]))

	var outcome string
	switch {
	case recorded.Choice == profile.DontKnow:
		outcome = fmt.Sprintf("The child did not know, which counts as a wrong answer; the right option is %s. "+
			"Go through the solution with the child step by step, simply and kindly.", right)
	case recorded.Correct:
		outcome = fmt.Sprintf("The child chose %s, and it is right. Praise briefly, and go through the solution if "+
			"they want it.", right)
	default:
		outcome = fmt.Sprintf("The child chose %s) %s, and it is wrong; the right option is %s. %s",
			recorded.Choice, quoted(task.Options[recorded.Choice]), right, mistakeText(recorded.Trap, repeated))
	}
	return joined(lead, outcome, aboutTheStep, s.standingText(recorded)) + "\nSolution: " + quoted(recorded.Solution)
}

// mistakeText is how to explain a wrong letter: from what went wrong on the
// way to it, when the task says, and then the solution — and, when the child
// has made the same mistake before, a short reminder of it in the model's own
// words, since the service keeps none of its own.
func mistakeText(trap profile.Distractor, repeated bool) string {
	if trap.Text == "" {
		return "Go through the solution step by step, kindly."
	}
	text := "Start from what went wrong on the way to it: " + quoted(trap.Text) +
		". Then go through the solution step by step, kindly."
	if repeated {
		text += " The child has made this mistake before among the latest answers: end with one short " +
			"reminder of it, in your own words, that the child can keep in mind next time."
	}
	return text
}

// standingText is where the child stands after the answer: the rating in the
// topic before and after, or, while the trial series runs, how far it has got
// — and on the answer that ends it, that it has ended.
func (s *Service) standingText(recorded *profile.Recorded) string {
	switch {
	case recorded.Trial == rating.TrialAnswers:
		return fmt.Sprintf("That was the last of the %d tasks of the trial series: from now on the child has a "+
			"rating.", rating.TrialAnswers)
	case recorded.Trial > 0:
		return fmt.Sprintf("The trial series is under way: %d of %d tasks done. It finds where the child stands; "+
			"the rating comes after it.", recorded.Trial, rating.TrialAnswers)
	}
	return fmt.Sprintf("The rating in %s went from %d to %d.", s.topicName(recorded.Topic),
		shownRating(recorded.LevelBefore), shownRating(recorded.LevelAfter))
}

// notOnTheCard is an answer to a task that is not the one on the child's card:
// skipped, left for a newer one, or never given. Nothing is recorded. A task
// whose answer the window keeps is told that its answer is recorded, lest the
// model tell the child it was lost; the words say what is on the card now, and
// never repeat an id that arrived without the profile knowing it. The card that
// sent the answer stays on its task.
func (s *Service) notOnTheCard(p *profile.Profile, taskID string) Reply[answeredOut] {
	lead := "The task_id is not the task on the child's card: that task was skipped, left for a newer one or " +
		"never given. Nothing was recorded."
	if answer, found := p.AnswerTo(taskID); found {
		lead = fmt.Sprintf("Task %s was answered already, and that answer, %s, is recorded; the task has left the "+
			"card since, and nothing was recorded now.", answer.TaskID, howItWent(answer.Correct))
	}
	var current string
	switch task, request := p.CurrentTask, p.OpenRequest; {
	case task == nil && request != nil && request.Awaited(s.window, s.now()):
		current = fmt.Sprintf("There is no task on the card yet: request %s is open, and its task is still to be "+
			"handed in with submit_task.", request.ID)
	case task == nil:
		current = "There is no task on the card: ask for a new one with next_task when the child wants another."
	case task.Answered != nil:
		current = fmt.Sprintf("The task on the card, %s, has been answered already: ask for a new task with "+
			"next_task when the child wants another.", task.ID)
	default:
		current = fmt.Sprintf("Task %s is on the card: record the child's answer to it with its id.", task.ID)
	}
	return Reply[answeredOut]{
		Text:    joined(lead, current, s.lastAnswerText(p)),
		Payload: answeredOut{Screen: screenTask, Status: statusStale, Code: codeStaleTask, LastAnswer: lastAnswerOf(p)},
	}
}

// answerRefused is an answer that is none: neither a letter of the options nor
// "I don't know". It is refused by the rule it broke, never by what arrived;
// nothing is recorded, and the card stays on its task.
func answerRefused(p *profile.Profile, broken profile.Broken) Reply[answeredOut] {
	out, line := problemsOf([]profile.Problem{{Field: "answer", Broken: broken}})
	return Reply[answeredOut]{
		Text: "Nothing was recorded. Fix this argument and call submit_answer again: " + line + ".",
		Payload: answeredOut{
			Screen: screenTask, Status: statusRejected, Code: codeInvalidArguments, Problems: out, LastAnswer: lastAnswerOf(p),
		},
	}
}

// answerFields are what the line about an answer keeps of it: where the task
// stood, whether it was right, the trap it fell for by its catalog id, whether
// the hint was opened, whether it was "I don't know", and the pace; the chance
// of a right answer the task was handed out at, to two places, and who chose
// the task, which is what tells whether a chance came true; and which answer
// of the trial series it was, or, after the series, which of the child's
// answers it was, as a range, which is what tells whether the estimate falls
// behind a child as the answers pile up. answers is how many the child has
// given, this one included. Never a letter — neither the child's nor the right
// one — and the topic and the trap held to the catalog, since both are read
// from a file a person can edit.
func (s *Service) answerFields(recorded *profile.Recorded, answers int) []zap.Field {
	trap := ""
	if recorded.Trap.Trap != "" {
		trap = s.trapLabel(recorded.Trap.Trap)
	}
	fields := []zap.Field{
		zap.String("topic", s.topicLabel(recorded.Topic)),
		zap.String("level", string(recorded.GradeLevel)),
		zap.Int("difficulty", recorded.Difficulty),
		zap.Bool("correct", recorded.Correct),
		zap.String("trap", trap),
		zap.Bool("hint_used", recorded.HintUsed),
		zap.Bool("confused", recorded.Choice == profile.DontKnow),
		zap.String("pace", string(recorded.Pace)),
		zap.Float64("chance", math.Round(recorded.Probability*100)/100),
		zap.String("tutor_mode", chooserOf(recorded.TutorMode)),
		zap.Int("trial", recorded.Trial),
	}
	if recorded.Trial == 0 {
		fields = append(fields, zap.String("answers_bucket", answersBucket(answers)))
	}
	return fields
}

// chooserUnknown is who chose a task handed out before the card kept that.
const chooserUnknown = "unknown"

// chooserOf is who chose a task as a line names it: the rule, the model or a
// person, and unknown for a task that does not say.
func chooserOf(mode profile.TutorMode) string {
	if mode.Known() {
		return string(mode)
	}
	return chooserUnknown
}

// answerRanges are the ranges an answer after the trial series is counted in,
// each by the last answer it holds: the first takes the answers just after the
// series, while the estimate is still settling, and each after it about twice
// as many as the one before. The answers past the last are counted together.
var answerRanges = []int{20, 50, 100, 200}

// answersBucket is the range an answer after the trial series falls in, given
// how many answers the child has given with it, named by its first and last
// answer — 6-20, 21-50, 51-100, 101-200 — or, past the last range, by the
// first answer past it, 201+.
func answersBucket(answers int) string {
	from := rating.TrialAnswers + 1
	for _, upTo := range answerRanges {
		if answers <= upTo {
			return fmt.Sprintf("%d-%d", from, upTo)
		}
		from = upTo + 1
	}
	return fmt.Sprintf("%d+", from)
}
