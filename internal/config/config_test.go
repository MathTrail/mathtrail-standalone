package config_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
)

// The sealing keys of these tests: thirty-two bytes the reader only has to
// recognise as a key, and that nothing here ever seals anything with.
var (
	sealKey         = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("mathtrail"), 4)[:32])
	previousSealKey = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("rotated!!"), 4)[:32])
)

// withSealKey puts the one required secret in front of an environment, so that
// a case about something else does not have to carry it. A case that is about
// the key says so again and wins, because the last value of a variable is the
// one that is read.
func withSealKey(environ ...string) []string {
	return append([]string{"MATHTRAIL_SEAL_KEY_CURRENT=" + sealKey}, environ...)
}

func TestDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(withSealKey())
	if err != nil {
		t.Fatalf("LoadFrom() error = %v, want nil", err)
	}

	if cfg.Port != config.DefaultPort {
		t.Errorf("Port = %q, want %q", cfg.Port, config.DefaultPort)
	}
	if want := "http://localhost:" + config.DefaultPort; cfg.PublicURL != want {
		t.Errorf("PublicURL = %q, want %q", cfg.PublicURL, want)
	}
	if cfg.LogLevel != config.DefaultLogLevel {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, config.DefaultLogLevel)
	}
	if cfg.LogFormat != config.DefaultLogFormat {
		t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, config.DefaultLogFormat)
	}
	if cfg.ReadHeaderTimeout != config.DefaultReadHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want %v", cfg.ReadHeaderTimeout, config.DefaultReadHeaderTimeout)
	}
	if cfg.ShutdownTimeout != config.DefaultShutdownTimeout {
		t.Errorf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, config.DefaultShutdownTimeout)
	}
	if cfg.DevAuth || cfg.Deployed() {
		t.Errorf("DevAuth = %v, Deployed() = %v, want both false", cfg.DevAuth, cfg.Deployed())
	}
	if cfg.SolverSteps != config.DefaultSolverSteps {
		t.Errorf("SolverSteps = %d, want %d", cfg.SolverSteps, uint64(config.DefaultSolverSteps))
	}
	if cfg.SolverTimeout != config.DefaultSolverTimeout {
		t.Errorf("SolverTimeout = %v, want %v", cfg.SolverTimeout, config.DefaultSolverTimeout)
	}
	if cfg.SolverConcurrency != config.DefaultSolverConcurrency {
		t.Errorf("SolverConcurrency = %d, want %d", cfg.SolverConcurrency, config.DefaultSolverConcurrency)
	}
}

func TestValuesAreRead(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom([]string{
		"PORT=9000",
		"MATHTRAIL_SEAL_KEY_CURRENT=" + sealKey,
		"MATHTRAIL_PUBLIC_URL=https://mathtrail.example/",
		"MATHTRAIL_LOG_LEVEL=debug",
		"MATHTRAIL_LOG_FORMAT=console",
		"MATHTRAIL_HTTP_READ_TIMEOUT=45s",
		"MATHTRAIL_SHUTDOWN_TIMEOUT=2m",
		"MATHTRAIL_DEV_AUTH=true",
		"MATHTRAIL_SOLVER_STEPS=250000",
		"MATHTRAIL_SOLVER_TIMEOUT=500ms",
		"MATHTRAIL_SOLVER_CONCURRENCY=2",
	})
	if err != nil {
		t.Fatalf("LoadFrom() error = %v, want nil", err)
	}

	if cfg.Port != "9000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9000")
	}
	if got := cfg.Origin(); got != "https://mathtrail.example" {
		t.Errorf("Origin() = %q, want the URL without its trailing slash", got)
	}
	if cfg.LogLevel != "debug" || cfg.LogFormat != "console" {
		t.Errorf("LogLevel = %q, LogFormat = %q, want debug and console", cfg.LogLevel, cfg.LogFormat)
	}
	if cfg.ReadTimeout != 45*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", cfg.ReadTimeout, 45*time.Second)
	}
	if cfg.ShutdownTimeout != 2*time.Minute {
		t.Errorf("ShutdownTimeout = %v, want %v", cfg.ShutdownTimeout, 2*time.Minute)
	}
	if !cfg.DevAuth {
		t.Error("DevAuth = false, want true")
	}
	if cfg.SolverSteps != 250_000 {
		t.Errorf("SolverSteps = %d, want %d", cfg.SolverSteps, 250_000)
	}
	if cfg.SolverTimeout != 500*time.Millisecond {
		t.Errorf("SolverTimeout = %v, want %v", cfg.SolverTimeout, 500*time.Millisecond)
	}
	if cfg.SolverConcurrency != 2 {
		t.Errorf("SolverConcurrency = %d, want 2", cfg.SolverConcurrency)
	}
}

// A variable a deployment has stopped setting must not override a default with
// nothing.
func TestEmptyValueCountsAsUnset(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(withSealKey("MATHTRAIL_LOG_LEVEL=", "PORT="))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v, want nil", err)
	}
	if cfg.LogLevel != config.DefaultLogLevel {
		t.Errorf("LogLevel = %q, want the default %q", cfg.LogLevel, config.DefaultLogLevel)
	}
	if cfg.Port != config.DefaultPort {
		t.Errorf("Port = %q, want the default %q", cfg.Port, config.DefaultPort)
	}
}

// The whole of 127.0.0.0/8 is this machine, and so is every name reserved
// under .localhost: those may be served over plain http, nothing else may.
func TestLoopbackHostsMayUsePlainHTTP(t *testing.T) {
	t.Parallel()

	for _, host := range []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://127.0.1.1:8080",
		"http://[::1]:8080",
		"http://app.localhost:8080",
		"http://LOCALHOST:8080",
		"http://localhost.:8080",
		"http://App.LocalHost.:8080",
	} {
		t.Run(host, func(t *testing.T) {
			t.Parallel()

			if _, err := config.LoadFrom(withSealKey("MATHTRAIL_PUBLIC_URL=" + host)); err != nil {
				t.Errorf("LoadFrom(%q) error = %v, want nil", host, err)
			}
		})
	}

	for _, host := range []string{
		"http://mathtrail.example",
		"http://10.0.0.1",
		"http://localhost.example.com",
	} {
		t.Run(host, func(t *testing.T) {
			t.Parallel()

			if _, err := config.LoadFrom(withSealKey("MATHTRAIL_PUBLIC_URL=" + host)); err == nil {
				t.Errorf("LoadFrom(%q) error = nil, want https to be required", host)
			}
		})
	}
}

// A process carries other systems' credentials in its environment. None of
// them belongs in the reader that holds this service's configuration.
func TestOnlyDeclaredVariablesAreRead(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(withSealKey(
		"DB_PASSWORD=hunter2",
		"AWS_SECRET_ACCESS_KEY=secret",
		"PATH=/usr/bin",
		"PORT=9200",
	))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v, want nil", err)
	}
	if cfg.Port != "9200" {
		t.Errorf("Port = %q, want the declared variable to be read", cfg.Port)
	}
}

// The decoder names the variables it could not read, all of them at once.
// This is the test that notices if it ever stops.
func TestDecodeFailuresNameEveryBadVariable(t *testing.T) {
	t.Parallel()

	_, err := config.LoadFrom([]string{
		"MATHTRAIL_HTTP_READ_TIMEOUT=half a minute",
		"MATHTRAIL_DEV_AUTH=yes please",
	})
	if err == nil {
		t.Fatal("LoadFrom() error = nil, want a refusal")
	}
	for _, name := range []string{"MATHTRAIL_HTTP_READ_TIMEOUT", "MATHTRAIL_DEV_AUTH"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error = %q, want it to name %s", err, name)
		}
	}
}

func TestDeployed(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(withSealKey(
		"K_SERVICE=mathtrail",
		"MATHTRAIL_PUBLIC_URL=https://mathtrail.example",
	))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v, want nil", err)
	}
	if !cfg.Deployed() {
		t.Error("Deployed() = false, want true when K_SERVICE is set")
	}
}

// Every refusal has to name the variable behind it: a message that says only
// "invalid configuration" makes somebody read the code to deploy a service.
func TestRefusals(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		environ []string
		wantVar string
	}{
		{
			name:    "port is not a number",
			environ: []string{"PORT=eighty"},
			wantVar: "PORT",
		},
		{
			name:    "port is out of range",
			environ: []string{"PORT=70000"},
			wantVar: "PORT",
		},
		{
			name:    "public url is missing in a deployment",
			environ: []string{"K_SERVICE=mathtrail"},
			wantVar: "MATHTRAIL_PUBLIC_URL",
		},
		{
			name:    "public url is not a url",
			environ: []string{"MATHTRAIL_PUBLIC_URL=https://mathtrail.example:eighty"},
			wantVar: "MATHTRAIL_PUBLIC_URL",
		},
		{
			// Nothing but the host is missing, so no other rule refuses it.
			name:    "public url has no host",
			environ: []string{"MATHTRAIL_PUBLIC_URL=https://"},
			wantVar: "MATHTRAIL_PUBLIC_URL",
		},
		{
			name:    "public url is not https",
			environ: []string{"MATHTRAIL_PUBLIC_URL=http://mathtrail.example"},
			wantVar: "MATHTRAIL_PUBLIC_URL",
		},
		{
			name:    "public url has a path",
			environ: []string{"MATHTRAIL_PUBLIC_URL=https://mathtrail.example/mcp"},
			wantVar: "MATHTRAIL_PUBLIC_URL",
		},
		{
			name:    "public url has a query",
			environ: []string{"MATHTRAIL_PUBLIC_URL=https://mathtrail.example?tenant=1"},
			wantVar: "MATHTRAIL_PUBLIC_URL",
		},
		{
			name:    "public url has a fragment",
			environ: []string{"MATHTRAIL_PUBLIC_URL=https://mathtrail.example#top"},
			wantVar: "MATHTRAIL_PUBLIC_URL",
		},
		{
			name:    "public url carries credentials",
			environ: []string{"MATHTRAIL_PUBLIC_URL=https://user:secret@mathtrail.example"},
			wantVar: "MATHTRAIL_PUBLIC_URL",
		},
		{
			name:    "log level is unknown",
			environ: []string{"MATHTRAIL_LOG_LEVEL=loud"},
			wantVar: "MATHTRAIL_LOG_LEVEL",
		},
		{
			name:    "log format is unknown",
			environ: []string{"MATHTRAIL_LOG_FORMAT=xml"},
			wantVar: "MATHTRAIL_LOG_FORMAT",
		},
		{
			name:    "the sealing key is not base64",
			environ: []string{"MATHTRAIL_SEAL_KEY_CURRENT=not base64 at all!"},
			wantVar: "MATHTRAIL_SEAL_KEY_CURRENT",
		},
		{
			name:    "the sealing key is the wrong length",
			environ: []string{"MATHTRAIL_SEAL_KEY_CURRENT=" + base64.StdEncoding.EncodeToString(make([]byte, 16))},
			wantVar: "MATHTRAIL_SEAL_KEY_CURRENT",
		},
		{
			name:    "the previous sealing key is not a key",
			environ: []string{"MATHTRAIL_SEAL_KEY_PREVIOUS=half a key"},
			wantVar: "MATHTRAIL_SEAL_KEY_PREVIOUS",
		},
		{
			// A rotation that only looks done: both variables naming one
			// key means nothing the retired key sealed can still be opened.
			name:    "both sealing keys are the same key",
			environ: []string{"MATHTRAIL_SEAL_KEY_PREVIOUS=" + sealKey},
			wantVar: "MATHTRAIL_SEAL_KEY_PREVIOUS",
		},
		{
			name:    "dev auth is not a boolean",
			environ: []string{"MATHTRAIL_DEV_AUTH=yes please"},
			wantVar: "MATHTRAIL_DEV_AUTH",
		},
		{
			name:    "a timeout is not a duration",
			environ: []string{"MATHTRAIL_HTTP_READ_TIMEOUT=half a minute"},
			wantVar: "MATHTRAIL_HTTP_READ_TIMEOUT",
		},
		{
			name:    "a timeout is zero",
			environ: []string{"MATHTRAIL_SHUTDOWN_TIMEOUT=0s"},
			wantVar: "MATHTRAIL_SHUTDOWN_TIMEOUT",
		},
		{
			name:    "a solver with no steps to spend",
			environ: []string{"MATHTRAIL_SOLVER_STEPS=0"},
			wantVar: "MATHTRAIL_SOLVER_STEPS",
		},
		{
			name:    "a solver with no clock",
			environ: []string{"MATHTRAIL_SOLVER_TIMEOUT=0s"},
			wantVar: "MATHTRAIL_SOLVER_TIMEOUT",
		},
		{
			name:    "a sandbox with no slots",
			environ: []string{"MATHTRAIL_SOLVER_CONCURRENCY=0"},
			wantVar: "MATHTRAIL_SOLVER_CONCURRENCY",
		},
		{
			name:    "telemetry is neither auto, on nor off",
			environ: []string{"MATHTRAIL_TELEMETRY=maybe"},
			wantVar: "MATHTRAIL_TELEMETRY",
		},
		{
			name:    "the sample ratio is above one",
			environ: []string{"MATHTRAIL_TELEMETRY_SAMPLE_RATIO=1.5"},
			wantVar: "MATHTRAIL_TELEMETRY_SAMPLE_RATIO",
		},
		{
			name:    "the sample ratio is negative",
			environ: []string{"MATHTRAIL_TELEMETRY_SAMPLE_RATIO=-0.1"},
			wantVar: "MATHTRAIL_TELEMETRY_SAMPLE_RATIO",
		},
		{
			name:    "the collector is not https",
			environ: []string{"MATHTRAIL_TELEMETRY_ENDPOINT=http://telemetry.example"},
			wantVar: "MATHTRAIL_TELEMETRY_ENDPOINT",
		},
		{
			name:    "the collector has no host",
			environ: []string{"MATHTRAIL_TELEMETRY_ENDPOINT=https:///v1"},
			wantVar: "MATHTRAIL_TELEMETRY_ENDPOINT",
		},
		{
			name:    "the collector is not an address at all",
			environ: []string{"MATHTRAIL_TELEMETRY_ENDPOINT=http://[::1"},
			wantVar: "MATHTRAIL_TELEMETRY_ENDPOINT",
		},
		{
			name:    "a way out shorter than a second",
			environ: []string{"MATHTRAIL_SHUTDOWN_TIMEOUT=500ms"},
			wantVar: "MATHTRAIL_SHUTDOWN_TIMEOUT",
		},
		{
			name:    "the collector names a user",
			environ: []string{"MATHTRAIL_TELEMETRY_ENDPOINT=https://someone:hunter2@telemetry.example"},
			wantVar: "MATHTRAIL_TELEMETRY_ENDPOINT",
		},
		{
			name:    "the collector carries a query",
			environ: []string{"MATHTRAIL_TELEMETRY_ENDPOINT=https://telemetry.example?key=value"},
			wantVar: "MATHTRAIL_TELEMETRY_ENDPOINT",
		},
		{
			// This machine is let off https, not off the exporter's own rule.
			name:    "the collector on this machine speaks no protocol the exporter does",
			environ: []string{"MATHTRAIL_TELEMETRY_ENDPOINT=ftp://localhost:4318"},
			wantVar: "MATHTRAIL_TELEMETRY_ENDPOINT",
		},
		{
			// The whole point of the switch: in a deployment it is refused,
			// not warned about.
			name: "dev auth in a deployment",
			environ: []string{
				"MATHTRAIL_DEV_AUTH=true",
				"K_SERVICE=mathtrail",
				"MATHTRAIL_PUBLIC_URL=https://mathtrail.example",
			},
			wantVar: "MATHTRAIL_DEV_AUTH",
		},
		{
			// A line a person reads at a terminal is not one a collector can
			// trust: it escapes nothing a request carries.
			name: "the console format in a deployment",
			environ: []string{
				"MATHTRAIL_LOG_FORMAT=console",
				"K_SERVICE=mathtrail",
				"MATHTRAIL_PUBLIC_URL=https://mathtrail.example",
			},
			wantVar: "MATHTRAIL_LOG_FORMAT",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := config.LoadFrom(withSealKey(tc.environ...))
			if err == nil {
				t.Fatalf("LoadFrom() error = nil, want an error naming %s", tc.wantVar)
			}
			if !errors.Is(err, config.ErrInvalid) {
				t.Errorf("errors.Is(err, ErrInvalid) = false, want true; err = %v", err)
			}
			if !strings.Contains(err.Error(), tc.wantVar) {
				t.Errorf("error = %q, want it to name %s", err, tc.wantVar)
			}
		})
	}
}

// Load is LoadFrom over the process environment; this is the one test that
// proves the two are connected.
func TestLoadReadsTheProcessEnvironment(t *testing.T) {
	t.Setenv("PORT", "9100")
	t.Setenv("MATHTRAIL_LOG_LEVEL", "warn")
	t.Setenv("MATHTRAIL_SEAL_KEY_CURRENT", sealKey)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Port != "9100" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9100")
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "warn")
	}
}

// The one secret with no default. A service handed no sealing key is refused
// by name: inventing one would work until the first restart and then sign
// every parent out without ever saying why.
func TestTheSealingKeyIsRequired(t *testing.T) {
	t.Parallel()

	_, err := config.LoadFrom(nil)
	if !errors.Is(err, config.ErrInvalid) {
		t.Fatalf("LoadFrom() error = %v, want it to wrap ErrInvalid", err)
	}
	if !strings.Contains(err.Error(), "MATHTRAIL_SEAL_KEY_CURRENT") {
		t.Errorf("error = %q, want it to name MATHTRAIL_SEAL_KEY_CURRENT", err)
	}
}

// While a key is being rotated the service carries two, and neither variable
// is read anywhere but here.
func TestBothSealingKeysAreRead(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(withSealKey("MATHTRAIL_SEAL_KEY_PREVIOUS=" + previousSealKey))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v, want nil", err)
	}
	if cfg.SealKeyCurrent != sealKey {
		t.Error("SealKeyCurrent is not the key that was set")
	}
	if cfg.SealKeyPrevious != previousSealKey {
		t.Error("SealKeyPrevious is not the key that was set")
	}
}

// A previous key of whitespace alone is no previous key, as the key ring reads
// it: a secret emptied at the end of a rotation arrives as a lone newline, and
// a service that refused to start over it would be down for nothing.
func TestAPreviousKeyOfWhitespaceIsNone(t *testing.T) {
	t.Parallel()

	for name, previous := range map[string]string{"a lone newline": "\n", "spaces": "   "} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := config.LoadFrom(withSealKey("MATHTRAIL_SEAL_KEY_PREVIOUS=" + previous)); err != nil {
				t.Errorf("LoadFrom(previous key %q) error = %v, want nil", previous, err)
			}
		})
	}
}

// A refusal about a key names the variable and nothing else: the value behind
// it is a secret, and an error message is read in places a secret may not go.
func TestARefusalNeverCarriesTheSealingKey(t *testing.T) {
	t.Parallel()

	// Base64 of too few bytes: shaped like a key, and refused like one.
	nearlyAKey := base64.StdEncoding.EncodeToString([]byte("too short to seal with"))

	_, err := config.LoadFrom([]string{"MATHTRAIL_SEAL_KEY_CURRENT=" + nearlyAKey})
	if err == nil {
		t.Fatal("LoadFrom() error = nil, want a refusal")
	}
	if strings.Contains(err.Error(), nearlyAKey) {
		t.Errorf("error = %q, want it to carry nothing of the key", err)
	}
}

// Telemetry is configured by four variables, and all four have a setting that
// works without one being given.
func TestTelemetryDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(withSealKey())
	if err != nil {
		t.Fatalf("LoadFrom() error = %v, want nil", err)
	}

	if cfg.Telemetry != config.DefaultTelemetry {
		t.Errorf("Telemetry = %q, want %q", cfg.Telemetry, config.DefaultTelemetry)
	}
	if cfg.TelemetryEndpoint != config.DefaultTelemetryEndpoint {
		t.Errorf("TelemetryEndpoint = %q, want %q", cfg.TelemetryEndpoint, config.DefaultTelemetryEndpoint)
	}
	if cfg.TelemetrySampleRatio != config.DefaultTelemetrySampleRatio {
		t.Errorf("TelemetrySampleRatio = %v, want %v", cfg.TelemetrySampleRatio, config.DefaultTelemetrySampleRatio)
	}
	if cfg.GCPProjectID != "" {
		t.Errorf("GCPProjectID = %q, want empty", cfg.GCPProjectID)
	}
}

// The switch says where telemetry is exported. Its usual setting asks the one
// question that distinguishes the two places this runs: a deployment is
// watched, a developer's machine is not.
func TestTelemetryFollowsTheDeployment(t *testing.T) {
	t.Parallel()

	deployed := []string{
		"K_SERVICE=mathtrail",
		"MATHTRAIL_PUBLIC_URL=https://mathtrail.example",
	}

	for _, c := range []struct {
		name    string
		environ []string
		want    bool
	}{
		{name: "auto on a developer's machine", want: false},
		{name: "auto in a deployment", environ: deployed, want: true},
		{
			name:    "off in a deployment",
			environ: append([]string{"MATHTRAIL_TELEMETRY=off"}, deployed...),
			want:    false,
		},
		{
			name:    "on off a deployment",
			environ: []string{"MATHTRAIL_TELEMETRY=on"},
			want:    true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := config.LoadFrom(withSealKey(c.environ...))
			if err != nil {
				t.Fatalf("LoadFrom() error = %v, want nil", err)
			}
			if got := cfg.TelemetryEnabled(); got != c.want {
				t.Errorf("TelemetryEnabled() = %v, want %v", got, c.want)
			}
		})
	}
}

// The drain and the close share the way out rather than take it each: the
// two together are the whole of it, and the close always keeps a share for
// the last delivery of what the service recorded.
func TestTheWayOutIsSharedRatherThanTakenTwice(t *testing.T) {
	t.Parallel()

	for _, budget := range []time.Duration{time.Second, config.DefaultShutdownTimeout, 7 * time.Second} {
		t.Run(budget.String(), func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{ShutdownTimeout: budget}
			if got := cfg.DrainTimeout() + cfg.CloseTimeout(); got != budget {
				t.Errorf("drain + close = %v, want the whole way out, %v", got, budget)
			}
			if cfg.CloseTimeout() != budget/4 {
				t.Errorf("close = %v, want the quarter of %v the drain leaves it", cfg.CloseTimeout(), budget)
			}
		})
	}
}
