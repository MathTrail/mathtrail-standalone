package content_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
)

// What the rule that chooses a child's next task asks of the catalogs: which
// topics a child of this grade can be given, how two traps are told apart when
// nothing else separates them, and what usually goes wrong in a topic.
//
// These three answers cross into a package that knows nothing about levels or
// reference tasks, so they are checked here, where both are in hand.

// The years a child can be in, and what they are grouped into.
func TestEveryGradeHasItsLevel(t *testing.T) {
	t.Parallel()

	for grade, want := range map[int]string{
		1: content.Level12, 2: content.Level12,
		3: content.Level34, 4: content.Level34,
		5: content.Level56, 6: content.Level56,
	} {
		got, known := content.LevelOf(grade)
		if !known || got != want {
			t.Errorf("LevelOf(%d) = %q, %v, want %q", grade, got, known, want)
		}
	}

	// A year outside school is told apart rather than rounded to the nearest.
	for _, grade := range []int{0, -1, 7, 100} {
		if got, known := content.LevelOf(grade); known {
			t.Errorf("LevelOf(%d) = %q, want no level at all", grade, got)
		}
	}
}

func TestTheTopicsOfAGrade(t *testing.T) {
	t.Parallel()

	c := loaded(t)
	for _, grade := range []int{1, 3, 5} {
		level, _ := content.LevelOf(grade)
		got := c.TopicsAt(grade)
		if len(got) == 0 {
			t.Fatalf("grade %d has no topics", grade)
		}

		// Catalog order, and every one of them offered at this level.
		var inOrder []string
		for _, topic := range c.Topics() {
			if topic.HasLevel(level) {
				inOrder = append(inOrder, topic.ID)
			}
		}
		if !slices.Equal(got, inOrder) {
			t.Errorf("TopicsAt(%d) = %v, want the catalog's own order %v", grade, got, inOrder)
		}
	}

	// A grade nobody is in has no topics rather than all of them.
	if got := c.TopicsAt(9); len(got) != 0 {
		t.Errorf("TopicsAt(9) = %v, want nothing", got)
	}
}

func TestTheTrapsInCatalogOrder(t *testing.T) {
	t.Parallel()

	c := loaded(t)
	got := c.TrapIDs()

	want := make([]string, 0, len(c.Traps()))
	for _, trap := range c.Traps() {
		want = append(want, trap.ID)
	}
	if !slices.Equal(got, want) {
		t.Errorf("TrapIDs() = %v, want %v", got, want)
	}
}

// What the reference tasks of a topic are usually built around, most frequent
// first. It is a count over the tasks themselves rather than a list kept
// beside the catalog, so it cannot fall out of date.
func TestTheTrapsOfATopicsReferenceTasks(t *testing.T) {
	t.Parallel()

	c := loaded(t)
	const topic, grade = "counting.gaps", 1
	level, _ := content.LevelOf(grade)

	got := c.ExampleTraps(topic, grade)
	if len(got) == 0 {
		t.Fatalf("%s has no traps among its reference tasks", topic)
	}

	counts := trapCounts(c, topic, level)
	if len(got) != len(counts) {
		t.Errorf("ExampleTraps() named %d traps, want %d", len(got), len(counts))
	}
	checkByFrequency(t, got, counts)

	// A topic nobody has, and a grade nobody is in, name nothing.
	if got := c.ExampleTraps("astrophysics.blackholes", grade); len(got) != 0 {
		t.Errorf("ExampleTraps() on a topic nobody has = %v, want nothing", got)
	}
	if got := c.ExampleTraps(topic, 9); len(got) != 0 {
		t.Errorf("ExampleTraps() at a grade nobody is in = %v, want nothing", got)
	}
}

// A child new to a topic of grades 5–6 is given the two traps most frequent
// among its reference tasks, and those tasks are labelled so that the two are
// the mistakes most typical of the topic. The pairs are pinned here, because
// one relabelled wrong option can tip a count, and a tie goes to the catalog's
// order, in which the traps added for these grades come last: a pair changes
// only when someone means it to. A topic whose tasks of 5–6 arrive names its
// pair here.
func TestANewChildOfGradesFiveAndSixIsGivenTheTopicsOwnTraps(t *testing.T) {
	t.Parallel()

	c := loaded(t)
	pinned := map[string][]string{
		"geometry.grid":       {"double_count", "missed_case"},
		"games.strategy":      {"ignored_condition", "first_move_assumed"},
		"logic.sets":          {"double_count", "answered_other_question"},
		"fractions.parts":     {"part_whole_swap", "answered_other_question"},
		"percent.basic":       {"percent_wrong_base", "part_whole_swap"},
		"ratio.sharing":       {"ratio_total_confusion", "wrong_operation"},
		"number.divisibility": {"remainder_vs_quotient", "off_by_one"},
	}
	for _, topic := range c.Topics() {
		want, isPinned := pinned[topic.ID]
		hasTasks := len(trapCounts(c, topic.ID, content.Level56)) > 0
		switch {
		case hasTasks && !isPinned:
			t.Errorf("%s has reference tasks of 5-6 and no pair of traps pinned here", topic.ID)
		case !hasTasks && isPinned:
			t.Errorf("%s has %v pinned here and no reference tasks of 5-6", topic.ID, want)
		case hasTasks:
			if got := c.ExampleTraps(topic.ID, 5); len(got) < 2 || !slices.Equal(got[:2], want) {
				t.Errorf("a child new to %s is given %v first, want %v", topic.ID, got[:min(2, len(got))], want)
			}
		}
	}
}

// trapCounts is how often each trap appears among the reference tasks of one
// topic at one level, counted the long way round.
func trapCounts(c *content.Content, topic, level string) map[string]int {
	counts := map[string]int{}
	examples := c.Examples()
	for i := range examples {
		task := &examples[i]
		if task.Topic != topic || task.GradeLevel != level {
			continue
		}
		for _, distractor := range task.Distractors {
			counts[distractor.Trap]++
		}
	}
	return counts
}

// checkByFrequency fails unless every trap named is one the reference tasks
// actually use, and the list runs from the most used downwards.
func checkByFrequency(t *testing.T, got []string, counts map[string]int) {
	t.Helper()

	for i, trap := range got {
		if counts[trap] == 0 {
			t.Errorf("ExampleTraps() named %q, which no reference task uses", trap)
		}
		if i > 0 && counts[got[i-1]] < counts[trap] {
			t.Errorf("ExampleTraps() = %v, want the most frequent first", got)
		}
	}
}
