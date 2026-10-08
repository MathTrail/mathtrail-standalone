package report

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// A line is held to an event the service is decided to write, to the fields
// decided for it, and to carrying nothing shaped like an email address, in a
// value, in a list, in an object or written with its percent-encoding; a
// number or a truth value holds no text.
func TestALineIsHeldToTheRulesOfTheLog(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		line string
		want []breach
	}{
		{"a line as the service writes it",
			`{"severity":"INFO","time":"2026-09-30T08:00:00Z","caller":"x.go:1","message":"tool_call","tool":"next_task",` +
				`"request_id":"r1","logging.googleapis.com/trace":"t","user":"u1","instance":"i-1","duration_ms":12}`,
			nil},
		{"an event nobody decided on", `{"message":"made_up","anything":"at all"}`,
			[]breach{{event: unknownEvent, rule: ruleUnknownEvent}}},
		{"an event whose words name a child", `{"message":"hello Masha from 10.1.2.3"}`,
			[]breach{{event: unknownEvent, rule: ruleUnknownEvent}}},
		{"a field named otherwise than the service names one", `{"message":"tool_call","Masha":"1"}`,
			[]breach{{event: "tool_call", field: withheld, rule: ruleUndecided}}},
		{"a field its event is not decided to carry", `{"message":"tool_call","pseudonym":"Comet"}`,
			[]breach{{event: "tool_call", field: "pseudonym", rule: ruleUndecided}}},
		{"an email address in a value", `{"message":"cimd_fetch","host":"alice@school.example"}`,
			[]breach{{event: "cimd_fetch", field: "host", rule: ruleEmail}}},
		{"one percent-encoded", `{"message":"http_request","query":"login_hint=alice%40school.example"}`,
			[]breach{{event: "http_request", field: "query", rule: ruleEmail}}},
		{"one escaped beside a stray percent sign", `{"message":"http_request","query":"q=100%&login_hint=kid%40school.example"}`,
			[]breach{{event: "http_request", field: "query", rule: ruleEmail}}},
		{"one escaped twice", `{"message":"http_request","query":"login_hint=kid%2540school.example"}`,
			[]breach{{event: "http_request", field: "query", rule: ruleEmail}}},
		{"one escaped three times", `{"message":"http_request","query":"login_hint=kid%252540school.example"}`,
			[]breach{{event: "http_request", field: "query", rule: ruleEmail}}},
		{"a percent sign and an at sign that hold no address",
			`{"message":"http_request","query":"q=100%","errors":"a @ b"}`,
			nil},
		{"one in a list", `{"message":"task_submitted","failed":["readability","bob@home.example"]}`,
			[]breach{{event: "task_submitted", field: "failed", rule: ruleEmail}}},
		{"one in an object", `{"message":"startup","error":{"from":"carol@work.example"}}`,
			[]breach{{event: "startup", field: "error", rule: ruleEmail}}},
		{"a number and a truth value", `{"message":"limit_hit","count":3,"user":"u1","limit":"daily_tasks"}`, nil},
		{"a field named like an email address", `{"message":"limit_hit","dan@home.example":"1"}`,
			[]breach{
				{event: "limit_hit", field: withheld, rule: ruleEmail},
				{event: "limit_hit", field: withheld, rule: ruleUndecided},
			}},
		{"an event named like one", `{"message":"erin@home.example"}`,
			[]breach{{event: unknownEvent, field: "message", rule: ruleEmail}, {event: unknownEvent, rule: ruleUnknownEvent}}},
		{"a task handed out, counted as the service counts it",
			`{"message":"task_accepted","topic":"logic.ordering","learner":"Ab3_x-9QzK1mN0pR","host":"claude","language":"pt",` +
				`"grade":2,"cohort":"2026-10","country":"US","region":"US-TX","signin_country":"unknown","drawing":true}`,
			nil},
		{"the drawing itself where whether there was one goes",
			`{"message":"task_accepted","topic":"counting.gaps","drawing":"A  B  C\n●──●──●"}`,
			[]breach{{event: "task_accepted", field: "drawing", rule: ruleShape}}},
		{"an answer and a topic mastered, counted as the service counts them",
			`{"message":"answer_recorded","learner":"Ab3_x-9QzK1mN0pR","grade":6,"cohort":"2026-01","topics_mastered":0}`,
			nil},
		{"a profile's own identifier where the name it is counted under goes",
			`{"message":"topic_mastered","learner":"3b241101-e2bb-4255-8caf-4136c566a962","topic":"logic.ordering","grade":3}`,
			[]breach{{event: "topic_mastered", field: "learner", rule: ruleShape}}},
		{"a country, a region and a host in words",
			`{"message":"task_accepted","country":"Russia","region":"Texas","host":"Claude Desktop"}`,
			[]breach{
				{event: "task_accepted", field: "country", rule: ruleShape},
				{event: "task_accepted", field: "host", rule: ruleShape},
				{event: "task_accepted", field: "region", rule: ruleShape},
			}},
		{"a grade nobody is in, a month there is none of and topics mastered below none",
			`{"message":"answer_recorded","grade":7,"cohort":"2026-13","topics_mastered":-1}`,
			[]breach{
				{event: "answer_recorded", field: "cohort", rule: ruleShape},
				{event: "answer_recorded", field: "grade", rule: ruleShape},
				{event: "answer_recorded", field: "topics_mastered", rule: ruleShape},
			}},
		{"a grade written as text", `{"message":"answer_recorded","grade":"2"}`,
			[]breach{{event: "answer_recorded", field: "grade", rule: ruleShape}}},
		{"a grade that is no whole number", `{"message":"answer_recorded","grade":2.5}`,
			[]breach{{event: "answer_recorded", field: "grade", rule: ruleShape}}},
		{"a host that names a server, on a line no child is counted from",
			`{"message":"cimd_fetch","host":"claude.ai","outcome":"ok"}`,
			nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var fields map[string]any
			if err := json.Unmarshal([]byte(test.line), &fields); err != nil {
				t.Fatalf("the line does not read: %v", err)
			}
			if got := sortedBreaches(audit(fields)); !slices.Equal(got, sortedBreaches(test.want)) {
				t.Errorf("audit(%s) = %+v, want %+v", test.line, got, test.want)
			}
		})
	}
}

// Every line a request writes may carry the request's own fields, and every
// line those the logger writes on all of them, whatever its event; an event
// nobody decided on carries none.
func TestTheFieldsOfEveryLineAreDecidedForEveryEvent(t *testing.T) {
	t.Parallel()

	for event := range events {
		for _, field := range slices.Concat(ofARequest, ofEveryLine) {
			if !Decided(event, field) {
				t.Errorf("Decided(%q, %q) = false, want a field every line may carry", event, field)
			}
		}
	}
	if Known("made_up") || Decided("made_up", "message") {
		t.Error("an event nobody decided on is known, or a field decided for it")
	}
	if Decided("tool_call", instanceField) {
		t.Error("the instance is decided for a line, want it the reader's alone: the service never writes it")
	}
}

// No breach the report writes repeats what a line held, whatever it held and
// wherever: its event is one the service is decided to write or says it is
// not, and its field is named as the service names a field or withheld.
func TestABreachNeverRepeatsWhatALineHeld(t *testing.T) {
	t.Parallel()

	words := gen.OneConstOf("tool_call", "limit_hit", "made_up", "user", "query", "pseudonym", "hello Masha",
		"Masha", "10.1.2.3", "alice@school.example", "bob%40home.example", "carol@work.example", "plain")
	properties := gopter.NewProperties(nil)
	properties.Property("every event and field of a breach is the service's own or stands in for one", prop.ForAll(
		func(event string, keys, values []string) bool {
			fields := map[string]any{"message": event}
			for i, key := range keys {
				fields[key] = values[i%len(values)]
			}
			for _, broken := range audit(fields) {
				if (!Known(broken.event) && broken.event != unknownEvent) ||
					(broken.field != "" && broken.field != withheld && !fieldName.MatchString(broken.field)) ||
					EmailIn(broken.event) != "" || EmailIn(broken.field) != "" {
					return false
				}
			}
			return true
		},
		words, gen.SliceOfN(4, words), gen.SliceOfN(3, words),
	))
	properties.TestingRun(t)
}

// A grade or a count of topics fits however it is held: read back from a
// line's text, which holds every number as a double, or taken from a logger's
// fields, which hold a whole number as one. A fraction, a number past either
// bound, and a number written as text or as anything else do not.
func TestAWholeNumberFitsHoweverItIsHeld(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		value any
		want  bool
	}{
		{"a double read back from a line", 2.0, true},
		{"a double that is no whole number", 2.5, false},
		{"a whole number as a logger holds it", int64(6), true},
		{"a whole number as the language holds it", 1, true},
		{"one below the lowest, as a logger holds it", int64(lowestGrade - 1), false},
		{"one above the highest, as the language holds it", highestGrade + 1, false},
		{"a double above the highest", 7.0, false},
		{"a number written as text", "2", false},
		{"a truth value", true, false},
		{"nothing", nil, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := wholeWithin(test.value, lowestGrade, highestGrade); got != test.want {
				t.Errorf("wholeWithin(%#v, %d, %d) = %v, want %v", test.value, lowestGrade, highestGrade, got, test.want)
			}
		})
	}
}

// Whatever the bounds, a whole number fits exactly when it lies between them,
// held as a double, as a logger's whole number or as the language's; and a
// number half way between two whole ones never fits. The numbers are drawn
// close to the bounds, where a slip of one would show.
func TestAWholeNumberFitsExactlyWhenItLiesWithinTheBounds(t *testing.T) {
	t.Parallel()

	lows, widths, offsets := gen.Int64Range(-1<<20, 1<<20), gen.Int64Range(0, 4), gen.Int64Range(-2, 6)
	properties := gopter.NewProperties(nil)
	properties.Property("a whole number fits as it lies, however it is held", prop.ForAll(
		func(low, width, offset int64) bool {
			high, value := low+width, low+offset
			within := low <= value && value <= high
			return wholeWithin(float64(value), low, high) == within &&
				wholeWithin(value, low, high) == within &&
				wholeWithin(int(value), low, high) == within
		},
		lows, widths, offsets,
	))
	properties.Property("a number half way between two whole ones never fits", prop.ForAll(
		func(low, width, offset int64) bool {
			return !wholeWithin(float64(low+offset)+0.5, low, low+width)
		},
		lows, widths, offsets,
	))
	properties.TestingRun(t)
}

// sortedBreaches are the breaches in one order, so that two lists of the same
// compare equal.
func sortedBreaches(breaches []breach) []breach {
	sorted := slices.Clone(breaches)
	slices.SortFunc(sorted, compareBreaches)
	return sorted
}
