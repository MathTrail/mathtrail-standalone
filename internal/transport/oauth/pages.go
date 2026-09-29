package oauthserver

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/language"

	"github.com/MathTrail/mathtrail-standalone/internal/widget"
)

// pageFiles are the pages a parent may see during a sign-in, the stylesheet
// they share, and their words in every language they are written in.
//
//go:embed pages
var pageFiles embed.FS

// pageLanguages are the languages the pages are written in. The first is the
// one a browser is answered in when it asks for none of them.
var pageLanguages = []language.Tag{language.English, language.Russian}

// fragment is a piece of a sentence as a page shows it: plain text, a name
// set apart from it, or a link.
type fragment struct {
	Text   string
	Strong bool
	Link   string
}

// pages draws the pages of a sign-in: the consent screen, and the page a
// sign-in stops at. They are the service's own, outside any chat, so they are
// drawn in the language the browser asks for, with the design's tokens
// inlined — the policy they are served under loads nothing.
type pages struct {
	consent *template.Template
	refusal *template.Template
	words   map[string]map[string]string
	matcher language.Matcher
	style   template.CSS
	site    string
}

// newPages reads the pages, their stylesheet and their words from the files
// given, and refuses words that differ in what they say between two
// languages: a key one language lacks would be a blank on a page in it.
func newPages(files fs.FS, site string) (*pages, error) {
	consent, err := template.ParseFS(files, "pages/consent.html", "pages/phrase.html")
	if err != nil {
		return nil, fmt.Errorf("oauth: read the consent page: %w", err)
	}
	refusal, err := template.ParseFS(files, "pages/refusal.html", "pages/phrase.html")
	if err != nil {
		return nil, fmt.Errorf("oauth: read the refusal page: %w", err)
	}
	style, err := fs.ReadFile(files, "pages/pages.css")
	if err != nil {
		return nil, fmt.Errorf("oauth: read the pages' stylesheet: %w", err)
	}

	words := make(map[string]map[string]string, len(pageLanguages))
	for _, tag := range pageLanguages {
		lang := tag.String()
		raw, err := fs.ReadFile(files, "pages/"+lang+".json")
		if err != nil {
			return nil, fmt.Errorf("oauth: read the words in %s: %w", lang, err)
		}
		var dictionary map[string]string
		if err := json.Unmarshal(raw, &dictionary); err != nil {
			return nil, fmt.Errorf("oauth: read the words in %s: %w", lang, err)
		}
		words[lang] = dictionary
	}
	first := words[pageLanguages[0].String()]
	for lang, dictionary := range words {
		if !slices.Equal(slices.Sorted(maps.Keys(dictionary)), slices.Sorted(maps.Keys(first))) {
			return nil, fmt.Errorf("oauth: the words in %s are not the words in %s", lang, pageLanguages[0])
		}
	}

	return &pages{
		consent: consent,
		refusal: refusal,
		words:   words,
		matcher: language.NewMatcher(pageLanguages),
		// Both stylesheets are the service's own, embedded when it is built:
		// nothing a request carries ever reaches them.
		style: template.CSS(widget.Tokens() + "\n" + string(style)), //nolint:gosec // the service's own stylesheets, embedded at build time
		site:  site,
	}, nil
}

// languageOf is the language a page is drawn in for a request: the best of
// the pages' languages for what the browser asks for, and the first of them
// when it asks for none.
func (p *pages) languageOf(r *http.Request) string {
	asked, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
	if err != nil || len(asked) == 0 {
		return pageLanguages[0].String()
	}
	_, index, confidence := p.matcher.Match(asked...)
	if confidence == language.No {
		return pageLanguages[0].String()
	}
	return pageLanguages[index].String()
}

// consentScreen is what the consent screen shows and posts.
type consentScreen struct {
	// client is what the client calls itself.
	client string
	// redirectURI is where the parent is sent back to.
	redirectURI string
	// request is the sealed request the screen posts back.
	request string
	// google is where allowing sends the parent next, when anywhere.
	google string
}

// consentView is the consent screen as its template reads it.
type consentView struct {
	Lang     string
	Style    template.CSS
	Words    map[string]string
	Heading  []fragment
	Return   []fragment
	Loopback bool
	Legal    []fragment
	Request  string
}

// showConsent draws the consent screen. It names the client as the client
// names itself, and the address the parent goes back to, which is what says
// who the client is: a name is only a string the client chose. An address on
// the parent's own computer gets a warning, since nothing proves which program
// listens there. The form may be sent to this server alone, and what it leads
// to — Google, or the client, when the parent declines — is let through too.
func (p *pages) showConsent(w http.ResponseWriter, r *http.Request, screen consentScreen) error {
	lang := p.languageOf(r)
	words := p.words[lang]
	client := shownName(screen.client)
	if client == "" {
		client = hostOf(screen.redirectURI)
	}
	return p.write(w, http.StatusOK, lang, p.consent, consentView{
		Lang:    lang,
		Style:   p.style,
		Words:   words,
		Heading: phrase(words["consent.heading"], map[string]fragment{"client": {Text: client, Strong: true}}),
		Return: phrase(words["consent.return"], map[string]fragment{
			"address": {Text: shownAddress(screen.redirectURI), Strong: true},
		}),
		Loopback: toThisComputer(screen.redirectURI),
		Legal: phrase(words["consent.legal"], map[string]fragment{
			"terms":   {Text: words["consent.terms"], Link: p.site + "/" + lang + "/terms/"},
			"privacy": {Text: words["consent.privacy"], Link: p.site + "/" + lang + "/privacy/"},
		}),
		Request: screen.request,
	}, screen.google, screen.redirectURI)
}

// mostMarks is how many marks may stack on one letter of a client's name as
// the consent screen shows it: enough for the letters of the scripts people
// write, and too few to pile up over the lines around the name.
const mostMarks = 3

// shownName is a client's name as the consent screen shows it: without the
// characters a reader cannot see — controls, and the marks that reorder,
// join or hide text — which would let a name read otherwise than it was sent,
// and with no more than a few marks stacked on any letter, which would let it
// spill over the words around it.
func shownName(name string) string {
	var shown strings.Builder
	stacked := 0
	for _, c := range name {
		switch {
		case unicode.In(c, unicode.Cc, unicode.Cf):
			continue
		case unicode.In(c, unicode.Mn, unicode.Me):
			stacked++
			if stacked > mostMarks {
				continue
			}
		default:
			stacked = 0
		}
		shown.WriteRune(c)
	}
	return shown.String()
}

// maxShownAddress is how much of the address the parent goes back to the
// consent screen shows: its host whole, and as much of its path as the rest
// of a line holds.
const maxShownAddress = 80

// shownAddress is the address the parent goes back to as the consent screen
// shows it: its host as hostOf spells it, with its port when it names one,
// and its path in ASCII, anything else escaped, shortened past what a line
// holds. The host says who the client is; the path says where on a host that
// serves many, since a page anybody may publish there is not the host's own.
func shownAddress(uri string) string {
	address, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	host := hostOf(uri)
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	if port := address.Port(); port != "" {
		host += ":" + port
	}
	// A host longer than the line still leaves its path a third of one.
	return host + shortened(address.EscapedPath(), max(maxShownAddress-len(host), maxShownAddress/3))
}

// shortened is a path in the room given, cut in its middle with an ellipsis
// when it is longer: its start and its end both show, since either may be
// what tells two pages of one host apart. Neither cut falls inside an escape.
func shortened(path string, room int) string {
	if len(path) <= room {
		return path
	}
	kept := max(room-1, 2)
	head, tail := kept/2, len(path)-(kept-kept/2)
	if escape := strings.LastIndexByte(path[:head], '%'); escape >= 0 && escape > head-3 {
		head = escape
	}
	if escape := strings.LastIndexByte(path[:tail], '%'); escape >= 0 && escape > tail-3 {
		tail = min(escape+3, len(path))
	}
	return path[:head] + "…" + path[tail:]
}

// stop is a sign-in stopped at a page of its own rather than sent back to the
// client: what goes wrong before the client and the address it gave are
// known, or once what came back can no longer be trusted to say where the
// parent came from.
type stop struct {
	// status is the answer's status.
	status int
	// page names the page's words: refusal.<page>.heading and .text.
	page string
	// code is the error in the protocol's words, which the page names for
	// whoever has to fix the client.
	code string
}

// The pages a sign-in stops at.
var (
	stoppedRequest      = stop{http.StatusBadRequest, "request", "invalid_request"}
	stoppedResponseType = stop{http.StatusBadRequest, "request", "unsupported_response_type"}
	stoppedTooLarge     = stop{http.StatusRequestEntityTooLarge, "request", "invalid_request"}
	stoppedClient       = stop{http.StatusBadRequest, "client", "invalid_client"}
	stoppedRedirect     = stop{http.StatusBadRequest, "redirect", "invalid_request"}
	stoppedExpired      = stop{http.StatusBadRequest, "expired", "invalid_request"}
	stoppedCookie       = stop{http.StatusBadRequest, "cookie", "invalid_request"}
	stoppedUnconfigured = stop{http.StatusServiceUnavailable, "unconfigured", "temporarily_unavailable"}
	stoppedFailed       = stop{http.StatusInternalServerError, "failed", "server_error"}
	stoppedBusy         = stop{http.StatusTooManyRequests, "busy", "temporarily_unavailable"}
)

// refusalView is a page a sign-in stops at, as its template reads it.
type refusalView struct {
	Lang    string
	Style   template.CSS
	Heading string
	Text    string
	Code    []fragment
}

// showRefusal draws the page a sign-in stopped at. It sends the parent
// nowhere: the page itself is the answer.
func (p *pages) showRefusal(w http.ResponseWriter, r *http.Request, how stop) error {
	lang := p.languageOf(r)
	words := p.words[lang]
	return p.write(w, how.status, lang, p.refusal, refusalView{
		Lang:    lang,
		Style:   p.style,
		Heading: words["refusal."+how.page+".heading"],
		Text:    words["refusal."+how.page+".text"],
		Code:    phrase(words["refusal.code"], map[string]fragment{"code": {Text: how.code, Strong: true}}),
	})
}

// busy is the page a parent's browser is shown when a sign-in is past its
// pace: too many requests from their network, so wait a minute and connect
// again. A page that could not be drawn has answered 500 already, which the
// request's own line records.
func (p *pages) busy(w http.ResponseWriter, r *http.Request) {
	_ = p.showRefusal(w, r, stoppedBusy)
}

// write answers with a page, drawn whole before anything is sent, so that a
// page that could not be drawn is a failure rather than half a page.
func (p *pages) write(w http.ResponseWriter, status int, lang string, page *template.Template, view any, formTargets ...string) error {
	var body bytes.Buffer
	if err := page.Execute(&body, view); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return fmt.Errorf("oauth: draw the page: %w", err)
	}
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	header.Set("Content-Language", lang)
	header.Set("Cache-Control", "no-store")
	header.Set("Content-Security-Policy", pagePolicy(formTargets))
	w.WriteHeader(status)
	_, _ = w.Write(body.Bytes())
	return nil
}

// pagePolicy is the content security policy of a page: it loads nothing, its
// styles are its own, it is framed by nobody, and its form may be sent to this
// server and on to the addresses given, since a browser holds the redirect
// that answers a form to the same list.
func pagePolicy(formTargets []string) string {
	sources := []string{"'self'"}
	for _, target := range formTargets {
		if source := policySource(target); source != "" {
			sources = append(sources, source)
		}
	}
	return "default-src 'none'; style-src 'unsafe-inline'; form-action " + strings.Join(sources, " ") +
		"; frame-ancestors 'none'"
}

// policySource is the origin of an address as a policy names it, or nothing
// when the address has none a policy can name. A policy names a host in ASCII
// letters, digits, dots and hyphens: a name in another script is named as IDNA
// spells it, the form a browser looks up, and a host IDNA does not spell — an
// IPv6 address, a character that would end the directive, an empty label —
// has no name there.
func policySource(address string) string {
	parsed, err := url.Parse(address)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return ""
	}
	host, spelled := asciiHost(parsed.Hostname())
	plain := !strings.ContainsFunc(host, func(c rune) bool {
		return (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '.' && c != '-'
	})
	emptyLabel := strings.HasPrefix(host, ".") || strings.Contains(host, "..")
	if !spelled || host == "" || !plain || emptyLabel {
		return ""
	}
	if port := parsed.Port(); port != "" {
		host += ":" + port
	}
	return parsed.Scheme + "://" + host
}

// phrase is a wording with its slots filled: every {name} in it becomes the
// fragment given for the name, and everything else stays plain text. A brace
// that names no slot is text like any other.
func phrase(wording string, slots map[string]fragment) []fragment {
	var fragments []fragment
	rest := wording
	for {
		open := strings.IndexByte(rest, '{')
		if open < 0 {
			break
		}
		length := strings.IndexByte(rest[open:], '}')
		if length < 0 {
			break
		}
		slot, named := slots[rest[open+1:open+length]]
		if !named {
			fragments = plain(fragments, rest[:open+1])
			rest = rest[open+1:]
			continue
		}
		fragments = append(plain(fragments, rest[:open]), slot)
		rest = rest[open+length+1:]
	}
	return plain(fragments, rest)
}

// plain adds text to a phrase, when there is any.
func plain(fragments []fragment, text string) []fragment {
	if text == "" {
		return fragments
	}
	return append(fragments, fragment{Text: text})
}
