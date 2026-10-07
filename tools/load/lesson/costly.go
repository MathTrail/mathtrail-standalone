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
	// Sets spends it on sets, the tables a helper builds, which hold far more
	// for every element than a list of the same elements.
	Sets = "sets"
	// Product spends it multiplying: a hundred thousand steps, and far more
	// work than a clock of the sandbox allows, since the numbers grow as they
	// are multiplied. Only the clock stops it.
	Product = "product"
	// Dicts spends it writing dictionaries in a loop: memory the language
	// takes by itself, a step for every dictionary whatever the dictionary
	// takes, and which no price of the sandbox reaches.
	Dicts = "dicts"
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
	// dictTurn is a turn of a comprehension that writes one dictionary.
	dictTurn = 9
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
	{Sets, func(steps uint64) string {
		return heldWork(fill(steps, maxElements, sets), "set(range(size))")
	}},
	{Product, func(uint64) string { return "    prod(range(1, 100001))\n" }},
	{Dicts, func(steps uint64) string {
		return fmt.Sprintf("    held = [{} for i in range(%d)]\n", steps/100*aim/dictTurn)
	}},
}

// What one call of a helper costs the budget, by the size it is called on:
// the elements it walks, the sequences it opens and the bytes it builds, and
// the turn of the loop that calls it.
func summed(size uint64) uint64 { return size + callTurn }

// tuples is a product of the two values repeated width times: the one
// sequence opened and its two elements walked, and the tuples built.
func tuples(width uint64) uint64 {
	return opened(1) + 2 + built(tuplesOf(1<<width, width)) + callTurn
}

// pairs is a product of a range with itself: both ranges opened and walked,
// and a pair built for every two elements.
func pairs(size uint64) uint64 {
	return opened(2) + 2*size + built(tuplesOf(size*size, 2)) + callTurn
}

// sets is a set of a range: the range walked, and a table built with an entry
// for every element.
func sets(size uint64) uint64 { return size + built(512+256*size) + callTurn }

// The price of the sandbox, as the solvers are sized to it: what a helper
// builds costs a step for every sixteen bytes, with a quarter on top.
func built(bytes uint64) uint64 { return (5*bytes + 63) / 64 }

// tuplesOf is what a list of count tuples of width values takes: the list, and
// for every tuple its place, its header and its values.
func tuplesOf(count, width uint64) uint64 { return 32 + count*(40+16*width) }

// opened is what a helper pays for the sequences it is handed, before it
// reads any of them.
func opened(sequences uint64) uint64 { return built(64 * sequences) }

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
		"task":       task.Body,
		"solver":     task.Solver,
		"self_check": task.SelfCheck,
	})
}
