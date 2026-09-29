package oauthserver

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/oauthex"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
)

// tokens issues the tokens a host holds for a sign-in, renews them, ends the
// grant they carry, and reads the access tokens requests to the resource come
// with. Nothing it issued is kept anywhere: a token carries its sign-in,
// sealed.
type tokens struct {
	issuer   string
	resource string
	scope    string
	clients  *clients
	codes    sealer
	access   sealer
	refresh  sealer
	google   googleauth.SignIn
	renewals ratelimit.Limiter
	events   *signInLog
	now      func() time.Time
}

// tokenAnswer is a host's tokens as the token endpoint hands them over
// (RFC 6749 5.1).
type tokenAnswer struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// serveToken answers a request for tokens (RFC 6749 4.1.3 and 6): a code this
// server issued, or a refresh token, is worth an access token and a new
// refresh token for the same sign-in. The request is judged first, and only a
// request that may have tokens reaches Google, when the tokens need a fresh
// Google access token inside them.
func (t *tokens) serveToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	form, refused := readForm(w, r)
	registration := t.clients.kindOf(form.Get("client_id"))
	event := eventAuthToken
	if form.Get("grant_type") == "refresh_token" {
		event = eventAuthRefresh
	}
	var s *session
	if refused == nil {
		switch form.Get("grant_type") {
		case "authorization_code":
			s, refused = t.sessionOfCode(form)
		case "refresh_token":
			s, refused = t.sessionOfRefresh(form)
		case "":
			refused = &refusal{http.StatusBadRequest, "invalid_request", "grant_type is required", "invalid_request"}
		default:
			refused = &refusal{http.StatusBadRequest, "unsupported_grant_type",
				"the grants here are authorization_code and refresh_token", "unsupported_grant_type"}
		}
	}
	if refused != nil {
		t.events.granted(r.Context(), event, registration, form, &ending{outcome: "refused", reason: refused.reason})
		writeRefusal(w, t.issuer, refused)
		return
	}

	given, err := t.issue(r.Context(), s)
	if err != nil {
		failure, end := failureOf(err)
		end.user = s.user
		if paced := new(*renewalPaced); errors.As(err, paced) {
			w.Header().Set("Retry-After", strconv.Itoa(max(1, int(math.Ceil((*paced).after.Seconds())))))
		}
		t.events.granted(r.Context(), event, registration, form, &end)
		writeRefusal(w, t.issuer, failure)
		return
	}
	t.events.granted(r.Context(), event, registration, form, &ending{outcome: "ok", user: s.user})
	writeJSON(w, http.StatusOK, &tokenAnswer{
		AccessToken:  given.access,
		TokenType:    "Bearer",
		ExpiresIn:    given.expiresIn,
		RefreshToken: given.refresh,
		Scope:        given.scope,
	})
}

// sessionOfCode is the session a code stands for, when the host presenting it
// may have it: a code this server issued, still good, to the client that
// presents it, with the verifier its challenge was made from — and, when the
// host names them, for the address the code was sent to and for the resource
// it was asked for.
func (t *tokens) sessionOfCode(form url.Values) (*session, *refusal) {
	clientID, verifier := form.Get("client_id"), form.Get("code_verifier")
	switch {
	case form.Get("code") == "" || clientID == "" || verifier == "":
		return nil, &refusal{http.StatusBadRequest, "invalid_request",
			"code, code_verifier and client_id are required", "invalid_request"}
	case !isVerifier(verifier):
		return nil, &refusal{http.StatusBadRequest, "invalid_request",
			fmt.Sprintf("code_verifier is %d to %d letters, digits and -._~", shortestVerifier, longestVerifier), "invalid_pkce"}
	}
	var code grantCode
	if err := t.openAs(t.codes, form.Get("code"), &code); err != nil {
		return nil, invalidGrant("the code is not one this server issued", "unknown_code")
	}

	age := t.now().Sub(time.Unix(code.IssuedAt, 0))
	switch {
	case age > codeLifetime || age < -clockSkew:
		return nil, invalidGrant("the code has expired: sign in again", "expired")
	case !sameDigest(digestOf(clientID), code.Client):
		return nil, invalidGrant("the code was issued to another client", "client")
	case form.Has("redirect_uri") && form.Get("redirect_uri") != code.RedirectURI:
		return nil, invalidGrant("redirect_uri is not the address the code was sent to", "redirect_uri")
	case !sameDigest(challengeOf(verifier), code.Challenge):
		return nil, invalidGrant("code_verifier is not the one the code's challenge was made from", "pkce")
	case !forResource(form, code.Resource):
		return nil, otherResource(code.Resource)
	}
	return &session{
		user:       code.User,
		client:     code.Client,
		resource:   code.Resource,
		scope:      code.Scope,
		signedInAt: time.Unix(code.IssuedAt, 0),
		google: googleGrant{
			accessToken:  code.GoogleAccessToken,
			accessExpiry: time.Unix(code.GoogleAccessExpiry, 0),
			refreshToken: code.GoogleRefreshToken,
		},
	}, nil
}

// sessionOfRefresh is the session a refresh token carries, when the host
// presenting it may go on with it: a refresh token this server issued, not
// past its end, to the client that presents it — and, when the host names
// them, for no scope beyond the one it grants and for the resource it is for.
// A refresh that names a scope is given the grant's own all the same: a new
// refresh token keeps the scope of the one it replaces (RFC 6749 6), and with
// one scope there is nothing narrower to give an access token.
func (t *tokens) sessionOfRefresh(form url.Values) (*session, *refusal) {
	clientID := form.Get("client_id")
	if form.Get("refresh_token") == "" || clientID == "" {
		return nil, &refusal{http.StatusBadRequest, "invalid_request",
			"refresh_token and client_id are required", "invalid_request"}
	}
	var grant refreshGrant
	if err := t.openAs(t.refresh, form.Get("refresh_token"), &grant); err != nil {
		return nil, invalidGrant("the refresh token is not one this server issued", "unknown_token")
	}

	switch {
	case t.now().After(time.Unix(grant.ExpiresAt, 0).Add(clockSkew)):
		return nil, invalidGrant("the refresh token has expired: sign in again", "expired")
	case !sameDigest(digestOf(clientID), grant.Client):
		return nil, invalidGrant("the refresh token was issued to another client", "client")
	case !within(form.Get("scope"), grant.Scope):
		return nil, &refusal{http.StatusBadRequest, "invalid_scope",
			"a refresh asks for no scope beyond " + grant.Scope, "invalid_scope"}
	case !forResource(form, grant.Resource):
		return nil, otherResource(grant.Resource)
	}
	return sessionOf(&grant), nil
}

// failureOf is what a host hears when its tokens could not be issued, and how
// the line tells it. A sign-in past its days, and a grant Google has ended, are
// refused, since only a new sign-in mends either. Google not answering is a failure the host may try again
// after. Anything else is this server's own: Google refusing the service's
// client fails every renewal alike until the deployment is fixed.
func failureOf(err error) (*refusal, ending) {
	switch {
	case errors.Is(err, errSessionEnded):
		return invalidGrant("the sign-in has ended: sign in again", "expired"),
			ending{outcome: "refused", reason: "expired"}
	case errors.Is(err, googleauth.ErrGrantEnded):
		return invalidGrant("the parent's access at Google has ended: sign in again", "grant_ended"),
			ending{outcome: "refused", reason: "grant_ended"}
	case errors.As(err, new(*renewalPaced)):
		return &refusal{http.StatusServiceUnavailable, "temporarily_unavailable",
				"this sign-in was renewed at Google too often in a short time: try again after the time Retry-After gives", "renewal_rate"},
			ending{outcome: "refused", reason: "renewal_rate"}
	case googleUnavailable(err):
		return &refusal{http.StatusServiceUnavailable, "temporarily_unavailable",
				"Google did not answer: try again later", "google_unavailable"},
			ending{outcome: "failed", reason: "google_unavailable", cause: err}
	case errors.Is(err, errUnconfigured):
		return &refusal{http.StatusServiceUnavailable, "temporarily_unavailable",
				"signing in is not set up on this server", "unconfigured"},
			ending{outcome: "failed", reason: "unconfigured", cause: err}
	case errors.Is(err, googleauth.ErrClient):
		return &refusal{http.StatusInternalServerError, "server_error",
				"the tokens could not be issued: try again later", "client_refused"},
			ending{outcome: "failed", reason: "client_refused", cause: err, ours: true}
	default:
		return &refusal{http.StatusInternalServerError, "server_error",
				"the tokens could not be issued: try again", "internal"},
			ending{outcome: "failed", reason: "internal", cause: err, ours: true}
	}
}

// invalidGrant is a code or a token refused as the protocol refuses one it
// will not honour (RFC 6749 5.2), with the reason its line gives.
func invalidGrant(description, reason string) *refusal {
	return &refusal{http.StatusBadRequest, "invalid_grant", description, reason}
}

// forResource reports whether a request is for the resource a grant was given
// for: it names none, or names that one, compared as the protocol compares a
// resource. A request for more than one resource is for one this server does
// not have.
func forResource(form url.Values, resource string) bool {
	given := form["resource"]
	return len(given) == 0 || (len(given) == 1 && oauthex.MatchesResource(given, resource))
}

// otherResource is a request refused for a resource other than the one its
// grant is for (RFC 8707).
func otherResource(resource string) *refusal {
	return &refusal{http.StatusBadRequest, "invalid_target", "the one resource here is " + resource, "invalid_target"}
}

// within reports whether a refresh asks for no scope beyond the one its grant
// gives.
func within(asked, granted string) bool {
	given := strings.Fields(granted)
	return !slices.ContainsFunc(strings.Fields(asked), func(word string) bool { return !slices.Contains(given, word) })
}

// The lengths a verifier may be (RFC 7636 4.1).
const (
	shortestVerifier = 43
	longestVerifier  = 128
)

// isVerifier reports whether a value is a verifier as RFC 7636 makes one:
// letters, digits and the four marks of an address that need no escape, and
// long enough to be no guess. A host whose verifier is shorter has only
// itself to blame when its code is stolen, but it is told rather than given
// tokens on the strength of it.
func isVerifier(value string) bool {
	if len(value) < shortestVerifier || len(value) > longestVerifier {
		return false
	}
	for _, c := range value {
		if !unreserved(c) {
			return false
		}
	}
	return true
}

// unreserved reports whether a character is one an address carries without an
// escape: a letter, a digit, or one of - . _ ~ (RFC 3986 2.3).
func unreserved(c rune) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || strings.ContainsRune("-._~", c)
}

// challengeOf is the S256 challenge of a verifier (RFC 7636 4.2).
func challengeOf(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

// sameDigest reports whether two digests are the same, taking as long either
// way.
func sameDigest(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
