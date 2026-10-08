package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// prepareTaskIn is what prepare_task takes: the language of the chat, and the
// model's own idea of where the next task should stand instead of where the
// rule sets it, with its reason. The limits are in the words rather than in
// the schema, and the rule checks them, saying what is wrong without
// repeating it.
type prepareTaskIn struct {
	Language   string `json:"language" jsonschema:"the language of the chat as a BCP 47 tag, such as en, ru or pt-BR: the task is written in it, unless the profile names a language for the lessons"`
	Topic      string `json:"topic,omitempty" jsonschema:"a topic of your own for the next task instead of the rule's, by its id from the list in next_task's description. Needs a reason"`
	GradeLevel string `json:"grade_level,omitempty" jsonschema:"a level of your own for the next task: 1-2, 3-4 or 5-6, the grades a task is written for. Needs a reason"`
	Difficulty int    `json:"difficulty,omitempty" jsonschema:"a difficulty of your own for the next task inside the level, 1 to 5. Needs a reason"`
	Reason     string `json:"reason,omitempty" jsonschema:"why the child needs your topic, level or difficulty next, in a sentence of at most 300 characters. Required with any of them"`
}

// choice is what the model asked for instead of the rule, as it was sent.
func (in *prepareTaskIn) choice() tutor.Choice {
	return tutor.Choice{
		Topic:      in.Topic,
		GradeLevel: rating.GradeLevel(in.GradeLevel),
		Difficulty: in.Difficulty,
		Reason:     in.Reason,
	}
}

// preparedRefusedOut is what prepare_task hands back when its arguments cannot
// set a task: the one result of the tool with a payload, its status marking it
// a refusal. Every other result is words alone, as get_package's are: the
// package travels in the words, and the tool draws no card.
type preparedRefusedOut struct {
	Status     string       `json:"status"`
	Code       string       `json:"code"`
	Problems   []problemOut `json:"problems"`
	LastAnswer *answerLine  `json:"last_answer"`
}

func (s *Service) prepareTaskTool() Tool {
	return Define(Spec{
		Name:  "prepare_task",
		Title: "Get the next task ready",
		Description: "Gets the next task ready. It hands you the package of the task to write now: the one the " +
			"child's card waits for, when there is one; otherwise the next task, which you write ahead while the " +
			"child works on the one on the card, and which is kept, sealed, until another is asked for — then " +
			"next_task puts it, at once, on the card it draws. When the next task is written already, or the day " +
			"has no room for another, it says there is nothing to write. Call it once a task is on the card, and as " +
			"the last step of every turn of a lesson. Always pass language, the language of the chat. Your own " +
			"idea of what the child needs next — an easier task after several misses, for example — goes here, as " +
			"topic, grade_level or difficulty with a short reason, and shapes the task written ahead. Write the task " +
			"to the package and hand it in with submit_task and its request_id. It writes to the profile's file in " +
			"the adult's Google Drive the request it opens for the task written ahead, and lets go of a task kept " +
			"that no longer fits the lesson.",
		Effect:     Adds,
		Idempotent: true,
	}, s.prepareTask)
}

// prepareTask hands the model the package of the task to write now.
func (s *Service) prepareTask(ctx context.Context, account store.Account, in prepareTaskIn) (Reply[any], error) {
	return afresh(ctx, func() (Reply[any], error) { return s.prepare(ctx, account, &in) })
}

// prepare is one read of the profile and the package of the task to write
// now: the one a card waits for, or the next one, written ahead. The profile
// is written only when a request is opened for the task written ahead, or a
// task written ahead that no longer fits the lesson is let go.
func (s *Service) prepare(ctx context.Context, account store.Account, in *prepareTaskIn) (Reply[any], error) {
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[any]{Text: "No task can be written yet. " + firstRunText}, nil
	case err != nil:
		return Reply[any]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	now := s.now()
	if open := s.waitedOpen(p, now); open != nil {
		noteTaskRequest(ctx, open.ID)
		return s.packageWords(p, open, fmt.Sprintf("Request %[1]s is open, and the child waits for its task on the "+
			"card: write it now, to the package below, and hand it in with submit_task and request_id %[1]s.", open.ID))
	}
	language, problems := languageOf(in.Language)
	if len(problems) > 0 {
		return prepareRefused(p, problems), nil
	}

	lesson := s.lessonNow(p, language, profile.Place{})
	gone := s.letGo(p, &lesson, now)
	reply, opened, err := s.prepared(p, in, language, now)
	if err != nil || reply.Payload != nil {
		return reply, err
	}
	if p.OpenRequest != nil {
		noteTaskRequest(ctx, p.OpenRequest.ID)
	}
	if len(gone) > 0 || opened != nil {
		p.Touch(s.version, now)
		if _, err := s.store.Save(ctx, account, p, revision); err != nil {
			return Reply[any]{}, fmt.Errorf("mcp: save the profile: %w", err)
		}
	}
	s.sayLetGo(ctx, account, gone)
	if opened != nil {
		s.events.write(ctx, account, eventTaskRequested, requestFields(opened, false)...)
	}
	return reply, nil
}

// prepared is what there is to write now, the task the child waits for set
// aside: the task being written ahead, its package handed back; nothing, when
// the next task is kept already or the day has no room for another; or the
// next task, its request opened here — and then returned, for the caller to
// write. A choice no request can be opened from is refused, with a payload.
func (s *Service) prepared(p *profile.Profile, in *prepareTaskIn, language string, now time.Time) (
	reply Reply[any], opened *profile.OpenRequest, err error,
) {
	if open := s.aheadOpen(p, now); open != nil {
		lead := aheadLead(open)
		if choice := in.choice(); choice.Made() {
			lead += " The arguments of this call were not applied: the request keeps what it was opened with."
		}
		reply, err = s.packageWords(p, open, lead)
		return reply, nil, err
	}
	if p.ReadyTask != nil {
		return Reply[any]{Text: joined("The next task is written already, and kept for the child until another is "+
			"asked for: there is nothing to write now. Say nothing about it.",
			s.lastAnswerText(p))}, nil, nil
	}
	if _, _, reached := s.daily.reached(p.Daily.Today(now)); reached {
		return Reply[any]{Text: joined("The child's tasks for today are over: there is nothing to write ahead, and "+
			"nothing to say about it.", s.lastAnswerText(p))}, nil, nil
	}

	brief, mode, err := s.aheadBrief(p, in.choice())
	var refused *tutor.ChoiceError
	switch {
	case errors.As(err, &refused):
		return prepareRefused(p, refused.Problems), nil, nil
	case err != nil:
		return Reply[any]{}, nil, fmt.Errorf("mcp: choose the task: %w", err)
	}
	opened, err = p.AskAhead(&brief, mode, p.Student.LessonLanguage(language), tutor.LessonTopic(p, s.content), now)
	if err != nil {
		return Reply[any]{}, nil, fmt.Errorf("mcp: ask for the task ahead: %w", err)
	}
	reply, err = s.packageWords(p, opened, aheadLead(opened))
	return reply, opened, err
}

// aheadBrief is the brief of the task written ahead, and who chose it: where
// the model asked, when it asked; one more like the task on the card when a
// person asked for that one, in case the child wants another like it; and
// where the rule sets it otherwise. One more like it that no brief can be
// built on any more — its topic since left out of the lessons or the catalog —
// is the rule's: the model passed nothing it could fix.
func (s *Service) aheadBrief(p *profile.Profile, own tutor.Choice) (profile.Brief, profile.TutorMode, error) {
	if again, asked := oneMoreLikeIt(p); asked && !own.Made() {
		if brief, mode, err := tutor.Ahead(p, s.content, again); err == nil {
			return brief, mode, nil
		}
	}
	return tutor.Ahead(p, s.content, own)
}

// oneMoreLikeIt is the choice of one more task like the one on the card, when
// a person asked for where that one stands.
func oneMoreLikeIt(p *profile.Profile) (tutor.Choice, bool) {
	task := p.CurrentTask
	if task == nil || !task.Asked {
		return tutor.Choice{}, false
	}
	return tutor.Choice{
		Topic: task.Topic, GradeLevel: task.GradeLevel, Difficulty: task.Difficulty,
		Reason: "One more like the task the adult asked for, in case another like it is wanted.",
	}, true
}

// aheadLead is what the model is told of the request for the task written
// ahead: what it is for, and that nothing is said of it.
func aheadLead(request *profile.OpenRequest) string {
	return fmt.Sprintf("Request %[1]s is open for the next task, written ahead: the child is working on the task on "+
		"the card, and this one is kept until another is asked for. Write it now, to the package below, and hand it "+
		"in with submit_task and request_id %[1]s. Say nothing about it, now or once it is kept.",
		request.ID)
}

// packageWords are the words prepare_task hands the model the package of a
// request in: the lead, what the task is written in, the last answer, and the
// package.
func (s *Service) packageWords(p *profile.Profile, request *profile.OpenRequest, lead string) (Reply[any], error) {
	pack, err := s.packageFor(p, request)
	if err != nil {
		return Reply[any]{}, err
	}
	return Reply[any]{
		Text: joined(lead, writtenInText(request), lessonLanguageText(&p.Student), s.lastAnswerText(p)) +
			packageText(pack),
	}, nil
}

// prepareRefused is a call no request can be opened from, told argument by
// argument. Nothing is written.
func prepareRefused(p *profile.Profile, problems []profile.Problem) Reply[any] {
	out, lines := problemsOf(problems)
	return Reply[any]{
		Text: "Nothing was prepared. Fix these arguments and call prepare_task again: " + lines + ".",
		Payload: preparedRefusedOut{
			Status: statusRejected, Code: codeInvalidArguments, Problems: out, LastAnswer: lastAnswerOf(p),
		},
	}
}
