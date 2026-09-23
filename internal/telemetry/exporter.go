package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// The paths the two signals are posted to. The protocol puts them under the
// collector's root, and the exporter is told the whole address rather than
// left to assemble one.
const (
	tracePath  = "/v1/traces"
	metricPath = "/v1/metrics"
)

// quotaProjectHeader names the project whose quota a request is counted
// against. Credentials alone do not say it, and a request that names no
// project is refused.
const quotaProjectHeader = "X-Goog-User-Project"

// exportScopes are what the credentials are asked for: permission to append
// spans and to write measurements, and nothing else the same credentials could
// otherwise reach.
var exportScopes = []string{
	"https://www.googleapis.com/auth/trace.append",
	"https://www.googleapis.com/auth/monitoring.write",
}

// credentialedClient signs every export with the credentials of the machine
// the process runs on. There is no key to carry: a deployment is issued
// short-lived tokens by the platform it runs on, and a developer's machine
// usually has none at all, which is why the caller treats a failure here as
// "nothing to export" rather than as a reason not to start.
func credentialedClient(ctx context.Context) (*http.Client, error) {
	// Detached from whatever is starting the process. A token source keeps the
	// context it was built with and refreshes through it, so a source built
	// from the context that ends at shutdown would be unable to sign the last
	// delivery — the one carrying everything the process had not sent yet.
	ctx = context.WithoutCancel(ctx)

	source, err := google.DefaultTokenSource(ctx, exportScopes...)
	if err != nil {
		return nil, fmt.Errorf("telemetry: application default credentials: %w", err)
	}
	return oauth2.NewClient(ctx, source), nil
}

// newTraceExporter posts finished spans to the collector.
func newTraceExporter(ctx context.Context, settings *Settings, client *http.Client, endpoint string) (*otlptrace.Exporter, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(endpoint),
		otlptracehttp.WithHTTPClient(client),
		otlptracehttp.WithHeaders(map[string]string{quotaProjectHeader: settings.ProjectID}),
		otlptracehttp.WithTimeout(exportTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: trace exporter: %w", err)
	}
	return exporter, nil
}

// newMetricReader collects the measurements and posts them to the collector on
// its own clock. Something else has to wake it where a process loses its
// processor between requests, and that is what an asked-for delivery is for.
func newMetricReader(ctx context.Context, settings *Settings, client *http.Client, endpoint string) (sdkmetric.Reader, error) {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(endpoint),
		otlpmetrichttp.WithHTTPClient(client),
		otlpmetrichttp.WithHeaders(map[string]string{quotaProjectHeader: settings.ProjectID}),
		otlpmetrichttp.WithTimeout(exportTimeout),
		otlpmetrichttp.WithTemporalitySelector(deltaTemporality),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: metric exporter: %w", err)
	}
	return sdkmetric.NewPeriodicReader(exporter,
		sdkmetric.WithInterval(MetricInterval),
		sdkmetric.WithTimeout(exportTimeout),
	), nil
}

// endpoints are the two addresses the signals are posted to, read out of the
// collector's root once — because it is one address, and checking it twice
// leaves the second check unable to fail and unable to be tested.
//
// The check matters more than it looks: handed something it cannot parse, the
// exporter keeps its own default of localhost and says nothing at all, so a
// mistyped address becomes spans that leave the process and arrive nowhere,
// with no line anywhere to explain it.
func endpoints(root string) (traces, metrics string, err error) {
	parsed, err := url.Parse(root)
	switch {
	case err != nil:
		return "", "", fmt.Errorf("telemetry: endpoint %q is not a URL: %w", root, err)
	case parsed.Scheme != "http" && parsed.Scheme != "https":
		return "", "", fmt.Errorf("telemetry: endpoint %q must be http or https", root)
	case parsed.Host == "":
		return "", "", fmt.Errorf("telemetry: endpoint %q has no host", root)
	}
	return parsed.JoinPath(tracePath).String(), parsed.JoinPath(metricPath).String(), nil
}

// deltaTemporality asks each delivery to carry what changed since the last one
// rather than a running total. A total has to be held in memory for the life
// of the process and resent whole every time; a change is what the collector
// stores anyway.
//
// The two kinds that can go down are the exception: the difference between two
// readings of a value that falls says nothing on its own, so they keep their
// totals.
func deltaTemporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	switch kind {
	case sdkmetric.InstrumentKindUpDownCounter, sdkmetric.InstrumentKindObservableUpDownCounter:
		return metricdata.CumulativeTemporality
	default:
		return metricdata.DeltaTemporality
	}
}
