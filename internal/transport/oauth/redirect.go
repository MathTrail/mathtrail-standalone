package oauthserver

import (
	"net/netip"
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
		return plainLoopback(address.Hostname())
	}
	return false
}

// plainLoopback reports whether a host is one a sign-in may send the parent
// back to over plain http: the parent's own computer, by one of the three
// names a client listening there gives it — localhost, in any case, as names
// are compared, 127.0.0.1 or ::1.
func plainLoopback(host string) bool {
	return strings.EqualFold(host, "localhost") || host == "127.0.0.1" || host == "::1"
}

// loopbackOrigins are how an address on the parent's own computer begins when
// its port is the client's to choose: plain http, to one of the three names
// a client listening there gives it, written as clients write them.
var loopbackOrigins = []string{"http://localhost", "http://127.0.0.1", "http://[::1]"}

// sameButThePort reports whether two addresses on the parent's own computer
// are the same but for their ports: each begins with one of loopbackOrigins,
// and has a port of digits after it or none, and the rest of the two is the
// same byte for byte.
func sameButThePort(registered, requested string) bool {
	first, isFirst := withoutPort(registered)
	second, isSecond := withoutPort(requested)
	return isFirst && isSecond && first == second
}

// withoutPort is an address on the parent's own computer with its port taken
// out, and whether it is one: one of loopbackOrigins, a colon and a port of
// one to five digits or no port at all, and then a path, a query or nothing —
// so that a host that merely begins like the computer's, such as
// localhost.example, is no address of it.
func withoutPort(uri string) (string, bool) {
	for _, origin := range loopbackOrigins {
		rest, isOrigin := strings.CutPrefix(uri, origin)
		if !isOrigin {
			continue
		}
		if port, hasPort := strings.CutPrefix(rest, ":"); hasPort {
			digits := len(port) - len(strings.TrimLeft(port, "0123456789"))
			if digits == 0 || digits > 5 {
				return "", false
			}
			rest = port[digits:]
		}
		if rest != "" && rest[0] != '/' && rest[0] != '?' {
			return "", false
		}
		return origin + rest, true
	}
	return "", false
}

// toThisComputer reports whether an address sends the parent back to the
// computer their browser runs on, however the address names it: as written,
// or as IDNA spells it for a browser to look up, which reads a name or an
// address in full-width letters and digits as the one in ASCII.
func toThisComputer(uri string) bool {
	address, err := url.Parse(uri)
	if err != nil {
		return false
	}
	host := address.Hostname()
	spelled, isSpelled := asciiHost(host)
	return loopbackHost(host) || isSpelled && loopbackHost(spelled)
}

// loopbackHost reports whether a host is the computer the browser runs on: an
// address of the loopback, or the unspecified one that reaches the same
// computer, however it is written — a browser reads 127.1 and 2130706433 as
// 127.0.0.1 — or the name localhost or a name under it, which a browser
// answers with the loopback itself, in any case and with or without the dot
// that ends a name.
func loopbackHost(host string) bool {
	if addr, isAddress := addressOf(host); isAddress {
		addr = addr.Unmap()
		return addr.IsLoopback() || addr.IsUnspecified()
	}
	name := strings.ToLower(strings.TrimSuffix(host, "."))
	return name == "localhost" || strings.HasSuffix(name, ".localhost")
}

// addressOf is the address a browser reads a host as, when it reads it as one:
// IPv6 as written, and IPv4 in any form a browser takes — one to four
// numbers, each in decimal, in octal behind a 0 or in hexadecimal behind 0x,
// the last filling the bytes the ones before it leave.
func addressOf(host string) (netip.Addr, bool) {
	if strings.Contains(host, ":") {
		addr, err := netip.ParseAddr(host)
		return addr, err == nil
	}
	parts := strings.Split(strings.TrimSuffix(host, "."), ".")
	if len(parts) > 4 {
		return netip.Addr{}, false
	}
	var address uint64
	for i, part := range parts {
		number, isNumber := ipv4Number(part)
		last := i == len(parts)-1
		switch {
		case !isNumber, !last && number > 255, last && number >= 1<<(8*(5-len(parts))):
			return netip.Addr{}, false
		case last:
			address += number
		default:
			address += number << (8 * (3 - i))
		}
	}
	return netip.AddrFrom4([4]byte{byte(address >> 24), byte(address >> 16), byte(address >> 8), byte(address)}), true
}

// ipv4Number reads one part of an IPv4 address as a browser does: decimal, or
// octal behind a 0, or hexadecimal behind 0x, where 0x alone is zero.
func ipv4Number(part string) (uint64, bool) {
	base := 10
	switch {
	case len(part) >= 2 && (part[:2] == "0x" || part[:2] == "0X"):
		part, base = part[2:], 16
		if part == "" {
			return 0, true
		}
	case len(part) >= 2 && part[0] == '0':
		part, base = part[1:], 8
	}
	number, err := strconv.ParseUint(part, base, 32)
	return number, err == nil
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
