package oauthserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// Whatever a registration sends, it is either registered or refused in the
// protocol's words — never a failure of the server — and a registration made
// opens into exactly what the answer said was registered.
func FuzzRegister(f *testing.F) {
	for _, seed := range []struct{ contentType, body string }{
		{"application/json", `{"redirect_uris":["https://claude.ai/api/mcp/auth_callback"],"client_name":"Claude"}`},
		{"application/json; charset=utf-8", `{"redirect_uris":["http://localhost:3000/cb","https://a.example/cb"]}`},
		{"application/json", `{"redirect_uris":["javascript:alert(1)"]}`},
		{"application/json", `{"redirect_uris":"https://a.example/cb"}`},
		{"application/json", `null`},
		{"text/plain", `{}`},
		{"", ``},
	} {
		f.Add(seed.contentType, seed.body)
	}

	f.Fuzz(func(t *testing.T, contentType, body string) {
		server, _ := knownServer(t)
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/oauth/register", strings.NewReader(body))
		request.Header.Set("Content-Type", contentType)
		answer := httptest.NewRecorder()
		server.Register.ServeHTTP(answer, request)

		switch answer.Code {
		case http.StatusBadRequest, http.StatusRequestEntityTooLarge:
			return
		case http.StatusCreated:
		default:
			t.Fatalf("status = %d for %q, want 201, 400 or 413", answer.Code, body)
		}
		var registered struct {
			ClientID     string   `json:"client_id"`
			RedirectURIs []string `json:"redirect_uris"`
			ClientName   string   `json:"client_name"`
		}
		if err := json.Unmarshal(answer.Body.Bytes(), &registered); err != nil {
			t.Fatalf("the answer to %q is not JSON: %v", body, err)
		}
		client, err := server.clients.resolve(t.Context(), registered.ClientID)
		if err != nil {
			t.Fatalf("the client_id registered for %q does not open: %v", body, err)
		}
		if !slices.Equal(client.RedirectURIs, registered.RedirectURIs) || client.Name != registered.ClientName {
			t.Errorf("the client_id opens into %+v, want what the answer said: %+v", client, registered)
		}
	})
}

// Whatever a client presents as its identifier, it is either no client or a
// client a parent may be sent back to — and it never takes the server down.
func FuzzClientID(f *testing.F) {
	server, _ := knownServer(f)
	issued, err := server.clients.register(registration{RedirectURIs: []string{"https://a.example/cb"}})
	if err != nil {
		f.Fatalf("register() error = %v, want nil", err)
	}
	for _, seed := range []string{
		issued,
		"https://claude.ai/oauth/mcp-oauth-client-metadata",
		"HTTPS:",
		"mt1.d.x.y",
		"mt1.d..",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, clientID string) {
		client, err := server.clients.resolve(t.Context(), clientID)
		if err != nil {
			return
		}
		if client.ID != clientID || len(client.RedirectURIs) == 0 ||
			slices.ContainsFunc(client.RedirectURIs, func(uri string) bool { return !redirectable(uri) }) {
			t.Errorf("resolve(%q) = %+v, want no client or one with somewhere a parent may be sent", clientID, client)
		}
	})
}

// knownServer is a server whose every client document is one the fetcher
// refuses: what a fuzzed identifier reaches is the registrations alone.
func knownServer(t testing.TB) (*Server, *documents) {
	t.Helper()

	fetcher := &documents{err: cimd.ErrUnreachable}
	known, _ := knownClients(t, ringOf(t, 'k'), fetcher)
	return &Server{Register: &registrar{clients: known, events: known.events, now: known.now}, clients: known}, fetcher
}

// fuzzedRedirect is where the one client of the fuzzed server sends a parent
// back to.
const fuzzedRedirect = "https://host.example/cb"

// googleStandIn answers as Google does, with no network: the address it sends
// a parent to carries the request, and the code "allowed" is worth a grant.
type googleStandIn struct{}

func (googleStandIn) AuthURL(state, _, _ string) string {
	return "https://accounts.example/auth?" + url.Values{"state": {state}}.Encode()
}

func (googleStandIn) Exchange(_ context.Context, code, _, _ string) (googleauth.Grant, error) {
	if code != "allowed" {
		return googleauth.Grant{}, googleauth.ErrCodeRefused
	}
	return googleauth.Grant{
		Subject: "a-subject", AccessToken: "an-access-token", RefreshToken: "a-refresh-token",
		Expiry: testDay.Add(time.Hour), Scopes: []string{googleauth.ScopeOpenID, googleauth.ScopeDriveFile},
	}, nil
}

// fuzzedServer is a server with one registered client, whose sign-in goes to
// the stand-in for Google.
func fuzzedServer(f *testing.F) (server *Server, clientID string) {
	f.Helper()

	server, err := New(&Settings{
		PublicURL: testIssuer,
		Scope:     "mcp",
		Seal:      ringOf(f, 'k'),
		Documents: &documents{err: cimd.ErrUnreachable},
		Logger:    zap.NewNop(),
		Google:    googleStandIn{},
		SiteURL:   testSite,
		Now:       func() time.Time { return testDay },
	})
	if err != nil {
		f.Fatalf("New() error = %v, want nil", err)
	}
	clientID, err = server.clients.register(registration{RedirectURIs: []string{fuzzedRedirect}})
	if err != nil {
		f.Fatalf("register() error = %v, want nil", err)
	}
	return server, clientID
}

// leadsOnlyWhereItMay fails an answer that is anything but a page or a way on
// to one of the places a sign-in may send the parent: Google, or the address
// the client registered, told who answered.
func leadsOnlyWhereItMay(t *testing.T, answer *httptest.ResponseRecorder, input string) {
	t.Helper()

	switch answer.Code {
	case http.StatusOK, http.StatusBadRequest:
		if answer.Header().Get("Location") != "" {
			t.Errorf("a page for %q leads to %q", input, answer.Header().Get("Location"))
		}
	case http.StatusFound:
		location := answer.Header().Get("Location")
		back := strings.HasPrefix(location, fuzzedRedirect+"?") && strings.Contains(location, "iss="+url.QueryEscape(testIssuer))
		if !back && !strings.HasPrefix(location, "https://accounts.example/auth?") {
			t.Errorf("%q leads to %q, want Google or the client's own address with the issuer", input, location)
		}
	default:
		t.Errorf("status = %d for %q, want a page or a redirect", answer.Code, input)
	}
}

// Whatever an authorization request asks, it is answered with a page or sent
// on to Google or to the address the client registered — never anywhere else,
// and never with a failure of the server's.
func FuzzAuthorize(f *testing.F) {
	server, clientID := fuzzedServer(f)
	valid := url.Values{
		"response_type": {"code"}, "client_id": {clientID}, "redirect_uri": {fuzzedRedirect},
		"code_challenge": {"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"}, "code_challenge_method": {"S256"},
		"state": {"s"}, "resource": {testIssuer + "/mcp"}, "scope": {"mcp"},
	}.Encode()
	for _, seed := range []string{
		valid,
		valid + "&resource=https://other.example/mcp",
		valid + "&redirect_uri=https://attacker.example/cb",
		strings.Replace(valid, "scope=mcp", "scope=admin", 1),
		"response_type=code&client_id=https://client.example/meta",
		"%zz",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, query string) {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/oauth/authorize", http.NoBody)
		request.URL.RawQuery = query
		answer := httptest.NewRecorder()
		server.Authorize.ServeHTTP(answer, request)
		leadsOnlyWhereItMay(t, answer, query)
	})
}

// Whatever Google's answer carries, and whatever cookie comes with it, the
// callback answers with a page or sends the parent back to the address the
// client registered — never anywhere else, and never with a failure of the
// server's.
func FuzzCallback(f *testing.F) {
	server, clientID := fuzzedServer(f)
	request := &flight{
		Client: digestOf(clientID), Registration: registrationDCR, RedirectURI: fuzzedRedirect, State: "s",
		Challenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", Resource: testIssuer + "/mcp", Scope: "mcp",
		Verifier: "v", Nonce: "n", Cookie: digestOf("the-cookie"), StartedAt: testDay.Unix(),
	}
	sealed, err := server.flow.sealFlight(request)
	if err != nil {
		f.Fatalf("sealFlight() error = %v, want nil", err)
	}
	state := url.Values{"state": {sealed}}.Encode()
	for _, seed := range []struct{ query, cookie string }{
		{state + "&code=allowed", "the-cookie"},
		{state + "&code=refused", "the-cookie"},
		{state + "&error=access_denied", "the-cookie"},
		{state + "&code=allowed", "another-cookie"},
		{state + "&state=again", "the-cookie"},
		{"state=mt1.s.x.y&code=allowed", "the-cookie"},
		{"", ""},
	} {
		f.Add(seed.query, seed.cookie)
	}

	f.Fuzz(func(t *testing.T, query, cookie string) {
		back := httptest.NewRequestWithContext(t.Context(), http.MethodGet, CallbackPath, http.NoBody)
		back.URL.RawQuery = query
		back.Header.Set("Cookie", csrfCookie+"="+cookie)
		answer := httptest.NewRecorder()
		server.Callback.ServeHTTP(answer, back)
		leadsOnlyWhereItMay(t, answer, query)
		if strings.HasPrefix(answer.Header().Get("Location"), "https://accounts.example/") {
			t.Errorf("the callback for %q sends the parent to Google again", query)
		}
	})
}
