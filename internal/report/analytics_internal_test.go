package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// analytics is where the SQL of the counts kept for years lives.
var analytics = filepath.Join("..", "..", "infra", "analytics")

// The lines copied to be counted for years are the lines the children are
// counted from, no more and no fewer: the list Terraform builds the copying
// from and the list this package holds the fields of those lines to are one.
func TestTheLinesKeptToBeCountedAreTheLinesChildrenAreCountedFrom(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join(analytics, "events.json"))
	if err != nil {
		t.Fatalf("read the events: %v", err)
	}
	var events []string
	if err := json.Unmarshal(raw, &events); err != nil {
		t.Fatalf("read the events: %v", err)
	}
	if !slices.Equal(events, countedFrom) {
		t.Errorf("the lines copied to be counted = %v, want %v", events, countedFrom)
	}
}

// columnWords are the only words a column of a table kept for years may be
// called: periods, what a group is told apart by, counts, and the sums a mean
// and its error are read from. A column called anything else — the name a
// child is counted under, an account, a request — is refused here before any
// table holds it.
var columnWords = []string{
	"day", "week", "month",
	"dimension", "value", "measure", "bucket", "tenure", "topic", "grade", "trap",
	"learners", "learners_week", "learners_month", "tasks", "answers", "topics_won",
	"correct", "hinted", "dont_know", "mastered", "won", "winners",
	"promised", "margins_squared", "answers_by_margins", "answers_squared",
}

// Every column of every table the counts are kept in for years is a word of
// the closed list: no table can keep a child's name, or anything else, beside
// the counts.
func TestATableKeptForYearsHoldsCountsAlone(t *testing.T) {
	t.Parallel()

	schemas, err := filepath.Glob(filepath.Join(analytics, "tables", "*.json"))
	if err != nil || len(schemas) == 0 {
		t.Fatalf("the schemas of the tables = %v, %v; want some", schemas, err)
	}
	for _, schema := range schemas {
		raw, err := os.ReadFile(schema)
		if err != nil {
			t.Fatalf("read %s: %v", schema, err)
		}
		var columns []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &columns); err != nil {
			t.Fatalf("read %s: %v", schema, err)
		}
		for _, column := range columns {
			if !slices.Contains(columnWords, column.Name) {
				t.Errorf("%s has a column %q, which is no word of the counts", filepath.Base(schema), column.Name)
			}
		}
	}
}

// The night counts the answers weighed against the chance promised in the
// ranges this package counts them in, so that the counts kept for years and
// the report read a chance alike: the same highest chance in hundredths for
// each range, under the same name, and the last for every chance past them.
func TestTheNightCountsChancesInTheReportsRanges(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join(analytics, "nightly.sql"))
	if err != nil {
		t.Fatalf("read the night: %v", err)
	}
	var counted []chanceRange
	for _, found := range regexp.MustCompile(`WHEN chance <= (\d+) THEN '([^']+)'`).FindAllStringSubmatch(string(raw), -1) {
		upTo, _ := strconv.Atoi(found[1])
		counted = append(counted, chanceRange{upTo: upTo, named: found[2]})
	}
	last := regexp.MustCompile(`WHEN chance <= \d+ THEN '[^']+'\s+ELSE '([^']+)' END`).FindStringSubmatch(string(raw))
	if last == nil {
		t.Fatal("the night names no range for a chance past the last it bounds")
	}
	counted = append(counted, chanceRange{upTo: 100, named: last[1]})
	if !slices.Equal(counted, chanceRanges) {
		t.Errorf("the night counts chances in %v, want the report's %v", counted, chanceRanges)
	}
}

// A public view of children told apart by one thing folds its small groups by
// the one rule written once, put in as it is written.
func TestThePublicViewsFoldByTheOneRule(t *testing.T) {
	t.Parallel()

	for _, view := range []string{"learners_by.sql", "learning.sql"} {
		raw, err := os.ReadFile(filepath.Join(analytics, "views", "public", view))
		if err != nil {
			t.Fatalf("read %s: %v", view, err)
		}
		if !strings.Contains(string(raw), "${fold}") {
			t.Errorf("%s does not fold its groups by the rule of views/public/fold.sql", view)
		}
	}
}

// No file of the SQL holds what templatefile reads otherwise than a name to
// fill in, so the files the tests fill in are the ones BigQuery is handed.
func TestTheSQLHoldsOnlyNamesToFillIn(t *testing.T) {
	t.Parallel()

	err := filepath.WalkDir(analytics, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".sql" {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, directive := range []string{"%{", "$${"} {
			if strings.Contains(string(raw), directive) {
				t.Errorf("%s holds %s, which templatefile reads as more than a name to fill in", path, directive)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the SQL: %v", err)
	}
}
