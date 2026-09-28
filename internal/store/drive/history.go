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

const (
	// recent is how many of the latest revisions before a damaged one a
	// recovery tries, beside the latest one kept forever.
	recent = 4
	// maxKept is how many revisions of the profile's file are kept forever.
	// The first write of each day keeps one; past this many the earliest are
	// deleted, since Drive keeps at most two hundred so and a revision kept
	// forever can be let go of only by deleting it. That is a hundred days of
	// use to put a damaged file back from.
	maxKept = 100
)

// candidates are the revisions a recovery tries, in the order it tries them:
// the four latest before the one the file holds now — which is the damage —
// the latest first, and then the latest one kept forever, when it is none of
// those. That last one is the one a run of damage cannot push out of reach: a
// fault that wrote five damaged states in a row leaves the four latest all
// damaged, and the state kept for its day is still there.
//
// Drive promises no order of a file's revisions, so they are put in the order
// they were made.
func candidates(history []drive.Revision) []drive.Revision {
	if len(history) < 2 {
		return nil
	}
	latestFirst := slices.Clone(history)
	slices.SortStableFunc(latestFirst, func(a, b drive.Revision) int { return b.ModifiedTime.Compare(a.ModifiedTime) })
	earlier := latestFirst[1:]
	picked := slices.Clone(earlier[:min(recent, len(earlier))])
	if kept := slices.IndexFunc(earlier, func(r drive.Revision) bool { return r.KeepForever }); kept >= len(picked) {
		picked = append(picked, earlier[kept])
	}
	return picked
}

// recover puts a damaged file back to the latest earlier state of it that
// reads, as a new revision — nothing is deleted, so a recovery made wrongly
// can be undone from the same history — and answers ErrRestored; or answers
// ErrCorrupted when no state it tries reads. The damage is the state the file
// was read in, and the file is written only while it still holds it.
//
// Drive gives what an earlier revision held only once it is kept forever, so
// each one tried is kept first, room made for them beforehand. They stay
// kept, and the earliest are deleted with the others past maxKept.
//
// A recovery that could not look is not one that found nothing: a history
// Drive would not list, a Drive out of reach, or one too full to keep a
// revision in, stops it and is said as it is, and nobody is offered a new
// start for it. A revision Drive no longer has, will not give, or that does
// not read is passed over; one Drive refused for a reason the store has no
// word for is passed over too, and then nothing is said to be unreadable.
func (s *driveStore) recover(ctx context.Context, parent parentsDrive, id, damage string) error {
	history, err := parent.history(ctx, id)
	if err != nil {
		return s.refused(parent, id, err)
	}
	s.makeRoom(ctx, parent, id, history, recent+1)
	tried := 0
	var unsure error
	for _, candidate := range candidates(history) {
		tried++
		raw, err := s.earlier(ctx, parent, id, candidate)
		switch {
		case stopsRecovery(refusalOf(err)):
			return refusalOf(err)
		case outcomeOf(err) == outcomeFailed:
			unsure = err
			continue
		case err != nil:
			continue
		}
		if p, err := store.Parse(raw); err == nil {
			return s.restore(ctx, parent, id, damage, raw, p, tried)
		}
	}
	if unsure != nil {
		return fmt.Errorf("drivestore: a revision could not be read: %w", unsure)
	}
	s.watch.recovered(ctx, parent.account, tried, 0, recoveredNothing)
	return fmt.Errorf("%w: no earlier state of the file reads", store.ErrCorrupted)
}

// stopsRecovery reports whether an error stops a recovery rather than passing
// over one revision: access taken back or ending, Drive out of reach or with
// no room to keep a revision in, the time up, or the caller gone.
func stopsRecovery(err error) bool {
	for _, stop := range []error{
		store.ErrAccessRevoked, store.ErrAccessExpired, store.ErrUnavailable, store.ErrStorageFull,
		context.DeadlineExceeded, context.Canceled,
	} {
		if errors.Is(err, stop) {
			return true
		}
	}
	return false
}

// earlier is what an earlier revision of the file held: kept forever first,
// when it is not, since Drive gives no other.
func (s *driveStore) earlier(ctx context.Context, parent parentsDrive, id string, candidate drive.Revision) ([]byte, error) {
	if !candidate.KeepForever {
		if err := parent.keep(ctx, id, candidate.ID); err != nil {
			return nil, err
		}
	}
	return parent.downloadRevision(ctx, id, candidate.ID)
}

// restore writes a state recovered from the file's history over the damage,
// as it was, unless the file changed while its history was read — mended by
// the parent, or written by another instance — in which case what it holds
// now is what counts, and reading again finds it.
func (s *driveStore) restore(ctx context.Context, parent parentsDrive, id, damage string, raw []byte, p *profile.Profile, tried int) error {
	giveBack, err := s.turns.take(ctx, parent.account.ID)
	if err != nil {
		return err
	}
	defer giveBack()

	current, err := parent.download(ctx, id)
	now := fileRead{id: id, raw: current}
	if failed := now.took(err); failed != nil {
		return s.refused(parent, id, failed)
	}
	if now.state() != damage {
		return fmt.Errorf("%w: the file changed while its history was read", store.ErrConflict)
	}
	if err := s.upload(ctx, parent, id, p, raw, false); err != nil {
		return err
	}
	s.ids.wrote(parent.account.ID, id, p.Revision)
	s.watch.recovered(ctx, parent.account, tried, p.Revision, recoveredRestored)
	return fmt.Errorf("%w: put back to the state numbered %d", store.ErrRestored, p.Revision)
}

// makeRoom deletes the earliest revisions of the file kept forever, as the
// history lists them, so that room more can be kept within maxKept: Drive
// keeps at most two hundred so, and lets go of one only when it is deleted.
// It runs before the calls that keep one, which might otherwise be refused
// for want of room. A revision already gone is as good as deleted. What goes
// wrong here goes wrong quietly, with the lines of the calls to say so: what
// follows goes on, and the next day's first write makes room again.
func (s *driveStore) makeRoom(ctx context.Context, parent parentsDrive, id string, history []drive.Revision, room int) {
	kept := slices.DeleteFunc(slices.Clone(history), func(r drive.Revision) bool { return !r.KeepForever })
	if len(kept)+room <= maxKept {
		return
	}
	slices.SortStableFunc(kept, func(a, b drive.Revision) int { return a.ModifiedTime.Compare(b.ModifiedTime) })
	for _, old := range kept[:min(len(kept), len(kept)+room-maxKept)] {
		if err := parent.delete(ctx, id, old.ID); err != nil && !errors.Is(err, drive.ErrNotFound) {
			return
		}
	}
}
