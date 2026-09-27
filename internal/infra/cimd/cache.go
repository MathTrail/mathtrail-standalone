package cimd

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// shortestKept is how long a document is kept however its answer asked,
	// so that the sign-ins after a first one are answered from memory rather
	// than each fetching the document again.
	shortestKept = 5 * time.Minute
	// longestKept is how long a document is kept at most, so that a client
	// that changes its document is not held to the old one past a day.
	longestKept = 24 * time.Hour
	// cacheEntries is how many documents are kept at once. The hosts that
	// connect are a handful; the bound is what stops a stream of addresses
	// nobody uses twice from growing the process's memory.
	cacheEntries = 256
)

// cache keeps the documents that were good, by the identifier they were
// fetched for, until each one's time is up. A failure is never kept: the next
// request for it is a fetch.
type cache struct {
	mu      sync.Mutex
	entries map[string]entry
}

// entry is one kept document and when it stops being fresh.
type entry struct {
	document Document
	expires  time.Time
}

func newCache() *cache {
	return &cache{entries: make(map[string]entry)}
}

// get returns a copy of the document kept for an identifier while it is
// fresh, so that nothing a caller does to it reaches the next caller.
func (c *cache) get(clientID string, now time.Time) (Document, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	kept, found := c.entries[clientID]
	if !found || !now.Before(kept.expires) {
		return Document{}, false
	}
	document := kept.document
	document.RedirectURIs = slices.Clone(document.RedirectURIs)
	return document, true
}

// put keeps a document until it expires. A new identifier in a full cache
// makes room first: every document whose time is up goes, and if that frees
// nothing, the one closest to its end.
func (c *cache) put(clientID string, document Document, now, expires time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.entries[clientID]; !found && len(c.entries) >= cacheEntries {
		c.makeRoom(now)
	}
	document.RedirectURIs = slices.Clone(document.RedirectURIs)
	c.entries[clientID] = entry{document: document, expires: expires}
}

// makeRoom frees at least one place in a full cache.
func (c *cache) makeRoom(now time.Time) {
	soonest, found := "", false
	for clientID, kept := range c.entries {
		if !now.Before(kept.expires) {
			delete(c.entries, clientID)
			continue
		}
		if !found || kept.expires.Before(c.entries[soonest].expires) {
			soonest, found = clientID, true
		}
	}
	if len(c.entries) >= cacheEntries {
		delete(c.entries, soonest)
	}
}

// lifetimeOf is how long an answer's document is kept: what its max-age
// allows, less the time a cache on the way has already kept it for, within
// the floor and the ceiling above. An answer that gives no max-age, or asks not
// to be kept at all, is kept for the floor.
func lifetimeOf(header http.Header) time.Duration {
	fresh, given := maxAge(header)
	if !given {
		return shortestKept
	}
	kept, _ := seconds(header.Get("Age"))
	return min(max(time.Duration(fresh-kept)*time.Second, shortestKept), longestKept)
}

// maxAge is the max-age an answer's Cache-Control gives, in seconds.
func maxAge(header http.Header) (int64, bool) {
	for directive := range strings.SplitSeq(strings.Join(header.Values("Cache-Control"), ","), ",") {
		name, value, _ := strings.Cut(strings.TrimSpace(directive), "=")
		if strings.EqualFold(name, "max-age") {
			return seconds(value)
		}
	}
	return 0, false
}

// mostSeconds is as many seconds as a header is read as giving: more than any
// document is kept for, and few enough to count in a duration.
const mostSeconds = 1 << 32

// seconds reads a count of seconds as a cache header writes one. A count too
// long to hold is read as the most there could be; anything else that is not a
// count is not one.
func seconds(value string) (int64, bool) {
	count, err := strconv.ParseInt(strings.Trim(strings.TrimSpace(value), `"`), 10, 64)
	switch {
	case errors.Is(err, strconv.ErrRange) && count > 0:
		return mostSeconds, true
	case err != nil || count < 0:
		return 0, false
	}
	return min(count, mostSeconds), true
}
