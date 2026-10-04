//go:build analytics

package analytics_test

import (
	"fmt"
	"slices"
	"testing"
	"time"
)

// openedLongAgo writes the line the bucket's first day of lines falls on, 40
// days before yesterday, so that the week and the month of yesterday are
// counted whole whatever day the test runs on. The child on it is not counted:
// the first day of lines may have begun without them. It is the first day the
// night counts from that it says.
func openedLongAgo(t *testing.T, s *space) time.Time {
	t.Helper()

	first := today().AddDate(0, 0, -41)
	s.write(t, accepted(learnerOf(999), noon(first), nil))
	return first.AddDate(0, 0, 1)
}

// A night counts every day from the first the lines hold whole to yesterday,
// a day with nothing on it among them: the children a task was handed to that
// day, over the seven days to it and over its month so far — none, where the
// lines do not reach back that far — and the tasks, the answers and the topics
// won. A line the log repeated counts once, a line that names no child counts
// nothing, and today is not counted until it is over.
func TestTheNightCountsEveryDayAndTheWeekAndTheMonthToIt(t *testing.T) {
	s := newSpace(t)
	countedFrom := openedLongAgo(t, s)
	yesterday := today().AddDate(0, 0, -1)
	twoDaysAgo := today().AddDate(0, 0, -2)

	repeated := answered(learnerOf(1), noon(yesterday), nil)
	noName := accepted(learnerOf(9), noon(yesterday), nil)
	delete(noName.fields, "learner")
	s.write(t,
		accepted(learnerOf(1), noon(twoDaysAgo), nil),
		accepted(learnerOf(1), noon(yesterday), nil),
		accepted(learnerOf(2), yesterday.Add(24*time.Hour-time.Second), nil),
		accepted(learnerOf(3), today().Add(time.Second), nil),
		repeated, repeated, noName,
		won(learnerOf(1), noon(yesterday), nil),
	)
	s.countTheNight(t)

	got := s.rows(t, "SELECT day, learners, IFNULL(CAST(learners_week AS STRING), 'NULL'), IFNULL(CAST(learners_month AS STRING), 'NULL'), tasks, answers, topics_won FROM `${impact}.daily` ORDER BY day")
	if want := int(yesterday.Sub(countedFrom).Hours()/24) + 1; len(got) != want {
		t.Fatalf("the night counted %d days, want %d: every day from %s to %s", len(got), want, day(countedFrom), day(yesterday))
	}
	if first := got[0]; first[0] != day(countedFrom) || first[2] != "NULL" {
		t.Errorf("the first day counted = %v, want %s with no week: the lines do not reach back to its first day", first, day(countedFrom))
	}
	if first := got[0]; countedFrom.Day() != 1 && first[3] != "NULL" {
		t.Errorf("the first day counted = %v, want no month: the lines do not reach back to its first day", first)
	}
	for _, want := range [][]string{
		{day(twoDaysAgo), "1", "1"},
		{day(yesterday), "2", "2", "2", "2", "1", "1"},
	} {
		row := rowOf(got, want[0])
		if row == nil || !slices.Equal(row[1:len(want)], want[1:]) {
			t.Errorf("the day %s = %v, want %v", want[0], row, want)
		}
	}
	if row := rowOf(got, day(today())); row != nil {
		t.Errorf("today was counted, %v, before it is over", row)
	}
	months := s.rows(t, "SELECT month, learners FROM `${impact}.learners_monthly` WHERE dimension = 'all' ORDER BY month")
	if row := rowOf(months, day(firstOfMonth(today()))); firstOfMonth(today()).Equal(firstOfMonth(yesterday)) && (len(row) < 2 || row[1] != "2") {
		t.Errorf("the month so far = %v, want the two children of the days before today", row)
	} else if !firstOfMonth(today()).Equal(firstOfMonth(yesterday)) && row != nil {
		t.Errorf("the month begun today = %v, want none until today is over", row)
	}
}

// The children the load tool or MCP Inspector handed a task to are no
// children a family has: nothing their names are on is counted, an answer
// written before the tool's task and a topic won included.
func TestTheChildrenOfATestingToolAreLeftOut(t *testing.T) {
	s := newSpace(t)
	openedLongAgo(t, s)
	yesterday := today().AddDate(0, 0, -1)
	threeDaysAgo := today().AddDate(0, 0, -3)

	s.write(t,
		answered(learnerOf(1), noon(threeDaysAgo), nil),
		accepted(learnerOf(1), noon(yesterday), map[string]any{"host": "load"}),
		won(learnerOf(1), noon(yesterday), nil),
		accepted(learnerOf(2), noon(yesterday), map[string]any{"host": "inspector"}),
		accepted(learnerOf(3), noon(yesterday), map[string]any{"host": "chatgpt"}),
	)
	s.countTheNight(t)

	days := s.rows(t, "SELECT day, learners, tasks, answers, topics_won FROM `${impact}.daily` ORDER BY day")
	if row := rowOf(days, day(threeDaysAgo)); len(row) < 4 || row[3] != "0" {
		t.Errorf("the day %s = %v, want no answer: its child was the load tool's", day(threeDaysAgo), row)
	}
	if row := rowOf(days, day(yesterday)); row == nil || !slices.Equal(row[1:], []string{"1", "1", "0", "0"}) {
		t.Errorf("yesterday = %v, want the one child of the chat, and no topic won", row)
	}
	hosts := s.rows(t, "SELECT value, learners FROM `${impact}.learners_monthly` WHERE dimension = 'host' ORDER BY value")
	if want := [][]string{{"chatgpt", "1"}}; !equalRows(hosts, want) {
		t.Errorf("the children of the month by host = %v, want %v", hosts, want)
	}
}

// A child is counted once in a week and in a month, under the value of its
// latest task there: one that moved from Claude to ChatGPT and from grade 2 to
// grade 3 is a ChatGPT child of grade 3, so that the values of a dimension add
// up to all and no child is counted under two of them.
func TestAChildIsCountedOnceUnderItsLatestValue(t *testing.T) {
	s := newSpace(t)
	openedLongAgo(t, s)
	yesterday := today().AddDate(0, 0, -1)

	s.write(t,
		accepted(learnerOf(1), yesterday.Add(8*time.Hour), map[string]any{"host": "claude", "grade": 2}),
		accepted(learnerOf(1), yesterday.Add(20*time.Hour), map[string]any{"host": "chatgpt", "grade": 3}),
		accepted(learnerOf(2), yesterday.Add(10*time.Hour), map[string]any{"host": "claude", "grade": 2}),
	)
	s.countTheNight(t)

	for _, table := range []string{"learners_weekly", "learners_monthly"} {
		got := s.rows(t, "SELECT dimension, value, learners FROM `${impact}."+table+"` WHERE dimension IN ('all', 'host', 'grade') ORDER BY dimension, value")
		want := [][]string{{"all", "all", "2"}, {"grade", "2", "1"}, {"grade", "3", "1"}, {"host", "chatgpt", "1"}, {"host", "claude", "1"}}
		if !equalRows(got, want) {
			t.Errorf("%s = %v, want %v", table, got, want)
		}
	}
}

// Only a week or a month the lines hold from its first day is counted: one
// that began before the first day of lines would be counted from part of its
// days, and is left out.
func TestOnlyAWholeWeekOrMonthIsCounted(t *testing.T) {
	s := newSpace(t)
	first := today().AddDate(0, 0, -12)
	countedFrom := first.AddDate(0, 0, 1)
	s.write(t,
		accepted(learnerOf(999), noon(first), nil),
		accepted(learnerOf(1), noon(countedFrom), nil),
		accepted(learnerOf(2), noon(today().AddDate(0, 0, -1)), nil),
	)
	s.countTheNight(t)

	weeksFrom := countedFrom
	for weeksFrom.Weekday() != time.Monday {
		weeksFrom = weeksFrom.AddDate(0, 0, 1)
	}
	monthsFrom := firstOfMonth(countedFrom)
	if monthsFrom.Before(countedFrom) {
		monthsFrom = monthsFrom.AddDate(0, 1, 0)
	}
	if got := s.rows(t, fmt.Sprintf("SELECT week FROM `${impact}.learners_weekly` WHERE week < DATE '%s'", day(weeksFrom))); len(got) != 0 {
		t.Errorf("weeks counted from part of their days: %v, want none before %s", got, day(weeksFrom))
	}
	if got := s.rows(t, fmt.Sprintf("SELECT month FROM `${impact}.learners_monthly` WHERE month < DATE '%s'", day(monthsFrom))); len(got) != 0 {
		t.Errorf("months counted from part of their days: %v, want none before %s", got, day(monthsFrom))
	}
	if got := s.rows(t, fmt.Sprintf("SELECT day FROM `${impact}.daily` WHERE day < DATE '%s'", day(countedFrom))); len(got) != 0 {
		t.Errorf("days counted before the first whole one: %v", got)
	}
}

// A day counted whole keeps its numbers once its lines begin to go: the nights
// after count it no longer, rather than count its week and its month from part
// of their days, or count a child the Inspector handed a task to earlier that
// month, whose task has gone from the lines.
func TestADayCountedWholeKeepsItsNumbersAsTheLinesGo(t *testing.T) {
	s := newSpace(t)
	s.write(t, accepted(learnerOf(999), noon(today().AddDate(0, 0, -58)), nil))
	// The tenth of the month twenty days back: its week, its month and the
	// month of its week are all in the lines.
	tenth := firstOfMonth(today().AddDate(0, 0, -20)).AddDate(0, 0, 9)
	seventeenth := tenth.AddDate(0, 0, 7)
	s.write(t,
		accepted(learnerOf(1), noon(tenth.AddDate(0, 0, -3)), nil),
		accepted(learnerOf(1), noon(tenth), nil),
		accepted(learnerOf(1), noon(seventeenth), nil),
		accepted(learnerOf(2), noon(tenth.AddDate(0, 0, -8)), map[string]any{"host": "inspector"}),
		accepted(learnerOf(2), noon(tenth), nil),
		accepted(learnerOf(2), noon(seventeenth), nil),
	)
	s.countTheNight(t)
	days := fmt.Sprintf("SELECT day, learners, IFNULL(CAST(learners_week AS STRING), 'NULL'), IFNULL(CAST(learners_month AS STRING), 'NULL'), tasks FROM `${impact}.daily` WHERE day = DATE '%s'", day(tenth))
	weeks := fmt.Sprintf("SELECT week, learners FROM `${impact}.learners_weekly` WHERE dimension = 'all' AND week = DATE_TRUNC(DATE '%s', WEEK(MONDAY))", day(seventeenth))
	countedDay, countedWeek := s.rows(t, days), s.rows(t, weeks)
	if want := [][]string{{day(tenth), "1", "1", "1", "1"}}; !equalRows(countedDay, want) {
		t.Fatalf("the tenth = %v, want %v: one child, the Inspector's left out", countedDay, want)
	}
	if len(countedWeek) != 1 || countedWeek[0][1] != "1" {
		t.Fatalf("the week of the seventeenth = %v, want one child, the Inspector's left out", countedWeek)
	}

	// The lines before the seventh go, as the bucket lets them go: the first
	// day the bucket holds whole moves into the tenth's week, and the
	// Inspector's task of the second is gone.
	s.rows(t, fmt.Sprintf("DELETE FROM `${logs}` WHERE timestamp < TIMESTAMP '%s 00:00:00+00'", day(tenth.AddDate(0, 0, -3))))
	s.countTheNight(t)
	if again := s.rows(t, days); !equalRows(again, countedDay) {
		t.Errorf("the tenth, once its lines began to go = %v, want it kept as %v", again, countedDay)
	}
	if again := s.rows(t, weeks); !equalRows(again, countedWeek) {
		t.Errorf("the week of the seventeenth, once its month's lines began to go = %v, want it kept as %v", again, countedWeek)
	}
}

// A night counted again gives the same rows, so a run made by hand changes
// nothing a night would not; and the rows of a period the lines no longer hold
// whole are left as they were last counted.
func TestANightCountedAgainGivesTheSameRows(t *testing.T) {
	s := newSpace(t)
	openedLongAgo(t, s)
	yesterday := today().AddDate(0, 0, -1)
	longAgo := today().AddDate(0, 0, -100)
	s.rows(t, fmt.Sprintf("INSERT INTO `${impact}.daily` (day, learners, learners_week, learners_month, tasks, answers, topics_won) VALUES (DATE '%s', 7, 7, 7, 9, 9, 1)", day(longAgo)))
	s.write(t,
		accepted(learnerOf(1), noon(yesterday), map[string]any{"host": "chatgpt"}),
		answered(learnerOf(1), noon(yesterday), map[string]any{"correct": false, "trap": "t1"}),
		won(learnerOf(1), noon(yesterday), nil),
	)

	s.countTheNight(t)
	first := everyRow(t, s)
	s.countTheNight(t)
	if again := everyRow(t, s); !slices.EqualFunc(again, first, slices.Equal) {
		t.Errorf("the second night counted\n%v\nwant the first night's\n%v", again, first)
	}
	if row := rowOf(first, "daily "+day(longAgo)); row == nil {
		t.Errorf("the day %s, which the lines no longer hold, lost its row", day(longAgo))
	}
}

// How much the children of a month did and how well: the tasks and the active
// days of each, in ranges; the answers by the months since a profile was made,
// right, after the hint and I don't know, and the most topics each had
// mastered; each topic; and the traps by topic and grade, by topic over every
// grade and by grade over every topic.
func TestHowMuchAndHowWellTheChildrenDidIsCounted(t *testing.T) {
	s := newSpace(t)
	openedLongAgo(t, s)
	yesterday := today().AddDate(0, 0, -1)
	month := firstOfMonth(yesterday)
	cohort := month.Format("2006-01")
	olderCohort := month.AddDate(0, -2, 0).Format("2006-01")

	at := func(hour int) time.Time { return yesterday.Add(time.Duration(hour) * time.Hour) }
	// The third child's lines name no grade on one answer and no month its
	// profile was made in on the other: the first counts over every grade
	// alone, and the second takes the month the other names.
	noGrade := answered(learnerOf(3), at(12), map[string]any{"cohort": cohort, "correct": false, "trap": "t1"})
	delete(noGrade.fields, "grade")
	noCohort := answered(learnerOf(3), at(13), nil)
	delete(noCohort.fields, "cohort")
	s.write(t,
		accepted(learnerOf(1), at(8), nil),
		accepted(learnerOf(1), at(9), nil),
		accepted(learnerOf(1), at(10), nil),
		answered(learnerOf(1), at(8), map[string]any{"cohort": cohort, "hint_used": true, "topics_mastered": 1}),
		answered(learnerOf(1), at(9), map[string]any{"cohort": cohort, "correct": false, "trap": "t1", "topics_mastered": 2}),
		answered(learnerOf(1), at(10), map[string]any{"cohort": cohort, "correct": false, "confused": true, "topics_mastered": 2}),
		accepted(learnerOf(2), at(11), map[string]any{"grade": 4}),
		answered(learnerOf(2), at(11), map[string]any{"cohort": olderCohort, "grade": 4, "correct": false, "trap": "t1"}),
		won(learnerOf(2), at(11), map[string]any{"grade": 4}),
		noGrade, noCohort,
	)
	s.countTheNight(t)

	monthly := func(table, columns string) [][]string {
		return s.rows(t, fmt.Sprintf("SELECT %s FROM `${impact}.%s` WHERE month = DATE '%s' ORDER BY 1, 2, 3", columns, table, day(month)))
	}
	for _, check := range []struct {
		table, columns string
		want           [][]string
	}{
		{"dose_monthly", "measure, bucket, learners", [][]string{{"active_days", "1", "2"}, {"tasks", "1", "1"}, {"tasks", "2-4", "1"}}},
		{"learning_monthly", "tenure, learners, answers, correct, hinted, dont_know, mastered", [][]string{
			{"0", "2", "5", "2", "1", "1", "2"},
			{"2", "1", "1", "0", "0", "0", "0"},
		}},
		{"topics_monthly", "topic, learners, answers, correct, dont_know, won, winners", [][]string{
			{"logic.ordering", "3", "6", "2", "1", "1", "1"},
		}},
		{"traps_monthly", "IFNULL(topic, 'every topic'), IFNULL(CAST(grade AS STRING), 'every grade'), trap, answers, learners", [][]string{
			{"every topic", "2", "t1", "1", "1"},
			{"every topic", "4", "t1", "1", "1"},
			{"logic.ordering", "2", "t1", "1", "1"},
			{"logic.ordering", "4", "t1", "1", "1"},
			{"logic.ordering", "every grade", "t1", "3", "3"},
		}},
	} {
		if got := monthly(check.table, check.columns); !equalRows(got, check.want) {
			t.Errorf("%s = %v, want %v", check.table, got, check.want)
		}
	}
}

// rowOf is the row whose first cell is the key, or nil.
func rowOf(rows [][]string, key string) []string {
	for _, row := range rows {
		if len(row) > 0 && row[0] == key {
			return row
		}
	}
	return nil
}

// equalRows reports whether two lists of rows are the same, row by row.
func equalRows(got, want [][]string) bool {
	return slices.EqualFunc(got, want, slices.Equal)
}

// everyRow is every row of every table of the counts, each led by its table's
// name and the first column, so that two nights can be laid side by side.
// Every table has four columns or more, and its first four tell its rows
// apart.
func everyRow(t *testing.T, s *space) [][]string {
	t.Helper()

	var all [][]string
	for _, table := range tableNames(t) {
		for _, row := range s.rows(t, "SELECT * FROM `${impact}."+table+"` ORDER BY 1, 2, 3, 4") {
			all = append(all, append([]string{table + " " + row[0]}, row[1:]...))
		}
	}
	return all
}
