package tutor_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The rule, the ratings and the profile together, with no model: simulated
// children are set the rule's tasks from the real catalogs and answer them by
// the chance the ratings give a child of their true level. It is how the trial
// series is held to what it is for — finding where a child stands when their
// grade put them a level off — and to what it must not cost a child whose
// grade was right.
//
// Each child is played twice, with the series and without one, from the same
// seed. Without it the child starts with the answers of a series already
// behind them, so the ordinary steps begin a little narrower than a new
// child's would; the difference the series makes is far larger than that.

const (
	// cohort is how many children each case plays: enough for the shares
	// below to mean something, few enough to run with every other test.
	cohort = 200
	// lesson is how many answers each child gives.
	lesson = 20
	// longRun is the run of failures in a row that counts as long: a child
	// who fails four tasks running is a child the lesson has lost.
	longRun = 4
)

// outcome is what a cohort came to.
type outcome struct {
	// within is the share of children the series left within one level of
	// where they truly stand.
	within float64
	// distance is how far from where they truly stand the series left them,
	// on average.
	distance float64
	// longRuns is the share of children who failed longRun tasks in a row or
	// more at some point of the lesson.
	longRuns float64
}

// cachedCatalog answers the traps of reference tasks from a table made once,
// so that thousands of lessons do not count six hundred tasks each time.
type cachedCatalog struct {
	tutor.Catalog
	traps map[string][]string
}

func (c cachedCatalog) ExampleTraps(topic string, level rating.GradeLevel) []string {
	return c.traps[topic+"|"+string(level)]
}

func cached(shipped *content.Content) cachedCatalog {
	c := cachedCatalog{Catalog: shipped, traps: map[string][]string{}}
	for _, topic := range shipped.TopicIDs() {
		for _, level := range rating.GradeLevels() {
			c.traps[topic+"|"+string(level)] = shipped.ExampleTraps(topic, level)
		}
	}
	return c
}

func TestTheTrialSeriesFindsWhereAChildStands(t *testing.T) {
	t.Parallel()

	catalog := cached(embedded(t))
	const grade = 3
	for _, tc := range []struct {
		name   string
		offset float64 // where the child truly stands against the start of their grade
	}{
		{"a child whose grade is a level too high", -2.5},
		{"a child whose grade is right", 0},
		{"a child whose grade is a level too low", 2.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			with := play(t, catalog, grade, tc.offset, true)
			without := play(t, catalog, grade, tc.offset, false)
			t.Logf("with the series: %.0f%% within a level after it, %.2f off on average, %.0f%% with %d failures running; "+
				"without: %.0f%%, %.2f, %.0f%%",
				100*with.within, with.distance, 100*with.longRuns, longRun,
				100*without.within, without.distance, 100*without.longRuns)

			checkCohort(t, tc.offset, with, without)
		})
	}
}

// checkCohort holds what the series did for a cohort to what it is for.
func checkCohort(t *testing.T, offset float64, with, without outcome) {
	t.Helper()

	if with.within < 0.9 {
		t.Errorf("the series left %.0f%% of children within a level of where they stand, want 90%% or more",
			100*with.within)
	}
	if offset == 0 {
		// The price of the series is paid here: a slip early in it moves a
		// child a long way. It has to stay a price and not become a lesson
		// lost.
		if with.distance > 1 || with.longRuns > 0.15 {
			t.Errorf("children whose grade is right were left %.2f from where they stand and %.0f%% failed "+
				"%d tasks running, want at most 1 and 15%%", with.distance, 100*with.longRuns, longRun)
		}
		return
	}
	if with.distance > without.distance/2 {
		t.Errorf("the series left children %.2f from where they stand, and a start without it %.2f; "+
			"want the series to halve it at least", with.distance, without.distance)
	}
	if offset < 0 && with.longRuns > without.longRuns/2 {
		t.Errorf("%.0f%% of children placed too high failed %d tasks running with the series and %.0f%% without, "+
			"want the series to halve it at least", 100*with.longRuns, longRun, 100*without.longRuns)
	}
}

// play runs a cohort through a lesson each and says what it came to.
func play(t *testing.T, catalog tutor.Catalog, grade int, offset float64, series bool) outcome {
	t.Helper()

	var result outcome
	for child := range cohort {
		random := rand.New(rand.NewPCG(uint64(grade), uint64(child)))
		truth := rating.Start(grade) + offset
		placed, run := lessonOf(t, catalog, grade, truth, series, random)

		if math.Abs(placed-truth) < 2.5 {
			result.within++
		}
		result.distance += math.Abs(placed - truth)
		if run >= longRun {
			result.longRuns++
		}
	}
	result.within /= cohort
	result.distance /= cohort
	result.longRuns /= cohort
	return result
}

// lessonOf plays one child through a lesson: where the child is placed once
// the series is over, and the longest run of failures in the whole lesson.
func lessonOf(t *testing.T, catalog tutor.Catalog, grade int, truth float64, series bool,
	random *rand.Rand) (placed float64, longest int) {
	t.Helper()

	now := time.Date(2026, 9, 1, 17, 0, 0, 0, time.UTC)
	p := profile.New(profile.Student{Grade: grade, Pseudonym: "Otter", Interests: []string{"space"}}, "0.0.0-test", now)
	if !series {
		p.Ratings.Answers = rating.TrialAnswers
	}

	run := 0
	for number := range lesson {
		correct := answerOne(t, catalog, p, number, truth, now, random)
		if correct {
			run = 0
		} else {
			run++
		}
		longest = max(longest, run)
		if number == rating.TrialAnswers-1 {
			placed = p.Ratings.Theta
		}
		now = now.AddDate(0, 0, 1)
	}
	return placed, longest
}

// answerOne sets the child the rule's next task, answers it by the chance a
// child of this true level has at it, and records the answer.
func answerOne(t *testing.T, catalog tutor.Catalog, p *profile.Profile, number int, truth float64,
	now time.Time, random *rand.Rand) bool {
	t.Helper()

	brief, _, err := tutor.Next(p, catalog, tutor.Choice{})
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}
	id := fmt.Sprintf("tsk_%02d", number)
	p.CurrentTask = &profile.CurrentTask{
		Difficulty: brief.Difficulty, Fingerprint: "sketch", GradeLevel: brief.GradeLevel, Hint: "hint",
		ID: id, InstructionsVersion: "v", IssuedAt: profile.At(now), Language: "en",
		Options: map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"},
		Sealed:  "sealed", Topic: brief.TargetConcept, Wording: "a task",
	}
	summary := p.Topics[brief.TargetConcept]
	summary.LastIssued = profile.DateOf(now)
	if summary.Traps == nil {
		summary.Traps = map[string]int{}
	}
	p.Topics[brief.TargetConcept] = summary

	point := rating.Point{GradeLevel: brief.GradeLevel, Difficulty: brief.Difficulty}
	correct := random.Float64() < rating.Probability(truth, point.Beta())
	answer := profile.Answered{TaskID: id, Correct: correct, At: now.Add(2 * time.Minute)}
	if !correct {
		answer.Chosen, answer.Trap = "B", brief.TrapsToUse[0]
	}
	if _, err := p.Record(answer); err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}
	return correct
}
