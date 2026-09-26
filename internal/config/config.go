// Package config reads the service's configuration from the environment.
//
// Environment variables are the only source: there is no file, no flag and no
// remote source. They are read here and nowhere else, so that every value has
// one origin and one validation. Viper does the reading — as it does in the
// platform's other services — which keeps the names, the defaults and the
// types of a growing table in one declaration.
//
// Only the variables the service actually uses today are here. The ones the
// sign-in, the limits and the solver will need arrive with the code that
// consumes them, because a required variable with no reader is a deployment
// that fails for no reason.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry/collector"
)

// The three settings of the telemetry switch. A deployment is watched and a
// developer's machine is not, so the usual answer is "whichever this is"; the
// other two are for the times that is wrong.
const (
	// TelemetryAuto exports when the process is running as a deployed service.
	TelemetryAuto = "auto"
	// TelemetryOn exports wherever the process runs.
	TelemetryOn = "on"
	// TelemetryOff exports nowhere.
	TelemetryOff = "off"
)

// Defaults for every optional variable. They are named constants rather than
// literals in Load so that the documentation and the code cannot drift.
const (
	DefaultPort              = "8080"
	DefaultLogLevel          = "info"
	DefaultLogFormat         = "json"
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultReadTimeout       = 30 * time.Second
	DefaultWriteTimeout      = 60 * time.Second
	DefaultIdleTimeout       = 120 * time.Second
	DefaultShutdownTimeout   = 9 * time.Second

	DefaultSolverSteps       = 25_000_000
	DefaultSolverTimeout     = 2 * time.Second
	DefaultSolverConcurrency = 4

	DefaultTelemetry            = TelemetryAuto
	DefaultTelemetryEndpoint    = "https://telemetry.googleapis.com"
	DefaultTelemetrySampleRatio = 0.1
)

// MinShutdownTimeout is the shortest way out a service may be given: the close
// after the drain has a quarter of it, and a quarter of less is no time to
// deliver anything in.
const MinShutdownTimeout = time.Second

// ErrInvalid is returned by Load and Validate; callers branch on it with
// errors.Is.
var ErrInvalid = errors.New("config: invalid")

// Config is the whole configuration of the process. Every field names the
// variable it comes from, so that a refusal and the deployment that caused it
// use the same word.
type Config struct {
	// Port is the TCP port to listen on.
	Port string `mapstructure:"PORT"`
	// PublicURL is the service's own address: the OAuth issuer, the canonical
	// resource of the MCP endpoint and the only Host served.
	PublicURL string `mapstructure:"MATHTRAIL_PUBLIC_URL"`
	// LogLevel is a zap level name: debug, info, warn or error.
	LogLevel string `mapstructure:"MATHTRAIL_LOG_LEVEL"`
	// LogFormat is "json" for a log collector and "console" to read by eye.
	LogFormat string `mapstructure:"MATHTRAIL_LOG_FORMAT"`

	ReadHeaderTimeout time.Duration `mapstructure:"MATHTRAIL_HTTP_READ_HEADER_TIMEOUT"`
	ReadTimeout       time.Duration `mapstructure:"MATHTRAIL_HTTP_READ_TIMEOUT"`
	WriteTimeout      time.Duration `mapstructure:"MATHTRAIL_HTTP_WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `mapstructure:"MATHTRAIL_HTTP_IDLE_TIMEOUT"`
	// ShutdownTimeout is the whole way out once a stop is asked for: the
	// server's drain and the close after it together. The platform stops the
	// process ten seconds after asking it to, and time spent past that is not
	// spent at all, so the default stops a second short of it: the signal takes
	// a moment to arrive, and the last lines are written after the close.
	ShutdownTimeout time.Duration `mapstructure:"MATHTRAIL_SHUTDOWN_TIMEOUT"`

	// SolverSteps is how many instructions one run of a solver may execute
	// before it is stopped.
	SolverSteps uint64 `mapstructure:"MATHTRAIL_SOLVER_STEPS"`
	// SolverTimeout is the wall clock of one run of a solver.
	SolverTimeout time.Duration `mapstructure:"MATHTRAIL_SOLVER_TIMEOUT"`
	// SolverConcurrency is how many solvers may run at once in this process.
	SolverConcurrency int `mapstructure:"MATHTRAIL_SOLVER_CONCURRENCY"`

	// Telemetry is TelemetryAuto, TelemetryOn or TelemetryOff: whether traces
	// and metrics leave the process at all.
	Telemetry string `mapstructure:"MATHTRAIL_TELEMETRY"`
	// TelemetryEndpoint is the root URL of the collector they are sent to.
	TelemetryEndpoint string `mapstructure:"MATHTRAIL_TELEMETRY_ENDPOINT"`
	// TelemetrySampleRatio is the share of traces kept, whatever a request
	// arrives saying about its own.
	TelemetrySampleRatio float64 `mapstructure:"MATHTRAIL_TELEMETRY_SAMPLE_RATIO"`

	// GCPProjectID names the Google Cloud project the telemetry is filed
	// under. The collector learns it from nowhere else.
	GCPProjectID string `mapstructure:"MATHTRAIL_GCP_PROJECT_ID"`

	// SealKeyCurrent is the key everything is sealed and unsealed with,
	// standard base64 of 32 random bytes. It is a secret: it belongs in no log
	// line and in no error message, and only the identifier derived from it may
	// be named anywhere.
	SealKeyCurrent string `mapstructure:"MATHTRAIL_SEAL_KEY_CURRENT"`
	// SealKeyPrevious is the key from before a rotation, which still opens what
	// it sealed. Empty outside a rotation, and a secret like the one above.
	SealKeyPrevious string `mapstructure:"MATHTRAIL_SEAL_KEY_PREVIOUS"`

	// DevAuth replaces the Google sign-in with a stub. It is refused whenever
	// K_SERVICE is set.
	DevAuth bool `mapstructure:"MATHTRAIL_DEV_AUTH"`

	// Service is the name K_SERVICE carries. It is the signal a deployment has
	// and a developer's machine does not, which is what makes it the gate for
	// the dev switches.
	Service string `mapstructure:"K_SERVICE"`
}

// Deployed reports whether the process is running as a deployed service.
func (c *Config) Deployed() bool { return c.Service != "" }

// TelemetryEnabled reports whether traces and metrics leave the process.
func (c *Config) TelemetryEnabled() bool {
	switch c.Telemetry {
	case TelemetryOn:
		return true
	case TelemetryOff:
		return false
	default:
		return c.Deployed()
	}
}

// DrainTimeout is the part of the way out the server waits for the requests
// still in flight: three quarters of it.
//
// The drain and the close after it share one budget rather than take one
// each, in fixed shares, so that a slow request cannot eat the time the
// telemetry's last delivery needs.
func (c *Config) DrainTimeout() time.Duration { return c.ShutdownTimeout - c.CloseTimeout() }

// CloseTimeout is the part of the way out kept for closing what the service ran
// on, once the server has drained or a start has failed halfway: a quarter of
// it.
func (c *Config) CloseTimeout() time.Duration { return c.ShutdownTimeout / 4 }

// Origin is the public URL as a bare scheme and host, without a trailing slash.
func (c *Config) Origin() string { return strings.TrimSuffix(c.PublicURL, "/") }

// Load reads the process environment.
func Load() (*Config, error) { return LoadFrom(os.Environ()) }

// LoadFrom reads the configuration from a set of KEY=VALUE strings, in the
// shape os.Environ returns them, applies the defaults and validates the
// result. The source is an argument so that a test can hand over an
// environment instead of changing the one the process runs in.
//
// A variable set to an empty string counts as unset: an empty value is what a
// deployment leaves behind when it stops setting something, and it should not
// silently override a default with nothing.
func LoadFrom(environ []string) (*Config, error) {
	v := viper.New()

	// Every key is declared, including the ones whose default is empty, so
	// that the names, the defaults and the types sit in one place.
	v.SetDefault("PORT", DefaultPort)
	v.SetDefault("MATHTRAIL_PUBLIC_URL", "")
	v.SetDefault("MATHTRAIL_LOG_LEVEL", DefaultLogLevel)
	v.SetDefault("MATHTRAIL_LOG_FORMAT", DefaultLogFormat)
	v.SetDefault("MATHTRAIL_HTTP_READ_HEADER_TIMEOUT", DefaultReadHeaderTimeout)
	v.SetDefault("MATHTRAIL_HTTP_READ_TIMEOUT", DefaultReadTimeout)
	v.SetDefault("MATHTRAIL_HTTP_WRITE_TIMEOUT", DefaultWriteTimeout)
	v.SetDefault("MATHTRAIL_HTTP_IDLE_TIMEOUT", DefaultIdleTimeout)
	v.SetDefault("MATHTRAIL_SHUTDOWN_TIMEOUT", DefaultShutdownTimeout)
	v.SetDefault("MATHTRAIL_SOLVER_STEPS", DefaultSolverSteps)
	v.SetDefault("MATHTRAIL_SOLVER_TIMEOUT", DefaultSolverTimeout)
	v.SetDefault("MATHTRAIL_SOLVER_CONCURRENCY", DefaultSolverConcurrency)
	v.SetDefault("MATHTRAIL_TELEMETRY", DefaultTelemetry)
	v.SetDefault("MATHTRAIL_TELEMETRY_ENDPOINT", DefaultTelemetryEndpoint)
	v.SetDefault("MATHTRAIL_TELEMETRY_SAMPLE_RATIO", DefaultTelemetrySampleRatio)
	v.SetDefault("MATHTRAIL_GCP_PROJECT_ID", "")
	v.SetDefault("MATHTRAIL_SEAL_KEY_CURRENT", "")
	v.SetDefault("MATHTRAIL_SEAL_KEY_PREVIOUS", "")
	v.SetDefault("MATHTRAIL_DEV_AUTH", false)
	v.SetDefault("K_SERVICE", "")

	// Only the declared keys are taken from the environment. Everything else a
	// process carries — credentials of other systems, tokens, the shell's own
	// variables — has no business inside this reader, where a future dump of
	// its contents would carry them straight into a log.
	declared := make(map[string]struct{}, len(v.AllKeys()))
	for _, key := range v.AllKeys() {
		declared[key] = struct{}{}
	}
	for _, entry := range environ {
		name, value, found := strings.Cut(entry, "=")
		if !found || value == "" {
			continue
		}
		if _, ours := declared[strings.ToLower(name)]; !ours {
			continue
		}
		v.Set(name, value)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		// The decoder names every variable it could not read, and it names all
		// of them rather than the first, so its message is passed through as
		// it is. A test holds it to that.
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	// Off a deployment the public address is almost always the loopback one,
	// and making every developer set it would be ceremony. In a deployment it
	// cannot be guessed, and guessing it wrong would issue tokens under the
	// wrong issuer, so there it is required.
	if cfg.PublicURL == "" && !cfg.Deployed() {
		cfg.PublicURL = "http://localhost:" + cfg.Port
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate reports the first problem it finds, naming the variable behind it.
func (c *Config) Validate() error {
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("%w: PORT must be a number from 1 to 65535, got %q", ErrInvalid, c.Port)
	}

	if err := c.validatePublicURL(); err != nil {
		return err
	}

	if err := c.validateSealKeys(); err != nil {
		return err
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("%w: MATHTRAIL_LOG_LEVEL must be debug, info, warn or error, got %q", ErrInvalid, c.LogLevel)
	}

	switch c.LogFormat {
	case "json", "console":
	default:
		return fmt.Errorf("%w: MATHTRAIL_LOG_FORMAT must be json or console, got %q", ErrInvalid, c.LogFormat)
	}
	// The console format is for a person at a terminal. It escapes nothing, so a
	// newline a request carries would start a line of its own, and a collector
	// reads no severity out of it: a deployment writes JSON.
	if c.LogFormat == "console" && c.Deployed() {
		return fmt.Errorf("%w: MATHTRAIL_LOG_FORMAT must be json when K_SERVICE is set", ErrInvalid)
	}

	if err := c.validateTimeouts(); err != nil {
		return err
	}

	if err := c.validateSolver(); err != nil {
		return err
	}

	if err := c.validateTelemetry(); err != nil {
		return err
	}

	// The one switch that trades safety for convenience, and the one place it
	// is stopped: in a deployment it is refused outright, not warned about.
	if c.DevAuth && c.Deployed() {
		return fmt.Errorf("%w: MATHTRAIL_DEV_AUTH must not be set when K_SERVICE is set", ErrInvalid)
	}

	return nil
}

// validatePublicURL refuses an address the service could not be itself at. The
// issuer, the redirect URI and the canonical resource are all built from it, so
// anything it carries beyond a scheme and a host ends up inside a token that
// somebody else has to match exactly.
func (c *Config) validatePublicURL() error {
	if c.PublicURL == "" {
		return fmt.Errorf("%w: MATHTRAIL_PUBLIC_URL must be set when K_SERVICE is set", ErrInvalid)
	}
	parsed, err := url.Parse(c.PublicURL)
	switch {
	case err != nil:
		return fmt.Errorf("%w: MATHTRAIL_PUBLIC_URL is not a URL: %q", ErrInvalid, c.PublicURL)
	case parsed.Host == "":
		return fmt.Errorf("%w: MATHTRAIL_PUBLIC_URL has no host: %q", ErrInvalid, c.PublicURL)
	case parsed.Scheme != "https" && !isLoopback(parsed.Hostname()):
		// The issuer, the redirect URI and the canonical resource must be
		// https everywhere except a developer's own machine.
		return fmt.Errorf("%w: MATHTRAIL_PUBLIC_URL must use https outside localhost: %q", ErrInvalid, c.PublicURL)
	case parsed.Path != "" && parsed.Path != "/":
		return fmt.Errorf("%w: MATHTRAIL_PUBLIC_URL must have no path: %q", ErrInvalid, c.PublicURL)
	case parsed.RawQuery != "" || parsed.Fragment != "":
		return fmt.Errorf("%w: MATHTRAIL_PUBLIC_URL must have no query and no fragment: %q", ErrInvalid, c.PublicURL)
	case parsed.User != nil:
		return fmt.Errorf("%w: MATHTRAIL_PUBLIC_URL must carry no credentials: %q", ErrInvalid, c.PublicURL)
	}

	return nil
}

// validateSolver refuses limits nothing could run under: a solver with no
// budget would either run for ever or not at all, and a sandbox with no slots
// would hold every run it was given for ever.
func (c *Config) validateSolver() error {
	if c.SolverSteps == 0 {
		return fmt.Errorf("%w: MATHTRAIL_SOLVER_STEPS must be at least 1", ErrInvalid)
	}
	if c.SolverConcurrency < 1 {
		return fmt.Errorf("%w: MATHTRAIL_SOLVER_CONCURRENCY must be at least 1, got %d",
			ErrInvalid, c.SolverConcurrency)
	}
	return nil
}

// validateSealKeys refuses a key the service could not seal with, naming the
// variable that carries it. The key itself never reaches the message: what it
// says is the shape that is wrong, never the value that is wrong.
func (c *Config) validateSealKeys() error {
	// The one secret with no sensible default: a service given no key could
	// only invent one, and every token it issued would die with the process. A
	// key is read with the whitespace around it ignored, since a secret read out
	// of a file arrives with a newline on the end, and a value of whitespace
	// alone is no key.
	if strings.TrimSpace(c.SealKeyCurrent) == "" {
		return fmt.Errorf(
			"%w: MATHTRAIL_SEAL_KEY_CURRENT must be set to %d random bytes in standard base64",
			ErrInvalid, seal.KeySize)
	}

	// What makes the two a ring the service can seal with — each a key, and not
	// the same key twice — is the ring's own to say, and it says which of the
	// two it refused.
	if _, err := seal.NewKeyRing(c.SealKeyCurrent, c.SealKeyPrevious); err != nil {
		variable := "MATHTRAIL_SEAL_KEY_CURRENT"
		if errors.Is(err, seal.ErrPreviousKey) {
			variable = "MATHTRAIL_SEAL_KEY_PREVIOUS"
		}
		return fmt.Errorf("%w: %s: %w", ErrInvalid, variable, err)
	}
	return nil
}

// isLoopback reports whether a host can only mean this machine. The whole of
// 127.0.0.0/8 is loopback, not just its first address, and every name under
// .localhost is reserved for it. A name is read as the name system reads it:
// without regard to case, and the same with or without the dot that ends a
// fully qualified one.
func isLoopback(host string) bool {
	name := strings.TrimSuffix(strings.ToLower(host), ".")
	if name == "localhost" || strings.HasSuffix(name, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// validateTimeouts refuses a duration nothing could run under, naming its
// variable: every one has to be positive, and the way out long enough to be
// shared.
func (c *Config) validateTimeouts() error {
	for _, timeout := range []struct {
		name  string
		value time.Duration
	}{
		{"MATHTRAIL_HTTP_READ_HEADER_TIMEOUT", c.ReadHeaderTimeout},
		{"MATHTRAIL_HTTP_READ_TIMEOUT", c.ReadTimeout},
		{"MATHTRAIL_HTTP_WRITE_TIMEOUT", c.WriteTimeout},
		{"MATHTRAIL_HTTP_IDLE_TIMEOUT", c.IdleTimeout},
		{"MATHTRAIL_SHUTDOWN_TIMEOUT", c.ShutdownTimeout},
		{"MATHTRAIL_SOLVER_TIMEOUT", c.SolverTimeout},
	} {
		if timeout.value <= 0 {
			return fmt.Errorf("%w: %s must be a positive duration, got %v", ErrInvalid, timeout.name, timeout.value)
		}
	}
	if c.ShutdownTimeout < MinShutdownTimeout {
		return fmt.Errorf("%w: MATHTRAIL_SHUTDOWN_TIMEOUT must be at least %v, got %v",
			ErrInvalid, MinShutdownTimeout, c.ShutdownTimeout)
	}
	return nil
}

// validateTelemetry refuses a value that is not one of the settings, naming
// the variable behind it. Whether anything can actually be exported is a
// question for the deployment, and it is answered by exporting nothing rather
// than by refusing to start.
func (c *Config) validateTelemetry() error {
	switch c.Telemetry {
	case TelemetryAuto, TelemetryOn, TelemetryOff:
	default:
		return fmt.Errorf("%w: MATHTRAIL_TELEMETRY must be %s, %s or %s, got %q",
			ErrInvalid, TelemetryAuto, TelemetryOn, TelemetryOff, c.Telemetry)
	}

	if c.TelemetrySampleRatio < 0 || c.TelemetrySampleRatio > 1 {
		return fmt.Errorf("%w: MATHTRAIL_TELEMETRY_SAMPLE_RATIO must be a share from 0 to 1, got %v",
			ErrInvalid, c.TelemetrySampleRatio)
	}

	if err := c.validateTelemetryEndpoint(); err != nil {
		return err
	}
	return nil
}

// validateTelemetryEndpoint refuses an address the exporter could not post to,
// by the exporter's own rule for one, and an address that would carry the
// exporter's credential in the clear.
func (c *Config) validateTelemetryEndpoint() error {
	root, err := collector.Root(c.TelemetryEndpoint)
	if err != nil {
		return fmt.Errorf("%w: MATHTRAIL_TELEMETRY_ENDPOINT: %w", ErrInvalid, err)
	}
	// Telemetry carries no secret, but it does carry a credential in every
	// request, and a credential travels over https or not at all.
	if root.Scheme != "https" && !isLoopback(root.Hostname()) {
		return fmt.Errorf("%w: MATHTRAIL_TELEMETRY_ENDPOINT must use https outside localhost: %q",
			ErrInvalid, c.TelemetryEndpoint)
	}
	return nil
}
