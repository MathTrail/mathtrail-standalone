package mcpserver

import (
	"context"
	"errors"
	"testing"
	"time"
)

// A call waits for the writes of its account begun before it, and for none of
// another account's; a call whose caller gives up stops waiting.
func TestACallWaitsForTheWritesOfItsAccountBegunBeforeIt(t *testing.T) {
	t.Parallel()

	writes := newPending()
	_, end := writes.begin("masha")
	_, otherEnd := writes.begin("petya")
	defer otherEnd()

	waited := make(chan error, 1)
	go func() { waited <- writes.wait(t.Context(), "masha") }()
	select {
	case err := <-waited:
		t.Fatalf("wait() = %v while the write begun before it was under way, want it to wait", err)
	case <-time.After(50 * time.Millisecond):
	}
	end()
	select {
	case err := <-waited:
		if err != nil {
			t.Errorf("wait() = %v once the write before it ended, want nil: another account's write holds nothing", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("wait() still waits once the write before it ended, want it done")
	}

	_, stillGoing := writes.begin("masha")
	defer stillGoing()
	gaveUp, cancel := context.WithCancel(t.Context())
	cancel()
	if err := writes.wait(gaveUp, "masha"); !errors.Is(err, context.Canceled) {
		t.Errorf("wait() for a caller that gave up = %v, want context.Canceled", err)
	}
}

// A request is held open for its writes no longer than its bound, however
// long a write takes.
func TestARequestIsHeldNoLongerThanItsBound(t *testing.T) {
	t.Parallel()

	held := &heldWrites{}
	held.add(make(chan struct{}))
	const bound = 50 * time.Millisecond
	began := time.Now()
	held.wait(bound)
	if took := time.Since(began); took < bound || took > 5*time.Second {
		t.Errorf("wait() took %v for a write that never ends, want the bound, %v", took, bound)
	}
}
