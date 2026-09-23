package httpserver_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
)

// Telemetry that is missing a piece is refused where it is assembled, with the
// piece named.
//
// Each of the three would otherwise fail in a way of its own and none of them
// would say what was wrong: a missing tracer is silently replaced by one that
// records nothing, a missing meter stops the process here, and a missing
// delivery waits to stop it on the first request whose trace is kept.
func TestHalfFilledTelemetryIsRefusedByName(t *testing.T) {
	t.Parallel()

	complete := httpserver.Observability{
		Traces: tracenoop.NewTracerProvider(),
		Meters: metricnoop.NewMeterProvider(),
		Flush:  func(context.Context) error { return nil },
	}

	for _, c := range []struct {
		name    string
		missing func(*httpserver.Observability)
		want    string
	}{
		{name: "no tracer", missing: func(o *httpserver.Observability) { o.Traces = nil }, want: "Traces"},
		{name: "no meter", missing: func(o *httpserver.Observability) { o.Meters = nil }, want: "Meters"},
		{name: "no delivery", missing: func(o *httpserver.Observability) { o.Flush = nil }, want: "Flush"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			obs := complete
			c.missing(&obs)

			_, err := httpserver.NewRouter(httpserver.NewHealthHandler(), zap.NewNop(), obs)
			if !errors.Is(err, httpserver.ErrObservability) {
				t.Fatalf("NewRouter() error = %v, want it to refuse", err)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error = %q, want it to name %q", err, c.want)
			}
		})
	}
}
