package scenario_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/scenario"
	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
)

// kinds are how many calls of the run came to each kind of answer, by the
// name a report writes it under.
func kinds(run *report.Run) map[string]int {
	counted := map[string]int{}
	for i := range run.Calls {
		counted[run.Calls[i].Kind.String()]++
	}
	return counted
}

// failedNone is a case that fails if any of the runs failed.
func failedNone(t *testing.T, came []report.Run) {
	t.Helper()

	for i := range came {
		if hards := came[i].Verdict(); len(hards) != 0 {
			t.Errorf("%s: Verdict() = %+v, want nothing", came[i].Scenario, hards)
		}
	}
}

// Saturation examines what the sandbox has a slot for, turned away as stale,
// and turns away as busy what it has none for in time: the service keeping its
// answers in time rather than failing. The service comes back after it.
func TestSaturationTurnsAwayWhatTheSandboxHasNoSlotFor(t *testing.T) {
	t.Parallel()

	// One slot, every run of which the clock ends, and a short wait for it: a
	// queue at once. The ceiling is out of every loop's reach.
	const ceiling = 1_000_000_000_000
	target := servicetest.Start(t,
		"MATHTRAIL_SOLVER_CONCURRENCY=1", "MATHTRAIL_SOLVER_TIMEOUT=300ms", "MATHTRAIL_SOLVER_WAIT=100ms",
		"MATHTRAIL_SOLVER_STEPS=1000000000000", "MATHTRAIL_RATE_USER_PER_MIN=600")
	o := optionsOf(t, scenario.Saturation)
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 20, 2*time.Second, 3, ceiling, 10*time.Second

	came := runs(t.Context(), t, &o, target)
	if len(came) != 2 {
		t.Fatalf("saturation came to %d runs, want the attack and the recovery", len(came))
	}
	if attack := kinds(&came[0]); attack["stale:stale_request"] == 0 || attack["failure:busy"] == 0 {
		t.Errorf("the hand-ins came to %v, want some examined and turned away as stale, and some as busy", attack)
	}
	failedNone(t, came)
}

// Every costly solver, handed in again and again, leaves the service as it
// found it: nothing it should never do, and a lesson given right after.
func TestEveryCostlySolverLeavesTheServiceAlive(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t,
		"MATHTRAIL_SOLVER_STEPS=200000", "MATHTRAIL_SOLVER_TIMEOUT=300ms", "MATHTRAIL_RATE_USER_PER_MIN=600")
	o := optionsOf(t, scenario.Adversarial)
	o.Variants = lesson.Costly()
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 600*time.Millisecond, 2, lesson.MinSteps, 10*time.Second

	came := runs(t.Context(), t, &o, target)
	if len(came) != 2*len(o.Variants) {
		t.Fatalf("adversarial came to %d runs, want an attack and a recovery for each of %d variants", len(came), len(o.Variants))
	}
	for i := 0; i < len(came); i += 2 {
		if kinds(&came[i])["stale:stale_request"] == 0 {
			t.Errorf("%s: no hand-in was examined: %v", came[i].Scenario, kinds(&came[i]))
		}
	}
	failedNone(t, came)
}

// A hand-in the service answers otherwise than a task for no request is —
// judged and refused, here — fails the attack: it was turned away neither as
// stale, nor as busy, nor by a pace.
func TestAHandInAnsweredOtherwiseFailsTheAttack(t *testing.T) {
	t.Parallel()

	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"submit_task": func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				Content:           []mcp.Content{&mcp.TextContent{Text: "Refused, attempt 1 of 3."}},
				StructuredContent: map[string]any{"status": "rejected", "code": "solver_error"},
			}, nil
		},
	}, nil)
	o := optionsOf(t, scenario.Saturation)
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 10, 300*time.Millisecond, 1, lesson.MinSteps, 300*time.Millisecond

	came := runs(t.Context(), t, &o, target)
	if broken := strings.Join(came[0].Broken, "; "); !strings.Contains(broken, "rejected:solver_error") {
		t.Errorf("%s: Broken = %q, want the attack failed for the hand-ins refused", came[0].Scenario, broken)
	}
}

// An attack whose every hand-in was turned away before its solver ran — busy,
// here — never reached the sandbox it is meant to load, and fails.
func TestAnAttackThatExaminedNothingFails(t *testing.T) {
	t.Parallel()

	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"submit_task": func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{
				Text: "MathTrail is checking too many tasks right now, so this task was not checked and no attempt was spent.",
			}}}, nil
		},
	}, nil)
	o := optionsOf(t, scenario.Saturation)
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 10, 300*time.Millisecond, 1, lesson.MinSteps, 300*time.Millisecond

	came := runs(t.Context(), t, &o, target)
	if broken := strings.Join(came[0].Broken, "; "); !strings.Contains(broken, "no hand-in was examined") {
		t.Errorf("%s: Broken = %q, want the attack failed for examining nothing", came[0].Scenario, broken)
	}
}

// A lesson of the recovery that a pace held back is walked again, by another
// child, until the pace lets it through: a pace shedding what is left of an
// attack is the service doing as it should, not a service that did not come
// back.
func TestARecoveryWaitsOutAPace(t *testing.T) {
	t.Parallel()

	var held atomic.Int32
	target := inFront(t, servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600"), func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The first profile read of each of the first two lessons.
			if r.Header.Get("Mcp-Name") == "get_profile" && held.Add(1) <= 2 {
				refuseAsAPace(t, w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	o := optionsOf(t, scenario.Saturation)
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 500*time.Millisecond, 1, lesson.MinSteps, 10*time.Second

	came := runs(t.Context(), t, &o, target)
	recovered := &came[len(came)-1]
	if hards := recovered.Verdict(); len(hards) != 0 || recovered.Units != 1 || held.Load() < 3 {
		t.Errorf("%s: Verdict() = %+v, %d task accepted, %d profile reads; want the lesson walked again until it passed",
			recovered.Scenario, hards, recovered.Units, held.Load())
	}
}

// A recovery is timed from the first time the service was asked, so that its
// length is how long the service took to come back — here, a probe that fails
// for more than half a second — and a service that did come back fails
// nothing.
func TestARecoveryIsTimedFromTheFirstProbe(t *testing.T) {
	t.Parallel()

	const down = 600 * time.Millisecond
	var firstProbe atomic.Int64
	target := inFront(t, servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600"), func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				firstProbe.CompareAndSwap(0, time.Now().UnixNano())
				if time.Since(time.Unix(0, firstProbe.Load())) < down {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	})
	o := optionsOf(t, scenario.Saturation)
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 500*time.Millisecond, 1, lesson.MinSteps, 10*time.Second

	came := runs(t.Context(), t, &o, target)
	recovered := &came[len(came)-1]
	if hards := recovered.Verdict(); len(hards) != 0 || recovered.Ended.Sub(recovered.Began) < down {
		t.Errorf("%s: Verdict() = %+v, lasting %v; want nothing failed, over at least the %v the probe failed for",
			recovered.Scenario, hards, recovered.Ended.Sub(recovered.Began), down)
	}
}

// A run that ended before the scenario was asked to stop keeps what it came
// to — here, a service that did not come back after the first attack — while
// the attack the stop cut short is marked as stopped, and no recovery follows
// it.
func TestARunThatEndedBeforeTheStopKeepsWhatItCameTo(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(t.Context())
	defer stop()
	var probed atomic.Bool
	target := inFront(t, servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600"), func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/health":
				probed.Store(true)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			case probed.Load() && r.Header.Get("Mcp-Name") == "submit_task":
				// The second attack has begun, and it is the one cut short.
				stop()
			}
			next.ServeHTTP(w, r)
		})
	})
	o := optionsOf(t, scenario.Adversarial)
	o.Variants = []string{lesson.Loop, lesson.Appends}
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 500*time.Millisecond, 1, lesson.MinSteps, 500*time.Millisecond

	came := runs(ctx, t, &o, target)
	if len(came) != 3 {
		t.Fatalf("adversarial came to %d runs, want the first attack, its recovery and the second attack", len(came))
	}
	first, cut := &came[1], &came[2]
	if first.Stopped || !strings.Contains(strings.Join(first.Broken, "; "), "did not come back") || !cut.Stopped {
		t.Errorf("the first recovery is stopped %v with %q, and the second attack stopped %v; want the recovery "+
			"failed as it ended, and the attack cut short marked as stopped", first.Stopped, first.Broken, cut.Stopped)
	}
}

// refuseAsAPace answers a request of the protocol the way a pace refuses a
// message: an error of the protocol's, too many requests.
func refuseAsAPace(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("read the request: %v", err)
		return
	}
	message, err := jsonrpc.DecodeMessage(body)
	request, isRequest := message.(*jsonrpc.Request)
	if err != nil || !isRequest {
		t.Errorf("the request is %T (%v), want a request of the protocol", message, err)
		return
	}
	refusal, err := jsonrpc.EncodeMessage(&jsonrpc.Response{
		ID: request.ID, Error: &jsonrpc.Error{Code: -32000, Message: "too many requests; try again in a moment"},
	})
	if err != nil {
		t.Errorf("encode the refusal: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(refusal)
}

// A service that does not come back after an attack fails the run, however
// well it took the attack itself.
func TestAServiceThatDoesNotComeBackFailsTheRun(t *testing.T) {
	t.Parallel()

	target := inFront(t, servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600"), func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	o := optionsOf(t, scenario.Saturation)
	o.Rate, o.Duration, o.Children, o.Steps, o.Timeout = 5, 500*time.Millisecond, 1, lesson.MinSteps, time.Second

	came := runs(t.Context(), t, &o, target)
	recovered := &came[len(came)-1]
	var said []string
	for _, hard := range recovered.Verdict() {
		said = append(said, hard.What)
	}
	if !strings.Contains(strings.Join(said, "; "), "did not come back") {
		t.Errorf("%s: Verdict() = %q, want the run failed for a service that did not come back", recovered.Scenario, said)
	}
}
