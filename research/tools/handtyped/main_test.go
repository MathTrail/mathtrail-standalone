package main

import (
	"bytes"
	"os"
	"path/filepath"
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
