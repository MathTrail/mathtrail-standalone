package profile_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// The first answers of a child are a trial series: the start is only a guess
// made from the grade, so the level is estimated from all of them at once, and
// the corrections of the topics wait until the child's place is known. These
// tests hold the profile to that, answer by answer.

// newChild is a profile made today for a child of this grade, with nothing
// answered yet.
func newChild(t *testing.T, grade int) *profile.Profile {
	t.Helper()

	return profile.New(profile.Student{Grade: grade, Pseudonym: "Otter"}, "0.0.0-test", issued)
}

// answer puts a task at this point in flight and answers it.
func answer(t *testing.T, p *profile.Profile, number int, topic string, point rating.Point, correct bool) profile.Recorded {
	t.Helper()

	answeringAt(t, p, topic, point)
	p.CurrentTask.ID = fmt.Sprintf("tsk_%02d", number)
	given := profile.Answered{
		TaskID:  p.CurrentTask.ID,
		Correct: correct,
		At:      issued.Add(time.Duration(number) * time.Hour),
	}
	if !correct {
		given.Chosen, given.Trap = "B", "off_by_one"
	}
	recorded, err := p.Record(given)
	if err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}
	return recorded
}

// trialAnswer is one answer of a series as a test writes it down.
type trialAnswer struct {
	topic   string
	point   rating.Point
	correct bool
}

// aSeries is five answers on five topics, at the points a child placed a level
// too high would meet, and a sixth after the series.
func aSeries() []trialAnswer {
	middle := func(difficulty int) rating.Point {
		return rating.Point{GradeLevel: rating.Grades34, Difficulty: difficulty}
	}
	youngest := func(difficulty int) rating.Point {
		return rating.Point{GradeLevel: rating.Grades12, Difficulty: difficulty}
	}
	return []trialAnswer{
		{"logic.ordering", middle(2), false},
		{"combinatorics.enumeration", youngest(3), true},
		{"logic.knights_liars", middle(1), false},
		{"counting.gaps", youngest(2), true},
		{"time.clocks", youngest(2), true},
		{"time.clocks", youngest(3), false},
	}
}

// A profile starts where its grade is on the ladder, and nowhere else.
func TestANewProfileStartsWhereItsGradeIs(t *testing.T) {
	t.Parallel()

	for grade := profile.MinGrade; grade <= profile.MaxGrade; grade++ {
		p := newChild(t, grade)
		if p.Ratings.Start != rating.Start(grade) || p.Ratings.Theta != rating.Start(grade) {
			t.Errorf("grade %d starts at %v with θ %v, want both at %v", grade, p.Ratings.Start, p.Ratings.Theta, rating.Start(grade))
		}
		if !p.Ratings.InTrial() {
			t.Errorf("grade %d: a child who has answered nothing is not in the trial series", grade)
		}
	}
}

// The series is the first five answers, and the fifth is the last of it: each
// of them sets the level to the estimate from the start and every answer so
// far, and the sixth moves it by a step, as every answer after it will.
func TestTheTrialSeriesEndsOnItsFifthAnswer(t *testing.T) {
	t.Parallel()

	p := newChild(t, 3)
	var so []rating.Answer
	for number, next := range aSeries() {
		before := p.Ratings
		topicBefore := p.Topics[next.topic]
		recorded := answer(t, p, number+1, next.topic, next.point, next.correct)

		if number < rating.TrialAnswers {
			so = append(so, rating.Answer{Point: next.point, Correct: next.correct})
			if recorded.Trial != number+1 {
				t.Errorf("answer %d is trial answer %d, want %d", number+1, recorded.Trial, number+1)
			}
			if want := rating.Estimate(rating.Start(3), so); p.Ratings.Theta != want {
				t.Errorf("after answer %d θ = %v, want the estimate %v", number+1, p.Ratings.Theta, want)
			}
			continue
		}

		if recorded.Trial != 0 || p.Ratings.InTrial() {
			t.Errorf("answer %d is trial answer %d, want the series over", number+1, recorded.Trial)
		}
		checkAStep(t, p, next, rating.State{
			Theta: before.Theta, Delta: topicBefore.Delta,
			Answers: before.Answers, TopicAnswers: topicBefore.Answers,
		})
	}
	if p.Ratings.InTrial() {
		t.Error("the series is still running after six answers")
	}
}

// checkAStep holds the profile to one answer after the series: both levels
// moved from where they stood by the step of an ordinary answer.
func checkAStep(t *testing.T, p *profile.Profile, answered trialAnswer, before rating.State) {
	t.Helper()

	want := rating.Update(before, answered.point.Beta(), answered.correct)
	if p.Ratings.Theta != want.Theta || p.Topics[answered.topic].Delta != want.Delta {
		t.Errorf("θ = %v and δ = %v, want the step to %v and %v",
			p.Ratings.Theta, p.Topics[answered.topic].Delta, want.Theta, want.Delta)
	}
}

// During the series the correction of a topic stays where it is — a trial
// answer says where the child stands, not what they know of one topic — while
// everything else about the answer is written as it always is.
func TestATrialAnswerLeavesTheTopicWhereItIs(t *testing.T) {
	t.Parallel()

	p := newChild(t, 1)
	point := rating.Point{GradeLevel: rating.Grades12, Difficulty: 2}
	recorded := answer(t, p, 1, "counting.gaps", point, false)

	topic := p.Topics["counting.gaps"]
	if topic.Delta != 0 {
		t.Errorf("delta = %v after a trial answer, want it where it was", topic.Delta)
	}
	if topic.Answers != 1 || topic.Correct != 0 || topic.Traps["off_by_one"] != 1 || topic.WrongStreak != 1 {
		t.Errorf("the topic reads %+v, want one answer, none correct, the trap counted and a run of one", topic)
	}
	if p.Ratings.Answers != 1 || p.Ratings.ConsecutiveFailures != 1 {
		t.Errorf("ratings = %+v, want one answer and one failure", p.Ratings)
	}
	if p.Ratings.Theta >= rating.Start(1) {
		t.Errorf("θ = %v after a wrong answer, want it below the start %v", p.Ratings.Theta, rating.Start(1))
	}
	if recorded.LevelAfter != p.Ratings.Theta || recorded.LevelBefore != rating.Start(1) {
		t.Errorf("recorded the level in the topic from %v to %v, want from the start to θ", recorded.LevelBefore, recorded.LevelAfter)
	}
	if want := rating.Probability(rating.Start(1), point.Beta()); recorded.Probability != want {
		t.Errorf("the chance = %v, want %v, the one the task was handed out at", recorded.Probability, want)
	}
	last := p.Recent[len(p.Recent)-1]
	if last.GradeLevel != point.GradeLevel || last.Difficulty != point.Difficulty || last.Correct {
		t.Errorf("the window ends with %+v, want the wrong answer at %+v", last, point)
	}
}

// The grade is a label once the profile is made: a child whose grade changes
// in the middle of the series, and again after it, ends exactly where a child
// answering the same tasks the same way ends, and the series is not started
// again.
func TestAChangeOfGradeMovesNothing(t *testing.T) {
	t.Parallel()

	kept, changed := newChild(t, 3), newChild(t, 3)
	for number, next := range aSeries() {
		switch number {
		case 2:
			changed.Student.Grade = 5
		case 5:
			changed.Student.Grade = 1
		}
		keptRecorded := answer(t, kept, number+1, next.topic, next.point, next.correct)
		changedRecorded := answer(t, changed, number+1, next.topic, next.point, next.correct)
		if keptRecorded != changedRecorded {
			t.Errorf("answer %d recorded %+v, and %+v with the grade changed", number+1, keptRecorded, changedRecorded)
		}
	}

	if kept.Ratings != changed.Ratings {
		t.Errorf("ratings = %+v with the grade changed, want %+v", changed.Ratings, kept.Ratings)
	}
	if !reflect.DeepEqual(kept.Topics, changed.Topics) {
		t.Errorf("topics = %+v with the grade changed, want %+v", changed.Topics, kept.Topics)
	}
	if changed.Ratings.Start != rating.Start(3) {
		t.Errorf("start = %v, want the start of the grade the profile was made with", changed.Ratings.Start)
	}
}

// The window is never pruned below the length of the series, so it holds every
// answer of the series while the series runs; and what the series reads from
// it is the last of its entries, as many as the answers the level rests on —
// a window restored or edited into holding more is read no further back.
func TestTheTrialReadsItsAnswersFromTheWindow(t *testing.T) {
	t.Parallel()

	if profile.MinRecent < rating.TrialAnswers-1 {
		t.Fatalf("the window may be pruned to %d answers, and the series reads back %d",
			profile.MinRecent, rating.TrialAnswers-1)
	}

	p := parseFixture(t, "masha") // three answers in, two of them wrong
	window := p.Recent[len(p.Recent)-p.Ratings.Answers:]
	earlier := profile.Answer{
		AnsweredAt: profile.At(issued.Add(-time.Hour)), Correct: true, Difficulty: 5,
		GradeLevel: rating.Grades56, Pace: profile.PaceNormal, TaskID: "tsk_before", Topic: "logic.ordering",
	}
	p.Recent = append([]profile.Answer{earlier}, p.Recent...)

	point := rating.Point{GradeLevel: rating.Grades34, Difficulty: 1}
	answer(t, p, 4, "logic.ordering", point, true)

	var want []rating.Answer
	for _, entry := range window {
		want = append(want, rating.Answer{Point: rating.Point{GradeLevel: entry.GradeLevel, Difficulty: entry.Difficulty}, Correct: entry.Correct})
	}
	want = append(want, rating.Answer{Point: point, Correct: true})
	if estimate := rating.Estimate(p.Ratings.Start, want); p.Ratings.Theta != estimate {
		t.Errorf("θ = %v, want %v, the estimate from the three answers of the series and this one", p.Ratings.Theta, estimate)
	}
}
