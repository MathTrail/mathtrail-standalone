package picture

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// Total is the equality a solution comes to, written large under the picture
// of the solution: the labels it names, and what it ends in after its last =,
// which is the answer.
type Total struct {
	Labels []string
	Result string
}

// totalPath is where a total stands in the task the model writes.
const totalPath = "solution_total"

// totalRule is what a total must be, as a refusal says it.
var totalRule = fmt.Sprintf("must be the equality the solution comes to: labels and numbers, written with the "+
	"question's decimal mark, joined by + − × ÷ = ( ) and spaces, ending in = and the answer, one number or a "+
	"label of at most %d Latin capitals, with no ?, at most %d characters long", MaxLabelCharacters,
	MaxTotalCharacters)

// ReadTotal reads the total under a picture of a solution, its numbers written
// with the decimal mark of the lesson: the labels it names and what it ends in.
// What it ends in is the answer, a number of any length the total holds, or a
// label; the total is written under the picture rather than in it, so a number
// there is held to no label's length. A total that breaks its rule is told by
// its path, never by what it held.
func ReadTotal(text string, decimals Decimals) (Total, []Problem) {
	written := strings.TrimSpace(text)
	last := strings.LastIndex(written, "=")
	result := strings.TrimSpace(written[last+1:])
	if utf8.RuneCountInString(written) > MaxTotalCharacters || last < 0 || strings.ContainsAny(written, "?<>") ||
		!decimals.joined(written) || !decimals.endsATotal(result) {
		return Total{}, []Problem{{Path: totalPath, Rule: totalRule}}
	}
	labels := slices.DeleteFunc(noteTokens(written, decimals), func(token string) bool { return !capitals(token) })
	return Total{Labels: labels, Result: result}, nil
}

// endsATotal says whether a text may stand for the answer a total ends in: a
// number of any length, or a label.
func (d Decimals) endsATotal(text string) bool {
	return d.isNumber(text) || d.isLabel(text)
}
