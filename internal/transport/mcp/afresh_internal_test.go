package mcpserver

import (
	"context"
	"errors"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// A conflict whose caller has given up is not run again: the pause before the
// next run ends with the call, and says both why the run failed and why it was
// not made again.
func TestAConflictWhoseCallerGaveUpIsNotRunAgain(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	runs := 0
	_, err := afresh(ctx, func() (Reply[int], error) {
		runs++
		cancel()
		return Reply[int]{}, store.ErrConflict
	})
	if runs != 1 {
		t.Errorf("runs = %d after the caller gave up, want 1", runs)
	}
	if !errors.Is(err, store.ErrConflict) || !errors.Is(err, context.Canceled) {
		t.Errorf("afresh() error = %v, want %v and %v", err, store.ErrConflict, context.Canceled)
	}
}

// Every pause after a conflict is within its bounds.
func TestAPauseAfterAConflictIsShort(t *testing.T) {
	t.Parallel()

	for range 1000 {
		if wait := pause(); wait < 50e6 || wait >= 150e6 {
			t.Fatalf("pause() = %v, want 50 to 150 ms", wait)
		}
	}
}
