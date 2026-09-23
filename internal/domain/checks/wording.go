package checks

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// quantity is a count and its unit, in words that stay right for one.
func quantity(count int, unit string) string {
	if count == 1 {
		return "1 " + strings.TrimSuffix(unit, "s")
	}
	return fmt.Sprintf("%d %s", count, unit)
}

// listed quotes ids and joins them the way a sentence lists things.
func listed(ids []string) string {
	quoted := make([]string, len(ids))
	for i, id := range ids {
		quoted[i] = fmt.Sprintf("%q", id)
	}
	return inWords(quoted, 0)
}

// inWords joins items the way a sentence lists things — "3", "3 and 7", "1, 2
// and 3" — and ends with how many more there are, if there are: "1, 2 and 4
// more".
func inWords(items []string, more int) string {
	if more > 0 {
		items = append(slices.Clip(items), strconv.Itoa(more)+" more")
	}
	if len(items) < 2 {
		return strings.Join(items, "")
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}
