package mcpserver_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store/memory"
)

// placePayload is where the family lives as a payload of the profile's tools
// carries it.
type placePayload struct {
	Profile *struct {
		Country *string `json:"country"`
		Region  *string `json:"region"`
	} `json:"profile"`
	Problems []problemPayload `json:"problems"`
	Changed  bool             `json:"changed"`
}

// The country and the state an adult gives are kept as codes of the list,
// handed back as such, and told in words — the state by its name too — and a
// form that moves the family to another country takes the state away with it.
func TestThePlaceOfTheFamilyIsKeptAndToldAsCodes(t *testing.T) {
	t.Parallel()

	kept := memory.New()
	_, session := lesson(t, kept)

	made := call(t, session, "save_profile", map[string]any{
		"pseudonym": "Comet", "grade": 3, "country": " us ", "region": "us-tx",
	})
	if payload := payloadOf[placePayload](t, made); payload.Profile == nil ||
		text(payload.Profile.Country) != "US" || text(payload.Profile.Region) != "US-TX" {
		t.Fatalf("payload = %+v, want US and US-TX", payload.Profile)
	}
	if words := textOf(t, made); !strings.Contains(words, "Country: US, state: US-TX (Texas).") {
		t.Errorf("the words are %q, want the country and the state told", words)
	}
	if p, _ := loadKept(t, kept); p.Student.Country != "US" || p.Student.Region != "US-TX" {
		t.Errorf("kept: %q and %q, want US and US-TX", p.Student.Country, p.Student.Region)
	}

	moved := payloadOf[placePayload](t, call(t, session, "edit_profile", map[string]any{"country": "FR"}))
	if !moved.Changed || moved.Profile == nil || text(moved.Profile.Country) != "FR" || moved.Profile.Region != nil {
		t.Errorf("payload = %+v, want FR and no state", moved.Profile)
	}
	if progress := payloadOf[placePayload](t, call(t, session, "read_progress", map[string]any{})); progress.Profile == nil ||
		text(progress.Profile.Country) != "FR" || progress.Profile.Region != nil {
		t.Errorf("the progress carries %+v, want FR and no state", progress.Profile)
	}

	cleared := call(t, session, "save_profile", map[string]any{"country": ""})
	if payload := payloadOf[placePayload](t, cleared); payload.Profile == nil || payload.Profile.Country != nil {
		t.Errorf("payload = %+v, want no country", payload.Profile)
	}
	if words := textOf(t, cleared); !strings.Contains(words, "No country is given.") {
		t.Errorf("the words are %q, want them to say no country is given", words)
	}
}

// A country or a state the list does not have is refused by its field, by a
// code a card can say in words of its own, and nothing is written.
func TestAPlaceOffTheListIsRefusedByItsField(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		arguments map[string]any
		field     string
	}{
		{"a country by three letters", map[string]any{"country": "USA"}, "country"},
		{"a state with no country", map[string]any{"region": "US-TX"}, "region"},
		{"a state of another country", map[string]any{"country": "FR", "region": "US-TX"}, "region"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kept := keptWith(t, "masha")
			was, _ := loadKept(t, kept)
			_, session := lesson(t, kept)

			for _, tool := range []string{"save_profile", "edit_profile"} {
				refused := payloadOf[placePayload](t, call(t, session, tool, tc.arguments))
				if len(refused.Problems) != 1 || refused.Problems[0].Field != tc.field ||
					refused.Problems[0].Code != profile.CodeNotOneOf {
					t.Errorf("%s problems = %+v, want %s refused as %s", tool, refused.Problems, tc.field, profile.CodeNotOneOf)
				}
			}
			if now, _ := loadKept(t, kept); now.Revision != was.Revision {
				t.Errorf("kept at revision %d, want %d: nothing written", now.Revision, was.Revision)
			}
		})
	}
}

// text is the text a pointer holds, or nothing.
func text(held *string) string {
	if held == nil {
		return ""
	}
	return *held
}
