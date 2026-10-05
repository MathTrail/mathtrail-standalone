package report

import (
	"cmp"
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

// mastery is a mastery as a test states it: the child it was shown to, the
// answers that followed it, and the one it was taken back at, none when it was
// not.
func mastery(child string, answers, takenBackAt int) shownMastery {
	return shownMastery{child: child, Followed: Followed{Answers: answers, TakenBackAt: takenBackAt}}
}

// asStated is each mastery as a test states it, without the run of wrong
// answers it last stood at.
func asStated(masteries []shownMastery) []shownMastery {
	stated := make([]shownMastery, len(masteries))
	for i, m := range masteries {
		stated[i] = mastery(m.child, m.Answers, m.TakenBackAt)
	}
	return stated
}

// sortedMasteries are masteries in one order whatever order they were counted
// in.
func sortedMasteries(masteries []shownMastery) []shownMastery {
	sorted := asStated(masteries)
	slices.SortFunc(sorted, func(a, b shownMastery) int {
		return cmp.Or(cmp.Compare(a.child, b.child), cmp.Compare(a.Answers, b.Answers), cmp.Compare(a.TakenBackAt, b.TakenBackAt))
	})
	return sorted
}

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
		{"two wrong in a row", "RWW", mastery("u1", 3, 3)},
		{"a right answer between", "WRWR", mastery("u1", 4, 0)},
		{"wrong with the hint", "hW", mastery("u1", 2, 2)},
		{"I don't know", "W?", mastery("u1", 2, 2)},
		{"right with the hint between", "WHW", mastery("u1", 3, 0)},
		{"answers after it was taken back", "WWRRW", mastery("u1", 2, 2)},
		{"no answer after it", "", mastery("u1", 0, 0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tallied(t, lesson("claude", "u1", "counting.gaps", 0, tc.outcomes)...).masteries[claude]
			if !slices.Equal(asStated(got), []shownMastery{tc.want}) {
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

	want := []shownMastery{mastery("u1", 3, 0)}
	if got := tallied(t, lines...).masteries[claude]; !slices.Equal(asStated(got), want) {
		t.Errorf("the masteries are %+v, want %+v", got, want)
	}
}

// A topic shown as mastered again, at a higher level, ends the mastery before
// it, which was not taken back, and is followed from there on its own.
func TestATopicShownAgainEndsTheMasteryBeforeIt(t *testing.T) {
	t.Parallel()

	lines := slices.Concat(lesson("claude", "u1", "counting.gaps", 0, "RW"), lesson("claude", "u1", "counting.gaps", 3, "WW"))
	want := []shownMastery{mastery("u1", 3, 0), mastery("u1", 2, 2)}
	if got := tallied(t, lines...).masteries[claude]; !slices.Equal(asStated(got), want) {
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
	want := []shownMastery{mastery("u1", 4, 4)}
	if got := tallied(t, lines...).masteries[claude]; !slices.Equal(asStated(got), want) {
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
	want := []shownMastery{mastery("u1", 3, 0)}
	if got := c.masteries[claude]; len(c.masteries) != 1 || !slices.Equal(asStated(got), want) {
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
		mastery("u1", 2, 2),
		mastery("u2", 2, 0),
		mastery("u3", 4, 4),
		mastery("u4", 10, 0),
	}
	on := heldOnOf(masteries)
	if got := 1 - on.held(); math.Abs(got-5.0/8) > 1e-12 {
		t.Errorf("the share taken back by the %dth answer is %.4f, want %.4f", TakenBackBy, got, 5.0/8)
	}
	if got := settledBy(masteries, TakenBackBy); got != 3 {
		t.Errorf("%d masteries are settled by the %dth answer, want 3", got, TakenBackBy)
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
		mastery("u1", 2, 2), mastery("u1", 3, 3),
		mastery("u2", 10, 0), mastery("u2", 12, 0),
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
// followed far enough to settle; the row counts the settled ones either way.
func TestAShareOfTooFewSettledMasteriesSaysSo(t *testing.T) {
	t.Parallel()

	for _, settled := range []int{fewestMasteries - 1, fewestMasteries} {
		var masteries []shownMastery
		for i := range settled {
			m := mastery(fmt.Sprintf("u%d", i%3), TakenBackBy+i%3, 0)
			if i%2 == 0 {
				m.Answers, m.TakenBackAt = 2+i%9, 2+i%9
			}
			masteries = append(masteries, m)
		}
		for i := range 5 {
			masteries = append(masteries, mastery(fmt.Sprintf("u%d", i), 3, 0))
		}
		c := &counts{masteries: map[group][]shownMastery{claude: masteries}}
		row := c.masteriesTable().rows[0]
		saysTooFew := row[9] == tooFew && row[10] == tooFew
		if wantTooFew := settled < fewestMasteries; saysTooFew != wantTooFew || row[8] != strconv.Itoa(settled) {
			t.Errorf("%d settled masteries: the row reads %v, want them counted and too few: %t", settled, row, wantTooFew)
		}
	}
}

// Every version taken together counts each mastery once, under the version of
// the answer that showed it: one taken back by an answer under a later
// version is taken back under the version that showed it, and the masteries
// of every version are summed in the last row.
func TestEveryVersionCountsTheMasteriesOfEveryVersionOnce(t *testing.T) {
	t.Parallel()

	late := answerAt{user: "u3", call: "u3-late", topic: "counting.gaps", minute: 2}
	lines := slices.Concat(
		underVersion("v1", lesson("claude", "u1", "counting.gaps", 0, "WW")),
		underVersion("v1", lesson("chatgpt", "u3", "counting.gaps", 0, "W")),
		underVersion("v2", []string{late.line(), callLine(late.call, "chatgpt")}),
		underVersion("v2", lesson("claude", "u2", "counting.gaps", 10, "RRRRRRRRRR")),
	)

	var got [][9]string
	for _, row := range tallied(t, lines...).masteriesTable().rows {
		got = append(got, [9]string(row[:9]))
	}
	want := [][9]string{
		{"v1", "chatgpt", "1", "1", "1", "0", "0", "0", "1"},
		{"v1", "claude", "1", "1", "1", "0", "0", "0", "1"},
		{"v1", everyHost, "2", "2", "2", "0", "0", "0", "2"},
		{"v2", "claude", "1", "1", "0", "0", "0", "1", "1"},
		{"v2", everyHost, "1", "1", "0", "0", "0", "1", "1"},
		{everyVersion, everyHost, "3", "3", "2", "0", "0", "1", "3"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("the rows read %v, want %v", got, want)
	}
}

// The share taken back and its standard error are written to three places,
// as the bands they are read against are. Forty children with one mastery
// each — ten taken back at the second answer, ten followed through two
// answers alone, ten taken back at the fourth and ten followed through all
// ten — give a share of 3/4 · 1/2 taken from one, 0.625, and an error by child
// of 0.0931, which two places would write as 0.62 and 0.09.
func TestTheShareTakenBackAndItsErrorAreWrittenToThreePlaces(t *testing.T) {
	t.Parallel()

	var masteries []shownMastery
	for i := range 10 {
		masteries = append(masteries,
			mastery(fmt.Sprintf("a%d", i), 2, 2), mastery(fmt.Sprintf("b%d", i), 2, 0),
			mastery(fmt.Sprintf("c%d", i), 4, 4), mastery(fmt.Sprintf("d%d", i), 10, 0))
	}
	c := &counts{masteries: map[group][]shownMastery{claude: masteries}}
	if got := c.masteriesTable().rows[0][8:11]; !slices.Equal(got, []string{"30", "0.625", "0.093"}) {
		t.Errorf("the settled masteries, the share and its error read %v, want [30 0.625 0.093]", got)
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
				if m.TakenBackAt != 0 && (m.TakenBackAt != m.Answers || m.TakenBackAt < profile.MasteryLostAfter) {
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
