package telemetry_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap/zaptest"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
)

// Every run of a solver leaves a span saying what became of it and what it
// cost — and nothing about what it arrived at. The letters are the answer, and
// the answer stays hidden until a child has given theirs.
func TestARunLeavesItsCostAndNotItsAnswer(t *testing.T) {
	t.Parallel()

	watched, spans, reader := observedRunner(t, &fakeRunner{result: solver.Result{
		Status:   solver.StatusOK,
		Letters:  []string{"C"},
		Steps:    4321,
		Duration: 7 * time.Millisecond,
	}})

	if _, err := watched.Run(t.Context(), "def solve(options): pass", solver.Options{}); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	ended := spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(ended))
	}
	if ended[0].Name() != "solver_run" {
		t.Errorf("span name = %q, want %q", ended[0].Name(), "solver_run")
	}

	recorded := map[attribute.Key]string{}
	for _, attr := range ended[0].Attributes() {
		recorded[attr.Key] = attr.Value.String()
		if strings.Contains(attr.Value.String(), "C") && attr.Key != "solver.status" {
			t.Errorf("span attribute %s = %q, want the answer nowhere near it", attr.Key, attr.Value.String())
		}
	}
	for key, want := range map[attribute.Key]string{
		"solver.status":      "ok",
		"solver.steps":       "4321",
		"solver.duration_ms": "7",
	} {
		if recorded[key] != want {
			t.Errorf("%s = %q, want %q", key, recorded[key], want)
		}
	}

	if steps := histogramSum(t, reader, "solver_run_steps"); steps != 4321 {
		t.Errorf("solver_run_steps recorded %d, want 4321", steps)
	}
}

// A program that failed is a fact about what the chat's model wrote, not a
// failure of this service: the span carries the status and stays successful.
func TestAFailedProgramIsNotAFailedSpan(t *testing.T) {
	t.Parallel()

	watched, spans, _ := observedRunner(t, &fakeRunner{result: solver.Result{
		Status: solver.StatusTimeout,
		Steps:  10_000_000,
	}})

	if _, err := watched.Run(t.Context(), "while True: pass", solver.Options{}); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	ended := spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(ended))
	}
	if ended[0].Status().Code == codes.Error {
		t.Error("the span is marked failed, want a solver's own failure to be an attribute")
	}
}

// A run that never happened is this service's problem, and the span says so.
func TestARunThatNeverHappenedIsRecordedAsAnError(t *testing.T) {
	t.Parallel()

	watched, spans, _ := observedRunner(t, &fakeRunner{err: errors.New("no slot came free")})

	if _, err := watched.Run(t.Context(), "", solver.Options{}); err == nil {
		t.Fatal("Run() error = nil, want the refusal passed through")
	}

	ended := spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(ended))
	}
	if len(ended[0].Events()) == 0 {
		t.Error("the span records nothing about the refusal, want it named")
	}
}

// The fields that tie a line to a trace are there when there is a trace to tie
// it to, and absent otherwise: a line with nothing to point at should point at
// nothing rather than at an empty project.
func TestALineIsTiedToItsTraceOnlyWhenThereIsOne(t *testing.T) {
	t.Parallel()

	traces := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	t.Cleanup(func() { _ = traces.Shutdown(context.Background()) })

	traced, span := traces.Tracer("test").Start(t.Context(), "unit")
	defer span.End()

	for _, c := range []struct {
		name      string
		ctx       context.Context
		projectID string
		want      int
	}{
		{name: "a span and a project", ctx: traced, projectID: "a-project", want: 3},
		{name: "a span and no project", ctx: traced, want: 0},
		{name: "a project and no span", ctx: t.Context(), projectID: "a-project", want: 0},
	} {
		if got := telemetry.LogFields(c.ctx, c.projectID); len(got) != c.want {
			t.Errorf("LogFields(%s) gave %d fields, want %d", c.name, len(got), c.want)
		}
	}
}

// Redaction blanks what is there and adds nothing: a span that never carried
// an address must not gain a field saying it lost one.
func TestRedactionAddsNothingThatWasNotThere(t *testing.T) {
	t.Parallel()

	spans := tracetest.NewSpanRecorder()
	traces := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(telemetry.Redactor()),
		sdktrace.WithSpanProcessor(spans),
	)
	t.Cleanup(func() { _ = traces.Shutdown(context.Background()) })

	_, span := traces.Tracer("test").Start(t.Context(), "unit")
	span.End()

	ended := spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(ended))
	}
	if attrs := ended[0].Attributes(); len(attrs) != 0 {
		t.Errorf("the span carries %v, want nothing added to a span that carried nothing", attrs)
	}
}

// observedRunner wraps a runner in telemetry that keeps what it records.
func observedRunner(t *testing.T, next solver.Runner) (solver.Runner, *tracetest.SpanRecorder, *sdkmetric.ManualReader) {
	t.Helper()

	spans := tracetest.NewSpanRecorder()
	traces := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(spans),
	)
	t.Cleanup(func() { _ = traces.Shutdown(context.Background()) })

	reader := sdkmetric.NewManualReader()
	watched, err := telemetry.ObserveSolver(next, traces, sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))
	if err != nil {
		t.Fatalf("ObserveSolver() error = %v, want nil", err)
	}
	return watched, spans, reader
}

// histogramSum is everything one histogram was given, added up.
func histogramSum(t *testing.T, reader *sdkmetric.ManualReader, name string) int64 {
	t.Helper()

	var collected metricdata.ResourceMetrics
	if err := reader.Collect(t.Context(), &collected); err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}

	for _, scope := range collected.ScopeMetrics {
		for _, recorded := range scope.Metrics {
			if recorded.Name != name {
				continue
			}
			histogram, ok := recorded.Data.(metricdata.Histogram[int64])
			if !ok {
				t.Fatalf("%s is a %T, want a histogram of counts", name, recorded.Data)
			}
			var total int64
			for _, point := range histogram.DataPoints {
				total += point.Sum
			}
			return total
		}
	}
	t.Fatalf("nothing was recorded under %q", name)
	return 0
}

// fakeRunner is a sandbox that returns what a case needs it to, so that the
// case is about what was recorded rather than about an interpreter.
type fakeRunner struct {
	result solver.Result
	err    error
}

func (f *fakeRunner) Run(context.Context, string, solver.Options) (solver.Result, error) {
	return f.result, f.err
}

// A run that never happened is this service's problem, and the span has to
// look failed rather than merely carry a note about it.
func TestARefusedRunMarksItsSpanFailed(t *testing.T) {
	t.Parallel()

	watched, spans, _ := observedRunner(t, &fakeRunner{err: errors.New("no slot came free")})

	if _, err := watched.Run(t.Context(), "", solver.Options{}); err == nil {
		t.Fatal("Run() error = nil, want the refusal passed through")
	}

	ended := spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(ended))
	}
	if code := ended[0].Status().Code; code != codes.Error {
		t.Errorf("span status = %v, want %v", code, codes.Error)
	}
}

// Redaction is what keeps out of a trace what the logs are not allowed to
// carry, and it has to work on whatever put the attribute there rather than on
// one library's habits.
func TestAForbiddenAttributeIsBlankedWhoeverSetIt(t *testing.T) {
	t.Parallel()

	spans := tracetest.NewSpanRecorder()
	traces := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(telemetry.Redactor()),
		sdktrace.WithSpanProcessor(spans),
	)
	t.Cleanup(func() { _ = traces.Shutdown(context.Background()) })

	_, span := traces.Tracer("test").Start(t.Context(), "unit", trace.WithAttributes(
		attribute.String("client.address", "203.0.113.9"),
		attribute.String("user_agent.original", "a-browser-nobody-wrote"),
		attribute.String("http.route", "/kept"),
	))
	span.End()

	ended := spans.Ended()
	if len(ended) != 1 {
		t.Fatalf("the recorder kept %d spans, want 1", len(ended))
	}

	for _, attr := range ended[0].Attributes() {
		switch attr.Key {
		case "client.address", "user_agent.original":
			if attr.Value.String() != "redacted" {
				t.Errorf("%s = %q, want it blanked", attr.Key, attr.Value.String())
			}
		case "http.route":
			if attr.Value.String() != "/kept" {
				t.Errorf("%s = %q, want it left alone", attr.Key, attr.Value.String())
			}
		}
	}
}

// An address the exporter could not post to stops the telemetry from being
// built at all, rather than being worked around quietly.
func TestAnEndpointThatIsNotAnAddressStopsTheBuild(t *testing.T) {
	t.Parallel()

	_, err := telemetry.New(t.Context(), &telemetry.Settings{
		Enabled:    true,
		Endpoint:   "telemetry.example",
		ProjectID:  "a-project",
		HTTPClient: http.DefaultClient,
		Detector:   stubDetector{},
	}, zaptest.NewLogger(t))

	if err == nil {
		t.Fatal("New() error = nil, want the address named")
	}
	if !strings.Contains(err.Error(), "telemetry.example") {
		t.Errorf("error = %q, want it to name the address", err)
	}
}
