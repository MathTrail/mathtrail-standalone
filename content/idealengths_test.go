package content_test

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
)

var update = flag.Bool("update", false, "record in testdata/idea-lengths.json the lists of ideas that grew or are new")

// ideaLengths is where the length of every topic's list of ideas at each level
// is kept, by the topic's id and then by the level.
const ideaLengths = "testdata/idea-lengths.json"

// A topic's list of ideas grows and never grows shorter. A family's file keeps
// the places of its round in the list, and a list made shorter, or taken away,
// gives those places other ideas, which is a new version of the file. So a
// list shorter than the snapshot holds, or gone, is refused even with -update,
// and one that grew, or a new one, is recorded with -update, its diff read
// before it is kept. What this cannot see is an idea taken out and another
// written in: the length stays, and the places after it move.
func TestAListOfIdeasNeverGrowsShorter(t *testing.T) {
	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	now := lengthsOf(loaded)
	kept := keptLengths(t)

	wantNoneShorter(t, now, kept)
	grown := grownSince(now, kept)
	switch {
	case len(grown) > 0 && !*update:
		t.Errorf("these lists of ideas grew or are new, %v: run the test with -update to record them", grown)
	case *update && !t.Failed():
		recordLengths(t, now)
	}
}

// lengthsOf are the lengths of every topic's lists of ideas, by the topic's id
// and then by the level.
func lengthsOf(loaded *content.Content) map[string]map[string]int {
	now := map[string]map[string]int{}
	for _, topic := range loaded.TopicIDs() {
		for _, level := range loaded.LevelsOf(topic) {
			if now[topic] == nil {
				now[topic] = map[string]int{}
			}
			now[topic][string(level)] = loaded.IdeaCount(topic, level)
		}
	}
	return now
}

// wantNoneShorter holds every list the snapshot keeps to being there and to
// being no shorter than it was.
func wantNoneShorter(t *testing.T, now, kept map[string]map[string]int) {
	t.Helper()

	for topic, levels := range kept {
		for level, was := range levels {
			length, there := now[topic][level]
			switch {
			case !there:
				t.Errorf("%s at %s has no list of ideas any more, and it held %d: a family's file may hold places in it", topic, level, was)
			case length < was:
				t.Errorf("%s at %s holds %d ideas, fewer than the %d recorded: a family's file may hold places past the end", topic, level, length, was)
			}
		}
	}
}

// keptLengths are the lengths the snapshot holds. A snapshot gone is a
// failure, -update or not: an empty one would let any list grow shorter.
func keptLengths(t *testing.T) map[string]map[string]int {
	t.Helper()

	written, err := os.ReadFile(ideaLengths)
	if errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("%s is gone, and with it what no list may grow shorter than: restore it", ideaLengths)
	}
	if err != nil {
		t.Fatalf("read %s: %v", ideaLengths, err)
	}
	var kept map[string]map[string]int
	if err := json.Unmarshal(written, &kept); err != nil {
		t.Fatalf("read %s: %v", ideaLengths, err)
	}
	return kept
}

// grownSince are the lists, by topic and level, that are longer now than the
// snapshot holds or that it does not hold, in order.
func grownSince(now, kept map[string]map[string]int) []string {
	var grown []string
	for topic, levels := range now {
		for level, length := range levels {
			if was, there := kept[topic][level]; !there || length > was {
				grown = append(grown, fmt.Sprintf("%s at %s, %d", topic, level, length))
			}
		}
	}
	slices.Sort(grown)
	return grown
}

// recordLengths writes the lengths of now as the snapshot.
func recordLengths(t *testing.T, now map[string]map[string]int) {
	t.Helper()

	written, err := json.MarshalIndent(now, "", "\t")
	if err != nil {
		t.Fatalf("write %s: %v", ideaLengths, err)
	}
	if err := os.WriteFile(ideaLengths, append(written, '\n'), 0o600); err != nil {
		t.Fatalf("write %s: %v", ideaLengths, err)
	}
}
