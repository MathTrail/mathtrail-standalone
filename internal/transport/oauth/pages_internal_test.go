package oauthserver

import (
	"encoding/json"
	"flag"
	"io/fs"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// update rewrites the snapshots of the pages from what the pages draw now. The
// snapshots are the service's own, so the diff that comes of it is what a
// review reads.
var update = flag.Bool("update", false, "rewrite the snapshots of the sign-in's pages")

// pagesOf reads the pages as the server does, linking to the test site.
func pagesOf(t *testing.T) *pages {
	t.Helper()

	screens, err := newPages(testSite)
	if err != nil {
		t.Fatalf("newPages() error = %v, want nil", err)
	}
	return screens
}

// wordsIn reads the words of a language as they are embedded.
func wordsIn(t *testing.T, lang string) map[string]string {
	t.Helper()

	raw, err := fs.ReadFile(pageFiles, "pages/"+lang+".json")
	if err != nil {
		t.Fatalf("reading the words in %s: %v", lang, err)
	}
	var words map[string]string
	if err := json.Unmarshal(raw, &words); err != nil {
		t.Fatalf("the words in %s are not a flat object of strings: %v", lang, err)
	}
	return words
}

// Every language says the same things: the same keys, every one the pages ask
// for, the words of every page a sign-in stops at, and in each wording the
// same slots as in English.
func TestEveryLanguageSaysTheSameThings(t *testing.T) {
	t.Parallel()

	english := wordsIn(t, "en")
	asked := wordsAskedFor(t)
	for _, tag := range pageLanguages {
		words := wordsIn(t, tag.String())
		if !slices.Equal(slices.Sorted(maps.Keys(words)), slices.Sorted(maps.Keys(english))) {
			t.Errorf("the words in %s have other keys than the words in English", tag)
		}
		for _, key := range asked {
			if strings.TrimSpace(words[key]) == "" {
				t.Errorf("the words in %s have nothing for %q", tag, key)
			}
		}
		for key, wording := range words {
			if got, want := slotsOf(wording), slotsOf(english[key]); !slices.Equal(got, want) {
				t.Errorf("%s in %s has the slots %q, want those of English, %q", key, tag, got, want)
			}
		}
	}
}

// wordsAskedFor are the keys the pages ask for: the ones their templates name,
// the words of every page a sign-in stops at, and the wordings the code fills.
func wordsAskedFor(t *testing.T) []string {
	t.Helper()

	var asked []string
	for _, name := range []string{"consent.html", "refusal.html"} {
		page, err := fs.ReadFile(pageFiles, "pages/"+name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		for _, found := range regexp.MustCompile(`index \.Words "([^"]+)"`).FindAllStringSubmatch(string(page), -1) {
			asked = append(asked, found[1])
		}
	}
	for _, stopped := range []stop{stoppedRequest, stoppedClient, stoppedRedirect, stoppedExpired, stoppedCookie, stoppedUnconfigured, stoppedFailed} {
		asked = append(asked, "refusal."+stopped.page+".heading", "refusal."+stopped.page+".text")
	}
	return append(asked, "consent.heading", "consent.return", "consent.legal", "consent.terms", "consent.privacy", "refusal.code")
}

// slotsOf are the slots a wording names, sorted.
func slotsOf(wording string) []string {
	return slices.Sorted(slices.Values(regexp.MustCompile(`\{[a-z]+\}`).FindAllString(wording, -1)))
}

// A page is drawn in the language the browser asks for most, among the ones
// the pages are written in, and in English when it asks for none of them.
func TestAPageIsInTheLanguageTheBrowserAsksFor(t *testing.T) {
	t.Parallel()

	screens := pagesOf(t)
	for asked, want := range map[string]string{
		"":                           "en",
		"ru":                         "ru",
		"ru-RU,ru;q=0.9,en;q=0.8":    "ru",
		"en-GB,en;q=0.9,ru;q=0.8":    "en",
		"de-DE,de;q=0.9":             "en",
		"de-DE,de;q=0.9,ru;q=0.5":    "ru",
		"fr;q=1.0, ru;q=0.1":         "ru",
		"not a language at all, ;;;": "en",
	} {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/authorize", http.NoBody)
		request.Header.Set("Accept-Language", asked)
		if got := screens.languageOf(request); got != want {
			t.Errorf("languageOf(%q) = %q, want %q", asked, got, want)
		}
	}
}

// A wording's slots are filled with what is given for them, and everything
// else of it stays text: a brace that names no slot, and one never closed.
func TestAPhraseFillsItsSlots(t *testing.T) {
	t.Parallel()

	slots := map[string]fragment{"client": {Text: "Claude", Strong: true}, "terms": {Text: "terms", Link: "https://s.example/t"}}
	for _, tc := range []struct {
		wording string
		want    []fragment
	}{
		{"", nil},
		{"no slots", []fragment{{Text: "no slots"}}},
		{"{client}", []fragment{{Text: "Claude", Strong: true}}},
		{"Connect {client} now", []fragment{{Text: "Connect "}, {Text: "Claude", Strong: true}, {Text: " now"}}},
		{"{client}{terms}", []fragment{{Text: "Claude", Strong: true}, {Text: "terms", Link: "https://s.example/t"}}},
		{"{nobody} {client}", []fragment{{Text: "{"}, {Text: "nobody} "}, {Text: "Claude", Strong: true}}},
		{"open {client", []fragment{{Text: "open {client"}}},
	} {
		if got := phrase(tc.wording, slots); !slices.Equal(got, tc.want) {
			t.Errorf("phrase(%q) = %+v, want %+v", tc.wording, got, tc.want)
		}
	}
}

// A policy names an origin it can name, and leaves out one it cannot, which
// would either be read as a directive of its own or not be read at all.
func TestAPolicyNamesOnlyTheOriginsItCan(t *testing.T) {
	t.Parallel()

	for address, want := range map[string]string{
		"https://claude.ai/api/mcp/auth_callback":              "https://claude.ai",
		"https://accounts.google.com/o/oauth2/v2/auth?state=x": "https://accounts.google.com",
		"http://127.0.0.1:33418/callback":                      "http://127.0.0.1:33418",
		"http://localhost:6274/oauth/callback":                 "http://localhost:6274",
		"https://пример.рф/callback":                           "https://xn--e1afmkfd.xn--p1ai",
		"https://Claude.AI/api/mcp/auth_callback":              "https://claude.ai",
		"http://[::1]:33418/callback":                          "",
		"https://a;sandbox.example/cb":                         "",
		"https://a_b.example/cb":                               "",
		"custom-app://callback":                                "",
		"not an address":                                       "",
		"":                                                     "",
	} {
		if got := policySource(address); got != want {
			t.Errorf("policySource(%q) = %q, want %q", address, got, want)
		}
	}
	if got := pagePolicy([]string{"", "http://[::1]:1/cb", "https://host.example/cb"}); got !=
		"default-src 'none'; style-src 'unsafe-inline'; form-action 'self' https://host.example; frame-ancestors 'none'" {
		t.Errorf("pagePolicy() = %q, want the one origin it can name after this server's own", got)
	}
}

// consentFor draws the consent screen for a client and an address back, in a
// language.
func consentFor(t *testing.T, client, redirectURI, lang string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/authorize", http.NoBody)
	request.Header.Set("Accept-Language", lang)
	answer := httptest.NewRecorder()
	if err := pagesOf(t).showConsent(answer, request, consentScreen{
		client:      client,
		redirectURI: redirectURI,
		request:     "mt1.s.KEYID.the-sealed-request",
		google:      "https://accounts.google.com/o/oauth2/v2/auth?state=mt1.s.KEYID.the-sealed-request",
	}); err != nil {
		t.Fatalf("showConsent() error = %v, want nil", err)
	}
	return answer
}

// The consent screen names the client as it names itself, escaped like any
// other text and without what a reader cannot see, and the host the parent
// goes back to, spelled as a browser goes there; an address on the parent's
// own computer is warned about, since nothing proves who listens there.
func TestTheConsentScreenSaysWhoAndWhere(t *testing.T) {
	t.Parallel()

	// A host whose first letter is a Cyrillic one that looks Latin, and a
	// name with a mark that turns the text after it around and one that
	// hides a character.
	cyrillic, reversing, hidden := string(rune(0x0441)), string(rune(0x202E)), string(rune(0x200B))
	for _, tc := range []struct {
		name        string
		client      string
		redirectURI string
		shows       []string
		hides       []string
	}{
		{"a host over https", "Claude", "https://claude.ai/api/mcp/auth_callback",
			[]string{"<bdi>Claude</bdi>", "<bdi>claude.ai</bdi>"}, []string{`class="auth-warning"`}},
		{"a name that is markup", "<script>alert(1)</script>", "https://claude.ai/cb",
			[]string{"&lt;script&gt;alert(1)&lt;/script&gt;"}, []string{"<script>"}},
		{"no name", "", "https://chatgpt.com/connector/oauth/abc",
			[]string{"<bdi>chatgpt.com</bdi></strong> to MathTrail"}, nil},
		{"a program on this computer", "Inspector", "http://localhost:6274/oauth/callback",
			[]string{`class="auth-warning"`, "<bdi>localhost</bdi>"}, nil},
		{"a host that passes for another", "Claude", "https://" + cyrillic + "laude.ai/cb",
			[]string{"<bdi>xn--laude-0ye.ai</bdi>"}, []string{cyrillic}},
		{"a name that hides what it says", "Cla" + hidden + "ude" + reversing + " Pro", "https://claude.ai/cb",
			[]string{"<bdi>Claude Pro</bdi>"}, []string{reversing, hidden}},
		{"a host no browser can spell", "Native", "http://[::1]:33418/cb",
			[]string{"<bdi>::1</bdi>"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			answer := consentFor(t, tc.client, tc.redirectURI, "en")
			body := answer.Body.String()
			for _, want := range tc.shows {
				if !strings.Contains(body, want) {
					t.Errorf("the consent screen has no %q", want)
				}
			}
			for _, unwanted := range tc.hides {
				if strings.Contains(body, unwanted) {
					t.Errorf("the consent screen has %q", unwanted)
				}
			}
		})
	}
}

// A page is served as the service's own document: in the language it is drawn
// in, kept by nobody on the way, loading nothing, framed by nobody, and with a
// form that may be sent here and on to where it leads.
func TestAPageIsServedAsTheServicesOwn(t *testing.T) {
	t.Parallel()

	answer := consentFor(t, "Claude", "https://claude.ai/api/mcp/auth_callback", "ru")
	header := answer.Header()
	for name, want := range map[string]string{
		"Content-Type":     "text/html; charset=utf-8",
		"Content-Language": "ru",
		"Cache-Control":    "no-store",
		"Content-Security-Policy": "default-src 'none'; style-src 'unsafe-inline'; " +
			"form-action 'self' https://accounts.google.com https://claude.ai; frame-ancestors 'none'",
	} {
		if got := header.Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	body := answer.Body.String()
	for _, want := range []string{`<html lang="ru">`, "--surface:", testSite + "/ru/terms/", testSite + "/ru/privacy/", "Разрешить"} {
		if !strings.Contains(body, want) {
			t.Errorf("the page has no %q", want)
		}
	}
	for _, slot := range []string{"{client}", "{host}", "{terms}", "{privacy}"} {
		if strings.Contains(body, slot) {
			t.Errorf("the page shows the slot %s unfilled", slot)
		}
	}
}

// The pages as a parent sees them, in every language, against the snapshots
// the review last read. The stylesheet is left out of the comparison: it is
// the design tokens and the pages' own, which are checked where they are
// kept.
func TestThePagesLookAsTheReviewLastSawThem(t *testing.T) {
	t.Parallel()

	style := regexp.MustCompile(`(?s)<style>.*?</style>`)
	for _, tag := range pageLanguages {
		lang := tag.String()
		refusal := httptest.NewRecorder()
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/callback", http.NoBody)
		request.Header.Set("Accept-Language", lang)
		if err := pagesOf(t).showRefusal(refusal, request, stoppedCookie); err != nil {
			t.Fatalf("showRefusal() error = %v, want nil", err)
		}
		for name, answer := range map[string]*httptest.ResponseRecorder{
			"consent." + lang + ".html": consentFor(t, "Claude", "https://claude.ai/api/mcp/auth_callback", lang),
			"refusal." + lang + ".html": refusal,
		} {
			matchesSnapshot(t, name, style.ReplaceAllString(answer.Body.String(), "<style>…</style>"))
		}
	}
}

// matchesSnapshot holds a page to its snapshot, or rewrites the snapshot when
// the tests run to update it.
func matchesSnapshot(t *testing.T, name, drawn string) {
	t.Helper()

	snapshot := filepath.Join("testdata", "pages", name)
	if *update {
		if err := os.WriteFile(snapshot, []byte(drawn), 0o644); err != nil {
			t.Fatalf("writing %s: %v", snapshot, err)
		}
		return
	}
	kept, err := os.ReadFile(snapshot)
	if err != nil {
		t.Fatalf("reading %s: %v; run the test with -update to write it", snapshot, err)
	}
	if drawn != string(kept) {
		t.Errorf("%s is drawn otherwise than its snapshot; run with -update and review the diff\ngot:\n%s", name, drawn)
	}
}
