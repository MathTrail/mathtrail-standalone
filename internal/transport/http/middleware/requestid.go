// Package middleware holds the gin middleware every request passes through.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header a request is correlated by.
	RequestIDHeader = "X-Request-ID"

	// RequestIDKey is the gin context key holding the request id.
	RequestIDKey = "request_id"
)

// maxRequestIDLength bounds what a client may dictate: an id is a correlation
// handle, and an unbounded one is a way to write anything into a log line.
const maxRequestIDLength = 64

// RequestID gives every request an id and echoes it back. A client that sends
// a usable one keeps it, so that a trace through somebody else's logs and a
// trace through ours use the same word; anything else is replaced.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if !usableRequestID(id) {
			id = newRequestID()
		}
		c.Set(RequestIDKey, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// usableRequestID reports whether a client's id can be carried as it stands.
//
// The alphabet is deliberately narrow: letters, digits and the four separators
// every correlation format in the wild is built from — a UUID, a hex trace id
// and a W3C traceparent all fit, and nothing else needs to. What it keeps out
// is a value that means something to a reader rather than to a machine: a
// newline, a control byte, a quote. Neither our JSON log nor the header writer
// would be fooled by one today, but the id also travels into whatever we
// correlate with next, and this is the one place that can be said once.
func usableRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for _, b := range []byte(id) {
		switch {
		case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		case b == '-', b == '_', b == '.', b == ':':
		default:
			return false
		}
	}
	return true
}

// newRequestID prefers a version 7 identifier, whose leading bits are a
// timestamp, so that ids sort by the order the requests arrived.
func newRequestID() string {
	if v7, err := uuid.NewV7(); err == nil {
		return v7.String()
	}
	return uuid.New().String()
}
