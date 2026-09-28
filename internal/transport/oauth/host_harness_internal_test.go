package oauthserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// hostReply is an answer as a host's server gets it from the token or the
// revocation endpoint: its status, its headers, and the fields of its body,
// when it has one.
type hostReply struct {
	status int
	header http.Header
	fields map[string]any
}

// field is a field of the answer's body as text, or nothing.
func (r *hostReply) field(name string) string {
	value, _ := r.fields[name].(string)
	return value
}

// asHost posts a form to an endpoint of the server as a host's own server
// does — with no browser and no cookie — with the headers given.
func (h *signIn) asHost(t *testing.T, path string, form url.Values, header http.Header) hostReply {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, h.served.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("NewRequest(%s) error = %v, want nil", path, err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for name, values := range header {
		request.Header[name] = values
	}
	response, err := h.served.Client().Do(request)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("reading the answer to POST %s: %v", path, err)
	}
	answered := hostReply{status: response.StatusCode, header: response.Header}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &answered.fields); err != nil {
			t.Fatalf("POST %s answered %q, which is not JSON: %v", path, body, err)
		}
	}
	return answered
}

// changedForm is a form with what a case changes in it: a value replaced, and
// an empty value left out.
func changedForm(form, changed url.Values) url.Values {
	for name, values := range changed {
		form[name] = values
		if len(values) == 1 && values[0] == "" {
			delete(form, name)
		}
	}
	return form
}

// signedInCode is the code a host is handed for a parent who signed in and
// allowed everything asked.
func (h *signIn) signedInCode(t *testing.T, clientID string) string {
	t.Helper()

	parent := h.browser(t)
	back := parent.get(h.google.Allow(h.toGoogle(parent, clientID)))
	return answerAt(t, back).Get("code")
}

// exchange trades a code for the host's tokens, as the host that asked for it
// does, with what a case changes in the request.
func (h *signIn) exchange(t *testing.T, code, clientID string, changed url.Values) hostReply {
	t.Helper()

	return h.asHost(t, "/oauth/token", changedForm(url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {hostVerifier},
		"client_id":     {clientID},
		"redirect_uri":  {hostRedirect},
		"resource":      {h.served.URL + "/mcp"},
	}, changed), nil)
}

// renew trades a refresh token for new tokens, as the host that holds it
// does, with what a case changes in the request.
func (h *signIn) renew(t *testing.T, refreshToken, clientID string, changed url.Values) hostReply {
	t.Helper()

	return h.asHost(t, "/oauth/token", changedForm(url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
		"resource":      {h.served.URL + "/mcp"},
	}, changed), nil)
}

// revoke ends a host's grant with one of its tokens, as the host that holds it
// does, with what a case changes in the request.
func (h *signIn) revoke(t *testing.T, token, clientID string, changed url.Values) hostReply {
	t.Helper()

	return h.asHost(t, "/oauth/revoke", changedForm(url.Values{
		"token":     {token},
		"client_id": {clientID},
	}, changed), nil)
}

// tokensFor is what a host holds once a parent signed in: its access token
// and its refresh token.
func (h *signIn) tokensFor(t *testing.T, clientID string) (access, refresh string) {
	t.Helper()

	answer := h.exchange(t, h.signedInCode(t, clientID), clientID, nil)
	if answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/token = %d %v, want the host's tokens", answer.status, answer.fields)
	}
	return answer.field("access_token"), answer.field("refresh_token")
}

// openedRefresh is a refresh token opened as the token endpoint opens it.
func (h *signIn) openedRefresh(t *testing.T, token string) refreshGrant {
	t.Helper()

	var grant refreshGrant
	if err := openFor(h.ring, seal.PurposeRefresh, token, h.served.URL, &grant); err != nil {
		t.Fatalf("the refresh token does not open as a refresh token of this issuer: %v", err)
	}
	return grant
}

// openedAccess is an access token opened as the resource opens it.
func (h *signIn) openedAccess(t *testing.T, token string) accessGrant {
	t.Helper()

	var grant accessGrant
	if err := openFor(h.ring, seal.PurposeAccess, token, h.served.URL, &grant); err != nil {
		t.Fatalf("the access token does not open as an access token of this issuer: %v", err)
	}
	return grant
}

// openFor opens a sealed value of a purpose, bound to an issuer, into what it
// carries.
func openFor(ring *seal.KeyRing, purpose seal.Purpose, value, issuer string, carried any) error {
	plain, err := ring.Open(purpose, value, issuer)
	if err != nil {
		return err
	}
	return json.Unmarshal(plain, carried)
}

// sealedFor seals what a token carries for a purpose, bound to an issuer, as
// only this server's key could: a token a case makes with what it wants in it.
func sealedFor(t testing.TB, ring *seal.KeyRing, purpose seal.Purpose, issuer string, carried any) string {
	t.Helper()

	plain, err := json.Marshal(carried)
	if err != nil {
		t.Fatalf("encoding a token: %v", err)
	}
	sealed, err := ring.Seal(purpose, plain, issuer)
	if err != nil {
		t.Fatalf("Seal() error = %v, want nil", err)
	}
	return sealed
}
