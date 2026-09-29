package drivestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

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
	giveBack, err := s.turns.take(ctx, account.ID)
	if err != nil {
		return "", fmt.Errorf("drivestore: create: %w", err)
	}
	defer giveBack()

	parent := s.driveOf(account)
	// A profile is never replaced by a new one the parent did not ask for,
	// readable or not, and one in the bin is the parent's to restore.
	_, err = s.search(ctx, parent)
	switch {
	case err == nil:
		return "", fmt.Errorf("drivestore: create: %w: the account has a profile already", store.ErrConflict)
	case !errors.Is(err, store.ErrNotFound):
		return "", fmt.Errorf("drivestore: create: %w", err)
	}
	revision, err := s.make(ctx, parent, p, raw)
	if err != nil {
		return "", fmt.Errorf("drivestore: create: %w", err)
	}
	return revision, nil
}

// make keeps a profile in a new file, in the folder its marker finds or in a
// new one, and remembers the file.
func (s *driveStore) make(ctx context.Context, parent parentsDrive, p *profile.Profile, raw []byte) (store.Revision, error) {
	folder, err := s.folder(ctx, parent)
	if err != nil {
		return "", err
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
		return "", refusalOf(err)
	}
	s.ids.remember(parent.account.ID, made.ID, s.now())
	s.ids.wrote(parent.account.ID, made.ID, p.Revision)
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
	giveBack, err := s.turns.take(ctx, account.ID)
	if err != nil {
		return "", fmt.Errorf("drivestore: save: %w", err)
	}
	defer giveBack()

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

	current, err := s.unchanged(ctx, parent, account, id, read)
	if err != nil {
		return "", fmt.Errorf("drivestore: save: %w", err)
	}
	keep := firstOfTheDay(current, p)
	if keep {
		// Making room takes calls of its own, and another write may land in
		// the meantime; the file is held to what the profile was computed
		// from once more, so that between the last look and the write there
		// is one upload again.
		s.roomForOne(ctx, parent, id)
		if _, err := s.unchanged(ctx, parent, account, id, read); err != nil {
			return "", fmt.Errorf("drivestore: save: %w", err)
		}
	}
	if err := s.upload(ctx, parent, id, p, raw, keep); err != nil {
		return "", fmt.Errorf("drivestore: save: %w", err)
	}
	s.ids.wrote(account.ID, id, p.Revision)
	return revisionOf(id, p.Revision, raw), nil
}

// upload writes a profile over what the file holds — its revision kept
// forever when keep says so — and holds the file Drive answers with to still
// being the profile.
//
// The marker stays as it was made: Drive keeps the properties an update does
// not name. The copy of the version goes with every write, so that it never
// falls behind the file.
//
// Drive writes into a file in the bin, or one set aside since it was read, as
// into any other, and says so only afterwards. Such a write counts for
// nothing: the file is forgotten, and the write refused as a conflict, so
// that whoever made it reads again and finds where the profile is now.
//
// An upload asked to keep its revision that Drive refuses for a reason the
// store has no word for — the most revisions it keeps, among them — is made
// again without: the write matters more than the revision kept, and Drive did
// not take the one it refused.
func (s *driveStore) upload(ctx context.Context, parent parentsDrive, id string, p *profile.Profile, raw []byte, keep bool) error {
	change := drive.Change{
		Meta:    &drive.File{MimeType: fileType, AppProperties: map[string]string{schemaKey: strconv.Itoa(p.SchemaVersion)}},
		Content: raw,
		Keep:    keep,
	}
	changed, err := parent.update(ctx, id, &change)
	if err != nil && keep && outcomeOf(err) == outcomeFailed {
		change.Keep = false
		changed, err = parent.update(ctx, id, &change)
	}
	if err != nil {
		return s.refused(parent, id, err)
	}
	var why string
	switch {
	case changed.Trashed:
		why = conflictInBin
	case changed.AppProperties[markerKey] != profileMarker:
		why = conflictSetAside
	default:
		return nil
	}
	s.ids.forget(parent.account.ID, id)
	s.watch.conflict(ctx, parent.account, p.Revision-1, p.Revision, why)
	return fmt.Errorf("%w: the file stopped being the profile before the write", store.ErrConflict)
}

// folder is the folder a new profile is made in: the one found by its marker —
// the most recently changed, which Drive lists first — or a new one when the
// parent never had one, or deleted it.
func (s *driveStore) folder(ctx context.Context, parent parentsDrive) (string, error) {
	found, err := parent.find(ctx, folderMarker, false)
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

// firstOfTheDay reports whether a write is the first of its day, by the day
// the file was last written on, as the file itself says: that write's
// revision is kept forever, so that a damaged file can be put back to how it
// was on each day it was used. A write lost to another changes nothing in the
// file, so the next one still finds the day's first write to make; and a file
// that does not say when it was written is kept, as one nothing is known of.
// A write dated before the file's day — an instance whose clock runs behind
// another's, around midnight — is no new day.
func firstOfTheDay(current []byte, p *profile.Profile) bool {
	var written struct {
		UpdatedAt profile.Time `json:"updated_at"`
	}
	if json.Unmarshal(current, &written) != nil || written.UpdatedAt.IsZero() {
		return true
	}
	return dayOf(p.UpdatedAt.Time) > dayOf(written.UpdatedAt.Time)
}

// dayOf is the day a moment falls on, as UTC counts days, in a form that
// orders as the days do.
func dayOf(moment time.Time) string {
	return moment.UTC().Format(time.DateOnly)
}

// counterOf is the profile's number the bytes hold, or none when they do not
// read as a profile.
func counterOf(raw []byte) int {
	var held struct {
		Revision int `json:"revision"`
	}
	if json.Unmarshal(raw, &held) != nil {
		return 0
	}
	return held.Revision
}

// unchanged reads the file once more and holds it to what the profile was
// computed from, and is what the file holds. Drive keeps whatever is uploaded
// last, so a change made since — another tab, another instance, the parent by
// hand — is a conflict, and is not written over.
func (s *driveStore) unchanged(ctx context.Context, parent parentsDrive, account store.Account, id string, read revision) ([]byte, error) {
	current, err := parent.download(ctx, id)
	if err != nil {
		return nil, s.refused(parent, id, err)
	}
	if digestOf(current) != read.digest {
		s.watch.conflict(ctx, account, read.counter, counterOf(current), conflictChanged)
		return nil, fmt.Errorf("%w: the file changed after it was read", store.ErrConflict)
	}
	return current, nil
}

// roomForOne makes room for one more revision of the file kept forever, by
// the history as Drive lists it now. A history Drive would not list leaves
// the room as it is, and the write goes on.
func (s *driveStore) roomForOne(ctx context.Context, parent parentsDrive, id string) {
	history, err := parent.history(ctx, id)
	if err != nil {
		return
	}
	s.makeRoom(ctx, parent, id, history, 1)
}
