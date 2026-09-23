// Package app wires the service together and runs it.
//
// The container is hand-written on purpose: there is no DI framework, the
// order of construction is the order of the code, and the order of shutdown is
// its reverse.
package app

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
	"github.com/MathTrail/mathtrail-standalone/internal/version"
)

// Container holds everything the process needs while it runs, and knows how to
// close it again. Today it holds almost nothing; the shape is what matters,
// because every later part is added to exactly one place.
type Container struct {
	Config    *config.Config
	Logger    *zap.Logger
	Content   *content.Content
	Seal      *seal.KeyRing
	Telemetry *telemetry.Telemetry
	Solver    solver.Runner
	Router    http.Handler

	// closers run in reverse order of registration, so that a resource is
	// always closed before whatever it was built from.
	closers []func(context.Context) error
}

// NewContainer builds everything. If construction fails halfway, whatever was
// already built is closed before the error is returned: a half-built container
// must not leak a connection or a goroutine.
func NewContainer(ctx context.Context, cfg *config.Config, log *zap.Logger) (*Container, error) {
	c := &Container{Config: cfg, Logger: log}

	// The content is read and checked before anything is served. A catalog that
	// lost an entry, or a reference task that no longer matches it, is a fault
	// nobody can repair at runtime, so it stops the process here rather than
	// reaching a child in the middle of a lesson.
	embedded, err := content.Load()
	if err != nil {
		c.Close(ctx)
		return nil, err
	}
	c.Content = embedded
	log.Info("content loaded",
		zap.Int("topics", len(embedded.Topics())),
		zap.Int("traps", len(embedded.Traps())),
		zap.Int("skills", len(embedded.Skills())),
		zap.Int("reference_tasks", embedded.ExampleCount()),
		zap.String("instructions_version", embedded.InstructionsVersion()),
	)

	// The key ring is built next, and by the same rule: a key that cannot be
	// read is a service that could issue nothing and open nothing, so it says
	// so now rather than at the first sign-in.
	ring, err := seal.NewKeyRing(cfg.SealKeyCurrent, cfg.SealKeyPrevious)
	if err != nil {
		c.Close(ctx)
		return nil, err
	}
	c.Seal = ring
	log.Info("seal keys loaded",
		zap.String("key_id", ring.CurrentKeyID()),
		zap.Bool("previous_key", ring.PreviousKeyID() != ""),
	)

	// Traces and metrics come before anything worth tracing, and they are
	// built whether or not they are exported: everything below records spans
	// without knowing which of the two it is. This is also the first thing
	// with something to release, so it is the first closer registered and the
	// last one run.
	tel, err := telemetry.New(ctx, &telemetry.Settings{
		Enabled:     cfg.TelemetryEnabled(),
		Endpoint:    cfg.TelemetryEndpoint,
		SampleRatio: cfg.TelemetrySampleRatio,
		ProjectID:   cfg.GCPProjectID,
		Version:     version.Version,
	}, log)
	if err != nil {
		c.Close(ctx)
		return nil, err
	}
	c.Telemetry = tel
	c.closers = append(c.closers, tel.Shutdown)

	// The sandbox is built once and shared. Its vocabulary is the same for
	// every run, and its slots belong to the process: they are what keeps a
	// handful of solvers from taking every core the instance has.
	sandbox, err := starlark.New(starlark.Limits{
		Steps:       cfg.SolverSteps,
		Timeout:     cfg.SolverTimeout,
		Concurrency: cfg.SolverConcurrency,
	})
	if err != nil {
		c.Close(ctx)
		return nil, err
	}
	// Wrapped rather than instrumented in place: what is worth recording is
	// the whole of a run's result, and the interpreter stays free of anything
	// that watches it.
	observed, err := telemetry.ObserveSolver(sandbox, tel.TracerProvider(), tel.MeterProvider())
	if err != nil {
		c.Close(ctx)
		return nil, err
	}
	c.Solver = observed
	log.Info("solver sandbox built",
		zap.Uint64("steps", cfg.SolverSteps),
		zap.Duration("timeout", cfg.SolverTimeout),
		zap.Int("concurrency", cfg.SolverConcurrency),
	)

	c.Router, err = httpserver.NewRouter(httpserver.NewHealthHandler(), log, httpserver.Observability{
		Traces:    tel.TracerProvider(),
		Meters:    tel.MeterProvider(),
		Flush:     tel.ForceFlush,
		ProjectID: cfg.GCPProjectID,
	})
	if err != nil {
		c.Close(ctx)
		return nil, err
	}
	return c, nil
}

// Close releases everything in reverse order of construction. It logs each
// failure and keeps going: one resource refusing to close must not leave the
// others open.
func (c *Container) Close(ctx context.Context) {
	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := c.closers[i](ctx); err != nil {
			c.Logger.Error("close failed", zap.Error(err))
		}
	}
}
