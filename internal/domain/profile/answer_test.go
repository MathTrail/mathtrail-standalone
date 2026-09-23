package profile_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// issued is when the task in flight of the fixtures was handed out.
var issued = time.Date(2026, 9, 4, 18, 0, 0, 0, time.UTC)

// answering puts a task of this topic and difficulty in flight, so that a test
// can answer it. It is the shape a task has when it reaches the child, with
// nothing of the answer in it.
func answering(t *testing.T, p *profile.Profile, topic string, difficulty int) string {
	t.Helper()

	id := fmt.Sprintf("tsk_%s_%d", topic, difficulty)
	p.CurrentTask = &profile.CurrentTask{
		Difficulty:          difficulty,
		Fingerprint:         "sketch",
		Hint:                "Count them one at a time.",
		ID:                  id,
		InstructionsVersion: "357968db0310",
		IssuedAt:            profile.At(issued),
		Language:            "en-US",
		Options:             map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"},
		Sealed:              "mt1.t.kQ7fWx.placeholder",
		Topic:               topic,
		Wording:             "How many?",
	}
	return id
}

// One answer, and everything it moves. This is the whole of what a lesson
// leaves behind, and the test reads like the list a person would check by hand.
func TestAnAnswerMovesEverythingItShould(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	const topic = "counting.gaps"
	before, beforeTopic := p.Ratings, p.Topics[topic]
	id := answering(t, p, topic, 3)

	recorded, err := p.Record(profile.Answered{
		TaskID:  id,
		Chosen:  "C",
		Correct: true,
		At:      issued.Add(90 * time.Second),
	})
	if err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}

	after := p.Topics[topic]
	if p.Ratings.Answers != before.Answers+1 || after.Answers != beforeTopic.Answers+1 {
		t.Errorf("answers = %d and %d in the topic, want one more than %d and %d",
			p.Ratings.Answers, after.Answers, before.Answers, beforeTopic.Answers)
	}
	if after.Correct != beforeTopic.Correct+1 {
		t.Errorf("correct = %d, want one more than %d", after.Correct, beforeTopic.Correct)
	}
	if p.Ratings.Theta <= before.Theta || after.Delta <= beforeTopic.Delta {
		t.Error("a correct answer left the levels where they were")
	}
	if p.Ratings.ConsecutiveFailures != 0 {
		t.Errorf("consecutive_failures = %d, want a correct answer to clear it", p.Ratings.ConsecutiveFailures)
	}
	if recorded.Pace != profile.PaceNormal {
		t.Errorf("pace = %q, want normal for an answer after ninety seconds", recorded.Pace)
	}

	// The window gained the answer, and the task is no longer in flight: it
	// was answered, and it is answered once.
	last := p.Recent[len(p.Recent)-1]
	if last.TaskID != id || last.Topic != topic || !last.Correct {
		t.Errorf("the window ends with %+v, want this answer", last)
	}
	if last.Chosen != "" || last.Trap != "" {
		t.Errorf("a correct answer was written with chosen %q and trap %q, want neither", last.Chosen, last.Trap)
	}
	if p.CurrentTask != nil {
		t.Error("the task is still in flight after being answered")
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile that can be written", err)
	}
}

// A wrong answer counts the trap it led to and starts a run of failures.
func TestAWrongAnswerIsRecordedWithItsTrap(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	const topic, trap = "counting.gaps", "off_by_one"
	before := p.Topics[topic].Traps[trap]
	id := answering(t, p, topic, 3)

	if _, err := p.Record(profile.Answered{
		TaskID: id, Chosen: "A", Trap: trap, At: issued.Add(4 * time.Minute),
	}); err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}

	if got := p.Topics[topic].Traps[trap]; got != before+1 {
		t.Errorf("%s was counted %d times, want one more than %d", trap, got, before)
	}
	if p.Ratings.ConsecutiveFailures != 1 {
		t.Errorf("consecutive_failures = %d, want 1", p.Ratings.ConsecutiveFailures)
	}
	last := p.Recent[len(p.Recent)-1]
	if last.Chosen != "A" || last.Trap != trap {
		t.Errorf("the window ends with chosen %q and trap %q, want A and %s", last.Chosen, last.Trap, trap)
	}
	if last.Pace != profile.PaceSlow {
		t.Errorf("pace = %q, want slow for an answer after four minutes", last.Pace)
	}
}

// An answer for a task nobody is working on changes nothing, and says which
// of the two things went wrong.
func TestAnAnswerToSomethingElseIsRefused(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	before := p.Ratings

	if _, err := p.Record(profile.Answered{TaskID: "tsk_nothing", At: issued}); !errors.Is(err, profile.ErrNoTask) {
		t.Errorf("Record() error = %v, want %v", err, profile.ErrNoTask)
	}

	answering(t, p, "counting.gaps", 3)
	_, err := p.Record(profile.Answered{TaskID: "tsk_answered_yesterday", At: issued})
	if !errors.Is(err, profile.ErrOtherTask) {
		t.Errorf("Record() error = %v, want %v", err, profile.ErrOtherTask)
	}
	if p.Ratings != before || p.CurrentTask == nil {
		t.Error("a refused answer changed the profile")
	}
}

// How long the child took, measured by the service from the moment it handed
// the task out.
func TestThePaceOfAnAnswer(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		took time.Duration
		want profile.Pace
	}{
		{name: "inside a minute", took: 30 * time.Second, want: profile.PaceFast},
		{name: "just under a minute", took: 59 * time.Second, want: profile.PaceFast},
		{name: "a minute exactly", took: time.Minute, want: profile.PaceNormal},
		{name: "three minutes exactly", took: 3 * time.Minute, want: profile.PaceNormal},
		{name: "past three minutes", took: 4 * time.Minute, want: profile.PaceSlow},
		{name: "a clock that ran backwards", took: -time.Hour, want: profile.PaceNormal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "dima")
			id := answering(t, p, "counting.gaps", 3)

			recorded, err := p.Record(profile.Answered{TaskID: id, Correct: true, At: issued.Add(tc.took)})
			if err != nil {
				t.Fatalf("Record() error = %v, want nil", err)
			}
			if recorded.Pace != tc.want {
				t.Errorf("pace = %q, want %q", recorded.Pace, tc.want)
			}
		})
	}
}

// The window holds twenty answers and no more, and what falls out of it was
// already counted in the summary of its topic — which is what lets it fall out
// at all.
func TestTheWindowKeepsItsSizeAndLosesNothingTheRuleReads(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha") // no history, so the window fills from empty
	const topic = "counting.gaps"

	const answers = profile.MaxRecent + 5
	for i := range answers {
		id := answering(t, p, topic, 3)
		p.CurrentTask.ID = fmt.Sprintf("%s_%02d", id, i)
		if _, err := p.Record(profile.Answered{
			TaskID:  p.CurrentTask.ID,
			Correct: i%2 == 0,
			Trap:    "off_by_one",
			At:      issued.Add(time.Duration(i) * time.Hour),
		}); err != nil {
			t.Fatalf("Record() error = %v, want nil", err)
		}
	}

	if len(p.Recent) != profile.MaxRecent {
		t.Errorf("the window holds %d answers, want %d", len(p.Recent), profile.MaxRecent)
	}
	// The oldest answers are gone from the window and still counted here.
	summary := p.Topics[topic]
	if summary.Answers != answers {
		t.Errorf("the summary counts %d answers, want all %d", summary.Answers, answers)
	}
	if want := (answers + 1) / 2; summary.Correct != want {
		t.Errorf("the summary counts %d correct, want %d", summary.Correct, want)
	}
	if summary.Traps["off_by_one"] != answers/2 {
		t.Errorf("the summary counts %d traps, want %d", summary.Traps["off_by_one"], answers/2)
	}
	if p.Ratings.Answers != answers {
		t.Errorf("the ratings rest on %d answers, want %d", p.Ratings.Answers, answers)
	}
	// And the window still ends with the last answer, which is what the rule
	// reads to know what to go over again.
	if last := p.Recent[len(p.Recent)-1]; last.Topic != topic {
		t.Errorf("the window ends with %s, want the topic just answered", last.Topic)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile that can be written", err)
	}
}

// A hard task answered correctly, with no hint, is what a run toward mastery
// is made of — and the corridor is where "hard enough" is measured.
func TestWhatCountsTowardMastering(t *testing.T) {
	t.Parallel()

	const topic = "counting.gaps"

	for _, tc := range []struct {
		name       string
		difficulty int
		correct    bool
		hint       bool
		wantStreak int
		wantWrong  int
	}{
		{name: "a hard task answered", difficulty: 5, correct: true, wantStreak: 1},
		{name: "an easy task answered", difficulty: 1, correct: true, wantStreak: 0},
		{name: "a hard task answered with the hint", difficulty: 5, correct: true, hint: true, wantStreak: 0},
		{name: "a hard task failed", difficulty: 5, wantStreak: 0, wantWrong: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "sasha")
			id := answering(t, p, topic, tc.difficulty)
			recorded, err := p.Record(profile.Answered{
				TaskID: id, Correct: tc.correct, HintUsed: tc.hint, At: issued.Add(time.Minute),
			})
			if err != nil {
				t.Fatalf("Record() error = %v, want nil", err)
			}

			// The line the criterion draws is the middle of the corridor, and
			// the case is only meaningful if the task sits on the right side.
			hard := recorded.Probability <= rating.CorridorMiddle
			if hard != (tc.difficulty == 5) {
				t.Fatalf("a task of difficulty %d had a chance of %v; the case assumes otherwise",
					tc.difficulty, recorded.Probability)
			}
			if got := p.Topics[topic].TopStreak; got != tc.wantStreak {
				t.Errorf("top_streak = %d, want %d", got, tc.wantStreak)
			}
			if got := p.Topics[topic].WrongStreak; got != tc.wantWrong {
				t.Errorf("wrong_streak = %d, want %d", got, tc.wantWrong)
			}
		})
	}
}

// A correct answer ends a run of failures. The run is what the rule reads to
// decide whether to go over the same ground again, so an answer that goes
// right has to clear it.
func TestACorrectAnswerEndsTheRunOfFailures(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	// The run is set here rather than taken from a fixture: what this test is
	// about is that a correct answer ends one, and it should say so itself.
	p.Ratings.ConsecutiveFailures = 2

	id := answering(t, p, "time.clocks", 3)
	if _, err := p.Record(profile.Answered{TaskID: id, Correct: true, At: issued.Add(time.Minute)}); err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}
	if p.Ratings.ConsecutiveFailures != 0 {
		t.Errorf("consecutive_failures = %d, want a correct answer to end the run", p.Ratings.ConsecutiveFailures)
	}
}
