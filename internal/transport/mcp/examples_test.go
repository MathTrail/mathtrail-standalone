package mcpserver_test

import (
	"encoding/json"
	"maps"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// The reference tasks of a package turn with the tasks of its topic the child
// has left behind, so that answers to other topics coming between do not bring
// a topic's last ones back. Ten answers in all since the race was asked for,
// one of them to the race's own topic, bring the next turn of its reference
// tasks, where a turn by every answer would come round to the same ones.
func TestATopicComingBackTenAnswersLaterIsShownOtherReferenceTasks(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	race := askForTheRace(t, session, kept)
	before := exampleQuestions(t, fetchPackage(t, session, race.ID))

	changeKept(t, kept, func(p *profile.Profile) {
		ordering, gaps := p.Topics["logic.ordering"], p.Topics["counting.gaps"]
		ordering.Answers++
		gaps.Answers += 9
		p.Topics["logic.ordering"], p.Topics["counting.gaps"] = ordering, gaps
		p.Ratings.Answers += 10
	})
	after := exampleQuestions(t, fetchPackage(t, session, race.ID))
	if maps.Equal(before, after) {
		t.Errorf("ten answers later, one of them on its topic, the race is shown %q again, want other reference tasks",
			slices.Sorted(maps.Keys(after)))
	}
}

// A task still being written ahead when the child skips the task on the card
// is waited for from then on, and the model may fetch its package again to
// finish it. The skip leaves the task on the card behind among the tasks of
// its topic, as an answer would, so the package shows the reference tasks it
// showed before.
func TestATaskWrittenAheadIsShownTheSameReferenceTasksOnceTheCardIsSkipped(t *testing.T) {
	t.Parallel()

	kept := racer(t)
	_, session := lesson(t, kept)
	raceHandedOut(t, session, kept)
	before := exampleQuestions(t, prepared(t, session, aheadChoice))
	ahead := aheadOpen(t, kept)

	call(t, session, "next_task", map[string]any{"language": "en"})
	if p, _ := loadKept(t, kept); p.OpenRequest == nil || p.OpenRequest.ID != ahead.ID || p.OpenRequest.Ahead ||
		p.CurrentTask != nil {
		t.Fatalf("the profile holds %+v and %+v, want the request ahead waited for and the race skipped",
			p.OpenRequest, p.CurrentTask)
	}
	after := exampleQuestions(t, fetchPackage(t, session, ahead.ID))
	if !maps.Equal(before, after) {
		t.Errorf("once the race is skipped the package shows %q, want the reference tasks it showed before, %q",
			slices.Sorted(maps.Keys(after)), slices.Sorted(maps.Keys(before)))
	}
}

// exampleQuestions are the questions of the reference tasks the package in a
// tool's words shows, whatever their order: which of them a package shows is
// what turns, and their order follows from it.
func exampleQuestions(t *testing.T, words string) map[string]bool {
	t.Helper()

	shown := map[string]bool{}
	for _, question := range exampleQuestionsIn(t, words) {
		shown[question] = true
	}
	return shown
}

// exampleQuestionsIn are the questions of the reference tasks the package in a
// tool's words shows, in the order it shows them.
func exampleQuestionsIn(t *testing.T, words string) []string {
	t.Helper()

	var examples []struct {
		Question string `json:"question"`
	}
	if err := json.Unmarshal(packageIn(t, words)["examples"], &examples); err != nil || len(examples) == 0 {
		t.Fatalf("the package shows the reference tasks %v (%v), want some", examples, err)
	}
	questions := make([]string, 0, len(examples))
	for _, example := range examples {
		questions = append(questions, example.Question)
	}
	return questions
}
