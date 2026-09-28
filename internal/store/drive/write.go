package drivestore

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

func (s *driveStore) Create(ctx context.Context, account store.Account, p *profile.Profile) (store.Revision, error) {
	if err := ready(ctx, account); err != nil {
		return "", fmt.Errorf("drivestore: create: %w", err)
	}
	raw, err := profile.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("drivestore: create: %w", err)
	}
	parent := s.driveOf(account)
	// A profile is never replaced by a new one the parent did not ask for,
	// readable or not.
	_, err = s.search(ctx, parent)
	switch {
	case err == nil:
		return "", fmt.Errorf("drivestore: create: %w: the account has a profile already", store.ErrConflict)
	case !errors.Is(err, store.ErrNotFound):
		return "", fmt.Errorf("drivestore: create: %w", err)
	}
	folder, err := s.folder(ctx, parent)
	if err != nil {
		return "", fmt.Errorf("drivestore: create: %w", err)
	}
	made, err := parent.create(ctx, &drive.File{
		Name:     fileName,
		MimeType: fileType,
		Parents:  []string{folder},
		AppProperties: map[string]string{
			markerKey: profileMarker,
			schemaKey: strconv.Itoa(p.SchemaVersion),
		},
	}, raw)
	if err != nil {
		return "", fmt.Errorf("drivestore: create: %w", refusalOf(err))
	}
	s.ids.remember(account.ID, made.ID)
	return revisionOf(made.ID, p.Revision, raw), nil
}

func (s *driveStore) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	if err := ready(ctx, account); err != nil {
		return "", fmt.Errorf("drivestore: save: %w", err)
	}
	raw, err := profile.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("drivestore: save: %w", err)
	}
	parent := s.driveOf(account)
	id, err := s.locate(ctx, parent)
	if err != nil {
		return "", fmt.Errorf("drivestore: save: %w", err)
	}
	read, readable := parseRevision(expected)
	switch {
	case !readable || read.file != id:
		return "", fmt.Errorf("drivestore: save: %w: the profile was not read from the account's file", store.ErrConflict)
	case p.Revision != read.counter+1:
		return "", fmt.Errorf("drivestore: save: the profile's revision is %d, want %d: a write moves it on by exactly one",
			p.Revision, read.counter+1)
	}

	// Drive keeps whatever is uploaded last, so the file is read once more
	// and held to what the profile was computed from. A change made since —
	// another tab, another instance, the parent by hand — is not written over.
	current, err := parent.download(ctx, id)
	if err != nil {
		return "", fmt.Errorf("drivestore: save: %w", s.refused(parent, id, err))
	}
	if digestOf(current) != read.digest {
		return "", fmt.Errorf("drivestore: save: %w: the file changed after it was read", store.ErrConflict)
	}
	// The marker stays as it was made: Drive keeps the properties an update
	// does not name. The copy of the version goes with every write, so that
	// it never falls behind the file.
	err = parent.update(ctx, id, &drive.File{
		MimeType:      fileType,
		AppProperties: map[string]string{schemaKey: strconv.Itoa(p.SchemaVersion)},
	}, raw)
	if err != nil {
		return "", fmt.Errorf("drivestore: save: %w", s.refused(parent, id, err))
	}
	return revisionOf(id, p.Revision, raw), nil
}

// folder is the folder a new profile is made in: the one found by its marker —
// the most recently changed, which Drive lists first — or a new one when the
// parent never had one, or deleted it.
func (s *driveStore) folder(ctx context.Context, parent parentsDrive) (string, error) {
	found, err := parent.find(ctx, folderMarker)
	if err != nil {
		return "", refusalOf(err)
	}
	if len(found) > 0 {
		return found[0].ID, nil
	}
	made, err := parent.create(ctx, &drive.File{
		Name:          folderName,
		MimeType:      drive.FolderType,
		AppProperties: map[string]string{markerKey: folderMarker},
	}, nil)
	if err != nil {
		return "", refusalOf(err)
	}
	return made.ID, nil
}
