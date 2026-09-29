package scenario_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MathTrail/mathtrail-standalone/tools/load/scenario"
	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
)

// Only the greedy are held back: the child calling past the pace of an
// account, and the address fetching past the pace of an address. The children
// walking their lessons beside them and the quiet addresses never are, and the
// service comes back.
func TestOnlyTheGreedyAreHeldBack(t *testing.T) {
	t.Parallel()

	// Paces a run of two seconds goes past, and a pace of the instance nothing
	// here reaches.
	target := servicetest.Start(t,
		"MATHTRAIL_RATE_USER_PER_MIN=60", "MATHTRAIL_RATE_IP_PER_MIN=120", "MATHTRAIL_RATE_INSTANCE_PER_MIN=6000")
	o := optionsOf(t, scenario.Limits)
	o.Rate, o.Duration, o.Pace, o.QuietEvery, o.Timeout = 50, 2*time.Second, 0, 100*time.Millisecond, 10*time.Second

	came := runs(t.Context(), t, &o, target)
	if len(came) != 5 {
		t.Fatalf("limits came to %d runs, want the four parts and the recovery", len(came))
	}
	for _, part := range []struct {
		run         int
		wantHeld    bool
		description string
	}{
		{0, true, "the greedy child"},
		{1, false, "the other children"},
		{2, true, "the greedy address"},
		{3, false, "the quiet addresses"},
	} {
		counted := kinds(&came[part.run])
		if held := counted["limited:paced"] > 0; held != part.wantHeld || counted["ok"] == 0 {
			t.Errorf("%s came to %v; want held back: %v, and answered as asked at least once", part.description, counted, part.wantHeld)
		}
	}
	failedNone(t, came)
}

// A pace that holds back the wrong one fails the run: a greedy child it never
// held back, and a quiet address it did.
func TestAPaceThatHoldsBackTheWrongOneFailsTheRun(t *testing.T) {
	t.Parallel()

	// A pace of an account nothing here reaches, and one of an address that
	// holds back every fetch after the first.
	target := servicetest.Start(t,
		"MATHTRAIL_RATE_USER_PER_MIN=6000", "MATHTRAIL_RATE_IP_PER_MIN=1", "MATHTRAIL_RATE_INSTANCE_PER_MIN=6000")
	o := optionsOf(t, scenario.Limits)
	o.Rate, o.Duration, o.Pace, o.QuietEvery, o.Timeout, o.Children, o.Tasks = 20, time.Second, 0, 100*time.Millisecond, 10*time.Second, 1, 1

	came := runs(t.Context(), t, &o, target)
	for _, want := range []struct {
		run  int
		says string
	}{
		{0, "never held back"},
		{3, "answered otherwise than served"},
	} {
		if broken := strings.Join(came[want.run].Broken, "; "); !strings.Contains(broken, want.says) {
			t.Errorf("%s: Broken = %q, want the run failed as %q", came[want.run].Scenario, broken, want.says)
		}
	}
}

// A greedy child answered otherwise than as asked or held back fails its part,
// even with nothing hard among its answers.
func TestAGreedyChildAnsweredOtherwiseFailsItsPart(t *testing.T) {
	t.Parallel()

	target := servicetest.Fake(t, map[string]mcp.ToolHandler{
		"get_profile": func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				Content:           []mcp.Content{&mcp.TextContent{Text: "No more tasks today."}},
				StructuredContent: map[string]any{"status": "limited", "code": "daily_tasks"},
			}, nil
		},
	}, nil)
	o := optionsOf(t, scenario.Limits)
	o.Rate, o.Duration, o.Pace, o.QuietEvery, o.Timeout, o.Children, o.Tasks = 20, 300*time.Millisecond, 0, 100*time.Millisecond, 300*time.Millisecond, 1, 1

	came := runs(t.Context(), t, &o, target)
	const says = "answered otherwise than as asked or held back: limited:daily_tasks"
	if broken := strings.Join(came[0].Broken, "; "); !strings.Contains(broken, says) {
		t.Errorf("%s: Broken = %q, want the run failed as %q", came[0].Scenario, broken, says)
	}
}
