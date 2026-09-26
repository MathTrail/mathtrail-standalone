package middleware

import "github.com/gin-gonic/gin"

// probeRoutes answer a machine rather than a person. The platform asks for them
// constantly and the answer is the same every time.
var probeRoutes = map[string]struct{}{
	"/health": {},
}

// isProbe reports whether a request matched one of those.
//
// Three middlewares ask, and all three leave a probe out for the same reason:
// there are far more of them than there are children, so a probe in the log
// buries the lines that mean something, a probe in a trace is a span nobody
// will ever open, and a probe in the counters makes "requests answered" a
// measure of how often the platform checked rather than of how much the
// service was used.
//
// It asks by the route, as everything else here names a request, so that the
// path a caller wrote is never read at all.
func isProbe(c *gin.Context) bool {
	_, probe := probeRoutes[route(c)]
	return probe
}
