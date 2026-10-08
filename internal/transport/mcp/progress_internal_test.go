package mcpserver

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
)

// The topics of the progress are told as the child met them, each with what it
// is. A child who has met every topic within reach is told nothing of topics
// still to meet, not an empty list of them; a child who has met none is told
// so, and then the topics there are to meet.
func TestTheTopicsAreToldAsTheChildMetThem(t *testing.T) {
	t.Parallel()

	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	s := &Service{content: loaded}

	for _, tc := range []struct {
		name   string
		topics []progress.Topic
		want   string
	}{
		{"every topic within reach met", []progress.Topic{
			{ID: "logic.ordering", Answers: 2, Correct: 1, InReach: true},
			{ID: "counting.gaps", Skipped: 1, InReach: true},
		}, "Topics: Ordering (Restore an order from comparisons: who stands behind whom, who is older or taller), " +
			"1 of 2 right; Gaps and boundaries (Count gaps versus objects: saw cuts in a log, trees in a row, floors of " +
			"a building), no answers yet, 1 skipped."},
		{"none met yet", []progress.Topic{
			{ID: "logic.ordering", InReach: true}, {ID: "counting.gaps", InReach: true},
		}, "No topic has been answered yet. Not met yet, and within reach now: Ordering, Gaps and boundaries."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := s.topicsText(tc.topics); got != tc.want {
				t.Errorf("topicsText() =\n%q\nwant\n%q", got, tc.want)
			}
		})
	}
}
