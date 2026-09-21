package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
)

// panicStackSkip drops the recovery machinery from the top of the logged
// stack, so it starts at the code that actually panicked.
const panicStackSkip = 3

// ZapRecovery turns a panic into a logged error and a plain JSON answer. One
// request must not take the process down: another child is in the middle of a
// task on the same instance.
//
// One kind of panic never reaches here: net/http documents ErrAbortHandler as
// the way to drop a request silently, and the framework aborts on it without
// calling this, alongside a connection that died under it. A test holds it to
// that, because answering 500 to a silent abort would break a contract of the
// standard library.
func ZapRecovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		logger.Error("panic",
			zap.Any("error", recovered), // zap routes an error here to the same field zap.Error would
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("request_id", RequestIDFrom(c)),
			zap.StackSkip("stack", panicStackSkip),
		)

		// A handler that panicked halfway through its answer has already sent
		// a status and headers. Writing a second one corrupts what the client
		// is reading; all that is left to do is stop.
		if c.Writer.Written() {
			c.Abort()
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, apierror.Response{
			Code:    apierror.CodeInternal,
			Message: "an unexpected error occurred",
		})
	})
}

// RequestIDFrom returns the request id set by RequestID, or an empty string.
func RequestIDFrom(c *gin.Context) string {
	if value, ok := c.Get(RequestIDKey); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}
