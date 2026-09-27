package oauthserver

import (
	"net/url"
	"strings"
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
