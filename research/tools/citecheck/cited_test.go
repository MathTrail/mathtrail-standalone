package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCitedKeysReadsWhatLaTeXRecorded(t *testing.T) {
	t.Parallel()
	aux := `\relax
\citation{polya2014solve}
\citation{schoenfeld2016learning,polya2014solve}
\@writefile{toc}{\contentsline {section}{\numberline {1}Introduction}{1}{}\protected@file@percent }
\citation{rasch1960}
\citation{*}
\bibdata{refs,web}
\bibcite{polya2014solve}{1}`
	got := CitedKeys(aux)
	want := []string{"polya2014solve", "schoenfeld2016learning", "rasch1960", "*"}
	if !slices.Equal(got, want) {
		t.Errorf("CitedKeys = %v, want %v", got, want)
	}
}

func TestCitedKeysTakesOnlyCitationLines(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, aux string
	}{
		{"a bibliography's own record", `\bibcite{hidden}{1}`},
		{"a label", `\newlabel{sec:citation}{{1}{1}}`},
		{"no citation at all", `\relax`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := CitedKeys(c.aux); len(got) != 0 {
				t.Errorf("CitedKeys(%q) = %v, want none", c.aux, got)
			}
		})
	}
}

func TestCitedFailsOnAKeyNoBibliographyHolds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write := func(name, content string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	refs := write("refs.bib", "@article{checked2020work,\n  title = {A Work},\n}\n")
	web := write("web.bib", "@misc{official2024law,\n  title = {A Law},\n}\n")
	good := write("good.aux", "\\citation{checked2020work}\n\\citation{official2024law}\n\\citation{*}\n")
	lead := write("lead.aux", "\\citation{checked2020work,arxiv2512.23036}\n")

	var out bytes.Buffer
	code := run(context.Background(), []string{"-bib", refs, "-also", web, "cited", good}, Registries{}, &out, &out)
	if code != 0 || !strings.Contains(out.String(), "3 cited keys in 1 files") {
		t.Errorf("cited on keys both files hold = %d, %q; want 0 and a count", code, out.String())
	}

	out.Reset()
	code = run(context.Background(), []string{"-bib", refs, "-also", web, "cited", good, lead}, Registries{}, &out, &out)
	if code != 1 || !strings.Contains(out.String(), "arxiv2512.23036") {
		t.Errorf("cited on a lead = %d, %q; want 1 and the lead's key named", code, out.String())
	}

	out.Reset()
	code = run(context.Background(), []string{"-bib", refs, "cited", good}, Registries{}, &out, &out)
	if code != 1 || !strings.Contains(out.String(), "official2024law") {
		t.Errorf("cited without -also = %d, %q; want 1: web.bib's key is unknown without it", code, out.String())
	}

	out.Reset()
	code = run(context.Background(), []string{"-bib", refs, "cited", filepath.Join(dir, "missing.aux")}, Registries{}, &out, &out)
	if code != 2 {
		t.Errorf("cited on a file that is not there = %d, want 2", code)
	}
}
