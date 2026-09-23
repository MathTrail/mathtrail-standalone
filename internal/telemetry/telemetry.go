// Package telemetry gives the service traces and metrics beside its logs.
//
// The logs stay the source of every number the product asks for: they are
// written whatever happens, they are read without a backend, and nothing here
// replaces them. What this package adds is the shape a log line cannot hold —
// which call happened inside which request, and how long each part of it took.
//
// Two facts about the runtime decide most of the design. The platform traces
// incoming requests itself and puts a sampling decision in the request, so our
// spans join that trace rather than starting one; and it takes the processor
// away once a response has been returned, so anything still held in memory
// after that may never be sent. Hence a parent-based sampler and a delivery
// that happens while the request is still being answered.
//
// The providers are built whether or not anything is exported. Instrumented
// code is written once and asks no questions: off a deployment the same spans
// are created, sampled away and dropped, and no credential is needed to start.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	// ServiceName is the name every span and every metric is attributed to. It
	// is ours rather than the platform's: a service renamed in a console must
	// not split a year of history in two.
	ServiceName = "mathtrail"

	// FlushTimeout bounds a delivery that happens while a request is still
	// being answered. It is short on purpose: a trace is worth less than the
	// answer, and an exporter that has stopped responding must never be what
	// a child waits for.
	FlushTimeout = 200 * time.Millisecond

	// MetricInterval is the shortest gap between two metric deliveries, and
	// the interval of the periodic one. A delivery asked for sooner is
	// declined: each one costs bytes of a monthly allowance, and an aggregate
	// does not change fast enough to be worth more of them.
	MetricInterval = time.Minute

	// exportTimeout bounds one attempt to deliver a batch in the background,
	// where no request is waiting and only the next attempt is delayed.
	exportTimeout = 5 * time.Second
)

// Settings are what the process knows and this package needs.
type Settings struct {
	// Enabled reports whether anything leaves the process. When it is false
	// the providers are still built, and everything they produce is dropped.
	Enabled bool

	// Endpoint is the root URL of the collector. Each signal's own path is
	// appended to it, so it carries a scheme and a host and nothing else.
	Endpoint string

	// SampleRatio is the share of traces kept when a request arrives with no
	// sampling decision of its own. A request that carries one is obeyed.
	SampleRatio float64

	// ProjectID names the Google Cloud project the data belongs to. It travels
	// as a resource attribute, which is the only way the collector learns it.
	ProjectID string

	// Version is the build every span is attributed to.
	Version string

	// HTTPClient carries the credentials the exporter signs with. Nil asks for
	// the credentials of the machine the process runs on.
	HTTPClient *http.Client

	// Detector describes the platform underneath: its region, its revision and
	// the instance this process is. Nil asks the platform itself, which costs
	// a call to its metadata service and only happens when export is on.
	Detector resource.Detector
}

// Telemetry owns the providers and what they hold in memory.
type Telemetry struct {
	traces  *sdktrace.TracerProvider
	metrics *sdkmetric.MeterProvider
	enabled bool

	// mu guards lastMetricFlush, which a delivery reads and writes from
	// whatever goroutine is answering a request.
	mu              sync.Mutex
	lastMetricFlush time.Time
}

// New builds the providers, and attaches exporters to them when the settings
// ask for it.
//
// It refuses nothing that can be worked around. Credentials that cannot be
// found, or a platform that will not describe itself, are reported and then
// left behind: a service must start and answer children whether or not anybody
// is watching it.
func New(ctx context.Context, settings *Settings, log *zap.Logger) (*Telemetry, error) {
	client, err := exportClient(ctx, settings)
	if err != nil {
		// The credentials are the one part of this that belongs to the
		// deployment rather than to the code, and a service that cannot find
		// them is still a service.
		log.Warn("telemetry export unavailable", zap.Error(err))
	}
	exporting := client != nil

	res := newResource(ctx, exporting, settings, log)

	if !exporting {
		log.Info("telemetry built", zap.Bool("export", false))
		return &Telemetry{
			traces: sdktrace.NewTracerProvider(
				sdktrace.WithResource(res),
				// Nothing is kept, so nothing is built to keep it. The
				// decision still reaches the request, which is what lets a
				// delivery know there is nothing to deliver.
				sdktrace.WithSampler(sdktrace.NeverSample()),
			),
			metrics: sdkmetric.NewMeterProvider(sdkmetric.WithResource(res)),
		}, nil
	}

	traceEndpoint, metricEndpoint, err := endpoints(settings.Endpoint)
	if err != nil {
		return nil, err
	}
	traceExporter, err := newTraceExporter(ctx, settings, client, traceEndpoint)
	if err != nil {
		return nil, err
	}
	metricReader, err := newMetricReader(ctx, settings, client, metricEndpoint)
	if err != nil {
		return nil, err
	}

	log.Info("telemetry built",
		zap.Bool("export", true),
		zap.String("endpoint", settings.Endpoint),
		zap.Float64("sample_ratio", settings.SampleRatio),
	)
	return &Telemetry{
		enabled: true,
		traces: sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			// Parent-based, because the platform in front of this process has
			// already decided for every request that reached it; the ratio is
			// the backstop for a request that arrived without a decision.
			sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(settings.SampleRatio))),
			// Before the batcher, so that nothing forbidden is ever queued for
			// sending rather than removed on the way out.
			sdktrace.WithSpanProcessor(Redactor()),
			sdktrace.WithBatcher(traceExporter, sdktrace.WithExportTimeout(exportTimeout)),
		),
		metrics: sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(metricReader),
		),
	}, nil
}

// exportClient is the client the exporters sign their requests with, or nil
// when nothing is to be exported. An error means export was asked for and
// could not be set up; nil and no error means it was not asked for.
func exportClient(ctx context.Context, settings *Settings) (*http.Client, error) {
	switch {
	case !settings.Enabled:
		return nil, nil
	case settings.ProjectID == "":
		// The collector reads the project out of the data itself, and files
		// data that names none nowhere at all.
		return nil, errors.New("telemetry: no Google Cloud project to file the data under")
	case settings.HTTPClient != nil:
		return settings.HTTPClient, nil
	default:
		return credentialedClient(ctx)
	}
}

// TracerProvider is where instrumented code gets its tracer.
func (t *Telemetry) TracerProvider() trace.TracerProvider { return t.traces }

// MeterProvider is where instrumented code gets its meter.
func (t *Telemetry) MeterProvider() metric.MeterProvider { return t.metrics }

// ForceFlush delivers what is held in memory, now. It is meant to be called
// while a request is still being answered, because after that the processor
// may be gone.
//
// Spans are always offered, because a request that has just finished is
// exactly what is waiting to be sent. Measurements are offered only once an
// interval has passed: every delivery costs bytes whether or not anything
// changed.
func (t *Telemetry) ForceFlush(ctx context.Context) error {
	if !t.enabled {
		return nil
	}

	// Its own deadline, and never one longer than the caller already has:
	// whoever is waiting for an answer must not end up waiting for a
	// collector. WithTimeout keeps the earlier of the two.
	ctx, cancel := context.WithTimeout(ctx, FlushTimeout)
	defer cancel()

	err := from("traces", t.traces.ForceFlush(ctx))
	if t.metricsDue(time.Now()) {
		err = errors.Join(err, from("metrics", t.metrics.ForceFlush(ctx)))
	}
	if err != nil {
		return fmt.Errorf("telemetry: flush: %w", err)
	}
	return nil
}

// Shutdown delivers what is left and releases both providers.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if !t.enabled {
		return nil
	}
	err := errors.Join(
		from("traces", t.traces.Shutdown(ctx)),
		from("metrics", t.metrics.Shutdown(ctx)),
	)
	if err != nil {
		return fmt.Errorf("telemetry: shutdown: %w", err)
	}
	return nil
}

// from names which of the two providers an error came from. Both are offered
// the same call and either may refuse it, and "the flush failed" without the
// half that failed is a line nobody can act on.
func from(signal string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", signal, err)
}

// metricsDue reports whether enough time has passed to deliver measurements
// again, and books the delivery when it has.
func (t *Telemetry) metricsDue(now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if now.Sub(t.lastMetricFlush) < MetricInterval {
		return false
	}
	t.lastMetricFlush = now
	return true
}

// ErrorHandler sends what the SDK could not deliver to the service's own log,
// as one line with a named cause. Without it a failed delivery is reported
// through the standard library's logger, which writes a line no collector of
// ours can read and no field of ours can be found in.
//
// It is handed back rather than installed, because there is one such handler
// for the whole binary: that makes it a setting of the process, and a
// constructor that reached for it would change what every other part of the
// same binary reports.
func ErrorHandler(log *zap.Logger) otel.ErrorHandler {
	return otel.ErrorHandlerFunc(func(err error) {
		log.Error("telemetry_export_failed", zap.Error(err))
	})
}
