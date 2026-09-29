package profile_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// issued is when the task on the card of the fixtures was handed out.
var issued = time.Date(2026, 9, 4, 18, 0, 0, 0, time.UTC)

// The letters of the task these tests put on the card, as secret() seals it:
// C is the right option, and A is wrong with off_by_one behind it.
const (
	rightLetter = "C"
	wrongLetter = "A"
	wrongTrap   = "off_by_one"
)

// answering puts a task of this topic and difficulty on the card, at the level
// of the child's own grade, so that a test can answer it.
func answering(t *testing.T, p *profile.Profile, topic string, difficulty int) string {
	t.Helper()

	level, _ := rating.GradeLevelOf(p.Student.Grade)
	return answeringAt(t, p, topic, rating.Point{GradeLevel: level, Difficulty: difficulty})
}

// answeringAt puts a task of this topic on the card at this point of the
// ladder.
func answeringAt(t *testing.T, p *profile.Profile, topic string, point rating.Point) string {
	t.Helper()

	return answeringAs(t, p, fmt.Sprintf("tsk_%s_%s_%d", topic, point.GradeLevel, point.Difficulty), topic, point)
}

// answeringAs puts a task on the card under an id the test chooses. It is the
// shape a task has when it reaches the child: what gives its answer away is
// sealed, and tied to the id — which is why a test that needs an id of its own
// names it here rather than changing it afterwards.
func answeringAs(t *testing.T, p *profile.Profile, id, topic string, point rating.Point) string {
	t.Helper()

	p.CurrentTask = &profile.CurrentTask{
		Difficulty:          point.Difficulty,
		Fingerprint:         "sketch",
		GradeLevel:          point.GradeLevel,
		Hint:                "Count them one at a time.",
		ID:                  id,
		InstructionsVersion: "357968db0310",
		IssuedAt:            profile.At(issued),
		Language:            "en-US",
		Options:             map[string]string{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5"},
		Topic:               topic,
		Wording:             "How many?",
	}
	if err := p.SealTask(newSealer(t), secret()); err != nil {
		t.Fatalf("SealTask() error = %v, want nil", err)
	}
	return id
}

// give answers the task on the card with this choice at this moment, and is
// what was recorded.
func give(t *testing.T, p *profile.Profile, id, choice string, at time.Time) profile.Recorded {
	t.Helper()

	recorded, err := p.Record(profile.Answered{TaskID: id, Choice: choice, At: at}, newSealer(t))
	if err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}
	return recorded
}

// One answer, and everything it moves. This is the whole of what a lesson
// leaves behind, and the test reads like the list a person would check by hand.
// The child is past the trial series, so the answer moves both levels by a
// step.
func TestAnAnswerMovesEverythingItShould(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "olya")
	if p.Ratings.InTrial() {
		t.Fatalf("the fixture is in the trial series with %d answers, and the case needs it over", p.Ratings.Answers)
	}
	const topic = "counting.gaps"
	before, beforeTopic := p.Ratings, p.Topics[topic]
	levelBefore := p.LevelIn(topic)
	id := answering(t, p, topic, 3)

	recorded := give(t, p, id, rightLetter, issued.Add(90*time.Second))

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

	// The window gained the answer, and the task keeps it on the card: it was
	// answered, and it is answered once.
	last := p.Recent[len(p.Recent)-1]
	if last.TaskID != id || last.Topic != topic || !last.Correct || last.Confused {
		t.Errorf("the window ends with %+v, want this answer", last)
	}
	if last.Chosen != "" || last.Trap != "" {
		t.Errorf("a correct answer was written with chosen %q and trap %q, want neither", last.Chosen, last.Trap)
	}
	wantTold(t, &recorded, rightLetter, true, profile.Distractor{})
	wantKept(t, p, profile.Given{Choice: rightLetter, LevelAfter: p.LevelIn(topic), LevelBefore: levelBefore})
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile that can be written", err)
	}
}

// wantTold holds an answer to what it is told: the child's choice and whether
// it was right, the right option, the trap behind a wrong one and the
// solution — the part the seal says, as the first telling says it.
func wantTold(t *testing.T, recorded *profile.Recorded, choice string, correct bool, trap profile.Distractor) {
	t.Helper()

	if recorded.Choice != choice || recorded.Correct != correct || recorded.Right != rightLetter ||
		recorded.Trap != trap || recorded.Solution != secret().Solution || recorded.Again {
		t.Errorf("told %+v, want %s, right %v, the trap %+v, %s as the right option and the solution",
			recorded, choice, correct, trap, rightLetter)
	}
}

// wantKept holds the task on the card to keeping what it was given: no longer
// waiting for an answer, and able to tell the one it had again.
func wantKept(t *testing.T, p *profile.Profile, want profile.Given) {
	t.Helper()

	if p.CurrentTask == nil || p.CurrentTask.Answered == nil || *p.CurrentTask.Answered != want || p.InFlight() != nil {
		t.Errorf("the task on the card is %+v, want it answered with %+v", p.CurrentTask, want)
	}
}

// A wrong answer counts the trap it led to and starts a run of failures, and
// tells the child what went wrong on the way to the option they chose.
func TestAWrongAnswerIsRecordedWithItsTrap(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	const topic = "counting.gaps"
	before := p.Topics[topic].Traps[wrongTrap]
	id := answering(t, p, topic, 3)

	recorded := give(t, p, id, wrongLetter, issued.Add(4*time.Minute))

	if got := p.Topics[topic].Traps[wrongTrap]; got != before+1 {
		t.Errorf("%s was counted %d times, want one more than %d", wrongTrap, got, before)
	}
	if p.Ratings.ConsecutiveFailures != 1 {
		t.Errorf("consecutive_failures = %d, want 1", p.Ratings.ConsecutiveFailures)
	}
	last := p.Recent[len(p.Recent)-1]
	if last.Chosen != wrongLetter || last.Trap != wrongTrap || last.Confused {
		t.Errorf("the window ends with chosen %q and trap %q, want %s and %s", last.Chosen, last.Trap, wrongLetter, wrongTrap)
	}
	if last.Pace != profile.PaceSlow {
		t.Errorf("pace = %q, want slow for an answer after four minutes", last.Pace)
	}
	wantTold(t, &recorded, wrongLetter, false, secret().Distractors[wrongLetter])
}

// "I don't know" is an answer, and a wrong one: the card showed the solution
// for it. It moves the levels exactly as a wrong letter does, and runs of
// failures with them, and it names no option and no trap — so the map of the
// child's mistakes, which is made of traps, does not count it.
func TestIDontKnowIsAWrongAnswerThatTookNoTrap(t *testing.T) {
	t.Parallel()

	const topic = "counting.gaps"
	unsure, mistaken := parseFixture(t, "olya"), parseFixture(t, "olya")
	traps := unsure.Topics[topic].Traps[wrongTrap]

	told := give(t, unsure, answering(t, unsure, topic, 3), profile.DontKnow, issued.Add(time.Minute))
	give(t, mistaken, answering(t, mistaken, topic, 3), wrongLetter, issued.Add(time.Minute))

	if unsure.Ratings != mistaken.Ratings || unsure.Topics[topic].Delta != mistaken.Topics[topic].Delta {
		t.Errorf("I don't know moved the levels to %+v, want %+v as a wrong letter does",
			unsure.Ratings, mistaken.Ratings)
	}
	if unsure.Ratings.ConsecutiveFailures == 0 || unsure.Topics[topic].WrongStreak != mistaken.Topics[topic].WrongStreak {
		t.Errorf("the runs are %d and %d, want them to count a failure", unsure.Ratings.ConsecutiveFailures,
			unsure.Topics[topic].WrongStreak)
	}
	if got := unsure.Topics[topic].Traps[wrongTrap]; got != traps {
		t.Errorf("%s was counted %d times after I don't know, want %d as before", wrongTrap, got, traps)
	}

	last := unsure.Recent[len(unsure.Recent)-1]
	if !last.Confused || last.Correct || last.Chosen != "" || last.Trap != "" {
		t.Errorf("the window ends with %+v, want a wrong answer marked confused, with no option and no trap", last)
	}
	wantTold(t, &told, profile.DontKnow, false, profile.Distractor{})
	if err := unsure.Validate(); err != nil {
		t.Errorf("Validate() error = %v, want a profile that can be written", err)
	}
}

// An answer is recorded once. Sent again — by a second tab, after a reply lost
// on the way, by the model recording what the card already did — it moves
// nothing and is told what the first recorded, word for word, whatever it
// chose itself.
func TestAnAnswerIsRecordedOnce(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "olya")
	id := answering(t, p, "counting.gaps", 3)
	first := give(t, p, id, rightLetter, issued.Add(time.Minute))
	ratings, topics, window, kept := p.Ratings, p.Topics["counting.gaps"], len(p.Recent), *p.CurrentTask.Answered

	for _, choice := range []string{rightLetter, wrongLetter, profile.DontKnow} {
		again := give(t, p, id, choice, issued.Add(time.Hour))
		if !again.Again {
			t.Errorf("sent again with %s, it was recorded anew: %+v", choice, again)
		}
		if again.Choice != first.Choice || again.Correct != first.Correct || again.Right != first.Right ||
			again.Trap != first.Trap || again.Solution != first.Solution || again.LevelBefore != first.LevelBefore ||
			again.LevelAfter != first.LevelAfter || again.Trial != first.Trial || again.HintUsed != first.HintUsed {
			t.Errorf("sent again with %s, it was told %+v, want what was recorded: %+v", choice, again, first)
		}
	}
	if p.Ratings != ratings || p.Topics["counting.gaps"].Answers != topics.Answers || len(p.Recent) != window ||
		*p.CurrentTask.Answered != kept {
		t.Error("an answer sent again moved the profile")
	}
}

// A task whose seal will not open — the key that sealed it has been retired —
// cannot be checked, so it takes no answer, and nothing moves.
func TestATaskWhoseSealWillNotOpenTakesNoAnswer(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "olya")
	id := answering(t, p, "counting.gaps", 3)
	before := p.Ratings

	_, err := p.Record(profile.Answered{TaskID: id, Choice: rightLetter, At: issued.Add(time.Minute)}, newOtherSealer(t))
	if !errors.Is(err, profile.ErrSealed) {
		t.Errorf("Record() error = %v, want %v", err, profile.ErrSealed)
	}
	if p.Ratings != before || p.CurrentTask.Answered != nil || p.InFlight() == nil {
		t.Error("an answer to a task that could not be checked moved the profile")
	}

	p.DiscardTask()
	if p.CurrentTask != nil || len(p.Recent) != len(parseFixture(t, "olya").Recent) {
		t.Errorf("the task was discarded with %d entries in the window, want it gone and nothing written", len(p.Recent))
	}
}

// An answer for a task nobody is working on changes nothing, and says which
// of the two things went wrong.
func TestAnAnswerToSomethingElseIsRefused(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	before := p.Ratings

	_, err := p.Record(profile.Answered{TaskID: "tsk_nothing", Choice: rightLetter, At: issued}, newSealer(t))
	if !errors.Is(err, profile.ErrNoTask) {
		t.Errorf("Record() error = %v, want %v", err, profile.ErrNoTask)
	}

	answering(t, p, "counting.gaps", 3)
	_, err = p.Record(profile.Answered{TaskID: "tsk_answered_yesterday", Choice: rightLetter, At: issued}, newSealer(t))
	if !errors.Is(err, profile.ErrOtherTask) {
		t.Errorf("Record() error = %v, want %v", err, profile.ErrOtherTask)
	}
	if p.Ratings != before || p.InFlight() == nil {
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
			recorded := give(t, p, answering(t, p, "counting.gaps", 3), rightLetter, issued.Add(tc.took))
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
	point := rating.Point{GradeLevel: rating.Grades34, Difficulty: 3}

	const answers = profile.MaxRecent + 5
	for i := range answers {
		id := answeringAs(t, p, fmt.Sprintf("tsk_%02d", i), topic, point)
		choice := wrongLetter
		if i%2 == 0 {
			choice = rightLetter
		}
		give(t, p, id, choice, issued.Add(time.Duration(i)*time.Hour))
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
	if summary.Traps[wrongTrap] != answers/2 {
		t.Errorf("the summary counts %d traps, want %d", summary.Traps[wrongTrap], answers/2)
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
		choice     string
		hint       bool
		wantStreak int
		wantWrong  int
	}{
		{name: "a hard task answered", difficulty: 5, choice: rightLetter, wantStreak: 1},
		{name: "an easy task answered", difficulty: 1, choice: rightLetter, wantStreak: 0},
		{name: "a hard task answered with the hint", difficulty: 5, choice: rightLetter, hint: true, wantStreak: 0},
		{name: "a hard task failed", difficulty: 5, choice: wrongLetter, wantStreak: 0, wantWrong: 1},
		{name: "a hard task not known", difficulty: 5, choice: profile.DontKnow, wantStreak: 0, wantWrong: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "sasha")
			id := answering(t, p, topic, tc.difficulty)
			recorded, err := p.Record(profile.Answered{
				TaskID: id, Choice: tc.choice, HintUsed: tc.hint, At: issued.Add(time.Minute),
			}, newSealer(t))
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

	give(t, p, answering(t, p, "time.clocks", 3), rightLetter, issued.Add(time.Minute))
	if p.Ratings.ConsecutiveFailures != 0 {
		t.Errorf("consecutive_failures = %d, want a correct answer to end the run", p.Ratings.ConsecutiveFailures)
	}
}

// An answer that is neither a letter of the options nor "I don't know" is
// refused before anything moves: the history keeps the letter of a wrong
// answer, and a profile holding one that names no option could never be
// written back. What a person typed is read by ChoiceOf; this takes only what
// it reads.
func TestAnAnswerNamingNoOptionIsRefused(t *testing.T) {
	t.Parallel()

	for _, choice := range []string{"BC", "b", "F", " A", "", "??"} {
		t.Run(choice, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "dima")
			id := answering(t, p, "counting.gaps", 3)
			before := p.Ratings

			_, err := p.Record(profile.Answered{TaskID: id, Choice: choice, At: issued.Add(time.Minute)}, newSealer(t))
			if !errors.Is(err, profile.ErrNoSuchOption) {
				t.Errorf("Record() error = %v, want %v", err, profile.ErrNoSuchOption)
			}
			if p.InFlight() == nil || p.Ratings != before {
				t.Error("Record() moved the profile, want it left as it was")
			}
			if _, err := profile.Marshal(p); err != nil {
				t.Errorf("Marshal() after the refusal error = %v, want the profile still writable", err)
			}
		})
	}
}

// An answer is read as a child or a model types it: a letter in either case,
// or a question mark, with spaces around it. Anything else is refused by a
// rule that never repeats what was typed.
func TestAnAnswerIsReadAsItWasTyped(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		typed string
		want  string
	}{
		{"C", "C"},
		{" c ", "C"},
		{"e", "E"},
		{"\tA\n", "A"},
		{"?", profile.DontKnow},
		{" ? ", profile.DontKnow},
	} {
		if choice, rule := profile.ChoiceOf(tc.typed); choice != tc.want || rule != "" {
			t.Errorf("ChoiceOf(%q) = %q, %q, want %q and no rule", tc.typed, choice, rule, tc.want)
		}
	}

	for _, typed := range []string{"", "F", "AB", "C)", "c.", "zebra", "¿", "Ｃ", "?!"} {
		choice, rule := profile.ChoiceOf(typed)
		if choice != "" || rule == "" {
			t.Errorf("ChoiceOf(%q) = %q, %q, want no choice and a rule", typed, choice, rule)
		}
		if strings.TrimSpace(typed) != "" && strings.Contains(rule, typed) {
			t.Errorf("ChoiceOf(%q) says %q, want a rule that does not repeat what was typed", typed, rule)
		}
	}
}

// A file that says null where the topics go — one a person edited by hand —
// still takes an answer: the topic is written into a summary of its own
// rather than into nothing.
func TestAFileWithNoTopicsStillTakesAnAnswer(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")
	id := answering(t, p, "counting.gaps", 3)
	p.Topics = nil

	give(t, p, id, wrongLetter, issued.Add(time.Minute))
	if summary := p.Topics["counting.gaps"]; summary.Answers != 1 || summary.Traps[wrongTrap] != 1 {
		t.Errorf("topics = %+v, want the answer and its trap counted", p.Topics)
	}
}

// An answer told again needs the seal as much as the first did. Once the seal
// will not open, nothing is told, and the answer that was recorded stays as it
// was.
func TestAnAnswerToldAgainNeedsTheSeal(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "olya")
	id := answering(t, p, "counting.gaps", 3)
	give(t, p, id, rightLetter, issued.Add(time.Minute))
	ratings, window := p.Ratings, len(p.Recent)

	_, err := p.Record(profile.Answered{TaskID: id, Choice: rightLetter, At: issued.Add(time.Hour)}, newOtherSealer(t))
	if !errors.Is(err, profile.ErrSealed) {
		t.Errorf("Record() error = %v, want %v", err, profile.ErrSealed)
	}
	if p.Ratings != ratings || len(p.Recent) != window || p.CurrentTask.Answered == nil {
		t.Error("an answer told again with its seal lost moved the profile")
	}
}

// Only an answer recorded here opens the task to be told again. A task the
// file says was answered — written into it by anybody but this service — does
// not open, so no answer is told before the child has given one; nor does an
// answer whose letter was changed in the file, nor one taken out of it to be
// answered anew.
func TestOnlyAnAnswerRecordedHereIsToldAgain(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		edit func(t *testing.T, p *profile.Profile, id string)
	}{
		{"marked answered, never answered", func(_ *testing.T, p *profile.Profile, _ string) {
			p.CurrentTask.Answered = &profile.Given{Choice: wrongLetter}
		}},
		{"answered with another letter", func(t *testing.T, p *profile.Profile, id string) {
			give(t, p, id, wrongLetter, issued.Add(time.Minute))
			p.CurrentTask.Answered.Choice = rightLetter
		}},
		{"its answer taken out", func(t *testing.T, p *profile.Profile, id string) {
			give(t, p, id, wrongLetter, issued.Add(time.Minute))
			p.CurrentTask.Answered = nil
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parseFixture(t, "olya")
			id := answering(t, p, "counting.gaps", 3)
			tc.edit(t, p, id)
			ratings, window := p.Ratings, len(p.Recent)

			told, err := p.Record(profile.Answered{TaskID: id, Choice: rightLetter, At: issued.Add(time.Hour)}, newSealer(t))
			if !errors.Is(err, profile.ErrSealed) || told.Right != "" || told.Solution != "" {
				t.Errorf("Record() = %+v, %v; want %v and nothing told", told, err, profile.ErrSealed)
			}
			if p.Ratings != ratings || len(p.Recent) != window {
				t.Error("an answer to an edited task moved the profile")
			}
		})
	}
}
