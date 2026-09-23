package telemetry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap/zaptest"
	"golang.org/x/oauth2"

	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

const testToken = "a-token-nobody-issued"

// A span that finished must reach the collector when a delivery is asked for,
// at the path the protocol gives it, signed with the credentials it was built
// with and naming the project it belongs to.
func TestSpansReachTheCollector(t *testing.T) {
	collector := newCollector(t, http.StatusOK)
	tel := newTelemetry(t, collector, &telemetry.Settings{SampleRatio: 1})

	_, span := tel.TracerProvider().Tracer("test").Start(t.Context(), "unit")
	span.End()

	if err := tel.ForceFlush(t.Context()); err != nil {
		t.Fatalf("ForceFlush() error = %v, want nil", err)
	}

	got := collector.takenAt("/v1/traces")
	if len(got) != 1 {
		t.Fatalf("collector saw %d span deliveries, want 1", len(got))
	}
	if auth := got[0].header.Get("Authorization"); auth != "Bearer "+testToken {
		t.Errorf("Authorization = %q, want the token the client was built with", auth)
	}
	if project := got[0].header.Get("X-Goog-User-Project"); project != "a-project" {
		t.Errorf("X-Goog-User-Project = %q, want %q", project, "a-project")
	}
	if got[0].size == 0 {
		t.Error("the collector was sent an empty body, want the encoded span")
	}
}

// Measurements go to their own path, and only once an interval has passed:
// every delivery costs bytes whether or not anything changed.
func TestMeasurementsAreNotDeliveredTwiceInAnInterval(t *testing.T) {
	collector := newCollector(t, http.StatusOK)
	tel := newTelemetry(t, collector, &telemetry.Settings{SampleRatio: 0})

	counter, err := tel.MeterProvider().Meter("test").Int64Counter("unit_total")
	if err != nil {
		t.Fatalf("Int64Counter() error = %v, want nil", err)
	}

	for range 3 {
		counter.Add(t.Context(), 1)
		if err := tel.ForceFlush(t.Context()); err != nil {
			t.Fatalf("ForceFlush() error = %v, want nil", err)
		}
	}

	if deliveries := len(collector.takenAt("/v1/metrics")); deliveries != 1 {
		t.Errorf("the collector saw %d metric deliveries, want 1 within one interval", deliveries)
	}
}

// A collector that refuses everything must cost the caller its deadline and
// nothing else: the process keeps running and keeps recording.
func TestARefusingCollectorDoesNotStopTheService(t *testing.T) {
	collector := newCollector(t, http.StatusServiceUnavailable)
	tel := newTelemetry(t, collector, &telemetry.Settings{SampleRatio: 1})

	_, span := tel.TracerProvider().Tracer("test").Start(t.Context(), "unit")
	span.End()

	ctx, cancel := context.WithTimeout(t.Context(), telemetry.FlushTimeout)
	defer cancel()
	if err := tel.ForceFlush(ctx); err == nil {
		t.Error("ForceFlush() error = nil, want the refusal reported")
	}

	// Still usable afterwards: the failure belonged to one delivery, not to
	// the provider.
	_, after := tel.TracerProvider().Tracer("test").Start(t.Context(), "after")
	after.End()

	// Without the detachment the shutdown would be handed a context the test
	// framework has already cancelled, and would deliver nothing; the second
	// is what keeps a shutdown that hangs from hanging the test.
	closing, cancelClosing := context.WithTimeout(context.WithoutCancel(t.Context()), time.Second)
	defer cancelClosing()
	_ = tel.Shutdown(closing)
}

// Switched off, the providers still exist and still hand out tracers, because
// instrumented code must not know which it is. Nothing leaves the process.
func TestSwitchedOffNothingLeavesTheProcess(t *testing.T) {
	collector := newCollector(t, http.StatusOK)

	tel, err := telemetry.New(t.Context(), &telemetry.Settings{
		Enabled:     false,
		Endpoint:    collector.url,
		SampleRatio: 1,
		HTTPClient:  signedClient(t),
		ProjectID:   "a-project",
		Detector:    stubDetector{},
	}, zaptest.NewLogger(t))
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	_, span := tel.TracerProvider().Tracer("test").Start(t.Context(), "unit")
	span.End()
	if err := tel.ForceFlush(t.Context()); err != nil {
		t.Fatalf("ForceFlush() error = %v, want nil", err)
	}
	if err := tel.Shutdown(t.Context()); err != nil {
		t.Fatalf("Shutdown() error = %v, want nil", err)
	}

	if seen := collector.taken(); len(seen) != 0 {
		t.Errorf("the collector saw %d requests, want none", len(seen))
	}
}

// Credentials that cannot be found are a deployment's problem, not a reason to
// refuse to start: the service comes up and exports nothing.
func TestMissingCredentialsLeaveTheServiceRunning(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", t.TempDir()+"/nothing-here.json")

	collector := newCollector(t, http.StatusOK)
	tel, err := telemetry.New(t.Context(), &telemetry.Settings{
		Enabled:     true,
		Endpoint:    collector.url,
		SampleRatio: 1,
		ProjectID:   "a-project",
		Detector:    stubDetector{},
	}, zaptest.NewLogger(t))
	if err != nil {
		t.Fatalf("New() error = %v, want a running service", err)
	}

	_, span := tel.TracerProvider().Tracer("test").Start(t.Context(), "unit")
	span.End()
	if err := tel.ForceFlush(t.Context()); err != nil {
		t.Fatalf("ForceFlush() error = %v, want nil", err)
	}
	if seen := collector.taken(); len(seen) != 0 {
		t.Errorf("the collector saw %d requests, want none without credentials", len(seen))
	}
}

// A platform that will not describe itself costs the spans an attribute, not
// the service its start.
func TestAFailingDetectorDoesNotStopTheStart(t *testing.T) {
	collector := newCollector(t, http.StatusOK)
	tel := newTelemetry(t, collector, &telemetry.Settings{
		SampleRatio: 1,
		Detector:    stubDetector{err: context.DeadlineExceeded},
	})

	_, span := tel.TracerProvider().Tracer("test").Start(t.Context(), "unit")
	span.End()
	if err := tel.ForceFlush(t.Context()); err != nil {
		t.Fatalf("ForceFlush() error = %v, want nil", err)
	}
	if len(collector.takenAt("/v1/traces")) != 1 {
		t.Error("the span was not delivered, want the start to survive a silent platform")
	}
}

// newTelemetry builds telemetry pointed at a collector, with a client that
// signs like a real one and a platform that says nothing unless the case says
// otherwise.
func newTelemetry(t *testing.T, c *collector, settings *telemetry.Settings) *telemetry.Telemetry {
	t.Helper()

	settings.Enabled = true
	settings.Endpoint = c.url
	settings.ProjectID = "a-project"
	settings.HTTPClient = signedClient(t)
	if settings.Detector == nil {
		settings.Detector = stubDetector{}
	}

	tel, err := telemetry.New(t.Context(), settings, zaptest.NewLogger(t))
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	t.Cleanup(func() {
		closing, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tel.Shutdown(closing)
	})
	return tel
}

// request is what the collector saw, without its body: the bytes are protobuf
// and what matters here is that they arrived.
type request struct {
	path   string
	header http.Header
	size   int64
}

// collector stands in for the one a deployment posts to.
type collector struct {
	url    string
	status int

	mu   sync.Mutex
	seen []request
}

func newCollector(t *testing.T, status int) *collector {
	t.Helper()

	c := &collector{status: status}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size, _ := io.Copy(io.Discard, r.Body)

		c.mu.Lock()
		c.seen = append(c.seen, request{path: r.URL.Path, header: r.Header.Clone(), size: size})
		c.mu.Unlock()

		w.WriteHeader(c.status)
	}))
	t.Cleanup(server.Close)

	c.url = server.URL
	return c
}

func (c *collector) taken() []request {
	c.mu.Lock()
	defer c.mu.Unlock()

	return append([]request(nil), c.seen...)
}

// takenAt is what the collector saw at one path. A delivery of spans and a
// delivery of measurements are two requests, and every case here cares about
// one of them.
func (c *collector) takenAt(path string) []request {
	var at []request
	for _, seen := range c.taken() {
		if seen.path == path {
			at = append(at, seen)
		}
	}
	return at
}

// stubDetector is a platform that answers whatever a case needs it to.
type stubDetector struct {
	err error
}

func (d stubDetector) Detect(context.Context) (*resource.Resource, error) {
	if d.err != nil {
		return nil, d.err
	}
	return resource.NewWithAttributes(semconv.SchemaURL, semconv.CloudRegion("nowhere")), nil
}

// A deployment whose project has not reached it yet is a deployment that
// exports nothing, not one that refuses to answer children. The name arrives
// with the configuration, and an image can be rolled before it.
func TestWithoutAProjectNothingLeavesTheProcess(t *testing.T) {
	collector := newCollector(t, http.StatusOK)

	tel, err := telemetry.New(t.Context(), &telemetry.Settings{
		Enabled:     true,
		Endpoint:    collector.url,
		SampleRatio: 1,
		HTTPClient:  signedClient(t),
		Detector:    stubDetector{},
	}, zaptest.NewLogger(t))
	if err != nil {
		t.Fatalf("New() error = %v, want a running service", err)
	}

	_, span := tel.TracerProvider().Tracer("test").Start(t.Context(), "unit")
	span.End()
	if err := tel.ForceFlush(t.Context()); err != nil {
		t.Fatalf("ForceFlush() error = %v, want nil", err)
	}
	if seen := collector.taken(); len(seen) != 0 {
		t.Errorf("the collector saw %d requests, want none without a project", len(seen))
	}
}

// signedClient is what a real export would be signed with, so that a case
// about something else — a switch, a missing project — fails when that
// something else stops being what holds the data in.
func signedClient(t *testing.T) *http.Client {
	t.Helper()

	return oauth2.NewClient(t.Context(),
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: testToken, TokenType: "Bearer"}))
}

// A collector that never answers must not become the thing a child waits for.
// The delivery carries its own deadline, so a caller that passed none still
// gets its goroutine back long before the exporter's own five seconds.
func TestADeliveryBoundsItselfWhenTheCallerDidNot(t *testing.T) {
	tel := newTelemetry(t, newSlowCollector(t), &telemetry.Settings{SampleRatio: 1})

	_, span := tel.TracerProvider().Tracer("test").Start(t.Context(), "unit")
	span.End()

	started := time.Now()
	//nolint:usetesting // the point of the case is a caller that brought no deadline
	err := tel.ForceFlush(context.Background())
	took := time.Since(started)

	if err == nil {
		t.Error("ForceFlush() error = nil, want the deadline reported")
	}
	if took > time.Second {
		t.Errorf("ForceFlush() took %v, want it bounded by %v", took, telemetry.FlushTimeout)
	}
}

// A caller who has already run out of time is told so rather than made to wait
// again: the deadline the delivery adds is a ceiling, never a floor.
func TestADeliveryObeysACallerWithNoTimeLeft(t *testing.T) {
	tel := newTelemetry(t, newSlowCollector(t), &telemetry.Settings{SampleRatio: 1})

	_, span := tel.TracerProvider().Tracer("test").Start(t.Context(), "unit")
	span.End()

	spent, cancel := context.WithCancel(t.Context())
	cancel()

	started := time.Now()
	if err := tel.ForceFlush(spent); err == nil {
		t.Error("ForceFlush() error = nil, want the cancellation reported")
	}
	if took := time.Since(started); took > telemetry.FlushTimeout {
		t.Errorf("ForceFlush() took %v, want it to return at once", took)
	}
}

// newSlowCollector answers later than a delivery is allowed to wait. That is
// the failure a deadline exists for: not a refusal, which comes back at once,
// but a collector that holds the connection open.
//
// It does answer in the end, and it gives up the moment the caller has: a
// collector that never answered would leave the exporter retrying into a
// closing test for as long as its own retry budget allows.
func newSlowCollector(t *testing.T) *collector {
	t.Helper()

	c := &collector{status: http.StatusOK}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(4 * telemetry.FlushTimeout):
		case <-r.Context().Done():
		}
		w.WriteHeader(c.status)
	}))
	t.Cleanup(server.Close)

	c.url = server.URL
	return c
}
