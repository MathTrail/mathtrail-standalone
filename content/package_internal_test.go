package content

import (
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
