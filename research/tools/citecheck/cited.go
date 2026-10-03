package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// citationLine is how LaTeX records a citation in the .aux file it writes:
// one \citation{…} line for every \cite or \nocite it met, in whatever file
// the document read it from, with the keys as it finally understood them —
// comments, verbatim text and URLs already set aside.
var citationLine = regexp.MustCompile(`^\\citation\{([^}]*)\}\s*$`)

// CitedKeys are the keys an .aux file records, each once, in the order first
// cited. \nocite{*} cites every entry of the bibliographies, and is recorded
// as the key "*".
func CitedKeys(aux string) []string {
	var keys []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(aux, "\n") {
		match := citationLine.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			continue
		}
		for _, key := range strings.Split(match[1], ",") {
			key = strings.TrimSpace(key)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			keys = append(keys, key)
		}
	}
	return keys
}

// checkCited reports every key the .aux files record that none of the
// bibliographies holds. It is clean when every cited key is found, which is
// what a paper's build asks before BibTeX runs: a citation of a lead that
// nobody has read and checked must not reach a PDF as though it had been.
// "*" is every entry of the bibliographies, so it is always found.
func checkCited(auxPaths, bibPaths []string, out io.Writer) (bool, error) {
	known, err := knownKeys(bibPaths)
	if err != nil {
		return false, err
	}
	known["*"] = true
	clean, cited := true, 0
	for _, path := range auxPaths {
		keys, missing, err := citedIn(path, known)
		if err != nil {
			return false, err
		}
		cited += len(keys)
		if len(missing) > 0 {
			clean = false
			fmt.Fprintf(out, "%s: cites keys no bibliography holds: %s\n", path, strings.Join(missing, ", "))
		}
	}
	if clean {
		fmt.Fprintf(out, "%d cited keys in %d files, every one in %s\n", cited, len(auxPaths), strings.Join(bibPaths, " or "))
	}
	return clean, nil
}

// knownKeys are the keys of every entry of the bibliographies.
func knownKeys(bibPaths []string) (map[string]bool, error) {
	known := make(map[string]bool)
	for _, path := range bibPaths {
		entries, err := readBib(path)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			known[e.Key] = true
		}
	}
	return known, nil
}

// citedIn are the keys one .aux file records, and those of them nothing
// holds, sorted.
func citedIn(path string, known map[string]bool) (keys, missing []string, err error) {
	text, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}
	keys = CitedKeys(string(text))
	for _, key := range keys {
		if !known[key] {
			missing = append(missing, key)
		}
	}
	slices.Sort(missing)
	return keys, missing, nil
}
