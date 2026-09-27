package profile

import (
	"errors"
	"maps"
	"slices"
	"time"

	"github.com/google/uuid"
)

// ErrNoRequest means no task is being written: there is no request for an
// attempt to count against or a task to answer.
var ErrNoRequest = errors.New("profile: no request open")

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
// replaced, since the caller has decided nobody waits for it any more.
func (p *Profile) Ask(brief *Brief, mode TutorMode, language string, now time.Time) *OpenRequest {
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
// the task in flight, standing where the request asked, with everything that
// gives its answer away sealed; the request is closed. The task's fingerprint
// joins those of the tasks already given, the oldest leaving when there are
// too many, the topic records the day, and the day's count of accepted tasks —
// the unit of the daily limit — goes up.
//
// A task still in flight — which only a file edited by hand holds beside an
// open request, since asking for a task skips the one before it — is recorded
// as skipped: it leaves flight without an answer, like any other.
//
// Nothing changes if the task cannot be sealed: a task in flight with its
// answer in the open is worse than none.
func (p *Profile) Issue(written *Written, secret TaskSecret, sealer Sealer, now time.Time) (*CurrentTask, error) {
	request := p.OpenRequest
	if request == nil {
		return nil, ErrNoRequest
	}

	flying := p.CurrentTask
	p.CurrentTask = &CurrentTask{
		Difficulty:          request.Brief.Difficulty,
		Drawing:             written.Drawing,
		Fingerprint:         written.Fingerprint,
		GradeLevel:          request.Brief.GradeLevel,
		Hint:                written.Hint,
		ID:                  "tsk_" + uuid.NewString(),
		InstructionsVersion: written.InstructionsVersion,
		IssuedAt:            At(now),
		Language:            request.Language,
		Options:             maps.Clone(written.Options),
		Topic:               request.Brief.TargetConcept,
		Wording:             written.Wording,
	}
	if err := p.SealTask(sealer, secret); err != nil {
		p.CurrentTask = flying
		return nil, err
	}

	if flying != nil {
		p.leave(flying, now)
	}
	p.OpenRequest = nil
	p.TaskFingerprints = append(p.TaskFingerprints, written.Fingerprint)
	if extra := len(p.TaskFingerprints) - MaxFingerprints; extra > 0 {
		p.TaskFingerprints = slices.Clone(p.TaskFingerprints[extra:])
	}
	topic := p.Topics[p.CurrentTask.Topic]
	topic.LastIssued = DateOf(now)
	p.putTopic(p.CurrentTask.Topic, &topic)
	p.CountAccepted(now)
	return p.CurrentTask, nil
}

// Skip records that the task in flight was left without an answer: the child
// asked for another. The task goes into the window as skipped, at this
// moment, and into the count of its topic, and it leaves flight — nobody is
// waiting for its answer any more. Nothing that learns from answers reads the
// entry, and nothing else about the child moves: there was no answer to learn
// from. It reports the entry, and false when no task was in flight.
//
// The task's fingerprint stays among those of the tasks given, so the task
// does not come back; the generation it cost stays counted, since the task did
// reach the child.
func (p *Profile) Skip(now time.Time) (Answer, bool) {
	task := p.CurrentTask
	if task == nil {
		return Answer{}, false
	}
	p.CurrentTask = nil
	return p.leave(task, now), true
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
