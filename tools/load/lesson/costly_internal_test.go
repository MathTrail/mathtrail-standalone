package lesson

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// planned is what the pieces of a plan cost the budget together.
func planned(pieces []piece, cost func(size uint64) uint64) uint64 {
	var total uint64
	for _, p := range pieces {
		total += p.times * cost(p.size)
	}
	return total
}

// Whatever the ceiling, a solver made of pieces plans to spend nine to ten
// tenths of it, and every costly solver is short enough for the sandbox to
// read.
func TestEveryCostlySolverIsSizedToAnyCeiling(t *testing.T) {
	t.Parallel()

	ceilings := gen.UInt64Range(MinSteps, 1_000_000_000_000)
	properties := gopter.NewProperties(gopter.DefaultTestParameters())
	for _, plan := range []struct {
		name    string
		largest uint64
		cost    func(size uint64) uint64
	}{
		{Helper, maxElements, summed},
		{Tuples, widestTuples, tuples},
		{Pairs, widestPairs, pairs},
	} {
		properties.Property(plan.name+" plans nine to ten tenths of the ceiling", prop.ForAll(
			func(steps uint64) bool {
				total := planned(fill(steps, plan.largest, plan.cost), plan.cost)
				return total >= steps/100*least && total <= steps/100*most
			}, ceilings))
	}
	properties.Property("every costly solver is short enough to read", prop.ForAll(
		func(steps uint64) bool {
			for _, name := range Costly() {
				task, err := CostlyTask(name, steps)
				if err != nil || len(task.Solver) > 8*1024 {
					return false
				}
			}
			return true
		}, ceilings))
	properties.TestingRun(t)
}
