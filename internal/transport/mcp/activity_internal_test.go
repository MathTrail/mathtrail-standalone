package mcpserver

import (
	"context"
	"testing"
)

// A call is counted by the chat host the boundary says it came from, and a
// call that came by no boundary by an unknown one, rather than by none.
func TestACallIsCountedByTheHostItCameFrom(t *testing.T) {
	t.Parallel()

	if got := hostFrom(withHost(context.Background(), "claude")); got != "claude" {
		t.Errorf("hostFrom() of a call from claude = %q, want claude", got)
	}
	if got := hostFrom(context.Background()); got != "unknown" {
		t.Errorf("hostFrom() of a call by no boundary = %q, want unknown", got)
	}
}
