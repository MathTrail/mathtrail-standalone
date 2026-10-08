package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// shownTotal is a month's total the public views would show.
const shownTotal = `{"learners": 45, "answers": 1235, "promised_mean": 0.754, "correct_share": 0.734}`

// snapshotOf writes a snapshot with every key, each given as JSON.
func snapshotOf(month, total, chances, keptUp string) string {
	return fmt.Sprintf(`{"month": %s, "total": %s, "chances": %s, "kept_up": %s}`, month, total, chances, keptUp)
}

// shownSnapshot is the made-up snapshot of a month whose total is shown that
// the bench of the site's page reads, so that a change to the snapshot's shape
// reaches these tests as soon as it reaches the page.
var shownSnapshot = filepath.Join("..", "..", "..", "tools", "learners", "testdata", "live.json")

// readSnapshot reads a snapshot a test starts from.
func readSnapshot(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestParseReadsTheStateOfEachSnapshot(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, path string
		want       Numbers
	}{
		// The snapshot the service gave before its first month was counted whole.
		{"no month counted", filepath.Join("testdata", "coming.json"), Numbers{State: stateComing}},
		{"too few children", filepath.Join("testdata", "too-few.json"), Numbers{State: stateTooFew, Year: 2026, Month: 11}},
		{"a total shown", shownSnapshot, Numbers{State: stateShown, Year: 2026, Month: 11, Learners: 45, Answers: 1235, PromisedMean: 0.754, CorrectShare: 0.734}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(readSnapshot(t, c.path))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got != c.want {
				t.Errorf("Parse = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestParseRefusesWhatThePublicViewsWouldNotShow(t *testing.T) {
	t.Parallel()
	month := `"2026-11"`
	cases := []struct {
		name, snapshot, want string
	}{
		{"a total of fewer children than the rule needs",
			snapshotOf(month, `{"learners": 5, "answers": 100, "promised_mean": 0.7, "correct_share": 0.7}`, `[]`, `[]`), "only behind 10 children and 30 answers"},
		{"a total of fewer answers than the rule needs",
			snapshotOf(month, `{"learners": 10, "answers": 25, "promised_mean": 0.7, "correct_share": 0.7}`, `[]`, `[]`), "only behind 10 children and 30 answers"},
		{"children not rounded as the rule rounds them",
			snapshotOf(month, `{"learners": 12, "answers": 100, "promised_mean": 0.7, "correct_share": 0.7}`, `[]`, `[]`), "round every count to 5"},
		{"answers not rounded as the rule rounds them",
			snapshotOf(month, `{"learners": 10, "answers": 101, "promised_mean": 0.7, "correct_share": 0.7}`, `[]`, `[]`), "round every count to 5"},
		{"a share right above one",
			snapshotOf(month, `{"learners": 10, "answers": 100, "promised_mean": 0.7, "correct_share": 1.2}`, `[]`, `[]`), "both are shares"},
		{"a chance promised below nothing",
			snapshotOf(month, `{"learners": 10, "answers": 100, "promised_mean": -0.1, "correct_share": 0.7}`, `[]`, `[]`), "both are shares"},
		{"a total of no month", snapshotOf(`null`, shownTotal, `[]`, `[]`), "numbers of no month"},
		{"ranges of no month", snapshotOf(`null`, `null`, `[{}]`, `[]`), "numbers of no month"},
		{"ranges of a month with no total", snapshotOf(month, `null`, `[]`, `[{}]`), "no total of"},
		{"a month past the twelfth", snapshotOf(`"2026-13"`, `null`, `[]`, `[]`), "want a year and a month"},
		{"a month in words", snapshotOf(`"November 2026"`, `null`, `[]`, `[]`), "want a year and a month"},
		{"a total holding a null",
			snapshotOf(month, `{"learners": null, "answers": 100, "promised_mean": 0.7, "correct_share": 0.7}`, `[]`, `[]`), "holds a null"},
		{"a key it does not know", `{"month": null, "total": null, "chances": [], "kept_up": [], "deployment": "x"}`, "unknown field"},
		{"a key of the total it does not know",
			snapshotOf(month, `{"learners": 45, "answers": 1235, "promised_mean": 0.754, "correct_share": 0.734, "children": 45}`, `[]`, `[]`), "unknown field"},
		{"a key left out", `{"month": null, "total": null, "chances": []}`, "left out"},
		{"a key of the total left out",
			snapshotOf(month, `{"learners": 45, "answers": 1235, "promised_mean": 0.754}`, `[]`, `[]`), "left out"},
		{"two values", snapshotOf(`null`, `null`, `[]`, `[]`) + snapshotOf(`null`, `null`, `[]`, `[]`), "more than one value"},
		{"no JSON at all", "month=2026-11", "read the snapshot"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parse([]byte(c.snapshot))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("Parse(%s) = %+v, %v; want an error containing %q", c.snapshot, got, err, c.want)
			}
		})
	}
}

func TestRenderWritesTheKeysOfItsStateAlone(t *testing.T) {
	t.Parallel()
	const header = "# The service's real use, as its latest monthly snapshot of public totals shows it.\n"
	cases := []struct {
		name    string
		numbers Numbers
		want    string
	}{
		{"no month counted", Numbers{State: stateComing}, header + "state=0\n"},
		{"too few children", Numbers{State: stateTooFew, Year: 2026, Month: 11}, header + "state=1\nyear=2026\nmonth=11\n"},
		{"a total shown",
			Numbers{State: stateShown, Year: 2027, Month: 1, Learners: 45, Answers: 1235, PromisedMean: 0.754, CorrectShare: 0.7},
			header + "state=2\nyear=2027\nmonth=1\nlearners=45\nanswers=1235\npromised_mean=0.754\ncorrect_share=0.7\n"},
	}
	// The lines the paper's macros read: a key of lower-case letters, digits
	// and underscores, and a value; or a comment.
	macroLine := regexp.MustCompile(`^([a-z0-9_]+=\S+|#.*)$`)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := Render(c.numbers)
			if got != c.want {
				t.Errorf("Render(%+v) =\n%s\nwant\n%s", c.numbers, got, c.want)
			}
			for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
				if !macroLine.MatchString(line) {
					t.Errorf("Render wrote %q, which the paper's macros cannot read", line)
				}
			}
		})
	}
}

func TestRunReadsNoSnapshotAsNoMonthCounted(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out := filepath.Join(dir, "generated", "live.txt")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-snapshot", filepath.Join(dir, "live.json"), "-out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("run = %d, stderr %s", code, stderr.String())
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(got), "\nstate=0\n") {
		t.Errorf("with no snapshot the file reads\n%s\nwant state=0 alone", got)
	}
}

func TestRunLeavesTheFileAsItWasWhenTheSnapshotIsRefused(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	snapshot := filepath.Join(dir, "live.json")
	out := filepath.Join(dir, "live.txt")
	if err := os.WriteFile(snapshot, readSnapshot(t, shownSnapshot), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-snapshot", snapshot, "-out", out}, &stdout, &stderr); code != 0 {
		t.Fatalf("run = %d, stderr %s", code, stderr.String())
	}
	good, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(good), "\nlearners=45\n") {
		t.Fatalf("the written file lacks the month's children:\n%s", good)
	}

	hidden := snapshotOf(`"2026-12"`, `{"learners": 5, "answers": 40, "promised_mean": 0.7, "correct_share": 0.7}`, `[]`, `[]`)
	if werr := os.WriteFile(snapshot, []byte(hidden), 0o600); werr != nil {
		t.Fatal(werr)
	}
	stderr.Reset()
	if code := run([]string{"-snapshot", snapshot, "-out", out}, &stdout, &stderr); code != 1 {
		t.Fatalf("run with a total the views hide = %d, want 1", code)
	}
	after, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, good) {
		t.Errorf("a refused snapshot changed the file:\n%s\nwant\n%s", after, good)
	}
	if !strings.Contains(stderr.String(), "only behind 10 children") {
		t.Errorf("stderr = %q, want it to name the rule the total breaks", stderr.String())
	}
}

func TestRunFailsWhenTheFileCannotBeWritten(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	// A folder stands where the file would be written.
	code := run([]string{"-snapshot", filepath.Join(dir, "live.json"), "-out", dir}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "write "+dir) {
		t.Errorf("run with a folder for the file = %d, stderr %q; want 1 and the write that failed", code, stderr.String())
	}
}

func TestRunNeedsTheSnapshotAndTheFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
	}{
		{"no file", []string{"-snapshot", "live.json"}},
		{"no snapshot", []string{"-out", "live.txt"}},
		{"a word too many", []string{"-snapshot", "live.json", "-out", "live.txt", "extra"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			if code := run(c.args, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage:") {
				t.Errorf("run(%q) = %d, stderr %q; want 2 and the usage", c.args, code, stderr.String())
			}
		})
	}
}

// The rule a total is held to is the rule the public view of the months'
// totals shows it by: the same least children and answers, and the same
// multiple its counts are rounded to, as the view's SQL writes them.
func TestTheRuleIsTheRuleOfThePublicView(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "..", "..", "infra", "analytics", "views", "public", "chances_total.sql")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	leasts := regexp.MustCompile(`learners >= (\d+) AND counted\.answers >= (\d+)`).FindAllStringSubmatch(sql, -1)
	roundings := regexp.MustCompile(`ROUND\([\w.]+ / (\d+)\) \* (\d+)`).FindAllStringSubmatch(sql, -1)
	if len(leasts) == 0 || len(roundings) == 0 {
		t.Fatalf("%s shows its totals by no rule this test can read", path)
	}
	for _, found := range leasts {
		if found[1] != strconv.Itoa(rule.Learners) || found[2] != strconv.Itoa(rule.Answers) {
			t.Errorf("%s shows a total of %s children and %s answers or more, and the paper's are held to %d and %d",
				path, found[1], found[2], rule.Learners, rule.Answers)
		}
	}
	for _, found := range roundings {
		if found[1] != strconv.Itoa(rule.RoundedTo) || found[2] != found[1] {
			t.Errorf("%s rounds a count by %s and %s, and the paper's are held to %d", path, found[1], found[2], rule.RoundedTo)
		}
	}
}
