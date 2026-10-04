package report

import (
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// events are the lines the service writes, each with the fields its event may
// carry beside those of a request and of every line. A field that is not here
// is one nobody decided a line may carry, and that is how something personal
// would get in: the rule is not to clean a value before it is written, but
// never to write it.
var events = map[string][]string{
	// The process starting and stopping. A start or a stop that failed is the
	// same event, with what failed as its error.
	"startup":            {"version", "commit", "date", "port", "public_url", "deployed", "dev_auth", "error"},
	"shutdown":           {"started", "error"},
	"listening":          {"addr"},
	"shutdown requested": nil,
	"stopped":            {"cut_short"},
	"close failed":       {"error"},

	// The service as it is built.
	"content loaded":       {"topics", "traps", "skills", "reference_tasks", "instructions_version"},
	"seal keys loaded":     {"key_id", "previous_key"},
	"telemetry built":      {"export", "endpoint", "sample_ratio"},
	"solver sandbox built": {"steps", "timeout", "concurrency", "wait", "gomaxprocs", "memory_limit"},
	"profile store":        {"in_drive"},
	"google sign-in":       {"configured"},
	"learner key":          {"configured"},
	"country database":     {"configured", "type", "built"},
	"limits set": {
		"user_per_min", "ip_per_min", "instance_per_min", "renewal_per_min", "daily_tasks", "daily_failed", "trap_repeats",
	},

	// Traces and measurements that could not be set up or delivered.
	"telemetry export unavailable":  {"error"},
	"telemetry resource incomplete": {"error"},
	"telemetry_failed":              {"error"},
	"telemetry_flush_failed":        {"error"},
	"telemetry_flush_panicked":      {"panic"},

	// Every request, whatever it asked for, and a panic above its own line.
	"http_request": {"status", "method", "route", "duration", "body_size", "query", "query_others", "errors", "panic", "stack"},
	"panic":        {"panic", "stack", "method", "route"},

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
	"mcp_panic":      {"panic", "stack", "method", "user"},
	"task_requested": {"topic", "level", "difficulty", "goal", "tutor_mode", "already_open", "instructions_version", "user"},
	"task_skipped":   {"topic", "level", "difficulty", "instructions_version", "user"},
	"solver_run":     {"status", "steps", "duration_ms", "instructions_version", "user"},
	"task_submitted": {
		"attempt", "outcome", "primary", "failed", "minor_issues", "duration_ms", "solver_steps", "solver_ms",
		"instructions_version", "user",
	},
	// A task handed out, an answer and a topic mastered are what the children
	// are counted from, and each names its child by the name it is counted
	// under that month, which leads back to no child.
	"task_accepted": {
		"topic", "level", "difficulty", "attempts", "seconds_since_request", "instructions_version", "user",
		"learner", "host", "language", "grade", "cohort", "country", "region", "signin_country",
	},
	// The chance beside the user is the child's rating, answer by answer, as
	// far as two places of a chance tell it. The topic, the level, the
	// difficulty and whether each answer was right let the rating be rebuilt
	// too, but only from the child's first answer on, with the start guessed
	// from the first tasks, and only while the log keeps every answer since;
	// the chance tells it outright. It is the one number of a line the rating
	// sets.
	"answer_recorded": {
		"topic", "level", "difficulty", "correct", "trap", "hint_used", "confused", "pace",
		"chance", "tutor_mode", "trial", "answers_bucket", "instructions_version", "user",
		"learner", "grade", "cohort", "topics_mastered",
	},
	"topic_mastered": {"learner", "topic", "grade", "instructions_version", "user"},
	"limit_hit":      {"limit", "count", "user"},

	// The parent's Drive.
	"drive_call":         {"op", "duration_ms", "retries", "outcome", "user"},
	"drive_conflict":     {"read_revision", "found_revision", "reason", "user"},
	"drive_stale_read":   {"written_revision", "read_revision", "outcome", "user"},
	"drive_recovered":    {"tried", "restored_revision", "outcome", "user"},
	"drive_started_over": {"set_aside", "set_aside_from_bin", "user"},
}

// The fields a line written inside a request carries to tie it to the trace a
// reader opens it in, spelled the way a log collector expects them.
const (
	traceField   = "logging.googleapis.com/trace"
	spanField    = "logging.googleapis.com/spanId"
	sampledField = "logging.googleapis.com/trace_sampled"
)

// instanceField is the instance that wrote a line, which the service never
// writes, so no event is decided to carry it: the platform names it in the
// entry it keeps the line in, and the reader of the platform's log puts it
// beside the line's own fields.
const instanceField = "instance"

// ofARequest are the fields any line written inside a request may carry: the
// request it belongs to, and the trace a reader opens it in.
var ofARequest = []string{"request_id", traceField, spanField, sampledField}

// ofEveryLine are the fields the logger writes on every line: the severity, the
// event, the time and the place in the code that wrote it.
var ofEveryLine = []string{"severity", "message", "time", "caller"}

// emailAddress is anything shaped like an email address, whoever wrote it.
var emailAddress = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}`)

// Known says whether the service writes lines of the event.
func Known(event string) bool {
	_, known := events[event]
	return known
}

// Decided says whether a line of the event may carry the field: one decided
// for the event, one of any line written inside a request, or one every line
// has. No field is decided for an event nobody decided on.
func Decided(event, field string) bool {
	own, known := events[event]
	return known && (slices.Contains(own, field) || slices.Contains(ofARequest, field) || slices.Contains(ofEveryLine, field))
}

// countedFrom are the events the children are counted from. They are the lines
// kept the longest, so what each of their fields may hold is decided too.
var countedFrom = []string{"task_accepted", "answer_recorded", "topic_mastered"}

// shapes are the forms the text fields of the events the children are counted
// from hold: a word of a closed list or a code of a fixed form, so that nothing
// a person typed — a name, the words of a profile, an address — fits in one,
// and an identifier logged by mistake, a profile's own among them, does not
// fit the name a child is counted under. A form says what a code looks like
// rather than which codes there are, so that the log of a build with a longer
// list of countries reads as well as this one's.
var shapes = map[string]*regexp.Regexp{
	"learner":        regexp.MustCompile(`^[A-Za-z0-9_-]{16}$`),
	"cohort":         regexp.MustCompile(`^\d{4}-(0[1-9]|1[012])$`),
	"language":       regexp.MustCompile(`^([a-z]{2,3}|other)$`),
	"host":           regexp.MustCompile(`^(claude|chatgpt|inspector|load|other|unknown)$`),
	"country":        regexp.MustCompile(`^([A-Z]{2}|unknown|other)$`),
	"signin_country": regexp.MustCompile(`^([A-Z]{2}|unknown|other)$`),
	"region":         regexp.MustCompile(`^([A-Z]{2}-[A-Z0-9]{1,3}|unknown|other)$`),
}

// The numbers those events carry, each a whole number in its range: the grade
// a parent can give, and how many topics a child has mastered, which no
// catalog holds a thousand of.
const (
	lowestGrade, highestGrade = 1, 6
	mostTopicsMastered        = 1000
)

// Fits says whether a value is one the field of the event may hold. Only the
// fields of the events the children are counted from are held to a form; any
// other field may hold whatever its event is decided to carry.
func Fits(event, field string, value any) bool {
	if !slices.Contains(countedFrom, event) {
		return true
	}
	switch field {
	case "grade":
		return wholeWithin(value, lowestGrade, highestGrade)
	case "topics_mastered":
		return wholeWithin(value, 0, mostTopicsMastered)
	}
	shape, formed := shapes[field]
	if !formed {
		return true
	}
	text, isText := value.(string)
	return isText && shape.MatchString(text)
}

// wholeWithin says whether a value is a whole number from low to high, as a
// line read back from its text holds one or as a logger holds it.
func wholeWithin(value any, low, high int64) bool {
	var whole int64
	switch number := value.(type) {
	case float64:
		if number != math.Trunc(number) {
			return false
		}
		whole = int64(number)
	case int64:
		whole = number
	case int:
		whole = int64(number)
	default:
		return false
	}
	return whole >= low && whole <= high
}

// unescapes is how many times a text's percent-escapes are undone in search of
// an address: once for a text escaped once, as an address, a query and a form
// carry one, and again for one escaped over again.
const unescapes = 3

// EmailIn is the email address a text carries, as it is written or once its
// percent-escapes are undone; empty when it carries none. An address has an
// at sign, written or escaped, so a text with neither an at sign nor a percent
// sign carries none, which most texts of a line are.
func EmailIn(text string) string {
	if !strings.ContainsAny(text, "@%") {
		return ""
	}
	for range unescapes {
		if found := emailAddress.FindString(text); found != "" {
			return found
		}
		undone := unescape(text)
		if undone == text {
			return ""
		}
		text = undone
	}
	return emailAddress.FindString(text)
}

// unescape undoes every percent-escape of a text that reads as one, each on its
// own, and leaves the rest as it is: a stray percent sign beside an escape
// must not keep the escape from being read.
func unescape(text string) string {
	if !strings.Contains(text, "%") {
		return text
	}
	var undone strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] == '%' && i+2 < len(text) {
			if b, err := strconv.ParseUint(text[i+1:i+3], 16, 8); err == nil {
				undone.WriteByte(byte(b))
				i += 2
				continue
			}
		}
		undone.WriteByte(text[i])
	}
	return undone.String()
}
