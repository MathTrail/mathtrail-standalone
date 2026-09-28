package drivestore

import (
	"context"
	"errors"
	"testing"
	"time"
)

// A write whose caller gives up while it waits for the account's turn never
// takes it, and an account nobody writes for any more takes no memory.
func TestATurnGivenUpOnWhileWaitingIsNeverTaken(t *testing.T) {
	t.Parallel()

	queue := newTurns()
	giveBack, err := queue.take(t.Context(), "mia")
	if err != nil {
		t.Fatalf("take() error = %v, want the turn", err)
	}
	gaveUp, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := queue.take(gaveUp, "mia"); !errors.Is(err, context.Canceled) {
		t.Errorf("take() given up on while another holds the turn: error = %v, want %v", err, context.Canceled)
	}
	if other, err := queue.take(t.Context(), "leo"); err != nil {
		t.Errorf("take() for another account while mia's is held: error = %v, want the turn", err)
	} else {
		other()
	}
	giveBack()
	if held := len(queue.held); held != 0 {
		t.Errorf("turns held after every write is done = %d, want none", held)
	}
}

// The store's own wait waits for as long as asked, unless the context ends
// first.
func TestTheTimerWaitsOrEndsWithItsContext(t *testing.T) {
	t.Parallel()

	if err := timer(t.Context(), time.Millisecond); err != nil {
		t.Errorf("timer() error = %v, want nil", err)
	}
	ended, cancel := context.WithCancel(t.Context())
	cancel()
	if err := timer(ended, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("timer() with an ended context: error = %v, want %v", err, context.Canceled)
	}
}
