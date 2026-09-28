// Package ratelimittest holds paces for the tests of whatever a pace stands in
// front of: one no test comes near, and ones on a clock that stands still, so
// that nothing a test spends comes back to it while it runs.
package ratelimittest

import (
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
)

// roomyPerMinute is more requests a minute than any test makes.
const roomyPerMinute = 1 << 20

// stillClock is a clock that never moves.
func stillClock() time.Time { return time.Date(2026, time.September, 28, 9, 0, 0, 0, time.UTC) }

// Roomy is a pace no test comes near, for the tests of anything but the paces.
func Roomy(t testing.TB) ratelimit.Limiter {
	t.Helper()
	return Shared(t, roomyPerMinute)
}

// Keyed is a pace of so many requests a minute that counts every key apart, on
// a clock that stands still.
func Keyed(t testing.TB, perMinute int) ratelimit.Limiter {
	t.Helper()

	limiter, err := ratelimit.New(ratelimit.Settings{PerMinute: perMinute, Keys: ratelimit.MaxKeys, Now: stillClock})
	if err != nil {
		t.Fatalf("ratelimit.New() error = %v, want nil", err)
	}
	return limiter
}

// Shared is a pace of so many requests a minute that every key shares, on a
// clock that stands still.
func Shared(t testing.TB, perMinute int) ratelimit.Limiter {
	t.Helper()

	limiter, err := ratelimit.NewShared(perMinute, stillClock)
	if err != nil {
		t.Fatalf("ratelimit.NewShared() error = %v, want nil", err)
	}
	return limiter
}
