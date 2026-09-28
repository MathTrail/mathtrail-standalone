package drivestore

import "sync"

// maxRemembered is how many accounts' files an instance remembers at once.
// Past it, one it remembers is forgotten for each it learns: a memory that
// could grow without bound is itself a way to bring an instance down, and a
// file forgotten costs one search the next time it is needed.
const maxRemembered = 4096

// fileIDs remembers which file holds whose profile while the instance runs,
// keyed by the identifier the sign-in derived. It is a memory in the strict
// sense: what it holds may be wrong by the time it is used — the file
// deleted, the profile moved to another — and whoever finds that out forgets
// the entry and searches again. Nothing durable is keyed by the identifier,
// which changes at the first sign-in after the keys rotate.
type fileIDs struct {
	mu  sync.Mutex
	ids map[string]string
}

// newFileIDs is a memory that holds nothing yet.
func newFileIDs() *fileIDs {
	return &fileIDs{ids: map[string]string{}}
}

// recall is the file remembered for an account.
func (c *fileIDs) recall(account string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	id, remembered := c.ids[account]
	return id, remembered
}

// remember keeps the file that holds an account's profile. When the memory is
// full, some other account's file is forgotten to make room: which one does
// not matter, since any of them costs one search to learn again.
func (c *fileIDs) remember(account, id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, known := c.ids[account]; !known && len(c.ids) >= maxRemembered {
		for someone := range c.ids {
			delete(c.ids, someone)
			break
		}
	}
	c.ids[account] = id
}

// forget drops what is remembered for an account, but only while it is still
// the file found wrong: a call that learnt of another file in the meantime
// keeps what it learnt.
func (c *fileIDs) forget(account, id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ids[account] == id {
		delete(c.ids, account)
	}
}
