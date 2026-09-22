// Package httpserver builds the HTTP surface of the service.
package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
	"github.com/MathTrail/mathtrail-standalone/internal/transport/http/middleware"
)

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
func NewRouter(health *HealthHandler, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	// A known path asked with the wrong method answers 405 rather than 404:
	// the difference is what tells a client it got the path right. The list of
	// methods that would have worked is added by the framework.
	router.HandleMethodNotAllowed = true

	// Order matters: an id first so that everything downstream can log it,
	// recovery next so that it catches panics from the handlers below, and the
	// request log last so that it sees the status recovery produced.
	router.Use(middleware.RequestID())
	router.Use(middleware.ZapRecovery(logger))
	router.Use(middleware.ZapLogger(logger))

	router.NoRoute(notFound)
	router.NoMethod(methodNotAllowed)

	// Not /healthz: the serverless frontend in front of this process answers
	// that exact path itself, with its own 404, and the request never arrives.
	router.GET("/health", health.Health)

	return router
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
