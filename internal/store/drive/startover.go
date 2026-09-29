package drivestore

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// StartOver keeps a new profile in place of one this build cannot read:
// damaged, or of a newer build.
//
// The file that is the profile is downloaded once more, and a profile that
// reads — mended, or put back, since the parent was told it did not — is never
// replaced. The new profile is then made as a first one, and only then is the
// old file set aside — renamed by the day, and marked as set aside instead of
// as the profile — so that a new start that fails halfway leaves the account
// as it found it, or with the new profile beside the old one, and never with
// none. So is every other file that carries the marker and cannot be read
// either, and every one in the bin, which is left there so that restoring it
// later makes no second profile beside the new one. A file beside it that
// reads is a profile too, and stays as it is, for the parent to find. Nothing
// is deleted, and nothing a file holds is changed.
//
// An instance that still remembers a damaged file set aside reads it as
// damaged, and searches again before it does anything about the damage: so it
// finds the new profile. One that remembers a file set aside from the bin,
// which reads, goes on with it for as long as it trusts what it remembers, and
// the first write it makes there sends it to the search.
func (s *driveStore) StartOver(ctx context.Context, account store.Account, p *profile.Profile) (store.Revision, error) {
	if err := ready(ctx, account); err != nil {
		return "", fmt.Errorf("drivestore: start over: %w", err)
	}
	raw, err := profile.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("drivestore: start over: %w", err)
	}
	giveBack, err := s.turns.take(ctx, account.ID)
	if err != nil {
		return "", fmt.Errorf("drivestore: start over: %w", err)
	}
	defer giveBack()

	parent := s.driveOf(account)
	unreadable, inBin, err := s.toSetAside(ctx, parent)
	if err != nil {
		return "", fmt.Errorf("drivestore: start over: %w", err)
	}
	revision, err := s.make(ctx, parent, p, raw)
	if err != nil {
		return "", fmt.Errorf("drivestore: start over: %w", err)
	}
	files := slices.Concat(unreadable, inBin)
	for i := range files {
		err = s.setAside(ctx, parent, files[i].ID)
		if err != nil {
			return "", fmt.Errorf("drivestore: start over: %w", err)
		}
	}
	s.watch.startedOver(ctx, account, len(unreadable), len(inBin))
	return revision, nil
}

// toSetAside is the files a new start sets aside: those that carry the
// profile's marker and cannot be read, and those in the bin. The first of the
// ones outside the bin is the profile, and has to be one this build cannot
// read; a file beside it that some build reads — this one, or a newer — is a
// profile too, and stays.
func (s *driveStore) toSetAside(ctx context.Context, parent parentsDrive) (unreadable, inBin []drive.File, err error) {
	beside, err := parent.find(ctx, profileMarker, false)
	if err != nil {
		return nil, nil, refusalOf(err)
	}
	for i := range beside {
		var read reading
		read, err = s.reads(ctx, parent, beside[i].ID)
		switch {
		case i == 0 && errors.Is(err, errGone):
			return nil, nil, fmt.Errorf("%w: the profile's file went while it was set aside", store.ErrConflict)
		case i == 0 && err == nil && read == readHere:
			return nil, nil, fmt.Errorf("%w: the profile can be read", store.ErrConflict)
		case errors.Is(err, errGone), i > 0 && err == nil && read != readByNone:
			// An older file gone in the meantime has nothing to set aside, and
			// one that reads is a profile, never replaced.
		case err != nil:
			return nil, nil, err
		default:
			unreadable = append(unreadable, beside[i])
		}
	}
	inBin, err = parent.find(ctx, profileMarker, true)
	if err != nil {
		return nil, nil, refusalOf(err)
	}
	return unreadable, inBin, nil
}

// errGone is a file Drive no longer has.
var errGone = errors.New("the file is gone")

// reading is which build reads a file carrying the profile's marker.
type reading int

const (
	readByNone reading = iota
	readHere
	readByNewer
)

// reads is which build reads a file carrying the profile's marker as a
// profile: this one, a newer one, or none. A file too large for any profile is
// read by none; one Drive no longer has is errGone.
func (s *driveStore) reads(ctx context.Context, parent parentsDrive, id string) (reading, error) {
	current, err := parent.download(ctx, id)
	switch {
	case errors.Is(err, drive.ErrNotFound):
		s.ids.forget(parent.account.ID, id)
		return readByNone, errGone
	case errors.Is(err, drive.ErrTooLarge):
		return readByNone, nil
	case err != nil:
		return readByNone, refusalOf(err)
	}
	switch _, err = store.Parse(current); {
	case err == nil:
		return readHere, nil
	case errors.Is(err, profile.ErrNewer):
		return readByNewer, nil
	}
	return readByNone, nil
}

// setAside renames a file by the day and marks it as set aside, so that no
// search for the profile finds it again. A file gone for good in the meantime
// has nothing left to set aside.
func (s *driveStore) setAside(ctx context.Context, parent parentsDrive, id string) error {
	_, err := parent.update(ctx, id, &drive.Change{Meta: &drive.File{
		Name:          setAsideName(s.now()),
		AppProperties: map[string]string{markerKey: setAsideMarker},
	}})
	s.ids.forget(parent.account.ID, id)
	if err != nil && !errors.Is(err, drive.ErrNotFound) {
		return refusalOf(err)
	}
	return nil
}
