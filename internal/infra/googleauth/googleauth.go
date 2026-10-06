// Package googleauth signs a parent in with Google.
//
// Towards Google the service is an ordinary OAuth client. It sends the
// parent's browser to Google's consent screen with the one scope it needs —
// the files it creates in their Drive, and no other — and trades the code
// Google sends back for the grant. The scope is asked for alone: once a
// request asks for more than one scope, a sign-in's counted, Google shows
// every scope that is not a sign-in's with a box it leaves unticked, and a
// parent who presses on past the box has not granted that scope.
//
// Who signed in is asked of Drive: its about resource names the account an
// access token reaches Drive as, by the identifier Drive's permissions know it
// by. That is the one identifier of an account a grant of the Drive alone can
// learn. Google's own identifier for the account comes with a sign-in's scope,
// which would bring the box back, and Google's tokeninfo endpoint leaves it out
// of what it says of a token granted the Drive alone. The token asked with is
// the one the exchange of this request's code has just given, over a
// connection the service opened itself to Google, so it is this client's own,
// and Drive's word comes over another such connection.
//
// After the sign-in the grant is renewed with its refresh token, for as long
// as Google honours it, and revoked when the parent disconnects.
package googleauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
)

// ScopeDriveFile is the one scope asked of Google: the files the service
// creates in the parent's Drive — no other file.
const ScopeDriveFile = "https://www.googleapis.com/auth/drive.file"

// callTimeout is how long one call to Google may take: an exchange of a code
// with the question of whose Drive its access token reaches, a renewal, or a
// revocation.
const callTimeout = 10 * time.Second

// Endpoints are where a provider's sign-in is.
type Endpoints struct {
	// Auth is where the parent's browser is sent to sign in.
	Auth string
	// Token is where a code is exchanged for the grant.
	Token string
	// About is where Drive is asked whose Drive an access token reaches.
	About string
	// Revoke is where a grant is ended.
	Revoke string
}

// Accounts are Google's own endpoints: those its discovery document at
// https://accounts.google.com/.well-known/openid-configuration names, and
// Drive's about resource, as version 3 of Drive's API has it.
var Accounts = Endpoints{ //nolint:gosec // Google's published endpoints, not a credential
	Auth:   "https://accounts.google.com/o/oauth2/v2/auth",
	Token:  "https://oauth2.googleapis.com/token",
	About:  drive.Google + "/drive/v3/about",
	Revoke: "https://oauth2.googleapis.com/revoke",
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
	// Now is the clock a grant is dated by.
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
	case s.Endpoints.Auth == "", s.Endpoints.Token == "", s.Endpoints.About == "", s.Endpoints.Revoke == "":
		return fmt.Errorf("%w: Endpoints must name all four", ErrSettings)
	case s.Now == nil:
		return fmt.Errorf("%w: Now must be set", ErrSettings)
	}
	return nil
}

// SignIn is a parent's sign-in at Google, as the authorization server walks
// through it: where to send the browser, what the code it comes back with is
// worth, and — for as long as the grant lasts — a new access token, and the
// grant's end.
type SignIn interface {
	// AuthURL is the address the parent's browser is sent to. The state
	// comes back unchanged with the browser; the verifier's challenge binds
	// the code it comes back with to this one request.
	AuthURL(state, verifier string) string
	// Exchange trades the code the browser came back with for the grant,
	// proving the verifier the request was made with, and accepts it only
	// once it reaches the parent's Drive and Drive has said whose it is.
	Exchange(ctx context.Context, code, verifier string) (Grant, error)
	// Refresh asks Google for a new access token with the grant's refresh
	// token.
	Refresh(ctx context.Context, refreshToken string) (Renewal, error)
	// Revoke ends a grant at Google, given any token of it: Google ends the
	// whole grant, the refresh token and every access token alike. A token
	// Google no longer honours ends nothing, and says so with ErrNotHonoured.
	Revoke(ctx context.Context, token string) error
}

// Grant is what a parent's sign-in at Google gave the service.
type Grant struct {
	// PermissionID is the account's identifier at Drive, as Drive's
	// permissions show it. It identifies a real person to anybody who shares
	// a file with them, so it is turned into an identifier of the service's
	// own and goes no further.
	PermissionID string
	// AccessToken is what Drive is called with, until Expiry.
	AccessToken string
	// Expiry is when the access token stops being good.
	Expiry time.Time
	// RefreshToken is what a new access token is asked for with.
	RefreshToken string
}

// Renewal is what Google gave for a grant's refresh token.
type Renewal struct {
	// AccessToken is what Drive is called with, until Expiry.
	AccessToken string
	// Expiry is when the access token stops being good.
	Expiry time.Time
	// RefreshToken is the refresh token the grant goes on with: the one it
	// was renewed with, unless Google gave another in its place.
	RefreshToken string
}

// The refusals of the sign-in's calls to Google. Every one of them is checked
// with errors.Is, and none of them carries a token or anything Google said in
// its own words.
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
	// way its endpoints do.
	ErrUnavailable = errors.New("googleauth: Google did not answer")
	// ErrIdentity means Drive did not say who signed in: it called the
	// access token none of Google's, or named no account.
	ErrIdentity = errors.New("googleauth: the identity is not proven")
	// ErrNoRefresh means the grant carries no refresh token, without which
	// the sign-in would end with its first access token.
	ErrNoRefresh = errors.New("googleauth: no refresh token")
	// ErrNoDrive means the grant does not reach the parent's Drive: Google
	// granted less than the one scope asked for, as it does when a parent
	// leaves a box on its screen unticked. Drive, the one that could say who
	// signed in, is not asked.
	ErrNoDrive = errors.New("googleauth: the Drive was not granted")
	// ErrGrantEnded means Google no longer honours the grant's refresh token:
	// the parent took the access back, the grant went unused for months, or
	// it fell past the number of refresh tokens Google keeps for a client.
	// Only a new sign-in gives another.
	ErrGrantEnded = errors.New("googleauth: the grant has ended")
	// ErrNotHonoured means Google no longer honours the token a grant was to
	// be ended with: the grant has ended already, or the token has — an access
	// token past its hour, a refresh token Google gave another in place of.
	// Nothing was ended with it.
	ErrNotHonoured = errors.New("googleauth: Google no longer honours the token")
)

// google is the sign-in at a provider with Google's endpoints.
type google struct {
	oauth     *oauth2.Config
	client    *http.Client
	aboutURL  string
	revokeURL string
	now       func() time.Time
}

// New builds the sign-in, or refuses settings it could not work with.
func New(settings *Settings) (SignIn, error) {
	if err := settings.validate(); err != nil {
		return nil, err
	}
	endpoints := settings.Endpoints
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
			Scopes:      []string{ScopeDriveFile},
		},
		client:    &http.Client{Timeout: callTimeout},
		aboutURL:  endpoints.About,
		revokeURL: endpoints.Revoke,
		now:       settings.Now,
	}, nil
}

// AuthURL asks Google for the one scope, and for a refresh token on every
// sign-in: Google gives one only when the parent is shown its consent
// screen, and the service keeps nothing from a sign-in before.
func (g *google) AuthURL(state, verifier string) string {
	return g.oauth.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.S256ChallengeOption(verifier),
	)
}
