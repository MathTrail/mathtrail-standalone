package middleware

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

// Limit holds a request, by the address it came from, to each ceiling given in
// turn, and answers at the first it is over: when to ask again, that no cache
// may keep the answer, and the answer given — a page for a person's browser,
// the router's refusal for a program. A request a later ceiling turns away is
// given back to the ones before it, so that it costs the address nothing.
//
// A ceiling reached is written down once for a flood, as a limit_hit warning
// naming the ceiling. The line never carries the address: it is somebody's.
func Limit(logger *zap.Logger, projectID string, answer gin.HandlerFunc, ceilings ...ratelimit.Ceiling) gin.HandlerFunc {
	return func(c *gin.Context) {
		refusedBy, verdict := ratelimit.Admit(clientAddress(c.Request), ceilings...)
		if verdict.Allowed {
			c.Next()
			return
		}
		if verdict.Began {
			logger.Warn("limit_hit", append([]zap.Field{
				zap.String("limit", refusedBy),
				zap.String("request_id", RequestIDFrom(c)),
			}, telemetry.LogFields(c.Request.Context(), projectID)...)...)
		}
		c.Header("Retry-After", strconv.Itoa(max(1, int(math.Ceil(verdict.RetryAfter.Seconds())))))
		c.Header("Cache-Control", "no-store")
		answer(c)
		c.Abort()
	}
}

// TooMany answers a request past its pace the way the router refuses anything
// on its own account: 429, in the shape every such refusal has.
func TooMany(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, apierror.Response{
		Code:    apierror.CodeTooManyRequests,
		Message: "too many requests; try again in a moment",
	})
}

// Page answers a request past its pace with the page given, which says so in
// words a person reads, and with the status the page gives.
func Page(page http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		page.ServeHTTP(c.Writer, c.Request)
	}
}
