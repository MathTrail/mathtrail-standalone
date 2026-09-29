package oauthserver

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit/ratelimittest"
)

// day is a day, as the lifetimes of a sign-in count them.
const day = 24 * time.Hour

// A refresh token is worth new tokens for the same sign-in, and a new refresh
// token in its place, lasting thirty days from now. Google is not asked for
// anything while the Google token inside has more than seven minutes left.
func TestARefreshTokenIsWorthNewTokens(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, client)
	h.clock.advance(10 * time.Minute)
	answer := h.renew(t, refresh, client, nil)

	if answer.status != http.StatusOK || answer.field("scope") != "mcp" || answer.fields["expires_in"] != 900.0 {
		t.Fatalf("POST /oauth/token = %d %v, want new tokens for mcp, good for 900 seconds", answer.status, answer.fields)
	}
	if answer.field("refresh_token") == refresh {
		t.Error("the refresh token given back is the one presented, want a new one in its place")
	}
	renewed := h.openedRefresh(t, answer.field("refresh_token"))
	if renewed.SignedInAt != testDay.Unix() || renewed.ExpiresAt != testDay.Add(10*time.Minute+30*day).Unix() {
		t.Errorf("the new refresh token was signed in at %d and ends at %d, want the sign-in's time and thirty days from now",
			renewed.SignedInAt, renewed.ExpiresAt)
	}
	if access := h.openedAccess(t, answer.field("access_token")); access.GoogleAccessToken != googletest.AccessToken {
		t.Errorf("the access token carries %q, want Google's access token as it was", access.GoogleAccessToken)
	}
	if got := h.google.Renewals(); got != 0 {
		t.Errorf("Google renewed the grant %d times, want none", got)
	}

	user := h.ring.UserID("google-sub:" + googletest.Subject)
	if lines := h.lines(eventAuthRefresh); len(lines) != 1 || lines[0]["outcome"] != "ok" || lines[0]["user"] != user {
		t.Errorf("auth_refresh lines = %v, want one ok line naming the user", lines)
	}
	noLineCarries(t, h, googletest.AccessToken, googletest.RefreshToken, googletest.Subject, refresh, answer.field("access_token"), answer.field("refresh_token"))
}

// The Google token inside a host's tokens is renewed at Google when under seven
// minutes of it are left, and not before: a host is never handed an access
// token good for less than four minutes.
func TestAGoogleTokenWithLittleLeftIsRenewedFirst(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		left     time.Duration
		renewals int
		google   string
	}{
		{"seven minutes and a second left", 7*time.Minute + time.Second, 0, googletest.AccessToken},
		{"a second short of seven minutes left", 7*time.Minute - time.Second, 1, googletest.RenewedAccessToken},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			_, refresh := h.tokensFor(t, client)
			h.clock.advance(googletest.ExpiresIn*time.Second - tc.left)
			answer := h.renew(t, refresh, client, nil)

			if answer.status != http.StatusOK {
				t.Fatalf("POST /oauth/token = %d %v, want new tokens", answer.status, answer.fields)
			}
			if got := h.google.Renewals(); got != tc.renewals {
				t.Errorf("Google renewed the grant %d times, want %d", got, tc.renewals)
			}
			access := h.openedAccess(t, answer.field("access_token"))
			if access.GoogleAccessToken != tc.google {
				t.Errorf("the access token carries %q, want %q", access.GoogleAccessToken, tc.google)
			}
			if left := access.ExpiresAt - h.clock.Now().Unix(); left < int64(4*time.Minute/time.Second) {
				t.Errorf("the access token is good for %d seconds, want four minutes at least", left)
			}
		})
	}
}

// When Google renews a grant with a refresh token of its own in place of the
// one it was asked with, the host's new refresh token carries that one on.
func TestARefreshTokenGoogleGaveInPlaceIsCarriedOn(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, client)
	h.google.Misbehave(&googletest.Answer{Rotate: true})
	h.clock.advance(time.Hour)
	answer := h.renew(t, refresh, client, nil)

	if answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/token = %d %v, want new tokens", answer.status, answer.fields)
	}
	if got := h.openedRefresh(t, answer.field("refresh_token")).GoogleRefreshToken; got != googletest.RotatedRefreshToken {
		t.Errorf("the new refresh token carries Google's %q, want %q", got, googletest.RotatedRefreshToken)
	}
}

// A refresh token is worth new tokens to the host it was issued to alone, and
// only until it ends: to the client it was issued to, for no scope beyond its
// own and for its resource. A grant Google has ended is refused like an ended
// token, since only a new sign-in mends it; Google not answering is a failure
// the host may try again after; Google refusing the service's own client is an
// error of this server's.
func TestARefreshIsRefusedToAnybodyElse(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		changed func(h *signIn, access string) url.Values
		google  *googletest.Answer
		ended   bool
		after   time.Duration
		status  int
		code    string
		reason  string
	}{
		{name: "a token thirty days and a minute old", after: 30*day + 61*time.Second,
			status: 400, code: "invalid_grant", reason: "expired"},
		{name: "another client", changed: func(h *signIn, _ string) url.Values {
			return url.Values{"client_id": {h.register(hostRedirect, "Another host")}}
		}, status: 400, code: "invalid_grant", reason: "client"},
		{name: "no client", changed: presenting("client_id", ""), status: 400, code: "invalid_request", reason: "invalid_request"},
		{name: "a scope beyond the grant", changed: presenting("scope", "mcp admin"), status: 400, code: "invalid_scope", reason: "invalid_scope"},
		{name: "another resource", changed: presenting("resource", "https://other.example/mcp"), status: 400, code: "invalid_target", reason: "invalid_target"},
		{name: "a token nobody issued", changed: presenting("refresh_token", "mt1.r.xxxxxx.yyyy"), status: 400, code: "invalid_grant", reason: "unknown_token"},
		{name: "an access token", changed: func(_ *signIn, access string) url.Values {
			return url.Values{"refresh_token": {access}}
		}, status: 400, code: "invalid_grant", reason: "unknown_token"},
		{name: "no token", changed: presenting("refresh_token", ""), status: 400, code: "invalid_request", reason: "invalid_request"},
		{name: "a grant Google ended", ended: true, after: time.Hour,
			status: 400, code: "invalid_grant", reason: "grant_ended"},
		{name: "Google failing", google: &googletest.Answer{Status: 503}, after: time.Hour,
			status: 503, code: "temporarily_unavailable", reason: "google_unavailable"},
		{name: "Google renewing with a token too short to use", google: &googletest.Answer{Lifetime: 30}, after: time.Hour,
			status: 503, code: "temporarily_unavailable", reason: "google_unavailable"},
		{name: "Google refusing our client", google: &googletest.Answer{Status: 401, ErrorCode: "invalid_client"}, after: time.Hour,
			status: 500, code: "server_error", reason: "client_refused"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			access, refresh := h.tokensFor(t, client)
			var changed url.Values
			if tc.changed != nil {
				changed = tc.changed(h, access)
			}
			if tc.ended {
				h.revokeAtGoogle(t, googletest.RefreshToken)
			}
			if tc.google != nil {
				h.google.Misbehave(tc.google)
			}
			h.clock.advance(tc.after)
			answer := h.renew(t, refresh, client, changed)
			wantRefused(t, h, &answer, tc.status, tc.code, eventAuthRefresh, tc.reason)
		})
	}
}

// presenting is a change of one parameter of a refresh.
func presenting(name, value string) func(*signIn, string) url.Values {
	return func(*signIn, string) url.Values { return url.Values{name: {value}} }
}

// revokeAtGoogle ends the grant at the stand-in for Google, as a parent does in
// their Google account, behind the server's back.
func (h *signIn) revokeAtGoogle(t *testing.T, token string) {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, h.google.URL+"/revoke",
		strings.NewReader(url.Values{"token": {token}}.Encode()))
	if err != nil {
		t.Fatalf("NewRequest() error = %v, want nil", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := h.google.Client().Do(request)
	if err != nil {
		t.Fatalf("revoking at Google: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("revoking at Google = %d, want 200", response.StatusCode)
	}
}

// A refresh that failed on Google's side leaves a warning, and one that failed
// on this server's side an error, each with what went wrong.
func TestARefreshThatFailedSaysSoInItsLine(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		how   googletest.Answer
		level zapcore.Level
	}{
		{"Google failing", googletest.Answer{Status: 503}, zapcore.WarnLevel},
		{"Google refusing our client", googletest.Answer{Status: 401, ErrorCode: "invalid_client"}, zapcore.ErrorLevel},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			_, refresh := h.tokensFor(t, client)
			h.google.Misbehave(&tc.how)
			h.clock.advance(time.Hour)
			h.renew(t, refresh, client, nil)

			entries := h.logs.FilterMessage(eventAuthRefresh).All()
			if len(entries) != 1 || entries[0].Level != tc.level || entries[0].ContextMap()["error"] == nil {
				t.Errorf("auth_refresh lines = %v, want one %s line with the error", entries, tc.level)
			}
		})
	}
}

// A sign-in lasts ninety days from the moment the parent signed in, however
// often its tokens are renewed: no token issued for it ends later than that —
// an access token issued five minutes before is good for five minutes — and
// none is worth anything past it, not even within the minute a refresh token
// is allowed past its own end.
func TestASignInLastsNinetyDaysHoweverOftenItIsRenewed(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, client)
	ceiling := testDay.Add(90 * day).Unix()
	for renewed := 1; renewed <= 3; renewed++ {
		h.clock.advance(29 * day)
		answer := h.renew(t, refresh, client, nil)
		if answer.status != http.StatusOK {
			t.Fatalf("renewal %d, on day %d: POST /oauth/token = %d %v, want new tokens", renewed, 29*renewed, answer.status, answer.fields)
		}
		refresh = answer.field("refresh_token")
		if ends := h.openedRefresh(t, refresh).ExpiresAt; ends > ceiling {
			t.Errorf("renewal %d, on day %d: the refresh token ends at %d, past the sign-in's ninety days at %d", renewed, 29*renewed, ends, ceiling)
		}
	}
	if ends := h.openedRefresh(t, refresh).ExpiresAt; ends != ceiling {
		t.Errorf("the refresh token of day 87 ends at %d, want the sign-in's ninety days at %d", ends, ceiling)
	}

	h.clock.advance(3*day - 5*time.Minute)
	last := h.renew(t, refresh, client, nil)
	if last.status != http.StatusOK || last.fields["expires_in"] != 300.0 {
		t.Fatalf("five minutes before the ninetieth day: POST /oauth/token = %d %v, want an access token good for 300 seconds",
			last.status, last.fields)
	}

	h.clock.advance(5*time.Minute + 30*time.Second)
	answer := h.renew(t, last.field("refresh_token"), client, nil)
	wantRefused(t, h, &answer, http.StatusBadRequest, "invalid_grant", eventAuthRefresh, "expired")
}

// The clocks of two instances may disagree by a minute: a refresh token is
// still worth new tokens within the minute past its end.
func TestARefreshTokenIsJudgedWithAMinuteOfSkew(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, client)
	h.clock.advance(30*day + 59*time.Second)

	if answer := h.renew(t, refresh, client, nil); answer.status != http.StatusOK {
		t.Errorf("POST /oauth/token = %d %v, want a token 59 s past its end renewed", answer.status, answer.fields)
	}
}

// A server with no Google client configured can neither renew a Google token
// nor end a grant at Google: a host that asks it to is told to try again,
// rather than that its grant has ended.
func TestAServerWithoutGoogleNeitherRenewsNorRevokes(t *testing.T) {
	t.Parallel()

	h := newSignInWithoutGoogle(t)
	client := h.register(hostRedirect, hostName)
	refresh := sealedFor(t, h.ring, seal.PurposeRefresh, h.served.URL, &refreshGrant{
		User: "a-user", Client: digestOf(client), Resource: h.served.URL + "/mcp", Scope: "mcp", SignedInAt: testDay.Unix(),
		GoogleAccessToken: googletest.AccessToken, GoogleAccessExpiry: testDay.Unix(), GoogleRefreshToken: googletest.RefreshToken,
		IssuedAt: testDay.Unix(), ExpiresAt: testDay.Add(30 * day).Unix(),
	})

	renewed := h.renew(t, refresh, client, nil)
	wantRefused(t, h, &renewed, http.StatusServiceUnavailable, "temporarily_unavailable", eventAuthRefresh, "unconfigured")
	revoked := h.revoke(t, refresh, client, nil)
	wantRefused(t, h, &revoked, http.StatusServiceUnavailable, "temporarily_unavailable", eventAuthRevoke, "unconfigured")
}

// A refresh may name the scope it asks for, within the grant's, and the new
// tokens keep the grant's whole scope, as a new refresh token has to.
func TestARefreshKeepsTheScopeOfItsGrant(t *testing.T) {
	t.Parallel()

	for _, asked := range []string{"mcp", "mcp mcp", " mcp "} {
		t.Run(asked, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			_, refresh := h.tokensFor(t, client)
			answer := h.renew(t, refresh, client, url.Values{"scope": {asked}})

			if answer.status != http.StatusOK || answer.field("scope") != "mcp" {
				t.Fatalf("POST /oauth/token = %d %v, want new tokens for mcp", answer.status, answer.fields)
			}
			if got := h.openedRefresh(t, answer.field("refresh_token")).Scope; got != "mcp" {
				t.Errorf("the new refresh token grants %q, want mcp", got)
			}
		})
	}
}

// A refresh refused before it is read further — a client that tries to prove
// itself in a header, as the protocol library's own client does first — is a
// refresh in its line all the same.
func TestARefreshRefusedForItsHeaderIsARefreshInItsLine(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, client)
	answer := h.asHost(t, "/oauth/token", url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refresh}},
		http.Header{"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte(url.QueryEscape(client)+":"))}})

	wantRefused(t, h, &answer, http.StatusUnauthorized, "invalid_client", eventAuthRefresh, "invalid_client")
}

// An old refresh token used over and over renews the grant at Google each
// time, since the Google token it carries has ended; the account's pace of
// renewals holds that back. Past the pace a renewal is refused as a pause the
// host waits out, Google is not called, and the pace reached leaves one line,
// which names the account by its user identifier alone.
func TestRenewalsAtGoogleAreHeldToTheAccountsPace(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	h.server.tokens.renewals = ratelimittest.Keyed(t, 1)
	client := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, client)
	h.clock.advance(time.Hour)

	if first := h.renew(t, refresh, client, nil); first.status != http.StatusOK {
		t.Fatalf("the first renewal: POST /oauth/token = %d %v, want new tokens", first.status, first.fields)
	}
	again := h.renew(t, refresh, client, nil)
	wantRefused(t, h, &again, http.StatusServiceUnavailable, "temporarily_unavailable", eventAuthRefresh, "renewal_rate")
	if got := again.header.Get("Retry-After"); got != "60" {
		t.Errorf("Retry-After = %q, want the minute a renewal of this pace takes to come back", got)
	}
	if got := h.google.Renewals(); got != 1 {
		t.Errorf("Google renewed the grant %d times, want 1", got)
	}
	user := h.ring.UserID("google-sub:" + googletest.Subject)
	if lines := h.lines(eventLimitHit); len(lines) != 1 || lines[0]["limit"] != limitRenewals || lines[0]["user"] != user {
		t.Errorf("limit_hit lines = %v, want one naming %s and the user %s", lines, limitRenewals, user)
	}
}
