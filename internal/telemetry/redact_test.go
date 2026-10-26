package telemetry_test

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

// ownTracer is the name a tracer of the service's own goes by: anything under
// the module's path.
const ownTracer = "github.com/MathTrail/mathtrail-standalone/internal/telemetry/redact_test"

// libraryTracer is the name of a tracer the service does not own.
const libraryTracer = "go.opentelemetry.io/contrib/a-library"

// A library records whatever its version records, so of the attributes it
// starts a span with only those on the list stay: the route, a known method,
// the scheme and the protocol. Anything else keeps its key and loses its
// value before any processor after the redactor reads it — the ports and the
// host a caller wrote, and an attribute no version had when the list was
// written.
func TestALibrarysSpanStartsWithOnlyWhatIsAllowed(t *testing.T) {
	t.Parallel()

	p := newPipeline(t)
	_, span := p.traces.Tracer(libraryTracer).Start(t.Context(), "GET /kept", trace.WithAttributes(
		attribute.String("http.route", "/kept"),
		attribute.String("http.request.method", "GET"),
		attribute.String("url.scheme", "https"),
		attribute.String("network.protocol.name", "http"),
		attribute.String("network.protocol.version", "1.1"),
		attribute.String("server.address", "mcp.example"),
		attribute.Int("server.port", 6543),
		attribute.Int("network.peer.port", 51234),
		attribute.String("a.later.version.adds.this", "masha-ivanova"),
	))
	span.End()

	inside, _ := p.only(t)
	wantAttributes(t, "the span inside the process", inside.Attributes(), map[attribute.Key]string{
		"http.route":                "/kept",
		"http.request.method":       "GET",
		"url.scheme":                "https",
		"network.protocol.name":     "http",
		"network.protocol.version":  "1.1",
		"server.address":            "redacted",
		"server.port":               "redacted",
		"network.peer.port":         "redacted",
		"a.later.version.adds.this": "redacted",
	})
}

// What a library gives its span after the start, as an attribute, an event or
// a link, is held to a list on its way out: the status code, the size of the
// answer and the type of an error stay, and everything else leaves blanked —
// an error's text recorded as an event among it.
func TestALibrarysSpanLeavesWithOnlyWhatIsAllowed(t *testing.T) {
	t.Parallel()

	p := newPipeline(t)
	_, span := p.traces.Tracer(libraryTracer).Start(t.Context(), "GET /kept",
		trace.WithAttributes(attribute.String("http.route", "/kept")),
		trace.WithLinks(trace.Link{
			SpanContext: linkedSpan(),
			Attributes:  []attribute.KeyValue{attribute.String("a.link.says", "masha-ivanova")},
		}),
	)
	span.SetAttributes(
		attribute.Int("http.response.status_code", 500),
		attribute.Int("http.response.body.size", 12),
		attribute.String("error.type", "*net.OpError"),
		attribute.String("url.path", "/children/masha-ivanova"),
		attribute.String("a.later.version.adds.this", "masha-ivanova"),
	)
	span.AddEvent("exception", trace.WithAttributes(
		attribute.String("exception.message", "write tcp 10.0.0.2:8080->203.0.113.9:51234: broken pipe"),
	))
	span.End()

	_, left := p.only(t)
	wantAttributes(t, "the span that left", left.Attributes, map[attribute.Key]string{
		"http.route":                "/kept",
		"http.response.status_code": "500",
		"http.response.body.size":   "12",
		"error.type":                "*net.OpError",
		"url.path":                  "redacted",
		"a.later.version.adds.this": "redacted",
	})
	if len(left.Events) != 1 || left.Events[0].Name != "exception" {
		t.Fatalf("events = %+v, want the one exception, by its name", left.Events)
	}
	wantAttributes(t, "the event that left", left.Events[0].Attributes, map[attribute.Key]string{
		"exception.message": "redacted",
	})
	if len(left.Links) != 1 || left.Links[0].SpanContext.SpanID() != linkedSpan().SpanID() {
		t.Fatalf("links = %+v, want the one link, to the span it points at", left.Links)
	}
	wantAttributes(t, "the link that left", left.Links[0].Attributes, map[attribute.Key]string{
		"a.link.says": "redacted",
	})
}

// The service's own span keeps what its code chose — at its start, after it,
// on its events and on its links — and loses only what the logs may not carry,
// whoever set it and whenever.
func TestTheServicesOwnSpanKeepsAllButTheForbidden(t *testing.T) {
	t.Parallel()

	p := newPipeline(t)
	_, span := p.traces.Tracer(ownTracer).Start(t.Context(), "drive get",
		trace.WithAttributes(
			attribute.String("mathtrail.drive.op", "get"),
			attribute.String("client.address", "203.0.113.9"),
			attribute.String("user_agent.original", "a-browser-nobody-wrote"),
		),
		trace.WithLinks(trace.Link{
			SpanContext: linkedSpan(),
			Attributes: []attribute.KeyValue{
				attribute.String("mathtrail.link.why", "the call before"),
				attribute.String("client.address", "203.0.113.9"),
			},
		}),
	)
	span.SetAttributes(
		attribute.String("mathtrail.drive.outcome", "ok"),
		attribute.String("url.full", "https://mcp.example/oauth/callback?code=a-code"),
	)
	span.AddEvent("retried", trace.WithAttributes(
		attribute.Int("mathtrail.drive.retries", 1),
		attribute.String("url.full", "https://mcp.example/oauth/callback?code=a-code"),
	))
	span.End()

	inside, left := p.only(t)
	wantAttributes(t, "the span inside the process", inside.Attributes(), map[attribute.Key]string{
		"mathtrail.drive.op":      "get",
		"client.address":          "redacted",
		"user_agent.original":     "redacted",
		"mathtrail.drive.outcome": "ok",
		// Set after the start, where no processor can change it: it is
		// blanked on its way out.
		"url.full": "https://mcp.example/oauth/callback?code=a-code",
	})
	wantAttributes(t, "the span that left", left.Attributes, map[attribute.Key]string{
		"mathtrail.drive.op":      "get",
		"client.address":          "redacted",
		"user_agent.original":     "redacted",
		"mathtrail.drive.outcome": "ok",
		"url.full":                "redacted",
	})
	if len(left.Events) != 1 {
		t.Fatalf("events = %+v, want the one the service recorded", left.Events)
	}
	wantAttributes(t, "the event that left", left.Events[0].Attributes, map[attribute.Key]string{
		"mathtrail.drive.retries": "1",
		"url.full":                "redacted",
	})
	if len(left.Links) != 1 {
		t.Fatalf("links = %+v, want the one the service made", left.Links)
	}
	wantAttributes(t, "the link that left", left.Links[0].Attributes, map[attribute.Key]string{
		"mathtrail.link.why": "the call before",
		"client.address":     "redacted",
	})
}

// Redaction blanks what is there and adds nothing: a span that never carried
// an address must not gain a field saying it lost one.
func TestRedactionAddsNothingThatWasNotThere(t *testing.T) {
	t.Parallel()

	for name, tracer := range map[string]string{"the service's own": ownTracer, "a library's": libraryTracer} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := newPipeline(t)
			_, span := p.traces.Tracer(tracer).Start(t.Context(), "unit")
			span.End()

			inside, left := p.only(t)
			if attrs := inside.Attributes(); len(attrs) != 0 {
				t.Errorf("the span inside the process carries %v, want nothing added to a span that carried nothing", attrs)
			}
			if len(left.Attributes) != 0 {
				t.Errorf("the span that left carries %v, want nothing added to a span that carried nothing", left.Attributes)
			}
		})
	}
}

// pipeline is a tracer provider put together the way the service puts its own
// together — the redactor, and an exporter behind Stripped — with a recorder
// beside them of the spans as they ended inside the process.
type pipeline struct {
	traces   *sdktrace.TracerProvider
	recorded *tracetest.SpanRecorder
	exported *tracetest.InMemoryExporter
}

// newPipeline builds one that keeps every span.
func newPipeline(t *testing.T) *pipeline {
	t.Helper()

	p := &pipeline{recorded: tracetest.NewSpanRecorder(), exported: tracetest.NewInMemoryExporter()}
	p.traces = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(telemetry.Redactor()),
		sdktrace.WithSpanProcessor(p.recorded),
		sdktrace.WithSyncer(telemetry.Stripped(p.exported)),
	)
	t.Cleanup(func() { _ = p.traces.Shutdown(context.Background()) })
	return p
}

// only is the one span the case made, as it ended inside the process and as
// it left it.
func (p *pipeline) only(t *testing.T) (sdktrace.ReadOnlySpan, tracetest.SpanStub) {
	t.Helper()

	inside, left := p.recorded.Ended(), p.exported.GetSpans()
	if len(inside) != 1 || len(left) != 1 {
		t.Fatalf("the recorder kept %d spans and the exporter was handed %d, want 1 each", len(inside), len(left))
	}
	return inside[0], left[0]
}

// wantAttributes holds attributes to exactly the keys and values given, each
// value as its String writes it.
func wantAttributes(t *testing.T, of string, got []attribute.KeyValue, want map[attribute.Key]string) {
	t.Helper()

	seen := make(map[attribute.Key]string, len(got))
	for _, attr := range got {
		seen[attr.Key] = attr.Value.String()
	}
	for key, value := range want {
		if have, ok := seen[key]; !ok {
			t.Errorf("%s has no %s, want %q", of, key, value)
		} else if have != value {
			t.Errorf("%s: %s = %q, want %q", of, key, have, value)
		}
	}
	for key, have := range seen {
		if _, expected := want[key]; !expected {
			t.Errorf("%s: %s = %q, want no such attribute", of, key, have)
		}
	}
}

// linkedSpan is a span somewhere else that a link can point at.
func linkedSpan() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01},
		SpanID:     trace.SpanID{0x02},
		TraceFlags: trace.FlagsSampled,
	})
}
