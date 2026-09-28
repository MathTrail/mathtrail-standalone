package drivestore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// recheck is how long a read that came back with an earlier state than one
// this instance wrote waits before it reads once more.
const recheck = 250 * time.Millisecond

func (s *driveStore) Load(ctx context.Context, account store.Account) (*profile.Profile, store.Revision, error) {
	if err := ready(ctx, account); err != nil {
		return nil, "", fmt.Errorf("drivestore: load: %w", err)
	}
	parent := s.driveOf(account)
	read, err := s.read(ctx, parent)
	if err != nil {
		return nil, "", fmt.Errorf("drivestore: load: %w", err)
	}
	p, err := store.Parse(read.raw)
	if errors.Is(err, store.ErrCorrupted) && read.remembered {
		// The file this instance remembers may have been set aside by
		// another since, its damage and all, with a new profile started in
		// its place: damage is never mended in a file before the marker says
		// it is still the profile.
		s.ids.forget(account.ID, read.id)
		if read, err = s.searchAndRead(ctx, parent); err != nil {
			return nil, "", fmt.Errorf("drivestore: load: %w", err)
		}
		p, err = store.Parse(read.raw)
	}
	switch {
	case errors.Is(err, store.ErrCorrupted) && (read.oversized || errors.Is(err, profile.ErrMalformed)):
		return nil, "", fmt.Errorf("drivestore: load: %w", s.recover(ctx, parent, read.id, read.state()))
	case err != nil:
		// A profile that breaks a rule of this build is damage too, but not
		// one to roll back: a newer build may have written it without the
		// version it needed, or the parent edited it by hand, and putting an
		// earlier state over either would lose it. It is told, and left.
		return nil, "", fmt.Errorf("drivestore: load: %w", err)
	}
	if read.written > p.Revision {
		if p, read, err = s.caughtUp(ctx, parent, read, p); err != nil {
			return nil, "", fmt.Errorf("drivestore: load: %w", err)
		}
	}
	return p, revisionOf(read.id, p.Revision, read.raw), nil
}

// fileRead is the account's profile file as a read found it.
type fileRead struct {
	id  string
	raw []byte
	// oversized says the file holds more than any profile, and nothing of it
	// was read: it is damage as surely as bytes that do not parse.
	oversized bool
	// remembered says the file is the one this instance remembered, rather
	// than one a search found just now.
	remembered bool
	// written is the profile's number this instance last wrote to the file,
	// or none.
	written int
}

// tooLarge is the state of a file too large to read, told apart from the
// digest of any bytes.
const tooLarge = "too large"

// state is what the file held when it was read, as a later read tells it
// apart: the digest of its bytes, or tooLarge.
func (r *fileRead) state() string {
	if r.oversized {
		return tooLarge
	}
	return digestOf(r.raw)
}

// read finds the account's profile and downloads it: the file this instance
// remembers first, and the one a search finds when it remembers none, or when
// the one it remembers is gone.
func (s *driveStore) read(ctx context.Context, parent parentsDrive) (fileRead, error) {
	if known, remembered := s.ids.recall(parent.account.ID, s.now()); remembered {
		raw, err := parent.download(ctx, known.file)
		if !errors.Is(err, drive.ErrNotFound) {
			read := fileRead{id: known.file, raw: raw, remembered: true, written: known.written}
			return read, refusalOf(read.took(err))
		}
		// The file went from under the memory — deleted for good, or no longer
		// one the service may reach — and a search says where the profile is
		// now.
		s.ids.forget(parent.account.ID, known.file)
	}
	return s.searchAndRead(ctx, parent)
}

// searchAndRead finds the account's profile by its marker and downloads it.
func (s *driveStore) searchAndRead(ctx context.Context, parent parentsDrive) (fileRead, error) {
	found, err := s.search(ctx, parent)
	if err != nil {
		return fileRead{}, err
	}
	raw, err := parent.download(ctx, found[0].ID)
	read := fileRead{id: found[0].ID, raw: raw}
	if failed := read.took(err); failed != nil {
		return fileRead{}, s.refused(parent, found[0].ID, failed)
	}
	return read, nil
}

// took is what became of the download of the file: a file too large to read
// is taken as damage, and anything else Drive answered stays as it was.
func (r *fileRead) took(err error) error {
	if errors.Is(err, drive.ErrTooLarge) {
		r.oversized = true
		return nil
	}
	return err
}

// caughtUp holds a read to the profile's number this instance last wrote to
// the file: Drive does not promise that a read right after a write sees it,
// and the lesson reads straight after it writes. One that comes back earlier
// is read once more after a moment. One that is still behind is refused as
// that, and the number is let go of, so that the next call reads the file as
// it is — whatever put it back.
func (s *driveStore) caughtUp(ctx context.Context, parent parentsDrive, read fileRead, early *profile.Profile) (*profile.Profile, fileRead, error) {
	account := parent.account
	if err := s.sleep(ctx, recheck); err != nil {
		return nil, fileRead{}, err
	}
	raw, err := parent.download(ctx, read.id)
	if err != nil {
		return nil, fileRead{}, s.refused(parent, read.id, err)
	}
	p, err := store.Parse(raw)
	if err != nil {
		return nil, fileRead{}, err
	}
	if p.Revision < read.written {
		// A read that stayed behind the number once is not held to it again:
		// a file put back on purpose is read as it is by the next call.
		s.ids.wrote(account.ID, read.id, 0)
		s.watch.staleRead(ctx, account, read.written, p.Revision, staleBehind)
		return nil, fileRead{}, fmt.Errorf("%w: read %d, wrote %d", store.ErrBehind, p.Revision, read.written)
	}
	s.watch.staleRead(ctx, account, read.written, early.Revision, staleCaughtUp)
	read.raw = raw
	return p, read, nil
}

func (s *driveStore) Export(ctx context.Context, account store.Account) (store.Location, error) {
	if err := ready(ctx, account); err != nil {
		return store.Location{}, fmt.Errorf("drivestore: export: %w", err)
	}
	parent := s.driveOf(account)
	found, err := s.search(ctx, parent)
	if err != nil {
		return store.Location{}, fmt.Errorf("drivestore: export: %w", err)
	}
	file := found[0]
	location := store.Location{File: file.Name, Link: file.WebViewLink}
	for i := range found[1:] {
		other := &found[1+i]
		location.Others = append(location.Others, store.Elsewhere{File: other.Name, Link: other.WebViewLink})
	}
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

// locate is the file that holds the account's profile: the one this instance
// remembers, or the one a search finds.
func (s *driveStore) locate(ctx context.Context, parent parentsDrive) (string, error) {
	if known, remembered := s.ids.recall(parent.account.ID, s.now()); remembered {
		return known.file, nil
	}
	found, err := s.search(ctx, parent)
	if err != nil {
		return "", err
	}
	return found[0].ID, nil
}

// search finds the account's profile by its marker, and remembers the file it
// found; it answers every file that carries the marker, the one that is the
// profile first. Of several, the most recently changed is the profile, and
// Drive lists it first: two are there only after an unusual sequence — a file
// in the bin, a new one made, the old one restored — and merging them would be
// guesswork.
//
// A search leaves out the bin, and one that finds nothing looks there: a
// profile the parent put in the bin is theirs to restore, and not a profile
// missing, which a first sign-in would make again beside it.
func (s *driveStore) search(ctx context.Context, parent parentsDrive) ([]drive.File, error) {
	found, err := parent.find(ctx, profileMarker, false)
	switch {
	case err != nil:
		return nil, refusalOf(err)
	case len(found) > 0:
		s.ids.remember(parent.account.ID, found[0].ID, s.now())
		return found, nil
	}
	binned, err := parent.find(ctx, profileMarker, true)
	switch {
	case err != nil:
		return nil, refusalOf(err)
	case len(binned) > 0:
		return nil, store.ErrInBin
	}
	return nil, store.ErrNotFound
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
