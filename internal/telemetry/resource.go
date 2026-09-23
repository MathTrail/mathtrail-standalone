package telemetry

import (
	"context"

	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap"
)

// projectIDKey names the Google Cloud project the data is filed under. It is
// not part of any convention: the collector reads this one attribute to decide
// where a span belongs, and a span without it belongs nowhere.
const projectIDKey = "gcp.project_id"

// newResource describes who is sending the data: this service, this build, and
// — when there is somewhere to send it — the region, the revision and the
// instance underneath.
//
// The description of the platform is asked for only when something is being
// exported, because asking costs a call to a metadata service that exists on a
// deployment and nowhere else.
//
// A platform that answers only partly is not a reason to refuse to start. What
// comes back is used, and what is missing is named in the log: a span that
// cannot say which instance produced it is still a span worth having.
func newResource(ctx context.Context, exporting bool, settings *Settings, log *zap.Logger) *resource.Resource {
	options := []resource.Option{
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(ServiceName),
			semconv.ServiceVersion(settings.Version),
			attribute.String(projectIDKey, settings.ProjectID),
		),
	}
	if exporting {
		options = append(options, resource.WithDetectors(platform(settings)))
	}

	res, err := resource.New(ctx, options...)
	if err != nil {
		log.Warn("telemetry resource incomplete", zap.Error(err))
	}
	return res
}

// platform is what describes the machine underneath, the real one unless a
// caller supplied its own.
func platform(settings *Settings) resource.Detector {
	if settings.Detector != nil {
		return settings.Detector
	}
	return gcp.NewDetector()
}
