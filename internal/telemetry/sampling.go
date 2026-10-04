package telemetry

import (
	"encoding/binary"
	"fmt"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// sampler keeps a trace at the configured share, whatever the request it
// arrived in says about it. That decision is not the platform's alone: a
// client sets it in the header too, and one asking for every request to be
// traced would be choosing what this service spends on traces and how long
// its callers wait for deliveries. A span of our own follows its parent, so a
// trace is kept or dropped whole.
func sampler(ratio float64) sdktrace.Sampler {
	share := newTraceShare(ratio)
	return sdktrace.ParentBased(share,
		sdktrace.WithRemoteParentSampled(share),
		sdktrace.WithRemoteParentNotSampled(share),
	)
}

// traceShare keeps a trace when its identifier, stirred, falls below bound,
// which is to 2⁶³ what the share is to all traces. The decision is the
// identifier's, so a trace is decided alike on every instance, however many
// requests carry it.
//
// It reads the whole identifier, where the standard ratio sampler reads the
// right half alone. The platform's front end makes the identifier of a request
// that arrives without one, and it fills the right half once for a run of
// requests and only the left half anew for each: read by the right half, a run
// was kept or dropped whole, and an hour of requests kept a fourteenth of the
// share configured.
type traceShare struct {
	bound       uint64
	description string
}

// newTraceShare keeps the share of traces given, from none to all. A share that
// is not above nought keeps none.
func newTraceShare(ratio float64) traceShare {
	var bound uint64
	switch {
	case ratio >= 1:
		bound = 1 << 63
	case ratio > 0:
		bound = uint64(ratio * (1 << 63))
	}
	return traceShare{bound: bound, description: fmt.Sprintf("TraceShare{%g}", ratio)}
}

// ShouldSample keeps the trace or drops it, and hands the parent's trace state
// on either way.
//
//nolint:gocritic // hugeParam: the SDK's sampler takes its parameters by value
func (s traceShare) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	result := sdktrace.SamplingResult{
		Decision:   sdktrace.Drop,
		Tracestate: trace.SpanContextFromContext(p.ParentContext).TraceState(),
	}
	if stirred(p.TraceID)>>1 < s.bound {
		result.Decision = sdktrace.RecordAndSample
	}
	return result
}

// Description names the sampler and its share.
func (s traceShare) Description() string { return s.description }

// stirred folds an identifier into one number that every bit of it reaches, so
// that whichever part of the identifier changes from one trace to the next,
// traces spread evenly over the numbers. The right half is stirred before it
// is folded into the left, so that two halves alike do not cancel out.
func stirred(id trace.TraceID) uint64 {
	return finalized(binary.BigEndian.Uint64(id[:8]) ^ finalized(binary.BigEndian.Uint64(id[8:])))
}

// finalized is the last step of MurmurHash3's 64-bit hash: a change to any bit
// of x changes about half the bits of the result, and no two numbers come out
// as one.
func finalized(x uint64) uint64 {
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	x ^= x >> 33
	return x
}
