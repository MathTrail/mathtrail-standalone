package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// redactedValue is what stands in for a value this service may not keep. The
// key survives, so that a reader can see the shape of what was recorded and
// that something was deliberately left out of it.
const redactedValue = "redacted"

// redactedKeys are the attributes whose values never leave this process.
//
// The first five are what off-the-shelf HTTP instrumentation adds by itself:
// who the request came from, by which hop, and what it was sent with, and the
// path and the verb exactly as the caller wrote them. The service's own log
// carries none of that — it names a request by the route it matched — and a
// trace is stored somewhere else again, under a retention of its own, so what
// was never written down in one place must not appear in the other. The route
// and a method from the known ones stay on the span, as they are in the log.
// The last two are here before anything sets them: a query string is where a
// sign-in carries its codes.
//
// The peer's port is deliberately not on the list, though it sits beside an
// address that is. On its own an ephemeral source port names nobody, and it is
// a number rather than a string: blanking it would change what the attribute
// is in order to hide what it never said.
var redactedKeys = map[attribute.Key]struct{}{
	"client.address":               {},
	"network.peer.address":         {},
	"user_agent.original":          {},
	"url.path":                     {},
	"http.request.method_original": {},
	"url.query":                    {},
	"url.full":                     {},
}

// redactor blanks those attributes on every span, as it starts.
type redactor struct{}

// Redactor keeps out of the traces what the logs are not allowed to carry.
//
// It works by overwriting rather than by asking instrumentation not to record:
// what a library attaches is the library's to change, and a list of forbidden
// keys that is enforced after the fact holds whatever version of it is
// installed.
func Redactor() sdktrace.SpanProcessor { return redactor{} }

// OnStart overwrites the forbidden attributes a span was started with.
//
// The interpreter of the specification sets a span's start attributes before
// any processor is told the span exists, which is what makes this the earliest
// point where they can be read, and the only one before they could be read by
// anybody else. Nothing is added: a key that was not there stays away, so a
// span that never had an address does not gain a field saying it lost one.
func (redactor) OnStart(_ context.Context, span sdktrace.ReadWriteSpan) {
	var blanked []attribute.KeyValue
	for _, recorded := range span.Attributes() {
		if _, forbidden := redactedKeys[recorded.Key]; forbidden {
			blanked = append(blanked, recorded.Key.String(redactedValue))
		}
	}
	if len(blanked) > 0 {
		span.SetAttributes(blanked...)
	}
}

// OnEnd is where redaction cannot happen, and the gap that leaves is worth
// naming rather than hiding.
func (redactor) OnEnd(sdktrace.ReadOnlySpan) {
	// Deliberately nothing. A processor is handed a finished span for reading
	// only, so an attribute added after the span started cannot be blanked
	// here. What that leaves open is whatever is set between a start and an
	// end — which today is this service's own code, chosen rather than
	// filtered, and the text of a span's status, which the exporter is handed
	// without (withoutStatusText). The guard test reads spans that have
	// already ended, so a library that starts adding an attribute later is
	// caught there rather than shipped.
}

// Shutdown releases nothing, because this holds nothing.
func (redactor) Shutdown(context.Context) error { return nil }

// ForceFlush delivers nothing, because this keeps nothing.
func (redactor) ForceFlush(context.Context) error { return nil }

// withoutStatusText hands the exporter every span with the text of its status
// taken off and its code kept.
//
// The text is written as a span ends, after any processor could change it,
// and it is not ours to choose: the web framework's instrumentation fills it
// with the errors the request collected, and a failed write among them names
// both ends of the connection — the address of whoever was at the other end.
// The exporter is the last place anything can be done about it, so it is done
// there. The code still says whether the request failed, which is what a trace
// is read for; what failed is in the request's own line in the log.
func withoutStatusText(exporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return statusless{exporter}
}

// statusless is an exporter that sees spans through bareStatus.
type statusless struct{ sdktrace.SpanExporter }

// ExportSpans passes the spans on, each without the text of its status.
func (e statusless) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	bare := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, span := range spans {
		bare[i] = bareStatus{span}
	}
	return e.SpanExporter.ExportSpans(ctx, bare)
}

// bareStatus is a span whose status says whether it failed and not why.
type bareStatus struct{ sdktrace.ReadOnlySpan }

// Status is the span's status without its text.
func (s bareStatus) Status() sdktrace.Status {
	return sdktrace.Status{Code: s.ReadOnlySpan.Status().Code}
}
