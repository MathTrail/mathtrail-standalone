// Command macros writes the file every number of the paper is taken from.
//
// A paper states no number typed by hand: each one is computed by a script
// over saved data and reaches the LaTeX source as a macro, \stat{source}{key},
// which fails the build when nothing defines it. This command reads the files
// of computed numbers — `key=value` lines, such as the facts the evidence
// ledger computes at a commit, or the results of an experiment — and writes a
// \statdef line for each.
//
// Usage:
//
//	macros -out <file.tex> [-identifying key,key] <name>=<path> [<name>=<path> ...]
//
// A value that would name the system to the reviewers of an anonymous build —
// a key -identifying lists, such as a commit, or a web address or the project's
// name in any value — is written as \statiddef instead, which that build hides.
//
// The output is written whole or not at all: a source that cannot be read, or
// a key given twice, leaves the last good file in place.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("macros", flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", "", "the LaTeX file to write")
	identifying := flags.String("identifying", "", "keys whose values the anonymous build hides, comma-separated")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if err := write(*out, keySet(*identifying), flags.Args(), stdout); err != nil {
		fmt.Fprintln(stderr, "macros:", err)
		return 1
	}
	return 0
}

func write(out string, identifying map[string]bool, args []string, stdout io.Writer) error {
	if out == "" || len(args) == 0 {
		return errors.New("usage: macros -out <file.tex> [-identifying key,key] <name>=<path> [<name>=<path> ...]")
	}
	all := make([]Numbers, 0, len(args))
	seen := make(map[string]bool)
	for _, arg := range args {
		source, err := ParseSource(arg)
		if err != nil {
			return err
		}
		if seen[source.Name] {
			return fmt.Errorf("%q names two sources", source.Name)
		}
		seen[source.Name] = true
		text, err := os.ReadFile(filepath.Clean(source.Path))
		if err != nil {
			return fmt.Errorf("read %s: %w", source.Path, err)
		}
		parsed, err := Parse(string(text))
		if err != nil {
			return fmt.Errorf("%s: %w", source.Path, err)
		}
		all = append(all, Numbers{Source: source, Values: parsed})
	}
	return replace(out, Render(all, identifying), stdout)
}

// keySet reads a comma-separated list of keys.
func keySet(list string) map[string]bool {
	keys := make(map[string]bool)
	for _, key := range strings.Split(list, ",") {
		if key = strings.TrimSpace(key); key != "" {
			keys[key] = true
		}
	}
	return keys
}

// replace writes the file beside its final place and moves it there, so that a
// build never reads half a file.
func replace(out, content string, stdout io.Writer) error {
	dir := filepath.Dir(out)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("make %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".macros-*")
	if err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write %s: %w", out, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	// A temporary file is made readable by its owner alone; the numbers are
	// read by whoever builds the paper, as any other source of it.
	if err := os.Chmod(tmp.Name(), 0o644); err != nil { //nolint:gosec // G302: a source of the paper, meant to be read by anyone who builds it
		return fmt.Errorf("write %s: %w", out, err)
	}
	if err := os.Rename(tmp.Name(), out); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	fmt.Fprintf(stdout, "%s written\n", out)
	return nil
}
