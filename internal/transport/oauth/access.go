package oauthserver

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/oauthex"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// errNoAccount is an access token that signs nobody in. It wraps why.
var errNoAccount = errors.New("oauth: the token signs nobody in")

// account is the account an access token signs a request to the resource in
// as, and when the token ends. The token is one this server sealed, for
// itself, as an access token; it has not ended, a minute of disagreement
// between clocks allowed; its audience is this resource, compared as the
// protocol compares a resource; and it grants this server's scope. The account
// is reached with the Google access token the token carries, until the moment
// that token ends, so that nothing is asked of Drive that would outlive it, and
// it carries the country its sign-in was made from.
//
// Anything else signs nobody in, and its line says why — a key retired since,
// or no token of this server's at all — while the refusal the client hears
// does not.
func (t *tokens) account(ctx context.Context, token string) (store.Account, time.Time, error) {
	var grant accessGrant
	reason := ""
	switch err := t.openAs(t.access, token, &grant); {
	case errors.Is(err, seal.ErrUnknownKey):
		reason = "unknown_key"
	case err != nil:
		reason = "unreadable"
	case t.now().After(time.Unix(grant.ExpiresAt, 0).Add(clockSkew)):
		reason = "expired"
	case !oauthex.MatchesResource([]string{grant.Resource}, t.resource):
		reason = "audience"
	case !slices.Contains(strings.Fields(grant.Scope), t.scope):
		reason = "scope"
	}
	if reason != "" {
		t.events.bearerRefused(ctx, reason)
		return store.Account{}, time.Time{}, fmt.Errorf("%w: %s", errNoAccount, reason)
	}
	account := store.NewAccount(grant.User, grant.GoogleAccessToken, time.Unix(grant.GoogleAccessExpiry, 0))
	account.SignInCountry = grant.Country
	return account, time.Unix(grant.ExpiresAt, 0), nil
}
