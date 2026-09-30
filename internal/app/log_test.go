package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/report"
)

// events are the lines a whole lesson leaves, each with the fields it may
// carry. A field that is not here is one nobody decided a line may carry, and
// that is how something personal would get in: the rule is not to clean a
// value before it is written, but never to write it.
var events = map[string][]string{
	// The service as it is built.
	"content loaded":       {"topics", "traps", "skills", "reference_tasks", "instructions_version"},
	"seal keys loaded":     {"key_id", "previous_key"},
	"telemetry built":      {"export", "endpoint", "sample_ratio"},
	"solver sandbox built": {"steps", "timeout", "concurrency", "wait", "gomaxprocs", "memory_limit"},
	"profile store":        {"in_drive"},
	"google sign-in":       {"configured"},
	"limits set": {
		"user_per_min", "ip_per_min", "instance_per_min", "renewal_per_min", "daily_tasks", "daily_failed", "trap_repeats",
	},

	// Every request, whatever it asked for.
	"http_request": {"status", "method", "route", "duration", "body_size", "query", "query_others", "errors", "panic", "stack"},

	// The sign-in.
	"auth_register":  {"registration", "redirect_host", "outcome", "reason", "error"},
	"cimd_fetch":     {"host", "cached", "duration_ms", "outcome"},
	"auth_reject":    {"step", "reason", "error"},
	"auth_authorize": {"registration", "redirect_host", "resource", "scope", "outcome", "reason"},
	"auth_consent":   {"registration", "redirect_host", "outcome"},
	"auth_callback":  {"registration", "redirect_host", "outcome", "reason", "user", "error"},
	"auth_token":     {"registration", "resource", "outcome", "reason", "user", "error"},
	"auth_bearer":    {"outcome", "reason"},
	"auth_refresh":   {"registration", "resource", "outcome", "reason", "user", "error"},
	"auth_revoke":    {"registration", "token", "outcome", "reason", "user", "error"},

	// The lesson.
	"tool_call": {
		"tool", "outcome", "status", "error", "duration_ms", "instructions_version", "protocol_version", "client",
		"user", "panic", "stack",
	},
	"drive_call":     {"op", "duration_ms", "retries", "outcome", "user"},
	"task_requested": {"topic", "level", "difficulty", "goal", "tutor_mode", "already_open", "instructions_version", "user"},
	"solver_run":     {"status", "steps", "duration_ms", "instructions_version", "user"},
	"task_submitted": {
		"attempt", "outcome", "primary", "failed", "minor_issues", "duration_ms", "solver_steps", "solver_ms",
		"instructions_version", "user",
	},
	"task_accepted": {"topic", "level", "difficulty", "attempts", "seconds_since_request", "instructions_version", "user"},
	"answer_recorded": {
		"topic", "level", "difficulty", "correct", "trap", "hint_used", "confused", "pace", "instructions_version", "user",
	},
	"limit_hit": {"limit", "count", "user"},
}

// ofARequest are the fields any line written inside a request may carry: the
// request it belongs to, and the trace a reader opens it in.
var ofARequest = []string{
	"request_id", "logging.googleapis.com/trace", "logging.googleapis.com/spanId", "logging.googleapis.com/trace_sampled",
}

// emailAddress is anything shaped like an email address, whoever wrote it.
var emailAddress = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}`)

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
		"| claude | submit_task | 2 | 1 | 1 | 0 | 0 |",
		"| claude | submit_answer | 2 | 2 | 0 | 0 | 0 |",
	} {
		if !strings.Contains(added.String(), row) {
			t.Errorf("the report of the lesson has no row starting %q:\n%s", row, added.String())
		}
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
		if _, known := events[lines[i].Message]; !known {
			t.Errorf("a line of %q, which the lesson is not known to leave: %v", lines[i].Message, lines[i].ContextMap())
		}
		for _, field := range strayFields(&lines[i]) {
			t.Errorf("a %s line carries %s, a field nobody decided it may", lines[i].Message, field)
		}
	}
	for event := range events {
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
	if address := emailAddress.FindString(text); address != "" {
		found = append(found, "the email address "+address)
	}
	return found
}

// strayFields are the fields of a line its event is not decided to carry.
// Every field of a line of an event nobody decided on is stray.
func strayFields(line *observer.LoggedEntry) []string {
	var stray []string
	for field := range line.ContextMap() {
		if !slices.Contains(events[line.Message], field) && !slices.Contains(ofARequest, field) {
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
