package tutor_test

import (
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// pausing is a child past the trial series with three topics met, of the
// three-topic catalog: counting and clocks within reach, and ordering on
// pause, its easiest task at a chance of 0.65, as the child the probes were
// made for stood. tasks is how many tasks are behind the child, all answered.
func pausing(t *testing.T, tasks int) *profile.Profile {
	t.Helper()

	p := settled(t)
	p.Ratings.Answers = tasks
	p.Ratings.Theta = -0.5
	for topic, delta := range map[string]float64{"counting.gaps": 0.5, "logic.ordering": -1.25, "time.clocks": 0} {
		issued(p, topic, day, 3)
		summary := p.Topics[topic]
		summary.Delta = delta
		p.Topics[topic] = summary
	}
	return p
}

func TestATopicOnPauseIsOneMetAndOutOfReach(t *testing.T) {
	t.Parallel()
	c := threeTopics()
	p := pausing(t, 20)

	if got, want := tutor.PausedTopics(p, c), []string{"logic.ordering"}; !slices.Equal(got, want) {
		t.Fatalf("PausedTopics() = %v, want %v", got, want)
	}

	// A topic out of reach the child never answered is not on pause: it has
	// not been reached yet.
	unmet := pausing(t, 20)
	summary := unmet.Topics["logic.ordering"]
	summary.Answers, summary.Correct = 0, 0
	unmet.Topics["logic.ordering"] = summary
	if got := tutor.PausedTopics(unmet, c); len(got) != 0 {
		t.Errorf("PausedTopics() of a topic out of reach never answered = %v, want none", got)
	}

	// During the trial series nothing is on pause: the series is finding where
	// the child stands.
	trial := pausing(t, rating.TrialAnswers-1)
	if got := tutor.PausedTopics(trial, c); len(got) != 0 {
		t.Errorf("PausedTopics() during the trial series = %v, want none", got)
	}
}

// A topic of a level the child has not grown into, met once because a person
// or the model chose it, is on pause like any topic met and out of reach, but
// no probe takes it until the overall level reaches its easiest task: probed
// every sixth task, it would be failed again and again, and a right answer
// guessed among five options would lift the child into tasks of a level far
// above their own. Once the overall level comes to it, the topic's own
// correction alone keeps it out of reach, and the probes take it.
func TestATopicOnPauseBeyondTheOverallLevelWaitsForIt(t *testing.T) {
	t.Parallel()
	c := withAnOlderTopic()
	p := pausing(t, 11) // eleven tasks behind: the next takes a probe's turn
	issued(p, "percent.basic", day, 1)
	older := p.Topics["percent.basic"]
	older.Correct, older.Delta = 0, -1
	p.Topics["percent.basic"] = older

	if got := tutor.PausedTopics(p, c); !slices.Contains(got, "percent.basic") {
		t.Fatalf("PausedTopics() = %v, want percent.basic among them: it was met, and is out of reach", got)
	}
	if got := tutor.Probeable(p, c); slices.Contains(got, "percent.basic") {
		t.Errorf("Probeable() = %v at an overall level of %.1f, want no percent.basic: its easiest task, of grades "+
			"5–6, is beyond that level", got, p.Ratings.Theta)
	}
	brief, _, err := tutor.Next(p, c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if brief.TargetConcept != "logic.ordering" || !strings.Contains(brief.Rationale, "probes") {
		t.Errorf("Next() set %s with the rationale %q, want a probe of logic.ordering, the topic on pause the "+
			"overall level reaches", brief.TargetConcept, brief.Rationale)
	}

	p.Ratings.Theta = 3.6 // the overall level of a child of grades 5–6, which reaches the topic's easiest task
	if got := tutor.PausedTopics(p, c); !slices.Equal(got, []string{"percent.basic"}) {
		t.Fatalf("PausedTopics() = %v at an overall level of 3.6, want [percent.basic]", got)
	}
	brief, _, err = tutor.Next(p, c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if brief.TargetConcept != "percent.basic" || !strings.Contains(brief.Rationale, "probes") {
		t.Errorf("Next() set %s with the rationale %q, want a probe of percent.basic", brief.TargetConcept, brief.Rationale)
	}
}

// A task the rule set on a topic that has since gone on pause beyond the
// overall level — written ahead while it was within reach, the answer before
// it bringing the level down — is no probe: it is answered as any task, and
// its right answer lifts nothing into reach.
func TestATaskOnATopicBeyondTheOverallLevelIsNoProbe(t *testing.T) {
	t.Parallel()
	c := withAnOlderTopic()
	p := pausing(t, 11)
	issued(p, "percent.basic", day, 1)
	older := p.Topics["percent.basic"]
	older.Correct, older.Delta = 0, -1
	p.Topics["percent.basic"] = older

	if tutor.Probes(p, c, "percent.basic", profile.TutorRule) {
		t.Errorf("Probes(percent.basic) = true at an overall level of %.1f, want false: its easiest task is beyond it",
			p.Ratings.Theta)
	}
	if !tutor.Probes(p, c, "logic.ordering", profile.TutorRule) {
		t.Error("Probes(logic.ordering) = false, want true: the overall level reaches its easiest task")
	}
	p.CurrentTask = &profile.CurrentTask{ID: "tsk_older", Topic: "percent.basic", TutorMode: profile.TutorRule}
	if judging := tutor.JudgingOf(p, c); judging.Probe {
		t.Error("JudgingOf() of a task on percent.basic is a probe, want none: no lift may follow its answer")
	}
}

// While the lessons are kept to a topic a person chose, the rule's turns do
// not run, so no topic is on pause and none is probed — not even the chosen
// topic when it stands out of reach, which every task is set on.
func TestNothingIsOnPauseWhileTheLessonsAreKeptToATopic(t *testing.T) {
	t.Parallel()
	c := threeTopics()
	for _, kept := range []string{"time.clocks", "logic.ordering"} {
		p := pausing(t, 11)
		p.Student.LessonTopic = kept
		if got := tutor.PausedTopics(p, c); len(got) != 0 {
			t.Errorf("with the lessons kept to %s, PausedTopics() = %v, want none", kept, got)
		}
	}
}

func TestNothingIsOnPauseAtTheBottomOfTheLadder(t *testing.T) {
	t.Parallel()
	c := threeTopics()
	p := pausing(t, 11)
	p.Ratings.Theta = -6 // below every topic: the rule falls back on the easiest the catalog has

	if got := tutor.PausedTopics(p, c); len(got) != 0 {
		t.Errorf("PausedTopics() below every topic = %v, want none", got)
	}
	if tutor.ProbeDue(p) {
		brief, _, err := tutor.Next(p, c, tutor.Choice{})
		if err != nil {
			t.Fatalf("Next() error = %v, want nil", err)
		}
		if strings.Contains(brief.Rationale, "probes") {
			t.Errorf("Rationale = %q at the bottom of the ladder, want no probe", brief.Rationale)
		}
	}
}

// A probe takes the last of every six tasks once the trial series is over,
// but never the sixth task, the first after a series nothing was skipped in,
// and the topics take their turns between.
func TestAProbeTakesTheLastOfEverySixTasks(t *testing.T) {
	t.Parallel()
	c := threeTopics()
	for tasks := rating.TrialAnswers; tasks <= 24; tasks++ {
		brief, _, err := tutor.Next(pausing(t, tasks), c, tutor.Choice{})
		if err != nil {
			t.Fatalf("Next() at %d tasks error = %v, want nil", tasks, err)
		}
		want := tasks == 11 || tasks == 17 || tasks == 23
		if got := brief.TargetConcept == "logic.ordering"; got != want {
			t.Errorf("at %d tasks behind the child the topic is %s, a probe of logic.ordering %t, want %t",
				tasks, brief.TargetConcept, got, want)
		}
	}
}

func TestAProbeIsTheEasiestTaskOfItsTopic(t *testing.T) {
	t.Parallel()
	brief, mode, err := tutor.Next(pausing(t, 11), threeTopics(), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if brief.TargetConcept != "logic.ordering" || brief.GradeLevel != rating.Grades12 || brief.Difficulty != 1 {
		t.Errorf("the probe is %s, difficulty %d of %s, want logic.ordering, difficulty 1 of 1-2",
			brief.TargetConcept, brief.Difficulty, brief.GradeLevel)
	}
	if brief.PedagogicalGoal != profile.GoalNewTopic || mode != profile.TutorRule {
		t.Errorf("the probe's goal and mode are %s and %s, want new_topic and rule", brief.PedagogicalGoal, mode)
	}
	for _, said := range []string{"one task in 6 probes a topic on pause", "logic.ordering has waited longest",
		"a little too hard"} {
		if !strings.Contains(brief.Rationale, said) {
			t.Errorf("Rationale = %q, want it to say %q", brief.Rationale, said)
		}
	}
}

func TestAFailureBeingWorkedOverComesBeforeAProbe(t *testing.T) {
	t.Parallel()
	p := pausing(t, 11)
	failedAt(p, "counting.gaps", 1)

	brief, _, err := tutor.Next(p, threeTopics(), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if brief.TargetConcept != "counting.gaps" || brief.PedagogicalGoal != profile.GoalReinforce {
		t.Errorf("Next() = %s for %s, want counting.gaps reinforced", brief.TargetConcept, brief.PedagogicalGoal)
	}
}

// A failure in a topic on pause is not worked over, since its easiest task is
// out of reach; on a probe's turn, the probe sets it anyway.
func TestAFailureInATopicOnPauseFallsToItsProbe(t *testing.T) {
	t.Parallel()
	p := pausing(t, 11)
	failedAt(p, "logic.ordering", 1)

	brief, _, err := tutor.Next(p, threeTopics(), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if brief.TargetConcept != "logic.ordering" || brief.PedagogicalGoal != profile.GoalNewTopic {
		t.Errorf("Next() = %s for %s, want a probe of logic.ordering", brief.TargetConcept, brief.PedagogicalGoal)
	}
}

func TestAPersonsTopicAndTheModelsChoiceComeBeforeAProbe(t *testing.T) {
	t.Parallel()
	c := threeTopics()

	kept := pausing(t, 11)
	kept.Student.LessonTopic = "time.clocks"
	brief, mode, err := tutor.Next(kept, c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if brief.TargetConcept != "time.clocks" || mode != profile.TutorPerson {
		t.Errorf("with the lessons kept to time.clocks, Next() = %s by %s, want time.clocks by person",
			brief.TargetConcept, mode)
	}

	brief, mode, err = tutor.Next(pausing(t, 11), c,
		tutor.Choice{Topic: "counting.gaps", Reason: "The adult asked for gaps."})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if brief.TargetConcept != "counting.gaps" || mode != profile.TutorLLM {
		t.Errorf("with the model asking for counting.gaps, Next() = %s by %s, want counting.gaps by llm",
			brief.TargetConcept, mode)
	}
}

func TestOfTopicsOnPauseTheOneThatWaitedLongestIsProbed(t *testing.T) {
	t.Parallel()
	p := pausing(t, 11)
	summary := p.Topics["time.clocks"]
	summary.Delta = -1.3
	summary.LastIssued = profile.DateOf(day.AddDate(0, 0, -2))
	p.Topics["time.clocks"] = summary

	brief, _, err := tutor.Next(p, threeTopics(), tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	if brief.TargetConcept != "time.clocks" {
		t.Errorf("of logic.ordering, given today, and time.clocks, given two days ago, the probe is %s, want time.clocks",
			brief.TargetConcept)
	}
}

func TestJudgingOfMarksOnlyARulesTaskOnATopicOnPause(t *testing.T) {
	t.Parallel()
	c := threeTopics()
	for _, tc := range []struct {
		name     string
		topic    string
		mode     profile.TutorMode
		answered bool
		want     bool
	}{
		{"the rule's task on a topic on pause", "logic.ordering", profile.TutorRule, false, true},
		{"a person's task on a topic on pause", "logic.ordering", profile.TutorPerson, false, false},
		{"the model's task on a topic on pause", "logic.ordering", profile.TutorLLM, false, false},
		{"the rule's task on a topic within reach", "counting.gaps", profile.TutorRule, false, false},
		{"the rule's task on a topic on pause, answered", "logic.ordering", profile.TutorRule, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := pausing(t, 11)
			p.CurrentTask = &profile.CurrentTask{ID: "tsk_1", Topic: tc.topic, TutorMode: tc.mode}
			if tc.answered {
				p.CurrentTask.Answered = &profile.Given{Choice: "C"}
			}
			judging := tutor.JudgingOf(p, c)
			if judging.Probe != tc.want {
				t.Errorf("JudgingOf().Probe = %t, want %t", judging.Probe, tc.want)
			}
			if !slices.Equal(judging.Taught, c.LevelsOf(tc.topic)) {
				t.Errorf("JudgingOf().Taught = %v, want %v", judging.Taught, c.LevelsOf(tc.topic))
			}
		})
	}
	if judging := tutor.JudgingOf(pausing(t, 11), c); judging.Probe || judging.Taught != nil {
		t.Errorf("JudgingOf() with no task on the card = %+v, want nothing", judging)
	}
}

// probed hands the probe of a child on pause out and answers it with choice,
// C being right, and says how the answer was recorded and where the topic
// stood before it.
func probed(t *testing.T, choice string) (*profile.Profile, profile.Recorded, float64) {
	t.Helper()

	sealer, err := taskSeal()
	if err != nil {
		t.Fatalf("build the seal: %v", err)
	}
	c := threeTopics()
	p := pausing(t, 11)
	brief, mode, err := tutor.Next(p, c, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	task := handedOut(t, p, tenIdeas, &brief, mode, sealer, day)
	before := p.LevelIn(task.Topic)
	recorded, err := p.Record(profile.Answered{TaskID: task.ID, Choice: choice, At: day.Add(time.Minute)},
		sealer, tutor.JudgingOf(p, c))
	if err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}
	return p, recorded, before
}

func TestARightAnswerToAProbeTakesItsTopicOffThePause(t *testing.T) {
	t.Parallel()
	p, recorded, before := probed(t, "C")

	if !recorded.Probe || !recorded.Correct {
		t.Fatalf("the answer was recorded as probe %t, correct %t, want both", recorded.Probe, recorded.Correct)
	}
	if !slices.Contains(tutor.WithinReach(p, threeTopics()), "logic.ordering") {
		t.Errorf("after a right answer to its probe, logic.ordering stands at %v, out of reach", p.LevelIn("logic.ordering"))
	}
	if got := p.LevelIn("logic.ordering"); got < before+rating.ProbeLift {
		t.Errorf("after a right answer to its probe, logic.ordering stands at %v, want at least %v", got, before+rating.ProbeLift)
	}
	if recorded.LevelAfter != p.LevelIn("logic.ordering") {
		t.Errorf("the answer tells of a level of %v, want the lifted one, %v", recorded.LevelAfter, p.LevelIn("logic.ordering"))
	}
}

func TestAWrongAnswerToAProbeLeavesItsTopicOnPause(t *testing.T) {
	t.Parallel()
	p, recorded, before := probed(t, "B")

	if !recorded.Probe || recorded.Correct {
		t.Fatalf("the answer was recorded as probe %t, correct %t, want a wrong answer to a probe", recorded.Probe, recorded.Correct)
	}
	if !slices.Contains(tutor.PausedTopics(p, threeTopics()), "logic.ordering") {
		t.Errorf("after a wrong answer to its probe, logic.ordering is not on pause")
	}
	if got := p.LevelIn("logic.ordering"); got >= before {
		t.Errorf("after a wrong answer, logic.ordering stands at %v, want below %v, where it stood", got, before)
	}
}

// fiveTopics is the three topics and two more, so that a circle of turns is
// long enough to show a topic coming round twice in it.
func fiveTopics() catalog {
	c := threeTopics()
	c.topics = append(c.topics, "time.calendar", "arithmetic.tricks")
	c.examples["time.calendar"] = []string{"off_by_one", "missed_case"}
	c.examples["arithmetic.tricks"] = []string{"wrong_operation", "missed_case"}
	return c
}

// A child behind in one topic of five, all met and within reach: over thirty
// tasks left on the card one after another, which moves no rating, the topic
// behind comes round twice as often as each of the rest.
func TestATopicBehindComesRoundTwiceAsOften(t *testing.T) {
	t.Parallel()
	sealer, err := taskSeal()
	if err != nil {
		t.Fatalf("build the seal: %v", err)
	}
	c := fiveTopics()
	p := settled(t)
	p.Ratings.Answers = 40
	for _, topic := range c.topics {
		issued(p, topic, day, 3)
	}
	summary := p.Topics["time.clocks"]
	summary.Delta = -1 // rank 1 against an overall rank 3, and its easiest task still in reach
	p.Topics["time.clocks"] = summary
	if !p.Behind("time.clocks") || slices.ContainsFunc(c.topics, func(topic string) bool {
		return topic != "time.clocks" && p.Behind(topic)
	}) {
		t.Fatalf("the child is not behind in time.clocks alone")
	}

	counts := map[string]int{}
	now := day
	for range 30 {
		brief, mode, err := tutor.Next(p, c, tutor.Choice{})
		if err != nil {
			t.Fatalf("Next() error = %v, want nil", err)
		}
		counts[handedOut(t, p, tenIdeas, &brief, mode, sealer, now).Topic]++
		now = now.Add(time.Minute)
	}
	for _, topic := range c.topics {
		want := 5
		if topic == "time.clocks" {
			want = 10
		}
		if counts[topic] != want {
			t.Errorf("over 30 tasks %s came %d times, want %d: %v", topic, counts[topic], want, counts)
		}
	}
}

// The properties of the turns hold only as well as the days they play meet
// probes; these days, children of every grade and standing, have to meet some,
// and right answers to some, or the properties would pass by never looking.
func TestTheSimulatedDaysMeetProbes(t *testing.T) {
	t.Parallel()
	catalog := cached(embedded(t))
	sealer, err := taskSeal()
	if err != nil {
		t.Fatalf("build the seal: %v", err)
	}
	probes, right := 0, 0
	random := rand.New(rand.NewPCG(7, 7))
	for at := range 48 {
		d := &daySeed{Grade: 1 + at%6, Offset: random.Float64()*5 - 2.5, Dice: random.Uint64(), Left: random.Float64() / 2}
		met, rightly := probesOf(playDay(t, catalog, sealer, d, longDay))
		probes, right = probes+met, right+rightly
	}
	if probes == 0 || right == 0 {
		t.Errorf("48 simulated days met %d probes, %d answered rightly, want some of each", probes, right)
	}
}

// probesOf is how many probes a simulated day set, and how many of them were
// answered rightly.
func probesOf(steps []step) (probes, right int) {
	for _, s := range steps {
		if !s.probe() {
			continue
		}
		probes++
		if s.right {
			right++
		}
	}
	return probes, right
}
