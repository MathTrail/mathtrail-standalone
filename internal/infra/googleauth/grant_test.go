package googleauth_test

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
)

// signedIn is a sign-in against the fake that has already been given its
// grant.
func signedIn(t *testing.T, google *googletest.Server) googleauth.SignIn {
	t.Helper()

	signIn := signInAt(t, google)
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier)))
	if _, err := signIn.Exchange(t.Context(), code, verifier); err != nil {
		t.Fatalf("Exchange() error = %v, want the grant", err)
	}
	return signIn
}

// noTokenIn fails an error that repeats a token of the grant.
func noTokenIn(t *testing.T, err error) {
	t.Helper()

	for _, token := range []string{googletest.AccessToken, googletest.RenewedAccessToken, googletest.RefreshToken} {
		if strings.Contains(err.Error(), token) {
			t.Errorf("error = %q, want it to carry no token", err)
		}
	}
}

// A grant's refresh token is worth a new access token, dated by this clock,
// and the grant goes on with the refresh token it had: Google gives no other.
func TestAGrantIsRenewedWithItsRefreshToken(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	renewal, err := signedIn(t, google).Refresh(t.Context(), googletest.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v, want a new access token", err)
	}
	want := googleauth.Renewal{
		AccessToken:  googletest.RenewedAccessToken,
		Expiry:       someDay.Add(googletest.ExpiresIn * time.Second),
		RefreshToken: googletest.RefreshToken,
	}
	if renewal != want {
		t.Errorf("Refresh() = %+v, want %+v", renewal, want)
	}
}

// clock is a clock a case moves on, from another goroutine too.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) read() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *clock) advance(by time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(by)
}

// An access token's life is counted from the moment it was asked for, at the
// exchange and at a renewal alike. Google counts it from the moment it gave
// the token, which comes after, so a grant is never taken to last longer than
// it does — however long the answer took to arrive.
func TestAnAccessTokensLifeIsCountedFromTheRequest(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	moving := &clock{now: someDay}
	// Every answer takes half a minute on its way back.
	google.Misbehave(&googletest.Answer{Meanwhile: func() { moving.advance(30 * time.Second) }})
	asked := settings(google)
	asked.Now = moving.read
	signIn, err := googleauth.New(asked)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}

	sent := moving.read()
	code := codeIn(t, google.Allow(signIn.AuthURL(state, verifier)))
	grant, err := signIn.Exchange(t.Context(), code, verifier)
	if err != nil {
		t.Fatalf("Exchange() error = %v, want the grant", err)
	}
	if want := sent.Add(googletest.ExpiresIn * time.Second); !grant.Expiry.Equal(want) {
		t.Errorf("Exchange() expiry = %v, want %v, counted from the request", grant.Expiry, want)
	}

	sent = moving.read()
	renewal, err := signIn.Refresh(t.Context(), googletest.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v, want a new access token", err)
	}
	if want := sent.Add(googletest.ExpiresIn * time.Second); !renewal.Expiry.Equal(want) {
		t.Errorf("Refresh() expiry = %v, want %v, counted from the request", renewal.Expiry, want)
	}
}

// When Google answers a renewal with another refresh token, that one is what
// the grant goes on with.
func TestARenewalGoesOnWithTheRefreshTokenGoogleGave(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signedIn(t, google)
	google.Misbehave(&googletest.Answer{Rotate: true})

	renewal, err := signIn.Refresh(t.Context(), googletest.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v, want a new access token", err)
	}
	if renewal.RefreshToken != googletest.RotatedRefreshToken {
		t.Errorf("Refresh() refresh token = %q, want the one Google gave in its place", renewal.RefreshToken)
	}
}

// Every way a renewal goes wrong ends in the refusal that names it: a grant
// Google no longer honours has ended, and only a new sign-in gives another.
func TestARenewalGoneWrongIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		how   googletest.Answer
		token string
		want  error
	}{
		{name: "a refresh token Google never gave", token: "1//another-refresh-token", want: googleauth.ErrGrantEnded},
		{name: "a grant Google refuses", how: googletest.Answer{Status: 400, ErrorCode: "invalid_grant"}, want: googleauth.ErrGrantEnded},
		{name: "a refusal with no words", how: googletest.Answer{Status: 400}, want: googleauth.ErrGrantEnded},
		{name: "our client refused", how: googletest.Answer{Status: 401, ErrorCode: "invalid_client"}, want: googleauth.ErrClient},
		{name: "Google asking for a pause", how: googletest.Answer{Status: 429}, want: googleauth.ErrUnavailable},
		{name: "Google failing", how: googletest.Answer{Status: 500, ErrorCode: "internal_failure"}, want: googleauth.ErrUnavailable},
		{name: "no lifetime for the access token", how: googletest.Answer{NoLifetime: true}, want: googleauth.ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			google := newGoogle(t)
			signIn := signedIn(t, google)
			google.Misbehave(&tc.how)

			renewal, err := signIn.Refresh(t.Context(), or(tc.token, googletest.RefreshToken))
			if !errors.Is(err, tc.want) {
				t.Fatalf("Refresh() = %+v, %v; want %v", renewal, err, tc.want)
			}
			noTokenIn(t, err)
		})
	}
}

// A grant is ended with any of its tokens, the refresh token or an access
// token alike, and Google renews it no more.
func TestAGrantIsEndedWithAnyOfItsTokens(t *testing.T) {
	t.Parallel()

	for _, token := range []string{googletest.RefreshToken, googletest.AccessToken} {
		t.Run(token, func(t *testing.T) {
			t.Parallel()

			google := newGoogle(t)
			signIn := signedIn(t, google)
			if err := signIn.Revoke(t.Context(), token); err != nil {
				t.Fatalf("Revoke() error = %v, want nil", err)
			}
			if got := google.Revocations(); !slices.Equal(got, []string{token}) {
				t.Errorf("Google was asked to end the grant with %q, want %q alone", got, token)
			}
			if _, err := signIn.Refresh(t.Context(), googletest.RefreshToken); !errors.Is(err, googleauth.ErrGrantEnded) {
				t.Errorf("Refresh() after Revoke() error = %v, want %v", err, googleauth.ErrGrantEnded)
			}
		})
	}
}

// A grant that has already ended cannot be ended again: Google no longer
// honours its token, and says so.
func TestAGrantAlreadyEndedIsNotEndedAgain(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signedIn(t, google)
	if err := signIn.Revoke(t.Context(), googletest.RefreshToken); err != nil {
		t.Fatalf("Revoke() error = %v, want nil", err)
	}
	if err := signIn.Revoke(t.Context(), googletest.RefreshToken); !errors.Is(err, googleauth.ErrNotHonoured) {
		t.Errorf("Revoke() again error = %v, want %v", err, googleauth.ErrNotHonoured)
	}
}

// A revocation Google could not answer is Google not answering, which a
// caller may try again; one Google refused for another reason is an error
// that says so, and neither repeats the token.
func TestARevocationGoneWrongIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		how         googletest.Answer
		unavailable bool
	}{
		{name: "Google failing", how: googletest.Answer{RevokeStatus: 503}, unavailable: true},
		{name: "Google asking for a pause", how: googletest.Answer{RevokeStatus: 429}, unavailable: true},
		{name: "a request Google could not read", how: googletest.Answer{RevokeStatus: 400, RevokeError: "invalid_request"}},
		{name: "a refusal with no words", how: googletest.Answer{RevokeStatus: 403}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			google := newGoogle(t)
			signIn := signedIn(t, google)
			google.Misbehave(&tc.how)

			err := signIn.Revoke(t.Context(), googletest.RefreshToken)
			if err == nil || errors.Is(err, googleauth.ErrUnavailable) != tc.unavailable {
				t.Fatalf("Revoke() error = %v, want one that is ErrUnavailable: %v", err, tc.unavailable)
			}
			noTokenIn(t, err)
		})
	}
}

// A Google nobody can reach neither renews nor ends a grant, and says it did
// not answer.
func TestAGoneGoogleNeitherRenewsNorRevokes(t *testing.T) {
	t.Parallel()

	google := newGoogle(t)
	signIn := signedIn(t, google)
	google.Close()

	if _, err := signIn.Refresh(t.Context(), googletest.RefreshToken); !errors.Is(err, googleauth.ErrUnavailable) {
		t.Errorf("Refresh() error = %v, want %v", err, googleauth.ErrUnavailable)
	}
	if err := signIn.Revoke(t.Context(), googletest.RefreshToken); !errors.Is(err, googleauth.ErrUnavailable) {
		t.Errorf("Revoke() error = %v, want %v", err, googleauth.ErrUnavailable)
	}
}
