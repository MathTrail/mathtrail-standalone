package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Host serves the service under its own name only. A request that arrives
// under any other — the platform's own address for the service, or a name
// somebody else pointed at it — is answered 404 with an empty body, so that
// there is exactly one issuer and one set of metadata, and the address a
// token is meant for is the only one it works at. A probe is answered under
// any name, because the platform asks by its own address.
func Host(public *url.URL) gin.HandlerFunc {
	served := comparableHost(public.Host)
	return func(c *gin.Context) {
		if isProbe(c) || comparableHost(c.Request.Host) == served {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusNotFound)
	}
}

// Origin refuses a request a browser sent from a page of any origin but the
// service's own, with 403 and an empty body. A request that carries no Origin
// passes: the header is a browser's, a chat host's own client sends none, and
// only a browser can be made to send a request on somebody else's behalf.
func Origin(public *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" || sameOrigin(origin, public) {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusForbidden)
	}
}

// sameOrigin says whether an Origin header names the service's own origin. A
// value that is not an origin at all, such as the "null" a sandboxed page
// sends, names nobody.
func sameOrigin(origin string, public *url.URL) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Scheme, public.Scheme) &&
		comparableHost(parsed.Host) == comparableHost(public.Host)
}

// Sender is the address a request came from, as the platform saw it.
//
// That is the last entry of X-Forwarded-For, the hop the platform appends;
// anything a client wrote into the header stands before it, so whatever a
// client claims there is never read. A request that reached the process with
// no such header — on a developer's machine, with no platform in front — came
// from the address of its connection. An entry that is no address names
// nobody, and the connection's address stands in for it; it is false when that
// is no address either.
func Sender(r *http.Request) (netip.Addr, bool) {
	if forwarded := r.Header.Values("X-Forwarded-For"); len(forwarded) > 0 {
		hops := forwarded[len(forwarded)-1]
		if address, readable := addressOf(hops[strings.LastIndexByte(hops, ',')+1:]); readable {
			return address, true
		}
	}
	return addressOf(r.RemoteAddr)
}

// clientAddress is who a request is counted as, before anybody has signed in:
// the address it came from, as Sender reads it, or the connection's address
// as it was written when nothing reads as one.
//
// An IPv6 address is counted by its /64, the network one household is given:
// counted whole, a single household could spend a fresh allowance on every
// address it has.
func clientAddress(r *http.Request) string {
	address, readable := Sender(r)
	switch {
	case !readable:
		return r.RemoteAddr
	case address.Is4():
		return address.String()
	}
	return netip.PrefixFrom(address, 64).Masked().String()
}

// addressOf reads one address, with a port after it or none, as plain IPv4 or
// IPv6: an IPv4 address written as IPv6 is read as IPv4, and a zone is dropped.
func addressOf(text string) (netip.Addr, bool) {
	text = strings.TrimSpace(text)
	address, err := netip.ParseAddr(text)
	if err != nil {
		withPort, portErr := netip.ParseAddrPort(text)
		if portErr != nil {
			return netip.Addr{}, false
		}
		address = withPort.Addr()
	}
	return address.Unmap().WithZone(""), true
}

// comparableHost is a host written the way two names for it are compared: in
// lower case, and without a port that is the default of either scheme, since
// a client may write that port or leave it out.
func comparableHost(host string) string {
	host = strings.ToLower(host)
	name, port, err := net.SplitHostPort(host)
	if err != nil || (port != "80" && port != "443") {
		return host
	}
	if strings.Contains(name, ":") {
		return "[" + name + "]"
	}
	return name
}
