package oauthserver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit/ratelimittest"
)

// An access token signs a request to the resource in as the account the
// parent signed in as, reached with the Google access token it carries, until
// it ends — and the Google token until its own end.
func TestAnAccessTokenSignsItsAccountIn(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	access, _ := h.tokensFor(t, client)

	account, ends, err := h.server.Account(t.Context(), access)
	if err != nil {
		t.Fatalf("Account() error = %v, want the account", err)
	}
	if want := h.ring.UserID("google-drive:" + googletest.PermissionID); account.ID != want {
		t.Errorf("Account() = %q, want %q", account.ID, want)
	}
	if account.Token() != googletest.AccessToken {
		t.Errorf("the account is reached with %q, want Google's access token", account.Token())
	}
	// The Google token's end travels with it, so that the store starts no
	// call to Drive it would not last through.
	if want := testDay.Add(googletest.ExpiresIn * time.Second); !account.Ends().Equal(want) {
		t.Errorf("the account's Google token ends at %v, want %v", account.Ends(), want)
	}
	if want := testDay.Add(15 * time.Minute); !ends.Equal(want) {
		t.Errorf("Account() ends at %v, want %v", ends, want)
	}
	if lines := h.lines(eventAuthBearer); len(lines) != 0 {
		t.Errorf("auth_bearer lines = %v, want none for a token that signs its account in", lines)
	}
}

// An access token signs nobody in unless this server sealed it, for itself, as
// an access token, under a key it still carries — and unless it has not
// ended, is for this resource and grants this scope. Its line says why, and
// the error carries nothing of the token.
func TestAnAccessTokenThatSignsNobodyIn(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		token  func(t *testing.T, h *signIn, access, refresh string) string
		after  time.Duration
		reason string
	}{
		{name: "words", token: func(*testing.T, *signIn, string, string) string { return "not-a-token" }, reason: "unreadable"},
		{name: "a refresh token", token: func(_ *testing.T, _ *signIn, _, refresh string) string { return refresh }, reason: "unreadable"},
		{name: "a token of another issuer", token: func(t *testing.T, h *signIn, _, _ string) string {
			return sealedFor(t, h.ring, seal.PurposeAccess, "https://other.example", h.anAccessGrant())
		}, reason: "unreadable"},
		{name: "a token under a key the server no longer carries", token: func(t *testing.T, h *signIn, _, _ string) string {
			return sealedFor(t, ringOf(t, 'r'), seal.PurposeAccess, h.served.URL, h.anAccessGrant())
		}, reason: "unknown_key"},
		{name: "a token that ended", token: func(_ *testing.T, _ *signIn, access, _ string) string { return access },
			after: 16*time.Minute + time.Second, reason: "expired"},
		{name: "a token for another resource", token: func(t *testing.T, h *signIn, _, _ string) string {
			grant := h.anAccessGrant()
			grant.Resource = "https://other.example/mcp"
			return sealedFor(t, h.ring, seal.PurposeAccess, h.served.URL, grant)
		}, reason: "audience"},
		{name: "a token for another scope", token: func(t *testing.T, h *signIn, _, _ string) string {
			grant := h.anAccessGrant()
			grant.Scope = "admin"
			return sealedFor(t, h.ring, seal.PurposeAccess, h.served.URL, grant)
		}, reason: "scope"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			access, refresh := h.tokensFor(t, client)
			token := tc.token(t, h, access, refresh)
			h.clock.advance(tc.after)

			_, _, err := h.server.Account(t.Context(), token)
			if !errors.Is(err, errNoAccount) {
				t.Fatalf("Account() error = %v, want %v", err, errNoAccount)
			}
			if strings.Contains(err.Error(), token) {
				t.Errorf("Account() error = %q, want it to carry nothing of the token", err)
			}
			if lines := h.lines(eventAuthBearer); len(lines) != 1 || lines[0]["outcome"] != "refused" || lines[0]["reason"] != tc.reason {
				t.Errorf("auth_bearer lines = %v, want one refusal for %s", lines, tc.reason)
			}
		})
	}
}

// anAccessGrant is what an access token of this server carries, for a case to
// change and seal.
func (h *signIn) anAccessGrant() *accessGrant {
	return &accessGrant{
		User: "a-user", Client: digestOf("a-client"), Resource: h.served.URL + "/mcp", Scope: "mcp",
		GoogleAccessToken: googletest.AccessToken, GoogleAccessExpiry: testDay.Add(time.Hour).Unix(),
		IssuedAt: testDay.Unix(), ExpiresAt: testDay.Add(15 * time.Minute).Unix(),
	}
}

// The clocks of two instances may disagree by a minute: an access token still
// signs its account in within the minute past its end.
func TestAnAccessTokenIsJudgedWithAMinuteOfSkew(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	access, _ := h.tokensFor(t, client)
	h.clock.advance(15*time.Minute + 59*time.Second)

	if _, _, err := h.server.Account(t.Context(), access); err != nil {
		t.Errorf("Account() error = %v, want a token 59 s past its end accepted", err)
	}
}

// A rotation of the key does not end a sign-in: a token sealed under the key
// that became the previous one still signs its account in and is renewed, and
// the tokens it is renewed with are sealed under the new key, for the same
// account. Once the old key is retired, its tokens sign nobody in, and the
// renewed ones still do.
func TestATokenOutlivesARotationOfItsKey(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	access, refresh := h.tokensFor(t, client)
	user := h.ring.UserID("google-drive:" + googletest.PermissionID)

	rotated := h.serverUnder(t, keyOf('n'), keyOf('k'))
	if account, _, err := rotated.Account(t.Context(), access); err != nil || account.ID != user {
		t.Fatalf("after the rotation Account() = %q, %v; want %q", account.ID, err, user)
	}
	renewed := serveForm(t, rotated.Token, url.Values{
		"grant_type": {"refresh_token"}, "refresh_token": {refresh}, "client_id": {client},
	})
	if renewed.Code != http.StatusOK {
		t.Fatalf("after the rotation POST /oauth/token = %d %s, want new tokens", renewed.Code, renewed.Body)
	}
	newAccess := fieldOf(t, renewed, "access_token")
	newKey := ringOfKeys(t, keyOf('n'), "").CurrentKeyID()
	if !strings.HasPrefix(newAccess, "mt1.a."+newKey+".") {
		t.Errorf("the renewed access token %q is not sealed under the new key %s", newAccess[:12], newKey)
	}
	if account, _, err := rotated.Account(t.Context(), newAccess); err != nil || account.ID != user {
		t.Errorf("the renewed access token signs in %q, %v; want the same account %q", account.ID, err, user)
	}

	retired := h.serverUnder(t, keyOf('n'), "")
	if _, _, err := retired.Account(t.Context(), access); !errors.Is(err, errNoAccount) {
		t.Errorf("once the old key is retired Account() error = %v, want %v", err, errNoAccount)
	}
	if _, _, err := retired.Account(t.Context(), newAccess); err != nil {
		t.Errorf("once the old key is retired the renewed token: Account() error = %v, want nil", err)
	}
}

// keyOf is a sealing key made of the byte given, as a deployment carries it.
func keyOf(fill byte) string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{fill}, seal.KeySize))
}

// ringOfKeys is a ring of a current key and a previous one.
func ringOfKeys(t *testing.T, current, previous string) *seal.KeyRing {
	t.Helper()

	ring, err := seal.NewKeyRing(current, previous)
	if err != nil {
		t.Fatalf("NewKeyRing() error = %v, want nil", err)
	}
	return ring
}

// serverUnder is the server of the case as another instance would be after a
// rotation: the same issuer, clock and Google, under the keys given.
func (h *signIn) serverUnder(t *testing.T, current, previous string) *Server {
	t.Helper()

	google, err := googleauth.New(&googleauth.Settings{
		ClientID: googletest.ClientID, ClientSecret: googletest.ClientSecret, RedirectURL: h.served.URL + CallbackPath,
		Endpoints: h.google.Endpoints(), Now: h.clock.Now,
	})
	if err != nil {
		t.Fatalf("googleauth.New() error = %v, want nil", err)
	}
	server, err := New(&Settings{
		PublicURL: h.served.URL, Scope: "mcp", Seal: ringOfKeys(t, current, previous), Documents: h.documents,
		Logger: zap.NewNop(), Renewals: ratelimittest.Roomy(t), Google: google, SiteURL: testSite, Now: h.clock.Now,
	})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	return server
}

// serveForm posts a form to a handler, as a host's server does.
func serveForm(t *testing.T, handler http.Handler, form url.Values) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	answer := httptest.NewRecorder()
	handler.ServeHTTP(answer, request)
	return answer
}

// fieldOf is a text field of a JSON answer.
func fieldOf(t *testing.T, answer *httptest.ResponseRecorder, name string) string {
	t.Helper()

	var fields map[string]any
	if err := json.Unmarshal(answer.Body.Bytes(), &fields); err != nil {
		t.Fatalf("the answer %q is not JSON: %v", answer.Body, err)
	}
	value, _ := fields[name].(string)
	return value
}
