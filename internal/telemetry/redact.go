package telemetry

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// redactedValue is what stands in for a value this service may not keep. The
// key survives, so that a reader can see the shape of what was recorded and
// that something was deliberately left out of it.
const redactedValue = "redacted"

// ownScope begins the name of every tracer of this service's own. A span
// started under any other name is a library's.
const ownScope = "github.com/MathTrail/mathtrail-standalone/"

// keptAtStart are the only attributes a library's span keeps of those it
// starts with: the route and a method from the known ones, which name a
// request as the log names it, and the scheme and the protocol it came over.
//
// A library records whatever its version records, and the HTTP
// instrumentation records a good deal about the caller: the address and the
// port of the connection, the host and the port the caller wrote in its Host
// header, the browser, the path and the verb as the caller wrote them. The
// service's own log carries none of that, and a trace is stored somewhere else
// again, under a retention of its own, so what was never written down in one
// place must not appear in the other. A list of what may stay holds whatever
// the next version of a library adds; a list of what may not would let it
// through.
var keptAtStart = keys(
	"http.route",
	"http.request.method",
	"url.scheme",
	"network.protocol.name",
	"network.protocol.version",
)

// keptAtEnd are the only attributes a library's span keeps of those it adds as
// it ends: the status code, the size of the answer and the type of an error,
// which say how the request went and nothing about who sent it.
var keptAtEnd = keys(
	"http.response.status_code",
	"http.response.body.size",
	"error.type",
)

// forbiddenOwn are the attributes a span of this service's own never keeps.
// Its own code chooses what it records, so a list of what may not stay is
// enough there; these are what the HTTP instrumentation records about a
// caller, kept off the service's spans too in case a span is ever started
// with a request's attributes. A query string is where a sign-in carries its
// codes.
var forbiddenOwn = keys(
	"client.address",
	"network.peer.address",
	"user_agent.original",
	"url.path",
	"http.request.method_original",
	"url.query",
	"url.full",
)

// keys is a set of attribute keys.
func keys(names ...attribute.Key) map[attribute.Key]struct{} {
	set := make(map[attribute.Key]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}

// ours reports whether a span was started by a tracer of this service's own.
func ours(span sdktrace.ReadOnlySpan) bool {
	return strings.HasPrefix(span.InstrumentationScope().Name, ownScope)
}

// rule says which attributes keep their values.
type rule func(attribute.Key) bool

// notForbidden is the rule of the service's own spans: every attribute but a
// forbidden one.
func notForbidden(key attribute.Key) bool {
	_, forbidden := forbiddenOwn[key]
	return !forbidden
}

// onLists is the rule of a library's span: only the attributes on one of the
// lists.
func onLists(lists ...map[attribute.Key]struct{}) rule {
	return func(key attribute.Key) bool {
		for _, list := range lists {
			if _, kept := list[key]; kept {
				return true
			}
		}
		return false
	}
}

// nothing is the rule that keeps no attribute.
func nothing(attribute.Key) bool { return false }

// The rules a library's span is held to: as it starts, and as it leaves.
var (
	libraryAtStart = onLists(keptAtStart)
	libraryLeaving = onLists(keptAtStart, keptAtEnd)
)

// redactor blanks, as each span starts, the attributes it may not keep.
type redactor struct{}

// Redactor keeps out of the traces what the logs are not allowed to carry.
//
// It works by overwriting rather than by asking instrumentation not to record:
// what a library attaches is the library's to change, and a rule enforced
// after the fact holds whatever version of it is installed.
func Redactor() sdktrace.SpanProcessor { return redactor{} }

// OnStart overwrites the attributes a span was started with that it may not
// keep.
//
// The interpreter of the specification sets a span's start attributes before
// any processor is told the span exists, which is what makes this the earliest
// point where they can be read, and the only one before they could be read by
// anybody else. Nothing is added: a key that was not there stays away, so a
// span that never had an address does not gain a field saying it lost one.
func (redactor) OnStart(_ context.Context, span sdktrace.ReadWriteSpan) {
	keep := rule(notForbidden)
	if !ours(span) {
		keep = libraryAtStart
	}
	var hidden []attribute.KeyValue
	for _, recorded := range span.Attributes() {
		if !keep(recorded.Key) {
			hidden = append(hidden, recorded.Key.String(redactedValue))
		}
	}
	if len(hidden) > 0 {
		span.SetAttributes(hidden...)
	}
}

// OnEnd is where redaction cannot happen: a processor is handed a finished
// span for reading only. What a span is given between its start and its end
// is held to the same rule on its way out instead (Stripped).
func (redactor) OnEnd(sdktrace.ReadOnlySpan) {
	// Deliberately nothing: a finished span can be read here and no longer
	// changed.
}

// Shutdown releases nothing, because this holds nothing.
func (redactor) Shutdown(context.Context) error { return nil }

// ForceFlush delivers nothing, because this keeps nothing.
func (redactor) ForceFlush(context.Context) error { return nil }

// Stripped hands the exporter every span without what may not leave the
// process once the span has ended.
//
// That is, first, the text of its status. It is written as a span ends, after
// any processor could change it, and it is not ours to choose: the web
// framework's instrumentation fills it with the errors the request collected,
// and a failed write among them names both ends of the connection — the
// address of whoever was at the other end. The code still says whether the
// request failed, which is what a trace is read for; what failed is in the
// request's own line in the log.
//
// It is, next, what a span was given after it started, which no processor
// could change either. A span of the service's own leaves with every
// attribute but a forbidden one, on itself, its events and its links. A
// library's leaves with only the attributes on either of its lists, and with
// no attribute on its events and links: a library chooses those, and an error
// it records as an event carries its text. A name of the first list that a
// library sets again after the start leaves with the value it was set to,
// since what a span ends with does not say when it was set.
func Stripped(exporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return stripped{exporter}
}

// stripped is an exporter that sees spans through strippedSpan.
type stripped struct{ sdktrace.SpanExporter }

// ExportSpans passes the spans on, each stripped.
func (e stripped) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	bare := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, span := range spans {
		bare[i] = strip(span)
	}
	return e.SpanExporter.ExportSpans(ctx, bare)
}

// strippedSpan is a span as it may leave the process: its attributes held to
// one rule, and those of its events and links to another.
type strippedSpan struct {
	sdktrace.ReadOnlySpan
	keep, keepOnParts rule
}

// strip reads which rules hold for the span.
func strip(span sdktrace.ReadOnlySpan) strippedSpan {
	if ours(span) {
		return strippedSpan{ReadOnlySpan: span, keep: notForbidden, keepOnParts: notForbidden}
	}
	return strippedSpan{ReadOnlySpan: span, keep: libraryLeaving, keepOnParts: nothing}
}

// Status is the span's status without its text.
func (s strippedSpan) Status() sdktrace.Status {
	return sdktrace.Status{Code: s.ReadOnlySpan.Status().Code}
}

// Attributes are the span's attributes, each one it may not keep blanked.
func (s strippedSpan) Attributes() []attribute.KeyValue {
	return blanked(s.ReadOnlySpan.Attributes(), s.keep)
}

// Events are the span's events, with their names and times, and each
// attribute they may not keep blanked.
func (s strippedSpan) Events() []sdktrace.Event {
	events := s.ReadOnlySpan.Events()
	bare := make([]sdktrace.Event, len(events))
	for i, event := range events {
		event.Attributes = blanked(event.Attributes, s.keepOnParts)
		bare[i] = event
	}
	return bare
}

// Links are the span's links, to the spans they point at, and each attribute
// they may not keep blanked.
func (s strippedSpan) Links() []sdktrace.Link {
	links := s.ReadOnlySpan.Links()
	bare := make([]sdktrace.Link, len(links))
	for i, link := range links {
		link.Attributes = blanked(link.Attributes, s.keepOnParts)
		bare[i] = link
	}
	return bare
}

// blanked are the attributes with the value of each one the rule does not keep
// replaced, in the order they came. Attributes it keeps whole come back as
// they are; otherwise the copy is a fresh slice, so the span's own stays as it
// was.
func blanked(attrs []attribute.KeyValue, keep rule) []attribute.KeyValue {
	var out []attribute.KeyValue
	for i, attr := range attrs {
		if keep(attr.Key) {
			if out != nil {
				out[i] = attr
			}
			continue
		}
		if out == nil {
			out = make([]attribute.KeyValue, len(attrs))
			copy(out, attrs[:i])
		}
		out[i] = attr.Key.String(redactedValue)
	}
	if out == nil {
		return attrs
	}
	return out
}
