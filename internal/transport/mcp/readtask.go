package mcpserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// readTaskIn is what read_task takes: the request whose task a card waits for.
type readTaskIn struct {
	RequestID string `json:"request_id" jsonschema:"the id of the request the card waits for the task of"`
}

// awaitedOut is how the task a card waits for stands: still being written,
// with how many of its attempts the checks have turned down; on the child's
// card, as the card shows it, with nothing that gives its answer away; or not
// coming, the request over. It never carries a status: a card asking how a
// task stands is answered, not refused, so the line its call leaves counts it
// among the answers.
type awaitedOut struct {
	Screen     string      `json:"screen"`
	Code       string      `json:"code,omitempty"`
	Refused    int         `json:"refused,omitempty"`
	LastAnswer *answerLine `json:"last_answer"`
	Child      *childLine  `json:"child"`
	Task       *cardOut    `json:"task"`
	// Language is the lesson's: the task's once it is on the card, and the
	// request's while it is being written.
	Language string `json:"language,omitempty"`
	// TopicChoice is what the card needs to keep the lessons to a topic, once
	// the task is on it and the trial series is over.
	TopicChoice *topicChoiceOut `json:"topic_choice,omitempty"`
}

func (s *Service) readTaskTool() Tool {
	return Define(Spec{
		Name:  "read_task",
		Title: "Read how the awaited task stands, for the card",
		Description: "How the task a card drawn by next_task waits for stands: still being written, on the " +
			"child's card, or not coming. Only the card calls it.",
		Effect:     Reads,
		Idempotent: true,
		WidgetOnly: true,
	}, s.readTask)
}

// readTask tells a card how the task it waits for stands.
func (s *Service) readTask(ctx context.Context, account store.Account, in readTaskIn) (Reply[awaitedOut], error) {
	return afresh(ctx, func() (Reply[awaitedOut], error) { return s.taskAwaited(ctx, account, in.RequestID) })
}

// taskAwaited is one read of the profile, and how the task of the request
// asked about stands in it. Nothing is written. A task kept ready comes on the
// card next_task draws when the child asks for the next task, never to one
// waiting for the request it was written for — no card waits for a request
// written ahead — so such a card is told none is coming, as for a request
// that is over.
func (s *Service) taskAwaited(ctx context.Context, account store.Account, requestID string) (Reply[awaitedOut], error) {
	p, _, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[awaitedOut]{Text: firstRunText, Payload: awaitedOut{Screen: screenFirstRun}}, nil
	case err != nil:
		return Reply[awaitedOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	awaited := awaitedOut{LastAnswer: lastAnswerOf(p), Child: childLineOf(&p.Student)}
	switch state, refused := p.TaskFor(requestID, s.window, s.now()); state {
	case profile.TaskOnTheCard:
		task := p.CurrentTask
		awaited.Screen, awaited.Task, awaited.Language = screenTask, cardOf(task), task.Language
		if awaited.TopicChoice, err = s.topicChoiceOf(p, s.now()); err != nil {
			return Reply[awaitedOut]{}, fmt.Errorf("mcp: the choice of the topic: %w", err)
		}
		return Reply[awaitedOut]{Text: fmt.Sprintf("Task %s is on the child's card.", task.ID), Payload: awaited}, nil
	case profile.TaskBeingWritten:
		awaited.Screen, awaited.Refused, awaited.Language = screenComing, refused, p.OpenRequest.Language
		return Reply[awaitedOut]{Text: "The task is still being written.", Payload: awaited}, nil
	}
	awaited.Screen, awaited.Code = screenWaiting, codeStaleRequest
	return Reply[awaitedOut]{Text: "No task is coming for this request: it is over.", Payload: awaited}, nil
}
