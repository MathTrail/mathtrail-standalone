package report

import (
	"fmt"
	"maps"
	"math"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// answerAt is an answer as a test writes its line: the child it counts, the
// call it came in, its topic, the minute it came at, which of the child's
// answers it was, and whether it was right, came after the hint, or was "I
// don't know". Its task was the rule's, at a chance of one half.
type answerAt struct {
	user, call, topic, answers string
	minute                     int
	correct, hint, unknown     bool
}

// line is the answer's line as the service writes it.
func (a answerAt) line() string {
	return fmt.Sprintf(`{"message":"answer_recorded","time":%q,"topic":%q,"correct":%t,"hint_used":%t,"confused":%t,`+
		`"chance":0.5,"tutor_mode":"rule","trial":0,"answers_bucket":%q,"instructions_version":"v1","request_id":%q,"user":%q}`,
		lessonMinute(a.minute), a.topic, a.correct, a.hint, a.unknown, a.answers, a.call, a.user)
}

// callLine is the line of a tool call the host made, which names the host
// for every line of the call.
func callLine(call, host string) string {
	return fmt.Sprintf(`{"message":"tool_call","tool":"submit_answer","outcome":"ok","duration_ms":80,"client":%q,"request_id":%q}`,
		host, call)
}

// lessonMinute is a moment so many minutes into a day of lessons, as a line
// writes it.
func lessonMinute(minute int) string {
	return time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC).Add(time.Duration(minute) * time.Minute).Format(time.RFC3339Nano)
}

// weighedAt weighs an answer of a child, right or wrong at a chance of one
// half, into the count, in a range of the child's answers.
func weighedAt(c *counts, g group, user, answers string, correct bool) {
	chance := 0.5
	answered := line{User: user, Correct: correct, Chance: &chance, TutorMode: string(profile.TutorRule), AnswersBucket: answers}
	c.answer(&answered, g)
}

// A range of the child's answers is earlier when it is wholly within the
// first fifty, later when it is wholly within the 101st to the 200th, and
// neither when it crosses a boundary, is open at its end or is named
// otherwise.
func TestARangeOfAnswersIsEarlierLaterOrNeither(t *testing.T) {
	t.Parallel()

	for named, want := range map[string]side{
		"6-20": earlier, "21-50": earlier, "51-100": neither, "101-200": later, "201+": neither,
		"40-60": neither, "150-250": neither, "named otherwise": neither, "": neither,
	} {
		if got := sideOf(named); got != want {
			t.Errorf("sideOf(%q) = %d, want %d", named, got, want)
		}
	}
}

// Only a child with answers in both ranges is set against itself: a child who
// stopped before the later range, one whose earlier answers fell before the
// lines began, and the answers between the two ranges are left out.
func TestOnlyAChildWithAnswersInBothRangesIsSetAgainstItself(t *testing.T) {
	t.Parallel()

	c, g := weighing(), group{version: "v1", host: "claude"}
	weighedAt(c, g, "both", "6-20", true)
	weighedAt(c, g, "both", "51-100", true)
	weighedAt(c, g, "both", "101-200", false)
	weighedAt(c, g, "earlier only", "21-50", false)
	weighedAt(c, g, "later only", "101-200", true)

	result := compared(c.pairsByGroup()[g])
	if result.children != 1 || result.earlier.answers != 1 || result.later.answers != 1 || result.difference != -1 {
		t.Errorf("compared %+v, want the one child with a right answer earlier and a wrong one later, 1 less", result)
	}
}

// The error of the later less the earlier is counted by child, a child's two
// ranges together. Three children answer four times in each range at a chance
// of one half: one right every time; one right once early on and twice later;
// one right once early on and never later. Each range comes out as promised,
// so the difference is nothing, and its error is √(1/48) = 0.144, where the
// errors of the two ranges taken as if from other children, 0.25 and 0.289,
// would give √(7/48) = 0.382: a child better than promised in one range leans
// that way in the other, and setting it against itself takes that out.
func TestTheLaterLessTheEarlierIsCountedByChild(t *testing.T) {
	t.Parallel()

	c, g := weighing(), group{version: "v1", host: "claude"}
	for user, rights := range map[string][2][4]bool{
		"u1": {{true, true, true, true}, {true, true, true, true}},
		"u2": {{true, false, false, false}, {true, true, false, false}},
		"u3": {{true, false, false, false}, {false, false, false, false}},
	} {
		for _, correct := range rights[0] {
			weighedAt(c, g, user, "6-20", correct)
		}
		for _, correct := range rights[1] {
			weighedAt(c, g, user, "101-200", correct)
		}
	}

	result := compared(c.pairsByGroup()[g])
	if result.difference != 0 || !result.hasError || math.Abs(result.standardError-math.Sqrt(1.0/48)) > 1e-12 {
		t.Errorf("compared %+v, want a difference of 0 and an error of %.4f", result, math.Sqrt(1.0/48))
	}
	earlierError, _ := c.keptUp[keptUpIn{g, "6-20"}].standardError()
	laterError, _ := c.keptUp[keptUpIn{g, "101-200"}].standardError()
	if apart := math.Hypot(earlierError, laterError); math.Abs(apart-math.Sqrt(7.0/48)) > 1e-12 {
		t.Errorf("the two ranges' errors taken apart give %.4f, want %.4f", apart, math.Sqrt(7.0/48))
	}
}

// Children who come out right more often than promised by one margin early on
// and by another later show the later less the earlier: the difference is
// within three standard errors of it, however each child leans on its own,
// which falls out of the difference. The draws are seeded, so the case is the
// same every time rather than wrong one time in a few hundred.
func TestTheLaterLessTheEarlierShowsHowMuchTheMarginMoved(t *testing.T) {
	t.Parallel()

	for i, margins := range [][2]float64{{0, 0}, {0, 0.06}, {0.05, -0.03}, {-0.04, 0.04}} {
		random := rand.New(rand.NewPCG(uint64(i)+1, 2026))
		c, g := weighing(), group{version: "v1", host: "claude"}
		for child := range 150 {
			lean := 0.1 * (random.Float64() - 0.5)
			for answer := range 40 {
				answers, margin := "6-20", margins[0]
				if answer >= 20 {
					answers, margin = "101-200", margins[1]
				}
				chance := math.Round((0.6+0.25*random.Float64())*100) / 100
				answered := line{
					User: fmt.Sprintf("child-%d", child), Correct: random.Float64() < chance+margin+lean, Chance: &chance,
					TutorMode: string(profile.TutorRule), AnswersBucket: answers,
				}
				c.answer(&answered, g)
			}
		}
		result := compared(c.pairsByGroup()[g])
		want := margins[1] - margins[0]
		if off := math.Abs(result.difference - want); !result.hasError || off > 3*result.standardError {
			t.Errorf("margins %v show %+.3f ± %.3f, want %+.2f within three standard errors",
				margins, result.difference, result.standardError, want)
		}
	}
}

// A later range promised more by the same amount for every answer comes out
// lower by that amount, and no less certain: the difference moves by the
// amount and its error stays, whatever the children's answers.
func TestAPromiseRaisedLaterLowersTheDifferenceByAsMuch(t *testing.T) {
	t.Parallel()

	inRange := gopter.CombineGens(gen.IntRange(0, 6), gen.Float64Range(-1, 1)).Map(func(drawn []any) ofChild {
		answers, _ := drawn[0].(int)
		mean, _ := drawn[1].(float64)
		return ofChild{answers: answers, sum: float64(answers) * mean}
	})
	child := gen.SliceOfN(2, inRange).Map(func(ranges []ofChild) *pair { return &pair{earlier: ranges[0], later: ranges[1]} })
	properties := gopter.NewProperties(nil)
	properties.Property("the difference moves by the amount the later promise rose, and its error stays", prop.ForAll(
		func(children []*pair, raised float64) bool {
			byChild, raisedByChild := map[string]*pair{}, map[string]*pair{}
			for i, p := range children {
				user := fmt.Sprintf("child-%d", i)
				byChild[user] = p
				moved := *p
				moved.later.sum -= raised * float64(p.later.answers)
				raisedByChild[user] = &moved
			}
			before, after := compared(byChild), compared(raisedByChild)
			if before.children == 0 {
				return after.children == 0
			}
			return math.Abs(after.difference-(before.difference-raised)) < 1e-9 &&
				after.hasError == before.hasError && math.Abs(after.standardError-before.standardError) < 1e-9
		},
		gen.SliceOf(child), gen.Float64Range(-0.3, 0.3),
	))
	properties.TestingRun(t)
}

// A range of fewer answers than the tables read says so in place of its
// number, and so does the difference it is part of; a range of as many shows
// them.
func TestAComparisonOfTooFewAnswersSaysSo(t *testing.T) {
	t.Parallel()

	for _, earlierAnswers := range []int{fewestAnswers - 1, fewestAnswers} {
		c, g := weighing(), group{version: "v1", host: "claude"}
		for answer := range earlierAnswers {
			weighedAt(c, g, fmt.Sprintf("u%d", answer%2), "6-20", answer%3 == 0)
		}
		for answer := range fewestAnswers {
			weighedAt(c, g, fmt.Sprintf("u%d", answer%2), "101-200", answer%2 == 0)
		}
		row := c.laterTable().rows[0]
		saysTooFew := row[4] == tooFew && row[7] == tooFew && row[8] == tooFew
		if wantTooFew := earlierAnswers < fewestAnswers; saysTooFew != wantTooFew || row[6] == tooFew {
			t.Errorf("%d earlier answers: the row reads %v, want the earlier side too few: %t", earlierAnswers, row, wantTooFew)
		}
	}
}

// The hosts of a version are set side by side and then taken together, and
// what the load tool or MCP Inspector did is left out: their calls, and a
// child either of them handed a task to, whichever host its answers came
// through.
func TestEveryHostTakesAVersionsHostsTogetherAndLeavesTestingOut(t *testing.T) {
	t.Parallel()

	var lines []string
	for _, child := range []struct{ host, user string }{
		{"claude", "u1"}, {"claude", "u2"}, {"chatgpt", "u3"}, {"load", "u4"}, {"claude", "tester"},
	} {
		for number := range 4 {
			for _, answered := range []answerAt{
				{user: child.user, answers: "6-20", minute: number, correct: number%2 == 0},
				{user: child.user, answers: "101-200", minute: 100 + number, correct: number%2 == 1},
			} {
				answered.call = fmt.Sprintf("%s-%d", child.user, answered.minute)
				answered.topic = "counting.gaps"
				lines = append(lines, answered.line(), callLine(answered.call, child.host))
			}
		}
	}
	lines = append(lines, `{"message":"task_accepted","host":"inspector","user":"tester","request_id":"t1","instructions_version":"v1"}`)

	children := map[string]string{}
	var hosts []string
	for _, row := range tallied(t, lines...).laterTable().rows {
		hosts = append(hosts, row[1])
		children[row[1]] = row[2]
	}
	want := map[string]string{"chatgpt": "1", "claude": "2", everyHost: "3"}
	if !slices.Equal(hosts, []string{"chatgpt", "claude", everyHost}) || !maps.Equal(children, want) {
		t.Errorf("the rows are of hosts %v with children %v, want %v in that order", hosts, children, want)
	}
}
