package tutor_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The catalog of these tests is written out here rather than taken from the
// binary: three topics and four traps are enough to say what the rule does,
// and a test that leans on four hundred reference tasks says it less clearly.
type catalog struct {
	topics   []string
	traps    []string
	examples map[string][]string
}

func (c catalog) TopicsAt(int) []string                     { return c.topics }
func (c catalog) TrapIDs() []string                         { return c.traps }
func (c catalog) ExampleTraps(topic string, _ int) []string { return c.examples[topic] }

func threeTopics() catalog {
	return catalog{
		topics: []string{"counting.gaps", "logic.ordering", "time.clocks"},
		traps:  []string{"off_by_one", "missed_case", "reversed_relation", "wrong_operation"},
		examples: map[string][]string{
			"counting.gaps":  {"off_by_one", "missed_case"},
			"logic.ordering": {"reversed_relation", "missed_case"},
			"time.clocks":    {"wrong_operation", "off_by_one"},
		},
	}
}

var day = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

// child is a profile to take apart: a grade, two interests, and nothing
// answered yet.
func child(t *testing.T) *profile.Profile {
	t.Helper()

	return profile.New(profile.Student{
		Grade:     2,
		Interests: []string{"cats", "trains"},
		Pseudonym: "Otter",
	}, "0.0.0-test", day)
}

// issued marks a topic as given on a day, with counts that make it a topic the
// child has actually met.
func issued(p *profile.Profile, topic string, on time.Time, answers int) {
	summary := p.Topics[topic]
	summary.LastIssued = profile.DateOf(on)
	summary.Answers += answers
	summary.Correct += answers
	if summary.Traps == nil {
		summary.Traps = map[string]int{}
	}
	p.Topics[topic] = summary
}

func brief(t *testing.T, p *profile.Profile, c catalog) profile.Brief {
	t.Helper()

	built, mode, err := tutor.Next(p, c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if mode != profile.TutorRule {
		t.Errorf("mode = %q, want %q when the model asked for nothing", mode, profile.TutorRule)
	}
	return built
}

// A child who has answered nothing gets the first topic of the catalog: none
// has been given, and catalog order settles it.
func TestATopicNeverGivenComesFirst(t *testing.T) {
	t.Parallel()

	got := brief(t, child(t), threeTopics())
	if got.TargetConcept != "counting.gaps" {
		t.Errorf("topic = %q, want the first of the catalog", got.TargetConcept)
	}
	if got.PedagogicalGoal != profile.GoalNewTopic {
		t.Errorf("goal = %q, want %q", got.PedagogicalGoal, profile.GoalNewTopic)
	}
}

// Once every topic has been given, the one waiting longest goes next.
func TestTheTopicGivenLongestAgoGoesNext(t *testing.T) {
	t.Parallel()

	p := child(t)
	issued(p, "counting.gaps", day, 1)
	issued(p, "logic.ordering", day.AddDate(0, 0, -3), 1)
	issued(p, "time.clocks", day.AddDate(0, 0, -1), 1)

	if got := brief(t, p, threeTopics()); got.TargetConcept != "logic.ordering" {
		t.Errorf("topic = %q, want the one given longest ago", got.TargetConcept)
	}
}

// A topic never given beats a topic given at any time, however long ago.
func TestNeverGivenBeatsGivenLongAgo(t *testing.T) {
	t.Parallel()

	p := child(t)
	issued(p, "counting.gaps", day.AddDate(-1, 0, 0), 1)

	if got := brief(t, p, threeTopics()); got.TargetConcept != "logic.ordering" {
		t.Errorf("topic = %q, want the first topic never given", got.TargetConcept)
	}
}

// A mastered topic steps out of the rotation.
func TestAMasteredTopicIsSkipped(t *testing.T) {
	t.Parallel()

	p := child(t)
	mastered := profile.DateOf(day)
	summary := p.Topics["counting.gaps"]
	summary.MasteredSince = &mastered
	summary.Traps = map[string]int{}
	p.Topics["counting.gaps"] = summary

	if got := brief(t, p, threeTopics()); got.TargetConcept != "logic.ordering" {
		t.Errorf("topic = %q, want the mastered one to be skipped", got.TargetConcept)
	}
}

// And when there is nothing left to master, everything comes back: a child who
// has mastered the lot still gets a task.
func TestEverythingMasteredBringsItAllBack(t *testing.T) {
	t.Parallel()

	p := child(t)
	mastered := profile.DateOf(day)
	for i, topic := range threeTopics().topics {
		summary := p.Topics[topic]
		summary.MasteredSince = &mastered
		summary.LastIssued = profile.DateOf(day.AddDate(0, 0, -i))
		summary.Traps = map[string]int{}
		p.Topics[topic] = summary
	}

	// time.clocks was issued furthest back, so it is the one that waited.
	if got := brief(t, p, threeTopics()); got.TargetConcept != "time.clocks" {
		t.Errorf("topic = %q, want the choice to run over every topic", got.TargetConcept)
	}
}

// After a failure the child works the same ground again.
func TestAFailureIsWorkedOverAgain(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Ratings.ConsecutiveFailures = 2
	p.Recent = []profile.Answer{{
		AnsweredAt: profile.At(day),
		Correct:    false,
		Difficulty: 3,
		Pace:       profile.PaceNormal,
		TaskID:     "tsk_1",
		Topic:      "time.clocks",
		Chosen:     "A",
		Trap:       "wrong_operation",
	}}
	issued(p, "time.clocks", day, 0)

	got := brief(t, p, threeTopics())
	if got.PedagogicalGoal != profile.GoalReinforce {
		t.Errorf("goal = %q, want %q", got.PedagogicalGoal, profile.GoalReinforce)
	}
	if got.TargetConcept != "time.clocks" {
		t.Errorf("topic = %q, want the one just failed", got.TargetConcept)
	}
	if got.Rationale == "" || !strings.Contains(got.Rationale, "2 failures in a row") {
		t.Errorf("rationale = %q, want it to name the run of failures", got.Rationale)
	}
}

// A profile can arrive with a run of failures and nothing in the window —
// restored from an older revision, or edited by hand. The rule reads the
// window rather than trusting it, and falls through to a new topic.
func TestAFailureWithNothingInTheWindow(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Ratings.ConsecutiveFailures = 3

	got := brief(t, p, threeTopics())
	if got.PedagogicalGoal != profile.GoalNewTopic {
		t.Errorf("goal = %q, want a new topic when there is nothing to work over", got.PedagogicalGoal)
	}
	if got.TargetConcept != "counting.gaps" {
		t.Errorf("topic = %q, want the first of the catalog", got.TargetConcept)
	}
}

// The settings go round the interests, counted by the answers behind the
// child — a number that never goes backwards, unlike the window.
func TestTheSettingGoesRoundTheInterests(t *testing.T) {
	t.Parallel()

	for answers, want := range map[int]string{0: "cats", 1: "trains", 2: "cats", 57: "trains"} {
		p := child(t)
		p.Ratings.Answers = answers

		if got := brief(t, p, threeTopics()); got.Setting != want {
			t.Errorf("after %d answers the setting is %q, want %q", answers, got.Setting, want)
		}
	}
}

// A child with no interests is a child the model dresses the task for. It is
// not an error: the profile is complete without them.
func TestNoInterestsMeansNoSetting(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Student.Interests = nil

	if got := brief(t, p, threeTopics()); got.Setting != "" {
		t.Errorf("setting = %q, want none for a child with no interests", got.Setting)
	}
}

// What the child actually gets wrong comes first, most frequent first.
func TestTheChildsOwnMistakesComeFirst(t *testing.T) {
	t.Parallel()

	// Every topic has been given and counting.gaps longest ago, so that is
	// the one the rule picks and the one whose traps are read.
	p := child(t)
	issued(p, "counting.gaps", day.AddDate(0, 0, -9), 4)
	issued(p, "logic.ordering", day, 1)
	issued(p, "time.clocks", day, 1)
	summary := p.Topics["counting.gaps"]
	summary.Traps = map[string]int{"missed_case": 3, "reversed_relation": 1, "off_by_one": 2}
	p.Topics["counting.gaps"] = summary

	got := brief(t, p, threeTopics())
	if want := []string{"missed_case", "off_by_one"}; !slices.Equal(got.TrapsToUse, want) {
		t.Errorf("traps = %v, want %v", got.TrapsToUse, want)
	}
}

// askedCatalog has no reference tasks, and counts how often it was asked for
// their traps.
type askedCatalog struct {
	catalog
	asked *int
}

func (c askedCatalog) ExampleTraps(string, int) []string {
	*c.asked++
	return nil
}

// A brief the child's own mistakes fill asks the reference tasks for nothing:
// they stand in for the mistakes a child has not made yet, and there are none
// to stand in for.
func TestOwnMistakesEnoughAskTheExamplesForNothing(t *testing.T) {
	t.Parallel()

	p := child(t)
	issued(p, "counting.gaps", day.AddDate(0, 0, -9), 4)
	issued(p, "logic.ordering", day, 1)
	issued(p, "time.clocks", day, 1)
	summary := p.Topics["counting.gaps"]
	summary.Traps = map[string]int{"missed_case": 3, "off_by_one": 2}
	p.Topics["counting.gaps"] = summary

	asked := 0
	got, _, err := tutor.Next(p, askedCatalog{catalog: threeTopics(), asked: &asked}, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if want := []string{"missed_case", "off_by_one"}; !slices.Equal(got.TrapsToUse, want) {
		t.Errorf("traps = %v, want %v", got.TrapsToUse, want)
	}
	if asked != 0 {
		t.Errorf("the reference tasks were asked %d times, want none", asked)
	}
}

// One mistake of the child's own, and the second comes from the reference
// tasks of that topic — without repeating the first.
func TestOneOwnMistakeIsToppedUpFromTheExamples(t *testing.T) {
	t.Parallel()

	p := child(t)
	issued(p, "counting.gaps", day.AddDate(0, 0, -9), 1)
	issued(p, "logic.ordering", day, 1)
	issued(p, "time.clocks", day, 1)
	summary := p.Topics["counting.gaps"]
	summary.Traps = map[string]int{"off_by_one": 1}
	p.Topics["counting.gaps"] = summary

	got := brief(t, p, threeTopics())
	if want := []string{"off_by_one", "missed_case"}; !slices.Equal(got.TrapsToUse, want) {
		t.Errorf("traps = %v, want %v", got.TrapsToUse, want)
	}
}

// A topic the child has never met has no mistakes of its own, and the brief
// still names two: the wrong options have to lead somewhere.
func TestATopicNeverMetTakesBothFromTheExamples(t *testing.T) {
	t.Parallel()

	got := brief(t, child(t), threeTopics())
	if want := []string{"off_by_one", "missed_case"}; !slices.Equal(got.TrapsToUse, want) {
		t.Errorf("traps = %v, want %v", got.TrapsToUse, want)
	}
}

// Everything the child may not meet is carried over whole.
func TestTheProhibitionsAreCarriedOver(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Student.ExcludedSkills = []string{"multiplication", "fractions"}

	got := brief(t, p, threeTopics())
	if !slices.Equal(got.ExcludedSkills, p.Student.ExcludedSkills) {
		t.Errorf("excluded skills = %v, want %v", got.ExcludedSkills, p.Student.ExcludedSkills)
	}

	// And what is handed out shares nothing with the profile it came from.
	got.ExcludedSkills[0] = "edited by a caller"
	if p.Student.ExcludedSkills[0] != "multiplication" {
		t.Error("editing the brief changed the profile it was built from")
	}
}

// A grade the catalogs do not cover is a child nothing can be chosen for, and
// the rule says so instead of handing back an empty brief.
func TestAGradeWithNoTopicsIsRefused(t *testing.T) {
	t.Parallel()

	p := child(t)
	if _, _, err := tutor.Next(p, catalog{}, tutor.Choice{}); err == nil {
		t.Error("Next() error = nil, want a refusal when the grade has no topics")
	}
}

// The number of answers comes out of a file the parent can open and edit, so
// it can arrive as anything. A count below zero is nonsense, and the answer to
// nonsense is a task the child can be given — not a panic that takes the
// request down with it.
func TestACountBelowZeroStillGetsASetting(t *testing.T) {
	t.Parallel()

	for _, answers := range []int{-1, -2, -7, -1000} {
		p := child(t)
		p.Ratings.Answers = answers

		got := brief(t, p, threeTopics())
		if !slices.Contains(p.Student.Interests, got.Setting) {
			t.Errorf("after %d answers the setting is %q, want one of %v",
				answers, got.Setting, p.Student.Interests)
		}
	}
}

// A failure on a topic the child's grade no longer takes — the grade changed,
// or the topic left the catalog — is not worked over again: no task could be
// written for it, and a run of failures that waits for an accepted task would
// never end. Something new of the grade is set instead, and the rationale says
// why.
func TestAFailureOnATopicNoLongerTakenIsNotWorkedOver(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Ratings.ConsecutiveFailures = 2
	p.Recent = []profile.Answer{{
		AnsweredAt: profile.At(day), Difficulty: 3, Pace: profile.PaceNormal,
		TaskID: "tsk_1", Topic: "time.calendar", Chosen: "A", Trap: "off_by_one",
	}}

	got, _, err := tutor.Next(p, threeTopics(), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if !slices.Contains(threeTopics().topics, got.TargetConcept) || got.PedagogicalGoal != profile.GoalNewTopic {
		t.Errorf("brief = %s, %s, want a new topic of this grade", got.TargetConcept, got.PedagogicalGoal)
	}
	if !strings.Contains(got.Rationale, "time.calendar, which is not a topic of this grade") {
		t.Errorf("rationale = %q, want it to say why the failure was not worked over", got.Rationale)
	}
}

// A brief carries empty lists rather than none: the model hands it back as it
// received it, and a list left out is refused where an empty one says there
// is nothing in it. A topic with no reference tasks at the child's level, met
// for the first time, still names traps: the catalog's own stand in.
func TestABriefNeverLeavesAListOut(t *testing.T) {
	t.Parallel()

	topics := threeTopics()
	delete(topics.examples, "counting.gaps")
	got, _, err := tutor.Next(child(t), topics, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.ExcludedSkills == nil || got.Constraints == nil || len(got.TrapsToUse) == 0 {
		t.Errorf("brief = excluded %#v, constraints %#v, traps %#v, want every list present and traps named",
			got.ExcludedSkills, got.Constraints, got.TrapsToUse)
	}
}

// With every topic of the grade mastered, the rotation takes all of them back,
// and the rationale says so rather than calling one of them unmastered.
func TestWithEveryTopicMasteredTheRationaleSaysSo(t *testing.T) {
	t.Parallel()

	p := child(t)
	mastered := profile.DateOf(day)
	for _, topic := range threeTopics().topics {
		p.Topics[topic] = profile.Topic{MasteredSince: &mastered}
	}

	got, _, err := tutor.Next(p, threeTopics(), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if !strings.Contains(got.Rationale, "every topic of this grade is mastered") {
		t.Errorf("rationale = %q, want it to say every topic is mastered", got.Rationale)
	}
}

// gradedCatalog is a catalog whose reference tasks differ by grade, which is
// what a topic taught at more than one level looks like.
type gradedCatalog struct {
	catalog
	byGrade map[int][]string
}

func (c gradedCatalog) ExampleTraps(_ string, grade int) []string { return c.byGrade[grade] }

// A topic with no reference tasks at the child's grade borrows the traps of
// the grades nearest it, the nearest first, until the brief names as many as
// it should — before the catalog's first traps, which may have nothing to do
// with the topic.
func TestATopicWithoutTasksAtTheGradeBorrowsTheNearest(t *testing.T) {
	t.Parallel()

	topics := gradedCatalog{
		catalog: threeTopics(),
		byGrade: map[int][]string{1: {"reversed_relation"}, 4: {"wrong_operation", "off_by_one"}},
	}
	got, _, err := tutor.Next(child(t), topics, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if want := []string{"reversed_relation", "wrong_operation"}; !slices.Equal(got.TrapsToUse, want) {
		t.Errorf("traps = %v, want the nearest grades', nearest first: %v", got.TrapsToUse, want)
	}
}
