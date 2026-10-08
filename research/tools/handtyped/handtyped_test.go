package main

import (
	"regexp"
	"slices"
	"testing"
)

// names are the names with digits the tests let through.
var names = []*regexp.Regexp{regexp.MustCompile(`RQ[12]`), regexp.MustCompile(`Glicko-2`)}

// numbers are the numbers Find reports, in order.
func numbers(found []Finding) []string {
	out := make([]string, len(found))
	for i, f := range found {
		out[i] = f.Number
	}
	return out
}

func TestFindReportsOnlyNumbersNoMacroPrinted(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		text string
		want []string
	}{
		{"a number in the text", "only 56.6 of 150 dialogues", []string{"56.6", "150"}},
		{"a number a macro prints", `only \stat{literature}{gupta_percent}\pct{} of them`, nil},
		{"a rounding a macro is given", `\statr{ea3}{key}{2} and \statpct{ea3}{key}{0}`, nil},
		{"a range over listed keys", `\statmin{ea3}{2}{a_g0_r1,b_g0_r1}`, nil},
		{"a range of magnitudes over listed keys", `\statabsmin{ea3}{2}{a_g2_r6,b_g2_r6} to \statabsmax{ea3}{2}{a_g2_r6,b_g2_r6}`, nil},
		{"a comparison the build guards", `lower\statless{ea3}{glicko2_g4_rms_200}{service_g4_rms_200}`, nil},
		{"a sign the build guards", `as it was\statvszero{ea3}{pc4_floor_0_01_g3}{=}.`, nil},
		{"prose in brackets after a reference", "As \\ref{sec:3} [see 12 cases\nmore text with 603 tasks", []string{"12", "603"}},
		{"a bracket after a label is prose", `\label{x} [9 of 10]`, []string{"9", "10"}},
		{"an optional argument before the braces is quoted", `\cite[p.~12]{key}`, nil},
		{"a percent sign inside an address is no comment", `see \url{https://a.org/x%20y} then 42 and } 55`, []string{"42", "55"}},
		{"a number the text chooses", `at a chance of \given{0.6}`, nil},
		{"a reference, a label and a citation", `Section~\ref{sec:3} \label{eq:2} \cite{glickman2022glicko2}`, nil},
		{"a comment", "text % 42 is a note\nmore", nil},
		{"an escaped percent is no comment", `50\% of them`, []string{"50"}},
		{"a column width", `\begin{tabular}{@{}p{3.1cm}r@{}}`, nil},
		{"a length in a command", `\hspace{2pt} and 1.5em`, nil},
		{"a number before a unit's word", "3 in 10 children", []string{"3", "10"}},
		{"an escaped line break is no formula", `line\\[2pt] then 2 and 1 here \[ x \]`, []string{"2", "1"}},
		{"an escaped dollar is no formula", `\$2 and \$1`, []string{"2", "1"}},
		{"a name with a digit", "RQ1 and Glicko-2 but RQ3", []string{"3"}},
		{"a formula's structure", `$1 - c$, $x^2$, $S \in \{0, 1\}$, $\ln 10$, $\frac{(\theta - \theta_0)^2}{2\sigma_0^2}$`, nil},
		{"a number in a formula", `$P = 0.85$ and \[ K = 0.05\,n \]`, []string{"0.85", "0.05"}},
		{"a later task's mark", `\TBD{K05: the authors of S63}`, nil},
		{"a comment on the last line", "7 % 42 is a note", []string{"7"}},
		{"a formula that never closes", "$x^2 + 5", []string{"5"}},
		{"a power in a formula", "$2^{10}$ cells", []string{"10"}},
		{"an escaped brace inside a quoted argument", `\given{\}0.6} and 5`, []string{"5"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := numbers(Find(c.text, names)); !slices.Equal(got, c.want) {
				t.Errorf("Find(%q) = %q, want %q", c.text, got, c.want)
			}
		})
	}
}

// A number is reported on the line it stands on, even inside a formula that
// runs over several lines.
func TestFindNamesTheLineOfANumber(t *testing.T) {
	t.Parallel()
	found := Find("first line\n\\[\n  K = 1 + 3\n\\]\nlast 7", names)
	got := make([]int, len(found))
	for i, f := range found {
		got[i] = f.Line
	}
	if want := []int{3, 5}; !slices.Equal(got, want) {
		t.Errorf("lines %v, want %v", got, want)
	}
}

// A name is a word that carries digits: a pattern of the names file that
// matches digits alone lets nothing through, however it was written.
func TestFindLetsThroughOnlyNamesWithLetters(t *testing.T) {
	t.Parallel()
	digitsOnly := []*regexp.Regexp{regexp.MustCompile(`\d{5}`), regexp.MustCompile(`\d+-\d+`)}
	found := Find("we saw 12345 tasks over 10-20 days of RQ1", append(digitsOnly, names...))
	if got, want := numbers(found), []string{"12345", "10", "20"}; !slices.Equal(got, want) {
		t.Errorf("numbers = %q, want %q", got, want)
	}
}

// A group that never closes swallows the rest of the text rather than
// letting the checker read past the end.
func TestFindSurvivesAnUnclosedGroup(t *testing.T) {
	t.Parallel()
	if got := Find(`\stat{product}{key 12`, names); len(got) != 0 {
		t.Errorf("Find with an unclosed group = %v, want nothing", got)
	}
}

// A figure is read with its layout blanked out: a number in a node's text, a
// label, a legend or a typed tick is typed by hand, whatever syntax prints it,
// while the coordinates, option values and macro values around it are layout.
func TestFigureWordsKeepWhatAFigurePrints(t *testing.T) {
	t.Parallel()
	figure := "\\def\\xModel{3.3}\n" +
		"\\draw[msg] (\\xModel,-1.0) -- node[note] {asks for 3 tasks} (6.6,-1.0);\n" +
		"\\node[party] (card) at (0,0) {Child\\\\[-1pt]{\\scriptsize the card}};\n" +
		"\\begin{axis}[width=3.8cm, enlarge y limits=0.04, grid style={gray!25}, xlabel={tasks in the corridor, \\%}, ylabel={over 200 answers}]\n" +
		"\\node at (0,0) [draw] {4 more};\n" +
		"\\addlegendentry{5 rules} \\pgfplotsset{compat=1.18, rules/.style={height=5.0cm}}\n"
	found := Find(FigureWords(figure), names)
	if got, want := numbers(found), []string{"3", "200", "4", "5"}; !slices.Equal(got, want) {
		t.Errorf("numbers in the figure's words = %q, want %q", got, want)
	}
	if lines := []int{2, 4, 5, 6}; len(found) != len(lines) {
		t.Errorf("found %d numbers, want %d", len(found), len(lines))
	} else {
		for i, f := range found {
			if f.Line != lines[i] {
				t.Errorf("%q on line %d, want %d", f.Number, f.Line, lines[i])
			}
		}
	}
}

// Parentheses and settings are layout only in the forms a figure lays itself
// out with: a coordinate pair and a known key set to a number. A count in a
// node's text, in parentheses or set equal in words, is typed by hand.
func TestFigureWordsReadCountsInTheirText(t *testing.T) {
	t.Parallel()
	figure := "\\node {asks for tasks (of 12)};\n" +
		"\\node {answers, n = 200};\n" +
		"\\draw (\\xModel,-4.75) -- (0.6,0) -- (card);\n" +
		"\\node[minimum width=22mm, inner sep=2pt, rounded corners=2pt, xshift=4pt] at (0,-0.3) {x};\n"
	if got, want := numbers(Find(FigureWords(figure), names)), []string{"12", "200"}; !slices.Equal(got, want) {
		t.Errorf("numbers in the figure's words = %q, want %q", got, want)
	}
}
