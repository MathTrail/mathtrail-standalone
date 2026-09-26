package tutor_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The catalog of these tests is written out here rather than taken from the
// binary: three topics, four traps and the traps of their reference tasks are
// enough to say what the rule does, and a test that leans on six hundred
// reference tasks says it less clearly. A topic is taught at every level
// unless levels says otherwise.
type catalog struct {
	topics   []string
	levels   map[string][]rating.GradeLevel
	traps    []string
	examples map[string][]string
}

func (c catalog) TopicIDs() []string { return c.topics }
func (c catalog) TrapIDs() []string  { return c.traps }

func (c catalog) LevelsOf(topic string) []rating.GradeLevel {
	if levels, set := c.levels[topic]; set {
		return levels
	}
	if slices.Contains(c.topics, topic) {
		return rating.GradeLevels()
	}
	return nil
}

func (c catalog) ExampleTraps(topic string, _ rating.GradeLevel) []string { return c.examples[topic] }

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

// withAnOlderTopic is the three topics behind one taught at grades 5–6 alone,
// first in catalog order, so that the rule reaches it before any other when it
// is within reach at all.
func withAnOlderTopic() catalog {
	c := threeTopics()
	c.topics = append([]string{"percent.basic"}, c.topics...)
	c.levels = map[string][]rating.GradeLevel{"percent.basic": {rating.Grades56}}
	c.examples["percent.basic"] = []string{"wrong_operation", "missed_case"}
	return c
}

var day = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

// child is a profile to take apart: a grade, two interests, and nothing
// answered yet, so the child is at the start of the trial series.
func child(t *testing.T) *profile.Profile {
	t.Helper()

	return profile.New(profile.Student{
		Grade:     2,
		Interests: []string{"cats", "trains"},
		Pseudonym: "Otter",
	}, "0.0.0-test", day)
}

// settled is a child past the trial series, at the same level: the rule works
// over a failure again only once the series is over.
func settled(t *testing.T) *profile.Profile {
	t.Helper()

	p := child(t)
	p.Ratings.Answers = rating.TrialAnswers
	return p
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

// mastered marks a topic mastered at a level.
func mastered(p *profile.Profile, topic string, level rating.GradeLevel) {
	summary := p.Topics[topic]
	since := profile.DateOf(day)
	summary.MasteredSince, summary.MasteredLevel = &since, &level
	if summary.Traps == nil {
		summary.Traps = map[string]int{}
	}
	p.Topics[topic] = summary
}

// failedAt leaves the window ending with a failure in this topic.
func failedAt(p *profile.Profile, topic string, failures int) {
	p.Ratings.ConsecutiveFailures = failures
	p.Recent = []profile.Answer{{
		AnsweredAt: profile.At(day),
		Difficulty: 3,
		GradeLevel: rating.Grades12,
		Pace:       profile.PaceNormal,
		TaskID:     "tsk_1",
		Topic:      topic,
		Chosen:     "A",
		Trap:       "wrong_operation",
	}}
}

func brief(t *testing.T, p *profile.Profile, c tutor.Catalog) profile.Brief {
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

// A topic mastered where its tasks now come from steps out of the rotation.
func TestAMasteredTopicIsSkipped(t *testing.T) {
	t.Parallel()

	p := child(t)
	mastered(p, "counting.gaps", rating.Grades12)

	if got := brief(t, p, threeTopics()); got.TargetConcept != "logic.ordering" {
		t.Errorf("topic = %q, want the mastered one to be skipped", got.TargetConcept)
	}
}

// Mastery is held at a level. A topic mastered among the tasks of grades 1–2
// is back in the rotation once the child's tasks in it come from grades 3–4:
// it has never been met there.
func TestAMasteredTopicComesBackWhenItsTasksMoveUp(t *testing.T) {
	t.Parallel()

	p := child(t)
	mastered(p, "counting.gaps", rating.Grades12)
	summary := p.Topics["counting.gaps"]
	summary.Delta = 2.5 // a child far past the tasks of grades 1–2 in this topic
	p.Topics["counting.gaps"] = summary

	got := brief(t, p, threeTopics())
	if got.TargetConcept != "counting.gaps" || got.GradeLevel != rating.Grades34 {
		t.Errorf("brief = %s at %s, want the topic back, its tasks now of %s",
			got.TargetConcept, got.GradeLevel, rating.Grades34)
	}
}

// And when there is nothing left to master, everything comes back: a child who
// has mastered the lot still gets a task.
func TestEverythingMasteredBringsItAllBack(t *testing.T) {
	t.Parallel()

	p := child(t)
	for i, topic := range threeTopics().topics {
		mastered(p, topic, rating.Grades12)
		summary := p.Topics[topic]
		summary.LastIssued = profile.DateOf(day.AddDate(0, 0, -i))
		p.Topics[topic] = summary
	}

	// time.clocks was issued furthest back, so it is the one that waited.
	if got := brief(t, p, threeTopics()); got.TargetConcept != "time.clocks" {
		t.Errorf("topic = %q, want the choice to run over every topic", got.TargetConcept)
	}
}

// After a failure the child works the same ground again, once the trial
// series is over.
func TestAFailureIsWorkedOverAgain(t *testing.T) {
	t.Parallel()

	p := settled(t)
	failedAt(p, "time.clocks", 2)
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

// The trial series is finding where the child stands: it moves to a topic of
// its own each time and does not go over a failure again.
func TestTheTrialSeriesDoesNotWorkOverAFailure(t *testing.T) {
	t.Parallel()

	p := child(t)
	failedAt(p, "counting.gaps", 1)
	p.Ratings.Answers = 1
	issued(p, "counting.gaps", day, 0)

	got := brief(t, p, threeTopics())
	if got.PedagogicalGoal != profile.GoalNewTopic || got.TargetConcept != "logic.ordering" {
		t.Errorf("brief = %s, %s, want the next topic never given", got.PedagogicalGoal, got.TargetConcept)
	}
	if !strings.Contains(got.Rationale, "trial series, task 2 of 5") {
		t.Errorf("rationale = %q, want it to say the trial series is running", got.Rationale)
	}
}

// A profile can arrive with a run of failures and nothing in the window —
// restored from an older revision, or edited by hand. The rule reads the
// window rather than trusting it, and falls through to a new topic.
func TestAFailureWithNothingInTheWindow(t *testing.T) {
	t.Parallel()

	p := settled(t)
	p.Ratings.ConsecutiveFailures = 3

	got := brief(t, p, threeTopics())
	if got.PedagogicalGoal != profile.GoalNewTopic {
		t.Errorf("goal = %q, want a new topic when there is nothing to work over", got.PedagogicalGoal)
	}
	if got.TargetConcept != "counting.gaps" {
		t.Errorf("topic = %q, want the first of the catalog", got.TargetConcept)
	}
}

// A failure on a topic now out of reach — the failures themselves moved it
// there, or the topic left the catalog — is not worked over again: the task
// would be failed again, and a run of failures that waits for an accepted task
// would never end. Something new within reach is set instead, and the
// rationale says why.
func TestAFailureOnATopicOutOfReachIsNotWorkedOver(t *testing.T) {
	t.Parallel()

	for _, failed := range []string{"percent.basic", "time.calendar"} {
		p := settled(t)
		failedAt(p, failed, 2)

		got, _, err := tutor.Next(p, withAnOlderTopic(), tutor.Choice{})
		if err != nil {
			t.Fatalf("Next() error = %v, want nil", err)
		}
		if got.TargetConcept == failed || got.PedagogicalGoal != profile.GoalNewTopic {
			t.Errorf("brief = %s, %s, want a new topic within reach", got.TargetConcept, got.PedagogicalGoal)
		}
		if !strings.Contains(got.Rationale, failed+", which is out of reach now") {
			t.Errorf("rationale = %q, want it to say why the failure was not worked over", got.Rationale)
		}
	}
}

// A topic whose easiest task the child would mostly fail is not set, however
// early it stands in the catalog; the rule moves on to one within reach.
func TestATopicOutOfReachIsNotSet(t *testing.T) {
	t.Parallel()

	got := brief(t, child(t), withAnOlderTopic())
	if got.TargetConcept != "counting.gaps" {
		t.Errorf("topic = %q, want the first topic within reach", got.TargetConcept)
	}
}

// The same topic comes within reach as the child grows: its easiest task is in
// the corridor once the child stands high enough, and the rule sets it at the
// point the corridor recommends.
func TestATopicComesInReachAsTheChildGrows(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Ratings.Theta = 3.6

	got := brief(t, p, withAnOlderTopic())
	if got.TargetConcept != "percent.basic" || got.GradeLevel != rating.Grades56 || got.Difficulty != 1 {
		t.Errorf("brief = %s at difficulty %d of %s, want percent.basic at its easiest point",
			got.TargetConcept, got.Difficulty, got.GradeLevel)
	}
}

// A child below every topic of the catalog is still given a task: the easiest
// the catalog has, from the topics the ladder starts with.
func TestAChildBelowTheLadderIsGivenItsBottom(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Ratings.Theta = -5

	got := brief(t, p, withAnOlderTopic())
	if got.TargetConcept != "counting.gaps" || got.GradeLevel != rating.Grades12 || got.Difficulty != 1 {
		t.Errorf("brief = %s at difficulty %d of %s, want the first topic of grades 1-2 at its easiest",
			got.TargetConcept, got.Difficulty, got.GradeLevel)
	}
}

// The brief asks for the point the corridor recommends among all the topic's
// points, whatever the child's grade: a child of grade 3 a little above the
// start is nearer the middle of the corridor at difficulty 5 of grades 1–2
// than at difficulty 2 of their own.
func TestTheBriefAsksForTheRecommendedPoint(t *testing.T) {
	t.Parallel()

	p := settled(t)
	p.Student.Grade = 3
	p.Ratings.Theta = 2.79

	got := brief(t, p, threeTopics())
	if got.GradeLevel != rating.Grades12 || got.Difficulty != 5 {
		t.Errorf("brief asks for difficulty %d of %s, want difficulty 5 of %s",
			got.Difficulty, got.GradeLevel, rating.Grades12)
	}
}

// The grade is never read: two children at the same place, one of grade 1 and
// one of grade 6, are given the same brief.
func TestTheGradeIsNeverRead(t *testing.T) {
	t.Parallel()

	young, old := settled(t), settled(t)
	young.Student.Grade, old.Student.Grade = 1, 6
	for _, p := range []*profile.Profile{young, old} {
		p.Ratings.Theta = 1.2
		failedAt(p, "logic.ordering", 1)
	}

	if a, b := brief(t, young, threeTopics()), brief(t, old, threeTopics()); !reflect.DeepEqual(a, b) {
		t.Errorf("grade 1 is given %+v and grade 6 %+v, want the same brief", a, b)
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

func (c askedCatalog) ExampleTraps(string, rating.GradeLevel) []string {
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

// A catalog with no topic taught anywhere is one nothing can be chosen from,
// and the rule says so instead of handing back an empty brief.
func TestACatalogWithNoTopicIsRefused(t *testing.T) {
	t.Parallel()

	nowhere := threeTopics()
	nowhere.levels = map[string][]rating.GradeLevel{"counting.gaps": nil, "logic.ordering": nil, "time.clocks": nil}
	for _, c := range []catalog{{}, nowhere} {
		if _, _, err := tutor.Next(child(t), c, tutor.Choice{}); err == nil {
			t.Error("Next() error = nil, want a refusal when no topic is taught at any level")
		}
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

// A brief carries empty lists rather than none: the model hands it back as it
// received it, and a list left out is refused where an empty one says there
// is nothing in it. A topic with no reference tasks, met for the first time,
// still names traps: the catalog's own stand in.
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

// With every topic within reach mastered, the rotation takes all of them back,
// and the rationale says so rather than calling one of them unmastered.
func TestWithEveryTopicMasteredTheRationaleSaysSo(t *testing.T) {
	t.Parallel()

	p := child(t)
	for _, topic := range threeTopics().topics {
		mastered(p, topic, rating.Grades12)
	}

	got, _, err := tutor.Next(p, threeTopics(), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if !strings.Contains(got.Rationale, "every topic within reach is mastered") {
		t.Errorf("rationale = %q, want it to say every topic is mastered", got.Rationale)
	}
}

// levelCatalog is a catalog whose reference tasks differ by level, which is
// what a topic taught at more than one level looks like.
type levelCatalog struct {
	catalog
	byLevel map[rating.GradeLevel][]string
}

func (c levelCatalog) ExampleTraps(_ string, level rating.GradeLevel) []string {
	return c.byLevel[level]
}

// A topic with no reference tasks at the level of the brief borrows the traps
// of the levels nearest it, the nearest first, until the brief names as many
// as it should — before the catalog's first traps, which may have nothing to
// do with the topic.
func TestATopicWithoutTasksAtItsLevelBorrowsTheNearest(t *testing.T) {
	t.Parallel()

	topics := levelCatalog{
		catalog: threeTopics(),
		byLevel: map[rating.GradeLevel][]string{
			rating.Grades34: {"reversed_relation"},
			rating.Grades56: {"wrong_operation", "off_by_one"},
		},
	}
	got, _, err := tutor.Next(child(t), topics, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if got.GradeLevel != rating.Grades12 {
		t.Fatalf("the brief is of %s, and the case needs it of %s", got.GradeLevel, rating.Grades12)
	}
	if want := []string{"reversed_relation", "wrong_operation"}; !slices.Equal(got.TrapsToUse, want) {
		t.Errorf("traps = %v, want the nearest levels', nearest first: %v", got.TrapsToUse, want)
	}
}

// The trial series takes a topic never given while one is within reach, and
// when every one of them has been given — tasks handed out and never answered
// count too — it takes the one given longest ago. The rationale says which,
// and promises no new topic it has not set.
func TestATrialTaskSaysWhatItRestedOn(t *testing.T) {
	t.Parallel()

	p := child(t)
	p.Ratings.Answers = 1
	issued(p, "counting.gaps", day.AddDate(0, 0, -3), 0)
	issued(p, "logic.ordering", day.AddDate(0, 0, -2), 0)
	issued(p, "time.clocks", day.AddDate(0, 0, -1), 0)

	got := brief(t, p, threeTopics())
	want := "Rule: trial series, task 2 of 5; counting.gaps is the unmastered topic within reach given longest ago, so new_topic."
	if !strings.HasPrefix(got.Rationale, want) {
		t.Errorf("rationale = %q, want it to begin %q", got.Rationale, want)
	}
}
