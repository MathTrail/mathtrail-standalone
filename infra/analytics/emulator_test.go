//go:build analytics

package analytics_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// emulatorEnv names the address the BigQuery emulator answers at. The recipe
// that runs these tests starts the emulator and sets it.
const emulatorEnv = "MATHTRAIL_BIGQUERY_EMULATOR"

// project is the project the emulator was started with.
const project = "mathtrail"

// emulator is a BigQuery emulator, reached over the REST API BigQuery has.
type emulator struct {
	base   string
	client *http.Client
}

func emulatorOf(t *testing.T) *emulator {
	t.Helper()

	base := os.Getenv(emulatorEnv)
	if base == "" {
		t.Fatalf("%s is not set: these tests run with just analytics-test, which starts the emulator", emulatorEnv)
	}
	return &emulator{base: strings.TrimSuffix(base, "/") + "/bigquery/v2/projects/" + project, client: &http.Client{Timeout: time.Minute}}
}

// post sends a body to a path of the API and is what it answered. An answer
// that carries an error fails the test with it.
func (e *emulator) post(t *testing.T, path string, body any) map[string]any {
	t.Helper()

	sent, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("encode the request to %s: %v", path, err)
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, e.base+path, bytes.NewReader(sent))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := e.client.Do(request)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer func() { _ = response.Body.Close() }()
	var answer map[string]any
	if err := json.NewDecoder(response.Body).Decode(&answer); err != nil {
		t.Fatalf("POST %s: read the answer: %v", path, err)
	}
	if failure, failed := answer["error"].(map[string]any); failed {
		t.Fatalf("POST %s: %v", path, failure["message"])
	}
	return answer
}

// columns are the names of the columns a query answers with.
func (e *emulator) columns(t *testing.T, sql string) []string {
	t.Helper()

	answer := e.post(t, "/queries", map[string]any{"query": sql, "useLegacySql": false})
	schema, _ := answer["schema"].(map[string]any)
	fields, _ := schema["fields"].([]any)
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i], _ = field.(map[string]any)["name"].(string)
	}
	return names
}

// query runs SQL — a statement or a script — and is the rows it ends with,
// each value as the API writes it, or "NULL".
func (e *emulator) query(t *testing.T, sql string) [][]string {
	t.Helper()

	answer := e.post(t, "/queries", map[string]any{"query": sql, "useLegacySql": false})
	listed, _ := answer["rows"].([]any)
	rows := make([][]string, 0, len(listed))
	for _, row := range listed {
		fields, _ := row.(map[string]any)
		cells, _ := fields["f"].([]any)
		values := make([]string, len(cells))
		for i, cell := range cells {
			value, _ := cell.(map[string]any)
			values[i] = fmt.Sprint(value["v"])
			if values[i] == "<nil>" {
				values[i] = "NULL"
			}
		}
		rows = append(rows, values)
	}
	return rows
}

// spaces numbers the datasets each test makes, so that no two share any, and
// run sets this run's apart from those of an earlier run against the same
// emulator.
var (
	spaces atomic.Int64
	run    = strconv.FormatInt(time.Now().UnixNano(), 36)
)

// space is a test's own datasets in the emulator — the raw lines, the counts
// and the two sets of views — and the names the SQL is written with, filled in
// as Terraform fills them.
type space struct {
	e     *emulator
	names map[string]string
}

// newSpace makes a test's datasets: the raw lines in a table of the shape the
// linked dataset of a log bucket has, the tables of the counts from their
// schemas, and empty datasets for the views.
func newSpace(t *testing.T) *space {
	t.Helper()

	e := emulatorOf(t)
	prefix := fmt.Sprintf("r%s_t%d_", run, spaces.Add(1))
	for _, dataset := range []string{"logs", "impact", "public", "private"} {
		e.post(t, "/datasets", map[string]any{"datasetReference": map[string]string{"projectId": project, "datasetId": prefix + dataset}})
	}
	s := &space{e: e, names: map[string]string{
		"logs":    project + "." + prefix + "logs._AllLogs",
		"impact":  project + "." + prefix + "impact",
		"public":  project + "." + prefix + "public",
		"private": project + "." + prefix + "private",
		"events":  eventsList(t),
		"fold":    read(t, "views/public/fold.sql"),
	}}
	e.query(t, "CREATE TABLE `"+s.names["logs"]+"` (timestamp TIMESTAMP, log_name STRING, insert_id STRING, json_payload JSON)")
	for _, table := range tableNames(t) {
		var fields []any
		if err := json.Unmarshal([]byte(read(t, "tables/"+table+".json")), &fields); err != nil {
			t.Fatalf("read the schema of %s: %v", table, err)
		}
		e.post(t, "/datasets/"+prefix+"impact/tables", map[string]any{
			"tableReference": map[string]string{"projectId": project, "datasetId": prefix + "impact", "tableId": table},
			"schema":         map[string]any{"fields": fields},
		})
	}
	return s
}

// countTheNight runs the nightly script.
func (s *space) countTheNight(t *testing.T) {
	t.Helper()
	s.e.query(t, s.render(t, "nightly.sql", nil))
}

// makeViews makes every view, in the order Terraform makes them.
func (s *space) makeViews(t *testing.T) {
	t.Helper()

	view := func(name, file string, more map[string]string) {
		s.e.query(t, "CREATE VIEW `"+name+"` AS "+s.render(t, file, more))
	}
	for _, file := range sqlFiles(t, "views/private") {
		view(s.names["private"]+"."+strings.TrimSuffix(file, ".sql"), "views/private/"+file, nil)
	}
	view(s.names["public"]+".closed_months", "views/public/closed_months.sql", nil)
	for _, dimension := range dimensions {
		view(s.names["public"]+".learners_by_"+dimension, "views/public/learners_by.sql", map[string]string{"dimension": dimension})
	}
	for _, file := range sqlFiles(t, "views/public") {
		if name := strings.TrimSuffix(file, ".sql"); name != "fold" && name != "closed_months" && name != "learners_by" {
			view(s.names["public"]+"."+name, "views/public/"+file, nil)
		}
	}
}

// rows are the rows of a query of the space's, its names filled in.
func (s *space) rows(t *testing.T, sql string) [][]string {
	t.Helper()
	return s.e.query(t, fill(t, sql, s.names))
}

// render is a file of SQL with the space's names filled in, and more beside
// them, and with its comment lines left out. That much is the emulator's: it
// splits a script into statements without reading its comments, so an
// apostrophe in one opens a string, and a comment before the first statement
// hides that what follows is a script. BigQuery reads the files as they are.
func (s *space) render(t *testing.T, file string, more map[string]string) string {
	t.Helper()

	names := map[string]string{}
	for name, value := range s.names {
		names[name] = value
	}
	for name, value := range more {
		names[name] = value
	}
	var kept []string
	for _, line := range strings.Split(fill(t, read(t, file), names), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

// placeholder is a name the SQL is written with, as templatefile writes one.
var placeholder = regexp.MustCompile(`\$\{(\w+)\}`)

// fill puts the names in, as templatefile does; a name the SQL uses and is
// not given fails the test, as it fails Terraform.
func fill(t *testing.T, sql string, names map[string]string) string {
	t.Helper()

	return placeholder.ReplaceAllStringFunc(sql, func(found string) string {
		name := placeholder.FindStringSubmatch(found)[1]
		value, given := names[name]
		if !given {
			t.Fatalf("the SQL names ${%s}, which nothing fills in", name)
		}
		return value
	})
}

// read is a file of this package's.
func read(t *testing.T, file string) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.FromSlash(file))
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	return string(raw)
}

// eventsList is the list of events the lines are copied for, as the SQL takes
// it: the names in quotes, one after another.
func eventsList(t *testing.T) string {
	t.Helper()

	var events []string
	if err := json.Unmarshal([]byte(read(t, "events.json")), &events); err != nil {
		t.Fatalf("read the events: %v", err)
	}
	quoted := make([]string, len(events))
	for i, event := range events {
		quoted[i] = `"` + event + `"`
	}
	return strings.Join(quoted, ", ")
}

// tableNames are the tables of the counts, by their schemas.
func tableNames(t *testing.T) []string {
	t.Helper()

	var names []string
	for _, file := range filesIn(t, "tables", ".json") {
		names = append(names, strings.TrimSuffix(file, ".json"))
	}
	return names
}

// sqlFiles are the files of SQL in a folder of views.
func sqlFiles(t *testing.T, folder string) []string {
	t.Helper()
	return filesIn(t, folder, ".sql")
}

func filesIn(t *testing.T, folder, suffix string) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.FromSlash(folder))
	if err != nil {
		t.Fatalf("list %s: %v", folder, err)
	}
	var files []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), suffix) {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files
}

// dimensions are what the children are told apart by, one at a time.
var dimensions = []string{"host", "language", "grade", "country", "region", "signin_country", "cohort"}

// line is one line of the service's as the linked dataset keeps it: when it
// was written, its id in the log, and its fields.
type line struct {
	at       time.Time
	insertID string
	fields   map[string]any
}

// write puts lines into the space's raw lines.
func (s *space) write(t *testing.T, lines ...line) {
	t.Helper()

	values := make([]string, len(lines))
	for i, l := range lines {
		payload, err := json.Marshal(l.fields)
		if err != nil {
			t.Fatalf("encode a line: %v", err)
		}
		values[i] = fmt.Sprintf("(TIMESTAMP '%s', 'projects/%s/logs/run.googleapis.com%%2Fstdout', '%s', JSON '%s')",
			l.at.UTC().Format("2006-01-02 15:04:05.000000+00"), project, l.insertID, quoteInSQL(string(payload)))
	}
	s.e.query(t, "INSERT INTO `"+s.names["logs"]+"` (timestamp, log_name, insert_id, json_payload) VALUES "+strings.Join(values, ", "))
}

// quoteInSQL is a text as it stands inside a string of SQL in single quotes.
func quoteInSQL(text string) string {
	return strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(text)
}

// lines number the lines the tests write, so that each has an id of its own.
var lines atomic.Int64

// fieldsOf are the fields a line of an event carries, the given ones over the
// usual ones.
func fieldsOf(event, learner string, usual, given map[string]any) map[string]any {
	fields := map[string]any{"message": event, "learner": learner}
	for name, value := range usual {
		fields[name] = value
	}
	for name, value := range given {
		fields[name] = value
	}
	return fields
}

// accepted is the line of a task handed to a child, on Claude in English to a
// child of grade 2 in Texas unless given otherwise.
func accepted(learner string, at time.Time, given map[string]any) line {
	return line{at: at, insertID: fmt.Sprintf("line-%d", lines.Add(1)), fields: fieldsOf("task_accepted", learner, map[string]any{
		"host": "claude", "language": "en", "grade": 2, "cohort": "2026-01",
		"country": "US", "region": "US-TX", "signin_country": "US",
	}, given)}
}

// answered is the line of an answer of a child, right and with no hint unless
// given otherwise.
func answered(learner string, at time.Time, given map[string]any) line {
	return line{at: at, insertID: fmt.Sprintf("line-%d", lines.Add(1)), fields: fieldsOf("answer_recorded", learner, map[string]any{
		"topic": "logic.ordering", "trap": "", "correct": true, "hint_used": false, "confused": false,
		"grade": 2, "cohort": "2026-01", "topics_mastered": 0,
	}, given)}
}

// won is the line of an answer that brought a topic among those mastered.
func won(learner string, at time.Time, given map[string]any) line {
	return line{at: at, insertID: fmt.Sprintf("line-%d", lines.Add(1)), fields: fieldsOf("topic_mastered", learner, map[string]any{
		"topic": "logic.ordering", "grade": 2,
	}, given)}
}

// learnerOf is the name the n-th child of a test is counted under: sixteen
// characters, as the service writes it.
func learnerOf(n int) string {
	return fmt.Sprintf("child%011d", n)
}

// today is the day in UTC the emulator counts from, and the days around it.
func today() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour)
}

// day is the day as the tables write it.
func day(at time.Time) string {
	return at.Format(time.DateOnly)
}

// noon is the middle of a day, away from either of its ends.
func noon(at time.Time) time.Time {
	return at.Add(12 * time.Hour)
}

// firstOfMonth is the first day of the month a day falls in.
func firstOfMonth(at time.Time) time.Time {
	return time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, time.UTC)
}
