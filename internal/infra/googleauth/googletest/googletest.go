// Package googletest stands in for Google's sign-in, as far as the service can
// see it: a token endpoint that holds a code to the request it was issued
// for and renews the grant it gave, the endpoint that says whom an access
// token was issued for, the endpoint that ends a grant, and a parent who
// allows or declines what a request asks for.
//
// It is test code that lives in a package rather than a test file, because
// the tests of more than one package sign a parent in, and a test file cannot
// be shared between packages.
package googletest

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// The client the service is at this Google, and who signs in. The tokens are
// made up, and nothing of Google's.
//
//nolint:gosec // made-up tokens of the stand-in, which no Google ever issued
const (
	ClientID     = "mathtrail.apps.googleusercontent.com"
	ClientSecret = "a-secret-for-tests"
	Subject      = "110169484474386276334"
	AccessToken  = "ya29.an-access-token"
	RefreshToken = "1//a-refresh-token"
	// RenewedAccessToken is the access token a renewal gives.
	RenewedAccessToken = "ya29.a-renewed-access-token"
	// RotatedRefreshToken is the refresh token a renewal gives in place of
	// the one it was asked with, when the answer says to rotate it.
	RotatedRefreshToken = "1//a-rotated-refresh-token"
	// ExpiresIn is how many seconds an access token is good for.
	ExpiresIn = 3599
)

// Answer is how this Google answers, which a case changes to make it misbehave
// in one way. The zero Answer is Google behaving.
type Answer struct {
	// Status and ErrorCode answer an exchange or a renewal with a refusal
	// instead.
	Status    int
	ErrorCode string
	// Scope is the scope granted in place of the one asked for.
	Scope string
	// Lifetime is how many seconds an access token is good for, in place of
	// ExpiresIn.
	Lifetime int
	// NoRefresh and NoLifetime leave the refresh token or the access token's
	// lifetime out.
	NoRefresh  bool
	NoLifetime bool
	// Rotate answers a renewal with another refresh token.
	Rotate bool
	// RevokeStatus and RevokeError answer a revocation with a refusal instead.
	RevokeStatus int
	RevokeError  string
	// InfoStatus answers the question of whom a token is for with a refusal
	// of that status instead, and InfoNotJSON with a page that is no answer
	// of Google's at all.
	InfoStatus  int
	InfoNotJSON bool
	// Audience names another client as the one the token was issued to,
	// NoAudience leaves that client out, and NoSubject leaves out the account
	// the token was issued for.
	Audience   string
	NoAudience bool
	NoSubject  bool
	// Meanwhile runs while a token request is at this Google, before it is
	// answered: whatever a case needs to happen on the way, such as time
	// passing.
	Meanwhile func()
}

// asked is what a sign-in asked Google for, as this Google keeps it for the
// code it issued.
type asked struct {
	challenge, redirectURI, scope string
}

// Server is the stand-in for Google. It gives one grant, whose tokens are the
// constants above, and a revocation ends it until the next exchange gives it
// again.
type Server struct {
	*httptest.Server
	t   testing.TB
	now func() time.Time

	mu          sync.Mutex
	codes       map[string]asked
	answer      Answer
	ended       bool
	renewals    int
	questions   int
	revocations []string
}

// New starts a stand-in for Google, which dates what it says of its tokens by
// the clock given, and stops it when the test ends.
func New(t testing.TB, now func() time.Time) *Server {
	t.Helper()

	google := &Server{t: t, now: now, codes: map[string]asked{}}
	routes := http.NewServeMux()
	routes.HandleFunc("POST /token", google.token)
	routes.HandleFunc("POST /tokeninfo", google.tokenInfo)
	routes.HandleFunc("POST /revoke", google.revoke)
	google.Server = httptest.NewServer(routes)
	t.Cleanup(google.Close)
	return google
}

// Endpoints are where this Google is.
func (g *Server) Endpoints() googleauth.Endpoints {
	return googleauth.Endpoints{
		Auth:      g.URL + "/auth",
		Token:     g.URL + "/token",
		TokenInfo: g.URL + "/tokeninfo",
		Revoke:    g.URL + "/revoke",
	}
}

// Renewals is how many renewals this Google has answered with an access
// token.
func (g *Server) Renewals() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.renewals
}

// Questions is how many times this Google was asked whom an access token was
// issued for.
func (g *Server) Questions() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.questions
}

// Revocations are the tokens this Google was asked to end a grant with, in
// the order it was asked.
func (g *Server) Revocations() []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return slices.Clone(g.revocations)
}

// Allow is the parent allowing the sign-in an address asks for: Google issues
// a code for the request, and sends the browser back to the request's
// redirect URI with the code and the state. That address is returned.
func (g *Server) Allow(address string) string {
	g.t.Helper()

	query := g.read(address)
	g.mu.Lock()
	code := "a-code-" + strconv.Itoa(len(g.codes)+1)
	g.codes[code] = asked{
		challenge:   query.Get("code_challenge"),
		redirectURI: query.Get("redirect_uri"),
		scope:       query.Get("scope"),
	}
	g.mu.Unlock()
	return query.Get("redirect_uri") + "?" + url.Values{"code": {code}, "state": {query.Get("state")}}.Encode()
}

// Deny is the parent declining at Google: the browser goes back to the
// request's redirect URI with access_denied and the state. That address is
// returned.
func (g *Server) Deny(address string) string {
	g.t.Helper()

	query := g.read(address)
	return query.Get("redirect_uri") + "?" + url.Values{"error": {"access_denied"}, "state": {query.Get("state")}}.Encode()
}

// Misbehave sets how the exchanges from now on are answered.
func (g *Server) Misbehave(how *Answer) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.answer = *how
}

// read is the query of an address a parent was sent to sign in at.
func (g *Server) read(address string) url.Values {
	g.t.Helper()

	parsed, err := url.Parse(address)
	if err != nil {
		g.t.Fatalf("the sign-in address %q does not parse: %v", address, err)
	}
	return parsed.Query()
}

func (g *Server) token(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	how := g.answer
	g.mu.Unlock()
	if how.Meanwhile != nil {
		how.Meanwhile()
	}

	switch {
	case how.Status != 0:
		writeJSON(w, how.Status, map[string]string{"error": how.ErrorCode})
	case r.PostFormValue("client_id") != ClientID, r.PostFormValue("client_secret") != ClientSecret:
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
	case r.PostFormValue("grant_type") == "authorization_code":
		g.exchange(w, r, &how)
	case r.PostFormValue("grant_type") == "refresh_token":
		g.renew(w, r, &how)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
	}
}

// exchange answers a code with the grant: a code this Google issued, for the
// address it was issued for and the verifier of its challenge. The grant it
// gives is a new one, whatever ended the last.
func (g *Server) exchange(w http.ResponseWriter, r *http.Request, how *Answer) {
	g.mu.Lock()
	request, issued := g.codes[r.PostFormValue("code")]
	g.mu.Unlock()

	if !issued || r.PostFormValue("redirect_uri") != request.redirectURI ||
		ChallengeOf(r.PostFormValue("code_verifier")) != request.challenge {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	g.mu.Lock()
	g.ended = false
	g.mu.Unlock()

	body := map[string]any{
		"access_token": AccessToken,
		"expires_in":   lifetime(how),
		"token_type":   "Bearer",
		"scope":        request.scope,
	}
	if how.Scope != "" {
		body["scope"] = how.Scope
	}
	if !how.NoRefresh {
		body["refresh_token"] = RefreshToken
	}
	if how.NoLifetime {
		delete(body, "expires_in")
	}
	writeJSON(w, http.StatusOK, body)
}

// renew answers a refresh token of the grant, while the grant lasts, with a
// new access token — and, when the answer says to rotate, a new refresh token.
func (g *Server) renew(w http.ResponseWriter, r *http.Request, how *Answer) {
	token := r.PostFormValue("refresh_token")
	g.mu.Lock()
	honoured := !g.ended && (token == RefreshToken || token == RotatedRefreshToken)
	if honoured {
		g.renewals++
	}
	g.mu.Unlock()

	if !honoured {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	body := map[string]any{
		"access_token": RenewedAccessToken,
		"expires_in":   lifetime(how),
		"token_type":   "Bearer",
		"scope":        googleauth.ScopeDriveFile,
	}
	if how.Rotate {
		body["refresh_token"] = RotatedRefreshToken
	}
	if how.NoLifetime {
		delete(body, "expires_in")
	}
	writeJSON(w, http.StatusOK, body)
}

// tokenInfo says whom an access token of the grant was issued for, and to
// which client, while the grant lasts. It takes the question as Google's own
// client library puts it — a form posted with the token in the Authorization
// header — and refuses any other question, and any other token, as Google
// refuses a token it does not know.
func (g *Server) tokenInfo(w http.ResponseWriter, r *http.Request) {
	token, bearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	asForm := r.Header.Get("Content-Type") == "application/x-www-form-urlencoded"
	g.mu.Lock()
	how := g.answer
	g.questions++
	ofTheGrant := bearer && asForm && !g.ended && (token == AccessToken || token == RenewedAccessToken)
	g.mu.Unlock()

	switch {
	case how.InfoStatus != 0:
		w.WriteHeader(how.InfoStatus)
		return
	case !ofTheGrant:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_token", "error_description": "Invalid Value"})
		return
	case how.InfoNotJSON:
		// The server names the page's type from its first bytes.
		_, _ = io.WriteString(w, "<html><body>a page that is no answer of Google's</body></html>")
		return
	}
	// Google writes every value as a string, numbers included.
	body := map[string]string{
		"azp":         ClientID,
		"aud":         ClientID,
		"sub":         Subject,
		"scope":       googleauth.ScopeDriveFile,
		"exp":         strconv.FormatInt(g.now().Add(ExpiresIn*time.Second).Unix(), 10),
		"expires_in":  strconv.Itoa(ExpiresIn),
		"access_type": "offline",
	}
	if how.Audience != "" {
		body["azp"], body["aud"] = how.Audience, how.Audience
	}
	if how.NoAudience {
		delete(body, "azp")
		delete(body, "aud")
	}
	if how.NoSubject {
		delete(body, "sub")
	}
	writeJSON(w, http.StatusOK, body)
}

// revoke ends the grant, given any token of it. A token of a grant already
// ended, or of none, is refused as Google refuses it.
func (g *Server) revoke(w http.ResponseWriter, r *http.Request) {
	token := r.PostFormValue("token")
	g.mu.Lock()
	how := g.answer
	g.revocations = append(g.revocations, token)
	ofTheGrant := !g.ended && slices.Contains([]string{AccessToken, RenewedAccessToken, RefreshToken, RotatedRefreshToken}, token)
	if how.RevokeStatus == 0 && ofTheGrant {
		g.ended = true
	}
	g.mu.Unlock()

	switch {
	case how.RevokeStatus != 0:
		writeJSON(w, how.RevokeStatus, map[string]string{"error": how.RevokeError})
	case !ofTheGrant:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_token"})
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// lifetime is how many seconds an access token of the answer is good for.
func lifetime(how *Answer) int {
	if how.Lifetime > 0 {
		return how.Lifetime
	}
	return ExpiresIn
}

// ChallengeOf is the S256 challenge of a verifier (RFC 7636).
func ChallengeOf(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
