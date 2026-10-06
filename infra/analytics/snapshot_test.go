//go:build analytics

package analytics_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// snapshotSQL is the SQL the snapshot of the page "Research"'s live numbers is
// written from.
const snapshotSQL = "site/live.sql"

// benchSnapshot is the snapshot the bench makes the page's data of its tests
// from, which the bench reads as the snapshot's SQL writes it.
var benchSnapshot = filepath.Join("..", "..", "tools", "learners", "testdata", "live.json")

// snapshotViews are the public views the snapshot reads.
var snapshotViews = []string{"closed_months", "chances_total", "chances", "kept_up"}

// makeSnapshotViews makes the views the snapshot reads, and no other: every
// view a space makes costs the emulator memory it keeps, and a space with all
// of them beside the counted months runs it out of room.
func (s *space) makeSnapshotViews(t *testing.T) {
	t.Helper()

	for _, name := range snapshotViews {
		s.e.query(t, "CREATE VIEW `"+s.names["public"]+"."+name+"` AS "+s.render(t, "views/public/"+name+".sql", nil))
	}
}

// snapshotOf is the snapshot the SQL gives of a space: the one value of the
// one row it answers with.
func snapshotOf(t *testing.T, s *space) string {
	t.Helper()

	rows := s.e.query(t, s.render(t, snapshotSQL, nil))
	if len(rows) != 1 || len(rows[0]) != 1 {
		t.Fatalf("the snapshot's SQL answers %v, want one row of one value", rows)
	}
	return rows[0][0]
}

// wantSnapshot holds a snapshot to the one written out, value for value, its
// ranges in the order written.
func wantSnapshot(t *testing.T, got, want string) {
	t.Helper()

	if !reflect.DeepEqual(decoded(t, got), decoded(t, want)) {
		t.Errorf("the snapshot = %s, want %s", got, want)
	}
}

// decoded is a text of JSON, read.
func decoded(t *testing.T, text string) any {
	t.Helper()

	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatalf("read %q: %v", text, err)
	}
	return value
}

// keysOf are the paths of every key a value of JSON holds, an array's items
// under the array's path with [] after it.
func keysOf(value any, at string, into map[string]bool) {
	switch v := value.(type) {
	case map[string]any:
		for key, inner := range v {
			into[at+"."+key] = true
			keysOf(inner, at+"."+key, into)
		}
	case []any:
		for _, item := range v {
			keysOf(item, at+"[]", into)
		}
	}
}

// keyList is the paths of a value's keys, in order.
func keyList(value any) []string {
	keys := map[string]bool{}
	keysOf(value, "", keys)
	list := make([]string, 0, len(keys))
	for key := range keys {
		list = append(list, key)
	}
	slices.Sort(list)
	return list
}

// The snapshot is of the latest month counted whole, though an earlier one
// showed more: February's children were too few for its views to show any of
// its numbers, so the snapshot names February and shows neither its total nor
// a range, and January's numbers are not in it. March, under way, is no month
// of it.
func TestTheSnapshotIsOfTheLatestMonthCountedWhole(t *testing.T) {
	s := counted(t)

	wantSnapshot(t, snapshotOf(t, s), `{"month": "2026-02", "total": null, "chances": [], "kept_up": []}`)
}

// Before a month is counted whole the snapshot names none and shows nothing.
// Once one is, it shows the month's answers over every range of chance and the
// ranges of chance and of the child's answers the public views show, from the
// lowest, whatever order they were counted in, and none of those they hide;
// and it holds the keys the bench reads a snapshot by, no more and no fewer.
func TestTheSnapshotShowsWhatThePublicViewsShow(t *testing.T) {
	s := newSpace(t)
	s.makeSnapshotViews(t)

	wantSnapshot(t, snapshotOf(t, s), `{"month": null, "total": null, "chances": [], "kept_up": []}`)

	// January counted every day, February only its first ten.
	var days []string
	for d := january; d.Before(february.AddDate(0, 0, 10)); d = d.AddDate(0, 0, 1) {
		days = append(days, fmt.Sprintf("(DATE '%s', 5, NULL, NULL, 6, 7, 0)", day(d)))
	}
	s.rows(t, "INSERT INTO `${impact}.daily` (day, learners, learners_week, learners_month, tasks, answers, topics_won) VALUES "+strings.Join(days, ", "))
	// By range of chance: ten children and thirty answers, shown; twelve and
	// forty-one, shown as ten and forty; nine children, not shown. By the range
	// of the children's answers: ten children of one margin, shown; ten five
	// each way of their promise, shown; nine children, not shown.
	s.rows(t, "INSERT INTO `${impact}.chance_monthly` (month, bucket, learners, answers, correct, promised) VALUES "+
		"(DATE '2026-01-01', '0.70-0.77', 10, 30, 22, 2220), "+
		"(DATE '2026-01-01', '0.60-0.69', 12, 41, 28, 2665), "+
		"(DATE '2026-01-01', '0.78-0.85', 9, 200, 160, 16200), "+
		"(DATE '2026-01-01', NULL, 45, 400, 300, 30000)")
	s.rows(t, "INSERT INTO `${impact}.kept_up_monthly` (month, bucket, learners, answers, correct, promised, margins_squared, answers_by_margins, answers_squared) VALUES "+
		"(DATE '2026-01-01', '21-50', 10, 30, 20, 2700, 49000, -2100, 90), "+
		"(DATE '2026-01-01', '6-20', 10, 30, 24, 2400, 9000, 0, 90), "+
		"(DATE '2026-01-01', '101-200', 9, 300, 200, 21000, 1000, 10, 10000)")

	got := snapshotOf(t, s)
	wantSnapshot(t, got, `{
		"month": "2026-01",
		"total": {"learners": 45, "answers": 400, "promised_mean": 0.75, "correct_share": 0.75},
		"chances": [
			{"chance": "0.60-0.69", "learners": 10, "answers": 40, "promised_mean": 0.65, "correct_share": 0.683},
			{"chance": "0.70-0.77", "learners": 10, "answers": 30, "promised_mean": 0.74, "correct_share": 0.733}
		],
		"kept_up": [
			{"answers_range": "6-20", "learners": 10, "answers": 30, "came_true_less_promised": 0, "standard_error": 0.033},
			{"answers_range": "21-50", "learners": 10, "answers": 30, "came_true_less_promised": -0.233, "standard_error": 0}
		]
	}`)

	raw, err := os.ReadFile(benchSnapshot)
	if err != nil {
		t.Fatalf("read the bench's snapshot: %v", err)
	}
	if made, read := keyList(decoded(t, got)), keyList(decoded(t, string(raw))); !slices.Equal(made, read) {
		t.Errorf("the snapshot's SQL writes the keys %v, and the bench reads %v", made, read)
	}
}

// The snapshot's SQL names the public views and nothing else: a table it could
// name beside them holds what no rule has made fit to show.
func TestTheSnapshotReadsThePublicViewsAlone(t *testing.T) {
	t.Parallel()

	for _, named := range placeholder.FindAllStringSubmatch(read(t, snapshotSQL), -1) {
		if named[1] != "public" {
			t.Errorf("the snapshot's SQL names ${%s}, want ${public} alone", named[1])
		}
	}
}

// The snapshot is one value in one column, live, the name the table it is
// kept in is read by.
func TestTheSnapshotIsOneValueNamedLive(t *testing.T) {
	s := counted(t)

	if got := s.e.columns(t, s.render(t, snapshotSQL, nil)); !slices.Equal(got, []string{"live"}) {
		t.Errorf("the snapshot's SQL answers with the columns %v, want [live] alone", got)
	}
}
