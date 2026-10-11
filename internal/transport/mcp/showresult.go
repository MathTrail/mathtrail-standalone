package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The refusals of show_result beyond a task that is not the one on the card.
const (
	// codeNotAnswered is the task on the card with no answer yet: nothing of
	// it can be shown, since everything there is to show gives the answer away.
	codeNotAnswered = "not_answered"
	// codeToldNoMore is a task whose answer is recorded and whose seal no
	// longer opens: the answer counts, and only telling how it went is gone.
	codeToldNoMore = "told_no_more"
)

// An answer the card recorded is written to the file after the card has its
// result, and the instance writing it may not be the one asked to show it, so
// a task found on the card without its answer is read again, a moment later,
// before it is told it has none.
const (
	// answerReads is how many times the profile is read for the answer.
	answerReads = 3
	// answerLandsIn is the pause before each read after the first.
	answerLandsIn = 700 * time.Millisecond
)

// showResultIn is what show_result takes: the task whose answer the card is
// to show.
type showResultIn struct {
	TaskID string `json:"task_id" jsonschema:"the id of the task on the child's card, once its answer is recorded"`
}

// shownOut is what the card of how an answer went is handed: how the answer
// went, as the card that took it shows it too — or why there is nothing to
// show, with whose card it is and in which language. A host may show the model
// the payload in place of the words, so the last answer is here as well.
type shownOut struct {
	Screen     string      `json:"screen"`
	Status     string      `json:"status,omitempty"`
	Code       string      `json:"code,omitempty"`
	LastAnswer *answerLine `json:"last_answer"`
	wentOut
}

// answeredTaskOut is the task as the card of its answer names it: its id and
// topic, the language it was written in, its five options, which the verdict
// names by their texts and which give nothing away once the answer is in, and
// its topic's page on the site, by the slug, with whether it is published.
type answeredTaskOut struct {
	ID       string            `json:"id"`
	Topic    string            `json:"topic"`
	Language string            `json:"language"`
	Options  map[string]string `json:"options"`
	Slug     string            `json:"slug"`
	SitePage bool              `json:"site_page"`
}

func (s *Service) showResultTool() Tool {
	return Define(Spec{
		Name:  "show_result",
		Title: "Show how the answer went",
		Description: "Draws the card of how the child's answer to the task on the card went, once the answer is " +
			"recorded by you with submit_answer: whether it was right and which option is, what went wrong on the " +
			"way to a wrong one and whether that mistake has come up before, the solution step by step with its " +
			"picture, the topic, and how the child's rating in the topic moved — during the trial " +
			"series, how many of its tasks are done instead — with the buttons for another task and for the topic. " +
			"An answer given on the card turns that card into the same. Call it with the task's id. A task with no " +
			"answer yet shows nothing of it, and a " +
			"task left behind by the next one shows no more than whether its answer was right. It reads the " +
			"profile's file in the adult's Google Drive and changes nothing.",
		Effect:     Reads,
		Idempotent: true,
		DrawsCard:  true,
	}, s.showResult)
}

// showResult tells how the answer to the task on the card went, on a card of
// its own: to the card, which shows the whole of it, and to the model, which
// adds a sentence at most beside it, or explains it in words where no card is
// shown. Nothing is written.
func (s *Service) showResult(ctx context.Context, account store.Account, in showResultIn) (Reply[shownOut], error) {
	p, recorded, err := s.toldOnceWritten(ctx, account, in.TaskID)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[shownOut]{
			Text:    "There is no answer to show: there is no task. " + firstRunText,
			Payload: shownOut{Screen: screenFirstRun, Status: statusStale, Code: codeStaleTask},
		}, nil
	case errors.Is(err, profile.ErrNoTask), errors.Is(err, profile.ErrOtherTask):
		return s.notShown(p, in.TaskID), nil
	case errors.Is(err, profile.ErrNotAnswered):
		return s.notAnswered(p), nil
	case errors.Is(err, profile.ErrSealed):
		return s.toldNoMore(p), nil
	case err != nil:
		return Reply[shownOut]{}, fmt.Errorf("mcp: tell how the answer went: %w", err)
	}
	noteTaskRequest(ctx, profile.RequestIDFor(p.CurrentTask.ID))
	return s.shown(p, &recorded)
}

// toldOnceWritten reads the profile and tells how the answer to the task went,
// reading again a moment later while the task on the card has no answer, until
// the reads are spent. A call given up on ends the wait.
func (s *Service) toldOnceWritten(ctx context.Context, account store.Account, taskID string) (*profile.Profile, profile.Recorded, error) {
	for read := 1; ; read++ {
		p, _, err := s.store.Load(ctx, account)
		if err != nil {
			return nil, profile.Recorded{}, err
		}
		recorded, err := p.Told(taskID, s.sealer)
		if !errors.Is(err, profile.ErrNotAnswered) || read == answerReads {
			return p, recorded, err
		}
		if err := sleep(ctx, s.answerPause); err != nil {
			return nil, profile.Recorded{}, err
		}
	}
}

// shown is how the answer went, as its card shows it and as the words tell it
// to the model.
func (s *Service) shown(p *profile.Profile, recorded *profile.Recorded) (Reply[shownOut], error) {
	choice, err := s.topicChoiceOf(p, s.now())
	if err != nil {
		return Reply[shownOut]{}, fmt.Errorf("mcp: the choice of the topic: %w", err)
	}
	went, repeated := s.wentOf(p, recorded, choice)
	return Reply[shownOut]{
		Text:    s.shownText(p.CurrentTask, recorded, repeated),
		Payload: shownOut{Screen: screenResult, LastAnswer: lastAnswerOf(p), wentOut: went},
	}, nil
}

// answeredTaskOf is the task on the card as the card of its answer names it.
func (s *Service) answeredTaskOf(task *profile.CurrentTask) *answeredTaskOut {
	out := &answeredTaskOut{ID: task.ID, Topic: task.Topic, Language: task.Language, Options: task.Options}
	if topic, found := s.content.Topic(task.Topic); found {
		out.Slug, out.SitePage = topic.Slug, topic.SitePage
	}
	return out
}

// shownText tells the model how the answer went, as the card below shows it,
// for it to add a sentence at most beside the card, or to explain it in full
// without one.
func (s *Service) shownText(task *profile.CurrentTask, recorded *profile.Recorded, repeated bool) string {
	lead := fmt.Sprintf("Where cards are shown, the card below shows how the answer to task %s went: the verdict, "+
		"the trap, the solution step by step and the rating. %s", task.ID, besideACard)
	return joined(lead, outcomeText(task, recorded, repeated), aboutTheStep, s.standingText(recorded)) +
		"\nSolution: " + quoted(recorded.Solution)
}

// notShown is a task that is not the one on the child's card: left for a newer
// one, skipped, or never given. Nothing of it is shown. A task whose answer the
// window keeps is told how that answer went, lest the model tell the child it
// was lost; the words never repeat an id that arrived without the profile
// knowing it.
func (s *Service) notShown(p *profile.Profile, taskID string) Reply[shownOut] {
	lead := "The task_id is not the task on the child's card: that task was skipped, left for a newer one or " +
		"never given. Nothing of it can be shown."
	if answer, found := p.AnswerTo(taskID); found {
		lead = fmt.Sprintf("Task %s has left the card since its answer, %s, was recorded: how it went can no "+
			"longer be shown.", answer.TaskID, howItWent(answer.Correct))
	}
	return Reply[shownOut]{
		Text:    joined(lead, s.showableNowText(p), s.lastAnswerText(p)),
		Payload: s.refusedOut(p, statusStale, codeStaleTask),
	}
}

// showableNowText says what is on the card now, and what of it can be shown.
func (s *Service) showableNowText(p *profile.Profile) string {
	switch task, request := p.CurrentTask, s.waitedOpen(p, s.now()); {
	case task == nil && request != nil:
		return fmt.Sprintf("There is no task on the card yet: request %s is open, and its task is still to be "+
			"handed in with submit_task.", request.ID)
	case task == nil:
		return "There is no task on the card: ask for a new one with next_task when another is asked for."
	case task.Answered != nil:
		return fmt.Sprintf("The task on the card, %s, has its answer: the card that took it shows how it went, and "+
			"for an answer given in the chat show_result with its id draws that card.", task.ID)
	}
	return fmt.Sprintf("Task %s is on the card, with no answer yet.", p.CurrentTask.ID)
}

// notAnswered is the task on the card before its answer: there is nothing to
// show yet, and nothing of the task is shown. An answer the card has only just
// recorded may not have reached the file even after the wait, and the model is
// told so, lest it tell the adult the child has not answered.
func (s *Service) notAnswered(p *profile.Profile) Reply[shownOut] {
	return Reply[shownOut]{
		Text: fmt.Sprintf("Task %s on the card has no answer yet, so there is nothing to show. Call show_result once "+
			"the child has answered: on the card, which records the answer itself, or in the chat, recorded with "+
			"submit_answer. An answer you have only just recorded with submit_answer may still be on its way into "+
			"the file: then, and only then, call show_result again in a moment.", p.CurrentTask.ID),
		Payload: s.refusedOut(p, statusRejected, codeNotAnswered),
	}
}

// toldNoMore is the task on the card with its answer recorded and its seal
// lost: the answer counts, and only how it went can no longer be told. Nothing
// is written: the next task asked for takes the card over.
func (s *Service) toldNoMore(p *profile.Profile) Reply[shownOut] {
	return Reply[shownOut]{
		Text: fmt.Sprintf("The answer to task %s is recorded and counts, but how it went can no longer be told: "+
			"say so in a sentence, and offer another task.", p.CurrentTask.ID),
		Payload: s.refusedOut(p, statusStale, codeToldNoMore),
	}
}

// refusedOut is the payload of a card with nothing to show: whose card it is,
// in the language of the task on the card when there is one, and why.
func (s *Service) refusedOut(p *profile.Profile, status, code string) shownOut {
	out := shownOut{
		Screen: screenResult, Status: status, Code: code, LastAnswer: lastAnswerOf(p),
		wentOut: wentOut{Child: childLineOf(&p.Student)},
	}
	if p.CurrentTask != nil {
		out.Language = p.CurrentTask.Language
	}
	return out
}
