package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// attempts is how many times a read of the profile, and the write that
// follows it, are made before what stops them is told as it is.
const attempts = 3

// afresh makes a read of the profile, and the write that follows it, and
// makes them again from a fresh read when the store says the profile moved on
// in between: somebody else — another tab, another instance, the parent by
// hand — wrote it, which is a conflict, or deleted it for good before the
// write, which the store finds as no profile.
//
// Nothing the first run computed is sent again. Every run reads what is there
// now and decides from it, so what the other writer did already is found done
// and told as such — the request already open, the answer already recorded,
// the task already accepted — and what it did not do is done on top of what
// it did; a profile deleted is found gone, and told as that. A short pause,
// spread at random, keeps two runs that keep meeting from meeting again. The
// third time is told as it is.
func afresh[Out any](ctx context.Context, run func() (Reply[Out], error)) (Reply[Out], error) {
	for attempt := 1; ; attempt++ {
		reply, err := run()
		movedOn := errors.Is(err, store.ErrConflict) || errors.Is(err, store.ErrNotFound)
		if !movedOn || attempt == attempts {
			return reply, err
		}
		if waited := sleep(ctx, pause()); waited != nil {
			return Reply[Out]{}, errors.Join(err, waited)
		}
	}
}

// pause is how long a run waits before it is made again: 50 to 150 ms.
func pause() time.Duration {
	//nolint:gosec // the spread only keeps two writers apart; nothing depends on guessing it
	return 50*time.Millisecond + rand.N(100*time.Millisecond)
}

// sleep waits for d, or until ctx ends.
func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("mcp: wait before trying again: %w", ctx.Err())
	}
}
