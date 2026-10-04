package oauthserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
)

// The lifetimes of what a host holds.
const (
	// codeLifetime is how long a code is worth the host's tokens. A server
	// that keeps nothing cannot mark a code used, so the time it can be used
	// in is made small instead.
	codeLifetime = 60 * time.Second
	// accessLifetime is how long an access token is good for, at most: short
	// enough that one stolen is a small loss, long enough that renewing it is
	// rare.
	accessLifetime = 15 * time.Minute
	// refreshLifetime is how long a refresh token lasts unused. Each use
	// gives another, so a host in weekly use never has to sign in again.
	refreshLifetime = 30 * 24 * time.Hour
	// sessionLifetime is how long a sign-in lasts from the moment the parent
	// signed in, however often its tokens are renewed. Nothing records which
	// refresh tokens were used, so this is what bounds a stolen one.
	sessionLifetime = 90 * 24 * time.Hour
	// requestSpan is the longest a tool goes on calling Drive after its request
	// is let in: at most six calls, each given the ten seconds a call to Drive
	// has — a first profile after a file gone from under the memory — or five
	// beside two runs of a solver of two seconds each. A request cut off, or
	// answered late, does not stop the tool, so its own work is the bound; a
	// call to Drive given longer, more calls, or a wait for a solver's slot
	// move it.
	requestSpan = 60 * time.Second
	// googleMargin is how long before the Google access token inside it an
	// access token ends, so that a request it lets in reaches Drive with a
	// Google token that has not ended. The resource takes a token for
	// clockSkew past its end; the clock that checks it may run clockSkew
	// behind the one that issued it; and the tool then goes on calling Drive
	// for requestSpan. Less than that, and the last requests of a token meet
	// an ended Google token, which Drive answers as it answers access taken
	// back.
	googleMargin = 2*clockSkew + requestSpan
	// googleRenewal is how little may be left of the Google access token a
	// host's tokens are issued with before Google is asked for a new one: the
	// margin above and four minutes more, so that an access token is good for
	// four minutes at least whenever Google's own lasts seven.
	googleRenewal = googleMargin + 4*time.Minute
)

// session is a sign-in as the host's tokens carry it from one to the next:
// whom the parent signed in as and when, and from which country, the client it
// is for, the resource and the scope, and Google's grant. The country is the
// one of the sign-in itself: a renewal comes from the host's servers and keeps
// it as it was.
type session struct {
	user       string
	country    string
	client     string
	resource   string
	scope      string
	signedInAt time.Time
	google     googleGrant
}

// googleGrant is Google's grant as a session carries it: the access token the
// account's Drive is reached with, until when, and the refresh token a new one
// is asked for with.
type googleGrant struct {
	accessToken  string
	accessExpiry time.Time
	refreshToken string
}

// accessGrant is what an access token carries: whom it signs a request in as
// and the country that sign-in was made from — empty in a token of a sign-in
// made before the country was kept — the digest of the client it was issued
// to, the resource it is for — its audience — and the scope, the Google access
// token the account's Drive is reached with, and when it was issued and ends.
// It carries no Google refresh token: one stolen buys minutes of access to a
// single file, and no way to renew it.
type accessGrant struct {
	User               string `json:"user"`
	Country            string `json:"country,omitempty"`
	Client             string `json:"client"`
	Resource           string `json:"resource"`
	Scope              string `json:"scope"`
	GoogleAccessToken  string `json:"google_access_token"`
	GoogleAccessExpiry int64  `json:"google_access_expiry"`
	IssuedAt           int64  `json:"issued_at"`
	ExpiresAt          int64  `json:"expires_at"`
}

// refreshGrant is what a refresh token carries: the whole session, Google's
// refresh token among it, and when the token was issued and ends.
type refreshGrant struct {
	User               string `json:"user"`
	Country            string `json:"country,omitempty"`
	Client             string `json:"client"`
	Resource           string `json:"resource"`
	Scope              string `json:"scope"`
	SignedInAt         int64  `json:"signed_in_at"`
	GoogleAccessToken  string `json:"google_access_token"`
	GoogleAccessExpiry int64  `json:"google_access_expiry"`
	GoogleRefreshToken string `json:"google_refresh_token"`
	IssuedAt           int64  `json:"issued_at"`
	ExpiresAt          int64  `json:"expires_at"`
}

// sessionOf is the session a refresh token carries.
func sessionOf(grant *refreshGrant) *session {
	return &session{
		user:       grant.User,
		country:    grant.Country,
		client:     grant.Client,
		resource:   grant.Resource,
		scope:      grant.Scope,
		signedInAt: time.Unix(grant.SignedInAt, 0),
		google: googleGrant{
			accessToken:  grant.GoogleAccessToken,
			accessExpiry: time.Unix(grant.GoogleAccessExpiry, 0),
			refreshToken: grant.GoogleRefreshToken,
		},
	}
}

// issued is what a host is handed for a session: an access token and a
// refresh token, how many seconds the access token is good for, and the scope
// both grant.
type issued struct {
	access    string
	refresh   string
	expiresIn int64
	scope     string
}

// The ways a session could not be given tokens, beside Google's own.
var (
	// errUnconfigured is a session that needs Google on a server with no
	// Google client configured.
	errUnconfigured = errors.New("oauth: no Google sign-in is configured")
	// errShortGrant is a Google access token that ends too soon to issue an
	// access token with.
	errShortGrant = errors.New("oauth: Google's access token ends too soon")
	// errSessionEnded is a session past the days a sign-in lasts.
	errSessionEnded = errors.New("oauth: the sign-in has ended")
)

// issue hands a host its tokens for a session, with a Google access token
// fresh enough to issue them with: one with little left is renewed at Google
// first. The access token ends before the Google token inside it, and neither
// token outlives the session: a session that has ended is given none.
func (t *tokens) issue(ctx context.Context, s *session) (issued, error) {
	// Counted in whole seconds, the unit the tokens carry and the host is
	// told in.
	sessionEnd := s.signedInAt.Unix() + int64(sessionLifetime/time.Second)
	if sessionEnd <= t.now().Unix() {
		return issued{}, errSessionEnded
	}
	if err := t.freshen(ctx, s); err != nil {
		return issued{}, err
	}

	now := t.now().Unix()
	accessEnd := min(now+int64(accessLifetime/time.Second),
		s.google.accessExpiry.Unix()-int64(googleMargin/time.Second), sessionEnd)
	if accessEnd <= now {
		return issued{}, fmt.Errorf("%w: it ends at %d, now is %d", errShortGrant, s.google.accessExpiry.Unix(), now)
	}
	refreshEnd := min(now+int64(refreshLifetime/time.Second), sessionEnd)

	access, err := t.sealAs(t.access, &accessGrant{
		User:               s.user,
		Country:            s.country,
		Client:             s.client,
		Resource:           s.resource,
		Scope:              s.scope,
		GoogleAccessToken:  s.google.accessToken,
		GoogleAccessExpiry: s.google.accessExpiry.Unix(),
		IssuedAt:           now,
		ExpiresAt:          accessEnd,
	})
	if err != nil {
		return issued{}, err
	}
	refresh, err := t.sealAs(t.refresh, &refreshGrant{
		User:               s.user,
		Country:            s.country,
		Client:             s.client,
		Resource:           s.resource,
		Scope:              s.scope,
		SignedInAt:         s.signedInAt.Unix(),
		GoogleAccessToken:  s.google.accessToken,
		GoogleAccessExpiry: s.google.accessExpiry.Unix(),
		GoogleRefreshToken: s.google.refreshToken,
		IssuedAt:           now,
		ExpiresAt:          refreshEnd,
	})
	if err != nil {
		return issued{}, err
	}
	return issued{access: access, refresh: refresh, expiresIn: accessEnd - now, scope: s.scope}, nil
}

// freshen renews the session's Google access token at Google when little is
// left of it, and goes on with the refresh token the renewal names: the one
// it had, or the one Google gave in its place.
//
// A renewal is held to the pace of its account. A grant needs one about
// every hour, but a refresh token stays good to its own end, and one issued an
// hour ago carries a Google token past its end: without the pace, whoever
// kept one could have this server call Google with every use of it.
func (t *tokens) freshen(ctx context.Context, s *session) error {
	if s.google.accessExpiry.Sub(t.now()) >= googleRenewal {
		return nil
	}
	if t.google == nil {
		return errUnconfigured
	}
	if verdict := t.renewals.Take(s.user); !verdict.Allowed {
		if verdict.Began {
			t.events.limited(ctx, limitRenewals, s.user)
		}
		return &renewalPaced{after: verdict.RetryAfter}
	}
	renewal, err := t.google.Refresh(ctx, s.google.refreshToken)
	if err != nil {
		return fmt.Errorf("oauth: renew the Google access token: %w", err)
	}
	s.google.accessToken, s.google.accessExpiry, s.google.refreshToken = renewal.AccessToken, renewal.Expiry, renewal.RefreshToken
	return nil
}

// renewalPaced is a renewal at Google past the pace of its account, and how
// long the next one waits.
type renewalPaced struct {
	after time.Duration
}

func (*renewalPaced) Error() string { return "oauth: renewed at Google too often" }

// sealAs seals what a token carries, bound to this issuer, with the purpose
// of the ring given.
func (t *tokens) sealAs(ring sealer, carried any) (string, error) {
	plain, err := json.Marshal(carried)
	if err != nil {
		return "", fmt.Errorf("oauth: encode a token: %w", err)
	}
	sealed, err := ring.Seal(plain, t.issuer)
	if err != nil {
		return "", fmt.Errorf("oauth: seal a token: %w", err)
	}
	return sealed, nil
}

// openAs opens a token this server sealed, for itself, with the purpose of the
// ring given, into what it carries.
func (t *tokens) openAs(ring sealer, token string, carried any) error {
	plain, err := ring.Open(token, t.issuer)
	if err != nil {
		return err
	}
	return json.Unmarshal(plain, carried)
}

// googleUnavailable reports whether a session could not be given tokens
// because Google did not answer, or answered with a token too short to use:
// a host may try again later.
func googleUnavailable(err error) bool {
	return errors.Is(err, googleauth.ErrUnavailable) || errors.Is(err, errShortGrant)
}
