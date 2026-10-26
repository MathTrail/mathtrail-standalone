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
	"time"

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

// interest is what every child of the load is interested in: the setting of
// one brief in three is made of it.
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
	answer := child.Call(ctx, "read_task", map[string]any{"request_id": request.ID})
	if answer.Kind != session.Answered {
		return answer
	}
	var told awaited
	if err := answer.Payload(&told); err != nil || !told.writing() {
		answer.Kind = session.Mismatched
	}
	return answer
}

// awaited is as much of what a card is told of the task it waits for as says
// how it stands.
type awaited struct {
	Screen  string `json:"screen"`
	Refused int    `json:"refused"`
	Task    *struct {
		ID       string `json:"id"`
		Question string `json:"question"`
	} `json:"task"`
}

// writing says whether a card was told its task is still being written.
func (a *awaited) writing() bool { return a.Screen == "coming" && a.Task == nil }

// shows says whether a card was told it shows the task the hand-in put on it.
func (a *awaited) shows(card Card) bool {
	return a.Screen == "task" && a.Task != nil && a.Task.ID == card.TaskID && a.Task.Question == card.Question
}

// writingAfterLanding is how many answers that the task is still being
// written, to questions asked once the hand-in has landed, the held question
// takes before it settles: a read straight after a write may lag behind it
// once, and no card waits past a second.
const writingAfterLanding = 2

// CardPace is how often a card asks how its task stands, from the start of
// one question to the start of the next, a question held for news included.
const CardPace = 4 * time.Second

// HeldFor is how long the service holds a card's question that would only
// hear what the card knows already: a call has to be given longer.
const HeldFor = 4 * time.Second

// TriesTurnedDown is how many tries the checks have turned down, as an answer
// that the task is still being written tells them, and none as any other
// answer does: what a card says it has heard of on its next question.
func TriesTurnedDown(answer *session.Answer) int {
	told, _ := stillWriting(answer)
	return told.Refused
}

// stillWriting is what an answer that the task is still being written tells a
// card, and whether the answer is one.
func stillWriting(answer *session.Answer) (awaited, bool) {
	var told awaited
	if answer.Kind != session.Answered || answer.Payload(&told) != nil || !told.writing() {
		return awaited{}, false
	}
	return told, true
}

// Held is the card's question once the card has heard its task is being
// written, as the card asks it: saying the tries turned down it has heard of,
// so that the service holds it while the task is still being written with as
// many and answers as the hand-in lands, or as its hold ends, and asked again,
// at the card's pace, while the answer is the task still being written. It
// runs beside the hand-in.
type Held struct {
	stop     context.CancelFunc
	landed   chan struct{}
	landedAt time.Time
	done     chan outcome
}

// outcome is what the held question was told, and whether it settled on its
// last answer rather than being cut short.
type outcome struct {
	answers []session.Answer
	settled bool
}

// Hold starts the card's held question for the task of a request, the card
// having heard of heard tries turned down, asking no more often than once a
// pace from the start of each question.
func Hold(ctx context.Context, child *session.Child, request Request, heard int, pace time.Duration) *Held {
	ours, stop := context.WithCancel(ctx)
	h := &Held{stop: stop, landed: make(chan struct{}), done: make(chan outcome, 1)}
	go func() { h.done <- h.ask(ctx, ours, child, request, heard, pace) }()
	return h
}

// ask asks the held question until it settles: on an answer other than the
// task still being written, or on writingAfterLanding of those to questions
// asked once the hand-in has landed. A question it was let go of in the middle
// of is not among the answers; one the run's end cut short is, as the run's
// other calls are.
func (h *Held) ask(run, ours context.Context, child *session.Child, request Request, heard int, pace time.Duration) outcome {
	var answers []session.Answer
	afterLanding := 0
	for {
		answer := child.Call(ours, "read_task", map[string]any{"request_id": request.ID, "refused": heard})
		if answer.Kind.Class == session.Stopped {
			if run.Err() != nil {
				answers = append(answers, answer)
			}
			return outcome{answers: answers}
		}
		answers = append(answers, answer)
		told, writing := stillWriting(&answer)
		if !writing {
			return outcome{answers: answers, settled: true}
		}
		heard = max(heard, told.Refused)
		if h.landedBy(answer.Started) {
			afterLanding++
			if afterLanding == writingAfterLanding {
				return outcome{answers: answers, settled: true}
			}
		}
		// A card asks no more often than its pace, so a question answered at
		// once waits the rest of it before the next.
		if !waitUntil(ours, answer.Started.Add(pace)) {
			return outcome{answers: answers}
		}
	}
}

// waitUntil waits for the moment given, and says whether it came before the
// context was done.
func waitUntil(ctx context.Context, moment time.Time) bool {
	timer := time.NewTimer(time.Until(moment))
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// landedBy says whether the hand-in had landed by the moment given.
func (h *Held) landedBy(at time.Time) bool {
	select {
	case <-h.landed:
		return !at.Before(h.landedAt)
	default:
		return false
	}
}

// Card is every answer the held question was told once the hand-in has put
// card on the child's card: the task still being written, until the answer it
// settled on, which has to show card's task. Any other answer it settled on
// belongs to another call, and is told as a mismatch; a question the run's end
// cut short is left as it is.
func (h *Held) Card(card Card) []session.Answer {
	h.landedAt = time.Now()
	close(h.landed)
	told := <-h.done
	h.stop()
	if last := len(told.answers) - 1; told.settled && last >= 0 && told.answers[last].Kind == session.Answered {
		var on awaited
		if err := told.answers[last].Payload(&on); err != nil || !on.shows(card) {
			told.answers[last].Kind = session.Mismatched
		}
	}
	return told.answers
}

// Drop lets the held question go, the hand-in having put no task on the card,
// and is the answers it was told before.
func (h *Held) Drop() []session.Answer {
	h.stop()
	return (<-h.done).answers
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
