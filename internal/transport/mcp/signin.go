package mcpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"

	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// Scope is the one scope a token for the endpoint grants.
const Scope = "mcp"

// DevAccount is the identifier of the one account the development sign-in
// signs every request in as.
const DevAccount = "dev"

// SignIn decides whose account a request to the endpoint acts for. It hands
// the request on with that account in its context, or answers it with a
// refusal and hands on nothing.
type SignIn func(next http.Handler) http.Handler

// accountKey holds the account a request acts for. It is a type of its own, so
// that nothing else a context carries can be taken for it.
type accountKey struct{}

// withAccount is how a sign-in tells the endpoint whose request it let in.
//
// The context is only the bridge between the HTTP layer, where a request is
// signed in, and the protocol library, which calls the tools: it carries the
// account across and no further. From the frame of a call onwards the account
// travels as an argument — to the tool, and from the tool to the store — so
// that a call that needs somebody's profile says so in its signature.
func withAccount(ctx context.Context, account store.Account) context.Context {
	return context.WithValue(ctx, accountKey{}, account)
}

// accountFrom is the account a sign-in put in ctx, if one did.
func accountFrom(ctx context.Context) (store.Account, bool) {
	account, signedIn := ctx.Value(accountKey{}).(store.Account)
	return account, signedIn
}

// DevSignIn signs every request in as the development account, whether it
// carries a credential or not: a client that cannot send one, such as a chat
// host's custom connector without a sign-in of its own, is let in too. It
// exists so that the tools can be used on a developer's machine before any
// real sign-in does, and the configuration refuses to start a deployed
// service with it.
func DevSignIn(next http.Handler) http.Handler {
	account := store.NewAccount(DevAccount, "")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(withAccount(r.Context(), account)))
	})
}

// TokenReader reads the access token a request carries: the account it signs
// the request in as, and when the token ends. A token it returns an account
// for grants the endpoint's scope — the reader holds it to that, since the
// scope is read from the token. A token that signs nobody in is an error, and
// why stays on this side of the endpoint: the refusal a client hears says
// nothing about the token.
type TokenReader func(ctx context.Context, token string) (store.Account, time.Time, error)

// tokenClockSkew is how far the clock of the instance that issued a token and
// this one may disagree about when it ends.
const tokenClockSkew = 60 * time.Second

// accountExtra is where the information of a token the reader vouched for
// carries the account, from the check of the token to the endpoint.
const accountExtra = "account"

// BearerSignIn lets a request in with an access token the reader vouches for,
// as the account the token signs in. Any other request — one with no token,
// or with a token the reader does not vouch for — is answered the way a
// client expects of a service that wants a token: 401, and a challenge naming
// where the resource's metadata is, which is where a client begins a sign-in,
// and the scope a token needs.
//
// The check is the protocol library's own, and the reader is what it asks.
func BearerSignIn(read TokenReader, resourceMetadataURL string) SignIn {
	verify := func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		account, expires, err := read(ctx, token)
		if err != nil {
			// The library answers with the text of the error it is handed, so
			// it is handed its own, which says nothing about the token.
			return nil, auth.ErrInvalidToken
		}
		return &auth.TokenInfo{
			Scopes:     []string{Scope},
			Expiration: expires,
			UserID:     account.ID,
			Extra:      map[string]any{accountExtra: account},
		}, nil
	}
	require := auth.RequireBearerToken(verify, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL: resourceMetadataURL,
		Scopes:              []string{Scope},
		ClockSkew:           tokenClockSkew,
	})
	return func(next http.Handler) http.Handler {
		return require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The library let the request in with what verify said of its
			// token, the account among it.
			if info := auth.TokenInfoFromContext(r.Context()); info != nil {
				if account, signedIn := info.Extra[accountExtra].(store.Account); signedIn {
					r = r.WithContext(withAccount(r.Context(), account))
				}
			}
			next.ServeHTTP(w, r)
		}))
	}
}
