package report

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
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

// underVersion is lines as a test writes them, each under a version of the
// instructions of its own choosing rather than the one they name.
func underVersion(version string, lines []string) []string {
	moved := make([]string, len(lines))
	for i, l := range lines {
		moved[i] = strings.ReplaceAll(l, `"instructions_version":"v1"`, `"instructions_version":"`+version+`"`)
	}
	return moved
}

// sitting is the lines of four answers of a child in a range of its own,
// through a host and under a version of the instructions, one a minute from
// the minute given, right and wrong in turn.
func sitting(user, host, version, answers string, minute int) []string {
	var lines []string
	for number := range 4 {
		answered := answerAt{
			user: user, call: fmt.Sprintf("%s-%d", user, minute+number), topic: "counting.gaps", answers: answers,
			minute: minute + number, correct: number%2 == 0,
		}
		lines = append(lines, answered.line(), callLine(answered.call, host))
	}
	return underVersion(version, lines)
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

// A range of answers that holds answers on both sides of where one of the two
// ranges begins or ends crosses an edge, as one the service would write only
// if it counted in other ranges; the ranges it writes now cross none, and the
// table's words name one that does.
func TestARangeThatCrossesAnEdgeIsNamed(t *testing.T) {
	t.Parallel()

	for named, crosses := range map[string]bool{
		"6-20": false, "21-50": false, "51-100": false, "101-200": false, "201+": false, "named otherwise": false,
		"41-60": true, "91-110": true, "151-300": true, "151+": true,
	} {
		if got := crossesAnEdge(named); got != crosses {
			t.Errorf("crossesAnEdge(%q) = %t, want %t", named, got, crosses)
		}
	}

	c, g := weighing(), group{version: "v1", host: "claude"}
	for _, answers := range []string{"6-20", "151-300", "101-150", "41-60"} {
		weighedAt(c, g, "u1", answers, true)
	}
	if about := c.laterAbout(); !strings.HasSuffix(about, "which neither takes: 41-60 and 151-300.") {
		t.Errorf("the table's words end %q, want them to name 41-60 and 151-300", about[max(len(about)-80, 0):])
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

// The hosts of a version are set side by side and then taken together, every
// version taken together after them all, and what the load tool or MCP
// Inspector did is left out of every row: their calls, and a child either of
// them handed a task to, whichever host its answers came through.
func TestEveryHostAndThenEveryVersionComeLastAndTestingIsLeftOut(t *testing.T) {
	t.Parallel()

	var lines []string
	for _, child := range []struct {
		host, version, user string
		minute              int
	}{
		{"claude", "v2", "u1", 0}, {"claude", "v1", "u2", 10}, {"chatgpt", "v1", "u3", 20}, {"load", "v1", "u4", 30},
		{"claude", "v1", "tester", 40},
	} {
		lines = slices.Concat(lines,
			sitting(child.user, child.host, child.version, "6-20", child.minute),
			sitting(child.user, child.host, child.version, "101-200", 100+child.minute))
	}
	lines = append(lines, `{"message":"task_accepted","host":"inspector","user":"tester","request_id":"t1","instructions_version":"v1"}`)

	var got [][3]string
	for _, row := range tallied(t, lines...).laterTable().rows {
		got = append(got, [3]string{row[0], row[1], row[2]})
	}
	want := [][3]string{
		{"v2", "claude", "1"}, {"v2", everyHost, "1"}, {"v1", "chatgpt", "1"}, {"v1", "claude", "1"}, {"v1", everyHost, "2"},
		{everyVersion, everyHost, "3"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("the rows are %v, want %v", got, want)
	}
}

// Every version taken together sets a child against itself across versions:
// a child whose earlier answers came under one version and its later under
// another is set against itself there alone, whichever hosts the answers came
// through, and a version's own rows read only the answers given under it.
func TestEveryVersionSetsAChildAgainstItselfAcrossVersions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		lines []string
		want  [][5]string
	}{
		{
			"both ranges under one version",
			slices.Concat(sitting("u1", "claude", "v1", "6-20", 0), sitting("u1", "claude", "v1", "101-200", 100)),
			[][5]string{
				{"v1", "claude", "1", "4", "4"}, {"v1", everyHost, "1", "4", "4"}, {everyVersion, everyHost, "1", "4", "4"},
			},
		},
		{
			"the ranges under two versions",
			slices.Concat(sitting("u1", "claude", "v1", "6-20", 0), sitting("u1", "claude", "v2", "101-200", 100)),
			[][5]string{{everyVersion, everyHost, "1", "4", "4"}},
		},
		{
			"the ranges under two versions, through two hosts",
			slices.Concat(sitting("u1", "claude", "v1", "6-20", 0), sitting("u1", "chatgpt", "v2", "101-200", 100)),
			[][5]string{{everyVersion, everyHost, "1", "4", "4"}},
		},
		{
			"one earlier range under each version",
			slices.Concat(sitting("u1", "claude", "v1", "6-20", 0), sitting("u1", "claude", "v2", "21-50", 50),
				sitting("u1", "claude", "v2", "101-200", 100)),
			[][5]string{
				{"v2", "claude", "1", "4", "4"}, {"v2", everyHost, "1", "4", "4"}, {everyVersion, everyHost, "1", "8", "4"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got [][5]string
			for _, row := range tallied(t, tc.lines...).laterTable().rows {
				got = append(got, [5]string{row[0], row[1], row[2], row[3], row[5]})
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("the rows read %v, want %v", got, tc.want)
			}
		})
	}
}

// drawnSitting is a run of a child's answers in a topic as the property below
// draws it: the version, host and child, the range of the child's answers it
// falls in, whether each was right, and whether the first showed the topic
// mastered.
type drawnSitting struct {
	version, host, user, topic, answers string
	correct                             []bool
	shows                               bool
}

// linesOf are the lines of the sittings drawn, each answer at a minute and in
// a call of its own, with the task MCP Inspector handed the child it tests.
func linesOf(sittings []drawnSitting) []string {
	lines := []string{`{"message":"task_accepted","host":"inspector","user":"tester","request_id":"t1","instructions_version":"v1"}`}
	minute := 0
	for _, drawn := range sittings {
		var own []string
		for i, correct := range drawn.correct {
			answered := answerAt{
				user: drawn.user, call: fmt.Sprintf("c%d", minute), topic: drawn.topic, answers: drawn.answers,
				minute: minute, correct: correct,
			}
			own = append(own, answered.line(), callLine(answered.call, drawn.host))
			if i == 0 && drawn.shows {
				own = append(own, masteredLine(drawn.user, answered.call, drawn.topic, minute))
			}
			minute++
		}
		lines = append(lines, underVersion(drawn.version, own)...)
	}
	return lines
}

// inOneVersion are the sittings with every version made the first.
func inOneVersion(sittings []drawnSitting) []drawnSitting {
	one := slices.Clone(sittings)
	for i := range one {
		one[i].version = "v1"
	}
	return one
}

// Every version taken together reads the lines as if they had one version:
// whatever versions the answers and the masteries came under, the row of
// every version sets each child against itself, and counts the masteries,
// as one version's row over every host would were all the lines under it.
func TestEveryVersionReadsTheLinesAsIfTheyHadOneVersion(t *testing.T) {
	t.Parallel()

	oneSitting := gopter.CombineGens(
		gen.OneConstOf("v1", "v2", "v3"), gen.OneConstOf("claude", "chatgpt", "load"),
		gen.OneConstOf("u1", "u2", "u3", "tester"), gen.OneConstOf("counting.gaps", "parity.alternation"),
		gen.OneConstOf("6-20", "21-50", "51-100", "101-200"), gen.SliceOf(gen.Bool()), gen.Bool(),
	).Map(func(drawn []any) drawnSitting {
		version, _ := drawn[0].(string)
		host, _ := drawn[1].(string)
		user, _ := drawn[2].(string)
		topic, _ := drawn[3].(string)
		answers, _ := drawn[4].(string)
		correct, _ := drawn[5].([]bool)
		shows, _ := drawn[6].(bool)
		return drawnSitting{version, host, user, topic, answers, correct, shows}
	})
	parameters := gopter.DefaultTestParameters()
	parameters.MaxSize = 20
	properties := gopter.NewProperties(parameters)
	properties.Property("the row of every version is one version's row over every host", prop.ForAll(
		func(sittings []drawnSitting) bool {
			apart, err := readAll(strings.NewReader(strings.Join(linesOf(sittings), "\n")))
			if err != nil {
				return false
			}
			together, err := readAll(strings.NewReader(strings.Join(linesOf(inOneVersion(sittings)), "\n")))
			if err != nil {
				return false
			}
			asDrawn, asOne := tally(apart), tally(together)
			pooled, one := group{version: everyVersion, host: everyHost}, group{version: "v1", host: everyHost}
			return compared(asDrawn.pairsByGroup()[pooled]) == compared(asOne.pairsByGroup()[one]) &&
				slices.Equal(sortedMasteries(asDrawn.masteriesByGroup()[pooled]), sortedMasteries(asOne.masteriesByGroup()[one]))
		},
		gen.SliceOf(oneSitting),
	))
	properties.TestingRun(t)
}
