package profile

// SealTask puts the part of the task on the card that gives its answer away
// beyond reach, tied to this child and this task, as it stands: unanswered,
// or answered with the letter it was given. The service seals a task as it
// hands it out, through Issue and Keep; the tests outside the package put a
// task on the card by hand, under an id of their choosing, and seal it here.
func (p *Profile) SealTask(sealer Sealer, secret *TaskSecret) error {
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
