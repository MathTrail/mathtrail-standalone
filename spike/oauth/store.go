// Throwaway T04 spike: in-memory storage for a stub OAuth authorization server
// that approves everyone without Google. Deleted in T17.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// newToken returns a random URL-safe token; used for client IDs, secrets,
// authorization codes, access tokens and refresh tokens alike.
func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) // spike: no recovery path for a broken CSPRNG
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// dcrClient is a client registered via Dynamic Client Registration (RFC 7591).
type dcrClient struct {
	RedirectURIs            []string
	TokenEndpointAuthMethod string
	ClientSecret            string // empty for "none" (public clients)
	ClientName              string
}

// cimdClient is a client resolved via a Client ID Metadata Document, cached
// briefly so we don't refetch it on every /authorize and /token call.
type cimdClient struct {
	RedirectURIs []string
	ClientName   string
	FetchedAt    time.Time
}

// authCode is a single-use authorization code minted by /authorize.
type authCode struct {
	ClientID            string
	RedirectURI         string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
	Scope               string
	ExpiresAt           time.Time
	Used                bool
}

// accessToken and refreshToken are opaque, in-memory bearer/refresh tokens.
type accessToken struct {
	ClientID  string
	Scope     string
	Resource  string
	ExpiresAt time.Time
}

type refreshToken struct {
	ClientID  string
	Scope     string
	Resource  string
	ExpiresAt time.Time
}

// store is the whole spike's state: every map is protected by mu. A real
// authorization server would persist this; here everything resets when the
// process restarts, which is fine for a spike nobody re-authenticates against
// twice.
type store struct {
	mu            sync.Mutex
	dcrClients    map[string]*dcrClient    // client_id -> record
	cimdCache     map[string]*cimdClient   // client_id (the URL) -> record
	codes         map[string]*authCode     // code -> record
	accessTokens  map[string]*accessToken  // token -> record
	refreshTokens map[string]*refreshToken // token -> record
}

func newStore() *store {
	return &store{
		dcrClients:    make(map[string]*dcrClient),
		cimdCache:     make(map[string]*cimdClient),
		codes:         make(map[string]*authCode),
		accessTokens:  make(map[string]*accessToken),
		refreshTokens: make(map[string]*refreshToken),
	}
}
