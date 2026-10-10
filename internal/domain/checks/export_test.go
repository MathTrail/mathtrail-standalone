package checks

import (
	"maps"
	"slices"
)

// ListingLanguages are the languages whose lessons list two labels with words
// of their own, so that a test of what the cards speak can hold each of them
// to a list.
func ListingLanguages() []string { return slices.Sorted(maps.Keys(listingWords)) }
