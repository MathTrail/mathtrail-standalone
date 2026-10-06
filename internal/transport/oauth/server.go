// Package oauthserver is the service's own authorization server: a chat host
// signs a parent in against it and gets a token for the MCP endpoint.
//
// It keeps nothing between requests. What a stateful server would keep in a
// database is kept by somebody else. A client that names itself by an HTTPS
// address keeps its own record, in the document at that address. A client
// that registered here carries its record sealed inside the identifier it was
// given. A sign-in under way travels sealed with the parent — in the consent
// screen's form, and in the state Google carries back — and the browser holds
// the cookie that ties it to them and the one that remembers whom they
// approved. A finished sign-in is carried, sealed, by the tokens the host
// holds. This package serves the metadata a host discovers the server by, the
// registration endpoint, the parent's way through a sign-in — the
// authorization request, the consent screen and Google's answer — and the
// host's tokens: their issue and renewal, the end of the grant they carry, and
// the check of the access token every request to the resource comes with.
package oauthserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// Settings are what the authorization server is built from.
type Settings struct {
	// PublicURL is the service's own address and the issuer: a scheme and a
	// host, with no path and no trailing slash.
	PublicURL string
	// Scope is the one scope a token grants, the one the resource asks for.
	Scope string
	// Seal is the key ring what the server issues is sealed under.
	Seal *seal.KeyRing
	// Documents fetches the metadata documents clients name themselves by.
	Documents cimd.Fetcher
	// Logger writes the lines a sign-in leaves.
	Logger *zap.Logger
	// ProjectID is the project the traces a line points to belong to.
	ProjectID string
	// Google is the sign-in a parent is sent on to. Nil where no Google client
	// is configured — a developer's machine — and a sign-in then stops at a
	// page that says so.
	Google googleauth.SignIn
	// Renewals is the pace each account's grant is renewed at Google at,
	// keyed by the account: a refresh token stays good to its own end, and an
	// old one carries a Google token past its end, so every use of it would
	// otherwise be a call to Google.
	Renewals ratelimit.Limiter
	// SiteURL is the site the consent screen links the terms and the privacy
	// policy on: a scheme and a host, with no path and no trailing slash.
	SiteURL string
	// CountryOf is the country a request came from, by its code in the list
	// of countries, or empty when that is not known. It is asked once a
	// sign-in, of the request Google sends the parent's browser back with —
	// the one request of a sign-in that comes from the family rather than from
	// the servers of their chat. Nil where no database of countries is
	// configured, and no country is then known.
	CountryOf func(r *http.Request) string
	// Reviewer is the sign-in of a directory's reviewers, as the demo account.
	// Nil where none is configured, and the consent screen then offers none.
	// It needs Google, which renews the demo account's grant.
	Reviewer *Reviewer
	// Now is the clock everything the server issues is dated by, and a fetch
	// is timed by.
	Now func() time.Time
}

// ErrSettings is returned when the server is given settings it could not work
// with; callers branch on it with errors.Is.
var ErrSettings = errors.New("oauth: settings")

// validate refuses a missing part, naming it, rather than building a server
// that fails on its first request.
func (s *Settings) validate() error {
	switch {
	case s.Scope == "":
		return fmt.Errorf("%w: Scope must be set", ErrSettings)
	case s.Seal == nil:
		return fmt.Errorf("%w: Seal must be set", ErrSettings)
	case s.Documents == nil:
		return fmt.Errorf("%w: Documents must be set", ErrSettings)
	case s.Logger == nil:
		return fmt.Errorf("%w: Logger must be set", ErrSettings)
	case s.Renewals == nil:
		return fmt.Errorf("%w: Renewals must be set", ErrSettings)
	case s.Now == nil:
		return fmt.Errorf("%w: Now must be set", ErrSettings)
	}
	if !isOrigin(s.PublicURL) {
		return fmt.Errorf("%w: PublicURL must be a scheme and a host alone", ErrSettings)
	}
	if !isOrigin(s.SiteURL) {
		return fmt.Errorf("%w: SiteURL must be a scheme and a host alone", ErrSettings)
	}
	return s.validateReviewer()
}

// validateReviewer refuses a reviewer's sign-in that could sign nobody in.
func (s *Settings) validateReviewer() error {
	switch {
	case s.Reviewer == nil:
		return nil
	case s.Reviewer.Password == "", s.Reviewer.Subject == "", s.Reviewer.RefreshToken == "":
		return fmt.Errorf("%w: Reviewer must name a password, a subject and a refresh token", ErrSettings)
	case s.Google == nil:
		return fmt.Errorf("%w: Reviewer needs Google, which renews its grant", ErrSettings)
	}
	return nil
}

// isOrigin reports whether an address is a scheme and a host, and nothing else.
func isOrigin(address string) bool {
	parsed, err := url.Parse(address)
	return err == nil && parsed.Scheme != "" && parsed.Host != "" && parsed.Path == "" &&
		!strings.ContainsAny(address, "?#") && parsed.User == nil
}

// Server is the authorization server as the router and the resource use it:
// the documents a host discovers it by, its endpoints, the reader of the
// access tokens requests to the resource carry, and where a refusal of the
// resource sends a host to begin.
type Server struct {
	// ResourceMetadata serves the protected resource's metadata (RFC 9728),
	// at the address derived from the resource's own path.
	ResourceMetadata http.Handler
	// ServerMetadata serves the authorization server's metadata (RFC 8414).
	ServerMetadata http.Handler
	// Register registers a client (RFC 7591).
	Register http.Handler
	// Authorize answers an authorization request (RFC 6749 4.1.1) with the
	// consent screen, or on to Google.
	Authorize http.Handler
	// Consent answers the consent screen.
	Consent http.Handler
	// Callback answers Google's redirect with the code the client exchanges.
	Callback http.Handler
	// Drive draws the page that asks a parent who left Google's box for the
	// Drive unticked to go back and tick it, at DrivePath.
	Drive http.Handler
	// Token issues a host's tokens, for a code or a refresh token
	// (RFC 6749 4.1.3 and 6).
	Token http.Handler
	// Revoke ends a host's grant (RFC 7009).
	Revoke http.Handler
	// Busy is the page a parent's browser is shown when the sign-in is past
	// its pace, in the parent's language.
	Busy http.Handler
	// Account reads the access token a request to the resource carries: the
	// account it signs the request in as, with the country its sign-in was
	// made from, and when the token ends. A token that signs nobody in is an
	// error.
	Account func(ctx context.Context, token string) (store.Account, time.Time, error)
	// ResourceMetadataURL is the address of the resource's metadata, which a
	// refusal of the resource names.
	ResourceMetadataURL string
	// DemoAccounts are the identifiers of the account the reviewers' sign-in
	// signs in as — the one a sign-in is given now, and during a rotation of
	// the keys the one a sign-in made before it carries — or none where no
	// such sign-in is configured. Its child is no child, and its answers are a
	// reviewer's: whatever counts the children leaves it out.
	DemoAccounts []string

	clients *clients
	flow    *flow
	tokens  *tokens
}

// New builds the authorization server, or refuses settings it could not work
// with.
func New(settings *Settings) (*Server, error) {
	if err := settings.validate(); err != nil {
		return nil, err
	}
	issuer := settings.PublicURL
	events := &signInLog{logger: settings.Logger, projectID: settings.ProjectID}
	known := &clients{
		registrations: settings.Seal.For(seal.PurposeClient),
		issuer:        issuer,
		documents:     settings.Documents,
		events:        events,
		now:           settings.Now,
	}
	serverMetadata, err := serverMetadataHandler(issuer, settings.Scope)
	if err != nil {
		return nil, err
	}
	screens, err := newPages(pageFiles, settings.SiteURL)
	if err != nil {
		return nil, err
	}
	countryOf := settings.CountryOf
	if countryOf == nil {
		countryOf = func(*http.Request) string { return "" }
	}
	reviewer := newReviewerSignIn(settings.Reviewer, settings.Seal.UserIDs)
	signIn := &flow{
		issuer:    issuer,
		resource:  issuer + resourcePath,
		scope:     settings.Scope,
		clients:   known,
		flights:   settings.Seal.For(seal.PurposeState),
		consents:  settings.Seal.For(seal.PurposeConsent),
		codes:     settings.Seal.For(seal.PurposeCode),
		userID:    settings.Seal.UserID,
		countryOf: countryOf,
		google:    settings.Google,
		reviewer:  reviewer,
		renewals:  settings.Renewals,
		pages:     screens,
		events:    events,
		now:       settings.Now,
	}
	grants := &tokens{
		issuer:   issuer,
		resource: issuer + resourcePath,
		scope:    settings.Scope,
		clients:  known,
		codes:    settings.Seal.For(seal.PurposeCode),
		access:   settings.Seal.For(seal.PurposeAccess),
		refresh:  settings.Seal.For(seal.PurposeRefresh),
		google:   settings.Google,
		renewals: settings.Renewals,
		demo:     demoUsers(reviewer),
		events:   events,
		now:      settings.Now,
	}
	return &Server{
		ResourceMetadata:    resourceMetadataHandler(issuer, settings.Scope),
		ServerMetadata:      serverMetadata,
		Register:            &registrar{clients: known, events: events, now: settings.Now},
		Authorize:           http.HandlerFunc(signIn.authorize),
		Consent:             http.HandlerFunc(signIn.consent),
		Callback:            http.HandlerFunc(signIn.callback),
		Drive:               http.HandlerFunc(signIn.drive),
		Token:               http.HandlerFunc(grants.serveToken),
		Revoke:              http.HandlerFunc(grants.serveRevoke),
		Busy:                http.HandlerFunc(screens.busy),
		Account:             grants.account,
		ResourceMetadataURL: issuer + resourceMetadataPath,
		DemoAccounts:        slices.Clone(demoUsers(reviewer)),
		clients:             known,
		flow:                signIn,
		tokens:              grants,
	}, nil
}
