package report

import (
	"cmp"
	"regexp"
	"slices"
)

// The rules of the log a line can break.
const (
	ruleUnknownEvent = "an event the service is not decided to write"
	ruleUndecided    = "a field its event is not decided to carry"
	ruleEmail        = "text shaped like an email address"
	ruleShape        = "a value its field may not hold"
)

// What a breach says in place of a name it must not repeat: the words of an
// event nobody decided on, which may be anything at all, a child's name or an
// address among them, and a field named otherwise than a field of the service
// is.
const (
	unknownEvent = "(an event not in the table)"
	withheld     = "(withheld)"
)

// fieldName is how the service names a field, and a log collector its own
// keys: a lowercase letter first, then letters, digits, dots, slashes, dashes
// and underscores.
var fieldName = regexp.MustCompile(`^[a-z][a-zA-Z0-9_./-]{0,63}$`)

// breach is one way a line breaks the rules of the log: the event, the field
// and the rule. It names what the line held never.
type breach struct{ event, field, rule string }

// audit holds a line of the service's to the rules of the log and is every
// way it breaks them: an event the service is not decided to write, a field
// its event is not decided to carry, a value of a form its field may not hold,
// and text shaped like an email address in any of its values or names.
func audit(fields map[string]any) []breach {
	event, _ := fields["message"].(string)
	label := event
	var found []breach
	if !Known(event) {
		label = unknownEvent
		found = append(found, breach{event: label, rule: ruleUnknownEvent})
	}
	for field, value := range fields {
		// The instance is the reader's to add, beside what the service wrote.
		if Known(event) && !Decided(event, field) && field != instanceField {
			found = append(found, breach{event: label, field: named(field), rule: ruleUndecided})
		}
		if Known(event) && Decided(event, field) && !Fits(event, field, value) {
			found = append(found, breach{event: label, field: named(field), rule: ruleShape})
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

// named is a field's name as a breach gives it: as it is when it is named the
// way the service names a field, and withheld otherwise.
func named(name string) string {
	if !fieldName.MatchString(name) {
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
