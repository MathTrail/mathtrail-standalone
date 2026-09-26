package content_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// What the rule that chooses a child's next task asks of the catalogs: every
// topic in catalog order, the levels each is taught at, how two traps are told
// apart when nothing else separates them, and what usually goes wrong in a
// topic at a level.
//
// These answers cross into a package that knows nothing about reference
// tasks, so they are checked here, where the tasks are in hand.

// The rule is written against the catalogs the binary ships, and the content
// is what answers it.
var _ tutor.Catalog = (*content.Content)(nil)

// Every topic, in the catalog's own order, and the levels each is taught at as
// the catalog lists them. A topic nobody has is taught nowhere.
func TestTheTopicsAndTheLevelsTheyAreTaughtAt(t *testing.T) {
	t.Parallel()

	c := loaded(t)
	var inOrder []string
	for _, topic := range c.Topics() {
		inOrder = append(inOrder, topic.ID)
		if got := c.LevelsOf(topic.ID); !slices.Equal(got, topic.GradeLevels) {
			t.Errorf("LevelsOf(%s) = %v, want %v", topic.ID, got, topic.GradeLevels)
		}
	}
	if got := c.TopicIDs(); !slices.Equal(got, inOrder) {
		t.Errorf("TopicIDs() = %v, want the catalog's own order %v", got, inOrder)
	}
	if got := c.LevelsOf("astrophysics.blackholes"); len(got) != 0 {
		t.Errorf("LevelsOf() on a topic nobody has = %v, want none", got)
	}

	// What is handed out is a copy: a caller that rewrites it rewrites nothing
	// the next caller sees.
	levels := c.LevelsOf("logic.ordering")
	levels[0] = "9-10"
	if again := c.LevelsOf("logic.ordering"); again[0] == "9-10" {
		t.Error("LevelsOf() handed out the catalog's own list")
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

// What the reference tasks of a topic are usually built around, at one level,
// most frequent first. It is a count over the tasks themselves rather than a
// list kept beside the catalog, so it cannot fall out of date.
func TestTheTrapsOfATopicsReferenceTasks(t *testing.T) {
	t.Parallel()

	c := loaded(t)
	const topic, level = "counting.gaps", rating.Grades12

	got := c.ExampleTraps(topic, level)
	if len(got) == 0 {
		t.Fatalf("%s has no traps among its reference tasks", topic)
	}

	counts := trapCounts(c, topic, level)
	if len(got) != len(counts) {
		t.Errorf("ExampleTraps() named %d traps, want %d", len(got), len(counts))
	}
	checkByFrequency(t, got, counts)

	// A topic nobody has, and a level there is not, name nothing.
	if got := c.ExampleTraps("astrophysics.blackholes", level); len(got) != 0 {
		t.Errorf("ExampleTraps() on a topic nobody has = %v, want nothing", got)
	}
	if got := c.ExampleTraps(topic, "9-10"); len(got) != 0 {
		t.Errorf("ExampleTraps() at a level there is not = %v, want nothing", got)
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
		"geometry.grid":               {"double_count", "missed_case"},
		"games.strategy":              {"ignored_condition", "first_move_assumed"},
		"logic.sets":                  {"double_count", "answered_other_question"},
		"fractions.parts":             {"part_whole_swap", "answered_other_question"},
		"percent.basic":               {"percent_wrong_base", "part_whole_swap"},
		"ratio.sharing":               {"ratio_total_confusion", "wrong_operation"},
		"number.divisibility":         {"remainder_vs_quotient", "off_by_one"},
		"logic.ordering":              {"ignored_condition", "reversed_relation"},
		"logic.knights_liars":         {"trusted_statement", "ignored_condition"},
		"combinatorics.enumeration":   {"ignored_condition", "missed_case"},
		"counting.gaps":               {"off_by_one", "wrong_operation"},
		"time.clocks":                 {"ignored_condition", "stopped_early"},
		"time.calendar":               {"ignored_condition", "answered_other_question"},
		"arithmetic.tricks":           {"stopped_early", "wrong_operation"},
		"pigeonhole.basic":            {"best_case_not_worst", "off_by_one"},
		"parity.alternation":          {"wrong_parity", "ignored_condition"},
		"algorithms.weighing_pouring": {"missed_case", "ignored_condition"},
	}
	for _, topic := range c.Topics() {
		want, isPinned := pinned[topic.ID]
		hasTasks := len(trapCounts(c, topic.ID, rating.Grades56)) > 0
		switch {
		case hasTasks && !isPinned:
			t.Errorf("%s has reference tasks of 5-6 and no pair of traps pinned here", topic.ID)
		case !hasTasks && isPinned:
			t.Errorf("%s has %v pinned here and no reference tasks of 5-6", topic.ID, want)
		case hasTasks:
			if got := c.ExampleTraps(topic.ID, rating.Grades56); len(got) < 2 || !slices.Equal(got[:2], want) {
				t.Errorf("a child new to %s is given %v first, want %v", topic.ID, got[:min(2, len(got))], want)
			}
		}
	}
}

// trapCounts is how often each trap appears among the reference tasks of one
// topic at one level, counted the long way round.
func trapCounts(c *content.Content, topic string, level rating.GradeLevel) map[string]int {
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
