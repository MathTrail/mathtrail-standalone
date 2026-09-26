package store

import (
	"errors"
	"fmt"

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
