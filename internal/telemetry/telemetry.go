// Package telemetry gives the service traces and metrics beside its logs.
//
// The logs stay the source of every number the product asks for: they are
// written whatever happens, they are read without a backend, and nothing here
// replaces them. What this package adds is the shape a log line cannot hold —
// which call happened inside which request, and how long each part of it took.
//
// Two facts about the runtime decide most of the design. The platform traces
// incoming requests itself and says so in the request, so our spans join that
// trace rather than starting one; and it takes the processor away once a
// response has been returned, so anything still held in memory after that may
// never be sent. Hence spans that join the trace a request arrived in, and a
// delivery that happens while the request is still being answered.
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

	"github.com/MathTrail/mathtrail-standalone/internal/telemetry/collector"
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

	// exportTimeout bounds a delivery made in the background, retries and
	// all, where no request is waiting and only the next delivery is delayed.
	// The processors that make the deliveries carry it: the exporters' own
	// timeout applies only to a client they build themselves, and they are
	// handed one.
	exportTimeout = 5 * time.Second
)

// Settings are what the process knows and this package needs.
type Settings struct {
	// Enabled reports whether anything leaves the process. When it is false
	// the providers are still built, and everything they produce is dropped.
	Enabled bool

	// Endpoint is the root URL of the collector. Each signal's own path is
	// appended to it, so it carries a scheme, a host and at most a path to put
	// them under, as collector.Root holds it to.
	Endpoint string

	// SampleRatio is the share of traces kept, whatever decision a request
	// arrives with.
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

	if client == nil {
		res := newResource(ctx, false, settings, log)
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

	// The address before the platform: it is checked without leaving the
	// process, and a start that is going to fail on it should not first wait
	// for the platform to describe itself. And it is checked at all because,
	// handed something it cannot parse, the exporter keeps its own default of
	// localhost and says nothing, so a mistyped address would become spans that
	// leave the process and arrive nowhere.
	root, err := collector.Root(settings.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("telemetry: read the collector's address: %w", err)
	}
	traceEndpoint, metricEndpoint := endpoints(root)
	res := newResource(ctx, true, settings, log)

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
			sdktrace.WithSampler(sampler(settings.SampleRatio)),
			// Before the batcher, so that nothing forbidden is ever queued for
			// sending rather than removed on the way out.
			sdktrace.WithSpanProcessor(Redactor()),
			sdktrace.WithBatcher(withoutStatusText(traceExporter), sdktrace.WithExportTimeout(exportTimeout)),
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
// may be gone — and by every request, because a request is the only time a
// deployed instance has a processor at all.
//
// Spans are sent when the request asking kept its trace: a request that has
// just finished is exactly what is waiting to be sent, and one that kept none
// would only wait behind another's delivery. Measurements are sent once an
// interval has passed, whichever request asks: they belong to no one request,
// and waiting for one whose trace was kept would leave a quiet instance's
// numbers in memory for hours. Every delivery costs bytes whether or not
// anything changed, which is what the interval is for.
func (t *Telemetry) ForceFlush(ctx context.Context, spans bool) error {
	if !t.enabled {
		return nil
	}
	measurements := t.metricsDue(time.Now())
	if !spans && !measurements {
		return nil
	}

	// Its own deadline, and never one longer than the caller already has:
	// whoever is waiting for an answer must not end up waiting for a
	// collector. WithTimeout keeps the earlier of the two.
	ctx, cancel := context.WithTimeout(ctx, FlushTimeout)
	defer cancel()

	// Side by side rather than one after the other, so that each has the
	// whole deadline. The first delivery of a new instance spends most of it
	// on the connection, and measurements queued behind it would miss theirs
	// — and a delivery that misses its deadline loses what it carried.
	var measured, sent error
	var delivering sync.WaitGroup
	if measurements {
		delivering.Go(func() { measured = from("metrics", t.metrics.ForceFlush(ctx)) })
	}
	if spans {
		sent = from("traces", t.traces.ForceFlush(ctx))
	}
	delivering.Wait()

	if err := errors.Join(sent, measured); err != nil {
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

// ErrorHandler sends what the SDK reports going wrong to the service's own
// log, as one line with a named cause. Without it the report goes through the
// standard library's logger, which writes a line no collector of ours can read
// and no field of ours can be found in.
//
// Most of what arrives is a delivery that failed, but not all of it: the SDK
// reports a resource it could not merge or a setting of its own it could not
// read the same way. So the line is named for the telemetry rather than for
// the delivery, and the cause says which it was.
//
// It is handed back rather than installed, because there is one such handler
// for the whole binary: that makes it a setting of the process, and a
// constructor that reached for it would change what every other part of the
// same binary reports.
func ErrorHandler(log *zap.Logger) otel.ErrorHandler {
	return otel.ErrorHandlerFunc(func(err error) {
		log.Error("telemetry_failed", zap.Error(err))
	})
}

// sampler keeps a trace at the configured share, whatever the request it
// arrived in says about it. That decision is not the platform's alone: a
// client sets it in the header too, and one asking for every request to be
// traced would be choosing what this service spends on traces and how long
// its callers wait for deliveries. A span of our own follows its parent, so a
// trace is kept or dropped whole, and the share is read off the trace's own
// identifier, so every process a trace passes through decides it alike.
func sampler(ratio float64) sdktrace.Sampler {
	share := sdktrace.TraceIDRatioBased(ratio)
	return sdktrace.ParentBased(share,
		sdktrace.WithRemoteParentSampled(share),
		sdktrace.WithRemoteParentNotSampled(share),
	)
}
