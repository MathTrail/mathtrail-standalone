package drivestore

import (
	"context"
	"sync"
)

// turns lets the writes of one account's profile go one at a time within the
// instance, and those of different accounts side by side.
//
// A write reads the file once more and then uploads, and Drive cannot hold
// the two together; within one instance, this can. Two tabs whose calls land
// on the same instance then never meet between the read and the upload: the
// second reads what the first wrote, and is refused as a conflict rather than
// laid over it. What is left is two instances, which share nothing.
type turns struct {
	mu   sync.Mutex
	held map[string]*turn
}

// turn is one account's: a place taken by the write under way, and how many
// writes are under way or waiting, so that an account nobody writes for takes
// no memory.
type turn struct {
	taken   chan struct{}
	waiting int
}

// newTurns is turns with nobody writing yet.
func newTurns() *turns {
	return &turns{held: map[string]*turn{}}
}

// take waits for the account's turn, or for ctx to end, and answers what gives
// the turn back. A write whose caller gave up while it waited is refused with
// the context's error, and never started.
func (t *turns) take(ctx context.Context, account string) (giveBack func(), err error) {
	t.mu.Lock()
	mine := t.held[account]
	if mine == nil {
		mine = &turn{taken: make(chan struct{}, 1)}
		t.held[account] = mine
	}
	mine.waiting++
	t.mu.Unlock()

	select {
	case mine.taken <- struct{}{}:
		return func() {
			<-mine.taken
			t.leave(account, mine)
		}, nil
	case <-ctx.Done():
		t.leave(account, mine)
		return nil, ctx.Err()
	}
}

// leave counts a write out of the account's turn, and forgets the turn when
// nobody else is in it.
func (t *turns) leave(account string, mine *turn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	mine.waiting--
	if mine.waiting == 0 {
		delete(t.held, account)
	}
}
