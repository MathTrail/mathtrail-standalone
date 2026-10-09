package mcpserver_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark/starlarktest"
	"github.com/MathTrail/mathtrail-standalone/internal/learner"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
	"github.com/MathTrail/mathtrail-standalone/internal/telemetry"
	mcpserver "github.com/MathTrail/mathtrail-standalone/internal/transport/mcp"
)

// The tools of the lesson are served here as the service serves them: over a
// store in memory, to the development account, with the content the binary
// carries, the solvers run in the service's own sandbox and the answers sealed
// with a real key. None of it reaches out of the process, so none of it is
// faked. The fixtures are the prototype's seed profiles: Masha and Sasha are
// in the trial series, Olya is past it.

// shipped is the content the binary carries, read once for every case.
var shipped = sync.OnceValues(content.Load)

// sandbox is the sandbox the solvers of these cases run in: the service's own,
// with the limits a deployment starts with but for the wait for a slot and the
// clock under the race detector, built once. Every case of the package shares
// it, many at a time, and how long a queue of their runs may stand is not what
// any of them is about.
var sandbox = sync.OnceValues(func() (solver.Runner, error) {
	return starlark.New(starlark.Limits{
		Steps:       config.DefaultSolverSteps,
		Timeout:     starlarktest.Clock(config.DefaultSolverTimeout),
		Concurrency: config.DefaultSolverConcurrency,
		Wait:        time.Minute,
	})
})

// lessonDay is when every profile of these cases is written, until a case
// moves its clock on.
var lessonDay = time.Date(2026, 9, 27, 18, 0, 0, 0, time.UTC)

// devAccount is whom the development sign-in signs every call in as.
var devAccount = store.NewAccount(mcpserver.DevAccount, "", time.Time{})

// clock is a clock a case moves on by hand.
type clock struct {
	mu sync.Mutex
	at time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.at
}

func (c *clock) advance(by time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = c.at.Add(by)
}

// lesson serves every tool of the lesson over kept, at a moment that does not
// move, and connects a client.
func lesson(t *testing.T, kept store.Storage) (*harness, *mcp.ClientSession) {
	t.Helper()
	return lessonWith(t, kept, &clock{at: lessonDay}, nil)
}

// lessonWith is lesson on a clock the case moves, and in a sandbox of the
// case's own when it hands one in rather than nil. Whatever runs the solvers
// is observed the way the service observes its sandbox, so that its runs are
// recorded under the call.
func lessonWith(t *testing.T, kept store.Storage, moving *clock, runner solver.Runner) (*harness, *mcp.ClientSession) {
	t.Helper()

	h := newHarness(t)
	service := lessonService(t, h, kept, moving, runner)
	h.start(t, mcpserver.DevSignIn, slices.Concat(service.ProfileTools(), service.TaskTools())...)
	session := h.connect(t, "")
	return h, session
}

// lessonService is the tools of the lesson over kept, recording into the
// harness, on a clock the case moves, and in a sandbox of the case's own when
// it hands one in rather than nil — with whatever else a case changes in
// their parts.
func lessonService(t *testing.T, h *harness, kept store.Storage, moving *clock, runner solver.Runner, changes ...func(*mcpserver.Parts)) *mcpserver.Service {
	t.Helper()

	loaded, err := shipped()
	if err != nil {
		t.Fatalf("load the content: %v", err)
	}
	if runner == nil {
		if runner, err = sandbox(); err != nil {
			t.Fatalf("build the sandbox: %v", err)
		}
	}

	observed, err := telemetry.ObserveSolver(runner, h.traces, h.meters)
	if err != nil {
		t.Fatalf("ObserveSolver() error = %v", err)
	}
	parts := &mcpserver.Parts{
		Store:       kept,
		Content:     loaded,
		Reviewer:    checks.NewReviewer(loaded, observed),
		Sealer:      sealer(t),
		Window:      config.DefaultRequestWindow,
		Daily:       mcpserver.Daily{Tasks: config.DefaultDailyTasks, Failed: config.DefaultDailyFailed},
		TrapRepeats: config.DefaultTrapRepeats,
		Now:         moving.now,
		Version:     "test",
		Logger:      h.log,
		Traces:      h.traces,
		ProjectID:   projectID,
		Learners:    learners(t),
		SiteURL:     config.DefaultSiteURL,
	}
	for _, change := range changes {
		change(parts)
	}
	service, err := mcpserver.NewService(parts)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}

// learners is the key the children of these cases are counted under, the same
// every time, so that a case can work out the name a line has to carry.
func learners(t testing.TB) *learner.Key {
	t.Helper()

	secret := sha256.Sum256([]byte("the counting key of these cases"))
	key, err := learner.NewKey(base64.StdEncoding.EncodeToString(secret[:]))
	if err != nil {
		t.Fatalf("learner.NewKey() error = %v", err)
	}
	return key
}

// sealer is the seal of these cases: a key ring made from a key of their own,
// the same every time, so that a case can open what a tool sealed.
func sealer(t testing.TB) profile.Sealer {
	t.Helper()

	key := sha256.Sum256([]byte("the key of these cases"))
	ring, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(key[:]), "")
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v", err)
	}
	return ring.For(seal.PurposeTaskAnswer)
}

// keptWith is a store holding one fixture as the development account's
// profile.
func keptWith(t *testing.T, student string) store.Storage {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "profiles", student+".json"))
	if err != nil {
		t.Fatalf("read the fixture of %s: %v", student, err)
	}
	p, err := profile.Parse(raw)
	if err != nil {
		t.Fatalf("parse the fixture of %s: %v", student, err)
	}
	return keptAsIs(t, p)
}

// keptAsIs is a store holding this profile as the development account's.
func keptAsIs(t *testing.T, p *profile.Profile) store.Storage {
	t.Helper()

	kept := memory.New()
	if _, err := kept.Create(context.Background(), devAccount, p); err != nil {
		t.Fatalf("keep the profile: %v", err)
	}
	return kept
}

// loadKept is the development account's profile as the store holds it now.
func loadKept(t *testing.T, kept store.Storage) (*profile.Profile, store.Revision) {
	t.Helper()

	p, revision, err := kept.Load(context.Background(), devAccount)
	if err != nil {
		t.Fatalf("Load() error = %v, want the profile", err)
	}
	return p, revision
}

// payloadOf is a result's payload, read into what a case expects of it.
func payloadOf[T any](t *testing.T, result *mcp.CallToolResult) T {
	t.Helper()

	var payload T
	if err := json.Unmarshal(rawPayload(t, result), &payload); err != nil {
		t.Fatalf("the payload does not read: %v", err)
	}
	return payload
}

// rawPayload is a result's payload as it went over the wire.
func rawPayload(t *testing.T, result *mcp.CallToolResult) []byte {
	t.Helper()

	if result.IsError {
		t.Fatalf("the call failed: %s", textOf(t, result))
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("the payload does not encode: %v", err)
	}
	return raw
}
