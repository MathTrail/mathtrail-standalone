package oauthserver

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

const (
	// maxHostState is the longest state a host may send: it rides in the
	// request sealed into the address Google is sent, and back again.
	maxHostState = 1024
	// verifierBytes and nonceBytes are how many random bytes this server's
	// own verifier and nonce towards Google are made of: a verifier of 43
	// characters, the shortest RFC 7636 allows, and a nonce of 128 bits.
	verifierBytes = 32
	nonceBytes    = 16
)

// authorize answers an authorization request (RFC 6749 4.1.1) with the
// consent screen, or straight on to Google when this browser's parent has
// approved the client before.
//
// The checks run in an order that matters. Until the client is known and the
// address it gave is one it registered, nothing is sent anywhere: a page says
// what is wrong, since an address nobody vouched for is the one place a
// refusal must not lead. From then on the client hears of it at that address.
func (f *flow) authorize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		f.stopAt(w, r, stepAuthorize, stoppedRequest, "invalid_request")
		return
	}
	if how, reason := checkResponse(query); reason != "" {
		f.stopAt(w, r, stepAuthorize, how, reason)
		return
	}
	client, err := f.clients.resolve(r.Context(), query.Get("client_id"))
	if err != nil {
		f.stopAt(w, r, stepAuthorize, stoppedClient, "invalid_client")
		return
	}
	if !client.Redirects(query.Get("redirect_uri")) {
		f.stopAt(w, r, stepAuthorize, stoppedRedirect, "invalid_redirect_uri")
		return
	}

	request, refusal := f.requestOf(query, client)
	if refusal != nil {
		f.events.authorized(r.Context(), request, query, "refused", refusal.reason)
		f.sendBack(w, r, refusal.status, request, refused(refusal.code, refusal.description))
		return
	}
	f.begin(w, r, client, request, query)
}

// checkResponse holds a request to what this server answers: a code, with a
// PKCE challenge by the one method it supports. A request without one is
// refused rather than answered with less. What it is refused with, and why,
// is empty when there is nothing to refuse.
func checkResponse(query url.Values) (how stop, reason string) {
	for _, name := range []string{"response_type", "client_id", "redirect_uri", "code_challenge", "code_challenge_method"} {
		if len(query[name]) > 1 {
			return stoppedRequest, "invalid_request"
		}
	}
	switch {
	case query.Get("response_type") == "":
		return stoppedRequest, "invalid_request"
	case query.Get("response_type") != "code":
		return stoppedResponseType, "unsupported_response_type"
	case query.Get("code_challenge_method") != "S256", !isChallenge(query.Get("code_challenge")):
		return stoppedRequest, "invalid_pkce"
	}
	return stop{}, ""
}

// isChallenge reports whether a value is an S256 challenge: the SHA-256 of a
// verifier, as base64url without padding (RFC 7636 4.2).
func isChallenge(value string) bool {
	digest, err := base64.RawURLEncoding.Strict().DecodeString(value)
	return err == nil && len(digest) == sha256.Size
}

// requestOf is the request a query asks for from a known client at an address
// it registered, or the refusal the client hears when this server cannot give
// what it asks for. The request is returned either way, since a refusal is
// sent to its address with its state.
func (f *flow) requestOf(query url.Values, client *Client) (*flight, *refusal) {
	request := &flight{
		Client:       digestOf(client.ID),
		Registration: client.Registration,
		RedirectURI:  query.Get("redirect_uri"),
		State:        query.Get("state"),
		Challenge:    query.Get("code_challenge"),
		Resource:     f.resource,
		Scope:        f.scope,
		StartedAt:    f.now().Unix(),
	}
	resources := query["resource"]
	switch {
	case len(query["state"]) > 1, len(query["scope"]) > 1:
		return request, &refusal{http.StatusFound, "invalid_request",
			"state and scope are each given once at most", "invalid_request"}
	case len(request.State) > maxHostState:
		return request, &refusal{http.StatusFound, "invalid_request",
			fmt.Sprintf("state is at most %d characters", maxHostState), "state_too_long"}
	case len(resources) > 1, len(resources) == 1 && !oauthex.MatchesResource(resources, f.resource):
		return request, &refusal{http.StatusFound, "invalid_target",
			"the one resource here is " + f.resource, "invalid_target"}
	case scopeWord(query, f.scope) == "other":
		return request, &refusal{http.StatusFound, "invalid_scope",
			"the one scope here is " + f.scope, "invalid_scope"}
	}
	return request, nil
}

// begin starts the sign-in a request asks for. The browser is given the cookie
// the sign-in is tied to, and the parent the consent screen — or, when they
// approved this client before in this browser, the way on to Google.
func (f *flow) begin(w http.ResponseWriter, r *http.Request, client *Client, request *flight, query url.Values) {
	cookie := randomValue(cookieBytes)
	request.Cookie = digestOf(cookie)
	request.Verifier = randomValue(verifierBytes)
	request.Nonce = randomValue(nonceBytes)
	sealed, err := f.sealFlight(request)
	if err != nil {
		f.fail(w, r, stepAuthorize, err)
		return
	}

	setCSRFCookie(w, cookie)
	if f.approved(r, request) {
		if f.google == nil {
			f.stopAt(w, r, stepAuthorize, stoppedUnconfigured, "unconfigured")
			return
		}
		f.events.authorized(r.Context(), request, query, "to_google", "")
		//nolint:gosec // Google's own address, carrying the request this server sealed
		http.Redirect(w, r, f.google.AuthURL(sealed, request.Verifier, request.Nonce), http.StatusFound)
		return
	}

	screen := consentScreen{client: client.Name, redirectURI: request.RedirectURI, request: sealed}
	if f.google != nil {
		screen.google = f.google.AuthURL(sealed, request.Verifier, request.Nonce)
	}
	if err := f.pages.showConsent(w, r, screen); err != nil {
		f.events.failed(r.Context(), stepAuthorize, err)
		return
	}
	f.events.authorized(r.Context(), request, query, "consent", "")
}

// scopeWord is the scope a request asks for, in the closed words of its line:
// none, this server's one scope, or other — which is refused.
func scopeWord(query url.Values, scope string) string {
	asked := strings.Fields(query.Get("scope"))
	for _, each := range asked {
		if each != scope {
			return "other"
		}
	}
	if len(asked) == 0 {
		return "none"
	}
	return scope
}

// resourceWord is whether a request named the resource it is for, in the
// closed words of its line: a request that names none is given this server's,
// and the line is what shows how often that happens.
func resourceWord(query url.Values) string {
	if query.Has("resource") {
		return "given"
	}
	return "absent"
}
