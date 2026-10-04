package country_test

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/country"
)

func TestACountryIsKnownByItsCode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		code string
		want bool
	}{
		{"US", true},
		{"FR", true},
		{"XK", true},
		{"AX", true},
		{"us", false},
		{"USA", false},
		{"ZZ", false},
		{"EU", false},
		{"UK", false},
		{"SU", false},
		{"", false},
	} {
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()

			if got := country.Known(tc.code); got != tc.want {
				t.Errorf("Known(%q) = %v, want %v", tc.code, got, tc.want)
			}
		})
	}
}

func TestARegionBelongsToItsCountry(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		country, region string
		want            bool
	}{
		{"a state", "US", "US-TX", true},
		{"the district", "US", "US-DC", true},
		{"a territory, which is a country of its own", "US", "US-PR", false},
		{"a state without its country's prefix", "US", "TX", false},
		{"a state written in lower case", "US", "us-tx", false},
		{"a state of another country", "CA", "US-TX", false},
		{"a country the list names no regions of", "CA", "CA-ON", false},
		{"no country", "", "US-TX", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := country.KnownRegion(tc.country, tc.region); got != tc.want {
				t.Errorf("KnownRegion(%q, %q) = %v, want %v", tc.country, tc.region, got, tc.want)
			}
		})
	}
}

func TestARegionIsNamedInEnglish(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ country, region, want string }{
		{"US", "US-TX", "Texas"},
		{"US", "US-DC", "District of Columbia"},
		{"US", "US-ZZ", ""},
		{"FR", "US-TX", ""},
		{"", "US-TX", ""},
		{"", "", ""},
	} {
		if got := country.RegionName(tc.country, tc.region); got != tc.want {
			t.Errorf("RegionName(%q, %q) = %q, want %q", tc.country, tc.region, got, tc.want)
		}
	}
}
