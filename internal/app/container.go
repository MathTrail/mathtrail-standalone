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

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	httpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/http"
)

// Container holds everything the process needs while it runs, and knows how to
// close it again. Today it holds almost nothing; the shape is what matters,
// because every later part is added to exactly one place.
type Container struct {
	Config *config.Config
	Logger *zap.Logger
	Router http.Handler

	// closers run in reverse order of registration, so that a resource is
	// always closed before whatever it was built from.
	closers []func(context.Context) error
}

// NewContainer builds everything. If construction fails halfway, whatever was
// already built is closed before the error is returned: a half-built container
// must not leak a connection or a goroutine.
func NewContainer(ctx context.Context, cfg *config.Config, log *zap.Logger) (*Container, error) {
	c := &Container{Config: cfg, Logger: log}
	c.Router = httpserver.NewRouter(httpserver.NewHealthHandler(), log)
	_ = ctx // nothing here blocks yet; the signature is the one every later part needs
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
