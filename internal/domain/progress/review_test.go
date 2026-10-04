package progress_test

import (
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The review is worked out from the profile alone and says the same of it
// every time: which topics are strong and which to develop, which are too
// early to judge, and the steps to take.

// repeatsAt is the threshold the reviews of these tests count a mistake that
// repeats at: the service's own.
const repeatsAt = 2

// half is half the corridor, how far a topic stands from the overall level, or
// moves over the week, before the review says so.
var half = rating.CorridorWidth() / 2

// reviewed is the review of the child whose profile this is, today, as the
// progress works it out.
func reviewed(t *testing.T, p *profile.Profile) *progress.Review {
	t.Helper()

	return reviewedIn(t, p, embedded(t))
}

// reviewedIn is reviewed over a catalog loaded already, for the properties,
// which review a profile at every run.
func reviewedIn(t *testing.T, p *profile.Profile, catalog *content.Content) *progress.Review {
	t.Helper()

	summary := summaryOf(t, p, catalog)
	return progress.ReviewOf(p, catalog, &summary, progress.Mistakes(p.Recent, catalog, repeatsAt), repeatsAt, today)
}

// judgedChild is Petya long past the trial series, every topic he met resting
// on enough answers to judge and standing level with the overall one, none of
// them mastered, nothing in his window, and no week to tell.
func judgedChild(t *testing.T) *profile.Profile {
	t.Helper()

	p := fixture(t, "petya")
	p.Ratings.Answers = 40
	p.Recent = nil
	for id, topic := range p.Topics {
		topic.Answers, topic.Correct, topic.Delta = progress.JudgedAnswers+1, 4, 0
		topic.WrongStreak, topic.TopStreak = 0, 0
		topic.MasteredSince, topic.MasteredLevel = nil, nil
		p.Topics[id] = topic
	}
	return p
}

// with changes the topic of the profile by id as change says.
func with(p *profile.Profile, id string, change func(topic *profile.Topic)) {
	topic := p.Topics[id]
	change(&topic)
	p.Topics[id] = topic
}

// mastered marks the topic mastered at the level of Petya's grade.
func masteredAt(topic *profile.Topic) {
	level := rating.Grades34
	since := profile.DateOf(today.AddDate(0, 0, -3))
	topic.MasteredSince, topic.MasteredLevel = &since, &level
}

// right is a right answer in topic, with the hint used when hinted.
func right(topic string, hinted bool) profile.Answer {
	return profile.Answer{Topic: topic, Correct: true, HintUsed: hinted, TaskID: "task_" + topic}
}

// weekBegan keeps where the topics' corrections stood as the week began, two
// days ago, the overall level where it stands.
func weekBegan(p *profile.Profile, deltas map[string]float64) {
	p.RatingDays = []profile.RatingDay{{
		Date:   profile.DateOf(today.AddDate(0, 0, -2)),
		Deltas: deltas,
		Theta:  p.Ratings.Theta,
	}}
}

// topicsOf are the topics of judged, in order.
func topicsOf(judged []progress.Judged) []string {
	ids := make([]string, 0, len(judged))
	for _, topic := range judged {
		ids = append(ids, topic.Topic)
	}
	return ids
}

func TestThereIsNoReviewDuringTheTrialSeries(t *testing.T) {
	t.Parallel()

	for _, student := range []string{"sasha", "masha", "dima"} {
		if review := reviewed(t, fixture(t, student)); review != nil {
			t.Errorf("the review of %s is %+v during the trial series, want none", student, *review)
		}
	}
}

// Olya and Petya are past the trial series, and have met each of their topics
// too few times to judge: each is too early to judge, in the catalog's order,
// and no mistake of theirs repeats.
func TestTheSeedChildrenPastTheTrialSeriesAreTooEarlyToJudge(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		student string
		early   []string
	}{
		{"olya", []string{"combinatorics.enumeration", "counting.gaps", "time.calendar"}},
		{"petya", []string{
			"logic.ordering", "logic.knights_liars", "combinatorics.enumeration", "counting.gaps", "pigeonhole.basic", "arithmetic.tricks",
		}},
	} {
		review := reviewed(t, fixture(t, tc.student))

		want := &progress.Review{Strong: []progress.Judged{}, Develop: []progress.Judged{}, Early: tc.early, Steps: []progress.Step{}}
		if !reflect.DeepEqual(review, want) {
			t.Errorf("the review of %s is %+v, want %+v", tc.student, review, want)
		}
	}
}

// A child whose topics all stand level with the overall one, with nothing
// against them, has a review with nothing in it — and every list empty rather
// than missing.
func TestAReviewWithNothingToSayIsEmpty(t *testing.T) {
	t.Parallel()

	review := reviewed(t, judgedChild(t))

	want := &progress.Review{Strong: []progress.Judged{}, Develop: []progress.Judged{}, Early: []string{}, Steps: []progress.Step{}}
	if !reflect.DeepEqual(review, want) {
		t.Errorf("review = %+v, want one with every list empty", review)
	}
}

// A topic resting on fewer answers than mastering takes is too early to judge,
// however it stands: its level then says more of the tasks it met than of the
// child.
func TestATopicWithTooFewAnswersIsTooEarlyToJudge(t *testing.T) {
	t.Parallel()

	p := judgedChild(t)
	with(p, "counting.gaps", func(topic *profile.Topic) {
		topic.Answers, topic.Delta, topic.WrongStreak = progress.JudgedAnswers-1, -2*half, 3
	})
	with(p, "logic.ordering", func(topic *profile.Topic) { topic.Answers = 0 })

	review := reviewed(t, p)

	if !slices.Equal(review.Early, []string{"counting.gaps"}) {
		t.Errorf("early = %v, want the topic with too few answers alone, and none with no answer", review.Early)
	}
	if len(review.Develop) != 0 {
		t.Errorf("develop = %v, want no verdict on a topic too early to judge", topicsOf(review.Develop))
	}
}

// Each reason a topic is named for, at its threshold and just short of it.
func TestEachReasonIsGivenAtItsThreshold(t *testing.T) {
	t.Parallel()

	const topic = "counting.gaps"
	for _, tc := range []struct {
		name    string
		change  func(p *profile.Profile)
		strong  []progress.Reason
		develop []progress.Reason
	}{
		{"half a rank above", func(p *profile.Profile) { with(p, topic, func(t *profile.Topic) { t.Delta = half }) },
			[]progress.Reason{progress.ReasonHigh}, nil},
		{"just short of half a rank above", func(p *profile.Profile) { with(p, topic, func(t *profile.Topic) { t.Delta = half - 0.01 }) },
			nil, nil},
		{"mastered", func(p *profile.Profile) { with(p, topic, masteredAt) },
			[]progress.Reason{progress.ReasonMastered}, nil},
		{"half a rank below", func(p *profile.Profile) { with(p, topic, func(t *profile.Topic) { t.Delta = -half }) },
			nil, []progress.Reason{progress.ReasonLow}},
		{"just short of half a rank below", func(p *profile.Profile) { with(p, topic, func(t *profile.Topic) { t.Delta = -half + 0.01 }) },
			nil, nil},
		{"wrong twice in a row", func(p *profile.Profile) {
			with(p, topic, func(t *profile.Topic) { t.WrongStreak = progress.FailuresInARow })
		},
			nil, []progress.Reason{progress.ReasonFailures}},
		{"wrong once", func(p *profile.Profile) {
			with(p, topic, func(t *profile.Topic) { t.WrongStreak = progress.FailuresInARow - 1 })
		},
			nil, nil},
		{"the hint in half its answers", func(p *profile.Profile) {
			p.Recent = []profile.Answer{right(topic, true), right(topic, true), right(topic, false), right(topic, false)}
		}, nil, []progress.Reason{progress.ReasonHints}},
		{"the hint in half its answers, a task left without one beside them", func(p *profile.Profile) {
			p.Recent = []profile.Answer{
				right(topic, true), right(topic, true), right(topic, false), right(topic, false), {Topic: topic, Skipped: true},
			}
		}, nil, []progress.Reason{progress.ReasonHints}},
		{"the hint in fewer than half", func(p *profile.Profile) {
			p.Recent = []profile.Answer{right(topic, true), right(topic, true), right(topic, false), right(topic, false), right(topic, false)}
		}, nil, nil},
		{"the hint in one answer of two", func(p *profile.Profile) {
			p.Recent = []profile.Answer{right(topic, true), right(topic, false)}
		}, nil, nil},
		{"a trap fallen for as often as repeats", func(p *profile.Profile) {
			p.Recent = []profile.Answer{wrong(topic, "off_by_one"), wrong(topic, "off_by_one")}
		}, nil, []progress.Reason{progress.ReasonTrap}},
		{"a trap fallen for once", func(p *profile.Profile) {
			p.Recent = []profile.Answer{wrong(topic, "off_by_one"), wrong("logic.ordering", "off_by_one")}
		}, nil, nil},
		{"fallen half a rank over the week", func(p *profile.Profile) { weekBegan(p, map[string]float64{topic: half}) },
			nil, []progress.Reason{progress.ReasonFell}},
		{"fallen less over the week", func(p *profile.Profile) { weekBegan(p, map[string]float64{topic: half - 0.01}) },
			nil, nil},
		{"risen half a rank over the week, and strong", func(p *profile.Profile) {
			with(p, topic, func(t *profile.Topic) { t.Delta = half })
			weekBegan(p, map[string]float64{topic: 0})
		}, []progress.Reason{progress.ReasonHigh, progress.ReasonRose}, nil},
		{"risen over the week, and not strong", func(p *profile.Profile) {
			with(p, topic, func(t *profile.Topic) { t.Delta = 0 })
			weekBegan(p, map[string]float64{topic: -half})
		}, nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := judgedChild(t)
			tc.change(p)
			review := reviewed(t, p)

			if got := reasonsOf(review.Strong, topic); !slices.Equal(got, tc.strong) {
				t.Errorf("strong for %v, want %v", got, tc.strong)
			}
			if got := reasonsOf(review.Develop, topic); !slices.Equal(got, tc.develop) {
				t.Errorf("to develop for %v, want %v", got, tc.develop)
			}
		})
	}
}

// reasonsOf are the reasons topic is named for in judged, or nil when it is
// not named there.
func reasonsOf(judged []progress.Judged, topic string) []progress.Reason {
	for _, named := range judged {
		if named.Topic == topic {
			return named.Reasons
		}
	}
	return nil
}

// A topic with a reason to develop it is never called strong, however high it
// stands: the review is there to help.
func TestATopicToDevelopIsNeverStrong(t *testing.T) {
	t.Parallel()

	p := judgedChild(t)
	with(p, "counting.gaps", func(topic *profile.Topic) { topic.Delta = 2 * half; masteredAt(topic) })
	p.Recent = []profile.Answer{right("counting.gaps", true), right("counting.gaps", true)}

	review := reviewed(t, p)

	if len(review.Strong) != 0 || !slices.Equal(topicsOf(review.Develop), []string{"counting.gaps"}) {
		t.Errorf("strong %v and to develop %v, want it to develop alone", topicsOf(review.Strong), topicsOf(review.Develop))
	}
}

// The ones to develop are the topic the rule sets next first, so that the
// first step agrees with it, then the lowest, and three at most; the strong
// ones are the mastered first, then the highest, and three at most.
func TestTheTopicsNamedAreOrderedAndThreeAtMost(t *testing.T) {
	t.Parallel()

	p := judgedChild(t)
	low := map[string]float64{
		"arithmetic.tricks": -3 * half, "combinatorics.enumeration": -2 * half, "counting.gaps": -4 * half, "logic.ordering": -1.5 * half,
	}
	for id, delta := range low {
		with(p, id, func(topic *profile.Topic) { topic.Delta = delta })
	}
	if next := summaryOf(t, p, embedded(t)).Next.Topic; low[next] != 0 {
		t.Fatalf("the rule sets %s next, want a topic of none of the low ones for this case", next)
	}

	if got := topicsOf(reviewed(t, p).Develop); !slices.Equal(got, []string{"counting.gaps", "arithmetic.tricks", "combinatorics.enumeration"}) {
		t.Errorf("to develop %v, want the three lowest, the lowest first", got)
	}

	strong := judgedChild(t)
	with(strong, "logic.ordering", func(topic *profile.Topic) { topic.Delta = 3 * half })
	with(strong, "pigeonhole.basic", func(topic *profile.Topic) { topic.Delta = 2 * half })
	with(strong, "logic.knights_liars", func(topic *profile.Topic) { topic.Delta = 4 * half })
	with(strong, "counting.gaps", masteredAt)
	if got := topicsOf(reviewed(t, strong).Strong); !slices.Equal(got, []string{"counting.gaps", "logic.knights_liars", "logic.ordering"}) {
		t.Errorf("strong %v, want the mastered first, then the highest, three of them", got)
	}
}

// When the rule's next topic is one to develop, it leads, whatever stands
// lower.
func TestTheRulesTopicLeadsTheOnesToDevelop(t *testing.T) {
	t.Parallel()

	p := judgedChild(t)
	with(p, "counting.gaps", func(topic *profile.Topic) { topic.Delta = -4 * half })
	with(p, "logic.ordering", func(topic *profile.Topic) { topic.Delta = -half; topic.WrongStreak = progress.FailuresInARow })
	p.Ratings.ConsecutiveFailures = progress.FailuresInARow
	p.Recent = []profile.Answer{wrong("logic.ordering", "reversed_relation")}
	catalog := embedded(t)
	if next := summaryOf(t, p, catalog).Next; next.Topic != "logic.ordering" {
		t.Fatalf("the rule sets %+v next, want it to work over logic.ordering", next)
	}

	if got := topicsOf(reviewed(t, p).Develop); len(got) < 2 || got[0] != "logic.ordering" {
		t.Errorf("to develop %v, want the rule's topic first", got)
	}
}

// A topic to develop the rule could not set now is not named: no step could
// be taken in it.
func TestATopicOutOfReachIsNotOneToDevelop(t *testing.T) {
	t.Parallel()

	p := judgedChild(t)
	// Knights and liars is taught from grades 3 and 4 on: this far below the
	// overall level it is out of reach.
	with(p, "logic.knights_liars", func(topic *profile.Topic) { topic.Delta = -2 })
	summary := summaryOf(t, p, embedded(t))
	for _, topic := range summary.Topics {
		if topic.ID == "logic.knights_liars" && topic.InReach {
			t.Fatal("logic.knights_liars is within reach, want it out of reach for this case")
		}
	}

	if got := topicsOf(reviewed(t, p).Develop); slices.Contains(got, "logic.knights_liars") {
		t.Errorf("to develop %v, want no topic out of reach", got)
	}
}

// A topic to develop whose last answer was right has begun to move.
func TestATopicToDevelopWhoseLastAnswerWasRightIsMoving(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		wrongInARow int
		moving      bool
	}{{0, true}, {1, false}} {
		p := judgedChild(t)
		with(p, "counting.gaps", func(topic *profile.Topic) { topic.Delta, topic.WrongStreak = -half, tc.wrongInARow })

		review := reviewed(t, p)
		if len(review.Develop) != 1 || review.Develop[0].Moving != tc.moving {
			t.Errorf("with %d wrong in a row, to develop %+v, want it moving: %v", tc.wrongInARow, review.Develop, tc.moving)
		}
	}
}

// Each topic to develop is given the step of what keeps it back, and the
// mistake that repeats most across the topics is given one of its own, unless
// a topic's step gave its advice already.
func TestEachStepAdvisesOnWhatKeepsTheTopicBack(t *testing.T) {
	t.Parallel()

	p := judgedChild(t)
	with(p, "counting.gaps", func(topic *profile.Topic) { topic.Delta = -4 * half })
	with(p, "arithmetic.tricks", func(topic *profile.Topic) { topic.Delta = -3 * half })
	with(p, "logic.ordering", func(topic *profile.Topic) { topic.Delta = -2 * half })
	weekBegan(p, map[string]float64{"arithmetic.tricks": -half, "counting.gaps": -4 * half, "logic.ordering": -2 * half})
	p.Recent = []profile.Answer{
		wrong("counting.gaps", "off_by_one"), wrong("counting.gaps", "off_by_one"),
		wrong("pigeonhole.basic", "ignored_condition"), wrong("combinatorics.enumeration", "ignored_condition"),
		wrong("pigeonhole.basic", "ignored_condition"),
	}

	review := reviewed(t, p)

	want := []progress.Step{
		{Kind: progress.StepTrap, Topic: "counting.gaps", Trap: "off_by_one"},
		{Kind: progress.StepRhythm, Topic: "arithmetic.tricks"},
		{Kind: progress.StepPractice, Topic: "logic.ordering"},
		{Kind: progress.StepTrap, Trap: "ignored_condition"},
	}
	if !reflect.DeepEqual(review.Steps, want) {
		t.Errorf("steps = %+v, want %+v", review.Steps, want)
	}
}

// The step of a topic whose only reason is the hint is to try without it, and
// the mistake that repeats most is not advised again when a topic's step gave
// its advice already — nor is the next one advised in its place.
func TestNoAdviceIsGivenTwice(t *testing.T) {
	t.Parallel()

	p := judgedChild(t)
	p.Recent = []profile.Answer{
		wrong("counting.gaps", "off_by_one"), wrong("counting.gaps", "off_by_one"), wrong("counting.gaps", "off_by_one"),
		right("logic.ordering", true), right("logic.ordering", true),
		wrong("pigeonhole.basic", "double_count"), wrong("arithmetic.tricks", "double_count"),
	}

	review := reviewed(t, p)

	// The two topics stand level, and the one first in the catalog comes first.
	want := []progress.Step{
		{Kind: progress.StepUnaided, Topic: "logic.ordering"},
		{Kind: progress.StepTrap, Topic: "counting.gaps", Trap: "off_by_one"},
	}
	if !reflect.DeepEqual(review.Steps, want) {
		t.Errorf("steps = %+v, want %+v", review.Steps, want)
	}
}

// Of two mistakes that repeat as often, the one whose advice no topic's step
// gave is advised for every topic, whichever was fallen for last.
func TestOfMistakesAsFrequentTheOneNotYetAdvisedIsGiven(t *testing.T) {
	t.Parallel()

	p := judgedChild(t)
	p.Recent = []profile.Answer{
		wrong("pigeonhole.basic", "double_count"), wrong("arithmetic.tricks", "double_count"),
		wrong("logic.knights_liars", "double_count"),
		wrong("counting.gaps", "off_by_one"), wrong("counting.gaps", "off_by_one"), wrong("counting.gaps", "off_by_one"),
	}

	review := reviewed(t, p)

	want := []progress.Step{
		{Kind: progress.StepTrap, Topic: "counting.gaps", Trap: "off_by_one"},
		{Kind: progress.StepTrap, Trap: "double_count"},
	}
	if !reflect.DeepEqual(review.Steps, want) {
		t.Errorf("steps = %+v, want %+v", review.Steps, want)
	}
}

// reviewTopics are the topics the generated profiles of the properties are
// made of: Petya's, and one the catalog no longer has.
var reviewTopics = []string{
	"arithmetic.tricks", "combinatorics.enumeration", "counting.gaps", "logic.knights_liars", "logic.ordering", "pigeonhole.basic",
	"counting.cats",
}

// reviewTraps are the traps the generated answers fall for, one of them the
// catalog no longer has, and none.
var reviewTraps = []string{"", "off_by_one", "double_count", "ignored_condition", "counted_the_cat"}

// children generate what generatedChild makes a child of: each topic's state,
// the window's picks, each topic's correction, and how far it moved over the
// week — a little over two half corridors either way at most, so that the week
// says something about as often as not.
func children() []gopter.Gen {
	return []gopter.Gen{
		gen.SliceOfN(len(reviewTopics), gen.IntRange(0, 60)),
		gen.SliceOf(gen.IntRange(0, 100)),
		gen.SliceOfN(len(reviewTopics), gen.Float64Range(-2, 2)),
		gen.SliceOfN(len(reviewTopics), gen.Float64Range(-1, 1)),
	}
}

// generatedChild is the judged child with the topics as generated: each
// topic's answers, correction, run of wrong answers, mastery and how far its
// correction moved over the week, and the window's answers, each in a topic,
// right or wrong, with the hint or without, and a trap behind it or none. The
// judged child is left as it was.
func generatedChild(judged *profile.Profile, states, picks []int, deltas, moves []float64) *profile.Profile {
	p := *judged
	p.Topics = maps.Clone(judged.Topics)
	starts := map[string]float64{}
	for i, id := range reviewTopics {
		state := states[i]
		with(&p, id, func(topic *profile.Topic) {
			topic.Answers = state % 9
			topic.Delta = deltas[i]
			topic.WrongStreak = state % 4
			if state%5 == 0 {
				masteredAt(topic)
			}
		})
		if state%3 != 0 {
			starts[id] = deltas[i] - moves[i]
		}
	}
	weekBegan(&p, starts)
	p.Recent = nil
	for _, pick := range picks {
		topic, trap := reviewTopics[pick%len(reviewTopics)], reviewTraps[pick%len(reviewTraps)]
		if trap == "" {
			p.Recent = append(p.Recent, right(topic, pick%2 == 0))
		} else {
			p.Recent = append(p.Recent, wrong(topic, trap))
		}
	}
	return &p
}

// Whatever the topics and the window say, no topic is both strong and one to
// develop; each side names three at most; every step is for a topic to
// develop, or for none, and names only topics and traps of the catalog; and
// one profile is always reviewed the same.
func TestTheReviewHoldsItsProperties(t *testing.T) {
	t.Parallel()

	catalog, judged := embedded(t), judgedChild(t)
	properties := gopter.NewProperties(nil)
	properties.Property("apart, three at most, steps for the topics to develop, of the catalog, the same every time", prop.ForAll(
		func(states, picks []int, deltas, moves []float64) bool {
			p := generatedChild(judged, states, picks, deltas, moves)
			review := reviewedIn(t, p, catalog)
			return reflect.DeepEqual(review, reviewedIn(t, p, catalog)) &&
				apartAndFew(review) && stepsForTheTopicsToDevelop(review) && ofTheCatalog(review, catalog)
		},
		children()...,
	))
	properties.TestingRun(t)
}

// apartAndFew says no topic of the review is both strong and one to develop,
// and neither side names more than ReviewedTopics.
func apartAndFew(review *progress.Review) bool {
	strong, develop := topicsOf(review.Strong), topicsOf(review.Develop)
	return len(strong) <= progress.ReviewedTopics && len(develop) <= progress.ReviewedTopics &&
		!slices.ContainsFunc(strong, func(id string) bool { return slices.Contains(develop, id) })
}

// stepsForTheTopicsToDevelop says every step of the review is for a topic to
// develop, or for none.
func stepsForTheTopicsToDevelop(review *progress.Review) bool {
	develop := topicsOf(review.Develop)
	return !slices.ContainsFunc(review.Steps, func(step progress.Step) bool {
		return step.Topic != "" && !slices.Contains(develop, step.Topic)
	})
}

// ofTheCatalog says every topic and every trap the review names is one the
// catalog has.
func ofTheCatalog(review *progress.Review, catalog *content.Content) bool {
	topics := slices.Concat(topicsOf(review.Strong), topicsOf(review.Develop), review.Early)
	var traps []string
	for _, judged := range review.Develop {
		traps = append(traps, judged.Trap)
	}
	for _, step := range review.Steps {
		topics, traps = append(topics, step.Topic), append(traps, step.Trap)
	}
	return !slices.ContainsFunc(topics, func(id string) bool { return id != "" && !catalog.HasTopic(id) }) &&
		!slices.ContainsFunc(traps, func(id string) bool { return id != "" && !catalog.HasTrap(id) })
}

// The task on the card, with its sealed part, is no part of the review: it is
// what the child is doing now, not what the child knows. Only the card
// changes: the day each topic was last handed out on, which the rule reads,
// stays as it was.
func TestTheTaskOnTheCardIsNoPartOfTheReview(t *testing.T) {
	t.Parallel()

	catalog, judged := embedded(t), judgedChild(t)
	card := fixture(t, "masha").CurrentTask
	properties := gopter.NewProperties(nil)
	properties.Property("the same review with a task on the card or none", prop.ForAll(
		func(states, picks []int, deltas, moves []float64) bool {
			p := generatedChild(judged, states, picks, deltas, moves)
			without := reviewedIn(t, p, catalog)
			p.CurrentTask = card
			return reflect.DeepEqual(reviewedIn(t, p, catalog), without)
		},
		children()...,
	))
	properties.TestingRun(t)
}
