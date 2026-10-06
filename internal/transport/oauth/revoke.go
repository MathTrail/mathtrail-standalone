package oauthserver

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// The kinds of token a host holds, as the line of a revocation names them.
const (
	kindAccess  = "access"
	kindRefresh = "refresh"
)

// held is what a token a host holds says of the grant it belongs to: which
// kind of token it is, the digest of the client it was issued to, whom it
// signs in, the token of Google's that ends the grant, and when that token
// stops being honoured — never, for a refresh token, as far as this server
// can tell.
type held struct {
	kind        string
	client      string
	user        string
	googleToken string
	googleEnds  time.Time
}

// serveRevoke ends a host's grant (RFC 7009). A server that keeps nothing
// cannot take back a token it issued, so revoking only its own would change
// nothing: the grant is ended at Google, the one place it can end, and every
// token of it ends there with it — the tokens of any other chat the parent
// connected included. The demo account's grant is the one exception, since
// the reviewers' sign-in goes on renewing it: a token of it ends nothing, and
// expires as a token does. A token none of this server's is answered as
// revoked, as the protocol asks, since there is nothing a client could do
// about it; which kind of token the client says it sent is only a hint, and
// needs no reading, since either kind is tried.
func (t *tokens) serveRevoke(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	form, refused := readForm(w, r)
	registration := t.clients.kindOf(form.Get("client_id"))
	if refused == nil && (form.Get("token") == "" || form.Get("client_id") == "") {
		refused = &refusal{http.StatusBadRequest, "invalid_request", "token and client_id are required", "invalid_request"}
	}
	if refused != nil {
		t.events.revoked(r.Context(), registration, "", &ending{outcome: "refused", reason: refused.reason})
		writeRefusal(w, t.issuer, refused)
		return
	}

	token, found := t.heldGrant(form.Get("token"))
	switch {
	case !found:
		t.events.revoked(r.Context(), registration, "", &ending{outcome: "ignored", reason: "unknown_token"})
		w.WriteHeader(http.StatusOK)
		return
	case !sameDigest(digestOf(form.Get("client_id")), token.client):
		refused = invalidGrant("the token was issued to another client", "client")
		t.events.revoked(r.Context(), registration, token.kind, &ending{outcome: "refused", reason: refused.reason, user: token.user})
		writeRefusal(w, t.issuer, refused)
		return
	case slices.Contains(t.demo, token.user):
		// The demo account's grant is the one the reviewers' sign-in renews.
		// Ended at Google, it would end that sign-in for every reviewer until
		// it is captured again — and a reviewer disconnects once the review is
		// over.
		t.events.revoked(r.Context(), registration, token.kind, &ending{outcome: "ignored", reason: "demo_account", user: token.user})
		w.WriteHeader(http.StatusOK)
		return
	case !token.googleEnds.IsZero() && !token.googleEnds.After(t.now()):
		// Google no longer honours the token inside, which can end nothing:
		// the host's refresh token is what ends this grant.
		t.events.revoked(r.Context(), registration, token.kind, &ending{outcome: "ignored", reason: "google_token_ended", user: token.user})
		w.WriteHeader(http.StatusOK)
		return
	}

	switch err := t.endAtGoogle(r.Context(), token.googleToken); {
	case errors.Is(err, googleauth.ErrNotHonoured):
		// The grant ended before, or the token did; either way nothing was
		// ended now, and the line says so rather than calling it revoked.
		t.events.revoked(r.Context(), registration, token.kind, &ending{outcome: "ignored", reason: "not_honoured", user: token.user})
	case err != nil:
		failure, end := revocationFailureOf(err)
		end.user = token.user
		t.events.revoked(r.Context(), registration, token.kind, &end)
		writeRefusal(w, t.issuer, failure)
		return
	default:
		t.events.revoked(r.Context(), registration, token.kind, &ending{outcome: "revoked", user: token.user})
	}
	w.WriteHeader(http.StatusOK)
}

// heldGrant reads a token a host holds, of either kind: a refresh token ends
// the grant with Google's refresh token inside it, an access token with
// Google's access token, which Google ends the whole grant for as well while
// it honours that token. A token is read whatever its own end, since a token
// past it may still name a grant alive at Google.
func (t *tokens) heldGrant(token string) (held, bool) {
	var refresh refreshGrant
	if t.openAs(t.refresh, token, &refresh) == nil {
		return held{kind: kindRefresh, client: refresh.Client, user: refresh.User, googleToken: refresh.GoogleRefreshToken}, true
	}
	var access accessGrant
	if t.openAs(t.access, token, &access) == nil {
		return held{
			kind: kindAccess, client: access.Client, user: access.User,
			googleToken: access.GoogleAccessToken, googleEnds: time.Unix(access.GoogleAccessExpiry, 0),
		}, true
	}
	return held{}, false
}

// endAtGoogle ends a grant at Google with a token of it.
func (t *tokens) endAtGoogle(ctx context.Context, googleToken string) error {
	if t.google == nil {
		return errUnconfigured
	}
	return t.google.Revoke(ctx, googleToken)
}

// revocationFailureOf is what a client hears when its grant could not be
// ended, and how the line tells it: Google not answering is a failure the
// client may try again after, as the protocol has it for a token not yet
// revoked; anything else is this server's own.
func revocationFailureOf(err error) (*refusal, ending) {
	switch {
	case errors.Is(err, googleauth.ErrUnavailable):
		return &refusal{http.StatusServiceUnavailable, "temporarily_unavailable",
				"Google did not answer, and the grant has not ended: try again later", "google_unavailable"},
			ending{outcome: "failed", reason: "google_unavailable", cause: err}
	case errors.Is(err, errUnconfigured):
		return &refusal{http.StatusServiceUnavailable, "temporarily_unavailable",
				"signing in is not set up on this server", "unconfigured"},
			ending{outcome: "failed", reason: "unconfigured", cause: err}
	default:
		return &refusal{http.StatusInternalServerError, "server_error",
				"the grant could not be ended: try again later", "internal"},
			ending{outcome: "failed", reason: "internal", cause: err, ours: true}
	}
}
