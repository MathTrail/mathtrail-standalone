package profile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrSealed means the sealed part of the task on the card cannot be read: the
// key that sealed it has been retired, or it was sealed under a shape this
// build does not know. A task is worth less than a profile, so the answer is
// to tell the child this one can no longer be checked and give them another.
var ErrSealed = errors.New("profile: the sealed part of the task cannot be read")

// SecretVersion is the shape of the sealed part. It moves on its own: what is
// sealed lives as long as one task, and a task on the card is not worth a
// migration.
const SecretVersion = 1

// Sealer protects a value the service must keep to itself and reads it back.
// The profile declares what it needs rather than reaching for whoever
// provides it: this package computes, and what it computes with is handed in.
//
// The binding is what the value belongs to, and opening it again requires the
// same binding: without that, one child's sealed answer would open in another
// child's profile. The binding is made of what the profile itself says, so it
// holds against a value moved from one file to another, not against a file
// edited to say what another one says.
type Sealer interface {
	Seal(plaintext []byte, binding ...string) (string, error)
	Open(value string, binding ...string) ([]byte, error)
}

// Distractor is one wrong option: the mistake it comes from, and what to say
// to a child who picked it.
type Distractor struct {
	// Text explains the mistake in the child's own terms — not the solution
	// again, but what went wrong on the way to this option.
	Text string `json:"text"`
	// Trap is the catalog id of that mistake.
	Trap string `json:"trap"`
}

// TaskSecret is everything about a task that would give its answer away. It
// travels as one object rather than as a field per option on purpose: four
// explanations in the open would name the four wrong options and hand over
// the fifth by elimination.
type TaskSecret struct {
	// Answer is the letter of the correct option.
	Answer string `json:"answer"`
	// Distractors is what every other option leads to, keyed by letter.
	Distractors map[string]Distractor `json:"distractors"`
	// Solution is the worked answer the child is shown afterwards.
	Solution string `json:"solution"`
	// Solver is the program that confirmed exactly one option is right. It is
	// never run again: it is the record that this task was checked at all.
	Solver string `json:"solver"`
	// Version is the shape of this object.
	Version int `json:"version"`
}

// SealTask puts the part of the task on the card that gives its answer away
// beyond reach, tied to this child and this task, as it stands: unanswered,
// or answered with the letter it was given.
func (p *Profile) SealTask(sealer Sealer, secret TaskSecret) error {
	if p.CurrentTask == nil {
		return ErrNoTask
	}
	value, err := sealSecret(sealer, secret, p.binding())
	if err != nil {
		return err
	}
	p.CurrentTask.Sealed = value
	return nil
}

// sealSecret is a task's secret sealed against a binding.
func sealSecret(sealer Sealer, secret TaskSecret, binding []string) (string, error) {
	secret.Version = SecretVersion
	var plaintext bytes.Buffer
	encoder := json.NewEncoder(&plaintext)
	// Escaped for a web page, every <, > and & would take six bytes, and the
	// sealed value a third more again; nothing sealed is ever put in a page,
	// and the file carries the value whole.
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(secret); err != nil {
		return "", fmt.Errorf("profile: seal the task: %w", err)
	}
	value, err := sealer.Seal(plaintext.Bytes(), binding...)
	if err != nil {
		return "", fmt.Errorf("profile: seal the task: %w", err)
	}
	return value, nil
}

// OpenTask reads that part back. Only an answer to the task opens it — the
// first, which it judges, and any sent again, which it is told to — so it is
// never read before the child has answered, while the answer is still a
// secret. It opens against the task as it stands, so a task the file claims
// was answered opens only if its answer was recorded here, with that letter.
func (p *Profile) OpenTask(sealer Sealer) (TaskSecret, error) {
	if p.CurrentTask == nil {
		return TaskSecret{}, ErrNoTask
	}

	plaintext, err := sealer.Open(p.CurrentTask.Sealed, p.binding()...)
	if err != nil {
		// Why it would not open is not something the caller can do anything
		// about, and it is not something to repeat to a child either.
		return TaskSecret{}, fmt.Errorf("%w: %w", ErrSealed, err)
	}

	var secret TaskSecret
	if err := json.Unmarshal(plaintext, &secret); err != nil {
		return TaskSecret{}, fmt.Errorf("%w: %w", ErrSealed, err)
	}
	if secret.Version != SecretVersion {
		return TaskSecret{}, fmt.Errorf("%w: sealed as version %d, this build reads %d",
			ErrSealed, secret.Version, SecretVersion)
	}
	return secret, nil
}

// answeredMark sets the binding of an answered task apart from that of one
// still waiting for its answer, whatever the letter.
const answeredMark = "answered"

// binding is what a sealed task belongs to: this child, this task of theirs,
// and — once it is answered — the letter it was answered with. Moved to
// another profile or set against another task, it stops opening; and a task
// marked answered in the file by anybody but this service stops opening too,
// so that telling an answer again never tells it before the child has given
// one.
func (p *Profile) binding() []string {
	task := p.CurrentTask
	if task.Answered == nil {
		return unansweredBinding(p.StudentID, task.ID)
	}
	return answeredBinding(p.StudentID, task.ID, task.Answered.Choice)
}

// unansweredBinding is the binding of a task still waiting for its answer: on
// the card, or kept ready to be handed out under the same id, which is why a
// kept task opens on the card as it is.
func unansweredBinding(studentID, taskID string) []string {
	return []string{studentID, taskID}
}

// answeredBinding is the binding of a task answered with a letter.
func answeredBinding(studentID, taskID, choice string) []string {
	return []string{studentID, taskID, answeredMark, choice}
}
