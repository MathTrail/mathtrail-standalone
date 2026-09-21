package middleware

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// maskedQueryKeys are the query parameters whose values never reach a log.
// The sign-in carries codes, tokens and verifiers in query strings, and a log
// line is the one place they must not end up. Compared without case.
var maskedQueryKeys = map[string]struct{}{
	"access_token":  {},
	"authorization": {},
	"client_secret": {},
	"code":          {},
	"code_verifier": {},
	"id_token":      {},
	"key":           {},
	"password":      {},
	"refresh_token": {},
	"secret":        {},
	"state":         {},
	"token":         {},
}

// silentPaths answer a probe rather than a person. Logging every one of them
// buries the lines that mean something.
var silentPaths = map[string]struct{}{
	"/healthz": {},
}

// ZapLogger logs one line per request: what was asked, what was answered and
// how long it took. Successful probes are skipped; anything that failed is
// always logged.
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		if _, silent := silentPaths[path]; silent && status < 400 {
			return
		}

		// A handler that wrote no body leaves the size at -1, which reads as
		// nonsense in a log line; an answer with no body has a body of zero.
		bodySize := max(c.Writer.Size(), 0)

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Duration("duration", time.Since(start)),
			zap.Int("body_size", bodySize),
			zap.String("request_id", RequestIDFrom(c)),
		}
		if query != "" {
			fields = append(fields, zap.String("query", maskQuery(query)))
		}
		if errs := c.Errors.ByType(gin.ErrorTypePrivate); len(errs) > 0 {
			fields = append(fields, zap.String("error", errs.String()))
		}

		switch {
		case status >= 500:
			logger.Error("http_request", fields...)
		case status >= 400:
			logger.Warn("http_request", fields...)
		default:
			logger.Info("http_request", fields...)
		}
	}
}

// maskQuery keeps the shape of a query string and drops the secrets in it. A
// query that cannot be parsed is dropped whole rather than guessed at.
//
// It always parses, and deliberately has no shortcut for queries that look
// free of secrets: a parameter name may be percent-encoded, so "%63ode=abc"
// carries a code that no search of the raw text finds. Parsing is the only
// reading of a query that agrees with the one the handlers get.
func maskQuery(raw string) string {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return "unparsable"
	}
	for key, list := range values {
		if _, masked := maskedQueryKeys[strings.ToLower(key)]; !masked {
			continue
		}
		for i := range list {
			list[i] = "masked"
		}
	}
	return values.Encode()
}
