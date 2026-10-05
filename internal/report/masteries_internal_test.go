package report

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// masteredLine is the line of a topic shown as mastered, written in the call
// of the answer that earned it.
func masteredLine(user, call, topic string, minute int) string {
	return fmt.Sprintf(`{"message":"topic_mastered","time":%q,"topic":%q,"instructions_version":"v1","request_id":%q,"user":%q}`,
		lessonMinute(minute), topic, call, user)
}

// lesson is the lines of a child's answers in a topic, one a minute from the
// minute given, each in a call of its own the host made: first a right answer
// the topic was shown as mastered on, then an answer for each outcome — R
// right, W wrong, H right with the hint, h wrong with the hint, ? "I don't
// know".
func lesson(host, user, topic string, minute int, outcomes string) []string {
	shown := fmt.Sprintf("%s-%s-%d", user, topic, minute)
	lines := []string{
		answerAt{user: user, call: shown, topic: topic, minute: minute, correct: true}.line(),
		masteredLine(user, shown, topic, minute),
		callLine(shown, host),
	}
	for i, outcome := range outcomes {
		answered := answerAt{
			user: user, call: fmt.Sprintf("%s-%s-%d", user, topic, minute+1+i), topic: topic, minute: minute + 1 + i,
			correct: outcome == 'R' || outcome == 'H', hint: outcome == 'H' || outcome == 'h', unknown: outcome == '?',
		}
		lines = append(lines, answered.line(), callLine(answered.call, host))
	}
	return lines
}

// claude is the group the lessons of these tests are counted in.
var claude = group{version: "v1", host: "claude"}

// A mastery is taken back as the service takes it back, at the second wrong
// answer in a row in its topic: a wrong answer with the hint and "I don't
// know" are wrong answers, a right answer with the hint ends the run as any
// right answer does, and the answers after it was taken back are not
// followed. One never taken back is followed to the end of the lines.
func TestAMasteryIsTakenBackAtTheSecondWrongAnswerInARow(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, outcomes string
		want           shownMastery
	}{
		{"two wrong in a row", "RWW", shownMastery{child: "u1", followed: 3, takenBackAt: 3}},
		{"a right answer between", "WRWR", shownMastery{child: "u1", followed: 4}},
		{"wrong with the hint", "hW", shownMastery{child: "u1", followed: 2, takenBackAt: 2}},
		{"I don't know", "W?", shownMastery{child: "u1", followed: 2, takenBackAt: 2}},
		{"right with the hint between", "WHW", shownMastery{child: "u1", followed: 3}},
		{"answers after it was taken back", "WWRRW", shownMastery{child: "u1", followed: 2, takenBackAt: 2}},
		{"no answer after it", "", shownMastery{child: "u1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tallied(t, lesson("claude", "u1", "counting.gaps", 0, tc.outcomes)...).masteries[claude]
			if !slices.Equal(got, []shownMastery{tc.want}) {
				t.Errorf("after %q the masteries are %+v, want %+v", tc.outcomes, got, tc.want)
			}
		})
	}
}

// Only the child's answers in the topic follow its mastery: another topic of
// the same child's, and the same topic of another child's, take nothing back.
func TestOnlyTheChildsAnswersInTheTopicFollowItsMastery(t *testing.T) {
	t.Parallel()

	lines := lesson("claude", "u1", "counting.gaps", 0, "RRW")
	for minute, user := range map[int]string{10: "u1", 20: "u2"} {
		topic := "parity.alternation"
		if user == "u2" {
			topic = "counting.gaps"
		}
		for i := range 2 {
			wrong := answerAt{user: user, call: fmt.Sprintf("w%d-%d", minute, i), topic: topic, minute: minute + i}
			lines = append(lines, wrong.line(), callLine(wrong.call, "claude"))
		}
	}

	want := []shownMastery{{child: "u1", followed: 3}}
	if got := tallied(t, lines...).masteries[claude]; !slices.Equal(got, want) {
		t.Errorf("the masteries are %+v, want %+v", got, want)
	}
}

// A topic shown as mastered again, at a higher level, ends the mastery before
// it, which was not taken back, and is followed from there on its own.
func TestATopicShownAgainEndsTheMasteryBeforeIt(t *testing.T) {
	t.Parallel()

	lines := slices.Concat(lesson("claude", "u1", "counting.gaps", 0, "RW"), lesson("claude", "u1", "counting.gaps", 3, "WW"))
	want := []shownMastery{{child: "u1", followed: 3}, {child: "u1", followed: 2, takenBackAt: 2}}
	if got := tallied(t, lines...).masteries[claude]; !slices.Equal(got, want) {
		t.Errorf("the masteries are %+v, want %+v", got, want)
	}
}

// A line of a topic mastered whose answer is not among the lines shows
// nothing: the answer it is read with is the one that earned it.
func TestALineOfAMasteryWithoutItsAnswerShowsNothing(t *testing.T) {
	t.Parallel()

	lines := []string{masteredLine("u1", "lost", "counting.gaps", 0)}
	for i := range 3 {
		wrong := answerAt{user: "u1", call: fmt.Sprintf("w%d", i), topic: "counting.gaps", minute: 1 + i}
		lines = append(lines, wrong.line(), callLine(wrong.call, "claude"))
	}
	if got := tallied(t, lines...).masteries; len(got) != 0 {
		t.Errorf("the masteries are %+v, want none", got)
	}
}

// The answers are followed in the order they were given, whatever order the
// lines come in: the log hands them over newest first.
func TestAMasteryIsFollowedInTheOrderTheAnswersWereGiven(t *testing.T) {
	t.Parallel()

	lines := lesson("claude", "u1", "counting.gaps", 0, "WRWW")
	slices.Reverse(lines)
	want := []shownMastery{{child: "u1", followed: 4, takenBackAt: 4}}
	if got := tallied(t, lines...).masteries[claude]; !slices.Equal(got, want) {
		t.Errorf("the masteries are %+v, want %+v", got, want)
	}
}

// What the load tool or MCP Inspector did is left out — their calls, and a
// child either of them handed a task to, whichever host its answers came
// through —, and a line the log repeated is followed once.
func TestTestingIsLeftOutAndARepeatIsFollowedOnce(t *testing.T) {
	t.Parallel()

	lines := slices.Concat(
		lesson("load", "u2", "counting.gaps", 0, "WW"),
		lesson("claude", "tester", "counting.gaps", 0, "WW"),
		[]string{`{"message":"task_accepted","host":"inspector","user":"tester","request_id":"t1","instructions_version":"v1"}`},
	)
	for _, l := range lesson("claude", "u1", "counting.gaps", 0, "WRW") {
		lines = append(lines, l, l)
	}

	c := tallied(t, lines...)
	want := []shownMastery{{child: "u1", followed: 3}}
	if got := c.masteries[claude]; len(c.masteries) != 1 || !slices.Equal(got, want) {
		t.Errorf("the masteries are %+v, want only %+v", c.masteries, want)
	}
}

// The share taken back counts a mastery the lines stopped following early for
// the answers it was followed through. Of four masteries, one is taken back at
// the second answer, one followed through two answers alone, one taken back
// at the fourth and one followed through all ten: three of four are held past
// the second answer and one of the two left past the fourth, so 5/8 are taken
// back. Leaving the one followed early out would give 2/3, and counting it as
// held, 1/2.
func TestTheShareTakenBackCountsAMasteryForTheAnswersItWasFollowed(t *testing.T) {
	t.Parallel()

	masteries := []shownMastery{
		{child: "u1", followed: 2, takenBackAt: 2},
		{child: "u2", followed: 2},
		{child: "u3", followed: 4, takenBackAt: 4},
		{child: "u4", followed: 10},
	}
	if got := 1 - heldThrough(masteries, takenBackBy); math.Abs(got-5.0/8) > 1e-12 {
		t.Errorf("the share taken back by the %dth answer is %.4f, want %.4f", takenBackBy, got, 5.0/8)
	}
	if got := settledBy(masteries, takenBackBy); got != 3 {
		t.Errorf("%d masteries are settled by the %dth answer, want 3", got, takenBackBy)
	}
}

// The error of the share taken back is counted by child, since one child's
// masteries lean together. Two children with two masteries each, one child's
// both taken back and the other's both held, give an error of 0.5, the spread
// of two children; counted one mastery at a time, it would be √(1/12) =
// 0.289. One child's masteries give no error at all.
func TestTheErrorOfTheShareTakenBackIsCountedByChild(t *testing.T) {
	t.Parallel()

	masteries := []shownMastery{
		{child: "u1", followed: 2, takenBackAt: 2}, {child: "u1", followed: 3, takenBackAt: 3},
		{child: "u2", followed: 10}, {child: "u2", followed: 12},
	}
	if got, read := takenBackError(masteries); !read || math.Abs(got-0.5) > 1e-12 {
		t.Errorf("the error by child is %.4f (read: %t), want 0.5", got, read)
	}

	byMastery := slices.Clone(masteries)
	for i := range byMastery {
		byMastery[i].child = strconv.Itoa(i)
	}
	if got, _ := takenBackError(byMastery); math.Abs(got-math.Sqrt(1.0/12)) > 1e-12 {
		t.Errorf("the error by mastery is %.4f, want %.4f", got, math.Sqrt(1.0/12))
	}
	if _, read := takenBackError(masteries[:2]); read {
		t.Error("one child's masteries gave an error, want none")
	}
}

// A share that fewer masteries settle than the table reads says so, and one
// that as many settle shows it, however many more were shown and not
// followed far enough to settle.
func TestAShareOfTooFewSettledMasteriesSaysSo(t *testing.T) {
	t.Parallel()

	for _, settled := range []int{fewestMasteries - 1, fewestMasteries} {
		var masteries []shownMastery
		for i := range settled {
			m := shownMastery{child: fmt.Sprintf("u%d", i%3), followed: takenBackBy + i%3}
			if i%2 == 0 {
				m.followed, m.takenBackAt = 2+i%9, 2+i%9
			}
			masteries = append(masteries, m)
		}
		for i := range 5 {
			masteries = append(masteries, shownMastery{child: fmt.Sprintf("u%d", i), followed: 3})
		}
		c := &counts{masteries: map[group][]shownMastery{claude: masteries}}
		row := c.masteriesTable().rows[0]
		saysTooFew := row[8] == tooFew && row[9] == tooFew
		if wantTooFew := settled < fewestMasteries; saysTooFew != wantTooFew {
			t.Errorf("%d settled masteries: the row reads %v, want too few: %t", settled, row, wantTooFew)
		}
	}
}

// However the answers come, every mastery shown is counted once among those
// taken back early, in the middle, late, or not at all; one taken back was
// taken back at the last answer it was followed through, and no sooner than
// the run that takes it back.
func TestEveryMasteryShownIsCountedOnceAndTakenBackByARun(t *testing.T) {
	t.Parallel()

	type taught struct{ user, topic, outcomes string }
	lessons := gen.SliceOf(gopter.CombineGens(
		gen.OneConstOf("u1", "u2", "u3"), gen.OneConstOf("counting.gaps", "parity.alternation"),
		gen.SliceOf(gen.OneConstOf('R', 'W', 'H', 'h', '?')),
	).Map(func(drawn []any) taught {
		user, _ := drawn[0].(string)
		topic, _ := drawn[1].(string)
		outcomes, _ := drawn[2].([]rune)
		return taught{user, topic, string(outcomes)}
	}))
	properties := gopter.NewProperties(nil)
	properties.Property("each mastery is counted once, and taken back only by a run of wrong answers", prop.ForAll(
		func(drawn []taught) bool {
			var lines []string
			minute := 0
			for _, each := range drawn {
				lines = append(lines, lesson("claude", each.user, each.topic, minute, each.outcomes)...)
				minute += len(each.outcomes) + 1
			}
			read, err := readAll(strings.NewReader(strings.Join(lines, "\n")))
			if err != nil {
				return false
			}
			c := tally(read)
			for _, m := range c.masteries[claude] {
				if m.takenBackAt != 0 && (m.takenBackAt != m.followed || m.takenBackAt < profile.MasteryLostAfter) {
					return false
				}
			}
			return len(c.masteries[claude]) == len(drawn) && countsAddUp(c.masteriesTable())
		},
		lessons,
	))
	properties.TestingRun(t)
}

// countsAddUp says whether every row of the table of masteries counts each
// mastery shown once, among those taken back early, in the middle, late or
// not at all.
func countsAddUp(t *table) bool {
	for _, row := range t.rows {
		shown, err := strconv.Atoi(row[2])
		if err != nil {
			return false
		}
		for _, cell := range row[4:8] {
			counted, err := strconv.Atoi(cell)
			if err != nil {
				return false
			}
			shown -= counted
		}
		if shown != 0 {
			return false
		}
	}
	return true
}
