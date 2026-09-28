// Package googleauth signs a parent in with Google.
//
// Towards Google the service is an ordinary OAuth client. It sends the
// parent's browser to Google's consent screen with the two scopes it needs —
// who the parent is, and the files it creates in their Drive — and trades the
// code Google sends back for the grant. Who signed in is read from the ID
// token that comes with the grant, and only once the token's signature has
// been checked against the keys Google publishes: the token arrives over a
// connection the service opened itself, and the signature is what makes it
// Google's word rather than whatever answered there.
package googleauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// The scopes asked of Google: who the parent is, and the files the service
// creates in their Drive — no other file.
const (
	ScopeOpenID    = "openid"
	ScopeDriveFile = "https://www.googleapis.com/auth/drive.file"
)

const (
	// callTimeout is how long one call to Google may take: an exchange of a
	// code, or a fetch of the keys its ID tokens are signed with.
	callTimeout = 10 * time.Second
	// clockSkew is how far Google's clock and this one may disagree about
	// when an ID token stops being good.
	clockSkew = 60 * time.Second
)

// Endpoints are where a provider's sign-in is.
type Endpoints struct {
	// Issuer is who its ID tokens say issued them.
	Issuer string
	// Auth is where the parent's browser is sent to sign in.
	Auth string
	// Token is where a code is exchanged for the grant.
	Token string
	// Keys is where the keys its ID tokens are signed with are published.
	Keys string
}

// Accounts are Google's own endpoints, as its discovery document at
// https://accounts.google.com/.well-known/openid-configuration names them.
var Accounts = Endpoints{ //nolint:gosec // Google's published endpoints, not a credential
	Issuer: "https://accounts.google.com",
	Auth:   "https://accounts.google.com/o/oauth2/v2/auth",
	Token:  "https://oauth2.googleapis.com/token",
	Keys:   "https://www.googleapis.com/oauth2/v3/certs",
}

// Settings are what the sign-in is built from.
type Settings struct {
	// ClientID is the service's OAuth client at Google.
	ClientID string
	// ClientSecret is that client's secret. It is a secret: it is sent to
	// the token endpoint and nowhere else.
	ClientSecret string
	// RedirectURL is where Google sends the parent's browser back to: the
	// service's own callback, as the client at Google has it registered.
	RedirectURL string
	// Endpoints are where the sign-in is: Accounts, unless a test stands in.
	Endpoints Endpoints
	// Now is the clock an ID token is judged by and a grant is dated by.
	Now func() time.Time
}

// ErrSettings is returned when the sign-in is given settings it could not
// work with; callers branch on it with errors.Is.
var ErrSettings = errors.New("googleauth: settings")

// validate refuses a missing part, naming it, rather than building a sign-in
// that fails on the first parent.
func (s *Settings) validate() error {
	switch {
	case s.ClientID == "":
		return fmt.Errorf("%w: ClientID must be set", ErrSettings)
	case s.ClientSecret == "":
		return fmt.Errorf("%w: ClientSecret must be set", ErrSettings)
	case s.RedirectURL == "":
		return fmt.Errorf("%w: RedirectURL must be set", ErrSettings)
	case s.Endpoints.Issuer == "", s.Endpoints.Auth == "", s.Endpoints.Token == "", s.Endpoints.Keys == "":
		return fmt.Errorf("%w: Endpoints must name all four", ErrSettings)
	case s.Now == nil:
		return fmt.Errorf("%w: Now must be set", ErrSettings)
	}
	return nil
}

// SignIn is a parent's sign-in at Google, as the authorization server walks
// through it: where to send the browser, and what the code it comes back with
// is worth.
type SignIn interface {
	// AuthURL is the address the parent's browser is sent to. The state
	// comes back unchanged with the browser; the verifier's challenge and
	// the nonce bind what comes back to this one request.
	AuthURL(state, verifier, nonce string) string
	// Exchange trades the code the browser came back with for the grant,
	// proving the verifier the request was made with, and accepts it only
	// with an ID token that is Google's, for this client, still good, and
	// carrying the nonce.
	Exchange(ctx context.Context, code, verifier, nonce string) (Grant, error)
}

// Grant is what a parent's sign-in at Google gave the service.
type Grant struct {
	// Subject is the Google account's own identifier. It identifies a real
	// person across every service Google signs them in to, so it is turned
	// into an identifier of the service's own and goes no further.
	Subject string
	// AccessToken is what Drive is called with, until Expiry.
	AccessToken string
	// Expiry is when the access token stops being good.
	Expiry time.Time
	// RefreshToken is what a new access token is asked for with.
	RefreshToken string
	// Scopes are the scopes the parent granted, which may be fewer than the
	// ones asked for: Google lets a parent untick a permission.
	Scopes []string
}

// The refusals of Exchange. Every one of them is checked with errors.Is, and
// none of them carries a token or anything Google said in its own words.
var (
	// ErrCodeRefused means Google refused the code: it expired, it was used
	// already, or it was not issued for this client and this verifier.
	ErrCodeRefused = errors.New("googleauth: the code was refused")
	// ErrClient means Google refused the service's own client: its
	// identifier or its secret is not one Google knows. It is a fault of the
	// deployment's, not the parent's, and no sign-in succeeds until it is
	// fixed.
	ErrClient = errors.New("googleauth: Google refused the service's client")
	// ErrUnavailable means Google could not be asked, or did not answer the
	// way its token endpoint does.
	ErrUnavailable = errors.New("googleauth: Google did not answer")
	// ErrIdentity means the answer does not prove who signed in: no ID
	// token, a signature that is not Google's, a token for another client,
	// one expired, or one without the nonce of this request.
	ErrIdentity = errors.New("googleauth: the identity is not proven")
	// ErrNoRefresh means the grant carries no refresh token, without which
	// the sign-in would end with its first access token.
	ErrNoRefresh = errors.New("googleauth: no refresh token")
)

// google is the sign-in at a provider with Google's endpoints.
type google struct {
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
	client   *http.Client
	now      func() time.Time
}

// New builds the sign-in, or refuses settings it could not work with. The
// keys ID tokens are checked against are fetched when the first one arrives,
// kept, and fetched again when one arrives signed with a key not yet seen —
// which is how Google's rotation of its keys reaches the service.
func New(settings *Settings) (SignIn, error) {
	if err := settings.validate(); err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: callTimeout}
	endpoints := settings.Endpoints
	keys := oidc.NewRemoteKeySet(oidc.ClientContext(context.Background(), client), endpoints.Keys)
	return &google{
		oauth: &oauth2.Config{
			ClientID:     settings.ClientID,
			ClientSecret: settings.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:   endpoints.Auth,
				TokenURL:  endpoints.Token,
				AuthStyle: oauth2.AuthStyleInParams,
			},
			RedirectURL: settings.RedirectURL,
			Scopes:      []string{ScopeOpenID, ScopeDriveFile},
		},
		verifier: oidc.NewVerifier(endpoints.Issuer, keys, &oidc.Config{
			ClientID: settings.ClientID,
			// The token is judged a minute behind this clock, so that one
			// Google's clock calls good is not refused for the minute the
			// two may disagree by.
			Now: func() time.Time { return settings.Now().Add(-clockSkew) },
		}),
		client: client,
		now:    settings.Now,
	}, nil
}

// AuthURL asks Google for both scopes, and for a refresh token on every
// sign-in: Google gives one only when the parent is shown its consent
// screen, and the service keeps nothing from a sign-in before.
func (g *google) AuthURL(state, verifier, nonce string) string {
	return g.oauth.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.S256ChallengeOption(verifier),
		oidc.Nonce(nonce),
	)
}
