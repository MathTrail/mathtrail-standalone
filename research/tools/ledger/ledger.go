package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

// Claim is one fact a paper states about the product or its prototype, with
// its proof.
type Claim struct {
	ID       string   `json:"id"`
	Group    string   `json:"group"`
	Text     string   `json:"claim"`
	Status   string   `json:"status"`
	Evidence []Anchor `json:"evidence"`
	// Numbers are numbers of the claim a paper prints, each under the key its
	// macro reads; each must stand in a line that proves the claim.
	Numbers map[string]string `json:"numbers,omitempty"`
	UsedIn  string            `json:"used_in"`
}

// Anchor is one proof: either the line of a file that matches a pattern, or a
// fact a script computed at the commit. A fact may carry the value the claim's
// text relies on, so that a later commit with another value fails the run.
type Anchor struct {
	File   string `json:"file,omitempty"`
	Match  string `json:"match,omitempty"`
	Stat   string `json:"stat,omitempty"`
	Expect string `json:"expect,omitempty"`
}

// statuses are the kinds of claim the ledger allows:
//   - built: code and tests do it;
//   - specified: only the specification, the diagrams or a decision log describe it;
//   - planned: an open product task will build it;
//   - measured: a number a report or the specification recorded, not reproduced by a script;
//   - unverified: what a research note reports from a source nobody here has
//     checked yet;
//   - remark: a document that disagrees with the code, for the author;
//   - context: rests on a file no one else can check.
var statuses = map[string]bool{
	"built": true, "specified": true, "planned": true, "measured": true,
	"unverified": true, "remark": true, "context": true,
}

// commitPattern is what a pinned commit must look like: a hash long enough
// that git cannot take it for another object.
var commitPattern = regexp.MustCompile(`^[0-9a-f]{12,40}$`)

// Source returns a file's content at the pinned commit.
type Source func(path string) ([]byte, error)

// Stats are the facts a script computed at one commit of one repository.
type Stats struct {
	Repository string // the public name a reader finds the commit under
	Commit     string
	Date       string
	Values     map[string]string
}

// Proof is one resolved anchor: a line of a file at the commit, or a fact a
// script computed at the commit.
type Proof struct {
	Place string // "path:line", or "key = value" for a fact
	Fact  bool
	Text  string // the line itself, or the value of the fact
}

// Row is a claim with its proofs resolved to places a reader can look up.
type Row struct {
	Claim  Claim
	Proofs []Proof
}

// ParseClaims reads the claims file and rejects a claim that could never be
// resolved: a duplicate or missing ID or group, an unknown status, no proof,
// an anchor that is neither a file and a pattern nor a fact, or a pattern
// that does not compile. The file is one list of claims and nothing else, so
// claims left after it by a bad merge cannot drop out unseen.
func ParseClaims(data []byte) ([]Claim, error) {
	var claims []Claim
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("parse claims: something follows the list of claims")
	}
	if len(claims) == 0 {
		return nil, errors.New("parse claims: no claims")
	}
	seen := make(map[string]bool, len(claims))
	numberedBy := make(map[string]string)
	var problems []error
	for i := range claims {
		c := &claims[i]
		where := fmt.Sprintf("claim %d (%s)", i+1, c.ID)
		switch {
		case c.ID == "":
			problems = append(problems, fmt.Errorf("%s: no id", where))
		case seen[c.ID]:
			problems = append(problems, fmt.Errorf("%s: id used twice", where))
		}
		seen[c.ID] = true
		problems = append(problems, checkClaim(where, c)...)
		for _, key := range numberKeys(c) {
			if other, taken := numberedBy[key]; taken {
				problems = append(problems, fmt.Errorf("%s: number %s is %s's already", where, key, other))
			}
			numberedBy[key] = c.ID
		}
	}
	return claims, errors.Join(problems...)
}

// numberKeyPattern is what the paper's macros accept as a key.
var numberKeyPattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// numberKeys are a claim's number keys in order, so that what the tool reports
// and writes does not depend on the order of a map.
func numberKeys(c *Claim) []string {
	keys := make([]string, 0, len(c.Numbers))
	for key := range c.Numbers {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// standsIn says whether a number stands in a text as a whole number, not as a
// part of a longer one: "33" stands in "33 pairs" but not in "133" or "0.33".
func standsIn(number, text string) bool {
	pattern := regexp.MustCompile(`(^|[^0-9.,])` + regexp.QuoteMeta(number) + `($|[^0-9.,]|[.,]($|[^0-9]))`)
	return pattern.MatchString(text)
}

// checkClaim finds what makes one claim unresolvable on its own, whatever the
// other claims say.
func checkClaim(where string, c *Claim) []error {
	var problems []error
	if strings.TrimSpace(c.Text) == "" {
		problems = append(problems, fmt.Errorf("%s: no claim text", where))
	}
	if strings.TrimSpace(c.Group) == "" {
		problems = append(problems, fmt.Errorf("%s: no group", where))
	}
	if !statuses[c.Status] {
		problems = append(problems, fmt.Errorf("%s: unknown status %q", where, c.Status))
	}
	if len(c.Evidence) == 0 {
		problems = append(problems, fmt.Errorf("%s: no evidence", where))
	}
	for j, a := range c.Evidence {
		byFile := a.File != "" && a.Match != ""
		byStat := a.Stat != ""
		if byFile == byStat || (a.File == "") != (a.Match == "") {
			problems = append(problems, fmt.Errorf("%s: evidence %d must be a file with a match, or a stat", where, j+1))
		}
		if a.Expect != "" && !byStat {
			problems = append(problems, fmt.Errorf("%s: evidence %d expects a value but is not a stat", where, j+1))
		}
		if _, err := regexp.Compile(a.Match); err != nil {
			problems = append(problems, fmt.Errorf("%s: evidence %d: pattern %q: %w", where, j+1, a.Match, err))
		}
	}
	return append(problems, numberProblems(where, c)...)
}

// numberProblems finds the numbers of a claim the paper's macros could not
// read: a key they do not accept, or a number with no value.
func numberProblems(where string, c *Claim) []error {
	var problems []error
	for _, key := range numberKeys(c) {
		if !numberKeyPattern.MatchString(key) {
			problems = append(problems, fmt.Errorf("%s: number key %q is not lower-case letters, digits and underscores", where, key))
		}
		if strings.TrimSpace(c.Numbers[key]) == "" {
			problems = append(problems, fmt.Errorf("%s: number %s has no value", where, key))
		}
	}
	return problems
}

// ParseStats reads key=value lines, each key once. The repository and the
// commit they were computed at are what the ledger is pinned to, so both must
// be there, the commit in full enough form.
func ParseStats(text string) (Stats, error) {
	values := make(map[string]string)
	for n, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Stats{}, fmt.Errorf("stats line %d: no '=': %q", n+1, line)
		}
		key = strings.TrimSpace(key)
		if _, seen := values[key]; seen {
			return Stats{}, fmt.Errorf("stats line %d: %q is given twice", n+1, key)
		}
		values[key] = strings.TrimSpace(value)
	}
	stats := Stats{Repository: values["repository"], Commit: values["commit"], Date: values["commit_date"], Values: values}
	if stats.Repository == "" {
		return Stats{}, errors.New("stats name no repository")
	}
	if !commitPattern.MatchString(stats.Commit) {
		return Stats{}, fmt.Errorf("stats name no commit of at least 12 hex digits: %q", stats.Commit)
	}
	if stats.Date == "" {
		return Stats{}, errors.New("stats name no commit_date")
	}
	return stats, nil
}

// Resolve finds every claim's proofs. It reports every missing or changed
// proof at once, so moving to a new commit shows the whole list of claims to
// revisit.
func Resolve(claims []Claim, show Source, stats Stats) ([]Row, error) {
	files := make(map[string][]byte)
	read := func(path string) ([]byte, error) {
		if content, ok := files[path]; ok {
			return content, nil
		}
		content, err := show(path)
		if err != nil {
			return nil, err
		}
		files[path] = content
		return content, nil
	}

	rows := make([]Row, 0, len(claims))
	var problems []error
	for i := range claims {
		c := &claims[i]
		row := Row{Claim: *c}
		for _, a := range c.Evidence {
			proof, err := resolveAnchor(a, read, stats)
			if err != nil {
				problems = append(problems, fmt.Errorf("%s: %w", c.ID, err))
				continue
			}
			row.Proofs = append(row.Proofs, proof)
		}
		for _, key := range numberKeys(c) {
			if !slices.ContainsFunc(row.Proofs, func(p Proof) bool { return standsIn(c.Numbers[key], p.Text) }) {
				problems = append(problems, fmt.Errorf("%s: number %s = %s stands in none of its proofs", c.ID, key, c.Numbers[key]))
			}
		}
		rows = append(rows, row)
	}
	return rows, errors.Join(problems...)
}

// RenderNumbers writes the claims' numbers as key=value lines for the paper's
// macros, each claim's under a comment that names it, in the claims' order.
func RenderNumbers(rows []Row, stats Stats) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Numbers the claims of the evidence ledger state, each found in a line that proves its claim at %s.\n", stats.Commit)
	for i := range rows {
		c := &rows[i].Claim
		if len(c.Numbers) == 0 {
			continue
		}
		fmt.Fprintf(&b, "# %s\n", c.ID)
		for _, key := range numberKeys(c) {
			fmt.Fprintf(&b, "%s=%s\n", key, c.Numbers[key])
		}
	}
	return b.String()
}

func resolveAnchor(a Anchor, read Source, stats Stats) (Proof, error) {
	if a.Stat != "" {
		value := stats.Values[a.Stat]
		switch {
		case value == "":
			return Proof{}, fmt.Errorf("no value for fact %q in the stats", a.Stat)
		case a.Expect != "" && value != a.Expect:
			return Proof{}, fmt.Errorf("fact %q is %s, the claim says %s", a.Stat, value, a.Expect)
		}
		return Proof{Place: a.Stat + " = " + value, Fact: true, Text: value}, nil
	}
	pattern, err := regexp.Compile(a.Match)
	if err != nil {
		return Proof{}, fmt.Errorf("pattern %q: %w", a.Match, err)
	}
	content, err := read(a.File)
	if err != nil {
		return Proof{}, err
	}
	// A pattern that matches several lines might point at the wrong one after
	// the file changes, and nothing would notice; so a proof is one line.
	line, matches := Locate(content, pattern)
	switch {
	case matches == 0:
		return Proof{}, fmt.Errorf("no line of %s matches %q", a.File, a.Match)
	case matches > 1:
		return Proof{}, fmt.Errorf("%d lines of %s match %q; a proof must match exactly one", matches, a.File, a.Match)
	}
	return Proof{Place: fmt.Sprintf("%s:%d", a.File, line), Text: strings.Split(string(content), "\n")[line-1]}, nil
}

// Locate returns the number of the first line that matches, counting from 1,
// and how many lines match in all.
func Locate(content []byte, pattern *regexp.Regexp) (first, matches int) {
	for n, line := range strings.Split(string(content), "\n") {
		if !pattern.MatchString(line) {
			continue
		}
		if matches == 0 {
			first = n + 1
		}
		matches++
	}
	return first, matches
}

// Render writes the rows as Markdown tables, one per group in the order the
// groups first appear, under a line naming the repository, the commit and its
// date. Every file proof is qualified by the commit.
func Render(rows []Row, stats Stats) string {
	var order []string
	byGroup := make(map[string][]*Row)
	for i := range rows {
		group := rows[i].Claim.Group
		if _, ok := byGroup[group]; !ok {
			order = append(order, group)
		}
		byGroup[group] = append(byGroup[group], &rows[i])
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Pinned commit: %s of %s, %s. Every fact below was computed at that commit.\n", code(stats.Commit), code(stats.Repository), cell(stats.Date))
	for _, group := range order {
		fmt.Fprintf(&b, "\n### %s\n\n", group)
		b.WriteString("| ID | Claim | Status | Evidence | Used in |\n|---|---|---|---|---|\n")
		for _, r := range byGroup[group] {
			proofs := make([]string, len(r.Proofs))
			for j, p := range r.Proofs {
				if p.Fact {
					proofs[j] = code(p.Place)
				} else {
					proofs[j] = code(p.Place + "@" + stats.Commit)
				}
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
				cell(r.Claim.ID), cell(r.Claim.Text), cell(r.Claim.Status), strings.Join(proofs, "<br>"), cell(r.Claim.UsedIn))
		}
	}
	return b.String()
}

// cell keeps a value inside one Markdown table cell.
func cell(s string) string {
	return strings.NewReplacer("|", `\|`, "\n", " ").Replace(s)
}

// code writes a value as inline code that stays inside its table cell: the
// fence is longer than any run of backticks in the value, so none of them can
// close it early.
func code(s string) string {
	fence := "`"
	for strings.Contains(s, fence) {
		fence += "`"
	}
	pad := ""
	if strings.HasPrefix(s, "`") || strings.HasSuffix(s, "`") {
		pad = " "
	}
	return fence + pad + cell(s) + pad + fence
}

// markers are the two lines that enclose a generated section.
func markers(section string) (begin, end string) {
	return "<!-- ledger:" + section + ":begin -->", "<!-- ledger:" + section + ":end -->"
}

// sectionBounds finds where a section's content starts and stops; both markers
// must appear exactly once and in order.
func sectionBounds(document, section string) (start, stop int, err error) {
	begin, end := markers(section)
	if strings.Count(document, begin) != 1 || strings.Count(document, end) != 1 {
		return 0, 0, fmt.Errorf("the ledger needs exactly one %q and one %q", begin, end)
	}
	start = strings.Index(document, begin) + len(begin)
	stop = strings.Index(document, end)
	if stop < start {
		return 0, 0, fmt.Errorf("%q comes before %q", end, begin)
	}
	return start, stop, nil
}

// Section returns what stands between the section's markers.
func Section(document, section string) (string, error) {
	start, stop, err := sectionBounds(document, section)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(document[start:stop], "\n"), nil
}

// ReplaceSection puts content between the section's markers.
func ReplaceSection(document, section, content string) (string, error) {
	start, stop, err := sectionBounds(document, section)
	if err != nil {
		return "", err
	}
	return document[:start] + "\n" + content + document[stop:], nil
}
