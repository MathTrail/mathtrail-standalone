package main

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

// files is a Source over an in-memory set of files, standing in for a commit.
func files(contents map[string]string) Source {
	return func(path string) ([]byte, error) {
		content, ok := contents[path]
		if !ok {
			return nil, errors.New("no such file at the commit")
		}
		return []byte(content), nil
	}
}

// pinned is a set of stats at a fixed commit with the given facts.
func pinned(values map[string]string) Stats {
	return Stats{Repository: "github.com/example/product", Commit: "e1c315303bf6", Date: "2026-09-25", Values: values}
}

func TestLocateFindsTheFirstMatchingLineAndCountsTheRest(t *testing.T) {
	t.Parallel()
	content := []byte("package rating\n\nconst Guess = 0.2\nconst Guess = 0.3\n")
	cases := []struct {
		name    string
		pattern string
		line    int
		matches int
	}{
		{"two matches", `^const Guess`, 3, 2},
		{"one match", `^package`, 1, 1},
		{"no match", `^func Update`, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			line, matches := Locate(content, regexp.MustCompile(c.pattern))
			if line != c.line || matches != c.matches {
				t.Errorf("Locate(%q) = %d, %d; want %d, %d", c.pattern, line, matches, c.line, c.matches)
			}
		})
	}
}

func TestResolveRefusesAProofThatMatchesMoreThanOneLine(t *testing.T) {
	t.Parallel()
	claims := []Claim{{ID: "C001", Evidence: []Anchor{{File: "a.go", Match: "^const"}}}}
	_, err := Resolve(claims, files(map[string]string{"a.go": "const X = 1\nconst Y = 2\n"}), pinned(nil))
	if err == nil || !strings.Contains(err.Error(), "2 lines of a.go match") {
		t.Errorf("Resolve error = %v, want one saying 2 lines of a.go match", err)
	}
}

func TestParseClaimsAcceptsEveryKindOfProof(t *testing.T) {
	t.Parallel()
	json := `[
		{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"file":"go.mod","match":"^module"}]},
		{"id":"C2","group":"G","claim":"b","status":"measured","evidence":[{"stat":"topics","expect":"17"}]},
		{"id":"C3","group":"H","claim":"c","status":"remark","evidence":[{"stat":"skills"}],"used_in":"A"}
	]`
	claims, err := ParseClaims([]byte(json))
	if err != nil {
		t.Fatalf("ParseClaims: %v", err)
	}
	if len(claims) != 3 || claims[1].Evidence[0].Expect != "17" {
		t.Errorf("ParseClaims = %+v, want three claims, the second expecting 17", claims)
	}
}

func TestParseClaimsAcceptsEveryStatusTheLedgerDefines(t *testing.T) {
	t.Parallel()
	for _, status := range []string{"built", "specified", "planned", "measured", "unverified", "remark", "context"} {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			json := `[{"id":"C1","group":"G","claim":"a","status":"` + status + `","evidence":[{"stat":"x"}]}]`
			if _, err := ParseClaims([]byte(json)); err != nil {
				t.Errorf("ParseClaims with status %q: %v", status, err)
			}
		})
	}
}

func TestParseClaimsRejectsClaimsThatCannotBeResolved(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		json string
		want string
	}{
		{"duplicate id", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"stat":"x"}]},{"id":"C1","group":"G","claim":"b","status":"built","evidence":[{"stat":"x"}]}]`, "id used twice"},
		{"missing id", `[{"group":"G","claim":"a","status":"built","evidence":[{"stat":"x"}]}]`, "no id"},
		{"missing group", `[{"id":"C1","claim":"a","status":"built","evidence":[{"stat":"x"}]}]`, "no group"},
		{"missing text", `[{"id":"C1","group":"G","claim":" ","status":"built","evidence":[{"stat":"x"}]}]`, "no claim text"},
		{"unknown status", `[{"id":"C1","group":"G","claim":"a","status":"done","evidence":[{"stat":"x"}]}]`, `unknown status "done"`},
		{"no evidence", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[]}]`, "no evidence"},
		{"file without a match", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"file":"go.mod"}]}]`, "must be a file with a match, or a stat"},
		{"both a file and a stat", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"file":"go.mod","match":"x","stat":"y"}]}]`, "must be a file with a match, or a stat"},
		{"a stat with a pattern", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"stat":"x","match":"y"}]}]`, "must be a file with a match, or a stat"},
		{"a stat with a file", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"stat":"x","file":"go.mod"}]}]`, "must be a file with a match, or a stat"},
		{"expected value on a file", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"file":"go.mod","match":"x","expect":"1"}]}]`, "expects a value but is not a stat"},
		{"pattern that does not compile", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"file":"go.mod","match":"GET("}]}]`, `pattern "GET("`},
		{"unknown field", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"stat":"x"}],"source":"?"}]`, "unknown field"},
		{"claims after the list", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"stat":"x"}]}][{"id":"C2"}]`, "something follows the list"},
		{"a number key the macros cannot read", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"stat":"x"}],"numbers":{"Pairs":"33"}}]`, `number key "Pairs"`},
		{"a number with no value", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"stat":"x"}],"numbers":{"pairs":" "}}]`, "number pairs has no value"},
		{"a number key two claims state", `[{"id":"C1","group":"G","claim":"a","status":"built","evidence":[{"stat":"x"}],"numbers":{"pairs":"33"}},{"id":"C2","group":"G","claim":"b","status":"built","evidence":[{"stat":"x"}],"numbers":{"pairs":"34"}}]`, "number pairs is C1's already"},
		{"no list", `null`, "no claims"},
		{"empty list", `[]`, "no claims"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseClaims([]byte(c.json))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("ParseClaims error = %v, want one containing %q", err, c.want)
			}
		})
	}
}

func TestParseStatsTakesTheCommitFromTheFacts(t *testing.T) {
	t.Parallel()
	stats, err := ParseStats("repository=github.com/example/product\n commit = e1c315303bf6 \ncommit_date=2026-09-25\n\ntopics =17\n")
	if err != nil {
		t.Fatalf("ParseStats: %v", err)
	}
	if stats.Repository != "github.com/example/product" || stats.Commit != "e1c315303bf6" || stats.Date != "2026-09-25" {
		t.Errorf("ParseStats repository, commit, date = %q, %q, %q; want github.com/example/product, e1c315303bf6, 2026-09-25", stats.Repository, stats.Commit, stats.Date)
	}
	if got := stats.Values["topics"]; got != "17" {
		t.Errorf(`Values["topics"] = %q, want "17"`, got)
	}
}

func TestParseStatsRefusesFactsWithoutAPinnedCommit(t *testing.T) {
	t.Parallel()
	const repo = "repository=github.com/example/product\n"
	cases := []struct {
		name string
		text string
		want string
	}{
		{"no repository", "commit=e1c315303bf6\ncommit_date=2026-09-25\n", "no repository"},
		{"no commit", repo + "commit_date=2026-09-25\ntopics=17\n", "no commit"},
		{"short commit", repo + "commit=e1c3153\ncommit_date=2026-09-25\n", "no commit"},
		{"a branch, not a commit", repo + "commit=main\ncommit_date=2026-09-25\n", "no commit"},
		{"no date", repo + "commit=e1c315303bf6\ntopics=17\n", "no commit_date"},
		{"line without a value", repo + "commit=e1c315303bf6\ncommit_date=2026-09-25\ntopics\n", "no '='"},
		{"key given twice", repo + "commit=e1c315303bf6\ncommit_date=2026-09-25\ncommit=0123456789ab\n", `"commit" is given twice`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseStats(c.text)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("ParseStats error = %v, want one containing %q", err, c.want)
			}
		})
	}
}

func TestResolveReportsEveryMissingProofAtOnce(t *testing.T) {
	t.Parallel()
	claims := []Claim{
		{ID: "C001", Evidence: []Anchor{{File: "a.go", Match: "^const X"}}},
		{ID: "C002", Evidence: []Anchor{{File: "a.go", Match: "^const Y"}}},
		{ID: "C003", Evidence: []Anchor{{File: "b.go", Match: "x"}}},
		{ID: "C004", Evidence: []Anchor{{Stat: "missing"}}},
	}
	rows, err := Resolve(claims, files(map[string]string{"a.go": "package a\nconst X = 1\n"}), pinned(nil))
	if err == nil {
		t.Fatal("Resolve found every proof, want errors for C002, C003 and C004")
	}
	for _, id := range []string{"C002", "C003", "C004"} {
		if !strings.Contains(err.Error(), id) {
			t.Errorf("Resolve error does not name %s: %v", id, err)
		}
	}
	if got := rows[0].Proofs; len(got) != 1 || got[0] != (Proof{Place: "a.go:2", Text: "const X = 1"}) {
		t.Errorf("proof of C001 = %+v, want [{Place:a.go:2 Text:const X = 1}]", got)
	}
}

// facts are the stats the tests of fact anchors resolve against.
var facts = pinned(map[string]string{"topics": "17", "traps": "", "skills": "25"})

func TestResolveProvesAFactByItsValue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		anchor Anchor
		proof  Proof
	}{
		{"value as stated", Anchor{Stat: "topics", Expect: "17"}, Proof{Place: "topics = 17", Fact: true, Text: "17"}},
		{"no value stated", Anchor{Stat: "skills"}, Proof{Place: "skills = 25", Fact: true, Text: "25"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			rows, err := Resolve([]Claim{{ID: "C001", Evidence: []Anchor{c.anchor}}}, files(nil), facts)
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if got := rows[0].Proofs; len(got) != 1 || got[0] != c.proof {
				t.Errorf("proofs = %+v, want [%+v]", got, c.proof)
			}
		})
	}
}

func TestResolveRefusesAFactThatIsEmptyOrNotWhatTheClaimSays(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		anchor Anchor
		want   string
	}{
		{"value changed", Anchor{Stat: "topics", Expect: "18"}, `fact "topics" is 17, the claim says 18`},
		{"empty value", Anchor{Stat: "traps"}, `no value for fact "traps"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := Resolve([]Claim{{ID: "C001", Evidence: []Anchor{c.anchor}}}, files(nil), facts)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("Resolve error = %v, want one containing %q", err, c.want)
			}
		})
	}
}

func TestRenderQualifiesLinesByCommitAndKeepsCellsWhole(t *testing.T) {
	t.Parallel()
	rows := []Row{
		{Claim: Claim{ID: "C001", Group: "Model", Text: "P = 0.2 + 0.8·σ(x) | floor", Status: "built", UsedIn: "A"}, Proofs: []Proof{{Place: "rating.go:28"}, {Place: "topics = 17", Fact: true}}},
	}
	got := Render(rows, pinned(nil))
	for _, want := range []string{"Pinned commit: `e1c315303bf6` of `github.com/example/product`, 2026-09-25.", "### Model", "`rating.go:28@e1c315303bf6`", "`topics = 17`", `σ(x) \| floor`} {
		if !strings.Contains(got, want) {
			t.Errorf("Render output lacks %q:\n%s", want, got)
		}
	}
}

func TestRenderKeepsPipesAndBackticksInsideTheirCells(t *testing.T) {
	t.Parallel()
	rows := []Row{
		{Claim: Claim{ID: "C|1", Group: "G", Text: "a", Status: "built|x", UsedIn: "A"}, Proofs: []Proof{{Place: "k = a|b`c", Fact: true}, {Place: "odd`name.go:3"}}},
	}
	got := Render(rows, pinned(nil))
	var row string
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "| C") {
			row = line
		}
	}
	// Five cells have six separators; an escaped pipe is not one.
	if n := strings.Count(row, "|") - strings.Count(row, `\|`); n != 6 {
		t.Errorf("row has %d separators, want 6: %s", n, row)
	}
	for _, want := range []string{"``k = a\\|b`c``", "``odd`name.go:3@e1c315303bf6``"} {
		if !strings.Contains(row, want) {
			t.Errorf("row lacks %s: %s", want, row)
		}
	}
}

func TestRenderGivesEveryGroupOneTableInOrderOfFirstAppearance(t *testing.T) {
	t.Parallel()
	rows := []Row{
		{Claim: Claim{ID: "C001", Group: "Model"}},
		{Claim: Claim{ID: "C002", Group: "Checks"}},
		{Claim: Claim{ID: "C003", Group: "Model"}},
	}
	got := Render(rows, pinned(nil))
	if n := strings.Count(got, "### Model"); n != 1 {
		t.Errorf("Render wrote %d headings for Model, want 1:\n%s", n, got)
	}
	model, checks := strings.Index(got, "### Model"), strings.Index(got, "### Checks")
	first, third := strings.Index(got, "| C001 |"), strings.Index(got, "| C003 |")
	if model > first || first > third || third > checks {
		t.Errorf("Render order: want Model with C001 and C003, then Checks:\n%s", got)
	}
}

func TestReplaceSectionNeedsBothMarkersOnceAndInOrder(t *testing.T) {
	t.Parallel()
	const begin, end = "<!-- ledger:product:begin -->", "<!-- ledger:product:end -->"
	cases := []struct {
		name     string
		document string
		want     string
		fails    bool
	}{
		{"replaced", "x\n" + begin + "\nold\n" + end + "\ny", "x\n" + begin + "\nnew\n" + end + "\ny", false},
		{"no markers", "x", "", true},
		{"end before begin", end + begin, "", true},
		{"marker twice", begin + end + begin + end, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := ReplaceSection(c.document, "product", "new\n")
			if c.fails {
				if err == nil {
					t.Errorf("ReplaceSection = %q, want an error", got)
				}
				return
			}
			if err != nil || got != c.want {
				t.Errorf("ReplaceSection = %q, %v; want %q", got, err, c.want)
			}
		})
	}
}

func TestSectionReadsBackWhatReplaceSectionWrote(t *testing.T) {
	t.Parallel()
	document := "x\n<!-- ledger:product:begin -->\nold\n<!-- ledger:product:end -->\ny"
	updated, err := ReplaceSection(document, "product", "new\n")
	if err != nil {
		t.Fatalf("ReplaceSection: %v", err)
	}
	for _, c := range []struct {
		document, want string
	}{
		{updated, "new\n"},
		{document, "old\n"},
	} {
		got, err := Section(c.document, "product")
		if err != nil || got != c.want {
			t.Errorf("Section = %q, %v; want %q", got, err, c.want)
		}
	}
}

// A number a claim states reaches the paper only when it stands in a line or a
// fact that proves the claim, as a whole number rather than a part of another.
func TestResolveFindsEachNumberInAProofOfItsClaim(t *testing.T) {
	t.Parallel()
	source := files(map[string]string{"spec.md": "Over 62,653 pairs the sketches decide 33 pairs differently, all within 0.1 of it.\n"})
	evidence := []Anchor{{File: "spec.md", Match: "^Over"}, {Stat: "topics"}}
	cases := []struct {
		name    string
		numbers map[string]string
		want    string // what the error says, or nothing for a number that stands
	}{
		{"in the line and in the fact", map[string]string{"pairs": "62,653", "differ": "33", "within": "0.1", "topics": "17"}, ""},
		{"in no proof", map[string]string{"pairs": "62,654"}, "number pairs = 62,654 stands in none of its proofs"},
		{"a part of a longer number", map[string]string{"part": "653"}, "number part = 653 stands in none"},
		{"a part of a decimal", map[string]string{"part": "1"}, "number part = 1 stands in none"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := Resolve([]Claim{{ID: "C001", Evidence: evidence, Numbers: c.numbers}}, source, facts)
			switch {
			case c.want == "" && err != nil:
				t.Errorf("Resolve: %v, want every number to stand", err)
			case c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)):
				t.Errorf("Resolve error = %v, want one containing %q", err, c.want)
			}
		})
	}
}

// The numbers file names the commit, lists each claim's numbers under its id
// in the order of their keys, and leaves out a claim that states none.
func TestRenderNumbersListsEachClaimsNumbersUnderItsID(t *testing.T) {
	t.Parallel()
	rows := []Row{
		{Claim: Claim{ID: "C001", Numbers: map[string]string{"within": "0.1", "pairs": "62,653"}}},
		{Claim: Claim{ID: "C002"}},
		{Claim: Claim{ID: "C003", Numbers: map[string]string{"tasks": "15"}}},
	}
	got := RenderNumbers(rows, facts)
	want := "# Numbers the claims of the evidence ledger state, each found in a line that proves its claim at e1c315303bf6.\n" +
		"# C001\npairs=62,653\nwithin=0.1\n# C003\ntasks=15\n"
	if got != want {
		t.Errorf("RenderNumbers =\n%s\nwant\n%s", got, want)
	}
}
