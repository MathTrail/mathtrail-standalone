// Command ledger renders the evidence ledger's table of claims: every claim in
// a claims file, with the line that proves it at one pinned commit, or with the
// number a script computed at that commit. The commit is the one the stats file
// names, so the table and the numbers always describe the same state of the
// system. The proofs are read from a local git repository: the product's own,
// or a clone of another public one.
//
// A claim fails the run when one of its patterns matches no line at the
// commit, or more than one, or when a fact differs from the value the claim
// states; moving the ledger to a later commit therefore names the claims to
// read again. A pattern that still matches exactly one line is taken to be the
// right line, which is why patterns quote enough of it to be unmistakable, and
// why the new table's diff is read when the commit moves.
//
// Usage:
//
//	ledger -claims <claims.json> -stats <stats.txt> -ledger <ledger.md> -section <name> [-numbers <file>] [-repo <git dir>] [-check]
//
// The table replaces what stands between the markers
// "<!-- ledger:<name>:begin -->" and "<!-- ledger:<name>:end -->" in the ledger.
// With -numbers, the numbers the claims state — each found in a line that
// proves its claim — are written to that file as key=value lines, for the
// paper's macros. With -check nothing is written: the run also fails when the
// ledger's section, or the numbers file, differs from what the claims would
// render.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "ledger:", err)
		os.Exit(1)
	}
}

func run() error {
	claimsPath := flag.String("claims", "", "the claims file")
	statsPath := flag.String("stats", "", "the key=value facts computed at the pinned commit")
	ledgerPath := flag.String("ledger", "", "the ledger file whose section is rendered")
	section := flag.String("section", "", "the name in the section's markers")
	repo := flag.String("repo", ".", "the git repository the proofs are read from")
	numbersPath := flag.String("numbers", "", "the file the claims' numbers are written to, if any")
	check := flag.Bool("check", false, "fail if the section or the numbers are out of date instead of rewriting them")
	flag.Parse()
	if *claimsPath == "" || *statsPath == "" || *ledgerPath == "" || *section == "" {
		return errors.New("-claims, -stats, -ledger and -section are all required")
	}

	claims, err := readClaims(*claimsPath)
	if err != nil {
		return err
	}
	stats, err := readStats(*statsPath)
	if err != nil {
		return err
	}
	rows, err := Resolve(claims, gitShow(context.Background(), *repo, stats.Commit), stats)
	if err != nil {
		return err
	}
	table := Render(rows, stats)
	numbers := RenderNumbers(rows, stats)

	if *check {
		if err := checkLedger(*ledgerPath, *section, table); err != nil {
			return err
		}
		if err := checkNumbers(*numbersPath, numbers); err != nil {
			return err
		}
		fmt.Printf("%d claims, every proof found at %s, and %s is up to date\n", len(rows), stats.Commit, *ledgerPath)
		return nil
	}
	if err := writeLedger(*ledgerPath, *section, table); err != nil {
		return err
	}
	if *numbersPath != "" {
		if err := replaceFile(filepath.Clean(*numbersPath), []byte(numbers), 0o644); err != nil {
			return fmt.Errorf("write numbers: %w", err)
		}
	}
	fmt.Printf("%d claims written to %s\n", len(rows), *ledgerPath)
	return nil
}

// checkNumbers fails when the numbers file, if one is named, is not what the
// claims render now.
func checkNumbers(path, numbers string) error {
	if path == "" {
		return nil
	}
	current, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("read numbers: %w", err)
	}
	if string(current) != numbers {
		return fmt.Errorf("%s is out of date: it differs from the numbers the claims state", path)
	}
	return nil
}

// checkLedger fails when the ledger's section is not the table the claims
// render now.
func checkLedger(path, section, table string) error {
	ledger, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("read ledger: %w", err)
	}
	current, err := Section(string(ledger), section)
	if err != nil {
		return err
	}
	if current != table {
		return fmt.Errorf("%s is out of date: its %q section differs from what the claims render", path, section)
	}
	return nil
}

// writeLedger puts the table into the ledger's section. The ledger keeps the
// permissions it had: this tool only rewrites a section.
func writeLedger(path, section, table string) error {
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("read ledger: %w", err)
	}
	ledger, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read ledger: %w", err)
	}
	updated, err := ReplaceSection(string(ledger), section, table)
	if err != nil {
		return err
	}
	if err := replaceFile(path, []byte(updated), info.Mode().Perm()); err != nil {
		return fmt.Errorf("write ledger: %w", err)
	}
	return nil
}

// replaceFile writes the content beside the file and renames it over the file,
// so a write that fails part-way leaves the file as it was.
func replaceFile(path string, content []byte, perm os.FileMode) error {
	next, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	// Once the rename has happened there is nothing left to remove.
	defer func() { _ = os.Remove(next.Name()) }()
	if _, err := next.Write(content); err != nil {
		_ = next.Close()
		return err
	}
	if err := next.Close(); err != nil {
		return err
	}
	if err := os.Chmod(next.Name(), perm); err != nil {
		return err
	}
	return os.Rename(next.Name(), path)
}

func readClaims(path string) ([]Claim, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read claims: %w", err)
	}
	return ParseClaims(data)
}

func readStats(path string) (Stats, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return Stats{}, fmt.Errorf("read stats: %w", err)
	}
	return ParseStats(string(data))
}

// gitShow reads files as they were at the commit, never from a working tree.
func gitShow(ctx context.Context, repo, commit string) Source {
	return func(path string) ([]byte, error) {
		out, err := exec.CommandContext(ctx, "git", "-C", repo, "show", commit+":"+path).Output() //nolint:gosec // G204: the program is fixed; the repository and the object name come from the person running the tool
		if err != nil {
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				return nil, fmt.Errorf("git show %s:%s: %s", commit, path, strings.TrimSpace(string(exit.Stderr)))
			}
			return nil, fmt.Errorf("git show %s:%s: %w", commit, path, err)
		}
		return out, nil
	}
}
