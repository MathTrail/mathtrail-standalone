package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
)

// fixtureRegistries answers as Crossref and DataCite answered when the records
// under testdata/ were fetched: a DOI with a record there is served, any other
// is not found, and a DOI starting with 10.5555/fail makes the server fail.
func fixtureRegistries(t *testing.T) Registries {
	t.Helper()
	serve := func(dir, prefix string) http.HandlerFunc {
		// The records are read through a root, so that no path a request
		// holds can reach a file outside them.
		records, err := os.OpenRoot(filepath.Join("testdata", dir))
		if err != nil {
			t.Fatalf("open the %s records: %v", dir, err)
		}
		t.Cleanup(func() { _ = records.Close() })
		return func(w http.ResponseWriter, r *http.Request) {
			doi := strings.TrimPrefix(r.URL.Path, prefix)
			if strings.HasPrefix(doi, "10.5555/fail") {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			name := strings.ReplaceAll(strings.ToLower(doi), "/", "_") + ".json"
			body, err := records.ReadFile(name)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			_, _ = w.Write(body)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/crossref/works/", serve("crossref", "/crossref/works/"))
	mux.HandleFunc("/datacite/dois/", serve("datacite", "/datacite/dois/"))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return Registries{Client: server.Client(), Crossref: server.URL + "/crossref", DataCite: server.URL + "/datacite", UserAgent: "citecheck-test"}
}

func TestLookupReadsACrossrefRecord(t *testing.T) {
	t.Parallel()

	work, err := fixtureRegistries(t).Lookup(context.Background(), "10.1016/j.compedu.2011.02.003")

	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if work.Registry != "Crossref" || work.Kind != "journal-article" {
		t.Errorf("registry and kind = %s %s, want Crossref journal-article", work.Registry, work.Kind)
	}
	families := []string{}
	for _, a := range work.Authors {
		families = append(families, a.Family)
	}
	if want := []string{"Klinkenberg", "Straatemeier", "van der Maas"}; !slices.Equal(families, want) {
		t.Errorf("authors = %q, want %q", families, want)
	}
	if !slices.Equal(work.Years, []int{2011}) {
		t.Errorf("years = %v, want [2011]", work.Years)
	}
	if want := []string{"Computers & Education"}; !slices.Equal(work.Venues, want) {
		t.Errorf("venues = %q, want %q: the entity must be decoded", work.Venues, want)
	}
	if work.Pages != "1813-1824" || work.Volume != "57" || work.Issue != "2" {
		t.Errorf("volume, issue, pages = %s, %s, %s; want 57, 2, 1813-1824", work.Volume, work.Issue, work.Pages)
	}
}

func TestLookupNamesARetractionOnceAndCleansTheTitle(t *testing.T) {
	t.Parallel()

	work, err := fixtureRegistries(t).Lookup(context.Background(), "10.1177/1758835919874651")

	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	// Retraction Watch and the publisher both report the retraction.
	if want := []string{"retraction (10.1177/17588359211061903)"}; !slices.Equal(work.Notices, want) {
		t.Errorf("notices = %q, want %q", work.Notices, want)
	}
	if strings.Contains(work.Title, "<i>") || strings.Contains(work.Title, "\n") {
		t.Errorf("title = %q, want no tags or line breaks", work.Title)
	}
}

func TestLookupAsksDataCiteForWhatCrossrefDoesNotHold(t *testing.T) {
	t.Parallel()

	work, err := fixtureRegistries(t).Lookup(context.Background(), "10.48550/arXiv.2503.16460")

	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if work.Registry != "DataCite" || work.Title != "Beyond Final Answers: Evaluating Large Language Models for Math Tutoring" {
		t.Errorf("registry and title = %s %q, want DataCite and the preprint's title", work.Registry, work.Title)
	}
	if len(work.Authors) != 5 || work.Authors[4].Family != "MacLellan" || work.Authors[4].Given != "Christopher J." {
		t.Errorf("authors = %+v, want five, the last Christopher J. MacLellan", work.Authors)
	}
	if !slices.Equal(work.Years, []int{2025}) {
		t.Errorf("years = %v, want [2025]", work.Years)
	}
}

func TestLookupReportsWhatNoRegistryHoldsAndWhatFails(t *testing.T) {
	t.Parallel()
	registries := fixtureRegistries(t)

	_, missing := registries.Lookup(context.Background(), "10.9999/no-such-doi")
	_, failing := registries.Lookup(context.Background(), "10.5555/fail")

	if !errors.Is(missing, ErrNotFound) {
		t.Errorf("unknown DOI: err = %v, want ErrNotFound", missing)
	}
	if failing == nil || errors.Is(failing, ErrNotFound) {
		t.Errorf("failing registry: err = %v, want a failure that is not ErrNotFound", failing)
	}
}

func TestEscapeDOIKeepsSlashesAndEscapesTheRest(t *testing.T) {
	t.Parallel()

	got := escapeDOI("10.1002/(SICI)1097-4571(199806)49:8<693::AID-ASI4>3.0.CO;2-O")

	if want := "10.1002/%28SICI%291097-4571%28199806%2949:8%3C693::AID-ASI4%3E3.0.CO%3B2-O"; got != want {
		t.Errorf("escapeDOI = %s, want %s", got, want)
	}
}

func TestLookupLeavesACorrectionOutOfTheNotices(t *testing.T) {
	t.Parallel()

	work, err := fixtureRegistries(t).Lookup(context.Background(), "10.1088/1361-6595/aaebdb")

	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if len(work.Notices) > 0 {
		t.Errorf("notices = %q, want none: a correction does not put a work in doubt", work.Notices)
	}
}

func TestLookupAsksDataCiteFirstForArXivAndZenodo(t *testing.T) {
	t.Parallel()
	registries := fixtureRegistries(t)
	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(broken.Close)
	registries.Crossref = broken.URL

	work, err := registries.Lookup(context.Background(), "10.48550/arXiv.2503.16460")

	if err != nil || work.Registry != "DataCite" {
		t.Errorf("Lookup with Crossref down = %s, %v; want DataCite's record, Crossref never asked", work.Registry, err)
	}
}

func TestCleanDropsTagsAndKeepsInequalities(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"When a < b and c > d: a lemma":                    "When a < b and c > d: a lemma",
		"Growth <i>via</i>\n   inactivation":               "Growth via inactivation",
		`Roots of <mml:math display="inline">x</mml:math>`: "Roots of x",
		"Computers &amp; Education":                        "Computers & Education",
	} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			if got := clean(in); got != want {
				t.Errorf("clean(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

// registry stands in for one registry that answers every request with one
// status and body, and counts the requests it was sent.
type registry struct {
	*httptest.Server
	asked atomic.Int32
}

func answering(t *testing.T, status int, body string) *registry {
	t.Helper()
	r := &registry{}
	r.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		r.asked.Add(1)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(r.Close)
	return r
}

// registriesOf asks the two stand-ins, one as Crossref and one as DataCite.
func registriesOf(crossref, datacite *registry) Registries {
	return Registries{Client: crossref.Client(), Crossref: crossref.URL, DataCite: datacite.URL, UserAgent: "citecheck-test"}
}

// A registry that answers with something other than a record has not said
// that it does not hold the DOI: the lookup fails, and the other registry,
// which would say it holds no such DOI, is not asked in its place.
func TestARecordThatIsNotJSONIsAFailureNotAMissingWork(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		registry, doi string
	}{
		{"crossref", "10.1016/j.compedu.2011.02.003"},
		{"datacite", "10.48550/arXiv.2503.16460"},
	} {
		t.Run(tc.registry, func(t *testing.T) {
			t.Parallel()
			page := answering(t, http.StatusOK, "<html><body>A notice about the service</body></html>")
			missing := answering(t, http.StatusNotFound, "")
			registries := registriesOf(page, missing)
			if tc.registry == "datacite" {
				registries = registriesOf(missing, page)
			}

			_, err := registries.Lookup(context.Background(), tc.doi)

			if err == nil || errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), tc.registry+": read the record of "+tc.doi) {
				t.Errorf("Lookup error = %v, want %s unable to read the record, which is not ErrNotFound", err, tc.registry)
			}
			if n := missing.asked.Load(); n != 0 {
				t.Errorf("the other registry was asked %d times, want none", n)
			}
		})
	}
}

// A DataCite record gives its titles by type, its creators as one name or
// split in two, its year as a string, and its container by its title; each is
// read where a Work keeps it, and an alternative title, before the others or
// after them, is neither the title nor the subtitle.
func TestLookupReadsADataCiteRecordOfEveryShape(t *testing.T) {
	t.Parallel()
	record := `{"data": {"attributes": {
		"doi": "10.5281/zenodo.42",
		"titles": [
			{"title": "An Alternative Name", "titleType": "AlternativeTitle"},
			{"title": "Counting &amp; Reasoning"},
			{"title": "A Study of Olympiad Tasks", "titleType": "Subtitle"},
			{"title": "Another Alternative Name", "titleType": "AlternativeTitle"}
		],
		"creators": [{"name": "Family, Given"}, {"name": "An Organisation"}, {"name": "Doe, Jane", "givenName": "Jane", "familyName": "Doe"}],
		"publicationYear": "2021",
		"publisher": "Zenodo",
		"container": {"title": "Series of Datasets"},
		"types": {"resourceTypeGeneral": "Dataset"}
	}}}`
	crossref := answering(t, http.StatusNotFound, "")

	work, err := registriesOf(crossref, answering(t, http.StatusOK, record)).Lookup(context.Background(), "10.5281/zenodo.42")

	want := Work{
		DOI: "10.5281/zenodo.42", Registry: "DataCite", Kind: "Dataset",
		Title: "Counting & Reasoning", Subtitle: "A Study of Olympiad Tasks",
		Authors: []Person{{"Family", "Given"}, {"An Organisation", ""}, {"Doe", "Jane"}},
		Years:   []int{2021}, Venues: []string{"Series of Datasets"}, Publisher: "Zenodo",
	}
	if err != nil || !reflect.DeepEqual(work, want) {
		t.Errorf("Lookup = %+v, %v; want %+v", work, err, want)
	}
	if n := crossref.asked.Load(); n != 0 {
		t.Errorf("Crossref was asked %d times for a Zenodo DOI, want none", n)
	}
}

// Crossref names an organisation among the authors by one name, with no
// family name; that name is read as the family name, so that the check and
// the key have one to go by.
func TestLookupReadsAnOrganisationCrossrefNamesAsAnAuthor(t *testing.T) {
	t.Parallel()
	record := `{"message": {"DOI": "10.1000/report.1", "type": "report", "title": ["Guidance on Tutoring"],
		"author": [{"name": "An Agency"}, {"given": "Ann", "family": "Smith"}],
		"issued": {"date-parts": [[2023]]}}}`

	work, err := registriesOf(answering(t, http.StatusOK, record), answering(t, http.StatusNotFound, "")).Lookup(context.Background(), "10.1000/report.1")

	if want := []Person{{"An Agency", ""}, {"Smith", "Ann"}}; err != nil || !reflect.DeepEqual(work.Authors, want) {
		t.Errorf("authors = %+v, %v; want %+v", work.Authors, err, want)
	}
}

// DataCite gives a creator's family name apart when it knows it, and the name
// as "Family, Given" otherwise; either way the person is split as Crossref
// would split them, and a name with no comma is an organisation's.
func TestDataCitePersonSplitsANameAsCrossrefWould(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		what, name, given, family string
		want                      Person
	}{
		{"split by DataCite", "Doe, Jane", "Jane", "Doe", Person{"Doe", "Jane"}},
		{"one name with a comma", "Doe, Jane", "", "", Person{"Doe", "Jane"}},
		{"a comma with no space after it", "van der Maas,H.L.J.", "", "", Person{"van der Maas", "H.L.J."}},
		{"a family name alone", "Somebody Else", "", "Doe", Person{"Doe", ""}},
		{"an organisation", "An Organisation", "", "", Person{"An Organisation", ""}},
		{"markup and spacing", "  Computers &amp;  Education Lab ", "", "", Person{"Computers & Education Lab", ""}},
	} {
		t.Run(tc.what, func(t *testing.T) {
			t.Parallel()
			if got := dataCitePerson(tc.name, tc.given, tc.family); got != tc.want {
				t.Errorf("dataCitePerson(%q, %q, %q) = %+v, want %+v", tc.name, tc.given, tc.family, got, tc.want)
			}
		})
	}
}
