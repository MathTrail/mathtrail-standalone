package profile_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Mastering a topic and losing it again, answer by answer. The criterion is
// automatic on purpose — nobody marks a topic mastered by hand — so what it
// takes and what it costs are worth reading as a sequence rather than as a
// rule quoted back.

const masteredTopic = "counting.gaps"

// ownLevel is a difficulty of the level the child's grade falls into.
func ownLevel(t *testing.T, p *profile.Profile, difficulty int) rating.Point {
	t.Helper()

	level, _ := rating.GradeLevelOf(p.Student.Grade)
	return rating.Point{GradeLevel: level, Difficulty: difficulty}
}

// answerHard answers a task of difficulty 5, which for a child near the
// starting level is at the harder half of the corridor, and answerEasy one of
// difficulty 1, which is a warm-up and proves nothing either way.
func answerHard(t *testing.T, p *profile.Profile, number int, correct, hint bool) profile.Recorded {
	t.Helper()

	return answerAt(t, p, ownLevel(t, p, 5), number, correct, hint)
}

func answerEasy(t *testing.T, p *profile.Profile, number int, correct bool) profile.Recorded {
	t.Helper()

	return answerAt(t, p, ownLevel(t, p, 1), number, correct, false)
}

func answerAt(t *testing.T, p *profile.Profile, point rating.Point, number int, correct, hint bool) profile.Recorded {
	t.Helper()

	id := answeringAs(t, p, fmt.Sprintf("tsk_%02d", number), masteredTopic, point)
	choice := wrongLetter
	if correct {
		choice = rightLetter
	}

	recorded, err := p.Record(profile.Answered{
		TaskID:   id,
		Choice:   choice,
		HintUsed: hint,
		At:       issued.Add(time.Duration(number) * time.Hour),
	}, newSealer(t), everyLevel)
	if err != nil {
		t.Fatalf("Record() error = %v, want nil", err)
	}
	return recorded
}

// settled is a child of so many answers at this level overall, standing in
// the topic at this correction over so many of its own, in a file the
// cautious estimate's masteries are kept in.
func settled(t *testing.T, theta float64, answers int, delta float64, inTopic int) *profile.Profile {
	t.Helper()

	p := parseFixture(t, "sasha")
	p.Ratings.Theta, p.Ratings.Answers = theta, answers
	p.Ratings.MasteryRule = profile.MasteryRuleCautious
	p.Topics[masteredTopic] = profile.Topic{Delta: delta, Answers: inTopic, Correct: inTopic, Traps: map[string]int{}}
	return p
}

// wantMasteredAt holds the topic to mastery at this level, since the day of
// this moment.
func wantMasteredAt(t *testing.T, p *profile.Profile, level rating.GradeLevel, since time.Time) {
	t.Helper()

	topic := p.Topics[masteredTopic]
	if topic.MasteredLevel == nil || *topic.MasteredLevel != level {
		t.Errorf("mastered at %v, want %s", topic.MasteredLevel, level)
	}
	if want := profile.DateOf(since); topic.MasteredSince == nil || !topic.MasteredSince.Equal(want.Time) {
		t.Errorf("mastered since %v, want %v", topic.MasteredSince, want)
	}
}

// Worked example 4 of the specification, in its masteries: a child of 79
// answers, at 1.2, in a topic at 0.5 over nine answers. A right answer to a
// task of 3-4 masters the topic at 1-2, the highest level the cautious level
// clears; one wrong answer after it is a slip, and a second in a row loses it.
func TestTheWorkedExampleOfMastery(t *testing.T) {
	t.Parallel()

	p := settled(t, 1.2, 79, 0.5, 9)

	first := answerAt(t, p, rating.Point{GradeLevel: rating.Grades34, Difficulty: 1}, 1, true, false)
	if !first.Mastered {
		t.Fatal("the first answer did not master the topic, want it mastered at 1-2")
	}
	wantMasteredAt(t, p, rating.Grades12, issued.Add(time.Hour))

	if slip := answerAt(t, p, rating.Point{GradeLevel: rating.Grades12, Difficulty: 4}, 2, false, false); slip.Mastered || slip.Unmastered {
		t.Errorf("one wrong answer: mastered %v, unmastered %v; want a slip that changes nothing", slip.Mastered, slip.Unmastered)
	}
	wantMasteredAt(t, p, rating.Grades12, issued.Add(time.Hour))

	if second := answerAt(t, p, rating.Point{GradeLevel: rating.Grades12, Difficulty: 3}, 3, false, false); !second.Unmastered {
		t.Error("two wrong answers in a row left the topic mastered")
	}
	if topic := p.Topics[masteredTopic]; topic.MasteredSince != nil || topic.MasteredLevel != nil {
		t.Errorf("mastered since %v at %v after two wrong answers in a row, want both cleared", topic.MasteredSince, topic.MasteredLevel)
	}
}

// A topic needs five answers behind it before anything is declared, however
// sure the estimate is: a lucky start must not master a topic on the third
// task. Then a right and unaided answer masters it — and a right answer with
// the hint, or a wrong one, does not.
func TestMasteryWaitsForFiveAnswersAndARightUnaidedOne(t *testing.T) {
	t.Parallel()

	easy := rating.Point{GradeLevel: rating.Grades12, Difficulty: 3}
	for _, tc := range []struct {
		name          string
		correct, hint bool
		mastered      bool
	}{
		{"right and unaided", true, false, true},
		{"right with the hint", true, true, false},
		{"wrong", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := settled(t, 3, 100, 0, 0)
			for number := 1; number < profile.MasteryAnswers; number++ {
				if recorded := answerAt(t, p, easy, number, true, false); recorded.Mastered {
					t.Fatalf("answer %d mastered the topic, before it had %d behind it", number, profile.MasteryAnswers)
				}
			}
			fifth := answerAt(t, p, easy, profile.MasteryAnswers, tc.correct, tc.hint)
			if fifth.Mastered != tc.mastered {
				t.Errorf("the fifth answer mastered the topic: %v, want %v", fifth.Mastered, tc.mastered)
			}
		})
	}
}

// Mastery is held at the level declared and rises only: a right answer that
// clears the level held again adds nothing; one to a task of a higher level
// whose middle task the cautious level clears masters the topic there, with a
// new day; and a right answer to a task below the level held moves nothing
// down.
func TestMasteryIsHeldAtItsLevelAndOnlyRises(t *testing.T) {
	t.Parallel()

	p := settled(t, 4.5, 200, 0, 9)
	since, level := profile.DateOf(issued), rating.Grades12
	topic := p.Topics[masteredTopic]
	topic.MasteredSince, topic.MasteredLevel = &since, &level
	p.Topics[masteredTopic] = topic

	if again := answerAt(t, p, rating.Point{GradeLevel: rating.Grades12, Difficulty: 3}, 1, true, false); again.Mastered {
		t.Error("a right answer at the level held mastered the topic again")
	}
	wantMasteredAt(t, p, rating.Grades12, issued)

	if higher := answerAt(t, p, rating.Point{GradeLevel: rating.Grades34, Difficulty: 3}, 30, true, false); !higher.Mastered {
		t.Fatal("a right answer at 3-4 the cautious level clears did not master the topic there")
	}
	wantMasteredAt(t, p, rating.Grades34, issued.Add(30*time.Hour))

	if lower := answerAt(t, p, rating.Point{GradeLevel: rating.Grades12, Difficulty: 3}, 31, true, false); lower.Mastered {
		t.Error("a right answer below the level held mastered the topic again")
	}
	wantMasteredAt(t, p, rating.Grades34, issued.Add(30*time.Hour))
}

// mastered is a child whose topic the cautious estimate has mastered, at the
// level of their grade.
func mastered(t *testing.T) *profile.Profile {
	t.Helper()

	p := settled(t, 3, 100, 0, 10)
	since, level := profile.DateOf(issued), rating.Grades34
	topic := p.Topics[masteredTopic]
	topic.MasteredSince, topic.MasteredLevel = &since, &level
	p.Topics[masteredTopic] = topic
	return p
}

// One wrong answer is a slip. Two in a row are not, and the topic goes back
// into the rotation.
func TestMasteryIsLostByTwoWrongAnswersAndNotOne(t *testing.T) {
	t.Parallel()

	p := mastered(t)
	if slip := answerHard(t, p, 1, false, false); slip.Unmastered {
		t.Fatal("one wrong answer undid the mastery, want a slip to cost nothing")
	}
	if p.Topics[masteredTopic].MasteredSince == nil {
		t.Fatal("mastered_since was cleared by one wrong answer")
	}

	second := answerHard(t, p, 2, false, false)
	if !second.Unmastered {
		t.Error("two wrong answers in a row left the topic mastered")
	}
	if p.Topics[masteredTopic].MasteredSince != nil || p.Topics[masteredTopic].MasteredLevel != nil {
		t.Error("mastered_since or mastered_level survived two wrong answers in a row")
	}
	if got := p.Topics[masteredTopic].WrongStreak; got != profile.MasteryLostAfter {
		t.Errorf("wrong_streak = %d, want %d", got, profile.MasteryLostAfter)
	}
}

// Any right answer ends a run of failures, an easy one too: the child did not
// get it wrong twice in a row, which is all that losing mastery is about.
func TestARightAnswerBreaksTheRunOfFailures(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		right func(t *testing.T, p *profile.Profile, number int) profile.Recorded
	}{
		{"a hard one", func(t *testing.T, p *profile.Profile, number int) profile.Recorded {
			t.Helper()
			return answerHard(t, p, number, true, false)
		}},
		{"an easy one", func(t *testing.T, p *profile.Profile, number int) profile.Recorded {
			t.Helper()
			return answerEasy(t, p, number, true)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := mastered(t)
			answerHard(t, p, 1, false, false)
			tc.right(t, p, 2)
			if got := p.Topics[masteredTopic].WrongStreak; got != 0 {
				t.Errorf("wrong_streak = %d after a right answer, want it cleared", got)
			}
			if lost := answerHard(t, p, 3, false, false); lost.Unmastered {
				t.Error("the topic was unmastered by two failures with a right answer between them")
			}
		})
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

// top_streak counts what the earlier rule counted, for a build of that rule
// reading the file: right answers in a row at the middle of the corridor or
// harder and unaided, held where they are by an easy or a hinted right answer,
// ended by a wrong one, and spent once three of them complete in a topic of
// five answers or more. A task of 5-6 is hard for a child of grade 4 however
// the first answers move them, and one of 1-2 is easy.
func TestTopStreakCountsWhatTheEarlierRuleCounted(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "sasha")
	hard := rating.Point{GradeLevel: rating.Grades56, Difficulty: 5}
	easy := rating.Point{GradeLevel: rating.Grades12, Difficulty: 1}
	for number, step := range []struct {
		name          string
		point         rating.Point
		correct, hint bool
		want          int
	}{
		{"a hard right answer", hard, true, false, 1},
		{"a second one", hard, true, false, 2},
		{"an easy right answer", easy, true, false, 2},
		{"a hard right answer with the hint", hard, true, true, 2},
		{"a third hard one, in a topic of five answers", hard, true, false, 0},
		{"a hard right answer after it", hard, true, false, 1},
		{"a wrong answer", hard, false, false, 0},
	} {
		answerAt(t, p, step.point, number+1, step.correct, step.hint)
		if got := p.Topics[masteredTopic].TopStreak; got != step.want {
			t.Errorf("top_streak = %d after %s, want %d", got, step.name, step.want)
		}
	}
}

// A file the earlier rule wrote holds its masteries, two in three of them
// declared too soon. They count for nothing, and the first answer recorded
// into the file clears every one and marks the file — before that answer is
// judged, so the answer can master a topic by the cautious estimate itself.
func TestTheMasteriesOfTheEarlierRuleAreCleared(t *testing.T) {
	t.Parallel()

	p := settled(t, 3, 100, 0, 4)
	p.Ratings.MasteryRule = ""
	since, level := profile.DateOf(issued), rating.Grades34
	p.Topics["logic.ordering"] = profile.Topic{Answers: 6, Correct: 6, MasteredSince: &since, MasteredLevel: &level, Traps: map[string]int{}}
	if p.MasteriesStand() {
		t.Fatal("a file not marked counts its masteries, want them counted for nothing")
	}

	recorded := answerAt(t, p, rating.Point{GradeLevel: rating.Grades12, Difficulty: 3}, 1, true, false)
	if !p.MasteriesStand() || p.Ratings.MasteryRule != profile.MasteryRuleCautious {
		t.Errorf("mastery_rule = %q after an answer, want %q", p.Ratings.MasteryRule, profile.MasteryRuleCautious)
	}
	if earlier := p.Topics["logic.ordering"]; earlier.MasteredSince != nil || earlier.MasteredLevel != nil {
		t.Errorf("the earlier rule's mastery survived the first answer: since %v at %v", earlier.MasteredSince, earlier.MasteredLevel)
	}
	if !recorded.Mastered {
		t.Error("the answer that cleared the file did not master its own topic, want it judged by the cautious estimate")
	}
}

// A profile starts as the cautious estimate's: there is nothing in it to
// clear.
func TestANewProfileKeepsTheCautiousEstimatesMasteries(t *testing.T) {
	t.Parallel()

	p := profile.New(profile.Student{Pseudonym: "Comet", Grade: 3}, "test", issued)
	if !p.MasteriesStand() {
		t.Errorf("a new profile's mastery_rule is %q, want %q", p.Ratings.MasteryRule, profile.MasteryRuleCautious)
	}
}
