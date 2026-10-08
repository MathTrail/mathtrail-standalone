package report

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// update rewrites the report the lines in testdata add up to. The report is
// this package's own output, so a change to it is regenerated and the diff
// read in the review.
var update = flag.Bool("update", false, "rewrite testdata/report.md from testdata/lines.jsonl")

// The lines of a few lessons, as the service writes them, add up to the report
// kept beside them: two versions of the instructions, two chat hosts and a
// task whose call is not among the lines, a request refused to the last
// attempt, answers weighed against their chance — enough of them in one cell
// to read, too few in another — and answers of every kind the weighing leaves
// out, the limits reached, every way a tool call ends, and lines that are not
// the service's; and a minute of a deployed service's log, with two instances,
// calls to Drive, kept and dropped traces, deliveries that failed, and lines
// that break the rules of the log.
func TestTheLinesOfSomeLessonsAddUpToTheirReport(t *testing.T) {
	t.Parallel()

	in, err := os.Open(filepath.Join("testdata", "lines.jsonl"))
	if err != nil {
		t.Fatalf("open the lines: %v", err)
	}
	defer func() { _ = in.Close() }()
	var out bytes.Buffer
	if err = Run(in, &out); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	kept := filepath.Join("testdata", "report.md")
	if *update {
		if err = os.WriteFile(kept, out.Bytes(), 0o600); err != nil {
			t.Fatalf("rewrite the report: %v", err)
		}
	}
	want, err := os.ReadFile(kept)
	if err != nil {
		t.Fatalf("read the report: %v", err)
	}
	if got := out.String(); got != string(want) {
		t.Errorf("the report is\n%s\nwant\n%s", got, want)
	}
}

// A task is counted for the chat host of the tool call it came in, whose line
// is the one line of a call that names the host — and for no host anyone knows
// when that line is not among the ones read.
func TestATaskIsCountedForTheHostOfItsCall(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"task_requested","already_open":false,"instructions_version":"v1","request_id":"r1"}`,
		`{"message":"tool_call","tool":"next_task","outcome":"ok","client":"chatgpt","request_id":"r1"}`,
		`{"message":"task_requested","already_open":false,"instructions_version":"v1","request_id":"r2"}`,
	)
	for _, host := range []string{"chatgpt", callNotRead} {
		if counted := c.tasks[group{version: "v1", host: host}]; counted == nil || counted.asked != 1 {
			t.Errorf("tasks asked for by %s = %+v, want 1", host, counted)
		}
	}
}

// A call is known by its span wherever its lines name one, rather than by its
// request's id, which a client may have sent and sent again: two calls that
// came with one id are each counted for their own host.
func TestACallIsKnownByItsSpanBeforeItsRequest(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"task_requested","instructions_version":"v1","request_id":"same","logging.googleapis.com/spanId":"s1"}`,
		`{"message":"tool_call","tool":"next_task","client":"claude","request_id":"same","logging.googleapis.com/spanId":"s1"}`,
		`{"message":"task_requested","instructions_version":"v1","request_id":"same","logging.googleapis.com/spanId":"s2"}`,
		`{"message":"tool_call","tool":"next_task","client":"chatgpt","request_id":"same","logging.googleapis.com/spanId":"s2"}`,
	)
	for _, host := range []string{"claude", "chatgpt"} {
		if counted := c.tasks[group{version: "v1", host: host}]; counted == nil || counted.asked != 1 {
			t.Errorf("tasks asked for by %s = %+v, want 1", host, counted)
		}
	}
}

// The drawings of a group are counted among the tasks whose lines say whether
// they came with one: where only some say, the count stands beside how many
// said, so that it is not read against every task accepted, and where none
// says, no count is made up.
func TestADrawingIsCountedAmongTheLinesThatSaySo(t *testing.T) {
	t.Parallel()

	accepted := func(drawing string) string {
		return `{"message":"task_accepted","topic":"logic.ordering","attempts":1,"seconds_since_request":10,` +
			`"instructions_version":"v1"` + drawing + `}`
	}
	for _, test := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"every line says", []string{accepted(`,"drawing":true`), accepted(`,"drawing":false`)}, "1"},
		{"some lines say", []string{accepted(`,"drawing":true`), accepted(""), accepted(`,"drawing":false`)}, "1 of 2"},
		{"no line says", []string{accepted(""), accepted("")}, notLogged},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rows := tallied(t, test.lines...).acceptedTable().rows
			if len(rows) != 1 {
				t.Fatalf("accepted tasks in %d rows, want 1", len(rows))
			}
			if got := rows[0][3]; got != test.want {
				t.Errorf("with a drawing = %q, want %q", got, test.want)
			}
		})
	}
}

// handedInLine is a hand-in of this whole size, in bytes, with what it retired
// and what was mended in it, its other parts a quarter of the whole each.
func handedInLine(total int, retired, mended string) string {
	return fmt.Sprintf(`{"message":"task_submitted","attempt":1,"outcome":"accepted","failed":[],`+
		`"task_bytes":%d,"self_check_bytes":%d,"solver_bytes":%d,"core_idea_bytes":%d,"total_bytes":%d,`+
		`"retired":[%s],"mended":[%s],"instructions_version":"v1"}`,
		total/4, total/4, total/4, total/8, total, retired, mended)
}

// A hand-in is measured by its parts: the median of each, and the whole's 90th
// percentile, for each group and format.
func TestAHandInIsMeasuredByItsParts(t *testing.T) {
	t.Parallel()

	rows := tallied(t, handedInLine(1000, "", ""), handedInLine(2000, "", ""), handedInLine(4000, "", "")).
		handInsTable().rows
	want := []string{"v1", callNotRead, formatNow, "3", "0", "500", "500", "500", "250", "2000", "4000"}
	if len(rows) != 1 || !slices.Equal(rows[0], want) {
		t.Errorf("hand-ins = %v, want one row %v", rows, want)
	}
}

// A hand-in of the format before — one that brought what the format no longer
// reads — is counted apart from one of the format asked for now, so that the
// size of the one does not hide the size of the other.
func TestAHandInOfTheFormatBeforeIsCountedApart(t *testing.T) {
	t.Parallel()

	rows := tallied(t, handedInLine(2000, "", ""), handedInLine(5000, `"brief","task.design_thought_process"`, "")).
		handInsTable().rows
	if len(rows) != 2 || rows[0][2] != formatBefore || rows[0][9] != "5000" || rows[1][2] != formatNow ||
		rows[1][9] != "2000" {
		t.Errorf("hand-ins = %v, want the format before at 5000 apart from the format now at 2000", rows)
	}
}

// A hand-in whose line was written before hand-ins were measured says nothing
// of its size, and is counted as not logged rather than as nothing at all.
func TestALineWrittenBeforeHandInsWereMeasuredIsNotLogged(t *testing.T) {
	t.Parallel()

	rows := tallied(t,
		`{"message":"task_submitted","attempt":1,"outcome":"accepted","failed":[],"instructions_version":"v1"}`,
	).handInsTable().rows
	if len(rows) != 1 || rows[0][2] != notLogged || rows[0][3] != "1" || rows[0][9] != notLogged {
		t.Errorf("hand-ins = %v, want one counted, its size not logged", rows)
	}
}

// What was mended is counted by its field, the field mended most often first.
func TestWhatWasMendedIsCountedByItsField(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		handedInLine(2000, "", `"task.correct_answer"`),
		handedInLine(2000, "", `"task.correct_answer","self_check.issues"`),
		handedInLine(2000, "", ""),
	)
	want := [][]string{{"v1", "task.correct_answer", "2"}, {"v1", "self_check.issues", "1"}}
	if rows := c.mendsTable().rows; !slices.EqualFunc(rows, want, slices.Equal[[]string]) {
		t.Errorf("mends = %v, want %v", rows, want)
	}
	if row := c.handInsTable().rows[0]; row[4] != "2" {
		t.Errorf("hand-ins mended = %s, want 2 of the 3", row[4])
	}
}

// Tasks written ahead and let go are counted by the version, the reason and
// whether the task had been written and kept or was still being written, a
// row for each. Under each version, in the order it first appears, the row
// that let go the most comes first, whatever the other row of its reason
// holds; rows that let go as many come by their reason, and of one reason the
// tasks still being written before the ones kept.
func TestTasksLetGoAreCountedByWhyAndWhetherTheyWereWritten(t *testing.T) {
	t.Parallel()

	dropped := func(at, version, reason string, written bool, request string) string {
		return fmt.Sprintf(`{"message":"task_dropped","time":%q,"reason":%q,"written":%t,`+
			`"instructions_version":%q,"request_id":%q}`, at, reason, written, version, request)
	}
	rows := tallied(t,
		dropped("2026-10-02T09:00:00Z", "v2", "topic", true, "r1"),
		dropped("2026-10-02T09:01:00Z", "v2", "topic", true, "r2"),
		dropped("2026-10-02T09:02:00Z", "v2", "topic", true, "r3"),
		dropped("2026-10-02T09:03:00Z", "v2", "topic", false, "r4"),
		dropped("2026-10-02T09:04:00Z", "v2", "language", false, "r5"),
		dropped("2026-10-02T09:05:00Z", "v2", "language", false, "r6"),
		dropped("2026-10-02T09:06:00Z", "v2", "skill", true, "r7"),
		dropped("2026-10-02T09:07:00Z", "v2", "skill", false, "r8"),
		dropped("2026-10-01T09:00:00Z", "v1", "skill", true, "r9"),
	).letGoTable().rows
	want := [][]string{
		{"v1", "skill", "kept", "1"},
		{"v2", "topic", "kept", "3"},
		{"v2", "language", "being written", "2"},
		{"v2", "skill", "being written", "1"},
		{"v2", "skill", "kept", "1"},
		{"v2", "topic", "being written", "1"},
	}
	if !slices.EqualFunc(rows, want, slices.Equal[[]string]) {
		t.Errorf("tasks let go = %v, want %v", rows, want)
	}
}

// A request handed back while it is still open is the request already
// counted, not another one.
func TestARequestHandedBackIsNotAskedForAgain(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"task_requested","already_open":false,"instructions_version":"v1","request_id":"r1"}`,
		`{"message":"task_requested","already_open":true,"instructions_version":"v1","request_id":"r2"}`,
	)
	if got := c.tasks[group{version: "v1", host: callNotRead}].asked; got != 1 {
		t.Errorf("requests asked for = %d, want 1", got)
	}
}

// A request runs out of attempts when its last attempt is refused, and at no
// other attempt.
func TestARequestRunsOutOfAttemptsAtItsLastRefusal(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"task_submitted","attempt":2,"outcome":"rejected","primary":"readability","failed":["readability"],"instructions_version":"v1"}`,
		`{"message":"task_submitted","attempt":3,"outcome":"accepted","failed":[],"instructions_version":"v1"}`,
		`{"message":"task_submitted","attempt":3,"outcome":"rejected","primary":"readability","failed":["readability"],"instructions_version":"v1"}`,
	)
	counted := c.tasks[group{version: "v1", host: callNotRead}]
	if counted.handedIn != 3 || counted.refused != 2 || counted.outOfAttempts != 1 {
		t.Errorf("tasks = %+v, want 3 handed in, 2 refused and 1 out of attempts", counted)
	}
}

// A refused attempt is counted by one check, the one its refusal is counted
// by, and fails every check it names: a check that is both is counted once as
// each.
func TestARefusedAttemptIsCountedByOneCheckAndFailsEach(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"task_submitted","attempt":1,"outcome":"rejected","primary":"solver_disagrees","failed":["solver_disagrees","readability"],"instructions_version":"v1"}`,
		`{"message":"task_submitted","attempt":1,"outcome":"accepted","primary":"","failed":[],"instructions_version":"v1"}`,
	)
	for check, want := range map[string]refusals{
		"solver_disagrees": {counted: 1, failed: 1},
		"readability":      {counted: 0, failed: 1},
	} {
		if got := c.refusals[refusedBy{version: "v1", check: check}]; got == nil || *got != want {
			t.Errorf("refusals by %s = %+v, want %+v", check, got, want)
		}
	}
	if len(c.refusals) != 2 {
		t.Errorf("checks that refused = %v, want the two the refused attempt names", c.refusals)
	}
}

// The versions of the instructions come in the order they first appear,
// whatever the order of the lines: Cloud Logging hands the newest first.
func TestVersionsComeInTheOrderTheyFirstAppear(t *testing.T) {
	t.Parallel()

	c := tallied(t,
		`{"message":"task_requested","time":"2026-09-29T09:00:00Z","instructions_version":"aaaa"}`,
		`{"message":"task_requested","time":"2026-09-28T09:00:00Z","instructions_version":"ffff"}`,
		`{"message":"task_requested","time":"2026-09-27T09:00:00Z","instructions_version":"ffff"}`,
		`{"message":"task_requested","time":"2026-09-28T10:00:00Z","instructions_version":"aaaa"}`,
	)
	if want := []string{"ffff", "aaaa"}; !slices.Equal(c.versions, want) {
		t.Errorf("versions = %v, want %v", c.versions, want)
	}
}

// Only the service's own lines are read: a line of the runtime's, and an
// object that names no event, are left out and counted, and a blank line is
// no line at all. The last line is read without a newline after it.
func TestWhatIsNotALineOfTheServiceIsLeftOut(t *testing.T) {
	t.Parallel()

	read, err := readAll(strings.NewReader("panic: runtime error\n\n" +
		`{"severity":"INFO"}` + "\n  \n" +
		`{"message":"limit_hit","limit":"user_rate"}`))
	if err != nil {
		t.Fatalf("readAll() error = %v, want nil", err)
	}
	if len(read.lines) != 1 || read.lines[0].Limit != "user_rate" || read.others != 2 || read.unreadable != 0 {
		t.Errorf("readAll() = %+v, want the one limit and 2 other lines", read)
	}
}

// A line of the service's that does not read as the report expects is owned
// up to rather than left out as another's: a report that silently counted it
// out would read as a log with less in it.
func TestALineOfTheServiceThatDoesNotReadIsOwnedUpTo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := Run(strings.NewReader(`{"message":"task_submitted","failed":"readability"}`+"\n"+`{"severity":"INFO"}`), &out)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if said := out.String(); !strings.Contains(said, "1 lines of the service's did not read as this report expects") ||
		!strings.Contains(said, "1 more lines, not the service's, were left out.") {
		t.Errorf("the report begins %q, want the unreadable line owned up to apart from the other", said[:min(len(said), 200)])
	}
}

// A whole number reads the same whether the service wrote it, as 2, or Cloud
// Logging gave it back, as 2.0.
func TestAWholeNumberReadsAsWrittenOrAsADouble(t *testing.T) {
	t.Parallel()

	for _, written := range []string{"", ".0"} {
		c := tallied(t,
			`{"message":"task_submitted","attempt":3`+written+`,"outcome":"rejected","primary":"readability","failed":["readability"],"instructions_version":"v1"}`,
			`{"message":"task_accepted","attempts":2`+written+`,"seconds_since_request":95`+written+`,"instructions_version":"v1"}`,
			`{"message":"tool_call","tool":"next_task","outcome":"ok","client":"claude","duration_ms":120`+written+`}`,
		)
		counted := c.tasks[group{version: "v1", host: callNotRead}]
		if counted.outOfAttempts != 1 || !slices.Equal(counted.attempts, []int{2}) || !slices.Equal(counted.seconds, []int64{95}) {
			t.Errorf("numbers written as %q: tasks = %+v, want 1 out of attempts, and 2 attempts in 95 s", written, counted)
		}
		if took := c.tools[toolOf{host: "claude", tool: "next_task"}].milliseconds; !slices.Equal(took, []int64{120}) {
			t.Errorf("numbers written as %q: a call took %v ms, want 120", written, took)
		}
	}
}

// A report of no lines says so, and says of every table that nothing is in it.
func TestAReportOfNoLinesSaysSo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := Run(strings.NewReader(""), &out); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if got := out.String(); !strings.Contains(got, "No lines of the service's were read.") ||
		strings.Count(got, "None in these lines.") != 21 || strings.Contains(got, "|") {
		t.Errorf("the report of no lines is\n%s\nwant it to say there are none, and no table", got)
	}
}

// A whole number is written as English counts with it, the teens with th
// whatever their last digit.
func TestAnOrdinalTakesItsEnglishEnding(t *testing.T) {
	t.Parallel()

	for n, want := range map[int]string{
		1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 10: "10th", 11: "11th", 12: "12th", 13: "13th",
		21: "21st", 22: "22nd", 23: "23rd", 50: "50th", 101: "101st", 111: "111th", 112: "112th", 200: "200th",
	} {
		if got := ordinal(n); got != want {
			t.Errorf("ordinal(%d) = %q, want %q", n, got, want)
		}
	}
}

// The nearest-rank percentile of a few values, where it can be worked out by
// hand.
func TestAPercentileIsTheNearestRank(t *testing.T) {
	t.Parallel()

	tens := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, tc := range []struct {
		sorted []int
		rank   int
		want   int
	}{
		{tens, 50, 5},
		{tens, 90, 9},
		{tens, 95, 10},
		{tens, 100, 10},
		{tens, 1, 1},
		{[]int{7}, 50, 7},
		{[]int{1, 3}, 50, 1},
	} {
		if got := atRank(tc.sorted, tc.rank); got != tc.want {
			t.Errorf("atRank(%v, %d) = %d, want %d", tc.sorted, tc.rank, got, tc.want)
		}
	}
}

// Whatever the values, a percentile is one of them, it never falls as the
// rank rises, and the hundredth is the largest.
func TestAPercentileHoldsItsProperties(t *testing.T) {
	t.Parallel()

	sorted := gen.SliceOf(gen.Int64Range(0, 1<<20)).
		SuchThat(func(values []int64) bool { return len(values) > 0 }).
		Map(func(values []int64) []int64 { return slices.Sorted(slices.Values(values)) })
	properties := gopter.NewProperties(nil)
	properties.Property("a percentile is one of the values", prop.ForAll(
		func(sorted []int64, rank int) bool { return slices.Contains(sorted, atRank(sorted, rank)) },
		sorted, gen.IntRange(1, 100),
	))
	properties.Property("a higher rank is never a smaller value", prop.ForAll(
		func(sorted []int64, ranks []int) bool {
			lower, higher := min(ranks[0], ranks[1]), max(ranks[0], ranks[1])
			return atRank(sorted, lower) <= atRank(sorted, higher)
		},
		sorted, gen.SliceOfN(2, gen.IntRange(1, 100)),
	))
	properties.Property("the hundredth is the largest", prop.ForAll(
		func(sorted []int64) bool { return atRank(sorted, 100) == slices.Max(sorted) },
		sorted,
	))
	properties.TestingRun(t)
}

// tallied is what the lines given add up to.
func tallied(t *testing.T, lines ...string) *counts {
	t.Helper()

	read, err := readAll(strings.NewReader(strings.Join(lines, "\n")))
	if err != nil || read.others != 0 || read.unreadable != 0 {
		t.Fatalf("readAll() = %+v, error %v; want every line read", read, err)
	}
	return tally(read)
}

// A log that breaks off while it is read, and a report nobody can take, are
// failures that say which of the two it was, and nothing is written of a log
// read only in part.
func TestAReadOrAWriteThatFailsSaysWhich(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	broken := io.MultiReader(strings.NewReader(`{"message":"limit_hit","limit":"user_rate"}`+"\n"), failing{})
	if err := Run(broken, &out); err == nil || !strings.Contains(err.Error(), "report: read the log") || out.Len() != 0 {
		t.Errorf("Run() on a log that breaks off = %v, wrote %d bytes; want the read named and nothing written", err, out.Len())
	}
	if err := Run(strings.NewReader(""), failing{}); err == nil || !strings.Contains(err.Error(), "report: write the report") {
		t.Errorf("Run() to a report nobody can take = %v, want the write named", err)
	}
}

// failing is a reader and a writer that fail at once.
type failing struct{}

func (failing) Read([]byte) (int, error)  { return 0, errors.New("the log broke off") }
func (failing) Write([]byte) (int, error) { return 0, errors.New("nobody takes the report") }

// A number that is no number at all makes its line one of the service's the
// report could not read.
func TestANumberThatIsNoNumberMakesItsLineUnreadable(t *testing.T) {
	t.Parallel()

	read, err := readAll(strings.NewReader(`{"message":"task_submitted","attempt":"two"}`))
	if err != nil || len(read.lines) != 0 || read.unreadable != 1 {
		t.Errorf("readAll() = %+v, %v; want the line counted as unreadable", read, err)
	}
}

// A refused attempt whose refusal names no check of its own is counted by
// none, and still fails the checks it names.
func TestARefusalThatNamesNoCheckIsCountedByNone(t *testing.T) {
	t.Parallel()

	c := tallied(t, `{"message":"task_submitted","attempt":1,"outcome":"rejected","primary":"","failed":["readability"],"instructions_version":"v1"}`)
	if got := c.refusals[refusedBy{version: "v1", check: "readability"}]; got == nil || *got != (refusals{counted: 0, failed: 1}) || len(c.refusals) != 1 {
		t.Errorf("refusals = %v, want readability failed once and no check without a name", c.refusals)
	}
}

// Lines with no time on them are counted, and the report says it cannot tell
// when they were written.
func TestLinesWithNoTimeAreSaidToHaveNone(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	if err := Run(strings.NewReader(`{"message":"limit_hit","limit":"user_rate"}`), &out); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "1 lines of the service's, with no time on them.") {
		t.Errorf("the report begins %q, want it to say the lines have no time", out.String()[:min(out.Len(), 120)])
	}
}

// A text in a cell of a table stays in its cell and its row: a bar is written
// as a bar and a line break as a space, so neither starts a cell or a row of
// its own, and a cell with nothing in it says so.
func TestATextStaysInItsCell(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	writeRow(&b, []string{"a|b", "first\nsecond\r\nthird\rfourth", ""})
	if got, want := b.String(), "| a\\|b | first second third fourth | (none) |\n"; got != want {
		t.Errorf("writeRow() = %q, want %q", got, want)
	}
}
