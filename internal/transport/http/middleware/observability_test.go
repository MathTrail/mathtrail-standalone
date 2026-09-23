package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
)

// What a caller reveals about itself just by asking. None of it may end up in
// a trace: the service's own log carries none of it, and a span is kept
// somewhere else again, under a retention nobody told a parent about.
const (
	callerAddress  = "203.0.113.9"
	peerAddress    = "192.0.2.1"
	callerAgent    = "a-browser-nobody-wrote"
	callerPassword = "hunter2"
)

// These build the real router, because the order these handlers run in is
// what half of them are for, and that order is the router's to decide.

// The guard. Nothing a request carried about who sent it, and nothing a child
// or a task could be recognised by, may appear in a span's attributes.
func TestASpanCarriesNothingThatIdentifiesTheCaller(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.get(t, "/no-such-endpoint?token="+callerPassword)

	spans := watched.spans.Ended()
	if len(spans) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(spans))
	}

	forbidden := []string{callerAddress, peerAddress, callerAgent, callerPassword}
	for _, attr := range spans[0].Attributes() {
		value := attr.Value.String()
		for _, secret := range forbidden {
			if strings.Contains(value, secret) {
				t.Errorf("span attribute %s = %q, want it to carry no %q", attr.Key, value, secret)
			}
		}
	}
}

// The three attributes an off-the-shelf instrumentation adds by itself keep
// their keys and lose their values, so that a reader can see that something
// was deliberately left out rather than never recorded.
func TestTheAddressAndTheAgentAreBlankedRatherThanDropped(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.get(t, "/no-such-endpoint")

	spans := watched.spans.Ended()
	if len(spans) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(spans))
	}

	blanked := map[string]bool{
		"client.address":       false,
		"network.peer.address": false,
		"user_agent.original":  false,
	}
	for _, attr := range spans[0].Attributes() {
		if _, watchedKey := blanked[string(attr.Key)]; !watchedKey {
			continue
		}
		if attr.Value.String() != "redacted" {
			t.Errorf("%s = %q, want it blanked", attr.Key, attr.Value.String())
		}
		blanked[string(attr.Key)] = true
	}
	for key, seen := range blanked {
		if !seen {
			t.Errorf("%s was not on the span at all, want it kept and blanked", key)
		}
	}
}

// A path nobody declared is counted as one kind. Counting it as itself would
// let a stranger invent a new series with every request, and the allowance
// these are measured against is bytes.
func TestAnUndeclaredRouteIsCountedAsOther(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.get(t, "/no-such-endpoint")

	for _, labels := range watched.labels(t, "http_request") {
		route, found := labels.Value("http.route")
		if !found {
			t.Fatal("http_request carries no route, want one")
		}
		if route.String() != "other" {
			t.Errorf("http.route = %q, want %q", route.String(), "other")
		}
	}
}

// Nothing a request carried may become a label either: a label is a series,
// and a series that a caller can invent is an allowance a caller can spend.
func TestMeasurementsCarryNothingACallerChose(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.get(t, "/no-such-endpoint?token="+callerPassword)

	forbidden := []string{callerAddress, peerAddress, callerAgent, callerPassword, "no-such-endpoint"}
	for _, name := range []string{"http_request", "http_request_duration"} {
		for _, labels := range watched.labels(t, name) {
			for _, label := range labels.ToSlice() {
				for _, secret := range forbidden {
					if strings.Contains(label.Value.String(), secret) {
						t.Errorf("%s label %s = %q, want it to carry no %q",
							name, label.Key, label.Value.String(), secret)
					}
				}
			}
		}
	}
}

// A request whose trace is thrown away has nothing to deliver, and a deployed
// instance pays for every attempt.
func TestADeliveryHappensOnlyForASampledRequest(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name    string
		sampler sdktrace.Sampler
		want    int
	}{
		{name: "the trace is kept", sampler: sdktrace.AlwaysSample(), want: 1},
		{name: "the trace is thrown away", sampler: sdktrace.NeverSample(), want: 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			watched := newWatchedRouter(t, c.sampler)
			watched.get(t, "/no-such-endpoint")

			if got := watched.flushes(); got != c.want {
				t.Errorf("the telemetry was asked to deliver %d times, want %d", got, c.want)
			}
		})
	}
}

// A probe is not worth a span: they arrive constantly and say the same thing
// every time.
func TestAProbeIsNotTraced(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.get(t, "/health")

	if spans := watched.spans.Ended(); len(spans) != 0 {
		t.Errorf("the recorder kept %d spans of a probe, want none", len(spans))
	}
}

// The line a person finds in a console names the trace it belongs to, so that
// they can open the request and see what happened inside it.
func TestTheRequestLineNamesItsTrace(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.get(t, "/no-such-endpoint")

	lines := watched.logs.FilterMessage("http_request").All()
	if len(lines) != 1 {
		t.Fatalf("the log holds %d request lines, want 1", len(lines))
	}

	fields := lines[0].ContextMap()
	trace, found := fields["logging.googleapis.com/trace"].(string)
	if !found {
		t.Fatal("the line names no trace, want one")
	}
	if !strings.HasPrefix(trace, "projects/a-project/traces/") {
		t.Errorf("trace = %q, want it under the project", trace)
	}
	if _, found := fields["logging.googleapis.com/spanId"]; !found {
		t.Error("the line names no span, want one")
	}
	if sampled, found := fields["logging.googleapis.com/trace_sampled"].(bool); !found || !sampled {
		t.Error("the line does not say the trace is kept, want it to")
	}
}

// watchedRouter is the real router with everything it records kept in memory.
type watchedRouter struct {
	handler http.Handler
	spans   *tracetest.SpanRecorder
	reader  *sdkmetric.ManualReader
	logs    *observer.ObservedLogs

	delivered atomic.Int64
}

// newWatchedRouter builds it, sampling everything unless a case says otherwise.
func newWatchedRouter(t *testing.T, sampler ...sdktrace.Sampler) *watchedRouter {
	t.Helper()

	kept := sdktrace.AlwaysSample()
	if len(sampler) == 1 {
		kept = sampler[0]
	}

	recorded, logs := observer.New(zapcore.DebugLevel)
	w := &watchedRouter{
		spans:  tracetest.NewSpanRecorder(),
		reader: sdkmetric.NewManualReader(),
		logs:   logs,
	}

	// The real redactor, beside a recorder that keeps what survived it: the
	// guard has to run against what the service actually installs.
	traces := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(kept),
		sdktrace.WithSpanProcessor(telemetry.Redactor()),
		sdktrace.WithSpanProcessor(w.spans),
	)
	t.Cleanup(func() { _ = traces.Shutdown(context.Background()) })

	meters := sdkmetric.NewMeterProvider(sdkmetric.WithReader(w.reader))
	t.Cleanup(func() { _ = meters.Shutdown(context.Background()) })

	router, err := httpserver.NewRouter(httpserver.NewHealthHandler(), zap.New(recorded), httpserver.Observability{
		Traces: traces,
		Meters: meters,
		// Reports what the context it was handed says, so that a case about
		// the context it is handed can tell the difference.
		Flush: func(ctx context.Context) error {
			w.delivered.Add(1)
			return ctx.Err()
		},
		ProjectID: "a-project",
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v, want nil", err)
	}
	w.handler = router
	return w
}

// get asks for one path, the way a caller behind a proxy would.
func (w *watchedRouter) get(t *testing.T, target string) {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody)
	req.Header.Set("X-Forwarded-For", callerAddress)
	req.Header.Set("User-Agent", callerAgent)
	w.handler.ServeHTTP(httptest.NewRecorder(), req)
}

// flushes is how many times the request chain asked for a delivery.
func (w *watchedRouter) flushes() int { return int(w.delivered.Load()) }

// labels are the label sets one metric was recorded under.
func (w *watchedRouter) labels(t *testing.T, name string) []attribute.Set {
	t.Helper()

	var collected metricdata.ResourceMetrics
	if err := w.reader.Collect(t.Context(), &collected); err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}

	var sets []attribute.Set
	for _, scope := range collected.ScopeMetrics {
		for _, recorded := range scope.Metrics {
			if recorded.Name != name {
				continue
			}
			sets = append(sets, pointsOf(t, recorded)...)
		}
	}
	if len(sets) == 0 {
		t.Fatalf("nothing was recorded under %q, want at least one point", name)
	}
	return sets
}

// pointsOf reads the label sets of whichever shape the metric turned out to be.
func pointsOf(t *testing.T, recorded metricdata.Metrics) []attribute.Set {
	t.Helper()

	var sets []attribute.Set
	switch data := recorded.Data.(type) {
	case metricdata.Sum[int64]:
		for _, point := range data.DataPoints {
			sets = append(sets, point.Attributes)
		}
	case metricdata.Histogram[float64]:
		for _, point := range data.DataPoints {
			sets = append(sets, point.Attributes)
		}
	default:
		t.Fatalf("%s is a %T, want a sum or a histogram", recorded.Name, recorded.Data)
	}
	return sets
}

// A caller who hung up must not take the request's spans with them. The
// delivery is detached from the request on purpose, because a batch that fails
// to send is dropped rather than kept for the next attempt.
func TestACallerWhoHungUpStillLeavesItsSpans(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)

	abandoned, hangUp := context.WithCancel(t.Context())
	req := httptest.NewRequestWithContext(abandoned, http.MethodGet, "/no-such-endpoint", http.NoBody)
	hangUp()
	watched.handler.ServeHTTP(httptest.NewRecorder(), req)

	if got := watched.flushes(); got != 1 {
		t.Errorf("the telemetry was asked to deliver %d times, want 1", got)
	}
	if lines := watched.logs.FilterMessage("telemetry_flush_failed").Len(); lines != 0 {
		t.Errorf("the log holds %d delivery failures, want none for a caller that simply left", lines)
	}
}

// Telemetry that is missing a piece is refused where it is assembled, with the
// piece named. Each of the three would otherwise fail in a way of its own, and
// none of them would say what was wrong.
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
