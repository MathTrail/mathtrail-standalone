package middleware

import (
	"net/url"
	"slices"
	"strings"
	"testing"
)

// Whatever a caller writes into a query, a line carries of a parameter with
// words of its own only those words and "other": nothing a caller made up
// gets through, however it is encoded or repeated.
func FuzzAQueryIsWrittenInItsProtocolsWords(f *testing.F) {
	f.Add("scope=mcp+openid&error=access_denied")
	f.Add("error=alice%40school.example&%73cope=Masha+Ivanova&prompt=login+consent")
	f.Add("SCOPE=mcp&Prompt=%20none%20&response_type=code+id_token&grant_type=refresh_token")

	f.Fuzz(func(t *testing.T, raw string) {
		for _, field := range queryFields(raw) {
			if field.Key == "query" && field.String != "unparsable" {
				holdToWords(t, field.String)
			}
		}
	})
}

// holdToWords fails the test for any word a written query carries that is
// neither a word of its parameter's protocol nor "other". An address is held
// to a rule of its own, and is not looked at here.
func holdToWords(t *testing.T, query string) {
	t.Helper()

	written, err := url.ParseQuery(query)
	if err != nil {
		t.Fatalf("the line wrote %q, which does not parse back: %v", query, err)
	}
	for name, values := range written {
		kind := loggedQueryKeys[strings.ToLower(name)]
		if kind.words == nil {
			continue
		}
		for _, word := range strings.Fields(strings.Join(values, " ")) {
			if word != otherLabel && !slices.Contains(kind.words, word) {
				t.Errorf("%s was written holding %q, which is not a word of its protocol", name, word)
			}
		}
	}
}
