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
	DefaultShutdownTimeout   = 10 * time.Second

	DefaultSolverSteps       = 10_000_000
	DefaultSolverTimeout     = 2 * time.Second
	DefaultSolverConcurrency = 4
)

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
	ShutdownTimeout   time.Duration `mapstructure:"MATHTRAIL_SHUTDOWN_TIMEOUT"`

	// SolverSteps is how many instructions one run of a solver may execute
	// before it is stopped.
	SolverSteps uint64 `mapstructure:"MATHTRAIL_SOLVER_STEPS"`
	// SolverTimeout is the wall clock of one run of a solver.
	SolverTimeout time.Duration `mapstructure:"MATHTRAIL_SOLVER_TIMEOUT"`
	// SolverConcurrency is how many solvers may run at once in this process.
	SolverConcurrency int `mapstructure:"MATHTRAIL_SOLVER_CONCURRENCY"`

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

	if err := c.validateSolver(); err != nil {
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
	// only invent one, and every token it issued would die with the process.
	if c.SealKeyCurrent == "" {
		return fmt.Errorf(
			"%w: MATHTRAIL_SEAL_KEY_CURRENT must be set to %d random bytes in standard base64",
			ErrInvalid, seal.KeySize)
	}

	current, err := seal.ParseKey(c.SealKeyCurrent)
	if err != nil {
		return fmt.Errorf("%w: MATHTRAIL_SEAL_KEY_CURRENT: %w", ErrInvalid, err)
	}

	// Outside a rotation there is no previous key, and that is the normal shape.
	if c.SealKeyPrevious == "" {
		return nil
	}
	previous, err := seal.ParseKey(c.SealKeyPrevious)
	if err != nil {
		return fmt.Errorf("%w: MATHTRAIL_SEAL_KEY_PREVIOUS: %w", ErrInvalid, err)
	}

	// Both variables holding one key is a rotation that only looks done: what
	// the retired key sealed would stop opening the moment it is deployed.
	if previous.ID() == current.ID() {
		return fmt.Errorf(
			"%w: MATHTRAIL_SEAL_KEY_PREVIOUS holds the same key as MATHTRAIL_SEAL_KEY_CURRENT; a rotation needs two",
			ErrInvalid)
	}
	return nil
}

// isLoopback reports whether a host can only mean this machine. The whole of
// 127.0.0.0/8 is loopback, not just its first address, and every name under
// .localhost is reserved for it.
func isLoopback(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
