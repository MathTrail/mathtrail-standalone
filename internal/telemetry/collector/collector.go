// Package collector holds the one rule the address of a telemetry collector is
// held to, apart from the exporters that post to it: whatever checks the
// address — the configuration, the telemetry itself — checks it with this,
// without carrying the exporters along.
package collector

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Root reads a collector's root address and holds it to the rule.
//
// A root is http or https, names a host, and may carry a path for the signals'
// own paths to go under, as a collector behind a prefix needs. It carries
// nothing else: a user or a password has no business in it, since the exporter
// signs its own requests, and a query or a fragment would end up after the
// signals' paths, where nobody meant it.
//
// An @ is where a user and a password end, and a parser that reads a forgotten
// scheme as the user's name, or cannot read the address at all, never finds
// them. So an address with one is refused, without being repeated, unless it
// was read as a scheme and a host with no user — where an @ is part of a path.
func Root(raw string) (*url.URL, error) {
	root, err := url.Parse(raw)
	if strings.Contains(raw, "@") && (err != nil || root.User != nil || root.Host == "") {
		return nil, errors.New("collector: the address carries a user or a password, and the exporter signs its own requests")
	}
	switch {
	case err != nil:
		return nil, fmt.Errorf("collector: %s is not a URL", shown(raw))
	case root.Scheme != "http" && root.Scheme != "https":
		return nil, fmt.Errorf("collector: %s must be http or https", shown(raw))
	case root.Host == "":
		return nil, fmt.Errorf("collector: %s has no host", shown(raw))
	case root.RawQuery != "" || root.ForceQuery || root.Fragment != "":
		return nil, fmt.Errorf("collector: %s has a query or a fragment, and the signals' paths go at its end", shown(raw))
	}
	return root, nil
}

// shown is how a refusal names an address: whole when it holds nothing a
// secret could hide in, and only as "the address" when it holds a query or a
// fragment, which is where a key would be.
func shown(raw string) string {
	if strings.ContainsAny(raw, "?#") {
		return "the address"
	}
	return fmt.Sprintf("address %q", raw)
}
