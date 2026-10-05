// Package config reads the service's configuration from the environment.
//
// Environment variables are the only source: there is no file, no flag and no
// remote source. They are read here and nowhere else, so that every value has
// one origin and one validation. Viper does the reading — as it does in the
// platform's other services — which keeps the names, the defaults and the
// types of a growing table in one declaration.
//
// Only the variables the service actually uses are here. A variable arrives
// with the code that consumes it, because a required variable with no reader
// is a deployment that fails for no reason.
package config

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/idna"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/learner"
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
	DefaultSolverConcurrency = 1
	DefaultSolverWait        = 3 * time.Second

	DefaultRequestWindow = 15 * time.Minute

	DefaultRateUserPerMin     = 60
	DefaultRateIPPerMin       = 20
	DefaultRateInstancePerMin = 200
	DefaultRateRenewalPerMin  = 12
	DefaultDailyTasks         = 20
	DefaultDailyFailed        = 5

	DefaultTrapRepeats = 2

	DefaultDriveTimeout = 10 * time.Second

	DefaultSiteURL = "https://mathtrail.app"

	DefaultTelemetry            = TelemetryAuto
	DefaultTelemetryEndpoint    = "https://telemetry.googleapis.com"
	DefaultTelemetrySampleRatio = 0.1
)

// MinShutdownTimeout is the shortest way out a service may be given: the close
// after the drain has a quarter of it, and a quarter of less is no time to
// deliver anything in.
const MinShutdownTimeout = time.Second

// MinRequestWindow is the shortest a task being written may be waited for: a
// model takes about a minute to write one, and a window shorter than that
// would turn every task it hands in into one for a request that is over.
const MinRequestWindow = time.Minute

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
	// SolverConcurrency is how many solvers may run at once in this process:
	// one for every processor the instance has, since the clock of a run is
	// wall time and a run sharing a processor would spend it on another's
	// work. It is not worked out from the processors the runtime sees, which
	// never counts fewer than two on a machine that has two.
	SolverConcurrency int `mapstructure:"MATHTRAIL_SOLVER_CONCURRENCY"`
	// SolverWait is how long a run of a solver waits for one of those places to
	// come free. A task whose run finds none in that time is not checked, and
	// spends no attempt.
	SolverWait time.Duration `mapstructure:"MATHTRAIL_SOLVER_WAIT"`

	// RequestWindow is how long a task being written is waited for. A request
	// older than this counts as abandoned: the next ask opens a new one, and a
	// task handed in for it is refused as one for a request that is over.
	RequestWindow time.Duration `mapstructure:"MATHTRAIL_REQUEST_WINDOW"`

	// RateUserPerMin is how many requests a signed-in account may send the
	// MCP endpoint of one instance in a minute.
	RateUserPerMin int `mapstructure:"MATHTRAIL_RATE_USER_PER_MIN"`
	// RateIPPerMin is how many requests one address may send the sign-in's
	// endpoints and documents of one instance in a minute.
	RateIPPerMin int `mapstructure:"MATHTRAIL_RATE_IP_PER_MIN"`
	// RateInstancePerMin is how many requests one instance takes in a minute
	// from everybody together: a fuse against something gone wrong, not a
	// share of anybody's.
	RateInstancePerMin int `mapstructure:"MATHTRAIL_RATE_INSTANCE_PER_MIN"`
	// RateRenewalPerMin is how many times in a minute one account's grant may
	// be renewed at Google through one instance. A grant needs one renewal
	// about every hour; the pace is what keeps an old refresh token, used over
	// and over, from having the service call Google every time.
	RateRenewalPerMin int `mapstructure:"MATHTRAIL_RATE_RENEWAL_PER_MIN"`
	// DailyTasks is how many tasks a child may be given in a day.
	DailyTasks int `mapstructure:"MATHTRAIL_DAILY_TASKS"`
	// DailyFailed is how many requests of a day may end with the model out of
	// attempts before no more are opened that day.
	DailyFailed int `mapstructure:"MATHTRAIL_DAILY_FAILED"`

	// TrapRepeats is how many of the latest answers a trap has to be behind to
	// count as a mistake that repeats: the map of misconceptions lists it, and
	// the answer that falls for it again says so.
	TrapRepeats int `mapstructure:"MATHTRAIL_TRAP_REPEATS"`

	// DriveTimeout is how long one call to a parent's Drive may take.
	DriveTimeout time.Duration `mapstructure:"MATHTRAIL_DRIVE_TIMEOUT"`

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

	// GoogleClientID is the service's OAuth client at Google, which a parent
	// signs in with. Required in a deployment; a developer's machine may go
	// without, and a sign-in there stops at a page that says so.
	GoogleClientID string `mapstructure:"MATHTRAIL_GOOGLE_CLIENT_ID"`
	// GoogleClientSecret is that client's secret. It belongs in no log line
	// and in no error message.
	GoogleClientSecret string `mapstructure:"MATHTRAIL_GOOGLE_CLIENT_SECRET"`

	// LearnerKey is the secret the name a child is counted under in the log is
	// derived from, standard base64 of 32 random bytes. It is a secret of its
	// own, never one of the sealing keys, and it is not rotated with them: a
	// name that changed in the middle of a month would count one child twice.
	// Required in a deployment; a developer's machine may go without, and its
	// process then makes one that lasts as long as it does.
	LearnerKey string `mapstructure:"MATHTRAIL_LEARNER_KEY"`

	// CountryDB is the file of the database the country a parent signs in from
	// is looked up in. Required in a deployment, whose image carries the file;
	// a developer's machine may go without, and then no country is known.
	CountryDB string `mapstructure:"MATHTRAIL_COUNTRY_DB"`

	// SiteURL is the site's address, where the consent screen links the terms
	// and the privacy policy.
	SiteURL string `mapstructure:"MATHTRAIL_SITE_URL"`

	// OpenAIChallenge is the token ChatGPT's plugin directory gives to prove
	// the service's domain is the plugin's, served as it is at the address the
	// directory reads it from. Empty, that address is not served.
	OpenAIChallenge string `mapstructure:"MATHTRAIL_OPENAI_CHALLENGE"`

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

// Site is the site's origin as a browser writes it: its scheme and its host in
// lower case, a name spelled in ASCII as IDNA spells it, and no port the
// scheme implies nor a trailing slash. The consent screen, the progress and
// the card that holds an address to the site's origin then name it alike.
func (c *Config) Site() string {
	parsed, err := url.Parse(c.SiteURL)
	if err != nil {
		return strings.TrimSuffix(c.SiteURL, "/")
	}
	scheme, host := strings.ToLower(parsed.Scheme), asciiHost(parsed.Hostname())
	if port := parsed.Port(); port != "" && port != defaultPorts[scheme] {
		return scheme + "://" + net.JoinHostPort(host, port)
	}
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return scheme + "://" + host
}

// defaultPorts are the ports a browser leaves out of an origin, by scheme.
var defaultPorts = map[string]string{"https": "443", "http": "80"}

// browserNames is IDNA as a browser reads the host of an address: mapped as
// for a lookup, the rules of mixed directions held, and none of the checks of
// hyphens and of letters a registry makes, which no browser makes.
var browserNames = idna.New(
	idna.MapForLookup(), idna.BidiRule(), idna.CheckHyphens(false), idna.StrictDomainName(false), idna.Transitional(false),
)

// asciiHost is a host as a browser writes it in an origin: a name spelled as
// IDNA spells it, which writes it in lower case, and the address of a machine
// in lower case.
func asciiHost(host string) string {
	if net.ParseIP(host) == nil {
		if ascii, err := browserNames.ToASCII(host); err == nil {
			return ascii
		}
	}
	return strings.ToLower(host)
}

// GoogleSignIn reports whether a Google client is configured, which is what a
// parent signs in with.
func (c *Config) GoogleSignIn() bool { return c.GoogleClientID != "" }

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
	v.SetDefault("MATHTRAIL_SOLVER_WAIT", DefaultSolverWait)
	v.SetDefault("MATHTRAIL_REQUEST_WINDOW", DefaultRequestWindow)
	v.SetDefault("MATHTRAIL_RATE_USER_PER_MIN", DefaultRateUserPerMin)
	v.SetDefault("MATHTRAIL_RATE_IP_PER_MIN", DefaultRateIPPerMin)
	v.SetDefault("MATHTRAIL_RATE_INSTANCE_PER_MIN", DefaultRateInstancePerMin)
	v.SetDefault("MATHTRAIL_RATE_RENEWAL_PER_MIN", DefaultRateRenewalPerMin)
	v.SetDefault("MATHTRAIL_DAILY_TASKS", DefaultDailyTasks)
	v.SetDefault("MATHTRAIL_DAILY_FAILED", DefaultDailyFailed)
	v.SetDefault("MATHTRAIL_TRAP_REPEATS", DefaultTrapRepeats)
	v.SetDefault("MATHTRAIL_DRIVE_TIMEOUT", DefaultDriveTimeout)
	v.SetDefault("MATHTRAIL_TELEMETRY", DefaultTelemetry)
	v.SetDefault("MATHTRAIL_TELEMETRY_ENDPOINT", DefaultTelemetryEndpoint)
	v.SetDefault("MATHTRAIL_TELEMETRY_SAMPLE_RATIO", DefaultTelemetrySampleRatio)
	v.SetDefault("MATHTRAIL_GCP_PROJECT_ID", "")
	v.SetDefault("MATHTRAIL_SEAL_KEY_CURRENT", "")
	v.SetDefault("MATHTRAIL_SEAL_KEY_PREVIOUS", "")
	v.SetDefault("MATHTRAIL_GOOGLE_CLIENT_ID", "")
	v.SetDefault("MATHTRAIL_GOOGLE_CLIENT_SECRET", "")
	v.SetDefault("MATHTRAIL_LEARNER_KEY", "")
	v.SetDefault("MATHTRAIL_COUNTRY_DB", "")
	v.SetDefault("MATHTRAIL_SITE_URL", DefaultSiteURL)
	v.SetDefault("MATHTRAIL_OPENAI_CHALLENGE", "")
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

	// A secret read out of a file arrives with a newline on the end, and Google
	// would refuse a client whose name or secret carried one.
	cfg.GoogleClientID = strings.TrimSpace(cfg.GoogleClientID)
	cfg.GoogleClientSecret = strings.TrimSpace(cfg.GoogleClientSecret)
	cfg.OpenAIChallenge = strings.TrimSpace(cfg.OpenAIChallenge)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate reports the first problem it finds, naming the variable behind it.
func (c *Config) Validate() error {
	for _, check := range []func() error{
		c.validatePort,
		c.validatePublicURL,
		c.validateSealKeys,
		c.validateLogging,
		c.validateTimeouts,
		c.validateSolver,
		c.validateLimits,
		c.validateTrapRepeats,
		c.validateTelemetry,
		c.validateDevAuth,
		c.validateGoogle,
		c.validateSiteURL,
		c.validateLearnerKey,
		c.validateCountryDB,
		c.validateOpenAIChallenge,
	} {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

// validatePort refuses a port the server could not listen on.
func (c *Config) validatePort() error {
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("%w: PORT must be a number from 1 to 65535, got %q", ErrInvalid, c.Port)
	}
	return nil
}

// validateDevAuth refuses the development sign-in in a deployment. It is the
// one switch that trades safety for convenience, and the one place it is
// stopped: in a deployment it is refused outright, not warned about.
func (c *Config) validateDevAuth() error {
	if c.DevAuth && c.Deployed() {
		return fmt.Errorf("%w: MATHTRAIL_DEV_AUTH must not be set when K_SERVICE is set", ErrInvalid)
	}
	return nil
}

// validateLearnerKey refuses a secret the children could not be counted
// under, naming the variable and never the value. A deployment needs one: a
// key each instance made for itself would give one child a name on every
// instance it reached. It has to be a secret of its own, because the same
// bytes as a sealing key would mean a name that changed when that key was
// rotated.
func (c *Config) validateLearnerKey() error {
	if c.LearnerKey == "" {
		if c.Deployed() {
			return fmt.Errorf("%w: MATHTRAIL_LEARNER_KEY must be set when K_SERVICE is set, to %d random bytes in standard base64",
				ErrInvalid, learner.KeySize)
		}
		return nil
	}
	if _, err := learner.NewKey(c.LearnerKey); err != nil {
		return fmt.Errorf("%w: MATHTRAIL_LEARNER_KEY: %w", ErrInvalid, err)
	}
	if sameSecret(c.LearnerKey, c.SealKeyCurrent) || sameSecret(c.LearnerKey, c.SealKeyPrevious) {
		return fmt.Errorf("%w: MATHTRAIL_LEARNER_KEY must be a secret of its own, not one of the sealing keys", ErrInvalid)
	}
	return nil
}

// sameSecret reports whether two secrets in standard base64 are the same
// bytes, whatever whitespace each was read with. A value that is no base64 is
// the same as nothing.
func sameSecret(one, other string) bool {
	oneBytes, oneErr := base64.StdEncoding.Strict().DecodeString(strings.TrimSpace(one))
	otherBytes, otherErr := base64.StdEncoding.Strict().DecodeString(strings.TrimSpace(other))
	return oneErr == nil && otherErr == nil && len(oneBytes) > 0 && bytes.Equal(oneBytes, otherBytes)
}

// validateCountryDB refuses a deployment that could not tell where a parent
// signs in from. Whether the file opens is the service's to find out as it
// starts; off a deployment it may be left out, and no country is known.
func (c *Config) validateCountryDB() error {
	if c.Deployed() && c.CountryDB == "" {
		return fmt.Errorf("%w: MATHTRAIL_COUNTRY_DB must be set when K_SERVICE is set", ErrInvalid)
	}
	return nil
}

// maxOpenAIChallenge bounds the token the service serves, so that a value
// pasted by mistake — a whole file, say — is refused rather than served.
const maxOpenAIChallenge = 512

// validateOpenAIChallenge refuses a value that cannot be the token, without
// repeating it: the token is served as it is, so anything but visible ASCII —
// a space or a line break inside it — would be served too, and found only
// when the directory read something else than it gave.
func (c *Config) validateOpenAIChallenge() error {
	visible := !strings.ContainsFunc(c.OpenAIChallenge, func(r rune) bool { return r <= ' ' || r > '~' })
	if len(c.OpenAIChallenge) > maxOpenAIChallenge || !visible {
		return fmt.Errorf("%w: MATHTRAIL_OPENAI_CHALLENGE must be at most %d characters of visible ASCII", ErrInvalid, maxOpenAIChallenge)
	}
	return nil
}

// validateLogging refuses a level or a format the logger does not have, and
// the format for a terminal where a collector reads the log.
func (c *Config) validateLogging() error {
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
	// The issuer, the redirect URI and the canonical resource must be https
	// everywhere except a developer's own machine.
	return validateOrigin("MATHTRAIL_PUBLIC_URL", c.PublicURL)
}

// validateOrigin refuses an address that is not a scheme and a host alone,
// over https — or over plain http on a developer's own machine — naming the
// variable that carries it.
func validateOrigin(variable, address string) error {
	parsed, err := url.Parse(address)
	switch {
	case err != nil:
		return fmt.Errorf("%w: %s is not a URL: %q", ErrInvalid, variable, address)
	case parsed.Host == "":
		return fmt.Errorf("%w: %s has no host: %q", ErrInvalid, variable, address)
	case parsed.Scheme != "https" && !isLoopback(parsed.Hostname()):
		return fmt.Errorf("%w: %s must use https outside localhost: %q", ErrInvalid, variable, address)
	case parsed.Path != "" && parsed.Path != "/":
		return fmt.Errorf("%w: %s must have no path: %q", ErrInvalid, variable, address)
	case strings.ContainsAny(address, "?#"):
		// Even an empty query or fragment: an address with the mark and
		// nothing after it is another string for whoever compares it.
		return fmt.Errorf("%w: %s must have no query and no fragment: %q", ErrInvalid, variable, address)
	case parsed.User != nil:
		return fmt.Errorf("%w: %s must carry no credentials: %q", ErrInvalid, variable, address)
	}
	return nil
}

// validateSiteURL refuses a site the consent screen could not link a parent
// to. A deployment links every parent there, so a site on a developer's own
// machine is refused on one, as plain http is everywhere else.
func (c *Config) validateSiteURL() error {
	if err := validateOrigin("MATHTRAIL_SITE_URL", c.SiteURL); err != nil {
		return err
	}
	if parsed, err := url.Parse(c.SiteURL); err == nil && net.ParseIP(parsed.Hostname()) == nil {
		if _, err := browserNames.ToASCII(parsed.Hostname()); err != nil {
			return fmt.Errorf("%w: MATHTRAIL_SITE_URL has a host IDNA cannot spell, which no browser opens: %q", ErrInvalid, c.SiteURL)
		}
	}
	if c.Deployed() && !strings.HasPrefix(strings.ToLower(c.SiteURL), "https://") {
		return fmt.Errorf("%w: MATHTRAIL_SITE_URL must use https when K_SERVICE is set: %q", ErrInvalid, c.SiteURL)
	}
	return nil
}

// validateGoogle refuses a Google client the service could sign nobody in
// with. A deployment needs one. A developer's machine may go without, and its
// sign-in then stops at a page that says so. Half a client is refused
// everywhere: it is a secret that never arrived, and it would fail at the
// first parent rather than at the start.
func (c *Config) validateGoogle() error {
	client, secret := c.GoogleClientID != "", c.GoogleClientSecret != ""
	switch {
	case c.Deployed() && !client:
		return fmt.Errorf("%w: MATHTRAIL_GOOGLE_CLIENT_ID must be set when K_SERVICE is set", ErrInvalid)
	case client && !secret:
		return fmt.Errorf("%w: MATHTRAIL_GOOGLE_CLIENT_SECRET must be set with MATHTRAIL_GOOGLE_CLIENT_ID", ErrInvalid)
	case secret && !client:
		return fmt.Errorf("%w: MATHTRAIL_GOOGLE_CLIENT_ID must be set with MATHTRAIL_GOOGLE_CLIENT_SECRET", ErrInvalid)
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

// validateLimits refuses a ceiling nobody could get under, naming its
// variable: a pace of no request a minute would refuse every request, and a
// day of no task would refuse every lesson.
func (c *Config) validateLimits() error {
	for _, ceiling := range []struct {
		name  string
		value int
	}{
		{"MATHTRAIL_RATE_USER_PER_MIN", c.RateUserPerMin},
		{"MATHTRAIL_RATE_IP_PER_MIN", c.RateIPPerMin},
		{"MATHTRAIL_RATE_INSTANCE_PER_MIN", c.RateInstancePerMin},
		{"MATHTRAIL_RATE_RENEWAL_PER_MIN", c.RateRenewalPerMin},
		{"MATHTRAIL_DAILY_TASKS", c.DailyTasks},
		{"MATHTRAIL_DAILY_FAILED", c.DailyFailed},
	} {
		if ceiling.value < 1 {
			return fmt.Errorf("%w: %s must be at least 1, got %d", ErrInvalid, ceiling.name, ceiling.value)
		}
	}
	return nil
}

// validateTrapRepeats refuses a threshold no mistake could repeat at: a trap
// met once is a slip, not a mistake that repeats, and one the history window
// cannot hold that many times would never repeat at all.
func (c *Config) validateTrapRepeats() error {
	if c.TrapRepeats < progress.FewestRepeats || c.TrapRepeats > profile.MaxRecent {
		return fmt.Errorf("%w: MATHTRAIL_TRAP_REPEATS must be from %d to %d, got %d",
			ErrInvalid, progress.FewestRepeats, profile.MaxRecent, c.TrapRepeats)
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
		{"MATHTRAIL_SOLVER_WAIT", c.SolverWait},
		{"MATHTRAIL_REQUEST_WINDOW", c.RequestWindow},
		{"MATHTRAIL_DRIVE_TIMEOUT", c.DriveTimeout},
	} {
		if timeout.value <= 0 {
			return fmt.Errorf("%w: %s must be a positive duration, got %v", ErrInvalid, timeout.name, timeout.value)
		}
	}
	if c.ShutdownTimeout < MinShutdownTimeout {
		return fmt.Errorf("%w: MATHTRAIL_SHUTDOWN_TIMEOUT must be at least %v, got %v",
			ErrInvalid, MinShutdownTimeout, c.ShutdownTimeout)
	}
	if c.RequestWindow < MinRequestWindow {
		return fmt.Errorf("%w: MATHTRAIL_REQUEST_WINDOW must be at least %v, got %v",
			ErrInvalid, MinRequestWindow, c.RequestWindow)
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
