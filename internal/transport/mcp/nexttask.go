package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The refusals of next_task. Neither spends anything: no request was opened.
const (
	// codeInvalidArguments is the refusal of arguments no request can be
	// opened from.
	codeInvalidArguments = "invalid_arguments"
	// codeLimitReached is the refusal of a day that has no room for another
	// task.
	codeLimitReached = "limit_reached"
)

// nextTaskIn is what next_task takes: the language the task is to be written
// in, and what the model wants instead of what the rule sets, with its reason.
// The limits are in the words rather than in the schema, and the rule checks
// them, saying what is wrong without repeating it.
type nextTaskIn struct {
	Language   string `json:"language" jsonschema:"the language of the chat as a BCP 47 tag, such as en, ru or pt-BR: the task is written in it, unless the profile names a language for the lessons"`
	Topic      string `json:"topic,omitempty" jsonschema:"a topic of your own instead of the rule's, by its id from the list in this tool's description. Needs a reason"`
	GradeLevel string `json:"grade_level,omitempty" jsonschema:"a level of your own: 1-2, 3-4 or 5-6, the grades a task is written for. Needs a reason"`
	Difficulty int    `json:"difficulty,omitempty" jsonschema:"a difficulty of your own inside the level, 1 to 5. Needs a reason"`
	Reason     string `json:"reason,omitempty" jsonschema:"why the child needs your topic, level or difficulty now, in a sentence of at most 300 characters. Required with any of them"`
}

// choice is what the model asked for instead of the rule, as it was sent: a
// topic or a level it did not send is empty, and anything it did send is held
// to the rules as it stands.
func (in *nextTaskIn) choice() tutor.Choice {
	return tutor.Choice{
		Topic:      in.Topic,
		GradeLevel: rating.GradeLevel(in.GradeLevel),
		Difficulty: in.Difficulty,
		Reason:     in.Reason,
	}
}

// differsFrom reports whether the call asks for anything the request open does
// not have: a choice of the model's own, a reason, or — when the parent chose
// no language for the lessons — another chat language. A language the parent
// chose wins over the chat's, so the chat's then asks for nothing, whatever
// the request was opened in.
func (in *nextTaskIn) differsFrom(request *profile.OpenRequest, student *profile.Student) bool {
	choice := in.choice()
	language, broken := profile.LanguageTag(in.Language)
	_, chosen := student.ChosenLanguage()
	return choice.Made() || choice.Reason != "" ||
		(!chosen && broken.Code == "" && language != "" && language != request.Language)
}

// requestRefusedOut is what next_task hands back when no request can be opened:
// which argument broke which rule, or that the day has no room for a task. It
// is the one result of the tool that has a payload, because its status is what
// marks it a refusal. Every other result is words alone: the package is written
// for the model and travels in the words, and a host that shows the model a
// result's payload in place of its words would otherwise hand it a request
// with nothing to write the task from. No card is drawn from next_task, so a
// payload would have no other reader.
type requestRefusedOut struct {
	Screen     string       `json:"screen"`
	Status     string       `json:"status"`
	Code       string       `json:"code"`
	Problems   []problemOut `json:"problems,omitempty"`
	LastAnswer *answerLine  `json:"last_answer"`
}

func (s *Service) nextTaskTool() Tool {
	return Define(Spec{
		Name:  "next_task",
		Title: "Ask for the next task",
		Description: "Asks for the child's next task and returns the package to write it from: the brief — the " +
			"topic, the level and the difficulty the rule sets — with reference tasks, the page on how to write and " +
			"hand in a task, and the request id to hand it in with. Always pass language, the language of the chat; " +
			"when the profile names a language for the lessons, the task is written in that one instead, and you " +
			"talk in it too. " +
			"Called again before the task of the open request is handed in, it hands back that request with its " +
			"package: hand in the task you wrote for it, or write it now, rather than asking again. A task on the " +
			"child's card with no answer yet is recorded as skipped, so ask for a new one only when the child wants " +
			"another. To set a topic, a level or a difficulty other than the rule's, pass it with a short reason. " +
			"Never put the child's name in a task." +
			"\n\nTopics, by id, with the levels each is taught at:\n" + s.topicList(),
		Idempotent: true,
	}, s.nextTask)
}

// topicList is the topic catalog as the model needs it to choose a topic of
// its own: the id and the levels it is taught at, a line each. What a topic is
// stays out, since the ids say it well enough: with it the list would push the
// description past what a host may keep of one, and the topics at its end would
// be the ones lost. The package describes the topic a task is written on.
func (s *Service) topicList() string {
	var lines strings.Builder
	for _, topic := range s.content.Topics() {
		levels := make([]string, 0, len(topic.GradeLevels))
		for _, level := range topic.GradeLevels {
			levels = append(levels, string(level))
		}
		fmt.Fprintf(&lines, "- %s (%s)\n", topic.ID, strings.Join(levels, ", "))
	}
	return strings.TrimSuffix(lines.String(), "\n")
}

// nextTask opens a request for a task and hands the model what to write it
// from — or hands back the request already open, while the task for it is
// still to be handed in. Everything it says is in the words; only a refusal
// has a payload.
func (s *Service) nextTask(ctx context.Context, account store.Account, in nextTaskIn) (Reply[any], error) {
	return afresh(ctx, func() (Reply[any], error) { return s.openRequest(ctx, account, in) })
}

// openRequest is one read of the profile and the write of the request it
// opens, or the request already open.
func (s *Service) openRequest(ctx context.Context, account store.Account, in nextTaskIn) (Reply[any], error) {
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[any]{Text: "No task can be asked for yet. " + firstRunText}, nil
	case err != nil:
		return Reply[any]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	now := s.now()
	if open := p.OpenRequest; open != nil && open.Awaited(s.window, now) {
		return s.stillOpen(ctx, account, p, now, in.differsFrom(open, &p.Student))
	}
	if limit, count, reached := s.daily.reached(p.Daily.Today(now)); reached {
		limitHit(ctx, s.events.logger, s.events.projectID, account.ID, limit, zap.Int("count", count))
		return s.dayIsFull(p), nil
	}

	language, problems := languageOf(in.Language)
	brief, mode, err := tutor.Next(p, s.content, in.choice())
	var refused *tutor.ChoiceError
	switch {
	case errors.As(err, &refused):
		problems = append(problems, refused.Problems...)
	case err != nil:
		return Reply[any]{}, fmt.Errorf("mcp: choose the task: %w", err)
	}
	if len(problems) > 0 {
		return argumentsRefused(p, problems), nil
	}

	skipped, wasSkipped := p.Skip(now)
	request := p.Ask(&brief, mode, p.Student.LessonLanguage(language), now)
	pack, err := s.packageFor(p, request)
	if err != nil {
		return Reply[any]{}, err
	}
	p.Touch(s.version, now)
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return Reply[any]{}, fmt.Errorf("mcp: save the profile: %w", err)
	}

	lead := ""
	if wasSkipped {
		s.events.write(ctx, account, eventTaskSkipped, s.skippedFields(skipped.Topic, skipped.GradeLevel, skipped.Difficulty)...)
		lead = fmt.Sprintf("Task %s, left on the child's card without an answer, is recorded as skipped.", skipped.TaskID)
	}
	s.events.write(ctx, account, eventTaskRequested, requestFields(request, false)...)
	return Reply[any]{
		Text: joined(lead, fmt.Sprintf("Request %s is open. Write one task in %s to the package below, and hand it "+
			"in with submit_task and request_id %s.", request.ID, request.Language, request.ID),
			lessonLanguageText(&p.Student), forYouAlone, s.lastAnswerText(p)) + packageText(pack),
	}, nil
}

// stillOpen hands back the request already open, as it was: its task is still
// to be handed in, and a second task written beside it would race the first
// for the same three attempts. So the words ask for the task written for it
// first, and only then offer the package, which comes again, whole, for a model
// that has not written the task or has lost what to write it from. Nothing is
// written.
func (s *Service) stillOpen(ctx context.Context, account store.Account, p *profile.Profile, now time.Time, ignored bool) (Reply[any], error) {
	request := p.OpenRequest
	pack, err := s.packageFor(p, request)
	if err != nil {
		return Reply[any]{}, err
	}
	s.events.write(ctx, account, eventTaskRequested, requestFields(request, true)...)

	age := max(0, int(now.Sub(request.OpenedAt.Time)/time.Second))
	lead := fmt.Sprintf("Request %s is already open, since %d seconds ago, in %s. If you have written its task, "+
		"hand it in with submit_task and request_id %s. If you have not, a turn cut short say, the task is yours "+
		"to write, to the package below. Do not ask for another.", request.ID, age, request.Language, request.ID)
	if ignored {
		lead += " The arguments of this call were not applied: the request keeps what it was opened with."
	}
	return Reply[any]{Text: joined(lead, lessonLanguageText(&p.Student), stillInText(p), forYouAlone, s.lastAnswerText(p)) +
		packageText(pack)}, nil
}

// dayIsFull is a call the day has no room for. Nothing is asked for and nothing
// written, and the model is told what the child is to hear: that the new tasks
// are over for today, when there will be more, and what the child can do
// meanwhile. Which ceiling it was stays in the line, since what the child hears
// is the same either way.
func (s *Service) dayIsFull(p *profile.Profile) Reply[any] {
	return Reply[any]{
		Text: joined("No task was asked for: there are no more new tasks for the child today. Tell the child so: "+
			"there are no more new tasks today, and there will be more tomorrow; meanwhile they can look at their "+
			"progress, or go back over the last task.",
			lessonLanguageText(&p.Student), s.lastAnswerText(p)),
		Payload: requestRefusedOut{Screen: screenWaiting, Status: statusLimited, Code: codeLimitReached, LastAnswer: lastAnswerOf(p)},
	}
}

// argumentsRefused is a call no request can be opened from, told argument by
// argument. Nothing is written.
func argumentsRefused(p *profile.Profile, problems []profile.Problem) Reply[any] {
	out, lines := problemsOf(problems)
	return Reply[any]{
		Text: "No task was asked for. Fix these arguments and call next_task again: " + lines + ".",
		Payload: requestRefusedOut{
			Screen: screenWaiting, Status: statusRejected, Code: codeInvalidArguments, Problems: out, LastAnswer: lastAnswerOf(p),
		},
	}
}

// languageOf reads the chat's language, by the rule the language the parent
// chose is read by. Every call needs one, even when the profile names a
// language of its own: a profile that names none would otherwise leave the
// lesson's language to a guess, and a task written in a language guessed is a
// generation wasted.
func languageOf(text string) (string, []profile.Problem) {
	tag, broken := profile.LanguageTag(text)
	if tag == "" && broken.Code == "" {
		broken = profile.Broken{Code: profile.CodeRequired,
			Rule: "is required: the language of the chat, as a BCP 47 tag such as en, ru or pt-BR"}
	}
	if broken.Code != "" {
		return "", []profile.Problem{{Field: "language", Broken: broken}}
	}
	return tag, nil
}

// packageFor is what the model is handed to write the request's task from: the
// brief the request keeps, the chances around it for this child now, and the
// child as the task is to be pitched at them.
func (s *Service) packageFor(p *profile.Profile, request *profile.OpenRequest) ([]byte, error) {
	pack, err := s.content.Package(&content.Request{
		Language:  request.Language,
		Brief:     request.Brief,
		Corridor:  tutor.CorridorIn(p, s.content, request.Brief.TargetConcept),
		Grade:     p.Student.Grade,
		Interests: p.Student.Interests,
		Notes:     p.Student.Notes,
		Answers:   p.Ratings.Answers,
	})
	if err != nil {
		return nil, fmt.Errorf("mcp: build the package: %w", err)
	}
	return pack, nil
}

// packageText puts the package after the words about it, where the model reads
// it. It travels here alone and never in the payload: a card is drawn from the
// payload, and the package describes the child and the task to come.
func packageText(pack []byte) string {
	return "\n\nPackage:\n" + string(pack)
}

// skippedFields are what the line about a skipped task keeps of it: where it
// stood. The task is read from the profile, which a person can edit, so the
// line names its topic only while the catalog has it.
func (s *Service) skippedFields(topic string, level rating.GradeLevel, difficulty int) []zap.Field {
	return []zap.Field{
		zap.String("topic", s.topicLabel(topic)),
		zap.String("level", string(level)),
		zap.Int("difficulty", difficulty),
	}
}

// requestFields are what the line about a request keeps of it: where the task
// is to stand, why, and who chose it.
func requestFields(request *profile.OpenRequest, alreadyOpen bool) []zap.Field {
	return []zap.Field{
		zap.String("topic", request.Brief.TargetConcept),
		zap.String("level", string(request.Brief.GradeLevel)),
		zap.Int("difficulty", request.Brief.Difficulty),
		zap.String("goal", string(request.Brief.PedagogicalGoal)),
		zap.String("tutor_mode", string(request.TutorMode)),
		zap.Bool("already_open", alreadyOpen),
	}
}
