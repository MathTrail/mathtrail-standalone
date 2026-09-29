package cimd

import (
	"fmt"
	"net/url"
	"strings"
)

// maxClientURL is the longest client identifier fetched. The addresses clients
// name themselves by are short, and a longer one is refused before anything is
// resolved.
const maxClientURL = 1024

// clientURL checks that a client identifier is an address a document may be
// fetched from, by the rules a client identifier keeps: https, a host, a path
// of its own, and nothing that makes the address mean something other than
// what it shows — no credentials, no fragment, no segment that stays where it
// is or walks the path back up. A query is allowed: the rules advise against
// one without forbidding it.
//
// The rules allow a port as well; only the one https has of its own is taken
// here. The document is served where every web page is, and any other port
// would let a stranger's identifier point the service at whatever else a
// public address runs.
func clientURL(clientID string) error {
	if len(clientID) > maxClientURL {
		return fmt.Errorf("%w: longer than %d characters", ErrClientURL, maxClientURL)
	}
	address, err := url.Parse(clientID)
	switch {
	case err != nil:
		return fmt.Errorf("%w: not an address", ErrClientURL)
	case address.Scheme != "https":
		return fmt.Errorf("%w: not https", ErrClientURL)
	case address.Hostname() == "":
		return fmt.Errorf("%w: no host", ErrClientURL)
	case address.Port() != "" && address.Port() != "443":
		return fmt.Errorf("%w: a port other than https's own", ErrClientURL)
	case address.User != nil:
		return fmt.Errorf("%w: carries credentials", ErrClientURL)
	case address.Fragment != "" || strings.Contains(clientID, "#"):
		return fmt.Errorf("%w: has a fragment", ErrClientURL)
	case address.Path == "" || address.Path == "/":
		return fmt.Errorf("%w: no path of its own", ErrClientURL)
	case walksThePath(address.Path):
		return fmt.Errorf("%w: a path segment of dots", ErrClientURL)
	}
	return nil
}

// walksThePath reports whether a path has a segment of one dot or two, which a
// server may resolve against the segments around it. The path is the decoded
// one, so a segment spelt in escapes is caught as well.
func walksThePath(path string) bool {
	for segment := range strings.SplitSeq(path, "/") {
		if segment == "." || segment == ".." {
			return true
		}
	}
	return false
}
