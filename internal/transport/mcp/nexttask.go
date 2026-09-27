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

// codeInvalidArguments is the refusal of arguments no request can be opened
// from. It spends nothing: no request was opened.
const codeInvalidArguments = "invalid_arguments"

// nextTaskIn is what next_task takes: the language the task is to be written
// in, and what the model wants instead of what the rule sets, with its reason.
// The limits are in the words rather than in the schema, and the rule checks
// them, saying what is wrong without repeating it.
type nextTaskIn struct {
	Language   string `json:"language" jsonschema:"the language of the chat as a BCP 47 tag, such as en, ru or pt-BR: the task is written in it"`
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
// not have: a choice of the model's own, a reason, or another language.
func (in *nextTaskIn) differsFrom(request *profile.OpenRequest) bool {
	choice := in.choice()
	language, rule := profile.LanguageTag(in.Language)
	return choice.Made() || choice.Reason != "" || (rule == "" && language != "" && language != request.Language)
}

// requestOut is what next_task hands back beside the package, which travels in
// the words alone: the request a task is to be handed in against, and whether
// it was open already — or, when none could be opened, which argument broke
// which rule.
type requestOut struct {
	Screen      string       `json:"screen"`
	Status      string       `json:"status,omitempty"`
	Code        string       `json:"code,omitempty"`
	Problems    []problemOut `json:"problems,omitempty"`
	LastAnswer  *answerLine  `json:"last_answer"`
	RequestID   string       `json:"request_id,omitempty"`
	AlreadyOpen bool         `json:"already_open"`
	AgeSeconds  int          `json:"age_seconds"`
}

func (s *Service) nextTaskTool() Tool {
	return Define(Spec{
		Name:  "next_task",
		Title: "Ask for the next task",
		Description: "Asks for the child's next task and returns the package to write it from: the brief — the " +
			"topic, the level and the difficulty the rule sets — with reference tasks, the page on how to write and " +
			"hand in a task, and the request id to hand it in with. Always pass language, the language of the chat. " +
			"If a task is already being written, the same request comes back, marked already open: hand in the task " +
			"for it rather than writing a second. A task on the child's card with no answer yet is recorded as " +
			"skipped, so ask for a new one only when the child wants another. To set a topic, a level or a " +
			"difficulty other than the rule's, pass it with a short reason. Never put the child's name in a task." +
			"\n\nTopics, by id, with what each is and the levels it is taught at:\n" + s.topicList(),
		Idempotent: true,
	}, s.nextTask)
}

// topicList is the topic catalog as the model needs it to choose a topic of
// its own: the id, what the topic is and the levels it is taught at, a line
// each.
func (s *Service) topicList() string {
	var lines strings.Builder
	for _, topic := range s.content.Topics() {
		levels := make([]string, 0, len(topic.GradeLevels))
		for _, level := range topic.GradeLevels {
			levels = append(levels, string(level))
		}
		fmt.Fprintf(&lines, "- %s: %s (%s)\n",
			topic.ID, strings.TrimSuffix(topic.Description, "."), strings.Join(levels, ", "))
	}
	return strings.TrimSuffix(lines.String(), "\n")
}

// nextTask opens a request for a task and hands the model what to write it
// from — or hands back the request already open, while a task is still being
// written for it.
func (s *Service) nextTask(ctx context.Context, account store.Account, in nextTaskIn) (Reply[requestOut], error) {
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
		return s.stillOpen(ctx, account, p, now, in.differsFrom(open))
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
	request := p.Ask(&brief, mode, language, now)
	pack, err := s.packageFor(p, request)
	if err != nil {
		return Reply[requestOut]{}, err
	}
	p.Touch(s.version, now)
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return Reply[requestOut]{}, fmt.Errorf("mcp: save the profile: %w", err)
	}

	lead := ""
	if wasSkipped {
		s.events.write(ctx, account, eventTaskSkipped, skippedFields(skipped.Topic, skipped.GradeLevel, skipped.Difficulty)...)
		lead = fmt.Sprintf("Task %s, left on the child's card without an answer, is recorded as skipped.", skipped.TaskID)
	}
	s.events.write(ctx, account, eventTaskRequested, requestFields(request, false)...)
	return Reply[requestOut]{
		Text: joined(lead, fmt.Sprintf("Request %s is open. Write one task in %s to the package below, and hand it "+
			"in with submit_task and request_id %s.", request.ID, request.Language, request.ID)) + packageText(pack),
		Payload: requestOut{Screen: screenWaiting, LastAnswer: lastAnswerOf(p), RequestID: request.ID},
	}, nil
}

// stillOpen hands back the request already open, as it was: a task is being
// written for it, and a second task written beside it would race the first for
// the same three attempts. The package comes again, whole, for a model that
// has lost it. Nothing is written.
func (s *Service) stillOpen(ctx context.Context, account store.Account, p *profile.Profile, now time.Time, ignored bool) (Reply[requestOut], error) {
	request := p.OpenRequest
	pack, err := s.packageFor(p, request)
	if err != nil {
		return Reply[requestOut]{}, err
	}
	s.events.write(ctx, account, eventTaskRequested, requestFields(request, true)...)

	age := max(0, int(now.Sub(request.OpenedAt.Time)/time.Second))
	lead := fmt.Sprintf("A task is already being written for request %s, opened %d seconds ago, in %s. If you have "+
		"written it, hand it in with submit_task and request_id %s; do not start a second one.",
		request.ID, age, request.Language, request.ID)
	if ignored {
		lead += " The arguments of this call were not applied: the request keeps what it was opened with."
	}
	return Reply[requestOut]{
		Text: lead + packageText(pack),
		Payload: requestOut{
			Screen: screenWaiting, LastAnswer: lastAnswerOf(p), RequestID: request.ID, AlreadyOpen: true, AgeSeconds: age,
		},
	}, nil
}

// argumentsRefused is a call no request can be opened from, told argument by
// argument. Nothing is written.
func argumentsRefused(p *profile.Profile, problems []profile.Problem) Reply[requestOut] {
	payload := requestOut{Screen: screenWaiting, Status: statusRejected, Code: codeInvalidArguments, LastAnswer: lastAnswerOf(p)}
	lines := make([]string, 0, len(problems))
	for _, problem := range problems {
		payload.Problems = append(payload.Problems, problemOut{Field: problem.Field, Rule: problem.Rule})
		lines = append(lines, problem.String())
	}
	return Reply[requestOut]{
		Text:    "No task was asked for. Fix these arguments and call next_task again: " + strings.Join(lines, "; ") + ".",
		Payload: payload,
	}
}

// languageOf reads the language a task is to be written in, by the rule the
// cards' own language is read by. A task needs one: written in a language
// guessed, it is a generation wasted.
func languageOf(text string) (string, []profile.Problem) {
	tag, rule := profile.LanguageTag(text)
	if tag == "" && rule == "" {
		rule = "is required: the language of the chat, as a BCP 47 tag such as en, ru or pt-BR"
	}
	if rule != "" {
		return "", []profile.Problem{{Field: "language", Rule: rule}}
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
// stood.
func skippedFields(topic string, level rating.GradeLevel, difficulty int) []zap.Field {
	return []zap.Field{
		zap.String("topic", topic),
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
