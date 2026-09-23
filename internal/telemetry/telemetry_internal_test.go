package telemetry

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
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

// What the SDK could not deliver goes to the service's own log, at the one
// level that says a person should look.
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
	if lines[0].Message != "telemetry_export_failed" {
		t.Errorf("message = %q, want %q", lines[0].Message, "telemetry_export_failed")
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

// Both signals' addresses come out of one root, and a root that is not an
// address is named rather than posted to. Left to itself the exporter keeps
// its own default and says nothing, so a mistyped address becomes spans that
// arrive nowhere.
func TestBothAddressesComeOutOfOneRoot(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name    string
		root    string
		traces  string
		refused bool
	}{
		{name: "a bare host", root: "https://telemetry.example", traces: "https://telemetry.example/v1/traces"},
		{name: "a trailing slash", root: "https://telemetry.example/", traces: "https://telemetry.example/v1/traces"},
		{name: "a port", root: "http://127.0.0.1:4318", traces: "http://127.0.0.1:4318/v1/traces"},
		{name: "no scheme", root: "telemetry.example", refused: true},
		{name: "a host and a port, no scheme", root: "localhost:4318", refused: true},
		{name: "a scheme we do not speak", root: "grpc://telemetry.example", refused: true},
		{name: "no host", root: "https:///v1", refused: true},
		{name: "empty", root: "", refused: true},
		{name: "not an address at all", root: "http://[::1", refused: true},
	} {
		traces, metrics, err := endpoints(c.root)
		switch {
		case c.refused:
			if err == nil {
				t.Errorf("endpoints(%s) error = nil, want the root named", c.name)
			}
		case err != nil:
			t.Errorf("endpoints(%s) error = %v, want nil", c.name, err)
		default:
			if traces != c.traces {
				t.Errorf("endpoints(%s) traces = %q, want %q", c.name, traces, c.traces)
			}
			if want := strings.Replace(c.traces, tracePath, metricPath, 1); metrics != want {
				t.Errorf("endpoints(%s) metrics = %q, want %q", c.name, metrics, want)
			}
		}
	}
}
