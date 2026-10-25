package checks

import (
	"maps"
	"slices"
	"testing"
)

// ListingLanguages are the languages whose lessons list two labels with words
// of their own, so that a test of what the cards speak can hold each of them
// to a list.
func ListingLanguages() []string { return slices.Sorted(maps.Keys(listingWords)) }

// ReferenceQuestion is the question of the reference task with this id, as the
// content holds it, so that a test can measure a copy of a real one.
func ReferenceQuestion(t *testing.T, id string) string {
	t.Helper()

	for _, reference := range readReferenceQuestions(t) {
		if reference.ID == id {
			return reference.Question
		}
	}
	t.Fatalf("the content holds no reference task %s", id)
	return ""
}
