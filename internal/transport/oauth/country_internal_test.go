package oauthserver

import (
	"net/http"
	"slices"
	"sync"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// The country of a sign-in is asked of the request the parent's browser comes
// back from Google with, and of no other: the requests before it are the
// browser's too but end nowhere, and those after it come from the servers of
// the chat. The code carries it to the host's tokens, the access token to the
// account it signs in, and a renewal keeps it as the sign-in made it.
func TestTheCountryOfASignInTravelsWithItsTokens(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	var (
		mu      sync.Mutex
		askedOf []string
	)
	h.server.flow.countryOf = func(r *http.Request) string {
		mu.Lock()
		defer mu.Unlock()
		askedOf = append(askedOf, r.URL.Path)
		return "NZ"
	}
	client := h.register(hostRedirect, hostName)

	code := h.signedInCode(t, client)
	if got := h.opened(code).Country; got != "NZ" {
		t.Errorf("the code carries the country %q, want NZ", got)
	}
	answer := h.exchange(t, code, client, nil)
	if answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/token = %d %v, want the host's tokens", answer.status, answer.fields)
	}
	access, refresh := answer.field("access_token"), answer.field("refresh_token")
	if got, gotRefresh := h.openedAccess(t, access).Country, h.openedRefresh(t, refresh).Country; got != "NZ" || gotRefresh != "NZ" {
		t.Errorf("the access token carries the country %q and the refresh token %q, want NZ in both", got, gotRefresh)
	}
	account, _, err := h.server.Account(t.Context(), access)
	if err != nil {
		t.Fatalf("Account() error = %v, want the account", err)
	}
	if account.SignInCountry != "NZ" {
		t.Errorf("Account() signs in an account from %q, want NZ", account.SignInCountry)
	}

	renewed := h.renew(t, refresh, client, nil)
	if renewed.status != http.StatusOK {
		t.Fatalf("POST /oauth/token = %d %v, want renewed tokens", renewed.status, renewed.fields)
	}
	if got, gotRefresh := h.openedAccess(t, renewed.field("access_token")).Country,
		h.openedRefresh(t, renewed.field("refresh_token")).Country; got != "NZ" || gotRefresh != "NZ" {
		t.Errorf("the renewed tokens carry the countries %q and %q, want NZ in both", got, gotRefresh)
	}

	mu.Lock()
	defer mu.Unlock()
	if !slices.Equal(askedOf, []string{CallbackPath}) {
		t.Errorf("the country was asked of %v, want the callback alone, once", askedOf)
	}
}

// A sign-in whose country is not known carries none, and a token of a sign-in
// made before the country was kept — one with no such field at all — signs
// its account in all the same, from no country.
func TestATokenWithNoCountrySignsInFromNone(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	access, _ := h.tokensFor(t, client)
	if got := h.openedAccess(t, access).Country; got != "" {
		t.Errorf("a sign-in of no known country carries %q, want none", got)
	}

	old := sealedFor(t, h.ring, seal.PurposeAccess, h.served.URL, h.anAccessGrant())
	account, _, err := h.server.Account(t.Context(), old)
	if err != nil {
		t.Fatalf("Account() error = %v, want the account", err)
	}
	if account.SignInCountry != "" {
		t.Errorf("Account() signs in an account from %q, want no country", account.SignInCountry)
	}
}
