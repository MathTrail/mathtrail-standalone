package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fixtureRegistries answers as Crossref and DataCite answered when the records
// under testdata/ were fetched: a DOI with a record there is served, any other
// is not found, and a DOI starting with 10.5555/fail makes the server fail.
func fixtureRegistries(t *testing.T) Registries {
	t.Helper()
	serve := func(dir, prefix string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			doi := strings.TrimPrefix(r.URL.Path, prefix)
			if strings.HasPrefix(doi, "10.5555/fail") {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
				return
			}
			name := strings.ReplaceAll(strings.ToLower(doi), "/", "_") + ".json"
			body, err := os.ReadFile(filepath.Join("testdata", dir, name))
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
