package cimd

import (
	"errors"
	"net/url"
	"strings"
	"testing"
)

// A client identifier is fetched only when it is an address of the kind the
// rules allow; anything else is refused before a name is resolved.
func TestOnlyAClientAddressIsFetched(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		clientID string
		fetched  bool
	}{
		{"claude's own", "https://claude.ai/oauth/mcp-oauth-client-metadata", true},
		{"with a port", "https://client.example.com:8443/meta.json", true},
		{"with a query", "https://client.example.com/meta?v=2", true},
		{"a literal address", "https://192.0.2.1/meta", true},

		{"plain http", "http://client.example.com/meta", false},
		{"another scheme", "ftp://client.example.com/meta", false},
		{"no path", "https://client.example.com", false},
		{"only the root", "https://client.example.com/", false},
		{"credentials", "https://user:secret@client.example.com/meta", false},
		{"a user alone", "https://user@client.example.com/meta", false},
		{"a fragment", "https://client.example.com/meta#part", false},
		{"an empty fragment", "https://client.example.com/meta#", false},
		{"two dots", "https://client.example.com/a/../meta", false},
		{"one dot", "https://client.example.com/./meta", false},
		{"dots in escapes", "https://client.example.com/%2e%2e/meta", false},
		{"an escaped slash around dots", "https://client.example.com/a%2F..%2Fmeta", false},
		{"no host", "https:///meta", false},
		{"opaque", "https:meta", false},
		{"not an address", "a client", false},
		{"an address that does not parse", "https://[::1/meta", false},
		{"nothing", "", false},
		{"too long", "https://client.example.com/" + strings.Repeat("a", maxClientURL), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := clientURL(tc.clientID)
			if fetched := err == nil; fetched != tc.fetched {
				t.Errorf("clientURL(%q) = %v; fetched: %v, want %v", tc.clientID, err, fetched, tc.fetched)
			}
			if err != nil && !errors.Is(err, ErrClientURL) {
				t.Errorf("clientURL(%q) = %v, want ErrClientURL", tc.clientID, err)
			}
		})
	}
}

// Whatever the identifier, it is either refused or an address that keeps
// every rule, and it never takes the check down.
func FuzzClientURL(f *testing.F) {
	for _, seed := range []string{
		"https://claude.ai/oauth/mcp-oauth-client-metadata",
		"https://client.example.com:8443/meta?v=2",
		"https://user@client.example.com/meta",
		"https://client.example.com/%2e%2e/meta",
		"https://[::1]:443/meta",
		"http://client.example.com/meta",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, clientID string) {
		if err := clientURL(clientID); err != nil {
			return
		}
		address, err := url.Parse(clientID)
		switch {
		case err != nil:
			t.Fatalf("clientURL(%q) = nil for something that does not parse: %v", clientID, err)
		case address.Scheme != "https", address.Hostname() == "", address.User != nil:
			t.Errorf("clientURL(%q) = nil for scheme %q, host %q, user %v", clientID, address.Scheme, address.Hostname(), address.User)
		case strings.Contains(clientID, "#"), address.Path == "", address.Path == "/", walksThePath(address.Path):
			t.Errorf("clientURL(%q) = nil for a fragment or a path %q", clientID, address.Path)
		case len(clientID) > maxClientURL:
			t.Errorf("clientURL(%q) = nil past %d characters", clientID, maxClientURL)
		}
	})
}
