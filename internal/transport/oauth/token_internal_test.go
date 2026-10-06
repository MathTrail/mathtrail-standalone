package oauthserver

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
)

// wantRefused holds an answer of the token or revocation endpoint to a refusal
// in the protocol's words, and the last line of the event to the reason.
func wantRefused(t *testing.T, h *signIn, answer *hostReply, status int, code, event, reason string) {
	t.Helper()

	if answer.status != status || answer.field("error") != code {
		t.Fatalf("answer = %d %v, want %d %s", answer.status, answer.fields, status, code)
	}
	if answer.field("error_description") == "" {
		t.Errorf("the refusal %s says nothing more, want a description", code)
	}
	if got := answer.header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	lines := h.lines(event)
	if len(lines) == 0 {
		t.Fatalf("no %s line, want one of the refusal", event)
	}
	if line := lines[len(lines)-1]; line["reason"] != reason {
		t.Errorf("%s reason = %v, want %s", event, line["reason"], reason)
	}
}

// A code a parent's sign-in ended in is worth the host's tokens, kept by
// nobody on the way: an access token for the resource, good for fifteen
// minutes, and a refresh token that lasts thirty days. Both carry whom the
// parent signed in as and the client they were issued to; the access token
// carries Google's access token and nothing more of Google's grant.
func TestACodeIsWorthTheHostsTokens(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	answer := h.exchange(t, h.signedInCode(t, client), client, nil)

	if answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/token = %d %v, want the host's tokens", answer.status, answer.fields)
	}
	if got := answer.header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if answer.field("token_type") != "Bearer" || answer.field("scope") != "mcp" || answer.fields["expires_in"] != 900.0 {
		t.Errorf("answer = %v, want a Bearer token for mcp, good for 900 seconds", answer.fields)
	}

	user := h.ring.UserID("google-drive:" + googletest.PermissionID)
	access := h.openedAccess(t, answer.field("access_token"))
	wantAccess := accessGrant{
		User: user, Client: digestOf(client), Resource: h.served.URL + "/mcp", Scope: "mcp",
		GoogleAccessToken: googletest.AccessToken, GoogleAccessExpiry: testDay.Add(googletest.ExpiresIn * time.Second).Unix(),
		IssuedAt: testDay.Unix(), ExpiresAt: testDay.Add(15 * time.Minute).Unix(),
	}
	if access != wantAccess {
		t.Errorf("the access token carries %+v, want %+v", access, wantAccess)
	}
	refresh := h.openedRefresh(t, answer.field("refresh_token"))
	wantRefresh := refreshGrant{
		User: user, Client: digestOf(client), Resource: h.served.URL + "/mcp", Scope: "mcp", SignedInAt: testDay.Unix(),
		GoogleAccessToken: googletest.AccessToken, GoogleAccessExpiry: testDay.Add(googletest.ExpiresIn * time.Second).Unix(),
		GoogleRefreshToken: googletest.RefreshToken, IssuedAt: testDay.Unix(), ExpiresAt: testDay.Add(30 * 24 * time.Hour).Unix(),
	}
	if refresh != wantRefresh {
		t.Errorf("the refresh token carries %+v, want %+v", refresh, wantRefresh)
	}

	lines := h.lines(eventAuthToken)
	if len(lines) != 1 || lines[0]["outcome"] != "ok" || lines[0]["user"] != user ||
		lines[0]["registration"] != registrationDCR || lines[0]["resource"] != "given" {
		t.Errorf("auth_token lines = %v, want one ok line naming the user", lines)
	}
	noLineCarries(t, h, googletest.AccessToken, googletest.RefreshToken, googletest.PermissionID, answer.field("access_token"), answer.field("refresh_token"))
}

// A host need not repeat what the code already holds: OAuth 2.1 leaves the
// address out of the exchange, a host may leave the resource out, and a
// resource with a slash at its end is the same resource.
func TestAHostMayLeaveOutWhatTheCodeHolds(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		changed url.Values
		after   time.Duration
	}{
		{"no redirect_uri", url.Values{"redirect_uri": {""}}, 0},
		{"no resource", url.Values{"resource": {""}}, 0},
		{"the resource with a slash", url.Values{"resource": {"/mcp/"}}, 0},
		{"a code 59 seconds old", nil, 59 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			code := h.signedInCode(t, client)
			if resource, named := tc.changed["resource"]; named && resource[0] != "" {
				tc.changed = url.Values{"resource": {h.served.URL + resource[0]}}
			}
			h.clock.advance(tc.after)
			if answer := h.exchange(t, code, client, tc.changed); answer.status != http.StatusOK {
				t.Errorf("POST /oauth/token = %d %v, want the host's tokens", answer.status, answer.fields)
			}
		})
	}
}

// A code is worth tokens to the host it was issued for alone, within its
// minute: to the client it was issued to, with the verifier its challenge was
// made from, for the address it went to and the resource it was asked for.
// Anything else is refused in the protocol's words, and the line says why.
func TestACodeIsRefusedToAnybodyElse(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		changed func(h *signIn) url.Values
		after   time.Duration
		status  int
		code    string
		reason  string
	}{
		{"another verifier", changing("code_verifier", hostVerifier+"x"), 0, 400, "invalid_grant", "pkce"},
		{"no verifier", changing("code_verifier", ""), 0, 400, "invalid_request", "invalid_request"},
		{"another client", func(h *signIn) url.Values {
			return url.Values{"client_id": {h.register(hostRedirect, "Another host")}}
		}, 0, 400, "invalid_grant", "client"},
		{"no client", changing("client_id", ""), 0, 400, "invalid_request", "invalid_request"},
		{"another address", changing("redirect_uri", "https://host.example/oauth/other"), 0, 400, "invalid_grant", "redirect_uri"},
		{"a blank address", changing("redirect_uri", " "), 0, 400, "invalid_grant", "redirect_uri"},
		{"a code a minute old", nil, 61 * time.Second, 400, "invalid_grant", "expired"},
		{"a code dated two minutes ahead", nil, -2 * time.Minute, 400, "invalid_grant", "expired"},
		{"another resource", changing("resource", "https://other.example/mcp"), 0, 400, "invalid_target", "invalid_target"},
		{"two resources", func(h *signIn) url.Values {
			return url.Values{"resource": {h.served.URL + "/mcp", "https://other.example/mcp"}}
		}, 0, 400, "invalid_target", "invalid_target"},
		{"a code nobody issued", changing("code", "mt1.c.xxxxxx.yyyy"), 0, 400, "invalid_grant", "unknown_code"},
		{"no code", changing("code", ""), 0, 400, "invalid_request", "invalid_request"},
		{"another grant", changing("grant_type", "password"), 0, 400, "unsupported_grant_type", "unsupported_grant_type"},
		{"no grant", changing("grant_type", ""), 0, 400, "invalid_request", "invalid_request"},
		{"a parameter twice", func(*signIn) url.Values {
			return url.Values{"code_verifier": {hostVerifier, hostVerifier}}
		}, 0, 400, "invalid_request", "invalid_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			code := h.signedInCode(t, client)
			var changed url.Values
			if tc.changed != nil {
				changed = tc.changed(h)
			}
			h.clock.advance(tc.after)
			answer := h.exchange(t, code, client, changed)
			wantRefused(t, h, &answer, tc.status, tc.code, eventAuthToken, tc.reason)
			noLineCarries(t, h, googletest.AccessToken, googletest.RefreshToken, googletest.PermissionID, code)
		})
	}
}

// A client of its own document on the parent's computer is sent back to the
// port it chose at the sign-in, which its document cannot name, and its code
// is worth tokens for that port alone: the exchange names the address the code
// went to, port and all, or is refused.
func TestAPortChosenAtTheSignInIsTheCodesAlone(t *testing.T) {
	t.Parallel()

	const (
		listed  = "http://127.0.0.1/callback"
		chosen  = "http://127.0.0.1:52011/callback"
		another = "http://127.0.0.1:52012/callback"
	)
	h := newSignIn(t)
	h.documents.fetched = cimd.Fetched{Document: cimd.Document{
		ClientID: hostDocument, ClientName: hostName, RedirectURIs: []string{listed},
	}}
	h.documents.err = nil
	parent := h.browser(t)
	back := parent.get(h.google.Allow(h.toGoogleAt(parent, h.authorizeURL(hostDocument, url.Values{"redirect_uri": {chosen}}))))
	address, err := url.Parse(back.location)
	if err != nil || address.Scheme+"://"+address.Host+address.Path != chosen || address.Query().Get("code") == "" {
		t.Fatalf("the parent was sent back to %q, want a code at %s", back.location, chosen)
	}
	code := address.Query().Get("code")

	refused := h.exchange(t, code, hostDocument, url.Values{"redirect_uri": {another}})
	wantRefused(t, h, &refused, http.StatusBadRequest, "invalid_grant", eventAuthToken, "redirect_uri")
	if answer := h.exchange(t, code, hostDocument, url.Values{"redirect_uri": {chosen}}); answer.status != http.StatusOK {
		t.Errorf("POST /oauth/token at the port chosen = %d %v, want the host's tokens", answer.status, answer.fields)
	}
}

// A code is worth tokens only with a verifier of the shape RFC 7636 gives
// one — 43 to 128 letters, digits and -._~ — even when the code's challenge
// was made from another.
func TestAVerifierIsOfTheShapeTheRulesGiveIt(t *testing.T) {
	t.Parallel()

	allowed := "AZaz09-._~"
	for _, tc := range []struct {
		name     string
		verifier string
		taken    bool
	}{
		{"the shortest", strings.Repeat(allowed, 5)[:shortestVerifier], true},
		{"the longest", strings.Repeat(allowed, 13)[:longestVerifier], true},
		{"one character short", strings.Repeat("v", shortestVerifier-1), false},
		{"one character long", strings.Repeat("v", longestVerifier+1), false},
		{"a character the rules leave out", strings.Repeat("v", shortestVerifier) + "+", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			parent := h.browser(t)
			asked := h.authorizeURL(client, url.Values{"code_challenge": {googletest.ChallengeOf(tc.verifier)}})
			code := answerAt(t, parent.get(h.google.Allow(h.toGoogleAt(parent, asked)))).Get("code")

			answer := h.exchange(t, code, client, url.Values{"code_verifier": {tc.verifier}})
			if !tc.taken {
				wantRefused(t, h, &answer, http.StatusBadRequest, "invalid_request", eventAuthToken, "invalid_pkce")
				return
			}
			if answer.status != http.StatusOK {
				t.Errorf("POST /oauth/token = %d %v, want the host's tokens", answer.status, answer.fields)
			}
		})
	}
}

// changing is a change of one parameter of a request.
func changing(name, value string) func(*signIn) url.Values {
	return func(*signIn) url.Values { return url.Values{name: {value}} }
}

// Clients here are public. A client that tries to prove itself in a header is
// told so by the refusal a client with a secret expects, and the same request
// that names the client in its body is answered: the code is still good.
func TestAClientThatProvesItselfInAHeaderIsToldItIsPublic(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	code := h.signedInCode(t, client)
	form := url.Values{
		"grant_type": {"authorization_code"}, "code": {code}, "code_verifier": {hostVerifier}, "redirect_uri": {hostRedirect},
	}
	answer := h.asHost(t, "/oauth/token", form, http.Header{"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte(url.QueryEscape(client)+":"))}})

	wantRefused(t, h, &answer, http.StatusUnauthorized, "invalid_client", eventAuthToken, "invalid_client")
	if got, want := answer.header.Get("WWW-Authenticate"), `Basic realm="`+h.served.URL+`"`; got != want {
		t.Errorf("WWW-Authenticate = %q, want %q", got, want)
	}
	form.Set("client_id", client)
	if answer := h.asHost(t, "/oauth/token", form, nil); answer.status != http.StatusOK {
		t.Errorf("the same request with client_id in its body = %d %v, want the host's tokens", answer.status, answer.fields)
	}
}

// A request for tokens is a form, as the protocol has it, and no larger than
// one ever needs to be.
func TestATokenRequestIsASmallForm(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		contentType string
		body        string
		status      int
		reason      string
	}{
		{"JSON", "application/json", `{"grant_type":"authorization_code"}`, http.StatusBadRequest, "invalid_request"},
		{"no type", "", "grant_type=authorization_code", http.StatusBadRequest, "invalid_request"},
		{"too large", "application/x-www-form-urlencoded", "code=" + strings.Repeat("a", maxTokenForm), http.StatusRequestEntityTooLarge, "too_large"},
		{"not a form", "application/x-www-form-urlencoded", "code=%zz", http.StatusBadRequest, "invalid_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, h.served.URL+"/oauth/token", strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("NewRequest() error = %v, want nil", err)
			}
			request.Header.Set("Content-Type", tc.contentType)
			response, err := h.served.Client().Do(request)
			if err != nil {
				t.Fatalf("POST /oauth/token: %v", err)
			}
			_ = response.Body.Close()
			if response.StatusCode != tc.status {
				t.Errorf("status = %d, want %d", response.StatusCode, tc.status)
			}
			if lines := h.lines(eventAuthToken); len(lines) != 1 || lines[0]["reason"] != tc.reason {
				t.Errorf("auth_token lines = %v, want one refusal for %s", lines, tc.reason)
			}
		})
	}
}

// An access token ends three minutes before the Google token inside it, when
// that comes sooner than its fifteen minutes.
func TestAnAccessTokenEndsBeforeTheGoogleTokenInsideIt(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	h.google.Misbehave(&googletest.Answer{Lifetime: 600})
	answer := h.exchange(t, h.signedInCode(t, client), client, nil)

	if answer.status != http.StatusOK || answer.fields["expires_in"] != 420.0 {
		t.Fatalf("POST /oauth/token = %d %v, want an access token good for 420 seconds", answer.status, answer.fields)
	}
	if access := h.openedAccess(t, answer.field("access_token")); access.ExpiresAt != testDay.Add(7*time.Minute).Unix() {
		t.Errorf("the access token ends at %d, want %d", access.ExpiresAt, testDay.Add(7*time.Minute).Unix())
	}
}

// A line names a client as what it is, without fetching anything: a client
// registered here as registered, an address as a client of its own document,
// and anything else as unknown — never as a registration it is not.
func TestALineNamesAClientAsWhatItIs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		clientID     func(h *signIn) string
		registration string
	}{
		{"a client registered here", func(h *signIn) string { return h.register(hostRedirect, "Another host") }, registrationDCR},
		{"a client of its own document", func(*signIn) string { return hostDocument }, registrationCIMD},
		{"words", func(*signIn) string { return "mt1.d.not-a-registration" }, registrationUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			h.exchange(t, h.signedInCode(t, client), tc.clientID(h), nil)

			if lines := h.lines(eventAuthToken); len(lines) != 1 || lines[0]["registration"] != tc.registration {
				t.Errorf("auth_token lines = %v, want the client named %s", lines, tc.registration)
			}
		})
	}
}
