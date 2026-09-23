package checks

import (
	"slices"
	"testing"
)

// Dropping what was already reported leaves the list it was handed as it was:
// a helper that rewrote its argument would change a list the caller may still
// be holding. A repeat right after the first is the order that shows it.
func TestDistinctLeavesWhatItWasGivenAlone(t *testing.T) {
	t.Parallel()

	a, b := structural("a"), structural("b")
	given := []Problem{a, a, b}
	if got := distinct(given); !slices.Equal(got, []Problem{a, b}) {
		t.Errorf("distinct() = %v, want %v", got, []Problem{a, b})
	}
	if !slices.Equal(given, []Problem{a, a, b}) {
		t.Errorf("distinct() rewrote what it was given into %v", given)
	}
}
