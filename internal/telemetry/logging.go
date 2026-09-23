package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// The three fields a log collector reads to tie a line to a trace. They are
// spelled the way it expects them, which is why they are the only keys in this
// service that are neither lowercase nor snake_case.
const (
	traceField   = "logging.googleapis.com/trace"
	spanField    = "logging.googleapis.com/spanId"
	sampledField = "logging.googleapis.com/trace_sampled"
)

// LogFields are what turns a line in a console into the trace it belongs to.
//
// They carry no fact about the request beyond the identifiers of a span that
// already exists, so they are safe on any line, including one written while
// something is going wrong.
//
// Nothing comes back when there is no span, and nothing comes back without a
// project either: the collector reads the trace as a path under a project, and
// a path under no project points at nothing. A line with no trace to expand
// into is a line, which is what this service had before any of this.
func LogFields(ctx context.Context, projectID string) []zap.Field {
	span := trace.SpanContextFromContext(ctx)
	if !span.IsValid() || projectID == "" {
		return nil
	}
	return []zap.Field{
		zap.String(traceField, "projects/"+projectID+"/traces/"+span.TraceID().String()),
		zap.String(spanField, span.SpanID().String()),
		zap.Bool(sampledField, span.IsSampled()),
	}
}
