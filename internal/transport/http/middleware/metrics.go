package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// metricScope is what these measurements are recorded under.
const metricScope = "github.com/MathTrail/mathtrail-standalone/internal/transport/http"

// otherLabel stands for anything outside the closed sets below. Every label of
// every metric is drawn from a list this service wrote down, because a label
// that can take any value turns one series into as many as a stranger cares to
// invent — and the allowance these are counted against is measured in bytes.
const otherLabel = "other"

// durationBuckets are where an answer's time is worth distinguishing. The
// service promises an answer inside a second, so the buckets are dense on
// either side of that and stop caring in detail past it.
var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// knownMethods is the closed set a request's method is reported as. Anything
// else is a caller inventing a verb, and it is counted as one kind.
var knownMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodPatch:   {},
	http.MethodDelete:  {},
	http.MethodOptions: {},
	http.MethodConnect: {},
	http.MethodTrace:   {},
}

// Metrics counts the requests and times them.
//
// Both measurements carry the same three labels, so that one can be read
// against the other: the route as this service declared it, the method, and
// the status that was answered.
func Metrics(meters metric.MeterProvider) (gin.HandlerFunc, error) {
	meter := meters.Meter(metricScope)

	requests, err := meter.Int64Counter(
		"http_request",
		metric.WithDescription("requests answered"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("middleware: http_request: %w", err)
	}

	duration, err := meter.Float64Histogram(
		"http_request_duration",
		metric.WithDescription("how long a request took to answer"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(durationBuckets...),
	)
	if err != nil {
		return nil, fmt.Errorf("middleware: http_request_duration: %w", err)
	}

	return func(c *gin.Context) {
		// A probe is left out whole, failed ones included. Counting only the
		// failures would leave "requests answered" meaning something nobody
		// could state in a sentence, and a probe that fails is already a line
		// in the log and a signal the platform has of its own.
		if isProbe(c.Request.URL.Path) {
			c.Next()
			return
		}

		started := time.Now()

		c.Next()

		// The route is read after the handler, because that is when the
		// framework knows which one matched.
		labels := metric.WithAttributes(
			attribute.String("http.route", route(c)),
			attribute.String("http.request.method", method(c)),
			attribute.Int("http.response.status_code", c.Writer.Status()),
		)
		requests.Add(c.Request.Context(), 1, labels)
		duration.Record(c.Request.Context(), time.Since(started).Seconds(), labels)
	}, nil
}

// route is the pattern the request matched, never the path it asked for: a
// path carries identifiers, and an identifier in a label is a new series for
// every child who ever uses the service.
func route(c *gin.Context) string {
	if matched := c.FullPath(); matched != "" {
		return matched
	}
	return otherLabel
}

// method is the request's verb when it is one of the known ones.
func method(c *gin.Context) string {
	if _, known := knownMethods[c.Request.Method]; known {
		return c.Request.Method
	}
	return otherLabel
}
