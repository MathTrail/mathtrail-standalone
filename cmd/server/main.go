// Command server runs the MathTrail service: one process, one port, and no
// state of its own between requests.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/MathTrail/mathtrail-standalone/internal/app"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/logger"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	"github.com/MathTrail/mathtrail-standalone/internal/version"
)

// main owns the two things a process owns and a function should not: the
// signals it listens for and the code it exits with. Everything else is run,
// which stops because its context was cancelled and says so by returning.
func main() {
	ctx, stop := signalled()
	err := run(ctx)
	stop()

	// The exit is the last statement of the process on purpose: os.Exit runs
	// no deferred call, so nothing may be left to one after this point. A
	// failure the log has recorded is not printed again; one from before there
	// was a log has nowhere else to go.
	if err != nil {
		if !errors.Is(err, errLogged) {
			fmt.Fprintf(os.Stderr, "mathtrail: %v\n", err)
		}
		os.Exit(1)
	}
}

// signalled is a context that ends when a stop is asked for, by either of the
// signals that ask for one. One signal asks; a second ends the process at once.
// Once the first has arrived the signals go back to what they do by default,
// rather than being caught for the whole of the way out, which can take
// seconds nobody pressing Ctrl-C twice means to wait.
func signalled() (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	context.AfterFunc(ctx, stop)
	return ctx, stop
}

// errLogged marks a failure the log has already recorded, so that it is
// reported once.
var errLogged = errors.New("logged")

// logged writes a failure to the log under the name of the moment it happened
// at, and marks it as written.
func logged(log *zap.Logger, moment string, err error) error {
	log.Error(moment, zap.Error(err))
	return fmt.Errorf("%w: %s: %w", errLogged, moment, err)
}

// run starts the service, serves until ctx is cancelled and closes it again.
// Every deferred call runs on the way out, which is why nothing here calls
// os.Exit.
func run(ctx context.Context) error {
	// The configuration is read before anything else exists, so a service that
	// is configured wrongly says so and exits instead of half-starting.
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, err := logger.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	// The web framework keeps its mode in a package variable, so it belongs
	// where the process is set up rather than inside whatever builds a router.
	gin.SetMode(gin.ReleaseMode)

	// One handler reports whatever the telemetry SDK finds going wrong, for the
	// whole binary, and it is a package variable for the same reason.
	otel.SetErrorHandler(telemetry.ErrorHandler(log))

	container, err := app.NewContainer(ctx, cfg, log)
	if err != nil {
		return logged(log, "startup", err)
	}
	// Deferred, so that the container is closed on every way out of here. It
	// gets the share of the way out the server's drain always leaves it.
	defer func() {
		closing, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.CloseTimeout())
		defer cancel()
		container.Close(closing)
	}()

	log.Info("startup",
		zap.String("version", version.Version),
		zap.String("commit", version.Commit),
		zap.String("date", version.Date),
		zap.String("port", cfg.Port),
		zap.String("public_url", cfg.PublicURL),
		zap.Bool("deployed", cfg.Deployed()),
		zap.Bool("dev_auth", cfg.DevAuth),
	)

	// Everything that runs for the life of the process shares one group: when
	// one part fails, the others are told to stop, and the deferred close runs
	// only after all of them have.
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return app.NewServer(container).Run(groupCtx) })

	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return logged(log, "shutdown", err)
	}
	log.Info("shutdown")
	return nil
}
