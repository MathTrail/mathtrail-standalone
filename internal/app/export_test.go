package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// NewContainerReaching builds the container as NewContainer does, reaching the
// sign-in and the Drive given in place of Google's own, and hands that to the
// tests outside the package, so that they can run the service as it is built
// against stand-ins for both.
func NewContainerReaching(ctx context.Context, cfg *config.Config, log *zap.Logger,
	signIn googleauth.Endpoints, driveRoot string,
) (*Container, error) {
	return newContainer(ctx, cfg, log, &outside{signIn: signIn, drive: driveRoot})
}
