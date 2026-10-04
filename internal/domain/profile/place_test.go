package profile_test

import (
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// A country and a region are kept as codes of the list, whatever case and
// spaces they were typed with; a region is one of its country's, the one the
// edit gives or the one kept; and a family that moves to another country
// leaves the old one's region behind.
func TestAPlaceIsKeptAsCodesOfTheList(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                   string
		country, region        string // what the profile holds before
		edit                   profile.Edit
		wantCountry, wantState string
	}{
		{"a country typed as a person types it", "", "", profile.Edit{Country: text(" us ")}, "US", ""},
		{"a country and its state", "", "", profile.Edit{Country: text("US"), Region: text("us-tx")}, "US", "US-TX"},
		{"a state of the country kept", "US", "", profile.Edit{Region: text("US-CA")}, "US", "US-CA"},
		{"another country leaves the state behind", "US", "US-TX", profile.Edit{Country: text("FR")}, "FR", ""},
		{"the same country keeps the state", "US", "US-TX", profile.Edit{Country: text("us")}, "US", "US-TX"},
		{"no country, and so no state", "US", "US-TX", profile.Edit{Country: text("")}, "", ""},
		{"no state", "US", "US-TX", profile.Edit{Region: text(" ")}, "US", ""},
		{"an edit of something else leaves the place alone", "US", "US-TX", profile.Edit{Grade: number(3)}, "US", "US-TX"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			student := newStudent()
			student.Country, student.Region = tc.country, tc.region
			p := profile.New(student, "1.2.3", editedAt)
			if _, problems := p.Change(&tc.edit, shipped, "1.2.4", editedAt); len(problems) > 0 {
				t.Fatalf("Change() problems = %v, want none", problems)
			}
			if p.Student.Country != tc.wantCountry || p.Student.Region != tc.wantState {
				t.Errorf("the profile holds %q and %q, want %q and %q",
					p.Student.Country, p.Student.Region, tc.wantCountry, tc.wantState)
			}
		})
	}
}

// A code the list does not have is refused by its field, as one of the few a
// value may be; so is a region of a country other than the one it would be
// kept with, and a territory, which is a country of its own.
func TestAPlaceOffTheListIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		country string // what the profile holds before
		edit    profile.Edit
		field   string
	}{
		{"a country by three letters", "", profile.Edit{Country: text("USA")}, "country"},
		{"a code of no country", "", profile.Edit{Country: text("ZZ")}, "country"},
		{"a state with no country", "", profile.Edit{Region: text("US-TX")}, "region"},
		{"a state of another country", "", profile.Edit{Country: text("FR"), Region: text("US-TX")}, "region"},
		{"a state of the country kept, by its letters alone", "US", profile.Edit{Region: text("TX")}, "region"},
		{"a territory, which is a country of its own", "US", profile.Edit{Region: text("US-PR")}, "region"},
		{"a region of a country the list has none of", "CA", profile.Edit{Region: text("CA-ON")}, "region"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			student := newStudent()
			student.Country = tc.country
			p := profile.New(student, "1.2.3", editedAt)
			changed, problems := p.Change(&tc.edit, shipped, "1.2.4", editedAt)
			if changed || !slices.Equal(fields(problems), []string{tc.field}) || problems[0].Code != profile.CodeNotOneOf {
				t.Fatalf("Change() = %v, %v; want nothing changed and %s refused as %s", changed, problems, tc.field, profile.CodeNotOneOf)
			}
		})
	}
}

// The file is the parent's to edit by hand, so it is held to the length of a
// code alone: a country the list does not have still reads, and one longer
// than any code can be does not.
func TestTheFileHoldsAPlaceToTheLengthOfACode(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		country, region string
		wantErr         bool
	}{
		{"codes of the list", "US", "US-TX", false},
		{"codes the list does not have", "ZZ", "XX-ABC", false},
		{"a country longer than a code", "USA", "", true},
		{"a region longer than a code", "US", "US-TEXAS", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			student := newStudent()
			student.Country, student.Region = tc.country, tc.region
			if err := profile.New(student, "1.2.3", editedAt).Validate(); (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, want an error: %v", err, tc.wantErr)
			}
		})
	}
}
