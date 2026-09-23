package telemetry

import (
	"context"
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

// A path is put under the collector's root, and a root that is not an address
// is refused by name rather than posted to.
func TestOneSignalsPathGoesUnderTheRoot(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name    string
		root    string
		want    string
		refused bool
	}{
		{name: "a bare host", root: "https://telemetry.example", want: "https://telemetry.example/v1/traces"},
		{name: "a trailing slash", root: "https://telemetry.example/", want: "https://telemetry.example/v1/traces"},
		{name: "a port", root: "http://127.0.0.1:4318", want: "http://127.0.0.1:4318/v1/traces"},
		{name: "not an address", root: "http://[::1", refused: true},
	} {
		got, err := signalURL(c.root, tracePath)
		switch {
		case c.refused && err == nil:
			t.Errorf("signalURL(%s) error = nil, want the root named", c.name)
		case !c.refused && err != nil:
			t.Errorf("signalURL(%s) error = %v, want nil", c.name, err)
		case !c.refused && got != c.want:
			t.Errorf("signalURL(%s) = %q, want %q", c.name, got, c.want)
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

// An address the exporter could not post to is named, not worked around. Left
// to itself the exporter keeps its own default and says nothing, so a mistyped
// address becomes spans that arrive nowhere.
func TestAnAddressThatIsNotOneIsRefused(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name string
		root string
	}{
		{name: "no scheme", root: "telemetry.example"},
		{name: "a host and a port, no scheme", root: "localhost:4318"},
		{name: "a scheme we do not speak", root: "grpc://telemetry.example"},
		{name: "no host", root: "https:///v1"},
		{name: "empty", root: ""},
	} {
		if _, err := signalURL(c.root, tracePath); err == nil {
			t.Errorf("signalURL(%s) error = nil, want the address named", c.name)
		}
	}
}
