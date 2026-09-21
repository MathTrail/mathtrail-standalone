package config_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
)

func TestDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(nil)
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
}

func TestValuesAreRead(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom([]string{
		"PORT=9000",
		"MATHTRAIL_PUBLIC_URL=https://mathtrail.example/",
		"MATHTRAIL_LOG_LEVEL=debug",
		"MATHTRAIL_LOG_FORMAT=console",
		"MATHTRAIL_HTTP_READ_TIMEOUT=45s",
		"MATHTRAIL_SHUTDOWN_TIMEOUT=2m",
		"MATHTRAIL_DEV_AUTH=true",
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
}

// A variable a deployment has stopped setting must not override a default with
// nothing.
func TestEmptyValueCountsAsUnset(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom([]string{"MATHTRAIL_LOG_LEVEL=", "PORT="})
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
	} {
		t.Run(host, func(t *testing.T) {
			t.Parallel()

			if _, err := config.LoadFrom([]string{"MATHTRAIL_PUBLIC_URL=" + host}); err != nil {
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

			if _, err := config.LoadFrom([]string{"MATHTRAIL_PUBLIC_URL=" + host}); err == nil {
				t.Errorf("LoadFrom(%q) error = nil, want https to be required", host)
			}
		})
	}
}

// A process carries other systems' credentials in its environment. None of
// them belongs in the reader that holds this service's configuration.
func TestOnlyDeclaredVariablesAreRead(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom([]string{
		"DB_PASSWORD=hunter2",
		"AWS_SECRET_ACCESS_KEY=secret",
		"PATH=/usr/bin",
		"PORT=9200",
	})
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

	cfg, err := config.LoadFrom([]string{
		"K_SERVICE=mathtrail",
		"MATHTRAIL_PUBLIC_URL=https://mathtrail.example",
	})
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := config.LoadFrom(tc.environ)
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
