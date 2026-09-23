package checks

import (
	"slices"
	"strings"
	"unicode"
)

// shingles is the set a question is compared by, sorted and without repeats,
// which is what makes the overlap of two of them one pass over both.
type shingles []string

// shinglesOf is the set of a question, of the kind its script calls for: word
// trigrams where words are separated by spaces, character bigrams where they
// are not. Word trigrams of a text with no word boundaries measure nothing.
func shinglesOf(text string) shingles {
	if Spaceless(text) {
		return ngrams(text, 1, 2)
	}
	return ngrams(text, 2, 3)
}

// ngrams is the normalisation of Postgres's pg_trgm, generalised to a length:
// the text is lowercased, everything that is neither a letter nor a digit
// becomes a boundary, each word is padded with lead spaces in front and one
// behind, and the set of its substrings of n characters is taken. With a lead
// of two and n of three it is pg_trgm exactly, which is what the prototype's
// threshold was measured with.
func ngrams(text string, lead, n int) shingles {
	var found []string
	for _, word := range strings.FieldsFunc(strings.ToLower(text), boundary) {
		padded := []rune(strings.Repeat(" ", lead) + word + " ")
		for i := 0; i+n <= len(padded); i++ {
			found = append(found, string(padded[i:i+n]))
		}
	}
	slices.Sort(found)
	return slices.Compact(found)
}

// boundary is a character that separates words: anything that is not a letter,
// a digit, or a mark that belongs to the letter before it.
func boundary(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsMark(r)
}

// jaccard is the share of the two sets that they have in common: what both
// hold, over what either holds. Two empty sets have nothing in common rather
// than everything, which is pg_trgm's answer too.
func (a shingles) jaccard(b shingles) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	common := 0
	for i, j := 0, 0; i < len(a) && j < len(b); {
		switch {
		case a[i] == b[j]:
			common++
			i++
			j++
		case a[i] < b[j]:
			i++
		default:
			j++
		}
	}
	return float64(common) / float64(len(a)+len(b)-common)
}

// Similarity is how alike two questions are, from nothing in common at 0 to
// the same set of shingles at 1. Each text is cut into the shingles its own
// script calls for, so two texts of different kinds share nothing — which is
// the right answer for a question in Chinese against one in English.
func Similarity(a, b string) float64 {
	return shinglesOf(a).jaccard(shinglesOf(b))
}
