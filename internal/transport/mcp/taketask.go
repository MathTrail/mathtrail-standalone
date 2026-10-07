package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// takeTaskIn is what take_task takes: the task on the card that asks for the
// next one.
type takeTaskIn struct {
	TaskID string `json:"task_id" jsonschema:"the id of the task on the card that asks for the next one"`
}

func (s *Service) takeTaskTool() Tool {
	return Define(Spec{
		Name:  "take_task",
		Title: "Take the next task, for the card",
		Description: "Takes the next task for the card of a task, when the child asks for another there: the task " +
			"written ahead and kept, on the card at once; or the task being written ahead, which the card then waits " +
			"for as for any task; or, with neither, a request opened by the rule for the model to write. The task " +
			"the card showed without an answer is recorded as skipped. Asked again after a task it brought, it " +
			"tells the same rather than skipping that one. Only the card calls it. It writes to the profile's file " +
			"in the adult's Google Drive the task it hands out, the request it opens or waits for, and the task it " +
			"records as skipped.",
		Effect:     Adds,
		Idempotent: true,
		WidgetOnly: true,
	}, s.takeTask)
}

// takeTask brings the card that asked the next task, or the wait for it.
func (s *Service) takeTask(ctx context.Context, account store.Account, in takeTaskIn) (Reply[requestOut], error) {
	return afresh(ctx, func() (Reply[requestOut], error) { return s.take(ctx, account, in.TaskID) })
}

// take is one read of the profile and the next task for the card that showed
// taskID: the one a take from that card brought already, its answer lost on
// the way; the task kept, handed out; the task being written ahead, waited
// for; or a request opened by the rule. A card whose task is no longer the
// one on the card is told its task is over.
func (s *Service) take(ctx context.Context, account store.Account, taskID string) (Reply[requestOut], error) {
	p, revision, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[requestOut]{Text: firstRunText, Payload: requestOut{Screen: screenFirstRun}}, nil
	case err != nil:
		return Reply[requestOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	now := s.now()
	if reply, taken, before := s.takenBefore(p, taskID, now); taken || before != nil {
		return reply, before
	}
	current := p.CurrentTask
	if current == nil || current.ID != taskID {
		return cardOver(p, taskID), nil
	}
	if limit, count, reached := s.daily.reached(p.Daily.Today(now)); reached {
		limitHit(ctx, s.events.logger, s.events.projectID, account.ID, limit, zap.Int("count", count))
		return s.dayIsFull(p), nil
	}

	lesson := s.lessonNow(p, current.Language, profile.Place{})
	gone := s.letGo(p, &lesson, now)
	var reply Reply[requestOut]
	switch {
	case s.aheadOpen(p, now) != nil:
		var request *profile.OpenRequest
		if request, _, err = s.waitForAhead(ctx, account, p, revision, taskID, false, now); err == nil {
			reply = takenComing(p, request)
		}
	case p.ReadyTask != nil:
		var task *profile.CurrentTask
		if task, err = s.handOutReady(ctx, account, p, revision, taskID, false, now); err == nil {
			reply, err = s.cardWith(p, task, now)
		}
	default:
		reply, err = s.askForTheCard(ctx, account, p, revision, taskID, now)
	}
	if err != nil {
		return Reply[requestOut]{}, err
	}
	s.sayLetGo(ctx, account, gone)
	return reply, nil
}

// takenBefore is what a take already made from the card that showed taskID
// brought, when that card asks again, the answer to its first ask lost on the
// way: the task it brought, on the card now, or the request it waits for. The
// card is told the same rather than having that task skipped. It reports
// whether there was such a take.
func (s *Service) takenBefore(p *profile.Profile, taskID string, now time.Time) (Reply[requestOut], bool, error) {
	if task := p.CurrentTask; task != nil && task.ID != taskID && task.TakenAfter == taskID {
		reply, err := s.cardWith(p, task, now)
		return reply, true, err
	}
	if open := s.waitedOpen(p, now); open != nil && open.TakenAfter == taskID {
		return takenComing(p, open), true, nil
	}
	return Reply[requestOut]{}, false, nil
}

// askForTheCard opens a request by the rule for the card that asked, with no
// task written ahead to give it: the task the card showed without an answer is
// recorded as skipped, and the card waits for the task the model writes. The
// lesson keeps its language: the parent's choice, or else that of the task the
// card showed.
func (s *Service) askForTheCard(ctx context.Context, account store.Account, p *profile.Profile, revision store.Revision,
	taskID string, now time.Time,
) (Reply[requestOut], error) {
	brief, mode, err := tutor.Next(p, s.content, tutor.Choice{})
	if err != nil {
		return Reply[requestOut]{}, fmt.Errorf("mcp: choose the task: %w", err)
	}
	language := p.Student.LessonLanguage(p.CurrentTask.Language)
	skipped, wasSkipped := p.Skip(now)
	request := p.Ask(&brief, mode, language, now)
	request.TakenAfter = taskID
	if err := s.canPackage(p, request); err != nil {
		return Reply[requestOut]{}, err
	}
	p.Touch(s.version, now)
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return Reply[requestOut]{}, fmt.Errorf("mcp: save the profile: %w", err)
	}
	if wasSkipped {
		s.events.write(ctx, account, eventTaskSkipped, s.skippedFields(skipped.Topic, skipped.GradeLevel, skipped.Difficulty)...)
	}
	s.events.write(ctx, account, eventTaskRequested, requestFields(request, false)...)
	return takenComing(p, request), nil
}

// cardWith is the card task is on now, handed out to the child, with what the
// card needs to keep the lessons to a topic.
func (s *Service) cardWith(p *profile.Profile, task *profile.CurrentTask, now time.Time) (Reply[requestOut], error) {
	choice, err := s.topicChoiceOf(p, now)
	if err != nil {
		return Reply[requestOut]{}, fmt.Errorf("mcp: the choice of the topic: %w", err)
	}
	return Reply[requestOut]{
		Text: fmt.Sprintf("Task %s is on the child's card.", task.ID),
		Payload: requestOut{
			Screen: screenTask, Task: cardOf(task), TopicChoice: choice, LastAnswer: lastAnswerOf(p),
			Child: childLineOf(&p.Student), Language: task.Language,
		},
	}, nil
}

// takenComing is the request a card that took the next task waits for, being
// written.
func takenComing(p *profile.Profile, request *profile.OpenRequest) Reply[requestOut] {
	return Reply[requestOut]{
		Text: fmt.Sprintf("The task of request %s is being written.", request.ID),
		Payload: requestOut{
			Screen: screenComing, RequestID: request.ID, LastAnswer: lastAnswerOf(p), Child: childLineOf(&p.Student),
			Language: request.Language,
		},
	}
}

// cardOver is what a card is told whose task is no longer the one on the
// card: a newer task took its place, asked for elsewhere, so the next one is
// not this card's to take. Nothing is written.
func cardOver(p *profile.Profile, taskID string) Reply[requestOut] {
	return Reply[requestOut]{
		Text: fmt.Sprintf("Task %s is no longer the one on the child's card.", taskID),
		Payload: requestOut{
			Screen: screenWaiting, Status: statusStale, Code: codeStaleTask, LastAnswer: lastAnswerOf(p),
			Child: childLineOf(&p.Student),
		},
	}
}
