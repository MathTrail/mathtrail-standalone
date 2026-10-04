package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/report"
)

// lessonEvents are the lines a whole lesson leaves. The fields each may carry
// are the report's to decide, in the one table every reader of the log holds
// lines to: here the lesson, and in the report the log of a deployment.
var lessonEvents = []string{
	// The service as it is built.
	"content loaded", "seal keys loaded", "telemetry built", "solver sandbox built", "profile store",
	"google sign-in", "limits set",
	// Every request, whatever it asked for.
	"http_request",
	// The sign-in.
	"auth_register", "cimd_fetch", "auth_reject", "auth_authorize", "auth_consent", "auth_callback", "auth_token",
	"auth_bearer", "auth_refresh", "auth_revoke",
	// The lesson.
	"tool_call", "drive_call", "task_requested", "solver_run", "task_submitted", "task_accepted", "answer_recorded",
	"limit_hit",
}

// A whole lesson, read line by line. A family signs in through the consent
// screen and Google, makes the child's profile, has the race refused and
// accepted, answers it wrongly, reads the progress, reaches the day's limit
// and ends the access, and every line the service writes on the way is held to
// two rules. It carries nothing a person, a model or a sign-in wrote: no
// pseudonym, interest or note, no word of the task and no letter of an answer,
// no token, code or cookie, no email address and no address a request came
// from. And it is a line of an event the lesson is known to leave, with only
// the fields decided for that event.
func TestTheLogOfAWholeLesson(t *testing.T) {
	t.Parallel()

	f := newFamily(t)
	f.signIn()
	f.presentAStrangersToken()
	f.holdALesson()
	f.renewAndEnd()
	lines := f.lines()

	t.Run("nothing personal", func(t *testing.T) { wantNothingPersonal(t, lines, f.secrets, f.words) })
	t.Run("only the fields of its event", func(t *testing.T) { wantOnlyTheFieldsOfTheirEvents(t, lines) })
	t.Run("adds up in the report", func(t *testing.T) { wantTheLessonInTheReport(t, lines) })
}

// wantTheLessonInTheReport holds the report to adding the lesson's own lines up
// into what the lesson was: one task asked for, however often it was asked,
// handed in twice, refused once for its structure and accepted at the second
// attempt, the day's limit reached, and each tool called as the host called
// it. So the report is held to the events and the fields the service writes,
// not to ones it supposes.
func wantTheLessonInTheReport(t *testing.T, lines []observer.LoggedEntry) {
	t.Helper()

	var written bytes.Buffer
	version := ""
	for i := range lines {
		fields := lines[i].ContextMap()
		if lines[i].Message == "content loaded" {
			version, _ = fields["instructions_version"].(string)
		}
		fields["message"], fields["time"] = lines[i].Message, lines[i].Time.Format(time.RFC3339Nano)
		// A duration is written in seconds, as the service's encoder writes it.
		for key, value := range fields {
			if took, isDuration := value.(time.Duration); isDuration {
				fields[key] = took.Seconds()
			}
		}
		encoded, err := json.Marshal(fields)
		if err != nil {
			t.Fatalf("a %s line does not encode: %v", lines[i].Message, err)
		}
		written.Write(append(encoded, '\n'))
	}
	var added strings.Builder
	if err := report.Run(&written, &added); err != nil {
		t.Fatalf("report.Run() error = %v, want nil", err)
	}

	for _, row := range []string{
		"| " + version + " | claude | 1 | 2 | 1 | 0 |",
		"| " + version + " | claude | 1 | 0 | 2.0 |",
		"| " + version + " | bad_structure | 1 | 1 |",
		"| daily_tasks | 1 |",
		"| claude | next_task | 3 | 2 | 1 | 0 | 0 |",
		"| claude | get_package | 1 | 1 | 0 | 0 | 0 |",
		"| claude | read_task | 2 | 2 | 0 | 0 | 0 |",
		"| claude | submit_task | 2 | 1 | 1 | 0 | 0 |",
		"| claude | submit_answer | 2 | 2 | 0 | 0 | 0 |",
		// The race was asked for by the model, so its answer is not weighed
		// against the chance: the task kept who chose it, and the line says.
		"are the answers to tasks the model chose (1).",
	} {
		if !strings.Contains(added.String(), row) {
			t.Errorf("the report of the lesson has no row starting %q:\n%s", row, added.String())
		}
	}
	// The report holds the lines to the rules this test holds them to, and
	// finds them kept: its own section, up to the next, says so.
	_, rules, _ := strings.Cut(added.String(), "## The rules of the log")
	rules, _, _ = strings.Cut(rules, "\n## ")
	if !strings.Contains(rules, "None in these lines.") {
		t.Errorf("the report finds the lesson's lines breaking the rules of the log:%s", rules)
	}
}

// wantNothingPersonal holds every text of every line to carrying nothing of a
// person's, a model's or a sign-in's: none of the secrets anywhere in it, and
// none of the words as the whole of it.
func wantNothingPersonal(t *testing.T, lines []observer.LoggedEntry, secrets, words []string) {
	t.Helper()

	for i := range lines {
		for _, text := range textsOf(&lines[i]) {
			for _, found := range personal(text, secrets, words) {
				t.Errorf("a %s line carries %s in %q", lines[i].Message, found, text)
			}
		}
	}
}

// wantOnlyTheFieldsOfTheirEvents holds every line to an event the lesson is
// known to leave and to that event's fields, and every such event to a line of
// the lesson, so that none is listed for nothing.
func wantOnlyTheFieldsOfTheirEvents(t *testing.T, lines []observer.LoggedEntry) {
	t.Helper()

	seen := map[string]bool{}
	for i := range lines {
		seen[lines[i].Message] = true
		if !slices.Contains(lessonEvents, lines[i].Message) || !report.Known(lines[i].Message) {
			t.Errorf("a line of %q, which the lesson is not known to leave: %v", lines[i].Message, lines[i].ContextMap())
		}
		for _, field := range strayFields(&lines[i]) {
			t.Errorf("a %s line carries %s, a field nobody decided it may", lines[i].Message, field)
		}
	}
	for _, event := range lessonEvents {
		if !seen[event] {
			t.Errorf("the lesson left no %s line, want the event it is listed for", event)
		}
	}
}

// personal is what of a person's, a model's or a sign-in's a text of a line
// carries: any of the secrets, one of the words standing as the whole text,
// or an email address.
func personal(text string, secrets, words []string) []string {
	var found []string
	for _, secret := range secrets {
		if strings.Contains(text, secret) {
			found = append(found, fmt.Sprintf("%q", secret))
		}
	}
	if slices.Contains(words, text) {
		found = append(found, fmt.Sprintf("%q as a value of its own", text))
	}
	if address := report.EmailIn(text); address != "" {
		found = append(found, "the email address "+address)
	}
	return found
}

// strayFields are the fields of a line its event is not decided to carry.
// Every field of a line of an event nobody decided on is stray.
func strayFields(line *observer.LoggedEntry) []string {
	var stray []string
	for field := range line.ContextMap() {
		if !report.Decided(line.Message, field) {
			stray = append(stray, field)
		}
	}
	return stray
}

// textsOf is every text a line carries: its message, and each of its values as
// it is written, a list and an error included — and each of those as it reads
// once its percent-encoding is undone, which is how an address, a query and a
// form carry an email address or a token.
func textsOf(line *observer.LoggedEntry) []string {
	texts := []string{line.Message}
	for _, value := range line.ContextMap() {
		texts = append(texts, textsIn(value)...)
	}
	for _, text := range texts {
		if decoded, err := url.QueryUnescape(text); err == nil && decoded != text {
			texts = append(texts, decoded)
		}
	}
	return texts
}

// textsIn is the texts a value of a line holds.
func textsIn(value any) []string {
	switch held := value.(type) {
	case string:
		return []string{held}
	case []any:
		var texts []string
		for _, item := range held {
			texts = append(texts, textsIn(item)...)
		}
		return texts
	}
	return []string{fmt.Sprint(value)}
}
