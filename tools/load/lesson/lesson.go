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

// Request is a request the service opened, by its id. The request keeps the
// brief its package carries, so a task is handed in without it.
type Request struct {
	ID string
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

// interest is what every child of the load is interested in, which the
// setting of a brief is made of.
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

// Ask asks for a task of the choice given, which draws the card it will come
// to, and then, as a model does at once, for the package to write it from, and
// reads the request out of the package's words. A request that cannot be read
// is no request, and the answers that failed to give one are still the answers
// the service gave; a package of another request than the card waits for
// belongs to another call, and is told as a mismatch.
func Ask(ctx context.Context, child *session.Child, choice Choice) (Request, []session.Answer, error) {
	asked := child.Call(ctx, "next_task", map[string]any{
		"language":    language,
		"topic":       choice.Topic,
		"grade_level": choice.GradeLevel,
		"difficulty":  choice.Difficulty,
		"reason":      choice.Reason,
	})
	if asked.Kind != session.Answered {
		return Request{}, []session.Answer{asked}, fmt.Errorf("lesson: next_task answered %s", asked.Kind)
	}
	var coming struct {
		Screen    string `json:"screen"`
		RequestID string `json:"request_id"`
	}
	if err := asked.Payload(&coming); err != nil || coming.Screen != "coming" || coming.RequestID == "" {
		asked.Kind = session.Mismatched
		return Request{}, []session.Answer{asked}, errors.New("lesson: next_task drew no card waiting for a request")
	}

	packed := child.Call(ctx, "get_package", map[string]any{"request_id": coming.RequestID})
	answers := []session.Answer{asked, packed}
	if packed.Kind != session.Answered {
		return Request{}, answers, fmt.Errorf("lesson: get_package answered %s", packed.Kind)
	}
	request, err := ReadRequest(packed.Text())
	if err == nil && request.ID != coming.RequestID {
		answers[1].Kind = session.Mismatched
		return Request{}, answers, errors.New("lesson: the package is of another request than the card waits for")
	}
	return request, answers, err
}

// AwaitWriting asks, as the card next_task drew does, how the task of a
// request stands while it is being written, and is what the card was told. An
// answer that says anything but that it is being written belongs to another
// call, and is told as a mismatch.
func AwaitWriting(ctx context.Context, child *session.Child, request Request) session.Answer {
	return await(ctx, child, request, func(a *awaited) bool { return a.Screen == "coming" && a.Task == nil })
}

// AwaitCard asks, as the card next_task drew does, how the task of a request
// stands once it was handed in, and is what the card was told. An answer that
// shows anything but the task the hand-in put on the card belongs to another
// call, and is told as a mismatch.
func AwaitCard(ctx context.Context, child *session.Child, request Request, card Card) session.Answer {
	return await(ctx, child, request, func(a *awaited) bool {
		return a.Screen == "task" && a.Task != nil && a.Task.ID == card.TaskID && a.Task.Question == card.Question
	})
}

// awaited is as much of what a card is told of the task it waits for as says
// how it stands.
type awaited struct {
	Screen string `json:"screen"`
	Task   *struct {
		ID       string `json:"id"`
		Question string `json:"question"`
	} `json:"task"`
}

// await asks how the task of a request stands, and holds what the card is told
// to what it should be.
func await(ctx context.Context, child *session.Child, request Request, should func(*awaited) bool) session.Answer {
	answer := child.Call(ctx, "read_task", map[string]any{"request_id": request.ID})
	if answer.Kind != session.Answered {
		return answer
	}
	var told awaited
	if err := answer.Payload(&told); err != nil || !should(&told) {
		answer.Kind = session.Mismatched
	}
	return answer
}

// HandIn hands a task in for a request, and reads the card it went onto. A
// card that shows another question or another child belongs to another call,
// whatever the service said of it, and the answer is told as a mismatch.
func HandIn(ctx context.Context, child *session.Child, request Request, task *Task, student Student) (Card, session.Answer) {
	answer := child.Call(ctx, "submit_task", map[string]any{
		"request_id": request.ID,
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

// ShowResult asks for the card of how the answer to the task on the card went,
// as the model does once the card has recorded the answer.
func ShowResult(ctx context.Context, child *session.Child, card Card) session.Answer {
	return child.Call(ctx, "show_result", map[string]any{"task_id": card.TaskID})
}

// Progress reads the child's progress.
func Progress(ctx context.Context, child *session.Child) session.Answer {
	return child.Call(ctx, "get_progress", map[string]any{})
}

// packageMark is where the words of get_package end and the package the task is
// written from begins.
const packageMark = "\n\nPackage:\n"

// requestID is the id of a request, as the words of get_package name it.
var requestID = regexp.MustCompile(`\breq_[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)

// ReadRequest reads the request the words of get_package hand the package of,
// as a model reads them: the id the words before the package name. A package
// with no brief after them is no package a task could be written to.
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
	return Request{ID: id}, nil
}
