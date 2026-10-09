// Package app wires the service together and runs it.
//
// The container is hand-written on purpose: there is no DI framework, the
// order of construction is the order of the code, and the order of shutdown is
// its reverse.
package app

import (
	"context"
	"errors"
	"math"
	"net/http"
	"runtime"
	"runtime/debug"
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
	"github.com/MathTrail/mathtrail-standalone/internal/infra/geoip"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/internal/learner"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	drivestore "github.com/MathTrail/mathtrail-standalone/internal/store/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
	"github.com/MathTrail/mathtrail-standalone/internal/transport/http/middleware"
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

// outside is where the service reaches the two services of Google's that a
// parent's sign-in and profile live with: its sign-in, and Drive.
type outside struct {
	signIn googleauth.Endpoints
	drive  string
}

// NewContainer builds everything, reaching Google's own sign-in and Drive. If
// construction fails halfway, whatever was already built is closed before the
// error is returned: a half-built container must not leak a connection or a
// goroutine.
func NewContainer(ctx context.Context, cfg *config.Config, log *zap.Logger) (*Container, error) {
	return newContainer(ctx, cfg, log, &outside{signIn: googleauth.Accounts, drive: drive.Google})
}

// newContainer builds everything, reaching Google where reach says.
func newContainer(ctx context.Context, cfg *config.Config, log *zap.Logger, reach *outside) (_ *Container, err error) {
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
		Wait:        cfg.SolverWait,
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
		zap.Duration("wait", cfg.SolverWait),
		// What the instance gave the runtime, beside the slots it was given:
		// the processors it schedules on and the soft limit its collector
		// keeps the heap under, so that a deployment shows what it got.
		zap.Int("gomaxprocs", runtime.GOMAXPROCS(0)),
		zap.Int64("memory_limit", softMemoryLimit()),
	)

	// A task is reviewed against the content above and run in the sandbox
	// above, the observed one, so that its solver's runs are recorded like
	// any other.
	c.Reviewer = checks.NewReviewer(embedded, c.Solver)

	c.Store, err = profileStore(cfg, log, tel.TracerProvider(), reach.drive)
	if err != nil {
		return nil, err
	}
	log.Info("profile store", zap.Bool("in_drive", !cfg.DevAuth))

	google, reviewer, err := signInsOf(cfg, reach.signIn, log)
	if err != nil {
		return nil, err
	}

	counted, err := c.censusOf(cfg, log)
	if err != nil {
		return nil, err
	}

	// Each account, each address before a sign-in, the instance as a whole and
	// each account's renewals at Google are held to a pace of their own; a
	// child's day, to the tasks it holds.
	paces, err := newPaces(cfg)
	if err != nil {
		return nil, err
	}

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
		Renewals:  paces.renewals,
		SiteURL:   cfg.Site(),
		CountryOf: counted.countryOf,
		Reviewer:  reviewer,
		Now:       time.Now,
	})
	if err != nil {
		return nil, err
	}

	// The tools of the lesson seal a task's answer under the key ring above,
	// with the purpose that keeps an answer apart from everything else it
	// seals.
	signIn := signInOf(cfg, signInServer)
	log.Info("limits set",
		zap.Int("user_per_min", cfg.RateUserPerMin),
		zap.Int("ip_per_min", cfg.RateIPPerMin),
		zap.Int("instance_per_min", cfg.RateInstancePerMin),
		zap.Int("renewal_per_min", cfg.RateRenewalPerMin),
		zap.Int("daily_tasks", cfg.DailyTasks),
		zap.Int("daily_failed", cfg.DailyFailed),
		zap.Int("trap_repeats", cfg.TrapRepeats),
	)
	lesson, err := mcpserver.NewService(&mcpserver.Parts{
		Store:        c.Store,
		Content:      embedded,
		Reviewer:     c.Reviewer,
		Sealer:       ring.For(seal.PurposeTaskAnswer),
		Window:       cfg.RequestWindow,
		Daily:        mcpserver.Daily{Tasks: cfg.DailyTasks, Failed: cfg.DailyFailed},
		TrapRepeats:  cfg.TrapRepeats,
		Now:          time.Now,
		Version:      version.Version,
		Logger:       log,
		Traces:       tel.TracerProvider(),
		ProjectID:    cfg.GCPProjectID,
		Learners:     counted.learners,
		DemoAccounts: signInServer.DemoAccounts,
		SiteURL:      cfg.Site(),
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
		Origin:              cfg.Origin(),
		Site:                cfg.Site(),
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
		Drive:            signInServer.Drive,
		Token:            signInServer.Token,
		Revoke:           signInServer.Revoke,
		Busy:             signInServer.Busy,
		Challenge:        cfg.OpenAIChallenge,
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
// holds back a child in the middle of a lesson. Beside them, every account's
// renewals at Google are counted apart: not requests let in, but calls to
// Google the sign-in makes for an account.
type paces struct {
	perAccount ratelimit.Limiter
	perAddress ratelimit.Limiter
	signIn     ratelimit.Limiter
	lessons    ratelimit.Limiter
	renewals   ratelimit.Limiter
}

// newPaces builds the paces from the numbers configured. Each is built whether
// or not one before it failed, so that a refusal names every number nothing
// could be let in at.
func newPaces(cfg *config.Config) (*paces, error) {
	perAccount, accountErr := ratelimit.New(ratelimit.Settings{PerMinute: cfg.RateUserPerMin, Keys: ratelimit.MaxKeys, Now: time.Now})
	perAddress, addressErr := ratelimit.New(ratelimit.Settings{PerMinute: cfg.RateIPPerMin, Keys: ratelimit.MaxKeys, Now: time.Now})
	signIn, signInErr := ratelimit.NewShared(cfg.RateInstancePerMin, time.Now)
	lessons, lessonsErr := ratelimit.NewShared(cfg.RateInstancePerMin, time.Now)
	renewals, renewalsErr := ratelimit.New(ratelimit.Settings{PerMinute: cfg.RateRenewalPerMin, Keys: ratelimit.MaxKeys, Now: time.Now})
	if err := errors.Join(accountErr, addressErr, signInErr, lessonsErr, renewalsErr); err != nil {
		return nil, err
	}
	return &paces{perAccount: perAccount, perAddress: perAddress, signIn: signIn, lessons: lessons, renewals: renewals}, nil
}

// reviewerSignIn is the sign-in of a directory's reviewers as the demo account,
// or nil when the deployment configures none.
func reviewerSignIn(cfg *config.Config) (*oauthserver.Reviewer, error) {
	grant, err := cfg.ReviewerSignIn()
	if grant == nil || err != nil {
		return nil, err
	}
	return &oauthserver.Reviewer{Password: cfg.ReviewerPassword, Subject: grant.Subject, RefreshToken: grant.RefreshToken}, nil
}

// signInOf is how the MCP endpoint lets a request in: with an access token
// the authorization server issued, as the account the token signs in, refusing
// any other by naming the resource's metadata, where a client begins a
// sign-in. The development sign-in refuses nobody instead: a request whose
// bearer credential is a name acts for an account of that name, and any other
// for one development account. The configuration refuses it on a deployment.
func signInOf(cfg *config.Config, server *oauthserver.Server) mcpserver.SignIn {
	if cfg.DevAuth {
		return mcpserver.DevSignIn
	}
	return mcpserver.BearerSignIn(server.Account, server.ResourceMetadataURL)
}

// census is what the lines that count the children need: the key the name a
// child is counted under each month is derived from, and the country a request
// came from, which the sign-in asks of the one request the parent's browser
// makes.
type census struct {
	learners  *learner.Key
	countryOf func(*http.Request) string
}

// censusOf builds what the lines that count the children need, and says what
// it built. The database of countries is opened here and closed on the way
// out: a file that does not open stops the process now, as the content does,
// rather than at the first parent who signs in.
func (c *Container) censusOf(cfg *config.Config, log *zap.Logger) (*census, error) {
	learners, err := learnerKey(cfg)
	if err != nil {
		return nil, err
	}
	log.Info("learner key", zap.Bool("configured", cfg.LearnerKey != ""))
	countryOf, err := c.countries(cfg, log)
	if err != nil {
		return nil, err
	}
	return &census{learners: learners, countryOf: countryOf}, nil
}

// learnerKey is the key the children are counted under in the log: the one
// configured, or, on a machine given none, one the process makes for itself,
// whose names mean something only while it runs. The configuration refuses
// that on a deployment.
func learnerKey(cfg *config.Config) (*learner.Key, error) {
	if cfg.LearnerKey == "" {
		return learner.RandomKey(), nil
	}
	return learner.NewKey(cfg.LearnerKey)
}

// countries opens the database of countries, when one is configured, and is
// the country a request came from as the sign-in asks it: the address the
// platform saw the request come from, looked up there. With none configured
// it is nil, and no country is known.
func (c *Container) countries(cfg *config.Config, log *zap.Logger) (func(*http.Request) string, error) {
	if cfg.CountryDB == "" {
		log.Info("country database", zap.Bool("configured", false))
		return nil, nil
	}
	database, err := geoip.Open(cfg.CountryDB)
	if err != nil {
		return nil, err
	}
	c.closers = append(c.closers, func(context.Context) error { return database.Close() })
	about := database.About()
	log.Info("country database",
		zap.Bool("configured", true),
		zap.String("type", about.Type),
		zap.String("built", about.Built.Format(time.DateOnly)),
	)
	return func(r *http.Request) string {
		address, readable := middleware.Sender(r)
		if !readable {
			return ""
		}
		return database.Country(address)
	}, nil
}

// signInsOf are the two ways in a sign-in has, and it says which it has: a
// parent's, through Google, and a directory's reviewer's, by the reviewers'
// password.
func signInsOf(cfg *config.Config, endpoints googleauth.Endpoints, log *zap.Logger) (googleauth.SignIn, *oauthserver.Reviewer, error) {
	google, err := googleSignIn(cfg, endpoints)
	if err != nil {
		return nil, nil, err
	}
	log.Info("google sign-in", zap.Bool("configured", google != nil))
	reviewer, err := reviewerSignIn(cfg)
	if err != nil {
		return nil, nil, err
	}
	log.Info("reviewer sign-in", zap.Bool("configured", reviewer != nil))
	return google, reviewer, nil
}

// googleSignIn is how a parent signs in with Google: through the service's own
// client there, at the endpoints given, coming back to the authorization
// server's callback. A machine with no client configured has no sign-in, and
// nil is what it gets; the configuration refuses that on a deployment.
func googleSignIn(cfg *config.Config, endpoints googleauth.Endpoints) (googleauth.SignIn, error) {
	if !cfg.GoogleSignIn() {
		return nil, nil
	}
	return googleauth.New(&googleauth.Settings{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.Origin() + oauthserver.CallbackPath,
		Endpoints:    endpoints,
		Now:          time.Now,
	})
}

// profileStore is where the profiles are kept. A profile lives in the Drive of
// the parent who signed in, reached at the root given with the Google token
// their sign-in carries. The development sign-in carries none, so under it
// profiles are kept in the memory of the process: what is written there is
// lost with it, and one instance knows nothing of what another kept. The
// configuration refuses that sign-in on a deployment.
func profileStore(cfg *config.Config, log *zap.Logger, traces trace.TracerProvider, root string) (store.Storage, error) {
	if cfg.DevAuth {
		return memory.New(), nil
	}
	calls := &drive.Settings{Root: root, Timeout: cfg.DriveTimeout}
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

// softMemoryLimit is the soft limit the runtime's collector keeps the heap under,
// as GOMEMLIMIT set it, or nothing when it was not set.
func softMemoryLimit() int64 {
	if limit := debug.SetMemoryLimit(-1); limit != math.MaxInt64 {
		return limit
	}
	return 0
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
