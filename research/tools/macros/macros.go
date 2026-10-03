package main

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Source is one file of computed numbers and the namespace its keys are cited
// under: `\stat{product}{reference_tasks}` names the key reference_tasks of the
// source called product.
type Source struct {
	Name string
	Path string
}

// namePattern and keyPattern are what a namespace and a key may be. Both end
// up inside a control sequence name, where only these characters are safe to
// hand to \csname without a reader of the paper's source misreading them.
var (
	namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
	keyPattern  = regexp.MustCompile(`^[a-z0-9_]+$`)
)

// Parse reads one file of computed numbers: a `key=value` line for each, blank
// lines and lines starting with `#` ignored. A key given twice is an error:
// the paper would cite one number and mean two.
func Parse(text string) (map[string]string, error) {
	values := make(map[string]string)
	for n, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: no '=': %q", n+1, line)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !keyPattern.MatchString(key) {
			return nil, fmt.Errorf("line %d: %q is not a key: lower-case letters, digits and underscores only", n+1, key)
		}
		if value == "" {
			return nil, fmt.Errorf("line %d: %q has no value", n+1, key)
		}
		if _, seen := values[key]; seen {
			return nil, fmt.Errorf("line %d: %q is given twice", n+1, key)
		}
		values[key] = value
	}
	if len(values) == 0 {
		return nil, errors.New("no numbers in it")
	}
	return values, nil
}

// latexSpecial turns the characters TeX gives a meaning of its own into the
// text they stand for, so that a value such as 82.1% prints as it reads rather
// than cutting the rest of its line off as a comment.
var latexSpecial = strings.NewReplacer(
	`\`, `\textbackslash{}`,
	`{`, `\{`,
	`}`, `\}`,
	`%`, `\%`,
	`#`, `\#`,
	`&`, `\&`,
	`_`, `\_`,
	`$`, `\$`,
	`~`, `\textasciitilde{}`,
	`^`, `\textasciicircum{}`,
)

// Numbers are the values one source holds, under its name.
type Numbers struct {
	Source Source
	Values map[string]string
}

// looksIdentifying catches a value that would name the system or where it
// lives, whatever its key: a web address, or the name of the project or of its
// organisation.
var looksIdentifying = regexp.MustCompile(`(?i)://|mathtrail|github\.com`)

// Identifying reports whether a value must be hidden from the reviewers of an
// anonymous build: its key is one the caller names — a commit, which a reviewer
// could look up — or the value itself looks like an address or a name.
func Identifying(key, value string, keys map[string]bool) bool {
	return keys[key] || looksIdentifying.MatchString(value)
}

// Render writes every number of every source as a \statdef line — or a
// \statiddef line for a value that identifies the system, which the anonymous
// build prints as a mark instead — the sources in the order given and the keys
// of each in sorted order, so that the same numbers always give the same file
// and a changed number shows in its diff.
func Render(all []Numbers, identifying map[string]bool) string {
	var b strings.Builder
	names := make([]string, len(all))
	for i, n := range all {
		names[i] = n.Source.Name + "=" + n.Source.Path
	}
	fmt.Fprintf(&b, "%% Every number the paper states, generated from %s.\n", strings.Join(names, ", "))
	b.WriteString("% Do not edit: change a source and generate the file again.\n")
	for _, n := range all {
		keys := make([]string, 0, len(n.Values))
		for key := range n.Values {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, key := range keys {
			def := "statdef"
			if Identifying(key, n.Values[key], identifying) {
				def = "statiddef"
			}
			fmt.Fprintf(&b, "\\%s{%s}{%s}{%s}\n", def, n.Source.Name, key, latexSpecial.Replace(n.Values[key]))
		}
	}
	return b.String()
}

// ParseSource reads a `name=path` argument.
func ParseSource(arg string) (Source, error) {
	name, path, ok := strings.Cut(arg, "=")
	if !ok || path == "" {
		return Source{}, fmt.Errorf("%q is not name=path", arg)
	}
	if !namePattern.MatchString(name) {
		return Source{}, fmt.Errorf("%q is not a name: a lower-case letter, then lower-case letters and digits", name)
	}
	return Source{Name: name, Path: path}, nil
}
