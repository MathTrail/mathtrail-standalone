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
	p.Ask(&next, mode, "en", day)
	task, err := p.Issue(&profile.Written{
		Wording: "a task", Options: map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"},
		Hint: "hint", Fingerprint: "sketch", InstructionsVersion: "v",
	}, profile.TaskSecret{
		Answer:      "C",
		Distractors: map[string]profile.Distractor{"B": {Trap: next.TrapsToUse[0], Text: "a slip"}},
		Solution:    "the solution",
	}, sealer, day)
	if err != nil {
		t.Fatalf("Issue() error = %v, want nil", err)
	}
	return task
}
