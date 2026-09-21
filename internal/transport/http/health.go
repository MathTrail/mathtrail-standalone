package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/MathTrail/mathtrail-standalone/internal/version"
)

// healthResponse is the body of the probe. It is a type rather than a map so
// that the shape of the answer is stated once, in one place, and a mistyped
// key is a compile error instead of a field nobody notices is missing.
type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// HealthHandler answers the liveness probe.
type HealthHandler struct{}

// NewHealthHandler creates a HealthHandler. It holds nothing today and is
// still built like every other handler, so that the router wires one shape.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health reports that the process is up. It checks nothing else on purpose:
// the service has no dependency it could usefully fail over, and a probe that
// called Google would turn somebody else's outage into our restart loop.
func (h *HealthHandler) Health(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, healthResponse{
		Status:  "ok",
		Version: version.Version,
		Commit:  version.Commit,
	})
}
