package drivestore

import (
	"sync"
	"time"
)

const (
	// maxRemembered is how many accounts' files an instance remembers at
	// once. Past it, one it remembers is forgotten for each it learns: a
	// memory that could grow without bound is itself a way to bring an
	// instance down, and a file forgotten costs one search the next time it
	// is needed.
	maxRemembered = 4096
	// maxAge is how long a file found by a search is trusted before it is
	// searched for again. Drive answers a file in the bin by its ID as it
	// answers any other, and only a search leaves the bin out: so a parent
	// who put the profile in the bin is seen by every instance within this
	// long, at the cost of one search per account in as long.
	maxAge = 10 * time.Minute
)

// remembered is what an instance knows of an account's profile: the file that
// holds it, when a search found that file, the profile's number this instance
// last wrote there — none, when it wrote nothing — and whether that write put
// back a damaged file.
type remembered struct {
	file    string
	learnt  time.Time
	written int
	mended  bool
}

// fileIDs remembers which file holds whose profile while the instance runs,
// keyed by the identifier the sign-in derived. It is a memory in the strict
// sense: what it holds may be wrong by the time it is used — the file
// deleted, put in the bin, the profile moved to another — and whoever finds
// that out forgets the entry and searches again. Nothing durable is keyed by
// the identifier, which changes at the first sign-in after the keys rotate.
type fileIDs struct {
	mu      sync.Mutex
	entries map[string]remembered
}

// newFileIDs is a memory that holds nothing yet.
func newFileIDs() *fileIDs {
	return &fileIDs{entries: map[string]remembered{}}
}

// recall is what is remembered for an account, unless it was learnt longer
// ago than a file is trusted.
func (c *fileIDs) recall(account string, now time.Time) (remembered, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, known := c.entries[account]
	if !known || now.Sub(entry.learnt) >= maxAge {
		return remembered{}, false
	}
	return entry, true
}

// remember keeps the file a search found, or a creation made, to hold an
// account's profile, as learnt now. The number this instance wrote stays with
// the file when the file found is the one it wrote to — a search that finds
// the same file again says nothing of whether a read has caught up with the
// write — and goes with a file found in its place. When the memory is full,
// some other account's file is forgotten to make room: which one does not
// matter, since any of them costs one search to learn again.
func (c *fileIDs) remember(account, file string, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	known, isKnown := c.entries[account]
	if !isKnown && len(c.entries) >= maxRemembered {
		for someone := range c.entries {
			delete(c.entries, someone)
			break
		}
	}
	entry := remembered{file: file, learnt: now}
	if isKnown && known.file == file {
		entry.written, entry.mended = known.written, known.mended
	}
	c.entries[account] = entry
}

// wrote records the profile's number this instance just wrote to the file:
// a read of the file that comes back with an earlier one is a Drive that has
// not caught up with the write. It changes nothing when another file is
// remembered by now.
func (c *fileIDs) wrote(account, file string, counter int) {
	c.note(account, file, counter, false)
}

// mended records the profile's number this instance just wrote over a damaged
// file to put it back: a read of the file that comes back as damage may be one
// from before the write, as one that comes back with an earlier number is.
func (c *fileIDs) mended(account, file string, counter int) {
	c.note(account, file, counter, true)
}

// note records a write of this instance to the file it remembers, and whether
// the write put back a damaged file.
func (c *fileIDs) note(account, file string, counter int, mended bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, known := c.entries[account]; known && entry.file == file {
		entry.written, entry.mended = counter, mended
		c.entries[account] = entry
	}
}

// forget drops what is remembered for an account, but only while it is still
// the file found wrong: a call that learnt of another file in the meantime
// keeps what it learnt.
func (c *fileIDs) forget(account, file string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries[account].file == file {
		delete(c.entries, account)
	}
}
