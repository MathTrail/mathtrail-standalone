package render

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
)

const stringsDir = "strings"

// stringSet is the short interface text of every locale — the masthead, the
// footer and the labels of the links between pages — keyed by locale and then
// by name.
type stringSet map[string]map[string]string

// readStrings loads one file per locale and refuses a set that is not the same
// everywhere. A locale missing a name would otherwise publish a page with a
// blank label on it, which is the kind of fault nobody notices in a language
// they do not read.
func readStrings(source fs.FS, locales []string, reference string) (stringSet, error) {
	set := make(stringSet, len(locales))
	for _, locale := range locales {
		file := path.Join(stringsDir, locale+".json")
		raw, err := fs.ReadFile(source, file)
		if err != nil {
			return nil, fmt.Errorf("render: read %s: %w", file, err)
		}
		var text map[string]string
		if err := json.Unmarshal(raw, &text); err != nil {
			return nil, fmt.Errorf("render: parse %s: %w", file, err)
		}
		if len(text) == 0 {
			return nil, fmt.Errorf("render: %s holds no text", file)
		}
		set[locale] = text
	}

	for _, locale := range locales {
		if locale == reference {
			continue
		}
		if err := sameNames(set[reference], set[locale], reference, locale); err != nil {
			return nil, err
		}
	}
	return set, nil
}

// sameNames reports the first name the two locales disagree about.
func sameNames(reference, other map[string]string, referenceLocale, otherLocale string) error {
	for _, name := range sortedKeys(reference) {
		if _, ok := other[name]; !ok {
			return fmt.Errorf("render: %s.json has no %q, which %s.json has", otherLocale, name, referenceLocale)
		}
	}
	for _, name := range sortedKeys(other) {
		if _, ok := reference[name]; !ok {
			return fmt.Errorf("render: %s.json has %q, which %s.json does not", otherLocale, name, referenceLocale)
		}
	}
	return nil
}

// sortedKeys lists a map's keys in a stable order, so that the same mistake is
// always reported the same way.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
