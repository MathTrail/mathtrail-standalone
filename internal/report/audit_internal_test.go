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
			[]breach{{event: "made_up", rule: ruleUnknownEvent}}},
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
			[]breach{{event: withheld, field: "message", rule: ruleEmail}, {event: withheld, rule: ruleUnknownEvent}}},
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

// No breach the report writes repeats an email address, whatever a line held
// and wherever: a name shaped like one is withheld.
func TestABreachNeverRepeatsAnEmailAddress(t *testing.T) {
	t.Parallel()

	words := gen.OneConstOf("tool_call", "limit_hit", "made_up", "user", "query", "pseudonym",
		"alice@school.example", "bob%40home.example", "carol@work.example", "plain")
	properties := gopter.NewProperties(nil)
	properties.Property("every event and field of a breach is free of an email address", prop.ForAll(
		func(event string, keys, values []string) bool {
			fields := map[string]any{"message": event}
			for i, key := range keys {
				fields[key] = values[i%len(values)]
			}
			for _, broken := range audit(fields) {
				if EmailIn(broken.event) != "" || EmailIn(broken.field) != "" {
					return false
				}
			}
			return true
		},
		words, gen.SliceOfN(4, words), gen.SliceOfN(3, words),
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
