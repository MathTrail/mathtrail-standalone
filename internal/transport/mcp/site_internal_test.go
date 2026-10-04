package mcpserver

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
)

// The site is written in the languages of its dictionaries, a file of each: a
// language the service links a page in is one every page is written in.
func TestTheSiteIsWrittenInTheLanguagesOfItsDictionaries(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("..", "..", "..", "web", "src", "site", "locales", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("the site's dictionaries are %v, %v; want some", files, err)
	}
	var written []string
	for _, file := range files {
		written = append(written, strings.TrimSuffix(filepath.Base(file), ".json"))
	}
	slices.Sort(written)
	listed := slices.Sorted(slices.Values(siteLanguages))
	if !slices.Equal(listed, written) {
		t.Errorf("siteLanguages = %v, want the site's dictionaries %v", listed, written)
	}
	if !slices.Contains(siteLanguages, everyPageLanguage) {
		t.Errorf("siteLanguages = %v, want %q among them", siteLanguages, everyPageLanguage)
	}
}

// A site's address is its origin alone, written one way whatever way it was
// configured: the card holds an address to the exact origin, so the words and
// the card name the site alike.
func TestASiteIsItsOriginWrittenOneWay(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		address string
		site    string
		isSite  bool
	}{
		{"https://mathtrail.app", "https://mathtrail.app", true},
		{"https://mathtrail.app/", "https://mathtrail.app", true},
		{"HTTPS://MathTrail.App", "https://mathtrail.app", true},
		{"https://mathtrail.app:443", "https://mathtrail.app", true},
		{"https://mathtrail.app:8443", "https://mathtrail.app:8443", true},
		{"http://localhost:8000", "http://localhost:8000", true},
		{"http://[::1]", "http://[::1]", true},
		{"", "", false},
		{"mathtrail.app", "", false},
		{"https://mathtrail.app/en/", "", false},
		{"https://mathtrail.app?child=Comet", "", false},
		{"https://mathtrail.app#traps", "", false},
		{"https://parent@mathtrail.app", "", false},
		{"ftp://mathtrail.app", "", false},
	} {
		site, isSite := siteOf(tc.address)
		if site != tc.site || isSite != tc.isSite {
			t.Errorf("siteOf(%q) = %q, %v; want %q, %v", tc.address, site, isSite, tc.site, tc.isSite)
		}
	}
}

// A topic's pages are linked in the language of the lesson when the site is
// written in it, or in the language it narrows down to, and otherwise in the
// one every page is written in.
func TestPagesAreLinkedInTheLanguageOfTheLessonTheSiteIsWrittenIn(t *testing.T) {
	t.Parallel()

	tag := func(text string) *string { return &text }
	for _, tc := range []struct {
		lesson *string
		want   string
	}{
		{nil, "en"},
		{tag("ru"), "ru"},
		{tag("ru-RU"), "ru"},
		{tag("en-GB"), "en"},
		{tag("pt-BR"), "en"},
		{tag("ar"), "en"},
		{tag("-"), "en"},
	} {
		lesson := "none"
		if tc.lesson != nil {
			lesson = *tc.lesson
		}
		if got := pageLanguage(tc.lesson); got != tc.want {
			t.Errorf("pageLanguage(%s) = %q, want %q", lesson, got, tc.want)
		}
	}
}

// A page's address is the site, the language and the topic, at the part asked
// for; a topic whose page is not published, or a service with no site, links
// nowhere.
func TestAPageIsLinkedOnlyOncePublished(t *testing.T) {
	t.Parallel()

	published := content.Topic{ID: "counting.gaps", Slug: "gaps-and-boundaries", SitePage: true}
	unpublished := content.Topic{ID: "counting.gaps", Slug: "gaps-and-boundaries"}
	for _, tc := range []struct {
		name  string
		site  string
		topic content.Topic
		want  string
	}{
		{"published, at its mistakes", "https://mathtrail.app", published, "https://mathtrail.app/ru/topics/gaps-and-boundaries/#traps"},
		{"not published yet", "https://mathtrail.app", unpublished, ""},
		{"no site", "", published, ""},
		{"no slug", "https://mathtrail.app", content.Topic{ID: "counting.gaps", SitePage: true}, ""},
	} {
		if got := pageAddress(tc.site, "ru", tc.topic, anchorTraps); got != tc.want {
			t.Errorf("%s: pageAddress() = %q, want %q", tc.name, got, tc.want)
		}
	}
}
