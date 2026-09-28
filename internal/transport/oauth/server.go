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
// approved. This package serves the metadata a host discovers the server by,
// the registration endpoint, and the parent's way through a sign-in: the
// authorization request, the consent screen and Google's answer.
package oauthserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
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
	// SiteURL is the site the consent screen links the terms and the privacy
	// policy on: a scheme and a host, with no path and no trailing slash.
	SiteURL string
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
	case s.Now == nil:
		return fmt.Errorf("%w: Now must be set", ErrSettings)
	}
	if !isOrigin(s.PublicURL) {
		return fmt.Errorf("%w: PublicURL must be a scheme and a host alone", ErrSettings)
	}
	if !isOrigin(s.SiteURL) {
		return fmt.Errorf("%w: SiteURL must be a scheme and a host alone", ErrSettings)
	}
	return nil
}

// isOrigin reports whether an address is a scheme and a host, and nothing else.
func isOrigin(address string) bool {
	parsed, err := url.Parse(address)
	return err == nil && parsed.Scheme != "" && parsed.Host != "" && parsed.Path == "" &&
		!strings.ContainsAny(address, "?#") && parsed.User == nil
}

// Server is the authorization server as the router mounts it: the documents a
// host discovers it by, its endpoints, and where a refusal of the resource
// sends a host to begin.
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
	// ResourceMetadataURL is the address of the resource's metadata, which a
	// refusal of the resource names.
	ResourceMetadataURL string

	clients *clients
	flow    *flow
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
	screens, err := newPages(settings.SiteURL)
	if err != nil {
		return nil, err
	}
	signIn := &flow{
		issuer:   issuer,
		resource: issuer + resourcePath,
		scope:    settings.Scope,
		clients:  known,
		flights:  settings.Seal.For(seal.PurposeState),
		consents: settings.Seal.For(seal.PurposeConsent),
		codes:    settings.Seal.For(seal.PurposeCode),
		userID:   settings.Seal.UserID,
		google:   settings.Google,
		pages:    screens,
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
		ResourceMetadataURL: issuer + resourceMetadataPath,
		clients:             known,
		flow:                signIn,
	}, nil
}
