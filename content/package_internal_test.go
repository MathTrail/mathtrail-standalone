package content

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// cell is a pool of reference tasks with so many at each difficulty, each
// named by its difficulty and its number.
func cell(counts map[int]int) []Example {
	var pool []Example
	for _, difficulty := range slices.Sorted(maps.Keys(counts)) {
		for i := 1; i <= counts[difficulty]; i++ {
			pool = append(pool, Example{ID: fmt.Sprintf("d%d-%d", difficulty, i), Difficulty: difficulty})
		}
	}
	return pool
}

// idsOf are the ids of some tasks, in their order.
func idsOf(tasks []Example) []string {
	ids := make([]string, len(tasks))
	for i := range tasks {
		ids[i] = tasks[i].ID
	}
	return ids
}

func TestTheExamplesAreTheNearestToTheDifficulty(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name                string
		counts              map[int]int
		difficulty, answers int
		want                []string
	}{
		{"three of the requested difficulty", map[int]int{2: 5, 3: 5}, 3, 0, []string{"d3-1", "d3-2", "d3-3"}},
		{"the next answer brings the next three", map[int]int{3: 5}, 3, 1, []string{"d3-2", "d3-3", "d3-4"}},
		{"the turn comes round", map[int]int{3: 5}, 3, 4, []string{"d3-5", "d3-1", "d3-2"}},
		{"five answers later, the same three", map[int]int{3: 5}, 3, 5, []string{"d3-1", "d3-2", "d3-3"}},
		{"too few at the difficulty: topped up from the nearest", map[int]int{1: 3, 3: 2, 4: 3}, 3, 0,
			[]string{"d3-1", "d3-2", "d4-1"}},
		{"of two equally near, the easier first", map[int]int{2: 3, 4: 3}, 3, 0, []string{"d2-1", "d2-2", "d2-3"}},
		{"none at the difficulty: the nearest there is", map[int]int{1: 1, 5: 3}, 4, 0, []string{"d5-1", "d5-2", "d5-3"}},
		{"fewer than three in all: all of them", map[int]int{3: 2}, 3, 0, []string{"d3-1", "d3-2"}},
		{"a count that is not positive still turns", map[int]int{3: 5}, 3, -1, []string{"d3-5", "d3-1", "d3-2"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := idsOf(nearest(cell(test.counts), test.difficulty, test.answers)); !slices.Equal(got, test.want) {
				t.Errorf("nearest() = %v, want %v", got, test.want)
			}
		})
	}
}

// drawnCell is a cell whose tasks of these ids carry a picture.
func drawnCell(counts map[int]int, drawn ...string) []Example {
	pool := cell(counts)
	for i := range pool {
		if slices.Contains(drawn, pool[i].ID) {
			pool[i].Picture = json.RawMessage(`{"kind":"ring","count":9}`)
		}
	}
	return pool
}

func TestAPackageShowsATaskThatDrawsWhereThePoolHasOne(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name                string
		counts              map[int]int
		drawn               []string
		difficulty, answers int
		want                []string
	}{
		{"one of those picked draws: they stay", map[int]int{3: 5}, []string{"d3-2"}, 3, 0,
			[]string{"d3-1", "d3-2", "d3-3"}},
		{"none of those picked draws: the last gives way", map[int]int{3: 5}, []string{"d3-5"}, 3, 0,
			[]string{"d3-1", "d3-2", "d3-5"}},
		{"the task that draws nearest the difficulty", map[int]int{1: 1, 3: 5, 4: 1}, []string{"d1-1", "d4-1"}, 3, 0,
			[]string{"d3-1", "d3-2", "d4-1"}},
		{"of two equally near, the easier", map[int]int{2: 1, 3: 5, 4: 1}, []string{"d2-1", "d4-1"}, 3, 0,
			[]string{"d3-1", "d3-2", "d2-1"}},
		{"tasks that draw at one difficulty take turns", map[int]int{3: 5, 4: 2}, []string{"d4-1", "d4-2"}, 3, 1,
			[]string{"d3-2", "d3-3", "d4-2"}},
		{"nothing draws: the picked stay", map[int]int{3: 5}, nil, 3, 0, []string{"d3-1", "d3-2", "d3-3"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			pool := drawnCell(test.counts, test.drawn...)
			picked := nearest(pool, test.difficulty, test.answers)
			if got := idsOf(withADrawnTask(picked, pool, test.difficulty, test.answers)); !slices.Equal(got, test.want) {
				t.Errorf("withADrawnTask(%v) = %v, want %v", idsOf(picked), got, test.want)
			}
		})
	}
}

// drawnPool is a pool with so many tasks at each difficulty from 1 on, the
// tasks whose flag is set carrying a picture.
func drawnPool(counts []int, drawn []bool) []Example {
	sizes := map[int]int{}
	for i, count := range counts {
		sizes[i+1] = count
	}
	pool := cell(sizes)
	for i := range pool {
		if i < len(drawn) && drawn[i] {
			pool[i].Picture = json.RawMessage(`{"kind":"ring","count":9}`)
		}
	}
	return pool
}

// Whatever the pool and wherever the turns have come to, the tasks a package
// shows are those nearest picks, but for the last of them where none of those
// draws, and no task is shown twice.
func TestATaskThatDrawsIsShownWheneverThePoolHasOne(t *testing.T) {
	t.Parallel()

	shown := func(counts []int, drawn []bool, difficulty, answers int) (picked, got []Example, pool []Example) {
		pool = drawnPool(counts, drawn)
		picked = nearest(pool, difficulty, answers)
		return picked, withADrawnTask(picked, pool, difficulty, answers), pool
	}
	properties := gopter.NewProperties(nil)
	for _, property := range []struct {
		name  string
		holds func(picked, got, pool []Example) bool
	}{
		{"as many tasks as nearest picks", func(picked, got, _ []Example) bool { return len(got) == len(picked) }},
		{"all but the last as nearest picks them", func(picked, got, _ []Example) bool {
			return len(got) == 0 || slices.Equal(idsOf(got[:len(got)-1]), idsOf(picked[:len(picked)-1]))
		}},
		{"a pick that draws is kept as it is", func(picked, got, _ []Example) bool {
			return !anyDraws(picked) || slices.Equal(idsOf(got), idsOf(picked))
		}},
		{"a task that draws is shown whenever the pool has one", func(_, got, pool []Example) bool {
			return anyDraws(got) == anyDraws(pool)
		}},
		{"no task is shown twice", func(_, got, _ []Example) bool {
			return len(slices.Compact(slices.Sorted(slices.Values(idsOf(got))))) == len(got)
		}},
	} {
		properties.Property(property.name, prop.ForAll(
			func(counts []int, drawn []bool, difficulty, answers int) bool {
				return property.holds(shown(counts, drawn, difficulty, answers))
			},
			gen.SliceOfN(5, gen.IntRange(0, 5)),
			gen.SliceOfN(25, gen.Bool()),
			gen.IntRange(1, 5),
			gen.IntRange(-1_000_000, 1_000_000),
		))
	}
	properties.TestingRun(t)
}

// solvedCell is a cell whose tasks of these ids carry a picture of their
// solution, and of those ids a picture of their own.
func solvedCell(counts map[int]int, solved, drawn []string) []Example {
	pool := drawnCell(counts, drawn...)
	for i := range pool {
		if slices.Contains(solved, pool[i].ID) {
			pool[i].SolutionPicture = json.RawMessage(`{"kind":"ring","count":8}`)
		}
	}
	return pool
}

func TestAPackageShowsATaskThatDrawsItsSolutionWhereThePoolHasOne(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name                string
		counts              map[int]int
		solved, drawn       []string
		difficulty, answers int
		want                []string
	}{
		{"one of those picked draws its solution: they stay", map[int]int{3: 5}, []string{"d3-2"}, nil, 3, 0,
			[]string{"d3-1", "d3-2", "d3-3"}},
		{"none of those picked draws its solution: the last gives way", map[int]int{3: 5}, []string{"d3-5"}, nil, 3, 0,
			[]string{"d3-1", "d3-2", "d3-5"}},
		{"a task picked for its picture stays: the one before it leaves", map[int]int{3: 5}, []string{"d3-5"},
			[]string{"d3-3"}, 3, 0, []string{"d3-1", "d3-3", "d3-5"}},
		{"a task that also draws takes the last place, where a task shown for its picture stood",
			map[int]int{3: 5, 5: 2}, []string{"d5-2"}, []string{"d5-1", "d5-2"}, 3, 0,
			[]string{"d3-1", "d3-2", "d5-2"}},
		{"every one picked has a picture: the last gives way", map[int]int{3: 5}, []string{"d3-5"},
			[]string{"d3-1", "d3-2", "d3-3"}, 3, 0, []string{"d3-1", "d3-2", "d3-5"}},
		{"the task that draws its solution nearest the difficulty", map[int]int{1: 1, 3: 5, 5: 1},
			[]string{"d1-1", "d5-1"}, nil, 4, 0, []string{"d3-1", "d3-2", "d5-1"}},
		{"tasks that draw their solution at one difficulty take turns", map[int]int{3: 5, 4: 2},
			[]string{"d4-1", "d4-2"}, nil, 3, 1, []string{"d3-2", "d3-3", "d4-2"}},
		{"nothing draws its solution: the picked stay", map[int]int{3: 5}, nil, nil, 3, 0,
			[]string{"d3-1", "d3-2", "d3-3"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			pool := solvedCell(test.counts, test.solved, test.drawn)
			picked := withADrawnTask(nearest(pool, test.difficulty, test.answers), pool, test.difficulty, test.answers)
			if got := idsOf(withADrawnSolution(picked, pool, test.difficulty, test.answers)); !slices.Equal(got, test.want) {
				t.Errorf("withADrawnSolution(%v) = %v, want %v", idsOf(picked), got, test.want)
			}
		})
	}
}

// Whatever the pool and wherever the turns have come to, the tasks a package
// shows once one that draws its solution is among them are those picked, in
// their order, but for one at most, which leaves for a task that draws its
// solution, last; a pick that draws its solution stays as it is, a task that
// draws its solution is shown whenever the pool has one, a task that draws
// stays among them whenever one was picked, and no task is shown twice.
func TestATaskThatDrawsItsSolutionIsShownWheneverThePoolHasOne(t *testing.T) {
	t.Parallel()

	shown := func(counts, drawing []int, difficulty, answers int) (picked, got, pool []Example) {
		pool = solvedPool(counts, drawing)
		picked = withADrawnTask(nearest(pool, difficulty, answers), pool, difficulty, answers)
		return picked, withADrawnSolution(picked, pool, difficulty, answers), pool
	}
	properties := gopter.NewProperties(nil)
	for _, property := range []struct {
		name  string
		holds func(picked, got, pool []Example) bool
	}{
		{"as many tasks as were picked", func(picked, got, _ []Example) bool { return len(got) == len(picked) }},
		{"those picked in their order, but one at most, which leaves for one last", func(picked, got, _ []Example) bool {
			return slices.Equal(idsOf(got), idsOf(picked)) || oneLeft(picked, got[:len(got)-1])
		}},
		{"a pick that draws its solution is kept as it is", func(picked, got, _ []Example) bool {
			return !anyDrawsItsSolution(picked) || slices.Equal(idsOf(got), idsOf(picked))
		}},
		{"a task that draws its solution is shown whenever the pool has one", func(_, got, pool []Example) bool {
			return len(got) == 0 || anyDrawsItsSolution(got) == anyDrawsItsSolution(pool)
		}},
		{"a task that draws stays shown whenever one was picked", func(picked, got, _ []Example) bool {
			return !anyDraws(picked) || anyDraws(got)
		}},
		{"no task is shown twice", func(_, got, _ []Example) bool {
			return len(slices.Compact(slices.Sorted(slices.Values(idsOf(got))))) == len(got)
		}},
	} {
		properties.Property(property.name, prop.ForAll(
			func(counts, drawing []int, difficulty, answers int) bool {
				return property.holds(shown(counts, drawing, difficulty, answers))
			},
			gen.SliceOfN(5, gen.IntRange(0, 5)),
			gen.SliceOfN(25, gen.IntRange(0, 3)),
			gen.IntRange(1, 5),
			gen.IntRange(-1_000_000, 1_000_000),
		))
	}
	properties.TestingRun(t)
}

// solvedPool is a pool with so many tasks at each difficulty from 1 on, each
// task drawn by a number from 0 to 3: its first bit gives it a picture, and
// its second a picture of its solution.
func solvedPool(counts, drawing []int) []Example {
	drawn := make([]bool, len(drawing))
	for i, bits := range drawing {
		drawn[i] = bits&1 != 0
	}
	pool := drawnPool(counts, drawn)
	for i := range pool {
		if i < len(drawing) && drawing[i]&2 != 0 {
			pool[i].SolutionPicture = json.RawMessage(`{"kind":"ring","count":8}`)
		}
	}
	return pool
}

// oneLeft says whether the tasks kept are those picked, in their order, with
// one of them left out.
func oneLeft(picked, kept []Example) bool {
	for leaving := range picked {
		if slices.Equal(idsOf(kept), idsOf(slices.Delete(slices.Clone(picked), leaving, leaving+1))) {
			return true
		}
	}
	return false
}

// A topic with nothing at the level of the brief is shown the level below; one
// with nothing at or below it, and a level there is not, are shown nothing.
func TestTheLevelBelowStandsInForAnEmptyOne(t *testing.T) {
	t.Parallel()

	c := &Content{examples: []Example{
		{ID: "t-12", Topic: "t", GradeLevel: rating.Grades12, Difficulty: 3},
		{ID: "t-34", Topic: "t", GradeLevel: rating.Grades34, Difficulty: 3},
		{ID: "u-56", Topic: "u", GradeLevel: rating.Grades56, Difficulty: 3},
	}}
	for _, test := range []struct {
		name  string
		topic string
		level rating.GradeLevel
		want  []string
	}{
		{"the level of the brief", "t", rating.Grades34, []string{"t-34"}},
		{"the level below, and only the one below", "t", rating.Grades56, []string{"t-34"}},
		{"nothing at or below the level", "u", rating.Grades12, nil},
		{"a level there is not", "t", "9-10", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := idsOf(c.examplesFor(test.topic, test.level, 3, 0)); !slices.Equal(got, test.want) {
				t.Errorf("examplesFor(%q, %s) = %v, want %v", test.topic, test.level, got, test.want)
			}
		})
	}
}

// The turns are fair: over any five answers in a row, each task of a cell of
// five is shown exactly three times, so that no example is a favourite.
func TestTheExamplesTakeFairTurns(t *testing.T) {
	t.Parallel()

	pool := cell(map[int]int{3: 5})
	properties := gopter.NewProperties(nil)
	properties.Property("over five answers in a row, each of five tasks is shown three times", prop.ForAll(
		func(start int) bool {
			shown := map[string]int{}
			for answers := start; answers < start+5; answers++ {
				for _, id := range idsOf(nearest(pool, 3, answers)) {
					shown[id]++
				}
			}
			for _, times := range shown {
				if times != 3 {
					return false
				}
			}
			return len(shown) == 5
		},
		gen.IntRange(-1_000_000, 1_000_000),
	))
	properties.TestingRun(t)
}
