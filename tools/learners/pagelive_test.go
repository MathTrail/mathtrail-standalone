package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/report"
)

// fixtureSnapshot is the made-up snapshot of the fixture, read afresh.
func fixtureSnapshot(t *testing.T) liveSnapshot {
	t.Helper()
	var s liveSnapshot
	if err := readWholeJSON(fixtureLive, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

// A snapshot is taken only when it shows what the views' rule lets out and
// what a month could give: every cell on ten children and thirty answers or
// more, counted to five and to no more than its month; every range one the
// service counts in, once, and no two overlapping; and every share, margin
// and error one that answers give.
func TestTheLiveNumbersRefuseASnapshotTheRuleWouldNotShow(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		edit func(s *liveSnapshot)
	}{
		{"numbers of no month", func(s *liveSnapshot) { s.Month = nil }},
		{"a month of no calendar", func(s *liveSnapshot) { month := "2026-13"; s.Month = &month }},
		{"ranges of a month with no total", func(s *liveSnapshot) { s.Total = nil }},
		{"a month of too few children", func(s *liveSnapshot) { s.Total.Learners = 5 }},
		{"a range of too few children", func(s *liveSnapshot) { s.Chances[0].Learners = 5 }},
		{"a range of too few answers", func(s *liveSnapshot) { s.Chances[0].Answers = 25 }},
		{"children not counted to five", func(s *liveSnapshot) { s.KeptUp[0].Learners = 12 }},
		{"answers not counted to five", func(s *liveSnapshot) { s.Chances[1].Answers = 121 }},
		{"a range of more answers than its month", func(s *liveSnapshot) { s.Chances[2].Answers = 1240 }},
		{"a range of more children than its month", func(s *liveSnapshot) { s.KeptUp[3].Learners = 50 }},
		{"a range of chance the report has not", func(s *liveSnapshot) { s.Chances[0].Chance = "0.55-0.65" }},
		{"a range of chance twice", func(s *liveSnapshot) { s.Chances[1] = s.Chances[0] }},
		{"a promise outside its range", func(s *liveSnapshot) { s.Chances[1].PromisedMean = 0.7 }},
		{"a month's promise past certain", func(s *liveSnapshot) { s.Total.PromisedMean = 1.1 }},
		{"a share past all", func(s *liveSnapshot) { s.Chances[0].CorrectShare = 1.2 }},
		{"a share under nothing", func(s *liveSnapshot) { s.Total.CorrectShare = -0.1 }},
		{"a range of answers named otherwise", func(s *liveSnapshot) { s.KeptUp[0].AnswersRange = "101–200" }},
		{"a range that ends before it begins", func(s *liveSnapshot) { s.KeptUp[2].AnswersRange = "100-51" }},
		{"ranges that overlap", func(s *liveSnapshot) { s.KeptUp[1].AnswersRange = "20-50" }},
		{"a range of every answer before another", func(s *liveSnapshot) { s.KeptUp[1].AnswersRange = "21+" }},
		{"a margin past one", func(s *liveSnapshot) { s.KeptUp[0].CameTrueLessPromised = -1.5 }},
		{"an error under nothing", func(s *liveSnapshot) { s.KeptUp[0].StandardError = -0.01 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := fixtureSnapshot(t)
			tc.edit(&s)
			if live, err := liveBlockOf(&s); err == nil {
				t.Errorf("took the snapshot as %s, want it refused", live.State)
			}
		})
	}
	s := fixtureSnapshot(t)
	if _, err := liveBlockOf(&s); err != nil {
		t.Errorf("the fixture's snapshot: error %v, want it taken", err)
	}
}

// A snapshot of another shape is refused before its numbers are read: one
// with a key a snapshot has not, one that leaves a key out, which would read
// as nothing, and one that is not there or not one value of JSON.
func TestTheLiveNumbersRefuseASnapshotOfAnotherShape(t *testing.T) {
	t.Parallel()
	raw := string(readBytes(t, fixtureLive))
	dir := t.TempDir()
	for _, tc := range []struct{ name, text string }{
		{"a key a snapshot has not", strings.Replace(raw, `"month"`, `"deployment": "mathtrail", "month"`, 1)},
		{"a key left out", strings.Replace(raw, `"came_true_less_promised": 0.018,`, "", 1)},
		{"a key spelled otherwise", strings.Replace(raw, `"month"`, `"Month"`, 1)},
		{"a number set to null", strings.Replace(raw, `"standard_error": 0.041`, `"standard_error": null`, 1)},
		{"two values", raw + "{}"},
		{"no JSON", "month: 2026-11"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := fileOfText(t, dir, strings.ReplaceAll(tc.name, " ", "-")+".json", tc.text)
			if _, err := liveOf(path); !errors.Is(err, errLiveSnapshot) {
				t.Errorf("error %v, want the snapshot refused", err)
			}
		})
	}
	if _, err := liveOf(filepath.Join(dir, "none.json")); !errors.Is(err, errLiveSnapshot) {
		t.Errorf("a snapshot that is not there: error %v, want it refused", err)
	}
}

// The live numbers are still to come with no snapshot, and with one of no
// month counted whole; too few with one of a month that shows nothing; and
// shown with one that shows the month's total. The table of goals says which.
func TestTheLiveNumbersSayWhatStateTheyAreIn(t *testing.T) {
	t.Parallel()
	numbers := numbersFile(t, smallNumbers(t))
	dir := t.TempDir()
	for _, tc := range []struct {
		name, snapshot, state, month, said string
	}{
		{"no snapshot", "", liveComing, "",
			"The live numbers are still to come: no snapshot of them names a month counted whole."},
		{"no month counted whole", `{"month": null, "total": null, "chances": [], "kept_up": []}`, liveComing, "",
			"The live numbers are still to come: no snapshot of them names a month counted whole."},
		{"a month of too few children", `{"month": "2026-12", "total": null, "chances": [], "kept_up": []}`, liveTooFew, "2026-12",
			"The live numbers of 2026-12 show nothing: too few children answered for the rule of 10 children and 30 answers."},
		{"a month shown", string(readBytes(t, fixtureLive)), liveReady, "2026-11",
			"The live numbers are of 2026-11: 45 children and 1235 answers, with 4 of the 6 ranges of chance and 4 ranges of the child's answers shown."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := pageFileArgs{numbers: numbers, inputs: keyOfBuild, commit: fixtureCommit, date: fixtureDate, paperCommit: paperCommitOf}
			if tc.snapshot != "" {
				a.live = fileOfText(t, dir, strings.ReplaceAll(tc.name, " ", "-")+".json", tc.snapshot)
			}
			file, _, err := pageFileOf(&a)
			if err != nil {
				t.Fatal(err)
			}
			if month := monthOf(file.Live.Month); file.Live.State != tc.state || month != tc.month {
				t.Errorf("the live numbers are %s, of the month %q; want %s, of %q", file.Live.State, month, tc.state, tc.month)
			}
			if said := goalsReport(&file, ""); !strings.Contains(said, tc.said) {
				t.Errorf("the table of goals says %q, want it to say %q", said, tc.said)
			}
		})
	}
}

// monthOf is the month the live numbers are of, or nothing.
func monthOf(month *string) string {
	if month == nil {
		return ""
	}
	return *month
}

// A range of chance is named by the lowest and the highest chance it holds,
// and a range of the child's answers by its first answer and its last, with
// no last for the range of every answer past the others; both come from the
// lowest, whatever order the snapshot lists them in.
func TestTheLiveNumbersNameEachRangeByWhatItSpans(t *testing.T) {
	t.Parallel()
	s := fixtureSnapshot(t)
	s.Chances = append(s.Chances, snapshotChance{"under 0.50", liveCounts{10, 30}, liveWeighed{0.45, 0.4}})
	s.KeptUp = append([]snapshotKeptUp{{"201+", liveCounts{10, 30}, liveMargin{0.05, 0.06}}}, s.KeptUp...)
	live, err := liveBlockOf(&s)
	if err != nil {
		t.Fatal(err)
	}
	var chances, answers []string
	for _, c := range live.Chances {
		chances = append(chances, fmt.Sprintf("%v-%v", c.From, c.To))
	}
	for _, k := range live.KeptUp {
		last := "on"
		if k.Last != nil {
			last = strconv.Itoa(*k.Last)
		}
		answers = append(answers, fmt.Sprintf("%d-%s", k.First, last))
	}
	if got, want := strings.Join(chances, " "), "0-0.49 0.5-0.59 0.6-0.69 0.7-0.77 0.78-0.85"; got != want {
		t.Errorf("the ranges of chance are %s, want %s", got, want)
	}
	if got, want := strings.Join(answers, " "), "6-20 21-50 51-100 101-200 201-on"; got != want {
		t.Errorf("the ranges of answers are %s, want %s", got, want)
	}
}

// A snapshot taken shows nothing the rule hides: whatever a snapshot holds,
// once it is taken, every cell of it stands on the rule's children and answers
// or more, counted to the rule's multiple and to no more than its month, and
// its state is the one its month and its total give.
func TestALiveSnapshotTakenShowsNothingTheRuleHides(t *testing.T) {
	t.Parallel()
	taken, refused := 0, 0
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 2000
	properties := gopter.NewProperties(parameters)
	properties.Property("a snapshot taken shows no cell the rule hides, in the state its month and total give", prop.ForAll(
		func(s liveSnapshot) string {
			live, err := liveBlockOf(&s)
			if err != nil {
				refused++
				return ""
			}
			if len(live.Chances) > 0 || len(live.KeptUp) > 0 {
				taken++
			}
			return hiddenShown(&s, &live)
		},
		genLiveSnapshot(),
	))
	properties.TestingRun(t)
	if taken == 0 || refused == 0 {
		t.Errorf("of the snapshots drawn %d were taken with ranges and %d refused, want some of both, or the property shows nothing", taken, refused)
	}
}

// hiddenShown says how live numbers taken from a snapshot stand in a state
// its month and total do not give, or show a cell the rule hides, and is
// nothing when they do neither.
func hiddenShown(s *liveSnapshot, live *liveBlock) string {
	want := liveReady
	switch {
	case s.Month == nil:
		want = liveComing
	case s.Total == nil:
		want = liveTooFew
	}
	if live.State != want {
		return fmt.Sprintf("the live numbers are %s, want %s", live.State, want)
	}
	if live.Total == nil {
		return ""
	}
	month := live.Total.liveCounts
	cells := []liveCounts{month}
	for _, c := range live.Chances {
		cells = append(cells, c.liveCounts)
	}
	for _, k := range live.KeptUp {
		cells = append(cells, k.liveCounts)
	}
	for _, c := range cells {
		if c.Learners < privacyRule.Learners || c.Answers < privacyRule.Answers ||
			c.Learners%privacyRule.RoundedTo != 0 || c.Answers%privacyRule.RoundedTo != 0 ||
			c.Learners > month.Learners || c.Answers > month.Answers {
			return fmt.Sprintf("a cell of %d children and %d answers is shown in a month of %d and %d", c.Learners, c.Answers, month.Learners, month.Answers)
		}
	}
	return ""
}

// genLiveSnapshot is a snapshot of a month, now and then of none, with a
// total, now and then with none, and of up to three ranges of each kind, most
// of them named as the views name them and counted as the rule counts, and
// some of them not.
func genLiveSnapshot() gopter.Gen {
	months := gen.Weighted([]gen.WeightedGen{{Weight: 4, Gen: gen.Const("2026-11")}, {Weight: 1, Gen: gen.Const("")}})
	total := gopter.CombineGens(genCount(2, 30), genCount(6, 400), gen.Float64Range(0.4, 1), gen.Float64Range(0, 1.05)).
		Map(func(values []any) *liveTotal {
			learners, _ := values[0].(int)
			answers, _ := values[1].(int)
			promised, _ := values[2].(float64)
			right, _ := values[3].(float64)
			return &liveTotal{liveCounts{learners, answers}, liveWeighed{promised, right}}
		})
	totals := gen.Weighted([]gen.WeightedGen{{Weight: 4, Gen: total}, {Weight: 1, Gen: gen.Const((*liveTotal)(nil))}})
	chances := gopter.CombineGens(gen.IntRange(0, 3), gen.SliceOfN(3, genSnapshotChance()))
	keptUp := gopter.CombineGens(gen.IntRange(0, 3), gen.SliceOfN(3, genSnapshotKeptUp()))
	return gopter.CombineGens(months, totals, chances, keptUp).Map(func(values []any) liveSnapshot {
		var s liveSnapshot
		if month, _ := values[0].(string); month != "" {
			s.Month = &month
		}
		s.Total, _ = values[1].(*liveTotal)
		someChances, _ := values[2].([]any)
		n, _ := someChances[0].(int)
		drawnChances, _ := someChances[1].([]snapshotChance)
		s.Chances = drawnChances[:n]
		someKeptUp, _ := values[3].([]any)
		m, _ := someKeptUp[0].(int)
		drawnKeptUp, _ := someKeptUp[1].([]snapshotKeptUp)
		s.KeptUp = drawnKeptUp[:m]
		return s
	})
}

// genCount is a count of a cell: most often a multiple of five from five times
// least to five times most, and now and then any count up to that.
func genCount(least, most int) gopter.Gen {
	return gen.Weighted([]gen.WeightedGen{
		{Weight: 9, Gen: gen.IntRange(least, most).Map(func(n int) int { return 5 * n })},
		{Weight: 1, Gen: gen.IntRange(0, 5*most)},
	})
}

// genSnapshotChance is a range of chance named as the report names one, or
// now and then not, promising on average a chance inside it, with a share
// right that is now and then past all.
func genSnapshotChance() gopter.Gen {
	ranges := report.ChanceRanges()
	return gopter.CombineGens(gen.IntRange(0, len(ranges)), genCount(1, 10), genCount(5, 80), gen.Float64Range(0, 1), gen.Float64Range(0, 1.05)).
		Map(func(values []any) snapshotChance {
			at, _ := values[0].(int)
			learners, _ := values[1].(int)
			answers, _ := values[2].(int)
			within, _ := values[3].(float64)
			right, _ := values[4].(float64)
			if at == len(ranges) {
				return snapshotChance{"0.55-0.65", liveCounts{learners, answers}, liveWeighed{0.6, right}}
			}
			r := ranges[at]
			promised := (float64(r.From) + within*float64(r.To-r.From)) / 100
			return snapshotChance{r.Name, liveCounts{learners, answers}, liveWeighed{promised, right}}
		})
}

// genSnapshotKeptUp is a range of the child's answers named as the service
// names one, or now and then not.
func genSnapshotKeptUp() gopter.Gen {
	names := gen.OneConstOf("6-20", "21-50", "51-100", "101-200", "201+", "x")
	return gopter.CombineGens(names, genCount(1, 10), genCount(5, 80), gen.Float64Range(-0.3, 0.3), gen.Float64Range(0, 0.1)).
		Map(func(values []any) snapshotKeptUp {
			name, _ := values[0].(string)
			learners, _ := values[1].(int)
			answers, _ := values[2].(int)
			margin, _ := values[3].(float64)
			standardError, _ := values[4].(float64)
			return snapshotKeptUp{name, liveCounts{learners, answers}, liveMargin{margin, standardError}}
		})
}

// The rule the live numbers are held to is the rule the public views show
// them by: the same least children and answers, and the same multiple their
// counts are rounded to, as each view's SQL writes them.
func TestTheLiveRuleIsTheRuleOfThePublicViews(t *testing.T) {
	t.Parallel()
	for _, view := range []string{"chances.sql", "chances_total.sql", "kept_up.sql"} {
		wantTheLiveRule(t, view)
	}
}

// The rule as a view's SQL writes it: the least children and answers of a
// cell shown, and the division and the multiplication a count is rounded by.
var (
	viewLeast   = regexp.MustCompile(`learners >= (\d+) AND counted\.answers >= (\d+)`)
	viewRounded = regexp.MustCompile(`ROUND\([\w.]+ / (\d+)\) \* (\d+)`)
)

// wantTheLiveRule holds the rule a public view's SQL shows its cells by to the
// rule the live numbers are held to.
func wantTheLiveRule(t *testing.T, view string) {
	t.Helper()
	sql := string(readBytes(t, filepath.Join("..", "..", "infra", "analytics", "views", "public", view)))
	leasts, roundings := viewLeast.FindAllStringSubmatch(sql, -1), viewRounded.FindAllStringSubmatch(sql, -1)
	if len(leasts) == 0 || len(roundings) == 0 {
		t.Errorf("%s shows its cells by no rule this test can read", view)
	}
	for _, found := range leasts {
		if found[1] != strconv.Itoa(privacyRule.Learners) || found[2] != strconv.Itoa(privacyRule.Answers) {
			t.Errorf("%s shows a cell of %s children and %s answers or more, and the live numbers are held to %d and %d",
				view, found[1], found[2], privacyRule.Learners, privacyRule.Answers)
		}
	}
	for _, found := range roundings {
		if found[1] != strconv.Itoa(privacyRule.RoundedTo) || found[2] != found[1] {
			t.Errorf("%s rounds a count by %s and %s, and the live numbers are held to %d", view, found[1], found[2], privacyRule.RoundedTo)
		}
	}
}
