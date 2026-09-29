package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// Parse reads the bytes a store keeps as a profile, and says which of two
// things a file that cannot be read is.
//
// A file that is not JSON, is not a profile, or is of an older shape this build
// knows no way up from is damage: ErrCorrupted, with the reason beside it.
// Nothing will ever read it as it stands, and what helps is an earlier state of
// it. A file written by a newer build is not damage and keeps profile.ErrNewer:
// the build that wrote it reads it perfectly well, and what helps is waiting.
func Parse(raw []byte) (*profile.Profile, error) {
	p, err := profile.Parse(raw)
	switch {
	case err == nil:
		return p, nil
	case errors.Is(err, profile.ErrNewer):
		return nil, err
	default:
		return nil, fmt.Errorf("%w: %w", ErrCorrupted, err)
	}
}

// RolloutWindow is how long a file a newer version wrote is taken for a rollout
// under way: an older instance answering meanwhile is told to wait. A rollout
// is over in minutes; a file written longer ago than this, or on a day it
// cannot be, was edited by hand or written by a version since withdrawn.
const RolloutWindow = 30 * time.Minute

// ParseAt reads the bytes a store keeps as a profile at a moment, as Parse
// does, and tells a file of a newer version by that moment: every read a store
// answers with goes through it, so that none tells a file no rollout explains
// to wait.
func ParseAt(raw []byte, now time.Time) (*profile.Profile, error) {
	p, err := Parse(raw)
	return p, newerAt(err, now)
}

// newerAt is a refusal as a store answers it at a moment: a file of a newer
// version written within RolloutWindow of now stays profile.ErrNewer, and
// one written further from it — or that does not say when — is ErrUnsupported
// beside it. Any other refusal is left as it is.
func newerAt(err error, now time.Time) error {
	var newer *profile.NewerError
	if !errors.As(err, &newer) {
		return err
	}
	if age := now.Sub(newer.WrittenAt); !newer.WrittenAt.IsZero() && age <= RolloutWindow && age >= -RolloutWindow {
		return err
	}
	return fmt.Errorf("%w: %w", ErrUnsupported, err)
}
