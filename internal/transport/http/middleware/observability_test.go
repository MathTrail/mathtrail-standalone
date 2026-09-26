package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
	"github.com/MathTrail/mathtrail-standalone/internal/transport/http/middleware"
)

// What a caller reveals about itself just by asking. None of it may end up in
// a trace: the service's own log carries none of it, and a span is kept
// somewhere else again, under a retention nobody told a parent about.
const (
	callerAddress  = "203.0.113.9"
	peerAddress    = "192.0.2.1"
	callerAgent    = "a-browser-nobody-wrote"
	callerPassword = "hunter2"
	callerName     = "masha-ivanova"
	callerVerb     = "MASHA"
)

// These build the real router, because the order these handlers run in is
// what half of them are for, and that order is the router's to decide.

// The guard. Nothing a request carried about who sent it, and nothing a child
// or a task could be recognised by, may appear in a span's name or attributes —
// the path and the verb as the caller wrote them included.
func TestASpanCarriesNothingThatIdentifiesTheCaller(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.do(t, callerVerb, "/children/"+callerName+"?token="+callerPassword)

	spans := watched.spans.Ended()
	if len(spans) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(spans))
	}

	forbidden := []string{callerAddress, peerAddress, callerAgent, callerPassword, callerName, callerVerb}
	for _, secret := range forbidden {
		if strings.Contains(spans[0].Name(), secret) {
			t.Errorf("span name = %q, want it to carry no %q", spans[0].Name(), secret)
		}
	}
	for _, attr := range spans[0].Attributes() {
		value := attr.Value.String()
		for _, secret := range forbidden {
			if strings.Contains(value, secret) {
				t.Errorf("span attribute %s = %q, want it to carry no %q", attr.Key, value, secret)
			}
		}
	}
}

// What an off-the-shelf instrumentation adds by itself about the caller keeps
// its key and loses its value, so that a reader can see that something was
// deliberately left out rather than never recorded.
func TestWhatTheCallerSentIsBlankedRatherThanDropped(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.do(t, callerVerb, "/no-such-endpoint")

	spans := watched.spans.Ended()
	if len(spans) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(spans))
	}

	blanked := map[string]bool{
		"client.address":               false,
		"network.peer.address":         false,
		"user_agent.original":          false,
		"url.path":                     false,
		"http.request.method_original": false,
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

// A panic is an answer like any other to what counts, logs and delivers: the
// request is logged once with the 500 it got, counted under that status, and
// its trace is delivered with its span marked as failed. A panic that unwound
// past them on its way to being caught would leave the request in none of the
// three.
func TestAPanicIsLoggedCountedAndDelivered(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	engine, isEngine := watched.handler.(*gin.Engine)
	if !isEngine {
		t.Fatalf("the router is a %T, want a *gin.Engine", watched.handler)
	}
	engine.GET("/boom", func(*gin.Context) { panic("boom") })

	watched.get(t, "/boom")

	lines := watched.logs.FilterMessage("http_request").All()
	if len(lines) != 1 || lines[0].ContextMap()["status"] != int64(http.StatusInternalServerError) {
		t.Errorf("http_request lines = %v, want one, of status 500", lines)
	}
	counted := false
	for _, labels := range watched.labels(t, "http_request") {
		if status, found := labels.Value("http.response.status_code"); found && status.AsInt64() == http.StatusInternalServerError {
			counted = true
		}
	}
	if !counted {
		t.Error("the request was not counted under status 500")
	}
	if watched.flushes() != 1 {
		t.Errorf("deliveries = %d, want 1", watched.flushes())
	}
	if spans := watched.spans.Ended(); len(spans) != 1 || spans[0].Status().Code != codes.Error {
		t.Errorf("ended spans = %v, want one marked as failed", spans)
	}
}

// A request a handler asked to drop is passed on to the server as the abort it
// is, and what it recorded is still delivered on the way: the spans of a
// request that went wrong are the ones worth reading. A delivery that panics
// on the way costs its line and leaves the abort as it was.
func TestAnAbortedRequestStillDeliversItsSpans(t *testing.T) {
	t.Parallel()

	for name, burning := range map[string]bool{
		"a delivery that works":  false,
		"a delivery that panics": true,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			watched := newWatchedRouter(t)
			watched.panicDelivery.Store(burning)
			engine, isEngine := watched.handler.(*gin.Engine)
			if !isEngine {
				t.Fatalf("the router is a %T, want a *gin.Engine", watched.handler)
			}
			engine.GET("/abort", func(*gin.Context) { panic(http.ErrAbortHandler) })

			if passed := servedPanic(t, engine, "/abort"); passed != http.ErrAbortHandler { //nolint:errorlint // the server compares it by identity
				t.Errorf("the server got %v, want %v", passed, http.ErrAbortHandler)
			}
			if watched.flushes() != 1 {
				t.Errorf("deliveries = %d, want 1", watched.flushes())
			}
			if lines := watched.logs.FilterMessage("panic").All(); len(lines) != 0 {
				t.Errorf("panic lines = %v, want none: a request dropped on purpose is no fault", lines)
			}
		})
	}
}

// A delivery that panics costs a line and not the request: the answer stands,
// and the panic is written by its kind alone.
func TestADeliveryThatPanicsCostsALine(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.panicDelivery.Store(true)
	watched.get(t, "/no-such-endpoint")

	lines := watched.logs.FilterMessage("telemetry_flush_panicked").All()
	if len(lines) != 1 || lines[0].ContextMap()["panic"] != "a value of type string" {
		t.Errorf("delivery lines = %v, want one naming the kind of the panic alone", lines)
	}
	if requests := watched.logs.FilterMessage("http_request").All(); len(requests) != 1 ||
		requests[0].ContextMap()["status"] != int64(http.StatusNotFound) {
		t.Errorf("request lines = %v, want the answer the request got, a 404", requests)
	}
}

// burningTracers hand out a tracer that panics as a span is opened: a panic in
// the middleware itself, above the recovery next to the handlers.
type burningTracers struct{ tracenoop.TracerProvider }

func (burningTracers) Tracer(string, ...trace.TracerOption) trace.Tracer { return burningTracer{} }

type burningTracer struct{ tracenoop.Tracer }

func (burningTracer) Start(context.Context, string, ...trace.SpanStartOption) (context.Context, trace.Span) {
	panic("the tracer is on fire")
}

// A panic in the middleware itself is caught by the last resort the router puts
// above everything: written once, by its kind alone, and answered, instead of
// reaching the server with its value.
func TestAPanicInTheMiddlewareIsCaughtByTheLastResort(t *testing.T) {
	t.Parallel()

	recorded, logs := observer.New(zapcore.DebugLevel)
	router, err := httpserver.NewRouter(httpserver.NewHealthHandler(), zap.New(recorded), httpserver.Observability{
		Traces: burningTracers{},
		Meters: metricnoop.NewMeterProvider(),
		Flush:  func(context.Context, bool) error { return nil },
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v, want nil", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/no-such-endpoint", http.NoBody))

	lines := logs.FilterMessage("panic").All()
	if len(lines) != 1 || lines[0].ContextMap()["panic"] != "a value of type string" {
		t.Errorf("panic lines = %v, want one naming the kind of the panic alone", lines)
	}
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

// A meter that cannot make one of the two instruments stops the counting
// before any request is served, and the refusal names the instrument: a
// service that counted nothing would say so to nobody.
func TestAnInstrumentThatCannotBeMadeIsNamed(t *testing.T) {
	t.Parallel()

	for _, refused := range []string{"http_request", "http_request_duration"} {
		t.Run(refused, func(t *testing.T) {
			t.Parallel()

			_, err := middleware.Metrics(refusingMeters{refused: refused})
			if !errors.Is(err, errMeterRefused) || !strings.Contains(err.Error(), refused+": ") {
				t.Errorf("Metrics() error = %v, want the refusal of %s passed on under its name", err, refused)
			}
		})
	}
}

// errMeterRefused is what a refusing meter answers.
var errMeterRefused = errors.New("the meter refuses")

// refusingMeters hand out a meter that refuses to make the one instrument a
// case names, and makes every other as a meter that records nothing does.
type refusingMeters struct {
	metricnoop.MeterProvider
	refused string
}

func (m refusingMeters) Meter(string, ...metric.MeterOption) metric.Meter {
	return refusingMeter{refused: m.refused}
}

type refusingMeter struct {
	metricnoop.Meter
	refused string
}

func (m refusingMeter) Int64Counter(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	if name == m.refused {
		return nil, errMeterRefused
	}
	return m.Meter.Int64Counter(name, options...)
}

func (m refusingMeter) Float64Histogram(
	name string, options ...metric.Float64HistogramOption,
) (metric.Float64Histogram, error) {
	if name == m.refused {
		return nil, errMeterRefused
	}
	return m.Meter.Float64Histogram(name, options...)
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

// A request whose trace is thrown away has no spans to deliver, and waiting on
// another's would cost it time for nothing; but it still offers what is due,
// so that measurements do not wait for a request whose trace is kept.
func TestSpansAreDeliveredOnlyForASampledRequest(t *testing.T) {
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
				t.Errorf("the telemetry was asked for spans %d times, want %d", got, c.want)
			}
			// Whatever became of the trace, the request offers the rest: the
			// measurements leave with whichever request finds them due.
			if got := watched.offers(); got != 1 {
				t.Errorf("the request offered %d deliveries, want 1", got)
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
	named, found := fields["logging.googleapis.com/trace"].(string)
	if !found {
		t.Fatal("the line names no trace, want one")
	}
	if !strings.HasPrefix(named, "projects/a-project/traces/") {
		t.Errorf("trace = %q, want it under the project", named)
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

	delivered      atomic.Int64
	offered        atomic.Int64
	refuseDelivery atomic.Bool
	panicDelivery  atomic.Bool
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
		Flush: func(ctx context.Context, spans bool) error {
			w.offered.Add(1)
			if spans {
				w.delivered.Add(1)
			}
			if w.panicDelivery.Load() {
				panic("the collector is on fire")
			}
			if w.refuseDelivery.Load() {
				return errors.New("the collector said no")
			}
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
	w.do(t, http.MethodGet, target)
}

// do asks for one path with any method at all, the way a caller behind a proxy
// would.
func (w *watchedRouter) do(t *testing.T, method, target string) {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, target, http.NoBody)
	req.Header.Set("X-Forwarded-For", callerAddress)
	req.Header.Set("User-Agent", callerAgent)
	w.handler.ServeHTTP(httptest.NewRecorder(), req)
}

// flushes is how many times the request chain asked for spans to be delivered.
func (w *watchedRouter) flushes() int { return int(w.delivered.Load()) }

// offers is how many times it offered a delivery at all, spans or not.
func (w *watchedRouter) offers() int { return int(w.offered.Load()) }

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

// A collector that refused the delivery costs a line in the log and nothing
// else: the response has been written by now, and nobody is waiting on this.
func TestARefusedDeliveryIsLoggedAndNothingMore(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.refuseDelivery.Store(true)
	watched.get(t, "/no-such-endpoint")

	if lines := watched.logs.FilterMessage("telemetry_flush_failed").Len(); lines != 1 {
		t.Errorf("the log holds %d delivery failures, want 1", lines)
	}
	if watched.logs.FilterMessage("http_request").Len() != 1 {
		t.Error("the request was not answered and logged as usual")
	}
}

// A verb nobody declared is counted as one kind, like an undeclared route:
// otherwise a stranger invents a series by inventing a word.
func TestAnUnknownMethodIsCountedAsOther(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)

	req := httptest.NewRequestWithContext(t.Context(), "BREW", "/no-such-endpoint", http.NoBody)
	watched.handler.ServeHTTP(httptest.NewRecorder(), req)

	for _, labels := range watched.labels(t, "http_request") {
		method, found := labels.Value("http.request.method")
		if !found {
			t.Fatal("http_request carries no method, want one")
		}
		if method.String() != "other" {
			t.Errorf("http.request.method = %q, want %q", method.String(), "other")
		}
	}
}

// A probe is left out of the counters as it is left out of the traces and the
// log. The platform asks for one constantly, and counted, they would make
// "requests answered" a measure of how often it checked rather than of how
// much the service was used.
func TestAProbeIsNotCounted(t *testing.T) {
	t.Parallel()

	watched := newWatchedRouter(t)
	watched.get(t, "/health")

	var collected metricdata.ResourceMetrics
	if err := watched.reader.Collect(t.Context(), &collected); err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}
	for _, scope := range collected.ScopeMetrics {
		for _, recorded := range scope.Metrics {
			t.Errorf("a probe was recorded under %q, want nothing counted", recorded.Name)
		}
	}
}
