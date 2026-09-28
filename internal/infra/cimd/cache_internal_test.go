package cimd

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// someDay is when these cases happen.
var someDay = time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

func documentNamed(name string) Document {
	return Document{ClientID: clientAt, ClientName: name, RedirectURIs: []string{"https://a.example/cb"}}
}

// A kept document is fresh until its time is up, and what a caller does to the
// copy it was handed never reaches the next caller.
func TestAKeptDocumentIsFreshUntilItsTimeIsUp(t *testing.T) {
	t.Parallel()

	kept := newCache()
	kept.put(clientAt, documentNamed("A"), someDay, someDay.Add(time.Hour))

	first, fresh := kept.get(clientAt, someDay.Add(time.Hour-time.Second))
	if !fresh || first.ClientName != "A" {
		t.Fatalf("get() before its time = %+v, %v; want the document", first, fresh)
	}
	first.RedirectURIs[0] = "https://changed.example/cb"
	if second, _ := kept.get(clientAt, someDay); second.RedirectURIs[0] != "https://a.example/cb" {
		t.Errorf("get() = %+v after a caller changed its copy, want the document as kept", second)
	}
	if _, fresh := kept.get(clientAt, someDay.Add(time.Hour)); fresh {
		t.Error("get() at its time = fresh, want it gone")
	}
	if _, fresh := kept.get("https://client.example.com/oauth/other", someDay); fresh {
		t.Error("get() for an address never fetched = fresh, want nothing")
	}
}

// A full cache of fresh documents makes room for a new one by letting the one
// closest to its end go, and only that one: it never grows past its bound.
func TestAFullCacheLetsTheDocumentClosestToItsEndGo(t *testing.T) {
	t.Parallel()

	kept := newCache()
	for i := range cacheEntries {
		kept.put(addressNumber(i), documentNamed("A"), someDay, someDay.Add(time.Hour+time.Duration(i)*time.Second))
	}
	kept.put("https://client.example.com/new", documentNamed("B"), someDay, someDay.Add(time.Hour))

	if len(kept.entries) != cacheEntries {
		t.Errorf("%d documents kept, want %d", len(kept.entries), cacheEntries)
	}
	if _, fresh := kept.get(addressNumber(0), someDay); fresh {
		t.Error("the document closest to its end is still kept, want it gone")
	}
	if _, fresh := kept.get(addressNumber(1), someDay); !fresh {
		t.Error("the next closest is gone as well, want only one to go")
	}
}

// A full cache lets its stale documents go before any fresh one.
func TestAFullCacheLetsItsStaleDocumentsGoFirst(t *testing.T) {
	t.Parallel()

	kept := newCache()
	for i := range cacheEntries {
		expires := someDay.Add(time.Hour)
		if i%2 == 0 {
			expires = someDay.Add(time.Minute)
		}
		kept.put(addressNumber(i), documentNamed("A"), someDay, expires)
	}
	later := someDay.Add(2 * time.Minute)
	kept.put("https://client.example.com/new", documentNamed("B"), later, later.Add(time.Hour))

	if want := cacheEntries/2 + 1; len(kept.entries) != want {
		t.Errorf("%d documents kept, want %d: the fresh half and the new one", len(kept.entries), want)
	}
	if _, fresh := kept.get(addressNumber(1), later); !fresh {
		t.Error("a fresh document went, want only the stale ones gone")
	}
}

func addressNumber(i int) string {
	return fmt.Sprintf("https://client.example.com/%d", i)
}

// A document is kept as long as its answer allows — its max-age, less what a
// cache on the way has already kept it for — but never for less than the floor
// nor longer than the ceiling.
func TestADocumentIsKeptWithinTheFloorAndTheCeiling(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		cacheControl []string
		age          string
		want         time.Duration
	}{
		{"nothing asked", nil, "", shortestKept},
		{"claude's five minutes", []string{"public, max-age=300"}, "", 5 * time.Minute},
		{"an hour", []string{"max-age=3600"}, "", time.Hour},
		{"quoted", []string{`max-age="600"`}, "", 10 * time.Minute},
		{"in capitals", []string{"MAX-AGE=600"}, "", 10 * time.Minute},
		{"over two headers", []string{"public", "max-age=900"}, "", 15 * time.Minute},
		{"ten seconds", []string{"max-age=10"}, "", shortestKept},
		{"not to be kept", []string{"no-store"}, "", shortestKept},
		{"a year", []string{"max-age=31536000"}, "", longestKept},
		{"past what a number holds", []string{"max-age=99999999999999999999"}, "", longestKept},
		{"negative past what a number holds", []string{"max-age=-99999999999999999999"}, "", shortestKept},
		{"negative", []string{"max-age=-1"}, "", shortestKept},
		{"not a number", []string{"max-age=soon"}, "", shortestKept},
		{"a cache on the way kept ten minutes of it", []string{"max-age=3600"}, "600", 50 * time.Minute},
		{"a cache on the way kept most of it", []string{"max-age=3600"}, "3500", shortestKept},
		{"a cache on the way kept all of it and more", []string{"max-age=3600"}, "99999999999999999999", shortestKept},
		{"an age with no max-age", nil, "600", shortestKept},
		{"an age that is not a number", []string{"max-age=3600"}, "soon", time.Hour},
		{"a year kept a day already", []string{"max-age=31536000"}, "86400", longestKept},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			for _, value := range tc.cacheControl {
				header.Add("Cache-Control", value)
			}
			if tc.age != "" {
				header.Set("Age", tc.age)
			}
			if got := lifetimeOf(header); got != tc.want {
				t.Errorf("lifetimeOf(%q, Age %q) = %v, want %v", tc.cacheControl, tc.age, got, tc.want)
			}
		})
	}
}
