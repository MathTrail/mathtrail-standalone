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
// The first three are what off-the-shelf HTTP instrumentation adds by itself:
// who the request came from, by which hop, and what it was sent with. The
// service's own log carries none of that, and a trace is stored somewhere else
// again, under a retention of its own — so an address that was never written
// down in one place must not appear in the other. The last two are here before
// anything sets them: a query string is where a sign-in carries its codes.
//
// The peer's port is deliberately not on the list, though it sits beside an
// address that is. On its own an ephemeral source port names nobody, and it is
// a number rather than a string: blanking it would change what the attribute
// is in order to hide what it never said.
var redactedKeys = map[attribute.Key]struct{}{
	"client.address":       {},
	"network.peer.address": {},
	"user_agent.original":  {},
	"url.query":            {},
	"url.full":             {},
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
	// here or anywhere else. What that leaves open is whatever is set between
	// a start and an end — which today is this service's own code, chosen
	// rather than filtered. The guard test reads spans that have already
	// ended, so a library that starts adding one later is caught there rather
	// than shipped.
}

// Shutdown releases nothing, because this holds nothing.
func (redactor) Shutdown(context.Context) error { return nil }

// ForceFlush delivers nothing, because this keeps nothing.
func (redactor) ForceFlush(context.Context) error { return nil }
