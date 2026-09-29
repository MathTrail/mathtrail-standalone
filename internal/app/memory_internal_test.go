package app

import (
	"math"
	"runtime/debug"
	"testing"
)

// The soft limit the log gives is the one the runtime was told, and nothing
// when it was told none, rather than the largest number there is.
//
// Not parallel: the limit belongs to the whole process, and it is put back
// before any test that runs in parallel starts.
func TestTheSoftMemoryLimitIsWhatTheRuntimeWasTold(t *testing.T) {
	told := debug.SetMemoryLimit(1 << 30)
	t.Cleanup(func() { debug.SetMemoryLimit(told) })

	if got := softMemoryLimit(); got != 1<<30 {
		t.Errorf("with a limit of 1 GiB: got %d, want %d", got, 1<<30)
	}
	debug.SetMemoryLimit(math.MaxInt64)
	if got := softMemoryLimit(); got != 0 {
		t.Errorf("with no limit: got %d, want 0", got)
	}
}
