package profile

import (
	"errors"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ErrNoRequest means no task is being written: there is no request for an
// attempt to count against or a task to answer.
var ErrNoRequest = errors.New("profile: no request open")

// TaskIDFor is the id of the task written for a request: the request's own,
// marked as a task's. A card that waits for a request finds its task by it, so
// the file need not say which request a task was written for.
func TaskIDFor(requestID string) string {
	return "tsk_" + strings.TrimPrefix(requestID, "req_")
}

// TaskState is how the task written for a request stands, as a card that
// waits for that request sees it.
type TaskState string

const (
	// TaskBeingWritten is a request still waited for: its task has not been
	// handed out yet.
	TaskBeingWritten TaskState = "being_written"
	// TaskKept is the request's task written ahead and kept ready: it comes to
	// a card when the child asks for the next task.
	TaskKept TaskState = "kept"
	// TaskOnTheCard is the request's task on the child's card, answered or not.
	TaskOnTheCard TaskState = "on_the_card"
	// TaskNotComing is a request whose task will not be on the card: it ran out
	// of attempts or of time, a newer request replaced it, its task kept ready
	// no longer fitted the lesson, or its task has left the card.
	TaskNotComing TaskState = "not_coming"
)

// TaskFor says how the task written for a request stands, and, while it is
// being written, how many of its attempts the checks have turned down. A task
// kept or handed out closes its request, so the task is looked for first: it
// is all that is left of the request.
func (p *Profile) TaskFor(requestID string, window time.Duration, now time.Time) (state TaskState, refused int) {
	if task := p.CurrentTask; task != nil && task.ID == TaskIDFor(requestID) {
		return TaskOnTheCard, 0
	}
	if ready := p.ReadyTask; ready != nil && ready.ID == TaskIDFor(requestID) {
		return TaskKept, 0
	}
	if request := p.OpenRequest; request != nil && request.ID == requestID && request.Awaited(window, now) {
		return TaskBeingWritten, request.Attempts
	}
	return TaskNotComing, 0
}

// Awaited reports whether a task written to this request is still waited for:
// the request was opened within the window, and it has an attempt left. A
// request no longer awaited is over, though nothing closes it: the next
// request replaces it, and a task handed in for it is refused as one for a
// request that is over.
//
// Two kinds of request are over that no clock of this service leaves behind,
// only a hand editing the file: one dated further ahead than the window, which
// would otherwise hold every new request back until its date came round, and
// one with no attempt left, which the last refusal closes.
func (r *OpenRequest) Awaited(window time.Duration, now time.Time) bool {
	age := now.Sub(r.OpenedAt.Time)
	return r.Attempts < MaxAttempts && -window <= age && age < window
}

// Ask records that the model has been asked for a task: the request it hands
// the task back against, with the brief it was given, the language the task
// is to be written in and nothing handed back yet. A request open before it is
// replaced, since the caller has decided nobody waits for it any more; and a
// task kept ready is let go, since asking for one to be written means the kept
// one is not to be handed out.
func (p *Profile) Ask(brief *Brief, mode TutorMode, language string, now time.Time) *OpenRequest {
	p.ReadyTask = nil
	p.OpenRequest = &OpenRequest{
		Brief:     *brief,
		ID:        "req_" + uuid.NewString(),
		Language:  language,
		OpenedAt:  At(now),
		TutorMode: mode,
	}
	return p.OpenRequest
}

// Refuse records an attempt the checks turned down, and reports whether it was
// the last one. The last closes the request with nothing handed to the child,
// and the day's count of failed generations goes up: the fuse that stops a
// model which cannot satisfy the checks from asking again for ever. The count
// of accepted tasks stays where it is — a refused attempt is not a task.
func (p *Profile) Refuse(now time.Time) (exhausted bool, err error) {
	if p.OpenRequest == nil {
		return false, ErrNoRequest
	}
	p.OpenRequest.Attempts++
	if p.OpenRequest.Attempts < MaxAttempts {
		return false, nil
	}
	p.OpenRequest = nil
	p.CountFailed(now)
	return true, nil
}

// Written is a task as the model wrote it and the checks accepted it: what the
// child is shown, the sketch that keeps it from coming back, and the version
// of the instructions it was written to. Where it stands — its topic, its level
// and difficulty and its language — is not among it: that is what the request
// asked for, and the checks refuse a task that asked for anything else.
type Written struct {
	Wording     string
	Drawing     string
	Options     map[string]string
	Hint        string
	Fingerprint string
	// InstructionsVersion is the version of the instructions the model had.
	InstructionsVersion string
}

// Issue hands the task written for the open request to the child. It becomes
// the task on the card, standing where the request asked, with everything that
// gives its answer away sealed, and under the id the request gives it; the
// request is closed, and the task is handed out as handOut says.
//
// A task still on the card — which only a file edited by hand holds beside an
// open request the child waits for, since asking for a task takes the one
// before it off the card — is replaced: when it has no answer it is recorded as
// skipped, like any other task left without one, and when it has one there is
// nothing left to record.
//
// Nothing changes if the task cannot be sealed: a task on the card with its
// answer in the open is worse than none.
func (p *Profile) Issue(written *Written, secret TaskSecret, sealer Sealer, now time.Time) (*CurrentTask, error) {
	request := p.OpenRequest
	if request == nil {
		return nil, ErrNoRequest
	}

	task := &CurrentTask{
		Asked:               request.Asked,
		Difficulty:          request.Brief.Difficulty,
		Drawing:             written.Drawing,
		Fingerprint:         written.Fingerprint,
		GradeLevel:          request.Brief.GradeLevel,
		Hint:                written.Hint,
		ID:                  TaskIDFor(request.ID),
		InstructionsVersion: written.InstructionsVersion,
		IssuedAt:            At(now),
		Language:            request.Language,
		Options:             maps.Clone(written.Options),
		TakenAfter:          request.TakenAfter,
		Topic:               request.Brief.TargetConcept,
		TutorMode:           request.TutorMode,
		Wording:             written.Wording,
	}
	sealed, err := sealSecret(sealer, secret, unansweredBinding(p.StudentID, task.ID))
	if err != nil {
		return nil, err
	}
	task.Sealed = sealed

	p.OpenRequest = nil
	p.handOut(task, now)
	return p.CurrentTask, nil
}

// handOut puts a task on the child's card. The task left there without an
// answer is recorded as skipped. The new task's fingerprint joins those of the
// tasks already given, the oldest leaving when there are too many; its topic
// records the day; and the day's count of accepted tasks — the unit of the
// daily limit — goes up.
func (p *Profile) handOut(task *CurrentTask, now time.Time) {
	if flying := p.InFlight(); flying != nil {
		p.leave(flying, now)
	}
	p.CurrentTask = task
	p.TaskFingerprints = append(p.TaskFingerprints, task.Fingerprint)
	if extra := len(p.TaskFingerprints) - MaxFingerprints; extra > 0 {
		p.TaskFingerprints = slices.Clone(p.TaskFingerprints[extra:])
	}
	topic := p.Topics[task.Topic]
	topic.LastIssued = DateOf(now)
	p.putTopic(task.Topic, &topic)
	p.CountAccepted(now)
}

// Skip takes the task off the card: the child asked for another. A task left
// without an answer goes into the window as skipped, at this moment, and into
// the count of its topic — nobody is waiting for its answer any more. Nothing
// that learns from answers reads the entry, and nothing else about the child
// moves: there was no answer to learn from. A task that has had its answer
// leaves with nothing recorded, since its answer is in the window already. It
// reports the entry, and false when no task was waiting for an answer.
//
// The task's fingerprint stays among those of the tasks given, so the task
// does not come back; the generation it cost stays counted, since the task did
// reach the child.
func (p *Profile) Skip(now time.Time) (Answer, bool) {
	task := p.InFlight()
	p.CurrentTask = nil
	if task == nil {
		return Answer{}, false
	}
	return p.leave(task, now), true
}

// DiscardTask takes the task off the card and records nothing about it. It is
// for a task whose seal can no longer be opened — the key that sealed it has
// been retired — which cannot be checked, so it cannot be answered; and a skip
// would tell the parent the child leafed past a task the service lost. The
// task's fingerprint and the generation it cost stay, as they do for a skip.
func (p *Profile) DiscardTask() {
	p.CurrentTask = nil
}

// leave records a task as left without an answer, at this moment: an entry of
// the window marked skipped, and one more in the count of its topic.
func (p *Profile) leave(task *CurrentTask, now time.Time) Answer {
	entry := Answer{
		AnsweredAt: At(now),
		Difficulty: task.Difficulty,
		GradeLevel: task.GradeLevel,
		Skipped:    true,
		TaskID:     task.ID,
		Topic:      task.Topic,
	}
	p.remember(&entry)
	topic := p.Topics[task.Topic]
	topic.Skipped++
	p.putTopic(task.Topic, &topic)
	return entry
}

// putTopic writes the summary of one topic back. A file that says null where
// the topics go is given a map to write into rather than a panic.
func (p *Profile) putTopic(id string, topic *Topic) {
	if p.Topics == nil {
		p.Topics = map[string]Topic{}
	}
	p.Topics[id] = *topic
}
