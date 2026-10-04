package mcpserver

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
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

// A topic's pages are linked in the language of the lesson when the site is
// written in it, or in the language it narrows down to, and otherwise in the
// one every page is written in.
func TestPagesAreLinkedInTheLanguageOfTheLessonTheSiteIsWrittenIn(t *testing.T) {
	t.Parallel()

	tag := func(text string) *string { return &text }
	for _, tc := range []struct {
		name   string
		lesson *string
		want   string
	}{
		{"no language chosen", nil, "en"},
		{"a language the site is written in", tag("ru"), "ru"},
		{"the same, written in capitals by hand", tag("RU"), "ru"},
		{"a region of it", tag("ru-RU"), "ru"},
		{"a region of the language of every page", tag("en-GB"), "en"},
		{"a region of a language the site is not written in", tag("pt-BR"), "en"},
		{"a language the site is not written in", tag("ar"), "en"},
		{"a tag of nothing", tag("-"), "en"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := pageLanguage(&profile.Student{UILanguage: tc.lesson}); got != tc.want {
				t.Errorf("pageLanguage() = %q, want %q", got, tc.want)
			}
		})
	}
}

// A page's address is the site, the language and the topic, at the part asked
// for; a topic whose page is not published links nowhere.
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
		{"no slug", "https://mathtrail.app", content.Topic{ID: "counting.gaps", SitePage: true}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := pageAddress(tc.site, "ru", &tc.topic, anchorTraps); got != tc.want {
				t.Errorf("pageAddress() = %q, want %q", got, tc.want)
			}
		})
	}
}
