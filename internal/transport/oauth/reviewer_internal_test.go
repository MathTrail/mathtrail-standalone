package oauthserver

import (
	"errors"
	"html"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/googleauth/googletest"
	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit/ratelimittest"
)

// The reviewer's sign-in of these cases: a password of thirty-two characters,
// as the recipe that makes one makes it, and a demo account other than the
// one the stand-in for Google signs a parent in as, whose grant that Google
// renews all the same.
const (
	reviewerPassword = "Xk3-vQ9_tLm2Wp7Rz4Yb8Nc1Jd6Hf5Gs"
	demoSubject      = "100000000000000000001"
)

// demoReviewer is the reviewer's sign-in of these cases.
func demoReviewer() *Reviewer {
	return &Reviewer{Password: reviewerPassword, Subject: demoSubject, RefreshToken: googletest.RefreshToken}
}

// signInAsReviewer takes a reviewer from the host's request to the consent
// screen, and posts the password typed there.
func (h *signIn) signInAsReviewer(reviewer *browser, clientID, typed string) reply {
	h.t.Helper()

	_, request := h.toConsent(reviewer, clientID)
	return reviewer.post(h.served.URL+"/oauth/consent", url.Values{
		"request": {request}, "decision": {"reviewer"}, "password": {typed},
	})
}

// reviewersTokens are the tokens a host holds once a reviewer signed in.
func (h *signIn) reviewersTokens(t *testing.T, clientID string) (access, refresh string) {
	t.Helper()

	code := answerAt(t, h.signInAsReviewer(h.browser(t), clientID, reviewerPassword)).Get("code")
	answer := h.exchange(t, code, clientID, nil)
	if answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/token = %d %v, want the host's tokens", answer.status, answer.fields)
	}
	return answer.field("access_token"), answer.field("refresh_token")
}

// The consent screen offers the reviewer's password only where a reviewer's
// sign-in is configured, in a form of its own that posts the same request.
func TestTheConsentScreenOffersThePasswordOnlyWhereThereIsAReviewer(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		h        func(t *testing.T) *signIn
		offers   bool
		requests int
	}{
		{"no reviewer", newSignIn, false, 1},
		{"a reviewer", func(t *testing.T) *signIn { t.Helper(); return newSignInWithReviewer(t, demoReviewer()) }, true, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := tc.h(t)
			screen, request := h.toConsent(h.browser(t), h.register(hostRedirect, hostName))
			if got := strings.Contains(screen.body, `name="password"`); got != tc.offers {
				t.Errorf("the consent screen offers a password: %v, want %v", got, tc.offers)
			}
			if got := strings.Count(screen.body, `name="request" value="`+request+`"`); got != tc.requests {
				t.Errorf("the consent screen posts its request in %d forms, want %d", got, tc.requests)
			}
		})
	}
}

// A reviewer who types the reviewer's password signs in as the demo account,
// with no step at Google: the host is sent back a code for the demo account,
// with a Google access token renewed for it, the one call made to Google, and
// no country, since a reviewer's network says nothing of where a family is.
// The sign-in leaves one line, and the browser keeps neither the sign-in's
// cookie nor an approval, so a reviewer who comes back is asked again.
func TestAReviewerSignsInAsTheDemoAccount(t *testing.T) {
	t.Parallel()

	h := newSignInWithReviewer(t, demoReviewer())
	// The service knows the country of a parent's browser; a reviewer's is
	// none of a family's.
	h.server.flow.countryOf = func(*http.Request) string { return "NZ" }
	clientID := h.register(hostRedirect, hostName)
	reviewer := h.browser(t)

	back := h.signInAsReviewer(reviewer, clientID, reviewerPassword)
	if back.status != http.StatusSeeOther {
		t.Fatalf("POST /oauth/consent = %d, want 303 back to the host", back.status)
	}
	answer := answerAt(t, back)
	if answer.Get("state") != hostState || answer.Get("iss") != h.served.URL || answer.Has("error") {
		t.Fatalf("the host was sent %v, want a code with its own state and this issuer", answer)
	}
	demo := h.ring.UserID("google-sub:" + demoSubject)
	want := grantCode{
		User:               demo,
		Client:             digestOf(clientID),
		RedirectURI:        hostRedirect,
		Challenge:          googletest.ChallengeOf(hostVerifier),
		Resource:           h.served.URL + "/mcp",
		Scope:              "mcp",
		GoogleAccessToken:  googletest.RenewedAccessToken,
		GoogleAccessExpiry: testDay.Add(googletest.ExpiresIn * time.Second).Unix(),
		GoogleRefreshToken: googletest.RefreshToken,
		IssuedAt:           testDay.Unix(),
	}
	if code := h.opened(answer.Get("code")); code != want {
		t.Errorf("the code opens into %+v, want %+v", code, want)
	}

	tokens := h.exchange(t, answer.Get("code"), clientID, nil)
	if tokens.status != http.StatusOK {
		t.Fatalf("POST /oauth/token = %d %v, want the host's tokens", tokens.status, tokens.fields)
	}
	if got := h.openedAccess(t, tokens.field("access_token")).User; got != demo {
		t.Errorf("the access token signs in %q, want the demo account %q", got, demo)
	}
	if got := h.google.Renewals(); got != 1 {
		t.Errorf("Google renewed the demo account's grant %d times, want once, for the consent screen's code", got)
	}
	for _, name := range []string{csrfCookie, consentCookie} {
		if reviewer.cookie("/oauth/consent", name) != "" {
			t.Errorf("the browser keeps %s after the reviewer's sign-in, want neither cookie", name)
		}
	}
	if lines := h.lines(eventAuthConsent); len(lines) != 1 || lines[0]["outcome"] != "reviewer" || lines[0]["user"] != demo {
		t.Errorf("auth_consent lines = %v, want one of the reviewer signed in as the demo account", lines)
	}
	noLineCarries(t, h, reviewerPassword, demoSubject, googletest.RenewedAccessToken, googletest.RefreshToken, answer.Get("code"))

	if again := reviewer.get(h.authorizeURL(clientID, nil)); again.status != http.StatusOK || !strings.Contains(again.body, `name="password"`) {
		t.Errorf("GET /oauth/authorize again = %d, want the consent screen with the password once more", again.status)
	}
}

// A password that is not the reviewer's goes no further than a page that says
// so: Google is asked nothing, the host hears nothing, and the line says why
// without what was typed.
func TestAPasswordNotTheReviewersGoesNoFurther(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		typed string
	}{
		{"nothing typed", ""},
		{"a letter wrong", reviewerPassword[:len(reviewerPassword)-1] + "x"},
		{"a letter short", reviewerPassword[:len(reviewerPassword)-1]},
		{"a letter more", reviewerPassword + "s"},
		{"in capitals", strings.ToUpper(reviewerPassword)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newSignInWithReviewer(t, demoReviewer())
			stopped := h.signInAsReviewer(h.browser(t), h.register(hostRedirect, hostName), tc.typed)
			wantStoppedAtThePasswordsPage(t, h, &stopped)
			if tc.typed != "" {
				noLineCarries(t, h, tc.typed)
			}
		})
	}
}

// wantStoppedAtThePasswordsPage holds a reviewer's sign-in to having stopped
// at the page that says the password is not the reviewers', with Google asked
// nothing and one line of why.
func wantStoppedAtThePasswordsPage(t *testing.T, h *signIn, stopped *reply) {
	t.Helper()

	if stopped.status != http.StatusForbidden || stopped.location != "" {
		t.Fatalf("POST /oauth/consent = %d to %q, want 403 and a page", stopped.status, stopped.location)
	}
	if heading := pageWords(t)["refusal.password.heading"]; !strings.Contains(stopped.body, html.EscapeString(heading)) {
		t.Errorf("the page does not say %q", heading)
	}
	if got := h.google.Renewals(); got != 0 {
		t.Errorf("Google renewed the demo account's grant %d times, want never", got)
	}
	if lines := h.lines(eventAuthReject); len(lines) != 1 || lines[0]["step"] != stepConsent || lines[0]["reason"] != "reviewer_password" {
		t.Errorf("auth_reject lines = %v, want one of the consent step with reviewer_password", lines)
	}
	if lines := h.lines(eventAuthConsent); len(lines) != 0 {
		t.Errorf("auth_consent lines = %v, want none", lines)
	}
}

// Where no reviewer's sign-in is configured, the reviewer's decision is a
// request the consent screen never made, whatever password comes with it.
func TestTheReviewersDecisionIsRefusedWhereThereIsNoReviewer(t *testing.T) {
	t.Parallel()

	h := newSignIn(t)
	stopped := h.signInAsReviewer(h.browser(t), h.register(hostRedirect, hostName), reviewerPassword)

	if stopped.status != http.StatusBadRequest || stopped.location != "" {
		t.Fatalf("POST /oauth/consent = %d to %q, want 400 and a page", stopped.status, stopped.location)
	}
	if got := h.google.Renewals(); got != 0 {
		t.Errorf("Google renewed a grant %d times, want never", got)
	}
	if lines := h.lines(eventAuthReject); len(lines) != 1 || lines[0]["reason"] != "invalid_request" {
		t.Errorf("auth_reject lines = %v, want one with invalid_request", lines)
	}
}

// When the reviewer's code cannot be issued, the host hears it at its own
// address, and the line says why: a grant Google no longer honours, Google
// refusing the service's own client, or a code that cannot be sealed is the
// deployment's to mend and an error of ours; Google not answering is a
// warning.
func TestAReviewerWhoseCodeCannotBeIssuedIsToldSo(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		grant  string
		spoil  func(h *signIn)
		code   string
		reason string
		level  zapcore.Level
	}{
		{"a grant Google ended", "1//a-grant-google-ended", func(*signIn) {}, "server_error", "grant_ended", zapcore.ErrorLevel},
		{"Google not answering", googletest.RefreshToken, func(h *signIn) {
			h.google.Misbehave(&googletest.Answer{Status: http.StatusServiceUnavailable, ErrorCode: "busy"})
		}, "temporarily_unavailable", "google_unavailable", zapcore.WarnLevel},
		{"the service's client refused", googletest.RefreshToken, func(h *signIn) {
			h.google.Misbehave(&googletest.Answer{Status: http.StatusUnauthorized, ErrorCode: "invalid_client"})
		}, "server_error", "client_refused", zapcore.ErrorLevel},
		{"a code that cannot be sealed", googletest.RefreshToken, func(h *signIn) { h.server.flow.codes = brokenSealer{} },
			"server_error", "internal", zapcore.ErrorLevel},
		{"the demo account's renewals past their pace", googletest.RefreshToken, func(h *signIn) {
			paced := ratelimittest.Keyed(h.t, 1)
			paced.Take(h.server.DemoAccounts[0])
			h.server.flow.renewals = paced
		}, "temporarily_unavailable", limitRenewals, zapcore.WarnLevel},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reviewer := demoReviewer()
			reviewer.RefreshToken = tc.grant
			h := newSignInWithReviewer(t, reviewer)
			tc.spoil(h)
			answer := answerAt(t, h.signInAsReviewer(h.browser(t), h.register(hostRedirect, hostName), reviewerPassword))

			if answer.Get("error") != tc.code || answer.Get("error_description") == "" || answer.Has("code") {
				t.Errorf("the host was sent %v, want %s with a description and no code", answer, tc.code)
			}
			entries := h.logs.FilterMessage(eventAuthConsent).All()
			if len(entries) != 1 {
				t.Fatalf("auth_consent lines = %d, want one", len(entries))
			}
			fields := entries[0].ContextMap()
			if fields["outcome"] != "failed" || fields["reason"] != tc.reason || entries[0].Level != tc.level {
				t.Errorf("auth_consent = %v at %v, want failed with %s at %v", fields, entries[0].Level, tc.reason, tc.level)
			}
		})
	}
}

// A host that ends the demo account's grant ends nothing at Google, with
// either of its tokens: the reviewers' sign-in goes on renewing that grant,
// and Google would end it for every reviewer at once.
func TestTheDemoAccountsGrantIsNotEndedAtGoogle(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{kindRefresh, kindAccess} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			h := newSignInWithReviewer(t, demoReviewer())
			clientID := h.register(hostRedirect, hostName)
			access, refresh := h.reviewersTokens(t, clientID)
			answer := h.revoke(t, map[string]string{kindRefresh: refresh, kindAccess: access}[kind], clientID, nil)

			if answer.status != http.StatusOK {
				t.Fatalf("POST /oauth/revoke = %d %v, want 200", answer.status, answer.fields)
			}
			if got := h.google.Revocations(); len(got) != 0 {
				t.Errorf("Google was asked to end a grant with %q, want nothing asked", got)
			}
			if lines := h.lines(eventAuthRevoke); len(lines) != 1 || lines[0]["outcome"] != "ignored" ||
				lines[0]["reason"] != "demo_account" || lines[0]["token"] != kind || lines[0]["user"] != h.server.DemoAccounts[0] {
				t.Errorf("auth_revoke lines = %v, want one of the demo account's %s token ignored", lines, kind)
			}
			h.clock.advance(time.Hour)
			if renewed := h.renew(t, refresh, clientID, nil); renewed.status != http.StatusOK {
				t.Errorf("a renewal after the revocation = %d %v, want the grant renewed as before", renewed.status, renewed.fields)
			}
		})
	}
}

// A parent's grant still ends at Google beside a reviewer's sign-in.
func TestAParentsGrantEndsAtGoogleBesideTheReviewers(t *testing.T) {
	t.Parallel()

	h := newSignInWithReviewer(t, demoReviewer())
	clientID := h.register(hostRedirect, hostName)
	_, refresh := h.tokensFor(t, clientID)
	if answer := h.revoke(t, refresh, clientID, nil); answer.status != http.StatusOK {
		t.Fatalf("POST /oauth/revoke = %d %v, want 200", answer.status, answer.fields)
	}
	if got := h.google.Revocations(); !slices.Equal(got, []string{googletest.RefreshToken}) {
		t.Errorf("Google was asked to end a grant with %q, want the parent's refresh token", got)
	}
}

// The server names the demo account, so that whatever counts the children can
// leave its child out, and a reviewer's tokens sign in that very account at
// the resource; with no reviewer's sign-in configured, it names none.
func TestTheServerNamesTheDemoAccount(t *testing.T) {
	t.Parallel()

	h := newSignInWithReviewer(t, demoReviewer())
	if got, want := h.server.DemoAccounts, []string{h.ring.UserID("google-sub:" + demoSubject)}; !slices.Equal(got, want) {
		t.Errorf("DemoAccounts = %q, want %q", got, want)
	}
	access, _ := h.reviewersTokens(t, h.register(hostRedirect, hostName))
	if account, _, err := h.server.Account(t.Context(), access); err != nil || account.ID != h.server.DemoAccounts[0] {
		t.Errorf("Account() = %q, %v, want the demo account", account.ID, err)
	}
	if got := newSignIn(t).server.DemoAccounts; len(got) != 0 {
		t.Errorf("DemoAccounts with no reviewer's sign-in = %q, want none", got)
	}
}

// After a rotation of the keys, a reviewer signed in before it is still the
// demo account's: an instance of the new keys names the account by both its
// identifiers, and a revocation of the reviewer's tokens ends nothing at
// Google there either.
func TestAReviewerSignedInBeforeARotationIsStillTheDemoAccounts(t *testing.T) {
	t.Parallel()

	h := newSignInWithReviewer(t, demoReviewer())
	clientID := h.register(hostRedirect, hostName)
	_, refresh := h.reviewersTokens(t, clientID)
	before := h.server.DemoAccounts[0]

	rotated := h.rotatedWithReviewer(t)
	if got := rotated.DemoAccounts; len(got) != 2 || got[1] != before || got[0] == before {
		t.Fatalf("DemoAccounts after the rotation = %q, want the new identifier and %q", got, before)
	}
	if answer := serveForm(t, rotated.Revoke, url.Values{"token": {refresh}, "client_id": {clientID}}); answer.Code != http.StatusOK {
		t.Fatalf("POST /oauth/revoke = %d %s, want 200", answer.Code, answer.Body)
	}
	if got := h.google.Revocations(); len(got) != 0 {
		t.Errorf("Google was asked to end a grant with %q, want the demo account's left alone", got)
	}
}

// rotatedWithReviewer is the server of the case as another instance would be
// after a rotation of the keys — the same issuer, clock and Google, the key of
// the case kept as the previous one —, with the reviewer's sign-in of these
// cases.
func (h *signIn) rotatedWithReviewer(t *testing.T) *Server {
	t.Helper()

	google, err := googleauth.New(&googleauth.Settings{
		ClientID: googletest.ClientID, ClientSecret: googletest.ClientSecret, RedirectURL: h.served.URL + CallbackPath,
		Endpoints: h.google.Endpoints(), Now: h.clock.Now,
	})
	if err != nil {
		t.Fatalf("googleauth.New() error = %v, want nil", err)
	}
	server, err := New(&Settings{
		PublicURL: h.served.URL, Scope: "mcp", Seal: ringOfKeys(t, keyOf('n'), keyOf('k')), Documents: h.documents,
		Logger: zap.NewNop(), Renewals: ratelimittest.Roomy(t), Google: google, SiteURL: testSite,
		Reviewer: demoReviewer(), Now: h.clock.Now,
	})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	return server
}

// A reviewer's sign-in that names no password, no account or no grant could
// sign nobody in, and the server refuses to be built with it.
func TestAReviewersSignInMissingAPartIsRefused(t *testing.T) {
	t.Parallel()

	google := newSignIn(t).server.flow.google
	for _, tc := range []struct {
		name     string
		reviewer Reviewer
	}{
		{"no password", Reviewer{Subject: demoSubject, RefreshToken: googletest.RefreshToken}},
		{"no account", Reviewer{Password: reviewerPassword, RefreshToken: googletest.RefreshToken}},
		{"no grant", Reviewer{Password: reviewerPassword, Subject: demoSubject}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := Settings{Google: google, Reviewer: &tc.reviewer}
			if err := settings.validateReviewer(); !errors.Is(err, ErrSettings) {
				t.Errorf("validateReviewer() error = %v, want ErrSettings", err)
			}
		})
	}
	whole := Settings{Google: google, Reviewer: demoReviewer()}
	if err := whole.validateReviewer(); err != nil {
		t.Errorf("validateReviewer() of a whole sign-in error = %v, want nil", err)
	}
}

// The reviewer's password is the one password admitted: whatever else is
// typed — another string, the password with something more, or with something
// less — is not.
func TestOnlyTheReviewersPasswordIsAdmitted(t *testing.T) {
	t.Parallel()

	reviewer := newReviewerSignIn(demoReviewer(), func(string) []string { return []string{"demo"} })
	properties := gopter.NewProperties(nil)
	properties.Property("a password is admitted exactly when it is the reviewer's", prop.ForAll(
		func(typed string) bool { return reviewer.admits(typed) == (typed == reviewerPassword) },
		gen.OneGenOf(
			gen.Const(reviewerPassword),
			gen.AnyString(),
			gen.AnyString().Map(func(more string) string { return reviewerPassword + more }),
			gen.IntRange(0, len(reviewerPassword)).Map(func(cut int) string { return reviewerPassword[:cut] }),
		),
	))
	properties.TestingRun(t)
}
