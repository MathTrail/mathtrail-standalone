package profile_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// Mastering a topic and losing it again, answer by answer. The criterion is
// automatic on purpose — nobody marks a topic mastered by hand — so what it
// takes and what it costs are worth reading as a sequence rather than as a
// rule quoted back.

const masteredTopic = "counting.gaps"

// answerHard answers a task of difficulty 5, which for a child near the
// starting level is at the harder half of the corridor, and answerEasy one of
// difficulty 1, which is a warm-up and proves nothing either way.
func answerHard(t *testing.T, p *profile.Profile, number int, correct, hint bool) profile.Recorded {
	t.Helper()

	return answerAt(t, p, 5, number, correct, hint)
}

func answerEasy(t *testing.T, p *profile.Profile, number int, correct bool) profile.Recorded {
	t.Helper()

	return answerAt(t, p, 1, number, correct, false)
}

func answerAt(t *testing.T, p *profile.Profile, difficulty, number int, correct, hint bool) profile.Recorded {
	t.Helper()

	answering(t, p, masteredTopic, difficulty)
	p.CurrentTask.ID = fmt.Sprintf("tsk_%02d", number)

	recorded, err := p.Record(profile.Answered{
		TaskID:   p.CurrentTask.ID,
		Correct:  correct,
		HintUsed: hint,
		Trap:     "off_by_one",
		At:       issued.Add(time.Duration(number) * time.Hour),
	})
	if err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}
	return recorded
}

// Five answers behind the topic, then three hard ones in a row with no hint.
// Not before: a lucky start must not master a topic on the third task.
func TestATopicIsMasteredByAStreakAndNotBefore(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")

	// The run is there from the first answer, and mastery is not: the topic
	// has nothing behind it yet.
	for number := 1; number <= profile.MasteryStreak; number++ {
		recorded := answerHard(t, p, number, true, false)
		if recorded.Mastered {
			t.Fatalf("the topic was mastered on answer %d, before it had %d behind it",
				number, profile.MasteryAnswers)
		}
	}
	if got := p.Topics[masteredTopic].TopStreak; got != profile.MasteryStreak {
		t.Fatalf("top_streak = %d, want %d", got, profile.MasteryStreak)
	}
	if p.Topics[masteredTopic].MasteredSince != nil {
		t.Fatal("the topic is mastered on three answers, want the count to hold it back")
	}

	// The fourth and fifth bring the count up, and the fifth is where both
	// halves of the criterion are true at once.
	if recorded := answerHard(t, p, 4, true, false); recorded.Mastered {
		t.Fatal("the topic was mastered on the fourth answer")
	}
	recorded := answerHard(t, p, 5, true, false)
	if !recorded.Mastered {
		t.Fatalf("the topic was not mastered on the fifth answer with a run of %d",
			p.Topics[masteredTopic].TopStreak)
	}
	if since := p.Topics[masteredTopic].MasteredSince; since == nil {
		t.Error("the topic is mastered and the day is not written down")
	} else if got := since.Format("2006-01-02"); got != "2026-09-04" {
		t.Errorf("mastered_since = %s, want the day of the answer that earned it", got)
	}

	// And it is earned once: the next correct answer says nothing new.
	if next := answerHard(t, p, 6, true, false); next.Mastered {
		t.Error("the topic was mastered a second time")
	}
}

// One wrong answer is a slip. Two in a row are not, and the topic goes back
// into the rotation.
func TestMasteryIsLostByTwoWrongAnswersAndNotOne(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")
	for number := 1; number <= profile.MasteryAnswers; number++ {
		answerHard(t, p, number, true, false)
	}
	if p.Topics[masteredTopic].MasteredSince == nil {
		t.Fatal("the topic was not mastered, so there is nothing to lose")
	}

	if slip := answerHard(t, p, 6, false, false); slip.Unmastered {
		t.Fatal("one wrong answer undid the mastery, want a slip to cost nothing")
	}
	if p.Topics[masteredTopic].MasteredSince == nil {
		t.Fatal("mastered_since was cleared by one wrong answer")
	}

	second := answerHard(t, p, 7, false, false)
	if !second.Unmastered {
		t.Error("two wrong answers in a row left the topic mastered")
	}
	if p.Topics[masteredTopic].MasteredSince != nil {
		t.Error("mastered_since survived two wrong answers in a row")
	}
	if got := p.Topics[masteredTopic].WrongStreak; got != profile.MasteryLostAfter {
		t.Errorf("wrong_streak = %d, want %d", got, profile.MasteryLostAfter)
	}
}

// A correct answer in between breaks the run of failures, so the two wrong
// ones have to be consecutive rather than merely two.
func TestACorrectAnswerBreaksTheRunOfFailures(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")
	for number := 1; number <= profile.MasteryAnswers; number++ {
		answerHard(t, p, number, true, false)
	}

	answerHard(t, p, 6, false, false)
	answerHard(t, p, 7, true, false)
	if got := p.Topics[masteredTopic].WrongStreak; got != 0 {
		t.Errorf("wrong_streak = %d after a correct answer, want it cleared", got)
	}
	if lost := answerHard(t, p, 8, false, false); lost.Unmastered {
		t.Error("the topic was unmastered by two failures with a correct answer between them")
	}
	if p.Topics[masteredTopic].MasteredSince == nil {
		t.Error("mastered_since was cleared by failures that were not consecutive")
	}
}

// Mastery is lost only from mastery. A topic that never earned it cannot lose
// it, however badly it goes.
func TestATopicThatWasNeverMasteredLosesNothing(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")
	for number := 1; number <= 4; number++ {
		if lost := answerHard(t, p, number, false, false); lost.Unmastered {
			t.Fatalf("answer %d unmastered a topic that was never mastered", number)
		}
	}
	if p.Topics[masteredTopic].MasteredSince != nil {
		t.Error("a topic answered wrongly four times is mastered")
	}
}

// Answering is not the criterion: a topic answered easily seven times over is
// a topic nobody has learned anything about. The run is what says the child
// can do the hard half of the corridor, and without it the count means nothing.
func TestEasyAnswersNeverAddUpToMastery(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")
	for number := 1; number <= profile.MasteryAnswers+2; number++ {
		if earned := answerEasy(t, p, number, true); earned.Mastered {
			t.Fatalf("easy answer %d mastered the topic", number)
		}
	}

	topic := p.Topics[masteredTopic]
	if topic.Answers < profile.MasteryAnswers {
		t.Fatalf("the topic has %d answers, and the case needs at least %d", topic.Answers, profile.MasteryAnswers)
	}
	if topic.TopStreak != 0 {
		t.Errorf("top_streak = %d after only easy answers, want none of them to count", topic.TopStreak)
	}
	if topic.MasteredSince != nil {
		t.Error("the topic was mastered by answers that proved nothing")
	}
}

// Any correct answer ends a run of failures, even one that earns nothing
// toward mastery: the child did not get it wrong twice in a row, which is all
// that losing mastery is about.
func TestAnEasyCorrectAnswerAlsoBreaksTheRunOfFailures(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")
	for number := 1; number <= profile.MasteryAnswers; number++ {
		answerHard(t, p, number, true, false)
	}
	if p.Topics[masteredTopic].MasteredSince == nil {
		t.Fatal("the topic was not mastered, so there is nothing to lose")
	}

	answerHard(t, p, 6, false, false)
	answerEasy(t, p, 7, true)
	if got := p.Topics[masteredTopic].WrongStreak; got != 0 {
		t.Errorf("wrong_streak = %d after an easy correct answer, want it cleared", got)
	}

	if lost := answerHard(t, p, 8, false, false); lost.Unmastered {
		t.Error("the topic was unmastered by two failures with a correct answer between them")
	}
}
