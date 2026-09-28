package drivestore

import (
	"context"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The lines the store leaves about what happened to a profile, beside the one
// every call to Drive leaves. Each carries the profile's own numbers and words
// of a closed list, and never the file's ID, its name or anything it holds.
const (
	// eventConflict is a write refused because the file was not the one it
	// was computed from.
	eventConflict = "drive_conflict"
	// eventStaleRead is a read that came back with an earlier state than one
	// this instance wrote.
	eventStaleRead = "drive_stale_read"
	// eventRecovered is a damaged file put back to its last state that reads,
	// or found to have none.
	eventRecovered = "drive_recovered"
	// eventStartedOver is a new profile kept in place of one nothing could
	// read, and the files set aside for it.
	eventStartedOver = "drive_started_over"
)

// Why a write was refused as a conflict.
const (
	// conflictChanged is a file whose bytes are no longer the ones read.
	conflictChanged = "changed"
	// conflictInBin is a file the parent put in the bin in the meantime.
	conflictInBin = "in_bin"
	// conflictSetAside is a file set aside in the meantime.
	conflictSetAside = "set_aside"
)

// How a read that came back early ended.
const (
	staleCaughtUp = "caught_up"
	staleBehind   = "behind"
)

// How a recovery ended.
const (
	recoveredRestored = "restored"
	recoveredNothing  = "nothing_readable"
)

// conflict writes that a write was refused, from which of the profile's
// numbers, and why. The number found is 0 when the file found reads as no
// profile.
func (w *watch) conflict(ctx context.Context, account store.Account, read, found int, why string) {
	w.logger.Info(eventConflict, append([]zap.Field{
		zap.Int("read_revision", read),
		zap.Int("found_revision", found),
		zap.String("reason", why),
	}, w.callerFields(ctx, account)...)...)
}

// staleRead writes that a read came back with an earlier number than the one
// written, and whether a second read caught up with it. One that did not is a
// warning: Drive is serving behind.
func (w *watch) staleRead(ctx context.Context, account store.Account, written, read int, outcome string) {
	fields := append([]zap.Field{
		zap.Int("written_revision", written),
		zap.Int("read_revision", read),
		zap.String("outcome", outcome),
	}, w.callerFields(ctx, account)...)
	if outcome == staleBehind {
		w.logger.Warn(eventStaleRead, fields...)
		return
	}
	w.logger.Info(eventStaleRead, fields...)
}

// recovered writes that a damaged file was put back, or could not be: how
// many revisions were tried, and the number of the one it was put back to.
// Either way something damaged a parent's file, which is worth a warning.
func (w *watch) recovered(ctx context.Context, account store.Account, tried, restored int, outcome string) {
	w.logger.Warn(eventRecovered, append([]zap.Field{
		zap.Int("tried", tried),
		zap.Int("restored_revision", restored),
		zap.String("outcome", outcome),
	}, w.callerFields(ctx, account)...)...)
}

// startedOver writes that a profile was started over, and how many files were
// set aside for it: beside it, and in the bin.
func (w *watch) startedOver(ctx context.Context, account store.Account, beside, inBin int) {
	w.logger.Info(eventStartedOver, append([]zap.Field{
		zap.Int("set_aside", beside),
		zap.Int("set_aside_from_bin", inBin),
	}, w.callerFields(ctx, account)...)...)
}
