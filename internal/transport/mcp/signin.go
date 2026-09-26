package mcpserver

import (
	"context"
	"net/http"

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

// NobodySignsIn lets no request in. It answers every one the way a request
// without a valid token is answered — 401, and a challenge naming the scope a
// token would need — so that the endpoint refuses in the shape a client
// expects of it, while no token can yet be issued at all.
func NobodySignsIn(next http.Handler) http.Handler {
	return auth.RequireBearerToken(refuseEveryToken, &auth.RequireBearerTokenOptions{
		Scopes: []string{Scope},
	})(next)
}

// refuseEveryToken is the verdict on any token while nothing can issue one.
// The error the library is handed ends up in the body of the refusal, so it is
// the library's own, which says nothing about the token.
func refuseEveryToken(context.Context, string, *http.Request) (*auth.TokenInfo, error) {
	return nil, auth.ErrInvalidToken
}
