package country

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"golang.org/x/text/language"
)

// kosovo is the one code of the list ISO 3166-1 does not assign: it lies in
// the range the standard leaves to its users, and Kosovo is known by it.
const kosovo = "XK"

// The list holds every code ISO 3166-1 assigns, and Kosovo's, each once and in
// order, so that a change to it is a line added or removed. The size is part
// of the test because a code lost in an edit is otherwise invisible.
func TestTheListIsTheCountriesOfISO3166(t *testing.T) {
	t.Parallel()

	countries := known.Countries
	if got, want := len(countries), 250; got != want {
		t.Errorf("the list has %d countries, want %d: the 249 of ISO 3166-1 and %s", got, want, kosovo)
	}
	if !slices.IsSorted(countries) {
		t.Errorf("the countries are not in order")
	}
	if compacted := slices.Compact(slices.Clone(countries)); len(compacted) != len(countries) {
		t.Errorf("a country is listed twice")
	}

	twoLetters := regexp.MustCompile(`^[A-Z]{2}$`)
	for _, code := range countries {
		if !twoLetters.MatchString(code) {
			t.Errorf("%q is not two capital letters", code)
			continue
		}
		if code == kosovo {
			continue
		}
		// A code the language library does not know as a country, under the
		// same spelling, is a typo, a code withdrawn from the standard, or one
		// it only reserves.
		region, err := language.ParseRegion(code)
		if err != nil || region.String() != code || !region.IsCountry() || region.IsPrivateUse() {
			t.Errorf("%q is not a country ISO 3166-1 assigns", code)
		}
	}
}

// The United States is the one country a profile may name a region of: its
// fifty states and the District of Columbia, as ISO 3166-2 codes them. The
// outlying areas are countries of their own in ISO 3166-1, and a family there
// names its country.
func TestTheRegionsAreTheStatesOfTheUnitedStates(t *testing.T) {
	t.Parallel()

	if got := len(known.Regions); got != 1 {
		t.Errorf("the list names regions of %d countries, want 1", got)
	}
	states := known.Regions["US"]
	if got, want := len(states), 51; got != want {
		t.Errorf("the United States has %d regions, want %d: fifty states and the District of Columbia", got, want)
	}

	state := regexp.MustCompile(`^US-[A-Z]{2}$`)
	for code, name := range states {
		if !state.MatchString(code) {
			t.Errorf("%q is not a state code of ISO 3166-2:US", code)
		}
		if strings.TrimSpace(name) == "" {
			t.Errorf("%s has no name", code)
		}
	}
	for _, outlying := range []string{"AS", "GU", "MP", "PR", "UM", "VI"} {
		if _, listed := states["US-"+outlying]; listed || !Known(outlying) {
			t.Errorf("US-%s is listed as a state, or %s is not a country of the list", outlying, outlying)
		}
	}
	for country := range known.Regions {
		if !Known(country) {
			t.Errorf("the list names regions of %q, which is no country of it", country)
		}
	}
}
