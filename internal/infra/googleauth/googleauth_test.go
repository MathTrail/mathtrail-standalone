package googleauth_test

import (
	"errors"
	"net/http/httptest"
	"net/url"
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
)

// someDay is when these sign-ins happen.
var someDay = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

// newGoogle stands in for Google, which dates what it says of its tokens by
// the clock of these cases.
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

// The parent is sent to Google for the Drive alone — a request for more than
// one scope shows each but a sign-in's with a box Google leaves unticked — and
// for a refresh token that Google gives only with its consent screen shown,
// with what binds the answer to this request: the state and the challenge of
// the verifier.
func TestTheParentIsSentToGoogleForTheDriveAlone(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	address, err := url.Parse(signInAt(t, google).AuthURL(state, verifier))
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
		"scope":                 "https://www.googleapis.com/auth/drive.file",
		"state":                 state,
		"access_type":           "offline",
		"prompt":                "consent",
		"code_challenge":        googletest.ChallengeOf(verifier),
		"code_challenge_method": "S256",
	} {
		if got := query.Get(parameter); got != want {
			t.Errorf("AuthURL() %s = %q, want %q", parameter, got, want)
		}
	}
	if query.Has("client_secret") {
		t.Error("AuthURL() carries the client secret, want it sent to the token endpoint alone")
	}
	// Scopes granted before would be asked for again beside the Drive, and
	// bring the box back.
	if query.Has("include_granted_scopes") {
		t.Error("AuthURL() asks for the scopes granted before, want the Drive alone")
	}
}

// A code the parent came back with is worth the grant: the tokens Drive is
// called with, dated by this clock, and who signed in, as Drive names the
// account the access token reaches it as.
func TestACodeIsWorthTheGrant(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signInAt(t, google)
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier)))

	grant, err := signIn.Exchange(t.Context(), code, verifier)
	if err != nil {
		t.Fatalf("Exchange() error = %v, want the grant", err)
	}
	if grant.PermissionID != googletest.PermissionID || grant.AccessToken != googletest.AccessToken || grant.RefreshToken != googletest.RefreshToken {
		t.Errorf("Exchange() = %+v, want the fake's account at Drive and tokens", grant)
	}
	if want := someDay.Add(googletest.ExpiresIn * time.Second); !grant.Expiry.Equal(want) {
		t.Errorf("Exchange() expiry = %v, want %v", grant.Expiry, want)
	}
	if asked := google.Questions(); asked != 1 {
		t.Errorf("Questions() = %d, want 1: Drive asked once whose the token is", asked)
	}
}

// A grant refused for what the exchange brought — no refresh token, an access
// token with no lifetime, or no Drive — is refused before Drive is asked whose
// it is.
func TestARefusedGrantIsNotAskedAbout(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		how  googletest.Answer
		want error
	}{
		{"no refresh token", googletest.Answer{NoRefresh: true}, googleauth.ErrNoRefresh},
		{"no lifetime for the access token", googletest.Answer{NoLifetime: true}, googleauth.ErrUnavailable},
		{"the Drive left out", googletest.Answer{Scope: "openid"}, googleauth.ErrNoDrive},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			google := newGoogle(t)
			signIn := signInAt(t, google)
			code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier)))
			google.Misbehave(&tc.how)

			if _, err := signIn.Exchange(t.Context(), code, verifier); !errors.Is(err, tc.want) {
				t.Fatalf("Exchange() error = %v, want %v", err, tc.want)
			}
			if asked := google.Questions(); asked != 0 {
				t.Errorf("Questions() = %d, want 0: Drive asked nothing about a grant refused anyway", asked)
			}
		})
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
		{name: "no refresh token", how: googletest.Answer{NoRefresh: true}, want: googleauth.ErrNoRefresh},
		{name: "no lifetime for the access token", how: googletest.Answer{NoLifetime: true}, want: googleauth.ErrUnavailable},
		{name: "the Drive left out", how: googletest.Answer{Scope: "openid"}, want: googleauth.ErrNoDrive},
		{name: "the token called none of Google's", how: googletest.Answer{AboutStatus: 401, AboutReason: "authError"}, want: googleauth.ErrIdentity},
		{name: "Drive asking for a pause", how: googletest.Answer{AboutStatus: 403, AboutReason: "userRateLimitExceeded"}, want: googleauth.ErrUnavailable},
		{name: "Drive asking for a pause by the other status", how: googletest.Answer{AboutStatus: 429, AboutReason: "rateLimitExceeded"}, want: googleauth.ErrUnavailable},
		{name: "Drive failing to say whose the token is", how: googletest.Answer{AboutStatus: 503, AboutReason: "backendError"}, want: googleauth.ErrUnavailable},
		{name: "Drive kept from the account", how: googletest.Answer{AboutStatus: 403, AboutReason: "domainPolicy"}, want: googleauth.ErrIdentity},
		{name: "a refusal of another kind", how: googletest.Answer{AboutStatus: 400, AboutReason: "badRequest"}, want: googleauth.ErrIdentity},
		{name: "a page that is no answer of Drive's", how: googletest.Answer{AboutNotJSON: true}, want: googleauth.ErrUnavailable},
		{name: "nobody who signed in", how: googletest.Answer{NoPermissionID: true}, want: googleauth.ErrIdentity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			google := newGoogle(t)
			signIn := signInAt(t, google)
			code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier)))
			google.Misbehave(&tc.how)

			grant, err := signIn.Exchange(t.Context(), or(tc.code, code), or(tc.verifier, verifier))
			if !errors.Is(err, tc.want) {
				t.Fatalf("Exchange() = %+v, %v; want %v", grant, err, tc.want)
			}
			if strings.Contains(err.Error(), googletest.RefreshToken) || strings.Contains(err.Error(), googletest.AccessToken) {
				t.Errorf("Exchange() error = %q, want it to carry no token", err)
			}
		})
	}
}

// A Google nobody can reach is Google not answering, not a refused code.
func TestAGoogleThatDoesNotAnswerIsUnavailable(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signInAt(t, google)
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier)))
	google.Close()

	if _, err := signIn.Exchange(t.Context(), code, verifier); !errors.Is(err, googleauth.ErrUnavailable) {
		t.Errorf("Exchange() error = %v, want %v", err, googleauth.ErrUnavailable)
	}
}

// A Google that gave the grant but whose Drive cannot be asked whose the token
// is, is Google not answering too, and the failure — which names the address
// it failed to reach — repeats no token: the token went in a header.
func TestADriveThatCannotSayWhoseTheTokenIsIsUnavailable(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	nobody := httptest.NewServer(nil)
	nobody.Close()
	given := settings(google)
	given.Endpoints.About = nobody.URL + "/drive/v3/about"
	signIn, err := googleauth.New(given)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier)))

	_, err = signIn.Exchange(t.Context(), code, verifier)
	if !errors.Is(err, googleauth.ErrUnavailable) {
		t.Fatalf("Exchange() error = %v, want %v", err, googleauth.ErrUnavailable)
	}
	if !strings.Contains(err.Error(), nobody.URL) || strings.Contains(err.Error(), googletest.AccessToken) {
		t.Errorf("Exchange() error = %q, want the address it failed to reach and no token", err)
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
		{"nobody to ask whose Drive a token reaches", func(s *googleauth.Settings) { s.Endpoints.About = "" }, "Endpoints"},
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
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier)))

	if _, err := signIn.Exchange(t.Context(), code, verifier); !errors.Is(err, googleauth.ErrClient) {
		t.Errorf("Exchange() error = %v, want %v", err, googleauth.ErrClient)
	}
}
