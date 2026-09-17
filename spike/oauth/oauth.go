// Throwaway T04 spike: a stub OAuth 2.1 authorization server that approves
// everyone without Google. It plays both the authorization-server and
// resource-server roles in one process, mirroring PRODUCT-V1's architecture
// for the real service (T47-T49), but with no real login and no encryption.
// Deleted in T17.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

const (
	codeTTL         = 60 * time.Second
	accessTokenTTL  = 90 * time.Second // short on purpose: a live run can watch a refresh happen
	refreshTokenTTL = 24 * time.Hour
	cimdCacheTTL    = 5 * time.Minute
	cimdMaxBody     = 64 << 10 // 64KB: a client metadata document has no business being bigger
)

type oauthServer struct {
	store  *store
	logger *slog.Logger
}

// issuer returns this request's own origin as the OAuth issuer. There is no
// way to know the public tunnel URL at process start (it's assigned by
// cloudflared after we're already listening), so every handler derives it
// fresh from the incoming Host header instead of a startup flag.
func issuer(r *http.Request) string {
	return "https://" + r.Host
}

// -- RFC 9728: OAuth 2.0 Protected Resource Metadata -----------------------

func (s *oauthServer) protectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	iss := issuer(r)
	meta := &oauthex.ProtectedResourceMetadata{
		Resource:               iss + "/mcp",
		AuthorizationServers:   []string{iss},
		ScopesSupported:        []string{"mcp"},
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "MathTrail OAuth spike (T04)",
	}
	auth.ProtectedResourceMetadataHandler(meta).ServeHTTP(w, r)
}

// -- RFC 8414: OAuth 2.0 Authorization Server Metadata ----------------------

func (s *oauthServer) authServerMetadata(w http.ResponseWriter, r *http.Request) {
	iss := issuer(r)
	meta := &oauthex.AuthServerMeta{
		Issuer:                                     iss,
		AuthorizationEndpoint:                      iss + "/authorize",
		TokenEndpoint:                              iss + "/token",
		RegistrationEndpoint:                       iss + "/register",
		ScopesSupported:                            []string{"mcp"},
		ResponseTypesSupported:                     []string{"code"},
		GrantTypesSupported:                        []string{"authorization_code", "refresh_token"},
		TokenEndpointAuthMethodsSupported:          []string{"none", "client_secret_post"},
		CodeChallengeMethodsSupported:              []string{"S256"},
		ClientIDMetadataDocumentSupported:          true,
		AuthorizationResponseIssParameterSupported: true,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(meta)
}

// -- RFC 7591: Dynamic Client Registration ----------------------------------

func (s *oauthServer) register(w http.ResponseWriter, r *http.Request) {
	var req oauthex.ClientRegistrationMetadata
	if err := json.NewDecoder(io.LimitReader(r.Body, cimdMaxBody)).Decode(&req); err != nil {
		s.writeRegistrationError(w, http.StatusBadRequest, "invalid_client_metadata", "body is not valid JSON")
		return
	}
	if len(req.RedirectURIs) == 0 {
		s.writeRegistrationError(w, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris is required")
		return
	}
	authMethod := req.TokenEndpointAuthMethod
	if authMethod == "" {
		// RFC 7591's own default is client_secret_basic, but MCP clients doing
		// PKCE are public clients; default to "none" to match that world.
		authMethod = "none"
	}

	clientID := "spike-" + newToken()
	rec := &dcrClient{
		RedirectURIs:            req.RedirectURIs,
		TokenEndpointAuthMethod: authMethod,
		ClientName:              req.ClientName,
	}
	if authMethod != "none" {
		rec.ClientSecret = newToken()
	}

	s.store.mu.Lock()
	s.store.dcrClients[clientID] = rec
	s.store.mu.Unlock()

	s.logger.Info("register",
		"registration", "dcr",
		"client_id", clientID,
		"client_name", req.ClientName,
		"redirect_uris", req.RedirectURIs,
		"token_endpoint_auth_method", authMethod,
		"application_type", req.ApplicationType,
	)

	resp := oauthex.ClientRegistrationResponse{
		ClientRegistrationMetadata: req,
		ClientID:                   clientID,
		ClientSecret:               rec.ClientSecret,
		ClientIDIssuedAt:           time.Now(),
	}
	resp.ClientRegistrationMetadata.TokenEndpointAuthMethod = authMethod
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&resp)
}

func (s *oauthServer) writeRegistrationError(w http.ResponseWriter, status int, code, desc string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(&oauthex.ClientRegistrationError{ErrorCode: code, ErrorDescription: desc})
}

// -- Client ID Metadata Documents -------------------------------------------

// looksLikeCIMD reports whether a client_id is an HTTPS URL with a path, per
// draft-ietf-oauth-client-id-metadata-document-00 §3 ("MUST use the https
// scheme and contain a path component").
func looksLikeCIMD(clientID string) bool {
	u, err := url.Parse(clientID)
	if err != nil || u.Scheme != "https" || u.Path == "" || u.Path == "/" {
		return false
	}
	return true
}

type cimdDocument struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
}

// fetchCIMD retrieves and validates a client's metadata document. Minimal
// SSRF hardening for a spike: HTTPS only (already checked by looksLikeCIMD),
// a short timeout and a capped response body. Production would also need
// private-IP / redirect-chasing protections per the spec's own security
// considerations section (client-id-metadata-document-00 §6); not done here.
func (s *oauthServer) fetchCIMD(clientID string) (*cimdClient, error) {
	s.store.mu.Lock()
	if cached, ok := s.store.cimdCache[clientID]; ok && time.Since(cached.FetchedAt) < cimdCacheTTL {
		s.store.mu.Unlock()
		return cached, nil
	}
	s.store.mu.Unlock()

	// 10s and one retry: a live run against claude.ai's real CIMD document hit
	// "context deadline exceeded" on a 5s timeout with no retry, even though a
	// plain curl to the same URL from the same devcontainer completed in
	// ~0.2s moments later — most likely a slow first DNS resolution inside
	// this process rather than the endpoint actually being slow.
	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		resp, err = client.Get(clientID) //nolint:noctx // spike; opts.Timeout already bounds this
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("fetch client-id-metadata-document: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client-id-metadata-document: unexpected status %d", resp.StatusCode)
	}
	var doc cimdDocument
	if err := json.NewDecoder(io.LimitReader(resp.Body, cimdMaxBody)).Decode(&doc); err != nil {
		return nil, fmt.Errorf("client-id-metadata-document: invalid JSON: %w", err)
	}
	if doc.ClientID != clientID {
		return nil, fmt.Errorf("client-id-metadata-document: client_id %q does not match document URL %q", doc.ClientID, clientID)
	}
	if len(doc.RedirectURIs) == 0 {
		return nil, fmt.Errorf("client-id-metadata-document: redirect_uris is empty")
	}

	rec := &cimdClient{RedirectURIs: doc.RedirectURIs, ClientName: doc.ClientName, FetchedAt: time.Now()}
	s.store.mu.Lock()
	s.store.cimdCache[clientID] = rec
	s.store.mu.Unlock()
	return rec, nil
}

// resolveClient finds a client's allowed redirect URIs, trying CIMD (if
// client_id looks like a URL) and then DCR-registered clients, per the
// priority order in client-registration.mdx (pre-registration is N/A here —
// this spike approves everyone, there's nothing to pre-register).
func (s *oauthServer) resolveClient(clientID string) (redirectURIs []string, name string, registration string, err error) {
	if looksLikeCIMD(clientID) {
		rec, err := s.fetchCIMD(clientID)
		if err != nil {
			return nil, "", "", err
		}
		return rec.RedirectURIs, rec.ClientName, "cimd", nil
	}
	s.store.mu.Lock()
	rec, ok := s.store.dcrClients[clientID]
	s.store.mu.Unlock()
	if !ok {
		return nil, "", "", fmt.Errorf("unknown client_id %q (not a CIMD URL, not registered via DCR)", clientID)
	}
	return rec.RedirectURIs, rec.ClientName, "dcr", nil
}

// -- /authorize --------------------------------------------------------------

func (s *oauthServer) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")
	resource := q.Get("resource")
	scope := q.Get("scope")

	if q.Get("response_type") != "code" {
		http.Error(w, "unsupported response_type", http.StatusBadRequest)
		return
	}
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		// Security Considerations: MCP clients MUST use S256; a server that
		// accepted "plain" or no PKCE at all would be non-compliant.
		http.Error(w, "code_challenge with method S256 is required", http.StatusBadRequest)
		return
	}

	redirectURIs, _, registration, err := s.resolveClient(clientID)
	if err != nil {
		s.logger.Info("authorize_reject", "reason", err.Error(), "client_id", clientID)
		http.Error(w, "invalid_client: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !slicesContain(redirectURIs, redirectURI) {
		s.logger.Info("authorize_reject", "reason", "redirect_uri not registered", "client_id", clientID, "redirect_uri", redirectURI)
		http.Error(w, "invalid_request: redirect_uri does not match a registered value", http.StatusBadRequest)
		return
	}

	// Always grant exactly "mcp", regardless of what was requested: this makes
	// the mcp:advanced tool's insufficient-scope / step-up demo (mcp.go)
	// deterministic instead of depending on what a live host happens to ask for.
	const grantedScope = "mcp"

	code := newToken()
	s.store.mu.Lock()
	s.store.codes[code] = &authCode{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Resource:            resource,
		Scope:               grantedScope,
		ExpiresAt:           time.Now().Add(codeTTL),
	}
	s.store.mu.Unlock()

	s.logger.Info("authorize",
		"registration", registration,
		"client_id", clientID,
		"redirect_uri", redirectURI,
		"resource", resource,
		"requested_scope", scope,
		"granted_scope", grantedScope,
		"code_challenge_method", codeChallengeMethod,
	)

	// No login screen at all: this spike approves everyone (T04's own brief).
	// A real consent screen for confused-deputy protection is T07's job.
	dest, _ := url.Parse(redirectURI)
	dq := dest.Query()
	dq.Set("code", code)
	if state != "" {
		dq.Set("state", state)
	}
	dq.Set("iss", issuer(r))
	dest.RawQuery = dq.Encode()
	http.Redirect(w, r, dest.String(), http.StatusFound)
}

// -- /token -------------------------------------------------------------------

func (s *oauthServer) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.writeTokenError(w, http.StatusBadRequest, "invalid_request", "malformed form body")
		return
	}
	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		s.tokenFromCode(w, r)
	case "refresh_token":
		s.tokenFromRefresh(w, r)
	default:
		s.writeTokenError(w, http.StatusBadRequest, "unsupported_grant_type", "")
	}
}

func (s *oauthServer) tokenFromCode(w http.ResponseWriter, r *http.Request) {
	code := r.PostForm.Get("code")
	verifier := r.PostForm.Get("code_verifier")
	redirectURI := r.PostForm.Get("redirect_uri")
	clientID := r.PostForm.Get("client_id")
	resource := r.PostForm.Get("resource")

	s.store.mu.Lock()
	rec, ok := s.store.codes[code]
	alreadyUsed := ok && rec.Used
	if ok {
		rec.Used = true // single-use: the next redemption attempt sees alreadyUsed
	}
	s.store.mu.Unlock()

	if !ok {
		s.logger.Info("token_reject", "grant_type", "authorization_code", "reason", "unknown code")
		s.writeTokenError(w, http.StatusBadRequest, "invalid_grant", "unknown code")
		return
	}
	if alreadyUsed {
		// RFC 6749 §10.5: a reused code must revoke tokens already issued for
		// it. This spike just rejects the replay; it doesn't walk back and
		// revoke — there's no live traffic that would exercise that path.
		s.logger.Info("token_reject", "grant_type", "authorization_code", "reason", "code already used", "client_id", clientID)
		s.writeTokenError(w, http.StatusBadRequest, "invalid_grant", "code already used")
		return
	}
	if time.Now().After(rec.ExpiresAt) {
		s.logger.Info("token_reject", "grant_type", "authorization_code", "reason", "code expired", "client_id", clientID)
		s.writeTokenError(w, http.StatusBadRequest, "invalid_grant", "code expired")
		return
	}
	if rec.ClientID != clientID || rec.RedirectURI != redirectURI {
		s.logger.Info("token_reject", "grant_type", "authorization_code", "reason", "client_id/redirect_uri mismatch", "client_id", clientID)
		s.writeTokenError(w, http.StatusBadRequest, "invalid_grant", "client_id or redirect_uri mismatch")
		return
	}
	if rec.Resource != "" && resource != "" && rec.Resource != resource {
		s.logger.Info("token_reject", "grant_type", "authorization_code", "reason", "resource mismatch", "authorize_resource", rec.Resource, "token_resource", resource)
		s.writeTokenError(w, http.StatusBadRequest, "invalid_target", "resource does not match the value from /authorize")
		return
	}
	if !verifyPKCE(verifier, rec.CodeChallenge) {
		s.logger.Info("token_reject", "grant_type", "authorization_code", "reason", "PKCE verification failed", "client_id", clientID)
		s.writeTokenError(w, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
		return
	}

	access := newToken()
	refresh := newToken()
	s.store.mu.Lock()
	s.store.accessTokens[access] = &accessToken{ClientID: clientID, Scope: rec.Scope, Resource: rec.Resource, ExpiresAt: time.Now().Add(accessTokenTTL)}
	s.store.refreshTokens[refresh] = &refreshToken{ClientID: clientID, Scope: rec.Scope, Resource: rec.Resource, ExpiresAt: time.Now().Add(refreshTokenTTL)}
	s.store.mu.Unlock()

	s.logger.Info("token_issue", "grant_type", "authorization_code", "client_id", clientID, "resource", rec.Resource, "scope", rec.Scope, "expires_in", int(accessTokenTTL.Seconds()))
	s.writeTokenResponse(w, access, refresh, rec.Scope)
}

func (s *oauthServer) tokenFromRefresh(w http.ResponseWriter, r *http.Request) {
	oldRefresh := r.PostForm.Get("refresh_token")
	clientID := r.PostForm.Get("client_id")

	s.store.mu.Lock()
	rec, ok := s.store.refreshTokens[oldRefresh]
	if ok {
		delete(s.store.refreshTokens, oldRefresh) // rotate: old refresh token is single-use
	}
	s.store.mu.Unlock()

	if !ok || time.Now().After(rec.ExpiresAt) {
		s.logger.Info("token_reject", "grant_type", "refresh_token", "reason", "unknown or expired refresh token", "client_id", clientID)
		s.writeTokenError(w, http.StatusBadRequest, "invalid_grant", "unknown or expired refresh token")
		return
	}
	if clientID != "" && rec.ClientID != clientID {
		s.logger.Info("token_reject", "grant_type", "refresh_token", "reason", "client_id mismatch")
		s.writeTokenError(w, http.StatusBadRequest, "invalid_grant", "client_id mismatch")
		return
	}

	access := newToken()
	newRefresh := newToken()
	s.store.mu.Lock()
	s.store.accessTokens[access] = &accessToken{ClientID: rec.ClientID, Scope: rec.Scope, Resource: rec.Resource, ExpiresAt: time.Now().Add(accessTokenTTL)}
	s.store.refreshTokens[newRefresh] = &refreshToken{ClientID: rec.ClientID, Scope: rec.Scope, Resource: rec.Resource, ExpiresAt: time.Now().Add(refreshTokenTTL)}
	s.store.mu.Unlock()

	s.logger.Info("token_refresh", "client_id", rec.ClientID, "resource", rec.Resource, "expires_in", int(accessTokenTTL.Seconds()))
	s.writeTokenResponse(w, access, newRefresh, rec.Scope)
}

func (s *oauthServer) writeTokenResponse(w http.ResponseWriter, access, refresh, scope string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token":  access,
		"token_type":    "Bearer",
		"expires_in":    int(accessTokenTTL.Seconds()),
		"refresh_token": refresh,
		"scope":         scope,
	})
}

func (s *oauthServer) writeTokenError(w http.ResponseWriter, status int, code, desc string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	body := map[string]string{"error": code}
	if desc != "" {
		body["error_description"] = desc
	}
	_ = json.NewEncoder(w).Encode(body)
}

func verifyPKCE(verifier, challenge string) bool {
	sum := sha256.Sum256([]byte(verifier))
	got := base64.RawURLEncoding.EncodeToString(sum[:])
	return got == challenge
}

func slicesContain(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// -- Bearer token verification for /mcp --------------------------------------

// verifyToken implements auth.TokenVerifier against the in-memory access
// token store.
func (s *oauthServer) verifyToken(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
	s.store.mu.Lock()
	rec, ok := s.store.accessTokens[token]
	s.store.mu.Unlock()
	if !ok {
		return nil, auth.ErrInvalidToken
	}
	if time.Now().After(rec.ExpiresAt) {
		s.logger.Info("token_expired", "client_id", rec.ClientID)
		return nil, auth.ErrInvalidToken
	}
	return &auth.TokenInfo{
		Scopes:     strings.Fields(rec.Scope),
		Expiration: rec.ExpiresAt,
		UserID:     rec.ClientID,
	}, nil
}
