package oauthserver

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"slices"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

// A host ends its grant with either of its tokens, and the grant ends at
// Google, with the token of Google's that each carries: the refresh token
// with Google's refresh token, the access token with Google's access token,
// which Google ends the whole grant for as well. From then on Google renews
// the grant no more, and neither does the server.
func TestAHostEndsItsGrantAtGoogle(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		kind   string
		google string
	}{
		{kindRefresh, googletest.RefreshToken},
		{kindAccess, googletest.AccessToken},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			access, refresh := h.tokensFor(t, client)
			token := map[string]string{kindRefresh: refresh, kindAccess: access}[tc.kind]
			answer := h.revoke(t, token, client, nil)

			if answer.status != http.StatusOK || answer.fields != nil {
				t.Fatalf("POST /oauth/revoke = %d %v, want 200 and nothing more", answer.status, answer.fields)
			}
			if got := h.google.Revocations(); !slices.Equal(got, []string{tc.google}) {
				t.Errorf("Google was asked to end the grant with %q, want %q", got, tc.google)
			}
			user := h.ring.UserID("google-drive:" + googletest.PermissionID)
			if lines := h.lines(eventAuthRevoke); len(lines) != 1 || lines[0]["outcome"] != "revoked" ||
				lines[0]["token"] != tc.kind || lines[0]["user"] != user {
				t.Errorf("auth_revoke lines = %v, want one line of the %s token revoked", lines, tc.kind)
			}

			h.clock.advance(time.Hour)
			renewed := h.renew(t, refresh, client, nil)
			wantRefused(t, h, &renewed, http.StatusBadRequest, "invalid_grant", eventAuthRefresh, "grant_ended")
			noLineCarries(t, h, googletest.AccessToken, googletest.RefreshToken, googletest.PermissionID, access, refresh)
		})
	}
}

// A token past its end still names a grant that may be alive at Google, and
// ends it.
func TestATokenPastItsEndStillEndsItsGrant(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, client)
	h.clock.advance(31 * day)

	if answer := h.revoke(t, refresh, client, nil); answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/revoke = %d %v, want 200", answer.status, answer.fields)
	}
	if got := h.google.Revocations(); !slices.Equal(got, []string{googletest.RefreshToken}) {
		t.Errorf("Google was asked to end the grant with %q, want the refresh token", got)
	}
}

// A token none of this server's is answered as revoked, as the protocol asks
// — there is nothing a client could do about it — and Google is asked
// nothing.
func TestATokenNoneOfOursIsAnsweredAsRevoked(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		token func(t *testing.T, h *signIn, client string) string
	}{
		{"words", func(*testing.T, *signIn, string) string { return "not-a-token" }},
		{"the client's own identifier", func(_ *testing.T, _ *signIn, client string) string { return client }},
		{"a code", func(t *testing.T, h *signIn, client string) string { return h.signedInCode(t, client) }},
		{"a token of another issuer", func(t *testing.T, h *signIn, _ string) string {
			return sealedFor(t, h.ring, seal.PurposeRefresh, "https://other.example", &refreshGrant{GoogleRefreshToken: googletest.RefreshToken})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			answer := h.revoke(t, tc.token(t, h, client), client, nil)

			if answer.status != http.StatusOK {
				t.Fatalf("POST /oauth/revoke = %d %v, want 200", answer.status, answer.fields)
			}
			if got := h.google.Revocations(); len(got) != 0 {
				t.Errorf("Google was asked to end a grant with %q, want nothing asked", got)
			}
			if lines := h.lines(eventAuthRevoke); len(lines) != 1 || lines[0]["outcome"] != "ignored" || lines[0]["reason"] != "unknown_token" {
				t.Errorf("auth_revoke lines = %v, want one line of a token ignored", lines)
			}
		})
	}
}

// A grant is ended by the host it was issued to alone, and one Google could
// not end is not called ended: a Google that did not answer is a failure the
// host may try again after, and Google refusing the request otherwise is an
// error of this server's.
func TestARevocationIsRefusedWhenItCannotBeDone(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		changed func(h *signIn) url.Values
		google  *googletest.Answer
		header  bool
		status  int
		code    string
		reason  string
	}{
		{name: "another client", changed: func(h *signIn) url.Values {
			return url.Values{"client_id": {h.register(hostRedirect, "Another host")}}
		}, status: 400, code: "invalid_grant", reason: "client"},
		{name: "no client", changed: func(*signIn) url.Values { return url.Values{"client_id": {""}} },
			status: 400, code: "invalid_request", reason: "invalid_request"},
		{name: "no token", changed: func(*signIn) url.Values { return url.Values{"token": {""}} },
			status: 400, code: "invalid_request", reason: "invalid_request"},
		{name: "a client proving itself in a header", header: true, status: 401, code: "invalid_client", reason: "invalid_client"},
		{name: "Google failing", google: &googletest.Answer{RevokeStatus: 503},
			status: 503, code: "temporarily_unavailable", reason: "google_unavailable"},
		{name: "Google refusing the request", google: &googletest.Answer{RevokeStatus: 400, RevokeError: "invalid_request"},
			status: 500, code: "server_error", reason: "internal"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			_, refresh := h.tokensFor(t, client)
			if tc.google != nil {
				h.google.Misbehave(tc.google)
			}
			var answer hostReply
			if tc.header {
				answer = h.asHost(t, "/oauth/revoke", url.Values{"token": {refresh}},
					http.Header{"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte(url.QueryEscape(client)+":"))}})
			} else {
				var changed url.Values
				if tc.changed != nil {
					changed = tc.changed(h)
				}
				answer = h.revoke(t, refresh, client, changed)
			}
			wantRefused(t, h, &answer, tc.status, tc.code, eventAuthRevoke, tc.reason)
		})
	}
}

// A revocation Google could not do leaves a warning when Google failed it and
// an error when this server did, each with what went wrong.
func TestARevocationThatFailedSaysSoInItsLine(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		how   googletest.Answer
		level zapcore.Level
	}{
		{"Google failing", googletest.Answer{RevokeStatus: 503}, zapcore.WarnLevel},
		{"Google refusing the request", googletest.Answer{RevokeStatus: 400, RevokeError: "invalid_request"}, zapcore.ErrorLevel},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignIn(t)
			client := h.register(hostRedirect, hostName)
			_, refresh := h.tokensFor(t, client)
			h.google.Misbehave(&tc.how)
			h.revoke(t, refresh, client, nil)

			entries := h.logs.FilterMessage(eventAuthRevoke).All()
			if len(entries) != 1 || entries[0].Level != tc.level || entries[0].ContextMap()["error"] == nil {
				t.Errorf("auth_revoke lines = %v, want one %s line with the error", entries, tc.level)
			}
		})
	}
}

// An access token whose Google token inside has ended can end nothing at
// Google: it is answered as the protocol asks, Google is asked nothing, and
// the line says why. The host's refresh token still ends the grant.
func TestAnAccessTokenPastItsGoogleTokenEndsNothing(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	access, refresh := h.tokensFor(t, client)
	h.clock.advance(googletest.ExpiresIn * time.Second)

	if answer := h.revoke(t, access, client, nil); answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/revoke = %d %v, want 200", answer.status, answer.fields)
	}
	if got := h.google.Revocations(); len(got) != 0 {
		t.Errorf("Google was asked to end the grant with %q, want nothing asked", got)
	}
	if lines := h.lines(eventAuthRevoke); len(lines) != 1 || lines[0]["outcome"] != "ignored" || lines[0]["reason"] != "google_token_ended" {
		t.Errorf("auth_revoke lines = %v, want one line of a token that could end nothing", lines)
	}

	if answer := h.revoke(t, refresh, client, nil); answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/revoke with the refresh token = %d %v, want 200", answer.status, answer.fields)
	}
	if got := h.google.Revocations(); !slices.Equal(got, []string{googletest.RefreshToken}) {
		t.Errorf("Google was asked to end the grant with %q, want the refresh token", got)
	}
}

// A grant ended already is not ended again: Google no longer honours the
// token, the host is answered as the protocol asks, and the line says that
// nothing was ended rather than calling it revoked.
func TestAGrantEndedAlreadyIsNotCalledRevokedAgain(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	client := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, client)
	for range 2 {
		if answer := h.revoke(t, refresh, client, nil); answer.status != http.StatusOK {
			t.Fatalf("POST /oauth/revoke = %d %v, want 200", answer.status, answer.fields)
		}
	}
	lines := h.lines(eventAuthRevoke)
	if len(lines) != 2 || lines[0]["outcome"] != "revoked" || lines[1]["outcome"] != "ignored" || lines[1]["reason"] != "not_honoured" {
		t.Errorf("auth_revoke lines = %v, want the grant revoked once and then nothing ended", lines)
	}
}
