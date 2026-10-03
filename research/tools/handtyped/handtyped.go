package main

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// Finding is a number left in a paper's text that no macro printed.
type Finding struct {
	Line   int
	Number string
	Text   string // the line it stands in
}

func (f Finding) String() string {
	return fmt.Sprintf("line %d: %q in %q", f.Line, f.Number, strings.TrimSpace(f.Text))
}

// quoting are the commands whose arguments may hold digits that are not the
// paper's numbers: the macros that print a computed number, with how it is
// rounded, and those that stop the build when one breaks a claim; a number the
// text chooses, marked as such; the references, labels, citations and files;
// and the marks of what a later task supplies.
var quoting = []string{
	"stat", "statr", "statpct", "statabs", "statnum", "statword", "Statword",
	"statmin", "statmax", "statminpct", "statmaxpct", "statabsmin", "statabsmax",
	"statzero", "statless", "statvszero", "given",
	"ref", "label", "cite", "input", "include", "includegraphics", "url",
	"begin", "end", "TBD", "documentclass", "usepackage", "bibliography", "bibliographystyle",
}

// dimension is a length in a layout command or a column, not a number of the
// paper. LaTeX writes it with its unit attached, which sets it apart from a
// number followed by a word: "3.8cm" is a length, "3 in 10" is not.
var dimension = regexp.MustCompile(`\d+(\.\d+)?(cm|mm|pt|em|ex|in|bp|pc|sp)\b`)

// digitRun is a number as the text writes it.
var digitRun = regexp.MustCompile(`\d+([.,]\d+)*`)

// Find returns every number in the text that stands outside the commands that
// may quote numbers and outside the names given, which carry digits of their
// own. In a formula the digits 0, 1 and 2 are its structure — 1 − c, the
// square, the answers 0 and 1 — and so is the base of ln 10; any other number
// there is a number of the paper too.
func Find(text string, names []*regexp.Regexp) []Finding {
	blank := []byte(text)
	blankOut := func(from, to int) {
		for i := from; i < to; i++ {
			if blank[i] != '\n' {
				blank[i] = ' '
			}
		}
	}
	for _, span := range comments(text) {
		blankOut(span[0], span[1])
	}
	for _, span := range quoted(string(blank)) {
		blankOut(span[0], span[1])
	}
	for _, pattern := range append(slices.Clone(names), dimension) {
		for _, span := range pattern.FindAllIndex(blank, -1) {
			// A name carries letters, and so does a length; a match of
			// digits alone is a number, whatever pattern matched it.
			if bytes.IndexFunc(blank[span[0]:span[1]], unicode.IsLetter) >= 0 {
				blankOut(span[0], span[1])
			}
		}
	}
	for _, span := range structural(string(blank)) {
		blankOut(span[0], span[1])
	}
	var found []Finding
	lines := strings.Split(text, "\n")
	for _, span := range digitRun.FindAllIndex(blank, -1) {
		line := strings.Count(text[:span[0]], "\n")
		found = append(found, Finding{Line: line + 1, Number: text[span[0]:span[1]], Text: lines[line]})
	}
	return found
}

// comments are the spans from an unescaped percent sign to the end of its line.
func comments(text string) [][2]int {
	var spans [][2]int
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\\':
			// An address prints its percent signs as they are.
			if strings.HasPrefix(text[i:], `\url{`) {
				if closing := group(text, i+len(`\url`)); closing >= 0 {
					i = closing
					continue
				}
			}
			i++ // an escaped character, \% among them, starts no comment
		case '%':
			end := strings.IndexByte(text[i:], '\n')
			if end < 0 {
				end = len(text) - i
			}
			spans = append(spans, [2]int{i, i + end})
			i += end
		}
	}
	return spans
}

// command is a control sequence's name where it starts.
var command = regexp.MustCompile(`\\([A-Za-z]+)`)

// quoted are the spans of the commands that may quote numbers, each with its
// optional and its braced arguments.
func quoted(text string) [][2]int {
	var spans [][2]int
	for _, at := range command.FindAllStringSubmatchIndex(text, -1) {
		if !slices.Contains(quoting, text[at[2]:at[3]]) {
			continue
		}
		spans = append(spans, [2]int{at[0], arguments(text, at[1])})
	}
	return spans
}

// arguments is where a command's arguments end: the bracketed groups before
// its braced ones and the braced groups, spaces between them allowed. A
// bracket after a braced group, or one that never closes, is the text's own;
// a brace that never closes swallows the rest, as LaTeX would.
func arguments(text string, from int) int {
	end, braced := from, false
	for i := from; i < len(text); {
		switch text[i] {
		case ' ', '\t':
			i++
			continue
		case '[':
			closing := group(text, i)
			if braced || closing < 0 {
				return end
			}
			end, i = closing+1, closing+1
			continue
		case '{':
			closing := group(text, i)
			if closing < 0 {
				return len(text)
			}
			end, i, braced = closing+1, closing+1, true
			continue
		}
		break
	}
	return end
}

// group is where the group opening at a brace or a bracket closes, counting
// the groups inside it, or -1 when it never does.
func group(text string, open int) int {
	opening, closing := text[open], byte('}')
	if opening == '[' {
		closing = ']'
	}
	depth := 0
	for i := open; i < len(text); i++ {
		switch text[i] {
		case '\\':
			i++
		case opening:
			depth++
		case closing:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// formulas are the spans of the text's formulas: inline between dollars or
// \( \), displayed between \[ \]. An escaped character opens none, so the line
// break \\[-1pt] and the amount \$8 stay text.
func formulas(text string) [][2]int {
	var spans [][2]int
	closers := map[string]string{"$": "$", `\(`: `\)`, `\[`: `\]`}
	for i := 0; i < len(text); i++ {
		var opener string
		switch {
		case strings.HasPrefix(text[i:], `\(`), strings.HasPrefix(text[i:], `\[`):
			opener = text[i : i+2]
		case text[i] == '\\':
			i++ // an escaped character, \\ and \$ among them, opens nothing
			continue
		case text[i] == '$':
			opener = "$"
		default:
			continue
		}
		end := closing(text, i+len(opener), closers[opener])
		spans = append(spans, [2]int{i, end})
		i = end - 1
	}
	return spans
}

// closing is where a formula that opened before from ends: just past its
// unescaped closer, or at the end of the text when it has none.
func closing(text string, from int, closer string) int {
	for i := from; i < len(text); i++ {
		if strings.HasPrefix(text[i:], closer) {
			return i + len(closer)
		}
		if text[i] == '\\' {
			i++
		}
	}
	return len(text)
}

// logBase is the base of ln 10, which a formula writes as a number.
var logBase = regexp.MustCompile(`\\ln\s*10`)

// structural are the spans of the numbers a formula's own structure needs: a
// whole 0, 1 or 2, and the base of ln 10.
func structural(text string) [][2]int {
	var spans [][2]int
	for _, formula := range formulas(text) {
		inside := text[formula[0]:formula[1]]
		for _, span := range logBase.FindAllStringIndex(inside, -1) {
			spans = append(spans, [2]int{formula[0] + span[0], formula[0] + span[1]})
		}
		for _, span := range digitRun.FindAllStringIndex(inside, -1) {
			if number := inside[span[0]:span[1]]; number == "0" || number == "1" || number == "2" {
				spans = append(spans, [2]int{formula[0] + span[0], formula[0] + span[1]})
			}
		}
	}
	return spans
}

// layout is what a figure states only to lay itself out: the name of an
// environment, whose options are read, a macro defined to a value, a
// coordinate pair of numbers or of a figure's macros, a key of TikZ or
// pgfplots set to a number, and a colour mixed by a share; the styles pgfplots
// is told to use are blanked apart. Only these forms are: parentheses or a
// setting in a node's words are read, and so is a key not listed here.
var layout = []*regexp.Regexp{
	regexp.MustCompile(`\\(begin|end)\{[^{}]*\}`),
	regexp.MustCompile(`\\def\\[A-Za-z]+\{[^{}]*\}`),
	regexp.MustCompile(`\(\s*(\\[A-Za-z]+|[-+]?(\d+(\.\d*)?|\.\d+))\s*,\s*(\\[A-Za-z]+|[-+]?(\d+(\.\d*)?|\.\d+))\s*\)`),
	regexp.MustCompile(`\b(x|y|xshift|yshift|width|height|enlarge [xy] limits|rounded corners|minimum width|minimum height|inner sep|outer sep|above|below|left|right|line width)\s*=\s*[-+]?(\d+(\.\d*)?|\.\d+)`),
	regexp.MustCompile(`!\d+`),
}

// FigureWords is a figure with its layout blanked out, lines kept where they
// were, so that Find reads what is left: the words the figure prints, in a
// node, a label, a legend or a tick, whatever syntax prints them. Whatever
// the blanking does not recognise is read, so an unforeseen place for a
// number is checked rather than passed.
func FigureWords(text string) string {
	kept := []byte(text)
	blankOut := func(from, to int) {
		for i := from; i < to; i++ {
			if kept[i] != '\n' {
				kept[i] = ' '
			}
		}
	}
	for _, at := range command.FindAllStringSubmatchIndex(text, -1) {
		if text[at[2]:at[3]] == "pgfplotsset" {
			blankOut(at[0], arguments(text, at[1]))
		}
	}
	for _, pattern := range layout {
		for _, span := range pattern.FindAllIndex(kept, -1) {
			blankOut(span[0], span[1])
		}
	}
	return string(kept)
}
