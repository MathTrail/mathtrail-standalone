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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	err := run(ctx)
	stop()

	// The exit is the last statement of the process on purpose: os.Exit runs
	// no deferred call, so nothing may be left to one after this point.
	if err != nil {
		fmt.Fprintf(os.Stderr, "mathtrail: %v\n", err)
		os.Exit(1)
	}
}

// run starts the service and serves until ctx is cancelled. Every deferred
// close runs on the way out, which is why nothing here calls os.Exit.
func run(ctx context.Context) error {
	// The configuration is read before anything else exists, so a service that
	// is configured wrongly says so and exits instead of half-starting.
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	defer func() { _ = log.Sync() }()

	// The web framework keeps its mode in a package variable, so it belongs
	// where the process is set up rather than inside whatever builds a router.
	gin.SetMode(gin.ReleaseMode)

	// One handler reports every delivery the telemetry SDK could not make, for
	// the whole binary, and it is a package variable for the same reason.
	otel.SetErrorHandler(telemetry.ErrorHandler(log))

	container, err := app.NewContainer(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("startup: %w", err)
	}
	defer func() {
		closing, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.ShutdownTimeout)
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
	// one part fails, the others are told to stop, and the deferred Close runs
	// only after all of them have.
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return app.NewServer(container).Run(groupCtx) })

	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("shutdown", zap.Error(err))
		return err
	}
	log.Info("shutdown")
	return nil
}
