package tutor_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// What is on the card when a task is written ahead: nothing yet, a task the
// child is working on, or one that has its answer.
const (
	cardEmpty = iota
	cardAwaited
	cardAnswered
)

// onTheCard is the task on the card, as card says it stands.
func onTheCard(card int) *profile.CurrentTask {
	switch card {
	case cardAwaited:
		return &profile.CurrentTask{ID: "task_on_card"}
	case cardAnswered:
		return &profile.CurrentTask{ID: "task_on_card", Answered: &profile.Given{Choice: "B"}}
	}
	return nil
}

// aheadBrief is the brief for the task after the one on the card.
func aheadBrief(t *testing.T, p *profile.Profile, c tutor.Catalog) profile.Brief {
	t.Helper()

	built, _, err := tutor.Ahead(p, c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Ahead() error = %v, want nil", err)
	}
	return built
}

// A task written ahead while the child works on the task on the card is
// dressed in the interest after the one that task was dressed in: its turn
// comes once the child has answered that one.
func TestATaskWrittenAheadIsDressedOneInterestOn(t *testing.T) {
	t.Parallel()

	for answers, want := range map[int]string{0: "trains", 1: "cats", 2: "trains", 57: "cats"} {
		p := child(t)
		p.Ratings.Answers = answers
		p.CurrentTask = onTheCard(cardAwaited)

		if got := aheadBrief(t, p, threeTopics()); got.Setting != want {
			t.Errorf("after %d answers the task ahead is dressed in %q, want %q", answers, got.Setting, want)
		}
	}
}

// A task written ahead while the child works on the task on the card counts
// the trial series from that task, so the last of the series is the last it
// names; the task after it says so.
func TestATaskWrittenAheadCountsTheTrialFromTheTaskOnTheCard(t *testing.T) {
	t.Parallel()

	for answers, want := range map[int]string{
		0: "trial series, task 2 of 5",
		3: "trial series, task 5 of 5",
		4: "the last task of the trial series is on the card",
	} {
		p := child(t)
		p.Ratings.Answers = answers
		p.CurrentTask = onTheCard(cardAwaited)

		got := aheadBrief(t, p, threeTopics())
		if !strings.Contains(got.Rationale, want) ||
			!strings.HasPrefix(got.Rationale, "Written ahead, before the answer to the task on the card.") {
			t.Errorf("after %d answers the rationale is %q, want it written before the answer and saying %q",
				answers, got.Rationale, want)
		}
	}
}

// A task written ahead once the task on the card has its answer — or with no
// task on the card at all — is the next task: its answer is behind the child
// already, so the task ahead is dressed and counted as the next one is.
func TestATaskWrittenAheadOnceTheCardIsAnsweredIsTheNextOne(t *testing.T) {
	t.Parallel()

	for _, card := range []int{cardAnswered, cardEmpty} {
		for answers, want := range map[int]string{0: "cats", 3: "trains"} {
			p := child(t)
			p.Ratings.Answers = answers
			p.CurrentTask = onTheCard(card)

			next := brief(t, p, threeTopics())
			got := aheadBrief(t, p, threeTopics())
			if got.Setting != want || got.Setting != next.Setting {
				t.Errorf("card %d, %d answers: the task ahead is dressed in %q, want %q, as the next one is",
					card, answers, got.Setting, want)
			}
			if got.Rationale != "Written ahead, before the child asks for it. "+next.Rationale {
				t.Errorf("card %d, %d answers: the rationale is %q, want the next one's, %q, written ahead",
					card, answers, got.Rationale, next.Rationale)
			}
		}
	}
}

// A task written ahead is chosen as the next one would be now: the same
// topic, level, difficulty, goal, traps and prohibitions, chosen by the same —
// only said to be written ahead, and, while the child works on the task on the
// card, dressed one interest on.
func TestATaskWrittenAheadIsChosenAsTheNextOneNow(t *testing.T) {
	t.Parallel()

	c := stepped()
	properties := gopter.NewProperties(nil)
	properties.Property("ahead is next, one interest on while the card's answer is awaited", prop.ForAll(
		func(s *seed, card int) bool {
			p := s.build()
			p.CurrentTask = onTheCard(card)
			next, nextMode, err := tutor.Next(p, c, tutor.Choice{})
			ahead, aheadMode, err2 := tutor.Ahead(p, c, tutor.Choice{})
			if err != nil || err2 != nil || nextMode != aheadMode {
				return false
			}
			if !strings.HasPrefix(ahead.Rationale, "Written ahead") || ahead.Setting != dressedAhead(p, card == cardAwaited) {
				return false
			}
			ahead.Rationale, ahead.Setting = next.Rationale, next.Setting
			return reflect.DeepEqual(ahead, next)
		},
		genSeed(),
		gen.IntRange(cardEmpty, cardAnswered),
	))

	properties.TestingRun(t)
}

// dressedAhead is the interest a task written ahead is dressed in: the one
// after the interest a task now would be dressed in while the answer to the
// task on the card is awaited, and that one otherwise.
func dressedAhead(p *profile.Profile, awaited bool) string {
	interests := p.Student.Interests
	if len(interests) == 0 {
		return ""
	}
	answers := p.Ratings.Answers
	if awaited {
		answers++
	}
	return interests[answers%len(interests)]
}
