//go:build analytics

package analytics_test

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// The months the public views are tried on: January and February 2026 counted
// every day, March only its first ten.
var (
	january  = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	february = time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	march    = time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
)

// countedOnce makes the space the view tests read, once for all of them: they
// only read it, and every space made costs the emulator memory it keeps.
var (
	countedOnce  sync.Once
	countedSpace *space
)

// counted is a space whose tables hold the counts of those months, as nights
// would have left them, and whose views are made: the first test to ask makes
// it, and the rest read the same.
func counted(t *testing.T) *space {
	t.Helper()

	countedOnce.Do(func() { countedSpace = countMonths(t) })
	if countedSpace == nil {
		t.Fatal("the counted months were not made: the first test to make them says why")
	}
	return countedSpace
}

// countMonths fills a space with the counts of those months and makes its
// views.
func countMonths(t *testing.T) *space {
	t.Helper()

	s := namedSpace(t)
	s.make(t)
	var days []string
	// Five children a day, six tasks and seven answers; a topic won on the
	// first two days of January alone, too few to show.
	for d := january; d.Before(march.AddDate(0, 0, 10)); d = d.AddDate(0, 0, 1) {
		won := 0
		if d.Before(january.AddDate(0, 0, 2)) {
			won = 1
		}
		days = append(days, fmt.Sprintf("(DATE '%s', 5, NULL, NULL, 6, 7, %d)", day(d), won))
	}
	s.rows(t, "INSERT INTO `${impact}.daily` (day, learners, learners_week, learners_month, tasks, answers, topics_won) VALUES "+strings.Join(days, ", "))

	learners := []string{
		// January: 50 children. By host, three groups of fewer than ten; by
		// cohort, the month's own of six beside the rest.
		"('2026-01-01', 'all', 'all', 50)",
		"('2026-01-01', 'host', 'claude', 30)", "('2026-01-01', 'host', 'chatgpt', 7)",
		"('2026-01-01', 'host', 'other', 4)", "('2026-01-01', 'host', 'unknown', 9)",
		"('2026-01-01', 'cohort', '2026-01', 6)", "('2026-01-01', 'cohort', '2025-12', 44)",
		"('2026-01-01', 'country', 'US', 38)", "('2026-01-01', 'country', 'GB', 12)",
		// February: eight children in all.
		"('2026-02-01', 'all', 'all', 8)", "('2026-02-01', 'host', 'claude', 8)",
		// March, under way: forty.
		"('2026-03-01', 'all', 'all', 40)", "('2026-03-01', 'host', 'claude', 40)",
	}
	s.rows(t, "INSERT INTO `${impact}.learners_monthly` (month, dimension, value, learners) SELECT DATE(month), dimension, value, learners FROM UNNEST([STRUCT('' AS month, '' AS dimension, '' AS value, 0 AS learners), "+
		strings.Join(learners, ", ")+"]) WHERE month != ''")
	s.rows(t, "INSERT INTO `${impact}.learning_monthly` (month, tenure, learners, answers, correct, hinted, dont_know, mastered) VALUES "+
		"(DATE '2026-01-01', 0, 6, 30, 20, 3, 2, 6), (DATE '2026-01-01', 1, 20, 100, 70, 10, 5, 40), (DATE '2026-01-01', 2, 24, 120, 90, 12, 6, 72), "+
		"(DATE '2026-02-01', 0, 8, 40, 30, 4, 2, 8)")
	s.rows(t, "INSERT INTO `${impact}.topics_monthly` (month, topic, learners, answers, correct, dont_know, won, winners) VALUES "+
		"(DATE '2026-01-01', 'logic.ordering', 15, 40, 30, 2, 12, 11), "+
		"(DATE '2026-01-01', 'counting.gaps', 9, 20, 10, 1, 5, 5), "+
		"(DATE '2026-01-01', 'time.clocks', 12, 30, 21, 0, 4, 3)")
	s.rows(t, "INSERT INTO `${impact}.traps_monthly` (month, topic, grade, trap, answers, learners) VALUES "+
		"(DATE '2026-01-01', 'logic.ordering', NULL, 'reversed_order', 14, 10), "+
		"(DATE '2026-01-01', 'logic.ordering', NULL, 'off_by_one', 6, 4), "+
		"(DATE '2026-01-01', NULL, 3, 'reversed_order', 16, 11), "+
		"(DATE '2026-01-01', 'logic.ordering', 3, 'reversed_order', 13, 10)")
	// The answers weighed against the chance promised. January by range of
	// chance: ten children and thirty answers, shown; nine children, and
	// twenty-nine answers, not shown; twelve children and forty-one answers,
	// shown as ten and forty. January over every range has forty-five
	// children, February eight, too few, and March is under way.
	s.rows(t, "INSERT INTO `${impact}.chance_monthly` (month, bucket, learners, answers, correct, promised) VALUES "+
		"(DATE '2026-01-01', '0.70-0.77', 10, 30, 22, 2220), "+
		"(DATE '2026-01-01', '0.78-0.85', 9, 200, 160, 16200), "+
		"(DATE '2026-01-01', 'over 0.85', 29, 29, 27, 2610), "+
		"(DATE '2026-01-01', '0.60-0.69', 12, 41, 28, 2665), "+
		"(DATE '2026-01-01', NULL, 45, 400, 300, 30000), "+
		"(DATE '2026-02-01', NULL, 8, 40, 30, 3000), "+
		"(DATE '2026-03-01', '0.70-0.77', 20, 60, 45, 4440), "+
		"(DATE '2026-03-01', NULL, 20, 60, 45, 4440)")
	// January by the range of the children's answers: in 6-20 ten children of
	// three answers each, five above their promise by 30 hundredths and five
	// below; in 21-50 ten children each 70 hundredths below on three answers,
	// which spreads them not at all; in 101-200 nine children. March is under
	// way.
	s.rows(t, "INSERT INTO `${impact}.kept_up_monthly` (month, bucket, learners, answers, correct, promised, margins_squared, answers_by_margins, answers_squared) VALUES "+
		"(DATE '2026-01-01', '6-20', 10, 30, 24, 2400, 9000, 0, 90), "+
		"(DATE '2026-01-01', '21-50', 10, 30, 20, 2700, 49000, -2100, 90), "+
		"(DATE '2026-01-01', '101-200', 9, 300, 200, 21000, 1000, 10, 10000), "+
		"(DATE '2026-03-01', '6-20', 20, 60, 45, 4440, 100, 0, 180)")
	s.makeViews(t)
	return s
}

// A public view shows only the months every day of which was counted: not a
// month under way, the month the counting began in, or one with a night that
// never ran.
func TestAPublicViewShowsOnlyMonthsCountedWhole(t *testing.T) {
	s := counted(t)

	got := s.rows(t, "SELECT month FROM `${public}.closed_months` ORDER BY month")
	if want := [][]string{{day(january)}, {day(february)}}; !equalRows(got, want) {
		t.Errorf("the months shown = %v, want %v", got, want)
	}
	if got := s.rows(t, "SELECT month FROM `${public}.learners_by_host` WHERE month = DATE '2026-03-01'"); len(got) != 0 {
		t.Errorf("March, under way, is shown: %v", got)
	}
}

// Children told apart by one thing show no group of fewer than ten and leave
// none out to be worked out from the rest: the groups of fewer than ten go into
// (others), and while (others) holds fewer than ten the smallest of the rest
// goes in too. A month of fewer than ten children in all shows nothing, and
// every number is to the nearest five.
func TestAPublicViewShowsNoGroupOfFewerThanTen(t *testing.T) {
	s := counted(t)

	for _, check := range []struct {
		view string
		want [][]string
	}{
		// 4, 7 and 9 are small; 4 + 7 is 11, enough.
		{"learners_by_host", [][]string{{day(january), "(others)", "20"}, {day(january), "claude", "30"}}},
		// 6 is small, and 6 alone is too few: the next, 44, goes in with it.
		{"learners_by_cohort", [][]string{{day(january), "(others)", "50"}}},
		// No group is small: 12 is shown as 10, 38 as 40.
		{"learners_by_country", [][]string{{day(january), "GB", "10"}, {day(january), "US", "40"}}},
	} {
		got := s.rows(t, "SELECT month, value, learners FROM `${public}."+check.view+"` ORDER BY month, value")
		if !equalRows(got, check.want) {
			t.Errorf("%s = %v, want %v", check.view, got, check.want)
		}
	}
}

// The months view shows a month of ten children or more, its numbers to the
// nearest five; the children who began that month are those the cohort view
// shows for the month's own cohort, and none where it folded them in with
// others — the two views set against each other must not give away a group of
// fewer than ten.
func TestTheMonthsViewGivesAwayNoSmallGroup(t *testing.T) {
	s := counted(t)

	got := s.rows(t, "SELECT month, learners, IFNULL(CAST(new_learners AS STRING), 'NULL'), tasks, answers, IFNULL(CAST(topics_won AS STRING), 'NULL'), tasks_per_child_week, active_days_per_child FROM `${public}.months` ORDER BY month")
	// January: 31 days of 5 children, 6 tasks and 7 answers, and two topics
	// won, too few to show.
	want := [][]string{{day(january), "50", "NULL", "185", "215", "NULL", "0.8", "3.1"}}
	if !equalRows(got, want) {
		t.Errorf("months = %v, want %v", got, want)
	}
}

// How the children answered, by the months since their profile was made:
// a child is in one group, so the small groups fold, and every share and mean
// stands on ten children or more. A month of fewer than ten shows nothing.
func TestTheLearningViewFoldsSmallGroups(t *testing.T) {
	s := counted(t)

	got := s.rows(t, "SELECT month, tenure, learners, answers, correct_share, hint_share, dont_know_share, topics_mastered_mean FROM `${public}.learning` ORDER BY month, tenure")
	// In January tenure 0 has 6 children, too few alone: tenure 1's 20 go in
	// with them. February's 8 children are too few to show at all.
	want := [][]string{
		{day(january), "(others)", "25", "130", "0.692", "0.1", "0.054", "1.8"},
		{day(january), "2", "25", "120", "0.75", "0.1", "0.05", "3"},
	}
	if !equalRows(got, want) {
		t.Errorf("learning = %v, want %v", got, want)
	}
}

// A topic or a trap is shown where ten children or more stand behind it, and
// the times a topic was won where ten children or more won it. A child is in
// many topics and traps, so none is folded in with others; the rows by topic
// and grade together are never shown.
func TestTopicsAndTrapsStandOnTenChildren(t *testing.T) {
	s := counted(t)

	for _, check := range []struct {
		sql  string
		want [][]string
	}{
		{"SELECT topic, learners, answers, correct_share, IFNULL(CAST(won AS STRING), 'NULL') FROM `${public}.topics` ORDER BY topic",
			[][]string{{"logic.ordering", "15", "40", "0.75", "10"}, {"time.clocks", "10", "30", "0.7", "NULL"}}},
		{"SELECT topic, trap, learners, answers FROM `${public}.traps_by_topic` ORDER BY topic, trap",
			[][]string{{"logic.ordering", "reversed_order", "10", "15"}}},
		{"SELECT CAST(grade AS STRING), trap, learners, answers FROM `${public}.traps_by_grade` ORDER BY grade, trap",
			[][]string{{"3", "reversed_order", "10", "15"}}},
	} {
		if got := s.rows(t, check.sql); !equalRows(got, check.want) {
			t.Errorf("%s = %v, want %v", check.sql, got, check.want)
		}
	}
}

// The numbers just impact prints for an application: by default the public
// views' whole months, the latest first, with the children in the United
// States beside them; given private, every month counted, exactly; and the
// days no night counted, from the first counted to yesterday.
func TestTheNumbersForAnApplicationReadTheViews(t *testing.T) {
	s := counted(t)

	public := s.e.query(t, s.render(t, "impact/public.sql", map[string]string{"months": "3"}))
	if want := [][]string{{"2026-01", "50", "NULL", "40", "NULL", "185", "215", "0.8", "3.1", "NULL"}}; !equalRows(public, want) {
		t.Errorf("the public numbers = %v, want %v", public, want)
	}
	private := s.e.query(t, s.render(t, "impact/private.sql", map[string]string{"months": "2"}))
	want := [][]string{
		{"2026-03", "40", "0", "NULL", "NULL", "60", "70", "1.1", "1.3", "0"},
		{"2026-02", "8", "0", "NULL", "NULL", "168", "196", "5.3", "17.5", "0"},
	}
	if !equalRows(private, want) {
		t.Errorf("the private numbers = %v, want %v", private, want)
	}
	missing := s.e.query(t, s.render(t, "impact/missing.sql", nil))
	if len(missing) == 0 || missing[0][0] != "2026-03-11" {
		t.Errorf("the days no night counted begin with %v, want 2026-03-11, the day after the last counted", missing)
	}
}

// The answers weighed against the chance promised show a range of chance, or
// a month over every range, where ten children and thirty answers or more
// stand behind it, its counts to the nearest five and its shares to three
// places. A range of fewer is not shown, and neither is a month not counted
// whole: a child's answers fall in many ranges, so none is folded in with
// others.
func TestAWeighedRangeStandsOnTenChildrenAndThirtyAnswers(t *testing.T) {
	s := counted(t)

	for _, check := range []struct {
		sql  string
		want [][]string
	}{
		{"SELECT month, chance, learners, answers, promised_mean, correct_share FROM `${public}.chances` ORDER BY month, chance",
			[][]string{
				{day(january), "0.60-0.69", "10", "40", "0.65", "0.683"},
				{day(january), "0.70-0.77", "10", "30", "0.74", "0.733"},
			}},
		{"SELECT month, learners, answers, promised_mean, correct_share FROM `${public}.chances_total` ORDER BY month",
			[][]string{{day(january), "45", "400", "0.75", "0.75"}}},
	} {
		if got := s.rows(t, check.sql); !equalRows(got, check.want) {
			t.Errorf("%s = %v, want %v", check.sql, got, check.want)
		}
	}
}

// How far the answers came out from the chance promised, by the range of the
// child's answers: right less the chance, on average, to three places, with
// its standard error counted by child from the sums the night keeps. Ten
// children five each way of their promise give sqrt(10/9 * 9000) / 3000, or
// 0.033; ten children of one margin give none, the sum under the root held at
// nothing where rounding takes it a hair below; nine children are not shown.
func TestTheErrorOfTheMeanIsCountedByChild(t *testing.T) {
	s := counted(t)

	got := s.rows(t, "SELECT month, answers_range, learners, answers, came_true_less_promised, standard_error FROM `${public}.kept_up` ORDER BY month, answers_range")
	want := [][]string{
		{day(january), "21-50", "10", "30", "-0.233", "0"},
		{day(january), "6-20", "10", "30", "0", "0.033"},
	}
	if !equalRows(got, want) {
		t.Errorf("kept_up = %v, want %v", got, want)
	}
	exact := s.rows(t, "SELECT month, answers_range, learners, answers, ROUND(came_true_less_promised, 3), ROUND(standard_error, 3) FROM `${private}.kept_up` WHERE learners >= 10 AND month < DATE '2026-03-01' ORDER BY month, answers_range")
	if !equalRows(exact, want) {
		t.Errorf("the exact kept_up, rounded = %v, want the public numbers %v", exact, want)
	}
}

// Every count a public view shows — of children, of tasks, of answers, of
// topics won — is ten or more and a multiple of five.
func TestEveryPublicCountIsTenOrMoreAndAMultipleOfFive(t *testing.T) {
	s := counted(t)

	views := publicViews(t)
	for _, dimension := range dimensions {
		views = append(views, "learners_by_"+dimension)
	}
	counts := map[string]bool{"learners": true, "new_learners": true, "tasks": true, "answers": true, "topics_won": true, "won": true}
	for _, view := range views {
		for _, column := range s.e.columns(t, fill(t, "SELECT * FROM `${public}."+view+"` LIMIT 1", s.names)) {
			if counts[column] {
				wantShownCounts(t, s, view, column)
			}
		}
	}
}

// wantShownCounts holds every number a column of a public view shows to ten
// or more and a multiple of five.
func wantShownCounts(t *testing.T, s *space, view, column string) {
	t.Helper()

	for _, row := range s.rows(t, fmt.Sprintf("SELECT CAST(%s AS STRING) FROM `${public}.%s` WHERE %s IS NOT NULL", column, view, column)) {
		if n, err := strconv.Atoi(row[0]); err != nil || n%5 != 0 || n < 10 {
			t.Errorf("%s.%s shows %s, want ten or more, and a multiple of five", view, column, row[0])
		}
	}
}
