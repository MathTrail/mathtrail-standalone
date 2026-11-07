package content_test

import (
	"slices"
	"testing"
)

// waitingForReferenceTasks are the topics of the catalog whose reference tasks
// are still being written. The tests that hold a topic to its reference tasks —
// their count, a solver template, a solution drawn at every level, a package
// that shows one — pass such a topic over, and a topic waits only while it has
// none, so it leaves this list with its first reference task.
var waitingForReferenceTasks = []string{"counting.figures", "geometry.cubes_paper"}

// waitsForReferenceTasks reports whether the topic's reference tasks are still
// being written.
func waitsForReferenceTasks(topic string) bool {
	return slices.Contains(waitingForReferenceTasks, topic)
}

// A topic waits for its reference tasks only while it is in the catalog and has
// none: one that has its first is held to them all, and a name that is no
// topic's would let nothing wait.
func TestATopicWaitsForItsReferenceTasksOnlyWhileItHasNone(t *testing.T) {
	t.Parallel()

	shipped := loaded(t)
	for _, topic := range waitingForReferenceTasks {
		if _, ok := shipped.Topic(topic); !ok {
			t.Errorf("%s waits for its reference tasks, and the catalog has no such topic", topic)
		}
	}
	examples := shipped.Examples()
	for i := range examples {
		if example := &examples[i]; waitsForReferenceTasks(example.Topic) {
			t.Errorf("%s waits for its reference tasks, and %s is one: take it off the list",
				example.Topic, example.ID)
		}
	}
}
