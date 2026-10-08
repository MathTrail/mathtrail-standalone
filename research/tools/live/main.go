// Command live writes the numbers of the service's real use that paper A
// prints, from the monthly snapshot of its public totals.
//
// The snapshot is the one the site's page "Research" draws: the latest month
// counted whole, with its total once ten children or more stand behind it. The
// paper prints that month and its total, so this command reads them and writes
// them as `key=value` lines, a source of the paper's numbers like any other:
//
//	state=0                                          no month counted whole yet, or no snapshot
//	state=1, year, month                             a month whose children were too few to show
//	state=2, year, month, learners, answers,
//	         promised_mean, correct_share            a month whose total is shown
//
// A snapshot that is not there is a month not yet counted. One that holds a
// key this command does not know, leaves one out, or shows a total the public
// views would hide is refused, and the file is left as it was.
//
// Usage:
//
//	live -snapshot <live.json> -out <file>
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("live", flag.ContinueOnError)
	flags.SetOutput(stderr)
	snapshotPath := flags.String("snapshot", "", "the monthly snapshot of the service's public totals")
	out := flags.String("out", "", "the file of key=value lines to write")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *snapshotPath == "" || *out == "" || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: live -snapshot <live.json> -out <file>")
		return 2
	}
	numbers, err := read(*snapshotPath)
	if err != nil {
		fmt.Fprintln(stderr, "live:", err)
		return 1
	}
	if err := write(*out, Render(numbers)); err != nil {
		fmt.Fprintln(stderr, "live:", err)
		return 1
	}
	fmt.Fprintf(stdout, "%s written\n", *out)
	return 0
}

// read parses the snapshot at path. One that is not there is a month not yet
// counted, as the site's page reads it.
func read(path string) (Numbers, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if errors.Is(err, fs.ErrNotExist) {
		return Numbers{State: stateComing}, nil
	}
	if err != nil {
		return Numbers{}, fmt.Errorf("read %s: %w", path, err)
	}
	numbers, err := Parse(data)
	if err != nil {
		return Numbers{}, fmt.Errorf("%s: %w", path, err)
	}
	return numbers, nil
}

// write puts the lines in the file, making its folder if need be. It runs only
// after the snapshot was read whole, so a refused snapshot changes nothing.
func write(out, text string) error {
	if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
		return fmt.Errorf("make %s: %w", filepath.Dir(out), err)
	}
	if err := os.WriteFile(filepath.Clean(out), []byte(text), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	return nil
}
