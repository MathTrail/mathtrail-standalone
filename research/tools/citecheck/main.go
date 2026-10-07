// Command citecheck keeps the research's bibliography honest: every entry of
// refs.bib must describe a work a registry actually holds, under the title,
// authors, year and venue the registry records, and no retracted work may be
// cited as sound.
//
// Usage:
//
//	citecheck [-bib literature/refs.bib] check
//	citecheck [-bib literature/refs.bib] add <DOI | arXiv:id>...
//	citecheck [-bib literature/refs.bib] [-also literature/web.bib] cited <file.aux>...
//
// check looks every entry up — by its DOI, or by the DOI arXiv registers for
// its preprint — and lists what disagrees. add writes new entries from the
// registries' records alone, so no field of the bibliography is typed from
// memory, and refuses a work that is already there or that a notice puts in
// doubt. cited asks no registry: it reads the citations LaTeX recorded in
// its .aux files and fails when one names a key that neither the bibliography
// nor the files named by -also holds, so a lead read by nobody cannot reach a
// paper's build.
//
// Crossref is asked first and DataCite for what Crossref does not hold: arXiv
// and Zenodo register their DOIs with DataCite. The exit status is 0 when every
// entry checks out, 1 when one does not, and 2 when the command could not do
// its job: a file it cannot read, or a registry that does not answer.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	registries := Registries{
		Client:    &http.Client{Timeout: 30 * time.Second},
		Crossref:  "https://api.crossref.org",
		DataCite:  "https://api.datacite.org",
		UserAgent: "mathtrail-research-citecheck (https://github.com/MathTrail/mathtrail-standalone)",
	}
	os.Exit(run(context.Background(), os.Args[1:], registries, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, registries Registries, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("citecheck", flag.ContinueOnError)
	flags.SetOutput(stderr)
	bibPath := flags.String("bib", "literature/refs.bib", "the bibliography to check or add to")
	also := flags.String("also", "", "more bibliographies a cited key may be found in, comma-separated; they are not checked against registries")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	command, ids := flags.Arg(0), flags.Args()
	var err error
	var clean bool
	switch {
	case command == "check" && len(ids) == 1:
		clean, err = checkFile(ctx, *bibPath, registries, stdout)
	case command == "add" && len(ids) > 1:
		clean, err = addToFile(ctx, *bibPath, ids[1:], registries, stdout)
	case command == "cited" && len(ids) > 1:
		clean, err = checkCited(ids[1:], bibliographies(*bibPath, *also), stdout)
	default:
		err = errors.New("usage: citecheck [-bib <file>] [-also <files>] check | add <DOI | arXiv:id> [more ids] | cited <file.aux> [more files]")
	}
	switch {
	case err != nil:
		fmt.Fprintln(stderr, "citecheck:", err)
		return 2
	case !clean:
		return 1
	}
	return 0
}

// checkFile checks every entry and reports each one; it is clean when no entry
// has a problem.
func checkFile(ctx context.Context, path string, registries Registries, out io.Writer) (bool, error) {
	entries, err := readBib(path)
	if err != nil {
		return false, err
	}
	failing, unanswered := 0, 0
	for _, entry := range entries {
		problems, err := checkEntry(ctx, entry, registries)
		switch {
		case err != nil:
			unanswered++
			fmt.Fprintf(out, "%s: not checked: %v\n", entry.Key, err)
		case len(problems) == 0:
			fmt.Fprintf(out, "%s: ok\n", entry.Key)
		default:
			failing++
			for _, p := range problems {
				fmt.Fprintf(out, "%s: %s\n", entry.Key, p)
			}
		}
	}
	fmt.Fprintf(out, "%d entries, %d with problems\n", len(entries), failing)
	if unanswered > 0 {
		return false, fmt.Errorf("%d of %d entries not checked: a registry did not answer", unanswered, len(entries))
	}
	return failing == 0, nil
}

// checkEntry lists an entry's problems. A DOI no registry holds is one; a
// registry that does not answer is an error, since the entry was not checked.
func checkEntry(ctx context.Context, entry Entry, registries Registries) ([]string, error) {
	doi, err := DOIOf(entry)
	if err != nil {
		return []string{"cannot be checked: " + err.Error()}, nil
	}
	work, err := registries.Lookup(ctx, doi)
	switch {
	case errors.Is(err, ErrNotFound):
		return []string{"no registry holds " + doi}, nil
	case err != nil:
		return nil, err
	}
	return Check(entry, &work), nil
}

// addToFile looks each id up and appends its entry. A work already in the file,
// or one a notice puts in doubt, is refused and reported; the rest are added.
func addToFile(ctx context.Context, path string, ids []string, registries Registries, out io.Writer) (bool, error) {
	text, err := os.ReadFile(filepath.Clean(path))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	entries, err := ParseBib(string(text))
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}
	known := catalogOf(entries)
	added, clean := string(text), true
	var unanswered []string
	for _, id := range ids {
		entry, refusal, err := known.admit(ctx, id, registries)
		switch {
		case err != nil:
			fmt.Fprintf(out, "%s: not added: %v\n", id, err)
			unanswered = append(unanswered, id)
		case refusal != "":
			fmt.Fprintf(out, "%s: %s\n", id, refusal)
			clean = false
		default:
			added = appendEntry(added, entry)
			fmt.Fprintf(out, "%s: added as %s\n", id, entry.Key)
		}
	}
	if added != string(text) {
		if err := os.WriteFile(filepath.Clean(path), []byte(added), 0o644); err != nil { //nolint:gosec // G306: the bibliography is a public file of the repository
			return false, fmt.Errorf("write %s: %w", path, err)
		}
	}
	if len(unanswered) > 0 {
		return false, fmt.Errorf("not added, since a registry did not answer: %s", strings.Join(unanswered, ", "))
	}
	return clean, nil
}

// appendEntry adds an entry to the text of a bibliography, after one blank line.
func appendEntry(text string, entry Entry) string {
	if text != "" && !strings.HasSuffix(text, "\n\n") {
		text = strings.TrimRight(text, "\n") + "\n\n"
	}
	return text + FormatEntry(entry)
}

// catalog is what a bibliography already holds: its keys, and the DOIs its
// entries are checked by, each with the key of its entry.
type catalog struct {
	keys map[string]bool
	dois map[string]string
}

func catalogOf(entries []Entry) catalog {
	c := catalog{keys: map[string]bool{}, dois: map[string]string{}}
	for _, e := range entries {
		c.keys[strings.ToLower(e.Key)] = true
		if doi, err := DOIOf(e); err == nil {
			c.dois[strings.ToLower(doi)] = e.Key
		}
	}
	return c
}

func (c catalog) taken(key string) bool { return c.keys[strings.ToLower(key)] }

// admit looks a work up and writes its entry, or says why it may not be added:
// it is there already, no registry holds it, or a notice puts it in doubt. A
// registry that does not answer is an error: nothing was decided.
func (c catalog) admit(ctx context.Context, id string, registries Registries) (entry Entry, refusal string, err error) {
	doi := DOIFromID(id)
	if key, present := c.dois[strings.ToLower(doi)]; present {
		return Entry{}, "already in the file as " + key, nil
	}
	work, err := registries.Lookup(ctx, doi)
	switch {
	case errors.Is(err, ErrNotFound):
		return Entry{}, "not added: no registry holds " + doi, nil
	case err != nil:
		return Entry{}, "", err
	case len(work.Notices) > 0:
		return Entry{}, fmt.Sprintf("not added: %s records a %s", work.Registry, strings.Join(work.Notices, ", ")), nil
	}
	entry = EntryFor(&work, c.taken)
	c.keys[strings.ToLower(entry.Key)], c.dois[strings.ToLower(doi)] = true, entry.Key
	return entry, "", nil
}

func readBib(path string) ([]Entry, error) {
	text, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	entries, err := ParseBib(string(text))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return entries, nil
}

// bibliographies are the file -bib names and those -also lists.
func bibliographies(bib, also string) []string {
	paths := []string{bib}
	for _, path := range strings.Split(also, ",") {
		if path = strings.TrimSpace(path); path != "" {
			paths = append(paths, path)
		}
	}
	return paths
}
