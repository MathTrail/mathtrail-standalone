package report

import (
	"cmp"
	"slices"
)

// The rules of the log a line can break.
const (
	ruleUnknownEvent = "an event the service is not decided to write"
	ruleUndecided    = "a field its event is not decided to carry"
	ruleEmail        = "text shaped like an email address"
)

// withheld stands in a breach for an event or a field whose own name is
// shaped like an email address: the report repeats no such text.
const withheld = "(withheld)"

// breach is one way a line breaks the rules of the log: the event, the field
// and the rule. It names what the line held never.
type breach struct{ event, field, rule string }

// audit holds a line of the service's to the rules of the log and is every
// way it breaks them: an event the service is not decided to write, a field
// its event is not decided to carry, and text shaped like an email address in
// any of its values or names.
func audit(fields map[string]any) []breach {
	event, _ := fields["message"].(string)
	label := named(event)
	var found []breach
	if !Known(event) {
		found = append(found, breach{event: label, rule: ruleUnknownEvent})
	}
	for field, value := range fields {
		// The instance is the reader's to add, beside what the service wrote.
		if Known(event) && !Decided(event, field) && field != instanceField {
			found = append(found, breach{event: label, field: named(field), rule: ruleUndecided})
		}
		if EmailIn(field) != "" || carriesEmail(value) {
			found = append(found, breach{event: label, field: named(field), rule: ruleEmail})
		}
	}
	return found
}

// compareBreaches orders breaches by event, field and rule.
func compareBreaches(a, b breach) int {
	return cmp.Or(cmp.Compare(a.event, b.event), cmp.Compare(a.field, b.field), cmp.Compare(a.rule, b.rule))
}

// named is a name as a breach gives it: as it is, unless it is itself shaped
// like an email address.
func named(name string) string {
	if EmailIn(name) != "" {
		return withheld
	}
	return name
}

// carriesEmail says whether a value of a line holds text shaped like an email
// address: a text, or any item or key of a list or an object. A number and a
// truth value hold no text.
func carriesEmail(value any) bool {
	switch held := value.(type) {
	case string:
		return EmailIn(held) != ""
	case []any:
		return slices.ContainsFunc(held, carriesEmail)
	case map[string]any:
		for key, item := range held {
			if EmailIn(key) != "" || carriesEmail(item) {
				return true
			}
		}
	}
	return false
}
