package tutor_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

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
// dressed as the task after that one: its turn comes once the child has left
// that one behind.
func TestATaskWrittenAheadIsDressedAsTheTaskAfterTheCard(t *testing.T) {
	t.Parallel()

	for answers, want := range map[int]string{0: "", 1: "", 2: "trains", 5: "cats", 56: "trains"} {
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
			if got.Rationale != "Written ahead, before it is asked for. "+next.Rationale {
				t.Errorf("card %d, %d answers: the rationale is %q, want the next one's, %q, written ahead",
					card, answers, got.Rationale, next.Rationale)
			}
		}
	}
}

// A task written ahead is chosen as the next one would be now: the same
// topic, level, difficulty, goal, traps, prohibitions and setting, chosen by
// the same — only said to be written ahead. While the child works on the task
// on the card, both are dressed as the task after that one.
func TestATaskWrittenAheadIsChosenAsTheNextOneNow(t *testing.T) {
	t.Parallel()

	c := stepped()
	properties := gopter.NewProperties(nil)
	properties.Property("ahead is next, dressed as the task after the card's while its answer is awaited", prop.ForAll(
		func(s *seed, card int) bool {
			p := s.build()
			p.CurrentTask = onTheCard(card)
			next, nextMode, err := tutor.Next(p, c, tutor.Choice{})
			ahead, aheadMode, err2 := tutor.Ahead(p, c, tutor.Choice{})
			after, err3 := dressedOnceBehind(s, c, card == cardAwaited)
			if err != nil || err2 != nil || err3 != nil || nextMode != aheadMode {
				return false
			}
			if !strings.HasPrefix(ahead.Rationale, "Written ahead") || ahead.Setting != after {
				return false
			}
			ahead.Rationale = next.Rationale
			return reflect.DeepEqual(ahead, next)
		},
		genSeed(),
		gen.IntRange(cardEmpty, cardAnswered),
	))

	properties.TestingRun(t)
}

// dressedOnceBehind is what the next task of a seed's child is dressed in once
// the task on the card is behind the child, answered or skipped, when its
// answer is awaited, and now otherwise: what a task written ahead has to be
// dressed in.
func dressedOnceBehind(s *seed, c tutor.Catalog, awaited bool) (string, error) {
	p := s.build()
	if awaited {
		p.Ratings.Answers++
	}
	next, _, err := tutor.Next(p, c, tutor.Choice{})
	return next.Setting, err
}

// The task on the card counts as behind the child only for a task written
// ahead of its answer: once among the answers, and once among the tasks of its
// own topic, never of another's.
func TestATaskWrittenAheadCountsTheTaskOnTheCardBehindTheChild(t *testing.T) {
	t.Parallel()

	awaited := func(topic string) *profile.CurrentTask { return &profile.CurrentTask{ID: "task_on_card", Topic: topic} }
	for _, tc := range []struct {
		name                string
		card                *profile.CurrentTask
		ahead               bool
		answers, topicTasks int
	}{
		{"written ahead of the answer to a task of the topic", awaited("counting.gaps"), true, 8, 5},
		{"written ahead of the answer to a task of another topic", awaited("logic.ordering"), true, 8, 4},
		{"written ahead once the task on the card has its answer", &profile.CurrentTask{
			ID: "task_on_card", Topic: "counting.gaps", Answered: &profile.Given{Choice: "B"},
		}, true, 7, 4},
		{"written ahead with no task on the card", nil, true, 7, 4},
		{"asked for rather than written ahead", awaited("counting.gaps"), false, 7, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := settled(t)
			p.Ratings.Answers = 7
			p.Topics["counting.gaps"] = profile.Topic{Answers: 3, Skipped: 1, Traps: map[string]int{}}
			p.CurrentTask = tc.card

			answers, topicTasks := tutor.Behind(p, "counting.gaps", tc.ahead)
			if answers != tc.answers || topicTasks != tc.topicTasks {
				t.Errorf("Behind() = %d answers and %d tasks of the topic, want %d and %d",
					answers, topicTasks, tc.answers, tc.topicTasks)
			}
		})
	}
}

// What a task written ahead counts behind the child is what the next task
// counts once the child has answered the task on the card, so that its
// package names the same idea and shows the same reference tasks before that
// answer as after it. A skip leaves the tasks of every topic where they were
// too, since a task left behind is counted among them.
func TestATaskWrittenAheadCountsTheSameBeforeTheAnswerToTheCardAsAfter(t *testing.T) {
	t.Parallel()

	c := stepped()
	sealer, err := taskSeal()
	if err != nil {
		t.Fatalf("build the seal: %v", err)
	}
	properties := gopter.NewProperties(nil)
	properties.Property("the answer to the task on the card moves no count", prop.ForAll(
		func(s *seed, choice string) bool {
			p := s.build()
			task := onTheCardFromTheRule(t, p, c, sealer)
			before := countedAhead(p, c)
			answer := profile.Answered{TaskID: task.ID, Choice: choice, At: day.Add(2 * time.Minute)}
			if _, err := p.Record(answer, sealer, c.LevelsOf(task.Topic)); err != nil {
				t.Errorf("Record() error = %v, want nil", err)
				return false
			}
			return reflect.DeepEqual(countedAhead(p, c), before)
		},
		genSeed(),
		gen.OneConstOf("B", "C"),
	))
	properties.Property("a skip of it moves no topic's tasks", prop.ForAll(
		func(s *seed) bool {
			p := s.build()
			onTheCardFromTheRule(t, p, c, sealer)
			before := countedAhead(p, c)
			if _, skipped := p.Skip(day.Add(2 * time.Minute)); !skipped {
				return false
			}
			after := countedAhead(p, c)
			for topic, counted := range before {
				if after[topic].topicTasks != counted.topicTasks {
					return false
				}
			}
			return true
		},
		genSeed(),
	))
	properties.TestingRun(t)
}

// behind is what Behind counts.
type behind struct{ answers, topicTasks int }

// countedAhead is what a task written ahead counts behind the child, for each
// topic of the catalog.
func countedAhead(p *profile.Profile, c tutor.Catalog) map[string]behind {
	counted := make(map[string]behind, len(c.TopicIDs()))
	for _, topic := range c.TopicIDs() {
		answers, topicTasks := tutor.Behind(p, topic, true)
		counted[topic] = behind{answers, topicTasks}
	}
	return counted
}

// onTheCardFromTheRule puts on the card the task the rule sets next, sealed as
// the service seals a task, for the child to answer: C is right and B a slip.
func onTheCardFromTheRule(t *testing.T, p *profile.Profile, c tutor.Catalog, sealer profile.Sealer) *profile.CurrentTask {
	t.Helper()

	next, mode, err := tutor.Next(p, c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	return handedOut(t, p, &next, mode, sealer, day)
}
