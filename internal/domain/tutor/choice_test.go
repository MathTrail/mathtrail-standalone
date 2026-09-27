package tutor_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
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

	p := settled(t)
	failedAt(p, "counting.gaps", 1)

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
// does. Named alone, it stays at the level the corridor recommends.
func TestTheModelMayStepOutsideTheCorridor(t *testing.T) {
	t.Parallel()

	p := child(t)
	recommended := brief(t, p, threeTopics())

	got, mode, err := tutor.Next(p, threeTopics(), tutor.Choice{Difficulty: 5, Reason: "a stretch on purpose"})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.Difficulty != 5 || got.GradeLevel != recommended.GradeLevel {
		t.Errorf("difficulty %d of %s, want the model's 5 at the recommended level %s",
			got.Difficulty, got.GradeLevel, recommended.GradeLevel)
	}
	if recommended.Difficulty == 5 {
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

// A level named alone is set at the difficulty its corridor recommends there:
// the model decided the level, and the ratings still say how hard a task of it
// the child can take.
func TestTheModelMayChooseTheLevel(t *testing.T) {
	t.Parallel()

	got, mode, err := tutor.Next(child(t), threeTopics(),
		tutor.Choice{GradeLevel: rating.Grades34, Reason: "the child is bored"})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.GradeLevel != rating.Grades34 || got.Difficulty != 1 || mode != profile.TutorLLM {
		t.Errorf("difficulty %d of %s by %q, want difficulty 1 of %s by the model",
			got.Difficulty, got.GradeLevel, mode, rating.Grades34)
	}
	if !strings.Contains(got.Rationale, "That is difficulty 1 of grades 3-4") {
		t.Errorf("rationale = %q, want it to say which task the level came to", got.Rationale)
	}
}

// Any topic of the catalog may be asked for, within reach or not: the model's
// reason is what makes the exception, and the topic is set at its easiest
// point when that is the nearest to the corridor.
func TestTheModelMayChooseATopicOutOfReach(t *testing.T) {
	t.Parallel()

	got, mode, err := tutor.Next(child(t), withAnOlderTopic(),
		tutor.Choice{Topic: "percent.basic", Reason: "the child asked for percentages"})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.TargetConcept != "percent.basic" || got.GradeLevel != rating.Grades56 || got.Difficulty != 1 ||
		mode != profile.TutorLLM {
		t.Errorf("brief = %s at difficulty %d of %s by %q, want percent.basic at its easiest point, by the model",
			got.TargetConcept, got.Difficulty, got.GradeLevel, mode)
	}
}

// What cannot be built is refused rather than half-built, with every rule the
// choice breaks named by the argument it came in: a topic the catalog does not
// have, or a level the topic is not taught at, leaves the examples, the traps
// and the limits standing on nothing, and a difficulty off the scale is no
// difficulty.
func TestAChoiceNoBriefCanBeBuiltFromIsRefused(t *testing.T) {
	t.Parallel()

	onlyLower := threeTopics()
	onlyLower.levels = map[string][]rating.GradeLevel{"counting.gaps": {rating.Grades12}}
	for _, tc := range []struct {
		name    string
		catalog catalog
		choice  tutor.Choice
		fields  []string
		says    string
	}{
		{
			name: "a topic nobody has", catalog: threeTopics(),
			choice: tutor.Choice{Topic: "astrophysics.blackholes", Reason: "stars"},
			fields: []string{"topic"}, says: "topic of the catalog",
		},
		{
			name: "a level there is not", catalog: threeTopics(),
			choice: tutor.Choice{GradeLevel: "7-8", Reason: "older"},
			fields: []string{"grade_level"}, says: "one of 1-2, 3-4, 5-6",
		},
		{
			name: "a level the topic is not taught at", catalog: withAnOlderTopic(),
			choice: tutor.Choice{Topic: "percent.basic", GradeLevel: rating.Grades12, Reason: "easier"},
			fields: []string{"grade_level"}, says: "taught at: 5-6",
		},
		{
			// Named alone, a level is set on the rule's topic, and the
			// refusal names that topic so that the model can pick another.
			name: "a level the rule's topic is not taught at", catalog: onlyLower,
			choice: tutor.Choice{GradeLevel: rating.Grades56, Reason: "harder"},
			fields: []string{"grade_level"}, says: "the rule's topic, counting.gaps, is taught at: 1-2",
		},
		{
			name: "a difficulty below the scale", catalog: threeTopics(),
			choice: tutor.Choice{Difficulty: 0 - 1, Reason: "easier"},
			fields: []string{"difficulty"}, says: "from 1 to 5",
		},
		{
			name: "a difficulty above the scale", catalog: threeTopics(),
			choice: tutor.Choice{Difficulty: 6, Reason: "harder"},
			fields: []string{"difficulty"}, says: "from 1 to 5",
		},
		{
			name: "everything at once", catalog: threeTopics(),
			choice: tutor.Choice{Topic: "astrophysics.blackholes", GradeLevel: "7-8", Difficulty: 9},
			fields: []string{"topic", "grade_level", "difficulty", "reason"}, says: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, mode, err := tutor.Next(child(t), tc.catalog, tc.choice)
			refused := refusal(t, err)
			if mode != "" || got.TargetConcept != "" {
				t.Errorf("Next() = %+v by %q, want nothing alongside a refusal", got, mode)
			}
			if fields := fieldsOf(refused); !slices.Equal(fields, tc.fields) {
				t.Errorf("the refusal names %v, want %v", fields, tc.fields)
			}
			if !strings.Contains(refused.Error(), tc.says) {
				t.Errorf("the refusal says %q, want it to say %q", refused.Error(), tc.says)
			}
		})
	}
}

// A choice of the model's comes with a reason, of a sentence or two: it is
// what makes the exception, and it goes into the brief beside the rule's own
// account. A reason alone chooses nothing, and the rule sets the task.
func TestAChoiceNeedsAReason(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		reason string
		says   string
	}{
		{"no reason", "", "is required"},
		{"a reason of spaces", "   ", "is required"},
		{"a reason past the limit", strings.Repeat("я", tutor.MaxReason+1), "at most 300 characters, not 301"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := tutor.Next(child(t), threeTopics(), tutor.Choice{Topic: "time.clocks", Reason: tc.reason})
			refused := refusal(t, err)
			if fields := fieldsOf(refused); !slices.Equal(fields, []string{"reason"}) || !strings.Contains(refused.Error(), tc.says) {
				t.Errorf("the refusal is %q, want the reason refused: %s", refused.Error(), tc.says)
			}
		})
	}

	longest, _, err := tutor.Next(child(t), threeTopics(),
		tutor.Choice{Topic: "time.clocks", Reason: strings.Repeat("я", tutor.MaxReason)})
	if err != nil || !strings.Contains(longest.Rationale, strings.Repeat("я", tutor.MaxReason)) {
		t.Errorf("Next() with the longest reason = %q, %v, want the reason kept whole", longest.Rationale, err)
	}

	// A reason alone, of any length, is dropped rather than held to a limit it
	// never reaches the brief to need.
	for _, reason := range []string{"the child is tired", strings.Repeat("tired ", tutor.MaxReason)} {
		alone, mode, err := tutor.Next(child(t), threeTopics(), tutor.Choice{Reason: reason})
		if err != nil || mode != profile.TutorRule || strings.Contains(alone.Rationale, "tired") {
			t.Errorf("a reason alone gives %q by %q, %v, want the rule's own brief", alone.Rationale, mode, err)
		}
	}
}

// A refusal names the argument and the rule, never what the model wrote: the
// model has its own words, and they need not travel any further.
func TestARefusalNeverRepeatsTheChoice(t *testing.T) {
	t.Parallel()

	topic, reason := "planets.rings-of-saturn", strings.Repeat("Saturn has rings. ", 20)
	_, _, err := tutor.Next(child(t), threeTopics(), tutor.Choice{Topic: topic, Reason: reason})
	refused := refusal(t, err)
	for _, said := range []string{topic, "Saturn"} {
		if strings.Contains(refused.Error(), said) {
			t.Errorf("the refusal %q repeats %q", refused.Error(), said)
		}
	}
}

// refusal is the choice error Next returned, or the end of the test.
func refusal(t *testing.T, err error) *tutor.ChoiceError {
	t.Helper()

	var refused *tutor.ChoiceError
	if !errors.As(err, &refused) {
		t.Fatalf("Next() error = %v, want a refusal of the choice", err)
	}
	return refused
}

// fieldsOf are the arguments a refusal names, in order.
func fieldsOf(refused *tutor.ChoiceError) []string {
	fields := make([]string, 0, len(refused.Problems))
	for _, problem := range refused.Problems {
		fields = append(fields, problem.Field)
	}
	return fields
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
// account: the point the rule would have set for the topic it suggested, not
// the one the model's topic is given. The two are compared afterwards, and a
// rule's account taken from the model's topic would leave nothing to compare.
func TestTheRulesAccountIsOfItsOwnTopic(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Topics["counting.gaps"] = profile.Topic{Delta: -1}
	p.Topics["time.clocks"] = profile.Topic{Delta: 2}

	got, _, err := tutor.Next(p, threeTopics(), tutor.Choice{Topic: "time.clocks", Reason: "clocks today"})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.Difficulty != 4 || got.GradeLevel != rating.Grades12 || !strings.Contains(got.Rationale, "counting.gaps") ||
		!strings.Contains(got.Rationale, "Difficulty 1 of grades 1-2 is") ||
		!strings.Contains(got.Rationale, "That is difficulty 4 of grades 1-2") {
		t.Errorf("difficulty %d of %s, rationale %q, want difficulty 4 of 1-2 for the model's topic, said where "+
			"it came from, and the rule's own difficulty 1 for counting.gaps",
			got.Difficulty, got.GradeLevel, got.Rationale)
	}
}

// The model's reason is one sentence of the rationale, ended once: with its own
// mark when it wrote one, and with a full stop when it did not.
func TestTheModelsReasonIsEndedOnce(t *testing.T) {
	t.Parallel()

	for _, reason := range []string{
		"clocks today", "The child asked for clocks.", "Clocks again?", "他需要复习时钟。",
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

// The model writes the choice, so whatever it writes, the rule either builds a
// brief that stands on the catalog or refuses the choice by the rules it
// breaks. It never falls over, and it never builds half a brief.
func FuzzChoice(f *testing.F) {
	f.Add("time.clocks", "3-4", 2, "the child asked for clocks")
	f.Add("", "5-6", 0, "harder today")
	f.Add("percent.basic", "1-2", 9, "")
	f.Add("", "", 0, "a reason alone")
	f.Add("counting.gaps\x00", "3-4 ", -3, strings.Repeat("\u202e", 400))

	f.Fuzz(func(t *testing.T, topic, level string, difficulty int, reason string) {
		c := withAnOlderTopic()
		choice := tutor.Choice{Topic: topic, GradeLevel: rating.GradeLevel(level), Difficulty: difficulty, Reason: reason}

		got, mode, err := tutor.Next(child(t), c, choice)
		if err != nil {
			var refused *tutor.ChoiceError
			if !errors.As(err, &refused) || len(refused.Problems) == 0 {
				t.Fatalf("Next() error = %v, want a refusal naming what to change", err)
			}
			return
		}
		if !slices.Contains(c.LevelsOf(got.TargetConcept), got.GradeLevel) ||
			got.Difficulty < profile.MinDifficulty || got.Difficulty > profile.MaxDifficulty {
			t.Fatalf("Next() = difficulty %d of %s on %q, want a point the catalog has",
				got.Difficulty, got.GradeLevel, got.TargetConcept)
		}
		if chose := topic != "" || level != "" || difficulty != 0; (mode == profile.TutorLLM) != chose {
			t.Fatalf("mode = %q for %+v, want the model's exactly when it chose", mode, choice)
		}
	})
}

// The model's reason is kept as any text typed into the file is kept: what
// shows nothing or turns the text round is dropped, a break is a space, and the
// limit counts what is kept.
func TestTheModelsReasonIsKeptAsTyped(t *testing.T) {
	t.Parallel()

	reason := strings.Repeat("\u202e", 50) + "Clocks,\nagain." + strings.Repeat("\u200b", tutor.MaxReason)
	got, _, err := tutor.Next(child(t), threeTopics(), tutor.Choice{Topic: "time.clocks", Reason: reason})
	if err != nil {
		t.Fatalf("Next() error = %v, want the reason taken by what it says", err)
	}
	if strings.ContainsAny(got.Rationale, "\u202e\u200b\n") || !strings.Contains(got.Rationale, "Clocks, again.") {
		t.Errorf("rationale = %q, want the reason as it reads", got.Rationale)
	}
}
