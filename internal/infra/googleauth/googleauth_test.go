package googleauth_test

import (
	"errors"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
)

// What a sign-in sends Google and what it proves at the exchange, and where
// Google sends the parent back to.
const (
	callback = "https://mcp.example/oauth/callback"
	state    = "mt1.s.KID.the-sealed-request"
	verifier = "a-verifier-of-forty-three-characters-or-more-000"
	nonce    = "a-nonce-of-this-request"
)

// someDay is when these sign-ins happen.
var someDay = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

// newGoogle stands in for Google, with ID tokens dated by the clock of these
// cases.
func newGoogle(t *testing.T) *googletest.Server {
	t.Helper()

	return googletest.New(t, func() time.Time { return someDay })
}

// codeIn is the code in the address Google sent the parent back to.
func codeIn(t *testing.T, back string) string {
	t.Helper()

	parsed, err := url.Parse(back)
	if err != nil {
		t.Fatalf("the address back %q does not parse: %v", back, err)
	}
	return parsed.Query().Get("code")
}

// settings are what a sign-in against the fake is built from.
func settings(google *googletest.Server) *googleauth.Settings {
	return &googleauth.Settings{
		ClientID:     googletest.ClientID,
		ClientSecret: googletest.ClientSecret,
		RedirectURL:  callback,
		Endpoints:    google.Endpoints(),
		Now:          func() time.Time { return someDay },
	}
}

// signInAt builds a sign-in against the fake.
func signInAt(t *testing.T, google *googletest.Server) googleauth.SignIn {
	t.Helper()

	signIn, err := googleauth.New(settings(google))
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	return signIn
}

// The parent is sent to Google with both scopes, for a refresh token that
// Google gives only with its consent screen shown, and with what binds the
// answer to this request: the state, the challenge of the verifier and the
// nonce.
func TestTheParentIsSentToGoogleForBothScopes(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	address, err := url.Parse(signInAt(t, google).AuthURL(state, verifier, nonce))
	if err != nil {
		t.Fatalf("AuthURL() does not parse: %v", err)
	}
	if got := address.Scheme + "://" + address.Host + address.Path; got != google.URL+"/auth" {
		t.Errorf("AuthURL() leads to %q, want %q", got, google.URL+"/auth")
	}
	query := address.Query()
	for parameter, want := range map[string]string{
		"response_type":         "code",
		"client_id":             googletest.ClientID,
		"redirect_uri":          callback,
		"scope":                 "openid https://www.googleapis.com/auth/drive.file",
		"state":                 state,
		"access_type":           "offline",
		"prompt":                "consent",
		"code_challenge":        googletest.ChallengeOf(verifier),
		"code_challenge_method": "S256",
		"nonce":                 nonce,
	} {
		if got := query.Get(parameter); got != want {
			t.Errorf("AuthURL() %s = %q, want %q", parameter, got, want)
		}
	}
	if query.Has("client_secret") {
		t.Error("AuthURL() carries the client secret, want it sent to the token endpoint alone")
	}
}

// A code the parent came back with is worth the grant: the tokens Drive is
// called with, dated by this clock, the scopes granted, and who signed in.
func TestACodeIsWorthTheGrant(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signInAt(t, google)
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier, nonce)))

	grant, err := signIn.Exchange(t.Context(), code, verifier, nonce)
	if err != nil {
		t.Fatalf("Exchange() error = %v, want the grant", err)
	}
	if grant.Subject != googletest.Subject || grant.AccessToken != googletest.AccessToken || grant.RefreshToken != googletest.RefreshToken {
		t.Errorf("Exchange() = %+v, want the fake's subject and tokens", grant)
	}
	if want := someDay.Add(googletest.ExpiresIn * time.Second); !grant.Expiry.Equal(want) {
		t.Errorf("Exchange() expiry = %v, want %v", grant.Expiry, want)
	}
	if want := []string{googleauth.ScopeOpenID, googleauth.ScopeDriveFile}; !slices.Equal(grant.Scopes, want) {
		t.Errorf("Exchange() scopes = %q, want %q", grant.Scopes, want)
	}
}

// Google lets a parent untick a permission; the grant says what was granted
// and leaves the verdict to the caller.
func TestAGrantSaysWhichScopesWereGranted(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signInAt(t, google)
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier, nonce)))
	google.Misbehave(&googletest.Answer{Scope: "openid"})

	grant, err := signIn.Exchange(t.Context(), code, verifier, nonce)
	if err != nil {
		t.Fatalf("Exchange() error = %v, want the grant", err)
	}
	if !slices.Equal(grant.Scopes, []string{googleauth.ScopeOpenID}) {
		t.Errorf("Exchange() scopes = %q, want openid alone", grant.Scopes)
	}
}

// Every way an exchange goes wrong ends in the refusal that names it, and the
// refusal repeats nothing Google said in its own words.
func TestAnExchangeGoneWrongIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		how      googletest.Answer
		code     string
		verifier string
		nonce    string
		want     error
	}{
		{name: "an expired code", how: googletest.Answer{Status: 400, ErrorCode: "invalid_grant"}, want: googleauth.ErrCodeRefused},
		{name: "a code never issued", code: "another-code", want: googleauth.ErrCodeRefused},
		{name: "another verifier", verifier: verifier + "x", want: googleauth.ErrCodeRefused},
		{name: "a refusal with no words", how: googletest.Answer{Status: 400}, want: googleauth.ErrCodeRefused},
		{name: "our client refused", how: googletest.Answer{Status: 401, ErrorCode: "invalid_client"}, want: googleauth.ErrClient},
		{name: "our client not let use a code", how: googletest.Answer{Status: 400, ErrorCode: "unauthorized_client"}, want: googleauth.ErrClient},
		{name: "Google asking for a pause", how: googletest.Answer{Status: 429}, want: googleauth.ErrUnavailable},
		{name: "Google failing", how: googletest.Answer{Status: 503, ErrorCode: "temporarily_unavailable"}, want: googleauth.ErrUnavailable},
		{name: "no ID token", how: googletest.Answer{NoIDToken: true}, want: googleauth.ErrIdentity},
		{name: "a signature nobody published", how: googletest.Answer{SignedBy: googletest.StrangersKey(), KeyID: "stranger"}, want: googleauth.ErrIdentity},
		{name: "a stranger's key under Google's id", how: googletest.Answer{SignedBy: googletest.StrangersKey(), KeyID: "published"}, want: googleauth.ErrIdentity},
		{name: "another issuer", how: googletest.Answer{Issuer: "https://accounts.example"}, want: googleauth.ErrIdentity},
		{name: "another client's token", how: googletest.Answer{Audience: "another.apps.googleusercontent.com"}, want: googleauth.ErrIdentity},
		{name: "another request's nonce", how: googletest.Answer{Nonce: "another-nonce"}, want: googleauth.ErrIdentity},
		{name: "an ID token expired past the skew", how: googletest.Answer{ExpiresAt: someDay.Add(-61 * time.Second)}, want: googleauth.ErrIdentity},
		{name: "no refresh token", how: googletest.Answer{NoRefresh: true}, want: googleauth.ErrNoRefresh},
		{name: "no lifetime for the access token", how: googletest.Answer{NoLifetime: true}, want: googleauth.ErrUnavailable},
		{name: "nobody who signed in", how: googletest.Answer{NoSubject: true}, want: googleauth.ErrIdentity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			google := newGoogle(t)
			signIn := signInAt(t, google)
			code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier, nonce)))
			google.Misbehave(&tc.how)

			proof, expected := or(tc.verifier, verifier), or(tc.nonce, nonce)
			grant, err := signIn.Exchange(t.Context(), or(tc.code, code), proof, expected)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Exchange() = %+v, %v; want %v", grant, err, tc.want)
			}
			if strings.Contains(err.Error(), googletest.RefreshToken) || strings.Contains(err.Error(), googletest.AccessToken) {
				t.Errorf("Exchange() error = %q, want it to carry no token", err)
			}
		})
	}
}

// Google's clock and this one may disagree by a minute: an ID token Google
// still calls good is not refused for it.
func TestAnIDTokenIsJudgedWithAMinuteOfSkew(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signInAt(t, google)
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier, nonce)))
	google.Misbehave(&googletest.Answer{ExpiresAt: someDay.Add(-59 * time.Second)})

	if _, err := signIn.Exchange(t.Context(), code, verifier, nonce); err != nil {
		t.Errorf("Exchange() error = %v, want an ID token 59 s past its end accepted", err)
	}
}

// A Google nobody can reach is Google not answering, not a refused code.
func TestAGoogleThatDoesNotAnswerIsUnavailable(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signInAt(t, google)
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier, nonce)))
	google.Close()

	if _, err := signIn.Exchange(t.Context(), code, verifier, nonce); !errors.Is(err, googleauth.ErrUnavailable) {
		t.Errorf("Exchange() error = %v, want %v", err, googleauth.ErrUnavailable)
	}
}

// A sign-in given a part it cannot work without refuses to be built and says
// which.
func TestASignInWithSomethingMissingIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		spoil func(*googleauth.Settings)
		want  string
	}{
		{"no client", func(s *googleauth.Settings) { s.ClientID = "" }, "ClientID"},
		{"no secret", func(s *googleauth.Settings) { s.ClientSecret = "" }, "ClientSecret"},
		{"no way back", func(s *googleauth.Settings) { s.RedirectURL = "" }, "RedirectURL"},
		{"no keys", func(s *googleauth.Settings) { s.Endpoints.Keys = "" }, "Endpoints"},
		{"no way to end a grant", func(s *googleauth.Settings) { s.Endpoints.Revoke = "" }, "Endpoints"},
		{"no clock", func(s *googleauth.Settings) { s.Now = nil }, "Now"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			given := settings(newGoogle(t))
			tc.spoil(given)
			_, err := googleauth.New(given)
			if !errors.Is(err, googleauth.ErrSettings) || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("New() error = %v, want ErrSettings naming %s", err, tc.want)
			}
		})
	}
}

// or is the value, or the fallback when there is none.
func or(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// A secret Google does not know is the service's own client refused — a fault
// of the deployment's — and never a code the parent should try again with.
func TestAWrongSecretIsOurClientRefused(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	given := settings(google)
	given.ClientSecret = "a-secret-rotated-away"
	signIn, err := googleauth.New(given)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier, nonce)))

	if _, err := signIn.Exchange(t.Context(), code, verifier, nonce); !errors.Is(err, googleauth.ErrClient) {
		t.Errorf("Exchange() error = %v, want %v", err, googleauth.ErrClient)
	}
}
