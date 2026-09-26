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
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
	"github.com/MathTrail/mathtrail-standalone/internal/version"
)

// Container holds everything the process needs while it runs, and knows how to
// close it again. Every part is built in one place, NewContainer, and released
// in one, Close.
type Container struct {
	Config    *config.Config
	Logger    *zap.Logger
	Content   *content.Content
	Seal      *seal.KeyRing
	Telemetry *telemetry.Telemetry
	Solver    solver.Runner
	Reviewer  checks.Reviewer
	Router    http.Handler

	// closers run in reverse order of registration, so that a resource is
	// always closed before whatever it was built from.
	closers []func(context.Context) error
}

// NewContainer builds everything. If construction fails halfway, whatever was
// already built is closed before the error is returned: a half-built container
// must not leak a connection or a goroutine.
func NewContainer(ctx context.Context, cfg *config.Config, log *zap.Logger) (_ *Container, err error) {
	c := &Container{Config: cfg, Logger: log}
	defer func() {
		if err == nil {
			return
		}
		// Detached, with the time a close is given: construction may have
		// failed because its context ended, and a close handed that context
		// would give up before it began.
		closing, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.CloseTimeout())
		defer cancel()
		c.Close(closing)
	}()

	// The content is read and checked before anything is served. A catalog that
	// lost an entry, or a reference task that no longer matches it, is a fault
	// nobody can repair at runtime, so it stops the process here rather than
	// reaching a child in the middle of a lesson.
	embedded, err := content.Load()
	if err != nil {
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
		return nil, err
	}
	// Wrapped rather than instrumented in place: what is worth recording is
	// the whole of a run's result, and the interpreter stays free of anything
	// that watches it.
	observed, err := telemetry.ObserveSolver(sandbox, tel.TracerProvider(), tel.MeterProvider())
	if err != nil {
		return nil, err
	}
	c.Solver = observed
	log.Info("solver sandbox built",
		zap.Uint64("steps", cfg.SolverSteps),
		zap.Duration("timeout", cfg.SolverTimeout),
		zap.Int("concurrency", cfg.SolverConcurrency),
	)

	// A task is reviewed against the content above and run in the sandbox
	// above, the observed one, so that its solver's runs are recorded like
	// any other.
	c.Reviewer = checks.NewReviewer(embedded, c.Solver, checks.DefaultDrawingLimits())

	// The MCP endpoint lets nobody in: this build can issue no token. The
	// development sign-in lets everybody in as one account, and the
	// configuration refuses it on a deployment.
	signIn := mcpserver.NobodySignsIn
	if cfg.DevAuth {
		signIn = mcpserver.DevSignIn
	}
	endpoint, err := mcpserver.NewHandler(&mcpserver.Settings{
		Instructions:        embedded.ServerInstructions(),
		InstructionsVersion: embedded.InstructionsVersion(),
		Version:             version.Version,
		SignIn:              signIn,
		Traces:              tel.TracerProvider(),
		Logger:              log,
		ProjectID:           cfg.GCPProjectID,
	})
	if err != nil {
		return nil, err
	}

	c.Router, err = httpserver.NewRouter(cfg.PublicURL, httpserver.Endpoints{
		Health: httpserver.NewHealthHandler(),
		MCP:    endpoint,
	}, log, httpserver.Observability{
		Traces:    tel.TracerProvider(),
		Meters:    tel.MeterProvider(),
		Flush:     tel.ForceFlush,
		ProjectID: cfg.GCPProjectID,
	})
	if err != nil {
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
