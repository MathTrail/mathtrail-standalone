package oauthserver

import (
	"encoding/json"
	"flag"
	"html"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// update rewrites the snapshots of the pages from what the pages draw now. The
// snapshots are the service's own, so the diff that comes of it is what a
// review reads.
var update = flag.Bool("update", false, "rewrite the snapshots of the sign-in's pages")

// pagesOf reads the pages as the server does, linking to the test site.
func pagesOf(t *testing.T) *pages {
	t.Helper()

	screens, err := newPages(pageFiles, testSite)
	if err != nil {
		t.Fatalf("newPages() error = %v, want nil", err)
	}
	return screens
}

// pageWords reads the pages' words as they are embedded.
func pageWords(t *testing.T) map[string]string {
	t.Helper()

	raw, err := fs.ReadFile(pageFiles, "pages/"+pageLanguage+".json")
	if err != nil {
		t.Fatalf("reading the pages' words: %v", err)
	}
	var words map[string]string
	if err := json.Unmarshal(raw, &words); err != nil {
		t.Fatalf("the pages' words are not a flat object of strings: %v", err)
	}
	return words
}

// The words say everything the pages ask for — the words of every page a
// sign-in stops at among them — and nothing they never ask for, and each
// wording the code fills names the slots it is filled with.
func TestTheWordsSayEverythingThePagesAskFor(t *testing.T) {
	t.Parallel()

	words := pageWords(t)
	asked := wordsAskedFor(t)
	for _, key := range asked {
		if strings.TrimSpace(words[key]) == "" {
			t.Errorf("the pages' words have nothing for %q", key)
		}
	}
	for key := range words {
		if !slices.Contains(asked, key) {
			t.Errorf("the pages' words have %q, which no page asks for", key)
		}
	}
	for key, want := range map[string][]string{
		"consent.heading": {"{client}"},
		"consent.return":  {"{address}"},
		"consent.legal":   {"{privacy}", "{terms}"},
		"refusal.code":    {"{code}"},
	} {
		if got := slotsOf(words[key]); !slices.Equal(got, want) {
			t.Errorf("%s names the slots %q, want %q", key, got, want)
		}
	}
}

// wordsAskedFor are the keys the pages ask for: the ones their templates name,
// the words of every page a sign-in stops at, and the wordings the code fills.
func wordsAskedFor(t *testing.T) []string {
	t.Helper()

	var asked []string
	for _, name := range []string{"consent.html", "drive.html", "refusal.html"} {
		page, err := fs.ReadFile(pageFiles, "pages/"+name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		for _, found := range regexp.MustCompile(`index \.Words "([^"]+)"`).FindAllStringSubmatch(string(page), -1) {
			asked = append(asked, found[1])
		}
	}
	for _, stopped := range []stop{stoppedRequest, stoppedClient, stoppedRedirect, stoppedExpired, stoppedCookie, stoppedUnconfigured, stoppedFailed, stoppedBusy, stoppedPassword} {
		asked = append(asked, "refusal."+stopped.page+".heading", "refusal."+stopped.page+".text")
	}
	return append(asked, "consent.heading", "consent.return", "consent.legal", "consent.terms", "consent.privacy", "refusal.code")
}

// slotsOf are the slots a wording names, sorted.
func slotsOf(wording string) []string {
	return slices.Sorted(slices.Values(regexp.MustCompile(`\{[a-z]+\}`).FindAllString(wording, -1)))
}

// The pages are in English whatever language the browser reads: the consent
// screen, and the page a sign-in stops at, for a browser that asks for Russian
// first and for English not at all.
func TestThePagesAreInEnglishWhateverTheBrowserReads(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	parent := h.browser(t)
	parent.language = "ru-RU,ru;q=0.9"
	screen, _ := h.toConsent(parent, h.register(hostRedirect, hostName))
	stopped := parent.get(h.authorizeURL("an-unknown-client", nil))
	words := pageWords(t)
	for name, tc := range map[string]struct {
		page   reply
		status int
		says   string
	}{
		"the consent screen":          {screen, http.StatusOK, words["consent.allow"]},
		"the page a sign-in stops at": {stopped, http.StatusBadRequest, words["refusal.client.heading"]},
	} {
		if tc.page.status != tc.status {
			t.Errorf("%s: status = %d, want %d", name, tc.page.status, tc.status)
		}
		if got := tc.page.header.Get("Content-Language"); got != "en" {
			t.Errorf("%s: Content-Language = %q, want en", name, got)
		}
		for _, want := range []string{`<html lang="en">`, html.EscapeString(tc.says)} {
			if !strings.Contains(tc.page.body, want) {
				t.Errorf("%s has no %q", name, want)
			}
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
		"https://login-.example.com/cb":                        "",
		"https://a..b.example/cb":                              "",
		"https://login-.opena%C4%B0.com/cb":                    "",
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

// consentFor draws the consent screen for a client and an address back.
func consentFor(t *testing.T, client, redirectURI string) *httptest.ResponseRecorder {
	t.Helper()

	answer := httptest.NewRecorder()
	if err := pagesOf(t).showConsent(answer, consentScreen{
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
// other text and without what a reader cannot see or what would spill over
// the lines around it, and the address the parent goes back to — its host
// spelled as a browser goes there, and its path in ASCII, cut short past a
// line; an address on the parent's own computer is warned about, since
// nothing proves who listens there.
func TestTheConsentScreenSaysWhoAndWhere(t *testing.T) {
	t.Parallel()

	// A host whose first letter is a Cyrillic one that looks Latin, and a
	// name with a mark that turns the text after it around and one that
	// hides a character.
	cyrillic, reversing, hidden := string(rune(0x0441)), string(rune(0x202E)), string(rune(0x200B))
	// Marks that stack on a letter: an acute accent, and the circumflex and
	// the dot below a Vietnamese e carries at once.
	acute, circumflex, dotBelow := string(rune(0x0301)), string(rune(0x0302)), string(rune(0x0323))
	for _, tc := range []struct {
		name        string
		client      string
		redirectURI string
		shows       []string
		hides       []string
	}{
		{"a host over https", "Claude", "https://claude.ai/api/mcp/auth_callback",
			[]string{"<bdi>Claude</bdi>", "<bdi>claude.ai/api/mcp/auth_callback</bdi>"}, []string{`class="auth-warning"`}},
		{"a name that is markup", "<script>alert(1)</script>", "https://claude.ai/cb",
			[]string{"&lt;script&gt;alert(1)&lt;/script&gt;"}, []string{"<script>"}},
		{"no name", "", "https://chatgpt.com/connector/oauth/abc",
			[]string{"<bdi>chatgpt.com</bdi></strong> to MathTrail"}, nil},
		{"a program on this computer", "Inspector", "http://localhost:6274/oauth/callback",
			[]string{`class="auth-warning"`, "<bdi>localhost:6274/oauth/callback</bdi>"}, nil},
		{"a program on this computer by another name", "Dev", "https://app.localhost./cb",
			[]string{`class="auth-warning"`}, nil},
		{"a host that passes for another", "Claude", "https://" + cyrillic + "laude.ai/cb",
			[]string{"<bdi>xn--laude-0ye.ai/cb</bdi>"}, []string{cyrillic}},
		{"a name that hides what it says", "Cla" + hidden + "ude" + reversing + " Pro", "https://claude.ai/cb",
			[]string{"<bdi>Claude Pro</bdi>"}, []string{reversing, hidden}},
		{"a host no browser can spell", "Native", "http://[::1]:33418/cb",
			[]string{"<bdi>[::1]:33418/cb</bdi>", `class="auth-warning"`}, nil},
		{"where on a host that serves many", "Script", "https://script.google.com/macros/s/AKfycbx/exec",
			[]string{"<bdi>script.google.com/macros/s/AKfycbx/exec</bdi>"}, nil},
		{"a path in another script, and one that turns text around", "Claude",
			"https://claude.ai/" + cyrillic + "/a%E2%80%AEb",
			[]string{"<bdi>claude.ai/%D1%81/a%E2%80%AEb</bdi>"}, []string{cyrillic, reversing}},
		{"a path longer than a line, its start and its end kept", "Long",
			"https://host.example/start-" + strings.Repeat("a", 200) + "-end",
			[]string{"<bdi>host.example/start-" + strings.Repeat("a", 26) + "…" + strings.Repeat("a", 30) + "-end</bdi>"},
			[]string{strings.Repeat("a", 31)}},
		{"a path of escapes, cut between them", "Turning", "https://claude.ai/" + strings.Repeat("%E2%80%AE", 20),
			[]string{"<bdi>claude.ai/" + strings.Repeat("%E2%80%AE", 3) + "%E2%80…%80%AE" + strings.Repeat("%E2%80%AE", 3) + "</bdi>"}, nil},
		{"a host longer than a line, its path kept", "Long",
			"https://" + strings.Repeat("h", 90) + ".example/cb",
			[]string{"<bdi>" + strings.Repeat("h", 90) + ".example/cb</bdi>"}, nil},
		{"a name that piles marks on a letter", "Z" + strings.Repeat(acute, 40) + "oe", "https://claude.ai/cb",
			[]string{"<bdi>Z" + strings.Repeat(acute, mostMarks) + "oe</bdi>"}, []string{strings.Repeat(acute, mostMarks+1)}},
		{"a name whose letters carry their marks", "Vie" + circumflex + dotBelow + "t", "https://claude.ai/cb",
			[]string{"<bdi>Vie" + circumflex + dotBelow + "t</bdi>"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			answer := consentFor(t, tc.client, tc.redirectURI)
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

// A page is served as the service's own document: in English, kept by nobody
// on the way, loading nothing, framed by nobody, with the terms and the policy
// it points to in English too, and with a form that may be sent here and on to
// where it leads.
func TestAPageIsServedAsTheServicesOwn(t *testing.T) {
	t.Parallel()

	answer := consentFor(t, "Claude", "https://claude.ai/api/mcp/auth_callback")
	header := answer.Header()
	for name, want := range map[string]string{
		"Content-Type":     "text/html; charset=utf-8",
		"Content-Language": "en",
		"Cache-Control":    "no-store",
		"Content-Security-Policy": "default-src 'none'; style-src 'unsafe-inline'; " +
			"form-action 'self' https://accounts.google.com https://claude.ai; frame-ancestors 'none'",
	} {
		if got := header.Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	body := answer.Body.String()
	for _, want := range []string{`<html lang="en">`, "--surface:", testSite + "/en/terms/", testSite + "/en/privacy/", "Allow"} {
		if !strings.Contains(body, want) {
			t.Errorf("the page has no %q", want)
		}
	}
	for _, slot := range []string{"{client}", "{address}", "{terms}", "{privacy}"} {
		if strings.Contains(body, slot) {
			t.Errorf("the page shows the slot %s unfilled", slot)
		}
	}
}

// The pages as a parent sees them, against the snapshots the review last
// read. The stylesheet is left out of the comparison: it is the design tokens
// and the pages' own, which are checked where they are kept.
func TestThePagesLookAsTheReviewLastSawThem(t *testing.T) {
	t.Parallel()

	style := regexp.MustCompile(`(?s)<style>.*?</style>`)
	refusal := httptest.NewRecorder()
	if err := pagesOf(t).showRefusal(refusal, stoppedCookie); err != nil {
		t.Fatalf("showRefusal() error = %v, want nil", err)
	}
	withReviewer := httptest.NewRecorder()
	if err := pagesOf(t).showConsent(withReviewer, consentScreen{
		client:      "Claude",
		redirectURI: "https://claude.ai/api/mcp/auth_callback",
		request:     "mt1.s.KEYID.the-sealed-request",
		google:      "https://accounts.google.com/o/oauth2/v2/auth?state=mt1.s.KEYID.the-sealed-request",
		reviewer:    true,
	}); err != nil {
		t.Fatalf("showConsent() error = %v, want nil", err)
	}
	drive := httptest.NewRecorder()
	if err := pagesOf(t).showDrive(drive, driveScreen{
		redirectURI: "https://claude.ai/api/mcp/auth_callback",
		request:     "mt1.s.KEYID.the-sealed-request",
		google:      "https://accounts.google.com/o/oauth2/v2/auth?state=mt1.s.KEYID.the-sealed-request",
	}); err != nil {
		t.Fatalf("showDrive() error = %v, want nil", err)
	}
	for name, answer := range map[string]*httptest.ResponseRecorder{
		"consent.en.html":          consentFor(t, "Claude", "https://claude.ai/api/mcp/auth_callback"),
		"consent.reviewer.en.html": withReviewer,
		"drive.en.html":            drive,
		"refusal.en.html":          refusal,
	} {
		matchesSnapshot(t, name, style.ReplaceAllString(answer.Body.String(), "<style>…</style>"))
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

// pageFilesWith are the embedded pages with some of their files replaced and
// some left out.
func pageFilesWith(t *testing.T, replaced map[string]string, dropped ...string) fstest.MapFS {
	t.Helper()

	files := fstest.MapFS{}
	err := fs.WalkDir(pageFiles, "pages", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(pageFiles, name)
		files[name] = &fstest.MapFile{Data: data}
		return err
	})
	if err != nil {
		t.Fatalf("reading the embedded pages: %v", err)
	}
	for name, data := range replaced {
		files[name] = &fstest.MapFile{Data: []byte(data)}
	}
	for _, name := range dropped {
		delete(files, name)
	}
	return files
}

// Pages that cannot be read, or words that are not there, stop the service
// before it serves a sign-in.
func TestPagesThatCannotBeReadStopTheService(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		replaced map[string]string
		dropped  []string
	}{
		{name: "no consent screen", dropped: []string{"pages/consent.html"}},
		{name: "no page that asks for the Drive", dropped: []string{"pages/drive.html"}},
		{name: "a refusal that does not parse", replaced: map[string]string{"pages/refusal.html": "{{ if }}"}},
		{name: "no stylesheet", dropped: []string{"pages/pages.css"}},
		{name: "no words", dropped: []string{"pages/en.json"}},
		{name: "words that are no words", replaced: map[string]string{"pages/en.json": "[]"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := newPages(pageFilesWith(t, tc.replaced, tc.dropped...), testSite); err == nil {
				t.Error("newPages() error = nil, want the pages refused")
			}
		})
	}
	t.Run("the pages as they are embedded", func(t *testing.T) {
		t.Parallel()

		if _, err := newPages(pageFilesWith(t, nil), testSite); err != nil {
			t.Errorf("newPages() error = %v, want nil", err)
		}
	})
}

// A parent's browser past the sign-in's pace is shown a page of its own: 429,
// what happened and what to do, in English, and kept by no cache.
func TestABrowserPastThePaceIsShownAPageOfItsOwn(t *testing.T) {
	t.Parallel()

	answer := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/authorize", http.NoBody)
	request.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	pagesOf(t).busy(answer, request)

	if answer.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", answer.Code, http.StatusTooManyRequests)
	}
	for name, want := range map[string]string{"Content-Language": "en", "Cache-Control": "no-store"} {
		if got := answer.Header().Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	words := pageWords(t)
	body := answer.Body.String()
	for _, key := range []string{"refusal.busy.heading", "refusal.busy.text"} {
		if !strings.Contains(body, html.EscapeString(words[key])) {
			t.Errorf("the page does not say %q", words[key])
		}
	}
}
