package report

import (
	"slices"
	"testing"
	"time"
)

// childrenRow is one day of the children's table, as it reads.
type childrenRow struct {
	day                              string
	children, tasks, answers, topics string
}

// childrenRows is the children's table of the lines, a row at a time.
func childrenRows(t *testing.T, lines ...string) []childrenRow {
	t.Helper()

	var rows []childrenRow
	for _, cells := range tallied(t, lines...).childrenTable().rows {
		if len(cells) != 5 {
			t.Fatalf("a row of the children's table = %v, want five cells", cells)
		}
		rows = append(rows, childrenRow{cells[0], cells[1], cells[2], cells[3], cells[4]})
	}
	return rows
}

// A day counts each child once, by the name it has that month, on a day a task
// was handed to it; the tasks, the answers and the topics won count as many as
// there were. A day is a day in UTC, whichever way the time was written.
func TestChildrenAreCountedADayAtATime(t *testing.T) {
	t.Parallel()

	got := childrenRows(t,
		`{"message":"task_accepted","learner":"aaaaaaaaaaaaaaaa","host":"claude","time":"2026-11-02T23:59:30Z"}`,
		`{"message":"task_accepted","learner":"aaaaaaaaaaaaaaaa","host":"claude","time":"2026-11-02T08:00:00Z"}`,
		`{"message":"task_accepted","learner":"bbbbbbbbbbbbbbbb","host":"chatgpt","time":"2026-11-02T20:00:00-05:00"}`,
		`{"message":"answer_recorded","learner":"aaaaaaaaaaaaaaaa","time":"2026-11-03T00:00:10Z"}`,
		`{"message":"topic_mastered","learner":"aaaaaaaaaaaaaaaa","time":"2026-11-03T00:00:10Z"}`,
	)
	want := []childrenRow{
		{"2026-11-02", "1", "2", "0", "0"},
		{"2026-11-03", "1", "1", "1", "1"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("the children's table = %v, want %v", got, want)
	}
}

// A child the load tool or MCP Inspector handed a task to is no child a family
// has: nothing its name is on is counted, an answer on another day included.
func TestAChildOfATestingToolIsNotCounted(t *testing.T) {
	t.Parallel()

	for _, host := range notChildren {
		t.Run(host, func(t *testing.T) {
			t.Parallel()

			got := childrenRows(t,
				`{"message":"task_accepted","learner":"aaaaaaaaaaaaaaaa","host":"claude","time":"2026-11-02T10:00:00Z"}`,
				`{"message":"answer_recorded","learner":"tttttttttttttttt","time":"2026-11-02T09:00:00Z"}`,
				`{"message":"task_accepted","learner":"tttttttttttttttt","host":"`+host+`","time":"2026-11-03T10:00:00Z"}`,
				`{"message":"topic_mastered","learner":"tttttttttttttttt","time":"2026-11-03T10:05:00Z"}`,
			)
			want := []childrenRow{{"2026-11-02", "1", "1", "0", "0"}}
			if !slices.Equal(got, want) {
				t.Errorf("the children's table = %v, want only the child of the chat", got)
			}
		})
	}
}

// A line the log handed over twice is one line, and a line with no name to
// count a child by counts no child: the lines of a build that wrote none.
func TestARepeatedLineAndALineWithNoNameAreNotCountedAgain(t *testing.T) {
	t.Parallel()

	line := `{"message":"task_accepted","learner":"aaaaaaaaaaaaaaaa","host":"claude","time":"2026-11-02T10:00:00Z"}`
	got := childrenRows(t, line, line,
		`{"message":"task_accepted","host":"claude","time":"2026-11-02T11:00:00Z"}`,
	)
	want := []childrenRow{{"2026-11-02", "1", "1", "0", "0"}}
	if !slices.Equal(got, want) {
		t.Errorf("the children's table = %v, want the line counted once", got)
	}
}

// Every day from the first to the last has a row, a day with nothing on it
// among them, as the counts kept for years have.
func TestEveryDayBetweenTheFirstAndTheLastHasARow(t *testing.T) {
	t.Parallel()

	got := childrenRows(t,
		`{"message":"task_accepted","learner":"aaaaaaaaaaaaaaaa","host":"claude","time":"2026-10-31T10:00:00Z"}`,
		`{"message":"answer_recorded","learner":"aaaaaaaaaaaaaaaa","time":"2026-11-02T10:00:00Z"}`,
	)
	want := []childrenRow{
		{"2026-10-31", "1", "1", "0", "0"},
		{"2026-11-01", "0", "0", "0", "0"},
		{"2026-11-02", "0", "0", "1", "0"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("the children's table = %v, want %v", got, want)
	}
}

// A line with no time falls on no day.
func TestALineWithNoTimeIsOnNoDay(t *testing.T) {
	t.Parallel()

	if got := childrenRows(t, `{"message":"task_accepted","learner":"aaaaaaaaaaaaaaaa","host":"claude"}`); len(got) != 0 {
		t.Errorf("the children's table = %v, want no day", got)
	}
	if got := dayOf(time.Date(2026, 11, 2, 23, 0, 0, 0, time.FixedZone("UTC-5", -5*3600))); got.Format(time.DateOnly) != "2026-11-03" {
		t.Errorf("dayOf() = %s, want the day in UTC, 2026-11-03", got.Format(time.DateOnly))
	}
}
