package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// getFrom reads a deposition from a server that answers as the handler does,
// and returns the error.
func getFrom(t *testing.T, handler http.HandlerFunc) error {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := Client{HTTP: server.Client(), URL: server.URL, Token: fakeToken}
	_, err := c.Get(t.Context(), server.URL+"/api/deposit/depositions/1")
	return err
}

// jsonOfLength is a JSON object that is exactly n bytes long.
func jsonOfLength(n int) []byte {
	return []byte("{" + strings.Repeat(" ", n-2) + "}")
}

// An answer is read whole or not at all: a page that is not the API's, one
// longer than any deposition, or one cut short is an error, never a
// deposition read from part of it. An answer exactly as long as the limit is
// still read.
func TestAnAnswerIsReadOnlyWhenItIsWholeAndTheAPIs(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		want    string // what the error says, or nothing for an answer that is read
	}{
		{"a page of HTML", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, "<html><body>Sign in</body></html>")
		}, "is not what the API gives"},
		{"an answer as long as the limit", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(jsonOfLength(mostAnswer))
		}, ""},
		{"an answer a byte past the limit", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(jsonOfLength(mostAnswer + 1))
		}, "the answer runs past 16777216 bytes"},
		{"an answer cut short", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "100")
			_, _ = io.WriteString(w, `{"id": 1`)
		}, "unexpected EOF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := getFrom(t, tc.handler)

			switch {
			case tc.want == "" && err != nil:
				t.Errorf("Get() error = %v, want the answer read", err)
			case tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)):
				t.Errorf("Get() error = %v, want one containing %q", err, tc.want)
			}
		})
	}
}

// A request that fails names what was asked and where, so a person can see
// which step broke, and never the token, since the error ends up in a log.
func TestARequestThatFailsNamesTheAddressButNotTheToken(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()
	c := Client{HTTP: &http.Client{}, URL: server.URL, Token: "a-secret-token"}
	link := server.URL + "/api/deposit/depositions/1"

	_, err := c.Get(t.Context(), link)

	if err == nil || !strings.Contains(err.Error(), "GET "+link) || strings.Contains(err.Error(), "a-secret-token") {
		t.Errorf("Get() from a server that is gone gives %v, want an error naming GET %s and not the token", err, link)
	}
}

// Zenodo gives the DOI it keeps for a draft as an object, or false before one
// is asked for; either way, and with no field at all, the DOI is read or is
// nothing, never a failure.
func TestReservedDOIIsTheDOIZenodoKeepsOrNothing(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, answer, want string
	}{
		{"a DOI kept", `{"metadata": {"prereserve_doi": {"doi": "10.5072/zenodo.7", "recid": 7}}}`, "10.5072/zenodo.7"},
		{"none asked for yet", `{"metadata": {"prereserve_doi": false}}`, ""},
		{"no field at all", `{"metadata": {}}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var d Deposition
			if err := json.Unmarshal([]byte(tc.answer), &d); err != nil {
				t.Fatal(err)
			}
			if got := d.ReservedDOI(); got != tc.want {
				t.Errorf("ReservedDOI() = %q, want %q", got, tc.want)
			}
		})
	}
}

// The DOI every version shares is the prefix of a version's DOI and the
// record's id; without either, the person is sent to the record's page rather
// than given a DOI that may not exist.
func TestConceptDOIIsBuiltOnlyFromWhatIsKnown(t *testing.T) {
	t.Parallel()
	const page = "the DOI Zenodo shows on the record's page"
	for _, tc := range []struct {
		name, doi, record, want string
	}{
		{"a version's DOI and the record", "10.5072/zenodo.1002", "1001", "10.5072/zenodo.1001"},
		{"no DOI", "", "1001", page},
		{"no record", "10.5072/zenodo.1002", "", page},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := conceptDOI(tc.doi, tc.record); got != tc.want {
				t.Errorf("conceptDOI(%q, %q) = %q, want %q", tc.doi, tc.record, got, tc.want)
			}
		})
	}
}

// A draft with no DOI kept says nothing of one, and still names its record.
func TestDOIsOfSaysOnlyWhatIsKnown(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, reserved, record, want string
	}{
		{"both", "10.5072/zenodo.1002", "1001", "Once published it is 10.5072/zenodo.1002.\nIts record is 1001; every version of it is cited by 10.5072/zenodo.1001.\n"},
		{"no DOI kept", "", "1001", "Its record is 1001; every version of it is cited by the DOI Zenodo shows on the record's page.\n"},
		{"neither", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := doisOf(tc.reserved, tc.record); got != tc.want {
				t.Errorf("doisOf(%q, %q) = %q, want %q", tc.reserved, tc.record, got, tc.want)
			}
		})
	}
}
