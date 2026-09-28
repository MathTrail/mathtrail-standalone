// Package lesson is the lesson as the load drives it: the steps the chat's
// model and a child take through the tools, the tasks the load hands in as a
// model would write them, and the words of the service read the way a model
// reads them.
package lesson

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// Student is who a child of the load is on their card.
type Student struct {
	Pseudonym string
	Grade     int
}

// Choice is what a task is asked for: its topic, level and difficulty, and
// the reason a model gives for choosing them rather than leaving it to the
// rule.
type Choice struct {
	Topic      string
	GradeLevel string
	Difficulty int
	Reason     string
}

// Request is a request the service opened: its id, and the brief a task
// written for it hands back as it was received.
type Request struct {
	ID    string
	Brief json.RawMessage
}

// Card is a task on a child's card, as submit_task handed it.
type Card struct {
	TaskID    string
	Question  string
	Pseudonym string
	Attempt   int
}

// language is the language every task of the load is written in.
const language = "en"

// interest is what every child of the load is interested in. A brief's
// setting is made of the child's interests, and a brief with none would be
// handed back refused.
const interest = "sport"

// Start opens the child's profile: it reads what is there, and saves the
// student, whom the card will show.
func Start(ctx context.Context, child *session.Child, student Student) []session.Answer {
	return []session.Answer{
		child.Call(ctx, "get_profile", map[string]any{}),
		child.Call(ctx, "save_profile", map[string]any{
			"pseudonym": student.Pseudonym,
			"grade":     student.Grade,
			"interests": []string{interest},
		}),
	}
}

// Ask asks for a task of the choice given, and reads the request the words of
// the answer open. A request that cannot be read is no request, and the answer
// that failed to give one is still the answer the service gave.
func Ask(ctx context.Context, child *session.Child, choice Choice) (Request, session.Answer, error) {
	answer := child.Call(ctx, "next_task", map[string]any{
		"language":    language,
		"topic":       choice.Topic,
		"grade_level": choice.GradeLevel,
		"difficulty":  choice.Difficulty,
		"reason":      choice.Reason,
	})
	if answer.Kind != session.Answered {
		return Request{}, answer, fmt.Errorf("lesson: next_task answered %s", answer.Kind)
	}
	request, err := ReadRequest(answer.Text())
	return request, answer, err
}

// HandIn hands a task in for a request, and reads the card it went onto. A
// card that shows another question or another child belongs to another call,
// whatever the service said of it, and the answer is told as a mismatch.
func HandIn(ctx context.Context, child *session.Child, request Request, task *Task, student Student) (Card, session.Answer) {
	answer := child.Call(ctx, "submit_task", map[string]any{
		"request_id": request.ID,
		"brief":      request.Brief,
		"task":       task.Body,
		"solver":     task.Solver,
		"self_check": task.SelfCheck,
	})
	if answer.Kind != session.Answered {
		return Card{}, answer
	}

	var handedIn struct {
		Attempt int `json:"attempt"`
		Child   *struct {
			Pseudonym string `json:"pseudonym"`
		} `json:"child"`
		Task *struct {
			ID       string `json:"id"`
			Question string `json:"question"`
		} `json:"task"`
	}
	if err := answer.Payload(&handedIn); err != nil || handedIn.Task == nil || handedIn.Child == nil {
		answer.Kind = session.Mismatched
		return Card{}, answer
	}
	card := Card{
		TaskID:    handedIn.Task.ID,
		Question:  handedIn.Task.Question,
		Pseudonym: handedIn.Child.Pseudonym,
		Attempt:   handedIn.Attempt,
	}
	if card.Question != task.Body.Question || card.Pseudonym != student.Pseudonym {
		answer.Kind = session.Mismatched
	}
	return card, answer
}

// AnswerTask answers the task on the card with the letter given, as the child
// would from the card.
func AnswerTask(ctx context.Context, child *session.Child, card Card, letter string) session.Answer {
	return child.Call(ctx, "submit_answer", map[string]any{"task_id": card.TaskID, "answer": letter})
}

// Progress reads the child's progress.
func Progress(ctx context.Context, child *session.Child) session.Answer {
	return child.Call(ctx, "get_progress", map[string]any{})
}

// packageMark is where the words of next_task end and the package the task is
// written from begins.
const packageMark = "\n\nPackage:\n"

// requestID is the id of a request, as the words of next_task name it.
var requestID = regexp.MustCompile(`\breq_[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)

// ReadRequest reads the request the words of next_task open, as a model reads
// them: the id the words before the package name, and the brief of the package
// after them.
func ReadRequest(words string) (Request, error) {
	lead, pack, found := strings.Cut(words, packageMark)
	if !found {
		return Request{}, errors.New("lesson: the words carry no package")
	}
	id := requestID.FindString(lead)
	if id == "" {
		return Request{}, errors.New("lesson: the words name no request")
	}
	var parts struct {
		Brief json.RawMessage `json:"brief"`
	}
	if err := json.Unmarshal([]byte(pack), &parts); err != nil {
		return Request{}, fmt.Errorf("lesson: read the package: %w", err)
	}
	if len(parts.Brief) == 0 || string(parts.Brief) == "null" {
		return Request{}, errors.New("lesson: the package carries no brief")
	}
	return Request{ID: id, Brief: parts.Brief}, nil
}
