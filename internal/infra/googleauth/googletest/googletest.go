// Package googletest stands in for Google's sign-in, as far as the service can
// see it: a token endpoint that holds a code to the request it was issued
// for, the keys its ID tokens are signed with, and a parent who allows or
// declines what a request asks for.
//
// It is test code that lives in a package rather than a test file, because
// the tests of more than one package sign a parent in, and a test file cannot
// be shared between packages.
package googletest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"

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
	// ExpiresIn is how many seconds an access token is good for.
	ExpiresIn = 3599
)

// publishedKeyID names the key this Google publishes.
const publishedKeyID = "published"

// The keys of these sign-ins: the one this Google publishes, and one nobody
// published. Each takes a moment to make, so each is made once.
var (
	publishedKey = sync.OnceValue(newKey)
	strangersKey = sync.OnceValue(newKey)
)

func newKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return key
}

// StrangersKey is a key this Google never published.
func StrangersKey() *rsa.PrivateKey { return strangersKey() }

// Answer is how this Google answers an exchange, which a case changes to make
// it misbehave in one way. The zero Answer is Google behaving.
type Answer struct {
	// Status and ErrorCode answer with a refusal instead.
	Status    int
	ErrorCode string
	// Scope is the scope granted in place of the one asked for.
	Scope string
	// NoRefresh, NoIDToken and NoLifetime leave the refresh token, the ID
	// token or the access token's lifetime out.
	NoRefresh  bool
	NoIDToken  bool
	NoLifetime bool
	// NoSubject leaves out of the ID token who signed in.
	NoSubject bool
	// SignedBy and KeyID sign the ID token with another key, under a name.
	SignedBy *rsa.PrivateKey
	KeyID    string
	// Issuer, Audience and Nonce put other values in the ID token.
	Issuer   string
	Audience string
	Nonce    string
	// ExpiresAt is when the ID token stops being good.
	ExpiresAt time.Time
}

// asked is what a sign-in asked Google for, as this Google keeps it for the
// code it issued.
type asked struct {
	challenge, nonce, redirectURI, scope string
}

// Server is the stand-in for Google.
type Server struct {
	*httptest.Server
	t   testing.TB
	now func() time.Time

	mu     sync.Mutex
	codes  map[string]asked
	answer Answer
}

// New starts a stand-in for Google, whose ID tokens are dated by the clock
// given, and stops it when the test ends.
func New(t testing.TB, now func() time.Time) *Server {
	t.Helper()

	google := &Server{t: t, now: now, codes: map[string]asked{}}
	routes := http.NewServeMux()
	routes.HandleFunc("POST /token", google.token)
	routes.HandleFunc("GET /keys", google.keys)
	google.Server = httptest.NewServer(routes)
	t.Cleanup(google.Close)
	return google
}

// Endpoints are where this Google is.
func (g *Server) Endpoints() googleauth.Endpoints {
	return googleauth.Endpoints{
		Issuer: g.URL,
		Auth:   g.URL + "/auth",
		Token:  g.URL + "/token",
		Keys:   g.URL + "/keys",
	}
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
		nonce:       query.Get("nonce"),
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
	request, issued := g.codes[r.PostFormValue("code")]
	g.mu.Unlock()

	switch {
	case how.Status != 0:
		writeJSON(w, how.Status, map[string]string{"error": how.ErrorCode})
		return
	case r.PostFormValue("grant_type") != "authorization_code",
		r.PostFormValue("client_id") != ClientID, r.PostFormValue("client_secret") != ClientSecret:
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	case !issued, r.PostFormValue("redirect_uri") != request.redirectURI,
		ChallengeOf(r.PostFormValue("code_verifier")) != request.challenge:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}

	body := map[string]any{
		"access_token": AccessToken,
		"expires_in":   ExpiresIn,
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
	if !how.NoIDToken {
		body["id_token"] = g.idToken(request, &how)
	}
	writeJSON(w, http.StatusOK, body)
}

// idToken is the ID token of a sign-in: this Google's, for the service's
// client, good for an hour and carrying the request's nonce, unless the
// answer says otherwise.
func (g *Server) idToken(request asked, how *Answer) string {
	claims := map[string]any{
		"iss":   g.URL,
		"aud":   ClientID,
		"sub":   Subject,
		"iat":   g.now().Unix(),
		"exp":   g.now().Add(time.Hour).Unix(),
		"nonce": request.nonce,
	}
	for claim, value := range map[string]string{"iss": how.Issuer, "aud": how.Audience, "nonce": how.Nonce} {
		if value != "" {
			claims[claim] = value
		}
	}
	if !how.ExpiresAt.IsZero() {
		claims["exp"] = how.ExpiresAt.Unix()
	}
	if how.NoSubject {
		delete(claims, "sub")
	}
	key, keyID := publishedKey(), publishedKeyID
	if how.SignedBy != nil {
		key, keyID = how.SignedBy, how.KeyID
	}
	return g.sign(key, keyID, claims)
}

func (g *Server) keys(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
		Key: &publishedKey().PublicKey, KeyID: publishedKeyID, Algorithm: string(jose.RS256), Use: "sig",
	}}})
}

// sign makes a JWT of the claims, signed with the key under the key id.
func (g *Server) sign(key *rsa.PrivateKey, keyID string, claims map[string]any) string {
	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       jose.JSONWebKey{Key: key, KeyID: keyID},
	}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		g.t.Errorf("NewSigner() error = %v, want nil", err)
		return ""
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		g.t.Errorf("encoding the claims: %v", err)
		return ""
	}
	signed, err := signer.Sign(payload)
	if err != nil {
		g.t.Errorf("Sign() error = %v, want nil", err)
		return ""
	}
	compact, err := signed.CompactSerialize()
	if err != nil {
		g.t.Errorf("CompactSerialize() error = %v, want nil", err)
		return ""
	}
	return compact
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
