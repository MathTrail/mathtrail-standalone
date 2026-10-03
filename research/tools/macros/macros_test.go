package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseReadsKeysAndValues(t *testing.T) {
	t.Parallel()
	got, err := Parse("# computed at a commit\n\ncommit=52ce86908135\ncoverage = 82.1%\n")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got["commit"] != "52ce86908135" || got["coverage"] != "82.1%" || len(got) != 2 {
		t.Errorf("Parse = %v, want commit and coverage", got)
	}
}

func TestParseRefusesWhatWouldMakeAMacroMeanTwoThings(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, text, want string
	}{
		{"a line with no =", "commit 52ce", "no '='"},
		{"a key with capitals", "Commit=1", "not a key"},
		{"a key with a hyphen", "grade-level=1", "not a key"},
		{"a key with no value", "commit=", "no value"},
		{"a key given twice", "a=1\na=2", "given twice"},
		{"nothing at all", "# only a comment\n", "no numbers"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(c.text)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("Parse(%q) error = %v, want one containing %q", c.text, err, c.want)
			}
		})
	}
}

func TestRenderEscapesWhatTeXWouldReadAsCommands(t *testing.T) {
	t.Parallel()
	got := Render([]Numbers{{
		Source: Source{Name: "product", Path: "stats.txt"},
		Values: map[string]string{"coverage": "82.1%", "odd": `a_b&c#d$e{f}g~h^i\j`},
	}}, nil)
	for _, want := range []string{
		`\statdef{product}{coverage}{82.1\%}`,
		`\statdef{product}{odd}{a\_b\&c\#d\$e\{f\}g\textasciitilde{}h\textasciicircum{}i\textbackslash{}j}`,
	} {
		if !strings.Contains(got, want+"\n") {
			t.Errorf("Render is missing the line %s; got:\n%s", want, got)
		}
	}
}

func TestRenderIsTheSameForTheSameNumbers(t *testing.T) {
	t.Parallel()
	all := []Numbers{
		{Source: Source{Name: "b", Path: "b.txt"}, Values: map[string]string{"z": "1", "y": "2", "x": "3"}},
		{Source: Source{Name: "a", Path: "a.txt"}, Values: map[string]string{"k": "4"}},
	}
	first := Render(all, nil)
	for range 20 {
		if again := Render(all, nil); again != first {
			t.Fatalf("Render changed between runs:\n%s\nthen\n%s", first, again)
		}
	}
	// The sources keep their order, and the keys of each come sorted.
	start := strings.Index(first, `\statdef`)
	if start < 0 {
		t.Fatalf("Render wrote no \\statdef line:\n%s", first)
	}
	body := first[start:]
	want := "\\statdef{b}{x}{3}\n\\statdef{b}{y}{2}\n\\statdef{b}{z}{1}\n\\statdef{a}{k}{4}\n"
	if body != want {
		t.Errorf("Render body =\n%s\nwant\n%s", body, want)
	}
}

func TestParseSourceHoldsTheNameToWhatAControlSequenceCanCarry(t *testing.T) {
	t.Parallel()
	cases := []struct {
		arg  string
		want Source
		ok   bool
	}{
		{"product=evidence/product-stats.txt", Source{Name: "product", Path: "evidence/product-stats.txt"}, true},
		{"faultinject2=results/summary.txt", Source{Name: "faultinject2", Path: "results/summary.txt"}, true},
		{"Product=x.txt", Source{}, false},
		{"fault_inject=x.txt", Source{}, false},
		{"product=", Source{}, false},
		{"product", Source{}, false},
	}
	for _, c := range cases {
		t.Run(c.arg, func(t *testing.T) {
			t.Parallel()
			got, err := ParseSource(c.arg)
			if (err == nil) != c.ok || got != c.want {
				t.Errorf("ParseSource(%q) = %+v, %v; want %+v, ok %v", c.arg, got, err, c.want, c.ok)
			}
		})
	}
}

func TestRunWritesTheFileWholeOrLeavesTheLastOne(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stats := filepath.Join(dir, "stats.txt")
	out := filepath.Join(dir, "generated", "numbers.tex")
	if err := os.WriteFile(stats, []byte("reference_tasks=603\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-out", out, "product=" + stats}, &stdout, &stderr); code != 0 {
		t.Fatalf("run = %d, stderr %s", code, stderr.String())
	}
	good, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(good), `\statdef{product}{reference_tasks}{603}`) {
		t.Fatalf("written file lacks the number:\n%s", good)
	}
	info, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o644 {
		t.Errorf("the numbers are written with mode %o, want 644: whoever builds the paper reads them", mode)
	}

	// A source that breaks the rules leaves the last good file as it was.
	if werr := os.WriteFile(stats, []byte("reference_tasks=603\nreference_tasks=604\n"), 0o600); werr != nil {
		t.Fatal(werr)
	}
	stderr.Reset()
	if code := run([]string{"-out", out, "product=" + stats}, &stdout, &stderr); code != 1 {
		t.Fatalf("run with a key given twice = %d, want 1", code)
	}
	after, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, good) {
		t.Errorf("a failed run changed the file:\n%s\nwant\n%s", after, good)
	}
	if !strings.Contains(stderr.String(), "given twice") {
		t.Errorf("stderr = %q, want it to name the key given twice", stderr.String())
	}
}

func TestRunRefusesTwoSourcesUnderOneName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stats := filepath.Join(dir, "stats.txt")
	if err := os.WriteFile(stats, []byte("a=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"-out", filepath.Join(dir, "n.tex"), "p=" + stats, "p=" + stats}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "names two sources") {
		t.Errorf("run = %d, stderr %q; want 1 and a word about the name given twice", code, stderr.String())
	}
}

func TestRenderMarksWhatWouldNameTheSystemToAReviewer(t *testing.T) {
	t.Parallel()
	got := Render([]Numbers{{
		Source: Source{Name: "product", Path: "stats.txt"},
		Values: map[string]string{
			"commit":     "52ce86908135",
			"repository": "github.com/MathTrail/mathtrail-standalone",
			"run_url":    "https://example.org/runs/123",
			"project":    "MathTrail",
			"topics":     "17",
		},
	}}, map[string]bool{"commit": true})
	for _, want := range []string{
		`\statiddef{product}{commit}{52ce86908135}`,
		`\statiddef{product}{repository}{github.com/MathTrail/mathtrail-standalone}`,
		`\statiddef{product}{run_url}{https://example.org/runs/123}`,
		`\statiddef{product}{project}{MathTrail}`,
		`\statdef{product}{topics}{17}`,
	} {
		if !strings.Contains(got, want+"\n") {
			t.Errorf("Render is missing the line %s; got:\n%s", want, got)
		}
	}
}
