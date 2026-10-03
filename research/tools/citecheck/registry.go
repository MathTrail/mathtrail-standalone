package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Work is what a registry records about a work: enough to check an entry
// against, and to write one.
type Work struct {
	DOI       string
	Registry  string // Crossref or DataCite
	Kind      string // the registry's own type, such as journal-article or Preprint
	Title     string
	Subtitle  string
	Authors   []Person
	Years     []int    // every year it records: of issue, in print, online
	Venues    []string // the containers it names: journal, proceedings, series
	Publisher string
	Volume    string
	Issue     string
	Pages     string
	Notices   []string // editorial notices that put the work in doubt
}

// Person is one author, as the registry splits the name.
type Person struct {
	Family string
	Given  string
}

// ErrNotFound means no registry holds the DOI: it names nothing, or nothing
// either registry knows.
var ErrNotFound = errors.New("neither Crossref nor DataCite knows this DOI")

// Registries looks works up in Crossref, and in DataCite for the DOIs
// Crossref does not hold — among them the ones arXiv and Zenodo register.
type Registries struct {
	Client    *http.Client
	Crossref  string // base URL of Crossref's REST API
	DataCite  string // base URL of DataCite's REST API
	UserAgent string
}

// dataCitePrefixes are DOI prefixes only DataCite registers: arXiv's and
// Zenodo's. Their works are looked up there first, sparing Crossref a request
// that would find nothing.
var dataCitePrefixes = []string{"10.48550/", "10.5281/"}

// Lookup finds what the registries record about a DOI: in the registry that
// most likely holds it, and in the other when that one does not.
func (r Registries) Lookup(ctx context.Context, doi string) (Work, error) {
	first, second := r.crossref, r.datacite
	for _, prefix := range dataCitePrefixes {
		if strings.HasPrefix(strings.ToLower(doi), prefix) {
			first, second = r.datacite, r.crossref
		}
	}
	work, err := first(ctx, doi)
	if !errors.Is(err, ErrNotFound) {
		return work, err
	}
	return second(ctx, doi)
}

// doubtful are the kinds of editorial update after which a work should not be
// cited as it stands; corrections and new versions are not among them.
var doubtful = []string{"retraction", "partial_retraction", "withdrawal", "removal", "expression_of_concern"}

func (r Registries) crossref(ctx context.Context, doi string) (Work, error) {
	body, err := r.get(ctx, r.Crossref+"/works/"+escapeDOI(doi))
	if err != nil {
		return Work{}, fmt.Errorf("crossref: %w", err)
	}
	var answer struct {
		Message struct {
			DOI             string
			Type            string
			Title           []string
			Subtitle        []string
			Author          []struct{ Given, Family, Name string }
			Issued          dateParts
			PublishedPrint  dateParts `json:"published-print"`
			PublishedOnline dateParts `json:"published-online"`
			ContainerTitle  []string  `json:"container-title"`
			Publisher       string
			Volume          string
			Issue           string
			Page            string
			UpdatedBy       []struct{ DOI, Type string } `json:"updated-by"`
		}
	}
	if err := json.Unmarshal(body, &answer); err != nil {
		return Work{}, fmt.Errorf("crossref: read the record of %s: %w", doi, err)
	}
	m := answer.Message
	work := Work{
		DOI: m.DOI, Registry: "Crossref", Kind: m.Type,
		Title: clean(first(m.Title)), Subtitle: clean(first(m.Subtitle)),
		Publisher: clean(m.Publisher), Volume: m.Volume, Issue: m.Issue, Pages: m.Page,
	}
	for _, a := range m.Author {
		family := a.Family
		if family == "" {
			family = a.Name // an organisation
		}
		work.Authors = append(work.Authors, Person{Family: clean(family), Given: clean(a.Given)})
	}
	for _, d := range []dateParts{m.Issued, m.PublishedPrint, m.PublishedOnline} {
		if year, ok := d.year(); ok && !slices.Contains(work.Years, year) {
			work.Years = append(work.Years, year)
		}
	}
	for _, c := range m.ContainerTitle {
		work.Venues = append(work.Venues, clean(c))
	}
	for _, u := range m.UpdatedBy {
		notice := fmt.Sprintf("%s (%s)", u.Type, u.DOI)
		if slices.Contains(doubtful, u.Type) && !slices.Contains(work.Notices, notice) {
			work.Notices = append(work.Notices, notice)
		}
	}
	return work, nil
}

func (r Registries) datacite(ctx context.Context, doi string) (Work, error) {
	body, err := r.get(ctx, r.DataCite+"/dois/"+escapeDOI(doi))
	if err != nil {
		return Work{}, fmt.Errorf("datacite: %w", err)
	}
	var answer struct {
		Data struct {
			Attributes struct {
				DOI    string
				Titles []struct {
					Title     string
					TitleType string `json:"titleType"`
				}
				Creators []struct {
					Name       string
					GivenName  string `json:"givenName"`
					FamilyName string `json:"familyName"`
				}
				PublicationYear json.RawMessage `json:"publicationYear"`
				Publisher       string
				Container       struct{ Title string }
				Types           struct {
					ResourceTypeGeneral string `json:"resourceTypeGeneral"`
				}
			}
		}
	}
	if err := json.Unmarshal(body, &answer); err != nil {
		return Work{}, fmt.Errorf("datacite: read the record of %s: %w", doi, err)
	}
	a := answer.Data.Attributes
	work := Work{DOI: a.DOI, Registry: "DataCite", Kind: a.Types.ResourceTypeGeneral, Publisher: clean(a.Publisher)}
	for _, t := range a.Titles {
		switch t.TitleType {
		case "":
			if work.Title == "" {
				work.Title = clean(t.Title)
			}
		case "Subtitle":
			work.Subtitle = clean(t.Title)
		}
	}
	for _, c := range a.Creators {
		work.Authors = append(work.Authors, dataCitePerson(c.Name, c.GivenName, c.FamilyName))
	}
	if year, err := strconv.Atoi(strings.Trim(string(a.PublicationYear), `"`)); err == nil {
		work.Years = []int{year}
	}
	if a.Container.Title != "" {
		work.Venues = []string{clean(a.Container.Title)}
	}
	return work, nil
}

// dataCitePerson splits a creator the way Crossref would: DataCite gives the
// family name apart when it knows it, and "Family, Given" otherwise.
func dataCitePerson(name, given, family string) Person {
	if family != "" {
		return Person{Family: clean(family), Given: clean(given)}
	}
	if f, g, found := strings.Cut(name, ","); found {
		return Person{Family: clean(f), Given: clean(g)}
	}
	return Person{Family: clean(name)}
}

// get fetches one record. A missing record is ErrNotFound, so the caller can
// ask the other registry; any other failure is reported as it is.
func (r Registries) get(ctx context.Context, address string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, http.NoBody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", r.UserAgent)
	request.Header.Set("Accept", "application/json")
	response, err := r.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	switch {
	case response.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case response.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%s answered %s", address, response.Status)
	}
	return io.ReadAll(io.LimitReader(response.Body, 16<<20))
}

// escapeDOI escapes a DOI for a URL path, keeping its slashes: the registries
// read the DOI from the path as written.
func escapeDOI(doi string) string {
	parts := strings.Split(doi, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// dateParts is how Crossref writes a date: [[year, month, day]], any part of
// which may be missing or null.
type dateParts struct {
	DateParts [][]*int `json:"date-parts"`
}

func (d dateParts) year() (int, bool) {
	if len(d.DateParts) == 0 || len(d.DateParts[0]) == 0 || d.DateParts[0][0] == nil {
		return 0, false
	}
	return *d.DateParts[0][0], true
}

// markup is an HTML or JATS tag, such as <i>, </sub> or <mml:math display="inline">:
// a name right after the bracket, so that "a < b" in a title is left alone.
var markup = regexp.MustCompile(`</?[A-Za-z][A-Za-z0-9:-]*(?:\s[^<>]*)?/?>`)

// clean turns a registry's string into plain text: Crossref titles carry HTML
// entities and tags such as <i>, and line breaks.
func clean(s string) string {
	return strings.Join(strings.Fields(html.UnescapeString(markup.ReplaceAllString(s, " "))), " ")
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
