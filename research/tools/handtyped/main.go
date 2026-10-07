// Command handtyped finds numbers typed by hand in a paper's LaTeX. Every
// number a paper states comes from a computed file through a macro, so a
// number left in its text is one nobody computed, which can drift from the
// data without anything noticing. A number the text chooses rather than
// reports — a point a function is evaluated at, a confidence level — is
// marked with \given and passes. The names that carry digits of their own, a
// research question or a cipher, come from a file: one regular expression a
// line, a line starting with # a comment.
//
// A figure given with -labels is read with its layout blanked out — the
// coordinates, the option values, the macros it defines — so that what is
// checked is what it prints.
//
// Usage:
//
//	handtyped -names <file> [-labels <figure.tex> ...] <file.tex> [<file.tex> ...]
//
// It prints every number typed by hand with its file and line, and fails if
// there is one.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("handtyped", flag.ContinueOnError)
	flags.SetOutput(stderr)
	namesPath := flags.String("names", "", "the names that carry digits, one regular expression a line")
	var figures []string
	flags.Func("labels", "a figure read for its words alone; may be given more than once", func(path string) error {
		figures = append(figures, path)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *namesPath == "" || flags.NArg() == 0 {
		fmt.Fprintln(stderr, "usage: handtyped -names <file> [-labels <figure.tex> ...] <file.tex> [<file.tex> ...]")
		return 2
	}
	count, err := check(*namesPath, figures, flags.Args(), stdout)
	if err != nil {
		fmt.Fprintln(stderr, "handtyped:", err)
		return 1
	}
	if count > 0 {
		fmt.Fprintf(stderr, "handtyped: %d numbers typed by hand: print each through a macro of computed numbers, or mark one the text chooses with \\given\n", count)
		return 1
	}
	return 0
}

// check prints every number typed by hand in the figures, read for their words
// alone, and in the papers, each with its file and line, and counts them.
func check(namesPath string, figures, papers []string, stdout io.Writer) (int, error) {
	names, err := readNames(namesPath)
	if err != nil {
		return 0, err
	}
	count := 0
	read := func(path string, words func(string) string) error {
		text, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return err
		}
		for _, f := range Find(words(string(text)), names) {
			fmt.Fprintf(stdout, "%s:%s\n", path, f)
			count++
		}
		return nil
	}
	for _, path := range figures {
		if err := read(path, FigureWords); err != nil {
			return 0, err
		}
	}
	whole := func(text string) string { return text }
	for _, path := range papers {
		if err := read(path, whole); err != nil {
			return 0, err
		}
	}
	return count, nil
}

// bareNumbers are numbers as a paper writes them, with no name around them; a
// pattern of the names file that matches one would hide such numbers.
var bareNumbers = []string{"0", "7", "42", "3.5", "0.775", "2026", "23,453"}

// readNames reads the names that carry digits, refusing a pattern that does
// not compile, one that matches the empty string, and one that matches a bare
// number, any of which would let numbers typed by hand through.
func readNames(path string) ([]*regexp.Regexp, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	var names []*regexp.Regexp
	scanner := bufio.NewScanner(file)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pattern, err := regexp.Compile(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, n, err)
		}
		if pattern.MatchString("") {
			return nil, fmt.Errorf("%s:%d: %q matches the empty string, which would let every number through", path, n, line)
		}
		if bare := slices.IndexFunc(bareNumbers, pattern.MatchString); bare >= 0 {
			return nil, fmt.Errorf("%s:%d: %q matches the bare number %s: a name carries letters, or it would hide numbers typed by hand", path, n, line, bareNumbers[bare])
		}
		names = append(names, pattern)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New(path + " names nothing")
	}
	return names, nil
}
