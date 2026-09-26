package telemetry

import (
	"context"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// A delivery of measurements is booked once an interval, and the first one is
// always due: a process that has just started has told nobody anything.
func TestMeasurementsComeDueOnceAnInterval(t *testing.T) {
	t.Parallel()

	tel := &Telemetry{}
	start := time.Now()

	// Steps, not cases: each answer depends on the ones before it, because
	// booking a delivery is what moves the interval on. Running these in any
	// other order, or apart from each other, would test something else.
	steps := []struct {
		name string
		at   time.Time
		want bool
	}{
		{"the first one", start, true},
		{"straight after", start, false},
		{"a moment short of the interval", start.Add(MetricInterval - time.Nanosecond), false},
		{"the interval, exactly", start.Add(MetricInterval), true},
	}
	for _, step := range steps {
		if got := tel.metricsDue(step.at); got != step.want {
			t.Errorf("metricsDue(%s) = %v, want %v", step.name, got, step.want)
		}
	}
}

// What the SDK reports going wrong goes to the service's own log, at the one
// level that says a person should look, under a name that does not claim every
// report is a failed delivery.
func TestAFailedDeliveryIsLoggedAsAnError(t *testing.T) {
	t.Parallel()

	recorded, logs := observer.New(zapcore.DebugLevel)
	ErrorHandler(zap.New(recorded)).Handle(context.DeadlineExceeded)

	lines := logs.All()
	if len(lines) != 1 {
		t.Fatalf("the handler wrote %d lines, want 1", len(lines))
	}
	if lines[0].Level != zapcore.ErrorLevel {
		t.Errorf("level = %v, want %v", lines[0].Level, zapcore.ErrorLevel)
	}
	if lines[0].Message != "telemetry_failed" {
		t.Errorf("message = %q, want %q", lines[0].Message, "telemetry_failed")
	}
}

// Nothing may ask the platform to describe itself when there is nowhere to
// send the description: off a deployment that is a call to a service that
// does not exist, on every start and in every test.
func TestThePlatformIsNotAskedWhenNothingIsExported(t *testing.T) {
	t.Parallel()

	asked := &countingDetector{}
	newResource(t.Context(), false, &Settings{Detector: asked}, zap.NewNop())

	if asked.calls != 0 {
		t.Errorf("the platform was asked %d times, want none", asked.calls)
	}
}

// Every span says which service, which build and which project it belongs to.
// The project is the one the collector reads to decide where the data goes.
func TestTheResourceNamesTheServiceTheBuildAndTheProject(t *testing.T) {
	t.Parallel()

	res := newResource(t.Context(), true, &Settings{
		Version:   "1.2.3",
		ProjectID: "a-project",
		Detector:  &countingDetector{},
	}, zap.NewNop())

	for key, want := range map[attribute.Key]string{
		"service.name":    ServiceName,
		"service.version": "1.2.3",
		projectIDKey:      "a-project",
	} {
		if got := valueOf(res, key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

// A delivery carries what changed, except for the two kinds whose change says
// nothing on its own.
func TestOnlyWhatCanFallKeepsItsTotal(t *testing.T) {
	t.Parallel()

	for kind, want := range map[sdkmetric.InstrumentKind]metricdata.Temporality{
		sdkmetric.InstrumentKindCounter:                 metricdata.DeltaTemporality,
		sdkmetric.InstrumentKindHistogram:               metricdata.DeltaTemporality,
		sdkmetric.InstrumentKindObservableCounter:       metricdata.DeltaTemporality,
		sdkmetric.InstrumentKindUpDownCounter:           metricdata.CumulativeTemporality,
		sdkmetric.InstrumentKindObservableUpDownCounter: metricdata.CumulativeTemporality,
	} {
		if got := deltaTemporality(kind); got != want {
			t.Errorf("deltaTemporality(%v) = %v, want %v", kind, got, want)
		}
	}
}

// valueOf reads one attribute of a resource, empty when it carries none.
func valueOf(res *resource.Resource, key attribute.Key) string {
	for _, attr := range res.Attributes() {
		if attr.Key == key {
			return attr.Value.String()
		}
	}
	return ""
}

// countingDetector is a platform that says nothing and remembers being asked.
type countingDetector struct {
	calls int
}

func (d *countingDetector) Detect(context.Context) (*resource.Resource, error) {
	d.calls++
	return nil, nil
}

// A step count is a measurement, and a measurement is a signed number. The
// budget a sandbox is given is far below the ceiling here, so the clamp never
// fires in a running service; it is what keeps a configuration nobody would
// write from turning a large number into a negative one.
func TestAStepCountNeverComesBackNegative(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name  string
		steps uint64
		want  int64
	}{
		{name: "none", steps: 0, want: 0},
		{name: "an ordinary run", steps: 4321, want: 4321},
		{name: "past what a measurement can hold", steps: math.MaxUint64, want: math.MaxInt64},
	} {
		if got := spent(c.steps); got != c.want {
			t.Errorf("spent(%s) = %d, want %d", c.name, got, c.want)
		}
	}
}

// Asked for no platform in particular, the package reaches for the real one.
// Building it costs nothing — it is asking it that costs a call.
func TestThePlatformDefaultsToTheRealOne(t *testing.T) {
	t.Parallel()

	if platform(&Settings{}) == nil {
		t.Error("platform() = nil, want the real detector")
	}
}

// Both signals' addresses come out of one root, each under whatever path the
// root carries.
func TestBothAddressesComeOutOfOneRoot(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name   string
		root   string
		traces string
	}{
		{name: "a bare host", root: "https://telemetry.example", traces: "https://telemetry.example/v1/traces"},
		{name: "a trailing slash", root: "https://telemetry.example/", traces: "https://telemetry.example/v1/traces"},
		{name: "a port", root: "http://127.0.0.1:4318", traces: "http://127.0.0.1:4318/v1/traces"},
		{name: "a prefix", root: "https://telemetry.example/otlp", traces: "https://telemetry.example/otlp/v1/traces"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			root, err := url.Parse(c.root)
			if err != nil {
				t.Fatalf("url.Parse(%q) error = %v, want nil", c.root, err)
			}
			traces, metrics := endpoints(root)
			wantMetrics := strings.Replace(c.traces, tracePath, metricPath, 1)
			if traces != c.traces || metrics != wantMetrics {
				t.Errorf("endpoints(%q) = %q, %q, want %q, %q", c.root, traces, metrics, c.traces, wantMetrics)
			}
		})
	}
}

// A start that is going to fail on its address does not first ask the
// platform to describe itself: that is a call that can take seconds, and it
// would be spent on a service that is not going to run.
func TestARefusedAddressAsksThePlatformNothing(t *testing.T) {
	t.Parallel()

	asked := &countingDetector{}
	_, err := New(t.Context(), &Settings{
		Enabled:    true,
		Endpoint:   "telemetry.example",
		ProjectID:  "a-project",
		HTTPClient: http.DefaultClient,
		Detector:   asked,
	}, zap.NewNop())
	if err == nil {
		t.Fatal("New() error = nil, want the address refused")
	}
	if asked.calls != 0 {
		t.Errorf("the platform was asked %d times, want none", asked.calls)
	}
}

// Spans and measurements are delivered side by side, each with the whole
// deadline. Each exporter below waits to see the other one start, so queued
// one behind the other the first waits out the deadline and the second finds
// it spent.
func TestSpansAndMeasurementsAreDeliveredSideBySide(t *testing.T) {
	t.Parallel()

	spansStarted, measurementsStarted := make(chan struct{}), make(chan struct{})
	traces := sdktrace.NewTracerProvider(sdktrace.WithBatcher(&waitingSpans{
		waiting: waiting{started: spansStarted, other: measurementsStarted},
	}))
	metrics := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(&waitingMeasurements{
		waiting: waiting{started: measurementsStarted, other: spansStarted},
	})))
	t.Cleanup(func() {
		_ = traces.Shutdown(context.Background())
		_ = metrics.Shutdown(context.Background())
	})
	tel := &Telemetry{traces: traces, metrics: metrics, enabled: true}

	_, span := traces.Tracer("test").Start(t.Context(), "request")
	span.End()
	requests, err := metrics.Meter("test").Int64Counter("requests")
	if err != nil {
		t.Fatalf("Int64Counter() error = %v, want nil", err)
	}
	requests.Add(t.Context(), 1)

	if err := tel.ForceFlush(t.Context(), true); err != nil {
		t.Errorf("ForceFlush() error = %v, want nil: one delivery waited for the other", err)
	}
}

// waiting is an exporter's half of a meeting: it says it has started, and
// waits until the other half has too, or the deadline passes.
type waiting struct {
	once    sync.Once
	started chan struct{}
	other   chan struct{}
}

func (w *waiting) meet(ctx context.Context) error {
	w.once.Do(func() { close(w.started) })
	select {
	case <-w.other:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// waitingSpans is a span exporter that delivers by meeting the other one.
type waitingSpans struct{ waiting }

func (e *waitingSpans) ExportSpans(ctx context.Context, _ []sdktrace.ReadOnlySpan) error {
	return e.meet(ctx)
}

func (*waitingSpans) Shutdown(context.Context) error { return nil }

// waitingMeasurements is a metric exporter that delivers by meeting the other
// one.
type waitingMeasurements struct{ waiting }

func (*waitingMeasurements) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	return deltaTemporality(kind)
}

func (*waitingMeasurements) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.DefaultAggregationSelector(kind)
}

func (e *waitingMeasurements) Export(ctx context.Context, _ *metricdata.ResourceMetrics) error {
	return e.meet(ctx)
}

func (*waitingMeasurements) ForceFlush(context.Context) error { return nil }

func (*waitingMeasurements) Shutdown(context.Context) error { return nil }

// What the exporter is handed says whether a span failed and not why: the
// text of a status is whatever errors the request collected, and a failed
// write names the address at the other end of it.
func TestTheExporterNeverSeesTheTextOfAStatus(t *testing.T) {
	t.Parallel()

	exported := tracetest.NewInMemoryExporter()
	traces := sdktrace.NewTracerProvider(sdktrace.WithSyncer(withoutStatusText(exported)))
	t.Cleanup(func() { _ = traces.Shutdown(context.Background()) })

	_, span := traces.Tracer("test").Start(t.Context(), "request")
	span.SetStatus(codes.Error, "write tcp 10.0.0.2:8080->203.0.113.9:51234: write: broken pipe")
	span.End()

	spans := exported.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("the exporter was handed %d spans, want 1", len(spans))
	}
	if status := spans[0].Status; status.Code != codes.Error || status.Description != "" {
		t.Errorf("status = %+v, want the failure without its text", status)
	}
}
