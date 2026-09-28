package drivestore

import (
	"context"
	"errors"
	"fmt"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

func (s *driveStore) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if err := ready(ctx, account); err != nil {
		return nil, "", fmt.Errorf("drivestore: load: %w", err)
	}
	id, raw, err := s.read(ctx, s.driveOf(account))
	if err != nil {
		return nil, "", fmt.Errorf("drivestore: load: %w", err)
	}
	p, err := store.Parse(raw)
	if err != nil {
		return nil, "", fmt.Errorf("drivestore: load: %w", err)
	}
	return p, revisionOf(id, p.Revision, raw), nil
}

func (s *driveStore) Export(ctx context.Context, account store.Account) (store.Location, error) {
	if err := ready(ctx, account); err != nil {
		return store.Location{}, fmt.Errorf("drivestore: export: %w", err)
	}
	parent := s.driveOf(account)
	file, err := s.search(ctx, parent)
	if err != nil {
		return store.Location{}, fmt.Errorf("drivestore: export: %w", err)
	}
	location := store.Location{File: file.Name, Link: file.WebViewLink}
	if len(file.Parents) == 0 {
		return location, nil
	}
	folder, err := parent.get(ctx, file.Parents[0])
	switch {
	case err == nil:
		location.Folder = folder.Name
	case errors.Is(err, drive.ErrNotFound):
		// A folder of the parent's own, which the service may not see: where
		// the file is has no name the service can tell, and the link still
		// opens it.
	default:
		return store.Location{}, fmt.Errorf("drivestore: export: %w", refusalOf(err))
	}
	return location, nil
}

// read finds the account's profile and downloads it: the file this instance
// remembers first, and the one a search finds when it remembers none, or when
// the one it remembers is gone.
func (s *driveStore) read(ctx context.Context, parent parentsDrive) (id string, raw []byte, err error) {
	if remembered, known := s.ids.recall(parent.account.ID); known {
		raw, err = parent.download(ctx, remembered)
		if !errors.Is(err, drive.ErrNotFound) {
			return remembered, raw, refusalOf(err)
		}
		// The file went from under the memory — deleted for good, or no longer
		// one the service may see — and a search says where the profile is now.
		s.ids.forget(parent.account.ID, remembered)
	}
	file, err := s.search(ctx, parent)
	if err != nil {
		return "", nil, err
	}
	raw, err = parent.download(ctx, file.ID)
	if err != nil {
		return "", nil, s.refused(parent, file.ID, err)
	}
	return file.ID, raw, nil
}

// locate is the file that holds the account's profile: the one this instance
// remembers, or the one a search finds.
func (s *driveStore) locate(ctx context.Context, parent parentsDrive) (string, error) {
	if id, remembered := s.ids.recall(parent.account.ID); remembered {
		return id, nil
	}
	file, err := s.search(ctx, parent)
	return file.ID, err
}

// search finds the account's profile by its marker, and remembers the file it
// found. Of several, the most recently changed is the profile, and Drive lists
// it first: two are there only after an unusual sequence — a file in the bin,
// a new one made, the old one restored — and merging them would be guesswork.
func (s *driveStore) search(ctx context.Context, parent parentsDrive) (drive.File, error) {
	found, err := parent.find(ctx, profileMarker)
	switch {
	case err != nil:
		return drive.File{}, refusalOf(err)
	case len(found) == 0:
		return drive.File{}, store.ErrNotFound
	}
	file := found[0]
	s.ids.remember(parent.account.ID, file.ID)
	return file, nil
}

// refused is what the store makes of a call about the account's file that
// Drive refused. A file Drive no longer has is a profile the account no
// longer has, and is forgotten; anything else is Drive's refusal, in the
// store's words where it has them.
func (s *driveStore) refused(parent parentsDrive, id string, err error) error {
	if errors.Is(err, drive.ErrNotFound) {
		s.ids.forget(parent.account.ID, id)
		return fmt.Errorf("%w: %w", store.ErrNotFound, err)
	}
	return refusalOf(err)
}
