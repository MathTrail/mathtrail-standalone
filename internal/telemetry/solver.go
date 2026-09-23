package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// solverScope is what the spans and the measurements of a run are recorded
// under, so that a reader can tell them from the ones a library produced.
const solverScope = "github.com/MathTrail/mathtrail-standalone/internal/domain/solver"

// stepBuckets are where the interesting differences are. A brute force either
// finishes in the first thousand instructions or walks a search space, and the
// question the numbers have to answer is how close the walks come to the
// ceiling — so the buckets thin out at the bottom and crowd towards it.
var stepBuckets = []float64{
	1_000, 10_000, 100_000, 500_000, 1_000_000, 2_500_000, 5_000_000, 10_000_000, 25_000_000,
}

// observedSolver reports what each run cost, and passes the run through
// untouched.
type observedSolver struct {
	next   solver.Runner
	tracer trace.Tracer
	steps  metric.Int64Histogram
}

// ObserveSolver wraps a runner so that every run leaves a span and a
// measurement of what it spent.
//
// It is a wrapper rather than a change to the sandbox because what is worth
// recording — the status, the cost, the time — is the whole of a run's result,
// and a result is what any runner returns. The interpreter stays free of all
// of this, and so do its tests.
func ObserveSolver(next solver.Runner, tracers trace.TracerProvider, meters metric.MeterProvider) (solver.Runner, error) {
	steps, err := meters.Meter(solverScope).Int64Histogram(
		"solver_run_steps",
		metric.WithDescription("what one run of a solver spent of its step budget"),
		metric.WithUnit("{step}"),
		metric.WithExplicitBucketBoundaries(stepBuckets...),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: solver_run_steps: %w", err)
	}
	return &observedSolver{
		next:   next,
		tracer: tracers.Tracer(solverScope),
		steps:  steps,
	}, nil
}

// Run records the run and reports exactly what the runner underneath did.
//
// A program that failed is not an error of this service: it is a fact about
// what the chat's model wrote, and the whole point of running it is to find
// that out. So the span stays successful and carries the status as an
// attribute; only a run that never happened marks the span as failed.
//
// What the program arrived at never reaches the span. The letters are the
// answer, and the answer stays hidden until a child has given theirs.
func (o *observedSolver) Run(ctx context.Context, source string, options solver.Options) (solver.Result, error) {
	ctx, span := o.tracer.Start(ctx, "solver_run")
	defer span.End()

	result, err := o.next.Run(ctx, source, options)
	if err != nil {
		// Both, because they say different things: the event carries the cause,
		// and the status is what makes the span show up as failed at all.
		// Recording the error alone leaves a run that never happened looking
		// like one that went fine.
		span.RecordError(err)
		span.SetStatus(codes.Error, "the run did not happen")
		return result, err
	}

	status := attribute.String("solver.status", string(result.Status))
	span.SetAttributes(
		status,
		attribute.Int64("solver.steps", spent(result.Steps)),
		attribute.Int64("solver.duration_ms", result.Duration.Milliseconds()),
	)
	o.steps.Record(ctx, spent(result.Steps), metric.WithAttributes(status))
	return result, nil
}

// spent is the step count as a number a measurement can carry. The budget the
// sandbox is given is far below the ceiling here, so the clamp never fires in
// a running service; it is what keeps a configuration nobody would write from
// turning a large number into a negative one.
func spent(steps uint64) int64 {
	const ceiling = uint64(1)<<63 - 1
	if steps > ceiling {
		return int64(ceiling)
	}
	return int64(steps)
}
