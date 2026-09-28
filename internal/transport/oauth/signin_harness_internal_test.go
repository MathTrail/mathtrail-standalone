package oauthserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// What a host signs a parent in with in these cases.
const (
	hostRedirect = "https://host.example/oauth/cb"
	hostName     = "Host"
	hostState    = "the-host's-own-state"
	// hostVerifier is the host's PKCE verifier, whose challenge the requests
	// carry.
	hostVerifier = "the-host-verifier-of-forty-three-characters-at-least"
)

// clock is a clock a case moves by hand.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *clock) advance(by time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(by)
}

// signIn is a server under test, served over TLS as a browser reaches it, with
// Google standing in and the lines it leaves kept.
type signIn struct {
	t         *testing.T
	server    *Server
	served    *httptest.Server
	routes    *http.ServeMux
	google    *googletest.Server
	ring      *seal.KeyRing
	clock     *clock
	documents *documents
	logs      *observer.ObservedLogs
}

// newSignIn builds a server whose sign-in goes to a stand-in for Google.
func newSignIn(t *testing.T) *signIn {
	t.Helper()
	return build(t, true, testDay)
}

// newSignInWithoutGoogle builds a server with no Google sign-in configured.
func newSignInWithoutGoogle(t *testing.T) *signIn {
	t.Helper()
	return build(t, false, testDay)
}

// newSignInNow builds a server whose clock starts at the real time: the check
// of a token at the resource is the protocol library's too, and the library
// judges a token by the real clock.
func newSignInNow(t *testing.T) *signIn {
	t.Helper()
	return build(t, true, time.Now())
}

func build(t *testing.T, withGoogle bool, start time.Time) *signIn {
	t.Helper()

	routes := http.NewServeMux()
	served := httptest.NewTLSServer(routes)
	t.Cleanup(served.Close)

	h := &signIn{
		t: t, served: served, routes: routes, ring: ringOf(t, 'k'), clock: &clock{now: start},
		documents: &documents{err: cimd.ErrUnreachable},
	}
	h.google = googletest.New(t, h.clock.Now)
	core, logs := observer.New(zapcore.DebugLevel)
	h.logs = logs

	settings := &Settings{
		PublicURL: served.URL,
		Scope:     "mcp",
		Seal:      h.ring,
		Documents: h.documents,
		Logger:    zap.New(core),
		SiteURL:   testSite,
		Now:       h.clock.Now,
	}
	if withGoogle {
		google, err := googleauth.New(&googleauth.Settings{
			ClientID:     googletest.ClientID,
			ClientSecret: googletest.ClientSecret,
			RedirectURL:  served.URL + CallbackPath,
			Endpoints:    h.google.Endpoints(),
			Now:          h.clock.Now,
		})
		if err != nil {
			t.Fatalf("googleauth.New() error = %v, want nil", err)
		}
		settings.Google = google
	}
	server, err := New(settings)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	h.server = server
	routes.Handle("POST /oauth/register", server.Register)
	routes.Handle("GET /oauth/authorize", server.Authorize)
	routes.Handle("POST /oauth/consent", server.Consent)
	routes.Handle("GET /oauth/callback", server.Callback)
	routes.Handle("POST /oauth/token", server.Token)
	routes.Handle("POST /oauth/revoke", server.Revoke)
	routes.Handle("GET /.well-known/oauth-protected-resource/mcp", server.ResourceMetadata)
	routes.Handle("GET /.well-known/oauth-authorization-server", server.ServerMetadata)
	return h
}

// register registers a host with the address it sends the parent back to,
// and is the client identifier it was given.
func (h *signIn) register(redirectURI, name string) string {
	h.t.Helper()

	client, err := h.server.clients.register(registration{
		RedirectURIs: []string{redirectURI}, ClientName: name, IssuedAt: testDay.Unix(),
	})
	if err != nil {
		h.t.Fatalf("register() error = %v, want nil", err)
	}
	return client
}

// authorizeURL is the address a host sends the parent to: a code, with the
// challenge of the host's verifier, for the resource here and its one scope.
// What a case sets in changed is changed, and an empty value is left out.
func (h *signIn) authorizeURL(clientID string, changed url.Values) string {
	query := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {hostRedirect},
		"code_challenge":        {googletest.ChallengeOf(hostVerifier)},
		"code_challenge_method": {"S256"},
		"state":                 {hostState},
		"resource":              {h.served.URL + "/mcp"},
		"scope":                 {"mcp"},
	}
	for name, values := range changed {
		query[name] = values
		if len(values) == 1 && values[0] == "" {
			delete(query, name)
		}
	}
	return h.served.URL + "/oauth/authorize?" + query.Encode()
}

// browser is a parent's browser against the server under test, for the test
// given: it keeps the cookies it is given as a browser does, and follows no
// redirect on its own, so a case sees every step.
func (h *signIn) browser(t *testing.T) *browser {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New() error = %v, want nil", err)
	}
	return &browser{t: t, served: h.served, client: &http.Client{
		Transport:     h.served.Client().Transport,
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// browser is a parent's browser.
type browser struct {
	t        *testing.T
	client   *http.Client
	served   *httptest.Server
	language string
}

// reply is an answer as a browser gets it, read whole.
type reply struct {
	status   int
	header   http.Header
	location string
	body     string
}

func (b *browser) get(address string) reply {
	b.t.Helper()

	request, err := http.NewRequestWithContext(b.t.Context(), http.MethodGet, address, http.NoBody)
	if err != nil {
		b.t.Fatalf("NewRequest(%q) error = %v, want nil", address, err)
	}
	return b.send(request)
}

func (b *browser) post(address string, form url.Values) reply {
	b.t.Helper()

	request, err := http.NewRequestWithContext(b.t.Context(), http.MethodPost, address, strings.NewReader(form.Encode()))
	if err != nil {
		b.t.Fatalf("NewRequest(%q) error = %v, want nil", address, err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return b.send(request)
}

func (b *browser) send(request *http.Request) reply {
	b.t.Helper()

	if b.language != "" {
		request.Header.Set("Accept-Language", b.language)
	}
	response, err := b.client.Do(request)
	if err != nil {
		b.t.Fatalf("%s %s: %v", request.Method, request.URL.Path, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		b.t.Fatalf("reading the answer to %s %s: %v", request.Method, request.URL.Path, err)
	}
	return reply{
		status:   response.StatusCode,
		header:   response.Header,
		location: response.Header.Get("Location"),
		body:     string(body),
	}
}

// cookie is the value of a cookie the browser would send to a path of the
// server, or nothing.
func (b *browser) cookie(path, name string) string {
	address, err := url.Parse(b.served.URL + path)
	if err != nil {
		b.t.Fatalf("url.Parse() error = %v, want nil", err)
	}
	for _, cookie := range b.client.Jar.Cookies(address) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// consentRequest is the sealed request a consent screen posts back.
var consentRequest = regexp.MustCompile(`name="request" value="([^"]+)"`)

// requestOn is the sealed request on a consent screen.
func requestOn(t *testing.T, screen reply) string {
	t.Helper()

	found := consentRequest.FindStringSubmatch(screen.body)
	if found == nil {
		t.Fatalf("the page holds no request to post back; status %d, body %q", screen.status, screen.body)
	}
	return found[1]
}

// toConsent takes a parent from the host's request to the consent screen.
func (h *signIn) toConsent(parent *browser, clientID string) (screen reply, request string) {
	h.t.Helper()
	return h.toConsentAt(parent, h.authorizeURL(clientID, nil))
}

// toConsentAt takes a parent from the address a host sent them to to the
// consent screen.
func (h *signIn) toConsentAt(parent *browser, address string) (screen reply, request string) {
	h.t.Helper()

	screen = parent.get(address)
	if screen.status != http.StatusOK {
		h.t.Fatalf("GET /oauth/authorize status = %d, want the consent screen; body %q", screen.status, screen.body)
	}
	return screen, requestOn(h.t, screen)
}

// toGoogle takes a parent from the host's request through the consent screen
// on to Google, and is the address they are sent to there.
func (h *signIn) toGoogle(parent *browser, clientID string) string {
	h.t.Helper()
	return h.toGoogleAt(parent, h.authorizeURL(clientID, nil))
}

// toGoogleAt takes a parent from the address a host sent them to through the
// consent screen on to Google, and is the address they are sent to there.
func (h *signIn) toGoogleAt(parent *browser, address string) string {
	h.t.Helper()

	_, request := h.toConsentAt(parent, address)
	allowed := parent.post(h.served.URL+"/oauth/consent", url.Values{"request": {request}, "decision": {"allow"}})
	if allowed.status != http.StatusSeeOther || !strings.HasPrefix(allowed.location, h.google.URL+"/auth?") {
		h.t.Fatalf("POST /oauth/consent = %d to %q, want 303 to Google", allowed.status, allowed.location)
	}
	return allowed.location
}

// answerAt is what a parent sent back to the host carries, or a failure when
// they were sent anywhere else.
func answerAt(t *testing.T, back reply) url.Values {
	t.Helper()

	if back.status != http.StatusFound && back.status != http.StatusSeeOther {
		t.Fatalf("status = %d, want the parent sent back to the host; body %q", back.status, back.body)
	}
	address, err := url.Parse(back.location)
	if err != nil {
		t.Fatalf("Location %q does not parse: %v", back.location, err)
	}
	if got := address.Scheme + "://" + address.Host + address.Path; got != hostRedirect {
		t.Fatalf("Location = %q leads to %q, want the host's %q", back.location, got, hostRedirect)
	}
	return address.Query()
}

// opened is the code a host was handed, opened as the exchange opens it.
func (h *signIn) opened(code string) grantCode {
	h.t.Helper()

	plain, err := h.ring.Open(seal.PurposeCode, code, h.served.URL)
	if err != nil {
		h.t.Fatalf("the code does not open as a code of this issuer: %v", err)
	}
	var opened grantCode
	if err := json.Unmarshal(plain, &opened); err != nil {
		h.t.Fatalf("the code opens into %q, which is not a code: %v", plain, err)
	}
	return opened
}

// lines are the lines of an event, each as the map of its fields.
func (h *signIn) lines(event string) []map[string]any {
	entries := h.logs.FilterMessage(event).All()
	lines := make([]map[string]any, 0, len(entries))
	for i := range entries {
		lines = append(lines, entries[i].ContextMap())
	}
	return lines
}
