package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// write puts a file into a fresh directory and returns its path.
func write(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunFailsOnANumberTypedByHandAndNamesIt(t *testing.T) {
	t.Parallel()
	namesFile := write(t, "names.txt", "# names with digits\nRQ[12]\n")
	clean := write(t, "clean.tex", `RQ1 holds \stat{ea1}{cases} cases.`)
	typed := write(t, "typed.tex", "line one\nRQ2 has 42 cases.\n")
	var out, errs bytes.Buffer
	if code := run([]string{"-names", namesFile, clean}, &out, &errs); code != 0 {
		t.Errorf("run on a clean file = %d, want 0; %s", code, errs.String())
	}
	if code := run([]string{"-names", namesFile, clean, typed}, &out, &errs); code != 1 {
		t.Errorf("run with a number typed by hand = %d, want 1", code)
	}
	if want := typed + `:line 2: "42"`; !strings.Contains(out.String(), want) {
		t.Errorf("output %q does not name %q", out.String(), want)
	}
}

func TestRunRefusesNamesThatCouldLetEveryNumberThrough(t *testing.T) {
	t.Parallel()
	paper := write(t, "paper.tex", "no numbers")
	cases := []struct {
		name, names, want string
	}{
		{"a pattern that does not compile", "RQ[12\n", "missing closing ]"},
		{"a pattern that matches the empty string", "RQ[12]\n.*\n", "matches the empty string"},
		{"a pattern that matches a bare number", "RQ[12]\n[A-Z]?[0-9]+\n", "matches the bare number"},
		{"no names", "# only a comment\n", "names nothing"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var out, errs bytes.Buffer
			if code := run([]string{"-names", write(t, "names.txt", c.names), paper}, &out, &errs); code != 1 || !strings.Contains(errs.String(), c.want) {
				t.Errorf("run = %d, %q; want 1 and an error containing %q", code, errs.String(), c.want)
			}
		})
	}
}

// reported are the numbers the command printed, each as "line: number".
func reported(stdout string) []string {
	var found []string
	for _, m := range regexp.MustCompile(`:line (\d+): "([^"]*)"`).FindAllStringSubmatch(stdout, -1) {
		found = append(found, m[1]+": "+m[2])
	}
	return found
}

// A figure named with -labels is read for the words it prints: the
// coordinates that lay it out pass, and the count in a node is reported, with
// the file and the line. The same file read as a paper reports every number.
func TestRunReadsAFigureForWhatItPrints(t *testing.T) {
	t.Parallel()
	namesFile := write(t, "names.txt", "RQ[12]\n")
	paper := write(t, "paper.tex", "No numbers here.\n")
	figure := write(t, "fig.tex", `\draw (0,0) -- (3.3,1) node {7 tasks};`+"\n")
	var out, errs bytes.Buffer

	code := run([]string{"-names", namesFile, "-labels", figure, paper}, &out, &errs)

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	if code != 1 || len(lines) != 1 || !strings.HasPrefix(lines[0], figure+`:line 1: "7" in "`) {
		t.Errorf("run with the figure as -labels = %d, %q; want 1 and the count alone, on line 1", code, out.String())
	}
	if !strings.Contains(errs.String(), "handtyped: 1 numbers typed by hand") {
		t.Errorf("stderr = %q, want one number counted", errs.String())
	}

	out.Reset()
	errs.Reset()
	code = run([]string{"-names", namesFile, figure}, &out, &errs)

	if want := []string{"1: 0,0", "1: 3.3,1", "1: 7"}; code != 1 || !slices.Equal(reported(out.String()), want) {
		t.Errorf("run with the figure as a paper = %d, %q; want 1 and %q", code, reported(out.String()), want)
	}
}

// A file the check cannot read fails it, naming the file: a figure or a paper
// nobody read cannot be said to hold no number typed by hand, and a names file
// read in part is not the list of names.
func TestRunFailsOnAFileItCannotRead(t *testing.T) {
	t.Parallel()
	namesFile := write(t, "names.txt", "RQ[12]\n")
	paper := write(t, "paper.tex", "No numbers here.\n")
	missing := filepath.Join(t.TempDir(), "missing.tex")
	longLine := write(t, "names.txt", "RQ[12]\n"+strings.Repeat("x", 70_000)+"\n")
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"the names", []string{"-names", missing, paper}, missing},
		{"a paper", []string{"-names", namesFile, missing}, missing},
		{"a figure", []string{"-names", namesFile, "-labels", missing, paper}, missing},
		{"a names file with a line too long to read", []string{"-names", longLine, paper}, "token too long"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var out, errs bytes.Buffer

			code := run(tc.args, &out, &errs)

			if code != 1 || !strings.Contains(errs.String(), "handtyped: ") || !strings.Contains(errs.String(), tc.want) {
				t.Errorf("run = %d, stderr %q; want 1 and %q", code, errs.String(), tc.want)
			}
		})
	}
}

// A call that names no names file or no paper is a mistake in the call, told
// apart from numbers found by its exit status, and answered with the usage.
func TestRunAnswersAWrongCallWithTheUsage(t *testing.T) {
	t.Parallel()
	namesFile := write(t, "names.txt", "RQ[12]\n")
	paper := write(t, "paper.tex", "No numbers here.\n")
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"nothing", nil, "usage: handtyped -names"},
		{"no names", []string{paper}, "usage: handtyped -names"},
		{"empty names", []string{"-names", "", paper}, "usage: handtyped -names"},
		{"no paper", []string{"-names", namesFile}, "usage: handtyped -names"},
		{"a figure but no paper", []string{"-names", namesFile, "-labels", paper}, "usage: handtyped -names"},
		{"an unknown flag", []string{"-names", namesFile, "-verbose", paper}, "flag provided but not defined: -verbose"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var out, errs bytes.Buffer

			code := run(tc.args, &out, &errs)

			if code != 2 || !strings.Contains(errs.String(), tc.want) || out.Len() != 0 {
				t.Errorf("run(%q) = %d, stdout %q, stderr %q; want 2, nothing on stdout and %q", tc.args, code, out.String(), errs.String(), tc.want)
			}
		})
	}
}
