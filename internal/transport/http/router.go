// Package httpserver builds the HTTP surface of the service.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
	"github.com/MathTrail/mathtrail-standalone/internal/transport/http/middleware"
)

// Observability is what the router needs in order to say what it did: where
// to record a span, where to count, what to call to deliver what a request
// leaves behind — its spans when its trace was kept, and the measurements when
// they are due — and which project a log line's trace belongs to.
//
// It is a value rather than a dependency on the telemetry itself, because a
// test of the routing has no use for any of it and can hand over providers
// that record into memory.
type Observability struct {
	Traces    trace.TracerProvider
	Meters    metric.MeterProvider
	Flush     func(ctx context.Context, spans bool) error
	ProjectID string
}

// ErrObservability is returned when the router is given telemetry it could not
// use; callers branch on it with errors.Is.
var ErrObservability = errors.New("http: observability")

// validate refuses a half-filled set, naming what is missing.
//
// Each of the three fails differently and none of them says why: a missing
// tracer is silently replaced by one that records nothing, a missing meter
// stops the process here, and a missing delivery panics after every request,
// which the delivery recovers and turns into a line of its own each time. One
// refusal at startup is worth more than three ways to find out later.
func (o Observability) validate() error {
	switch {
	case o.Traces == nil:
		return fmt.Errorf("%w: Traces must be set", ErrObservability)
	case o.Meters == nil:
		return fmt.Errorf("%w: Meters must be set", ErrObservability)
	case o.Flush == nil:
		return fmt.Errorf("%w: Flush must be set", ErrObservability)
	}
	return nil
}

// NewRouter wires the middleware and the routes.
//
// Today it serves one endpoint. The MCP endpoint, the OAuth endpoints and the
// metadata documents arrive with the code that serves them, and they are
// mounted here, so that the whole surface of the service can be read in one
// function.
//
// The framework's mode is a setting of the process, not of a router, so it is
// chosen where the process starts and never here: a constructor that reaches
// for a global changes what every other router in the same binary does.
func NewRouter(health *HealthHandler, logger *zap.Logger, obs Observability) (*gin.Engine, error) {
	if err := obs.validate(); err != nil {
		return nil, err
	}

	router := gin.New()

	// A known path asked with the wrong method answers 405 rather than 404:
	// the difference is what tells a client it got the path right. The list of
	// methods that would have worked is added by the framework.
	router.HandleMethodNotAllowed = true

	metrics, err := middleware.Metrics(obs.Meters)
	if err != nil {
		return nil, err
	}

	// Order matters. A panic unwinds through every handler it passes without
	// running what they do after the request, so the recovery that answers a
	// handler's panic stands last, next to the handlers, below the ones that
	// count, log and deliver: they then see the answer it made of the panic like
	// any other answer. Above them all stands a last resort for a panic of their
	// own. Between the two: an id, so that everything after it can log it, and
	// tracing, so that everything after it happens inside a span.
	router.Use(middleware.LastResort(logger))
	router.Use(middleware.RequestID())
	router.Use(middleware.Tracing(obs.Traces, obs.Flush, logger)...)
	router.Use(metrics)
	router.Use(middleware.ZapLogger(logger, obs.ProjectID))
	router.Use(middleware.ZapRecovery())

	router.NoRoute(notFound)
	router.NoMethod(methodNotAllowed)

	// Not /healthz: the serverless frontend in front of this process answers
	// that exact path itself, with its own 404, and the request never arrives.
	// HEAD is answered too, since a probe may ask with it.
	router.GET("/health", health.Health)
	router.HEAD("/health", health.Health)

	return router, nil
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, apierror.Response{
		Code:    apierror.CodeNotFound,
		Message: "no such endpoint",
	})
}

func methodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, apierror.Response{
		Code:    apierror.CodeMethodNotAllowed,
		Message: "this endpoint does not accept that method",
	})
}
