package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// metricScope is what these measurements are recorded under.
const metricScope = "github.com/MathTrail/mathtrail-standalone/internal/transport/http"

// durationBuckets are where an answer's time is worth distinguishing. The
// service promises an answer inside a second, so the buckets are dense on
// either side of that and stop caring in detail past it.
var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

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
		if isProbe(c) {
			c.Next()
			return
		}

		started := time.Now()

		c.Next()

		// The labels are made after the handlers, because the status they
		// carry is known only then.
		labels := metric.WithAttributes(
			attribute.String("http.route", route(c)),
			attribute.String("http.request.method", method(c)),
			attribute.Int("http.response.status_code", c.Writer.Status()),
		)
		requests.Add(c.Request.Context(), 1, labels)
		duration.Record(c.Request.Context(), time.Since(started).Seconds(), labels)
	}, nil
}
