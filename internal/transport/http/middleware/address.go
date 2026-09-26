package middleware

import (
	"net"
	"net/http"
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
