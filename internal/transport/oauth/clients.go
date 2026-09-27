package oauthserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/cimd"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// How a client is known, as a sign-in's lines name it.
const (
	// registrationCIMD is a client that names itself by the address of its
	// own metadata document.
	registrationCIMD = "cimd"
	// registrationDCR is a client that registered here, and carries its
	// registration sealed in its identifier.
	registrationDCR = "dcr"
)

// Client is a client a parent may be signed in for.
type Client struct {
	// ID is the identifier the client presented.
	ID string
	// Name is what the client calls itself. A client registered here may
	// have given none.
	Name string
	// RedirectURIs are where a sign-in may send the parent back to.
	RedirectURIs []string
	// Registration is how the client is known: registrationCIMD or
	// registrationDCR.
	Registration string
}

// Redirects reports whether a sign-in may send the parent back to an address:
// only when the client registered exactly that address — the same string,
// byte for byte, with no prefix, no normalisation and no pattern.
func (c *Client) Redirects(uri string) bool {
	return slices.Contains(c.RedirectURIs, uri)
}

// errUnknownClient is a client identifier that stands for no client a parent
// may be signed in for. It wraps why.
var errUnknownClient = errors.New("oauth: unknown client")

// clients turns an identifier into the client it stands for, and seals a
// registration into the identifier a client is given.
type clients struct {
	registrations seal.PurposeRing
	issuer        string
	documents     cimd.Fetcher
	events        *signInLog
	now           func() time.Time
}

// resolve is the client an identifier stands for. An HTTPS address is a
// client that names itself by its own document; anything else can only be a
// registration this server sealed, for itself. What neither reads as — a
// string nobody issued, one changed since, a value sealed for another purpose
// or for another issuer, one sealed under a key long retired — is no client.
func (c *clients) resolve(ctx context.Context, clientID string) (*Client, error) {
	if len(clientID) >= len("https:") && strings.EqualFold(clientID[:len("https:")], "https:") {
		return c.fromDocument(ctx, clientID)
	}
	return c.fromRegistration(clientID)
}

// register seals a registration into the identifier the client is given,
// bound to this issuer: the identifier is the registration, and it opens here
// alone.
func (c *clients) register(record registration) (string, error) {
	plain, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("oauth: encode the registration: %w", err)
	}
	clientID, err := c.registrations.Seal(plain, c.issuer)
	if err != nil {
		return "", fmt.Errorf("oauth: seal the registration: %w", err)
	}
	return clientID, nil
}

// fromRegistration opens a registration this server sealed into an
// identifier.
func (c *clients) fromRegistration(clientID string) (*Client, error) {
	plain, err := c.registrations.Open(clientID, c.issuer)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errUnknownClient, err)
	}
	var record registration
	if err := json.Unmarshal(plain, &record); err != nil || len(record.RedirectURIs) == 0 {
		return nil, fmt.Errorf("%w: a registration that names nothing", errUnknownClient)
	}
	return &Client{
		ID:           clientID,
		Name:         record.ClientName,
		RedirectURIs: record.RedirectURIs,
		Registration: registrationDCR,
	}, nil
}

// fromDocument reads a client from its own document. Of the addresses it
// lists, the ones a sign-in may never send the parent to are left out — they
// could never be matched — and a document that leaves none is no client, nor
// is one whose name is too long to show.
func (c *clients) fromDocument(ctx context.Context, clientID string) (*Client, error) {
	started := c.now()
	fetched, err := c.documents.Fetch(ctx, clientID)
	c.events.fetched(ctx, clientID, fetched.Cached, c.now().Sub(started), err)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errUnknownClient, err)
	}

	document := fetched.Document
	redirects := slices.DeleteFunc(slices.Clone(document.RedirectURIs), func(uri string) bool { return !redirectable(uri) })
	switch {
	case len(redirects) == 0:
		return nil, fmt.Errorf("%w: its document lists no address a parent may be sent back to", errUnknownClient)
	case utf8.RuneCountInString(document.ClientName) > maxClientName:
		return nil, fmt.Errorf("%w: its name is longer than %d characters", errUnknownClient, maxClientName)
	}
	return &Client{
		ID:           clientID,
		Name:         document.ClientName,
		RedirectURIs: redirects,
		Registration: registrationCIMD,
	}, nil
}
