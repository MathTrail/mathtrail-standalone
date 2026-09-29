package lesson

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// The costly solvers. Each spends as much of a run as the rules of the sandbox
// let one run spend, in a way of its own, and each ends by pointing at the one
// right option, so that a hand-in costs both of its runs whenever the clock
// lets the first one end.
const (
	// Loop spends the step budget on the interpreter's own steps: the plainest
	// way for a run to be expensive.
	Loop = "loop"
	// Appends spends it growing a list one element at a time.
	Appends = "appends"
	// Helper spends it inside sum, whose every element is a step the
	// interpreter never takes, and a slower one than those it does.
	Helper = "helper"
	// Tuples spends it on the widest tuples a helper builds.
	Tuples = "tuples"
	// Pairs spends it on pairs, the tuples that hold the most memory for every
	// step they cost.
	Pairs = "pairs"
	// Product spends it multiplying: a hundred thousand steps, and far more
	// work than a clock of the sandbox allows, since the numbers grow as they
	// are multiplied. Only the clock stops it.
	Product = "product"
)

// MinSteps is the smallest step ceiling the costly solvers are sized for.
// Below it, the budget would stop the hundred thousand numbers of Product
// before the clock does.
const MinSteps = 200_000

// A costly solver spends between nine and ten tenths of the ceiling it is
// sized to: near enough to the ceiling to be as costly as a run may be, and
// far enough from it that the interpreter's own steps around the work never
// take it over.
const (
	// aim is the share, in percent, a solver of one piece of work spends.
	aim = 95
	// least and most bound what a solver of many pieces spends.
	least = 90
	most  = 99
)

// The interpreter's own steps for what the solvers repeat, as counted on the
// language as the service runs it: one turn of a loop that adds one, and one
// that appends one. The tests hold every solver to its share, so a language
// that counts otherwise shows at once.
const (
	loopTurn   = 10
	appendTurn = 11
	// callTurn is a turn of the loop that calls a helper, beside what the
	// helper charges.
	callTurn = 14
)

// maxElements is the most any helper of the sandbox walks or builds in one
// call.
const maxElements = 1_000_000

// widestPairs is the widest a solver builds a product of pairs: a thousand of
// a thousand is the most one call may build.
const widestPairs = 1000

// widestTuples is the widest a product of two values may be: two to the
// nineteenth tuples is the most one call may build.
const widestTuples = 19

// costly is every costly solver, in the order an adversarial run takes them,
// each with the work it does for the ceiling given.
var costly = []struct {
	name string
	work func(steps uint64) string
}{
	{Loop, func(steps uint64) string {
		return fmt.Sprintf("    n = 0\n    for i in range(%d):\n        n += 1\n", steps/100*aim/loopTurn)
	}},
	{Appends, func(steps uint64) string {
		return fmt.Sprintf("    xs = []\n    for i in range(%d):\n        xs.append(i)\n", steps/100*aim/appendTurn)
	}},
	{Helper, func(steps uint64) string {
		return heldWork(fill(steps, maxElements, summed), "sum(range(size))")
	}},
	{Tuples, func(steps uint64) string {
		return heldWork(fill(steps, widestTuples, tuples), "product([0, 1], repeat=size)")
	}},
	{Pairs, func(steps uint64) string {
		return heldWork(fill(steps, widestPairs, pairs), "product(range(size), range(size))")
	}},
	{Product, func(uint64) string { return "    prod(range(1, 100001))\n" }},
}

// What one call of a helper costs the budget, by the size it is called on:
// the elements it walks and the values it builds, and the turn of the loop
// that calls it.
func summed(size uint64) uint64 { return size + callTurn }

// tuples is a product of the two values repeated width times: the two
// elements walked, and every value of every tuple.
func tuples(width uint64) uint64 { return 2 + width<<width + callTurn }

// pairs is a product of a range with itself: both ranges walked, and two
// values for every pair.
func pairs(size uint64) uint64 { return 2*size + 2*size*size + callTurn }

// Costly are the names of the costly solvers.
func Costly() []string {
	names := make([]string, len(costly))
	for i, solver := range costly {
		names[i] = solver.name
	}
	return names
}

// CostlyTask is a task whose solver is the costly one named, sized to the step
// ceiling given: the first of the written tasks, which the solver proves once
// it has spent what it came to spend.
func CostlyTask(name string, steps uint64) (Task, error) {
	if steps < MinSteps {
		return Task{}, fmt.Errorf("lesson: the costly solvers are sized for %d steps at least, not %d", MinSteps, steps)
	}
	for _, solver := range costly {
		if solver.name != name {
			continue
		}
		task := race()
		task.Solver = "def solve(options):\n" + solver.work(steps) +
			"    return match(options, " + strconv.Quote(task.Body.Options[task.Body.Correct]) + ")\n"
		return task, nil
	}
	return Task{}, fmt.Errorf("lesson: no costly solver %q; there are %s", name, strings.Join(Costly(), ", "))
}

// piece is one size of work a solver does, and how many times over.
type piece struct{ size, times uint64 }

// fill is the work a solver holds at once, the largest first: the largest
// size that fits under the most a solver may spend, as many times as it fits,
// then the largest that fits in what is left, until the solver has spent the
// least it should. It is how the budget is spent on work that comes in sizes,
// whose costs do not add up to any share exactly — such as tuples of a given
// width. Sizes run from one to the largest given, and cost more as they grow.
func fill(steps, largest uint64, cost func(size uint64) uint64) []piece {
	var pieces []piece
	var spent uint64
	for spent < steps/100*least {
		room := steps/100*most - spent
		size := largestWithin(largest, room, cost)
		if size == 0 {
			break
		}
		times := room / cost(size)
		pieces = append(pieces, piece{size: size, times: times})
		spent += times * cost(size)
	}
	return pieces
}

// largestWithin is the largest size, from one to the largest given, whose
// cost fits in the room, or nothing when none does.
func largestWithin(largest, room uint64, cost func(size uint64) uint64) uint64 {
	var fits uint64
	for low, high := uint64(1), largest; low <= high; {
		middle := low + (high-low)/2
		if cost(middle) <= room {
			fits, low = middle, middle+1
		} else {
			high = middle - 1
		}
	}
	return fits
}

// heldWork is the work of a solver that calls a helper on every size of the
// pieces, as many times as each is given, and holds what every call returns
// until the answer.
func heldWork(pieces []piece, call string) string {
	sizes := make([]string, len(pieces))
	for i, p := range pieces {
		sizes[i] = fmt.Sprintf("(%d, %d)", p.size, p.times)
	}
	return fmt.Sprintf("    held = []\n    for size, times in [%s]:\n        for i in range(times):\n            held.append(%s)\n",
		strings.Join(sizes, ", "), call)
}

// staleRequest is the id of a request nobody opened.
const staleRequest = "req_00000000-0000-4000-8000-000000000000"

// HandInStale hands a task in for a request nobody opened. The service runs
// its solver first, both runs, and only then finds no request to hold it
// against: the task is turned away as stale, nothing is spent and nothing is
// written, and the sandbox has done all its work. The child needs no profile
// for it.
func HandInStale(ctx context.Context, child *session.Child, task *Task) session.Answer {
	return child.Call(ctx, "submit_task", map[string]any{
		"request_id": staleRequest,
		"brief":      map[string]any{},
		"task":       task.Body,
		"solver":     task.Solver,
		"self_check": task.SelfCheck,
	})
}
