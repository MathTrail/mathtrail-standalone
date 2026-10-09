package picture

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"regexp"
	"slices"
	"strconv"
)

// reading is one pass over a description: the decimal mark its numbers are
// written with, and every problem found so far.
type reading struct {
	decimals Decimals
	problems []Problem
}

// fault records a rule the member at a path breaks.
func (r *reading) fault(path, rule string, args ...any) {
	r.problems = append(r.problems, Problem{Path: path, Rule: fmt.Sprintf(rule, args...)})
}

// object is one object of a description as it is read: where it stands, its
// members, and the members read so far, so that the rest can be named.
type object struct {
	*reading
	path    string
	members map[string]any
	read    map[string]bool
}

// object starts reading an object of the description that stands at a path.
func (r *reading) object(path string, members map[string]any) *object {
	return &object{reading: r, path: path, members: members, read: map[string]bool{}}
}

// at is the path of a member of the object.
func (o *object) at(name string) string { return o.path + "." + name }

// member is the value of a member, marked as read, and whether it holds
// anything: a member written as null, or as an empty text, is one left out.
func (o *object) member(name string) (any, bool) {
	o.read[name] = true
	value := o.members[name]
	if !holds(value) {
		return nil, false
	}
	return value, true
}

// holds says whether a value holds anything: null, and an empty text, hold
// nothing.
func holds(value any) bool {
	text, isText := value.(string)
	return value != nil && (!isText || text != "")
}

// missing records a member the object cannot do without, when it is one.
func (o *object) missing(name string, need bool) {
	if need {
		o.fault(o.at(name), "is missing")
	}
}

// The rules of the values a description is made of, as a refusal says them.
var (
	labelRule = fmt.Sprintf("must be a label: a Latin capital or a run of them, a number written with the "+
		"question's decimal mark and no separator between thousands, or ?, at most %d characters long",
		MaxLabelCharacters)
	timeRule = "must be a time from 0:00 to 23:59, written H:MM"
	noteRule = fmt.Sprintf("must be a note: an equality or an inequality of labels, numbers and ?, joined by "+
		"+ − × ÷ = < > ( ) and spaces, at most %d characters long", MaxNoteCharacters)
)

// objectRule is what a member must be that holds members of its own.
const objectRule = "must be an object"

// label reads a member that is a label, and the empty text when it is left
// out or is no label.
func (o *object) label(name string, need bool) string {
	value, held := o.member(name)
	if !held {
		o.missing(name, need)
		return ""
	}
	text, isText := value.(string)
	if !isText || !o.decimals.isLabel(text) {
		o.fault(o.at(name), "%s", labelRule)
		return ""
	}
	return text
}

// time reads a member that is a time of day.
func (o *object) time(name string, need bool) (Time, bool) {
	value, held := o.member(name)
	if !held {
		o.missing(name, need)
		return Time{}, false
	}
	text, _ := value.(string)
	read, isTime := ReadTime(text)
	if !isTime {
		o.fault(o.at(name), "%s", timeRule)
	}
	return read, isTime
}

// whole reads a member that is a whole number between two limits.
func (o *object) whole(name string, least, most int, need bool) (int, bool) {
	return o.bounded(name, least, most, need, fmt.Sprintf("must be a whole number from %d to %d", least, most))
}

// bounded reads a member that is a whole number between two limits, and says
// what it must be in words of its own where the limits are the picture's own
// numbers, which a refusal never names: the capacity of a container may be
// the answer.
func (o *object) bounded(name string, least, most int, need bool, rule string) (int, bool) {
	value, held := o.member(name)
	if !held {
		o.missing(name, need)
		return 0, false
	}
	number, isWhole := wholeOf(value)
	if !isWhole || number < least || number > most {
		o.fault(o.at(name), "%s", rule)
		return 0, false
	}
	return number, true
}

// wholeOf reads a JSON number that is a whole number: written as one, or as a
// number with no decimals left once it is read, as 3.0 is.
func wholeOf(value any) (int, bool) {
	number, isNumber := value.(json.Number)
	if !isNumber {
		return 0, false
	}
	if whole, err := strconv.ParseInt(string(number), 10, 32); err == nil {
		return int(whole), true
	}
	float, err := strconv.ParseFloat(string(number), 64)
	if err != nil || float != math.Trunc(float) || math.Abs(float) > math.MaxInt32 {
		return 0, false
	}
	return int(float), true
}

// flag reads a member that is true or false, and false when it is left out.
func (o *object) flag(name string) bool {
	value, held := o.member(name)
	if !held {
		return false
	}
	set, isFlag := value.(bool)
	if !isFlag {
		o.fault(o.at(name), "must be true or false")
	}
	return set
}

// word reads a member that is one of a few words.
func (o *object) word(name string, words ...string) {
	value, held := o.member(name)
	if !held {
		return
	}
	if text, isText := value.(string); !isText || !slices.Contains(words, text) {
		o.fault(o.at(name), "must be %s", oneOf(words))
	}
}

// oneOf names the words a member may be: "dot, ring or square".
func oneOf(words []string) string {
	if len(words) == 1 {
		return words[0]
	}
	text := words[0]
	for _, word := range words[1 : len(words)-1] {
		text += ", " + word
	}
	return text + " or " + words[len(words)-1]
}

// list reads a member that is a list holding between two numbers of items. A
// list longer than it may be is not read further, so that a flood of items
// makes one problem rather than a flood of them.
func (o *object) list(name string, least, most int, of string, need bool) []any {
	value, held := o.member(name)
	if !held {
		o.missing(name, need)
		return nil
	}
	items, isList := value.([]any)
	if !isList {
		o.fault(o.at(name), "must be a list of %s", of)
		return nil
	}
	if len(items) < least || len(items) > most {
		o.fault(o.at(name), "must hold %s %s", howMany(least, most), of)
		if len(items) > most {
			return nil
		}
	}
	return items
}

// howMany says how many items a list may hold: "exactly 2", "2 to 12", "at
// most 4".
func howMany(least, most int) string {
	switch least {
	case most:
		return fmt.Sprintf("exactly %d", least)
	case 0:
		return fmt.Sprintf("at most %d", most)
	default:
		return fmt.Sprintf("%d to %d", least, most)
	}
}

// objects reads a member that is a list of objects, each read in its own
// place. An item that is no object is said so, and stands as nil, so that the
// items around it keep their places.
func (o *object) objects(name string, least, most int, of string, need bool) []*object {
	items := o.list(name, least, most, of, need)
	read := make([]*object, len(items))
	for i, item := range items {
		path := fmt.Sprintf("%s.%d", o.at(name), i)
		members, isObject := item.(map[string]any)
		if !isObject {
			o.fault(path, objectRule)
			continue
		}
		read[i] = o.object(path, members)
	}
	return read
}

// labels reads a member that is a list of labels, leaving out each item that
// is not one.
func (o *object) labels(name string, least, most int, need bool) []string {
	var read []string
	for i, item := range o.list(name, least, most, "labels", need) {
		text, isText := item.(string)
		if !isText || !o.decimals.isLabel(text) {
			o.fault(fmt.Sprintf("%s.%d", o.at(name), i), "%s", labelRule)
			continue
		}
		read = append(read, text)
	}
	return read
}

// child reads a member that is an object, and nil when it is left out or is
// not one.
func (o *object) child(name string) *object {
	value, held := o.member(name)
	if !held {
		return nil
	}
	members, isObject := value.(map[string]any)
	if !isObject {
		o.fault(o.at(name), objectRule)
		return nil
	}
	return o.object(o.at(name), members)
}

// keyed reads a member that is an object whose keys are the model's own to
// choose, as the cells a grid marks are. Its entries are judged by whoever
// reads them, each under the path that ends in "*": a key may be the cell or
// the day that is the answer, and a refusal never names it.
func (o *object) keyed(name string) map[string]any {
	value, held := o.member(name)
	if !held {
		return nil
	}
	entries, isObject := value.(map[string]any)
	if !isObject {
		o.fault(o.at(name), objectRule)
		return nil
	}
	return entries
}

// alone says that an object which is a skip holds nothing else: no member
// but skip holds a value. The skip's members are still read as its item's, so
// that one the item does not have is named whatever it holds — its name may
// be anything the model wrote.
func (o *object) alone(rule string) {
	for name, value := range o.members {
		if name != "skip" && holds(value) {
			o.fault(o.path, "%s", rule)
			return
		}
	}
}

// close names every member of the object no reader asked for: a member a kind
// does not have is one the model believes it filled in.
func (o *object) close(what string) {
	for _, name := range slices.Sorted(maps.Keys(o.members)) {
		if !o.read[name] {
			o.fault(o.at(namedSafely(name)), "is no member of %s; leave it out", what)
		}
	}
}

// memberName is a name a member could have: a refusal quotes a member's name
// only when it is one, since anything else the model wrote may be a value.
var memberName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

// namedSafely is a member's name as a refusal may write it.
func namedSafely(name string) string {
	if memberName.MatchString(name) {
		return name
	}
	return "*"
}
