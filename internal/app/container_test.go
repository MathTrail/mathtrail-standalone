package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/app"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/geoip"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/geoip/geoiptest"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/learner"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
	oauthserver "github.com/MathTrail/mathtrail-standalone/internal/transport/oauth"
)

// A setting no part of the service could be built with stops the container
// where that part is built, with that part's own refusal, and no container is
// handed back: the process stops as it starts rather than at the first child
// it would fail. Each case spoils one setting of a configuration that builds.
func TestContainerRefusesWhatItCannotBuild(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		spoil func(t *testing.T, cfg *config.Config)
		// is is the refusal the part that fails wraps; says is what it says,
		// for a part that wraps no refusal of its own.
		is   error
		says string
	}{
		{name: "a solver that may take no step", spoil: func(_ *testing.T, cfg *config.Config) { cfg.SolverSteps = 0 },
			says: "starlark: steps must be at least one"},
		{name: "a solver given no time", spoil: func(_ *testing.T, cfg *config.Config) { cfg.SolverTimeout = 0 },
			says: "starlark: timeout must be a positive duration"},
		{name: "no solver at a time", spoil: func(_ *testing.T, cfg *config.Config) { cfg.SolverConcurrency = 0 },
			says: "starlark: concurrency must be at least one"},
		{name: "no wait for a solver's turn", spoil: func(_ *testing.T, cfg *config.Config) { cfg.SolverWait = 0 },
			says: "starlark: wait must be a positive duration"},
		{name: "calls to Drive given no time", spoil: func(_ *testing.T, cfg *config.Config) { cfg.DriveTimeout = 0 },
			is: drive.ErrSettings},
		{name: "a Google client with no secret", spoil: func(_ *testing.T, cfg *config.Config) {
			cfg.GoogleClientID, cfg.GoogleClientSecret = "mathtrail.apps.googleusercontent.com", ""
		}, is: googleauth.ErrSettings},
		{name: "a reviewers' grant that is not JSON", spoil: func(_ *testing.T, cfg *config.Config) {
			cfg.ReviewerPassword, cfg.ReviewerGrant = "a-password-for-the-reviewers", "not a grant"
		}, is: config.ErrInvalid},
		{name: "a learner key that is no key", spoil: func(_ *testing.T, cfg *config.Config) { cfg.LearnerKey = "not a key" },
			is: learner.ErrKey},
		{name: "a database of countries that is not there", spoil: func(t *testing.T, cfg *config.Config) {
			cfg.CountryDB = filepath.Join(t.TempDir(), "missing")
		}, is: geoip.ErrDatabase},
		{name: "a public address with a path", spoil: func(_ *testing.T, cfg *config.Config) { cfg.PublicURL = "http://localhost/mcp" },
			is: oauthserver.ErrSettings},
		{name: "no window to wait for a task in", spoil: func(_ *testing.T, cfg *config.Config) { cfg.RequestWindow = 0 },
			is: mcpserver.ErrSettings},
		{name: "a day with room for no task", spoil: func(_ *testing.T, cfg *config.Config) { cfg.DailyTasks = 0 },
			is: mcpserver.ErrSettings},
		{name: "a day with room for no failure", spoil: func(_ *testing.T, cfg *config.Config) { cfg.DailyFailed = 0 },
			is: mcpserver.ErrSettings},
		{name: "a mistake that never counts as repeated", spoil: func(_ *testing.T, cfg *config.Config) { cfg.TrapRepeats = 0 },
			is: mcpserver.ErrSettings},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := testConfig()
			tc.spoil(t, cfg)
			container, err := app.NewContainer(t.Context(), cfg, zaptest.NewLogger(t))
			if container != nil {
				container.Close(context.Background())
				t.Error("NewContainer() returned a container alongside its refusal, want nothing")
			}
			switch {
			case tc.is != nil && !errors.Is(err, tc.is):
				t.Errorf("NewContainer() error = %v, want %v", err, tc.is)
			case tc.says != "" && (err == nil || !strings.Contains(err.Error(), tc.says)):
				t.Errorf("NewContainer() error = %v, want one that says %q", err, tc.says)
			}
		})
	}
}

// A container whose construction fails halfway lets go of what it had opened
// by then — here the database of countries, opened before the tools of the
// lesson refused a window of none — so that a container that was never handed
// back leaves nothing open behind it.
func TestAContainerThatFailsHalfwayLetsGoOfTheCountryDatabase(t *testing.T) {
	t.Parallel()

	database := geoiptest.Write(t)
	wantMappingSeen(t, database)

	cfg := testConfig()
	cfg.CountryDB, cfg.RequestWindow = database, 0
	core, logs := observer.New(zapcore.InfoLevel)
	container, err := app.NewContainer(t.Context(), cfg, zap.New(core))
	if !errors.Is(err, mcpserver.ErrSettings) || container != nil {
		t.Fatalf("NewContainer() = %v, %v, want no container and the window refused", container, err)
	}
	// The container says so as it opens the database, so a construction that
	// stopped before it ever opened one cannot pass for one that let go of it.
	if opened := logs.FilterMessage("country database").FilterField(zap.Bool("configured", true)).Len(); opened != 1 {
		t.Fatalf("lines of the country database opened = %d, want 1: the case needs it opened before the window is refused", opened)
	}
	if mapped(t, database) {
		t.Error("the database of countries is still mapped once the container failed, want it let go of")
	}
}

// wantMappingSeen holds the platform to showing a database while it is open
// and no longer once it is closed, so that a case can tell from the mappings
// alone whether something let go of one. Where it shows neither — another
// system, or a database read into memory rather than mapped — the case cannot
// be told and is skipped.
func wantMappingSeen(t *testing.T, path string) {
	t.Helper()

	opened, err := geoip.Open(path)
	if err != nil {
		t.Fatalf("geoip.Open() error = %v, want the database open", err)
	}
	seen := mapped(t, path)
	if err := opened.Close(); err != nil {
		t.Fatalf("Close() error = %v, want the database closed", err)
	}
	if !seen {
		t.Skip("the database is not mapped into the process here, so letting go of it cannot be seen")
	}
	if mapped(t, path) {
		t.Fatal("the database is still mapped once closed, so letting go of it cannot be told apart")
	}
}

// mapped reports whether the file at path is mapped into the process, as the
// kernel lists the mappings; a system that lists none skips the case.
func mapped(t *testing.T, path string) bool {
	t.Helper()

	listed, err := os.ReadFile("/proc/self/maps")
	if err != nil {
		t.Skipf("the mappings of the process cannot be read here: %v", err)
	}
	return strings.Contains(string(listed), path)
}
