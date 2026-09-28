package oauthserver

import (
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

// maxRedirectURI is the longest address a sign-in may send the parent back
// to. The addresses clients use are short, and a longer one is refused rather
// than carried inside every identifier and every request it would ride in.
const maxRedirectURI = 512

// redirectable reports whether a sign-in may ever send the parent back to an
// address: an absolute one over https, or over plain http to the parent's own
// computer, where a client running on it listens. Never one with credentials
// or a fragment, which make an address mean something other than it shows.
func redirectable(uri string) bool {
	if len(uri) > maxRedirectURI || strings.Contains(uri, "#") {
		return false
	}
	address, err := url.Parse(uri)
	if err != nil || address.Opaque != "" || address.User != nil || address.Hostname() == "" {
		return false
	}
	switch address.Scheme {
	case "https":
		return true
	case "http":
		return loopbackHost(address.Hostname())
	}
	return false
}

// loopbackHost reports whether a host is the computer the browser runs on. A
// name is compared without regard to case, as names are.
func loopbackHost(host string) bool {
	return strings.EqualFold(host, "localhost") || host == "127.0.0.1" || host == "::1"
}

// asciiHost is a host name spelled as IDNA spells it for a browser to look
// up, and whether IDNA spells it at all.
func asciiHost(name string) (string, bool) {
	ascii, err := idna.Lookup.ToASCII(name)
	return ascii, err == nil
}

// hostOf is the host of an address, and nothing else of it, as a page or a
// line names where a client is or where it sends the parent back to. It is
// written in ASCII, a name as IDNA spells it: a name in another script cannot
// pass for a familiar one there — a Cyrillic letter that looks like a Latin
// one is spelled out as what it is — and a mark that turns text around, which
// an address may carry escaped, cannot reorder the page or the line it is read
// in. A host IDNA does not spell has everything but ASCII escaped, and only
// its ASCII letters written small, the one case IDNA writes: a letter of
// another script lowered could become an ASCII one, as a capital dotted I
// becomes an i.
func hostOf(uri string) string {
	address, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	if ascii, spelled := asciiHost(address.Hostname()); spelled {
		return ascii
	}
	lowered := strings.Map(func(c rune) rune {
		if 'A' <= c && c <= 'Z' {
			return c + 'a' - 'A'
		}
		return c
	}, address.Hostname())
	quoted := strconv.QuoteToASCII(lowered)
	return quoted[1 : len(quoted)-1]
}
