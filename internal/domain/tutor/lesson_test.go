package tutor_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The child or the adult may keep the lessons to one topic. Once the trial
// series is over every task is on it, at the point the rule sets on it; the
// model may still move the level or the difficulty with a reason, but not the
// topic, which is theirs to change.

// keptTo is the child past the trial series, the lessons kept to a topic.
func keptTo(t *testing.T, topic string) *profile.Profile {
	t.Helper()

	p := settled(t)
	p.Student.LessonTopic = topic
	return p
}

func TestAChosenTopicIsSetOnceTheSeriesIsOver(t *testing.T) {
	t.Parallel()

	c := threeTopics()
	ruled := brief(t, settled(t), c)
	if ruled.TargetConcept == "time.clocks" {
		t.Fatal("the rule sets clocks itself, so this case proves nothing")
	}

	p := keptTo(t, "time.clocks")
	got, mode, err := tutor.Next(p, c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.TargetConcept != "time.clocks" || mode != profile.TutorPerson {
		t.Errorf("topic %q by %q, want the chosen time.clocks by %q", got.TargetConcept, mode, profile.TutorPerson)
	}
	want := tutor.CorridorIn(p, c, "time.clocks").Recommended
	if got.GradeLevel != want.GradeLevel || got.Difficulty != want.Difficulty {
		t.Errorf("difficulty %d of %s, want the recommended point of the topic, %d of %s",
			got.Difficulty, got.GradeLevel, want.Difficulty, want.GradeLevel)
	}
	if got.PedagogicalGoal != ruled.PedagogicalGoal {
		t.Errorf("goal = %q, want the rule's %q", got.PedagogicalGoal, ruled.PedagogicalGoal)
	}
	// The traps follow the chosen topic, not the rule's.
	if got.TrapsToUse[0] != "wrong_operation" {
		t.Errorf("traps = %v, want those of time.clocks", got.TrapsToUse)
	}
}

// The trial series finds where the child stands by moving to a new topic each
// time: a topic chosen during it waits for its end.
func TestTheTrialSeriesSetsItsOwnTopicsBesideAChoice(t *testing.T) {
	t.Parallel()

	c := threeTopics()
	p := child(t)
	want := brief(t, p, c)
	p.Student.LessonTopic = "time.clocks"
	if want.TargetConcept == "time.clocks" {
		t.Fatal("the rule sets clocks itself, so this case proves nothing")
	}

	got, mode, err := tutor.Next(p, c, tutor.Choice{})
	if err != nil || mode != profile.TutorRule || !reflect.DeepEqual(got, want) {
		t.Errorf("Next() in the trial series = %+v by %q, %v; want the rule's own brief %+v", got, mode, err, want)
	}
}

// The file is the parent's to edit, and the catalog may lose a topic: one the
// catalog does not have is no choice at all.
func TestAChosenTopicTheCatalogLostIsNoChoice(t *testing.T) {
	t.Parallel()

	c := threeTopics()
	want := brief(t, settled(t), c)
	got, mode, err := tutor.Next(keptTo(t, "astronomy.stars"), c, tutor.Choice{})
	if err != nil || mode != profile.TutorRule || !reflect.DeepEqual(got, want) {
		t.Errorf("Next() = %+v by %q, %v; want the rule's own brief %+v", got, mode, err, want)
	}
}

// A topic out of reach may be chosen: it is set at the easiest point the
// child's level allows, as the model's own choice of it would be.
func TestAChosenTopicOutOfReachIsSetAtItsRecommendedPoint(t *testing.T) {
	t.Parallel()

	got, mode, err := tutor.Next(keptTo(t, "percent.basic"), withAnOlderTopic(), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.TargetConcept != "percent.basic" || got.GradeLevel != rating.Grades56 || got.Difficulty != 1 || mode != profile.TutorPerson {
		t.Errorf("difficulty %d of %s in %q by %q, want difficulty 1 of grades 5-6 in percent.basic by a person",
			got.Difficulty, got.GradeLevel, got.TargetConcept, mode)
	}
}

// A model that names the topic the lessons are kept to asks for nothing: it
// needs no reason, and the task is the person's choice.
func TestTheModelMayNameTheChosenTopicWithoutAReason(t *testing.T) {
	t.Parallel()

	got, mode, err := tutor.Next(keptTo(t, "time.clocks"), threeTopics(), tutor.Choice{Topic: "time.clocks"})
	if err != nil {
		t.Fatalf("Next() error = %v, want the topic chosen for the lessons taken without a reason", err)
	}
	if got.TargetConcept != "time.clocks" || mode != profile.TutorPerson {
		t.Errorf("topic %q by %q, want time.clocks by a person", got.TargetConcept, mode)
	}
}

// Another topic of the model's is refused while one is chosen, and the refusal
// names the chosen one, never what the model wrote.
func TestAnotherTopicOfTheModelsIsRefusedWhileOneIsChosen(t *testing.T) {
	t.Parallel()

	_, _, err := tutor.Next(keptTo(t, "time.clocks"), threeTopics(),
		tutor.Choice{Topic: "logic.ordering", Reason: "the child likes ordering"})
	refused := refusal(t, err)
	if got := fieldsOf(refused); len(got) != 1 || got[0] != "topic" {
		t.Fatalf("refused %v, want the topic alone", got)
	}
	problem := refused.Problems[0]
	if problem.Code != profile.CodeNotOneOf || !strings.Contains(problem.Rule, "time.clocks") ||
		strings.Contains(problem.Rule, "logic.ordering") {
		t.Errorf("problem = %+v, want not_one_of naming time.clocks and not the model's topic", problem)
	}
}

// A level of the model's is held to the chosen topic: one it is not taught at
// is refused, naming it; one it is taught at is the model's choice on top.
func TestALevelOfTheModelsIsHeldToTheChosenTopic(t *testing.T) {
	t.Parallel()

	c := withAnOlderTopic()
	_, _, err := tutor.Next(keptTo(t, "percent.basic"), c, tutor.Choice{GradeLevel: rating.Grades12, Reason: "easier"})
	refused := refusal(t, err)
	if len(refused.Problems) != 1 || refused.Problems[0].Field != "grade_level" ||
		refused.Problems[0].Code != profile.CodeNotTaught || !strings.Contains(refused.Problems[0].Rule, "percent.basic") {
		t.Errorf("refused %+v, want grade_level not_taught naming percent.basic", refused.Problems)
	}

	got, mode, err := tutor.Next(keptTo(t, "percent.basic"), c, tutor.Choice{GradeLevel: rating.Grades56, Reason: "a stretch"})
	if err != nil || got.TargetConcept != "percent.basic" || mode != profile.TutorLLM {
		t.Errorf("Next() = %q by %q, %v; want percent.basic, the level the model's", got.TargetConcept, mode, err)
	}
}

// The brief keeps every account: the topic the lessons are kept to and the
// point it is set at, a choice of the model's on top, and what the rule would
// have done.
func TestTheRationaleKeepsTheChoiceBesideTheRule(t *testing.T) {
	t.Parallel()

	c := threeTopics()
	alone, _, err := tutor.Next(keptTo(t, "time.clocks"), c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	for _, say := range []string{"The lessons are kept to topic time.clocks", "its corridor's", "Rule:", "counting.gaps"} {
		if !strings.Contains(alone.Rationale, say) {
			t.Errorf("rationale = %q, want it to carry %q", alone.Rationale, say)
		}
	}

	harder, mode, err := tutor.Next(keptTo(t, "time.clocks"), c, tutor.Choice{Difficulty: 5, Reason: "a stretch on purpose"})
	if err != nil || harder.Difficulty != 5 || harder.TargetConcept != "time.clocks" || mode != profile.TutorLLM {
		t.Fatalf("Next() = difficulty %d in %q by %q, %v; want the model's 5 on time.clocks", harder.Difficulty,
			harder.TargetConcept, mode, err)
	}
	for _, say := range []string{"a stretch on purpose", "The lessons are kept to topic time.clocks", "Rule:"} {
		if !strings.Contains(harder.Rationale, say) {
			t.Errorf("rationale = %q, want it to carry %q", harder.Rationale, say)
		}
	}
}

// Whatever the child, a topic chosen for the lessons is the topic of every
// task once the trial series is over, at its recommended point, by a person;
// and during the series it changes nothing at all.
func TestTheRuleKeepsToAChosenTopic(t *testing.T) {
	t.Parallel()

	c := stepped()
	properties := gopter.NewProperties(nil)

	properties.Property("a chosen topic is set after the series and waits during it", prop.ForAll(
		func(s *seed, which int) bool {
			lesson := c.topics[which]
			p := s.build()
			ruled, ruledMode, err := tutor.Next(p, c, tutor.Choice{})
			if err != nil {
				return false
			}
			p.Student.LessonTopic = lesson
			got, mode, err := tutor.Next(p, c, tutor.Choice{})
			if err != nil {
				return false
			}
			if p.Ratings.InTrial() {
				return mode == ruledMode && reflect.DeepEqual(got, ruled)
			}
			point := tutor.CorridorIn(p, c, lesson).Recommended
			return got.TargetConcept == lesson && mode == profile.TutorPerson &&
				got.GradeLevel == point.GradeLevel && got.Difficulty == point.Difficulty &&
				got.PedagogicalGoal == ruled.PedagogicalGoal
		},
		genSeed(), gen.IntRange(0, len(c.topics)-1),
	))

	properties.TestingRun(t)
}
