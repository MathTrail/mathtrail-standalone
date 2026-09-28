// Package app wires the service together and runs it.
//
// The container is hand-written on purpose: there is no DI framework, the
// order of construction is the order of the code, and the order of shutdown is
// its reverse.
package app

import (
	"context"
	"net/http"
	"slices"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	drivestore "github.com/MathTrail/mathtrail-standalone/internal/store/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
	oauthserver "github.com/MathTrail/mathtrail-standalone/internal/transport/oauth"
	"github.com/MathTrail/mathtrail-standalone/internal/version"
	"github.com/MathTrail/mathtrail-standalone/internal/widget"
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
	Store     store.Storage
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

	c.Store, err = profileStore(cfg, log, tel.TracerProvider())
	if err != nil {
		return nil, err
	}
	log.Info("profile store", zap.Bool("in_drive", !cfg.DevAuth))

	google, err := googleSignIn(cfg)
	if err != nil {
		return nil, err
	}
	log.Info("google sign-in", zap.Bool("configured", google != nil))

	// The authorization server seals what it issues under the same key ring,
	// each kind under a purpose of its own, and reads the documents clients
	// name themselves by through a fetcher that reaches public addresses only.
	signInServer, err := oauthserver.New(&oauthserver.Settings{
		PublicURL: cfg.Origin(),
		Scope:     mcpserver.Scope,
		Seal:      ring,
		Documents: cimd.NewFetcher(),
		Logger:    log,
		ProjectID: cfg.GCPProjectID,
		Google:    google,
		SiteURL:   cfg.Site(),
		Now:       time.Now,
	})
	if err != nil {
		return nil, err
	}

	// The tools of the lesson seal a task's answer under the key ring above,
	// with the purpose that keeps an answer apart from everything else it
	// seals.
	//
	// The MCP endpoint lets a request in with an access token the
	// authorization server issued, as the account the token signs in, and
	// refuses any other by naming the resource's metadata, where a client
	// begins a sign-in. The development sign-in lets everybody in as one
	// account instead, and the configuration refuses it on a deployment.
	signIn := mcpserver.BearerSignIn(signInServer.Account, signInServer.ResourceMetadataURL)
	if cfg.DevAuth {
		signIn = mcpserver.DevSignIn
	}
	// Each account, each address before a sign-in, and the instance as a whole
	// are held to a pace of their own; a child's day, to the tasks it holds.
	paces, err := newPaces(cfg)
	if err != nil {
		return nil, err
	}
	log.Info("limits set",
		zap.Int("user_per_min", cfg.RateUserPerMin),
		zap.Int("ip_per_min", cfg.RateIPPerMin),
		zap.Int("instance_per_min", cfg.RateInstancePerMin),
		zap.Int("daily_tasks", cfg.DailyTasks),
		zap.Int("daily_failed", cfg.DailyFailed),
	)
	lesson, err := mcpserver.NewService(&mcpserver.Parts{
		Store:     c.Store,
		Content:   embedded,
		Reviewer:  c.Reviewer,
		Sealer:    ring.For(seal.PurposeTaskAnswer),
		Window:    cfg.RequestWindow,
		Daily:     mcpserver.Daily{Tasks: cfg.DailyTasks, Failed: cfg.DailyFailed},
		Now:       time.Now,
		Version:   version.Version,
		Logger:    log,
		Traces:    tel.TracerProvider(),
		ProjectID: cfg.GCPProjectID,
	})
	if err != nil {
		return nil, err
	}
	endpoint, err := mcpserver.NewHandler(&mcpserver.Settings{
		Instructions:        embedded.ServerInstructions(),
		InstructionsVersion: embedded.InstructionsVersion(),
		Version:             version.Version,
		SignIn:              signIn,
		Limits:              mcpserver.Limits{PerAccount: paces.perAccount, Instance: paces.lessons},
		Traces:              tel.TracerProvider(),
		Logger:              log,
		ProjectID:           cfg.GCPProjectID,
		Widget:              widget.Page(),
	}, slices.Concat(lesson.ProfileTools(), lesson.TaskTools())...)
	if err != nil {
		return nil, err
	}

	c.Router, err = httpserver.NewRouter(cfg.PublicURL, &httpserver.Endpoints{
		Health:           httpserver.NewHealthHandler(),
		MCP:              endpoint,
		ResourceMetadata: signInServer.ResourceMetadata,
		ServerMetadata:   signInServer.ServerMetadata,
		Register:         signInServer.Register,
		Authorize:        signInServer.Authorize,
		Consent:          signInServer.Consent,
		Callback:         signInServer.Callback,
		Token:            signInServer.Token,
		Revoke:           signInServer.Revoke,
		Busy:             signInServer.Busy,
	}, httpserver.Limits{PerAddress: paces.perAddress, Instance: paces.signIn}, log, httpserver.Observability{
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

// paces are the paces requests are let in at, each counted in this instance's
// memory alone: every signed-in account's apart, every address's apart before
// anybody signs in, and the instance's at each of its two doors — the
// sign-in, which anybody reaches, and the lessons, which only a signed-in
// account does. The doors keep apart, so that a flood at the sign-in never
// holds back a child in the middle of a lesson.
type paces struct {
	perAccount ratelimit.Limiter
	perAddress ratelimit.Limiter
	signIn     ratelimit.Limiter
	lessons    ratelimit.Limiter
}

func newPaces(cfg *config.Config) (*paces, error) {
	perAccount, err := ratelimit.New(ratelimit.Settings{PerMinute: cfg.RateUserPerMin, Keys: ratelimit.MaxKeys, Now: time.Now})
	if err != nil {
		return nil, err
	}
	perAddress, err := ratelimit.New(ratelimit.Settings{PerMinute: cfg.RateIPPerMin, Keys: ratelimit.MaxKeys, Now: time.Now})
	if err != nil {
		return nil, err
	}
	signIn, err := ratelimit.NewShared(cfg.RateInstancePerMin, time.Now)
	if err != nil {
		return nil, err
	}
	lessons, err := ratelimit.NewShared(cfg.RateInstancePerMin, time.Now)
	if err != nil {
		return nil, err
	}
	return &paces{perAccount: perAccount, perAddress: perAddress, signIn: signIn, lessons: lessons}, nil
}

// googleSignIn is how a parent signs in with Google: through the service's own
// client there, coming back to the authorization server's callback. A machine
// with no client configured has no sign-in, and nil is what it gets; the
// configuration refuses that on a deployment.
func googleSignIn(cfg *config.Config) (googleauth.SignIn, error) {
	if !cfg.GoogleSignIn() {
		return nil, nil
	}
	return googleauth.New(&googleauth.Settings{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.Origin() + oauthserver.CallbackPath,
		Endpoints:    googleauth.Accounts,
		Now:          time.Now,
	})
}

// profileStore is where the profiles are kept. A profile lives in the Drive of
// the parent who signed in, reached with the Google token their sign-in
// carries. The development sign-in carries none, so under it profiles are kept
// in the memory of the process: what is written there is lost with it, and one
// instance knows nothing of what another kept. The configuration refuses that
// sign-in on a deployment.
func profileStore(cfg *config.Config, log *zap.Logger, traces trace.TracerProvider) (store.Storage, error) {
	if cfg.DevAuth {
		return memory.New(), nil
	}
	calls := &drive.Settings{Root: drive.Google, Timeout: cfg.DriveTimeout}
	files, err := drive.NewFiles(calls)
	if err != nil {
		return nil, err
	}
	revisions, err := drive.NewRevisions(calls)
	if err != nil {
		return nil, err
	}
	return drivestore.New(&drivestore.Settings{
		Files:     files,
		Revisions: revisions,
		Timeout:   cfg.DriveTimeout,
		Logger:    log,
		Traces:    traces,
		ProjectID: cfg.GCPProjectID,
	})
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
