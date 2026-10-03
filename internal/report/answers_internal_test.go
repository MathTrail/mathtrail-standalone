package report

import (
	"fmt"
	"maps"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// answerLine is the line about an answer as the service writes it, with what
// weighing it against its chance reads: whether it was right, the chance, who
// chose the task, which answer of the trial series it was, the range of the
// child's answers it fell in — none in the series — and whether it came after
// the hint.
func answerLine(chance float64, correct bool, chooser string, trial int, answers string, hint bool) string {
	inRange := ""
	if answers != "" {
		inRange = fmt.Sprintf(`,"answers_bucket":%q`, answers)
	}
	return fmt.Sprintf(`{"message":"answer_recorded","correct":%t,"hint_used":%t,"chance":%v,"tutor_mode":%q,"trial":%d%s,"instructions_version":"v1"}`,
		correct, hint, chance, chooser, trial, inRange)
}

// A chance falls in the range that holds it to two places, as the line writes
// it: the corridor's two ranges meet at its middle, and a chance past the last
// range is counted in it.
func TestAChanceFallsInItsRange(t *testing.T) {
	t.Parallel()

	for chance, want := range map[float64]string{
		0.2: "under 0.50", 0.49: "under 0.50", 0.5: "0.50-0.59", 0.69: "0.60-0.69",
		0.7: "0.70-0.77", 0.77: "0.70-0.77", 0.78: "0.78-0.85", 0.85: "0.78-0.85",
		0.86: "over 0.85", 1: "over 0.85", 1.2: "over 0.85",
	} {
		if got := chanceRanges[rangeOfChance(chance)].named; got != want {
			t.Errorf("a chance of %v falls in %s, want %s", chance, got, want)
		}
	}
}

// Only an answer to a task the rule chose, after the trial series and without
// the hint, is weighed against its chance: the model's tasks, a task whose
// chooser is not known, the trial series, the hint, and a line that carries no
// chance or no range of answers — one written before lines carried them, or one
// cut short — are left out of both tables. A whole number reads the
// same as Cloud Logging gives it back, 0.0 for 0.
func TestOnlyTheRulesAnswersAfterTheSeriesWithoutTheHintAreWeighed(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		answerLine(0.8, true, "rule", 0, "6-20", false),
		`{"message":"answer_recorded","correct":false,"hint_used":false,"chance":0.8,"tutor_mode":"rule","trial":0.0,"answers_bucket":"6-20","instructions_version":"v1"}`,
		answerLine(0.8, true, "llm", 0, "6-20", false),
		answerLine(0.8, true, "unknown", 0, "6-20", false),
		answerLine(0.6, false, "rule", 3, "", false),
		answerLine(0.8, true, "rule", 0, "6-20", true),
		`{"message":"answer_recorded","correct":true,"hint_used":false,"instructions_version":"v1"}`,
		`{"message":"answer_recorded","correct":true,"hint_used":false,"chance":0.8,"tutor_mode":"rule","trial":0,"instructions_version":"v1"}`,
	)
	g := group{version: "v1", host: callNotRead}
	weighed := c.promises[promisedIn{g, rangeOfChance(0.8)}]
	kept := c.keptUp[keptUpIn{g, "6-20"}]
	if weighed == nil || weighed.answers != 2 || weighed.right != 1 || kept == nil || kept.answers != 2 {
		t.Errorf("weighed %+v and %+v, want the rule's two answers in each table", weighed, kept)
	}
	want := map[string]int{leftModel: 1, leftUnknown: 1, leftTrial: 1, leftHint: 1, leftOld: 2}
	if !maps.Equal(c.answersLeftOut, want) {
		t.Errorf("left out %v, want %v", c.answersLeftOut, want)
	}
}

// However the answers come, each is counted once in each table or left out of
// both: the tables lose no answer and count none twice.
func TestEveryAnswerIsWeighedOnceInEachTableOrLeftOut(t *testing.T) {
	t.Parallel()

	answer := gopter.CombineGens(
		gen.Float64Range(0.2, 1), gen.Bool(), gen.OneConstOf("rule", "llm", "unknown", ""),
		gen.IntRange(0, 5), gen.OneConstOf("", "6-20", "21-50", "201+"), gen.OneConstOf(false, true),
	).Map(func(drawn []any) string {
		chance, _ := drawn[0].(float64)
		correct, _ := drawn[1].(bool)
		chooser, _ := drawn[2].(string)
		trial, _ := drawn[3].(int)
		answers, _ := drawn[4].(string)
		hint, _ := drawn[5].(bool)
		return answerLine(math.Round(chance*100)/100, correct, chooser, trial, answers, hint)
	})
	properties := gopter.NewProperties(nil)
	properties.Property("each answer is in one cell of each table, or left out of both", prop.ForAll(
		func(lines []string) bool {
			read, err := readAll(strings.NewReader(strings.Join(lines, "\n")))
			if err != nil || len(read.lines) != len(lines) {
				return false
			}
			c := tally(read)
			byChance, byAnswers, leftOut := 0, 0, 0
			for _, cell := range c.promises {
				byChance += cell.answers
			}
			for _, cell := range c.keptUp {
				byAnswers += cell.answers
			}
			for _, n := range c.answersLeftOut {
				leftOut += n
			}
			return byChance+leftOut == len(lines) && byAnswers+leftOut == len(lines)
		},
		gen.SliceOf(answer),
	))
	properties.TestingRun(t)
}

// Answers that come out right more often than they were promised, by the same
// margin whatever the chance, show that margin: the mean difference is within
// three standard errors of it. The draws are seeded, so the case is the same
// every time rather than wrong one time in a few hundred.
func TestAnswersRightMoreOftenThanPromisedShowByHowMuch(t *testing.T) {
	t.Parallel()

	for i, margin := range []float64{-0.1, -0.05, 0, 0.05, 0.1} {
		random := rand.New(rand.NewPCG(uint64(i)+1, 2026))
		c := weighing()
		g := group{version: "v1", host: "claude"}
		for answer := range 4000 {
			chance := math.Round((0.6+0.3*random.Float64())*100) / 100
			answered := line{
				User: fmt.Sprintf("child-%d", answer%200), Correct: random.Float64() < chance+margin, Chance: &chance,
				TutorMode: string(profile.TutorRule), AnswersBucket: "6-20",
			}
			c.answer(&answered, g)
		}
		kept := c.keptUp[keptUpIn{g, "6-20"}]
		standardError, read := kept.standardError()
		if off := math.Abs(kept.mean() - margin); !read || off > 3*standardError {
			t.Errorf("answers right %+.2f more often than promised show %+.3f, %.1f standard errors off",
				margin, kept.mean(), off/standardError)
		}
	}
}

// A cell of fewer answers than the tables read says so in place of its
// numbers, and a cell of as many shows them.
func TestACellOfTooFewAnswersSaysSo(t *testing.T) {
	t.Parallel()

	for _, answers := range []int{fewestAnswers - 1, fewestAnswers} {
		c := tallied(t, slices.Repeat([]string{answerLine(0.8, true, "rule", 0, "6-20", false)}, answers)...)
		promised, kept := c.promisesTable().rows[0], c.keptUpTable().rows[0]
		saysTooFew := slices.Equal(promised[5:], []string{tooFew, tooFew}) && slices.Equal(kept[5:], []string{tooFew, tooFew})
		if wantTooFew := answers < fewestAnswers; saysTooFew != wantTooFew {
			t.Errorf("%d answers: the cells read %v and %v, want too few: %t", answers, promised, kept, wantTooFew)
		}
	}
}

// The ranges of the child's answers come by the first answer each holds, as
// the service names them, and a range named otherwise comes after them.
func TestRangesOfAnswersComeInTheirOrder(t *testing.T) {
	t.Parallel()

	ranges := []string{"201+", "named otherwise", "21-50", "6-20", "101-200", "51-100"}
	slices.SortFunc(ranges, compareAnswerRanges)
	if want := []string{"6-20", "21-50", "51-100", "101-200", "201+", "named otherwise"}; !slices.Equal(ranges, want) {
		t.Errorf("the ranges come as %v, want %v", ranges, want)
	}
}

// A difference is written to two places with its sign, and one that rounds to
// nothing has none.
func TestADifferenceIsWrittenWithItsSign(t *testing.T) {
	t.Parallel()

	for difference, want := range map[float64]string{0.0125: "+0.01", -0.05: "-0.05", 0.004: "0.00", -0.004: "0.00", 0: "0.00"} {
		if got := signed(difference); got != want {
			t.Errorf("signed(%v) = %q, want %q", difference, got, want)
		}
	}
}

// The standard error of how the estimate keeps up is counted by child, since
// one child's answers lean together. Four children of 8 answers at a chance
// of 0.8, two of them wrong once and two twice, give a mean of +0.01 whose
// error is the spread of the four children's, 0.036, rather than the 0.070 the
// 32 answers would give counted one by one. One child's answers give no error
// at all, and the cell says so.
func TestTheStandardErrorIsCountedByChild(t *testing.T) {
	t.Parallel()

	g, chance := group{version: "v1", host: "claude"}, 0.8
	c := weighing()
	for child, wrong := range map[string]int{"u4": 1, "u5": 2, "u6": 1, "u7": 2} {
		for answer := range 8 {
			answered := line{
				User: child, Correct: answer >= wrong, Chance: &chance,
				TutorMode: string(profile.TutorRule), AnswersBucket: "6-20",
			}
			c.answer(&answered, g)
		}
	}
	if standardError, read := c.keptUp[keptUpIn{g, "6-20"}].standardError(); !read || math.Abs(standardError-0.0361) > 0.0005 {
		t.Errorf("the standard error by child is %.4f (read: %t), want 0.0361", standardError, read)
	}

	alone := weighing()
	for range 40 {
		answered := line{User: "u4", Correct: true, Chance: &chance, TutorMode: string(profile.TutorRule), AnswersBucket: "6-20"}
		alone.answer(&answered, g)
	}
	if _, read := alone.keptUp[keptUpIn{g, "6-20"}].standardError(); read {
		t.Error("one child's answers gave a standard error, want none")
	}
	if row := alone.keptUpTable().rows[0]; row[6] != oneChild {
		t.Errorf("the cell of one child reads %v, want %q for its error", row, oneChild)
	}
}

// weighing is an empty count to weigh answers into.
func weighing() *counts {
	return &counts{promises: map[promisedIn]*cameTrue{}, keptUp: map[keptUpIn]*keptUp{}, answersLeftOut: map[string]int{}}
}
