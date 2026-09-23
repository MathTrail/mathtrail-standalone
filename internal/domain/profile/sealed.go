package profile

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrSealed means the sealed part of the task in flight cannot be read: the
// key that sealed it has been retired, or it was sealed under a shape this
// build does not know. A task is worth less than a profile, so the answer is
// to tell the child this one can no longer be checked and give them another.
var ErrSealed = errors.New("profile: the sealed part of the task cannot be read")

// SecretVersion is the shape of the sealed part. It moves on its own: what is
// sealed lives as long as one task, and a task in flight is not worth a
// migration.
const SecretVersion = 1

// Sealer protects a value the service must keep to itself and reads it back.
// The profile declares what it needs rather than reaching for whoever
// provides it: this package computes, and what it computes with is handed in.
//
// The binding is what the value belongs to, and opening it again requires the
// same binding: without that, one child's sealed answer would open in another
// child's profile.
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

// SealTask puts the part of the task in flight that gives its answer away
// beyond reach, tied to this child and this task.
func (p *Profile) SealTask(sealer Sealer, secret TaskSecret) error {
	if p.CurrentTask == nil {
		return ErrNoTask
	}
	secret.Version = SecretVersion

	plaintext, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("profile: seal the task: %w", err)
	}
	sealed, err := sealer.Seal(plaintext, p.binding()...)
	if err != nil {
		return fmt.Errorf("profile: seal the task: %w", err)
	}
	p.CurrentTask.Sealed = sealed
	return nil
}

// OpenTask reads that part back. It is opened in one place and at one moment:
// after the child has answered, when the answer is no longer a secret.
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

// binding is what a sealed task belongs to: this child, and this task of
// theirs. Moved to another profile or set against another task, it stops
// opening.
func (p *Profile) binding() []string {
	return []string{p.StudentID, p.CurrentTask.ID}
}
