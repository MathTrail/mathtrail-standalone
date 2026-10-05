// Package starlarktest watches the slots of the sandbox from outside, for the
// tests that have to know every slot is taken before they go on, or that the
// slots came back once a run is over. It asks nothing of the sandbox that a
// caller could not: it runs programs, as any caller does — one that ends as
// soon as it has a slot, and one that keeps its slot until the sandbox stops it.
// It also sets the clock a test gives a run, which the race detector stretches.
package starlarktest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// probe is a program that holds a slot for no time at all: it does not parse,
// and the sandbox parses a program only once it holds a slot, so a run of it
// that finds one free is over at once.
const probe = "def solve(options):\n    return (\n"

// probeOptions are five options for the probe to be given.
var probeOptions = solver.Options{"4", "5", "6", "8", "12"}

// EverySlotTaken returns once no slot of the runner is free. Until then a
// probe finds one and is answered at once; from then on it finds none in the
// moment it is given to look, whichever of its own deadline or the sandbox's
// wait ends that moment.
func EverySlotTaken(t testing.TB, runner solver.Runner) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		_, err := runner.Run(ctx, probe, probeOptions)
		cancel()
		if err != nil {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("a slot was still free after 10s, want every one taken")
}

// holder is a program that keeps its slot until the sandbox stops it.
const holder = "def solve(options):\n    while True:\n        pass\n    return []\n"

// WantEverySlotFree holds the runner to having every one of its slots back:
// as many runs as it has slots, each holding one until the sandbox stops it,
// all start, where a slot that was never given back would leave one of them
// waiting until the sandbox turned it away. It takes as long as a run may
// last, and it proves something only of a runner whose clock outlasts its wait
// and whose budget of steps outlasts its clock.
func WantEverySlotFree(t testing.TB, runner solver.Runner, slots int) {
	t.Helper()

	refused := make([]error, slots)
	var together sync.WaitGroup
	for i := range slots {
		together.Go(func() { _, refused[i] = runner.Run(context.Background(), holder, probeOptions) })
	}
	together.Wait()
	for i, err := range refused {
		if err != nil {
			t.Errorf("run %d of %d, each holding a slot: error %v, want every slot free", i+1, slots, err)
		}
	}
}
