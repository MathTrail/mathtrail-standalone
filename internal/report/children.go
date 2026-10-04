package report

import (
	"maps"
	"slices"
	"time"
)

// eventTopicMastered is the line of an answer that brought a topic among those
// the progress calls mastered.
const eventTopicMastered = "topic_mastered"

// notChildren are the hosts whose calls hand a task to no child a family has:
// MCP Inspector and the service's own load tool. A child either of them handed
// a task to is counted nowhere, and neither is anything else its name is on.
var notChildren = []string{"inspector", "load"}

// childrenOn is what one day adds up to in the counts kept for years: the
// children a task was handed to, by the name each is counted under, and the
// tasks handed out, the answers recorded and the topics won.
type childrenOn struct {
	children            map[string]struct{}
	tasks, answers, won int
}

// childrenByDay adds up, a day at a time in UTC, the lines the children are
// counted from, by the rules the counts kept for years follow, so that the two
// can be laid side by side: a line with no name to count a child by is not
// counted, nor one the log repeated, nor a child the hosts of notChildren
// handed a task to. A child is counted by the name it has that month, which is
// all a line knows of it.
func childrenByDay(lines []line) map[time.Time]*childrenOn {
	left := map[string]bool{}
	for i := range lines {
		if l := &lines[i]; l.Message == eventTaskAccepted && slices.Contains(notChildren, l.Host) {
			left[l.Learner] = true
		}
	}
	days := map[time.Time]*childrenOn{}
	for i := range lines {
		l := &lines[i]
		if l.repeat || l.Learner == "" || left[l.Learner] || l.Time.IsZero() || !slices.Contains(countedFrom, l.Message) {
			continue
		}
		day := dayOf(l.Time)
		on := days[day]
		if on == nil {
			on = &childrenOn{children: map[string]struct{}{}}
			days[day] = on
		}
		switch l.Message {
		case eventTaskAccepted:
			on.children[l.Learner] = struct{}{}
			on.tasks++
		case eventAnswerRecorded:
			on.answers++
		case eventTopicMastered:
			on.won++
		}
	}
	return days
}

// dayOf is the day a moment falls on in UTC.
func dayOf(at time.Time) time.Time {
	year, month, day := at.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// childrenTable is every day from the first the lines count a child on to the
// last, a day with nothing on it among them: the counts kept for years have a
// row for each day too.
func (c *counts) childrenTable() *table {
	t := &table{columns: []string{"Day", "Children", "Tasks", "Answers", "Topics won"}, named: 1}
	if len(c.children) == 0 {
		return t
	}
	days := slices.SortedFunc(maps.Keys(c.children), time.Time.Compare)
	for day := days[0]; !day.After(days[len(days)-1]); day = day.AddDate(0, 0, 1) {
		on := c.children[day]
		if on == nil {
			on = &childrenOn{}
		}
		t.add(day.Format(time.DateOnly), number(len(on.children)), number(on.tasks), number(on.answers), number(on.won))
	}
	return t
}
