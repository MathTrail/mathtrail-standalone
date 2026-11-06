package picture

import (
	"encoding/json"
	"maps"
	"slices"
)

// asked is the mark of what a picture asks for: a place left for the answer
// says nothing a program could read.
const asked = "?"

// Facts are the members of a description that carry facts: those a program
// proving its task's answer from the picture has to read, by their names as
// the model wrote them, sorted. A description that is no object has none.
//
// The kind names the picture rather than saying anything of it; a member that
// only lays the picture out, as its kind lists them, says nothing a program
// needs; and a member that holds nothing — null, an empty text, the mark of
// what is asked, or nothing but those — has nothing to read. A list that is
// not empty holds at least how many it has, even of empty objects, as the
// posts of a fence do. A description of no kind the format has gives every
// member but its kind, since nothing says which of them only lay it out.
func Facts(raw json.RawMessage) []string {
	members, isObject := Members(raw)
	if !isObject {
		return nil
	}
	kind, _ := members["kind"].(string)
	layout := layoutOf(Kind(kind))
	var facts []string
	for _, name := range slices.Sorted(maps.Keys(members)) {
		if name != "kind" && !slices.Contains(layout, name) && holdsFacts(members[name]) {
			facts = append(facts, name)
		}
	}
	return facts
}

// layoutOf are the members that only lay a picture of a kind out, and none
// for a kind the format does not have.
func layoutOf(kind Kind) []string {
	for _, entry := range kinds {
		if entry.kind == kind {
			return entry.layout
		}
	}
	return nil
}

// holdsFacts says whether a value has anything in it to read: a number or a
// flag, a text other than an empty one or the mark of what is asked, an object
// with a member that holds facts, or a list with an item that is something.
func holdsFacts(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return typed != "" && typed != asked
	case map[string]any:
		for _, member := range typed {
			if holdsFacts(member) {
				return true
			}
		}
		return false
	case []any:
		return slices.ContainsFunc(typed, isSomething)
	default:
		return true
	}
}

// isSomething says whether an item of a list stands for a thing, as an empty
// object does in a row of them, rather than for nothing at all.
func isSomething(item any) bool {
	switch typed := item.(type) {
	case nil:
		return false
	case string:
		return typed != "" && typed != asked
	default:
		return true
	}
}
