package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// heldForNews is how long a card's question that would only hear what the card
// knows already waits for news. It is the card's pause between questions, so a
// card asking at its pace has a question waiting nearly all the time and hears
// of its task as the task lands, while it asks no more often than it did; and
// it is shorter than calls a host already carries.
const heldForNews = 4 * time.Second

// readTaskIn is what read_task takes: the request whose task a card waits for,
// and, once the card has heard the task is being written, how many tries
// turned down it has heard of. Absent, the question is answered at once, as a
// card drawn before the question could wait asks it.
type readTaskIn struct {
	RequestID string `json:"request_id" jsonschema:"the id of the request the card waits for the task of"`
	Refused   *int   `json:"refused,omitempty" jsonschema:"how many tries turned down the card has heard of, sent once it has heard the task is being written: while the file says the same, the answer waits a few seconds for news"`
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
		Description: fmt.Sprintf("How the task a card drawn by next_task waits for stands: still being written, "+
			"on the child's card, or not coming. Given the tries turned down the card has heard of, a task still "+
			"being written with as many is told of only once something changes, or after %d seconds. Only the card "+
			"calls it.", int(heldForNews/time.Second)),
		Effect:     Reads,
		Idempotent: true,
		WidgetOnly: true,
	}, s.readTask)
}

// readTask tells a card how the task it waits for stands. A card that says
// what it has heard already is not told it again at once: the question waits
// for a write of the account on this instance — the task accepted, a try
// turned down — and answers from the profile the write left. A write that
// leaves the task as the card knows it keeps the question waiting for the
// rest of its hold, and with no news by its end the question answers from the
// file as it read it when it came; the card's next question reads the file
// afresh. It listens before it reads, so a write that lands while the file is
// read is heard. The file is not read again when news comes: a read straight
// after a write may lag behind it, and once a read has lagged the store no
// longer holds the next read to the write, which a write made from that read
// would then undo.
func (s *Service) readTask(ctx context.Context, account store.Account, in readTaskIn) (Reply[awaitedOut], error) {
	stands := func() (Reply[awaitedOut], error) {
		return afresh(ctx, func() (Reply[awaitedOut], error) { return s.taskAwaited(ctx, account, in.RequestID) })
	}
	if in.Refused == nil {
		return stands()
	}
	heard, stop := s.news.listen(account.ID)
	//nolint:gocritic // stop is set anew each time news wakes the question, and the latest is the one to call.
	defer func() { stop() }()
	told, err := stands()
	if err != nil || !heardAlready(&told.Payload, *in.Refused) {
		return told, err
	}
	hold := time.NewTimer(s.hold)
	defer hold.Stop()
	for {
		written, woken := newsBefore(ctx, heard, hold.C)
		if !woken || written == nil {
			return told, nil
		}
		p, err := store.ParseAt(written, s.now())
		if err != nil {
			// The profile written could not be read back: the card hears what
			// it knew, and of the write at its next question.
			return told, nil
		}
		now, err := s.taskStands(ctx, p, in.RequestID)
		if err != nil || !heardAlready(&now.Payload, *in.Refused) {
			return now, err
		}
		told = now
		stop()
		heard, stop = s.news.listen(account.ID)
	}
}

// heardAlready says whether an answer tells the card only what it has heard:
// the task still being written, with as many tries turned down.
func heardAlready(told *awaitedOut, refused int) bool {
	return told.Screen == screenComing && told.Refused == refused
}

// newsBefore waits for the news the ear is open for, unless the hold or the
// call ends first, and is the profile the news brought, written down, and
// whether the news came.
func newsBefore(ctx context.Context, heard <-chan []byte, held <-chan time.Time) ([]byte, bool) {
	select {
	case written := <-heard:
		return written, true
	case <-held:
		return nil, false
	case <-ctx.Done():
		return nil, false
	}
}

// taskAwaited is one read of the profile, and how the task of the request
// asked about stands in it. Nothing is written.
func (s *Service) taskAwaited(ctx context.Context, account store.Account, requestID string) (Reply[awaitedOut], error) {
	p, _, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[awaitedOut]{Text: firstRunText, Payload: awaitedOut{Screen: screenFirstRun}}, nil
	case err != nil:
		return Reply[awaitedOut]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}
	return s.taskStands(ctx, p, requestID)
}

// taskStands is how the task of the request asked about stands in the
// profile. A task kept ready comes on the card next_task draws when the child
// asks for the next task, never to one waiting for the request it was written
// for — no card waits for a request written ahead — so such a card is told
// none is coming, as for a request that is over.
func (s *Service) taskStands(ctx context.Context, p *profile.Profile, requestID string) (Reply[awaitedOut], error) {
	awaited := awaitedOut{LastAnswer: lastAnswerOf(p), Child: childLineOf(&p.Student)}
	switch state, refused := p.TaskFor(requestID, s.window, s.now()); state {
	case profile.TaskOnTheCard:
		noteTaskRequest(ctx, requestID)
		task := p.CurrentTask
		choice, err := s.topicChoiceOf(p, s.now())
		if err != nil {
			return Reply[awaitedOut]{}, fmt.Errorf("mcp: the choice of the topic: %w", err)
		}
		awaited.Screen, awaited.Task, awaited.Language, awaited.TopicChoice = screenTask, cardOf(task), task.Language, choice
		return Reply[awaitedOut]{Text: fmt.Sprintf("Task %s is on the child's card.", task.ID), Payload: awaited}, nil
	case profile.TaskBeingWritten:
		noteTaskRequest(ctx, requestID)
		awaited.Screen, awaited.Refused, awaited.Language = screenComing, refused, p.OpenRequest.Language
		return Reply[awaitedOut]{Text: "The task is still being written.", Payload: awaited}, nil
	}
	awaited.Screen, awaited.Code = screenWaiting, codeStaleRequest
	return Reply[awaitedOut]{Text: "No task is coming for this request: it is over.", Payload: awaited}, nil
}
