package starlark

import (
	"fmt"
	"slices"

	"go.starlark.net/starlark"
)

// The four helpers here are most of what a brute force is made of, and they
// are the four Python keeps in itertools. Each of them returns the whole list
// rather than anything lazy: the list is what the ceiling of this sandbox can
// be about, and a solver that wants one tuple at a time can walk it.
//
// Every list comes out in the order Python produces it. The model writing a
// solver is one that knows Python, and a helper that quietly reordered its
// answers would be the hardest kind of difference to find — the solver would
// still run, and only the answer would be wrong.
//
// Elements are taken by position and not by value, again as in Python: a pool
// holding the same thing twice holds two elements, and the tuples say so.

// permutations lists every ordering of r of the elements, r being the whole
// pool unless it is given.
func permutations(steps uint64) *starlark.Builtin {
	return starlark.NewBuiltin("permutations", func(
		thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		pool, width, err := poolAndWidth(thread, steps, builtin, args, kwargs, "r?")
		if err != nil {
			return nil, err
		}

		// One ordering for an empty tuple, and one fewer element to choose
		// from at every place after the first.
		count := 1
		if width > len(pool) {
			count = 0
		}
		for place := 0; place < width && count > 0; place++ {
			count = times(count, len(pool)-place)
		}
		if err = charge(thread, steps, count, width); err != nil {
			return nil, fmt.Errorf("%s: %w", builtin.Name(), err)
		}
		return starlark.NewList(orderings(pool, width, count)), nil
	})
}

// combinations lists every choice of r of the elements, each in the order the
// pool has them.
func combinations(steps uint64) *starlark.Builtin {
	return starlark.NewBuiltin("combinations", func(
		thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		pool, width, err := poolAndWidth(thread, steps, builtin, args, kwargs, "r")
		if err != nil {
			return nil, err
		}

		count := choose(len(pool), width)
		if err = charge(thread, steps, count, width); err != nil {
			return nil, fmt.Errorf("%s: %w", builtin.Name(), err)
		}
		return starlark.NewList(choices(pool, width, count)), nil
	})
}

// combinationsWithReplacement is the same as combinations, except that an
// element may be chosen again.
func combinationsWithReplacement(steps uint64) *starlark.Builtin {
	return starlark.NewBuiltin("combinations_with_replacement", func(
		thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		pool, width, err := poolAndWidth(thread, steps, builtin, args, kwargs, "r")
		if err != nil {
			return nil, err
		}

		// A tuple of nothing is one tuple out of any pool at all, including no
		// pool, and the count below does not describe that: it would ask how
		// many ways there are to take nothing out of one fewer than none.
		// Otherwise one choice of width from a pool that may be drawn on again
		// is one choice of width from a pool of that many more.
		count := 1
		if width > 0 {
			count = choose(len(pool)+width-1, width)
		}
		if err = charge(thread, steps, count, width); err != nil {
			return nil, fmt.Errorf("%s: %w", builtin.Name(), err)
		}
		return starlark.NewList(choicesWithRepeats(pool, width, count)), nil
	})
}

// product lists every way of taking one element from each sequence in turn,
// the last sequence moving fastest. With repeat it is as if the sequences had
// been given that many times over.
func product(steps uint64) *starlark.Builtin {
	return starlark.NewBuiltin("product", func(
		thread *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		name := builtin.Name()
		// The sequences arrive one after another as they are written, so only
		// repeat is looked for among the keywords.
		repeat := 1
		if err := starlark.UnpackArgs(name, nil, kwargs, "repeat?", &repeat); err != nil {
			return nil, err
		}
		if repeat < 0 || repeat > maxElements {
			return nil, fmt.Errorf("%s: repeat is %d, and it must be between 0 and %d", name, repeat, maxElements)
		}

		given := make([][]starlark.Value, 0, len(args))
		for _, argument := range args {
			pool, err := elementsOf(thread, steps, name, argument)
			if err != nil {
				return nil, err
			}
			given = append(given, pool)
		}

		count, width := crossingCount(given, repeat)
		if err := charge(thread, steps, count, width); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		pools := make([][]starlark.Value, 0, width)
		for range repeat {
			pools = append(pools, given...)
		}
		return starlark.NewList(crossings(pools, count)), nil
	})
}

// orderings is every ordering of width of the pool: the leftmost place moves
// slowest, and each place takes the lowest element not already taken.
func orderings(pool []starlark.Value, width, count int) []starlark.Value {
	if count == 0 {
		return nil
	}
	tuples := make([]starlark.Value, 0, count)
	taken := make([]bool, len(pool))
	ordering := make(starlark.Tuple, width)

	var fill func(place int)
	fill = func(place int) {
		if place == width {
			tuples = append(tuples, slices.Clone(ordering))
			return
		}
		for i, element := range pool {
			if taken[i] {
				continue
			}
			taken[i] = true
			ordering[place] = element
			fill(place + 1)
			taken[i] = false
		}
	}
	fill(0)
	return tuples
}

// choices is every choice of width of the pool. The places hold positions in
// the pool and always rise; the last one that can still move is the one that
// moves, and everything after it starts again just behind it.
func choices(pool []starlark.Value, width, count int) []starlark.Value {
	if count == 0 {
		return nil
	}
	tuples := make([]starlark.Value, 0, count)
	places := make([]int, width)
	for i := range places {
		places[i] = i
	}
	for {
		tuples = append(tuples, tupleAt(pool, places))

		last := width - 1
		for last >= 0 && places[last] == last+len(pool)-width {
			last--
		}
		if last < 0 {
			return tuples
		}
		places[last]++
		for next := last + 1; next < width; next++ {
			places[next] = places[next-1] + 1
		}
	}
}

// choicesWithRepeats is the same walk over places that never fall rather than
// places that always rise, so the one that moves takes everything after it
// along to where it now stands.
func choicesWithRepeats(pool []starlark.Value, width, count int) []starlark.Value {
	if count == 0 {
		return nil
	}
	tuples := make([]starlark.Value, 0, count)
	places := make([]int, width)
	for {
		tuples = append(tuples, tupleAt(pool, places))

		last := width - 1
		for last >= 0 && places[last] == len(pool)-1 {
			last--
		}
		if last < 0 {
			return tuples
		}
		places[last]++
		for next := last + 1; next < width; next++ {
			places[next] = places[last]
		}
	}
}

// crossings is one element from each pool, every way round: a row of counters
// with the last one moving fastest, running to the end of its pool, going back
// to the start and moving the one before it.
func crossings(pools [][]starlark.Value, count int) []starlark.Value {
	if count == 0 {
		return nil
	}
	tuples := make([]starlark.Value, 0, count)
	places := make([]int, len(pools))
	for {
		tuple := make(starlark.Tuple, len(pools))
		for place, index := range places {
			tuple[place] = pools[place][index]
		}
		tuples = append(tuples, tuple)

		last := len(pools) - 1
		for last >= 0 {
			places[last]++
			if places[last] < len(pools[last]) {
				break
			}
			places[last] = 0
			last--
		}
		if last < 0 {
			return tuples
		}
	}
}

// crossingCount is how many tuples the pools make and how wide each one is.
// Every repetition multiplies the count again, so a pool of one or none needs
// no repeating to know the answer — and a count past the ceiling is a count
// nobody has to work out exactly.
func crossingCount(given [][]starlark.Value, repeat int) (count, width int) {
	once := 1
	for _, pool := range given {
		once = times(once, len(pool))
	}

	count = once
	if once > 1 {
		for range repeat - 1 {
			count = times(count, once)
			if count > maxElements {
				break
			}
		}
	}
	if repeat == 0 {
		count = 1
	}
	return count, len(given) * repeat
}

// poolAndWidth reads the two arguments the helpers above take: the sequence to
// draw from, and how many of it goes into each tuple. The second is declared
// to the unpacker as the helper wants it — required, or optional with the
// whole pool standing in.
func poolAndWidth(
	thread *starlark.Thread, steps uint64, builtin *starlark.Builtin,
	args starlark.Tuple, kwargs []starlark.Tuple, declaration string,
) ([]starlark.Value, int, error) {
	name := builtin.Name()
	var sequence, given starlark.Value
	if problem := starlark.UnpackArgs(name, args, kwargs, "seq", &sequence, declaration, &given); problem != nil {
		return nil, 0, problem
	}
	pool, err := elementsOf(thread, steps, name, sequence)
	if err != nil {
		return nil, 0, err
	}
	width, err := widthOf(name, given, len(pool))
	return pool, width, err
}

// widthOf is how many elements go into each tuple: what was asked for, or the
// whole pool when nothing was.
func widthOf(name string, given starlark.Value, whole int) (int, error) {
	if given == nil || given == starlark.None {
		return whole, nil
	}
	width, err := starlark.AsInt32(given)
	if err != nil {
		return 0, fmt.Errorf("%s: r must be a whole number: %w", name, err)
	}
	switch {
	case width < 0:
		return 0, fmt.Errorf("%s: r is %d, and a tuple cannot be shorter than empty", name, width)
	case width > maxElements:
		return 0, fmt.Errorf("%s: a tuple of %d values, and %d is the most that will be walked at once",
			name, width, maxElements)
	}
	return width, nil
}

// tupleAt is the elements of the pool standing at these places.
func tupleAt(pool []starlark.Value, places []int) starlark.Tuple {
	tuple := make(starlark.Tuple, len(places))
	for place, index := range places {
		tuple[place] = pool[index]
	}
	return tuple
}

// times multiplies and stops at one past the ceiling. A count too large to
// build is a count nobody needs the exact value of — and working it out would
// overflow long before it was reached.
func times(running, factor int) int {
	if running <= 0 || factor <= 0 {
		return 0
	}
	if factor > (maxElements+1)/running {
		return maxElements + 1
	}
	return running * factor
}

// choose is how many ways there are to take r of n, stopping at one past the
// ceiling for the same reason times does. Each step divides exactly, because
// the product of any r running numbers is divisible by the factorial of r, so
// nothing here has to hold a number larger than the ceiling and a factor.
//
// That product is counted in sixty-four bits whatever the machine. It reaches
// two million million, which is nothing next to what sixty-four bits hold and
// more than thirty-two would, and a count that wrapped round to a negative
// number would pass every ceiling there is.
func choose(n, r int) int {
	if r < 0 || r > n {
		return 0
	}
	if r > n-r {
		r = n - r
	}
	ways := int64(1)
	for i := range r {
		ways = ways * int64(n-i) / int64(i+1)
		if ways > maxElements {
			return maxElements + 1
		}
	}
	return int(ways)
}
