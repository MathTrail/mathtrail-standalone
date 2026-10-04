package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

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
// the request was opened in; and the topic the lessons are kept to, lesson,
// asks for nothing either, nor does a reason given for it alone.
func (in *nextTaskIn) differsFrom(request *profile.OpenRequest, student *profile.Student, lesson string) bool {
	choice := in.choice()
	if lesson != "" && choice.Topic == lesson {
		choice.Topic = ""
		if !choice.Made() {
			choice.Reason = ""
		}
	}
	language, broken := profile.LanguageTag(in.Language)
	_, chosen := student.ChosenLanguage()
	return choice.Made() || choice.Reason != "" ||
		(!chosen && broken.Code == "" && language != "" && language != request.Language)
}

// requestOut is what next_task hands the card it draws: the request a task is
// on its way for, with whom it is for and the language it is written in, or
// why no request was opened. The package to write the task from is in neither
// the payload nor the words, since the card reads both: the model fetches it
// with get_package. A host may show the model the payload in place of the
// words, so what the model acts on is here as well — the request, whether it
// was open already and since when, and the last answer.
type requestOut struct {
	Screen      string       `json:"screen"`
	Status      string       `json:"status,omitempty"`
	Code        string       `json:"code,omitempty"`
	Problems    []problemOut `json:"problems,omitempty"`
	RequestID   string       `json:"request_id,omitempty"`
	AlreadyOpen bool         `json:"already_open,omitempty"`
	AgeSeconds  int          `json:"age_seconds,omitempty"`
	LastAnswer  *answerLine  `json:"last_answer"`
	Child       *childLine   `json:"child"`
	// Language is the request's, which the card's words are in while it waits
	// and which its task is written in. It is empty where no request is open.
	Language string `json:"language,omitempty"`
}

func (s *Service) nextTaskTool() Tool {
	return Define(Spec{
		Name:  "next_task",
		Title: "Ask for the next task",
		Description: "Asks for the child's next task: opens a request and draws the card the task will appear on, " +
			"which shows the child a wait meanwhile. It does not write the task: you do. At once call get_package " +
			"with its request_id for what to write it from — the brief, reference tasks and the page on how to write " +
			"and hand in a task — and hand the task in with submit_task. Always pass language, the language of the " +
			"chat; when the profile names a language for the lessons, the task is written in that one instead, and " +
			"you talk in it too. " +
			"Called again before the task of the open request is handed in, it hands back that request: hand in " +
			"the task you wrote for it, or write it now, rather than asking again. A task on the child's card with " +
			"no answer yet is recorded as skipped, so ask for a new one only when the child wants another. To set a " +
			"topic, a level or a difficulty other than the rule's, pass it with a short reason. While the profile " +
			"keeps the lessons to a topic, lesson_topic, every task is on it once the trial series is over: pass no " +
			"topic of your own then. Never put the child's name in a task." +
			"\n\nTopics, by id, with the levels each is taught at:\n" + s.topicList(),
		Idempotent: true,
		DrawsCard:  true,
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

// nextTask opens a request for a task and draws the card it will come to — or
// hands back the request already open, while the task for it is still to be
// handed in.
func (s *Service) nextTask(ctx context.Context, account store.Account, in nextTaskIn) (Reply[requestOut], error) {
	return afresh(ctx, func() (Reply[requestOut], error) { return s.openRequest(ctx, account, in) })
}

// openRequest is one read of the profile and the write of the request it
// opens, or the request already open.
func (s *Service) openRequest(ctx context.Context, account store.Account, in nextTaskIn) (Reply[requestOut], error) {
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[requestOut]{
			Text:    "No task can be asked for yet. " + firstRunText,
			Payload: requestOut{Screen: screenFirstRun},
		}, nil
	case err != nil:
		return Reply[requestOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	now := s.now()
	if open := p.OpenRequest; open != nil && open.Awaited(s.window, now) {
		return s.stillOpen(ctx, account, p, now, in.differsFrom(open, &p.Student, tutor.LessonTopic(p, s.content)))
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
		return Reply[requestOut]{}, fmt.Errorf("mcp: choose the task: %w", err)
	}
	if len(problems) > 0 {
		return argumentsRefused(p, problems), nil
	}

	skipped, wasSkipped := p.Skip(now)
	request := p.Ask(&brief, mode, p.Student.LessonLanguage(language), now)
	if err := s.canPackage(p, request); err != nil {
		return Reply[requestOut]{}, err
	}
	p.Touch(s.version, now)
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return Reply[requestOut]{}, fmt.Errorf("mcp: save the profile: %w", err)
	}

	lead := ""
	if wasSkipped {
		s.events.write(ctx, account, eventTaskSkipped, s.skippedFields(skipped.Topic, skipped.GradeLevel, skipped.Difficulty)...)
		lead = fmt.Sprintf("Task %s, left on the child's card without an answer, is recorded as skipped.", skipped.TaskID)
	}
	s.events.write(ctx, account, eventTaskRequested, requestFields(request, false)...)
	return Reply[requestOut]{
		Text: joined(lead, fmt.Sprintf("Request %s is open, in %s. You write its task: get the package to write it "+
			"from with get_package and request_id %s, now, then hand the task in with submit_task and the same "+
			"request_id. Where cards are shown, the child sees a card waiting for the task, and the task on it once "+
			"it is accepted.", request.ID, request.Language, request.ID),
			lessonLanguageText(&p.Student), s.lessonTopicText(p), forYouAlone, s.lastAnswerText(p), noPackageTool),
		Payload: requestOut{
			Screen: screenComing, RequestID: request.ID, LastAnswer: lastAnswerOf(p), Child: childLineOf(&p.Student),
			Language: request.Language,
		},
	}, nil
}

// stillOpen hands back the request already open, as it was: its task is still
// to be handed in, and a second task written beside it would race the first
// for the same three attempts. So the words ask for the task written for it
// first, and only then send a model that has not written it — or has lost what
// to write it from — for the package. Nothing is written.
func (s *Service) stillOpen(ctx context.Context, account store.Account, p *profile.Profile, now time.Time, ignored bool) (Reply[requestOut], error) {
	request := p.OpenRequest
	if err := s.canPackage(p, request); err != nil {
		return Reply[requestOut]{}, err
	}
	s.events.write(ctx, account, eventTaskRequested, requestFields(request, true)...)

	age := max(0, int(now.Sub(request.OpenedAt.Time)/time.Second))
	lead := fmt.Sprintf("Request %s is already open, since %d seconds ago, in %s. If you have written its task, "+
		"hand it in with submit_task and request_id %s. If you have not, a turn cut short say, the task is yours "+
		"to write: get its package with get_package and the same request_id. Do not ask for another.",
		request.ID, age, request.Language, request.ID)
	if ignored {
		lead += " The arguments of this call were not applied: the request keeps what it was opened with."
	}
	return Reply[requestOut]{
		Text: joined(lead, lessonLanguageText(&p.Student), stillInText(p), forYouAlone, s.lastAnswerText(p), noPackageTool),
		Payload: requestOut{
			Screen: screenComing, RequestID: request.ID, AlreadyOpen: true, AgeSeconds: age, LastAnswer: lastAnswerOf(p),
			Child: childLineOf(&p.Student), Language: request.Language,
		},
	}, nil
}

// canPackage is whether the package a request's task is written from can be
// built. next_task builds it here only to be sure of that, and throws it away:
// the model fetches it with get_package, and a request it could never fetch
// one for would hold the lesson for the whole of its window.
func (s *Service) canPackage(p *profile.Profile, request *profile.OpenRequest) error {
	_, err := s.packageFor(p, request)
	return err
}

// dayIsFull is a call the day has no room for. Nothing is asked for and nothing
// written, and the model is told what the child is to hear: that the new tasks
// are over for today, when there will be more, and what the child can do
// meanwhile. The card says the same. Which ceiling it was stays in the line,
// since what the child hears is the same either way.
func (s *Service) dayIsFull(p *profile.Profile) Reply[requestOut] {
	return Reply[requestOut]{
		Text: joined("No task was asked for: there are no more new tasks for the child today. Tell the child so: "+
			"there are no more new tasks today, and there will be more tomorrow; meanwhile they can look at their "+
			"progress, or go back over the last task.",
			lessonLanguageText(&p.Student), s.lastAnswerText(p)),
		Payload: requestOut{
			Screen: screenWaiting, Status: statusLimited, Code: codeLimitReached, LastAnswer: lastAnswerOf(p),
			Child: childLineOf(&p.Student),
		},
	}
}

// argumentsRefused is a call no request can be opened from, told argument by
// argument. Nothing is written, and the card says no task comes to it.
func argumentsRefused(p *profile.Profile, problems []profile.Problem) Reply[requestOut] {
	out, lines := problemsOf(problems)
	return Reply[requestOut]{
		Text: "No task was asked for. Fix these arguments and call next_task again: " + lines + ".",
		Payload: requestOut{
			Screen: screenWaiting, Status: statusRejected, Code: codeInvalidArguments, Problems: out,
			LastAnswer: lastAnswerOf(p), Child: childLineOf(&p.Student),
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
