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

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/geoip/geoiptest"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
	"github.com/MathTrail/mathtrail-standalone/internal/report"
)

// lessonEvents are the lines a whole lesson leaves. The fields each may carry
// are the report's to decide, in the one table every reader of the log holds
// lines to: here the lesson, and in the report the log of a deployment.
var lessonEvents = []string{
	// The service as it is built.
	"content loaded", "seal keys loaded", "telemetry built", "solver sandbox built", "profile store",
	"google sign-in", "reviewer sign-in", "learner key", "country database", "limits set",
	// Every request, whatever it asked for.
	"http_request",
	// The sign-in.
	"auth_register", "cimd_fetch", "auth_reject", "auth_authorize", "auth_consent", "auth_callback", "auth_token",
	"auth_bearer", "auth_refresh", "auth_revoke",
	// The lesson.
	"tool_call", "drive_call", "task_requested", "solver_run", "task_submitted", "task_accepted", "answer_recorded",
	"write_after_answer", "limit_hit",
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
	t.Run("only the fields of its event", func(t *testing.T) { wantOnlyTheFieldsOfTheirEvents(t, lines, lessonEvents) })
	t.Run("adds up in the report", func(t *testing.T) { wantTheLessonInTheReport(t, lines) })
	t.Run("counts the child", func(t *testing.T) { wantTheChildCounted(t, lines) })
}

// The same lesson, held by a directory's reviewer who signed in with the
// reviewers' password as the demo account: every line holds to the same two
// rules — the password and the demo account's grant among what none may
// carry — and the lines a child is counted from name no child, since the demo
// account's is none. The host ending its access ends nothing at Google.
func TestTheLogOfAReviewersLesson(t *testing.T) {
	t.Parallel()

	f := newFamilyWith(t, func(cfg *config.Config) {
		cfg.ReviewerPassword = reviewerPassword
		cfg.ReviewerGrant = `{"subject":"` + googletest.Subject + `","refresh_token":"` + googletest.RefreshToken + `"}`
	})
	f.signInAsReviewer(reviewerPassword)
	f.presentAStrangersToken()
	f.holdALesson()
	f.renewAndEnd()
	lines := f.lines()

	t.Run("nothing personal", func(t *testing.T) { wantNothingPersonal(t, lines, f.secrets, f.words) })
	t.Run("only the fields of its event", func(t *testing.T) {
		// A reviewer's sign-in never comes back from Google.
		wantOnlyTheFieldsOfTheirEvents(t, lines, slices.DeleteFunc(slices.Clone(lessonEvents), func(event string) bool {
			return event == "auth_callback"
		}))
	})
	t.Run("counts no child", func(t *testing.T) { wantNoChildCounted(t, lines) })
	if revoked := f.google.Revocations(); len(revoked) != 0 {
		t.Errorf("Google was asked to end a grant %d times, want the demo account's left alone", len(revoked))
	}
}

// reviewerPassword is the reviewers' password of the reviewer's lesson.
const reviewerPassword = "Xk3-vQ9_tLm2Wp7Rz4Yb8Nc1Jd6Hf5Gs"

// wantNoChildCounted holds the lines a child is counted from to naming no
// child: they are written, as a family's are, and carry no name to count by.
func wantNoChildCounted(t *testing.T, lines []observer.LoggedEntry) {
	t.Helper()

	counted := 0
	for i := range lines {
		switch lines[i].Message {
		case "task_accepted", "answer_recorded", "topic_mastered":
			counted++
			if learner, named := lines[i].ContextMap()["learner"]; named {
				t.Errorf("a %s line names the demo account's child %v, want no name", lines[i].Message, learner)
			}
		}
	}
	if counted < 2 {
		t.Errorf("the lesson left %d lines a child is counted from, want its task and its answer", counted)
	}
}

// wantTheChildCounted holds the line about the task handed out to counting
// the child where the family is: the chat host by the family of hosts, the
// country the parent's browser came back from Google in, and nothing the
// parent did not say — the profile names no country. The answer's line counts
// the topics mastered, none yet.
func wantTheChildCounted(t *testing.T, lines []observer.LoggedEntry) {
	t.Helper()

	var accepted, recorded []map[string]any
	for i := range lines {
		switch lines[i].Message {
		case "task_accepted":
			accepted = append(accepted, lines[i].ContextMap())
		case "answer_recorded":
			recorded = append(recorded, lines[i].ContextMap())
		}
	}
	if len(accepted) != 1 || len(recorded) != 1 {
		t.Fatalf("the lesson left %d task_accepted and %d answer_recorded lines, want one of each", len(accepted), len(recorded))
	}
	for field, want := range map[string]any{
		"host": "claude", "language": "en", "grade": int64(2), "signin_country": geoiptest.Country,
		"country": "unknown", "region": "unknown",
	} {
		if got := accepted[0][field]; got != want {
			t.Errorf("task_accepted %s = %v, want %v", field, got, want)
		}
	}
	if got, want := recorded[0]["learner"], accepted[0]["learner"]; got != want {
		t.Errorf("answer_recorded counts the child as %v and task_accepted as %v, want one name", got, want)
	}
	if got := recorded[0]["topics_mastered"]; got != int64(0) {
		t.Errorf("answer_recorded topics_mastered = %v, want 0", got)
	}
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
		"| " + version + " | claude | 1 | 0 | 0 | 2.0 |",
		"| " + version + " | logic.ordering | 1 | 0 |",
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
	wantTheLessonInItsSections(t, added.String(), version)
}

// wantTheLessonInItsSections holds two sections of the report of the lesson,
// each up to the next. The card's question after the hand-in brought the
// task: the lines name the task's request and what each call answered, as the
// report reads them, so the section of tasks shown on the card counts it. And
// the report holds the lines to the rules this test holds them to, and finds
// them kept.
func wantTheLessonInItsSections(t *testing.T, added, version string) {
	t.Helper()

	if shown := sectionOf(added, "## Tasks accepted, shown on the card"); !strings.Contains(shown,
		"| "+version+" | claude | 1 |") {
		t.Errorf("the report of the lesson shows no task come to the card that waited for it:%s", shown)
	}
	if rules := sectionOf(added, "## The rules of the log"); !strings.Contains(rules, "None in these lines.") {
		t.Errorf("the report finds the lesson's lines breaking the rules of the log:%s", rules)
	}
}

// sectionOf is the section of the report added under the heading, up to the
// next.
func sectionOf(added, heading string) string {
	_, section, _ := strings.Cut(added, heading)
	section, _, _ = strings.Cut(section, "\n## ")
	return section
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

// wantOnlyTheFieldsOfTheirEvents holds every line to an event of those the
// lesson is known to leave and to that event's fields, and every such event to
// a line of the lesson, so that none is listed for nothing.
func wantOnlyTheFieldsOfTheirEvents(t *testing.T, lines []observer.LoggedEntry, events []string) {
	t.Helper()

	seen := map[string]bool{}
	for i := range lines {
		seen[lines[i].Message] = true
		if !slices.Contains(events, lines[i].Message) || !report.Known(lines[i].Message) {
			t.Errorf("a line of %q, which the lesson is not known to leave: %v", lines[i].Message, lines[i].ContextMap())
		}
		for _, field := range strayFields(&lines[i]) {
			t.Errorf("a %s line carries %s, a field nobody decided it may", lines[i].Message, field)
		}
		for field, value := range lines[i].ContextMap() {
			if !report.Fits(lines[i].Message, field, value) {
				t.Errorf("a %s line carries %v in %s, a value of no form the field may hold", lines[i].Message, value, field)
			}
		}
	}
	for _, event := range events {
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
