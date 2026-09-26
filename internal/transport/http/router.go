// Package httpserver builds the HTTP surface of the service.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

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

// Endpoints are the handlers the router serves.
type Endpoints struct {
	// Health answers the platform's probe.
	Health *HealthHandler
	// MCP is the MCP endpoint as it is served, its sign-in included.
	MCP http.Handler
}

// ErrEndpoints is returned when the router is given an address or endpoints it
// could not serve; callers branch on it with errors.Is.
var ErrEndpoints = errors.New("http: endpoints")

// validate refuses a missing handler, naming it, rather than mounting a route
// that panics on its first request.
func (e Endpoints) validate() error {
	switch {
	case e.Health == nil:
		return fmt.Errorf("%w: Health must be set", ErrEndpoints)
	case e.MCP == nil:
		return fmt.Errorf("%w: MCP must be set", ErrEndpoints)
	}
	return nil
}

// NewRouter wires the middleware and the routes.
//
// The public URL is the service's own address: only its host is served, and
// only its origin may call the MCP endpoint from a browser. The OAuth
// endpoints and the metadata documents arrive with the code that serves them,
// and they are mounted here too, so that the whole surface of the service can
// be read in one function.
//
// The framework's mode is a setting of the process, not of a router, so it is
// chosen where the process starts and never here: a constructor that reaches
// for a global changes what every other router in the same binary does.
func NewRouter(publicURL string, endpoints Endpoints, logger *zap.Logger, obs Observability) (*gin.Engine, error) {
	if err := endpoints.validate(); err != nil {
		return nil, err
	}
	if err := obs.validate(); err != nil {
		return nil, err
	}
	public, err := url.Parse(publicURL)
	if err != nil {
		return nil, fmt.Errorf("%w: the public URL is not an address: %w", ErrEndpoints, err)
	}
	if public.Host == "" {
		return nil, fmt.Errorf("%w: the public URL has no host", ErrEndpoints)
	}

	router := gin.New()

	// A known path asked with the wrong method answers 405 rather than 404:
	// the difference is what tells a client it got the path right. The list of
	// methods that would have worked is added by the framework.
	router.HandleMethodNotAllowed = true

	// A path is served under its one name. The framework would otherwise
	// answer the same path with a slash added or taken away by redirecting it
	// itself, before any handler runs — past the host check, the headers every
	// answer carries, and the request's own line.
	router.RedirectTrailingSlash = false

	metrics, err := middleware.Metrics(obs.Meters)
	if err != nil {
		return nil, err
	}

	// Order matters. A panic unwinds through every handler it passes without
	// running what they do after the request, so the recovery that answers a
	// handler's panic stands last, next to the handlers, below the ones that
	// count, log and deliver: they then see the answer it made of the panic like
	// any other answer. Above them all stands a last resort for a panic of their
	// own, and right below it the headers every answer carries, so that even
	// the last resort's answer has them. Then an id, so that everything after
	// it can log it, and tracing, so that everything after it happens inside a
	// span. The host is checked after the recovery, so that a request under a
	// name that is not ours is refused like any other answer: logged, counted
	// and traced.
	router.Use(middleware.LastResort(logger))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RequestID())
	router.Use(middleware.Tracing(obs.Traces, obs.Flush, logger)...)
	router.Use(metrics)
	router.Use(middleware.ZapLogger(logger, obs.ProjectID))
	router.Use(middleware.ZapRecovery())
	router.Use(middleware.Host(public))

	router.NoRoute(notFound)
	router.NoMethod(methodNotAllowed)

	// Not /healthz: the serverless frontend in front of this process answers
	// that exact path itself, with its own 404, and the request never arrives.
	// HEAD is answered too, since a probe may ask with it.
	router.GET("/health", endpoints.Health.Health)
	router.HEAD("/health", endpoints.Health.Health)

	// The MCP endpoint takes both methods the protocol names and lets the
	// protocol answer each: without sessions there is no stream to open, and
	// it refuses a GET itself.
	router.Match([]string{http.MethodGet, http.MethodPost}, "/mcp",
		middleware.Origin(public),
		gin.WrapH(endpoints.MCP),
	)

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
