package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// getPackageIn is what get_package takes: the request whose task is to be
// written.
type getPackageIn struct {
	RequestID string `json:"request_id" jsonschema:"the id of the request next_task opened"`
}

// packageRefusedOut is what get_package hands back when there is no package to
// hand: the one result of the tool with a payload, because its status is what
// marks it a refusal. Every other result is words alone: the package is
// written for the model and travels in the words, and a host that shows the
// model a result's payload in place of its words would otherwise hand it
// nothing to write the task from. The tool draws no card, so a payload would
// have no other reader.
type packageRefusedOut struct {
	Screen     string      `json:"screen"`
	Status     string      `json:"status"`
	Code       string      `json:"code"`
	LastAnswer *answerLine `json:"last_answer"`
}

func (s *Service) getPackageTool() Tool {
	return Define(Spec{
		Name:  "get_package",
		Title: "Get the package to write the task from",
		Description: "Returns the package to write the task of an open request from: the brief — the topic, the " +
			"level and the difficulty the rule sets — with reference tasks, and the page on how to write and hand in " +
			"a task. Call it with the request_id next_task gave, as soon as next_task has answered; write the task " +
			"to the package, and hand it in with submit_task and the same request_id. The package is for you alone: " +
			"show the child nothing of it.",
		ReadOnly:   true,
		Idempotent: true,
	}, s.getPackage)
}

// getPackage hands the model what to write the task of an open request from.
// Everything it says is in the words; only a refusal has a payload.
func (s *Service) getPackage(ctx context.Context, account store.Account, in getPackageIn) (Reply[any], error) {
	return afresh(ctx, func() (Reply[any], error) { return s.packageOfRequest(ctx, account, in.RequestID) })
}

// packageOfRequest is one read of the profile, and the package of the request
// asked about while it is the one open and still awaited. Nothing is written.
func (s *Service) packageOfRequest(ctx context.Context, account store.Account, requestID string) (Reply[any], error) {
	p, _, err := s.store.Load(ctx, account)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Reply[any]{Text: "No task has been asked for. " + firstRunText}, nil
	case err != nil:
		return Reply[any]{}, fmt.Errorf("mcp: read the profile: %w", err)
	}

	request, now := p.OpenRequest, s.now()
	if request == nil || request.ID != requestID || !request.Awaited(s.window, now) {
		return s.noPackage(p, now), nil
	}
	pack, err := s.packageFor(p, request)
	if err != nil {
		return Reply[any]{}, err
	}
	return Reply[any]{
		Text: joined(fmt.Sprintf("The package of request %s. Write one task in %s to it, and hand it in with "+
			"submit_task and request_id %s.", request.ID, request.Language, request.ID),
			lessonLanguageText(&p.Student), stillInText(p), forYouAlone) + packageText(pack),
	}, nil
}

// noPackage is the answer about a request that is not open, or not awaited any
// more — handed out, out of attempts, replaced or waited for too long — whose
// task there is nothing left to write for. The model is sent to what is waited
// for now: the request open, whose package it can have, the task the child is
// working on, or a new one to ask for.
func (s *Service) noPackage(p *profile.Profile, now time.Time) Reply[any] {
	lead := "The request_id is not the open request's: that request was handed out already, ran out of attempts, " +
		"was replaced by a newer one or waited too long, and there is no package for it."
	next := " Ask for a new task with next_task."
	switch task, open := p.InFlight(), p.OpenRequest; {
	case open != nil && open.Awaited(s.window, now):
		next = fmt.Sprintf(" Request %s is the open one, and the card waits for its task: get its package with "+
			"get_package and request_id %s.", open.ID, open.ID)
	case task != nil:
		next = fmt.Sprintf(" Task %s is on the child's card: wait for the child's answer to it.", task.ID)
	}
	return Reply[any]{
		Text:    lead + next,
		Payload: packageRefusedOut{Screen: screenWaiting, Status: statusStale, Code: codeStaleRequest, LastAnswer: lastAnswerOf(p)},
	}
}

// packageFor is what the model is handed to write the request's task from: the
// brief the request keeps, the chances around it for this child now, the child
// as the task is to be pitched at them, and how far through the topic's ideas
// the child's tasks of it have come.
func (s *Service) packageFor(p *profile.Profile, request *profile.OpenRequest) ([]byte, error) {
	topic := p.Topics[request.Brief.TargetConcept]
	pack, err := s.content.Package(&content.Request{
		Language:   request.Language,
		Brief:      request.Brief,
		Corridor:   tutor.CorridorIn(p, s.content, request.Brief.TargetConcept),
		Grade:      p.Student.Grade,
		Interests:  p.Student.Interests,
		Notes:      p.Student.Notes,
		Answers:    p.Ratings.Answers,
		TopicTasks: topic.Answers + topic.Skipped,
	})
	if err != nil {
		return nil, fmt.Errorf("mcp: build the package: %w", err)
	}
	return pack, nil
}

// packageText puts the package after the words about it, where the model reads
// it. It travels here alone and never in a payload, nor in the words of a tool
// that draws a card: a card is drawn from the whole of a result, and the
// package describes the child and the task to come.
func packageText(pack []byte) string {
	return "\n\nPackage:\n" + string(pack)
}
