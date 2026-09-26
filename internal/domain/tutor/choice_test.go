package tutor_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The model may ask for something other than what the rule suggests, and it
// has to say why. What it asks for is used; what the rule would have done is
// kept beside it, so that afterwards the two can be told apart rather than
// guessed at.

func TestTheModelMayChooseTheTopic(t *testing.T) {
	t.Parallel()

	p := child(t)
	asked := tutor.Choice{Topic: "time.clocks", Reason: "the child asked for clocks"}

	got, mode, err := tutor.Next(p, threeTopics(), asked)
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.TargetConcept != "time.clocks" {
		t.Errorf("topic = %q, want the one the model asked for", got.TargetConcept)
	}
	if mode != profile.TutorLLM {
		t.Errorf("mode = %q, want %q", mode, profile.TutorLLM)
	}

	// The traps follow the model's topic, not the rule's.
	if want := []string{"wrong_operation", "off_by_one"}; !slices.Equal(got.TrapsToUse, want) {
		t.Errorf("traps = %v, want the traps of the topic the model chose, %v", got.TrapsToUse, want)
	}
	for _, say := range []string{"the child asked for clocks", "Rule:", "counting.gaps"} {
		if !strings.Contains(got.Rationale, say) {
			t.Errorf("rationale = %q, want it to carry %q", got.Rationale, say)
		}
	}
}

// The goal stays the rule's. A brief may therefore say "reinforce" about a
// topic the child has never practised: that records "a failure just happened
// and something else was asked for", which is a fact about the child.
func TestTheGoalStaysTheRules(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Ratings.ConsecutiveFailures = 1
	p.Recent = []profile.Answer{{
		AnsweredAt: profile.At(day),
		Difficulty: 3,
		Pace:       profile.PaceNormal,
		TaskID:     "tsk_1",
		Topic:      "counting.gaps",
		Chosen:     "A",
		Trap:       "off_by_one",
	}}

	got, _, err := tutor.Next(p, threeTopics(), tutor.Choice{Topic: "time.clocks", Reason: "something lighter"})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.PedagogicalGoal != profile.GoalReinforce {
		t.Errorf("goal = %q, want the rule's %q even where the model chose the topic",
			got.PedagogicalGoal, profile.GoalReinforce)
	}
	if got.TargetConcept != "time.clocks" {
		t.Errorf("topic = %q, want the model's", got.TargetConcept)
	}
}

// A difficulty of the model's own is used as given: the corridor is a
// recommendation, and a deliberate step outside it is what a tutor sometimes
// does.
func TestTheModelMayStepOutsideTheCorridor(t *testing.T) {
	t.Parallel()

	p := child(t)
	recommended := brief(t, p, threeTopics()).Difficulty

	got, mode, err := tutor.Next(p, threeTopics(), tutor.Choice{Difficulty: 5, Reason: "a stretch on purpose"})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.Difficulty != 5 {
		t.Errorf("difficulty = %d, want the model's 5", got.Difficulty)
	}
	if recommended == 5 {
		t.Fatal("the rule already recommends 5, so this case proves nothing")
	}
	if mode != profile.TutorLLM {
		t.Errorf("mode = %q, want %q", mode, profile.TutorLLM)
	}
	// And the rule's own recommendation is still on the record.
	if !strings.Contains(got.Rationale, "a stretch on purpose") {
		t.Errorf("rationale = %q, want the model's reason", got.Rationale)
	}
}

// What cannot be built is refused rather than half-built: a topic the catalog
// does not have leaves the examples, the traps and the corridor standing on
// nothing.
func TestAChoiceNoBriefCanBeBuiltFromIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		choice tutor.Choice
	}{
		{name: "a topic nobody has", choice: tutor.Choice{Topic: "astrophysics.blackholes"}},
		{name: "a difficulty below the scale", choice: tutor.Choice{Difficulty: 0 - 1}},
		{name: "a difficulty above the scale", choice: tutor.Choice{Difficulty: 6}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, mode, err := tutor.Next(child(t), threeTopics(), tc.choice)
			if err == nil {
				t.Fatalf("Next() error = nil, want a refusal; brief = %+v", got)
			}
			if mode != "" {
				t.Errorf("mode = %q, want nothing alongside a refusal", mode)
			}
		})
	}
}

// A model that asks for exactly what the rule suggested still asked: the mode
// records who chose, not whether the two agreed.
func TestAskingForWhatTheRuleSaidIsStillTheModelChoosing(t *testing.T) {
	t.Parallel()

	p := child(t)
	rule := brief(t, p, threeTopics())

	_, mode, err := tutor.Next(p, threeTopics(), tutor.Choice{Topic: rule.TargetConcept, Reason: "agreed"})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if mode != profile.TutorLLM {
		t.Errorf("mode = %q, want %q", mode, profile.TutorLLM)
	}
}

// When the model sets another topic, the rationale still gives the rule's own
// account: the difficulty the rule would have set for the topic it suggested,
// not the one the model's topic is given. The two are compared afterwards, and
// a rule's account taken from the model's topic would leave nothing to compare.
func TestTheRulesAccountIsOfItsOwnTopic(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Topics["counting.gaps"] = profile.Topic{Delta: -1.5}
	p.Topics["time.clocks"] = profile.Topic{Delta: 1.5}

	got, _, err := tutor.Next(p, threeTopics(), tutor.Choice{Topic: "time.clocks", Reason: "clocks today"})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.Difficulty != 4 || !strings.Contains(got.Rationale, "counting.gaps") ||
		!strings.Contains(got.Rationale, "Difficulty 1 is") ||
		!strings.Contains(got.Rationale, "Difficulty 4 is the one its corridor recommends") {
		t.Errorf("difficulty %d, rationale %q, want 4 for the model's topic, said where it came from, "+
			"and the rule's own 1 for counting.gaps",
			got.Difficulty, got.Rationale)
	}
}

// The model's reason is one sentence of the rationale, ended once: with its own
// mark when it wrote one, and with a full stop when it did not.
func TestTheModelsReasonIsEndedOnce(t *testing.T) {
	t.Parallel()

	for _, reason := range []string{
		"clocks today", "The child asked for clocks.", "Clocks again?", "他需要复习时钟。", "  ",
	} {
		got, _, err := tutor.Next(child(t), threeTopics(), tutor.Choice{Topic: "time.clocks", Reason: reason})
		if err != nil {
			t.Fatalf("Next() error = %v, want nil", err)
		}
		if strings.Contains(got.Rationale, "..") || strings.Contains(got.Rationale, "?.") ||
			strings.Contains(got.Rationale, "。.") || strings.Contains(got.Rationale, ": .") {
			t.Errorf("rationale %q, want every sentence of it ended once", got.Rationale)
		}
	}
}
