package solver

import (
	"math/big"
	"regexp"
	"strings"
)

// Count is how many options a task offers.
const Count = 5

// letters label the options, in order. They are not exported as a variable
// because a package-level array is writable by whoever imports it, and the
// labels of the options are not somebody else's to change.
var letters = [Count]string{"A", "B", "C", "D", "E"}

// Letter is the label of the option in that place, counting from zero.
func Letter(place int) string { return letters[place] }

// Place is where a label stands, or -1 when it labels nothing.
func Place(letter string) int {
	for place, label := range letters {
		if label == letter {
			return place
		}
	}
	return -1
}

// Options are the option texts of one task, in the order of the labels.
type Options [Count]string

// relabel is how far the labels move between the two runs of a program: the
// text labelled A comes back labelled C, B comes back D, and so on around the
// five. Two is the smallest shift that leaves no option where it was and no
// pair of options swapped, so a program that repeats an old letter — or one
// that mirrors the list — cannot land on the right one by accident.
const relabel = 2

// Rotated is the same options under shifted labels: the second run of a
// program sees the same five texts, each under a different letter. A program
// that computes an answer and looks it up follows the text and returns the new
// label; a program that writes a letter out by hand returns the old one.
func (o *Options) Rotated() Options {
	var rotated Options
	for place, text := range o {
		rotated[(place+relabel)%Count] = text
	}
	return rotated
}

// Relabelled is where a letter of the first run stands in the second.
func Relabelled(letter string) string {
	place := Place(letter)
	if place < 0 {
		return ""
	}
	return letters[(place+relabel)%Count]
}

// Match is the letters whose option text means the same as a computed value.
//
// All of them, not the first: two options saying one thing is a fault of the
// task, and a verdict that quietly picked one of them would hide it. None of
// them is the most valuable answer this check gives — the computed answer is
// not among the options at all.
func (o *Options) Match(value string) []string {
	matched := []string{}
	wanted := number(value)
	for place, text := range o {
		if same(text, value, wanted) {
			matched = append(matched, letters[place])
		}
	}
	return matched
}

// same is one option text against one computed value. Numbers are compared as
// numbers, so that six, six point zero and a padded six are one answer; it is
// only when either side is not a number that the two are read as text.
func same(text, value string, wanted *big.Rat) bool {
	if wanted != nil {
		if have := number(text); have != nil {
			return have.Cmp(wanted) == 0
		}
	}
	return strings.EqualFold(strings.TrimSpace(text), strings.TrimSpace(value))
}

// decimal is the only shape read as a number: a whole number or a plain
// decimal fraction. Everything else stays text, which is what keeps "24/7"
// from being three and a bit, and it is deliberate that no unit is stripped —
// an option that says twelve centimetres is matched by passing those words.
var decimal = regexp.MustCompile(`^[+-]?\d+(\.\d+)?$`)

// number reads a value as an exact rational, or nil when it is not a number.
// Exact rather than floating: an answer counted by a brute force can be longer
// than a float can hold, and a comparison that rounds would call two different
// answers the same.
func number(s string) *big.Rat {
	s = strings.TrimSpace(s)
	if !decimal.MatchString(s) {
		return nil
	}
	parsed, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil
	}
	return parsed
}
