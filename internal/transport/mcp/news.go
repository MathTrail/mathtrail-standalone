package mcpserver

import (
	"context"
	"slices"
	"sync"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// A card waiting for its task asks how the task stands, and a question that
// would only hear what the card knows already is held until there is news. What
// changes a task being written is a write of its account — the task accepted,
// a try turned down — so a write that lands on this instance is news to the
// questions held here, and hands them the profile it wrote. One that lands on
// another instance is heard at the card's next question.

// news are the questions this instance holds open for news of an account, by
// account: each an ear the account's next write hands the profile it wrote.
type news struct {
	mu      sync.Mutex
	waiting map[string][]chan []byte
}

// newNews is an instance with no question held yet.
func newNews() *news {
	return &news{waiting: map[string][]chan []byte{}}
}

// listen is a question's ear for the account's next write, and what takes it
// away. A write takes away the ears it finds, and hands each the profile it
// wrote; an ear taken away first hears nothing.
func (n *news) listen(account string) (heard <-chan []byte, stop func()) {
	ear := make(chan []byte, 1)
	n.mu.Lock()
	n.waiting[account] = append(n.waiting[account], ear)
	n.mu.Unlock()
	return ear, func() {
		n.mu.Lock()
		defer n.mu.Unlock()
		left := slices.DeleteFunc(n.waiting[account], func(other chan []byte) bool { return other == ear })
		if len(left) == 0 {
			delete(n.waiting, account)
		} else {
			n.waiting[account] = left
		}
	}
}

// tell is a write of the account that has landed: every question listening
// for it is handed the profile as written, and listens no more. The profile is
// written down only when a question listens, and handed over as its bytes, so
// that the question reads its own copy while the writer goes on; one that
// cannot be written down is handed as nothing. Each ear is handed one profile,
// and has room for it, so a write never waits for a question.
func (n *news) tell(account string, written *profile.Profile) {
	n.mu.Lock()
	ears := n.waiting[account]
	delete(n.waiting, account)
	n.mu.Unlock()
	if len(ears) == 0 {
		return
	}
	raw, err := profile.Marshal(written)
	if err != nil {
		raw = nil
	}
	for _, ear := range ears {
		ear <- raw
	}
}

// telling is the store as every write of this instance reaches it: a profile
// written is news to the questions held for its account. Only a save tells. A
// question is held only over a profile that reads, with a task being written,
// and a first profile, a file put back or a new start is never written over
// such a profile.
type telling struct {
	store.Storage
	news *news
}

// Save writes the profile, and once it has landed tells the account's
// questions, handing them the profile written.
func (t telling) Save(ctx context.Context, account store.Account, p *profile.Profile, expected store.Revision) (store.Revision, error) {
	revision, err := t.Storage.Save(ctx, account, p, expected)
	if err == nil {
		t.news.tell(account.ID, p)
	}
	return revision, err
}
