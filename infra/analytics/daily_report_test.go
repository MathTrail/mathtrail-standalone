//go:build analytics

package analytics_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The tables the export of the costs writes, named as it names them after a
// billing account: the standard export the report reads, and the detailed one
// beside it, which repeats the same costs by resource and is never read.
const (
	standardExport = "gcp_billing_export_v1_000000_000000_000000"
	detailedExport = "gcp_billing_export_resource_v1_000000_000000_000000"
)

// costLine is one line of the export: when the usage it charges began, the
// project it was charged to, the charge and the credits against it.
type costLine struct {
	at      time.Time
	project string
	cost    float64
	credits []float64
}

// export makes a table of the export's shape, with the columns the report
// reads, and puts the lines in it.
func (s *space) export(t *testing.T, table string, lines ...costLine) {
	t.Helper()

	name := s.names["billing"] + "." + table
	s.e.query(t, "CREATE TABLE `"+name+"` (usage_start_time TIMESTAMP, project STRUCT<id STRING>, cost FLOAT64, credits ARRAY<STRUCT<amount FLOAT64>>)")
	if len(lines) == 0 {
		return
	}
	values := make([]string, len(lines))
	for i, l := range lines {
		credits := make([]string, len(l.credits))
		for j, amount := range l.credits {
			credits[j] = fmt.Sprintf("STRUCT(%v AS amount)", amount)
		}
		values[i] = fmt.Sprintf("(TIMESTAMP '%s', STRUCT('%s' AS id), %v, CAST([%s] AS ARRAY<STRUCT<amount FLOAT64>>))",
			l.at.UTC().Format("2006-01-02 15:04:05+00"), l.project, l.cost, strings.Join(credits, ", "))
	}
	s.e.query(t, "INSERT INTO `"+name+"` (usage_start_time, project, cost, credits) VALUES "+strings.Join(values, ", "))
}

// count puts the night's counts of days in, the children of each as given and
// the rest of its counts beside them.
func (s *space) count(t *testing.T, learners map[time.Time]int) {
	t.Helper()

	var days []string
	for d, n := range learners {
		days = append(days, fmt.Sprintf("(DATE '%s', %d, %d, %d, %d, %d, 1)", day(d), n, n+1, n+2, 2*n, 3*n))
	}
	s.rows(t, "INSERT INTO `${impact}.daily` (day, learners, learners_week, learners_month, tasks, answers, topics_won) VALUES "+strings.Join(days, ", "))
}

// makeDailyReport makes the daily report's view, as Terraform makes it once
// the export has its table.
func (s *space) makeDailyReport(t *testing.T) {
	t.Helper()
	s.e.query(t, "CREATE VIEW `"+s.names["private"]+".daily_report` AS "+s.render(t, "billing/daily_report.sql", nil))
}

// The report has a row for each of the 31 days to yesterday and for no other,
// whether or not the night or the export holds anything of it, and its flags
// tell the page which row is yesterday, which are of yesterday's month and how
// far back each is.
func TestTheDailyReportHasEachOfTheLast31DaysOnce(t *testing.T) {
	t.Parallel()

	s := newSpace(t)
	s.export(t, standardExport)
	yesterday := today().AddDate(0, 0, -1)
	s.count(t, map[time.Time]int{yesterday: 3, today().AddDate(0, 0, -31): 5, today().AddDate(0, 0, -32): 7})
	s.makeDailyReport(t)

	got := s.rows(t, "SELECT day, days_ago, yesterday, this_month, learners FROM `${private}.daily_report` ORDER BY day")
	if len(got) != 31 {
		t.Fatalf("got %d rows, want 31: %v", len(got), got)
	}
	for i, row := range got {
		back := 31 - i
		d := today().AddDate(0, 0, -back)
		learners := "NULL"
		switch back {
		case 1:
			learners = "3"
		case 31:
			learners = "5"
		}
		want := []string{day(d), strconv.Itoa(back), strconv.FormatBool(back == 1), strconv.FormatBool(firstOfMonth(d).Equal(firstOfMonth(yesterday))), learners}
		if strings.Join(row, " ") != strings.Join(want, " ") {
			t.Errorf("row %d: got %v, want %v", i, row, want)
		}
	}
}

// A day costs what the standard export charged this project for the usage
// that began on it, in UTC, less the credits against those charges, from the
// first day of the report to the last: another project's charges and the
// detailed export's lines are not counted, a day with no credits still has its
// charge, and the night's counts stand once beside the day's sum whatever the
// number of its lines.
func TestADayCostsWhatTheExportChargedTheProjectLessItsCredits(t *testing.T) {
	t.Parallel()

	s := newSpace(t)
	yesterday := today().AddDate(0, 0, -1)
	charged := []costLine{
		{at: yesterday.Add(30 * time.Minute), project: project, cost: 1.25, credits: []float64{-0.5, -0.25}},
		{at: noon(today().AddDate(0, 0, -31)), project: project, cost: 0.25},
		{at: yesterday.Add(23*time.Hour + 30*time.Minute), project: project, cost: 0.5},
		{at: noon(yesterday), project: "elsewhere", cost: 8},
		{at: yesterday.Add(-30 * time.Minute), project: project, cost: 0.75},
		{at: today().Add(30 * time.Minute), project: project, cost: 4},
	}
	s.export(t, standardExport, charged...)
	s.export(t, detailedExport, charged[:3]...)
	s.count(t, map[time.Time]int{yesterday: 3})
	s.makeDailyReport(t)

	got := s.rows(t, "SELECT days_ago, charged, credits, cost, learners FROM `${private}.daily_report` WHERE days_ago <= 3 OR days_ago = 31 ORDER BY day")
	want := [][]string{
		{"31", "0.25", "0", "0.25", "NULL"},
		{"3", "NULL", "NULL", "NULL", "NULL"},
		{"2", "0.75", "0", "0.75", "NULL"},
		{"1", "1.75", "-0.75", "1", "3"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if !sameNumbers(got[i], want[i]) {
			t.Errorf("%s days ago: got %v, want %v", want[i][0], got[i], want[i])
		}
	}
}

// sameNumbers is whether two rows hold the same numbers, however each writes
// them, and NULL where the other has NULL.
func sameNumbers(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] == "NULL" || want[i] == "NULL" {
			if got[i] != want[i] {
				return false
			}
			continue
		}
		g, errGot := strconv.ParseFloat(got[i], 64)
		w, errWant := strconv.ParseFloat(want[i], 64)
		if errGot != nil || errWant != nil || g != w {
			return false
		}
	}
	return true
}
